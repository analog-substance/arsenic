package config

import (
	"fmt"
	"sort"
)

type Scripts struct {
	Directory string           `yaml:"directory"`
	Phases    map[string]Phase `yaml:"phases"`
}

type Phase struct {
	Args    string            `yaml:"args"`
	Scripts map[string]Script `yaml:"scripts"`
}

type Script struct {
	Args    string `yaml:"args"`
	Script  string `yaml:"script"`
	Order   int    `yaml:"order"`
	Count   int    `yaml:"count"`
	Enabled bool   `yaml:"enabled"`

	TotalRuns int `yaml:"-"`
}

func NewScript(script string, order int, count int, enabled bool) Script {
	return Script{
		Script:  script,
		Order:   order,
		Enabled: enabled,
		Count:   count,
	}
}

func GetScripts(phase string) []Script {
	var scripts []Script
	p := Get().Scripts.Phases[phase]
	for _, script := range p.Scripts {
		script.Args = fmt.Sprintf("%s %s", p.Args, script.Args)
		scripts = append(scripts, script)
	}

	sort.SliceStable(scripts, func(i, j int) bool {
		return scripts[i].Order < scripts[j].Order
	})
	return scripts
}

func IteratePhaseScripts(phase string, done chan int) chan Script {
	scriptChan := make(chan Script)
	go func() {
		defer func() {
			close(scriptChan)
			<-done // Ensure we don't hang if done is sent something after very last script
		}()

		scripts := GetScripts(phase)
		if len(scripts) == 0 {
			return
		}

		for _, script := range scripts {
			if !script.Enabled {
				continue
			}

			for script.Count > script.TotalRuns {
				script.TotalRuns++

				select {
				case scriptChan <- script:
				case <-done:
					return
				}
			}
		}
	}()
	return scriptChan
}
