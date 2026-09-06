// Package config
package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTExpirationInSeconds int32
	DBUser                 string
	DBPassword             string
	DBHost                 string
	DBport                 string
	DBName                 string
	JWTSecretKey           string
	GoEnv                  string
}

// ENV variables
var ENV = initConfig()

func initConfig() *Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	return &Config{
		DBUser:                 GetEnv("DB_USER", "postgres"),
		DBPassword:             GetEnv("DB_PASSWORD", "postgres"),
		DBHost:                 GetEnv("DB_HOST", "localhost"),
		DBport:                 GetEnv("DB_PORT", "5432"),
		DBName:                 GetEnv("DB_NAME", "tui_chat_db"),
		JWTSecretKey:           GetEnv("JWT_SECRET_KEY", ""),
		JWTExpirationInSeconds: GetEnvAsInt("JWT_EXP", 259200),
		GoEnv:                  GetEnv("GO_ENV", "production"),
	}
}

func GetEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return value
}

func GetEnvAsInt(key string, fallback int32) int32 {
	if value, ok := os.LookupEnv(key); ok {
		valueInt, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return fallback
		}

		return int32(valueInt)
	}

	return fallback
}
