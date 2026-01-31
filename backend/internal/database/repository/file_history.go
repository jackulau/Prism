package repository

import (
	"database/sql"
	"fmt"
	"strings"
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
	// Attribution fields
	AgentID     *string `json:"agent_id,omitempty"`    // Which agent made the change
	AgentName   *string `json:"agent_name,omitempty"`  // Human-readable agent name
	ToolName    *string `json:"tool_name,omitempty"`   // Which tool was used
	MessageID   *string `json:"message_id,omitempty"`  // Related conversation message
	Description *string `json:"description,omitempty"` // User/agent-provided description
}

// VersionDiff represents the differences between two file history versions
type VersionDiff struct {
	HistoryID1 string     `json:"history_id_1"`
	HistoryID2 string     `json:"history_id_2"`
	FilePath   string     `json:"file_path"`
	Additions  int        `json:"additions"`
	Deletions  int        `json:"deletions"`
	Changes    []DiffLine `json:"changes"`
}

// DiffLine represents a single line in a diff
type DiffLine struct {
	Type    string `json:"type"` // "add", "delete", "unchanged"
	Content string `json:"content"`
	OldLine int    `json:"old_line,omitempty"`
	NewLine int    `json:"new_line,omitempty"`
}

// FileHistoryStats represents statistics about file history for a user
type FileHistoryStats struct {
	TotalEntries   int    `json:"total_entries"`
	TotalFiles     int    `json:"total_files"`
	TotalSizeBytes int64  `json:"total_size_bytes"`
	OldestEntry    string `json:"oldest_entry"`
	NewestEntry    string `json:"newest_entry"`
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
		`SELECT id, user_id, file_path, content, operation, created_at,
		        agent_id, agent_name, tool_name, message_id, description
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
		err := rows.Scan(&h.ID, &h.UserID, &h.FilePath, &h.Content, &h.Operation, &h.CreatedAt,
			&h.AgentID, &h.AgentName, &h.ToolName, &h.MessageID, &h.Description)
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
		`SELECT id, user_id, file_path, content, operation, created_at,
		        agent_id, agent_name, tool_name, message_id, description
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
		err := rows.Scan(&h.ID, &h.UserID, &h.FilePath, &h.Content, &h.Operation, &h.CreatedAt,
			&h.AgentID, &h.AgentName, &h.ToolName, &h.MessageID, &h.Description)
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
		`SELECT id, user_id, file_path, content, operation, created_at,
		        agent_id, agent_name, tool_name, message_id, description
		 FROM file_history WHERE id = ?`,
		id,
	).Scan(&h.ID, &h.UserID, &h.FilePath, &h.Content, &h.Operation, &h.CreatedAt,
		&h.AgentID, &h.AgentName, &h.ToolName, &h.MessageID, &h.Description)

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
		`SELECT id, user_id, file_path, content, operation, created_at,
		        agent_id, agent_name, tool_name, message_id, description
		 FROM file_history
		 WHERE user_id = ? AND file_path = ?
		 ORDER BY created_at DESC
		 LIMIT 1`,
		userID, filePath,
	).Scan(&h.ID, &h.UserID, &h.FilePath, &h.Content, &h.Operation, &h.CreatedAt,
		&h.AgentID, &h.AgentName, &h.ToolName, &h.MessageID, &h.Description)

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

// CreateWithAttribution creates a new file history entry with full attribution metadata
func (r *FileHistoryRepository) CreateWithAttribution(entry FileHistory) error {
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	_, err := r.db.Exec(
		`INSERT INTO file_history (id, user_id, file_path, content, operation, created_at,
		                           agent_id, agent_name, tool_name, message_id, description)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.ID, entry.UserID, entry.FilePath, entry.Content, entry.Operation, entry.CreatedAt,
		entry.AgentID, entry.AgentName, entry.ToolName, entry.MessageID, entry.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to create file history with attribution: %w", err)
	}

	return nil
}

// ListByTimeRange retrieves file history entries within a date range
func (r *FileHistoryRepository) ListByTimeRange(userID string, startTime, endTime time.Time, limit, offset int) ([]*FileHistory, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, file_path, content, operation, created_at,
		        agent_id, agent_name, tool_name, message_id, description
		 FROM file_history
		 WHERE user_id = ? AND created_at >= ? AND created_at <= ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		userID, startTime, endTime, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list file history by time range: %w", err)
	}
	defer rows.Close()

	var history []*FileHistory
	for rows.Next() {
		h := &FileHistory{}
		err := rows.Scan(&h.ID, &h.UserID, &h.FilePath, &h.Content, &h.Operation, &h.CreatedAt,
			&h.AgentID, &h.AgentName, &h.ToolName, &h.MessageID, &h.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file history: %w", err)
		}
		history = append(history, h)
	}

	return history, nil
}

// ListByAgent retrieves all file history entries made by a specific agent
func (r *FileHistoryRepository) ListByAgent(userID, agentID string, limit, offset int) ([]*FileHistory, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, file_path, content, operation, created_at,
		        agent_id, agent_name, tool_name, message_id, description
		 FROM file_history
		 WHERE user_id = ? AND agent_id = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		userID, agentID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list file history by agent: %w", err)
	}
	defer rows.Close()

	var history []*FileHistory
	for rows.Next() {
		h := &FileHistory{}
		err := rows.Scan(&h.ID, &h.UserID, &h.FilePath, &h.Content, &h.Operation, &h.CreatedAt,
			&h.AgentID, &h.AgentName, &h.ToolName, &h.MessageID, &h.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file history: %w", err)
		}
		history = append(history, h)
	}

	return history, nil
}

// GetStats returns statistics about file history for a user
func (r *FileHistoryRepository) GetStats(userID string) (*FileHistoryStats, error) {
	stats := &FileHistoryStats{}

	// Get total entries and total size
	err := r.db.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(LENGTH(content)), 0) FROM file_history WHERE user_id = ?`,
		userID,
	).Scan(&stats.TotalEntries, &stats.TotalSizeBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to get file history stats: %w", err)
	}

	// Get distinct files count
	err = r.db.QueryRow(
		`SELECT COUNT(DISTINCT file_path) FROM file_history WHERE user_id = ?`,
		userID,
	).Scan(&stats.TotalFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct files count: %w", err)
	}

	// Get oldest entry timestamp (as string for SQLite compatibility)
	var oldestTimeStr sql.NullString
	err = r.db.QueryRow(
		`SELECT MIN(created_at) FROM file_history WHERE user_id = ?`,
		userID,
	).Scan(&oldestTimeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get oldest entry: %w", err)
	}
	if oldestTimeStr.Valid && oldestTimeStr.String != "" {
		if t, parseErr := time.Parse("2006-01-02 15:04:05.999999999-07:00", oldestTimeStr.String); parseErr == nil {
			stats.OldestEntry = t.Format(time.RFC3339)
		} else if t, parseErr := time.Parse("2006-01-02T15:04:05Z", oldestTimeStr.String); parseErr == nil {
			stats.OldestEntry = t.Format(time.RFC3339)
		} else if t, parseErr := time.Parse(time.RFC3339, oldestTimeStr.String); parseErr == nil {
			stats.OldestEntry = t.Format(time.RFC3339)
		} else {
			stats.OldestEntry = oldestTimeStr.String
		}
	}

	// Get newest entry timestamp (as string for SQLite compatibility)
	var newestTimeStr sql.NullString
	err = r.db.QueryRow(
		`SELECT MAX(created_at) FROM file_history WHERE user_id = ?`,
		userID,
	).Scan(&newestTimeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get newest entry: %w", err)
	}
	if newestTimeStr.Valid && newestTimeStr.String != "" {
		if t, parseErr := time.Parse("2006-01-02 15:04:05.999999999-07:00", newestTimeStr.String); parseErr == nil {
			stats.NewestEntry = t.Format(time.RFC3339)
		} else if t, parseErr := time.Parse("2006-01-02T15:04:05Z", newestTimeStr.String); parseErr == nil {
			stats.NewestEntry = t.Format(time.RFC3339)
		} else if t, parseErr := time.Parse(time.RFC3339, newestTimeStr.String); parseErr == nil {
			stats.NewestEntry = t.Format(time.RFC3339)
		} else {
			stats.NewestEntry = newestTimeStr.String
		}
	}

	return stats, nil
}

// GetVersionDiff compares two history versions and returns the differences
func (r *FileHistoryRepository) GetVersionDiff(userID, historyID1, historyID2 string) (*VersionDiff, error) {
	// Get the first history entry
	h1, err := r.GetByID(historyID1)
	if err != nil {
		return nil, fmt.Errorf("failed to get history 1: %w", err)
	}
	if h1 == nil {
		return nil, fmt.Errorf("history entry 1 not found")
	}
	if h1.UserID != userID {
		return nil, fmt.Errorf("history entry 1 not found")
	}

	// Get the second history entry
	h2, err := r.GetByID(historyID2)
	if err != nil {
		return nil, fmt.Errorf("failed to get history 2: %w", err)
	}
	if h2 == nil {
		return nil, fmt.Errorf("history entry 2 not found")
	}
	if h2.UserID != userID {
		return nil, fmt.Errorf("history entry 2 not found")
	}

	// Compute the diff
	diff := computeDiff(h1.Content, h2.Content)
	diff.HistoryID1 = historyID1
	diff.HistoryID2 = historyID2
	diff.FilePath = h1.FilePath

	return diff, nil
}

// computeDiff computes a simple line-by-line diff between two strings
func computeDiff(content1, content2 string) *VersionDiff {
	lines1 := strings.Split(content1, "\n")
	lines2 := strings.Split(content2, "\n")

	diff := &VersionDiff{
		Changes: make([]DiffLine, 0),
	}

	// Simple LCS-based diff algorithm
	m, n := len(lines1), len(lines2)

	// Create LCS matrix
	lcs := make([][]int, m+1)
	for i := range lcs {
		lcs[i] = make([]int, n+1)
	}

	// Fill LCS matrix
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if lines1[i-1] == lines2[j-1] {
				lcs[i][j] = lcs[i-1][j-1] + 1
			} else {
				if lcs[i-1][j] > lcs[i][j-1] {
					lcs[i][j] = lcs[i-1][j]
				} else {
					lcs[i][j] = lcs[i][j-1]
				}
			}
		}
	}

	// Backtrack to find the diff
	i, j := m, n
	var changes []DiffLine

	for i > 0 || j > 0 {
		if i > 0 && j > 0 && lines1[i-1] == lines2[j-1] {
			changes = append(changes, DiffLine{
				Type:    "unchanged",
				Content: lines1[i-1],
				OldLine: i,
				NewLine: j,
			})
			i--
			j--
		} else if j > 0 && (i == 0 || lcs[i][j-1] >= lcs[i-1][j]) {
			changes = append(changes, DiffLine{
				Type:    "add",
				Content: lines2[j-1],
				NewLine: j,
			})
			diff.Additions++
			j--
		} else if i > 0 {
			changes = append(changes, DiffLine{
				Type:    "delete",
				Content: lines1[i-1],
				OldLine: i,
			})
			diff.Deletions++
			i--
		}
	}

	// Reverse the changes to get them in order
	for k := len(changes) - 1; k >= 0; k-- {
		diff.Changes = append(diff.Changes, changes[k])
	}

	return diff
}
