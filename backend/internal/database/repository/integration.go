package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jacklau/prism/internal/security"
)

// IntegrationSettings represents user integration settings
type IntegrationSettings struct {
	UserID     string
	Type       string // discord, slack, posthog
	Enabled    bool
	WebhookURL string // decrypted, only populated on read
	BotToken   string // decrypted, only populated on read
	ChannelID  string // for slack
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// IntegrationRepository handles integration settings database operations
type IntegrationRepository struct {
	db                *sql.DB
	encryptionService *security.EncryptionService
}

// NewIntegrationRepository creates a new integration repository
func NewIntegrationRepository(db *sql.DB, encryptionService *security.EncryptionService) *IntegrationRepository {
	return &IntegrationRepository{
		db:                db,
		encryptionService: encryptionService,
	}
}

// SetDiscordSettings stores or updates Discord integration settings
func (r *IntegrationRepository) SetDiscordSettings(userID string, webhookURL string, enabled bool) error {
	// Encrypt webhook URL if provided
	var webhookEncrypted, webhookNonce []byte
	var err error
	if webhookURL != "" {
		webhookEncrypted, webhookNonce, err = r.encryptionService.Encrypt([]byte(webhookURL))
		if err != nil {
			return fmt.Errorf("failed to encrypt webhook URL: %w", err)
		}
	}

	now := time.Now()
	_, err = r.db.Exec(`
		INSERT INTO discord_settings (user_id, webhook_url_encrypted, webhook_url_nonce, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			webhook_url_encrypted = excluded.webhook_url_encrypted,
			webhook_url_nonce = excluded.webhook_url_nonce,
			enabled = excluded.enabled,
			updated_at = excluded.updated_at
	`, userID, webhookEncrypted, webhookNonce, enabled, now, now)

	if err != nil {
		return fmt.Errorf("failed to set discord settings: %w", err)
	}
	return nil
}

// GetDiscordSettings retrieves Discord settings for a user
func (r *IntegrationRepository) GetDiscordSettings(userID string) (*IntegrationSettings, error) {
	var webhookEncrypted, webhookNonce []byte
	var enabled bool
	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(`
		SELECT webhook_url_encrypted, webhook_url_nonce, enabled, created_at, updated_at
		FROM discord_settings
		WHERE user_id = ?
	`, userID).Scan(&webhookEncrypted, &webhookNonce, &enabled, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get discord settings: %w", err)
	}

	settings := &IntegrationSettings{
		UserID:    userID,
		Type:      "discord",
		Enabled:   enabled,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	// Decrypt webhook URL
	if len(webhookEncrypted) > 0 && len(webhookNonce) > 0 {
		decrypted, err := r.encryptionService.Decrypt(webhookEncrypted, webhookNonce)
		if err == nil {
			settings.WebhookURL = string(decrypted)
		}
	}

	return settings, nil
}

// DeleteDiscordSettings removes Discord settings for a user
func (r *IntegrationRepository) DeleteDiscordSettings(userID string) error {
	_, err := r.db.Exec(`DELETE FROM discord_settings WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete discord settings: %w", err)
	}
	return nil
}

// SetSlackSettings stores or updates Slack integration settings
func (r *IntegrationRepository) SetSlackSettings(userID string, webhookURL string, channelID string, enabled bool) error {
	// Encrypt webhook URL if provided
	var webhookEncrypted, webhookNonce []byte
	var err error
	if webhookURL != "" {
		webhookEncrypted, webhookNonce, err = r.encryptionService.Encrypt([]byte(webhookURL))
		if err != nil {
			return fmt.Errorf("failed to encrypt webhook URL: %w", err)
		}
	}

	now := time.Now()
	_, err = r.db.Exec(`
		INSERT INTO slack_settings (user_id, webhook_url_encrypted, webhook_url_nonce, channel_id, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			webhook_url_encrypted = excluded.webhook_url_encrypted,
			webhook_url_nonce = excluded.webhook_url_nonce,
			channel_id = excluded.channel_id,
			enabled = excluded.enabled,
			updated_at = excluded.updated_at
	`, userID, webhookEncrypted, webhookNonce, channelID, enabled, now, now)

	if err != nil {
		return fmt.Errorf("failed to set slack settings: %w", err)
	}
	return nil
}

// GetSlackSettings retrieves Slack settings for a user
func (r *IntegrationRepository) GetSlackSettings(userID string) (*IntegrationSettings, error) {
	var webhookEncrypted, webhookNonce []byte
	var channelID sql.NullString
	var enabled bool
	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(`
		SELECT webhook_url_encrypted, webhook_url_nonce, channel_id, enabled, created_at, updated_at
		FROM slack_settings
		WHERE user_id = ?
	`, userID).Scan(&webhookEncrypted, &webhookNonce, &channelID, &enabled, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get slack settings: %w", err)
	}

	settings := &IntegrationSettings{
		UserID:    userID,
		Type:      "slack",
		Enabled:   enabled,
		ChannelID: channelID.String,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	// Decrypt webhook URL
	if len(webhookEncrypted) > 0 && len(webhookNonce) > 0 {
		decrypted, err := r.encryptionService.Decrypt(webhookEncrypted, webhookNonce)
		if err == nil {
			settings.WebhookURL = string(decrypted)
		}
	}

	return settings, nil
}

// DeleteSlackSettings removes Slack settings for a user
func (r *IntegrationRepository) DeleteSlackSettings(userID string) error {
	_, err := r.db.Exec(`DELETE FROM slack_settings WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete slack settings: %w", err)
	}
	return nil
}

// SetPostHogSettings stores or updates PostHog settings
func (r *IntegrationRepository) SetPostHogSettings(userID string, enabled bool) error {
	now := time.Now()
	_, err := r.db.Exec(`
		INSERT INTO posthog_settings (user_id, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			enabled = excluded.enabled,
			updated_at = excluded.updated_at
	`, userID, enabled, now, now)

	if err != nil {
		return fmt.Errorf("failed to set posthog settings: %w", err)
	}
	return nil
}

// GetPostHogSettings retrieves PostHog settings for a user
func (r *IntegrationRepository) GetPostHogSettings(userID string) (*IntegrationSettings, error) {
	var enabled bool
	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(`
		SELECT enabled, created_at, updated_at
		FROM posthog_settings
		WHERE user_id = ?
	`, userID).Scan(&enabled, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get posthog settings: %w", err)
	}

	return &IntegrationSettings{
		UserID:    userID,
		Type:      "posthog",
		Enabled:   enabled,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// DeletePostHogSettings removes PostHog settings for a user
func (r *IntegrationRepository) DeletePostHogSettings(userID string) error {
	_, err := r.db.Exec(`DELETE FROM posthog_settings WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete posthog settings: %w", err)
	}
	return nil
}

// GetAllSettings retrieves all integration settings for a user
func (r *IntegrationRepository) GetAllSettings(userID string) (map[string]*IntegrationSettings, error) {
	result := make(map[string]*IntegrationSettings)

	discord, err := r.GetDiscordSettings(userID)
	if err != nil {
		return nil, err
	}
	if discord != nil {
		result["discord"] = discord
	}

	slack, err := r.GetSlackSettings(userID)
	if err != nil {
		return nil, err
	}
	if slack != nil {
		result["slack"] = slack
	}

	posthog, err := r.GetPostHogSettings(userID)
	if err != nil {
		return nil, err
	}
	if posthog != nil {
		result["posthog"] = posthog
	}

	return result, nil
}
