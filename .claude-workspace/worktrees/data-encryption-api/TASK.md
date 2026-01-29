---
id: data-encryption-api
name: Data Encryption API for Integration Credentials
wave: 1
priority: 1
dependencies: []
estimated_hours: 4
tags:
- backend
- security
- database
---

## Objective

Create a generic `setDataConfig()` and `getDataConfig()` API for encrypting and storing arbitrary integration credentials in JSONB format, building on the existing AES-256-GCM encryption infrastructure.

## Context

The codebase already has robust encryption infrastructure in `backend/internal/security/crypto.go`:
- AES-256-GCM encryption with random nonces
- `Encrypt()` / `Decrypt()` methods
- Password hashing with Argon2id
- API key generation

Current integration-specific storage exists for:
- Provider API keys (`provider_keys` table)
- GitHub connections (`github_connections` table)
- Discord/Slack settings (dedicated tables with encrypted fields)

However, there's no **generic** encrypted config storage for arbitrary integration data. The `user_integrations` table has a `config TEXT` field that stores unencrypted JSON.

This task creates a unified API for storing encrypted configuration data for any integration type.

## Implementation

### 1. Create Generic Encrypted Config Repository

Create `backend/internal/database/repository/data_config.go`:

```go
package repository

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "time"

    "prism/internal/security"
)

// DataConfig represents an encrypted configuration entry
type DataConfig struct {
    ID           string                 `json:"id"`
    UserID       string                 `json:"user_id"`
    ConfigType   string                 `json:"config_type"`   // e.g., "stripe", "workos", "custom_api"
    ConfigKey    string                 `json:"config_key"`    // e.g., "webhook_config", "api_credentials"
    Data         map[string]interface{} `json:"data"`          // Decrypted JSON data
    CreatedAt    time.Time              `json:"created_at"`
    UpdatedAt    time.Time              `json:"updated_at"`
}

// DataConfigRepository handles encrypted configuration storage
type DataConfigRepository struct {
    db     *sql.DB
    crypto *security.EncryptionService
}

// NewDataConfigRepository creates a new repository instance
func NewDataConfigRepository(db *sql.DB, crypto *security.EncryptionService) *DataConfigRepository {
    return &DataConfigRepository{db: db, crypto: crypto}
}

// SetDataConfig encrypts and stores configuration data
// If config already exists for user/type/key, it updates; otherwise creates new
func (r *DataConfigRepository) SetDataConfig(userID, configType, configKey string, data map[string]interface{}) error {
    // Serialize data to JSON
    jsonData, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("failed to serialize config data: %w", err)
    }

    // Encrypt the JSON data
    encryptedData, nonce, err := r.crypto.Encrypt(jsonData)
    if err != nil {
        return fmt.Errorf("failed to encrypt config data: %w", err)
    }

    now := time.Now()

    // Upsert: insert or update on conflict
    _, err = r.db.Exec(`
        INSERT INTO data_configs (id, user_id, config_type, config_key, encrypted_data, data_nonce, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(user_id, config_type, config_key) DO UPDATE SET
            encrypted_data = excluded.encrypted_data,
            data_nonce = excluded.data_nonce,
            updated_at = excluded.updated_at
    `, generateID(), userID, configType, configKey, encryptedData, nonce, now, now)

    if err != nil {
        return fmt.Errorf("failed to store config data: %w", err)
    }

    return nil
}

// GetDataConfig retrieves and decrypts configuration data
// Returns nil, nil if not found (not an error)
func (r *DataConfigRepository) GetDataConfig(userID, configType, configKey string) (*DataConfig, error) {
    var config DataConfig
    var encryptedData, nonce []byte

    err := r.db.QueryRow(`
        SELECT id, user_id, config_type, config_key, encrypted_data, data_nonce, created_at, updated_at
        FROM data_configs
        WHERE user_id = ? AND config_type = ? AND config_key = ?
    `, userID, configType, configKey).Scan(
        &config.ID, &config.UserID, &config.ConfigType, &config.ConfigKey,
        &encryptedData, &nonce, &config.CreatedAt, &config.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, nil // Not found is not an error
    }
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve config data: %w", err)
    }

    // Decrypt the data
    decryptedData, err := r.crypto.Decrypt(encryptedData, nonce)
    if err != nil {
        return nil, fmt.Errorf("failed to decrypt config data: %w", err)
    }

    // Deserialize JSON
    if err := json.Unmarshal(decryptedData, &config.Data); err != nil {
        return nil, fmt.Errorf("failed to deserialize config data: %w", err)
    }

    return &config, nil
}

// DeleteDataConfig removes a configuration entry
func (r *DataConfigRepository) DeleteDataConfig(userID, configType, configKey string) error {
    result, err := r.db.Exec(`
        DELETE FROM data_configs
        WHERE user_id = ? AND config_type = ? AND config_key = ?
    `, userID, configType, configKey)

    if err != nil {
        return fmt.Errorf("failed to delete config data: %w", err)
    }

    rows, _ := result.RowsAffected()
    if rows == 0 {
        return sql.ErrNoRows
    }

    return nil
}

// ListDataConfigs returns all configs for a user and type (without decrypting data)
func (r *DataConfigRepository) ListDataConfigs(userID, configType string) ([]DataConfig, error) {
    rows, err := r.db.Query(`
        SELECT id, user_id, config_type, config_key, created_at, updated_at
        FROM data_configs
        WHERE user_id = ? AND config_type = ?
        ORDER BY config_key
    `, userID, configType)
    if err != nil {
        return nil, fmt.Errorf("failed to list configs: %w", err)
    }
    defer rows.Close()

    var configs []DataConfig
    for rows.Next() {
        var cfg DataConfig
        if err := rows.Scan(&cfg.ID, &cfg.UserID, &cfg.ConfigType, &cfg.ConfigKey, &cfg.CreatedAt, &cfg.UpdatedAt); err != nil {
            return nil, fmt.Errorf("failed to scan config row: %w", err)
        }
        configs = append(configs, cfg)
    }

    return configs, nil
}

// HasDataConfig checks if a configuration exists without retrieving it
func (r *DataConfigRepository) HasDataConfig(userID, configType, configKey string) (bool, error) {
    var exists int
    err := r.db.QueryRow(`
        SELECT 1 FROM data_configs
        WHERE user_id = ? AND config_type = ? AND config_key = ?
    `, userID, configType, configKey).Scan(&exists)

    if err == sql.ErrNoRows {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return true, nil
}
```

### 2. Add Database Migration

Add to `backend/internal/database/sqlite.go` (or create migration file):

```sql
CREATE TABLE IF NOT EXISTS data_configs (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    config_type TEXT NOT NULL,
    config_key TEXT NOT NULL,
    encrypted_data BLOB NOT NULL,
    data_nonce BLOB NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(user_id, config_type, config_key)
);

CREATE INDEX IF NOT EXISTS idx_data_configs_user_type ON data_configs(user_id, config_type);
```

### 3. Create HTTP Handler

Create `backend/internal/api/handlers/data_config.go`:

```go
package handlers

import (
    "github.com/gofiber/fiber/v2"
    "prism/internal/database/repository"
)

type DataConfigHandler struct {
    repo *repository.DataConfigRepository
}

func NewDataConfigHandler(repo *repository.DataConfigRepository) *DataConfigHandler {
    return &DataConfigHandler{repo: repo}
}

// SetConfig handles POST /api/v1/config/:type/:key
func (h *DataConfigHandler) SetConfig(c *fiber.Ctx) error {
    userID := c.Locals("userID").(string)
    configType := c.Params("type")
    configKey := c.Params("key")

    var data map[string]interface{}
    if err := c.BodyParser(&data); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid JSON body",
        })
    }

    if err := h.repo.SetDataConfig(userID, configType, configKey, data); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to store configuration",
        })
    }

    return c.JSON(fiber.Map{
        "success": true,
        "message": "Configuration saved",
    })
}

// GetConfig handles GET /api/v1/config/:type/:key
func (h *DataConfigHandler) GetConfig(c *fiber.Ctx) error {
    userID := c.Locals("userID").(string)
    configType := c.Params("type")
    configKey := c.Params("key")

    config, err := h.repo.GetDataConfig(userID, configType, configKey)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to retrieve configuration",
        })
    }

    if config == nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "Configuration not found",
        })
    }

    return c.JSON(fiber.Map{
        "data": config.Data,
        "updated_at": config.UpdatedAt,
    })
}

// DeleteConfig handles DELETE /api/v1/config/:type/:key
func (h *DataConfigHandler) DeleteConfig(c *fiber.Ctx) error {
    userID := c.Locals("userID").(string)
    configType := c.Params("type")
    configKey := c.Params("key")

    if err := h.repo.DeleteDataConfig(userID, configType, configKey); err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "Configuration not found",
        })
    }

    return c.JSON(fiber.Map{
        "success": true,
        "message": "Configuration deleted",
    })
}

// ListConfigs handles GET /api/v1/config/:type
func (h *DataConfigHandler) ListConfigs(c *fiber.Ctx) error {
    userID := c.Locals("userID").(string)
    configType := c.Params("type")

    configs, err := h.repo.ListDataConfigs(userID, configType)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to list configurations",
        })
    }

    // Return keys only (not decrypted data)
    keys := make([]string, len(configs))
    for i, cfg := range configs {
        keys[i] = cfg.ConfigKey
    }

    return c.JSON(fiber.Map{
        "keys": keys,
    })
}

// HasConfig handles GET /api/v1/config/:type/:key/exists
func (h *DataConfigHandler) HasConfig(c *fiber.Ctx) error {
    userID := c.Locals("userID").(string)
    configType := c.Params("type")
    configKey := c.Params("key")

    exists, err := h.repo.HasDataConfig(userID, configType, configKey)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to check configuration",
        })
    }

    return c.JSON(fiber.Map{
        "exists": exists,
    })
}
```

### 4. Register Routes

Add to route registration in `backend/internal/api/routes.go` (or equivalent):

```go
// Data Config routes (authenticated)
configHandler := handlers.NewDataConfigHandler(dataConfigRepo)
api.Post("/config/:type/:key", authMiddleware, configHandler.SetConfig)
api.Get("/config/:type/:key", authMiddleware, configHandler.GetConfig)
api.Delete("/config/:type/:key", authMiddleware, configHandler.DeleteConfig)
api.Get("/config/:type", authMiddleware, configHandler.ListConfigs)
api.Get("/config/:type/:key/exists", authMiddleware, configHandler.HasConfig)
```

### 5. Wire Up in Main

Update `backend/cmd/server/main.go` to initialize the repository:

```go
dataConfigRepo := repository.NewDataConfigRepository(db, encryptionService)
```

## Acceptance Criteria

- [ ] `data_configs` table created with proper schema
- [ ] `SetDataConfig()` encrypts and upserts configuration data
- [ ] `GetDataConfig()` decrypts and returns configuration data
- [ ] `DeleteDataConfig()` removes configuration entries
- [ ] `ListDataConfigs()` returns config keys without decrypting
- [ ] `HasDataConfig()` checks existence efficiently
- [ ] HTTP endpoints expose all operations with proper auth
- [ ] Unique constraint on (user_id, config_type, config_key)
- [ ] Index on (user_id, config_type) for efficient lookups
- [ ] All operations are user-scoped (users can only access their own configs)

## Files to Create/Modify

- `backend/internal/database/repository/data_config.go` - New repository
- `backend/internal/database/sqlite.go` - Add table creation
- `backend/internal/api/handlers/data_config.go` - New handler
- `backend/internal/api/routes.go` - Register routes
- `backend/cmd/server/main.go` - Wire up repository

## Integration Points

- **Provides**: Generic encrypted config storage API
- **Consumes**: Encryption service (`security.EncryptionService`)
- **Conflicts**: None - new independent module
