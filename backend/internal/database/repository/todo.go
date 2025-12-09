package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Todo represents a task item
type Todo struct {
	ID            string
	UserID        string
	WorkspacePath string
	Content       string
	ActiveForm    string
	Status        string // "pending", "in_progress", "completed"
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// TodoRepository handles todo database operations
type TodoRepository struct {
	db *sql.DB
}

// NewTodoRepository creates a new todo repository
func NewTodoRepository(db *sql.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

// ReplaceAll replaces all todos for a user/workspace with new ones
func (r *TodoRepository) ReplaceAll(userID, workspacePath string, todos []Todo) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing todos for this user/workspace
	_, err = tx.Exec(`DELETE FROM workspace_todos WHERE user_id = ? AND workspace_path = ?`, userID, workspacePath)
	if err != nil {
		return err
	}

	// Insert new todos
	stmt, err := tx.Prepare(`
		INSERT INTO workspace_todos (id, user_id, workspace_path, content, active_form, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for _, todo := range todos {
		id := todo.ID
		if id == "" {
			id = uuid.New().String()
		}
		_, err = stmt.Exec(id, userID, workspacePath, todo.Content, todo.ActiveForm, todo.Status, now, now)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetAll retrieves all todos for a user/workspace
func (r *TodoRepository) GetAll(userID, workspacePath string) ([]*Todo, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, workspace_path, content, active_form, status, created_at, updated_at
		FROM workspace_todos
		WHERE user_id = ? AND workspace_path = ?
		ORDER BY created_at ASC
	`, userID, workspacePath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []*Todo
	for rows.Next() {
		todo := &Todo{}
		err := rows.Scan(&todo.ID, &todo.UserID, &todo.WorkspacePath, &todo.Content, &todo.ActiveForm, &todo.Status, &todo.CreatedAt, &todo.UpdatedAt)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, rows.Err()
}

// GetByUserID retrieves all todos for a user across all workspaces
func (r *TodoRepository) GetByUserID(userID string) ([]*Todo, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, workspace_path, content, active_form, status, created_at, updated_at
		FROM workspace_todos
		WHERE user_id = ?
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []*Todo
	for rows.Next() {
		todo := &Todo{}
		err := rows.Scan(&todo.ID, &todo.UserID, &todo.WorkspacePath, &todo.Content, &todo.ActiveForm, &todo.Status, &todo.CreatedAt, &todo.UpdatedAt)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, rows.Err()
}

// DeleteAll removes all todos for a user/workspace
func (r *TodoRepository) DeleteAll(userID, workspacePath string) error {
	_, err := r.db.Exec(`DELETE FROM workspace_todos WHERE user_id = ? AND workspace_path = ?`, userID, workspacePath)
	return err
}
