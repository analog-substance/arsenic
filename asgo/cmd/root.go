package cmd

import (
	"bufio"
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"syscall"
	"sort"
	"time"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "asgo",
	Short: "Arsenic - Pentest Conventions",
	Long: `Arsenic - Pentest Conventions


`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.asgo.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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
	    log.Println(err)
	}

	defaultScriptDirs := []string{}
	defaultScriptDirs = append(defaultScriptDirs, "/opt/arsenic/scripts/")
	defaultScriptDirs = append(defaultScriptDirs, fmt.Sprintf("%s/opt/arsenic/scripts/", home))
	defaultScriptDirs = append(defaultScriptDirs, fmt.Sprintf("%s/asgo/scripts/", cwd))

	viper.SetDefault("scriptDirs", defaultScriptDirs)

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {

		// Search config in home directory with name ".as" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".asgo")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func getScripts(phase string) []string {
	scriptFile := make(map[string]string)
	for _, scriptDir := range(viper.GetStringSlice("scriptDirs")) {
		potentialScriptDir := fmt.Sprintf("%s/%s", scriptDir, phase)
		if _, err := os.Stat(potentialScriptDir); !os.IsNotExist(err) {
			files, err := ioutil.ReadDir(potentialScriptDir)
			if err != nil {
				fmt.Println(err)
			}

			for _, file := range files {
				scriptFile[file.Name()] = fmt.Sprintf("%s/%s", potentialScriptDir, file.Name())
			}
		}
	}

	filePaths := []string{}
	for _, file := range scriptFile {
		filePaths = append(filePaths, file)
	}

	sort.Strings(filePaths)
	return filePaths
}

func execScript(scriptPath string) int {
	cmd := exec.Command(scriptPath)

	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		log.Fatalf("cmd.Start: %v", err)
	}

	scannerEr := bufio.NewScanner(stderr)
	scannerEr.Split(bufio.ScanLines)
	go func() {
		for scannerEr.Scan() {
			m := scannerEr.Text()
			fmt.Println(m)
		}
	}()

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	go func() {
		for scanner.Scan() {
			m := scanner.Text()
			fmt.Println(m)
		}
	}()

	exitStatus := 0
	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitStatus = status.ExitStatus()
			}
		} else {
			log.Fatalf("cmd.Wait: %v", err)
		}
	}
	return exitStatus
}

func executePhaseScripts (phase string) {
	scripts := getScripts(phase)
	for len(scripts) > 0 {
		currentScript := scripts[0]
		if execScript(currentScript) == 0 {
			scripts = scripts[1:]
		} else {
			fmt.Printf("Script failed, gonna retry: %s", currentScript)
			time.Sleep(10 * time.Second)
		}
	}
	for _, reconScript := range getScripts(phase) {
		fmt.Printf("Running %s\n", reconScript)
	}
}
