package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Workspace represents a user's workspace/project directory
type Workspace struct {
	ID             string
	UserID         string
	Path           string
	Name           string
	IsCurrent      bool
	LastAccessedAt *time.Time
	CreatedAt      time.Time
}

// WorkspaceRepository handles workspace database operations
type WorkspaceRepository struct {
	db *sql.DB
}

// NewWorkspaceRepository creates a new workspace repository
func NewWorkspaceRepository(db *sql.DB) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

// Create creates a new workspace entry or returns existing one if path already exists
func (r *WorkspaceRepository) Create(userID, path, name string) (*Workspace, error) {
	// Check if workspace with this path already exists for user
	existing, err := r.GetByPath(userID, path)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		// Update last accessed time
		_ = r.UpdateLastAccessed(existing.ID)
		return existing, nil
	}

	// Create new workspace
	id := uuid.New().String()
	now := time.Now()

	_, err = r.db.Exec(
		`INSERT INTO user_workspaces (id, user_id, path, name, is_current, last_accessed_at, created_at)
		 VALUES (?, ?, ?, ?, 0, ?, ?)`,
		id, userID, path, name, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	return &Workspace{
		ID:             id,
		UserID:         userID,
		Path:           path,
		Name:           name,
		IsCurrent:      false,
		LastAccessedAt: &now,
		CreatedAt:      now,
	}, nil
}

// GetByID retrieves a workspace by ID
func (r *WorkspaceRepository) GetByID(id string) (*Workspace, error) {
	workspace := &Workspace{}
	var lastAccessedAt sql.NullTime

	err := r.db.QueryRow(
		`SELECT id, user_id, path, name, is_current, last_accessed_at, created_at
		 FROM user_workspaces WHERE id = ?`,
		id,
	).Scan(&workspace.ID, &workspace.UserID, &workspace.Path, &workspace.Name,
		&workspace.IsCurrent, &lastAccessedAt, &workspace.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	if lastAccessedAt.Valid {
		workspace.LastAccessedAt = &lastAccessedAt.Time
	}

	return workspace, nil
}

// GetByPath retrieves a workspace by user ID and path
func (r *WorkspaceRepository) GetByPath(userID, path string) (*Workspace, error) {
	workspace := &Workspace{}
	var lastAccessedAt sql.NullTime

	err := r.db.QueryRow(
		`SELECT id, user_id, path, name, is_current, last_accessed_at, created_at
		 FROM user_workspaces WHERE user_id = ? AND path = ?`,
		userID, path,
	).Scan(&workspace.ID, &workspace.UserID, &workspace.Path, &workspace.Name,
		&workspace.IsCurrent, &lastAccessedAt, &workspace.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	if lastAccessedAt.Valid {
		workspace.LastAccessedAt = &lastAccessedAt.Time
	}

	return workspace, nil
}

// GetCurrent retrieves the current workspace for a user
func (r *WorkspaceRepository) GetCurrent(userID string) (*Workspace, error) {
	workspace := &Workspace{}
	var lastAccessedAt sql.NullTime

	err := r.db.QueryRow(
		`SELECT id, user_id, path, name, is_current, last_accessed_at, created_at
		 FROM user_workspaces WHERE user_id = ? AND is_current = 1`,
		userID,
	).Scan(&workspace.ID, &workspace.UserID, &workspace.Path, &workspace.Name,
		&workspace.IsCurrent, &lastAccessedAt, &workspace.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get current workspace: %w", err)
	}

	if lastAccessedAt.Valid {
		workspace.LastAccessedAt = &lastAccessedAt.Time
	}

	return workspace, nil
}

// SetCurrent sets a workspace as the current workspace for a user
func (r *WorkspaceRepository) SetCurrent(userID, workspaceID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear current status from all user's workspaces
	_, err = tx.Exec(
		`UPDATE user_workspaces SET is_current = 0 WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to clear current workspace: %w", err)
	}

	// Set the new current workspace
	now := time.Now()
	_, err = tx.Exec(
		`UPDATE user_workspaces SET is_current = 1, last_accessed_at = ? WHERE id = ? AND user_id = ?`,
		now, workspaceID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to set current workspace: %w", err)
	}

	return tx.Commit()
}

// ListRecent retrieves recent workspaces for a user, ordered by last accessed time
func (r *WorkspaceRepository) ListRecent(userID string, limit int) ([]*Workspace, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := r.db.Query(
		`SELECT id, user_id, path, name, is_current, last_accessed_at, created_at
		 FROM user_workspaces
		 WHERE user_id = ?
		 ORDER BY COALESCE(last_accessed_at, '1970-01-01') DESC, created_at DESC
		 LIMIT ?`,
		userID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*Workspace
	for rows.Next() {
		workspace := &Workspace{}
		var lastAccessedAt sql.NullTime

		err := rows.Scan(&workspace.ID, &workspace.UserID, &workspace.Path, &workspace.Name,
			&workspace.IsCurrent, &lastAccessedAt, &workspace.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}

		if lastAccessedAt.Valid {
			workspace.LastAccessedAt = &lastAccessedAt.Time
		}

		workspaces = append(workspaces, workspace)
	}

	return workspaces, nil
}

// UpdateLastAccessed updates the last accessed time for a workspace
func (r *WorkspaceRepository) UpdateLastAccessed(workspaceID string) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE user_workspaces SET last_accessed_at = ? WHERE id = ?`,
		now, workspaceID,
	)
	if err != nil {
		return fmt.Errorf("failed to update last accessed: %w", err)
	}
	return nil
}

// UpdateName updates the name of a workspace
func (r *WorkspaceRepository) UpdateName(workspaceID, name string) error {
	_, err := r.db.Exec(
		`UPDATE user_workspaces SET name = ? WHERE id = ?`,
		name, workspaceID,
	)
	if err != nil {
		return fmt.Errorf("failed to update workspace name: %w", err)
	}
	return nil
}

// Delete removes a workspace by ID
func (r *WorkspaceRepository) Delete(workspaceID string) error {
	_, err := r.db.Exec(`DELETE FROM user_workspaces WHERE id = ?`, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}
	return nil
}

// DeleteByUserID removes all workspaces for a user
func (r *WorkspaceRepository) DeleteByUserID(userID string) error {
	_, err := r.db.Exec(`DELETE FROM user_workspaces WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user workspaces: %w", err)
	}
	return nil
}
