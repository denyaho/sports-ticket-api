package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerConfig
}

type ServerConfig struct {
	Port string
	Host string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
