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
	Env      string
	Reservation ReservationConfig
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

type ReservationConfig struct {
	ReservationExpiration time.Duration
	MaxSeats              int
}

func _getEnvBool(key string, defaultVal bool) bool {
	cookieSecureStr := getEnv(key, "")
	if cookieSecureStr == "" {
		return defaultVal
	}
	cookieSecure, err := strconv.ParseBool(cookieSecureStr)
	if err != nil {
		return defaultVal
	}
	return cookieSecure
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
		Env: getEnv("ENV", "development"),
		Reservation: ReservationConfig{
			ReservationExpiration: time.Duration(getEnvInt("HOLDTIME", 15)) * time.Minute,
			MaxSeats:              getEnvInt("MAXSEATS", 3),
		},
	}
	switch {
	case cfg.Database.User == "":
		return nil, fmt.Errorf("DB_USER is required")
	case cfg.Database.Password == "":
		return nil, fmt.Errorf("DB_PASSWORD is required")
	case cfg.Database.Name == "":
		return nil, fmt.Errorf("DB_NAME is required")
	case cfg.Database.Host == "":
		return nil, fmt.Errorf("DB_HOST is required")
	case !cfg.Auth.CookieSecure && cfg.Env == "production":
		return nil, fmt.Errorf("COOKIE_SECURE is false")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
