package config

import (
	"reflect"
	"testing"
)

func TestConfig_GetScripts(t *testing.T) {
	type fields struct {
		Scripts Scripts
	}
	type args struct {
		phase string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []Script
	}{
		{
			name: "Sorted by order",
			fields: fields{
				Scripts: Scripts{
					Phases: map[string]Phase{
						"test": {
							Scripts: map[string]Script{
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
							},
						},
					},
				},
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
			c := &Config{
				Scripts: tt.fields.Scripts,
			}

			got := c.GetScripts(tt.args.phase)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Config.GetScripts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_IterateScripts(t *testing.T) {
	type fields struct {
		Scripts Scripts
	}
	type args struct {
		phase string
		done  chan int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		doneAt int
		want   []Script
	}{
		{
			name: "Normal iteration",
			fields: fields{
				Scripts: Scripts{
					Phases: map[string]Phase{
						"test": {
							Scripts: map[string]Script{
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
							},
						},
					},
				},
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
			fields: fields{
				Scripts: Scripts{
					Phases: map[string]Phase{
						"test": {
							Scripts: map[string]Script{
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
							},
						},
					},
				},
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
			fields: fields{
				Scripts: Scripts{
					Phases: map[string]Phase{
						"test": {
							Scripts: map[string]Script{
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
							},
						},
					},
				},
			},
			args: args{
				phase: "test",
				done:  make(chan int),
			},
			doneAt: 2,
		},
		{
			name: "Done at end",
			fields: fields{
				Scripts: Scripts{
					Phases: map[string]Phase{
						"test": {
							Scripts: map[string]Script{
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
							},
						},
					},
				},
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
			c := &Config{
				Scripts: tt.fields.Scripts,
			}
			scripts := c.IterateScripts(tt.args.phase, tt.args.done)

			i := 0
			for got := range scripts {
				t.Log(got)
				if len(tt.want) != 0 && got != tt.want[i] {
					t.Errorf("iterateScripts() %d = %v, want %v", i, got, tt.want[i])
					tt.args.done <- 1
					break
				}

				if tt.doneAt != 0 && i == tt.doneAt {
					tt.args.done <- 1
					break
				}
				i++
			}
		})
	}
}
