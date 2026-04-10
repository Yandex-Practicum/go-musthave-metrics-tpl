package main

import (
	"database/sql"
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/logger"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/server"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
	"github.com/rs/zerolog/log"
)

func main() {
	addr := flag.String("a", ":8080", "адрес сервера (host:port)")
	logLevel := flag.String("l", "info", "уровень логирования")
	storeInterval := flag.Int("i", 300, "интервал сохранения на диск (сек)")
	filePath := flag.String("f", "metrics.json", "путь до файла хранения")
	restore := flag.Bool("r", true, "загружать данныепри старте")
	databaseDSN := flag.String("d", "", "строка подклюучения к PostgreSQL")
	flag.Parse()

	if v := os.Getenv("ADDRESS"); v != "" {
		*addr = v
	}

	if v := os.Getenv("STORE_INTERVAL"); v != "" {
		sec, err := strconv.Atoi(v)
		if err != nil {
			log.Fatal().Err(err).Msg("Неверное значение STORE_INTERVAL:")
		}
		*storeInterval = sec
	}

	if v := os.Getenv("FILE_STORAGE_PATH"); v != "" {
		*filePath = v
	}

	if v := os.Getenv("RESTORE"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			log.Fatal().Err(err).Msg("Неверное значение RESTORE")
		}
		*restore = b
	}

	if v := os.Getenv("DATABASE_DSN"); v != "" {
		*databaseDSN = v
	}

	if err := logger.Initialize(*logLevel); err != nil {
		log.Fatal().Err(err).Msg("ошибка инициализации логгера")
	}

	var repo storage.Repository
	var db *sql.DB

	if *databaseDSN != "" {
		var err error
		db, err = storage.NewPostgresDB(*databaseDSN)
		if err != nil {
			log.Fatal().Err(err).Msg("ошибка при подключения к БД")
		}
		repo = storage.NewPostgresStorege(db)
	} else {
		memRepo := storage.NewMemoryStorage()

		// загрузка метрик из файла при старте
		if *restore && *filePath != "" {
			if err := storage.LoadFromFile(memRepo, *filePath); err != nil {
				log.Error().Err(err).Msg("Не удалось загрузить метрик")
			}
		}
		// периодическое сохранение на диск
		if *filePath != "" && *storeInterval > 0 {
			go storage.RunSaver(memRepo, *filePath, time.Duration(*storeInterval)*time.Second)
		}
		//синхронная запись — передаём filePath в сервер
		if *filePath != "" && *storeInterval == 0 {
			memRepo.SetSyncFile(*filePath)
		}
		repo = memRepo
	}

	srv := server.New(*addr, repo, db)
	if err := srv.Run(); err != nil {
		log.Fatal().Err(err).Msg("ошибка запуска сервера")
	}

}
