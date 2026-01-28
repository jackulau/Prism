package handlers

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/sandbox"
)

// SandboxHandler handles sandbox-related HTTP requests
type SandboxHandler struct {
	manager *sandbox.Manager
}

// NewSandboxHandler creates a new sandbox handler
func NewSandboxHandler(manager *sandbox.Manager) *SandboxHandler {
	return &SandboxHandler{
		manager: manager,
	}
}

// CreateSandboxRequest represents a request to create a sandbox
type CreateSandboxRequest struct {
	Provider     string            `json:"provider,omitempty"` // "vercel" or "docker", empty for default
	Framework    string            `json:"framework,omitempty"` // "nextjs", "react", "vue", "vite", "static"
	NodeVersion  string            `json:"node_version,omitempty"`
	BuildCommand string            `json:"build_command,omitempty"`
	OutputDir    string            `json:"output_dir,omitempty"`
	EnvVars      map[string]string `json:"env_vars,omitempty"`
}

// DeploySandboxRequest represents a request to deploy files to a sandbox
type DeploySandboxRequest struct {
	Files map[string]string `json:"files"` // path -> base64-encoded content
}

// CreateSandbox creates a new sandbox
func (h *SandboxHandler) CreateSandbox(c *fiber.Ctx) error {
	var req CreateSandboxRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Convert framework string to type
	framework := sandbox.Framework(req.Framework)
	if framework == "" {
		framework = sandbox.FrameworkStatic
	}

	opts := &sandbox.CreateOptions{
		Framework:    framework,
		NodeVersion:  req.NodeVersion,
		BuildCommand: req.BuildCommand,
		OutputDir:    req.OutputDir,
		EnvVars:      req.EnvVars,
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	sb, err := h.manager.CreateSandbox(ctx, req.Provider, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to create sandbox: %v", err),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         sb.ID,
		"provider":   sb.Provider,
		"status":     sb.Status,
		"created_at": sb.CreatedAt,
		"metadata":   sb.Metadata,
	})
}

// DeploySandbox deploys files to a sandbox
func (h *SandboxHandler) DeploySandbox(c *fiber.Ctx) error {
	sandboxID := c.Params("id")
	provider := c.Query("provider", "")

	var req DeploySandboxRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if len(req.Files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "at least one file is required",
		})
	}

	// Decode base64 files
	files := make(map[string][]byte, len(req.Files))
	for path, content := range req.Files {
		decoded, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			// Try treating as plain text
			files[path] = []byte(content)
		} else {
			files[path] = decoded
		}
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Minute)
	defer cancel()

	result, err := h.manager.DeploySandbox(ctx, provider, sandboxID, files)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to deploy sandbox: %v", err),
		})
	}

	response := fiber.Map{
		"deployment_id": result.DeploymentID,
		"status":        result.Status,
		"created_at":    result.CreatedAt,
	}

	if result.PreviewURL != "" {
		response["preview_url"] = result.PreviewURL
	}

	if result.Error != "" {
		response["error"] = result.Error
	}

	if result.ReadyAt != nil {
		response["ready_at"] = result.ReadyAt
	}

	return c.JSON(response)
}

// GetSandbox gets the status of a sandbox
func (h *SandboxHandler) GetSandbox(c *fiber.Ctx) error {
	sandboxID := c.Params("id")
	provider := c.Query("provider", "")

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	sb, err := h.manager.GetSandbox(ctx, provider, sandboxID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("sandbox not found: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"id":         sb.ID,
		"provider":   sb.Provider,
		"status":     sb.Status,
		"preview_url": sb.PreviewURL,
		"created_at": sb.CreatedAt,
		"updated_at": sb.UpdatedAt,
		"metadata":   sb.Metadata,
	})
}

// GetSandboxLogs gets logs from a sandbox (SSE endpoint)
func (h *SandboxHandler) GetSandboxLogs(c *fiber.Ctx) error {
	sandboxID := c.Params("id")
	provider := c.Query("provider", "")

	// Set headers for SSE
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Minute)
	defer cancel()

	logChan, err := h.manager.GetLogs(ctx, provider, sandboxID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get logs: %v", err),
		})
	}

	// Use Fiber's streaming response
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		for {
			select {
			case log, ok := <-logChan:
				if !ok {
					// Channel closed, send done event
					w.WriteString("event: done\ndata: {}\n\n")
					w.Flush()
					return
				}

				// Format as SSE
				data := fmt.Sprintf(`{"timestamp":"%s","message":"%s","level":"%s","source":"%s"}`,
					log.Timestamp.Format(time.RFC3339),
					escapeJSON(log.Message),
					log.Level,
					log.Source,
				)
				w.WriteString(fmt.Sprintf("event: log\ndata: %s\n\n", data))
				w.Flush()

			case <-ctx.Done():
				return
			}
		}
	})

	return nil
}

// DeleteSandbox deletes a sandbox
func (h *SandboxHandler) DeleteSandbox(c *fiber.Ctx) error {
	sandboxID := c.Params("id")
	provider := c.Query("provider", "")

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	err := h.manager.DeleteSandbox(ctx, provider, sandboxID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to delete sandbox: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"id":      sandboxID,
	})
}

// GetPreviewURL gets the preview URL for a sandbox
func (h *SandboxHandler) GetPreviewURL(c *fiber.Ctx) error {
	sandboxID := c.Params("id")
	provider := c.Query("provider", "")

	url, err := h.manager.GetPreviewURL(provider, sandboxID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get preview URL: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"id":          sandboxID,
		"preview_url": url,
	})
}

// ListProviders lists available sandbox providers
func (h *SandboxHandler) ListProviders(c *fiber.Ctx) error {
	providers := h.manager.ListProviders()

	return c.JSON(fiber.Map{
		"providers": providers,
	})
}

// escapeJSON escapes a string for JSON
func escapeJSON(s string) string {
	result := ""
	for _, r := range s {
		switch r {
		case '"':
			result += `\"`
		case '\\':
			result += `\\`
		case '\n':
			result += `\n`
		case '\r':
			result += `\r`
		case '\t':
			result += `\t`
		default:
			result += string(r)
		}
	}
	return result
}
