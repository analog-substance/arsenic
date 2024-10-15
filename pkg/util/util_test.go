package util

import (
	"testing"
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
