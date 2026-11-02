package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"log/slog"
	"os"
)

// Loads secrets based on the configuration, returning the secrets
func ParseSecrets(AppConfig ApplicationConfiguration, logger *slog.Logger) (*tls.Config, error) {
	if logger == nil {
		logger = slog.Default()
	}
	// Setup
	tlsConfig := &tls.Config{
		RootCAs:            x509.NewCertPool(),
		Certificates:       make([]tls.Certificate, 0),
		InsecureSkipVerify: true, // TODO: Properly control
	}
	// Parse keys
	if AppConfig.Keys.CaCertsPath != "" {
		caCertb, err := os.ReadFile(AppConfig.Keys.CaCertsPath)
		if err != nil {
			return nil, err
		}
		for {
			var pemB *pem.Block
			pemB, caCertb = pem.Decode(caCertb)
			if pemB == nil {
				break
			}
			logger.Debug("Parsing CA certificate", "pem_type", pemB.Type)
			cert, err := x509.ParseCertificate(pemB.Bytes)
			if err != nil {
				logger.Error("Unable to parse CA certificate", "pem_type", pemB.Type, "error", err)
				return nil, err
			}
			tlsConfig.RootCAs.AddCert(cert)
		}
	}
	if AppConfig.Keys.ClientCertPath != "" && AppConfig.Keys.ClientKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(AppConfig.Keys.ClientCertPath, AppConfig.Keys.ClientKeyPath)
		if err != nil {
			logger.Error("Unable to parse client X509 key pair", "error", err)
			return nil, err
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	}
	return tlsConfig, nil
}
