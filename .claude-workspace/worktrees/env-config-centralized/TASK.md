---
id: env-config-centralized
name: Centralized Environment Configuration
wave: 1
priority: 1
dependencies: []
estimated_hours: 4
tags:
- backend
- config
- infrastructure
---

## Objective

Create a centralized, type-safe environment variable management system with validation, organized by category (Database, Auth, LLM, Analytics, Payments, GitHub).

## Context

The existing `backend/internal/config/config.go` already provides a solid foundation with:
- `godotenv` for loading `.env` files
- Helper functions (`getEnv`, `getIntEnv`, `getBoolEnv`, etc.)
- A `Config` struct with typed fields
- Basic security validation

This task extends the configuration to support the planned environment variables for:
- **Database**: PostgreSQL support (PGHOST, PGPORT, PGUSER, PGPASSWORD, PGDATABASE, PGSSL)
- **Auth**: WorkOS integration (WORKOS_API_KEY, WORKOS_CLIENT_ID, WORKOS_REDIRECT_URI, WORKOS_COOKIE_PASSWORD)
- **LLM**: Additional provider keys (OPENAI_API_KEY - for server-side usage)
- **Analytics**: PostHog (POSTHOG_API_KEY - already partially implemented)
- **Payments**: Stripe (STRIPE_SECRET_KEY, STRIPE_WEBHOOK_SECRET)
- **GitHub**: App credentials (GITHUB_APP_ID, GITHUB_PRIVATE_KEY, GITHUB_WEBHOOK_SECRET)

## Implementation

### 1. Extend Config Struct

Modify `backend/internal/config/config.go` to add new configuration sections:

```go
type Config struct {
    // ... existing fields ...

    // Database - PostgreSQL (for future migration from SQLite)
    Database DatabaseConfig

    // Auth - WorkOS
    WorkOS WorkOSConfig

    // LLM Providers - Server-side keys
    LLM LLMConfig

    // Analytics - PostHog
    Analytics AnalyticsConfig

    // Payments - Stripe
    Stripe StripeConfig

    // GitHub App
    GitHubApp GitHubAppConfig
}

type DatabaseConfig struct {
    URL      string // Existing DATABASE_URL
    Host     string // PGHOST
    Port     int    // PGPORT
    User     string // PGUSER
    Password string // PGPASSWORD
    Database string // PGDATABASE
    SSLMode  string // PGSSL (disable, require, verify-full)
}

type WorkOSConfig struct {
    APIKey         string // WORKOS_API_KEY
    ClientID       string // WORKOS_CLIENT_ID
    RedirectURI    string // WORKOS_REDIRECT_URI
    CookiePassword string // WORKOS_COOKIE_PASSWORD (32+ chars for encryption)
    WebhookSecret  string // WORKOS_WEBHOOK_SECRET
}

type LLMConfig struct {
    OpenAIAPIKey    string // OPENAI_API_KEY (server-side)
    AnthropicAPIKey string // ANTHROPIC_API_KEY (server-side)
}

type AnalyticsConfig struct {
    Enabled       bool   // POSTHOG_ENABLED (existing)
    APIKey        string // POSTHOG_API_KEY
    Endpoint      string // POSTHOG_ENDPOINT
    BatchSize     int    // POSTHOG_BATCH_SIZE
    FlushInterval int    // POSTHOG_FLUSH_INTERVAL
}

type StripeConfig struct {
    SecretKey     string // STRIPE_SECRET_KEY
    WebhookSecret string // STRIPE_WEBHOOK_SECRET
    PublishableKey string // STRIPE_PUBLISHABLE_KEY (for frontend)
}

type GitHubAppConfig struct {
    AppID         string // GITHUB_APP_ID
    PrivateKey    string // GITHUB_PRIVATE_KEY (PEM format)
    WebhookSecret string // GITHUB_WEBHOOK_SECRET
    ClientID      string // GITHUB_CLIENT_ID (existing)
    ClientSecret  string // GITHUB_CLIENT_SECRET (existing)
}
```

### 2. Add Configuration Loading

Update `Load()` function to populate new sections:

```go
func Load() (*Config, error) {
    _ = godotenv.Load()

    cfg := &Config{
        // ... existing fields ...

        Database: DatabaseConfig{
            URL:      getEnv("DATABASE_URL", "./data/prism.db"),
            Host:     getEnv("PGHOST", ""),
            Port:     getIntEnv("PGPORT", 5432),
            User:     getEnv("PGUSER", ""),
            Password: getEnv("PGPASSWORD", ""),
            Database: getEnv("PGDATABASE", ""),
            SSLMode:  getEnv("PGSSL", "disable"),
        },

        WorkOS: WorkOSConfig{
            APIKey:         getEnv("WORKOS_API_KEY", ""),
            ClientID:       getEnv("WORKOS_CLIENT_ID", ""),
            RedirectURI:    getEnv("WORKOS_REDIRECT_URI", ""),
            CookiePassword: getEnv("WORKOS_COOKIE_PASSWORD", ""),
            WebhookSecret:  getEnv("WORKOS_WEBHOOK_SECRET", ""),
        },

        // ... similar for other sections ...
    }

    if err := cfg.validate(); err != nil {
        return nil, err
    }

    return cfg, nil
}
```

### 3. Add Validation Methods

Create category-specific validation:

```go
func (c *Config) validate() error {
    if err := c.validateSecurity(); err != nil {
        return err
    }
    if err := c.validateDatabase(); err != nil {
        return err
    }
    if err := c.validateWorkOS(); err != nil {
        return err
    }
    if err := c.validateStripe(); err != nil {
        return err
    }
    return nil
}

func (c *Config) validateDatabase() error {
    // If PostgreSQL env vars are set, validate they're complete
    if c.Database.Host != "" {
        if c.Database.User == "" || c.Database.Database == "" {
            return fmt.Errorf("PGUSER and PGDATABASE required when PGHOST is set")
        }
    }
    return nil
}

func (c *Config) validateWorkOS() error {
    // If WorkOS is configured, validate cookie password length
    if c.WorkOS.APIKey != "" {
        if len(c.WorkOS.CookiePassword) < 32 {
            return fmt.Errorf("WORKOS_COOKIE_PASSWORD must be at least 32 characters")
        }
    }
    return nil
}

func (c *Config) validateStripe() error {
    // In production, if Stripe is configured, webhook secret is required
    if c.Environment == "production" && c.Stripe.SecretKey != "" {
        if c.Stripe.WebhookSecret == "" {
            log.Println("WARNING: STRIPE_WEBHOOK_SECRET not set in production")
        }
    }
    return nil
}
```

### 4. Add Helper Methods

```go
// IsPostgreSQLConfigured returns true if PostgreSQL connection details are set
func (c *Config) IsPostgreSQLConfigured() bool {
    return c.Database.Host != ""
}

// IsWorkOSConfigured returns true if WorkOS is configured
func (c *Config) IsWorkOSConfigured() bool {
    return c.WorkOS.APIKey != "" && c.WorkOS.ClientID != ""
}

// IsStripeConfigured returns true if Stripe is configured
func (c *Config) IsStripeConfigured() bool {
    return c.Stripe.SecretKey != ""
}

// GetDatabaseConnectionString returns the appropriate connection string
func (c *Config) GetDatabaseConnectionString() string {
    if c.IsPostgreSQLConfigured() {
        return fmt.Sprintf(
            "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
            c.Database.Host, c.Database.Port, c.Database.User,
            c.Database.Password, c.Database.Database, c.Database.SSLMode,
        )
    }
    return c.Database.URL
}
```

### 5. Update .env.example

Add new environment variables to the example file with clear documentation:

```bash
# ===============================
# DATABASE CONFIGURATION
# ===============================
# SQLite (default)
DATABASE_URL=./data/prism.db

# PostgreSQL (optional - set PGHOST to enable)
# PGHOST=localhost
# PGPORT=5432
# PGUSER=prism
# PGPASSWORD=your_password
# PGDATABASE=prism
# PGSSL=disable  # disable, require, verify-full

# ===============================
# AUTHENTICATION (WorkOS)
# ===============================
# WORKOS_API_KEY=sk_...
# WORKOS_CLIENT_ID=client_...
# WORKOS_REDIRECT_URI=http://localhost:8080/api/v1/auth/callback
# WORKOS_COOKIE_PASSWORD=minimum-32-character-secret-key-here
# WORKOS_WEBHOOK_SECRET=whsec_...

# ... etc for other categories
```

## Acceptance Criteria

- [ ] Config struct has organized sections for all environment variable categories
- [ ] New environment variables are loaded with appropriate defaults
- [ ] Validation ensures related variables are set together (e.g., PGHOST requires PGUSER)
- [ ] Helper methods identify which services are configured
- [ ] .env.example documents all new environment variables
- [ ] Existing functionality is not broken
- [ ] Log warnings for production security issues

## Files to Create/Modify

- `backend/internal/config/config.go` - Extend with new config sections
- `.env.example` - Add new environment variable documentation
- `backend/.env.example` - Mirror updates if exists

## Integration Points

- **Provides**: Centralized configuration for all backend services
- **Consumes**: Environment variables from system/file
- **Conflicts**: None - this is foundational infrastructure
