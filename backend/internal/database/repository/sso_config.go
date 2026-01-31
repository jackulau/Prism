package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/security"
)

// SSOConfigRepository handles SSO configuration database operations
type SSOConfigRepository struct {
	db            *sql.DB
	encryptionSvc *security.EncryptionService
}

// NewSSOConfigRepository creates a new SSO config repository
func NewSSOConfigRepository(db *sql.DB, encryptionSvc *security.EncryptionService) *SSOConfigRepository {
	return &SSOConfigRepository{
		db:            db,
		encryptionSvc: encryptionSvc,
	}
}

// Create creates a new SSO provider configuration
func (r *SSOConfigRepository) Create(ctx context.Context, config *security.SSOProviderConfig) error {
	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	if config.Status == "" {
		config.Status = security.SSOProviderStatusPending
	}

	// Serialize type-specific config
	configJSON, err := r.serializeConfig(config)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// Encrypt client secret if present
	var encryptedSecret, secretNonce []byte
	clientSecret := r.extractClientSecret(config)
	if clientSecret != "" {
		encryptedSecret, secretNonce, err = r.encryptionSvc.Encrypt([]byte(clientSecret))
		if err != nil {
			return fmt.Errorf("failed to encrypt client secret: %w", err)
		}
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO sso_configurations (
			id, organization_id, name, type, status, priority, enabled,
			config_json, encrypted_secret, secret_nonce,
			workos_connection_id, last_error,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		config.ID,
		config.OrganizationID,
		config.Name,
		string(config.Type),
		string(config.Status),
		config.Priority,
		config.Enabled,
		configJSON,
		encryptedSecret,
		secretNonce,
		config.WorkOSConnectionID,
		config.LastError,
		config.CreatedAt,
		config.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create SSO configuration: %w", err)
	}

	return nil
}

// GetByID retrieves an SSO provider configuration by ID
func (r *SSOConfigRepository) GetByID(ctx context.Context, id string) (*security.SSOProviderConfig, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, organization_id, name, type, status, priority, enabled,
		       config_json, encrypted_secret, secret_nonce,
		       workos_connection_id, last_error,
		       created_at, updated_at
		FROM sso_configurations WHERE id = ?
	`, id)

	return r.scanConfig(row)
}

// Update updates an SSO provider configuration
func (r *SSOConfigRepository) Update(ctx context.Context, config *security.SSOProviderConfig) error {
	config.UpdatedAt = time.Now()

	// Serialize type-specific config
	configJSON, err := r.serializeConfig(config)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// Encrypt client secret if present
	var encryptedSecret, secretNonce []byte
	clientSecret := r.extractClientSecret(config)
	if clientSecret != "" {
		encryptedSecret, secretNonce, err = r.encryptionSvc.Encrypt([]byte(clientSecret))
		if err != nil {
			return fmt.Errorf("failed to encrypt client secret: %w", err)
		}
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE sso_configurations SET
			name = ?, type = ?, status = ?, priority = ?, enabled = ?,
			config_json = ?, encrypted_secret = ?, secret_nonce = ?,
			workos_connection_id = ?, last_error = ?,
			updated_at = ?
		WHERE id = ?
	`,
		config.Name,
		string(config.Type),
		string(config.Status),
		config.Priority,
		config.Enabled,
		configJSON,
		encryptedSecret,
		secretNonce,
		config.WorkOSConnectionID,
		config.LastError,
		config.UpdatedAt,
		config.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update SSO configuration: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("SSO configuration not found")
	}

	return nil
}

// Delete deletes an SSO provider configuration
func (r *SSOConfigRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sso_configurations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete SSO configuration: %w", err)
	}
	return nil
}

// ListByOrganization lists all SSO configurations for an organization
func (r *SSOConfigRepository) ListByOrganization(ctx context.Context, orgID string) ([]*security.SSOProviderConfig, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, organization_id, name, type, status, priority, enabled,
		       config_json, encrypted_secret, secret_nonce,
		       workos_connection_id, last_error,
		       created_at, updated_at
		FROM sso_configurations
		WHERE organization_id = ?
		ORDER BY priority ASC, created_at ASC
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list SSO configurations: %w", err)
	}
	defer rows.Close()

	return r.scanConfigs(rows)
}

// GetByOrganizationAndType retrieves an SSO configuration by organization and type
func (r *SSOConfigRepository) GetByOrganizationAndType(ctx context.Context, orgID string, providerType security.SSOProviderType) (*security.SSOProviderConfig, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, organization_id, name, type, status, priority, enabled,
		       config_json, encrypted_secret, secret_nonce,
		       workos_connection_id, last_error,
		       created_at, updated_at
		FROM sso_configurations
		WHERE organization_id = ? AND type = ?
		LIMIT 1
	`, orgID, string(providerType))

	return r.scanConfig(row)
}

// GetActiveProviders retrieves all active SSO providers for an organization
func (r *SSOConfigRepository) GetActiveProviders(ctx context.Context, orgID string) ([]*security.SSOProviderConfig, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, organization_id, name, type, status, priority, enabled,
		       config_json, encrypted_secret, secret_nonce,
		       workos_connection_id, last_error,
		       created_at, updated_at
		FROM sso_configurations
		WHERE organization_id = ? AND enabled = 1 AND status = 'active'
		ORDER BY priority ASC
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list active SSO configurations: %w", err)
	}
	defer rows.Close()

	return r.scanConfigs(rows)
}

// SaveAttributeMappings saves attribute mappings for an SSO provider
func (r *SSOConfigRepository) SaveAttributeMappings(ctx context.Context, providerID string, mappings []security.AttributeMapping) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, m := range mappings {
		if m.ID == "" {
			m.ID = uuid.New().String()
		}
		m.SSOProviderID = providerID

		_, err := tx.ExecContext(ctx, `
			INSERT INTO sso_attribute_mappings (
				id, sso_provider_id, source_attribute, target_field,
				transform_type, transform_pattern
			) VALUES (?, ?, ?, ?, ?, ?)
		`,
			m.ID,
			m.SSOProviderID,
			m.SourceAttribute,
			m.TargetField,
			m.TransformType,
			m.TransformPattern,
		)
		if err != nil {
			return fmt.Errorf("failed to save attribute mapping: %w", err)
		}
	}

	return tx.Commit()
}

// GetAttributeMappings retrieves attribute mappings for an SSO provider
func (r *SSOConfigRepository) GetAttributeMappings(ctx context.Context, providerID string) ([]security.AttributeMapping, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, sso_provider_id, source_attribute, target_field,
		       transform_type, transform_pattern
		FROM sso_attribute_mappings
		WHERE sso_provider_id = ?
	`, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute mappings: %w", err)
	}
	defer rows.Close()

	var mappings []security.AttributeMapping
	for rows.Next() {
		var m security.AttributeMapping
		var transformType, transformPattern sql.NullString
		if err := rows.Scan(
			&m.ID,
			&m.SSOProviderID,
			&m.SourceAttribute,
			&m.TargetField,
			&transformType,
			&transformPattern,
		); err != nil {
			return nil, fmt.Errorf("failed to scan attribute mapping: %w", err)
		}
		m.TransformType = transformType.String
		m.TransformPattern = transformPattern.String
		mappings = append(mappings, m)
	}

	return mappings, nil
}

// DeleteAttributeMappings deletes all attribute mappings for an SSO provider
func (r *SSOConfigRepository) DeleteAttributeMappings(ctx context.Context, providerID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sso_attribute_mappings WHERE sso_provider_id = ?`, providerID)
	if err != nil {
		return fmt.Errorf("failed to delete attribute mappings: %w", err)
	}
	return nil
}

// scanConfig scans a single row into an SSOProviderConfig
func (r *SSOConfigRepository) scanConfig(row *sql.Row) (*security.SSOProviderConfig, error) {
	var config security.SSOProviderConfig
	var providerType, status string
	var configJSON []byte
	var encryptedSecret, secretNonce []byte
	var workosConnID, lastError sql.NullString

	err := row.Scan(
		&config.ID,
		&config.OrganizationID,
		&config.Name,
		&providerType,
		&status,
		&config.Priority,
		&config.Enabled,
		&configJSON,
		&encryptedSecret,
		&secretNonce,
		&workosConnID,
		&lastError,
		&config.CreatedAt,
		&config.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan SSO configuration: %w", err)
	}

	config.Type = security.SSOProviderType(providerType)
	config.Status = security.SSOProviderStatus(status)
	config.WorkOSConnectionID = workosConnID.String
	config.LastError = lastError.String

	// Deserialize config
	if err := r.deserializeConfig(&config, configJSON); err != nil {
		return nil, fmt.Errorf("failed to deserialize config: %w", err)
	}

	// Decrypt client secret if present
	if len(encryptedSecret) > 0 && len(secretNonce) > 0 {
		decrypted, err := r.encryptionSvc.Decrypt(encryptedSecret, secretNonce)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt client secret: %w", err)
		}
		r.setClientSecret(&config, string(decrypted))
	}

	return &config, nil
}

// scanConfigs scans multiple rows into SSOProviderConfigs
func (r *SSOConfigRepository) scanConfigs(rows *sql.Rows) ([]*security.SSOProviderConfig, error) {
	var configs []*security.SSOProviderConfig
	for rows.Next() {
		var config security.SSOProviderConfig
		var providerType, status string
		var configJSON []byte
		var encryptedSecret, secretNonce []byte
		var workosConnID, lastError sql.NullString

		err := rows.Scan(
			&config.ID,
			&config.OrganizationID,
			&config.Name,
			&providerType,
			&status,
			&config.Priority,
			&config.Enabled,
			&configJSON,
			&encryptedSecret,
			&secretNonce,
			&workosConnID,
			&lastError,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SSO configuration: %w", err)
		}

		config.Type = security.SSOProviderType(providerType)
		config.Status = security.SSOProviderStatus(status)
		config.WorkOSConnectionID = workosConnID.String
		config.LastError = lastError.String

		// Deserialize config
		if err := r.deserializeConfig(&config, configJSON); err != nil {
			return nil, fmt.Errorf("failed to deserialize config: %w", err)
		}

		// Decrypt client secret if present
		if len(encryptedSecret) > 0 && len(secretNonce) > 0 {
			decrypted, err := r.encryptionSvc.Decrypt(encryptedSecret, secretNonce)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt client secret: %w", err)
			}
			r.setClientSecret(&config, string(decrypted))
		}

		configs = append(configs, &config)
	}

	return configs, nil
}

// serializeConfig serializes the provider-specific configuration to JSON
func (r *SSOConfigRepository) serializeConfig(config *security.SSOProviderConfig) ([]byte, error) {
	data := make(map[string]interface{})

	switch config.Type {
	case security.SSOProviderTypeSAML:
		if config.SAMLConfig != nil {
			data["saml"] = config.SAMLConfig
		}
	case security.SSOProviderTypeOIDC:
		if config.OIDCConfig != nil {
			// Don't include client secret in JSON (stored separately encrypted)
			cfg := *config.OIDCConfig
			cfg.ClientSecret = ""
			data["oidc"] = cfg
		}
	case security.SSOProviderTypeOAuth:
		if config.OAuth2Config != nil {
			// Don't include client secret in JSON (stored separately encrypted)
			cfg := *config.OAuth2Config
			cfg.ClientSecret = ""
			data["oauth2"] = cfg
		}
	}

	return json.Marshal(data)
}

// deserializeConfig deserializes the provider-specific configuration from JSON
func (r *SSOConfigRepository) deserializeConfig(config *security.SSOProviderConfig, data []byte) error {
	if len(data) == 0 {
		return nil
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	switch config.Type {
	case security.SSOProviderTypeSAML:
		if samlData, ok := raw["saml"]; ok {
			config.SAMLConfig = &security.SAMLConfiguration{}
			if err := json.Unmarshal(samlData, config.SAMLConfig); err != nil {
				return err
			}
		}
	case security.SSOProviderTypeOIDC:
		if oidcData, ok := raw["oidc"]; ok {
			config.OIDCConfig = &security.OIDCConfiguration{}
			if err := json.Unmarshal(oidcData, config.OIDCConfig); err != nil {
				return err
			}
		}
	case security.SSOProviderTypeOAuth:
		if oauth2Data, ok := raw["oauth2"]; ok {
			config.OAuth2Config = &security.OAuth2Configuration{}
			if err := json.Unmarshal(oauth2Data, config.OAuth2Config); err != nil {
				return err
			}
		}
	}

	return nil
}

// extractClientSecret extracts the client secret from the config for encryption
func (r *SSOConfigRepository) extractClientSecret(config *security.SSOProviderConfig) string {
	switch config.Type {
	case security.SSOProviderTypeOIDC:
		if config.OIDCConfig != nil {
			return config.OIDCConfig.ClientSecret
		}
	case security.SSOProviderTypeOAuth:
		if config.OAuth2Config != nil {
			return config.OAuth2Config.ClientSecret
		}
	}
	return ""
}

// setClientSecret sets the client secret in the config after decryption
func (r *SSOConfigRepository) setClientSecret(config *security.SSOProviderConfig, secret string) {
	switch config.Type {
	case security.SSOProviderTypeOIDC:
		if config.OIDCConfig != nil {
			config.OIDCConfig.ClientSecret = secret
		}
	case security.SSOProviderTypeOAuth:
		if config.OAuth2Config != nil {
			config.OAuth2Config.ClientSecret = secret
		}
	}
}
