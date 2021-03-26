package util

import (
	"bufio"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"syscall"
	"strings"
	"time"
)

func GetScripts(phase string) []string {
	scriptMap := GetPhaseMap(phase)

	scripts := []string{}
	for _, script := range scriptMap {
		scripts = append(scripts, script)
	}
	return scripts
}

func GetPhaseMap(phase string) map[string]string {
	scriptMap := make(map[string]string)
	for _, varDir := range viper.GetStringSlice("varDirs") {
		potentialVarFile := fmt.Sprintf("%s/%s/default.txt", varDir, phase)

		if _, err := os.Stat(potentialVarFile); !os.IsNotExist(err) {
			scriptsLines, err := ReadLines(potentialVarFile)
			if err != nil {
				fmt.Println(err)
			}

			sort.Strings(scriptsLines)

			for _, scriptLine := range scriptsLines {
				scriptOrder := scriptLine[:2]
				scriptName := scriptLine[3:]
				scriptMap[scriptOrder] = scriptName
			}
		}
	}

	return scriptMap
}

func Override(phase string) {
	validPhases := []string{"discover","recon","hunt"}
	for _, p := range validPhases {
		if phase == p {
			phaseMap := GetPhaseMap(phase)
			content := []string{}

			for phaseKey, phaseCmd := range phaseMap {
				content = append(content, fmt.Sprintf("%s %s", phaseKey, phaseCmd))
			}

			err := os.MkdirAll("as/var/", 0755)
			if err == nil {
				err = ioutil.WriteFile(fmt.Sprintf("as/var/%s/default.txt", phase), []byte(strings.Join(content[:], "\n")), 0644)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
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

func ExecutePhaseScripts(phase string) {
	scripts := GetScripts(phase)
	for len(scripts) > 0 {
		currentScript := scripts[0]
		fmt.Printf("Running %s\n", currentScript)
		args := []string{}
		if ExecScript(currentScript, args) == 0 {
			scripts = scripts[1:]
		} else {
			fmt.Printf("Script failed, gonna retry: %s\n", currentScript)
			time.Sleep(10 * time.Second)
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
