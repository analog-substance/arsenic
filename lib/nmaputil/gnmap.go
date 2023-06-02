package nmaputil

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/analog-substance/arsenic/lib/host"
)

func GnmapSplit(path string, name string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	outputName := fmt.Sprintf("%s.gnmap", name)

	doneRe := regexp.MustCompile(`# Nmap done`)
	hostRe := regexp.MustCompile(`Host: ([0-9\.]+)`)

	var currentHost *host.Host
	var lines []string
	var header []string
	ip := ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// If we hit this string, there are no more hosts
		if doneRe.MatchString(line) {
			err = writeToFile(currentHost, outputName, []byte(strings.Join(lines, "\n")))
			if err != nil {
				return err
			}
			break
		}

		match := hostRe.FindStringSubmatch(line)
		if len(match) == 0 {
			header = append(header, line)
			continue
		}

		if match[1] != ip {
			// Write nmap file for current host before starting a new host
			if currentHost != nil {
				err = writeToFile(currentHost, outputName, []byte(strings.Join(lines, "\n")))
				if err != nil {
					return err
				}
			}

			ip = match[1]

			currentHost, err = getHost([]string{}, []string{ip})
			if err != nil {
				return err
			}

			lines = append(lines[:0], header...)
		}

		lines = append(lines, line)
	}

	return nil
}
