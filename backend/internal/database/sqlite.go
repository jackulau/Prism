package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func NewSQLite(databaseURL string) (*DB, error) {
	// Ensure the directory exists
	dir := filepath.Dir(databaseURL)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", databaseURL+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

func (db *DB) Migrate() error {
	migrations := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Sessions table
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			refresh_token_hash TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// User API keys (for external access)
		`CREATE TABLE IF NOT EXISTS user_api_keys (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			key_hash TEXT NOT NULL,
			key_prefix TEXT NOT NULL,
			last_used_at DATETIME,
			expires_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Provider API keys (encrypted)
		`CREATE TABLE IF NOT EXISTS provider_keys (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			provider TEXT NOT NULL,
			encrypted_key BLOB NOT NULL,
			key_nonce BLOB NOT NULL,
			is_active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, provider)
		)`,

		// GitHub connections
		`CREATE TABLE IF NOT EXISTS github_connections (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			github_user_id TEXT NOT NULL,
			github_username TEXT NOT NULL,
			encrypted_access_token BLOB NOT NULL,
			token_nonce BLOB NOT NULL,
			scopes TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id)
		)`,

		// Conversations
		`CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title TEXT,
			provider TEXT NOT NULL,
			model TEXT NOT NULL,
			system_prompt TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Messages
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			tool_calls TEXT,
			tool_call_id TEXT,
			tokens_used INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Tool executions
		`CREATE TABLE IF NOT EXISTS tool_executions (
			id TEXT PRIMARY KEY,
			message_id TEXT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
			tool_name TEXT NOT NULL,
			parameters TEXT NOT NULL,
			result TEXT,
			status TEXT NOT NULL,
			execution_time_ms INTEGER,
			container_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// User settings
		`CREATE TABLE IF NOT EXISTS user_settings (
			user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			default_provider TEXT,
			default_model TEXT,
			theme TEXT DEFAULT 'system',
			settings_json TEXT,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Tool settings (for per-tool approval configuration)
		`CREATE TABLE IF NOT EXISTS tool_settings (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			tool_name TEXT NOT NULL,
			requires_approval INTEGER DEFAULT 0,
			is_enabled INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, tool_name)
		)`,

		// File uploads
		`CREATE TABLE IF NOT EXISTS uploads (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			message_id TEXT REFERENCES messages(id) ON DELETE SET NULL,
			filename TEXT NOT NULL,
			file_type TEXT NOT NULL,
			file_size INTEGER NOT NULL,
			storage_path TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Discord integration settings
		`CREATE TABLE IF NOT EXISTS discord_settings (
			user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			webhook_url_encrypted BLOB,
			webhook_url_nonce BLOB,
			bot_token_encrypted BLOB,
			bot_token_nonce BLOB,
			enabled INTEGER DEFAULT 0,
			notify_on_conversation INTEGER DEFAULT 1,
			notify_on_tool_execution INTEGER DEFAULT 1,
			notify_on_error INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Slack integration settings
		`CREATE TABLE IF NOT EXISTS slack_settings (
			user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			webhook_url_encrypted BLOB,
			webhook_url_nonce BLOB,
			bot_token_encrypted BLOB,
			bot_token_nonce BLOB,
			channel_id TEXT,
			enabled INTEGER DEFAULT 0,
			notify_on_conversation INTEGER DEFAULT 1,
			notify_on_tool_execution INTEGER DEFAULT 1,
			notify_on_error INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// PostHog analytics settings (per-user overrides)
		`CREATE TABLE IF NOT EXISTS posthog_settings (
			user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			enabled INTEGER DEFAULT 1,
			track_conversations INTEGER DEFAULT 1,
			track_messages INTEGER DEFAULT 1,
			track_tool_usage INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// GitHub webhook configurations
		`CREATE TABLE IF NOT EXISTS github_webhooks (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			repo_full_name TEXT NOT NULL,
			webhook_secret_encrypted BLOB NOT NULL,
			webhook_secret_nonce BLOB NOT NULL,
			events TEXT,
			auto_run_enabled INTEGER DEFAULT 0,
			auto_run_triggers TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, repo_full_name)
		)`,

		// Webhook deliveries (audit log)
		`CREATE TABLE IF NOT EXISTS webhook_deliveries (
			id TEXT PRIMARY KEY,
			webhook_id TEXT NOT NULL REFERENCES github_webhooks(id) ON DELETE CASCADE,
			event TEXT NOT NULL,
			action TEXT,
			payload TEXT NOT NULL,
			status TEXT NOT NULL,
			error_message TEXT,
			processed_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Code execution results
		`CREATE TABLE IF NOT EXISTS code_executions (
			id TEXT PRIMARY KEY,
			delivery_id TEXT REFERENCES webhook_deliveries(id) ON DELETE SET NULL,
			user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
			command TEXT NOT NULL,
			environment TEXT NOT NULL,
			exit_code INTEGER,
			stdout TEXT,
			stderr TEXT,
			duration_ms INTEGER,
			started_at DATETIME NOT NULL,
			completed_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// MCP server connections (external MCP servers user connects to)
		`CREATE TABLE IF NOT EXISTS mcp_connections (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			url TEXT NOT NULL,
			api_key TEXT,
			enabled INTEGER DEFAULT 1,
			manifest TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_sync DATETIME,
			last_error TEXT
		)`,

		// MCP API keys (for external clients accessing Prism's MCP server)
		`CREATE TABLE IF NOT EXISTS mcp_api_keys (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			key_hash TEXT NOT NULL,
			key_prefix TEXT NOT NULL,
			permissions TEXT,
			last_used_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// File history for tracking changes made by agent
		`CREATE TABLE IF NOT EXISTS file_history (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			file_path TEXT NOT NULL,
			content TEXT NOT NULL,
			operation TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// User integrations (generic per-user integration configs)
		`CREATE TABLE IF NOT EXISTS user_integrations (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			type TEXT NOT NULL,
			enabled INTEGER DEFAULT 0,
			config TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, type)
		)`,

		// Add GitHub fields to users table (safe migrations with ALTER TABLE)
		`ALTER TABLE users ADD COLUMN github_token TEXT`,
		`ALTER TABLE users ADD COLUMN github_username TEXT`,
		`ALTER TABLE users ADD COLUMN github_connected_at DATETIME`,

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_conversations_user_id ON conversations(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tool_executions_message_id ON tool_executions(message_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_api_keys_user_id ON user_api_keys(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_api_keys_key_hash ON user_api_keys(key_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_uploads_user_id ON uploads(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_github_webhooks_user_id ON github_webhooks(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_github_webhooks_repo ON github_webhooks(repo_full_name)`,
		`CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_webhook_id ON webhook_deliveries(webhook_id)`,
		`CREATE INDEX IF NOT EXISTS idx_code_executions_delivery_id ON code_executions(delivery_id)`,
		`CREATE INDEX IF NOT EXISTS idx_code_executions_user_id ON code_executions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_mcp_connections_user_id ON mcp_connections(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_mcp_api_keys_user_id ON mcp_api_keys(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_mcp_api_keys_key_hash ON mcp_api_keys(key_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_user_integrations_user_id ON user_integrations(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_file_history_user_id ON file_history(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_file_history_file_path ON file_history(user_id, file_path)`,
	}

	for _, migration := range migrations {
		_, err := db.Exec(migration)
		if err != nil {
			// Ignore "duplicate column" errors from ALTER TABLE
			// SQLite returns "duplicate column name" when column already exists
			if strings.Contains(err.Error(), "duplicate column") {
				continue
			}
			return fmt.Errorf("failed to run migration: %w\nSQL: %s", err, migration)
		}
	}

	return nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}
