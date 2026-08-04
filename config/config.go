package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfigTmp
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	User     string
	Password string
	Name     string
	Host     string
	Port     string
	Driver   string
}

type AuthConfigTmp struct {
	JWTSecret       string
	JWTIssuer       string
	JWTAudience     string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	CookieDomain    string
	CookieSecure    bool
}

func _getEnvBool(key string, defaultVal bool) bool {
	CookieSecurestr := getEnv(key, "")
	if CookieSecurestr == "" {
		return defaultVal
	}
	CookieSecure, err := strconv.ParseBool(CookieSecurestr)
	if err != nil {
		return defaultVal
	}
	return CookieSecure
}

func Load(logger *slog.Logger) (*Config, error) {
	if err := godotenv.Load(); err != nil {
		logger.Info("No .env file found, using environment variables")
	}
	accessTokenTTL, err := strconv.Atoi(getEnv("ACCESSTTL", "15"))
	if err != nil {
		return nil, fmt.Errorf("invalid ACCESSTTL value: %w", err)
	}
	refreshTokenTTL, err := strconv.Atoi(getEnv("REFRESHTTL", "1440"))
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESHTTL value: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", ""),
			Host:     getEnv("DB_HOST", ""),
			Port:     getEnv("DB_PORT", "5432"),
			Driver:   getEnv("DB_DRIVER", "postgres"),
		},
		Auth: AuthConfigTmp{
			JWTSecret:       getEnv("JWTSECRET", ""),
			JWTIssuer:       getEnv("JWTISSUER", ""),
			JWTAudience:     getEnv("JWTAUDIENCE", ""),
			AccessTokenTTL:  time.Duration(accessTokenTTL) * time.Minute,
			RefreshTokenTTL: time.Duration(refreshTokenTTL) * time.Minute,
			CookieDomain:    getEnv("COOKIE_DOMAIN", ""),
			CookieSecure:    _getEnvBool("COOKIE_SECURE", true),
		},
	}
	if cfg.Database.User == "" {
		return nil, fmt.Errorf("DB_USER is required")
	} else if cfg.Database.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	} else if cfg.Database.Name == "" {
		return nil, fmt.Errorf("DB_NAME is required")
	} else if cfg.Database.Host == "" {
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
