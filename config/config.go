package config

import (
	"os"
	"fmt"
	"github.com/joho/godotenv"
	"log"
)

type Config struct {
	Server ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	User string
	Password string
	Name string
	Host string
	Port string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found,")
	}
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			User: getEnv("DB_USER",""),
			Password: getEnv("DB_PASSWORD",""),
			Name: getEnv("DB_NAME",""),
			Host: getEnv("DB_HOST",""),
			Port: getEnv("DB_PORT", "5432"),
		},
	}
	if cfg.Database.User == ""{
		return nil, fmt.Errorf("DB_USER is required")
	} else if cfg.Database.Password == ""{
		return nil, fmt.Errorf("DB_PASSWORD is required")
	} else if cfg.Database.Name == ""{
		return nil, fmt.Errorf("DB_NAME is required")
	} else if cfg.Database.Host == ""{
		return nil, fmt.Errorf("DB_HOST is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
