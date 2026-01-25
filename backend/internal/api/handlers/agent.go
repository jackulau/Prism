package handlers

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/database/repository"
)

// AgentHandler handles agent history endpoints
type AgentHandler struct {
	agentRepo *repository.AgentRepository
}

// NewAgentHandler creates a new agent handler
func NewAgentHandler(agentRepo *repository.AgentRepository) *AgentHandler {
	return &AgentHandler{
		agentRepo: agentRepo,
	}
}

// AgentDTO represents an agent response
type AgentDTO struct {
	ID             string     `json:"id"`
	ConversationID *string    `json:"conversation_id,omitempty"`
	Name           string     `json:"name"`
	Description    string     `json:"description,omitempty"`
	Provider       string     `json:"provider"`
	Model          string     `json:"model"`
	SystemPrompt   string     `json:"system_prompt,omitempty"`
	Status         string     `json:"status"`
	Error          string     `json:"error,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

// AgentResultDTO represents an agent result response
type AgentResultDTO struct {
	ID         string                 `json:"id"`
	AgentID    string                 `json:"agent_id"`
	TaskID     string                 `json:"task_id,omitempty"`
	Success    bool                   `json:"success"`
	Output     string                 `json:"output,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Usage      map[string]interface{} `json:"usage,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	DurationMS int64                  `json:"duration_ms"`
	CreatedAt  time.Time              `json:"created_at"`
}

// ListAgents lists all agents for the current user
func (h *AgentHandler) ListAgents(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	agents, err := h.agentRepo.GetByUserID(userID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list agents",
		})
	}

	// Get total count for pagination
	total, err := h.agentRepo.Count(userID)
	if err != nil {
		total = len(agents)
	}

	dtos := make([]AgentDTO, len(agents))
	for i, agent := range agents {
		dtos[i] = agentRecordToDTO(agent)
	}

	return c.JSON(fiber.Map{
		"agents": dtos,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetAgent gets an agent by ID
func (h *AgentHandler) GetAgent(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	agentID := c.Params("id")
	agent, err := h.agentRepo.GetByID(agentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get agent",
		})
	}
	if agent == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "agent not found",
		})
	}

	// Check ownership
	if agent.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	return c.JSON(agentRecordToDTO(agent))
}

// GetAgentResults gets all results for an agent
func (h *AgentHandler) GetAgentResults(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	agentID := c.Params("id")

	// Verify agent exists and belongs to user
	agent, err := h.agentRepo.GetByID(agentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get agent",
		})
	}
	if agent == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "agent not found",
		})
	}

	// Check ownership
	if agent.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	results, err := h.agentRepo.GetResults(agentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get agent results",
		})
	}

	dtos := make([]AgentResultDTO, len(results))
	for i, result := range results {
		dtos[i] = resultRecordToDTO(result)
	}

	return c.JSON(fiber.Map{
		"results": dtos,
	})
}

// DeleteAgent deletes an agent record
func (h *AgentHandler) DeleteAgent(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	agentID := c.Params("id")

	// Verify agent exists and belongs to user
	agent, err := h.agentRepo.GetByID(agentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get agent",
		})
	}
	if agent == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "agent not found",
		})
	}

	// Check ownership
	if agent.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	if err := h.agentRepo.Delete(agentID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete agent",
		})
	}

	return c.JSON(fiber.Map{
		"message": "agent deleted",
	})
}

// GetAgentsByConversation gets all agents for a conversation
func (h *AgentHandler) GetAgentsByConversation(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	conversationID := c.Params("conversationId")

	agents, err := h.agentRepo.GetByConversationID(conversationID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get agents",
		})
	}

	// Filter by ownership
	dtos := make([]AgentDTO, 0)
	for _, agent := range agents {
		if agent.UserID == userID {
			dtos = append(dtos, agentRecordToDTO(agent))
		}
	}

	return c.JSON(fiber.Map{
		"agents": dtos,
	})
}

// agentRecordToDTO converts a repository record to a DTO
func agentRecordToDTO(record *repository.AgentRecord) AgentDTO {
	return AgentDTO{
		ID:             record.ID,
		ConversationID: record.ConversationID,
		Name:           record.Name,
		Description:    record.Description,
		Provider:       record.Provider,
		Model:          record.Model,
		SystemPrompt:   record.SystemPrompt,
		Status:         record.Status,
		Error:          record.Error,
		CreatedAt:      record.CreatedAt,
		StartedAt:      record.StartedAt,
		CompletedAt:    record.CompletedAt,
	}
}

// resultRecordToDTO converts a result record to a DTO
func resultRecordToDTO(record *repository.AgentResultRecord) AgentResultDTO {
	dto := AgentResultDTO{
		ID:         record.ID,
		AgentID:    record.AgentID,
		TaskID:     record.TaskID,
		Success:    record.Success,
		Output:     record.Output,
		Error:      record.Error,
		DurationMS: record.DurationMS,
		CreatedAt:  record.CreatedAt,
	}

	// Parse usage JSON
	if record.UsageJSON != "" {
		var usage map[string]interface{}
		if err := json.Unmarshal([]byte(record.UsageJSON), &usage); err == nil {
			dto.Usage = usage
		}
	}

	// Parse metadata JSON
	if record.MetadataJSON != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(record.MetadataJSON), &metadata); err == nil {
			dto.Metadata = metadata
		}
	}

	return dto
}
