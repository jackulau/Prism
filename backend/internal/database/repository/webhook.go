package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/integrations/github"
	"github.com/jacklau/prism/internal/security"
)

// WebhookRepository handles webhook-related database operations
type WebhookRepository struct {
	db                *sql.DB
	encryptionService *security.EncryptionService
}

// NewWebhookRepository creates a new webhook repository
func NewWebhookRepository(db *sql.DB, encryptionService *security.EncryptionService) *WebhookRepository {
	return &WebhookRepository{
		db:                db,
		encryptionService: encryptionService,
	}
}

// Create creates a new webhook configuration
func (r *WebhookRepository) Create(config *github.WebhookConfig) error {
	config.ID = uuid.New().String()
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	// Encrypt the webhook secret
	encryptedSecret, nonce, err := r.encryptionService.Encrypt([]byte(config.WebhookSecret))
	if err != nil {
		return fmt.Errorf("failed to encrypt webhook secret: %w", err)
	}

	// Marshal events and triggers to JSON
	eventsJSON, err := json.Marshal(config.Events)
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	triggersJSON, err := json.Marshal(config.AutoRunTriggers)
	if err != nil {
		return fmt.Errorf("failed to marshal triggers: %w", err)
	}

	query := `
		INSERT INTO github_webhooks (
			id, user_id, repo_full_name, webhook_secret_encrypted, webhook_secret_nonce,
			events, auto_run_enabled, auto_run_triggers, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(query,
		config.ID,
		config.UserID,
		config.RepoFullName,
		encryptedSecret,
		nonce,
		string(eventsJSON),
		config.AutoRunEnabled,
		string(triggersJSON),
		config.CreatedAt,
		config.UpdatedAt,
	)

	return err
}

// GetByID retrieves a webhook configuration by ID
func (r *WebhookRepository) GetByID(id string) (*github.WebhookConfig, error) {
	query := `
		SELECT id, user_id, repo_full_name, webhook_secret_encrypted, webhook_secret_nonce,
			   events, auto_run_enabled, auto_run_triggers, created_at, updated_at
		FROM github_webhooks
		WHERE id = ?
	`

	return r.scanWebhook(r.db.QueryRow(query, id))
}

// GetByRepoName retrieves a webhook configuration by repository name
func (r *WebhookRepository) GetByRepoName(repoFullName string) (*github.WebhookConfig, error) {
	query := `
		SELECT id, user_id, repo_full_name, webhook_secret_encrypted, webhook_secret_nonce,
			   events, auto_run_enabled, auto_run_triggers, created_at, updated_at
		FROM github_webhooks
		WHERE repo_full_name = ?
		LIMIT 1
	`

	return r.scanWebhook(r.db.QueryRow(query, repoFullName))
}

// ListByUser retrieves all webhook configurations for a user
func (r *WebhookRepository) ListByUser(userID string) ([]*github.WebhookConfig, error) {
	query := `
		SELECT id, user_id, repo_full_name, webhook_secret_encrypted, webhook_secret_nonce,
			   events, auto_run_enabled, auto_run_triggers, created_at, updated_at
		FROM github_webhooks
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*github.WebhookConfig
	for rows.Next() {
		config, err := r.scanWebhookRow(rows)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}

	return configs, rows.Err()
}

// Update updates a webhook configuration
func (r *WebhookRepository) Update(config *github.WebhookConfig) error {
	config.UpdatedAt = time.Now()

	// Marshal events and triggers to JSON
	eventsJSON, err := json.Marshal(config.Events)
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	triggersJSON, err := json.Marshal(config.AutoRunTriggers)
	if err != nil {
		return fmt.Errorf("failed to marshal triggers: %w", err)
	}

	query := `
		UPDATE github_webhooks
		SET events = ?, auto_run_enabled = ?, auto_run_triggers = ?, updated_at = ?
		WHERE id = ?
	`

	_, err = r.db.Exec(query,
		string(eventsJSON),
		config.AutoRunEnabled,
		string(triggersJSON),
		config.UpdatedAt,
		config.ID,
	)

	return err
}

// Delete deletes a webhook configuration
func (r *WebhookRepository) Delete(id string) error {
	query := `DELETE FROM github_webhooks WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// CreateDelivery creates a new webhook delivery record
func (r *WebhookRepository) CreateDelivery(delivery *github.WebhookDelivery) error {
	delivery.ID = uuid.New().String()
	delivery.CreatedAt = time.Now()

	payloadJSON, err := json.Marshal(delivery.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	query := `
		INSERT INTO webhook_deliveries (
			id, webhook_id, event, action, payload, status, error_message, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(query,
		delivery.ID,
		delivery.WebhookID,
		delivery.Event,
		delivery.Action,
		string(payloadJSON),
		delivery.Status,
		delivery.ErrorMessage,
		delivery.CreatedAt,
	)

	return err
}

// UpdateDelivery updates a webhook delivery record
func (r *WebhookRepository) UpdateDelivery(delivery *github.WebhookDelivery) error {
	query := `
		UPDATE webhook_deliveries
		SET status = ?, error_message = ?, processed_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		delivery.Status,
		delivery.ErrorMessage,
		delivery.ProcessedAt,
		delivery.ID,
	)

	return err
}

// ListDeliveries lists webhook deliveries for a webhook
func (r *WebhookRepository) ListDeliveries(webhookID string, limit int) ([]*github.WebhookDelivery, error) {
	query := `
		SELECT id, webhook_id, event, action, payload, status, error_message, processed_at, created_at
		FROM webhook_deliveries
		WHERE webhook_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, webhookID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []*github.WebhookDelivery
	for rows.Next() {
		var delivery github.WebhookDelivery
		var payloadJSON string
		var processedAt sql.NullTime

		err := rows.Scan(
			&delivery.ID,
			&delivery.WebhookID,
			&delivery.Event,
			&delivery.Action,
			&payloadJSON,
			&delivery.Status,
			&delivery.ErrorMessage,
			&processedAt,
			&delivery.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(payloadJSON), &delivery.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		if processedAt.Valid {
			delivery.ProcessedAt = &processedAt.Time
		}

		deliveries = append(deliveries, &delivery)
	}

	return deliveries, rows.Err()
}

// CreateExecution creates a new code execution record
func (r *WebhookRepository) CreateExecution(result *github.CodeExecutionResult, deliveryID, userID string) error {
	query := `
		INSERT INTO code_executions (
			id, delivery_id, user_id, command, environment, exit_code,
			stdout, stderr, duration_ms, started_at, completed_at, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		result.ID,
		sql.NullString{String: deliveryID, Valid: deliveryID != ""},
		sql.NullString{String: userID, Valid: userID != ""},
		result.Command,
		result.Environment,
		result.ExitCode,
		result.Stdout,
		result.Stderr,
		result.Duration,
		result.StartedAt,
		result.CompletedAt,
		time.Now(),
	)

	return err
}

// ListExecutions lists code executions for a user
func (r *WebhookRepository) ListExecutions(userID string, limit int) ([]*github.CodeExecutionResult, error) {
	query := `
		SELECT id, command, environment, exit_code, stdout, stderr, duration_ms, started_at, completed_at
		FROM code_executions
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*github.CodeExecutionResult
	for rows.Next() {
		var result github.CodeExecutionResult
		err := rows.Scan(
			&result.ID,
			&result.Command,
			&result.Environment,
			&result.ExitCode,
			&result.Stdout,
			&result.Stderr,
			&result.Duration,
			&result.StartedAt,
			&result.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, &result)
	}

	return results, rows.Err()
}

// scanWebhook scans a single row into a WebhookConfig
func (r *WebhookRepository) scanWebhook(row *sql.Row) (*github.WebhookConfig, error) {
	var config github.WebhookConfig
	var encryptedSecret, nonce []byte
	var eventsJSON, triggersJSON string

	err := row.Scan(
		&config.ID,
		&config.UserID,
		&config.RepoFullName,
		&encryptedSecret,
		&nonce,
		&eventsJSON,
		&config.AutoRunEnabled,
		&triggersJSON,
		&config.CreatedAt,
		&config.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Decrypt the webhook secret
	secret, err := r.encryptionService.Decrypt(encryptedSecret, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt webhook secret: %w", err)
	}
	config.WebhookSecret = string(secret)

	// Unmarshal events
	if eventsJSON != "" {
		if err := json.Unmarshal([]byte(eventsJSON), &config.Events); err != nil {
			return nil, fmt.Errorf("failed to unmarshal events: %w", err)
		}
	}

	// Unmarshal triggers
	if triggersJSON != "" {
		if err := json.Unmarshal([]byte(triggersJSON), &config.AutoRunTriggers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal triggers: %w", err)
		}
	}

	return &config, nil
}

// scanWebhookRow scans a rows result into a WebhookConfig
func (r *WebhookRepository) scanWebhookRow(rows *sql.Rows) (*github.WebhookConfig, error) {
	var config github.WebhookConfig
	var encryptedSecret, nonce []byte
	var eventsJSON, triggersJSON string

	err := rows.Scan(
		&config.ID,
		&config.UserID,
		&config.RepoFullName,
		&encryptedSecret,
		&nonce,
		&eventsJSON,
		&config.AutoRunEnabled,
		&triggersJSON,
		&config.CreatedAt,
		&config.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Decrypt the webhook secret
	secret, err := r.encryptionService.Decrypt(encryptedSecret, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt webhook secret: %w", err)
	}
	config.WebhookSecret = string(secret)

	// Unmarshal events
	if eventsJSON != "" {
		if err := json.Unmarshal([]byte(eventsJSON), &config.Events); err != nil {
			return nil, fmt.Errorf("failed to unmarshal events: %w", err)
		}
	}

	// Unmarshal triggers
	if triggersJSON != "" {
		if err := json.Unmarshal([]byte(triggersJSON), &config.AutoRunTriggers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal triggers: %w", err)
		}
	}

	return &config, nil
}
