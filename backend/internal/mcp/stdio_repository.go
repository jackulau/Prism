package mcp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// StdioRepository handles stdio MCP server database operations
type StdioRepository struct {
	db *sql.DB
}

// NewStdioRepository creates a new stdio MCP repository
func NewStdioRepository(db *sql.DB) *StdioRepository {
	return &StdioRepository{db: db}
}

// Create creates a new stdio MCP server connection
func (r *StdioRepository) Create(server *StdioServer) error {
	argsJSON, err := json.Marshal(server.Args)
	if err != nil {
		argsJSON = []byte("[]")
	}

	envJSON, err := json.Marshal(server.Env)
	if err != nil {
		envJSON = []byte("[]")
	}

	_, err = r.db.Exec(
		`INSERT INTO mcp_stdio_servers (id, user_id, name, command, args, env, enabled, created_at, updated_at, last_error)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		server.ID, server.UserID, server.Name, server.Command, argsJSON, envJSON,
		server.Enabled, server.CreatedAt, server.UpdatedAt, server.LastError,
	)
	if err != nil {
		return fmt.Errorf("failed to create stdio MCP server: %w", err)
	}

	return nil
}

// GetByID retrieves a stdio MCP server by ID
func (r *StdioRepository) GetByID(id string) (*StdioServer, error) {
	server := &StdioServer{}
	var argsJSON, envJSON sql.NullString
	var lastError sql.NullString

	err := r.db.QueryRow(
		`SELECT id, user_id, name, command, args, env, enabled, created_at, updated_at, last_error
		 FROM mcp_stdio_servers WHERE id = ?`,
		id,
	).Scan(
		&server.ID, &server.UserID, &server.Name, &server.Command, &argsJSON, &envJSON,
		&server.Enabled, &server.CreatedAt, &server.UpdatedAt, &lastError,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stdio MCP server: %w", err)
	}

	server.LastError = lastError.String

	if argsJSON.Valid && argsJSON.String != "" {
		json.Unmarshal([]byte(argsJSON.String), &server.Args)
	}
	if envJSON.Valid && envJSON.String != "" {
		json.Unmarshal([]byte(envJSON.String), &server.Env)
	}

	return server, nil
}

// GetByUserID retrieves all stdio MCP servers for a user
func (r *StdioRepository) GetByUserID(userID string) ([]*StdioServer, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, name, command, args, env, enabled, created_at, updated_at, last_error
		 FROM mcp_stdio_servers WHERE user_id = ? ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get stdio MCP servers: %w", err)
	}
	defer rows.Close()

	servers := make([]*StdioServer, 0)
	for rows.Next() {
		server := &StdioServer{}
		var argsJSON, envJSON sql.NullString
		var lastError sql.NullString

		err := rows.Scan(
			&server.ID, &server.UserID, &server.Name, &server.Command, &argsJSON, &envJSON,
			&server.Enabled, &server.CreatedAt, &server.UpdatedAt, &lastError,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stdio MCP server: %w", err)
		}

		server.LastError = lastError.String

		if argsJSON.Valid && argsJSON.String != "" {
			json.Unmarshal([]byte(argsJSON.String), &server.Args)
		}
		if envJSON.Valid && envJSON.String != "" {
			json.Unmarshal([]byte(envJSON.String), &server.Env)
		}

		servers = append(servers, server)
	}

	return servers, nil
}

// Update updates a stdio MCP server connection
func (r *StdioRepository) Update(server *StdioServer) error {
	argsJSON, err := json.Marshal(server.Args)
	if err != nil {
		argsJSON = []byte("[]")
	}

	envJSON, err := json.Marshal(server.Env)
	if err != nil {
		envJSON = []byte("[]")
	}

	_, err = r.db.Exec(
		`UPDATE mcp_stdio_servers SET name = ?, command = ?, args = ?, env = ?, enabled = ?, updated_at = ?, last_error = ?
		 WHERE id = ?`,
		server.Name, server.Command, argsJSON, envJSON, server.Enabled, time.Now(), server.LastError, server.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update stdio MCP server: %w", err)
	}

	return nil
}

// Delete deletes a stdio MCP server connection
func (r *StdioRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM mcp_stdio_servers WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete stdio MCP server: %w", err)
	}
	return nil
}

// DeleteByUserID deletes all stdio MCP servers for a user
func (r *StdioRepository) DeleteByUserID(userID string) error {
	_, err := r.db.Exec(`DELETE FROM mcp_stdio_servers WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete stdio MCP servers: %w", err)
	}
	return nil
}

// LoadAllEnabled loads all enabled stdio MCP servers into the client
func (r *StdioRepository) LoadAllEnabled(client *StdioClient) error {
	rows, err := r.db.Query(
		`SELECT id, user_id, name, command, args, env, enabled, created_at, updated_at, last_error
		 FROM mcp_stdio_servers WHERE enabled = 1`,
	)
	if err != nil {
		return fmt.Errorf("failed to load stdio MCP servers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		server := &StdioServer{}
		var argsJSON, envJSON sql.NullString
		var lastError sql.NullString

		err := rows.Scan(
			&server.ID, &server.UserID, &server.Name, &server.Command, &argsJSON, &envJSON,
			&server.Enabled, &server.CreatedAt, &server.UpdatedAt, &lastError,
		)
		if err != nil {
			continue
		}

		server.LastError = lastError.String

		if argsJSON.Valid && argsJSON.String != "" {
			json.Unmarshal([]byte(argsJSON.String), &server.Args)
		}
		if envJSON.Valid && envJSON.String != "" {
			json.Unmarshal([]byte(envJSON.String), &server.Env)
		}

		// Add to client and start
		client.AddServer(server)
	}

	return nil
}

// CreateTable creates the mcp_stdio_servers table if it doesn't exist
func (r *StdioRepository) CreateTable() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS mcp_stdio_servers (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			name TEXT NOT NULL,
			command TEXT NOT NULL,
			args TEXT,
			env TEXT,
			enabled INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_error TEXT,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create mcp_stdio_servers table: %w", err)
	}

	// Create index for faster lookups
	_, err = r.db.Exec(`CREATE INDEX IF NOT EXISTS idx_mcp_stdio_servers_user_id ON mcp_stdio_servers(user_id)`)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}
