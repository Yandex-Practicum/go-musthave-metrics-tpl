package config

import (
	"flag"
	"os"
)

type HostConfig struct {
	Value string
	IsEnv bool
}

func GetHost() HostConfig {
	hostFlag := flag.String("a", "localhost:8080", "Host IP address and port.")
	flag.Parse()

	env, isEnv := os.LookupEnv("ADDRESS")
	if isEnv {
		return HostConfig{
			Value: env,
			IsEnv: isEnv,
		}
	}
	return HostConfig{
		Value: *hostFlag,
		IsEnv: false,
	}
}
