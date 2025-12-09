package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

// LSTool lists directory contents with detailed information
type LSTool struct {
	sandbox *sandbox.Service
}

// NewLSTool creates a new ls tool
func NewLSTool(sandbox *sandbox.Service) *LSTool {
	return &LSTool{sandbox: sandbox}
}

func (t *LSTool) Name() string {
	return "ls"
}

func (t *LSTool) Description() string {
	return "List directory contents with detailed information including file names, types, sizes, and modification times. Use this to explore directory structure."
}

func (t *LSTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"path": {
				Type:        "string",
				Description: "The directory path to list (optional, defaults to workspace root). Must be a relative path.",
			},
			"show_hidden": {
				Type:        "boolean",
				Description: "Show hidden files starting with '.' (default: false)",
			},
		},
		Required: []string{},
	}
}

// FileEntry represents a file or directory entry
type FileEntry struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Type     string `json:"type"` // "file", "directory", "symlink"
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
	Mode     string `json:"mode"`
}

func (t *LSTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	// Get workspace directory
	workDir, err := t.sandbox.GetOrCreateWorkDir(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Handle optional path parameter
	listDir := workDir
	if pathParam, ok := params["path"].(string); ok && pathParam != "" {
		cleanPath := filepath.Clean(pathParam)
		if strings.HasPrefix(cleanPath, "..") || filepath.IsAbs(cleanPath) {
			return nil, fmt.Errorf("path must be a relative path within the workspace")
		}
		listDir = filepath.Join(workDir, cleanPath)
	}

	showHidden := false
	if sh, ok := params["show_hidden"].(bool); ok {
		showHidden = sh
	}

	// Verify directory exists
	info, err := os.Stat(listDir)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory")
	}

	// Read directory entries
	entries, err := os.ReadDir(listDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var fileEntries []FileEntry
	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files unless requested
		if !showHidden && strings.HasPrefix(name, ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Determine type
		fileType := "file"
		if entry.IsDir() {
			fileType = "directory"
		} else if info.Mode()&os.ModeSymlink != 0 {
			fileType = "symlink"
		}

		// Get relative path from workspace root
		fullPath := filepath.Join(listDir, name)
		relPath, _ := filepath.Rel(workDir, fullPath)

		fileEntries = append(fileEntries, FileEntry{
			Name:     name,
			Path:     relPath,
			Type:     fileType,
			Size:     info.Size(),
			Modified: info.ModTime().Unix(),
			Mode:     info.Mode().String(),
		})
	}

	// Get relative path for the listed directory
	relDir, _ := filepath.Rel(workDir, listDir)
	if relDir == "." {
		relDir = ""
	}

	return map[string]interface{}{
		"path":    relDir,
		"entries": fileEntries,
		"count":   len(fileEntries),
	}, nil
}

func (t *LSTool) RequiresConfirmation() bool {
	return false
}
