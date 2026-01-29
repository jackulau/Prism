package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// DatabaseConfig holds PostgreSQL connection settings
type DatabaseConfig struct {
	URL      string // DATABASE_URL (SQLite path or connection string)
	Host     string // PGHOST
	Port     int    // PGPORT
	User     string // PGUSER
	Password string // PGPASSWORD
	Database string // PGDATABASE
	SSLMode  string // PGSSL (disable, require, verify-full)
}

// WorkOSConfig holds WorkOS authentication settings
type WorkOSConfig struct {
	APIKey         string // WORKOS_API_KEY
	ClientID       string // WORKOS_CLIENT_ID
	RedirectURI    string // WORKOS_REDIRECT_URI
	CookiePassword string // WORKOS_COOKIE_PASSWORD (32+ chars for encryption)
	WebhookSecret  string // WORKOS_WEBHOOK_SECRET
}

// LLMConfig holds server-side LLM provider API keys
type LLMConfig struct {
	OpenAIAPIKey    string // OPENAI_API_KEY (server-side)
	AnthropicAPIKey string // ANTHROPIC_API_KEY (server-side)
}

// AnalyticsConfig holds PostHog analytics settings
type AnalyticsConfig struct {
	Enabled       bool          // POSTHOG_ENABLED
	APIKey        string        // POSTHOG_API_KEY
	Endpoint      string        // POSTHOG_ENDPOINT
	BatchSize     int           // POSTHOG_BATCH_SIZE
	FlushInterval time.Duration // POSTHOG_FLUSH_INTERVAL
}

// StripeConfig holds Stripe payment settings
type StripeConfig struct {
	SecretKey      string // STRIPE_SECRET_KEY
	WebhookSecret  string // STRIPE_WEBHOOK_SECRET
	PublishableKey string // STRIPE_PUBLISHABLE_KEY (for frontend)
}

// GitHubAppConfig holds GitHub App and OAuth settings
type GitHubAppConfig struct {
	AppID         string // GITHUB_APP_ID
	PrivateKey    string // GITHUB_PRIVATE_KEY (PEM format)
	WebhookSecret string // GITHUB_WEBHOOK_SECRET
	ClientID      string // GITHUB_CLIENT_ID
	ClientSecret  string // GITHUB_CLIENT_SECRET
	RedirectURL   string // GITHUB_REDIRECT_URL
	WebhookEnabled bool  // GITHUB_WEBHOOK_ENABLED
}

type Config struct {
	// Server
	Port        string
	Host        string
	Environment string
	BaseURL     string
	FrontendURL string

	// Database - PostgreSQL or SQLite
	Database DatabaseConfig

	// Security
	EncryptionKey    string
	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration

	// Authentication - WorkOS
	WorkOS WorkOSConfig

	// GitHub App
	GitHubAppID            int64
	GitHubAppPrivateKey    string
	GitHubAppClientID      string
	GitHubAppClientSecret  string
	GitHubAppWebhookSecret string

	// Ollama
	OllamaHost string

	// LM Studio
	LMStudioHost string

	// Sandbox
	SandboxMemoryLimit string
	SandboxCPULimit    string
	SandboxTimeout     time.Duration
	SandboxPreviewURL  string

	// Rate Limiting
	RateLimitRequestsPerMinute int
	RateLimitBurst             int

	// CORS
	CORSAllowedOrigins string

	// Uploads
	UploadMaxSize int64
	UploadDir     string

	// Discord Integration
	DiscordEnabled    bool
	DiscordWebhookURL string
	DiscordBotToken   string

	// Slack Integration
	SlackEnabled    bool
	SlackWebhookURL string
	SlackBotToken   string
	SlackChannelID  string

	// Code Runner
	CodeRunnerEnabled     bool
	CodeRunnerDockerMode  bool
	CodeRunnerMemoryLimit string
	CodeRunnerCPULimit    string
	CodeRunnerTimeout     time.Duration

	// Guest Mode
	GuestModeEnabled bool

	// WorkOS Integration
	WorkOSEnabled       bool
	WorkOSAPIKey        string
	WorkOSClientID      string
	WorkOSWebhookSecret string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		// Server defaults
		Port:        getEnv("PORT", "8080"),
		Host:        getEnv("HOST", "0.0.0.0"),
		Environment: getEnv("ENVIRONMENT", "development"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),

		// Database - PostgreSQL or SQLite
		Database: DatabaseConfig{
			URL:      getEnv("DATABASE_URL", "./data/prism.db"),
			Host:     getEnv("PGHOST", ""),
			Port:     getIntEnv("PGPORT", 5432),
			User:     getEnv("PGUSER", ""),
			Password: getEnv("PGPASSWORD", ""),
			Database: getEnv("PGDATABASE", ""),
			SSLMode:  getEnv("PGSSL", "disable"),
		},

		// Security
		EncryptionKey:    getEnv("ENCRYPTION_KEY", ""),
		JWTSecret:        getEnv("JWT_SECRET", "change-me-in-production"),
		JWTAccessExpiry:  getDurationEnv("JWT_ACCESS_EXPIRY", 15*time.Minute),
		JWTRefreshExpiry: getDurationEnv("JWT_REFRESH_EXPIRY", 7*24*time.Hour),

		// Authentication - WorkOS
		WorkOS: WorkOSConfig{
			APIKey:         getEnv("WORKOS_API_KEY", ""),
			ClientID:       getEnv("WORKOS_CLIENT_ID", ""),
			RedirectURI:    getEnv("WORKOS_REDIRECT_URI", ""),
			CookiePassword: getEnv("WORKOS_COOKIE_PASSWORD", ""),
			WebhookSecret:  getEnv("WORKOS_WEBHOOK_SECRET", ""),
		},

		// GitHub App
		GitHubAppID:            getInt64Env("GITHUB_APP_ID", 0),
		GitHubAppPrivateKey:    getEnv("GITHUB_APP_PRIVATE_KEY", ""),
		GitHubAppClientID:      getEnv("GITHUB_APP_CLIENT_ID", ""),
		GitHubAppClientSecret:  getEnv("GITHUB_APP_CLIENT_SECRET", ""),
		GitHubAppWebhookSecret: getEnv("GITHUB_APP_WEBHOOK_SECRET", ""),

		// Ollama
		OllamaHost: getEnv("OLLAMA_HOST", "http://localhost:11434"),

		// LM Studio
		LMStudioHost: getEnv("LMSTUDIO_HOST", "http://localhost:1234"),

		// Sandbox
		SandboxMemoryLimit: getEnv("SANDBOX_MEMORY_LIMIT", "512m"),
		SandboxCPULimit:    getEnv("SANDBOX_CPU_LIMIT", "0.5"),
		SandboxTimeout:     getDurationEnv("SANDBOX_TIMEOUT", 60*time.Second),
		SandboxPreviewURL:  getEnv("SANDBOX_PREVIEW_URL", ""),

		// Rate Limiting
		RateLimitRequestsPerMinute: getIntEnv("RATE_LIMIT_REQUESTS_PER_MINUTE", 60),
		RateLimitBurst:             getIntEnv("RATE_LIMIT_BURST", 10),

		// CORS
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"),

		// Uploads
		UploadMaxSize: getInt64Env("UPLOAD_MAX_SIZE", 10*1024*1024), // 10MB
		UploadDir:     getEnv("UPLOAD_DIR", "./data/uploads"),

		// Discord Integration
		DiscordEnabled:    getBoolEnv("DISCORD_ENABLED", false),
		DiscordWebhookURL: getEnv("DISCORD_WEBHOOK_URL", ""),
		DiscordBotToken:   getEnv("DISCORD_BOT_TOKEN", ""),

		// Slack Integration
		SlackEnabled:    getBoolEnv("SLACK_ENABLED", false),
		SlackWebhookURL: getEnv("SLACK_WEBHOOK_URL", ""),
		SlackBotToken:   getEnv("SLACK_BOT_TOKEN", ""),
		SlackChannelID:  getEnv("SLACK_CHANNEL_ID", ""),

		// Code Runner
		CodeRunnerEnabled:     getBoolEnv("CODE_RUNNER_ENABLED", true),
		CodeRunnerDockerMode:  getBoolEnv("CODE_RUNNER_DOCKER_MODE", false),
		CodeRunnerMemoryLimit: getEnv("CODE_RUNNER_MEMORY_LIMIT", "512m"),
		CodeRunnerCPULimit:    getEnv("CODE_RUNNER_CPU_LIMIT", "0.5"),
		CodeRunnerTimeout:     getDurationEnv("CODE_RUNNER_TIMEOUT", 5*time.Minute),

		// Guest Mode - disabled by default for security
		GuestModeEnabled: getBoolEnv("GUEST_MODE_ENABLED", false),

		// WorkOS Integration - disabled by default
		WorkOSEnabled:       getBoolEnv("WORKOS_ENABLED", false),
		WorkOSAPIKey:        getEnv("WORKOS_API_KEY", ""),
		WorkOSClientID:      getEnv("WORKOS_CLIENT_ID", ""),
		WorkOSWebhookSecret: getEnv("WORKOS_WEBHOOK_SECRET", ""),
	}

	// Set legacy fields for backward compatibility
	cfg.DatabaseURL = cfg.Database.URL
	cfg.GitHubClientID = cfg.GitHub.ClientID
	cfg.GitHubClientSecret = cfg.GitHub.ClientSecret
	cfg.GitHubRedirectURL = cfg.GitHub.RedirectURL
	cfg.GitHubWebhookEnabled = cfg.GitHub.WebhookEnabled
	cfg.GitHubWebhookSecret = cfg.GitHub.WebhookSecret
	cfg.PostHogEnabled = cfg.Analytics.Enabled
	cfg.PostHogAPIKey = cfg.Analytics.APIKey
	cfg.PostHogEndpoint = cfg.Analytics.Endpoint
	cfg.PostHogBatchSize = cfg.Analytics.BatchSize
	cfg.PostHogFlushInterval = cfg.Analytics.FlushInterval

	// Validate configuration
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate runs all configuration validation checks
func (cfg *Config) validate() error {
	if err := cfg.validateSecurity(); err != nil {
		return err
	}
	if err := cfg.validateDatabase(); err != nil {
		return err
	}
	if err := cfg.validateWorkOS(); err != nil {
		return err
	}
	if err := cfg.validateStripe(); err != nil {
		return err
	}
	return nil
}

// validateSecurity checks security-critical configuration values
func (cfg *Config) validateSecurity() error {
	isProduction := cfg.Environment == "production"

	// Validate JWT secret
	if cfg.JWTSecret == "change-me-in-production" {
		if isProduction {
			return fmt.Errorf("JWT_SECRET must be changed from the default value in production. Generate one with: openssl rand -base64 48")
		}
		log.Println("WARNING: Using default JWT_SECRET. Set a secure value for production.")
	} else if len(cfg.JWTSecret) < 32 {
		if isProduction {
			return fmt.Errorf("JWT_SECRET must be at least 32 characters in production")
		}
		log.Println("WARNING: JWT_SECRET is shorter than recommended (32+ characters)")
	}

	// Validate encryption key
	if cfg.EncryptionKey == "" {
		if isProduction {
			return fmt.Errorf("ENCRYPTION_KEY must be set in production. Generate one with: openssl rand -hex 32")
		}
		log.Println("WARNING: ENCRYPTION_KEY not set. A random key will be generated (data will be lost on restart)")
	}

	// Warn about guest mode in production
	if cfg.GuestModeEnabled && isProduction {
		log.Println("WARNING: Guest mode is enabled in production. This allows unauthenticated access.")
	}

	return nil
}

// validateDatabase checks PostgreSQL configuration completeness
func (cfg *Config) validateDatabase() error {
	// If PostgreSQL host is set, validate that required fields are present
	if cfg.Database.Host != "" {
		if cfg.Database.User == "" || cfg.Database.Database == "" {
			return fmt.Errorf("PGUSER and PGDATABASE are required when PGHOST is set")
		}
	}
	return nil
}

// validateWorkOS checks WorkOS configuration
func (cfg *Config) validateWorkOS() error {
	// If WorkOS is configured, validate cookie password length
	if cfg.WorkOS.APIKey != "" {
		if cfg.WorkOS.ClientID == "" {
			return fmt.Errorf("WORKOS_CLIENT_ID is required when WORKOS_API_KEY is set")
		}
		if len(cfg.WorkOS.CookiePassword) > 0 && len(cfg.WorkOS.CookiePassword) < 32 {
			return fmt.Errorf("WORKOS_COOKIE_PASSWORD must be at least 32 characters")
		}
	}
	return nil
}

// validateStripe checks Stripe configuration
func (cfg *Config) validateStripe() error {
	isProduction := cfg.Environment == "production"

	// In production, if Stripe is configured, webhook secret is recommended
	if isProduction && cfg.Stripe.SecretKey != "" {
		if cfg.Stripe.WebhookSecret == "" {
			log.Println("WARNING: STRIPE_WEBHOOK_SECRET not set in production. Webhooks will not be verified.")
		}
	}
	return nil
}

// IsPostgreSQLConfigured returns true if PostgreSQL connection details are set
func (cfg *Config) IsPostgreSQLConfigured() bool {
	return cfg.Database.Host != ""
}

// IsWorkOSConfigured returns true if WorkOS is configured
func (cfg *Config) IsWorkOSConfigured() bool {
	return cfg.WorkOS.APIKey != "" && cfg.WorkOS.ClientID != ""
}

// IsStripeConfigured returns true if Stripe is configured
func (cfg *Config) IsStripeConfigured() bool {
	return cfg.Stripe.SecretKey != ""
}

// IsGitHubAppConfigured returns true if GitHub App is configured
func (cfg *Config) IsGitHubAppConfigured() bool {
	return cfg.GitHub.AppID != "" && cfg.GitHub.PrivateKey != ""
}

// GetDatabaseConnectionString returns the appropriate connection string
func (cfg *Config) GetDatabaseConnectionString() string {
	if cfg.IsPostgreSQLConfigured() {
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
			cfg.Database.Password, cfg.Database.Database, cfg.Database.SSLMode,
		)
	}
	return cfg.Database.URL
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if value == "true" || value == "1" || value == "yes" {
			return true
		}
		if value == "false" || value == "0" || value == "no" {
			return false
		}
	}
	return defaultValue
}
