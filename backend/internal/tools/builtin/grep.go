package builtin

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

const (
	maxFileSize    = 10 * 1024 * 1024 // 10MB max file size to grep
	maxOutputLines = 5000             // Max total output lines
	maxFilesToScan = 50000            // Max files to scan
)

// GrepTool searches file contents using regex patterns
type GrepTool struct {
	sandbox *sandbox.Service
}

// NewGrepTool creates a new grep tool
func NewGrepTool(sandbox *sandbox.Service) *GrepTool {
	return &GrepTool{sandbox: sandbox}
}

func (t *GrepTool) Name() string {
	return "grep"
}

func (t *GrepTool) Description() string {
	return `A powerful search tool for finding text patterns in files. Supports full regex syntax (e.g., "log.*Error", "function\s+\w+"). Use 'glob' parameter to filter files (e.g., "*.js", "**/*.tsx"). Output modes: "content" shows matching lines, "files_with_matches" shows only file paths (default), "count" shows match counts.`
}

func (t *GrepTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"pattern": {
				Type:        "string",
				Description: "The regular expression pattern to search for in file contents",
			},
			"path": {
				Type:        "string",
				Description: "File or directory to search in (optional, defaults to workspace root)",
			},
			"glob": {
				Type:        "string",
				Description: "Glob pattern to filter files (e.g., '*.js', '*.{ts,tsx}', 'src/**/*.go')",
			},
			"output_mode": {
				Type:        "string",
				Description: "Output mode: 'content' shows matching lines, 'files_with_matches' shows file paths (default), 'count' shows match counts",
				Enum:        []string{"content", "files_with_matches", "count"},
			},
			"context_before": {
				Type:        "number",
				Description: "Number of lines to show before each match (like grep -B)",
			},
			"context_after": {
				Type:        "number",
				Description: "Number of lines to show after each match (like grep -A)",
			},
			"case_insensitive": {
				Type:        "boolean",
				Description: "Case insensitive search (like grep -i)",
			},
			"head_limit": {
				Type:        "number",
				Description: "Limit output to first N results",
			},
		},
		Required: []string{"pattern"},
	}
}

// GrepMatch represents a single match in content mode
type GrepMatch struct {
	File       string   `json:"file"`
	LineNumber int      `json:"line_number"`
	Content    string   `json:"content"`
	Context    []string `json:"context,omitempty"`
}

// GrepCount represents match count for a file
type GrepCount struct {
	File  string `json:"file"`
	Count int    `json:"count"`
}

func (t *GrepTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
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
	searchPath := workDir
	if pathParam, ok := params["path"].(string); ok && pathParam != "" {
		cleanPath := filepath.Clean(pathParam)
		if strings.HasPrefix(cleanPath, "..") || filepath.IsAbs(cleanPath) {
			return nil, fmt.Errorf("path must be a relative path within the workspace")
		}
		searchPath = filepath.Join(workDir, cleanPath)
	}

	// Parse options
	outputMode := "files_with_matches"
	if mode, ok := params["output_mode"].(string); ok && mode != "" {
		outputMode = mode
	}

	caseInsensitive := false
	if ci, ok := params["case_insensitive"].(bool); ok {
		caseInsensitive = ci
	}

	contextBefore := 0
	if cb, ok := params["context_before"].(float64); ok {
		contextBefore = int(cb)
	}

	contextAfter := 0
	if ca, ok := params["context_after"].(float64); ok {
		contextAfter = int(ca)
	}

	headLimit := maxOutputLines
	if hl, ok := params["head_limit"].(float64); ok && hl > 0 {
		headLimit = int(hl)
	}

	globPattern := ""
	if g, ok := params["glob"].(string); ok {
		globPattern = g
	}

	// Compile regex
	regexPattern := pattern
	if caseInsensitive {
		regexPattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Check if searchPath is a file or directory
	info, err := os.Stat(searchPath)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}

	var filesToSearch []string
	if info.IsDir() {
		filesToSearch, err = t.collectFiles(searchPath, globPattern)
		if err != nil {
			return nil, fmt.Errorf("failed to collect files: %w", err)
		}
	} else {
		filesToSearch = []string{searchPath}
	}

	// Search based on output mode
	switch outputMode {
	case "content":
		matches, err := t.searchContent(filesToSearch, workDir, re, contextBefore, contextAfter, headLimit)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"matches": matches,
			"count":   len(matches),
		}, nil

	case "count":
		counts, total, err := t.searchCount(filesToSearch, workDir, re, headLimit)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"counts": counts,
			"total":  total,
		}, nil

	default: // files_with_matches
		files, err := t.searchFilesWithMatches(filesToSearch, workDir, re, headLimit)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"files": files,
			"count": len(files),
		}, nil
	}
}

func (t *GrepTool) collectFiles(dir, globPattern string) ([]string, error) {
	var files []string
	count := 0

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip hidden and common large directories
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || defaultSkipDirs[name] {
				return filepath.SkipDir
			}
			return nil
		}

		// Limit total files
		if count >= maxFilesToScan {
			return filepath.SkipAll
		}

		// Check glob pattern if specified
		if globPattern != "" {
			relPath, _ := filepath.Rel(dir, path)
			matched, _ := doublestar.Match(globPattern, relPath)
			if !matched {
				matched, _ = doublestar.Match(globPattern, d.Name())
			}
			if !matched {
				return nil
			}
		}

		// Skip large files
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.Size() > maxFileSize {
			return nil
		}

		files = append(files, path)
		count++
		return nil
	})

	return files, err
}

func (t *GrepTool) searchFilesWithMatches(files []string, workDir string, re *regexp.Regexp, limit int) ([]string, error) {
	var matchingFiles []string

	for _, filePath := range files {
		if len(matchingFiles) >= limit {
			break
		}

		if t.fileContainsMatch(filePath, re) {
			relPath, _ := filepath.Rel(workDir, filePath)
			matchingFiles = append(matchingFiles, relPath)
		}
	}

	return matchingFiles, nil
}

func (t *GrepTool) searchContent(files []string, workDir string, re *regexp.Regexp, contextBefore, contextAfter, limit int) ([]GrepMatch, error) {
	var matches []GrepMatch

	for _, filePath := range files {
		if len(matches) >= limit {
			break
		}

		fileMatches, err := t.searchFileContent(filePath, workDir, re, contextBefore, contextAfter, limit-len(matches))
		if err != nil {
			continue // Skip files with errors
		}
		matches = append(matches, fileMatches...)
	}

	return matches, nil
}

func (t *GrepTool) searchFileContent(filePath, workDir string, re *regexp.Regexp, contextBefore, contextAfter, limit int) ([]GrepMatch, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []GrepMatch
	var lines []string
	scanner := bufio.NewScanner(file)

	// Read all lines first (needed for context)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip binary content
		if !utf8.ValidString(line) {
			return nil, nil
		}
		lines = append(lines, line)
	}

	relPath, _ := filepath.Rel(workDir, filePath)

	for i, line := range lines {
		if len(matches) >= limit {
			break
		}

		if re.MatchString(line) {
			match := GrepMatch{
				File:       relPath,
				LineNumber: i + 1,
				Content:    line,
			}

			// Add context if requested
			if contextBefore > 0 || contextAfter > 0 {
				var context []string
				start := i - contextBefore
				if start < 0 {
					start = 0
				}
				end := i + contextAfter + 1
				if end > len(lines) {
					end = len(lines)
				}

				for j := start; j < end; j++ {
					if j != i {
						context = append(context, lines[j])
					}
				}
				match.Context = context
			}

			matches = append(matches, match)
		}
	}

	return matches, nil
}

func (t *GrepTool) searchCount(files []string, workDir string, re *regexp.Regexp, limit int) ([]GrepCount, int, error) {
	var counts []GrepCount
	total := 0

	for _, filePath := range files {
		if len(counts) >= limit {
			break
		}

		count, err := t.countMatches(filePath, re)
		if err != nil || count == 0 {
			continue
		}

		relPath, _ := filepath.Rel(workDir, filePath)
		counts = append(counts, GrepCount{
			File:  relPath,
			Count: count,
		})
		total += count
	}

	return counts, total, nil
}

func (t *GrepTool) countMatches(filePath string, re *regexp.Regexp) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if !utf8.ValidString(line) {
			return 0, nil // Skip binary files
		}
		if re.MatchString(line) {
			count++
		}
	}

	return count, nil
}

func (t *GrepTool) fileContainsMatch(filePath string, re *regexp.Regexp) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !utf8.ValidString(line) {
			return false // Skip binary files
		}
		if re.MatchString(line) {
			return true
		}
	}

	return false
}

func (t *GrepTool) RequiresConfirmation() bool {
	return false
}
