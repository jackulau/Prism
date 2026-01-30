package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Log represents an audit log entry in the database
type Log struct {
	ID            int64
	UserID        *string
	EventType     string
	EventCategory string
	Action        string
	ResourceType  *string
	ResourceID    *string
	IPAddress     *string
	UserAgent     *string
	Details       map[string]interface{}
	Success       bool
	CreatedAt     time.Time
}

// QueryOptions represents options for querying audit logs
type QueryOptions struct {
	UserID       *string
	EventType    *EventType
	Category     *EventCategory
	StartTime    *time.Time
	EndTime      *time.Time
	Success      *bool
	ResourceType *string
	ResourceID   *string
	Limit        int
	Offset       int
}

// Stats represents audit log statistics
type Stats struct {
	Since          time.Time               `json:"since"`
	CategoryCounts map[EventCategory]int64 `json:"category_counts"`
	AuthCounts     map[EventType]int64     `json:"auth_counts"`
	ProviderCounts map[EventType]int64     `json:"provider_counts"`
}

// Repository handles audit log database operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new audit repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new audit log entry
func (r *Repository) Create(entry *Log) error {
	var detailsJSON []byte
	var err error
	if entry.Details != nil {
		detailsJSON, err = json.Marshal(entry.Details)
		if err != nil {
			return fmt.Errorf("failed to marshal details: %w", err)
		}
	}

	successInt := 0
	if entry.Success {
		successInt = 1
	}

	result, err := r.db.Exec(
		`INSERT INTO audit_logs (user_id, event_type, event_category, action, resource_type, resource_id, ip_address, user_agent, details, success, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.UserID,
		entry.EventType,
		entry.EventCategory,
		entry.Action,
		entry.ResourceType,
		entry.ResourceID,
		entry.IPAddress,
		entry.UserAgent,
		string(detailsJSON),
		successInt,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	entry.ID = id

	return nil
}

// Query retrieves audit logs based on query options
func (r *Repository) Query(opts QueryOptions) ([]Log, int64, error) {
	var conditions []string
	var args []interface{}

	if opts.UserID != nil {
		conditions = append(conditions, "user_id = ?")
		args = append(args, *opts.UserID)
	}
	if opts.EventType != nil {
		conditions = append(conditions, "event_type = ?")
		args = append(args, string(*opts.EventType))
	}
	if opts.Category != nil {
		conditions = append(conditions, "event_category = ?")
		args = append(args, string(*opts.Category))
	}
	if opts.StartTime != nil {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, *opts.StartTime)
	}
	if opts.EndTime != nil {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, *opts.EndTime)
	}
	if opts.Success != nil {
		successInt := 0
		if *opts.Success {
			successInt = 1
		}
		conditions = append(conditions, "success = ?")
		args = append(args, successInt)
	}
	if opts.ResourceType != nil {
		conditions = append(conditions, "resource_type = ?")
		args = append(args, *opts.ResourceType)
	}
	if opts.ResourceID != nil {
		conditions = append(conditions, "resource_id = ?")
		args = append(args, *opts.ResourceID)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs %s", whereClause)
	var total int64
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Set default limit
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	// Get paginated results
	query := fmt.Sprintf(
		`SELECT id, user_id, event_type, event_category, action, resource_type, resource_id, ip_address, user_agent, details, success, created_at
		 FROM audit_logs %s ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		whereClause,
	)
	args = append(args, limit, opts.Offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []Log
	for rows.Next() {
		logEntry, err := r.scanLog(rows)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, *logEntry)
	}

	return logs, total, nil
}

// GetByUserID retrieves audit logs for a specific user
func (r *Repository) GetByUserID(userID string, limit, offset int) ([]Log, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	rows, err := r.db.Query(
		`SELECT id, user_id, event_type, event_category, action, resource_type, resource_id, ip_address, user_agent, details, success, created_at
		 FROM audit_logs WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []Log
	for rows.Next() {
		logEntry, err := r.scanLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, *logEntry)
	}

	return logs, nil
}

// GetRecent retrieves the most recent audit logs
func (r *Repository) GetRecent(limit int) ([]Log, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	rows, err := r.db.Query(
		`SELECT id, user_id, event_type, event_category, action, resource_type, resource_id, ip_address, user_agent, details, success, created_at
		 FROM audit_logs ORDER BY created_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent audit logs: %w", err)
	}
	defer rows.Close()

	var logs []Log
	for rows.Next() {
		logEntry, err := r.scanLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, *logEntry)
	}

	return logs, nil
}

// DeleteOlderThan deletes audit logs older than the specified date
func (r *Repository) DeleteOlderThan(date time.Time) (int64, error) {
	result, err := r.db.Exec("DELETE FROM audit_logs WHERE created_at < ?", date)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old audit logs: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get deleted count: %w", err)
	}

	return count, nil
}

// CountByEventType counts audit logs by event type within a time range
func (r *Repository) CountByEventType(category EventCategory, since time.Time) (map[EventType]int64, error) {
	rows, err := r.db.Query(
		`SELECT event_type, COUNT(*) as count FROM audit_logs
		 WHERE event_category = ? AND created_at >= ?
		 GROUP BY event_type`,
		string(category), since,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to count audit logs by event type: %w", err)
	}
	defer rows.Close()

	counts := make(map[EventType]int64)
	for rows.Next() {
		var eventType string
		var count int64
		if err := rows.Scan(&eventType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan count: %w", err)
		}
		counts[EventType(eventType)] = count
	}

	return counts, nil
}

// CountByCategory counts audit logs by category within a time range
func (r *Repository) CountByCategory(since time.Time) (map[EventCategory]int64, error) {
	rows, err := r.db.Query(
		`SELECT event_category, COUNT(*) as count FROM audit_logs
		 WHERE created_at >= ?
		 GROUP BY event_category`,
		since,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to count audit logs by category: %w", err)
	}
	defer rows.Close()

	counts := make(map[EventCategory]int64)
	for rows.Next() {
		var category string
		var count int64
		if err := rows.Scan(&category, &count); err != nil {
			return nil, fmt.Errorf("failed to scan count: %w", err)
		}
		counts[EventCategory(category)] = count
	}

	return counts, nil
}

// scanLog scans a single audit log row
func (r *Repository) scanLog(rows *sql.Rows) (*Log, error) {
	var logEntry Log
	var userID, resourceType, resourceID, ipAddress, userAgent, detailsJSON sql.NullString
	var successInt int

	err := rows.Scan(
		&logEntry.ID,
		&userID,
		&logEntry.EventType,
		&logEntry.EventCategory,
		&logEntry.Action,
		&resourceType,
		&resourceID,
		&ipAddress,
		&userAgent,
		&detailsJSON,
		&successInt,
		&logEntry.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan audit log: %w", err)
	}

	if userID.Valid {
		logEntry.UserID = &userID.String
	}
	if resourceType.Valid {
		logEntry.ResourceType = &resourceType.String
	}
	if resourceID.Valid {
		logEntry.ResourceID = &resourceID.String
	}
	if ipAddress.Valid {
		logEntry.IPAddress = &ipAddress.String
	}
	if userAgent.Valid {
		logEntry.UserAgent = &userAgent.String
	}
	if detailsJSON.Valid && detailsJSON.String != "" {
		if err := json.Unmarshal([]byte(detailsJSON.String), &logEntry.Details); err != nil {
			// Log error but don't fail - details are optional
			logEntry.Details = nil
		}
	}
	logEntry.Success = successInt == 1

	return &logEntry, nil
}

// Service provides audit logging functionality
type Service struct {
	repo *Repository
}

// NewService creates a new audit service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Log records an audit log entry
func (s *Service) Log(ctx context.Context, entry Entry) error {
	auditLog := &Log{
		UserID:        entry.UserID,
		EventType:     string(entry.EventType),
		EventCategory: string(entry.Category),
		Action:        entry.Action,
		Details:       entry.Details,
		Success:       entry.Success,
	}

	if entry.ResourceType != "" {
		auditLog.ResourceType = &entry.ResourceType
	}
	if entry.ResourceID != "" {
		auditLog.ResourceID = &entry.ResourceID
	}
	if entry.IPAddress != "" {
		auditLog.IPAddress = &entry.IPAddress
	}
	if entry.UserAgent != "" {
		auditLog.UserAgent = &entry.UserAgent
	}

	if err := s.repo.Create(auditLog); err != nil {
		log.Printf("Failed to create audit log: %v", err)
		return err
	}

	return nil
}

// LogFromRequest records an audit log entry with request metadata extracted from Fiber context
func (s *Service) LogFromRequest(c *fiber.Ctx, entry Entry) error {
	// Extract IP address
	ip := c.IP()
	if ip == "" {
		ip = c.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = c.Get("X-Real-IP")
	}
	entry.IPAddress = ip

	// Extract User-Agent
	entry.UserAgent = c.Get("User-Agent")

	return s.Log(context.Background(), entry)
}

// LogAsync records an audit log entry asynchronously (fire and forget)
func (s *Service) LogAsync(entry Entry) {
	go func() {
		if err := s.Log(context.Background(), entry); err != nil {
			log.Printf("Async audit log failed: %v", err)
		}
	}()
}

// LogFromRequestAsync records an audit log entry with request metadata asynchronously
func (s *Service) LogFromRequestAsync(c *fiber.Ctx, entry Entry) {
	// Extract IP address
	ip := c.IP()
	if ip == "" {
		ip = c.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = c.Get("X-Real-IP")
	}
	entry.IPAddress = ip

	// Extract User-Agent
	entry.UserAgent = c.Get("User-Agent")

	s.LogAsync(entry)
}

// Query retrieves audit logs based on query options
func (s *Service) Query(opts QueryOptions) ([]Log, int64, error) {
	return s.repo.Query(opts)
}

// GetUserLogs retrieves audit logs for a specific user
func (s *Service) GetUserLogs(userID string, limit, offset int) ([]Log, error) {
	return s.repo.GetByUserID(userID, limit, offset)
}

// GetRecentLogs retrieves the most recent audit logs
func (s *Service) GetRecentLogs(limit int) ([]Log, error) {
	return s.repo.GetRecent(limit)
}

// GetStats retrieves audit log statistics
func (s *Service) GetStats(since time.Time) (*Stats, error) {
	categoryCounts, err := s.repo.CountByCategory(since)
	if err != nil {
		return nil, err
	}

	authCounts, err := s.repo.CountByEventType(CategoryAuth, since)
	if err != nil {
		return nil, err
	}

	providerCounts, err := s.repo.CountByEventType(CategoryProvider, since)
	if err != nil {
		return nil, err
	}

	return &Stats{
		Since:          since,
		CategoryCounts: categoryCounts,
		AuthCounts:     authCounts,
		ProviderCounts: providerCounts,
	}, nil
}

// RotateLogs deletes audit logs older than the specified number of days
func (s *Service) RotateLogs(retentionDays int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	return s.repo.DeleteOlderThan(cutoffDate)
}

// Helper functions for creating common audit entries

// NewAuthEntry creates an audit entry for authentication events
func NewAuthEntry(userID *string, eventType EventType, action string, success bool) Entry {
	return Entry{
		UserID:    userID,
		EventType: eventType,
		Category:  CategoryAuth,
		Action:    action,
		Success:   success,
	}
}

// NewProviderEntry creates an audit entry for provider events
func NewProviderEntry(userID *string, eventType EventType, provider string, success bool) Entry {
	return Entry{
		UserID:       userID,
		EventType:    eventType,
		Category:     CategoryProvider,
		Action:       string(eventType),
		ResourceType: "provider",
		ResourceID:   provider,
		Success:      success,
	}
}

// NewSessionEntry creates an audit entry for session events
func NewSessionEntry(userID *string, eventType EventType, sessionID string, success bool) Entry {
	return Entry{
		UserID:       userID,
		EventType:    eventType,
		Category:     CategorySession,
		Action:       string(eventType),
		ResourceType: "session",
		ResourceID:   sessionID,
		Success:      success,
	}
}

// NewSettingsEntry creates an audit entry for settings changes
func NewSettingsEntry(userID *string, settingName string, details map[string]interface{}) Entry {
	return Entry{
		UserID:       userID,
		EventType:    EventSettingsChanged,
		Category:     CategorySettings,
		Action:       "settings_changed",
		ResourceType: "setting",
		ResourceID:   settingName,
		Details:      details,
		Success:      true,
	}
}
