package builtin

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

// ListFilesSandboxTool lists files in a repository directory with detailed metadata
type ListFilesSandboxTool struct {
	sandbox *sandbox.Service
}

// NewListFilesSandboxTool creates a new sandbox list files tool
func NewListFilesSandboxTool(sandbox *sandbox.Service) *ListFilesSandboxTool {
	return &ListFilesSandboxTool{sandbox: sandbox}
}

func (t *ListFilesSandboxTool) Name() string {
	return "sandbox_list_files"
}

func (t *ListFilesSandboxTool) Description() string {
	return "List files and directories in a repository path with detailed metadata including permissions, size, and modification time"
}

func (t *ListFilesSandboxTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"relativePath": {
				Type:        "string",
				Description: "Relative path within the repository to list (use '.' for root)",
			},
		},
		Required: []string{"relativePath"},
	}
}

// SandboxFileEntry represents a file or directory entry from ls -la output
type SandboxFileEntry struct {
	Name        string `json:"name"`
	IsDirectory bool   `json:"isDirectory"`
	Size        int64  `json:"size"`
	Permissions string `json:"permissions"`
	ModTime     string `json:"modTime"`
}

func (t *ListFilesSandboxTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	relativePath, ok := params["relativePath"].(string)
	if !ok || relativePath == "" {
		return nil, fmt.Errorf("relativePath parameter is required")
	}

	// Validate path - reject directory traversal attempts
	if err := validatePath(relativePath); err != nil {
		return nil, err
	}

	// Get user's work directory
	workDir, err := t.sandbox.GetOrCreateWorkDir(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Build full path
	fullPath := workDir
	if relativePath != "." && relativePath != "" {
		fullPath = filepath.Join(workDir, relativePath)
	}

	// Execute ls -la command
	cmd := exec.CommandContext(ctx, "ls", "-la", fullPath)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "No such file or directory") {
				return nil, fmt.Errorf("directory not found: %s", relativePath)
			}
			if strings.Contains(stderr, "Permission denied") {
				return nil, fmt.Errorf("permission denied: %s", relativePath)
			}
			return nil, fmt.Errorf("failed to list files: %s", stderr)
		}
		return nil, fmt.Errorf("failed to execute ls command: %w", err)
	}

	// Parse ls -la output
	entries, err := parseLsOutput(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ls output: %w", err)
	}

	return map[string]interface{}{
		"path":    relativePath,
		"entries": entries,
		"count":   len(entries),
	}, nil
}

func (t *ListFilesSandboxTool) RequiresConfirmation() bool {
	return false
}

// validatePath validates that the path is safe (no directory traversal)
func validatePath(path string) error {
	// Reject absolute paths
	if filepath.IsAbs(path) {
		return fmt.Errorf("absolute paths are not allowed")
	}

	// Clean the path and check for directory traversal
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("directory traversal is not allowed")
	}

	// Also check the original path for .. patterns that might bypass Clean
	if strings.Contains(path, "..") {
		return fmt.Errorf("directory traversal is not allowed")
	}

	return nil
}

// parseLsOutput parses the output of ls -la command into SandboxFileEntry slice
func parseLsOutput(output string) ([]SandboxFileEntry, error) {
	var entries []SandboxFileEntry

	scanner := bufio.NewScanner(strings.NewReader(output))

	// Regular expression to parse ls -la output
	// Format: -rw-r--r--  1 user group  1234 Jan  1 12:00 filename
	// Or:     drwxr-xr-x  2 user group  4096 Jan  1 12:00 dirname
	lsPattern := regexp.MustCompile(`^([drwxlsStT-]{10})\s+\d+\s+\S+\s+\S+\s+(\d+)\s+(\w+\s+\d+\s+[\d:]+)\s+(.+)$`)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and the "total" line
		if line == "" || strings.HasPrefix(line, "total ") {
			continue
		}

		matches := lsPattern.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		permissions := matches[1]
		sizeStr := matches[2]
		modTime := matches[3]
		name := matches[4]

		// Skip . and .. entries
		if name == "." || name == ".." {
			continue
		}

		size, _ := strconv.ParseInt(sizeStr, 10, 64)

		entry := SandboxFileEntry{
			Name:        name,
			IsDirectory: permissions[0] == 'd',
			Size:        size,
			Permissions: permissions,
			ModTime:     modTime,
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}
