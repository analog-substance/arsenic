package util

import (
	"bufio"
	"fmt"

	// "io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"

	// "strings"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

type ScriptConfig struct {
	Script  string
	Order   int
	Count   int
	Enabled bool
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

func GetWordlists(wordlistType string) []string {
	wordlistPaths := []string{}

	cwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	dirs := append([]string{cwd}, viper.GetStringSlice("wordlist-paths")...)

	wordlists := viper.GetStringSlice(fmt.Sprintf("wordlists.%s", wordlistType))
	for _, wordlist := range wordlists {
		for _, dir := range dirs {
			wordlistPath := path.Join(dir, wordlist)
			if fileExists(wordlistPath) {
				wordlistPaths = append(wordlistPaths, wordlistPath)
				break
			}
		}
	}

	return wordlistPaths
}

func ExecScript(scriptPath string, args []string) int {
	cmd := exec.Command(scriptPath, args...)

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

func ExecutePhaseScripts(phase string, args []string, dryRun bool) {
	phaseScripts := GetScripts(phase)
	for i := 0; ; i++ {
		scripts := phaseScripts
		scriptsRan := false
		for len(scripts) > 0 {
			currentScript := scripts[0]
			if currentScript.Enabled && currentScript.Count > i {
				fmt.Printf("Running %s %d\n", currentScript.Script, i)
				scriptsRan = true
				if dryRun {
					scripts = scripts[1:]
				} else {
					if ExecScript(currentScript.Script, args) == 0 {
						scripts = scripts[1:]
					} else {
						fmt.Printf("Script failed, gonna retry: %s\n", currentScript.Script)
						time.Sleep(10 * time.Second)
					}
				}
			} else {
				// not enabled.. remove
				scripts = scripts[1:]
			}
		}
		if !scriptsRan {
			break
		}
	}
}

func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

type NoopWriter struct {
}

func (w NoopWriter) Write(bytes []byte) (int, error) {
	return 0, nil
}
