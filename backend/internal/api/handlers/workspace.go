package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/sandbox"
	"github.com/sqweek/dialog"
)

// WorkspaceHandler handles workspace-related HTTP requests
type WorkspaceHandler struct {
	sandboxService *sandbox.Service
}

// NewWorkspaceHandler creates a new workspace handler
func NewWorkspaceHandler(sandboxService *sandbox.Service) *WorkspaceHandler {
	return &WorkspaceHandler{
		sandboxService: sandboxService,
	}
}

// Maximum allowed path length
const maxPathLength = 4096

// SetDirectory sets the working directory for the user's workspace
func (h *WorkspaceHandler) SetDirectory(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req struct {
		Directory string `json:"directory"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Directory == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "directory is required",
		})
	}

	// Validate path length to prevent DOS
	if len(req.Directory) > maxPathLength {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("directory path too long (max %d characters)", maxPathLength),
		})
	}

	// Validate no null bytes (can cause issues in C-based file systems)
	if strings.ContainsRune(req.Directory, '\x00') {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "directory path contains invalid characters",
		})
	}

	// Resolve the directory path
	dir := req.Directory
	if dir == "." {
		// Use current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get current directory",
			})
		}
		dir = cwd
	} else if !filepath.IsAbs(dir) {
		// Make relative paths absolute
		cwd, err := os.Getwd()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get current directory",
			})
		}
		dir = filepath.Join(cwd, dir)
	}

	// Clean the path
	dir = filepath.Clean(dir)

	// Verify the directory exists
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "directory does not exist",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to access directory: %v", err),
		})
	}

	if !info.IsDir() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "path is not a directory",
		})
	}

	// Set the workspace directory for the user
	if err := h.sandboxService.SetWorkDir(userID, dir); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to set workspace: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"path":    dir,
	})
}

// GetDirectory gets the current working directory for the user's workspace
func (h *WorkspaceHandler) GetDirectory(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	dir, err := h.sandboxService.GetOrCreateWorkDir(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get workspace: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"path": dir,
	})
}

// CloneGitHubRepo clones a GitHub repository into the workspace
func (h *WorkspaceHandler) CloneGitHubRepo(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req struct {
		RepoURL string `json:"repo_url"`
		Branch  string `json:"branch"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.RepoURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "repo_url is required",
		})
	}

	// Validate URL format (basic check)
	if !strings.HasPrefix(req.RepoURL, "https://github.com/") && !strings.HasPrefix(req.RepoURL, "git@github.com:") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid GitHub repository URL",
		})
	}

	// Get workspace directory
	workDir, err := h.sandboxService.GetOrCreateWorkDir(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get workspace: %v", err),
		})
	}

	// Extract repo name from URL
	repoName := extractRepoName(req.RepoURL)
	clonePath := filepath.Join(workDir, repoName)

	// Check if directory already exists
	if _, err := os.Stat(clonePath); err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "repository already exists in workspace",
			"path":  clonePath,
		})
	}

	// Build git clone command
	args := []string{"clone", "--depth", "1"}
	if req.Branch != "" {
		args = append(args, "-b", req.Branch)
	}
	args = append(args, req.RepoURL, clonePath)

	cmd := exec.Command("git", args...)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   fmt.Sprintf("failed to clone repository: %v", err),
			"details": string(output),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"path":    clonePath,
		"message": fmt.Sprintf("Successfully cloned %s", repoName),
	})
}

// OpenFolderPicker opens the native OS folder picker dialog and returns the selected path
func (h *WorkspaceHandler) OpenFolderPicker(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	// Get the current workspace as the starting directory
	startDir, _ := h.sandboxService.GetOrCreateWorkDir(userID)
	if startDir == "" {
		startDir, _ = os.Getwd()
	}

	// Open native folder picker dialog
	selectedPath, err := dialog.Directory().Title("Select Workspace Folder").SetStartDir(startDir).Browse()
	if err != nil {
		if err == dialog.ErrCancelled {
			return c.JSON(fiber.Map{
				"cancelled": true,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to open folder picker: %v", err),
		})
	}

	// Verify the selected path is a directory
	info, err := os.Stat(selectedPath)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "selected path does not exist",
		})
	}
	if !info.IsDir() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "selected path is not a directory",
		})
	}

	// Set the workspace directory
	if err := h.sandboxService.SetWorkDir(userID, selectedPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to set workspace: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"path":    selectedPath,
	})
}

// BrowseDirectories lists directories at a given path for folder picker UI
func (h *WorkspaceHandler) BrowseDirectories(c *fiber.Ctx) error {
	path := c.Query("path", "/")

	// Validate path length
	if len(path) > maxPathLength {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("path too long (max %d characters)", maxPathLength),
		})
	}

	// Validate no null bytes
	if strings.ContainsRune(path, '\x00') {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "path contains invalid characters",
		})
	}

	// Resolve the path
	resolvedPath := path
	if path == "." || path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get current directory",
			})
		}
		resolvedPath = cwd
	} else if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get current directory",
			})
		}
		resolvedPath = filepath.Join(cwd, path)
	}

	// Clean the path
	resolvedPath = filepath.Clean(resolvedPath)

	// Verify the path exists and is a directory
	info, err := os.Stat(resolvedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "path does not exist",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to access path: %v", err),
		})
	}

	if !info.IsDir() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "path is not a directory",
		})
	}

	// Read directory entries
	entries, err := os.ReadDir(resolvedPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to read directory: %v", err),
		})
	}

	// Filter to only directories
	type DirEntry struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}

	directories := make([]DirEntry, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			// Skip hidden directories (starting with .)
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			directories = append(directories, DirEntry{
				Name: entry.Name(),
				Path: filepath.Join(resolvedPath, entry.Name()),
			})
		}
	}

	// Calculate parent path
	parentPath := filepath.Dir(resolvedPath)
	if parentPath == resolvedPath {
		parentPath = "" // At root, no parent
	}

	return c.JSON(fiber.Map{
		"current_path": resolvedPath,
		"parent_path":  parentPath,
		"directories":  directories,
	})
}

// extractRepoName extracts the repository name from a GitHub URL
func extractRepoName(url string) string {
	// Handle HTTPS URLs: https://github.com/owner/repo.git
	// Handle SSH URLs: git@github.com:owner/repo.git
	url = strings.TrimSuffix(url, ".git")

	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		name := parts[len(parts)-1]
		return sanitizeRepoName(name)
	}

	// Fallback for SSH format
	if strings.Contains(url, ":") {
		parts = strings.Split(url, ":")
		if len(parts) > 1 {
			subparts := strings.Split(parts[1], "/")
			if len(subparts) > 0 {
				name := subparts[len(subparts)-1]
				return sanitizeRepoName(name)
			}
		}
	}

	return "cloned-repo"
}

// sanitizeRepoName removes any path traversal attempts and invalid characters from repo name
func sanitizeRepoName(name string) string {
	// Use filepath.Base to get just the final component (prevents ../ attacks)
	name = filepath.Base(name)

	// Remove any remaining dots at the start (prevents hidden files and . or ..)
	for strings.HasPrefix(name, ".") {
		name = strings.TrimPrefix(name, ".")
	}

	// If name is empty after sanitization, use default
	if name == "" {
		return "cloned-repo"
	}

	// Additional validation: only allow alphanumeric, dash, underscore, and dot (not at start)
	var result strings.Builder
	for i, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		} else if r == '.' && i > 0 {
			result.WriteRune(r)
		}
	}

	sanitized := result.String()
	if sanitized == "" {
		return "cloned-repo"
	}

	return sanitized
}

// ListRecentWorkspaces lists recent workspaces for the user
func (h *WorkspaceHandler) ListRecentWorkspaces(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	repo := h.sandboxService.GetWorkspaceRepository()
	if repo == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "workspace persistence not available",
		})
	}

	workspaces, err := repo.ListRecent(userID, 10)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to list workspaces: %v", err),
		})
	}

	// Convert to response format
	type WorkspaceResponse struct {
		ID             string `json:"id"`
		Path           string `json:"path"`
		Name           string `json:"name"`
		IsCurrent      bool   `json:"is_current"`
		LastAccessedAt string `json:"last_accessed_at,omitempty"`
	}

	result := make([]WorkspaceResponse, 0, len(workspaces))
	for _, ws := range workspaces {
		resp := WorkspaceResponse{
			ID:        ws.ID,
			Path:      ws.Path,
			Name:      ws.Name,
			IsCurrent: ws.IsCurrent,
		}
		if ws.LastAccessedAt != nil {
			resp.LastAccessedAt = ws.LastAccessedAt.Format("2006-01-02T15:04:05Z")
		}
		result = append(result, resp)
	}

	return c.JSON(fiber.Map{
		"workspaces": result,
	})
}

// RemoveWorkspace removes a workspace from the recent list
func (h *WorkspaceHandler) RemoveWorkspace(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	workspaceID := c.Params("id")

	if workspaceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "workspace id is required",
		})
	}

	repo := h.sandboxService.GetWorkspaceRepository()
	if repo == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "workspace persistence not available",
		})
	}

	// Verify the workspace belongs to the user
	workspace, err := repo.GetByID(workspaceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get workspace: %v", err),
		})
	}
	if workspace == nil || workspace.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workspace not found",
		})
	}

	if err := repo.Delete(workspaceID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to remove workspace: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

// SetCurrentWorkspace sets a workspace as the current one
func (h *WorkspaceHandler) SetCurrentWorkspace(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	workspaceID := c.Params("id")

	if workspaceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "workspace id is required",
		})
	}

	repo := h.sandboxService.GetWorkspaceRepository()
	if repo == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "workspace persistence not available",
		})
	}

	// Verify the workspace belongs to the user
	workspace, err := repo.GetByID(workspaceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get workspace: %v", err),
		})
	}
	if workspace == nil || workspace.UserID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workspace not found",
		})
	}

	// Verify the path still exists
	info, err := os.Stat(workspace.Path)
	if err != nil || !info.IsDir() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "workspace path no longer exists",
		})
	}

	// Set as current in database
	if err := repo.SetCurrent(userID, workspaceID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to set current workspace: %v", err),
		})
	}

	// Update in-memory cache via SetWorkDir
	if err := h.sandboxService.SetWorkDir(userID, workspace.Path); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to update workspace: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"path":    workspace.Path,
	})
}

// RenameFile renames a file in the workspace
func (h *WorkspaceHandler) RenameFile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req struct {
		SourcePath string `json:"source_path"`
		DestPath   string `json:"dest_path"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.SourcePath == "" || req.DestPath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "source_path and dest_path are required",
		})
	}

	if err := h.sandboxService.RenameFile(userID, req.SourcePath, req.DestPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

// CreateDirectory creates a directory in the workspace
func (h *WorkspaceHandler) CreateDirectory(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req struct {
		Path string `json:"path"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "path is required",
		})
	}

	if err := h.sandboxService.CreateDirectory(userID, req.Path); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}
