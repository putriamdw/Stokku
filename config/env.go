package config

import (
	"github.com/joho/godotenv"
	"os"
)

func LoadEnv() { _ = godotenv.Load() }

func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
