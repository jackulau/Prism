package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/approval"
	"github.com/jacklau/prism/internal/database/repository"
)

// ApprovalHandler handles approval workflow endpoints
type ApprovalHandler struct {
	engine  *approval.Engine
	repo    *repository.ApprovalRepository
	orgRepo *repository.OrganizationRepository
}

// NewApprovalHandler creates a new approval handler
func NewApprovalHandler(
	engine *approval.Engine,
	repo *repository.ApprovalRepository,
	orgRepo *repository.OrganizationRepository,
) *ApprovalHandler {
	return &ApprovalHandler{
		engine:  engine,
		repo:    repo,
		orgRepo: orgRepo,
	}
}

// CreateWorkflow creates a new approval workflow
func (h *ApprovalHandler) CreateWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req struct {
		OrganizationID string                       `json:"organization_id"`
		Name           string                       `json:"name"`
		Description    string                       `json:"description"`
		OperationType  string                       `json:"operation_type"`
		Steps          []approval.ApprovalStep      `json:"steps"`
		Conditions     *approval.WorkflowConditions `json:"conditions"`
		Active         bool                         `json:"active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Name == "" || req.OrganizationID == "" || req.OperationType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name, organization_id, and operation_type are required",
		})
	}

	if len(req.Steps) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "at least one step is required",
		})
	}

	// Verify user has permission in organization
	if h.orgRepo != nil {
		isMember, err := h.orgRepo.IsMember(req.OrganizationID, userID)
		if err != nil || !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied",
			})
		}
	}

	// Generate IDs for steps if not provided
	for i := range req.Steps {
		if req.Steps[i].ID == "" {
			req.Steps[i].ID = uuid.New().String()
		}
		req.Steps[i].Order = i
	}

	now := time.Now()
	workflow := &approval.ApprovalWorkflow{
		ID:             uuid.New().String(),
		OrganizationID: req.OrganizationID,
		Name:           req.Name,
		Description:    req.Description,
		OperationType:  approval.OperationType(req.OperationType),
		Steps:          req.Steps,
		Conditions:     req.Conditions,
		Active:         req.Active,
		CreatedAt:      now,
		UpdatedAt:      now,
		CreatedBy:      userID,
	}

	if err := h.repo.CreateWorkflow(workflow); err != nil {
		log.Printf("Failed to create workflow: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create workflow",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(workflow)
}

// GetWorkflow retrieves a workflow by ID
func (h *ApprovalHandler) GetWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	workflowID := c.Params("id")

	workflow, err := h.repo.GetWorkflowByID(workflowID)
	if err != nil {
		log.Printf("Failed to get workflow: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get workflow",
		})
	}
	if workflow == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workflow not found",
		})
	}

	// Verify membership
	if h.orgRepo != nil {
		isMember, err := h.orgRepo.IsMember(workflow.OrganizationID, userID)
		if err != nil || !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied",
			})
		}
	}

	return c.JSON(workflow)
}

// ListWorkflows lists workflows for an organization
func (h *ApprovalHandler) ListWorkflows(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	orgID := c.Query("organization_id")

	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization_id is required",
		})
	}

	// Verify membership
	if h.orgRepo != nil {
		isMember, err := h.orgRepo.IsMember(orgID, userID)
		if err != nil || !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied",
			})
		}
	}

	filter := &approval.WorkflowFilter{
		OrganizationID: orgID,
		Limit:          100,
	}

	if active := c.Query("active"); active != "" {
		activeVal := active == "true"
		filter.Active = &activeVal
	}

	workflows, err := h.repo.ListWorkflows(filter)
	if err != nil {
		log.Printf("Failed to list workflows: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list workflows",
		})
	}

	return c.JSON(fiber.Map{
		"workflows": workflows,
		"count":     len(workflows),
	})
}

// UpdateWorkflow updates a workflow
func (h *ApprovalHandler) UpdateWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	workflowID := c.Params("id")

	workflow, err := h.repo.GetWorkflowByID(workflowID)
	if err != nil || workflow == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workflow not found",
		})
	}

	// Verify membership
	if h.orgRepo != nil {
		isMember, err := h.orgRepo.IsMember(workflow.OrganizationID, userID)
		if err != nil || !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied",
			})
		}
	}

	var req struct {
		Name        string                       `json:"name"`
		Description string                       `json:"description"`
		Steps       []approval.ApprovalStep      `json:"steps"`
		Conditions  *approval.WorkflowConditions `json:"conditions"`
		Active      *bool                        `json:"active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Name != "" {
		workflow.Name = req.Name
	}
	if req.Description != "" {
		workflow.Description = req.Description
	}
	if len(req.Steps) > 0 {
		for i := range req.Steps {
			if req.Steps[i].ID == "" {
				req.Steps[i].ID = uuid.New().String()
			}
			req.Steps[i].Order = i
		}
		workflow.Steps = req.Steps
	}
	if req.Conditions != nil {
		workflow.Conditions = req.Conditions
	}
	if req.Active != nil {
		workflow.Active = *req.Active
	}

	workflow.UpdatedAt = time.Now()

	if err := h.repo.UpdateWorkflow(workflow); err != nil {
		log.Printf("Failed to update workflow: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update workflow",
		})
	}

	return c.JSON(workflow)
}

// DeleteWorkflow deletes a workflow
func (h *ApprovalHandler) DeleteWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	workflowID := c.Params("id")

	workflow, err := h.repo.GetWorkflowByID(workflowID)
	if err != nil || workflow == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "workflow not found",
		})
	}

	// Verify membership
	if h.orgRepo != nil {
		isMember, err := h.orgRepo.IsMember(workflow.OrganizationID, userID)
		if err != nil || !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied",
			})
		}
	}

	if err := h.repo.DeleteWorkflow(workflowID); err != nil {
		log.Printf("Failed to delete workflow: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete workflow",
		})
	}

	return c.JSON(fiber.Map{
		"deleted": true,
	})
}

// CreateRequest creates a new approval request
func (h *ApprovalHandler) CreateRequest(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	userEmail, _ := c.Locals("email").(string)

	var req struct {
		OrganizationID   string                 `json:"organization_id"`
		OperationType    string                 `json:"operation_type"`
		OperationDetails map[string]interface{} `json:"operation_details"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.OrganizationID == "" || req.OperationType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization_id and operation_type are required",
		})
	}

	// Verify membership
	if h.orgRepo != nil {
		isMember, err := h.orgRepo.IsMember(req.OrganizationID, userID)
		if err != nil || !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied",
			})
		}
	}

	request, err := h.engine.CreateRequest(
		c.Context(),
		req.OrganizationID,
		userID,
		userEmail,
		approval.OperationType(req.OperationType),
		req.OperationDetails,
	)
	if err != nil {
		log.Printf("Failed to create approval request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(request)
}

// GetRequest retrieves a request by ID
func (h *ApprovalHandler) GetRequest(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	requestID := c.Params("id")

	request, err := h.repo.GetRequestByID(requestID)
	if err != nil {
		log.Printf("Failed to get request: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get request",
		})
	}
	if request == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "request not found",
		})
	}

	// Verify membership
	if h.orgRepo != nil {
		isMember, err := h.orgRepo.IsMember(request.OrganizationID, userID)
		if err != nil || !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied",
			})
		}
	}

	// Get step progress
	progress, err := h.engine.GetStepProgress(requestID)
	if err != nil {
		log.Printf("Failed to get step progress: %v", err)
	}

	return c.JSON(fiber.Map{
		"request":  request,
		"progress": progress,
	})
}

// ListRequests lists approval requests
func (h *ApprovalHandler) ListRequests(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	orgID := c.Query("organization_id")

	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization_id is required",
		})
	}

	// Verify membership
	if h.orgRepo != nil {
		isMember, err := h.orgRepo.IsMember(orgID, userID)
		if err != nil || !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied",
			})
		}
	}

	filter := &approval.ApprovalRequestFilter{
		OrganizationID: orgID,
		Limit:          100,
	}

	if status := c.Query("status"); status != "" {
		filter.Status = []approval.ApprovalStatus{approval.ApprovalStatus(status)}
	}

	if opType := c.Query("operation_type"); opType != "" {
		filter.OperationType = []approval.OperationType{approval.OperationType(opType)}
	}

	requests, err := h.repo.ListRequests(filter)
	if err != nil {
		log.Printf("Failed to list requests: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list requests",
		})
	}

	return c.JSON(fiber.Map{
		"requests": requests,
		"count":    len(requests),
	})
}

// ListPendingForUser lists pending requests where user can approve
func (h *ApprovalHandler) ListPendingForUser(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	orgID := c.Query("organization_id")

	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization_id is required",
		})
	}

	// Get user roles (would need to be passed or looked up)
	var userRoles []string
	if h.orgRepo != nil {
		member, err := h.orgRepo.GetMember(orgID, userID)
		if err != nil || member == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied",
			})
		}
		userRoles = []string{member.Role}
	}

	requests, err := h.repo.GetPendingRequestsForApprover(orgID, userID, userRoles)
	if err != nil {
		log.Printf("Failed to get pending requests: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get pending requests",
		})
	}

	return c.JSON(fiber.Map{
		"requests": requests,
		"count":    len(requests),
	})
}

// Approve approves a request
func (h *ApprovalHandler) Approve(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	userEmail, _ := c.Locals("email").(string)
	requestID := c.Params("id")

	var req struct {
		Comment string `json:"comment"`
	}
	c.BodyParser(&req)

	request, err := h.repo.GetRequestByID(requestID)
	if err != nil || request == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "request not found",
		})
	}

	// Verify membership
	if h.orgRepo != nil {
		isMember, err := h.orgRepo.IsMember(request.OrganizationID, userID)
		if err != nil || !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied",
			})
		}
	}

	if err := h.engine.RecordDecision(c.Context(), requestID, userID, userEmail, approval.StatusApproved, req.Comment); err != nil {
		log.Printf("Failed to approve request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Get updated request
	updatedRequest, _ := h.repo.GetRequestByID(requestID)

	return c.JSON(fiber.Map{
		"approved": true,
		"request":  updatedRequest,
	})
}

// Reject rejects a request
func (h *ApprovalHandler) Reject(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	userEmail, _ := c.Locals("email").(string)
	requestID := c.Params("id")

	var req struct {
		Comment string `json:"comment"`
	}
	c.BodyParser(&req)

	request, err := h.repo.GetRequestByID(requestID)
	if err != nil || request == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "request not found",
		})
	}

	// Verify membership
	if h.orgRepo != nil {
		isMember, err := h.orgRepo.IsMember(request.OrganizationID, userID)
		if err != nil || !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied",
			})
		}
	}

	if err := h.engine.RecordDecision(c.Context(), requestID, userID, userEmail, approval.StatusRejected, req.Comment); err != nil {
		log.Printf("Failed to reject request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Get updated request
	updatedRequest, _ := h.repo.GetRequestByID(requestID)

	return c.JSON(fiber.Map{
		"rejected": true,
		"request":  updatedRequest,
	})
}

// Escalate escalates a request
func (h *ApprovalHandler) Escalate(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	requestID := c.Params("id")

	request, err := h.repo.GetRequestByID(requestID)
	if err != nil || request == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "request not found",
		})
	}

	// Verify membership
	if h.orgRepo != nil {
		isMember, err := h.orgRepo.IsMember(request.OrganizationID, userID)
		if err != nil || !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied",
			})
		}
	}

	if err := h.engine.EscalateRequest(c.Context(), requestID); err != nil {
		log.Printf("Failed to escalate request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Get updated request
	updatedRequest, _ := h.repo.GetRequestByID(requestID)

	return c.JSON(fiber.Map{
		"escalated": true,
		"request":   updatedRequest,
	})
}

// Cancel cancels a request
func (h *ApprovalHandler) Cancel(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	requestID := c.Params("id")

	request, err := h.repo.GetRequestByID(requestID)
	if err != nil || request == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "request not found",
		})
	}

	// Only requester or admin can cancel
	if request.RequesterID != userID {
		if h.orgRepo != nil {
			member, err := h.orgRepo.GetMember(request.OrganizationID, userID)
			if err != nil || member == nil || member.Role != "admin" {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "only requester or admin can cancel",
				})
			}
		}
	}

	if err := h.engine.CancelRequest(c.Context(), requestID, userID); err != nil {
		log.Printf("Failed to cancel request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"cancelled": true,
	})
}

// GetDecisions retrieves all decisions for a request
func (h *ApprovalHandler) GetDecisions(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	requestID := c.Params("id")

	request, err := h.repo.GetRequestByID(requestID)
	if err != nil || request == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "request not found",
		})
	}

	// Verify membership
	if h.orgRepo != nil {
		isMember, err := h.orgRepo.IsMember(request.OrganizationID, userID)
		if err != nil || !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied",
			})
		}
	}

	decisions, err := h.repo.GetDecisionsForRequest(requestID)
	if err != nil {
		log.Printf("Failed to get decisions: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get decisions",
		})
	}

	return c.JSON(fiber.Map{
		"decisions": decisions,
		"count":     len(decisions),
	})
}

// GetStats retrieves approval statistics for an organization
func (h *ApprovalHandler) GetStats(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	orgID := c.Query("organization_id")

	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "organization_id is required",
		})
	}

	// Verify membership
	if h.orgRepo != nil {
		isMember, err := h.orgRepo.IsMember(orgID, userID)
		if err != nil || !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied",
			})
		}
	}

	// Get counts for each status
	pendingCount, _ := h.repo.CountPendingByOrganization(orgID)

	// List all requests to calculate stats
	allRequests, err := h.repo.ListRequests(&approval.ApprovalRequestFilter{
		OrganizationID: orgID,
	})
	if err != nil {
		log.Printf("Failed to get stats: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get stats",
		})
	}

	approvedCount := 0
	rejectedCount := 0
	expiredCount := 0
	totalCount := len(allRequests)

	for _, req := range allRequests {
		switch req.Status {
		case approval.StatusApproved:
			approvedCount++
		case approval.StatusRejected:
			rejectedCount++
		case approval.StatusExpired:
			expiredCount++
		}
	}

	return c.JSON(fiber.Map{
		"total":    totalCount,
		"pending":  pendingCount,
		"approved": approvedCount,
		"rejected": rejectedCount,
		"expired":  expiredCount,
	})
}
