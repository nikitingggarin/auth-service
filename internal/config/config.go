package config

import (
	"os"
	"time"
)

type Config struct {
	ServerPort    string
	DBURL         string
	JWTSecret     string
	JWTExpiration time.Duration
}

func Load() *Config {
	return &Config{
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		DBURL:         getEnv("DB_URL", "postgres://postgres:password@localhost:5432/auth_service"),
		JWTSecret:     getEnv("JWT_SECRET", "default-secret-key-change-in-production"),
		JWTExpiration: 24 * time.Hour, // 24 часа
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
