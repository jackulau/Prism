package builtin

import (
	"context"
	"fmt"

	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

// userIDKey is the context key for user ID
type contextKey string

const UserIDKey contextKey = "userID"

// FileReadTool reads file content from the user's sandbox
type FileReadTool struct {
	sandbox *sandbox.Service
}

// NewFileReadTool creates a new file read tool
func NewFileReadTool(sandbox *sandbox.Service) *FileReadTool {
	return &FileReadTool{sandbox: sandbox}
}

func (t *FileReadTool) Name() string {
	return "file_read"
}

func (t *FileReadTool) Description() string {
	return "Read the contents of a file from the user's sandbox workspace. Returns the file content as a string."
}

func (t *FileReadTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"path": {
				Type:        "string",
				Description: "The relative path to the file to read (e.g., 'src/main.py' or 'index.html')",
			},
		},
		Required: []string{"path"},
	}
}

func (t *FileReadTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	path, ok := params["path"].(string)
	if !ok || path == "" {
		return nil, fmt.Errorf("path parameter is required")
	}

	content, err := t.sandbox.GetFileContent(userID, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return map[string]interface{}{
		"content": content,
		"path":    path,
	}, nil
}

func (t *FileReadTool) RequiresConfirmation() bool {
	return false
}

// FileWriteTool writes content to a file in the user's sandbox
type FileWriteTool struct {
	sandbox     *sandbox.Service
	historyRepo *repository.FileHistoryRepository
}

// NewFileWriteTool creates a new file write tool
func NewFileWriteTool(sandbox *sandbox.Service, historyRepo *repository.FileHistoryRepository) *FileWriteTool {
	return &FileWriteTool{sandbox: sandbox, historyRepo: historyRepo}
}

func (t *FileWriteTool) Name() string {
	return "file_write"
}

func (t *FileWriteTool) Description() string {
	return "Write content to a file in the user's sandbox workspace. Creates the file if it doesn't exist, or overwrites if it does. Parent directories are created automatically."
}

func (t *FileWriteTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"path": {
				Type:        "string",
				Description: "The relative path to the file to write (e.g., 'src/main.py' or 'index.html')",
			},
			"content": {
				Type:        "string",
				Description: "The content to write to the file",
			},
		},
		Required: []string{"path", "content"},
	}
}

func (t *FileWriteTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	path, ok := params["path"].(string)
	if !ok || path == "" {
		return nil, fmt.Errorf("path parameter is required")
	}

	content, ok := params["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content parameter is required")
	}

	// Save file history before writing (for existing files)
	if t.historyRepo != nil {
		existingContent, err := t.sandbox.GetFileContent(userID, path)
		if err == nil && existingContent != "" {
			// File exists, save its current content to history
			_, _ = t.historyRepo.Create(userID, path, existingContent, "update")
		} else {
			// New file, record creation
			_, _ = t.historyRepo.Create(userID, path, "", "create")
		}
	}

	if err := t.sandbox.WriteFile(userID, path, content); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"path":    path,
		"bytes":   len(content),
	}, nil
}

func (t *FileWriteTool) RequiresConfirmation() bool {
	return true // Writing files should require confirmation
}

// FileListTool lists files in the user's sandbox
type FileListTool struct {
	sandbox *sandbox.Service
}

// NewFileListTool creates a new file list tool
func NewFileListTool(sandbox *sandbox.Service) *FileListTool {
	return &FileListTool{sandbox: sandbox}
}

func (t *FileListTool) Name() string {
	return "file_list"
}

func (t *FileListTool) Description() string {
	return "List all files and directories in the user's sandbox workspace. Returns a tree structure of files with their names, paths, sizes, and modification times."
}

func (t *FileListTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type:       "object",
		Properties: map[string]llm.JSONProperty{},
		Required:   []string{},
	}
}

func (t *FileListTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	files, err := t.sandbox.ListFiles(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return map[string]interface{}{
		"files": files,
	}, nil
}

func (t *FileListTool) RequiresConfirmation() bool {
	return false
}

// FileDeleteTool deletes a file from the user's sandbox
type FileDeleteTool struct {
	sandbox     *sandbox.Service
	historyRepo *repository.FileHistoryRepository
}

// NewFileDeleteTool creates a new file delete tool
func NewFileDeleteTool(sandbox *sandbox.Service, historyRepo *repository.FileHistoryRepository) *FileDeleteTool {
	return &FileDeleteTool{sandbox: sandbox, historyRepo: historyRepo}
}

func (t *FileDeleteTool) Name() string {
	return "file_delete"
}

func (t *FileDeleteTool) Description() string {
	return "Delete a file or directory from the user's sandbox workspace. Use with caution as this operation cannot be undone."
}

func (t *FileDeleteTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"path": {
				Type:        "string",
				Description: "The relative path to the file or directory to delete",
			},
		},
		Required: []string{"path"},
	}
}

func (t *FileDeleteTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	path, ok := params["path"].(string)
	if !ok || path == "" {
		return nil, fmt.Errorf("path parameter is required")
	}

	// Save file history before deleting
	if t.historyRepo != nil {
		existingContent, err := t.sandbox.GetFileContent(userID, path)
		if err == nil {
			// Save the file content before deletion
			_, _ = t.historyRepo.Create(userID, path, existingContent, "delete")
		}
	}

	if err := t.sandbox.DeleteFile(userID, path); err != nil {
		return nil, fmt.Errorf("failed to delete file: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"path":    path,
	}, nil
}

func (t *FileDeleteTool) RequiresConfirmation() bool {
	return true // Deletion should require confirmation
}

// FileHistoryListTool lists file history entries
type FileHistoryListTool struct {
	historyRepo *repository.FileHistoryRepository
}

// NewFileHistoryListTool creates a new file history list tool
func NewFileHistoryListTool(historyRepo *repository.FileHistoryRepository) *FileHistoryListTool {
	return &FileHistoryListTool{historyRepo: historyRepo}
}

func (t *FileHistoryListTool) Name() string {
	return "file_history_list"
}

func (t *FileHistoryListTool) Description() string {
	return "List file history entries for a specific file or all files. Shows previous versions that can be restored."
}

func (t *FileHistoryListTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"path": {
				Type:        "string",
				Description: "Optional: The relative path to a specific file to get history for. If not provided, returns history for all files.",
			},
			"limit": {
				Type:        "number",
				Description: "Maximum number of history entries to return (default: 20)",
			},
		},
		Required: []string{},
	}
}

func (t *FileHistoryListTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	if t.historyRepo == nil {
		return nil, fmt.Errorf("file history not available")
	}

	limit := 20
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}

	path, hasPath := params["path"].(string)

	if hasPath && path != "" {
		// Get history for specific file
		history, err := t.historyRepo.ListByFilePath(userID, path, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to get file history: %w", err)
		}

		entries := make([]map[string]interface{}, len(history))
		for i, h := range history {
			entries[i] = map[string]interface{}{
				"id":         h.ID,
				"file_path":  h.FilePath,
				"operation":  h.Operation,
				"created_at": h.CreatedAt.Format("2006-01-02 15:04:05"),
				"size":       len(h.Content),
			}
		}

		return map[string]interface{}{
			"path":    path,
			"entries": entries,
			"count":   len(entries),
		}, nil
	}

	// Get all history for user
	history, err := t.historyRepo.ListByUserID(userID, limit, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get file history: %w", err)
	}

	entries := make([]map[string]interface{}, len(history))
	for i, h := range history {
		entries[i] = map[string]interface{}{
			"id":         h.ID,
			"file_path":  h.FilePath,
			"operation":  h.Operation,
			"created_at": h.CreatedAt.Format("2006-01-02 15:04:05"),
			"size":       len(h.Content),
		}
	}

	return map[string]interface{}{
		"entries": entries,
		"count":   len(entries),
	}, nil
}

func (t *FileHistoryListTool) RequiresConfirmation() bool {
	return false
}

// FileHistoryRestoreTool restores a file from history
type FileHistoryRestoreTool struct {
	sandbox     *sandbox.Service
	historyRepo *repository.FileHistoryRepository
}

// NewFileHistoryRestoreTool creates a new file history restore tool
func NewFileHistoryRestoreTool(sandbox *sandbox.Service, historyRepo *repository.FileHistoryRepository) *FileHistoryRestoreTool {
	return &FileHistoryRestoreTool{sandbox: sandbox, historyRepo: historyRepo}
}

func (t *FileHistoryRestoreTool) Name() string {
	return "file_history_restore"
}

func (t *FileHistoryRestoreTool) Description() string {
	return "Restore a file from a specific history entry. Use file_history_list to get available history entries."
}

func (t *FileHistoryRestoreTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"history_id": {
				Type:        "string",
				Description: "The ID of the history entry to restore from (get this from file_history_list)",
			},
		},
		Required: []string{"history_id"},
	}
}

func (t *FileHistoryRestoreTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	if t.historyRepo == nil {
		return nil, fmt.Errorf("file history not available")
	}

	historyID, ok := params["history_id"].(string)
	if !ok || historyID == "" {
		return nil, fmt.Errorf("history_id parameter is required")
	}

	// Get the history entry
	historyEntry, err := t.historyRepo.GetByID(historyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get history entry: %w", err)
	}
	if historyEntry == nil {
		return nil, fmt.Errorf("history entry not found")
	}

	// Verify the history entry belongs to this user
	if historyEntry.UserID != userID {
		return nil, fmt.Errorf("history entry not found")
	}

	// Save current file content to history before restoring
	existingContent, err := t.sandbox.GetFileContent(userID, historyEntry.FilePath)
	if err == nil && existingContent != "" {
		_, _ = t.historyRepo.Create(userID, historyEntry.FilePath, existingContent, "update")
	}

	// Restore the file content
	if err := t.sandbox.WriteFile(userID, historyEntry.FilePath, historyEntry.Content); err != nil {
		return nil, fmt.Errorf("failed to restore file: %w", err)
	}

	return map[string]interface{}{
		"success":   true,
		"path":      historyEntry.FilePath,
		"restored":  historyEntry.CreatedAt.Format("2006-01-02 15:04:05"),
		"operation": historyEntry.Operation,
		"bytes":     len(historyEntry.Content),
	}, nil
}

func (t *FileHistoryRestoreTool) RequiresConfirmation() bool {
	return true // Restoring files should require confirmation
}

// FileHistoryGetTool gets the content of a specific history entry
type FileHistoryGetTool struct {
	historyRepo *repository.FileHistoryRepository
}

// NewFileHistoryGetTool creates a new file history get tool
func NewFileHistoryGetTool(historyRepo *repository.FileHistoryRepository) *FileHistoryGetTool {
	return &FileHistoryGetTool{historyRepo: historyRepo}
}

func (t *FileHistoryGetTool) Name() string {
	return "file_history_get"
}

func (t *FileHistoryGetTool) Description() string {
	return "Get the content of a specific file history entry. Useful for viewing or comparing previous versions."
}

func (t *FileHistoryGetTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"history_id": {
				Type:        "string",
				Description: "The ID of the history entry to get content for",
			},
		},
		Required: []string{"history_id"},
	}
}

func (t *FileHistoryGetTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	if t.historyRepo == nil {
		return nil, fmt.Errorf("file history not available")
	}

	historyID, ok := params["history_id"].(string)
	if !ok || historyID == "" {
		return nil, fmt.Errorf("history_id parameter is required")
	}

	// Get the history entry
	historyEntry, err := t.historyRepo.GetByID(historyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get history entry: %w", err)
	}
	if historyEntry == nil {
		return nil, fmt.Errorf("history entry not found")
	}

	// Verify the history entry belongs to this user
	if historyEntry.UserID != userID {
		return nil, fmt.Errorf("history entry not found")
	}

	return map[string]interface{}{
		"id":         historyEntry.ID,
		"path":       historyEntry.FilePath,
		"content":    historyEntry.Content,
		"operation":  historyEntry.Operation,
		"created_at": historyEntry.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (t *FileHistoryGetTool) RequiresConfirmation() bool {
	return false
}

// FileRenameTool renames or moves a file in the user's sandbox
type FileRenameTool struct {
	sandbox     *sandbox.Service
	historyRepo *repository.FileHistoryRepository
}

// NewFileRenameTool creates a new file rename tool
func NewFileRenameTool(sandbox *sandbox.Service, historyRepo *repository.FileHistoryRepository) *FileRenameTool {
	return &FileRenameTool{sandbox: sandbox, historyRepo: historyRepo}
}

func (t *FileRenameTool) Name() string {
	return "file_rename"
}

func (t *FileRenameTool) Description() string {
	return "Rename or move a file or directory within the user's sandbox workspace. Can be used to rename files in place or move them to different locations within the workspace."
}

func (t *FileRenameTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"source_path": {
				Type:        "string",
				Description: "The current path of the file or directory to rename/move",
			},
			"dest_path": {
				Type:        "string",
				Description: "The new path for the file or directory",
			},
		},
		Required: []string{"source_path", "dest_path"},
	}
}

func (t *FileRenameTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	sourcePath, ok := params["source_path"].(string)
	if !ok || sourcePath == "" {
		return nil, fmt.Errorf("source_path parameter is required")
	}

	destPath, ok := params["dest_path"].(string)
	if !ok || destPath == "" {
		return nil, fmt.Errorf("dest_path parameter is required")
	}

	// Record in history before rename
	if t.historyRepo != nil {
		_, _ = t.historyRepo.Create(userID, sourcePath, "", "rename")
	}

	if err := t.sandbox.RenameFile(userID, sourcePath, destPath); err != nil {
		return nil, fmt.Errorf("failed to rename file: %w", err)
	}

	return map[string]interface{}{
		"success":     true,
		"source_path": sourcePath,
		"dest_path":   destPath,
	}, nil
}

func (t *FileRenameTool) RequiresConfirmation() bool {
	return true // Renaming/moving files should require confirmation
}

// FileCreateDirectoryTool creates a directory in the user's sandbox
type FileCreateDirectoryTool struct {
	sandbox *sandbox.Service
}

// NewFileCreateDirectoryTool creates a new file create directory tool
func NewFileCreateDirectoryTool(sandbox *sandbox.Service) *FileCreateDirectoryTool {
	return &FileCreateDirectoryTool{sandbox: sandbox}
}

func (t *FileCreateDirectoryTool) Name() string {
	return "file_mkdir"
}

func (t *FileCreateDirectoryTool) Description() string {
	return "Create a new directory in the user's sandbox workspace. Creates parent directories automatically if they don't exist."
}

func (t *FileCreateDirectoryTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"path": {
				Type:        "string",
				Description: "The path of the directory to create (relative to workspace root)",
			},
		},
		Required: []string{"path"},
	}
}

func (t *FileCreateDirectoryTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	path, ok := params["path"].(string)
	if !ok || path == "" {
		return nil, fmt.Errorf("path parameter is required")
	}

	if err := t.sandbox.CreateDirectory(userID, path); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"path":    path,
	}, nil
}

func (t *FileCreateDirectoryTool) RequiresConfirmation() bool {
	return false // Creating directories is relatively safe
}
