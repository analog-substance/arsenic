package util

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestIsIp(t *testing.T) {
	type args struct {
		ipOrHostname string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Should return true for valid IP",
			args{"11.11.11.11"},
			true,
		},
		{
			"Should return false for invalid IP",
			args{"999.11.320.11"},
			false,
		},
		{
			"Should return false domain",
			args{"example.com"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsIp(tt.args.ipOrHostname); got != tt.want {
				t.Errorf("IsIp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetScripts(t *testing.T) {
	type args struct {
		phase string
	}
	tests := []struct {
		name  string
		setup func()
		args  args
		want  []ScriptConfig
	}{
		{
			name: "Sorted by order",
			setup: func() {
				viper.Set("scripts.test", map[string]ScriptConfig{
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
			},
			args: args{phase: "test"},
			want: []ScriptConfig{
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

func TestExecScript(t *testing.T) {
	type args struct {
		scriptPath string
		args       []string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Exit 0",
			args: args{
				scriptPath: "../../tests/lib/util/normal.sh",
			},
			want: 0,
		},
		{
			name: "Exit 255",
			args: args{
				scriptPath: "../../tests/lib/util/exit-255.sh",
			},
			want: 255,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExecScript(tt.args.scriptPath, tt.args.args); got != tt.want {
				t.Errorf("ExecScript() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_iteratePhaseScripts(t *testing.T) {
	type args struct {
		phase string
		done  chan int
	}
	tests := []struct {
		name   string
		setup  func()
		args   args
		doneAt int
		want   []ScriptConfig
	}{
		{
			name: "Normal iteration",
			setup: func() {
				viper.Set("scripts.test", map[string]ScriptConfig{
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
			},
			args: args{
				phase: "test",
				done:  make(chan int),
			},
			want: []ScriptConfig{
				{
					Script:    "script-1-name",
					Enabled:   true,
					Order:     1,
					Count:     3,
					totalRuns: 1,
				},
				{
					Script:    "script-1-name",
					Enabled:   true,
					Order:     1,
					Count:     3,
					totalRuns: 2,
				},
				{
					Script:    "script-1-name",
					Enabled:   true,
					Order:     1,
					Count:     3,
					totalRuns: 3,
				},
				{
					Script:    "script-2-name",
					Enabled:   true,
					Order:     5,
					Count:     1,
					totalRuns: 1,
				},
			},
		},
		{
			name: "Some disabled or 0 count",
			setup: func() {
				viper.Set("scripts.test", map[string]ScriptConfig{
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
			},
			args: args{
				phase: "test",
				done:  make(chan int),
			},
			want: []ScriptConfig{
				{
					Script:    "script-1-name",
					Enabled:   true,
					Order:     1,
					Count:     3,
					totalRuns: 1,
				},
				{
					Script:    "script-1-name",
					Enabled:   true,
					Order:     1,
					Count:     3,
					totalRuns: 2,
				},
				{
					Script:    "script-1-name",
					Enabled:   true,
					Order:     1,
					Count:     3,
					totalRuns: 3,
				},
				{
					Script:    "script-4-name",
					Enabled:   true,
					Order:     4,
					Count:     1,
					totalRuns: 1,
				},
			},
		},
		{
			name: "Done in middle",
			setup: func() {
				viper.Set("scripts.test", map[string]ScriptConfig{
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
				viper.Set("scripts.test", map[string]ScriptConfig{
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
			scriptConfigChan := iteratePhaseScripts(tt.args.phase, tt.args.done)

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

func Test_executePhaseScripts(t *testing.T) {
	type args struct {
		phase  string
		args   []string
		dryRun bool
	}
	tests := []struct {
		name       string
		setup      func()
		args       args
		statusWant bool
		scriptWant string
	}{
		{
			name: "Normal",
			setup: func() {
				viper.Set("scripts.test", map[string]ScriptConfig{
					"script-1": {
						Script:  "../../tests/lib/util/normal.sh",
						Enabled: true,
						Order:   1,
						Count:   1,
					},
				})
			},
			args: args{
				phase: "test",
			},
			statusWant: true,
			scriptWant: "",
		},
		{
			name: "Exited script",
			setup: func() {
				viper.Set("scripts.test", map[string]ScriptConfig{
					"script-1": {
						Script:  "../../tests/lib/util/exit-255.sh",
						Enabled: true,
						Order:   1,
						Count:   1,
					},
				})
			},
			args: args{
				phase: "test",
			},
			statusWant: false,
			scriptWant: "../../tests/lib/util/exit-255.sh",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			statusGot, scriptGot := executePhaseScripts(tt.args.phase, tt.args.args, tt.args.dryRun)
			if statusGot != tt.statusWant {
				t.Errorf("executePhaseScripts() statusGot = %v, want %v", statusGot, tt.statusWant)
			}
			if scriptGot != tt.scriptWant {
				t.Errorf("executePhaseScripts() scriptGot = %v, want %v", scriptGot, tt.scriptWant)
			}
		})
	}
}

func TestReadLineByLine(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Multi line",
			args: args{
				r: strings.NewReader("line 1\nline 2\nline 3"),
			},
			want: []string{
				"line 1",
				"line 2",
				"line 3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotChan := ReadLineByLine(tt.args.r)

			var got []string
			for g := range gotChan {
				got = append(got, g)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadLineByLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadFileLineByLine(t *testing.T) {
	tempDir := t.TempDir()

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		setup   func()
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Multi line file",
			setup: func() {
				os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("line 1\nline 2\nline 3"), 0644)
			},
			args: args{
				path: filepath.Join(tempDir, "test.txt"),
			},
			want: []string{
				"line 1",
				"line 2",
				"line 3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			gotChan, err := ReadFileLineByLine(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFileLineByLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var got []string
			for g := range gotChan {
				got = append(got, g)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadFileLineByLine() = %v, want %v", got, tt.want)
			}
		})
	}
}
