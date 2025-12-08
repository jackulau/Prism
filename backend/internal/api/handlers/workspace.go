package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/sandbox"
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
