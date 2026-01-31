package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// ApprovalConfig represents user-specific tool approval configuration
type ApprovalConfig struct {
	UserID              string   `json:"-"`
	Enabled             bool     `json:"enabled"`
	AutoApproveReadOnly bool     `json:"auto_approve_read_only"`
	TrustedTools        []string `json:"trusted_tools"`
	MaxIterations       int      `json:"max_iterations"`
	UpdatedAt           time.Time `json:"-"`
}

// DefaultApprovalConfig returns the default approval configuration
func DefaultApprovalConfig(userID string) *ApprovalConfig {
	return &ApprovalConfig{
		UserID:              userID,
		Enabled:             false,
		AutoApproveReadOnly: false,
		TrustedTools:        []string{},
		MaxIterations:       10,
		UpdatedAt:           time.Now(),
	}
}

// ApprovalConfigRepository handles approval config database operations
type ApprovalConfigRepository struct {
	db *sql.DB
}

// NewApprovalConfigRepository creates a new approval config repository
func NewApprovalConfigRepository(db *sql.DB) *ApprovalConfigRepository {
	return &ApprovalConfigRepository{db: db}
}

// Get retrieves the approval config for a user
func (r *ApprovalConfigRepository) Get(userID string) (*ApprovalConfig, error) {
	var settingsJSON sql.NullString
	var updatedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT settings_json, updated_at
		FROM user_settings
		WHERE user_id = ?
	`, userID).Scan(&settingsJSON, &updatedAt)

	if err == sql.ErrNoRows {
		return DefaultApprovalConfig(userID), nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get approval config: %w", err)
	}

	config := DefaultApprovalConfig(userID)
	if settingsJSON.Valid && settingsJSON.String != "" {
		// Parse existing settings
		var settings map[string]interface{}
		if err := json.Unmarshal([]byte(settingsJSON.String), &settings); err == nil {
			// Extract approval_config from settings
			if approvalData, ok := settings["approval_config"]; ok {
				if approvalBytes, err := json.Marshal(approvalData); err == nil {
					json.Unmarshal(approvalBytes, config)
				}
			}
		}
	}

	config.UserID = userID
	if updatedAt.Valid {
		config.UpdatedAt = updatedAt.Time
	}

	return config, nil
}

// Save stores the approval config for a user
func (r *ApprovalConfigRepository) Save(config *ApprovalConfig) error {
	// First, get existing settings
	var existingJSON sql.NullString
	err := r.db.QueryRow(`SELECT settings_json FROM user_settings WHERE user_id = ?`, config.UserID).Scan(&existingJSON)

	var settings map[string]interface{}
	if err == sql.ErrNoRows {
		settings = make(map[string]interface{})
	} else if err != nil {
		return fmt.Errorf("failed to get existing settings: %w", err)
	} else if existingJSON.Valid && existingJSON.String != "" {
		if err := json.Unmarshal([]byte(existingJSON.String), &settings); err != nil {
			settings = make(map[string]interface{})
		}
	} else {
		settings = make(map[string]interface{})
	}

	// Update approval_config in settings
	settings["approval_config"] = map[string]interface{}{
		"enabled":               config.Enabled,
		"auto_approve_read_only": config.AutoApproveReadOnly,
		"trusted_tools":         config.TrustedTools,
		"max_iterations":        config.MaxIterations,
	}

	settingsBytes, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	now := time.Now()
	_, err = r.db.Exec(`
		INSERT INTO user_settings (user_id, settings_json, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			settings_json = excluded.settings_json,
			updated_at = excluded.updated_at
	`, config.UserID, string(settingsBytes), now)

	if err != nil {
		return fmt.Errorf("failed to save approval config: %w", err)
	}

	config.UpdatedAt = now
	return nil
}

// AddTrustedTool adds a tool to the trusted tools list
func (r *ApprovalConfigRepository) AddTrustedTool(userID, toolName string) error {
	config, err := r.Get(userID)
	if err != nil {
		return err
	}

	// Check if already trusted
	for _, t := range config.TrustedTools {
		if t == toolName {
			return nil // Already trusted
		}
	}

	config.TrustedTools = append(config.TrustedTools, toolName)
	return r.Save(config)
}

// RemoveTrustedTool removes a tool from the trusted tools list
func (r *ApprovalConfigRepository) RemoveTrustedTool(userID, toolName string) error {
	config, err := r.Get(userID)
	if err != nil {
		return err
	}

	// Remove from list
	newTrusted := make([]string, 0, len(config.TrustedTools))
	for _, t := range config.TrustedTools {
		if t != toolName {
			newTrusted = append(newTrusted, t)
		}
	}

	config.TrustedTools = newTrusted
	return r.Save(config)
}
