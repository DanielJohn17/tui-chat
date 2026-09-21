// Package config
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTExpirationInSeconds int32
	DBUser                 string
	DBPassword             string
	DBHost                 string
	DBport                 string
	DBName                 string
	DBSSLMode              string
	DatabaseURL            string
	JWTSecretKey           string
	GoEnv                  string
	Port                   string
}

// ENV variables
var ENV = initConfig()

func initConfig() *Config {
	_ = godotenv.Load(".env", "../../.env", "../../../.env")

	dbUser := GetEnv("DB_USER", GetEnv("PGUSER", "postgres"))
	dbPassword := GetEnv("DB_PASSWORD", GetEnv("PGPASSWORD", "postgres"))
	dbHost := GetEnv("DB_HOST", GetEnv("PGHOST", "localhost"))
	dbPort := GetEnv("DB_PORT", GetEnv("PGPORT", "5432"))
	dbName := GetEnv("DB_NAME", GetEnv("PGDATABASE", "tui_chat_db"))
	dbSSLMode := GetEnv("DB_SSLMODE", GetEnv("PGSSLMODE", "require"))

	rawDBURL := GetEnv("DATABASE_URL", GetEnv("DB_URL", ""))
	if rawDBURL == "" {
		if dbPassword != "" {
			rawDBURL = fmt.Sprintf(
				"postgres://%s:%s@%s:%s/%s?sslmode=%s",
				url.QueryEscape(dbUser),
				url.QueryEscape(dbPassword),
				dbHost,
				dbPort,
				dbName,
				dbSSLMode,
			)
		} else {
			rawDBURL = fmt.Sprintf(
				"postgres://%s@%s:%s/%s?sslmode=%s",
				url.QueryEscape(dbUser),
				dbHost,
				dbPort,
				dbName,
				dbSSLMode,
			)
		}
	}

	return &Config{
		DBUser:                 dbUser,
		DBPassword:             dbPassword,
		DBHost:                 dbHost,
		DBport:                 dbPort,
		DBName:                 dbName,
		DBSSLMode:              dbSSLMode,
		DatabaseURL:            rawDBURL,
		JWTSecretKey:           GetEnv("JWT_SECRET_KEY", "your-default-secret-key-change-in-production"),
		JWTExpirationInSeconds: GetEnvAsInt("JWT_EXP", 259200),
		GoEnv:                  GetEnv("GO_ENV", "production"),
		Port:                   GetEnv("PORT", "8080"),
	}
}

func (c *Config) GetDBURL() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		url.QueryEscape(c.DBUser),
		url.QueryEscape(c.DBPassword),
		c.DBHost,
		c.DBport,
		c.DBName,
		c.DBSSLMode,
	)
}

func (c *Config) GetServerAddr() string {
	port := c.Port
	if port == "" {
		port = "8080"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	return port
}

func GetEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}

	return value
}

func GetEnvAsInt(key string, fallback int32) int32 {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		valueInt, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return fallback
		}

		return int32(valueInt)
	}

	return fallback
}
