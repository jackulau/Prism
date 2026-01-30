package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/audit"
)

// AuditLogEntry represents a row in the audit_logs table
type AuditLogEntry struct {
	ID           string
	Timestamp    time.Time
	ActorID      string
	ActorEmail   string
	ActorType    string
	Action       string
	ResourceType string
	ResourceID   string
	ResourceName string
	IPAddress    string
	UserAgent    string
	SessionID    string
	OrgID        string
	Metadata     string
	BeforeState  string
	AfterState   string
	Success      bool
	ErrorMessage string
	LegalHold    bool
}

// AuditRepository handles audit log database operations
type AuditRepository struct {
	db *sql.DB
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create stores a new audit event
func (r *AuditRepository) Create(event *audit.AuditEvent) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	metadataJSON := "{}"
	if event.Metadata != nil {
		if data, err := json.Marshal(event.Metadata); err == nil {
			metadataJSON = string(data)
		}
	}

	beforeState := ""
	if event.BeforeState != nil {
		beforeState = string(event.BeforeState)
	}

	afterState := ""
	if event.AfterState != nil {
		afterState = string(event.AfterState)
	}

	_, err := r.db.Exec(`
		INSERT INTO audit_logs (
			id, timestamp, actor_id, actor_email, actor_type, action,
			resource_type, resource_id, resource_name, ip_address,
			user_agent, session_id, organization_id, metadata,
			before_state, after_state, success, error_message, legal_hold
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)`,
		event.ID, event.Timestamp, event.ActorID, event.ActorEmail, event.ActorType,
		string(event.Action), string(event.ResourceType), event.ResourceID, event.ResourceName,
		event.IPAddress, event.UserAgent, event.SessionID, event.OrgID,
		metadataJSON, beforeState, afterState, event.Success, event.ErrorMessage,
	)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// List retrieves audit events based on filter criteria
func (r *AuditRepository) List(filter audit.AuditFilter) ([]*audit.AuditEvent, int64, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.ActorID != "" {
		conditions = append(conditions, fmt.Sprintf("actor_id = ?%d", argIndex))
		args = append(args, filter.ActorID)
		argIndex++
	}

	if filter.OrgID != "" {
		conditions = append(conditions, fmt.Sprintf("organization_id = ?%d", argIndex))
		args = append(args, filter.OrgID)
		argIndex++
	}

	if filter.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = ?%d", argIndex))
		args = append(args, string(filter.Action))
		argIndex++
	}

	if filter.ResourceType != "" {
		conditions = append(conditions, fmt.Sprintf("resource_type = ?%d", argIndex))
		args = append(args, string(filter.ResourceType))
		argIndex++
	}

	if filter.ResourceID != "" {
		conditions = append(conditions, fmt.Sprintf("resource_id = ?%d", argIndex))
		args = append(args, filter.ResourceID)
		argIndex++
	}

	if filter.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp >= ?%d", argIndex))
		args = append(args, *filter.StartTime)
		argIndex++
	}

	if filter.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp <= ?%d", argIndex))
		args = append(args, *filter.EndTime)
		argIndex++
	}

	if filter.Success != nil {
		conditions = append(conditions, fmt.Sprintf("success = ?%d", argIndex))
		args = append(args, *filter.Success)
		argIndex++
	}

	if filter.IPAddress != "" {
		conditions = append(conditions, fmt.Sprintf("ip_address = ?%d", argIndex))
		args = append(args, filter.IPAddress)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Convert positional placeholders to SQLite style
	query := fmt.Sprintf(`SELECT COUNT(*) FROM audit_logs %s`, whereClause)
	query = convertPlaceholders(query)

	var total int64
	err := r.db.QueryRow(query, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	if filter.Limit <= 0 {
		filter.Limit = 100
	}

	selectQuery := fmt.Sprintf(`
		SELECT id, timestamp, actor_id, actor_email, actor_type, action,
			   resource_type, resource_id, resource_name, ip_address,
			   user_agent, session_id, organization_id, metadata,
			   before_state, after_state, success, error_message
		FROM audit_logs %s
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?`, whereClause)
	selectQuery = convertPlaceholders(selectQuery)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var events []*audit.AuditEvent
	for rows.Next() {
		var entry AuditLogEntry
		var actorEmail, resourceName, ipAddress, userAgent, sessionID, orgID sql.NullString
		var metadataJSON, beforeState, afterState sql.NullString
		var errorMsg sql.NullString

		err := rows.Scan(
			&entry.ID, &entry.Timestamp, &entry.ActorID, &actorEmail, &entry.ActorType,
			&entry.Action, &entry.ResourceType, &entry.ResourceID, &resourceName,
			&ipAddress, &userAgent, &sessionID, &orgID, &metadataJSON,
			&beforeState, &afterState, &entry.Success, &errorMsg,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}

		event := &audit.AuditEvent{
			ID:           entry.ID,
			Timestamp:    entry.Timestamp,
			ActorID:      entry.ActorID,
			ActorEmail:   actorEmail.String,
			ActorType:    entry.ActorType,
			Action:       audit.ActionType(entry.Action),
			ResourceType: audit.ResourceType(entry.ResourceType),
			ResourceID:   entry.ResourceID,
			ResourceName: resourceName.String,
			IPAddress:    ipAddress.String,
			UserAgent:    userAgent.String,
			SessionID:    sessionID.String,
			OrgID:        orgID.String,
			Success:      entry.Success,
			ErrorMessage: errorMsg.String,
		}

		if metadataJSON.Valid && metadataJSON.String != "" {
			var meta map[string]interface{}
			if err := json.Unmarshal([]byte(metadataJSON.String), &meta); err == nil {
				event.Metadata = meta
			}
		}

		if beforeState.Valid && beforeState.String != "" {
			event.BeforeState = json.RawMessage(beforeState.String)
		}

		if afterState.Valid && afterState.String != "" {
			event.AfterState = json.RawMessage(afterState.String)
		}

		events = append(events, event)
	}

	return events, total, nil
}

// GetByID retrieves a single audit event by ID
func (r *AuditRepository) GetByID(id string) (*audit.AuditEvent, error) {
	var entry AuditLogEntry
	var actorEmail, resourceName, ipAddress, userAgent, sessionID, orgID sql.NullString
	var metadataJSON, beforeState, afterState sql.NullString
	var errorMsg sql.NullString

	err := r.db.QueryRow(`
		SELECT id, timestamp, actor_id, actor_email, actor_type, action,
			   resource_type, resource_id, resource_name, ip_address,
			   user_agent, session_id, organization_id, metadata,
			   before_state, after_state, success, error_message
		FROM audit_logs WHERE id = ?`, id,
	).Scan(
		&entry.ID, &entry.Timestamp, &entry.ActorID, &actorEmail, &entry.ActorType,
		&entry.Action, &entry.ResourceType, &entry.ResourceID, &resourceName,
		&ipAddress, &userAgent, &sessionID, &orgID, &metadataJSON,
		&beforeState, &afterState, &entry.Success, &errorMsg,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	event := &audit.AuditEvent{
		ID:           entry.ID,
		Timestamp:    entry.Timestamp,
		ActorID:      entry.ActorID,
		ActorEmail:   actorEmail.String,
		ActorType:    entry.ActorType,
		Action:       audit.ActionType(entry.Action),
		ResourceType: audit.ResourceType(entry.ResourceType),
		ResourceID:   entry.ResourceID,
		ResourceName: resourceName.String,
		IPAddress:    ipAddress.String,
		UserAgent:    userAgent.String,
		SessionID:    sessionID.String,
		OrgID:        orgID.String,
		Success:      entry.Success,
		ErrorMessage: errorMsg.String,
	}

	if metadataJSON.Valid && metadataJSON.String != "" {
		var meta map[string]interface{}
		if err := json.Unmarshal([]byte(metadataJSON.String), &meta); err == nil {
			event.Metadata = meta
		}
	}

	if beforeState.Valid && beforeState.String != "" {
		event.BeforeState = json.RawMessage(beforeState.String)
	}

	if afterState.Valid && afterState.String != "" {
		event.AfterState = json.RawMessage(afterState.String)
	}

	return event, nil
}

// DeleteBefore deletes audit logs older than the specified timestamp
// If excludeLegalHold is true, records with legal_hold=true are not deleted
func (r *AuditRepository) DeleteBefore(timestamp time.Time, excludeLegalHold bool) (int64, error) {
	var result sql.Result
	var err error

	if excludeLegalHold {
		result, err = r.db.Exec(
			`DELETE FROM audit_logs WHERE timestamp < ? AND legal_hold = 0`,
			timestamp,
		)
	} else {
		result, err = r.db.Exec(
			`DELETE FROM audit_logs WHERE timestamp < ?`,
			timestamp,
		)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to delete audit logs: %w", err)
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return deleted, nil
}

// SetLegalHold sets or clears the legal hold flag for specific audit logs
func (r *AuditRepository) SetLegalHold(ids []string, hold bool) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+1)
	args[0] = hold
	for i, id := range ids {
		placeholders[i] = "?"
		args[i+1] = id
	}

	_, err := r.db.Exec(
		fmt.Sprintf(`UPDATE audit_logs SET legal_hold = ? WHERE id IN (%s)`, strings.Join(placeholders, ",")),
		args...,
	)
	if err != nil {
		return fmt.Errorf("failed to set legal hold: %w", err)
	}

	return nil
}

// SetLegalHoldByDateRange sets legal hold for all logs in a date range
func (r *AuditRepository) SetLegalHoldByDateRange(start, end time.Time, hold bool, orgID string) (int64, error) {
	var result sql.Result
	var err error

	if orgID != "" {
		result, err = r.db.Exec(
			`UPDATE audit_logs SET legal_hold = ? WHERE timestamp >= ? AND timestamp <= ? AND organization_id = ?`,
			hold, start, end, orgID,
		)
	} else {
		result, err = r.db.Exec(
			`UPDATE audit_logs SET legal_hold = ? WHERE timestamp >= ? AND timestamp <= ?`,
			hold, start, end,
		)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to set legal hold by date range: %w", err)
	}

	affected, _ := result.RowsAffected()
	return affected, nil
}

// GetStats returns statistics about audit logs
func (r *AuditRepository) GetStats(orgID string, startTime, endTime *time.Time) (*AuditStats, error) {
	var conditions []string
	var args []interface{}

	if orgID != "" {
		conditions = append(conditions, "organization_id = ?")
		args = append(args, orgID)
	}

	if startTime != nil {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, *startTime)
	}

	if endTime != nil {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, *endTime)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	stats := &AuditStats{
		ActionCounts:   make(map[string]int64),
		ResourceCounts: make(map[string]int64),
	}

	// Total count
	err := r.db.QueryRow(
		fmt.Sprintf(`SELECT COUNT(*) FROM audit_logs %s`, whereClause),
		args...,
	).Scan(&stats.TotalEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Failed count
	failedArgs := append(args, false)
	failedWhere := whereClause
	if whereClause == "" {
		failedWhere = "WHERE success = ?"
	} else {
		failedWhere = whereClause + " AND success = ?"
	}
	err = r.db.QueryRow(
		fmt.Sprintf(`SELECT COUNT(*) FROM audit_logs %s`, failedWhere),
		failedArgs...,
	).Scan(&stats.FailedEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed count: %w", err)
	}

	// Action counts
	rows, err := r.db.Query(
		fmt.Sprintf(`SELECT action, COUNT(*) FROM audit_logs %s GROUP BY action`, whereClause),
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get action counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var action string
		var count int64
		if err := rows.Scan(&action, &count); err == nil {
			stats.ActionCounts[action] = count
		}
	}

	// Resource type counts
	rows, err = r.db.Query(
		fmt.Sprintf(`SELECT resource_type, COUNT(*) FROM audit_logs %s GROUP BY resource_type`, whereClause),
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var resourceType string
		var count int64
		if err := rows.Scan(&resourceType, &count); err == nil {
			stats.ResourceCounts[resourceType] = count
		}
	}

	// Unique actors count
	err = r.db.QueryRow(
		fmt.Sprintf(`SELECT COUNT(DISTINCT actor_id) FROM audit_logs %s`, whereClause),
		args...,
	).Scan(&stats.UniqueActors)
	if err != nil {
		return nil, fmt.Errorf("failed to get unique actors: %w", err)
	}

	return stats, nil
}

// AuditStats represents aggregate statistics for audit logs
type AuditStats struct {
	TotalEvents    int64            `json:"total_events"`
	FailedEvents   int64            `json:"failed_events"`
	ActionCounts   map[string]int64 `json:"action_counts"`
	ResourceCounts map[string]int64 `json:"resource_counts"`
	UniqueActors   int64            `json:"unique_actors"`
}

// convertPlaceholders converts ?N style placeholders to ? for SQLite
func convertPlaceholders(query string) string {
	// SQLite uses ? for positional parameters, not ?N
	// This is a simple conversion
	for i := 1; i <= 20; i++ {
		query = strings.ReplaceAll(query, fmt.Sprintf("?%d", i), "?")
	}
	return query
}
