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
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/ahmetb/go-linq/v3"
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

func Mkdir(dirs ...string) []error {
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

func IsDomainInScope(domain string) bool {
	scope, err := GetScope("domains")
	if err != nil {
		return false
	}

	return linq.From(scope).AnyWith(func(i interface{}) bool {
		return strings.EqualFold(i.(string), domain)
	})
}

func IsIpInScope(ip string) bool {
	scope, err := GetScope("ips")
	if err != nil {
		return false
	}

	return linq.From(scope).AnyWith(func(i interface{}) bool {
		return strings.EqualFold(i.(string), ip)
	})
}

func GetScope(scopeType string) ([]string, error) {

	glob := fmt.Sprintf("scope-%s-*", scopeType)
	actualFile := fmt.Sprintf("scope-%s.txt", scopeType)
	blacklistFile := fmt.Sprintf("blacklist-%s.txt", scopeType)

	var blacklistRegexp []*regexp.Regexp
	if FileExists(blacklistFile) {
		lines, _ := ReadLines(blacklistFile)
		for _, line := range lines {
			if line == "" {
				continue
			}
			blacklistRegexp = append(blacklistRegexp, regexp.MustCompile(regexp.QuoteMeta(line)))
		}
	}

	files, _ := filepath.Glob(glob)
	scope := make(map[string]bool)

	for _, filename := range files {
		err := ReadLineByLine(filename, func(line string) {
			line = normalizeScope(line, scopeType)
			valid := true
			for _, re := range blacklistRegexp {
				if re.MatchString(line) {
					valid = false
					break
				}
			}
			if valid {
				scope[line] = true
			}
		})
		if err != nil {
			return nil, err
		}
	}

	// now lets open the actual scope file and add those. since they cant be blacklisted
	err := ReadLineByLine(actualFile, func(line string) {
		line = normalizeScope(line, scopeType)
		scope[line] = true
	})

	if err != nil {
		return nil, err
	}

	var scopeAr []string
	for scopeItem, _ := range scope {
		scopeAr = append(scopeAr, scopeItem)
	}

	sort.Strings(scopeAr)
	return scopeAr, nil
}

func normalizeScope(scopeItem, scopeType string) string {

	if scopeType == "domains" {
		re := regexp.MustCompile(`^\*\.`)
		scopeItem = re.ReplaceAllString(scopeItem, "")
	}

	return scopeItem
}
