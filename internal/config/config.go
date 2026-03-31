package config

import (
	"os"
	"time"
)

type Config struct {
	Port           string
	DBURL          string
	RedisURL       string
	JWTSecret      string
	JWTExpiry      time.Duration
	RefreshExpiry  time.Duration
}

func LoadConfig() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		DBURL:         getEnv("DB_URL", "postgres://postgres:postgres@localhost:5432/chatly?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:     getEnv("JWT_SECRET", "very-secret-key"),
		JWTExpiry:     getDurationEnv("JWT_EXPIRY", "15m"),
		RefreshExpiry: getDurationEnv("REFRESH_EXPIRY", "168h"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getDurationEnv(key, fallback string) time.Duration {
	val := getEnv(key, fallback)
	d, err := time.ParseDuration(val)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}
