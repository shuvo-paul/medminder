// Package config defines application configuration structure
// and functions for loading configuration values from environment variables
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

var ErrInvalidConfig = errors.New("invalid config")

type Config struct {
	AppPort  int
	Database DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  bool
}

func Load() (Config, error) {
	cfg := Config{}

	port, err := strconv.Atoi(getEnv("APP_PORT", "8080"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid APP_PORT: %w", err)
	}
	cfg.AppPort = port

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid DB_PORT: %w", err)
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

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
