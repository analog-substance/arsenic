package cmd

import (
	"context"
	"fmt"
	"github.com/analog-substance/arsenic/pkg/log"
	"log/slog"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/google/shlex"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/analog-substance/arsenic/pkg/config"
	"github.com/analog-substance/arsenic/pkg/engine"
)

var logger *slog.Logger

var cfgFile string
var configInitialized bool = false
var debug bool = false

var RootCmd = &cobra.Command{
	Use:     "arsenic",
	Version: "v0.2.0",
	Short:   "Pentesting conventions",
	Long: `Arsenic - Pentest Conventions


`,
	Args: cobra.ArbitraryArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if debug {
			log.LogLevel(slog.LevelDebug)
		}
		logger.Debug("debug mode", "debug", debug)
		return setOrRefreshConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}

		script, err := engine.NewScript(args[0])
		cobra.CheckErr(err)

		err = script.Run(args[1:])
		if err != nil && err != context.Canceled {
			fmt.Println(err)
			os.Exit(255)
		}
	},
}

func Execute() {
	if len(os.Args) > 1 && strings.Contains(os.Args[1], fmt.Sprintf("%c", os.PathSeparator)) {
		RootCmd.DisableFlagParsing = true
	}

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	logger = log.WithGroup("arsenic.cmd")

	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "the arsenic.yaml config file")
	RootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "the arsenic.yaml config file")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		logger.Error("error getting working dir", "err", err)
	}

	setConfigDefault("", config.Default(home))

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {

		// Search config in home directory with name "arsenic" (without extension).
		viper.AddConfigPath(cwd)
		viper.AddConfigPath(home)
		viper.SetConfigName("arsenic")
	}

	viper.AutomaticEnv() // read in environment variables that match
	viper.ReadInConfig()

	configInitialized = true
}

// If no config file exists, all possible keys in the defaults
// need to be registered with viper otherwise viper will only think
// the keys explicitly set via viper.SetDefault() exist.
func setConfigDefault(key string, value interface{}) {
	valueType := reflect.TypeOf(value)
	valueValue := reflect.ValueOf(value)

	if valueType.Kind() == reflect.Map {
		iter := valueValue.MapRange()
		for iter.Next() {
			k := iter.Key().Interface()
			v := iter.Value().Interface()
			setConfigDefault(fmt.Sprintf("%s.%s", key, k), v)
		}
	} else if valueType.Kind() == reflect.Struct {
		numFields := valueType.NumField()
		for i := 0; i < numFields; i++ {
			structField := valueType.Field(i)
			fieldValue := valueValue.Field(i)

			if !structField.IsExported() {
				continue
			}

			subKey := structField.Name
			yamlTag := structField.Tag.Get("yaml")
			if yamlTag == "-" {
				continue
			}

			if yamlTag != "" {
				subKey = yamlTag
			}

			fullKey := fmt.Sprintf("%s.%s", key, subKey)
			if key == "" {
				fullKey = subKey
			}

			setConfigDefault(fullKey, fieldValue.Interface())
		}
	} else {
		viper.SetDefault(key, value)
	}
}

func executePhaseScripts(phase string, args []string, dryRun bool) (bool, string) {
	done := make(chan int)
	defer close(done)

	scripts := config.Get().IterateScripts(phase, done)
	for script := range scripts {
		fmt.Printf("Running %s %d\n", script.Script, script.TotalRuns)
		if dryRun {
			continue
		}

		if ExecScript(script, args) != nil {
			done <- 1
			return false, script.Script
		}
	}

	return true, ""
}

func ExecutePhaseScripts(phase string, args []string, dryRun bool) {
	minWait := 10
	maxWait := 60

	for {
		status, script := executePhaseScripts(phase, args, dryRun)
		if status {
			return
		}

		fmt.Printf("Script failed, gonna retry: %s\n", script)

		timeToSleep := rand.Intn(maxWait-minWait) + minWait
		time.Sleep(time.Duration(timeToSleep) * time.Second)
	}
}

func ExecScript(script config.Script, args []string) error {
	if script.Args != "" {
		scriptArgs, err := shlex.Split(script.Args)
		if err != nil {
			return err
		}

		args = append(scriptArgs, args...)
	}

	scriptPath := script.Script
	if filepath.Ext(scriptPath) == ".tengo" {
		s, err := engine.NewScript(scriptPath)
		if err != nil {
			return err
		}

		err = s.Run(args)
		if err != nil && err != context.Canceled {
			return err
		}

		return nil
	}

	cmdCtx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(cmdCtx, scriptPath, args...)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// relay trapped signals to the spawned process
	terminate := false
	go func() {
		for sig := range sigs {
			terminate = true
			cmd.Process.Signal(sig)
			cancel()
		}
	}()

	defer func() {
		signal.Stop(sigs)
		close(sigs)
	}()

	if err := cmd.Start(); err != nil {
		logger.Error("cmd.Start err", "err", err)
		os.Exit(2)
	}

	var err error
	exitStatus := 0
	if err = cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if status.Signaled() {
					terminate = true
				}
				exitStatus = status.ExitStatus()
			}
		} else {
			logger.Error("cmd.Wait err", "err", err)
			os.Exit(2)
		}
	}

	if terminate {
		os.Exit(exitStatus)
	}
	return err
}
