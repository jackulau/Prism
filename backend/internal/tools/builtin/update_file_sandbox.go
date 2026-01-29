package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

// UpdateResult represents the result of a file update operation
type UpdateResult struct {
	Path      string `json:"path"`
	Created   bool   `json:"created"`
	Modified  bool   `json:"modified"`
	Size      int64  `json:"size"`
	Timestamp string `json:"timestamp"`
}

// UpdateFileSandboxTool writes or overwrites file contents in a repository
type UpdateFileSandboxTool struct {
	sandbox     *sandbox.Service
	historyRepo *repository.FileHistoryRepository
}

// NewUpdateFileSandboxTool creates a new update file sandbox tool
func NewUpdateFileSandboxTool(sandbox *sandbox.Service, historyRepo *repository.FileHistoryRepository) *UpdateFileSandboxTool {
	return &UpdateFileSandboxTool{sandbox: sandbox, historyRepo: historyRepo}
}

func (t *UpdateFileSandboxTool) Name() string {
	return "sandbox_update_file"
}

func (t *UpdateFileSandboxTool) Description() string {
	return "Write or overwrite a file in the repository with new content. Creates parent directories automatically if they don't exist."
}

func (t *UpdateFileSandboxTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"relativeFilePath": {
				Type:        "string",
				Description: "Relative path to the file within the repository (e.g., 'src/main.py' or 'config/settings.json')",
			},
			"content": {
				Type:        "string",
				Description: "The new content to write to the file",
			},
		},
		Required: []string{"relativeFilePath", "content"},
	}
}

func (t *UpdateFileSandboxTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract userID from context
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	// Get relativeFilePath from params
	relativeFilePath, ok := params["relativeFilePath"].(string)
	if !ok || relativeFilePath == "" {
		return nil, fmt.Errorf("relativeFilePath parameter is required")
	}

	// Get content from params
	content, ok := params["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content parameter is required")
	}

	// Validate path - reject directory traversal and absolute paths
	if err := validateFilePath(relativeFilePath); err != nil {
		return nil, err
	}

	// Check if file exists (for determining create vs update)
	existingContent, err := t.sandbox.GetFileContent(userID, relativeFilePath)
	fileExists := err == nil && existingContent != ""
	isCreate := !fileExists

	// Save file history before writing (if file exists)
	if t.historyRepo != nil {
		if fileExists {
			// File exists, save its current content to history
			_, _ = t.historyRepo.Create(userID, relativeFilePath, existingContent, "update")
		} else {
			// New file, record creation
			_, _ = t.historyRepo.Create(userID, relativeFilePath, "", "create")
		}
	}

	// Write new content to file (sandbox.WriteFile handles directory creation)
	if err := t.sandbox.WriteFile(userID, relativeFilePath, content); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Return success response with metadata
	result := UpdateResult{
		Path:      relativeFilePath,
		Created:   isCreate,
		Modified:  !isCreate,
		Size:      int64(len(content)),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	return map[string]interface{}{
		"success":   true,
		"path":      result.Path,
		"created":   result.Created,
		"modified":  result.Modified,
		"size":      result.Size,
		"timestamp": result.Timestamp,
	}, nil
}

func (t *UpdateFileSandboxTool) RequiresConfirmation() bool {
	return true // Writing files should require confirmation
}

// validateFilePath validates that the path is safe for file operations
func validateFilePath(path string) error {
	// Reject absolute paths
	if filepath.IsAbs(path) {
		return fmt.Errorf("absolute paths are not allowed")
	}

	// Clean the path and check for directory traversal
	cleanPath := filepath.Clean(path)
	if strings.HasPrefix(cleanPath, "..") || strings.Contains(cleanPath, string(os.PathSeparator)+"..") {
		return fmt.Errorf("directory traversal is not allowed")
	}

	// Check for null bytes (path injection)
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("invalid characters in path")
	}

	return nil
}
