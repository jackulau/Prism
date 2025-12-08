package sandbox

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/config"
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

// Build represents a build process
type Build struct {
	ID         string
	UserID     string
	WorkDir    string
	Command    string
	Args       []string
	Status     BuildStatus
	StartTime  time.Time
	EndTime    *time.Time
	Output     []string
	Error      string
	PreviewURL string
	cancel     context.CancelFunc
	mu         sync.Mutex
}

// OutputLine represents a line of build output
type OutputLine struct {
	Content   string
	Stream    string // "stdout" or "stderr"
	Timestamp time.Time
}

// OutputHandler is called for each line of output
type OutputHandler func(line OutputLine)

// FileInfo represents information about a file
type FileInfo struct {
	Name        string     `json:"name"`
	Path        string     `json:"path"`
	IsDirectory bool       `json:"is_directory"`
	Children    []FileInfo `json:"children,omitempty"`
	Size        int64      `json:"size,omitempty"`
	Modified    int64      `json:"modified,omitempty"`
}

// Service manages sandbox environments
type Service struct {
	config       *config.Config
	builds       map[string]*Build
	userWorkDirs map[string]string
	mu           sync.RWMutex
	baseDir      string
}

// NewService creates a new sandbox service
func NewService(cfg *config.Config) (*Service, error) {
	baseDir := filepath.Join(cfg.UploadDir, "sandboxes")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sandbox directory: %w", err)
	}

	return &Service{
		config:       cfg,
		builds:       make(map[string]*Build),
		userWorkDirs: make(map[string]string),
		baseDir:      baseDir,
	}, nil
}

// GetOrCreateWorkDir gets or creates a working directory for a user
func (s *Service) GetOrCreateWorkDir(userID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if dir, ok := s.userWorkDirs[userID]; ok {
		return dir, nil
	}

	dir := filepath.Join(s.baseDir, userID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create user sandbox directory: %w", err)
	}

	s.userWorkDirs[userID] = dir
	return dir, nil
}

// SetWorkDir sets a custom working directory for a user
func (s *Service) SetWorkDir(userID string, dir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify directory exists and is accessible
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("failed to access directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory")
	}

	s.userWorkDirs[userID] = dir
	return nil
}

// StartBuild starts a new build process
func (s *Service) StartBuild(userID, command string, args []string, outputHandler OutputHandler) (*Build, error) {
	workDir, err := s.GetOrCreateWorkDir(userID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.config.SandboxTimeout)

	build := &Build{
		ID:        uuid.New().String(),
		UserID:    userID,
		WorkDir:   workDir,
		Command:   command,
		Args:      args,
		Status:    BuildStatusRunning,
		StartTime: time.Now(),
		Output:    make([]string, 0),
		cancel:    cancel,
	}

	s.mu.Lock()
	s.builds[build.ID] = build
	s.mu.Unlock()

	go s.runBuild(ctx, build, outputHandler)

	return build, nil
}

// runBuild executes the build command
func (s *Service) runBuild(ctx context.Context, build *Build, outputHandler OutputHandler) {
	defer build.cancel()

	cmd := exec.CommandContext(ctx, build.Command, build.Args...)
	cmd.Dir = build.WorkDir

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		s.finishBuild(build, BuildStatusFailed, fmt.Sprintf("failed to create stdout pipe: %v", err))
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		s.finishBuild(build, BuildStatusFailed, fmt.Sprintf("failed to create stderr pipe: %v", err))
		return
	}

	if err := cmd.Start(); err != nil {
		s.finishBuild(build, BuildStatusFailed, fmt.Sprintf("failed to start command: %v", err))
		return
	}

	// Read output concurrently
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		s.readOutput(build, stdout, "stdout", outputHandler)
	}()

	go func() {
		defer wg.Done()
		s.readOutput(build, stderr, "stderr", outputHandler)
	}()

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			s.finishBuild(build, BuildStatusFailed, "build timed out")
		} else if ctx.Err() == context.Canceled {
			s.finishBuild(build, BuildStatusCancelled, "build cancelled")
		} else {
			s.finishBuild(build, BuildStatusFailed, fmt.Sprintf("build failed: %v", err))
		}
		return
	}

	s.finishBuild(build, BuildStatusSuccess, "")
}

// readOutput reads from a pipe and sends to the output handler
func (s *Service) readOutput(build *Build, r io.Reader, stream string, handler OutputHandler) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		build.mu.Lock()
		build.Output = append(build.Output, fmt.Sprintf("[%s] %s", stream, line))
		build.mu.Unlock()

		if handler != nil {
			handler(OutputLine{
				Content:   line,
				Stream:    stream,
				Timestamp: time.Now(),
			})
		}
	}
}

// finishBuild marks a build as finished
func (s *Service) finishBuild(build *Build, status BuildStatus, errorMsg string) {
	build.mu.Lock()
	defer build.mu.Unlock()

	now := time.Now()
	build.Status = status
	build.EndTime = &now
	build.Error = errorMsg
}

// StopBuild stops a running build
func (s *Service) StopBuild(buildID string) error {
	s.mu.RLock()
	build, ok := s.builds[buildID]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("build not found: %s", buildID)
	}

	if build.cancel != nil {
		build.cancel()
	}

	return nil
}

// GetBuild gets a build by ID
func (s *Service) GetBuild(buildID string) (*Build, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	build, ok := s.builds[buildID]
	if !ok {
		return nil, fmt.Errorf("build not found: %s", buildID)
	}

	return build, nil
}

// ListFiles lists files in a user's sandbox directory
func (s *Service) ListFiles(userID string) ([]FileInfo, error) {
	workDir, err := s.GetOrCreateWorkDir(userID)
	if err != nil {
		return nil, err
	}

	return s.walkDirectory(workDir, "")
}

// walkDirectory recursively walks a directory and returns file info
func (s *Service) walkDirectory(baseDir, relativePath string) ([]FileInfo, error) {
	fullPath := filepath.Join(baseDir, relativePath)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	files := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		// Skip hidden files and common ignore patterns
		if strings.HasPrefix(entry.Name(), ".") || entry.Name() == "node_modules" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			log.Printf("failed to get file info for %s: %v", entry.Name(), err)
			continue
		}

		filePath := filepath.Join(relativePath, entry.Name())
		fileInfo := FileInfo{
			Name:        entry.Name(),
			Path:        filePath,
			IsDirectory: entry.IsDir(),
			Size:        info.Size(),
			Modified:    info.ModTime().Unix(),
		}

		if entry.IsDir() {
			children, err := s.walkDirectory(baseDir, filePath)
			if err != nil {
				log.Printf("failed to walk directory %s: %v", filePath, err)
			} else {
				fileInfo.Children = children
			}
		}

		files = append(files, fileInfo)
	}

	return files, nil
}

// validateSandboxPath validates that a path is safely within the sandbox work directory
// This resolves symlinks to prevent directory traversal attacks
func (s *Service) validateSandboxPath(workDir, requestedPath string) (string, error) {
	// Sanitize the path to prevent obvious directory traversal
	cleanPath := filepath.Clean(requestedPath)
	if strings.HasPrefix(cleanPath, "..") {
		return "", fmt.Errorf("invalid file path")
	}

	fullPath := filepath.Join(workDir, cleanPath)

	// Resolve the work directory to handle any symlinks in its path
	resolvedWorkDir, err := filepath.EvalSymlinks(workDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve work directory: %w", err)
	}

	// Check if file/directory exists to resolve symlinks
	if _, statErr := os.Lstat(fullPath); statErr == nil {
		// Path exists, resolve any symlinks
		resolvedPath, err := filepath.EvalSymlinks(fullPath)
		if err != nil {
			return "", fmt.Errorf("failed to resolve path: %w", err)
		}

		// Verify the resolved path is within the resolved work directory
		if !strings.HasPrefix(resolvedPath, resolvedWorkDir) {
			return "", fmt.Errorf("invalid file path: path escapes sandbox")
		}
		return resolvedPath, nil
	}

	// Path doesn't exist yet (for writes) - verify the parent directory
	parentPath := filepath.Dir(fullPath)
	if parentInfo, statErr := os.Lstat(parentPath); statErr == nil {
		if parentInfo.Mode()&os.ModeSymlink != 0 {
			resolvedParent, err := filepath.EvalSymlinks(parentPath)
			if err != nil {
				return "", fmt.Errorf("failed to resolve parent path: %w", err)
			}
			if !strings.HasPrefix(resolvedParent, resolvedWorkDir) {
				return "", fmt.Errorf("invalid file path: parent escapes sandbox")
			}
		}
	}

	// For non-existent paths, also check the string prefix as a fallback
	if !strings.HasPrefix(fullPath, workDir) {
		return "", fmt.Errorf("invalid file path")
	}

	return fullPath, nil
}

// GetFileContent gets the content of a file
func (s *Service) GetFileContent(userID, filePath string) (string, error) {
	workDir, err := s.GetOrCreateWorkDir(userID)
	if err != nil {
		return "", err
	}

	// Validate path is within sandbox (handles symlink attacks)
	safePath, err := s.validateSandboxPath(workDir, filePath)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(safePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// WriteFile writes content to a file in the sandbox
func (s *Service) WriteFile(userID, filePath, content string) error {
	workDir, err := s.GetOrCreateWorkDir(userID)
	if err != nil {
		return err
	}

	// Validate path is within sandbox (handles symlink attacks)
	safePath, err := s.validateSandboxPath(workDir, filePath)
	if err != nil {
		return err
	}

	// Create parent directories if needed
	dir := filepath.Dir(safePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(safePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// DeleteFile deletes a file from the sandbox
func (s *Service) DeleteFile(userID, filePath string) error {
	workDir, err := s.GetOrCreateWorkDir(userID)
	if err != nil {
		return err
	}

	// Validate path is within sandbox (handles symlink attacks)
	safePath, err := s.validateSandboxPath(workDir, filePath)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(safePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetPreviewServer returns the URL for a preview server for a user
func (s *Service) GetPreviewServer(userID string) string {
	// Use configured preview URL if set, otherwise use frontend URL
	baseURL := s.config.SandboxPreviewURL
	if baseURL == "" {
		baseURL = s.config.FrontendURL
	}
	return fmt.Sprintf("%s/preview/%s", baseURL, userID)
}
