// Package config
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBport     string
	DBName     string
}

// ENV variables
var ENV = initConfig()

func initConfig() *Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	return &Config{
		DBUser:     GetEnv("DB_USER", "postgres"),
		DBPassword: GetEnv("DB_PASSWORD", "postgres"),
		DBHost:     GetEnv("DB_HOST", "localhost"),
		DBport:     GetEnv("DB_PORT", "5432"),
		DBName:     GetEnv("DB_NAME", "tui_chat_db"),
	}
}

func GetEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return value
}
