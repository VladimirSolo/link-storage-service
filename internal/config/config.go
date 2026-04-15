package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	CacheTTL    time.Duration
	CacheSweep  time.Duration
}

func Load() *Config {
	return &Config{
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/links?sslmode=disable"),
		CacheTTL:    getDuration("CACHE_TTL", 60*time.Second),
		CacheSweep:  getDuration("CACHE_SWEEP_INTERVAL", 2*time.Minute),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getDuration(key string, defaultVal time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return defaultVal
	}
	return d
}
