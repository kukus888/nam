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
		Level      string     `yaml:"level" default:"info"` // Log level, e.g. "debug", "info", "warn", "error"
		SlogLevel  slog.Level `yaml:"-"`                    // Internal representation of log level, e.g. 0 for debug, 1 for info, etc.
		FilePath   string     `yaml:"filepath"`             // Path to log file
		MaxSize    int        `yaml:"maxsize"`              // Maximum size in megabytes of the log file
		MaxBackups int        `yaml:"maxbackups"`           // Maximum number of old log files to retain
		MaxAge     int        `yaml:"maxage"`               // Maximum number of days to retain old log files
		Compress   bool       `yaml:"compress"`             // Whether to compress old log files
	} `yaml:"logging"`
	Services map[string]bool `yaml:"services"` // Map of service names to their enabled status
	Keys     struct {
		CaCertsPath    string `yaml:"cacerts"`
		ClientCertPath string `yaml:"clientcert"`
		ClientKeyPath  string `yaml:"clientkey"`
	} `yaml:"keys"`
	WebServer struct {
		Enabled bool   `yaml:"enabled"` // Whether the web server is enabled
		Address string `yaml:"address"` // Web server address, e.g. "0.0.0.0:8080"
		Mode    string `yaml:"mode"`    // Gin mode, e.g. "debug", "release", "test"
		TLS     struct {
			Enabled  bool   `yaml:"enabled"`
			CertPath string `yaml:"certpath"`
			KeyPath  string `yaml:"keypath"`
		} `yaml:"tls"`
	} `yaml:"webserver"`
}
