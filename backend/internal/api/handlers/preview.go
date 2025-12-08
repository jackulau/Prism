package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/sandbox"
)

// PreviewHandler handles preview-related HTTP requests
type PreviewHandler struct {
	sandboxService *sandbox.Service
}

// NewPreviewHandler creates a new preview handler
func NewPreviewHandler(sandboxService *sandbox.Service) *PreviewHandler {
	return &PreviewHandler{
		sandboxService: sandboxService,
	}
}

// ListFiles lists files in the user's sandbox
func (h *PreviewHandler) ListFiles(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	files, err := h.sandboxService.ListFiles(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to list files: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"files": files,
	})
}

// GetFile gets the content of a specific file
func (h *PreviewHandler) GetFile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	filePath := c.Params("*")

	if filePath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "file path is required",
		})
	}

	content, err := h.sandboxService.GetFileContent(userID, filePath)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get file: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"path":    filePath,
		"content": content,
	})
}

// WriteFile writes content to a file
func (h *PreviewHandler) WriteFile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "file path is required",
		})
	}

	if err := h.sandboxService.WriteFile(userID, req.Path, req.Content); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to write file: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"path":    req.Path,
	})
}

// DeleteFile deletes a file
func (h *PreviewHandler) DeleteFile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	filePath := c.Params("*")

	if filePath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "file path is required",
		})
	}

	if err := h.sandboxService.DeleteFile(userID, filePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to delete file: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

// ServePreview serves static files for preview
func (h *PreviewHandler) ServePreview(c *fiber.Ctx) error {
	userID := c.Params("userID")
	filePath := c.Params("*")

	if filePath == "" {
		filePath = "index.html"
	}

	content, err := h.sandboxService.GetFileContent(userID, filePath)
	if err != nil {
		// Try index.html if the path is a directory
		if !strings.Contains(filePath, ".") {
			indexPath := filepath.Join(filePath, "index.html")
			content, err = h.sandboxService.GetFileContent(userID, indexPath)
			if err != nil {
				return c.Status(fiber.StatusNotFound).SendString("File not found")
			}
		} else {
			return c.Status(fiber.StatusNotFound).SendString("File not found")
		}
	}

	// Set content type based on extension
	ext := filepath.Ext(filePath)
	contentType := getContentType(ext)
	c.Set("Content-Type", contentType)

	return c.SendString(content)
}

// GetBuild gets the status of a build
func (h *PreviewHandler) GetBuild(c *fiber.Ctx) error {
	buildID := c.Params("id")

	build, err := h.sandboxService.GetBuild(buildID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("build not found: %v", err),
		})
	}

	response := fiber.Map{
		"id":         build.ID,
		"status":     build.Status,
		"command":    build.Command,
		"start_time": build.StartTime,
	}

	if build.EndTime != nil {
		response["end_time"] = build.EndTime
		response["duration_ms"] = build.EndTime.Sub(build.StartTime).Milliseconds()
	}

	if build.Error != "" {
		response["error"] = build.Error
	}

	if build.PreviewURL != "" {
		response["preview_url"] = build.PreviewURL
	}

	return c.JSON(response)
}

// StopBuild stops a running build
func (h *PreviewHandler) StopBuild(c *fiber.Ctx) error {
	buildID := c.Params("id")

	if err := h.sandboxService.StopBuild(buildID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to stop build: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

// getContentType returns the content type for a file extension
func getContentType(ext string) string {
	contentTypes := map[string]string{
		".html": "text/html; charset=utf-8",
		".htm":  "text/html; charset=utf-8",
		".css":  "text/css; charset=utf-8",
		".js":   "application/javascript; charset=utf-8",
		".mjs":  "application/javascript; charset=utf-8",
		".json": "application/json; charset=utf-8",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".ico":  "image/x-icon",
		".woff": "font/woff",
		".woff2": "font/woff2",
		".ttf":  "font/ttf",
		".eot":  "application/vnd.ms-fontobject",
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".pdf":  "application/pdf",
		".txt":  "text/plain; charset=utf-8",
		".md":   "text/markdown; charset=utf-8",
		".xml":  "application/xml; charset=utf-8",
		".wasm": "application/wasm",
	}

	if ct, ok := contentTypes[strings.ToLower(ext)]; ok {
		return ct
	}

	return http.DetectContentType([]byte{})
}
