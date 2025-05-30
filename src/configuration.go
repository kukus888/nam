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
	if path == "" {
		return nil, errors.New("configuration file path cannot be empty")
	}
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
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"pass"` // TODO: add password encryption
		Name     string `yaml:"name"`
	} `yaml:"postgres"`
	Logging struct {
		Level     string     `yaml:"level" default:"info"` // Log level, e.g. "debug", "info", "warn", "error"
		SlogLevel slog.Level `yaml:"-"`                    // Internal representation of log level, e.g. 0 for debug, 1 for info, etc.
	} `yaml:"logging"`
}
