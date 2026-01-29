package builtin

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

const (
	// MaxFileSize is the maximum file size in bytes (10MB)
	MaxFileSize int64 = 10 * 1024 * 1024
)

// FileContent represents the structured response for file reading
type FileContent struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Size     int64  `json:"size"`
	Encoding string `json:"encoding"` // "utf-8" or "base64"
}

// ReadFileSandboxTool reads file contents from a repository using the sandbox
type ReadFileSandboxTool struct {
	sandbox *sandbox.Service
}

// NewReadFileSandboxTool creates a new sandbox file read tool
func NewReadFileSandboxTool(sandbox *sandbox.Service) *ReadFileSandboxTool {
	return &ReadFileSandboxTool{sandbox: sandbox}
}

func (t *ReadFileSandboxTool) Name() string {
	return "sandbox_read_file"
}

func (t *ReadFileSandboxTool) Description() string {
	return "Read the contents of a file in the repository. Returns the file content along with metadata including size and encoding."
}

func (t *ReadFileSandboxTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"relativeFilePath": {
				Type:        "string",
				Description: "Relative path to the file within the repository (e.g., 'src/main.py' or 'config/settings.json')",
			},
		},
		Required: []string{"relativeFilePath"},
	}
}

func (t *ReadFileSandboxTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	relativeFilePath, ok := params["relativeFilePath"].(string)
	if !ok || relativeFilePath == "" {
		return nil, fmt.Errorf("relativeFilePath parameter is required")
	}

	// Validate path for security
	if err := t.validatePath(relativeFilePath); err != nil {
		return nil, err
	}

	// Get the user's work directory
	workDir, err := t.sandbox.GetOrCreateWorkDir(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get work directory: %w", err)
	}

	// Construct the full path
	fullPath := filepath.Join(workDir, relativeFilePath)

	// Get file info to check size
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", relativeFilePath)
		}
		return nil, fmt.Errorf("failed to access file: %w", err)
	}

	if fileInfo.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file: %s", relativeFilePath)
	}

	// Check file size limit
	if fileInfo.Size() > MaxFileSize {
		return nil, fmt.Errorf("file exceeds maximum size limit of %d bytes (file size: %d bytes)", MaxFileSize, fileInfo.Size())
	}

	// Read file content using sandbox service
	content, err := t.sandbox.GetFileContent(userID, relativeFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Detect if content is binary
	encoding := "utf-8"
	outputContent := content

	if isBinary([]byte(content)) {
		encoding = "base64"
		outputContent = base64.StdEncoding.EncodeToString([]byte(content))
	}

	return FileContent{
		Path:     relativeFilePath,
		Content:  outputContent,
		Size:     fileInfo.Size(),
		Encoding: encoding,
	}, nil
}

func (t *ReadFileSandboxTool) RequiresConfirmation() bool {
	return false // Reading files is a read-only operation
}

// validatePath validates that the path is safe (no directory traversal, no absolute paths)
func (t *ReadFileSandboxTool) validatePath(path string) error {
	// Reject empty paths
	if path == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Reject absolute paths
	if filepath.IsAbs(path) {
		return fmt.Errorf("absolute paths are not allowed")
	}

	// Clean the path and check for directory traversal
	cleanPath := filepath.Clean(path)

	// Reject paths that try to escape the root
	if strings.HasPrefix(cleanPath, "..") {
		return fmt.Errorf("directory traversal is not allowed")
	}

	// Check for ".." components anywhere in the path
	parts := strings.Split(cleanPath, string(filepath.Separator))
	for _, part := range parts {
		if part == ".." {
			return fmt.Errorf("directory traversal is not allowed")
		}
	}

	return nil
}

// isBinary detects if the content is binary by checking for null bytes
// and non-printable characters
func isBinary(content []byte) bool {
	// Check for null bytes (common indicator of binary files)
	if bytes.Contains(content, []byte{0}) {
		return true
	}

	// Check a sample of the content for non-text characters
	// Only check first 8KB for performance
	sampleSize := 8192
	if len(content) < sampleSize {
		sampleSize = len(content)
	}

	sample := content[:sampleSize]
	nonPrintable := 0

	for _, b := range sample {
		// Allow common text characters: printable ASCII, tab, newline, carriage return
		if b < 32 && b != 9 && b != 10 && b != 13 {
			nonPrintable++
		}
	}

	// If more than 10% of the sample is non-printable, consider it binary
	threshold := sampleSize / 10
	return nonPrintable > threshold
}
