package builtin

import (
	"context"
	"fmt"
	"strings"

	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

// EditTool performs precise string replacements in files
type EditTool struct {
	sandbox     *sandbox.Service
	historyRepo *repository.FileHistoryRepository
}

// NewEditTool creates a new edit tool
func NewEditTool(sandbox *sandbox.Service, historyRepo *repository.FileHistoryRepository) *EditTool {
	return &EditTool{sandbox: sandbox, historyRepo: historyRepo}
}

func (t *EditTool) Name() string {
	return "edit"
}

func (t *EditTool) Description() string {
	return `Performs exact string replacements in files. The old_string must be unique in the file unless replace_all is true. Use this for precise edits instead of rewriting entire files. Always read the file first before editing.`
}

func (t *EditTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"file_path": {
				Type:        "string",
				Description: "The relative path to the file to modify",
			},
			"old_string": {
				Type:        "string",
				Description: "The exact text to replace (must exist in the file)",
			},
			"new_string": {
				Type:        "string",
				Description: "The text to replace it with (must be different from old_string)",
			},
			"replace_all": {
				Type:        "boolean",
				Description: "Replace all occurrences of old_string (default: false, requires unique match)",
			},
		},
		Required: []string{"file_path", "old_string", "new_string"},
	}
}

func (t *EditTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	filePath, ok := params["file_path"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("file_path parameter is required")
	}

	oldString, ok := params["old_string"].(string)
	if !ok {
		return nil, fmt.Errorf("old_string parameter is required")
	}

	newString, ok := params["new_string"].(string)
	if !ok {
		return nil, fmt.Errorf("new_string parameter is required")
	}

	if oldString == newString {
		return nil, fmt.Errorf("old_string and new_string must be different")
	}

	replaceAll := false
	if ra, ok := params["replace_all"].(bool); ok {
		replaceAll = ra
	}

	// Read current file content
	content, err := t.sandbox.GetFileContent(userID, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Count occurrences
	count := strings.Count(content, oldString)
	if count == 0 {
		return nil, fmt.Errorf("old_string not found in file")
	}

	if !replaceAll && count > 1 {
		return nil, fmt.Errorf("old_string found %d times in file. Either provide more context to make it unique, or use replace_all: true", count)
	}

	// Save to history before modifying
	if t.historyRepo != nil {
		_, _ = t.historyRepo.Create(userID, filePath, content, "edit")
	}

	// Perform replacement
	var newContent string
	if replaceAll {
		newContent = strings.ReplaceAll(content, oldString, newString)
	} else {
		newContent = strings.Replace(content, oldString, newString, 1)
	}

	// Write back
	if err := t.sandbox.WriteFile(userID, filePath, newContent); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return map[string]interface{}{
		"success":      true,
		"path":         filePath,
		"replacements": count,
	}, nil
}

func (t *EditTool) RequiresConfirmation() bool {
	return true
}

// MultiEditTool performs batch edits across multiple files
type MultiEditTool struct {
	sandbox     *sandbox.Service
	historyRepo *repository.FileHistoryRepository
}

// NewMultiEditTool creates a new multi-edit tool
func NewMultiEditTool(sandbox *sandbox.Service, historyRepo *repository.FileHistoryRepository) *MultiEditTool {
	return &MultiEditTool{sandbox: sandbox, historyRepo: historyRepo}
}

func (t *MultiEditTool) Name() string {
	return "multi_edit"
}

func (t *MultiEditTool) Description() string {
	return `Performs batch edits across multiple files in a single operation. Each edit specifies a file path, old string, and new string. All edits are validated before any are applied.`
}

func (t *MultiEditTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"edits": {
				Type:        "array",
				Description: "Array of edit operations. Each edit should have 'file_path', 'old_string', and 'new_string' fields.",
			},
		},
		Required: []string{"edits"},
	}
}

// EditOperation represents a single edit in a multi-edit batch
type EditOperation struct {
	FilePath  string `json:"file_path"`
	OldString string `json:"old_string"`
	NewString string `json:"new_string"`
}

// EditResult represents the result of a single edit operation
type EditResult struct {
	FilePath string `json:"file_path"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

func (t *MultiEditTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	editsRaw, ok := params["edits"].([]interface{})
	if !ok || len(editsRaw) == 0 {
		return nil, fmt.Errorf("edits parameter is required and must be a non-empty array")
	}

	// Parse edits
	edits := make([]EditOperation, 0, len(editsRaw))
	for i, editRaw := range editsRaw {
		editMap, ok := editRaw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("edit %d is not a valid object", i)
		}

		filePath, _ := editMap["file_path"].(string)
		oldString, _ := editMap["old_string"].(string)
		newString, _ := editMap["new_string"].(string)

		if filePath == "" || oldString == "" {
			return nil, fmt.Errorf("edit %d: file_path and old_string are required", i)
		}

		edits = append(edits, EditOperation{
			FilePath:  filePath,
			OldString: oldString,
			NewString: newString,
		})
	}

	// Phase 1: Validate all edits and collect original content
	type validatedEdit struct {
		edit           EditOperation
		originalContent string
		newContent     string
	}
	validatedEdits := make([]validatedEdit, 0, len(edits))

	for i, edit := range edits {
		content, err := t.sandbox.GetFileContent(userID, edit.FilePath)
		if err != nil {
			return nil, fmt.Errorf("edit %d: failed to read file %s: %w", i, edit.FilePath, err)
		}

		count := strings.Count(content, edit.OldString)
		if count == 0 {
			return nil, fmt.Errorf("edit %d: old_string not found in file %s", i, edit.FilePath)
		}
		if count > 1 {
			return nil, fmt.Errorf("edit %d: old_string found %d times in file %s (must be unique)", i, count, edit.FilePath)
		}

		newContent := strings.Replace(content, edit.OldString, edit.NewString, 1)
		validatedEdits = append(validatedEdits, validatedEdit{
			edit:           edit,
			originalContent: content,
			newContent:     newContent,
		})
	}

	// Phase 2: Save all originals to history
	if t.historyRepo != nil {
		for _, ve := range validatedEdits {
			_, _ = t.historyRepo.Create(userID, ve.edit.FilePath, ve.originalContent, "multi_edit")
		}
	}

	// Phase 3: Apply all edits
	results := make([]EditResult, 0, len(validatedEdits))
	for _, ve := range validatedEdits {
		err := t.sandbox.WriteFile(userID, ve.edit.FilePath, ve.newContent)
		if err != nil {
			results = append(results, EditResult{
				FilePath: ve.edit.FilePath,
				Success:  false,
				Error:    err.Error(),
			})
		} else {
			results = append(results, EditResult{
				FilePath: ve.edit.FilePath,
				Success:  true,
			})
		}
	}

	// Check overall success
	allSuccess := true
	for _, r := range results {
		if !r.Success {
			allSuccess = false
			break
		}
	}

	return map[string]interface{}{
		"success": allSuccess,
		"results": results,
		"count":   len(results),
	}, nil
}

func (t *MultiEditTool) RequiresConfirmation() bool {
	return true
}
