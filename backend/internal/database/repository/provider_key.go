package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ProviderKey represents an encrypted LLM provider API key
type ProviderKey struct {
	ID           string
	UserID       string
	Provider     string
	EncryptedKey []byte
	KeyNonce     []byte
	IsActive     bool
	CreatedAt    time.Time
}

// ProviderKeyRepository handles provider key database operations
type ProviderKeyRepository struct {
	db *sql.DB
}

// NewProviderKeyRepository creates a new provider key repository
func NewProviderKeyRepository(db *sql.DB) *ProviderKeyRepository {
	return &ProviderKeyRepository{db: db}
}

// SetKey stores or updates an encrypted API key for a provider
func (r *ProviderKeyRepository) SetKey(userID, provider string, encryptedKey, nonce []byte) error {
	id := uuid.New().String()
	now := time.Now()

	// Use UPSERT to insert or update
	_, err := r.db.Exec(`
		INSERT INTO provider_keys (id, user_id, provider, encrypted_key, key_nonce, is_active, created_at)
		VALUES (?, ?, ?, ?, ?, 1, ?)
		ON CONFLICT(user_id, provider) DO UPDATE SET
			encrypted_key = excluded.encrypted_key,
			key_nonce = excluded.key_nonce,
			is_active = 1
	`, id, userID, provider, encryptedKey, nonce, now)

	if err != nil {
		return fmt.Errorf("failed to set provider key: %w", err)
	}

	return nil
}

// GetKey retrieves an encrypted API key for a provider
func (r *ProviderKeyRepository) GetKey(userID, provider string) (*ProviderKey, error) {
	key := &ProviderKey{}

	err := r.db.QueryRow(`
		SELECT id, user_id, provider, encrypted_key, key_nonce, is_active, created_at
		FROM provider_keys
		WHERE user_id = ? AND provider = ? AND is_active = 1
	`, userID, provider).Scan(
		&key.ID,
		&key.UserID,
		&key.Provider,
		&key.EncryptedKey,
		&key.KeyNonce,
		&key.IsActive,
		&key.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get provider key: %w", err)
	}

	return key, nil
}

// DeleteKey removes an API key for a provider
func (r *ProviderKeyRepository) DeleteKey(userID, provider string) error {
	_, err := r.db.Exec(`
		DELETE FROM provider_keys
		WHERE user_id = ? AND provider = ?
	`, userID, provider)

	if err != nil {
		return fmt.Errorf("failed to delete provider key: %w", err)
	}

	return nil
}

// ListKeys retrieves all provider keys for a user (without the actual keys)
func (r *ProviderKeyRepository) ListKeys(userID string) ([]ProviderKey, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, provider, is_active, created_at
		FROM provider_keys
		WHERE user_id = ? AND is_active = 1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list provider keys: %w", err)
	}
	defer rows.Close()

	var keys []ProviderKey
	for rows.Next() {
		var key ProviderKey
		if err := rows.Scan(&key.ID, &key.UserID, &key.Provider, &key.IsActive, &key.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan provider key: %w", err)
		}
		keys = append(keys, key)
	}

	return keys, nil
}

// HasKey checks if a user has a key for a specific provider
func (r *ProviderKeyRepository) HasKey(userID, provider string) (bool, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM provider_keys
		WHERE user_id = ? AND provider = ? AND is_active = 1
	`, userID, provider).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check provider key: %w", err)
	}

	return count > 0, nil
}
