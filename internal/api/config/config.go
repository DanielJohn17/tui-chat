// Package config
package config

import (
	"errors"
	"flag"
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
	AppSharedSecret        string
	GoEnv                  string
	Port                   string
}

// ENV variables
var ENV = initConfig()

func initConfig() *Config {
	for _, envPath := range []string{".env", "../../.env", "../../../.env", "../../../../.env"} {
		if err := godotenv.Load(envPath); err == nil {
			break
		}
	}

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

	goEnv := GetEnv("GO_ENV", "production")
	jwtSecret := GetEnv("JWT_SECRET_KEY", "")
	appSecret := GetEnv("APP_SHARED_SECRET", GetEnv("CLIENT_SHARED_SECRET", ""))

	// In test environments, if no secret is provided, provide a test fallback so tests pass cleanly
	if isTestEnv() {
		if jwtSecret == "" {
			jwtSecret = DevDefaultSecretKey
		}
		if appSecret == "" {
			appSecret = "test-automated-suite-shared-secret-32-chars"
		}
	}

	if err := ValidateJWTSecret(jwtSecret, goEnv); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL CONFIG ERROR: %v\n", err)
		os.Exit(1)
	}
	if err := ValidateAppSecret(appSecret, goEnv); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL CONFIG ERROR: %v\n", err)
		os.Exit(1)
	}

	if jwtSecret == "" {
		jwtSecret = DevDefaultSecretKey
	}
	if appSecret == "" {
		appSecret = DevDefaultAppSecret
	}

	return &Config{
		DBUser:                 dbUser,
		DBPassword:             dbPassword,
		DBHost:                 dbHost,
		DBport:                 dbPort,
		DBName:                 dbName,
		DBSSLMode:              dbSSLMode,
		DatabaseURL:            rawDBURL,
		JWTSecretKey:           jwtSecret,
		AppSharedSecret:        appSecret,
		JWTExpirationInSeconds: GetEnvAsInt("JWT_EXP", 259200),
		GoEnv:                  goEnv,
		Port:                   GetEnv("PORT", "8080"),
	}
}

const (
	DefaultInsecureSecretKey = "your-default-secret-key-change-in-production"
	DevDefaultSecretKey      = "dev-insecure-jwt-secret-key-min-32-characters-long"
	MinJWTSecretLength       = 32

	DefaultInsecureAppSecret = "your-default-app-secret-change-in-production"
	DevDefaultAppSecret      = "dev-insecure-client-shared-secret-min-32-chars-long"
	MinAppSecretLength       = 32
	ClientSecretHeader       = "X-App-Secret"
	AltClientSecretHeader    = "X-Client-Secret"
)

// ValidateJWTSecret validates that the JWT secret is sufficiently secure.
// In production, an explicit, non-default secret of at least 32 characters is strictly required.
func ValidateJWTSecret(secret, goEnv string) error {
	if goEnv == "production" {
		if secret == "" || secret == DefaultInsecureSecretKey {
			return errors.New("JWT_SECRET_KEY environment variable is required in production and cannot use default fallback")
		}
		if len(secret) < MinJWTSecretLength {
			return fmt.Errorf("JWT_SECRET_KEY in production must be at least %d characters long for security", MinJWTSecretLength)
		}
	}
	return nil
}

// ValidateAppSecret verifies that the client shared secret meets security requirements.
// Requires the secret to be non-empty in all environments.
// In production, placeholder/dev secrets are forbidden and length must be >= 32 characters.
func ValidateAppSecret(secret, goEnv string) error {
	if secret == "" {
		return errors.New("APP_SHARED_SECRET environment variable is required and must be defined in your .env file")
	}
	if goEnv == "production" {
		if secret == DefaultInsecureAppSecret || secret == DevDefaultAppSecret {
			return errors.New("APP_SHARED_SECRET cannot use default or development placeholder secret in production")
		}
		if len(secret) < MinAppSecretLength {
			return fmt.Errorf("APP_SHARED_SECRET in production must be at least %d characters long for security", MinAppSecretLength)
		}
	}
	return nil
}

func isTestEnv() bool {
	return strings.HasSuffix(os.Args[0], ".test") || flag.Lookup("test.v") != nil
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
