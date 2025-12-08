package handlers

import (
	"encoding/json"
	"io"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/integrations"
	"github.com/jacklau/prism/internal/integrations/github"
	"github.com/jacklau/prism/internal/services/coderunner"
)

// GitHubHandler handles GitHub-related endpoints
type GitHubHandler struct {
	webhookRepo        *repository.WebhookRepository
	webhookHandler     *github.WebhookHandler
	codeRunner         *coderunner.Runner
	defaultSecret      string
	integrationManager *integrations.Manager
}

// NewGitHubHandler creates a new GitHub handler
func NewGitHubHandler(
	webhookRepo *repository.WebhookRepository,
	codeRunner *coderunner.Runner,
	defaultSecret string,
	integrationManager *integrations.Manager,
) *GitHubHandler {
	handler := &GitHubHandler{
		webhookRepo:        webhookRepo,
		webhookHandler:     github.NewWebhookHandler(),
		codeRunner:         codeRunner,
		defaultSecret:      defaultSecret,
		integrationManager: integrationManager,
	}

	// Register event processors
	handler.webhookHandler.RegisterProcessor(github.NewIssueProcessor(codeRunner))
	handler.webhookHandler.RegisterProcessor(github.NewIssueCommentProcessor(codeRunner))

	return handler
}

// HandleWebhook handles incoming GitHub webhooks
func (h *GitHubHandler) HandleWebhook(c *fiber.Ctx) error {
	// Get GitHub headers
	eventType := c.Get("X-GitHub-Event")
	deliveryID := c.Get("X-GitHub-Delivery")
	signature := c.Get("X-Hub-Signature-256")

	if eventType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing X-GitHub-Event header",
		})
	}

	// Read the raw body for signature verification
	body := c.Body()

	log.Printf("Received GitHub webhook: event=%s, delivery=%s", eventType, deliveryID)

	// Handle ping event (used to verify webhook setup)
	if eventType == "ping" {
		return c.JSON(fiber.Map{
			"message": "pong",
		})
	}

	// Parse the event to get the repository
	event, err := github.ParseEvent(eventType, body)
	if err != nil {
		log.Printf("Failed to parse webhook event: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "failed to parse event",
		})
	}

	// Get repository name from event
	repoFullName := github.GetRepoFullName(event)
	if repoFullName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "could not determine repository",
		})
	}

	// Look up webhook configuration
	config, err := h.webhookRepo.GetByRepoName(repoFullName)
	if err != nil {
		log.Printf("No webhook config found for %s, using default", repoFullName)
		// Use default secret if no specific config found
		if h.defaultSecret != "" {
			if err := github.VerifySignature(body, signature, h.defaultSecret); err != nil {
				log.Printf("Signature verification failed: %v", err)
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "invalid signature",
				})
			}
		}
		// Create a basic config for processing
		config = &github.WebhookConfig{
			RepoFullName:   repoFullName,
			AutoRunEnabled: false,
		}
	} else {
		// Verify signature with the webhook-specific secret
		if err := github.VerifySignature(body, signature, config.WebhookSecret); err != nil {
			log.Printf("Signature verification failed: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid signature",
			})
		}
	}

	// Track the webhook event
	if h.integrationManager != nil {
		h.integrationManager.Track(&integrations.Event{
			Type: "github.webhook_received",
			Data: map[string]interface{}{
				"event":      eventType,
				"action":     github.GetEventAction(event),
				"repository": repoFullName,
				"delivery":   deliveryID,
			},
		})
	}

	// Process the event asynchronously
	go func() {
		if err := h.webhookHandler.HandleWebhook(eventType, event, config); err != nil {
			log.Printf("Failed to process webhook: %v", err)
			if h.integrationManager != nil {
				h.integrationManager.TrackError("", "", "webhook_processing_error", err.Error())
			}
		}
	}()

	return c.JSON(fiber.Map{
		"message":  "webhook received",
		"delivery": deliveryID,
	})
}

// CreateWebhookConfig creates a new webhook configuration
func (h *GitHubHandler) CreateWebhookConfig(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req struct {
		RepoFullName    string                    `json:"repo_full_name"`
		WebhookSecret   string                    `json:"webhook_secret"`
		Events          []string                  `json:"events"`
		AutoRunEnabled  bool                      `json:"auto_run_enabled"`
		AutoRunTriggers []github.AutoRunTrigger   `json:"auto_run_triggers"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.RepoFullName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "repo_full_name is required",
		})
	}

	if req.WebhookSecret == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "webhook_secret is required",
		})
	}

	config := &github.WebhookConfig{
		UserID:          userID,
		RepoFullName:    req.RepoFullName,
		WebhookSecret:   req.WebhookSecret,
		Events:          req.Events,
		AutoRunEnabled:  req.AutoRunEnabled,
		AutoRunTriggers: req.AutoRunTriggers,
	}

	if err := h.webhookRepo.Create(config); err != nil {
		log.Printf("Failed to create webhook config: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create webhook configuration",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(config)
}

// GetWebhookConfigs returns all webhook configurations for the user
func (h *GitHubHandler) GetWebhookConfigs(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	configs, err := h.webhookRepo.ListByUser(userID)
	if err != nil {
		log.Printf("Failed to list webhook configs: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list webhook configurations",
		})
	}

	return c.JSON(fiber.Map{
		"configs": configs,
	})
}

// GetWebhookConfig returns a specific webhook configuration
func (h *GitHubHandler) GetWebhookConfig(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	configID := c.Params("id")

	config, err := h.webhookRepo.GetByID(configID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "webhook configuration not found",
		})
	}

	// Verify ownership
	if config.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	return c.JSON(config)
}

// UpdateWebhookConfig updates a webhook configuration
func (h *GitHubHandler) UpdateWebhookConfig(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	configID := c.Params("id")

	// Get existing config
	config, err := h.webhookRepo.GetByID(configID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "webhook configuration not found",
		})
	}

	// Verify ownership
	if config.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	var req struct {
		Events          []string                  `json:"events"`
		AutoRunEnabled  *bool                     `json:"auto_run_enabled"`
		AutoRunTriggers []github.AutoRunTrigger   `json:"auto_run_triggers"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Update fields
	if req.Events != nil {
		config.Events = req.Events
	}
	if req.AutoRunEnabled != nil {
		config.AutoRunEnabled = *req.AutoRunEnabled
	}
	if req.AutoRunTriggers != nil {
		config.AutoRunTriggers = req.AutoRunTriggers
	}

	if err := h.webhookRepo.Update(config); err != nil {
		log.Printf("Failed to update webhook config: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update webhook configuration",
		})
	}

	return c.JSON(config)
}

// DeleteWebhookConfig deletes a webhook configuration
func (h *GitHubHandler) DeleteWebhookConfig(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	configID := c.Params("id")

	// Get existing config
	config, err := h.webhookRepo.GetByID(configID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "webhook configuration not found",
		})
	}

	// Verify ownership
	if config.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	if err := h.webhookRepo.Delete(configID); err != nil {
		log.Printf("Failed to delete webhook config: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete webhook configuration",
		})
	}

	return c.JSON(fiber.Map{
		"message": "webhook configuration deleted",
	})
}

// TestWebhook allows testing a webhook configuration
func (h *GitHubHandler) TestWebhook(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	configID := c.Params("id")

	// Get existing config
	config, err := h.webhookRepo.GetByID(configID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "webhook configuration not found",
		})
	}

	// Verify ownership
	if config.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	// Create a test issue event
	testEvent := &github.IssueEvent{
		WebhookEvent: github.WebhookEvent{
			Action: "opened",
			Sender: &github.User{
				Login: "test-user",
			},
			Repo: &github.Repository{
				FullName: config.RepoFullName,
				HTMLURL:  "https://github.com/" + config.RepoFullName,
			},
		},
		Issue: &github.Issue{
			Number:  1,
			Title:   "Test Issue",
			Body:    "This is a test issue created by the webhook test endpoint.",
			HTMLURL: "https://github.com/" + config.RepoFullName + "/issues/1",
		},
	}

	// Process the test event
	err = h.webhookHandler.HandleWebhook("issues", testEvent, config)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "test failed",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "test completed successfully",
	})
}

// GetWebhookDeliveries returns recent webhook deliveries
func (h *GitHubHandler) GetWebhookDeliveries(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	configID := c.Params("id")

	// Get existing config
	config, err := h.webhookRepo.GetByID(configID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "webhook configuration not found",
		})
	}

	// Verify ownership
	if config.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	deliveries, err := h.webhookRepo.ListDeliveries(configID, 50)
	if err != nil {
		log.Printf("Failed to list webhook deliveries: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list deliveries",
		})
	}

	return c.JSON(fiber.Map{
		"deliveries": deliveries,
	})
}

// RunCode manually triggers a code execution
func (h *GitHubHandler) RunCode(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req github.CodeRunRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate command
	if err := coderunner.ValidateCommand(req.Command); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Track the code execution request
	if h.integrationManager != nil {
		h.integrationManager.Track(&integrations.Event{
			Type:   "code.execution_requested",
			UserID: userID,
			Data: map[string]interface{}{
				"environment": req.Environment,
				"command":     req.Command,
			},
		})
	}

	result, err := h.codeRunner.Run(&req)
	if err != nil {
		log.Printf("Code execution failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "code execution failed",
		})
	}

	// Track completion
	if h.integrationManager != nil {
		h.integrationManager.Track(&integrations.Event{
			Type:   "code.execution_completed",
			UserID: userID,
			Data: map[string]interface{}{
				"environment": req.Environment,
				"exit_code":   result.ExitCode,
				"duration_ms": result.Duration,
			},
		})
	}

	return c.JSON(result)
}

// parseWebhookPayload is a helper to parse the raw JSON payload
func parseWebhookPayload(body []byte) (map[string]interface{}, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// readBody reads the request body
func readBody(c *fiber.Ctx) ([]byte, error) {
	return io.ReadAll(c.Request().BodyStream())
}
