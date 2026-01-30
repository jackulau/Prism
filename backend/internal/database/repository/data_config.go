package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/security"
)

// DataConfig represents an encrypted configuration entry
type DataConfig struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"user_id"`
	ConfigType string                 `json:"config_type"`
	ConfigKey  string                 `json:"config_key"`
	Data       map[string]interface{} `json:"data"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// DataConfigRepository handles encrypted configuration storage
type DataConfigRepository struct {
	db     *sql.DB
	crypto *security.EncryptionService
}

// NewDataConfigRepository creates a new data config repository
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

	id := uuid.New().String()
	now := time.Now()

	// Upsert: insert or update on conflict
	_, err = r.db.Exec(`
		INSERT INTO data_configs (id, user_id, config_type, config_key, encrypted_data, data_nonce, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, config_type, config_key) DO UPDATE SET
			encrypted_data = excluded.encrypted_data,
			data_nonce = excluded.data_nonce,
			updated_at = excluded.updated_at
	`, id, userID, configType, configKey, encryptedData, nonce, now, now)

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

// ListConfigTypes returns all distinct config types for a user
func (r *DataConfigRepository) ListConfigTypes(userID string) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT DISTINCT config_type
		FROM data_configs
		WHERE user_id = ?
		ORDER BY config_type
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list config types: %w", err)
	}
	defer rows.Close()

	var types []string
	for rows.Next() {
		var configType string
		if err := rows.Scan(&configType); err != nil {
			return nil, fmt.Errorf("failed to scan config type: %w", err)
		}
		types = append(types, configType)
	}

	return types, nil
}
