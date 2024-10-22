package agent

import (
	"flag"

	"evgen3000/go-musthave-metrics-tpl.git/internal/config/utils"
)

type Config struct {
	PoolInterval   int
	ReportInterval int
	Host           string
}

func GetAgentConfig() *Config {
	reportIntervalFlag := flag.Int("r", 10, "Report interval in seconds.")
	pollIntervalFlag := flag.Int("p", 2, "Pool interval in seconds.")
	hostFlag := flag.String("a", "localhost:8080", "Host IP address and port.")

	host := utils.GetStringValue("ADDRESS", *hostFlag)
	pollInterval := utils.GetIntValue("POLL_INTERVAL", *pollIntervalFlag)
	reportInterval := utils.GetIntValue("REPORT_INTERVAL", *reportIntervalFlag)

	return &Config{
		PoolInterval:   pollInterval,
		ReportInterval: reportInterval,
		Host:           host,
	}
}
