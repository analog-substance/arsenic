package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/defektive/arsenic/lib/util"
	"github.com/spf13/cobra"
)

// scopeCmd represents the scope command
var scopeCmd = &cobra.Command{
	Use:   "scope",
	Short: "Print all scope",
	Long:  `Print all scope`,
	Run: func(cmd *cobra.Command, args []string) {
		domains, _ := getScope("domains")
		ips, _ := getScope("ips")

		scope := append(domains, ips...)
		for _, scopeItem := range scope {
			fmt.Println(scopeItem)
		}
	},
}

func init() {
	rootCmd.AddCommand(scopeCmd)
}

func getScope(scopeType string) ([]string, error) {

	glob := fmt.Sprintf("scope-%s-*", scopeType)
	actualFile := fmt.Sprintf("scope-%s.txt", scopeType)
	blacklistFile := fmt.Sprintf("blacklist-%s.txt", scopeType)

	var blacklistRegexp []*regexp.Regexp
	if util.FileExists(blacklistFile) {
		lines, _ := util.ReadLines(blacklistFile)
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
		file, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := normalizeScope(scanner.Text(), scopeType)
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
		}
	}

	// now lets open the actual scope file and add those. since they cant be blacklisted
	file, err := os.Open(actualFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := normalizeScope(scanner.Text(), scopeType)
		scope[line] = true
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
