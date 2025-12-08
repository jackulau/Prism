package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/database/repository"
)

// IntegrationHandler handles integration settings endpoints
type IntegrationHandler struct {
	integrationRepo *repository.IntegrationRepository
}

// NewIntegrationHandler creates a new integration handler
func NewIntegrationHandler(integrationRepo *repository.IntegrationRepository) *IntegrationHandler {
	return &IntegrationHandler{
		integrationRepo: integrationRepo,
	}
}

// IntegrationStatusResponse represents the status of all integrations
type IntegrationStatusResponse struct {
	Discord IntegrationStatus `json:"discord"`
	Slack   IntegrationStatus `json:"slack"`
	PostHog IntegrationStatus `json:"posthog"`
}

// IntegrationStatus represents the status of a single integration
type IntegrationStatus struct {
	Enabled   bool   `json:"enabled"`
	Connected bool   `json:"connected"`
	ChannelID string `json:"channel_id,omitempty"`
}

// SetIntegrationRequest represents a request to set integration settings
type SetIntegrationRequest struct {
	WebhookURL string `json:"webhook_url"`
	ChannelID  string `json:"channel_id,omitempty"`
	Enabled    bool   `json:"enabled"`
}

// GetStatus returns the status of all integrations for the current user
func (h *IntegrationHandler) GetStatus(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	settings, err := h.integrationRepo.GetAllSettings(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get integration settings",
		})
	}

	response := IntegrationStatusResponse{
		Discord: IntegrationStatus{Enabled: false, Connected: false},
		Slack:   IntegrationStatus{Enabled: false, Connected: false},
		PostHog: IntegrationStatus{Enabled: false, Connected: false},
	}

	if discord, ok := settings["discord"]; ok && discord != nil {
		response.Discord.Enabled = discord.Enabled
		response.Discord.Connected = discord.WebhookURL != ""
	}

	if slack, ok := settings["slack"]; ok && slack != nil {
		response.Slack.Enabled = slack.Enabled
		response.Slack.Connected = slack.WebhookURL != ""
		response.Slack.ChannelID = slack.ChannelID
	}

	if posthog, ok := settings["posthog"]; ok && posthog != nil {
		response.PostHog.Enabled = posthog.Enabled
		response.PostHog.Connected = posthog.Enabled // PostHog doesn't have webhook
	}

	return c.JSON(response)
}

// SetDiscord sets Discord integration settings
func (h *IntegrationHandler) SetDiscord(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req SetIntegrationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Enable if webhook URL is provided
	enabled := req.WebhookURL != ""
	if req.Enabled {
		enabled = true
	}

	if err := h.integrationRepo.SetDiscordSettings(userID, req.WebhookURL, enabled); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to save discord settings",
		})
	}

	return c.JSON(fiber.Map{
		"message":   "Discord integration configured successfully",
		"enabled":   enabled,
		"connected": req.WebhookURL != "",
	})
}

// DeleteDiscord removes Discord integration settings
func (h *IntegrationHandler) DeleteDiscord(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	if err := h.integrationRepo.DeleteDiscordSettings(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete discord settings",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Discord integration disconnected",
	})
}

// SetSlack sets Slack integration settings
func (h *IntegrationHandler) SetSlack(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req SetIntegrationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Enable if webhook URL is provided
	enabled := req.WebhookURL != ""
	if req.Enabled {
		enabled = true
	}

	if err := h.integrationRepo.SetSlackSettings(userID, req.WebhookURL, req.ChannelID, enabled); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to save slack settings",
		})
	}

	return c.JSON(fiber.Map{
		"message":   "Slack integration configured successfully",
		"enabled":   enabled,
		"connected": req.WebhookURL != "",
	})
}

// DeleteSlack removes Slack integration settings
func (h *IntegrationHandler) DeleteSlack(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	if err := h.integrationRepo.DeleteSlackSettings(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete slack settings",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Slack integration disconnected",
	})
}

// SetPostHog sets PostHog integration settings
func (h *IntegrationHandler) SetPostHog(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req SetIntegrationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.integrationRepo.SetPostHogSettings(userID, req.Enabled); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to save posthog settings",
		})
	}

	return c.JSON(fiber.Map{
		"message":   "PostHog integration configured successfully",
		"enabled":   req.Enabled,
		"connected": req.Enabled,
	})
}

// DeletePostHog removes PostHog integration settings
func (h *IntegrationHandler) DeletePostHog(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	if err := h.integrationRepo.DeletePostHogSettings(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete posthog settings",
		})
	}

	return c.JSON(fiber.Map{
		"message": "PostHog integration disabled",
	})
}
