package config

import (
	"flag"
	"log"
	"os"
	"strconv"

	"evgen3000/go-musthave-metrics-tpl.git/internal/logger"
)

type ServerConfig struct {
	Host          string
	StoreInterval int
	FilePath      string
	Restore       bool
	IsEnv         bool
}

type HostConfig struct {
	Host  string
	IsEnv bool
}

func GetHost() *HostConfig {
	hostFlag := flag.String("a", "localhost:8080", "Host IP address and port.")
	flag.Parse()

	env, isEnv := os.LookupEnv("ADDRESS")
	if isEnv {
		return &HostConfig{
			Host:  env,
			IsEnv: isEnv,
		}
	}
	return &HostConfig{
		Host:  *hostFlag,
		IsEnv: false,
	}
}

func GetServerConfig() *ServerConfig {
	soreIntervalFlag := flag.Int("i", 300, "Store interval in sec.")
	filePathFlag := flag.String("f", "./", "File storage location.")
	restoreFlag := flag.Bool("r", true, "Restore stored configuration.")
	host := GetHost()

	flag.Parse()

	storeIntervalEnv, isStoreIntervalEnv := os.LookupEnv("STORE_INTERVAL")
	filePathEnv, isFilePathEnv := os.LookupEnv("FILE_STORAGE_PATH")
	restoreEnv, isRestoreEnv := os.LookupEnv("RESTORE")

	if host.IsEnv && isStoreIntervalEnv && isFilePathEnv && isRestoreEnv {
		soreInterval, err := strconv.Atoi(storeIntervalEnv)
		if err != nil {
			logger.GetLogger().Fatal("Can't implement STORE_INTERVAL for code")
		}

		restore, err := strconv.ParseBool(restoreEnv)
		if err != nil {
			logger.GetLogger().Fatal("Can't implement RESTORE for code")
		}
		return &ServerConfig{
			Host:          host.Host,
			FilePath:      filePathEnv,
			StoreInterval: soreInterval,
			Restore:       restore,
			IsEnv:         true,
		}
	}
	return &ServerConfig{
		Host:          host.Host,
		FilePath:      *filePathFlag,
		StoreInterval: *soreIntervalFlag,
		Restore:       *restoreFlag,
		IsEnv:         false,
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

	reportIntervalFlag := flag.Int("r", 10, "Report interval in seconds.")
	pollIntervalFlag := flag.Int("p", 2, "Pool interval in seconds.")
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
			Host:           host.Host,
		}
	}
	return &AgentConfig{
		PoolInterval:   int64(*pollIntervalFlag),
		ReportInterval: int64(*reportIntervalFlag),
		Host:           host.Host,
	}
}
