package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Handles loading and storing NAM's configuration

func LoadAndParseConfiguration(path string) ApplicationConfiguration {
	var AppConfig ApplicationConfiguration
	appCfgB, e := os.ReadFile(path)
	if e != nil {
		panic(e)
	}
	yaml.Unmarshal(appCfgB, &AppConfig)
	return AppConfig
}

type ApplicationConfiguration struct {
	Connectors struct {
		SQLite SQLiteConnector `yaml:"sqlite"`
		Mongo  struct {
			ConnectionString string `yaml:"connectionString"`
		} `yaml:"mongo"`
	} `yaml:"connectors"`
}

type SQLiteConnector struct {
	DatabasePath string `yaml:"databasePath"`
}
