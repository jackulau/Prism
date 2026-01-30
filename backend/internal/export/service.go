package export

import (
	"archive/zip"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ExportFormat represents the output format for exports
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatCSV  ExportFormat = "csv"
	FormatZIP  ExportFormat = "zip"
)

// ExportType represents the type of data being exported
type ExportType string

const (
	ExportTypeUserData     ExportType = "user_data"
	ExportTypeAuditLogs    ExportType = "audit_logs"
	ExportTypeUsageReports ExportType = "usage_reports"
	ExportTypeGDPR         ExportType = "gdpr"
	ExportTypeCompliance   ExportType = "compliance"
)

// ExportStatus represents the current state of an export job
type ExportStatus string

const (
	StatusPending    ExportStatus = "pending"
	StatusProcessing ExportStatus = "processing"
	StatusCompleted  ExportStatus = "completed"
	StatusFailed     ExportStatus = "failed"
	StatusExpired    ExportStatus = "expired"
)

// ExportJob represents an export request and its state
type ExportJob struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	OrgID        string                 `json:"organization_id,omitempty"`
	Type         ExportType             `json:"type"`
	Format       ExportFormat           `json:"format"`
	Status       ExportStatus           `json:"status"`
	Progress     int                    `json:"progress"` // 0-100
	FilePath     string                 `json:"file_path,omitempty"`
	FileSize     int64                  `json:"file_size,omitempty"`
	DownloadURL  string                 `json:"download_url,omitempty"`
	DownloadKey  string                 `json:"-"` // Secret key for secure download
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
}

// ExportRepository is the interface for storing export jobs
type ExportRepository interface {
	Create(job *ExportJob) error
	Update(job *ExportJob) error
	GetByID(id string) (*ExportJob, error)
	GetByUserID(userID string, limit, offset int) ([]*ExportJob, error)
	GetByDownloadKey(key string) (*ExportJob, error)
	DeleteExpired() (int64, error)
}

// DataProvider is the interface for components that can provide export data
type DataProvider interface {
	GetUserData(ctx context.Context, userID string) (interface{}, error)
}

// Service handles data export operations
type Service struct {
	repo        ExportRepository
	exportDir   string
	baseURL     string
	linkExpiry  time.Duration
	maxFileAge  time.Duration

	mu         sync.Mutex
	processing map[string]context.CancelFunc
}

// NewService creates a new export service
func NewService(repo ExportRepository, exportDir, baseURL string) *Service {
	// Create export directory if it doesn't exist
	_ = os.MkdirAll(exportDir, 0755)

	return &Service{
		repo:        repo,
		exportDir:   exportDir,
		baseURL:     baseURL,
		linkExpiry:  24 * time.Hour, // Links expire in 24 hours
		maxFileAge:  7 * 24 * time.Hour, // Files deleted after 7 days
		processing:  make(map[string]context.CancelFunc),
	}
}

// CreateExportJob creates a new export job
func (s *Service) CreateExportJob(
	userID string,
	orgID string,
	exportType ExportType,
	format ExportFormat,
	params map[string]interface{},
) (*ExportJob, error) {
	job := &ExportJob{
		ID:         uuid.New().String(),
		UserID:     userID,
		OrgID:      orgID,
		Type:       exportType,
		Format:     format,
		Status:     StatusPending,
		Progress:   0,
		Parameters: params,
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.repo.Create(job); err != nil {
		return nil, fmt.Errorf("failed to create export job: %w", err)
	}

	return job, nil
}

// StartExport begins processing an export job asynchronously
func (s *Service) StartExport(ctx context.Context, jobID string, exportFunc func(ctx context.Context, job *ExportJob, progressCh chan<- int) error) error {
	job, err := s.repo.GetByID(jobID)
	if err != nil {
		return fmt.Errorf("failed to get export job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("export job not found")
	}
	if job.Status != StatusPending {
		return fmt.Errorf("export job is not pending")
	}

	// Create cancellable context
	exportCtx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	s.processing[jobID] = cancel
	s.mu.Unlock()

	// Update status
	now := time.Now().UTC()
	job.Status = StatusProcessing
	job.StartedAt = &now
	if err := s.repo.Update(job); err != nil {
		cancel()
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Process asynchronously
	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.processing, jobID)
			s.mu.Unlock()
		}()

		progressCh := make(chan int, 100)
		errCh := make(chan error, 1)

		// Progress tracker
		go func() {
			for progress := range progressCh {
				job.Progress = progress
				_ = s.repo.Update(job)
			}
		}()

		// Run export
		go func() {
			errCh <- exportFunc(exportCtx, job, progressCh)
			close(progressCh)
		}()

		// Wait for completion
		exportErr := <-errCh

		// Refresh job state
		job, _ = s.repo.GetByID(jobID)
		if job == nil {
			return
		}

		completedAt := time.Now().UTC()
		job.CompletedAt = &completedAt

		if exportErr != nil {
			job.Status = StatusFailed
			job.ErrorMessage = exportErr.Error()
		} else {
			job.Status = StatusCompleted
			job.Progress = 100

			// Generate secure download link
			if job.FilePath != "" {
				key, err := s.generateDownloadKey()
				if err == nil {
					job.DownloadKey = key
					expiry := completedAt.Add(s.linkExpiry)
					job.ExpiresAt = &expiry
					job.DownloadURL = fmt.Sprintf("%s/api/v1/exports/%s/download?key=%s", s.baseURL, job.ID, key)
				}

				// Get file size
				if info, err := os.Stat(job.FilePath); err == nil {
					job.FileSize = info.Size()
				}
			}
		}

		_ = s.repo.Update(job)
	}()

	return nil
}

// CancelExport cancels an in-progress export
func (s *Service) CancelExport(jobID string) error {
	s.mu.Lock()
	cancel, exists := s.processing[jobID]
	s.mu.Unlock()

	if !exists {
		return fmt.Errorf("export job is not processing")
	}

	cancel()

	job, err := s.repo.GetByID(jobID)
	if err != nil {
		return err
	}
	if job != nil {
		job.Status = StatusFailed
		job.ErrorMessage = "cancelled by user"
		now := time.Now().UTC()
		job.CompletedAt = &now
		return s.repo.Update(job)
	}

	return nil
}

// GetJob retrieves an export job by ID
func (s *Service) GetJob(jobID string) (*ExportJob, error) {
	return s.repo.GetByID(jobID)
}

// GetUserJobs retrieves all export jobs for a user
func (s *Service) GetUserJobs(userID string, limit, offset int) ([]*ExportJob, error) {
	return s.repo.GetByUserID(userID, limit, offset)
}

// GetDownloadReader returns a reader for downloading an export file
func (s *Service) GetDownloadReader(jobID, downloadKey string) (io.ReadCloser, string, int64, error) {
	job, err := s.repo.GetByID(jobID)
	if err != nil {
		return nil, "", 0, fmt.Errorf("failed to get export job: %w", err)
	}
	if job == nil {
		return nil, "", 0, fmt.Errorf("export job not found")
	}

	// Verify download key
	if job.DownloadKey != downloadKey {
		return nil, "", 0, fmt.Errorf("invalid download key")
	}

	// Check expiry
	if job.ExpiresAt != nil && time.Now().After(*job.ExpiresAt) {
		return nil, "", 0, fmt.Errorf("download link has expired")
	}

	// Open file
	file, err := os.Open(job.FilePath)
	if err != nil {
		return nil, "", 0, fmt.Errorf("failed to open export file: %w", err)
	}

	filename := filepath.Base(job.FilePath)
	return file, filename, job.FileSize, nil
}

// CleanupExpired removes expired export jobs and files
func (s *Service) CleanupExpired() error {
	// Delete expired jobs from database
	deleted, err := s.repo.DeleteExpired()
	if err != nil {
		return fmt.Errorf("failed to delete expired jobs: %w", err)
	}

	// Delete old files
	cutoff := time.Now().Add(-s.maxFileAge)
	err = filepath.Walk(s.exportDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(path)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to cleanup old files: %w", err)
	}

	_ = deleted // Log this if needed
	return nil
}

// generateDownloadKey creates a secure random download key
func (s *Service) generateDownloadKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// CreateFilePath returns a path for a new export file
func (s *Service) CreateFilePath(jobID string, format ExportFormat) string {
	ext := string(format)
	if format == FormatZIP {
		ext = "zip"
	}
	return filepath.Join(s.exportDir, fmt.Sprintf("%s.%s", jobID, ext))
}

// WriteJSON writes data as JSON to a file
func WriteJSON(filePath string, data interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// WriteCSV writes records to a CSV file
func WriteCSV(filePath string, headers []string, records [][]string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

// CreateZipArchive creates a zip archive with multiple files
func CreateZipArchive(zipPath string, files map[string][]byte) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for name, content := range files {
		writer, err := zipWriter.Create(name)
		if err != nil {
			return fmt.Errorf("failed to create file in zip: %w", err)
		}
		if _, err := writer.Write(content); err != nil {
			return fmt.Errorf("failed to write file in zip: %w", err)
		}
	}

	return nil
}

// AddFileToZip adds a file from disk to an existing zip writer
func AddFileToZip(zipWriter *zip.Writer, sourcePath, archiveName string) error {
	file, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("failed to create zip header: %w", err)
	}
	header.Name = archiveName
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create zip entry: %w", err)
	}

	_, err = io.Copy(writer, file)
	if err != nil {
		return fmt.Errorf("failed to copy file to zip: %w", err)
	}

	return nil
}
