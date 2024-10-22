package config

import (
	"flag"
	"os"
	"strconv"
)

type ServerConfig struct {
	Host          string
	StoreInterval int
	FilePath      string
	Restore       bool
}

type HostConfig struct {
	Host  string
	IsEnv bool
}

func getStringValue(envKey string, flagValue string) string {
	if value, ok := os.LookupEnv(envKey); ok {
		return value
	}
	return flagValue
}

func getIntValue(envKey string, flagValue int) int {
	if value, ok := os.LookupEnv(envKey); ok {
		intValue, err := strconv.Atoi(value)
		if err == nil {
			return intValue
		}
	}
	return flagValue
}

func getBoolValue(envKey string, flagValue bool) bool {
	if value, ok := os.LookupEnv(envKey); ok {
		boolValue, err := strconv.ParseBool(value)
		if err == nil {
			return boolValue
		}
	}
	return flagValue
}

func GetServerConfig() *ServerConfig {
	hostFlag := flag.String("a", "localhost:8080", "Host IP address and port.")
	storeIntervalFlag := flag.Int("i", 300, "Store interval in sec.")
	filePathFlag := flag.String("f", "storage.json", "File storage location.")
	restoreFlag := flag.Bool("r", true, "Restore stored configuration.")
	flag.Parse()

	storeInterval := getIntValue("STORE_INTERVAL", *storeIntervalFlag)
	filPath := getStringValue("FILE_STORE_PATH", *filePathFlag)
	restore := getBoolValue("RESTORE", *restoreFlag)
	host := getStringValue("ADDRESS", *hostFlag)

	return &ServerConfig{
		Host:          host,
		FilePath:      filPath,
		StoreInterval: storeInterval,
		Restore:       restore,
	}
}

type AgentConfig struct {
	PoolInterval   int
	ReportInterval int
	Host           string
}

func GetAgentConfig() *AgentConfig {
	reportIntervalFlag := flag.Int("r", 10, "Report interval in seconds.")
	pollIntervalFlag := flag.Int("p", 2, "Pool interval in seconds.")
	hostFlag := flag.String("a", "localhost:8080", "Host IP address and port.")

	host := getStringValue("ADDRESS", *hostFlag)
	pollInterval := getIntValue("POOL_INTERVAL", *pollIntervalFlag)
	reportInterval := getIntValue("REPORT_INTERVAL", *reportIntervalFlag)

	return &AgentConfig{
		PoolInterval:   pollInterval,
		ReportInterval: reportInterval,
		Host:           host,
	}
}
