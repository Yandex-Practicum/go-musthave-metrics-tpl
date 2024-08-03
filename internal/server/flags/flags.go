package flags

import (
	"log"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config структура конфигурации
type Config struct {
	ServerAddress   string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
}

// GetFlags устанавливает и получает флаги
func GetFlags() {
	// Define the flags and bind them to viper
	pflag.StringP("ServerAddress", "a", "localhost:8080", "HTTP server network address")
	pflag.IntP("StoreInterval", "i", 300, "Interval in seconds to store the current server readings to disk")
	pflag.StringP("FileStoragePath", "f", "/tmp/metrics-db.json", "Full filename where current values are saved")
	pflag.BoolP("Restore", "r", true, "Whether to load previously saved values from the specified file at server startup")

	// Parse the command-line flags
	pflag.Parse()

	// Check for unknown flags
	for _, arg := range pflag.Args() {
		if !strings.HasPrefix(arg, "-") {
			log.Fatalf("Unknown flag: %v", arg)
		}
	}

	// Bind the flags to viper
	bindFlagToViper("ServerAddress")
	bindFlagToViper("StoreInterval")
	bindFlagToViper("FileStoragePath")
	bindFlagToViper("Restore")

	// Set the environment variable names
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	bindEnvToViper("ServerAddress", "ADDRESS")
	bindEnvToViper("StoreInterval", "STORE_INTERVAL")
	bindEnvToViper("FileStoragePath", "FILE_STORAGE_PATH")
	bindEnvToViper("Restore", "RESTORE")

	// Read the environment variables
	viper.AutomaticEnv()
}

func bindFlagToViper(flagName string) {
	if err := viper.BindPFlag(flagName, pflag.Lookup(flagName)); err != nil {
		log.Println(err)
	}
}

func bindEnvToViper(viperKey, envKey string) {
	if err := viper.BindEnv(viperKey, envKey); err != nil {
		log.Println(err)
	}
}

// NewConfig создает новый экземпляр конфигурации
func NewConfig() *Config {
	GetFlags()
	return &Config{
		ServerAddress:   Address(),
		StoreInterval:   Interval(),
		FileStoragePath: FileStoragePath(),
		Restore:         Restore(),
	}
}

// Address возвращает адрес сервера
func Address() string {
	return viper.GetString("ServerAddress")
}

// Interval возвращает интервал сохранения текущих значений сервера на диск
func Interval() int {
	return viper.GetInt("StoreInterval")
}

// FileStoragePath возвращает путь к файлу хранения
func FileStoragePath() string {
	path := viper.GetString("FileStoragePath")
	if path == "=" {
		return ""
	}
	return path
}

// Restore возвращает флаг восстановления
func Restore() bool {
	return viper.GetBool("Restore")
}
