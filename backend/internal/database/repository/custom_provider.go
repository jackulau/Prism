package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/security"
)

// CustomProvider represents a user-configured OpenAI-compatible endpoint
type CustomProvider struct {
	ID            string
	UserID        string
	Name          string
	BaseURL       string
	EncryptedKey  []byte
	KeyNonce      []byte
	Models        string // JSON array of models
	SupportsTools bool
	SupportsVision bool
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// CustomProviderDTO is a data transfer object for custom providers (without encrypted data)
type CustomProviderDTO struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Name           string    `json:"name"`
	BaseURL        string    `json:"base_url"`
	HasAPIKey      bool      `json:"has_api_key"`
	Models         string    `json:"models"` // JSON array
	SupportsTools  bool      `json:"supports_tools"`
	SupportsVision bool      `json:"supports_vision"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CustomProviderRepository handles custom provider database operations
type CustomProviderRepository struct {
	db                *sql.DB
	encryptionService *security.EncryptionService
}

// NewCustomProviderRepository creates a new custom provider repository
func NewCustomProviderRepository(db *sql.DB, encryptionService *security.EncryptionService) *CustomProviderRepository {
	return &CustomProviderRepository{
		db:                db,
		encryptionService: encryptionService,
	}
}

// Create creates a new custom provider
func (r *CustomProviderRepository) Create(
	userID, name, baseURL, apiKey, models string,
	supportsTools, supportsVision bool,
) (*CustomProvider, error) {
	id := uuid.New().String()
	now := time.Now()

	var encryptedKey []byte
	var keyNonce []byte

	if apiKey != "" {
		var err error
		encryptedKey, keyNonce, err = r.encryptionService.Encrypt([]byte(apiKey))
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt API key: %w", err)
		}
	}

	_, err := r.db.Exec(`
		INSERT INTO custom_providers (
			id, user_id, name, base_url, encrypted_api_key, api_key_nonce,
			models, supports_tools, supports_vision, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)
	`, id, userID, name, baseURL, encryptedKey, keyNonce, models,
		supportsTools, supportsVision, now, now)

	if err != nil {
		return nil, fmt.Errorf("failed to create custom provider: %w", err)
	}

	return &CustomProvider{
		ID:             id,
		UserID:         userID,
		Name:           name,
		BaseURL:        baseURL,
		EncryptedKey:   encryptedKey,
		KeyNonce:       keyNonce,
		Models:         models,
		SupportsTools:  supportsTools,
		SupportsVision: supportsVision,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// GetByID retrieves a custom provider by ID
func (r *CustomProviderRepository) GetByID(id, userID string) (*CustomProvider, error) {
	provider := &CustomProvider{}

	err := r.db.QueryRow(`
		SELECT id, user_id, name, base_url, encrypted_api_key, api_key_nonce,
			   models, supports_tools, supports_vision, is_active, created_at, updated_at
		FROM custom_providers
		WHERE id = ? AND user_id = ?
	`, id, userID).Scan(
		&provider.ID,
		&provider.UserID,
		&provider.Name,
		&provider.BaseURL,
		&provider.EncryptedKey,
		&provider.KeyNonce,
		&provider.Models,
		&provider.SupportsTools,
		&provider.SupportsVision,
		&provider.IsActive,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get custom provider: %w", err)
	}

	return provider, nil
}

// GetByName retrieves a custom provider by name for a user
func (r *CustomProviderRepository) GetByName(userID, name string) (*CustomProvider, error) {
	provider := &CustomProvider{}

	err := r.db.QueryRow(`
		SELECT id, user_id, name, base_url, encrypted_api_key, api_key_nonce,
			   models, supports_tools, supports_vision, is_active, created_at, updated_at
		FROM custom_providers
		WHERE user_id = ? AND name = ?
	`, userID, name).Scan(
		&provider.ID,
		&provider.UserID,
		&provider.Name,
		&provider.BaseURL,
		&provider.EncryptedKey,
		&provider.KeyNonce,
		&provider.Models,
		&provider.SupportsTools,
		&provider.SupportsVision,
		&provider.IsActive,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get custom provider by name: %w", err)
	}

	return provider, nil
}

// List retrieves all custom providers for a user (without decrypted keys)
func (r *CustomProviderRepository) List(userID string) ([]CustomProviderDTO, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, name, base_url, encrypted_api_key IS NOT NULL,
			   models, supports_tools, supports_vision, is_active, created_at, updated_at
		FROM custom_providers
		WHERE user_id = ? AND is_active = 1
		ORDER BY name ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list custom providers: %w", err)
	}
	defer rows.Close()

	var providers []CustomProviderDTO
	for rows.Next() {
		var p CustomProviderDTO
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Name, &p.BaseURL, &p.HasAPIKey,
			&p.Models, &p.SupportsTools, &p.SupportsVision, &p.IsActive,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan custom provider: %w", err)
		}
		providers = append(providers, p)
	}

	return providers, nil
}

// ListAll retrieves all active custom providers for a user (with encrypted keys for internal use)
func (r *CustomProviderRepository) ListAll(userID string) ([]CustomProvider, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, name, base_url, encrypted_api_key, api_key_nonce,
			   models, supports_tools, supports_vision, is_active, created_at, updated_at
		FROM custom_providers
		WHERE user_id = ? AND is_active = 1
		ORDER BY name ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list all custom providers: %w", err)
	}
	defer rows.Close()

	var providers []CustomProvider
	for rows.Next() {
		var p CustomProvider
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Name, &p.BaseURL, &p.EncryptedKey, &p.KeyNonce,
			&p.Models, &p.SupportsTools, &p.SupportsVision, &p.IsActive,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan custom provider: %w", err)
		}
		providers = append(providers, p)
	}

	return providers, nil
}

// Update updates a custom provider
func (r *CustomProviderRepository) Update(
	id, userID string,
	name, baseURL, apiKey, models *string,
	supportsTools, supportsVision *bool,
) error {
	// Start building the update query dynamically
	query := "UPDATE custom_providers SET updated_at = ?"
	args := []interface{}{time.Now()}

	if name != nil {
		query += ", name = ?"
		args = append(args, *name)
	}
	if baseURL != nil {
		query += ", base_url = ?"
		args = append(args, *baseURL)
	}
	if apiKey != nil {
		if *apiKey == "" {
			// Clear the API key
			query += ", encrypted_api_key = NULL, api_key_nonce = NULL"
		} else {
			encryptedKey, nonce, err := r.encryptionService.Encrypt([]byte(*apiKey))
			if err != nil {
				return fmt.Errorf("failed to encrypt API key: %w", err)
			}
			query += ", encrypted_api_key = ?, api_key_nonce = ?"
			args = append(args, encryptedKey, nonce)
		}
	}
	if models != nil {
		query += ", models = ?"
		args = append(args, *models)
	}
	if supportsTools != nil {
		query += ", supports_tools = ?"
		args = append(args, *supportsTools)
	}
	if supportsVision != nil {
		query += ", supports_vision = ?"
		args = append(args, *supportsVision)
	}

	query += " WHERE id = ? AND user_id = ?"
	args = append(args, id, userID)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update custom provider: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("custom provider not found")
	}

	return nil
}

// Delete removes a custom provider (soft delete by setting is_active = 0)
func (r *CustomProviderRepository) Delete(id, userID string) error {
	result, err := r.db.Exec(`
		UPDATE custom_providers SET is_active = 0, updated_at = ?
		WHERE id = ? AND user_id = ?
	`, time.Now(), id, userID)

	if err != nil {
		return fmt.Errorf("failed to delete custom provider: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("custom provider not found")
	}

	return nil
}

// HardDelete permanently removes a custom provider
func (r *CustomProviderRepository) HardDelete(id, userID string) error {
	result, err := r.db.Exec(`
		DELETE FROM custom_providers
		WHERE id = ? AND user_id = ?
	`, id, userID)

	if err != nil {
		return fmt.Errorf("failed to hard delete custom provider: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("custom provider not found")
	}

	return nil
}

// DecryptAPIKey decrypts the API key for a custom provider
func (r *CustomProviderRepository) DecryptAPIKey(provider *CustomProvider) (string, error) {
	if provider.EncryptedKey == nil || len(provider.EncryptedKey) == 0 {
		return "", nil
	}

	decrypted, err := r.encryptionService.Decrypt(provider.EncryptedKey, provider.KeyNonce)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt API key: %w", err)
	}

	return string(decrypted), nil
}

// Exists checks if a custom provider with the given name exists for a user
func (r *CustomProviderRepository) Exists(userID, name string) (bool, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM custom_providers
		WHERE user_id = ? AND name = ? AND is_active = 1
	`, userID, name).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check custom provider existence: %w", err)
	}

	return count > 0, nil
}

// CountByUser returns the number of custom providers for a user
func (r *CustomProviderRepository) CountByUser(userID string) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM custom_providers
		WHERE user_id = ? AND is_active = 1
	`, userID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count custom providers: %w", err)
	}

	return count, nil
}
