package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

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

		// Guest Mode - enabled by default for easy access
		GuestModeEnabled: getBoolEnv("GUEST_MODE_ENABLED", true),
	}

	return cfg, nil
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
