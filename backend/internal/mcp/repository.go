package mcp

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// Repository handles MCP server database operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new MCP repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new MCP server connection
func (r *Repository) Create(server *RemoteServer) error {
	manifestJSON, err := json.Marshal(server.Manifest)
	if err != nil {
		manifestJSON = nil
	}

	_, err = r.db.Exec(
		`INSERT INTO mcp_connections (id, user_id, name, url, api_key, enabled, manifest, created_at, updated_at, last_sync, last_error)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		server.ID, server.UserID, server.Name, server.URL, server.APIKey,
		server.Enabled, manifestJSON, server.CreatedAt, server.UpdatedAt,
		server.LastSync, server.LastError,
	)
	if err != nil {
		return fmt.Errorf("failed to create MCP connection: %w", err)
	}

	return nil
}

// GetByID retrieves an MCP server by ID
func (r *Repository) GetByID(id string) (*RemoteServer, error) {
	server := &RemoteServer{}
	var manifestJSON sql.NullString
	var apiKey sql.NullString
	var lastSync sql.NullTime
	var lastError sql.NullString

	err := r.db.QueryRow(
		`SELECT id, user_id, name, url, api_key, enabled, manifest, created_at, updated_at, last_sync, last_error
		 FROM mcp_connections WHERE id = ?`,
		id,
	).Scan(
		&server.ID, &server.UserID, &server.Name, &server.URL, &apiKey,
		&server.Enabled, &manifestJSON, &server.CreatedAt, &server.UpdatedAt,
		&lastSync, &lastError,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get MCP connection: %w", err)
	}

	server.APIKey = apiKey.String
	server.LastError = lastError.String
	if lastSync.Valid {
		server.LastSync = &lastSync.Time
	}

	if manifestJSON.Valid && manifestJSON.String != "" {
		var manifest Manifest
		if err := json.Unmarshal([]byte(manifestJSON.String), &manifest); err == nil {
			server.Manifest = &manifest
		}
	}

	return server, nil
}

// GetByUserID retrieves all MCP servers for a user
func (r *Repository) GetByUserID(userID string) ([]*RemoteServer, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, name, url, api_key, enabled, manifest, created_at, updated_at, last_sync, last_error
		 FROM mcp_connections WHERE user_id = ? ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get MCP connections: %w", err)
	}
	defer rows.Close()

	servers := make([]*RemoteServer, 0)
	for rows.Next() {
		server := &RemoteServer{}
		var manifestJSON sql.NullString
		var apiKey sql.NullString
		var lastSync sql.NullTime
		var lastError sql.NullString

		err := rows.Scan(
			&server.ID, &server.UserID, &server.Name, &server.URL, &apiKey,
			&server.Enabled, &manifestJSON, &server.CreatedAt, &server.UpdatedAt,
			&lastSync, &lastError,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan MCP connection: %w", err)
		}

		server.APIKey = apiKey.String
		server.LastError = lastError.String
		if lastSync.Valid {
			server.LastSync = &lastSync.Time
		}

		if manifestJSON.Valid && manifestJSON.String != "" {
			var manifest Manifest
			if err := json.Unmarshal([]byte(manifestJSON.String), &manifest); err == nil {
				server.Manifest = &manifest
			}
		}

		servers = append(servers, server)
	}

	return servers, nil
}

// Update updates an MCP server connection
func (r *Repository) Update(server *RemoteServer) error {
	manifestJSON, err := json.Marshal(server.Manifest)
	if err != nil {
		manifestJSON = nil
	}

	_, err = r.db.Exec(
		`UPDATE mcp_connections SET name = ?, url = ?, api_key = ?, enabled = ?, manifest = ?, updated_at = ?, last_sync = ?, last_error = ?
		 WHERE id = ?`,
		server.Name, server.URL, server.APIKey, server.Enabled, manifestJSON,
		server.UpdatedAt, server.LastSync, server.LastError, server.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update MCP connection: %w", err)
	}

	return nil
}

// Delete deletes an MCP server connection
func (r *Repository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM mcp_connections WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete MCP connection: %w", err)
	}
	return nil
}

// DeleteByUserID deletes all MCP connections for a user
func (r *Repository) DeleteByUserID(userID string) error {
	_, err := r.db.Exec(`DELETE FROM mcp_connections WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete MCP connections: %w", err)
	}
	return nil
}

// LoadAllEnabled loads all enabled MCP connections into the client
func (r *Repository) LoadAllEnabled(client *Client) error {
	rows, err := r.db.Query(
		`SELECT id, user_id, name, url, api_key, enabled, manifest, created_at, updated_at, last_sync, last_error
		 FROM mcp_connections WHERE enabled = true`,
	)
	if err != nil {
		return fmt.Errorf("failed to load MCP connections: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		server := &RemoteServer{}
		var manifestJSON sql.NullString
		var apiKey sql.NullString
		var lastSync sql.NullTime
		var lastError sql.NullString

		err := rows.Scan(
			&server.ID, &server.UserID, &server.Name, &server.URL, &apiKey,
			&server.Enabled, &manifestJSON, &server.CreatedAt, &server.UpdatedAt,
			&lastSync, &lastError,
		)
		if err != nil {
			continue
		}

		server.APIKey = apiKey.String
		server.LastError = lastError.String
		if lastSync.Valid {
			server.LastSync = &lastSync.Time
		}

		if manifestJSON.Valid && manifestJSON.String != "" {
			var manifest Manifest
			if err := json.Unmarshal([]byte(manifestJSON.String), &manifest); err == nil {
				server.Manifest = &manifest
			}
		}

		// Add to client (will refresh manifest)
		client.AddServer(server)
	}

	return nil
}

// CreateTable creates the mcp_connections table if it doesn't exist
func (r *Repository) CreateTable() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS mcp_connections (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			name TEXT NOT NULL,
			url TEXT NOT NULL,
			api_key TEXT,
			enabled BOOLEAN DEFAULT true,
			manifest TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			last_sync DATETIME,
			last_error TEXT,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create mcp_connections table: %w", err)
	}

	// Create index for faster lookups
	_, err = r.db.Exec(`CREATE INDEX IF NOT EXISTS idx_mcp_connections_user_id ON mcp_connections(user_id)`)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}
