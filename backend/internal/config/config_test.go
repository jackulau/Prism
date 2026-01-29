package config

import (
	"os"
	"testing"
	"time"
)

func TestDatabaseConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid SQLite config",
			config: &Config{
				Database: DatabaseConfig{
					URL: "./data/prism.db",
				},
			},
			wantErr: false,
		},
		{
			name: "valid PostgreSQL config",
			config: &Config{
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "prism",
					Password: "secret",
					Database: "prism",
					SSLMode:  "disable",
				},
			},
			wantErr: false,
		},
		{
			name: "PostgreSQL host without user",
			config: &Config{
				Database: DatabaseConfig{
					Host:     "localhost",
					Database: "prism",
				},
			},
			wantErr: true,
		},
		{
			name: "PostgreSQL host without database",
			config: &Config{
				Database: DatabaseConfig{
					Host: "localhost",
					User: "prism",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validateDatabase()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDatabase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWorkOSConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "no WorkOS config",
			config: &Config{
				WorkOS: WorkOSConfig{},
			},
			wantErr: false,
		},
		{
			name: "valid WorkOS config",
			config: &Config{
				WorkOS: WorkOSConfig{
					APIKey:         "sk_test_123",
					ClientID:       "client_123",
					CookiePassword: "this-is-at-least-32-characters-long",
				},
			},
			wantErr: false,
		},
		{
			name: "WorkOS API key without client ID",
			config: &Config{
				WorkOS: WorkOSConfig{
					APIKey: "sk_test_123",
				},
			},
			wantErr: true,
		},
		{
			name: "WorkOS with short cookie password",
			config: &Config{
				WorkOS: WorkOSConfig{
					APIKey:         "sk_test_123",
					ClientID:       "client_123",
					CookiePassword: "tooshort",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validateWorkOS()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWorkOS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsPostgreSQLConfigured(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{
			name: "SQLite config",
			config: &Config{
				Database: DatabaseConfig{
					URL: "./data/prism.db",
				},
			},
			want: false,
		},
		{
			name: "PostgreSQL config",
			config: &Config{
				Database: DatabaseConfig{
					Host: "localhost",
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.IsPostgreSQLConfigured(); got != tt.want {
				t.Errorf("IsPostgreSQLConfigured() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsWorkOSConfigured(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{
			name: "not configured",
			config: &Config{
				WorkOS: WorkOSConfig{},
			},
			want: false,
		},
		{
			name: "only API key",
			config: &Config{
				WorkOS: WorkOSConfig{
					APIKey: "sk_test_123",
				},
			},
			want: false,
		},
		{
			name: "fully configured",
			config: &Config{
				WorkOS: WorkOSConfig{
					APIKey:   "sk_test_123",
					ClientID: "client_123",
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.IsWorkOSConfigured(); got != tt.want {
				t.Errorf("IsWorkOSConfigured() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsStripeConfigured(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{
			name: "not configured",
			config: &Config{
				Stripe: StripeConfig{},
			},
			want: false,
		},
		{
			name: "configured",
			config: &Config{
				Stripe: StripeConfig{
					SecretKey: "sk_test_123",
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.IsStripeConfigured(); got != tt.want {
				t.Errorf("IsStripeConfigured() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDatabaseConnectionString(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   string
	}{
		{
			name: "SQLite connection",
			config: &Config{
				Database: DatabaseConfig{
					URL: "./data/prism.db",
				},
			},
			want: "./data/prism.db",
		},
		{
			name: "PostgreSQL connection",
			config: &Config{
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "prism",
					Password: "secret",
					Database: "prism",
					SSLMode:  "disable",
				},
			},
			want: "host=localhost port=5432 user=prism password=secret dbname=prism sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.GetDatabaseConnectionString(); got != tt.want {
				t.Errorf("GetDatabaseConnectionString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoad_LegacyCompatibility(t *testing.T) {
	// Set minimal required environment variables
	os.Setenv("JWT_SECRET", "test-secret-that-is-at-least-32-characters-long")
	os.Setenv("ENCRYPTION_KEY", "test-encryption-key-with-32-chars")
	os.Setenv("GITHUB_CLIENT_ID", "test-client-id")
	os.Setenv("GITHUB_CLIENT_SECRET", "test-client-secret")
	os.Setenv("POSTHOG_ENABLED", "true")
	os.Setenv("POSTHOG_API_KEY", "test-posthog-key")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("ENCRYPTION_KEY")
		os.Unsetenv("GITHUB_CLIENT_ID")
		os.Unsetenv("GITHUB_CLIENT_SECRET")
		os.Unsetenv("POSTHOG_ENABLED")
		os.Unsetenv("POSTHOG_API_KEY")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify legacy fields are populated
	if cfg.GitHubClientID != cfg.GitHub.ClientID {
		t.Errorf("Legacy GitHubClientID = %v, want %v", cfg.GitHubClientID, cfg.GitHub.ClientID)
	}
	if cfg.GitHubClientSecret != cfg.GitHub.ClientSecret {
		t.Errorf("Legacy GitHubClientSecret = %v, want %v", cfg.GitHubClientSecret, cfg.GitHub.ClientSecret)
	}
	if cfg.PostHogEnabled != cfg.Analytics.Enabled {
		t.Errorf("Legacy PostHogEnabled = %v, want %v", cfg.PostHogEnabled, cfg.Analytics.Enabled)
	}
	if cfg.PostHogAPIKey != cfg.Analytics.APIKey {
		t.Errorf("Legacy PostHogAPIKey = %v, want %v", cfg.PostHogAPIKey, cfg.Analytics.APIKey)
	}
	if cfg.DatabaseURL != cfg.Database.URL {
		t.Errorf("Legacy DatabaseURL = %v, want %v", cfg.DatabaseURL, cfg.Database.URL)
	}
}

func TestLoad_NewConfigSections(t *testing.T) {
	// Set environment variables for new config sections
	os.Setenv("PGHOST", "test-host")
	os.Setenv("PGPORT", "5433")
	os.Setenv("PGUSER", "testuser")
	os.Setenv("PGPASSWORD", "testpass")
	os.Setenv("PGDATABASE", "testdb")
	os.Setenv("PGSSL", "require")

	os.Setenv("WORKOS_API_KEY", "sk_test")
	os.Setenv("WORKOS_CLIENT_ID", "client_test")
	os.Setenv("WORKOS_COOKIE_PASSWORD", "this-cookie-password-is-32-chars!")

	os.Setenv("OPENAI_API_KEY", "sk-openai-test")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test")

	os.Setenv("STRIPE_SECRET_KEY", "sk_stripe_test")
	os.Setenv("STRIPE_PUBLISHABLE_KEY", "pk_stripe_test")

	os.Setenv("GITHUB_APP_ID", "12345")
	os.Setenv("GITHUB_PRIVATE_KEY", "-----BEGIN RSA PRIVATE KEY-----")

	defer func() {
		os.Unsetenv("PGHOST")
		os.Unsetenv("PGPORT")
		os.Unsetenv("PGUSER")
		os.Unsetenv("PGPASSWORD")
		os.Unsetenv("PGDATABASE")
		os.Unsetenv("PGSSL")
		os.Unsetenv("WORKOS_API_KEY")
		os.Unsetenv("WORKOS_CLIENT_ID")
		os.Unsetenv("WORKOS_COOKIE_PASSWORD")
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
		os.Unsetenv("STRIPE_SECRET_KEY")
		os.Unsetenv("STRIPE_PUBLISHABLE_KEY")
		os.Unsetenv("GITHUB_APP_ID")
		os.Unsetenv("GITHUB_PRIVATE_KEY")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify Database config
	if cfg.Database.Host != "test-host" {
		t.Errorf("Database.Host = %v, want test-host", cfg.Database.Host)
	}
	if cfg.Database.Port != 5433 {
		t.Errorf("Database.Port = %v, want 5433", cfg.Database.Port)
	}
	if cfg.Database.SSLMode != "require" {
		t.Errorf("Database.SSLMode = %v, want require", cfg.Database.SSLMode)
	}

	// Verify WorkOS config
	if cfg.WorkOS.APIKey != "sk_test" {
		t.Errorf("WorkOS.APIKey = %v, want sk_test", cfg.WorkOS.APIKey)
	}
	if cfg.WorkOS.ClientID != "client_test" {
		t.Errorf("WorkOS.ClientID = %v, want client_test", cfg.WorkOS.ClientID)
	}

	// Verify LLM config
	if cfg.LLM.OpenAIAPIKey != "sk-openai-test" {
		t.Errorf("LLM.OpenAIAPIKey = %v, want sk-openai-test", cfg.LLM.OpenAIAPIKey)
	}
	if cfg.LLM.AnthropicAPIKey != "sk-ant-test" {
		t.Errorf("LLM.AnthropicAPIKey = %v, want sk-ant-test", cfg.LLM.AnthropicAPIKey)
	}

	// Verify Stripe config
	if cfg.Stripe.SecretKey != "sk_stripe_test" {
		t.Errorf("Stripe.SecretKey = %v, want sk_stripe_test", cfg.Stripe.SecretKey)
	}

	// Verify GitHub App config
	if cfg.GitHub.AppID != "12345" {
		t.Errorf("GitHub.AppID = %v, want 12345", cfg.GitHub.AppID)
	}
	if !cfg.IsGitHubAppConfigured() {
		t.Error("IsGitHubAppConfigured() = false, want true")
	}
}

func TestAnalyticsConfig_Defaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Analytics.Endpoint != "https://app.posthog.com" {
		t.Errorf("Analytics.Endpoint = %v, want https://app.posthog.com", cfg.Analytics.Endpoint)
	}
	if cfg.Analytics.BatchSize != 10 {
		t.Errorf("Analytics.BatchSize = %v, want 10", cfg.Analytics.BatchSize)
	}
	if cfg.Analytics.FlushInterval != 30*time.Second {
		t.Errorf("Analytics.FlushInterval = %v, want 30s", cfg.Analytics.FlushInterval)
	}
}
