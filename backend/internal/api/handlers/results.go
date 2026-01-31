package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/database/repository"
)

// ResultsHandler handles results aggregation endpoints
type ResultsHandler struct {
	executionRepo *repository.AgentExecutionRepository
}

// NewResultsHandler creates a new results handler
func NewResultsHandler(executionRepo *repository.AgentExecutionRepository) *ResultsHandler {
	return &ResultsHandler{
		executionRepo: executionRepo,
	}
}

// ExecutionSummaryDTO represents a summary of an execution
type ExecutionSummaryDTO struct {
	ID               string                 `json:"id"`
	UserID           string                 `json:"user_id"`
	Provider         string                 `json:"provider"`
	LLMProvider      string                 `json:"llm_provider"`
	Model            string                 `json:"model"`
	AgentName        string                 `json:"agent_name,omitempty"`
	Status           string                 `json:"status"`
	PromptTokens     int                    `json:"prompt_tokens"`
	CompletionTokens int                    `json:"completion_tokens"`
	TotalTokens      int                    `json:"total_tokens"`
	InputCost        float64                `json:"input_cost"`
	OutputCost       float64                `json:"output_cost"`
	TotalCost        float64                `json:"total_cost"`
	Currency         string                 `json:"currency"`
	Error            string                 `json:"error,omitempty"`
	DurationMs       int64                  `json:"duration_ms,omitempty"`
	StartedAt        *int64                 `json:"started_at,omitempty"`
	CompletedAt      *int64                 `json:"completed_at,omitempty"`
	CreatedAt        int64                  `json:"created_at"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// BatchResultsSummaryDTO represents aggregated batch results
type BatchResultsSummaryDTO struct {
	Total           int                   `json:"total"`
	Completed       int                   `json:"completed"`
	Failed          int                   `json:"failed"`
	Running         int                   `json:"running"`
	Pending         int                   `json:"pending"`
	Cancelled       int                   `json:"cancelled"`
	SuccessRate     float64               `json:"success_rate"`
	AvgDurationMs   int64                 `json:"avg_duration_ms"`
	TotalTokens     int                   `json:"total_tokens"`
	TotalCost       float64               `json:"total_cost"`
	Executions      []ExecutionSummaryDTO `json:"executions"`
	Limit           int                   `json:"limit"`
	Offset          int                   `json:"offset"`
}

// SwarmResultsSummaryDTO represents aggregated swarm results
type SwarmResultsSummaryDTO struct {
	Total           int                   `json:"total"`
	Completed       int                   `json:"completed"`
	Failed          int                   `json:"failed"`
	Running         int                   `json:"running"`
	Pending         int                   `json:"pending"`
	Cancelled       int                   `json:"cancelled"`
	SuccessRate     float64               `json:"success_rate"`
	AvgDurationMs   int64                 `json:"avg_duration_ms"`
	TotalTokens     int                   `json:"total_tokens"`
	TotalCost       float64               `json:"total_cost"`
	Executions      []ExecutionSummaryDTO `json:"executions"`
	Limit           int                   `json:"limit"`
	Offset          int                   `json:"offset"`
}

// AggregatedMetricsDTO represents aggregated metrics response
type AggregatedMetricsDTO struct {
	TotalExecutions       int             `json:"total_executions"`
	TotalPromptTokens     int             `json:"total_prompt_tokens"`
	TotalCompletionTokens int             `json:"total_completion_tokens"`
	TotalTokens           int             `json:"total_tokens"`
	TotalCost             float64         `json:"total_cost"`
	AvgTokensPerExecution float64         `json:"avg_tokens_per_execution"`
	AvgCostPerExecution   float64         `json:"avg_cost_per_execution"`
	StatusBreakdown       map[string]int  `json:"status_breakdown"`
	ProviderBreakdown     map[string]int  `json:"provider_breakdown"`
	ModelBreakdown        map[string]int  `json:"model_breakdown"`
	SuccessRate           float64         `json:"success_rate"`
	Period                string          `json:"period"`
}

// TimelinePointDTO represents a single point in the timeline
type TimelinePointDTO struct {
	Timestamp        int64   `json:"timestamp"`
	Date             string  `json:"date"`
	ExecutionCount   int     `json:"execution_count"`
	SuccessCount     int     `json:"success_count"`
	FailureCount     int     `json:"failure_count"`
	TotalTokens      int     `json:"total_tokens"`
	TotalCost        float64 `json:"total_cost"`
	AvgDurationMs    int64   `json:"avg_duration_ms"`
}

// TimelineResponseDTO represents timeline data response
type TimelineResponseDTO struct {
	Points      []TimelinePointDTO `json:"points"`
	Granularity string             `json:"granularity"`
	Since       int64              `json:"since"`
	Until       int64              `json:"until"`
}

// ExecutionDetailDTO represents detailed execution information
type ExecutionDetailDTO struct {
	ExecutionSummaryDTO
	Messages  []ExecutionMessageDTO   `json:"messages,omitempty"`
	ToolCalls []ExecutionToolCallDTO  `json:"tool_calls,omitempty"`
}

// ExecutionMessageDTO represents a message in an execution
type ExecutionMessageDTO struct {
	ID               string                   `json:"id"`
	Role             string                   `json:"role"`
	Content          string                   `json:"content"`
	ToolCalls        []map[string]interface{} `json:"tool_calls,omitempty"`
	ToolCallID       string                   `json:"tool_call_id,omitempty"`
	PromptTokens     int                      `json:"prompt_tokens"`
	CompletionTokens int                      `json:"completion_tokens"`
	CreatedAt        int64                    `json:"created_at"`
}

// ExecutionToolCallDTO represents a tool call in an execution
type ExecutionToolCallDTO struct {
	ID          string                 `json:"id"`
	ToolName    string                 `json:"tool_name"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Output      string                 `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Status      string                 `json:"status"`
	DurationMs  int64                  `json:"duration_ms"`
	CreatedAt   int64                  `json:"created_at"`
	CompletedAt *int64                 `json:"completed_at,omitempty"`
}

// ListBatchResults lists batch execution summaries
// GET /api/v1/results/batch
func (h *ResultsHandler) ListBatchResults(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Parse query parameters
	params := parseResultsQueryParams(c)

	// Get executions from repository
	executions, err := h.executionRepo.ListByUserID(userID, params.limit, params.offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list batch results",
		})
	}

	// Filter and aggregate
	filtered := filterExecutions(executions, params)
	summary := aggregateBatchResults(filtered, params)

	return c.JSON(summary)
}

// ListSwarmResults lists swarm execution summaries
// GET /api/v1/results/swarm
func (h *ResultsHandler) ListSwarmResults(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Parse query parameters
	params := parseResultsQueryParams(c)

	// Get executions from repository (swarm executions have specific metadata)
	executions, err := h.executionRepo.ListByUserID(userID, params.limit, params.offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list swarm results",
		})
	}

	// Filter for swarm executions and aggregate
	filtered := filterSwarmExecutions(executions, params)
	summary := aggregateSwarmResults(filtered, params)

	return c.JSON(summary)
}

// GetExecutionResults gets detailed results for a specific execution
// GET /api/v1/results/:executionId
func (h *ResultsHandler) GetExecutionResults(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	executionID := c.Params("executionId")
	if executionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "execution_id is required",
		})
	}

	// Get execution from repository
	execution, err := h.executionRepo.GetByID(executionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get execution",
		})
	}
	if execution == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "execution not found",
		})
	}

	// Check ownership
	if execution.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	// Convert to detailed DTO
	detail := executionToDetailDTO(execution)

	return c.JSON(detail)
}

// GetAggregatedMetrics gets aggregated metrics (tokens, costs, success rates)
// GET /api/v1/results/metrics
func (h *ResultsHandler) GetAggregatedMetrics(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Parse time range
	since := parseSinceParam(c)
	period := c.Query("period", "30d")

	// Get usage stats from repository
	stats, err := h.executionRepo.GetUsageStats(userID, since)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get aggregated metrics",
		})
	}

	// Get all executions for breakdown stats
	executions, err := h.executionRepo.ListByUserID(userID, 1000, 0)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get execution breakdown",
		})
	}

	// Filter by time range
	filteredExecutions := make([]*repository.AgentExecution, 0)
	for _, e := range executions {
		if e.CreatedAt.After(since) || e.CreatedAt.Equal(since) {
			filteredExecutions = append(filteredExecutions, e)
		}
	}

	// Calculate breakdowns
	statusBreakdown := make(map[string]int)
	providerBreakdown := make(map[string]int)
	modelBreakdown := make(map[string]int)
	successCount := 0

	for _, e := range filteredExecutions {
		statusBreakdown[e.Status]++
		providerBreakdown[e.Provider]++
		modelBreakdown[e.Model]++
		if e.Status == "completed" {
			successCount++
		}
	}

	totalExecs := len(filteredExecutions)
	successRate := 0.0
	avgTokens := 0.0
	avgCost := 0.0

	if totalExecs > 0 {
		successRate = float64(successCount) / float64(totalExecs) * 100
		avgTokens = float64(stats.TotalTokens) / float64(totalExecs)
		avgCost = stats.TotalCost / float64(totalExecs)
	}

	metrics := AggregatedMetricsDTO{
		TotalExecutions:       stats.TotalExecutions,
		TotalPromptTokens:     stats.TotalPromptTokens,
		TotalCompletionTokens: stats.TotalCompletionTokens,
		TotalTokens:           stats.TotalTokens,
		TotalCost:             stats.TotalCost,
		AvgTokensPerExecution: avgTokens,
		AvgCostPerExecution:   avgCost,
		StatusBreakdown:       statusBreakdown,
		ProviderBreakdown:     providerBreakdown,
		ModelBreakdown:        modelBreakdown,
		SuccessRate:           successRate,
		Period:                period,
	}

	return c.JSON(metrics)
}

// GetResultsTimeline gets time-series data for charts
// GET /api/v1/results/timeline
func (h *ResultsHandler) GetResultsTimeline(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Parse time range
	since := parseSinceParam(c)
	until := parseUntilParam(c)
	granularity := c.Query("granularity", "day")

	// Get executions from repository
	executions, err := h.executionRepo.ListByUserID(userID, 10000, 0)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get timeline data",
		})
	}

	// Filter by time range
	filteredExecutions := make([]*repository.AgentExecution, 0)
	for _, e := range executions {
		if (e.CreatedAt.After(since) || e.CreatedAt.Equal(since)) &&
			(e.CreatedAt.Before(until) || e.CreatedAt.Equal(until)) {
			filteredExecutions = append(filteredExecutions, e)
		}
	}

	// Aggregate into time buckets
	points := aggregateTimeline(filteredExecutions, since, until, granularity)

	response := TimelineResponseDTO{
		Points:      points,
		Granularity: granularity,
		Since:       since.UnixMilli(),
		Until:       until.UnixMilli(),
	}

	return c.JSON(response)
}

// queryParams holds parsed query parameters
type queryParams struct {
	since   time.Time
	until   time.Time
	status  string
	limit   int
	offset  int
	sort    string
	sortDir string
}

// parseResultsQueryParams parses common query parameters
func parseResultsQueryParams(c *fiber.Ctx) queryParams {
	params := queryParams{
		since:   parseSinceParam(c),
		until:   parseUntilParam(c),
		status:  c.Query("status"),
		sort:    c.Query("sort", "created_at"),
		sortDir: c.Query("sort_dir", "desc"),
	}

	// Parse limit
	limitStr := c.Query("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}
	params.limit = limit

	// Parse offset
	offsetStr := c.Query("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}
	params.offset = offset

	return params
}

// parseSinceParam parses the since query parameter
func parseSinceParam(c *fiber.Ctx) time.Time {
	sinceStr := c.Query("since")
	if sinceStr == "" {
		// Default to 30 days ago
		return time.Now().AddDate(0, 0, -30)
	}

	// Try parsing as Unix timestamp (milliseconds)
	if ts, err := strconv.ParseInt(sinceStr, 10, 64); err == nil {
		return time.UnixMilli(ts)
	}

	// Try parsing as RFC3339
	if t, err := time.Parse(time.RFC3339, sinceStr); err == nil {
		return t
	}

	// Default to 30 days ago
	return time.Now().AddDate(0, 0, -30)
}

// parseUntilParam parses the until query parameter
func parseUntilParam(c *fiber.Ctx) time.Time {
	untilStr := c.Query("until")
	if untilStr == "" {
		return time.Now()
	}

	// Try parsing as Unix timestamp (milliseconds)
	if ts, err := strconv.ParseInt(untilStr, 10, 64); err == nil {
		return time.UnixMilli(ts)
	}

	// Try parsing as RFC3339
	if t, err := time.Parse(time.RFC3339, untilStr); err == nil {
		return t
	}

	return time.Now()
}

// filterExecutions filters executions based on query params
func filterExecutions(executions []*repository.AgentExecution, params queryParams) []*repository.AgentExecution {
	filtered := make([]*repository.AgentExecution, 0)
	for _, e := range executions {
		// Filter by time range
		if e.CreatedAt.Before(params.since) {
			continue
		}
		if e.CreatedAt.After(params.until) {
			continue
		}

		// Filter by status
		if params.status != "" && e.Status != params.status {
			continue
		}

		filtered = append(filtered, e)
	}
	return filtered
}

// filterSwarmExecutions filters for swarm-type executions
func filterSwarmExecutions(executions []*repository.AgentExecution, params queryParams) []*repository.AgentExecution {
	filtered := make([]*repository.AgentExecution, 0)
	for _, e := range executions {
		// Check if execution is part of a swarm (via metadata)
		if e.Metadata != nil {
			if _, isSwarm := e.Metadata["swarm_id"]; isSwarm {
				// Apply standard filters
				if e.CreatedAt.Before(params.since) {
					continue
				}
				if e.CreatedAt.After(params.until) {
					continue
				}
				if params.status != "" && e.Status != params.status {
					continue
				}
				filtered = append(filtered, e)
			}
		}
	}
	return filtered
}

// aggregateBatchResults aggregates batch results into summary
func aggregateBatchResults(executions []*repository.AgentExecution, params queryParams) BatchResultsSummaryDTO {
	summary := BatchResultsSummaryDTO{
		Total:      len(executions),
		Executions: make([]ExecutionSummaryDTO, 0),
		Limit:      params.limit,
		Offset:     params.offset,
	}

	var totalDuration int64
	completedCount := 0

	for _, e := range executions {
		// Count by status
		switch e.Status {
		case "completed":
			summary.Completed++
			completedCount++
		case "failed":
			summary.Failed++
		case "running":
			summary.Running++
		case "pending":
			summary.Pending++
		case "cancelled":
			summary.Cancelled++
		}

		// Aggregate totals
		summary.TotalTokens += e.TotalTokens
		summary.TotalCost += e.TotalCost

		// Calculate duration
		if e.StartedAt != nil && e.CompletedAt != nil {
			totalDuration += e.CompletedAt.Sub(*e.StartedAt).Milliseconds()
		}

		// Add to executions list
		summary.Executions = append(summary.Executions, executionToSummaryDTO(e))
	}

	// Calculate averages
	if completedCount > 0 {
		summary.AvgDurationMs = totalDuration / int64(completedCount)
	}
	if summary.Total > 0 {
		summary.SuccessRate = float64(summary.Completed) / float64(summary.Total) * 100
	}

	return summary
}

// aggregateSwarmResults aggregates swarm results into summary
func aggregateSwarmResults(executions []*repository.AgentExecution, params queryParams) SwarmResultsSummaryDTO {
	summary := SwarmResultsSummaryDTO{
		Total:      len(executions),
		Executions: make([]ExecutionSummaryDTO, 0),
		Limit:      params.limit,
		Offset:     params.offset,
	}

	var totalDuration int64
	completedCount := 0

	for _, e := range executions {
		// Count by status
		switch e.Status {
		case "completed":
			summary.Completed++
			completedCount++
		case "failed":
			summary.Failed++
		case "running":
			summary.Running++
		case "pending":
			summary.Pending++
		case "cancelled":
			summary.Cancelled++
		}

		// Aggregate totals
		summary.TotalTokens += e.TotalTokens
		summary.TotalCost += e.TotalCost

		// Calculate duration
		if e.StartedAt != nil && e.CompletedAt != nil {
			totalDuration += e.CompletedAt.Sub(*e.StartedAt).Milliseconds()
		}

		// Add to executions list
		summary.Executions = append(summary.Executions, executionToSummaryDTO(e))
	}

	// Calculate averages
	if completedCount > 0 {
		summary.AvgDurationMs = totalDuration / int64(completedCount)
	}
	if summary.Total > 0 {
		summary.SuccessRate = float64(summary.Completed) / float64(summary.Total) * 100
	}

	return summary
}

// aggregateTimeline aggregates executions into timeline points
func aggregateTimeline(executions []*repository.AgentExecution, since, until time.Time, granularity string) []TimelinePointDTO {
	// Determine bucket size
	var bucketSize time.Duration
	var dateFormat string

	switch granularity {
	case "hour":
		bucketSize = time.Hour
		dateFormat = "2006-01-02 15:00"
	case "day":
		bucketSize = 24 * time.Hour
		dateFormat = "2006-01-02"
	case "week":
		bucketSize = 7 * 24 * time.Hour
		dateFormat = "2006-01-02"
	case "month":
		bucketSize = 30 * 24 * time.Hour
		dateFormat = "2006-01"
	default:
		bucketSize = 24 * time.Hour
		dateFormat = "2006-01-02"
	}

	// Create buckets map
	buckets := make(map[int64]*TimelinePointDTO)

	// Populate buckets with executions
	for _, e := range executions {
		// Truncate to bucket start
		bucketStart := e.CreatedAt.Truncate(bucketSize)
		ts := bucketStart.UnixMilli()

		if buckets[ts] == nil {
			buckets[ts] = &TimelinePointDTO{
				Timestamp: ts,
				Date:      bucketStart.Format(dateFormat),
			}
		}

		point := buckets[ts]
		point.ExecutionCount++
		point.TotalTokens += e.TotalTokens
		point.TotalCost += e.TotalCost

		if e.Status == "completed" {
			point.SuccessCount++
		} else if e.Status == "failed" {
			point.FailureCount++
		}

		// Add duration for averaging
		if e.StartedAt != nil && e.CompletedAt != nil {
			duration := e.CompletedAt.Sub(*e.StartedAt).Milliseconds()
			// Running average
			if point.ExecutionCount > 1 {
				point.AvgDurationMs = (point.AvgDurationMs*int64(point.ExecutionCount-1) + duration) / int64(point.ExecutionCount)
			} else {
				point.AvgDurationMs = duration
			}
		}
	}

	// Convert map to sorted slice
	points := make([]TimelinePointDTO, 0, len(buckets))

	// Generate all buckets in range (including empty ones)
	current := since.Truncate(bucketSize)
	for current.Before(until) || current.Equal(until) {
		ts := current.UnixMilli()
		if buckets[ts] != nil {
			points = append(points, *buckets[ts])
		} else {
			points = append(points, TimelinePointDTO{
				Timestamp: ts,
				Date:      current.Format(dateFormat),
			})
		}
		current = current.Add(bucketSize)
	}

	return points
}

// executionToSummaryDTO converts an execution to a summary DTO
func executionToSummaryDTO(e *repository.AgentExecution) ExecutionSummaryDTO {
	dto := ExecutionSummaryDTO{
		ID:               e.ID,
		UserID:           e.UserID,
		Provider:         e.Provider,
		LLMProvider:      e.LLMProvider,
		Model:            e.Model,
		AgentName:        e.AgentName,
		Status:           e.Status,
		PromptTokens:     e.PromptTokens,
		CompletionTokens: e.CompletionTokens,
		TotalTokens:      e.TotalTokens,
		InputCost:        e.InputCost,
		OutputCost:       e.OutputCost,
		TotalCost:        e.TotalCost,
		Currency:         e.Currency,
		Error:            e.Error,
		CreatedAt:        e.CreatedAt.UnixMilli(),
		Metadata:         e.Metadata,
	}

	if e.StartedAt != nil {
		ts := e.StartedAt.UnixMilli()
		dto.StartedAt = &ts
	}

	if e.CompletedAt != nil {
		ts := e.CompletedAt.UnixMilli()
		dto.CompletedAt = &ts
	}

	// Calculate duration
	if e.StartedAt != nil && e.CompletedAt != nil {
		dto.DurationMs = e.CompletedAt.Sub(*e.StartedAt).Milliseconds()
	}

	return dto
}

// executionToDetailDTO converts an execution to a detailed DTO
func executionToDetailDTO(e *repository.AgentExecution) ExecutionDetailDTO {
	return ExecutionDetailDTO{
		ExecutionSummaryDTO: executionToSummaryDTO(e),
		Messages:            []ExecutionMessageDTO{},
		ToolCalls:           []ExecutionToolCallDTO{},
	}
}
