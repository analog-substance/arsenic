package grep

import (
	"bufio"
	"os"
	"regexp"

	"github.com/analog-substance/arsenic/lib/util"
)

// LineByLine passes each line of the file matching the regex to the specified function
func LineByLine(path string, re *regexp.Regexp, action func(line string)) error {
	err := util.ReadLineByLine(path, func(line string) {
		if re.MatchString(line) {
			action(line)
		}
	})
	return err
}

// Matches returns maximum n number of matches from the file.
//   n > 0: at most n matches
//   n == 0: the result is nil (zero matches)
//   n < 0: all matches
func Matches(path string, re *regexp.Regexp, n int) []string {
	if n == 0 {
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
		if n > 0 && len(matches) == n {
			break
		}

		line := scanner.Text()
		if re.MatchString(line) {
			matches = append(matches, line)
		}
	}

	return matches
}

// Match returns true if any lines of the file match the regex
func Match(path string, re *regexp.Regexp) bool {
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
