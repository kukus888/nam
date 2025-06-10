package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// Loads secrets based on the configuration, returning the secrets
func ParseSecrets(AppConfig ApplicationConfiguration) (*tls.Config, error) {
	// Setup
	tlsConfig := &tls.Config{}
	// Parse keys
	if AppConfig.Keys.CaCertsPath != "" {
		caCertb, err := os.ReadFile(AppConfig.Keys.CaCertsPath)
		if err != nil {
			return nil, err
		}
		certs, err := x509.ParseCertificates(caCertb)
		if err != nil {
			fmt.Println("Unable to parse CA chain")
			return nil, err
		}
		for _, cert := range certs {
			tlsConfig.RootCAs.AddCert(cert)
		}
	}
	if AppConfig.Keys.ClientCertPath != "" && AppConfig.Keys.ClientKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(AppConfig.Keys.ClientCertPath, AppConfig.Keys.ClientKeyPath)
		if err != nil {
			return nil, err
		}
		tlsConfig.ClientCAs.AddCert(cert.Leaf)
	}
	return tlsConfig, nil
}
