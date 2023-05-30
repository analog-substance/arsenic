package config

import (
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

func setConfig() {
	var c Config
	viper.Unmarshal(&c)
	Set(&c)
}

func TestGetScripts(t *testing.T) {
	type args struct {
		phase string
	}
	tests := []struct {
		name  string
		setup func()
		args  args
		want  []Script
	}{
		{
			name: "Sorted by order",
			setup: func() {
				viper.Set("scripts.test", map[string]Script{
					"script-1": {
						Script: "script-1-name",
						Order:  1,
					},
					"script-2": {
						Script: "script-2-name",
						Order:  5,
					},
					"script-3": {
						Script: "script-3-name",
						Order:  6,
					},
					"script-4": {
						Script: "script-4-name",
						Order:  2,
					},
					"script-5": {
						Script: "script-5-name",
						Order:  2,
					},
				})

				setConfig()
			},
			args: args{phase: "test"},
			want: []Script{
				{
					Script: "script-1-name",
					Order:  1,
				},
				{
					Script: "script-4-name",
					Order:  2,
				},
				{
					Script: "script-5-name",
					Order:  2,
				},
				{
					Script: "script-2-name",
					Order:  5,
				},
				{
					Script: "script-3-name",
					Order:  6,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			if got := GetScripts(tt.args.phase); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetScripts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_IteratePhaseScripts(t *testing.T) {
	type args struct {
		phase string
		done  chan int
	}
	tests := []struct {
		name   string
		setup  func()
		args   args
		doneAt int
		want   []Script
	}{
		{
			name: "Normal iteration",
			setup: func() {
				viper.Set("scripts.test", map[string]Script{
					"script-1": {
						Script:  "script-1-name",
						Enabled: true,
						Order:   1,
						Count:   3,
					},
					"script-2": {
						Script:  "script-2-name",
						Enabled: true,
						Order:   5,
						Count:   1,
					},
				})

				setConfig()
			},
			args: args{
				phase: "test",
				done:  make(chan int),
			},
			want: []Script{
				{
					Script:    "script-1-name",
					Enabled:   true,
					Order:     1,
					Count:     3,
					TotalRuns: 1,
				},
				{
					Script:    "script-1-name",
					Enabled:   true,
					Order:     1,
					Count:     3,
					TotalRuns: 2,
				},
				{
					Script:    "script-1-name",
					Enabled:   true,
					Order:     1,
					Count:     3,
					TotalRuns: 3,
				},
				{
					Script:    "script-2-name",
					Enabled:   true,
					Order:     5,
					Count:     1,
					TotalRuns: 1,
				},
			},
		},
		{
			name: "Some disabled or 0 count",
			setup: func() {
				viper.Set("scripts.test", map[string]Script{
					"script-1": {
						Script:  "script-1-name",
						Enabled: true,
						Order:   1,
						Count:   3,
					},
					"script-2": {
						Script:  "script-2-name",
						Enabled: false,
						Order:   2,
						Count:   1,
					},
					"script-3": {
						Script:  "script-3-name",
						Enabled: true,
						Order:   3,
						Count:   0,
					},
					"script-4": {
						Script:  "script-4-name",
						Enabled: true,
						Order:   4,
						Count:   1,
					},
				})
				setConfig()
			},
			args: args{
				phase: "test",
				done:  make(chan int),
			},
			want: []Script{
				{
					Script:    "script-1-name",
					Enabled:   true,
					Order:     1,
					Count:     3,
					TotalRuns: 1,
				},
				{
					Script:    "script-1-name",
					Enabled:   true,
					Order:     1,
					Count:     3,
					TotalRuns: 2,
				},
				{
					Script:    "script-1-name",
					Enabled:   true,
					Order:     1,
					Count:     3,
					TotalRuns: 3,
				},
				{
					Script:    "script-4-name",
					Enabled:   true,
					Order:     4,
					Count:     1,
					TotalRuns: 1,
				},
			},
		},
		{
			name: "Done in middle",
			setup: func() {
				viper.Set("scripts.test", map[string]Script{
					"script-1": {
						Script:  "script-1-name",
						Enabled: true,
						Order:   1,
						Count:   3,
					},
					"script-2": {
						Script:  "script-2-name",
						Enabled: false,
						Order:   2,
						Count:   1,
					},
					"script-3": {
						Script:  "script-3-name",
						Enabled: true,
						Order:   3,
						Count:   0,
					},
					"script-4": {
						Script:  "script-4-name",
						Enabled: true,
						Order:   4,
						Count:   1,
					},
				})
				setConfig()
			},
			args: args{
				phase: "test",
				done:  make(chan int),
			},
			doneAt: 2,
		},
		{
			name: "Done at end",
			setup: func() {
				viper.Set("scripts.test", map[string]Script{
					"script-1": {
						Script:  "script-1-name",
						Enabled: true,
						Order:   1,
						Count:   3,
					},
					"script-2": {
						Script:  "script-2-name",
						Enabled: false,
						Order:   2,
						Count:   1,
					},
					"script-3": {
						Script:  "script-3-name",
						Enabled: true,
						Order:   3,
						Count:   0,
					},
					"script-4": {
						Script:  "script-4-name",
						Enabled: true,
						Order:   4,
						Count:   1,
					},
				})
				setConfig()
			},
			args: args{
				phase: "test",
				done:  make(chan int),
			},
			doneAt: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			scriptConfigChan := IteratePhaseScripts(tt.args.phase, tt.args.done)

			i := 0
			for got := range scriptConfigChan {
				t.Log(got)
				if len(tt.want) != 0 && got != tt.want[i] {
					t.Errorf("iteratePhaseScripts() %d = %v, want %v", i, got, tt.want[i])
					tt.args.done <- 1
					break
				}

				if tt.doneAt != 0 && i == tt.doneAt {
					tt.args.done <- 1
					break
				}
				i++
			}
			if tt.doneAt != 0 && i != tt.doneAt {
				t.Errorf("iteratePhaseScripts() done at: %v, want %v", i, tt.doneAt)
			}
			close(tt.args.done)
		})
	}
}
