package main

import (
	"flag"
	"os"
	"testing"
	"time"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name string
		args []string
		env  map[string]string
		want agentConfig
	}{
		{
			name: "defaults",
			args: []string{"test"},
			env:  map[string]string{},
			want: agentConfig{
				Addr:           "localhost:5050",
				PollInterval:   2 * time.Second,
				ReportInterval: 10 * time.Second,
			},
		},
		{
			name: "env_override",
			args: []string{"test"},
			env: map[string]string{
				"ADDRESS":         "localhost:9090",
				"POLL_INTERVAL":   "5",
				"REPORT_INTERVAL": "20",
			},
			want: agentConfig{
				Addr:           "localhost:9090",
				PollInterval:   5 * time.Second,
				ReportInterval: 20 * time.Second,
			},
		},
		{
			name: "env_priority_over_flags",
			args: []string{"test", "-a", ":7070"},
			env: map[string]string{
				"ADDRESS": ":8888",
			},
			want: agentConfig{
				Addr:           ":8888",
				PollInterval:   2 * time.Second,
				ReportInterval: 10 * time.Second,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
			os.Args = tt.args

			os.Unsetenv("ADDRESS")
			os.Unsetenv("POLL_INTERVAL")
			os.Unsetenv("REPORT_INTERVAL")
			for k, v := range tt.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			cfg := parseConfig()

			if cfg.Addr != tt.want.Addr {
				t.Errorf("Addr = %q, want %q", cfg.Addr, tt.want.Addr)
			}
			if cfg.PollInterval != tt.want.PollInterval {
				t.Errorf("PollInterval = %v, want %v", cfg.PollInterval, tt.want.PollInterval)
			}
			if cfg.ReportInterval != tt.want.ReportInterval {
				t.Errorf("ReportInterval = %v, want %v", cfg.ReportInterval, tt.want.ReportInterval)
			}
		})
	}
}
