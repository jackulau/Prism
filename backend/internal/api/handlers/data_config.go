package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/database/repository"
)

// DataConfigHandler handles encrypted data configuration endpoints
type DataConfigHandler struct {
	repo *repository.DataConfigRepository
}

// NewDataConfigHandler creates a new data config handler
func NewDataConfigHandler(repo *repository.DataConfigRepository) *DataConfigHandler {
	return &DataConfigHandler{repo: repo}
}

// SetConfig handles POST /api/v1/config/:type/:key
func (h *DataConfigHandler) SetConfig(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	configType := c.Params("type")
	configKey := c.Params("key")

	if configType == "" || configKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "config type and key are required",
		})
	}

	var data map[string]interface{}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid JSON body",
		})
	}

	if err := h.repo.SetDataConfig(userID, configType, configKey, data); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to store configuration",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Configuration saved",
	})
}

// GetConfig handles GET /api/v1/config/:type/:key
func (h *DataConfigHandler) GetConfig(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	configType := c.Params("type")
	configKey := c.Params("key")

	if configType == "" || configKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "config type and key are required",
		})
	}

	config, err := h.repo.GetDataConfig(userID, configType, configKey)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve configuration",
		})
	}

	if config == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "configuration not found",
		})
	}

	return c.JSON(fiber.Map{
		"data":       config.Data,
		"updated_at": config.UpdatedAt,
	})
}

// DeleteConfig handles DELETE /api/v1/config/:type/:key
func (h *DataConfigHandler) DeleteConfig(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	configType := c.Params("type")
	configKey := c.Params("key")

	if configType == "" || configKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "config type and key are required",
		})
	}

	if err := h.repo.DeleteDataConfig(userID, configType, configKey); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "configuration not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Configuration deleted",
	})
}

// ListConfigs handles GET /api/v1/config/:type
func (h *DataConfigHandler) ListConfigs(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	configType := c.Params("type")
	if configType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "config type is required",
		})
	}

	configs, err := h.repo.ListDataConfigs(userID, configType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list configurations",
		})
	}

	// Return keys only (not decrypted data)
	keys := make([]string, len(configs))
	for i, cfg := range configs {
		keys[i] = cfg.ConfigKey
	}

	return c.JSON(fiber.Map{
		"keys": keys,
	})
}

// HasConfig handles GET /api/v1/config/:type/:key/exists
func (h *DataConfigHandler) HasConfig(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	configType := c.Params("type")
	configKey := c.Params("key")

	if configType == "" || configKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "config type and key are required",
		})
	}

	exists, err := h.repo.HasDataConfig(userID, configType, configKey)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to check configuration",
		})
	}

	return c.JSON(fiber.Map{
		"exists": exists,
	})
}

// ListConfigTypes handles GET /api/v1/config
func (h *DataConfigHandler) ListConfigTypes(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	types, err := h.repo.ListConfigTypes(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list configuration types",
		})
	}

	return c.JSON(fiber.Map{
		"types": types,
	})
}
