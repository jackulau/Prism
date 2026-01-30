package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/database/repository"
)

// BuildConfigHandler handles build configuration endpoints
type BuildConfigHandler struct {
	repo *repository.BuildConfigRepository
}

// NewBuildConfigHandler creates a new build config handler
func NewBuildConfigHandler(repo *repository.BuildConfigRepository) *BuildConfigHandler {
	return &BuildConfigHandler{repo: repo}
}

// ==================== Request/Response Types ====================

// CreateBuildConfigRequest represents a request to create a build config
type CreateBuildConfigRequest struct {
	WorkspaceID    *string `json:"workspace_id"`
	OrgWorkspaceID *string `json:"org_workspace_id"`
	Name           string  `json:"name"`
	Description    *string `json:"description"`
	IsDefault      bool    `json:"is_default"`
}

// UpdateBuildConfigRequest represents a request to update a build config
type UpdateBuildConfigRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IsDefault   bool    `json:"is_default"`
}

// AddCommandRequest represents a request to add a command
type AddCommandRequest struct {
	Name             string  `json:"name"`
	Command          string  `json:"command"`
	WorkingDirectory *string `json:"working_directory"`
	RunOrder         int     `json:"run_order"`
	IsEnabled        bool    `json:"is_enabled"`
}

// UpdateCommandRequest represents a request to update a command
type UpdateCommandRequest struct {
	Name             string  `json:"name"`
	Command          string  `json:"command"`
	WorkingDirectory *string `json:"working_directory"`
	RunOrder         int     `json:"run_order"`
	IsEnabled        bool    `json:"is_enabled"`
}

// ReorderCommandsRequest represents a request to reorder commands
type ReorderCommandsRequest struct {
	Order []string `json:"order"`
}

// SetEnvVarRequest represents a request to set an env var
type SetEnvVarRequest struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}

// BuildConfigResponse represents a build config in API responses
type BuildConfigResponse struct {
	ID             string                   `json:"id"`
	WorkspaceID    *string                  `json:"workspace_id,omitempty"`
	OrgWorkspaceID *string                  `json:"org_workspace_id,omitempty"`
	UserID         string                   `json:"user_id"`
	Name           string                   `json:"name"`
	Description    *string                  `json:"description,omitempty"`
	IsDefault      bool                     `json:"is_default"`
	CreatedAt      int64                    `json:"created_at"`
	UpdatedAt      int64                    `json:"updated_at"`
	Commands       []BuildCommandResponse   `json:"commands,omitempty"`
	EnvVars        []BuildEnvVarResponse    `json:"env_vars,omitempty"`
}

// BuildCommandResponse represents a build command in API responses
type BuildCommandResponse struct {
	ID               string  `json:"id"`
	ConfigID         string  `json:"config_id"`
	Name             string  `json:"name"`
	Command          string  `json:"command"`
	WorkingDirectory *string `json:"working_directory,omitempty"`
	RunOrder         int     `json:"run_order"`
	IsEnabled        bool    `json:"is_enabled"`
	CreatedAt        int64   `json:"created_at"`
}

// BuildEnvVarResponse represents an env var in API responses
type BuildEnvVarResponse struct {
	ID        string `json:"id"`
	ConfigID  string `json:"config_id"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	IsSecret  bool   `json:"is_secret"`
	CreatedAt int64  `json:"created_at"`
}

// ==================== Config Endpoints ====================

// Create creates a new build configuration
func (h *BuildConfigHandler) Create(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req CreateBuildConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name is required",
		})
	}

	config := &repository.BuildConfig{
		WorkspaceID:    req.WorkspaceID,
		OrgWorkspaceID: req.OrgWorkspaceID,
		UserID:         userID,
		Name:           req.Name,
		Description:    req.Description,
		IsDefault:      req.IsDefault,
	}

	if err := h.repo.Create(config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create build config",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(h.configToResponse(config))
}

// Get retrieves a build configuration with details
func (h *BuildConfigHandler) Get(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id is required",
		})
	}

	config, err := h.repo.GetByIDWithDetails(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get build config",
		})
	}
	if config == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "build config not found",
		})
	}

	// Check ownership
	if config.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	// Mask secret env vars in response
	response := h.configToResponse(config)
	for i := range response.EnvVars {
		if response.EnvVars[i].IsSecret {
			response.EnvVars[i].Value = "********"
		}
	}

	return c.JSON(response)
}

// Update updates a build configuration
func (h *BuildConfigHandler) Update(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id is required",
		})
	}

	config, err := h.repo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get build config",
		})
	}
	if config == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "build config not found",
		})
	}

	// Check ownership
	if config.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	var req UpdateBuildConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name is required",
		})
	}

	config.Name = req.Name
	config.Description = req.Description
	config.IsDefault = req.IsDefault

	if err := h.repo.Update(config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update build config",
		})
	}

	return c.JSON(h.configToResponse(config))
}

// Delete deletes a build configuration
func (h *BuildConfigHandler) Delete(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id is required",
		})
	}

	config, err := h.repo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get build config",
		})
	}
	if config == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "build config not found",
		})
	}

	// Check ownership
	if config.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	if err := h.repo.Delete(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete build config",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "build config deleted",
	})
}

// List lists build configurations
func (h *BuildConfigHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	workspaceID := c.Query("workspace_id")
	orgWorkspaceID := c.Query("org_workspace_id")

	var configs []*repository.BuildConfig
	var err error

	if workspaceID != "" {
		configs, err = h.repo.ListByWorkspaceID(workspaceID)
	} else if orgWorkspaceID != "" {
		configs, err = h.repo.ListByOrgWorkspaceID(orgWorkspaceID)
	} else {
		configs, err = h.repo.ListByUserID(userID)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list build configs",
		})
	}

	// Filter by user ownership
	var filtered []*repository.BuildConfig
	for _, config := range configs {
		if config.UserID == userID {
			filtered = append(filtered, config)
		}
	}

	response := make([]BuildConfigResponse, len(filtered))
	for i, config := range filtered {
		response[i] = h.configToResponse(config)
	}

	return c.JSON(fiber.Map{
		"configs": response,
	})
}

// SetDefault sets a build configuration as default
func (h *BuildConfigHandler) SetDefault(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id is required",
		})
	}

	config, err := h.repo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get build config",
		})
	}
	if config == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "build config not found",
		})
	}

	// Check ownership
	if config.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	if config.WorkspaceID == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "config must have a workspace_id to set as default",
		})
	}

	if err := h.repo.SetDefault(id, *config.WorkspaceID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to set default",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "default config set",
	})
}

// ==================== Command Endpoints ====================

// AddCommand adds a command to a build configuration
func (h *BuildConfigHandler) AddCommand(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	configID := c.Params("id")
	if configID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "config id is required",
		})
	}

	// Verify config ownership
	config, err := h.repo.GetByID(configID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get build config",
		})
	}
	if config == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "build config not found",
		})
	}
	if config.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	var req AddCommandRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name is required",
		})
	}

	cmd := &repository.BuildCommand{
		ConfigID:         configID,
		Name:             req.Name,
		Command:          req.Command,
		WorkingDirectory: req.WorkingDirectory,
		RunOrder:         req.RunOrder,
		IsEnabled:        req.IsEnabled,
	}

	if err := h.repo.AddCommand(cmd); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(h.commandToResponse(cmd))
}

// UpdateCommand updates a build command
func (h *BuildConfigHandler) UpdateCommand(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	configID := c.Params("id")
	cmdID := c.Params("cmdId")
	if configID == "" || cmdID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "config id and command id are required",
		})
	}

	// Verify config ownership
	config, err := h.repo.GetByID(configID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get build config",
		})
	}
	if config == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "build config not found",
		})
	}
	if config.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	// Get the existing command to verify it belongs to this config
	existingCmd, err := h.repo.GetCommand(cmdID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get command",
		})
	}
	if existingCmd == nil || existingCmd.ConfigID != configID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "command not found",
		})
	}

	var req UpdateCommandRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name is required",
		})
	}

	cmd := &repository.BuildCommand{
		ID:               cmdID,
		ConfigID:         configID,
		Name:             req.Name,
		Command:          req.Command,
		WorkingDirectory: req.WorkingDirectory,
		RunOrder:         req.RunOrder,
		IsEnabled:        req.IsEnabled,
	}

	if err := h.repo.UpdateCommand(cmd); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(h.commandToResponse(cmd))
}

// DeleteCommand deletes a build command
func (h *BuildConfigHandler) DeleteCommand(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	configID := c.Params("id")
	cmdID := c.Params("cmdId")
	if configID == "" || cmdID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "config id and command id are required",
		})
	}

	// Verify config ownership
	config, err := h.repo.GetByID(configID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get build config",
		})
	}
	if config == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "build config not found",
		})
	}
	if config.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	// Get the existing command to verify it belongs to this config
	existingCmd, err := h.repo.GetCommand(cmdID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get command",
		})
	}
	if existingCmd == nil || existingCmd.ConfigID != configID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "command not found",
		})
	}

	if err := h.repo.DeleteCommand(cmdID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete command",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "command deleted",
	})
}

// ReorderCommands reorders commands within a build configuration
func (h *BuildConfigHandler) ReorderCommands(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	configID := c.Params("id")
	if configID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "config id is required",
		})
	}

	// Verify config ownership
	config, err := h.repo.GetByID(configID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get build config",
		})
	}
	if config == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "build config not found",
		})
	}
	if config.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	var req ReorderCommandsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if len(req.Order) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order is required",
		})
	}

	if err := h.repo.ReorderCommands(configID, req.Order); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to reorder commands",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "commands reordered",
	})
}

// ==================== Environment Variable Endpoints ====================

// SetEnvVar sets an environment variable
func (h *BuildConfigHandler) SetEnvVar(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	configID := c.Params("id")
	if configID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "config id is required",
		})
	}

	// Verify config ownership
	config, err := h.repo.GetByID(configID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get build config",
		})
	}
	if config == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "build config not found",
		})
	}
	if config.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	var req SetEnvVarRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	envVar := &repository.BuildEnvVar{
		ConfigID: configID,
		Key:      req.Key,
		Value:    req.Value,
		IsSecret: req.IsSecret,
	}

	if err := h.repo.SetEnvVar(envVar); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return response with value masked if secret
	value := req.Value
	if req.IsSecret {
		value = "********"
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"key":       req.Key,
		"is_secret": req.IsSecret,
		"value":     value,
		"message":   "environment variable set",
	})
}

// GetEnvVars retrieves environment variables (secrets masked)
func (h *BuildConfigHandler) GetEnvVars(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	configID := c.Params("id")
	if configID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "config id is required",
		})
	}

	// Verify config ownership
	config, err := h.repo.GetByID(configID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get build config",
		})
	}
	if config == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "build config not found",
		})
	}
	if config.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	envVars, err := h.repo.GetEnvVarsMasked(configID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get environment variables",
		})
	}

	response := make([]BuildEnvVarResponse, len(envVars))
	for i, ev := range envVars {
		response[i] = h.envVarToResponse(&ev)
	}

	return c.JSON(fiber.Map{
		"env_vars": response,
	})
}

// DeleteEnvVar deletes an environment variable
func (h *BuildConfigHandler) DeleteEnvVar(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	configID := c.Params("id")
	key := c.Params("key")
	if configID == "" || key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "config id and key are required",
		})
	}

	// Verify config ownership
	config, err := h.repo.GetByID(configID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get build config",
		})
	}
	if config == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "build config not found",
		})
	}
	if config.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	if err := h.repo.DeleteEnvVarByKey(configID, key); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "environment variable not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "environment variable deleted",
	})
}

// ==================== Helper Methods ====================

func (h *BuildConfigHandler) configToResponse(config *repository.BuildConfig) BuildConfigResponse {
	response := BuildConfigResponse{
		ID:             config.ID,
		WorkspaceID:    config.WorkspaceID,
		OrgWorkspaceID: config.OrgWorkspaceID,
		UserID:         config.UserID,
		Name:           config.Name,
		Description:    config.Description,
		IsDefault:      config.IsDefault,
		CreatedAt:      config.CreatedAt.UnixMilli(),
		UpdatedAt:      config.UpdatedAt.UnixMilli(),
	}

	if len(config.Commands) > 0 {
		response.Commands = make([]BuildCommandResponse, len(config.Commands))
		for i, cmd := range config.Commands {
			response.Commands[i] = h.commandToResponse(&cmd)
		}
	}

	if len(config.EnvVars) > 0 {
		response.EnvVars = make([]BuildEnvVarResponse, len(config.EnvVars))
		for i, ev := range config.EnvVars {
			response.EnvVars[i] = h.envVarToResponse(&ev)
		}
	}

	return response
}

func (h *BuildConfigHandler) commandToResponse(cmd *repository.BuildCommand) BuildCommandResponse {
	return BuildCommandResponse{
		ID:               cmd.ID,
		ConfigID:         cmd.ConfigID,
		Name:             cmd.Name,
		Command:          cmd.Command,
		WorkingDirectory: cmd.WorkingDirectory,
		RunOrder:         cmd.RunOrder,
		IsEnabled:        cmd.IsEnabled,
		CreatedAt:        cmd.CreatedAt.UnixMilli(),
	}
}

func (h *BuildConfigHandler) envVarToResponse(ev *repository.BuildEnvVar) BuildEnvVarResponse {
	return BuildEnvVarResponse{
		ID:        ev.ID,
		ConfigID:  ev.ConfigID,
		Key:       ev.Key,
		Value:     ev.Value,
		IsSecret:  ev.IsSecret,
		CreatedAt: ev.CreatedAt.UnixMilli(),
	}
}
