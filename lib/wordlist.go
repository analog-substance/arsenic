package lib

import (
	"bufio"
	"fmt"
	"github.com/analog-substance/arsenic/lib/set"
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/spf13/viper"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

func GenerateWordlist(wordlistType string, lineSet *set.Set) {

	for _, wordlistPath := range GetWordlists(wordlistType) {
		file, err := os.Open(wordlistPath)
		if err != nil {
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			rawLine := scanner.Text()

			if shouldIgnoreLine(wordlistType, rawLine) {
				continue
			}

			line := cleanLine(wordlistType, rawLine)
			lineSet.Add(line)
		}
	}
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
			if util.FileExists(wordlistPath) {
				wordlistPaths = append(wordlistPaths, wordlistPath)
				break
			}
		}
	}

	return wordlistPaths
}

func shouldIgnoreLine(wordlistType, line string) bool {
	if wordlistType == "web-content" {
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
