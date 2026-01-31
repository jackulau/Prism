package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/export"
)

// ExportJobRepository handles export job database operations
type ExportJobRepository struct {
	db *sql.DB
}

// NewExportJobRepository creates a new export job repository
func NewExportJobRepository(db *sql.DB) *ExportJobRepository {
	return &ExportJobRepository{db: db}
}

// Create stores a new export job
func (r *ExportJobRepository) Create(job *export.ExportJob) error {
	if job.ID == "" {
		job.ID = uuid.New().String()
	}

	paramsJSON := "{}"
	if job.Parameters != nil {
		if data, err := json.Marshal(job.Parameters); err == nil {
			paramsJSON = string(data)
		}
	}

	_, err := r.db.Exec(`
		INSERT INTO export_jobs (
			id, user_id, organization_id, type, format, status, progress,
			file_path, file_size, download_url, download_key, expires_at,
			error_message, parameters, created_at, started_at, completed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job.ID, job.UserID, job.OrgID, string(job.Type), string(job.Format),
		string(job.Status), job.Progress, job.FilePath, job.FileSize,
		job.DownloadURL, job.DownloadKey, job.ExpiresAt, job.ErrorMessage,
		paramsJSON, job.CreatedAt, job.StartedAt, job.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create export job: %w", err)
	}

	return nil
}

// Update updates an existing export job
func (r *ExportJobRepository) Update(job *export.ExportJob) error {
	paramsJSON := "{}"
	if job.Parameters != nil {
		if data, err := json.Marshal(job.Parameters); err == nil {
			paramsJSON = string(data)
		}
	}

	_, err := r.db.Exec(`
		UPDATE export_jobs SET
			status = ?, progress = ?, file_path = ?, file_size = ?,
			download_url = ?, download_key = ?, expires_at = ?,
			error_message = ?, parameters = ?, started_at = ?, completed_at = ?
		WHERE id = ?`,
		string(job.Status), job.Progress, job.FilePath, job.FileSize,
		job.DownloadURL, job.DownloadKey, job.ExpiresAt, job.ErrorMessage,
		paramsJSON, job.StartedAt, job.CompletedAt, job.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update export job: %w", err)
	}

	return nil
}

// GetByID retrieves an export job by ID
func (r *ExportJobRepository) GetByID(id string) (*export.ExportJob, error) {
	var job export.ExportJob
	var orgID, filePath, downloadURL, downloadKey, errorMsg, paramsJSON sql.NullString
	var fileSize sql.NullInt64
	var expiresAt, startedAt, completedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, user_id, organization_id, type, format, status, progress,
			   file_path, file_size, download_url, download_key, expires_at,
			   error_message, parameters, created_at, started_at, completed_at
		FROM export_jobs WHERE id = ?`, id,
	).Scan(
		&job.ID, &job.UserID, &orgID, &job.Type, &job.Format,
		&job.Status, &job.Progress, &filePath, &fileSize, &downloadURL,
		&downloadKey, &expiresAt, &errorMsg, &paramsJSON, &job.CreatedAt,
		&startedAt, &completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get export job: %w", err)
	}

	job.OrgID = orgID.String
	job.FilePath = filePath.String
	job.FileSize = fileSize.Int64
	job.DownloadURL = downloadURL.String
	job.DownloadKey = downloadKey.String
	job.ErrorMessage = errorMsg.String

	if expiresAt.Valid {
		job.ExpiresAt = &expiresAt.Time
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	if paramsJSON.Valid && paramsJSON.String != "" {
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(paramsJSON.String), &params); err == nil {
			job.Parameters = params
		}
	}

	return &job, nil
}

// GetByUserID retrieves all export jobs for a user
func (r *ExportJobRepository) GetByUserID(userID string, limit, offset int) ([]*export.ExportJob, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.db.Query(`
		SELECT id, user_id, organization_id, type, format, status, progress,
			   file_path, file_size, download_url, download_key, expires_at,
			   error_message, parameters, created_at, started_at, completed_at
		FROM export_jobs
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query export jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*export.ExportJob
	for rows.Next() {
		var job export.ExportJob
		var orgID, filePath, downloadURL, downloadKey, errorMsg, paramsJSON sql.NullString
		var fileSize sql.NullInt64
		var expiresAt, startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&job.ID, &job.UserID, &orgID, &job.Type, &job.Format,
			&job.Status, &job.Progress, &filePath, &fileSize, &downloadURL,
			&downloadKey, &expiresAt, &errorMsg, &paramsJSON, &job.CreatedAt,
			&startedAt, &completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan export job: %w", err)
		}

		job.OrgID = orgID.String
		job.FilePath = filePath.String
		job.FileSize = fileSize.Int64
		job.DownloadURL = downloadURL.String
		job.DownloadKey = downloadKey.String
		job.ErrorMessage = errorMsg.String

		if expiresAt.Valid {
			job.ExpiresAt = &expiresAt.Time
		}
		if startedAt.Valid {
			job.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			job.CompletedAt = &completedAt.Time
		}

		if paramsJSON.Valid && paramsJSON.String != "" {
			var params map[string]interface{}
			if err := json.Unmarshal([]byte(paramsJSON.String), &params); err == nil {
				job.Parameters = params
			}
		}

		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// GetByDownloadKey retrieves an export job by its download key
func (r *ExportJobRepository) GetByDownloadKey(key string) (*export.ExportJob, error) {
	var job export.ExportJob
	var orgID, filePath, downloadURL, downloadKey, errorMsg, paramsJSON sql.NullString
	var fileSize sql.NullInt64
	var expiresAt, startedAt, completedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, user_id, organization_id, type, format, status, progress,
			   file_path, file_size, download_url, download_key, expires_at,
			   error_message, parameters, created_at, started_at, completed_at
		FROM export_jobs WHERE download_key = ?`, key,
	).Scan(
		&job.ID, &job.UserID, &orgID, &job.Type, &job.Format,
		&job.Status, &job.Progress, &filePath, &fileSize, &downloadURL,
		&downloadKey, &expiresAt, &errorMsg, &paramsJSON, &job.CreatedAt,
		&startedAt, &completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get export job by download key: %w", err)
	}

	job.OrgID = orgID.String
	job.FilePath = filePath.String
	job.FileSize = fileSize.Int64
	job.DownloadURL = downloadURL.String
	job.DownloadKey = downloadKey.String
	job.ErrorMessage = errorMsg.String

	if expiresAt.Valid {
		job.ExpiresAt = &expiresAt.Time
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	if paramsJSON.Valid && paramsJSON.String != "" {
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(paramsJSON.String), &params); err == nil {
			job.Parameters = params
		}
	}

	return &job, nil
}

// DeleteExpired deletes export jobs that have expired
func (r *ExportJobRepository) DeleteExpired() (int64, error) {
	result, err := r.db.Exec(`
		DELETE FROM export_jobs
		WHERE expires_at IS NOT NULL AND expires_at < ?`,
		time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired export jobs: %w", err)
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return deleted, nil
}
