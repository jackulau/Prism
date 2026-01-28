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

type Config struct {
	// Server
	Port        string
	Host        string
	Environment string
	BaseURL     string
	FrontendURL string

	// Database
	DatabaseURL string

	// Security
	EncryptionKey     string
	JWTSecret         string
	JWTAccessExpiry   time.Duration
	JWTRefreshExpiry  time.Duration

	// GitHub OAuth
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string

	// Ollama
	OllamaHost string

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

	// Remote Access
	RemoteAccess RemoteAccessConfig
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

		// Database
		DatabaseURL: getEnv("DATABASE_URL", "./data/prism.db"),

		// Security
		EncryptionKey:    getEnv("ENCRYPTION_KEY", ""),
		JWTSecret:        getEnv("JWT_SECRET", "change-me-in-production"),
		JWTAccessExpiry:  getDurationEnv("JWT_ACCESS_EXPIRY", 15*time.Minute),
		JWTRefreshExpiry: getDurationEnv("JWT_REFRESH_EXPIRY", 7*24*time.Hour),

		// GitHub OAuth
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURL:  getEnv("GITHUB_REDIRECT_URL", "http://localhost:8080/api/v1/github/callback"),

		// Ollama
		OllamaHost: getEnv("OLLAMA_HOST", "http://localhost:11434"),

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

		// Remote Access - disabled by default for security
		RemoteAccess: RemoteAccessConfig{
			Enabled:        getBoolEnv("REMOTE_ACCESS_ENABLED", false),
			Port:           getIntEnv("REMOTE_ACCESS_PORT", 8443),
			Host:           getEnv("REMOTE_ACCESS_HOST", "0.0.0.0"),
			PasswordHash:   getEnv("REMOTE_ACCESS_PASSWORD_HASH", ""),
			TLSCertPath:    getEnv("REMOTE_ACCESS_TLS_CERT", ""),
			TLSKeyPath:     getEnv("REMOTE_ACCESS_TLS_KEY", ""),
			MaxConnections: getIntEnv("REMOTE_ACCESS_MAX_CONNECTIONS", 10),
			SessionTimeout: getDurationEnv("REMOTE_ACCESS_SESSION_TIMEOUT", 1*time.Hour),
			AllowedIPs:     getStringSliceEnv("REMOTE_ACCESS_ALLOWED_IPS", nil),
		},
	}

	// Validate security configuration in production
	if err := cfg.validateSecurity(); err != nil {
		return nil, err
	}

	// Validate remote access configuration
	if err := cfg.validateRemoteAccess(); err != nil {
		return nil, err
	}

	return cfg, nil
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

// validateRemoteAccess checks remote access configuration
func (cfg *Config) validateRemoteAccess() error {
	// Skip validation if remote access is disabled
	if !cfg.RemoteAccess.Enabled {
		return nil
	}

	isProduction := cfg.Environment == "production"

	// Validate port is different from main server port
	mainPort := 8080
	if p, err := strconv.Atoi(cfg.Port); err == nil {
		mainPort = p
	}
	if cfg.RemoteAccess.Port == mainPort {
		return fmt.Errorf("REMOTE_ACCESS_PORT (%d) must be different from main server PORT (%d)", cfg.RemoteAccess.Port, mainPort)
	}

	// Validate port range
	if cfg.RemoteAccess.Port < 1 || cfg.RemoteAccess.Port > 65535 {
		return fmt.Errorf("REMOTE_ACCESS_PORT must be between 1 and 65535, got %d", cfg.RemoteAccess.Port)
	}

	// Require password hash when enabled
	if cfg.RemoteAccess.PasswordHash == "" {
		if isProduction {
			return fmt.Errorf("REMOTE_ACCESS_PASSWORD_HASH must be set when remote access is enabled in production")
		}
		log.Println("WARNING: Remote access enabled without password. Set REMOTE_ACCESS_PASSWORD_HASH for security.")
	}

	// Require TLS in production
	if isProduction {
		if cfg.RemoteAccess.TLSCertPath == "" || cfg.RemoteAccess.TLSKeyPath == "" {
			return fmt.Errorf("REMOTE_ACCESS_TLS_CERT and REMOTE_ACCESS_TLS_KEY must be set for remote access in production")
		}
	} else {
		// Warn in development if TLS not configured
		if cfg.RemoteAccess.TLSCertPath == "" || cfg.RemoteAccess.TLSKeyPath == "" {
			log.Println("WARNING: Remote access TLS not configured. Connections will be unencrypted.")
		}
	}

	// Warn if no IP restrictions in production
	if isProduction && len(cfg.RemoteAccess.AllowedIPs) == 0 {
		log.Println("WARNING: Remote access has no IP restrictions. Consider setting REMOTE_ACCESS_ALLOWED_IPS.")
	}

	// Validate max connections
	if cfg.RemoteAccess.MaxConnections < 1 {
		return fmt.Errorf("REMOTE_ACCESS_MAX_CONNECTIONS must be at least 1, got %d", cfg.RemoteAccess.MaxConnections)
	}

	log.Printf("Remote access enabled on %s:%d (max %d connections, timeout %s)",
		cfg.RemoteAccess.Host, cfg.RemoteAccess.Port, cfg.RemoteAccess.MaxConnections, cfg.RemoteAccess.SessionTimeout)

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
