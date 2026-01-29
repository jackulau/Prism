package config

import (
	"os"
	"testing"
)

func TestRemoteAccessValidation(t *testing.T) {
	// Save original env vars
	origEnabled := os.Getenv("REMOTE_ACCESS_ENABLED")
	origPort := os.Getenv("REMOTE_ACCESS_PORT")
	origMainPort := os.Getenv("PORT")
	origEnv := os.Getenv("ENVIRONMENT")
	origPasswordHash := os.Getenv("REMOTE_ACCESS_PASSWORD_HASH")
	origTLSCert := os.Getenv("REMOTE_ACCESS_TLS_CERT")
	origTLSKey := os.Getenv("REMOTE_ACCESS_TLS_KEY")

	defer func() {
		os.Setenv("REMOTE_ACCESS_ENABLED", origEnabled)
		os.Setenv("REMOTE_ACCESS_PORT", origPort)
		os.Setenv("PORT", origMainPort)
		os.Setenv("ENVIRONMENT", origEnv)
		os.Setenv("REMOTE_ACCESS_PASSWORD_HASH", origPasswordHash)
		os.Setenv("REMOTE_ACCESS_TLS_CERT", origTLSCert)
		os.Setenv("REMOTE_ACCESS_TLS_KEY", origTLSKey)
	}()

	tests := []struct {
		name        string
		envVars     map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name: "disabled - no validation",
			envVars: map[string]string{
				"REMOTE_ACCESS_ENABLED": "false",
			},
			wantErr: false,
		},
		{
			name: "enabled - valid development config",
			envVars: map[string]string{
				"REMOTE_ACCESS_ENABLED": "true",
				"REMOTE_ACCESS_PORT":    "8443",
				"PORT":                  "8080",
				"ENVIRONMENT":           "development",
			},
			wantErr: false,
		},
		{
			name: "enabled - same port as main server",
			envVars: map[string]string{
				"REMOTE_ACCESS_ENABLED": "true",
				"REMOTE_ACCESS_PORT":    "8080",
				"PORT":                  "8080",
				"ENVIRONMENT":           "development",
			},
			wantErr:     true,
			errContains: "must be different from main server PORT",
		},
		{
			name: "enabled - invalid port",
			envVars: map[string]string{
				"REMOTE_ACCESS_ENABLED": "true",
				"REMOTE_ACCESS_PORT":    "0",
				"PORT":                  "8080",
				"ENVIRONMENT":           "development",
			},
			wantErr:     true,
			errContains: "must be between 1 and 65535",
		},
		{
			name: "production - missing password hash",
			envVars: map[string]string{
				"REMOTE_ACCESS_ENABLED":       "true",
				"REMOTE_ACCESS_PORT":          "8443",
				"PORT":                        "8080",
				"ENVIRONMENT":                 "production",
				"REMOTE_ACCESS_PASSWORD_HASH": "",
				"REMOTE_ACCESS_TLS_CERT":      "/path/to/cert",
				"REMOTE_ACCESS_TLS_KEY":       "/path/to/key",
				"JWT_SECRET":                  "a-very-long-secret-key-for-production-at-least-32-chars",
				"ENCRYPTION_KEY":              "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
			wantErr:     true,
			errContains: "REMOTE_ACCESS_PASSWORD_HASH must be set",
		},
		{
			name: "production - missing TLS",
			envVars: map[string]string{
				"REMOTE_ACCESS_ENABLED":       "true",
				"REMOTE_ACCESS_PORT":          "8443",
				"PORT":                        "8080",
				"ENVIRONMENT":                 "production",
				"REMOTE_ACCESS_PASSWORD_HASH": "somehash",
				"REMOTE_ACCESS_TLS_CERT":      "",
				"REMOTE_ACCESS_TLS_KEY":       "",
				"JWT_SECRET":                  "a-very-long-secret-key-for-production-at-least-32-chars",
				"ENCRYPTION_KEY":              "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
			wantErr:     true,
			errContains: "REMOTE_ACCESS_TLS_CERT and REMOTE_ACCESS_TLS_KEY must be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all remote access env vars
			os.Unsetenv("REMOTE_ACCESS_ENABLED")
			os.Unsetenv("REMOTE_ACCESS_PORT")
			os.Unsetenv("PORT")
			os.Unsetenv("ENVIRONMENT")
			os.Unsetenv("REMOTE_ACCESS_PASSWORD_HASH")
			os.Unsetenv("REMOTE_ACCESS_TLS_CERT")
			os.Unsetenv("REMOTE_ACCESS_TLS_KEY")
			os.Unsetenv("JWT_SECRET")
			os.Unsetenv("ENCRYPTION_KEY")

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			_, err := Load()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() expected error containing %q, got nil", tt.errContains)
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Load() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Load() unexpected error = %v", err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
