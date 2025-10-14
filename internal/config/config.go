package config

import (
	"os"
	// "strconv"
)

type Config struct {
	ServerPort string
	DBURL      string
	RedisURL   string
	JWTSecret  string
}

func Load() *Config {
	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		DBURL:      getEnv("DB_URL", "postgres://user:pass@localhost:5432/auth_db"),
		RedisURL:   getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:  getEnv("JWT_SECRET", "default-secret-key"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}