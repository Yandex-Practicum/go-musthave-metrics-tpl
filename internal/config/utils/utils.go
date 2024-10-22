package utils

import (
	"os"
	"strconv"
)

func GetStringValue(envKey string, flagValue string) string {
	if value, ok := os.LookupEnv(envKey); ok {
		return value
	}
	return flagValue
}

func GetIntValue(envKey string, flagValue int) int {
	if value, ok := os.LookupEnv(envKey); ok {
		intValue, err := strconv.Atoi(value)
		if err == nil {
			return intValue
		}
	}
	return flagValue
}

func GetBoolValue(envKey string, flagValue bool) bool {
	if value, ok := os.LookupEnv(envKey); ok {
		boolValue, err := strconv.ParseBool(value)
		if err == nil {
			return boolValue
		}
	}
	return flagValue
}
