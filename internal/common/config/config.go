// Package config defines application configuration structure
// and functions for loading configuration values from environment variables.
//
// Environment Variables:
//
//	APP_ENV    - Application environment: development or production (default: development)
//	APP_PORT   - HTTP server port (default: 8080)
//	DB_HOST     - Database host (default: localhost)
//	DB_PORT     - Database port (default: 5432)
//	DB_USER     - Database user (default: medminder)
//	DB_PASSWORD - Database password (default: medminder)
//	DB_NAME     - Database name (default: medminder)
//	DB_SSLMODE  - Enable SSL for database connection (default: false)
//	JWT_SECRET  - Secret key for JWT signing (required in production)
package config

import (
	"fmt"
	"os"
	"strconv"
)

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  bool
}

type JWTConfig struct {
	Secret string
}

type EmailConfig struct {
	BaseURL     string
	FromAddress string
	FromName    string
}

type Config struct {
	AppPort  int
	AppEnv   string
	Database DatabaseConfig
	JWT      JWTConfig
	Email    EmailConfig
}

func Load() (Config, error) {
	cfg := Config{}

	cfg.AppEnv = getEnv("APP_ENV", "development")

	port, err := strconv.Atoi(getEnv("APP_PORT", "8080"))
	if err != nil || port < 0 || port > 65535 {
		return Config{}, fmt.Errorf("invalid APP_PORT: must be a number between 0 and 65535")
	}
	cfg.AppPort = port

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil || dbPort < 0 || dbPort > 65535 {
		return Config{}, fmt.Errorf("invalid DB_PORT: must be a number between 0 and 65535")
	}

	sslMode, err := strconv.ParseBool(getEnv("DB_SSLMODE", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid DB_SSLMODE: %w", err)
	}

	cfg.Database = DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     dbPort,
		User:     getEnv("DB_USER", "medminder"),
		Password: getEnv("DB_PASSWORD", "medminder"),
		Name:     getEnv("DB_NAME", "medminder"),
		SSLMode:  sslMode,
	}

	jwtSecret := getEnv("JWT_SECRET", "")
	if cfg.AppEnv == "production" && jwtSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required in production")
	}
	if jwtSecret == "" {
		jwtSecret = "medminder-secret-key-change-in-production"
	}
	cfg.JWT = JWTConfig{Secret: jwtSecret}

	emailBaseURL := getEnv("EMAIL_SERVICE_BASE_URL", "http://localhost:9000")
	emailFromAddress := getEnv("EMAIL_FROM_ADDRESS", "noreply@medminder.app")
	emailFromName := getEnv("EMAIL_FROM_NAME", "MedMinder")
	cfg.Email = EmailConfig{
		BaseURL:     emailBaseURL,
		FromAddress: emailFromAddress,
		FromName:    emailFromName,
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
