package config

import "os"

type Config struct {
	Port    string
	DBPath  string
	GinMode string
}

func Load() Config {
	return Config{
		Port:    getEnv("PORT", "18007"),
		DBPath:  getEnv("DB_PATH", "./data/bookmarks.db"),
		GinMode: getEnv("GIN_MODE", "release"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
