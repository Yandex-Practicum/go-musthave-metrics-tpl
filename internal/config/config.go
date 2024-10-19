package config

import (
	"flag"
	"log"
	"os"
	"strconv"
)

type HostConfig struct {
	Value string
	IsEnv bool
}

func GetHost() *HostConfig {
	hostFlag := flag.String("a", "localhost:8080", "Host IP address and port.")
	flag.Parse()

	env, isEnv := os.LookupEnv("ADDRESS")
	if isEnv {
		return &HostConfig{
			Value: env,
			IsEnv: isEnv,
		}
	}
	return &HostConfig{
		Value: *hostFlag,
		IsEnv: false,
	}
}

type AgentConfig struct {
	PoolInterval   int64
	ReportInterval int64
	Host           string
}

func GetAgentConfig() *AgentConfig {
	reportIntervalEnv, isReportIntervalEnv := os.LookupEnv("REPORT_INTERVAL")
	pollIntervalEnv, isPollIntervalEnv := os.LookupEnv("POLL_INTERVAL")

	reportIntervalFlag := flag.Int("r", 5, "Report interval in seconds.")
	pollIntervalFlag := flag.Int("p", 1, "Pool interval in seconds.")
	host := GetHost()

	if host.IsEnv && isPollIntervalEnv && isReportIntervalEnv {
		pollInterval, err := strconv.ParseInt(pollIntervalEnv, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		reportInterval, err := strconv.ParseInt(reportIntervalEnv, 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		return &AgentConfig{
			PoolInterval:   pollInterval,
			ReportInterval: reportInterval,
			Host:           host.Value,
		}
	}
	return &AgentConfig{
		PoolInterval:   int64(*pollIntervalFlag),
		ReportInterval: int64(*reportIntervalFlag),
		Host:           host.Value,
	}
}
