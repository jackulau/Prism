package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FileHistory represents a historical version of a file
type FileHistory struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	FilePath  string    `json:"file_path"`
	Content   string    `json:"content"`
	Operation string    `json:"operation"` // "create", "update", "delete"
	CreatedAt time.Time `json:"created_at"`
}

// FileHistoryRepository handles file history database operations
type FileHistoryRepository struct {
	db *sql.DB
}

// NewFileHistoryRepository creates a new file history repository
func NewFileHistoryRepository(db *sql.DB) *FileHistoryRepository {
	return &FileHistoryRepository{db: db}
}

// Create creates a new file history entry
func (r *FileHistoryRepository) Create(userID, filePath, content, operation string) (*FileHistory, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := r.db.Exec(
		`INSERT INTO file_history (id, user_id, file_path, content, operation, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		id, userID, filePath, content, operation, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create file history: %w", err)
	}

	return &FileHistory{
		ID:        id,
		UserID:    userID,
		FilePath:  filePath,
		Content:   content,
		Operation: operation,
		CreatedAt: now,
	}, nil
}

// ListByFilePath retrieves file history for a specific file
func (r *FileHistoryRepository) ListByFilePath(userID, filePath string, limit int) ([]*FileHistory, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, file_path, content, operation, created_at
		 FROM file_history
		 WHERE user_id = ? AND file_path = ?
		 ORDER BY created_at DESC
		 LIMIT ?`,
		userID, filePath, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list file history: %w", err)
	}
	defer rows.Close()

	var history []*FileHistory
	for rows.Next() {
		h := &FileHistory{}
		err := rows.Scan(&h.ID, &h.UserID, &h.FilePath, &h.Content, &h.Operation, &h.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file history: %w", err)
		}
		history = append(history, h)
	}

	return history, nil
}

// ListByUserID retrieves all file history entries for a user
func (r *FileHistoryRepository) ListByUserID(userID string, limit, offset int) ([]*FileHistory, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, file_path, content, operation, created_at
		 FROM file_history
		 WHERE user_id = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list file history: %w", err)
	}
	defer rows.Close()

	var history []*FileHistory
	for rows.Next() {
		h := &FileHistory{}
		err := rows.Scan(&h.ID, &h.UserID, &h.FilePath, &h.Content, &h.Operation, &h.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file history: %w", err)
		}
		history = append(history, h)
	}

	return history, nil
}

// GetByID retrieves a specific file history entry
func (r *FileHistoryRepository) GetByID(id string) (*FileHistory, error) {
	h := &FileHistory{}
	err := r.db.QueryRow(
		`SELECT id, user_id, file_path, content, operation, created_at
		 FROM file_history WHERE id = ?`,
		id,
	).Scan(&h.ID, &h.UserID, &h.FilePath, &h.Content, &h.Operation, &h.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get file history: %w", err)
	}

	return h, nil
}

// GetLatestByFilePath gets the most recent history entry for a file
func (r *FileHistoryRepository) GetLatestByFilePath(userID, filePath string) (*FileHistory, error) {
	h := &FileHistory{}
	err := r.db.QueryRow(
		`SELECT id, user_id, file_path, content, operation, created_at
		 FROM file_history
		 WHERE user_id = ? AND file_path = ?
		 ORDER BY created_at DESC
		 LIMIT 1`,
		userID, filePath,
	).Scan(&h.ID, &h.UserID, &h.FilePath, &h.Content, &h.Operation, &h.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest file history: %w", err)
	}

	return h, nil
}

// DeleteOldEntries removes history entries older than the specified duration
// Keeps at least minKeep entries per file
func (r *FileHistoryRepository) DeleteOldEntries(userID string, olderThan time.Duration, minKeep int) error {
	cutoff := time.Now().Add(-olderThan)

	// Delete old entries but keep at least minKeep per file
	_, err := r.db.Exec(`
		DELETE FROM file_history
		WHERE user_id = ?
		AND created_at < ?
		AND id NOT IN (
			SELECT id FROM (
				SELECT id, file_path,
				ROW_NUMBER() OVER (PARTITION BY file_path ORDER BY created_at DESC) as rn
				FROM file_history WHERE user_id = ?
			) WHERE rn <= ?
		)`,
		userID, cutoff, userID, minKeep,
	)
	if err != nil {
		return fmt.Errorf("failed to delete old file history: %w", err)
	}

	return nil
}

// GetDistinctFiles returns all unique file paths with history for a user
func (r *FileHistoryRepository) GetDistinctFiles(userID string) ([]string, error) {
	rows, err := r.db.Query(
		`SELECT DISTINCT file_path FROM file_history WHERE user_id = ? ORDER BY file_path`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct files: %w", err)
	}
	defer rows.Close()

	var files []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, fmt.Errorf("failed to scan file path: %w", err)
		}
		files = append(files, path)
	}

	return files, nil
}
