package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// Loads secrets based on the configuration, returning the secrets
func ParseSecrets(AppConfig ApplicationConfiguration) (*tls.Config, error) {
	// Setup
	tlsConfig := &tls.Config{
		RootCAs:      x509.NewCertPool(),
		Certificates: make([]tls.Certificate, 0),
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
			fmt.Printf("Parsing CA: %s\n", pemB.Type)
			cert, err := x509.ParseCertificate(pemB.Bytes)
			if err != nil {
				fmt.Println("unable to parse CA chain")
				return nil, err
			}
			tlsConfig.RootCAs.AddCert(cert)
		}
	}
	if AppConfig.Keys.ClientCertPath != "" && AppConfig.Keys.ClientKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(AppConfig.Keys.ClientCertPath, AppConfig.Keys.ClientKeyPath)
		if err != nil {
			fmt.Println("unable to parse client X509 key pair")
			return nil, err
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	}
	return tlsConfig, nil
}
