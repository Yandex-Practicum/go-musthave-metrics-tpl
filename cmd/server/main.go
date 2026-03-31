package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/logger"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/server"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
)

func main() {
	addr := flag.String("a", ":8080", "адрес сервера (host:port)")
	logLevel := flag.String("l", "info", "уровень логирования")
	storeInterval := flag.Int("i", 300, "интервал сохранения на диск (сек)")
	filePath := flag.String("f", "metrics.json", "путь до файла хранения")
	restore := flag.Bool("r", true, "загружать данныепри старте")
	flag.Parse()

	if v := os.Getenv("ADDRESS"); v != "" {
		*addr = v
	}

	if v := os.Getenv("STORE_INTERVAL"); v != "" {
		sec, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("Неверное значение STORE_INTERVAL: %v", err)
		}
		*storeInterval = sec
	}

	if v := os.Getenv("FILE_STORAGE_PATH"); v != "" {
		*filePath = v
	}

	if v := os.Getenv("RESTORE"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			log.Fatalf("Неверное значение RESTORE: %v", err)
		}
		*restore = b
	}

	if err := logger.Initialize(*logLevel); err != nil {
		log.Fatal(err)
	}

	repo := storage.NewMemoryStorage()

	// загрузка метрик из файла при старте
	if *restore && *filePath != "" {
		if err := storage.LoadFromFile(repo, *filePath); err != nil {
			log.Printf("Не удалось загрузить метрик: %v", err)
		}
	}
	// периодическое сохранение на диск
	if *filePath != "" && *storeInterval > 0 {
		go storage.RunSaver(repo, *filePath, time.Duration(*storeInterval)*time.Second)
	}
	//синхронная запись — передаём filePath в сервер
	if *filePath != "" && *storeInterval == 0 {
		repo.SetSyncFile(*filePath)
	}

	srv := server.New(*addr, repo)
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
