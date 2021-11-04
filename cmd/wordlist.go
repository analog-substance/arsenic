package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/analog-substance/arsenic/lib/util"
	"github.com/spf13/cobra"
)

// wordlistCmd represents the wordlist command
var wordlistCmd = &cobra.Command{
	Use:       "wordlist (web-content)",
	Short:     "Generate a wordlist",
	Long:      `Generate a wordlist`,
	ValidArgs: []string{"web-content"},
	Args:      cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		generateWordlist(args[0])
	},
}

func generateWordlist(wordlistType string) {
	lineMap := make(map[string]bool)
	lines := []string{}

	for _, wordlistPath := range util.GetWordlists(wordlistType) {
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
			if _, ok := lineMap[line]; !ok {
				lines = append(lines, line)
			}
			lineMap[line] = true
		}
	}

	sort.Strings(lines)
	for _, line := range lines {
		fmt.Println(line)
	}
}

func cleanLine(wordlistType, line string) string {
	if wordlistType == "web-content" {
		re := regexp.MustCompile(`^(/+)`)
		line = re.ReplaceAllString(line, "")
	}
	return strings.TrimSpace(line)
}

func shouldIgnoreLine(wordlistType, line string) bool {
	if wordlistType == "web-content" {
		// this is why we can't have nice things
		re := regexp.MustCompile(`^(## Contribed by)|^/*(\?|\.$|#!?)|\.(gif|ico|jpe?g|png|js|css)$|^\^|\[[0-9a-zA-Z]\-[0-9a-zA-Z]\]|\*\.|\$$`)
		return re.MatchString(line)
	}
	return false
}

func init() {
	rootCmd.AddCommand(wordlistCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// wordlistCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// wordlistCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
