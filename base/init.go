package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	APIKey string
}

func LoadConfig() *Config {
	_ = godotenv.Load() // Optional: Don’t crash if .env is missing

	return &Config{
		APIKey: os.Getenv("API_KEY"),
	}
}
