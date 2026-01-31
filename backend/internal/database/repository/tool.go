package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Tool represents a tool in the tool catalog
type Tool struct {
	ID               string
	DisplayName      string
	SlugName         string
	Description      string
	IsModel          bool
	IsBuiltin        bool
	ProviderID       string
	ParametersSchema string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ToolRepository handles tool database operations
type ToolRepository struct {
	db *sql.DB
}

// NewToolRepository creates a new tool repository
func NewToolRepository(db *sql.DB) *ToolRepository {
	return &ToolRepository{db: db}
}

// Create creates a new tool
func (r *ToolRepository) Create(tool *Tool) error {
	if tool.ID == "" {
		tool.ID = uuid.New().String()
	}
	now := time.Now()
	tool.CreatedAt = now
	tool.UpdatedAt = now

	_, err := r.db.Exec(`
		INSERT INTO tools (id, display_name, slug_name, description, is_model, is_builtin, provider_id, parameters_schema, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, tool.ID, tool.DisplayName, tool.SlugName, tool.Description, tool.IsModel, tool.IsBuiltin, tool.ProviderID, tool.ParametersSchema, tool.CreatedAt, tool.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create tool: %w", err)
	}
	return nil
}

// GetByID retrieves a tool by ID
func (r *ToolRepository) GetByID(id string) (*Tool, error) {
	tool := &Tool{}
	var description, providerID, parametersSchema sql.NullString

	err := r.db.QueryRow(`
		SELECT id, display_name, slug_name, description, is_model, is_builtin, provider_id, parameters_schema, created_at, updated_at
		FROM tools
		WHERE id = ?
	`, id).Scan(&tool.ID, &tool.DisplayName, &tool.SlugName, &description, &tool.IsModel, &tool.IsBuiltin, &providerID, &parametersSchema, &tool.CreatedAt, &tool.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tool: %w", err)
	}

	tool.Description = description.String
	tool.ProviderID = providerID.String
	tool.ParametersSchema = parametersSchema.String

	return tool, nil
}

// GetBySlug retrieves a tool by slug name
func (r *ToolRepository) GetBySlug(slug string) (*Tool, error) {
	tool := &Tool{}
	var description, providerID, parametersSchema sql.NullString

	err := r.db.QueryRow(`
		SELECT id, display_name, slug_name, description, is_model, is_builtin, provider_id, parameters_schema, created_at, updated_at
		FROM tools
		WHERE slug_name = ?
	`, slug).Scan(&tool.ID, &tool.DisplayName, &tool.SlugName, &description, &tool.IsModel, &tool.IsBuiltin, &providerID, &parametersSchema, &tool.CreatedAt, &tool.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tool by slug: %w", err)
	}

	tool.Description = description.String
	tool.ProviderID = providerID.String
	tool.ParametersSchema = parametersSchema.String

	return tool, nil
}

// List retrieves all tools
func (r *ToolRepository) List() ([]*Tool, error) {
	rows, err := r.db.Query(`
		SELECT id, display_name, slug_name, description, is_model, is_builtin, provider_id, parameters_schema, created_at, updated_at
		FROM tools
		ORDER BY display_name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}
	defer rows.Close()

	return r.scanTools(rows)
}

// ListByProvider retrieves tools by provider ID
func (r *ToolRepository) ListByProvider(providerID string) ([]*Tool, error) {
	rows, err := r.db.Query(`
		SELECT id, display_name, slug_name, description, is_model, is_builtin, provider_id, parameters_schema, created_at, updated_at
		FROM tools
		WHERE provider_id = ?
		ORDER BY display_name ASC
	`, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools by provider: %w", err)
	}
	defer rows.Close()

	return r.scanTools(rows)
}

// ListModels retrieves all tools that are models (is_model = true)
func (r *ToolRepository) ListModels() ([]*Tool, error) {
	rows, err := r.db.Query(`
		SELECT id, display_name, slug_name, description, is_model, is_builtin, provider_id, parameters_schema, created_at, updated_at
		FROM tools
		WHERE is_model = 1
		ORDER BY display_name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer rows.Close()

	return r.scanTools(rows)
}

// Update updates an existing tool
func (r *ToolRepository) Update(tool *Tool) error {
	tool.UpdatedAt = time.Now()

	result, err := r.db.Exec(`
		UPDATE tools
		SET display_name = ?, slug_name = ?, description = ?, is_model = ?, provider_id = ?, parameters_schema = ?, updated_at = ?
		WHERE id = ?
	`, tool.DisplayName, tool.SlugName, tool.Description, tool.IsModel, tool.ProviderID, tool.ParametersSchema, tool.UpdatedAt, tool.ID)

	if err != nil {
		return fmt.Errorf("failed to update tool: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tool not found")
	}

	return nil
}

// Delete deletes a tool by ID
func (r *ToolRepository) Delete(id string) error {
	result, err := r.db.Exec(`DELETE FROM tools WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete tool: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tool not found")
	}

	return nil
}

// scanTools scans rows into a slice of Tool pointers
func (r *ToolRepository) scanTools(rows *sql.Rows) ([]*Tool, error) {
	var tools []*Tool

	for rows.Next() {
		tool := &Tool{}
		var description, providerID, parametersSchema sql.NullString

		err := rows.Scan(&tool.ID, &tool.DisplayName, &tool.SlugName, &description, &tool.IsModel, &tool.IsBuiltin, &providerID, &parametersSchema, &tool.CreatedAt, &tool.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tool: %w", err)
		}

		tool.Description = description.String
		tool.ProviderID = providerID.String
		tool.ParametersSchema = parametersSchema.String

		tools = append(tools, tool)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tools: %w", err)
	}

	return tools, nil
}
