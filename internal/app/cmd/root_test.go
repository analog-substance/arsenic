package cmd

import (
	"testing"

	"github.com/analog-substance/arsenic/lib/config"
)

func setConfig(name string, phase config.Phase) {
	c := config.Config{
		Scripts: config.Scripts{
			Phases: map[string]config.Phase{
				name: phase,
			},
		},
	}
	config.Set(&c)
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
				phase := config.Phase{
					Scripts: map[string]config.Script{
						"script-1": {
							Script:  "../tests/lib/util/normal.sh",
							Enabled: true,
							Order:   1,
							Count:   1,
						},
					},
				}

				setConfig("test", phase)
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
				phase := config.Phase{
					Scripts: map[string]config.Script{
						"script-1": {
							Script:  "../tests/lib/util/exit-255.sh",
							Enabled: true,
							Order:   1,
							Count:   1,
						},
					},
				}

				setConfig("test", phase)
			},
			args: args{
				phase: "test",
			},
			statusWant: false,
			scriptWant: "../tests/lib/util/exit-255.sh",
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

func TestExecScript(t *testing.T) {
	type args struct {
		script config.Script
		args   []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Exit 0",
			args: args{
				script: config.Script{
					Script: "../tests/lib/util/normal.sh",
				},
			},
			wantErr: false,
		},
		{
			name: "Exit 255",
			args: args{
				script: config.Script{
					Script: "../tests/lib/util/exit-255.sh",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ExecScript(tt.args.script, tt.args.args)
			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Errorf("ExecScript() = %v, want %v", gotErr, tt.wantErr)
			}
		})
	}
}
