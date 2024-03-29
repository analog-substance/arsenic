package config

import (
	"fmt"
	"sort"
)

type Scripts struct {
	Directory string           `yaml:"directory" mapstructure:"directory"`
	Phases    map[string]Phase `yaml:"phases" mapstructure:"phases"`
}

type Phase struct {
	Args    string            `yaml:"args" mapstructure:"args"`
	Scripts map[string]Script `yaml:"scripts" mapstructure:"scripts"`
}

type Script struct {
	Args    string `yaml:"args" mapstructure:"args"`
	Script  string `yaml:"script" mapstructure:"script"`
	Order   int    `yaml:"order" mapstructure:"order"`
	Count   int    `yaml:"count" mapstructure:"count"`
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`

	TotalRuns int `yaml:"-" mapstructure:"-"`
}

func NewScript(script string, order int, count int, enabled bool) Script {
	return Script{
		Script:  script,
		Order:   order,
		Enabled: enabled,
		Count:   count,
	}
}

func (c *Config) GetScripts(phase string) []Script {
	var scripts []Script
	p := c.Scripts.Phases[phase]
	for _, script := range p.Scripts {
		if p.Args != "" && script.Args != "" {
			script.Args = fmt.Sprintf("%s %s", p.Args, script.Args)
		} else if p.Args != "" {
			script.Args = p.Args
		}

		scripts = append(scripts, script)
	}

	sort.SliceStable(scripts, func(i, j int) bool {
		return scripts[i].Order < scripts[j].Order
	})
	return scripts
}

func (c *Config) IterateScripts(phase string, done chan int) chan Script {
	scriptChan := make(chan Script)
	go func() {
		defer func() {
			close(scriptChan)
			<-done // Ensure we don't hang if done is sent something after very last script
		}()

		scripts := c.GetScripts(phase)
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
