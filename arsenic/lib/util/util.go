package util

import (
	"bufio"
	"fmt"
	"reflect"

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
	Enabled bool
}

func NewScriptConfig(script string, order int, enabled bool) ScriptConfig {
	return ScriptConfig{
		Script:  script,
		Order:   order,
		Enabled: enabled,
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

func ExecutePhaseScripts(phase string, args []string) {
	scripts := GetScripts(phase)
	for len(scripts) > 0 {
		currentScript := scripts[0]
		if currentScript.Enabled {
			fmt.Printf("Running %s\n", currentScript.Script)
			if ExecScript(currentScript.Script, args) == 0 {
				scripts = scripts[1:]
			} else {
				fmt.Printf("Script failed, gonna retry: %s\n", currentScript.Script)
				time.Sleep(10 * time.Second)
			}
		} else {
			// not enabled.. remove
			scripts = scripts[1:]
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

func Any(x interface{}, predicate func(item interface{}) bool) bool {
	xValue := reflect.ValueOf(x)
	if xValue.Kind() != reflect.Slice {
		return false
	}

	length := xValue.Len()
	for i := 0; i < length; i++ {
		value := xValue.Index(i).Interface()
		if predicate(value) {
			return true
		}
	}
	return false
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

type StringSet struct {
	Set map[string]bool
}

func NewStringSet() StringSet {
	return StringSet{Set: map[string]bool{}}
}

// Add an element to a set
func (set *StringSet) Add(s string) bool {
	_, found := set.Set[s]
	set.Set[s] = true
	return !found
}

// AddRange adds a list of elements to a set
func (set *StringSet) AddRange(ss []string) {
	for _, s := range ss {
		set.Set[s] = true
	}
}

// Contains tests if an element is in a set
func (set *StringSet) Contains(s string) bool {
	_, found := set.Set[s]
	return found
}

// ContainsAny checks if any of the elements exist
func (set *StringSet) ContainsAny(ss []string) bool {
	for _, s := range ss {
		if set.Set[s] {
			return true
		}
	}
	return false
}

// Length returns the length of the Set
func (set *StringSet) Length() int {
	return len(set.Set)
}

// Slice returns the set as a slice
func (set *StringSet) Slice() []string {
	values := make([]string, len(set.Set))
	i := 0
	for s := range set.Set {
		values[i] = s
		i++
	}
	return values
}
