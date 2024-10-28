package storage

import (
	"log"
	"time"

	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage/file_manager"
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage/mem_storage"
)

type MemStorageConfig struct {
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
}

func NewMemStorage(config MemStorageConfig, fm *file_manager.FileManager) *mem_storage.MemStorage {
	var storage = &mem_storage.MemStorage{
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
