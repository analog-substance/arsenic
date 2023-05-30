package cmd

import (
	"testing"

	"github.com/analog-substance/arsenic/lib/config"
	"github.com/spf13/viper"
)

func setConfig() {
	var c config.Config
	viper.Unmarshal(&c)
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
				viper.Set("scripts.test", map[string]config.Script{
					"script-1": {
						Script:  "../tests/lib/util/normal.sh",
						Enabled: true,
						Order:   1,
						Count:   1,
					},
				})
				setConfig()
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
				viper.Set("scripts.test", map[string]config.Script{
					"script-1": {
						Script:  "../tests/lib/util/exit-255.sh",
						Enabled: true,
						Order:   1,
						Count:   1,
					},
				})
				setConfig()
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
		scriptPath string
		args       []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Exit 0",
			args: args{
				scriptPath: "../tests/lib/util/normal.sh",
			},
			wantErr: false,
		},
		{
			name: "Exit 255",
			args: args{
				scriptPath: "../tests/lib/util/exit-255.sh",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ExecScript(tt.args.scriptPath, tt.args.args)
			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Errorf("ExecScript() = %v, want %v", gotErr, tt.wantErr)
			}
		})
	}
}
