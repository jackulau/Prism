package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/agent/workflow"
)

// WorkflowHandler handles workflow-related HTTP requests
type WorkflowHandler struct {
	engine *workflow.Engine
}

// NewWorkflowHandler creates a new workflow handler
func NewWorkflowHandler(engine *workflow.Engine) *WorkflowHandler {
	return &WorkflowHandler{
		engine: engine,
	}
}

// CreateWorkflowRequest represents a request to create a workflow
type CreateWorkflowRequest struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	TemplateID   string                 `json:"template_id,omitempty"`
	Steps        []workflow.Step        `json:"steps,omitempty"`
	InitialState map[string]interface{} `json:"initial_state,omitempty"`
}

// CreateWorkflow creates a new workflow
func (h *WorkflowHandler) CreateWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req CreateWorkflowRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	var def *workflow.WorkflowDefinition

	if req.TemplateID != "" {
		// Create from template
		var err error
		def, err = workflow.CreateFromTemplate(req.TemplateID, req.InitialState)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		// Override name if provided
		if req.Name != "" {
			def.Name = req.Name
		}
		if req.Description != "" {
			def.Description = req.Description
		}
	} else {
		// Create from steps
		if len(req.Steps) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "steps or template_id is required",
			})
		}
		def = &workflow.WorkflowDefinition{
			Name:         req.Name,
			Description:  req.Description,
			Steps:        req.Steps,
			InitialState: req.InitialState,
		}
	}

	if def.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name is required",
		})
	}

	wf, err := h.engine.CreateWorkflow(userID, def)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"workflow": wf,
	})
}

// ListWorkflows returns workflows for the current user
func (h *WorkflowHandler) ListWorkflows(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	status := c.Query("status")
	name := c.Query("name")
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	filter := &workflow.WorkflowFilter{
		UserID: userID,
		Name:   name,
		Limit:  limit,
		Offset: offset,
	}

	if status != "" {
		filter.Status = []workflow.WorkflowStatus{workflow.WorkflowStatus(status)}
	}

	workflows, err := h.engine.ListWorkflows(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"workflows": workflows,
	})
}

// GetWorkflow returns a specific workflow
func (h *WorkflowHandler) GetWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	workflowID := c.Params("id")

	wf, err := h.engine.GetWorkflow(workflowID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if wf == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workflow not found",
		})
	}

	// Check ownership
	if wf.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	return c.JSON(fiber.Map{
		"workflow": wf,
	})
}

// StartWorkflow starts a workflow
func (h *WorkflowHandler) StartWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	workflowID := c.Params("id")

	// Verify ownership
	wf, err := h.engine.GetWorkflow(workflowID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if wf == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workflow not found",
		})
	}
	if wf.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	// Parse optional initial state update
	var body map[string]interface{}
	if err := c.BodyParser(&body); err == nil {
		if state, ok := body["state"].(map[string]interface{}); ok {
			for k, v := range state {
				wf.SetStateValue(k, v)
			}
		}
	}

	if err := h.engine.StartWorkflow(c.Context(), workflowID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "workflow started",
		"status":  "running",
	})
}

// PauseWorkflow pauses a running workflow
func (h *WorkflowHandler) PauseWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	workflowID := c.Params("id")

	// Verify ownership
	wf, err := h.engine.GetWorkflow(workflowID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if wf == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workflow not found",
		})
	}
	if wf.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	if err := h.engine.PauseWorkflow(c.Context(), workflowID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "workflow paused",
		"status":  "paused",
	})
}

// ResumeWorkflow resumes a paused workflow
func (h *WorkflowHandler) ResumeWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	workflowID := c.Params("id")

	// Verify ownership
	wf, err := h.engine.GetWorkflow(workflowID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if wf == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workflow not found",
		})
	}
	if wf.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	if err := h.engine.ResumeWorkflow(c.Context(), workflowID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "workflow resumed",
		"status":  "running",
	})
}

// CancelWorkflow cancels/deletes a workflow
func (h *WorkflowHandler) CancelWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	workflowID := c.Params("id")

	// Verify ownership
	wf, err := h.engine.GetWorkflow(workflowID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if wf == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workflow not found",
		})
	}
	if wf.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	if err := h.engine.CancelWorkflow(c.Context(), workflowID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "workflow cancelled",
		"status":  "cancelled",
	})
}

// GetWorkflowState returns the current state of a workflow
func (h *WorkflowHandler) GetWorkflowState(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	workflowID := c.Params("id")

	// Verify ownership
	wf, err := h.engine.GetWorkflow(workflowID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if wf == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workflow not found",
		})
	}
	if wf.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	return c.JSON(fiber.Map{
		"state":        wf.State,
		"status":       wf.Status,
		"current_step": wf.CurrentStep,
	})
}

// ProvideInputRequest represents a request to provide input to a waiting step
type ProvideInputRequest struct {
	StepID string      `json:"step_id"`
	Input  interface{} `json:"input"`
}

// ProvideInput provides input to a waiting workflow step
func (h *WorkflowHandler) ProvideInput(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	workflowID := c.Params("id")

	// Verify ownership
	wf, err := h.engine.GetWorkflow(workflowID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if wf == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workflow not found",
		})
	}
	if wf.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	var req ProvideInputRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.StepID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "step_id is required",
		})
	}

	if err := h.engine.ProvideInput(workflowID, req.StepID, req.Input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "input provided",
	})
}

// ListTemplates returns available workflow templates
func (h *WorkflowHandler) ListTemplates(c *fiber.Ctx) error {
	templates := workflow.ListTemplates()
	return c.JSON(fiber.Map{
		"templates": templates,
	})
}

// GetTemplate returns a specific workflow template
func (h *WorkflowHandler) GetTemplate(c *fiber.Ctx) error {
	templateID := c.Params("id")

	info := workflow.GetTemplateInfo(templateID)
	if info == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "template not found",
		})
	}

	def := workflow.GetTemplate(templateID)
	if def == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "template not found",
		})
	}

	return c.JSON(fiber.Map{
		"template": fiber.Map{
			"id":          info.ID,
			"name":        info.Name,
			"description": info.Description,
			"category":    info.Category,
			"tags":        info.Tags,
			"step_count":  info.StepCount,
			"definition":  def,
		},
	})
}
