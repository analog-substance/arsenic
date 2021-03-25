package util


import (
	"bufio"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"syscall"
	"sort"
	"time"
)

func GetScripts(phase string) []string {
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

func ExecutePhaseScripts (phase string) {
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
