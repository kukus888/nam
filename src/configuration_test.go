package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndParseConfigurationParsesSecuritySettings(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	configYAML := `postgres:
  dsn: "postgres://user:password@localhost:5432/postgres"
logging:
  level: warn
webserver:
  enabled: true
  mode: debug
  address: "0.0.0.0:8080"
  tls:
    enabled: false
    certpath: ""
    keypath: ""
keys:
  cacerts: ""
  clientcert: ""
  clientkey: ""
security:
  allow_insecure_tls: true
  jwt_key: "test-jwt-key"
  secrets_master_password: "test-master"
  secrets_salt: "test-salt"
services:
  HealthcheckService: true
`

	if err := os.WriteFile(configPath, []byte(configYAML), 0600); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	config, err := LoadAndParseConfiguration(configPath)
	if err != nil {
		t.Fatalf("LoadAndParseConfiguration() returned error: %v", err)
	}

	if !config.Security.AllowInsecureTLS {
		t.Fatal("expected security.allow_insecure_tls to be true")
	}
	if config.Security.JWTKey != "test-jwt-key" {
		t.Fatalf("unexpected jwt_key, got %q", config.Security.JWTKey)
	}
	if config.Security.SecretsMasterPassword != "test-master" {
		t.Fatalf("unexpected secrets_master_password, got %q", config.Security.SecretsMasterPassword)
	}
	if config.Security.SecretsSalt != "test-salt" {
		t.Fatalf("unexpected secrets_salt, got %q", config.Security.SecretsSalt)
	}
}

func TestLoadAndParseConfigurationDefaultsLogLevelToInfo(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	configYAML := `postgres:
  dsn: "postgres://user:password@localhost:5432/postgres"
logging:
  filepath: "application.log"
webserver:
  enabled: false
  mode: debug
  address: "127.0.0.1:8080"
  tls:
    enabled: false
    certpath: ""
    keypath: ""
keys:
  cacerts: ""
  clientcert: ""
  clientkey: ""
security:
  allow_insecure_tls: false
  jwt_key: ""
  secrets_master_password: ""
  secrets_salt: ""
services: {}
`

	if err := os.WriteFile(configPath, []byte(configYAML), 0600); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	config, err := LoadAndParseConfiguration(configPath)
	if err != nil {
		t.Fatalf("LoadAndParseConfiguration() returned error: %v", err)
	}

	if config.Logging.Level != "info" {
		t.Fatalf("expected default log level to be info, got %q", config.Logging.Level)
	}
}
