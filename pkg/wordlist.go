package pkg

import (
	"bufio"
	"github.com/analog-substance/arsenic/pkg/log"
	"github.com/analog-substance/util/fileutil"
	"io"
	"log/slog"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/analog-substance/arsenic/pkg/config"
	"github.com/analog-substance/util/set"
)

var logger *slog.Logger

func init() {
	logger = log.WithGroup("wordlists")
}

var validWordlistTypes []string

func GenerateWordlist(wordlistType string, lineSet *set.Set) {
	for _, wordlistPath := range GetWordlists(wordlistType) {
		if readWordlistFile(wordlistType, lineSet, wordlistPath) {
			return
		}
	}
}

func GetWordlists(wordlistType string) []string {
	var wordlistPaths []string

	cwd, err := os.Getwd()
	if err != nil {
		logger.Error(err.Error())
	}

	c := config.Get()
	dirs := append([]string{cwd}, c.Wordlists.Paths...)

	wordlists := c.Wordlists.Types[wordlistType]
	for _, wordlist := range wordlists {
		for _, dir := range dirs {
			wordlistPath := path.Join(dir, wordlist)
			if fileutil.FileExists(wordlistPath) {
				wordlistPaths = append(wordlistPaths, wordlistPath)
				break
			}
		}
	}

	return wordlistPaths
}

func readWordlistFile(wordlistType string, lineSet *set.Set, wordlistPath string) bool {
	file, err := os.Open(wordlistPath)
	if err != nil {
		return true
	}
	defer file.Close()

	return readWordlist(wordlistType, lineSet, file)
}

func readWordlist(wordlistType string, lineSet *set.Set, reader io.Reader) bool {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		rawLine := scanner.Text()

		if shouldIgnoreLine(wordlistType, rawLine) {
			continue
		}

		line := cleanLine(wordlistType, rawLine)
		lineSet.Add(line)
	}
	return false
}

func cleanLine(wordlistType, line string) string {
	if wordlistType == "web-content" {
		re := regexp.MustCompile(`^(/+)`)
		line = re.ReplaceAllString(line, "")
	} else if wordlistType == "subdomains" {
		re := regexp.MustCompile(`^\*\.`)
		line = re.ReplaceAllString(line, "")
		line = strings.ToLower(line)
	}
	return strings.TrimSpace(line)
}

func shouldIgnoreLine(wordlistType, line string) bool {
	if IsValidWordlistType(wordlistType) {
		// this is why we can't have nice things
		re := regexp.MustCompile(`^(## Contribed by)|^/*(\?|\.$|#!?)|\.(gif|ico|jpe?g|png|js|css)$|^\^|\[[0-9a-zA-Z]\-[0-9a-zA-Z]\]|\*\.|\$$`)
		return re.MatchString(line)
	}
	return false
}

func GetValidWordlistTypes() []string {
	if len(validWordlistTypes) == 0 {
		for wordlist := range config.Get().Wordlists.Types {
			validWordlistTypes = append(validWordlistTypes, wordlist)
		}
		sort.Strings(validWordlistTypes)
	}

	return validWordlistTypes
}

func IsValidWordlistType(wordlistType string) bool {
	wordlistTypes := GetValidWordlistTypes()
	for _, validType := range wordlistTypes {
		if wordlistType == validType {
			return true
		}
	}
	return false
}
