package server

import (
	"flag"

	"evgen3000/go-musthave-metrics-tpl.git/internal/config/utils"
)

type Config struct {
	Host          string
	StoreInterval int
	FilePath      string
	Restore       bool
}

func GetServerConfig() *Config {
	hostFlag := flag.String("a", "localhost:8080", "Host IP address and port.")
	storeIntervalFlag := flag.Int("i", 300, "Store interval in sec.")
	filePathFlag := flag.String("f", "storage.json", "File storage location.")
	restoreFlag := flag.Bool("r", true, "Restore stored configuration.")
	flag.Parse()

	storeInterval := utils.GetIntValue("STORE_INTERVAL", *storeIntervalFlag)
	filPath := utils.GetStringValue("FILE_STORE_PATH", *filePathFlag)
	restore := utils.GetBoolValue("RESTORE", *restoreFlag)
	host := utils.GetStringValue("ADDRESS", *hostFlag)

	return &Config{
		Host:          host,
		FilePath:      filPath,
		StoreInterval: storeInterval,
		Restore:       restore,
	}
}
