package util

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/net/publicsuffix"
)

const (
	DefaultDirPerms  fs.FileMode = 0755
	DefaultFilePerms fs.FileMode = 0644
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
			if FileExists(wordlistPath) {
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

func executePhaseScripts(phase string, args []string, dryRun bool) (bool, string) {
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
						return false, currentScript.Script
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

func ReadLineByLine(path string, action func(line string)) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		action(scanner.Text())
	}
	return nil
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func Mkdirs(dirs ...string) []error {
	var errors []error
	for _, dir := range dirs {
		err := os.MkdirAll(dir, DefaultDirPerms)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func WriteLines(path string, lines []string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, DefaultFilePerms)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, data := range lines {
		_, _ = writer.WriteString(data + "\n")
	}

	writer.Flush()
	return nil
}

func GrepLineByLine(path string, re *regexp.Regexp, action func(line string)) error {
	err := ReadLineByLine(path, func(line string) {
		if re.MatchString(line) {
			action(line)
		}
	})
	return err
}

func GrepLines(path string, re *regexp.Regexp, count int) []string {
	if count == 0 {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var matches []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if count > 0 && len(matches) == count {
			break
		}

		line := scanner.Text()
		if re.MatchString(line) {
			matches = append(matches, line)
		}
	}

	return matches
}

func GrepMatch(path string, re *regexp.Regexp) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if re.MatchString(scanner.Text()) {
			return true
		}
	}
	return false
}

type NoopWriter struct {
}

func (w NoopWriter) Write(bytes []byte) (int, error) {
	return 0, nil
}

func GetRootDomains(domains []string, pruneBlacklisted bool) []string {
	blacklistedRootDomains := viper.GetStringSlice("blacklist.root-domains")
	rootDomainMap := map[string]int{}
	rootDomains := []string{}
	for _, domain := range domains {
		rootDomain, _ := publicsuffix.EffectiveTLDPlusOne(domain)

		if len(rootDomain) > 0 {
			rootDomainMap[rootDomain] = 1
		}
	}

	for rootDomain := range rootDomainMap {
		addRootDomain := true
		if pruneBlacklisted {
			for _, badRootDomain := range blacklistedRootDomains {
				if strings.EqualFold(badRootDomain, rootDomain) {
					addRootDomain = false
					break
				}
			}
		}

		if addRootDomain {
			rootDomains = append(rootDomains, rootDomain)
		}
	}
	sort.Strings(rootDomains)
	return rootDomains
}
