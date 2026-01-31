package repository

import (
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/security"
)

// BuildConfig represents a build configuration
type BuildConfig struct {
	ID             string
	WorkspaceID    *string
	OrgWorkspaceID *string
	UserID         string
	Name           string
	Description    *string
	IsDefault      bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Commands       []BuildCommand
	EnvVars        []BuildEnvVar
}

// BuildCommand represents a build command within a config
type BuildCommand struct {
	ID               string
	ConfigID         string
	Name             string
	Command          string
	WorkingDirectory *string
	RunOrder         int
	IsEnabled        bool
	CreatedAt        time.Time
}

// BuildEnvVar represents an environment variable with encrypted value
type BuildEnvVar struct {
	ID       string
	ConfigID string
	Key      string
	Value    string // Decrypted value (not stored directly)
	IsSecret bool
	CreatedAt time.Time
}

// buildEnvVarEncrypted is the internal representation with encrypted data
type buildEnvVarEncrypted struct {
	ID             string
	ConfigID       string
	Key            string
	ValueEncrypted []byte
	ValueNonce     []byte
	IsSecret       bool
	CreatedAt      time.Time
}

// BuildConfigRepository handles build configuration database operations
type BuildConfigRepository struct {
	db     *sql.DB
	crypto *security.EncryptionService
}

// NewBuildConfigRepository creates a new build config repository
func NewBuildConfigRepository(db *sql.DB, crypto *security.EncryptionService) *BuildConfigRepository {
	return &BuildConfigRepository{db: db, crypto: crypto}
}

// envVarKeyRegex validates environment variable keys (alphanumeric + underscore)
var envVarKeyRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ValidateEnvVarKey validates an environment variable key
func ValidateEnvVarKey(key string) error {
	if key == "" {
		return fmt.Errorf("environment variable key cannot be empty")
	}
	if len(key) > 255 {
		return fmt.Errorf("environment variable key too long (max 255 characters)")
	}
	if !envVarKeyRegex.MatchString(key) {
		return fmt.Errorf("environment variable key must start with a letter or underscore and contain only alphanumeric characters and underscores")
	}
	return nil
}

// ValidateCommand validates a build command string
func ValidateCommand(command string) error {
	if command == "" {
		return fmt.Errorf("command cannot be empty")
	}
	if len(command) > 4096 {
		return fmt.Errorf("command too long (max 4096 characters)")
	}
	return nil
}

// Create creates a new build configuration
func (r *BuildConfigRepository) Create(config *BuildConfig) error {
	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	_, err := r.db.Exec(`
		INSERT INTO build_configs (id, workspace_id, org_workspace_id, user_id, name, description, is_default, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, config.ID, config.WorkspaceID, config.OrgWorkspaceID, config.UserID, config.Name, config.Description, config.IsDefault, config.CreatedAt, config.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create build config: %w", err)
	}

	return nil
}

// GetByID retrieves a build configuration by ID (without commands and env vars)
func (r *BuildConfigRepository) GetByID(id string) (*BuildConfig, error) {
	config := &BuildConfig{}

	err := r.db.QueryRow(`
		SELECT id, workspace_id, org_workspace_id, user_id, name, description, is_default, created_at, updated_at
		FROM build_configs
		WHERE id = ?
	`, id).Scan(
		&config.ID,
		&config.WorkspaceID,
		&config.OrgWorkspaceID,
		&config.UserID,
		&config.Name,
		&config.Description,
		&config.IsDefault,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get build config: %w", err)
	}

	return config, nil
}

// GetByIDWithDetails retrieves a build configuration with commands and env vars
func (r *BuildConfigRepository) GetByIDWithDetails(id string) (*BuildConfig, error) {
	config, err := r.GetByID(id)
	if err != nil || config == nil {
		return config, err
	}

	// Load commands
	commands, err := r.getCommands(id)
	if err != nil {
		return nil, err
	}
	config.Commands = commands

	// Load env vars (decrypted)
	envVars, err := r.GetEnvVars(id)
	if err != nil {
		return nil, err
	}
	config.EnvVars = envVars

	return config, nil
}

// Update updates a build configuration
func (r *BuildConfigRepository) Update(config *BuildConfig) error {
	config.UpdatedAt = time.Now()

	result, err := r.db.Exec(`
		UPDATE build_configs
		SET name = ?, description = ?, is_default = ?, updated_at = ?
		WHERE id = ?
	`, config.Name, config.Description, config.IsDefault, config.UpdatedAt, config.ID)

	if err != nil {
		return fmt.Errorf("failed to update build config: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("build config not found")
	}

	return nil
}

// Delete deletes a build configuration and all associated commands/env vars
func (r *BuildConfigRepository) Delete(id string) error {
	result, err := r.db.Exec(`DELETE FROM build_configs WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete build config: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("build config not found")
	}

	return nil
}

// ListByWorkspaceID lists all build configurations for a workspace
func (r *BuildConfigRepository) ListByWorkspaceID(workspaceID string) ([]*BuildConfig, error) {
	rows, err := r.db.Query(`
		SELECT id, workspace_id, org_workspace_id, user_id, name, description, is_default, created_at, updated_at
		FROM build_configs
		WHERE workspace_id = ?
		ORDER BY is_default DESC, name ASC
	`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list build configs: %w", err)
	}
	defer rows.Close()

	return r.scanConfigs(rows)
}

// ListByOrgWorkspaceID lists all build configurations for an org workspace
func (r *BuildConfigRepository) ListByOrgWorkspaceID(orgWorkspaceID string) ([]*BuildConfig, error) {
	rows, err := r.db.Query(`
		SELECT id, workspace_id, org_workspace_id, user_id, name, description, is_default, created_at, updated_at
		FROM build_configs
		WHERE org_workspace_id = ?
		ORDER BY is_default DESC, name ASC
	`, orgWorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list build configs: %w", err)
	}
	defer rows.Close()

	return r.scanConfigs(rows)
}

// ListByUserID lists all build configurations for a user
func (r *BuildConfigRepository) ListByUserID(userID string) ([]*BuildConfig, error) {
	rows, err := r.db.Query(`
		SELECT id, workspace_id, org_workspace_id, user_id, name, description, is_default, created_at, updated_at
		FROM build_configs
		WHERE user_id = ?
		ORDER BY is_default DESC, name ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list build configs: %w", err)
	}
	defer rows.Close()

	return r.scanConfigs(rows)
}

// GetDefault retrieves the default build configuration for a workspace
func (r *BuildConfigRepository) GetDefault(workspaceID string) (*BuildConfig, error) {
	config := &BuildConfig{}

	err := r.db.QueryRow(`
		SELECT id, workspace_id, org_workspace_id, user_id, name, description, is_default, created_at, updated_at
		FROM build_configs
		WHERE workspace_id = ? AND is_default = 1
	`, workspaceID).Scan(
		&config.ID,
		&config.WorkspaceID,
		&config.OrgWorkspaceID,
		&config.UserID,
		&config.Name,
		&config.Description,
		&config.IsDefault,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get default build config: %w", err)
	}

	return config, nil
}

// SetDefault sets a build configuration as the default for its workspace
func (r *BuildConfigRepository) SetDefault(id string, workspaceID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing default
	_, err = tx.Exec(`
		UPDATE build_configs
		SET is_default = 0
		WHERE workspace_id = ? AND is_default = 1
	`, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to clear default: %w", err)
	}

	// Set new default
	result, err := tx.Exec(`
		UPDATE build_configs
		SET is_default = 1, updated_at = ?
		WHERE id = ? AND workspace_id = ?
	`, time.Now(), id, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to set default: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("build config not found or does not belong to workspace")
	}

	return tx.Commit()
}

// scanConfigs scans multiple config rows
func (r *BuildConfigRepository) scanConfigs(rows *sql.Rows) ([]*BuildConfig, error) {
	var configs []*BuildConfig
	for rows.Next() {
		config := &BuildConfig{}
		if err := rows.Scan(
			&config.ID,
			&config.WorkspaceID,
			&config.OrgWorkspaceID,
			&config.UserID,
			&config.Name,
			&config.Description,
			&config.IsDefault,
			&config.CreatedAt,
			&config.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan build config: %w", err)
		}
		configs = append(configs, config)
	}
	return configs, nil
}

// ==================== Command Operations ====================

// AddCommand adds a command to a build configuration
func (r *BuildConfigRepository) AddCommand(cmd *BuildCommand) error {
	if err := ValidateCommand(cmd.Command); err != nil {
		return err
	}

	if cmd.ID == "" {
		cmd.ID = uuid.New().String()
	}
	cmd.CreatedAt = time.Now()

	_, err := r.db.Exec(`
		INSERT INTO build_commands (id, config_id, name, command, working_directory, run_order, is_enabled, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, cmd.ID, cmd.ConfigID, cmd.Name, cmd.Command, cmd.WorkingDirectory, cmd.RunOrder, cmd.IsEnabled, cmd.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to add build command: %w", err)
	}

	return nil
}

// UpdateCommand updates a build command
func (r *BuildConfigRepository) UpdateCommand(cmd *BuildCommand) error {
	if err := ValidateCommand(cmd.Command); err != nil {
		return err
	}

	result, err := r.db.Exec(`
		UPDATE build_commands
		SET name = ?, command = ?, working_directory = ?, run_order = ?, is_enabled = ?
		WHERE id = ?
	`, cmd.Name, cmd.Command, cmd.WorkingDirectory, cmd.RunOrder, cmd.IsEnabled, cmd.ID)

	if err != nil {
		return fmt.Errorf("failed to update build command: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("build command not found")
	}

	return nil
}

// DeleteCommand deletes a build command
func (r *BuildConfigRepository) DeleteCommand(id string) error {
	result, err := r.db.Exec(`DELETE FROM build_commands WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete build command: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("build command not found")
	}

	return nil
}

// GetCommand retrieves a single command by ID
func (r *BuildConfigRepository) GetCommand(id string) (*BuildCommand, error) {
	cmd := &BuildCommand{}

	err := r.db.QueryRow(`
		SELECT id, config_id, name, command, working_directory, run_order, is_enabled, created_at
		FROM build_commands
		WHERE id = ?
	`, id).Scan(
		&cmd.ID,
		&cmd.ConfigID,
		&cmd.Name,
		&cmd.Command,
		&cmd.WorkingDirectory,
		&cmd.RunOrder,
		&cmd.IsEnabled,
		&cmd.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get build command: %w", err)
	}

	return cmd, nil
}

// ReorderCommands reorders commands within a build configuration
func (r *BuildConfigRepository) ReorderCommands(configID string, order []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for i, cmdID := range order {
		_, err := tx.Exec(`
			UPDATE build_commands
			SET run_order = ?
			WHERE id = ? AND config_id = ?
		`, i, cmdID, configID)
		if err != nil {
			return fmt.Errorf("failed to update command order: %w", err)
		}
	}

	return tx.Commit()
}

// getCommands retrieves all commands for a config
func (r *BuildConfigRepository) getCommands(configID string) ([]BuildCommand, error) {
	rows, err := r.db.Query(`
		SELECT id, config_id, name, command, working_directory, run_order, is_enabled, created_at
		FROM build_commands
		WHERE config_id = ?
		ORDER BY run_order ASC
	`, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to get build commands: %w", err)
	}
	defer rows.Close()

	var commands []BuildCommand
	for rows.Next() {
		cmd := BuildCommand{}
		if err := rows.Scan(
			&cmd.ID,
			&cmd.ConfigID,
			&cmd.Name,
			&cmd.Command,
			&cmd.WorkingDirectory,
			&cmd.RunOrder,
			&cmd.IsEnabled,
			&cmd.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan build command: %w", err)
		}
		commands = append(commands, cmd)
	}

	return commands, nil
}

// ==================== Environment Variable Operations ====================

// SetEnvVar sets or updates an environment variable (encrypted)
func (r *BuildConfigRepository) SetEnvVar(envVar *BuildEnvVar) error {
	if err := ValidateEnvVarKey(envVar.Key); err != nil {
		return err
	}

	// Encrypt the value
	encryptedValue, nonce, err := r.crypto.Encrypt([]byte(envVar.Value))
	if err != nil {
		return fmt.Errorf("failed to encrypt env var value: %w", err)
	}

	id := uuid.New().String()
	now := time.Now()

	// Use UPSERT to insert or update
	_, err = r.db.Exec(`
		INSERT INTO build_env_vars (id, config_id, key, value_encrypted, value_nonce, is_secret, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(config_id, key) DO UPDATE SET
			value_encrypted = excluded.value_encrypted,
			value_nonce = excluded.value_nonce,
			is_secret = excluded.is_secret
	`, id, envVar.ConfigID, envVar.Key, encryptedValue, nonce, envVar.IsSecret, now)

	if err != nil {
		return fmt.Errorf("failed to set env var: %w", err)
	}

	return nil
}

// DeleteEnvVar deletes an environment variable
func (r *BuildConfigRepository) DeleteEnvVar(id string) error {
	result, err := r.db.Exec(`DELETE FROM build_env_vars WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete env var: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("env var not found")
	}

	return nil
}

// DeleteEnvVarByKey deletes an environment variable by config ID and key
func (r *BuildConfigRepository) DeleteEnvVarByKey(configID, key string) error {
	result, err := r.db.Exec(`DELETE FROM build_env_vars WHERE config_id = ? AND key = ?`, configID, key)
	if err != nil {
		return fmt.Errorf("failed to delete env var: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("env var not found")
	}

	return nil
}

// GetEnvVars retrieves all environment variables for a config (decrypted)
func (r *BuildConfigRepository) GetEnvVars(configID string) ([]BuildEnvVar, error) {
	rows, err := r.db.Query(`
		SELECT id, config_id, key, value_encrypted, value_nonce, is_secret, created_at
		FROM build_env_vars
		WHERE config_id = ?
		ORDER BY key ASC
	`, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to get env vars: %w", err)
	}
	defer rows.Close()

	var envVars []BuildEnvVar
	for rows.Next() {
		var encrypted buildEnvVarEncrypted
		if err := rows.Scan(
			&encrypted.ID,
			&encrypted.ConfigID,
			&encrypted.Key,
			&encrypted.ValueEncrypted,
			&encrypted.ValueNonce,
			&encrypted.IsSecret,
			&encrypted.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan env var: %w", err)
		}

		// Decrypt the value
		decryptedValue, err := r.crypto.Decrypt(encrypted.ValueEncrypted, encrypted.ValueNonce)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt env var value: %w", err)
		}

		envVars = append(envVars, BuildEnvVar{
			ID:        encrypted.ID,
			ConfigID:  encrypted.ConfigID,
			Key:       encrypted.Key,
			Value:     string(decryptedValue),
			IsSecret:  encrypted.IsSecret,
			CreatedAt: encrypted.CreatedAt,
		})
	}

	return envVars, nil
}

// GetEnvVarsMasked retrieves env vars with secret values masked
func (r *BuildConfigRepository) GetEnvVarsMasked(configID string) ([]BuildEnvVar, error) {
	envVars, err := r.GetEnvVars(configID)
	if err != nil {
		return nil, err
	}

	// Mask secret values
	for i := range envVars {
		if envVars[i].IsSecret {
			envVars[i].Value = "********"
		}
	}

	return envVars, nil
}
