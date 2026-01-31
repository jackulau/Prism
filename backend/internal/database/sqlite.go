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

		// Stdio MCP servers (local MCP servers connected via stdin/stdout)
		`CREATE TABLE IF NOT EXISTS mcp_stdio_servers (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			command TEXT NOT NULL,
			args TEXT,
			env TEXT,
			enabled INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_error TEXT
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

		// User workspaces for persistent project directory storage
		`CREATE TABLE IF NOT EXISTS user_workspaces (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			path TEXT NOT NULL,
			name TEXT,
			is_current INTEGER DEFAULT 0,
			last_accessed_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, path)
		)`,

		// Workspace todos for task tracking
		`CREATE TABLE IF NOT EXISTS workspace_todos (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			workspace_path TEXT NOT NULL,
			content TEXT NOT NULL,
			active_form TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Agent tasks for persistent task queue
		`CREATE TABLE IF NOT EXISTS agent_tasks (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			prompt TEXT NOT NULL,
			context TEXT,
			priority INTEGER DEFAULT 1,
			status TEXT DEFAULT 'pending',
			agent_config TEXT,
			metadata TEXT,
			result TEXT,
			error TEXT,
			callback_url TEXT,
			callback_data TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			started_at DATETIME,
			completed_at DATETIME
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

		// Organization-scoped workspaces for agent sessions
		`CREATE TABLE IF NOT EXISTS org_workspaces (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			organization_id TEXT NOT NULL,
			github_repository_name TEXT,
			worker_id TEXT,
			current_branch TEXT,
			slack_channel_id TEXT,
			slack_message_ts TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Add GitHub fields to users table (safe migrations with ALTER TABLE)
		`ALTER TABLE users ADD COLUMN github_token TEXT`,
		`ALTER TABLE users ADD COLUMN github_username TEXT`,
		`ALTER TABLE users ADD COLUMN github_connected_at DATETIME`,

		// Add WorkOS SSO fields to users table
		`ALTER TABLE users ADD COLUMN workos_id TEXT`,
		`ALTER TABLE users ADD COLUMN organization_id TEXT`,
		`ALTER TABLE users ADD COLUMN sso_connection_id TEXT`,
		`ALTER TABLE users ADD COLUMN sso_provider TEXT`,

		// Add last_used and use_count tracking to provider_keys
		`ALTER TABLE provider_keys ADD COLUMN last_used_at DATETIME`,
		`ALTER TABLE provider_keys ADD COLUMN use_count INTEGER DEFAULT 0`,

		// API key scopes table for granular permissions
		`CREATE TABLE IF NOT EXISTS api_key_scopes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			api_key_id TEXT NOT NULL,
			scope TEXT NOT NULL,
			FOREIGN KEY (api_key_id) REFERENCES user_api_keys(id) ON DELETE CASCADE
		)`,

		// Audit logs table for security event tracking
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT,
			event_type TEXT NOT NULL,
			event_category TEXT NOT NULL,
			action TEXT NOT NULL,
			resource_type TEXT,
			resource_id TEXT,
			ip_address TEXT,
			user_agent TEXT,
			details TEXT,
			success INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
		)`,

		// Add RBAC role field to users table
		`ALTER TABLE users ADD COLUMN role TEXT DEFAULT 'user'`,

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_api_key_scopes_key_id ON api_key_scopes(api_key_id)`,
		`CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)`,
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
		`CREATE INDEX IF NOT EXISTS idx_mcp_stdio_servers_user_id ON mcp_stdio_servers(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_integrations_user_id ON user_integrations(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tools_slug ON tools(slug_name)`,
		`CREATE INDEX IF NOT EXISTS idx_tools_provider ON tools(provider_id)`,
		`CREATE INDEX IF NOT EXISTS idx_file_history_user_id ON file_history(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_file_history_file_path ON file_history(user_id, file_path)`,
		`CREATE INDEX IF NOT EXISTS idx_user_workspaces_user_id ON user_workspaces(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_workspaces_current ON user_workspaces(user_id, is_current)`,
		`CREATE INDEX IF NOT EXISTS idx_workspace_todos_user_workspace ON workspace_todos(user_id, workspace_path)`,
		`CREATE INDEX IF NOT EXISTS idx_org_workspaces_organization_id ON org_workspaces(organization_id)`,
		`CREATE INDEX IF NOT EXISTS idx_org_workspaces_github_repo ON org_workspaces(organization_id, github_repository_name)`,
		`CREATE INDEX IF NOT EXISTS idx_org_workspaces_current_branch ON org_workspaces(current_branch)`,

		// SSO configurations table
		`CREATE TABLE IF NOT EXISTS sso_configurations (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			priority INTEGER DEFAULT 0,
			enabled INTEGER DEFAULT 0,
			config_json TEXT,
			encrypted_secret BLOB,
			secret_nonce BLOB,
			workos_connection_id TEXT,
			last_error TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// SSO attribute mappings table
		`CREATE TABLE IF NOT EXISTS sso_attribute_mappings (
			id TEXT PRIMARY KEY,
			sso_provider_id TEXT NOT NULL REFERENCES sso_configurations(id) ON DELETE CASCADE,
			source_attribute TEXT NOT NULL,
			target_field TEXT NOT NULL,
			transform_type TEXT,
			transform_pattern TEXT
		)`,

		// Organizations table (if not exists)
		`CREATE TABLE IF NOT EXISTS organizations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			workos_organization_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Organization members table (if not exists)
		`CREATE TABLE IF NOT EXISTS organization_members (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role TEXT NOT NULL DEFAULT 'member',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(organization_id, user_id)
		)`,

		// SSO indexes
		`CREATE INDEX IF NOT EXISTS idx_sso_configurations_org ON sso_configurations(organization_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sso_configurations_type ON sso_configurations(organization_id, type)`,
		`CREATE INDEX IF NOT EXISTS idx_sso_configurations_enabled ON sso_configurations(organization_id, enabled, status)`,
		`CREATE INDEX IF NOT EXISTS idx_sso_attribute_mappings_provider ON sso_attribute_mappings(sso_provider_id)`,
		`CREATE INDEX IF NOT EXISTS idx_organizations_workos ON organizations(workos_organization_id)`,
		`CREATE INDEX IF NOT EXISTS idx_organization_members_org ON organization_members(organization_id)`,
		`CREATE INDEX IF NOT EXISTS idx_organization_members_user ON organization_members(user_id)`,

		// Approval workflows
		`CREATE TABLE IF NOT EXISTS approval_workflows (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			operation_type TEXT NOT NULL,
			steps TEXT NOT NULL,
			conditions TEXT,
			metadata TEXT,
			active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT NOT NULL
		)`,

		// Approval requests
		`CREATE TABLE IF NOT EXISTS approval_requests (
			id TEXT PRIMARY KEY,
			workflow_id TEXT NOT NULL REFERENCES approval_workflows(id) ON DELETE CASCADE,
			organization_id TEXT NOT NULL,
			requester_id TEXT NOT NULL,
			requester_email TEXT,
			operation_type TEXT NOT NULL,
			operation_details TEXT,
			current_step INTEGER DEFAULT 0,
			total_steps INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			priority INTEGER DEFAULT 0,
			expires_at DATETIME,
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME
		)`,

		// Approval decisions (audit trail)
		`CREATE TABLE IF NOT EXISTS approval_decisions (
			id TEXT PRIMARY KEY,
			request_id TEXT NOT NULL REFERENCES approval_requests(id) ON DELETE CASCADE,
			step_order INTEGER NOT NULL,
			approver_id TEXT NOT NULL,
			approver_email TEXT,
			decision TEXT NOT NULL,
			comment TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			metadata TEXT
		)`,

		// Approval workflow indexes
		`CREATE INDEX IF NOT EXISTS idx_approval_workflows_org ON approval_workflows(organization_id)`,
		`CREATE INDEX IF NOT EXISTS idx_approval_workflows_operation ON approval_workflows(organization_id, operation_type)`,
		`CREATE INDEX IF NOT EXISTS idx_approval_workflows_active ON approval_workflows(organization_id, active)`,
		`CREATE INDEX IF NOT EXISTS idx_approval_requests_org ON approval_requests(organization_id)`,
		`CREATE INDEX IF NOT EXISTS idx_approval_requests_workflow ON approval_requests(workflow_id)`,
		`CREATE INDEX IF NOT EXISTS idx_approval_requests_requester ON approval_requests(requester_id)`,
		`CREATE INDEX IF NOT EXISTS idx_approval_requests_status ON approval_requests(organization_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_approval_requests_expires ON approval_requests(status, expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_approval_decisions_request ON approval_decisions(request_id)`,
		`CREATE INDEX IF NOT EXISTS idx_approval_decisions_approver ON approval_decisions(approver_id)`,
		`CREATE INDEX IF NOT EXISTS idx_approval_decisions_step ON approval_decisions(request_id, step_order)`,

		// Audit log indexes
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_event_type ON audit_logs(event_type)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_event_category ON audit_logs(event_category)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource_type, resource_id)`,

		// Build configuration tables
		`CREATE TABLE IF NOT EXISTS build_configs (
			id TEXT PRIMARY KEY,
			workspace_id TEXT,
			org_workspace_id TEXT,
			user_id TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			is_default INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS build_commands (
			id TEXT PRIMARY KEY,
			config_id TEXT NOT NULL,
			name TEXT NOT NULL,
			command TEXT NOT NULL,
			working_directory TEXT,
			run_order INTEGER DEFAULT 0,
			is_enabled INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (config_id) REFERENCES build_configs(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS build_env_vars (
			id TEXT PRIMARY KEY,
			config_id TEXT NOT NULL,
			key TEXT NOT NULL,
			value_encrypted BLOB NOT NULL,
			value_nonce BLOB NOT NULL,
			is_secret INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (config_id) REFERENCES build_configs(id) ON DELETE CASCADE,
			UNIQUE(config_id, key)
		)`,

		// Build config indexes
		`CREATE INDEX IF NOT EXISTS idx_build_configs_workspace ON build_configs(workspace_id)`,
		`CREATE INDEX IF NOT EXISTS idx_build_configs_org_workspace ON build_configs(org_workspace_id)`,
		`CREATE INDEX IF NOT EXISTS idx_build_configs_user ON build_configs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_build_commands_config ON build_commands(config_id)`,
		`CREATE INDEX IF NOT EXISTS idx_build_env_vars_config ON build_env_vars(config_id)`,
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
