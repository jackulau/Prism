package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/audit"
)

// AuditHandler handles audit log endpoints
type AuditHandler struct {
	auditService *audit.Service
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService *audit.Service) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// AuditLogResponse represents an audit log in API responses
type AuditLogResponse struct {
	ID            int64                  `json:"id"`
	UserID        *string                `json:"user_id,omitempty"`
	EventType     string                 `json:"event_type"`
	EventCategory string                 `json:"event_category"`
	Action        string                 `json:"action"`
	ResourceType  *string                `json:"resource_type,omitempty"`
	ResourceID    *string                `json:"resource_id,omitempty"`
	IPAddress     *string                `json:"ip_address,omitempty"`
	UserAgent     *string                `json:"user_agent,omitempty"`
	Details       map[string]interface{} `json:"details,omitempty"`
	Success       bool                   `json:"success"`
	CreatedAt     time.Time              `json:"created_at"`
}

// AuditLogsListResponse represents a paginated list of audit logs
type AuditLogsListResponse struct {
	Logs   []AuditLogResponse `json:"logs"`
	Total  int64              `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}

// AuditStatsResponse represents audit statistics
type AuditStatsResponse struct {
	Since          time.Time          `json:"since"`
	CategoryCounts map[string]int64   `json:"category_counts"`
	AuthCounts     map[string]int64   `json:"auth_counts"`
	ProviderCounts map[string]int64   `json:"provider_counts"`
}

// GetLogs retrieves audit logs with filtering and pagination (admin only)
func (h *AuditHandler) GetLogs(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Parse query parameters
	opts := audit.QueryOptions{}

	// Filter by user
	if filterUserID := c.Query("user_id"); filterUserID != "" {
		opts.UserID = &filterUserID
	}

	// Filter by category
	if category := c.Query("category"); category != "" {
		cat := audit.EventCategory(category)
		opts.Category = &cat
	}

	// Filter by event type
	if eventType := c.Query("event_type"); eventType != "" {
		et := audit.EventType(eventType)
		opts.EventType = &et
	}

	// Filter by resource
	if resourceType := c.Query("resource_type"); resourceType != "" {
		opts.ResourceType = &resourceType
	}
	if resourceID := c.Query("resource_id"); resourceID != "" {
		opts.ResourceID = &resourceID
	}

	// Filter by success
	if successStr := c.Query("success"); successStr != "" {
		success := successStr == "true" || successStr == "1"
		opts.Success = &success
	}

	// Filter by date range
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			startDate, err = time.Parse("2006-01-02", startDateStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "invalid start_date format (use RFC3339 or YYYY-MM-DD)",
				})
			}
		}
		opts.StartTime = &startDate
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			endDate, err = time.Parse("2006-01-02", endDateStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "invalid end_date format (use RFC3339 or YYYY-MM-DD)",
				})
			}
			// Set to end of day
			endDate = endDate.Add(24*time.Hour - time.Second)
		}
		opts.EndTime = &endDate
	}

	// Pagination
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid limit",
			})
		}
		opts.Limit = limit
	} else {
		opts.Limit = 50
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid offset",
			})
		}
		opts.Offset = offset
	}

	// Query audit logs
	logs, total, err := h.auditService.Query(opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to query audit logs",
		})
	}

	// Convert to response format
	response := AuditLogsListResponse{
		Logs:   make([]AuditLogResponse, len(logs)),
		Total:  total,
		Limit:  opts.Limit,
		Offset: opts.Offset,
	}

	for i, log := range logs {
		response.Logs[i] = convertToAuditLogResponse(log)
	}

	return c.JSON(response)
}

// GetMyLogs retrieves audit logs for the current user
func (h *AuditHandler) GetMyLogs(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Parse pagination
	limit := 50
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err == nil && o >= 0 {
			offset = o
		}
	}

	// Get user's audit logs
	logs, err := h.auditService.GetUserLogs(userID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get audit logs",
		})
	}

	// Convert to response format
	response := make([]AuditLogResponse, len(logs))
	for i, log := range logs {
		response[i] = convertToAuditLogResponse(log)
	}

	return c.JSON(fiber.Map{
		"logs":   response,
		"limit":  limit,
		"offset": offset,
	})
}

// GetStats retrieves audit log statistics (admin only)
func (h *AuditHandler) GetStats(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Default to last 24 hours
	since := time.Now().Add(-24 * time.Hour)

	if sinceStr := c.Query("since"); sinceStr != "" {
		parsedSince, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			parsedSince, err = time.Parse("2006-01-02", sinceStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "invalid since format (use RFC3339 or YYYY-MM-DD)",
				})
			}
		}
		since = parsedSince
	} else if periodStr := c.Query("period"); periodStr != "" {
		switch periodStr {
		case "1h":
			since = time.Now().Add(-1 * time.Hour)
		case "24h":
			since = time.Now().Add(-24 * time.Hour)
		case "7d":
			since = time.Now().AddDate(0, 0, -7)
		case "30d":
			since = time.Now().AddDate(0, 0, -30)
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid period (use 1h, 24h, 7d, or 30d)",
			})
		}
	}

	stats, err := h.auditService.GetStats(since)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get audit stats",
		})
	}

	// Convert to string keys for JSON
	response := AuditStatsResponse{
		Since:          stats.Since,
		CategoryCounts: make(map[string]int64),
		AuthCounts:     make(map[string]int64),
		ProviderCounts: make(map[string]int64),
	}

	for k, v := range stats.CategoryCounts {
		response.CategoryCounts[string(k)] = v
	}
	for k, v := range stats.AuthCounts {
		response.AuthCounts[string(k)] = v
	}
	for k, v := range stats.ProviderCounts {
		response.ProviderCounts[string(k)] = v
	}

	return c.JSON(response)
}

// convertToAuditLogResponse converts a audit.Log to AuditLogResponse
func convertToAuditLogResponse(log audit.Log) AuditLogResponse {
	return AuditLogResponse{
		ID:            log.ID,
		UserID:        log.UserID,
		EventType:     log.EventType,
		EventCategory: log.EventCategory,
		Action:        log.Action,
		ResourceType:  log.ResourceType,
		ResourceID:    log.ResourceID,
		IPAddress:     log.IPAddress,
		UserAgent:     log.UserAgent,
		Details:       log.Details,
		Success:       log.Success,
		CreatedAt:     log.CreatedAt,
	}
}
