// Package config defines application configuration structure
// and the Loader interface for retriving configuration values
package config

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

type Loader interface {
	Load() (Config, error)
}
