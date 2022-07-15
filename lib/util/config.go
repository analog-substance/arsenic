package util

import (
	"strconv"
	"strings"
)

// Don't know what this should be called, if anyone has any better names, go for it
type IgnoreService struct {
	Name  string `yaml:"name"`
	Ports string `yaml:"ports"`
	Flag  string `yaml:"flag"` // This is so we can add a flag saying that this was ignored to make people aware that ports were ignored
}

func (s IgnoreService) checkPort(port int) bool {
	for _, p := range strings.Split(s.Ports, ",") {
		if strings.Contains(p, "-") {
			min, err := strconv.Atoi(strings.Split(p, "-")[0])
			if err != nil {
				return false
			}

			max, err := strconv.Atoi(strings.Split(p, "-")[1])
			if err != nil {
				return false
			}

			if port >= min && port <= max {
				return true
			}
		} else {
			if strings.EqualFold(p, "all") {
				return true
			}

			converted, err := strconv.Atoi(p)
			if err != nil {
				return false
			}

			if port == converted {
				return true
			}
		}
	}
	return false
}

func (s IgnoreService) ShouldIgnore(service string, port int) bool {
	if !strings.EqualFold(s.Name, service) {
		return false
	}
	return s.checkPort(port)
}
