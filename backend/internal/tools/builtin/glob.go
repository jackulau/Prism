package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

// Default directories to skip during glob operations
var defaultSkipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".venv":        true,
	"__pycache__":  true,
	".cache":       true,
	"dist":         true,
	"build":        true,
	".next":        true,
}

// GlobTool finds files matching glob patterns
type GlobTool struct {
	sandbox *sandbox.Service
}

// NewGlobTool creates a new glob tool
func NewGlobTool(sandbox *sandbox.Service) *GlobTool {
	return &GlobTool{sandbox: sandbox}
}

func (t *GlobTool) Name() string {
	return "glob"
}

func (t *GlobTool) Description() string {
	return "Fast file pattern matching tool that finds files by name patterns. Supports glob patterns like '**/*.js', 'src/**/*.ts', '*.go'. Returns matching file paths sorted by modification time (newest first). Use this to find files by name patterns."
}

func (t *GlobTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"pattern": {
				Type:        "string",
				Description: "The glob pattern to match files against (e.g., '**/*.tsx', 'src/**/*.go', '*.md')",
			},
			"path": {
				Type:        "string",
				Description: "The directory to search in (optional, defaults to workspace root). Must be a relative path.",
			},
		},
		Required: []string{"pattern"},
	}
}

// FileMatch represents a matched file with metadata
type FileMatch struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
}

func (t *GlobTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	pattern, ok := params["pattern"].(string)
	if !ok || pattern == "" {
		return nil, fmt.Errorf("pattern parameter is required")
	}

	// Get workspace directory
	workDir, err := t.sandbox.GetOrCreateWorkDir(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Handle optional path parameter
	searchDir := workDir
	if pathParam, ok := params["path"].(string); ok && pathParam != "" {
		// Sanitize path
		cleanPath := filepath.Clean(pathParam)
		if strings.HasPrefix(cleanPath, "..") || filepath.IsAbs(cleanPath) {
			return nil, fmt.Errorf("path must be a relative path within the workspace")
		}
		searchDir = filepath.Join(workDir, cleanPath)

		// Verify the directory exists
		info, err := os.Stat(searchDir)
		if err != nil {
			return nil, fmt.Errorf("path does not exist: %s", pathParam)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("path is not a directory: %s", pathParam)
		}
	}

	// Find matching files
	matches, err := t.findMatches(searchDir, workDir, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search for files: %w", err)
	}

	// Sort by modification time (newest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Modified > matches[j].Modified
	})

	// Limit results
	const maxResults = 10000
	if len(matches) > maxResults {
		matches = matches[:maxResults]
	}

	return map[string]interface{}{
		"files": matches,
		"count": len(matches),
	}, nil
}

func (t *GlobTool) findMatches(searchDir, workDir, pattern string) ([]FileMatch, error) {
	var matches []FileMatch

	// Use doublestar for ** pattern support
	err := filepath.WalkDir(searchDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors and continue
		}

		// Skip hidden directories and common large directories
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || defaultSkipDirs[name] {
				return filepath.SkipDir
			}
			return nil
		}

		// Get relative path from search directory
		relPath, err := filepath.Rel(searchDir, path)
		if err != nil {
			return nil
		}

		// Check if file matches pattern
		matched, err := doublestar.Match(pattern, relPath)
		if err != nil {
			return nil
		}

		// Also try matching just the filename for simple patterns
		if !matched {
			matched, _ = doublestar.Match(pattern, d.Name())
		}

		if matched {
			info, err := d.Info()
			if err != nil {
				return nil
			}

			// Get path relative to workspace root
			workspaceRelPath, _ := filepath.Rel(workDir, path)

			matches = append(matches, FileMatch{
				Name:     d.Name(),
				Path:     workspaceRelPath,
				Size:     info.Size(),
				Modified: info.ModTime().Unix(),
			})
		}

		return nil
	})

	return matches, err
}

func (t *GlobTool) RequiresConfirmation() bool {
	return false
}
