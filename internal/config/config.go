package config

import (
	"os"
)

type Config struct {
	TarantoolHost string
	TarantoolUser string
	TarantoolPass string
	HTTPPort      string
}

func LoadConfig() *Config {
	cfg := &Config{
		TarantoolHost: getEnv("TARANTOOL_HOST", "tarantool:3301"),
		TarantoolUser: getEnv("TARANTOOL_USER", ""),
		TarantoolPass: getEnv("TARANTOOL_PASS", ""),
		HTTPPort:      getEnv("HTTP_PORT", "8080"),
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
