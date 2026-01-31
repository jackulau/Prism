package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// RemoteAccessConfig contains configuration for remote access functionality
type RemoteAccessConfig struct {
	// Enabled determines if remote access is available
	Enabled bool
	// Port for remote connections (separate from main server port)
	Port int
	// Host to bind for remote access (default: 0.0.0.0)
	Host string
	// PasswordHash is the Argon2id hashed password for remote access authentication
	PasswordHash string
	// TLSCertPath is the path to TLS certificate for secure connections
	TLSCertPath string
	// TLSKeyPath is the path to TLS private key
	TLSKeyPath string
	// MaxConnections is the maximum concurrent remote connections allowed
	MaxConnections int
	// SessionTimeout is the duration after which idle remote sessions expire
	SessionTimeout time.Duration
	// AllowedIPs is a whitelist of IP addresses/CIDRs allowed to connect (empty = allow all)
	AllowedIPs []string
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	URL      string
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// SessionConfig contains session timeout and cleanup settings
type SessionConfig struct {
	IdleTimeout     time.Duration // How long before an idle session expires (default: 30m)
	MaxPerUser      int           // Maximum concurrent sessions per user (default: 10)
	CleanupInterval time.Duration // How often to clean up expired/idle sessions (default: 5m)
}

// WorkOSConfig contains WorkOS authentication settings
type WorkOSConfig struct {
	APIKey         string
	ClientID       string
	RedirectURI    string
	CookiePassword string
	WebhookSecret  string
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

	// Session Management
	Session SessionConfig

	// Authentication - WorkOS
	WorkOS WorkOSConfig

	// GitHub App
	GitHubAppID            int64
	GitHubAppPrivateKey    string
	GitHubAppClientID      string
	GitHubAppClientSecret  string
	GitHubAppWebhookSecret string

	// GitHub OAuth (for user authentication)
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string

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

	// PostHog Analytics
	PostHogEnabled       bool
	PostHogAPIKey        string
	PostHogEndpoint      string
	PostHogBatchSize     int
	PostHogFlushInterval time.Duration
	PostHogProjectID     string // Project ID for PostHog query tools

	// PostHog Tools API (for MCP tools)
	PostHogToolsAPIKey  string
	PostHogToolsProject string
	PostHogToolsHost    string

	// GitHub Webhooks
	GitHubWebhookEnabled bool
	GitHubWebhookSecret  string

	// Code Runner
	CodeRunnerEnabled     bool
	CodeRunnerDockerMode  bool
	CodeRunnerMemoryLimit string
	CodeRunnerCPULimit    string
	CodeRunnerTimeout     time.Duration

	// Guest Mode
	GuestModeEnabled bool

	// WorkOS SSO Configuration
	WorkOSAPIKey         string
	WorkOSClientID       string
	WorkOSRedirectURI    string
	WorkOSCookiePassword string
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

		// Session Management
		Session: SessionConfig{
			IdleTimeout:     getDurationEnv("SESSION_IDLE_TIMEOUT", 30*time.Minute),
			MaxPerUser:      getIntEnv("SESSION_MAX_PER_USER", 10),
			CleanupInterval: getDurationEnv("SESSION_CLEANUP_INTERVAL", 5*time.Minute),
		},

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

		// GitHub OAuth (for user authentication)
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURL:  getEnv("GITHUB_REDIRECT_URL", ""),

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

		// PostHog Analytics
		PostHogEnabled:       getBoolEnv("POSTHOG_ENABLED", false),
		PostHogAPIKey:        getEnv("POSTHOG_API_KEY", ""),
		PostHogEndpoint:      getEnv("POSTHOG_ENDPOINT", "https://app.posthog.com"),
		PostHogBatchSize:     getIntEnv("POSTHOG_BATCH_SIZE", 10),
		PostHogFlushInterval: getDurationEnv("POSTHOG_FLUSH_INTERVAL", 30*time.Second),
		PostHogProjectID:     getEnv("POSTHOG_PROJECT_ID", ""),

		// PostHog Tools API (for MCP tools)
		PostHogToolsAPIKey:  getEnv("POSTHOG_TOOLS_API_KEY", ""),
		PostHogToolsProject: getEnv("POSTHOG_TOOLS_PROJECT_ID", ""),
		PostHogToolsHost:    getEnv("POSTHOG_TOOLS_HOST", "https://app.posthog.com"),

		// GitHub Webhooks
		GitHubWebhookEnabled: getBoolEnv("GITHUB_WEBHOOK_ENABLED", false),
		GitHubWebhookSecret:  getEnv("GITHUB_WEBHOOK_SECRET", ""),

		// Code Runner
		CodeRunnerEnabled:     getBoolEnv("CODE_RUNNER_ENABLED", true),
		CodeRunnerDockerMode:  getBoolEnv("CODE_RUNNER_DOCKER_MODE", false),
		CodeRunnerMemoryLimit: getEnv("CODE_RUNNER_MEMORY_LIMIT", "512m"),
		CodeRunnerCPULimit:    getEnv("CODE_RUNNER_CPU_LIMIT", "0.5"),
		CodeRunnerTimeout:     getDurationEnv("CODE_RUNNER_TIMEOUT", 5*time.Minute),

		// Guest Mode - disabled by default for security
		GuestModeEnabled: getBoolEnv("GUEST_MODE_ENABLED", false),

		// WorkOS SSO Configuration
		WorkOSAPIKey:         getEnv("WORKOS_API_KEY", ""),
		WorkOSClientID:       getEnv("WORKOS_CLIENT_ID", ""),
		WorkOSRedirectURI:    getEnv("WORKOS_REDIRECT_URI", "http://localhost:8080/api/v1/auth/sso/callback"),
		WorkOSCookiePassword: getEnv("WORKOS_COOKIE_PASSWORD", ""),
	}

	// Validate configuration
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	// Validate remote access configuration
	if err := cfg.validateRemoteAccess(); err != nil {
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

	// Warn if WorkOS SSO is not fully configured
	workosConfigured := cfg.WorkOSAPIKey != "" && cfg.WorkOSClientID != "" && cfg.WorkOSCookiePassword != ""
	if !workosConfigured {
		if cfg.WorkOSAPIKey != "" || cfg.WorkOSClientID != "" {
			log.Println("WARNING: WorkOS SSO is partially configured. Set WORKOS_API_KEY, WORKOS_CLIENT_ID, and WORKOS_COOKIE_PASSWORD to enable SSO.")
		}
	}

	return nil
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

func getStringSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Split by comma and trim whitespace
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}

// validateDatabase validates database configuration
func (cfg *Config) validateDatabase() error {
	// Database URL is required
	if cfg.Database.URL == "" && cfg.Database.Host == "" {
		log.Println("WARNING: No database configured. Using default SQLite database.")
	}
	return nil
}

// validateWorkOS validates WorkOS SSO configuration
func (cfg *Config) validateWorkOS() error {
	// WorkOS is optional, just validate if partially configured
	if cfg.WorkOS.APIKey != "" || cfg.WorkOS.ClientID != "" {
		if cfg.WorkOS.APIKey == "" || cfg.WorkOS.ClientID == "" {
			log.Println("WARNING: WorkOS is partially configured. Both WORKOS_API_KEY and WORKOS_CLIENT_ID are required.")
		}
	}
	return nil
}

// validateStripe validates Stripe payment configuration
func (cfg *Config) validateStripe() error {
	// Stripe is optional
	return nil
}

// validateRemoteAccess validates remote access configuration
func (cfg *Config) validateRemoteAccess() error {
	// Remote access is optional
	return nil
}
