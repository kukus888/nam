package main

import (
	"errors"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

// Handles loading and storing NAM's configuration

func LoadAndParseConfiguration(path string) (*ApplicationConfiguration, error) {
	var AppConfig ApplicationConfiguration
	AppConfig.Services = make(map[string]bool)
	if path == "" {
		return nil, errors.New("configuration file path cannot be empty")
	}
	// Read configuration file from path
	appCfgB, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	err := yaml.Unmarshal(appCfgB, &AppConfig)
	if err != nil {
		return nil, err
	}
	// Set default values for logging level if not set
	if AppConfig.Logging.Level == "" {
		AppConfig.Logging.Level = "info" // Default log level
	}
	// Convert log level string to integer representation
	switch AppConfig.Logging.Level {
	case "debug":
		AppConfig.Logging.SlogLevel = slog.LevelDebug
	case "info":
		AppConfig.Logging.SlogLevel = slog.LevelInfo
	case "warn":
		AppConfig.Logging.SlogLevel = slog.LevelWarn
	case "error":
		AppConfig.Logging.SlogLevel = slog.LevelError
	default:
		return nil, errors.New("invalid log level: " + AppConfig.Logging.Level + ". Allowed values are: debug, info, warn, error")
	}
	return &AppConfig, nil
}

type ApplicationConfiguration struct {
	Database struct {
		Dsn string `yaml:"dsn"`
	} `yaml:"postgres"`
	Logging struct {
		Level     string     `yaml:"level" default:"info"` // Log level, e.g. "debug", "info", "warn", "error"
		SlogLevel slog.Level `yaml:"-"`                    // Internal representation of log level, e.g. 0 for debug, 1 for info, etc.
	} `yaml:"logging"`
	Services map[string]bool `yaml:"services"` // Map of service names to their enabled status
	Keys     struct {
		CaCertsPath    string `yaml:"cacerts"`
		ClientCertPath string `yaml:"clientcert"`
		ClientKeyPath  string `yaml:"clientkey"`
	} `yaml:"keys"`
	WebServer struct {
		Port    int    `yaml:"port"`    // Port to run the web server on
		Mode    string `yaml:"mode"`    // Gin mode, e.g. "debug", "release", "test"
		Enabled bool   `yaml:"enabled"` // Whether the web server is enabled
	} `yaml:"webserver"`
}
