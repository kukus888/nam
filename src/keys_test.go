package main

import "testing"

func TestParseSecretsRespectsAllowInsecureTLS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		allowInsecureTLS     bool
		expectedInsecureSkip bool
	}{
		{
			name:                 "disabled",
			allowInsecureTLS:     false,
			expectedInsecureSkip: false,
		},
		{
			name:                 "enabled",
			allowInsecureTLS:     true,
			expectedInsecureSkip: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var cfg ApplicationConfiguration
			cfg.Security.AllowInsecureTLS = tt.allowInsecureTLS

			tlsConfig, err := ParseSecrets(cfg, nil)
			if err != nil {
				t.Fatalf("ParseSecrets() returned error: %v", err)
			}

			if tlsConfig == nil {
				t.Fatal("expected non-nil tls config")
			}

			if tlsConfig.InsecureSkipVerify != tt.expectedInsecureSkip {
				t.Fatalf("unexpected InsecureSkipVerify value, got %v want %v", tlsConfig.InsecureSkipVerify, tt.expectedInsecureSkip)
			}

			if tlsConfig.RootCAs == nil {
				t.Fatal("expected RootCAs to be initialized")
			}
		})
	}
}
