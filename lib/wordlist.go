package lib

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/analog-substance/arsenic/lib/set"
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/spf13/viper"
)

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
		log.Println(err)
	}

	dirs := append([]string{cwd}, viper.GetStringSlice("wordlist-paths")...)

	wordlists := viper.GetStringSlice(fmt.Sprintf("wordlists.%s", wordlistType))
	for _, wordlist := range wordlists {
		for _, dir := range dirs {
			wordlistPath := path.Join(dir, wordlist)
			if util.FileExists(wordlistPath) {
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

func shouldIgnoreLine(wordlistType, line string) bool {
	if wordlistType == "web-content" || wordlistType == "sqli" || wordlistType == "xss" {
		// this is why we can't have nice things
		re := regexp.MustCompile(`^(## Contribed by)|^/*(\?|\.$|#!?)|\.(gif|ico|jpe?g|png|js|css)$|^\^|\[[0-9a-zA-Z]\-[0-9a-zA-Z]\]|\*\.|\$$`)
		return re.MatchString(line)
	}
	return false
}

func cleanLine(wordlistType, line string) string {
	if wordlistType == "web-content" {
		re := regexp.MustCompile(`^(/+)`)
		line = re.ReplaceAllString(line, "")
	}
	return strings.TrimSpace(line)
}
