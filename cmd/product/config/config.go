// Package config provides configuration management for the product service.
package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config holds the configuration for the product service.
	Config struct {
		App  `yaml:"app"`
		HTTP `yaml:"http"`
		Log  `yaml:"logger"`
	}

	// App holds the application-level configuration.
	App struct {
		Name    string `yaml:"name" env:"APP_NAME" env-default:"product-service" env-description:"The name of the application"`
		Version string `yaml:"version" env:"APP_VERSION" env-default:"1.0.0" env-description:"The version of the application"`
	}

	// HTTP holds the HTTP server configuration.
	HTTP struct {
		Host string `yaml:"host" env:"HTTP_HOST" env-default:"0.0.0.0" env-description:"The host of the HTTP server"`
		Port string `env-required:"true" yaml:"port" env:"HTTP_PORT" env-description:"The port of the HTTP server"`
	}

	// Log holds the logging configuration.
	Log struct {
		Level string `yaml:"log_level" env:"LOG_LEVEL" env-default:"debug" env-description:"The log level of the application"`
	}
)

// NewConfig reads the configuration from the specified path and returns a Config instance.
// It uses the cleanenv package to read the configuration from a YAML file and environment variables.
// If there is an error reading the configuration, it returns an error.
// Parameters:
//   - path: The path to the configuration file.
//
// Returns:
//   - A pointer to a Config instance.
//   - An error if the configuration could not be read.
func NewConfig(path string) (*Config, error) {
	cfg := &Config{}

	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}
