package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// BuildStatus represents the status of a build
type BuildStatus string

const (
	BuildStatusPending   BuildStatus = "pending"
	BuildStatusRunning   BuildStatus = "running"
	BuildStatusSuccess   BuildStatus = "success"
	BuildStatusFailed    BuildStatus = "failed"
	BuildStatusCancelled BuildStatus = "cancelled"
)

// LogStream represents the type of log stream
type LogStream string

const (
	LogStreamStdout LogStream = "stdout"
	LogStreamStderr LogStream = "stderr"
)

// BuildHistory represents a build execution record
type BuildHistory struct {
	ID             string      `json:"id"`
	WorkspaceID    *string     `json:"workspace_id,omitempty"`
	OrgWorkspaceID *string     `json:"org_workspace_id,omitempty"`
	UserID         string      `json:"user_id"`
	Command        string      `json:"command"`
	Status         BuildStatus `json:"status"`
	ExitCode       *int        `json:"exit_code,omitempty"`
	StartedAt      time.Time   `json:"started_at"`
	CompletedAt    *time.Time  `json:"completed_at,omitempty"`
	DurationMs     *int64      `json:"duration_ms,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
}

// BuildLog represents a log entry for a build
type BuildLog struct {
	ID        string    `json:"id"`
	BuildID   string    `json:"build_id"`
	Stream    LogStream `json:"stream"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// BuildHistoryRepository handles build history database operations
type BuildHistoryRepository struct {
	db *sql.DB
}

// NewBuildHistoryRepository creates a new build history repository
func NewBuildHistoryRepository(db *sql.DB) *BuildHistoryRepository {
	return &BuildHistoryRepository{db: db}
}

// Create creates a new build history record
func (r *BuildHistoryRepository) Create(build *BuildHistory) error {
	if build.ID == "" {
		build.ID = uuid.New().String()
	}
	if build.CreatedAt.IsZero() {
		build.CreatedAt = time.Now()
	}

	_, err := r.db.Exec(
		`INSERT INTO build_history (id, workspace_id, org_workspace_id, user_id, command, status, exit_code, started_at, completed_at, duration_ms, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		build.ID, build.WorkspaceID, build.OrgWorkspaceID, build.UserID, build.Command,
		build.Status, build.ExitCode, build.StartedAt, build.CompletedAt, build.DurationMs, build.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create build history: %w", err)
	}

	return nil
}

// GetByID retrieves a build history record by ID
func (r *BuildHistoryRepository) GetByID(id string) (*BuildHistory, error) {
	build := &BuildHistory{}
	var status string
	err := r.db.QueryRow(
		`SELECT id, workspace_id, org_workspace_id, user_id, command, status, exit_code, started_at, completed_at, duration_ms, created_at
		 FROM build_history WHERE id = ?`,
		id,
	).Scan(&build.ID, &build.WorkspaceID, &build.OrgWorkspaceID, &build.UserID, &build.Command,
		&status, &build.ExitCode, &build.StartedAt, &build.CompletedAt, &build.DurationMs, &build.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get build history: %w", err)
	}

	build.Status = BuildStatus(status)
	return build, nil
}

// Update updates an existing build history record
func (r *BuildHistoryRepository) Update(build *BuildHistory) error {
	result, err := r.db.Exec(
		`UPDATE build_history SET workspace_id = ?, org_workspace_id = ?, command = ?, status = ?, exit_code = ?, started_at = ?, completed_at = ?, duration_ms = ?
		 WHERE id = ?`,
		build.WorkspaceID, build.OrgWorkspaceID, build.Command, build.Status, build.ExitCode,
		build.StartedAt, build.CompletedAt, build.DurationMs, build.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update build history: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("build history not found: %s", build.ID)
	}

	return nil
}

// Delete removes a build history record and its logs
func (r *BuildHistoryRepository) Delete(id string) error {
	result, err := r.db.Exec(`DELETE FROM build_history WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete build history: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("build history not found: %s", id)
	}

	return nil
}

// ListByUserID retrieves build history for a user with pagination
func (r *BuildHistoryRepository) ListByUserID(userID string, limit, offset int) ([]*BuildHistory, error) {
	rows, err := r.db.Query(
		`SELECT id, workspace_id, org_workspace_id, user_id, command, status, exit_code, started_at, completed_at, duration_ms, created_at
		 FROM build_history
		 WHERE user_id = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list build history: %w", err)
	}
	defer rows.Close()

	return r.scanBuildHistoryRows(rows)
}

// ListByWorkspaceID retrieves build history for a workspace with pagination
func (r *BuildHistoryRepository) ListByWorkspaceID(workspaceID string, limit, offset int) ([]*BuildHistory, error) {
	rows, err := r.db.Query(
		`SELECT id, workspace_id, org_workspace_id, user_id, command, status, exit_code, started_at, completed_at, duration_ms, created_at
		 FROM build_history
		 WHERE workspace_id = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		workspaceID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list build history by workspace: %w", err)
	}
	defer rows.Close()

	return r.scanBuildHistoryRows(rows)
}

// ListByOrgWorkspaceID retrieves build history for an org workspace with pagination
func (r *BuildHistoryRepository) ListByOrgWorkspaceID(orgWorkspaceID string, limit, offset int) ([]*BuildHistory, error) {
	rows, err := r.db.Query(
		`SELECT id, workspace_id, org_workspace_id, user_id, command, status, exit_code, started_at, completed_at, duration_ms, created_at
		 FROM build_history
		 WHERE org_workspace_id = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		orgWorkspaceID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list build history by org workspace: %w", err)
	}
	defer rows.Close()

	return r.scanBuildHistoryRows(rows)
}

// CountByUserID returns the total count of builds for a user
func (r *BuildHistoryRepository) CountByUserID(userID string) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM build_history WHERE user_id = ?`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count build history: %w", err)
	}
	return count, nil
}

// scanBuildHistoryRows scans rows into BuildHistory structs
func (r *BuildHistoryRepository) scanBuildHistoryRows(rows *sql.Rows) ([]*BuildHistory, error) {
	var builds []*BuildHistory
	for rows.Next() {
		build := &BuildHistory{}
		var status string
		err := rows.Scan(&build.ID, &build.WorkspaceID, &build.OrgWorkspaceID, &build.UserID, &build.Command,
			&status, &build.ExitCode, &build.StartedAt, &build.CompletedAt, &build.DurationMs, &build.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan build history: %w", err)
		}
		build.Status = BuildStatus(status)
		builds = append(builds, build)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating build history rows: %w", err)
	}

	return builds, nil
}

// AppendLog appends a log entry to a build
func (r *BuildHistoryRepository) AppendLog(log *BuildLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	_, err := r.db.Exec(
		`INSERT INTO build_logs (id, build_id, stream, content, timestamp)
		 VALUES (?, ?, ?, ?, ?)`,
		log.ID, log.BuildID, log.Stream, log.Content, log.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("failed to append build log: %w", err)
	}

	return nil
}

// GetLogs retrieves all logs for a build
func (r *BuildHistoryRepository) GetLogs(buildID string) ([]*BuildLog, error) {
	rows, err := r.db.Query(
		`SELECT id, build_id, stream, content, timestamp
		 FROM build_logs
		 WHERE build_id = ?
		 ORDER BY timestamp ASC`,
		buildID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get build logs: %w", err)
	}
	defer rows.Close()

	var logs []*BuildLog
	for rows.Next() {
		log := &BuildLog{}
		var stream string
		err := rows.Scan(&log.ID, &log.BuildID, &stream, &log.Content, &log.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan build log: %w", err)
		}
		log.Stream = LogStream(stream)
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating build logs: %w", err)
	}

	return logs, nil
}

// GetLogsSince retrieves logs for a build since a given timestamp (for streaming)
func (r *BuildHistoryRepository) GetLogsSince(buildID string, since time.Time) ([]*BuildLog, error) {
	rows, err := r.db.Query(
		`SELECT id, build_id, stream, content, timestamp
		 FROM build_logs
		 WHERE build_id = ? AND timestamp > ?
		 ORDER BY timestamp ASC`,
		buildID, since,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get build logs since: %w", err)
	}
	defer rows.Close()

	var logs []*BuildLog
	for rows.Next() {
		log := &BuildLog{}
		var stream string
		err := rows.Scan(&log.ID, &log.BuildID, &stream, &log.Content, &log.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan build log: %w", err)
		}
		log.Stream = LogStream(stream)
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating build logs: %w", err)
	}

	return logs, nil
}

// DeleteLogsByBuildID removes all logs for a specific build
func (r *BuildHistoryRepository) DeleteLogsByBuildID(buildID string) error {
	_, err := r.db.Exec(`DELETE FROM build_logs WHERE build_id = ?`, buildID)
	if err != nil {
		return fmt.Errorf("failed to delete build logs: %w", err)
	}
	return nil
}

// UpdateStatus is a convenience method to update only the status and related fields
func (r *BuildHistoryRepository) UpdateStatus(id string, status BuildStatus, exitCode *int, completedAt *time.Time, durationMs *int64) error {
	result, err := r.db.Exec(
		`UPDATE build_history SET status = ?, exit_code = ?, completed_at = ?, duration_ms = ?
		 WHERE id = ?`,
		status, exitCode, completedAt, durationMs, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update build status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("build history not found: %s", id)
	}

	return nil
}

// GetRunningBuilds retrieves all builds with status 'running'
func (r *BuildHistoryRepository) GetRunningBuilds(userID string) ([]*BuildHistory, error) {
	rows, err := r.db.Query(
		`SELECT id, workspace_id, org_workspace_id, user_id, command, status, exit_code, started_at, completed_at, duration_ms, created_at
		 FROM build_history
		 WHERE user_id = ? AND status = ?
		 ORDER BY started_at DESC`,
		userID, BuildStatusRunning,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get running builds: %w", err)
	}
	defer rows.Close()

	return r.scanBuildHistoryRows(rows)
}
