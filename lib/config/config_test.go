package config

import (
	"testing"
)

func TestIgnoreService_checkPort(t *testing.T) {
	type fields struct {
		Name  string
		Ports string
		Flag  string
	}
	type args struct {
		port int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Invalid min range",
			fields: fields{
				Ports: "asdf-100",
			},
			args: args{
				port: 1,
			},
			want: false,
		},
		{
			name: "Invalid max range",
			fields: fields{
				Ports: "1-asdf",
			},
			args: args{
				port: 1,
			},
			want: false,
		},
		{
			name: "Invalid port",
			fields: fields{
				Ports: "asdf",
			},
			args: args{
				port: 1,
			},
			want: false,
		},
		{
			name: "Valid port and port match",
			fields: fields{
				Ports: "100",
			},
			args: args{
				port: 100,
			},
			want: true,
		},
		{
			name: "Valid port and port doesn't match",
			fields: fields{
				Ports: "100",
			},
			args: args{
				port: 101,
			},
			want: false,
		},
		{
			name: "Valid range and port within range",
			fields: fields{
				Ports: "1-1000",
			},
			args: args{
				port: 500,
			},
			want: true,
		},
		{
			name: "Valid range and port not within range",
			fields: fields{
				Ports: "1-1000",
			},
			args: args{
				port: 1001,
			},
			want: false,
		},
		{
			name: "Multiple ports",
			fields: fields{
				Ports: "1,2,3-50,51-61,700",
			},
			args: args{
				port: 52,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := IgnoreService{
				Name:  tt.fields.Name,
				Ports: tt.fields.Ports,
				Flag:  tt.fields.Flag,
			}
			if got := s.checkPort(tt.args.port); got != tt.want {
				t.Errorf("IgnoreService.checkPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIgnoreService_ShouldIgnore(t *testing.T) {
	type fields struct {
		Name  string
		Ports string
		Flag  string
	}
	type args struct {
		service string
		port    int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Service doesn't match",
			fields: fields{
				Name: "test-service",
			},
			args: args{
				service: "not-matching",
				port:    0,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := IgnoreService{
				Name:  tt.fields.Name,
				Ports: tt.fields.Ports,
				Flag:  tt.fields.Flag,
			}
			if got := s.ShouldIgnore(tt.args.service, tt.args.port); got != tt.want {
				t.Errorf("IgnoreService.ShouldIgnore() = %v, want %v", got, tt.want)
			}
		})
	}
}
