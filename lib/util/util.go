package util

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

type ScriptConfig struct {
	Script  string
	Order   int
	Count   int
	Enabled bool

	totalRuns int
}

func NewScriptConfig(script string, order int, count int, enabled bool) ScriptConfig {
	return ScriptConfig{
		Script:  script,
		Order:   order,
		Enabled: enabled,
		Count:   count,
	}
}

func GetScripts(phase string) []ScriptConfig {
	scripts := map[string]ScriptConfig{}
	viper.UnmarshalKey(fmt.Sprintf("scripts.%s", phase), &scripts)
	phaseScripts := []ScriptConfig{}
	for _, scriptConfig := range scripts {
		phaseScripts = append(phaseScripts, scriptConfig)
	}

	sort.SliceStable(phaseScripts, func(i, j int) bool {
		return phaseScripts[i].Order < phaseScripts[j].Order
	})
	return phaseScripts
}

func ExecScript(scriptPath string, args []string) int {
	cmdCtx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(cmdCtx, scriptPath, args...)

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
		log.Fatalf("cmd.Start: %v", err)
	}

	exitStatus := 0
	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if status.Signaled() {
					terminate = true
				}
				exitStatus = status.ExitStatus()
			}
		} else {
			log.Fatalf("cmd.Wait: %v", err)
		}
	}

	if terminate {
		os.Exit(exitStatus)
	}
	return exitStatus
}

func iteratePhaseScripts(phase string, done chan int) chan ScriptConfig {
	scriptChan := make(chan ScriptConfig)
	go func() {
		defer func() {
			close(scriptChan)
			<-done // Ensure we don't hang if done is sent something after very last script
		}()

		scripts := GetScripts(phase)
		if len(scripts) == 0 {
			return
		}

		for _, script := range scripts {
			if !script.Enabled {
				continue
			}

			for script.Count > script.totalRuns {
				script.totalRuns++

				select {
				case scriptChan <- script:
				case <-done:
					return
				}
			}
		}
	}()
	return scriptChan
}

func executePhaseScripts(phase string, args []string, dryRun bool) (bool, string) {
	done := make(chan int)
	defer close(done)

	scriptChan := iteratePhaseScripts(phase, done)
	for script := range scriptChan {
		fmt.Printf("Running %s %d\n", script.Script, script.totalRuns)
		if dryRun {
			continue
		}

		if ExecScript(script.Script, args) != 0 {
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

func StringSliceEquals(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func GetReviewer(reviewerFlag string) string {
	if reviewerFlag == "operator" {
		envReviewer := os.Getenv("AS_REVIEWER")
		envUser := os.Getenv("USER")
		if len(envReviewer) > 0 {
			reviewerFlag = envReviewer
		} else if len(envUser) > 0 {
			reviewerFlag = envUser
		}
	}

	return reviewerFlag
}

type NoopWriter struct {
}

func (w NoopWriter) Write(bytes []byte) (int, error) {
	return 0, nil
}

func ToString(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
func ToStringSlice(v interface{}) []string {
	switch v := v.(type) {
	case []string:
		return v
	case []interface{}:
		b := make([]string, 0, len(v))
		for _, s := range v {
			if s != nil {
				b = append(b, ToString(s))
			}
		}
		return b
	default:
		val := reflect.ValueOf(v)
		switch val.Kind() {
		case reflect.Array, reflect.Slice:
			l := val.Len()
			b := make([]string, 0, l)
			for i := 0; i < l; i++ {
				value := val.Index(i).Interface()
				if value != nil {
					b = append(b, ToString(value))
				}
			}
			return b
		default:
			if v == nil {
				return []string{}
			}

			return []string{ToString(v)}
		}
	}
}

func IndexOf(data []string, item string) int {
	for k, v := range data {
		if item == v {
			return k
		}
	}
	return -1
}

func RemoveIndex(arr []string, idx int) []string {
	return append(arr[:idx], arr[idx+1:]...)
}

func Sanitize(s string) string {
	// Windows is most restrictive
	windows_regex := regexp.MustCompile("[<>:/\\|?*\"]+")
	s = windows_regex.ReplaceAllString(s, "_")
	return strings.TrimSpace(s)
}

func IsIp(ipOrHostname string) bool {
	if net.ParseIP(ipOrHostname) == nil {
		return false
	} else {
		return true
	}
}

func LogMsg(args ...interface{}) {
	Log("[+]", args...)
}
func LogWarn(args ...interface{}) {
	Log("[!]", args...)
}
func LogInfo(args ...interface{}) {
	Log("[-]", args...)
}

func Log(prefix string, args ...interface{}) {
	fmt.Printf("%s ", prefix)
	fmt.Print(args...)
	fmt.Println()
}
