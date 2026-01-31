package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// UserAPIKey represents a user's API key for external access
type UserAPIKey struct {
	ID         string
	UserID     string
	Name       string
	KeyHash    string
	KeyPrefix  string
	LastUsedAt *time.Time
	ExpiresAt  *time.Time
	CreatedAt  time.Time
}

// UserAPIKeyRepository handles user API key database operations
type UserAPIKeyRepository struct {
	db *sql.DB
}

// NewUserAPIKeyRepository creates a new user API key repository
func NewUserAPIKeyRepository(db *sql.DB) *UserAPIKeyRepository {
	return &UserAPIKeyRepository{db: db}
}

// Create creates a new API key with optional expiration and scopes
func (r *UserAPIKeyRepository) Create(userID, name, keyHash, prefix string, expiresAt *time.Time, scopes []string) (*UserAPIKey, error) {
	id := uuid.New().String()
	now := time.Now()

	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO user_api_keys (id, user_id, name, key_hash, key_prefix, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, userID, name, keyHash, prefix, expiresAt, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	// Insert scopes if provided
	if len(scopes) > 0 {
		for _, scope := range scopes {
			_, err = tx.Exec(`
				INSERT INTO api_key_scopes (api_key_id, scope)
				VALUES (?, ?)
			`, id, scope)
			if err != nil {
				return nil, fmt.Errorf("failed to create API key scope: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &UserAPIKey{
		ID:        id,
		UserID:    userID,
		Name:      name,
		KeyHash:   keyHash,
		KeyPrefix: prefix,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}, nil
}

// GetByUserID retrieves all API keys for a user
func (r *UserAPIKeyRepository) GetByUserID(userID string) ([]UserAPIKey, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, name, key_hash, key_prefix, last_used_at, expires_at, created_at
		FROM user_api_keys
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get API keys: %w", err)
	}
	defer rows.Close()

	var keys []UserAPIKey
	for rows.Next() {
		var key UserAPIKey
		var lastUsedAt, expiresAt sql.NullTime
		if err := rows.Scan(&key.ID, &key.UserID, &key.Name, &key.KeyHash, &key.KeyPrefix, &lastUsedAt, &expiresAt, &key.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		if lastUsedAt.Valid {
			key.LastUsedAt = &lastUsedAt.Time
		}
		if expiresAt.Valid {
			key.ExpiresAt = &expiresAt.Time
		}
		keys = append(keys, key)
	}

	return keys, nil
}

// GetByID retrieves an API key by its ID
func (r *UserAPIKeyRepository) GetByID(id string) (*UserAPIKey, error) {
	key := &UserAPIKey{}
	var lastUsedAt, expiresAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, user_id, name, key_hash, key_prefix, last_used_at, expires_at, created_at
		FROM user_api_keys
		WHERE id = ?
	`, id).Scan(&key.ID, &key.UserID, &key.Name, &key.KeyHash, &key.KeyPrefix, &lastUsedAt, &expiresAt, &key.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	if lastUsedAt.Valid {
		key.LastUsedAt = &lastUsedAt.Time
	}
	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Time
	}

	return key, nil
}

// GetByKeyHash retrieves an API key by its hash
func (r *UserAPIKeyRepository) GetByKeyHash(keyHash string) (*UserAPIKey, error) {
	key := &UserAPIKey{}
	var lastUsedAt, expiresAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, user_id, name, key_hash, key_prefix, last_used_at, expires_at, created_at
		FROM user_api_keys
		WHERE key_hash = ?
	`, keyHash).Scan(&key.ID, &key.UserID, &key.Name, &key.KeyHash, &key.KeyPrefix, &lastUsedAt, &expiresAt, &key.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	if lastUsedAt.Valid {
		key.LastUsedAt = &lastUsedAt.Time
	}
	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Time
	}

	return key, nil
}

// UpdateLastUsed updates the last used timestamp for an API key
func (r *UserAPIKeyRepository) UpdateLastUsed(id string) error {
	_, err := r.db.Exec(`
		UPDATE user_api_keys
		SET last_used_at = ?
		WHERE id = ?
	`, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}
	return nil
}

// Delete removes an API key
func (r *UserAPIKeyRepository) Delete(id string) error {
	result, err := r.db.Exec(`
		DELETE FROM user_api_keys
		WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("API key not found")
	}

	return nil
}

// DeleteExpired removes all expired API keys and returns the count
func (r *UserAPIKeyRepository) DeleteExpired() (int64, error) {
	result, err := r.db.Exec(`
		DELETE FROM user_api_keys
		WHERE expires_at IS NOT NULL AND expires_at < ?
	`, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired keys: %w", err)
	}

	count, _ := result.RowsAffected()
	return count, nil
}

// UpdateName updates the name of an API key
func (r *UserAPIKeyRepository) UpdateName(id, name string) error {
	result, err := r.db.Exec(`
		UPDATE user_api_keys
		SET name = ?
		WHERE id = ?
	`, name, id)
	if err != nil {
		return fmt.Errorf("failed to update API key name: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("API key not found")
	}

	return nil
}

// GetScopes retrieves all scopes for an API key
func (r *UserAPIKeyRepository) GetScopes(keyID string) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT scope
		FROM api_key_scopes
		WHERE api_key_id = ?
	`, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key scopes: %w", err)
	}
	defer rows.Close()

	var scopes []string
	for rows.Next() {
		var scope string
		if err := rows.Scan(&scope); err != nil {
			return nil, fmt.Errorf("failed to scan scope: %w", err)
		}
		scopes = append(scopes, scope)
	}

	return scopes, nil
}

// IsExpired checks if an API key is expired
func (k *UserAPIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}
