package storage

import (
	"log"
	"time"

	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage/filemanager"
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage/memstorage"
)

type MemStorageConfig struct {
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
}

func NewMemStorage(config MemStorageConfig, fm *filemanager.FileManager) *memstorage.MemStorage {
	var storage = &memstorage.MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}

	if config.Restore {
		err := fm.LoadData(config.FileStoragePath, storage)
		if err != nil {
			log.Fatal(err)
		}
	}
	return storage
}
