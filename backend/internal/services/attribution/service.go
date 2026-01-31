package attribution

import (
	"context"
	"time"

	"github.com/jacklau/prism/internal/agent"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/types"
)

// Service provides attribution functionality for tracking file changes
type Service struct {
	fileHistoryRepo *repository.FileHistoryRepository
}

// NewService creates a new attribution service
func NewService(fileHistoryRepo *repository.FileHistoryRepository) *Service {
	return &Service{
		fileHistoryRepo: fileHistoryRepo,
	}
}

// BuildAttributionContext creates an AttributionContext from the current execution context
func (s *Service) BuildAttributionContext(ctx context.Context, toolName string) *types.AttributionContext {
	attr := agent.BuildAttributionFromContext(ctx)

	// Set tool name if provided and not already set
	if toolName != "" && attr.ToolName == "" {
		attr.ToolName = toolName
	}

	return attr
}

// RecordFileChange creates a file history entry with full attribution
func (s *Service) RecordFileChange(
	userID, filePath, content, operation string,
	ctx context.Context,
	toolName string,
) (*repository.FileHistory, error) {
	attr := s.BuildAttributionContext(ctx, toolName)
	return s.fileHistoryRepo.CreateWithAttribution(userID, filePath, content, operation, attr)
}

// RecordFileChangeWithAttribution creates a file history entry with explicit attribution
func (s *Service) RecordFileChangeWithAttribution(
	userID, filePath, content, operation string,
	attr *types.AttributionContext,
) (*repository.FileHistory, error) {
	return s.fileHistoryRepo.CreateWithAttribution(userID, filePath, content, operation, attr)
}

// EnrichHistoryWithAttribution adds attribution to an existing history entry
func (s *Service) EnrichHistoryWithAttribution(
	historyID string,
	attr *types.AttributionContext,
) error {
	return s.fileHistoryRepo.UpdateAttribution(historyID, attr)
}

// GetAttributionSummary returns aggregated attribution data for a date range
func (s *Service) GetAttributionSummary(
	userID string,
	startDate, endDate time.Time,
) (*types.AttributionSummary, error) {
	return s.fileHistoryRepo.GetAttributionSummary(userID, startDate, endDate)
}

// GetAgentActivityReport generates a detailed report of an agent's file modifications
func (s *Service) GetAgentActivityReport(
	userID, agentID string,
) (*types.AgentActivityReport, error) {
	// Get all history entries for this agent
	history, err := s.fileHistoryRepo.ListByAgent(userID, agentID, 1000, 0)
	if err != nil {
		return nil, err
	}

	if len(history) == 0 {
		return &types.AgentActivityReport{
			AgentID:        agentID,
			TotalChanges:   0,
			FileChanges:    []types.FileChangeInfo{},
			OperationStats: make(map[string]int),
			TopFiles:       []types.FileActivityInfo{},
		}, nil
	}

	report := &types.AgentActivityReport{
		AgentID:        agentID,
		TotalChanges:   len(history),
		FileChanges:    make([]types.FileChangeInfo, 0, len(history)),
		OperationStats: make(map[string]int),
		TopFiles:       []types.FileActivityInfo{},
	}

	// Set agent name from first entry
	if history[0].AgentName != nil {
		report.AgentName = *history[0].AgentName
	}

	// Track file activity
	fileActivity := make(map[string]int)
	var firstChange, lastChange time.Time

	for _, h := range history {
		// Build file change info
		change := types.FileChangeInfo{
			FilePath:  h.FilePath,
			Operation: h.Operation,
			Timestamp: h.CreatedAt,
		}
		if h.ToolName != nil {
			change.ToolName = *h.ToolName
		}
		if h.ConversationID != nil {
			change.ConversationID = *h.ConversationID
		}
		report.FileChanges = append(report.FileChanges, change)

		// Update operation stats
		report.OperationStats[h.Operation]++

		// Track file activity
		fileActivity[h.FilePath]++

		// Track active period
		if firstChange.IsZero() || h.CreatedAt.Before(firstChange) {
			firstChange = h.CreatedAt
		}
		if lastChange.IsZero() || h.CreatedAt.After(lastChange) {
			lastChange = h.CreatedAt
		}
	}

	// Set active period
	if !firstChange.IsZero() {
		report.ActivePeriod = &types.ActivePeriod{
			FirstChange: firstChange,
			LastChange:  lastChange,
		}
	}

	// Build top files list (sorted by activity)
	type fileCount struct {
		path  string
		count int
	}
	var fileCounts []fileCount
	for path, count := range fileActivity {
		fileCounts = append(fileCounts, fileCount{path, count})
	}

	// Simple sort by count descending
	for i := 0; i < len(fileCounts); i++ {
		for j := i + 1; j < len(fileCounts); j++ {
			if fileCounts[j].count > fileCounts[i].count {
				fileCounts[i], fileCounts[j] = fileCounts[j], fileCounts[i]
			}
		}
	}

	// Take top 10
	topCount := 10
	if len(fileCounts) < topCount {
		topCount = len(fileCounts)
	}
	for i := 0; i < topCount; i++ {
		report.TopFiles = append(report.TopFiles, types.FileActivityInfo{
			FilePath:    fileCounts[i].path,
			ChangeCount: fileCounts[i].count,
		})
	}

	return report, nil
}

// GetConversationChanges returns all file changes made during a conversation
func (s *Service) GetConversationChanges(userID, conversationID string) ([]*repository.FileHistory, error) {
	return s.fileHistoryRepo.ListByConversation(userID, conversationID)
}

// GetWorkflowChanges returns all file changes made during a workflow execution
func (s *Service) GetWorkflowChanges(userID, workflowID string) ([]*repository.FileHistory, error) {
	return s.fileHistoryRepo.ListByWorkflow(userID, workflowID)
}

// GetToolChanges returns all file changes made by a specific tool
func (s *Service) GetToolChanges(userID, toolName string, limit, offset int) ([]*repository.FileHistory, error) {
	return s.fileHistoryRepo.ListByTool(userID, toolName, limit, offset)
}

// GetRecentActivityTimeline returns a timeline of recent changes grouped by time period
func (s *Service) GetRecentActivityTimeline(userID string, days int) (*types.AttributionSummary, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)
	return s.GetAttributionSummary(userID, startDate, endDate)
}
