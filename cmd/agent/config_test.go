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
				Addr:           "localhost:8080",
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
		{
			name: "all_env_vars",
			args: []string{"test"},
			env: map[string]string{
				"ADDRESS":         "localhost:1111",
				"POLL_INTERVAL":   "7",
				"REPORT_INTERVAL": "30",
				"RATE_LIMIT":      "5",
				"KEY":             "secret-key",
			},
			want: agentConfig{
				Addr:           "localhost:1111",
				PollInterval:   7 * time.Second,
				ReportInterval: 30 * time.Second,
				RateLimit:      5,
				HashKey:        "secret-key",
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
			os.Unsetenv("RATE_LIMIT")
			os.Unsetenv("KEY")
			for k, v := range tt.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			cfg, err := parseConfig()
			if err != nil {
				t.Fatal(err)
			}

			if cfg.Addr != tt.want.Addr {
				t.Errorf("Addr = %q, want %q", cfg.Addr, tt.want.Addr)
			}
			if cfg.PollInterval != tt.want.PollInterval {
				t.Errorf("PollInterval = %v, want %v", cfg.PollInterval, tt.want.PollInterval)
			}
			if cfg.ReportInterval != tt.want.ReportInterval {
				t.Errorf("ReportInterval = %v, want %v", cfg.ReportInterval, tt.want.ReportInterval)
			}
			if tt.want.RateLimit != 0 && cfg.RateLimit != tt.want.RateLimit {
				t.Errorf("RateLimit = %d, want %d", cfg.RateLimit, tt.want.RateLimit)
			}
			if tt.want.HashKey != "" && cfg.HashKey != tt.want.HashKey {
				t.Errorf("HashKey = %q, want %q", cfg.HashKey, tt.want.HashKey)
			}
		})
	}
}

func TestParseConfig_InvalidEnv(t *testing.T) {
	tests := []struct {
		name   string
		envKey string
		envVal string
	}{
		{name: "invalid_report_interval", envKey: "REPORT_INTERVAL", envVal: "abc"},
		{name: "invalid_poll_interval", envKey: "POLL_INTERVAL", envVal: "xyz"},
		{name: "invalid_rate_limit", envKey: "RATE_LIMIT", envVal: "notnum"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
			os.Args = []string{"test"}

			os.Unsetenv("ADDRESS")
			os.Unsetenv("POLL_INTERVAL")
			os.Unsetenv("REPORT_INTERVAL")
			os.Unsetenv("RATE_LIMIT")
			os.Unsetenv("KEY")

			os.Setenv(tt.envKey, tt.envVal)
			defer os.Unsetenv(tt.envKey)

			_, err := parseConfig()
			if err == nil {
				t.Errorf("ожидали ошибку для %s=%q", tt.envKey, tt.envVal)
			}
		})
	}
}
