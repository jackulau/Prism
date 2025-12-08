package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/database/repository"
)

// ChatHandler handles chat endpoints
type ChatHandler struct {
	conversationRepo *repository.ConversationRepository
	messageRepo      *repository.MessageRepository
}

// NewChatHandler creates a new chat handler
func NewChatHandler(conversationRepo *repository.ConversationRepository, messageRepo *repository.MessageRepository) *ChatHandler {
	return &ChatHandler{
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
	}
}

// ConversationDTO represents a conversation response
type ConversationDTO struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Provider     string    `json:"provider"`
	Model        string    `json:"model"`
	SystemPrompt string    `json:"system_prompt,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// MessageDTO represents a message response
type MessageDTO struct {
	ID         string                   `json:"id"`
	Role       string                   `json:"role"`
	Content    string                   `json:"content"`
	ToolCalls  []map[string]interface{} `json:"tool_calls,omitempty"`
	ToolCallID string                   `json:"tool_call_id,omitempty"`
	CreatedAt  time.Time                `json:"created_at"`
}

// CreateConversationRequest represents a request to create a conversation
type CreateConversationRequest struct {
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	SystemPrompt string `json:"system_prompt,omitempty"`
}

// UpdateConversationRequest represents a request to update a conversation
type UpdateConversationRequest struct {
	Title string `json:"title"`
}

// ListConversations lists all conversations for the current user
func (h *ChatHandler) ListConversations(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	conversations, err := h.conversationRepo.ListByUserID(userID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list conversations",
		})
	}

	dtos := make([]ConversationDTO, len(conversations))
	for i, conv := range conversations {
		dtos[i] = ConversationDTO{
			ID:           conv.ID,
			Title:        conv.Title,
			Provider:     conv.Provider,
			Model:        conv.Model,
			SystemPrompt: conv.SystemPrompt,
			CreatedAt:    conv.CreatedAt,
			UpdatedAt:    conv.UpdatedAt,
		}
	}

	return c.JSON(fiber.Map{
		"conversations": dtos,
	})
}

// CreateConversation creates a new conversation
func (h *ChatHandler) CreateConversation(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req CreateConversationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Provider == "" || req.Model == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provider and model are required",
		})
	}

	conv, err := h.conversationRepo.Create(userID, req.Provider, req.Model, req.SystemPrompt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create conversation",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(ConversationDTO{
		ID:           conv.ID,
		Title:        conv.Title,
		Provider:     conv.Provider,
		Model:        conv.Model,
		SystemPrompt: conv.SystemPrompt,
		CreatedAt:    conv.CreatedAt,
		UpdatedAt:    conv.UpdatedAt,
	})
}

// GetConversation gets a conversation by ID
func (h *ChatHandler) GetConversation(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	convID := c.Params("id")
	conv, err := h.conversationRepo.GetByID(convID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get conversation",
		})
	}
	if conv == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "conversation not found",
		})
	}

	// Check ownership
	if conv.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	return c.JSON(ConversationDTO{
		ID:           conv.ID,
		Title:        conv.Title,
		Provider:     conv.Provider,
		Model:        conv.Model,
		SystemPrompt: conv.SystemPrompt,
		CreatedAt:    conv.CreatedAt,
		UpdatedAt:    conv.UpdatedAt,
	})
}

// UpdateConversation updates a conversation
func (h *ChatHandler) UpdateConversation(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	convID := c.Params("id")
	conv, err := h.conversationRepo.GetByID(convID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get conversation",
		})
	}
	if conv == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "conversation not found",
		})
	}

	// Check ownership
	if conv.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	var req UpdateConversationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.conversationRepo.Update(convID, req.Title); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update conversation",
		})
	}

	conv.Title = req.Title
	return c.JSON(ConversationDTO{
		ID:           conv.ID,
		Title:        conv.Title,
		Provider:     conv.Provider,
		Model:        conv.Model,
		SystemPrompt: conv.SystemPrompt,
		CreatedAt:    conv.CreatedAt,
		UpdatedAt:    time.Now(),
	})
}

// DeleteConversation deletes a conversation
func (h *ChatHandler) DeleteConversation(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	convID := c.Params("id")
	conv, err := h.conversationRepo.GetByID(convID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get conversation",
		})
	}
	if conv == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "conversation not found",
		})
	}

	// Check ownership
	if conv.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	if err := h.conversationRepo.Delete(convID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete conversation",
		})
	}

	return c.JSON(fiber.Map{
		"message": "conversation deleted",
	})
}

// GetMessages gets all messages for a conversation
func (h *ChatHandler) GetMessages(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	convID := c.Params("id")
	conv, err := h.conversationRepo.GetByID(convID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get conversation",
		})
	}
	if conv == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "conversation not found",
		})
	}

	// Check ownership
	if conv.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	messages, err := h.messageRepo.ListByConversationID(convID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list messages",
		})
	}

	dtos := make([]MessageDTO, len(messages))
	for i, msg := range messages {
		var toolCalls []map[string]interface{}
		for _, tc := range msg.ToolCalls {
			toolCalls = append(toolCalls, map[string]interface{}{
				"id":         tc.ID,
				"name":       tc.Name,
				"parameters": tc.Parameters,
			})
		}

		dtos[i] = MessageDTO{
			ID:         msg.ID,
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCalls:  toolCalls,
			ToolCallID: msg.ToolCallID,
			CreatedAt:  msg.CreatedAt,
		}
	}

	return c.JSON(fiber.Map{
		"messages": dtos,
	})
}
