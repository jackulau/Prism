package export

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jacklau/prism/internal/audit"
)

// ComplianceReportType represents the type of compliance report
type ComplianceReportType string

const (
	ReportTypeAccessLog   ComplianceReportType = "access_log"
	ReportTypeChangeLog   ComplianceReportType = "change_log"
	ReportTypeSecurityLog ComplianceReportType = "security_log"
	ReportTypeUserActivity ComplianceReportType = "user_activity"
	ReportTypeSOC2        ComplianceReportType = "soc2"
	ReportTypeHIPAA       ComplianceReportType = "hipaa"
)

// ComplianceReportRequest defines parameters for a compliance report
type ComplianceReportRequest struct {
	ReportType  ComplianceReportType
	OrgID       string
	StartDate   time.Time
	EndDate     time.Time
	UserIDs     []string // Filter to specific users (optional)
	IncludeDetails bool   // Include full event details
}

// ComplianceReportMetadata contains report header information
type ComplianceReportMetadata struct {
	ReportID       string               `json:"report_id"`
	ReportType     ComplianceReportType `json:"report_type"`
	OrganizationID string               `json:"organization_id"`
	GeneratedAt    time.Time            `json:"generated_at"`
	PeriodStart    time.Time            `json:"period_start"`
	PeriodEnd      time.Time            `json:"period_end"`
	GeneratedBy    string               `json:"generated_by"`
	TotalEvents    int64                `json:"total_events"`
	Summary        ReportSummary        `json:"summary"`
}

// ReportSummary contains aggregate statistics
type ReportSummary struct {
	TotalActions      int64            `json:"total_actions"`
	UniqueUsers       int64            `json:"unique_users"`
	FailedActions     int64            `json:"failed_actions"`
	ActionBreakdown   map[string]int64 `json:"action_breakdown"`
	ResourceBreakdown map[string]int64 `json:"resource_breakdown"`
}

// AccessLogEntry represents an access event for compliance
type AccessLogEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	UserID       string    `json:"user_id"`
	UserEmail    string    `json:"user_email,omitempty"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id,omitempty"`
	ResourceName string    `json:"resource_name,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	Success      bool      `json:"success"`
	Details      string    `json:"details,omitempty"`
}

// ChangeLogEntry represents a data change event
type ChangeLogEntry struct {
	Timestamp    time.Time       `json:"timestamp"`
	UserID       string          `json:"user_id"`
	UserEmail    string          `json:"user_email,omitempty"`
	ChangeType   string          `json:"change_type"` // create, update, delete
	ResourceType string          `json:"resource_type"`
	ResourceID   string          `json:"resource_id"`
	ResourceName string          `json:"resource_name,omitempty"`
	BeforeState  json.RawMessage `json:"before_state,omitempty"`
	AfterState   json.RawMessage `json:"after_state,omitempty"`
	IPAddress    string          `json:"ip_address,omitempty"`
}

// SecurityLogEntry represents a security-relevant event
type SecurityLogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"` // login, logout, failed_login, permission_change
	UserID      string                 `json:"user_id,omitempty"`
	UserEmail   string                 `json:"user_email,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Success     bool                   `json:"success"`
	ErrorReason string                 `json:"error_reason,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// UserActivityEntry represents user activity for a time period
type UserActivityEntry struct {
	UserID        string           `json:"user_id"`
	UserEmail     string           `json:"user_email"`
	FirstActivity time.Time        `json:"first_activity"`
	LastActivity  time.Time        `json:"last_activity"`
	TotalActions  int64            `json:"total_actions"`
	ActionTypes   map[string]int64 `json:"action_types"`
	ResourceTypes map[string]int64 `json:"resource_types"`
	IPAddresses   []string         `json:"ip_addresses"`
}

// ComplianceReport is the base structure for all compliance reports
type ComplianceReport struct {
	Metadata ComplianceReportMetadata `json:"metadata"`
	Data     interface{}              `json:"data"`
}

// AccessLogReport is a report of access events
type AccessLogReport struct {
	Metadata ComplianceReportMetadata `json:"metadata"`
	Entries  []AccessLogEntry         `json:"entries"`
}

// ChangeLogReport is a report of data changes
type ChangeLogReport struct {
	Metadata ComplianceReportMetadata `json:"metadata"`
	Entries  []ChangeLogEntry         `json:"entries"`
}

// SecurityLogReport is a report of security events
type SecurityLogReport struct {
	Metadata ComplianceReportMetadata `json:"metadata"`
	Entries  []SecurityLogEntry       `json:"entries"`
}

// UserActivityReport is a report of user activity
type UserActivityReport struct {
	Metadata ComplianceReportMetadata `json:"metadata"`
	Users    []UserActivityEntry      `json:"users"`
}

// ComplianceReporter generates compliance reports from audit logs
type ComplianceReporter struct {
	auditLogger *audit.Logger
	exportDir   string
}

// NewComplianceReporter creates a new compliance reporter
func NewComplianceReporter(auditLogger *audit.Logger, exportDir string) *ComplianceReporter {
	return &ComplianceReporter{
		auditLogger: auditLogger,
		exportDir:   exportDir,
	}
}

// GenerateReport creates a compliance report based on the request
func (r *ComplianceReporter) GenerateReport(ctx context.Context, job *ExportJob, progressCh chan<- int) error {
	// Extract parameters from job
	reportType := ComplianceReportType(job.Type)
	if job.Parameters != nil {
		if rt, ok := job.Parameters["report_type"].(string); ok {
			reportType = ComplianceReportType(rt)
		}
	}

	var startDate, endDate time.Time
	if job.Parameters != nil {
		if sd, ok := job.Parameters["start_date"].(string); ok {
			startDate, _ = time.Parse(time.RFC3339, sd)
		}
		if ed, ok := job.Parameters["end_date"].(string); ok {
			endDate, _ = time.Parse(time.RFC3339, ed)
		}
	}
	if startDate.IsZero() {
		startDate = time.Now().AddDate(0, -1, 0) // Default: 1 month ago
	}
	if endDate.IsZero() {
		endDate = time.Now()
	}

	progressCh <- 10

	// Query audit logs
	filter := audit.AuditFilter{
		OrgID:     job.OrgID,
		StartTime: &startDate,
		EndTime:   &endDate,
		Limit:     10000,
	}

	events, total, err := r.auditLogger.Query(filter)
	if err != nil {
		return fmt.Errorf("failed to query audit logs: %w", err)
	}

	progressCh <- 30

	// Generate appropriate report
	var report interface{}
	switch reportType {
	case ReportTypeAccessLog:
		report = r.generateAccessLogReport(job, events, total, startDate, endDate)
	case ReportTypeChangeLog:
		report = r.generateChangeLogReport(job, events, total, startDate, endDate)
	case ReportTypeSecurityLog:
		report = r.generateSecurityLogReport(job, events, total, startDate, endDate)
	case ReportTypeUserActivity:
		report = r.generateUserActivityReport(job, events, total, startDate, endDate)
	case ReportTypeSOC2:
		report = r.generateSOC2Report(job, events, total, startDate, endDate)
	case ReportTypeHIPAA:
		report = r.generateHIPAAReport(job, events, total, startDate, endDate)
	default:
		return fmt.Errorf("unsupported report type: %s", reportType)
	}

	progressCh <- 70

	// Write output file
	filePath := fmt.Sprintf("%s/%s.json", r.exportDir, job.ID)
	if err := WriteJSON(filePath, report); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	job.FilePath = filePath
	progressCh <- 100

	return nil
}

func (r *ComplianceReporter) generateAccessLogReport(job *ExportJob, events []*audit.AuditEvent, total int64, start, end time.Time) *AccessLogReport {
	entries := make([]AccessLogEntry, 0, len(events))
	actionBreakdown := make(map[string]int64)
	resourceBreakdown := make(map[string]int64)
	uniqueUsers := make(map[string]bool)
	var failedCount int64

	for _, event := range events {
		entries = append(entries, AccessLogEntry{
			Timestamp:    event.Timestamp,
			UserID:       event.ActorID,
			UserEmail:    event.ActorEmail,
			Action:       string(event.Action),
			ResourceType: string(event.ResourceType),
			ResourceID:   event.ResourceID,
			ResourceName: event.ResourceName,
			IPAddress:    event.IPAddress,
			UserAgent:    event.UserAgent,
			Success:      event.Success,
			Details:      event.ErrorMessage,
		})

		actionBreakdown[string(event.Action)]++
		resourceBreakdown[string(event.ResourceType)]++
		uniqueUsers[event.ActorID] = true
		if !event.Success {
			failedCount++
		}
	}

	return &AccessLogReport{
		Metadata: ComplianceReportMetadata{
			ReportID:       job.ID,
			ReportType:     ReportTypeAccessLog,
			OrganizationID: job.OrgID,
			GeneratedAt:    time.Now().UTC(),
			PeriodStart:    start,
			PeriodEnd:      end,
			GeneratedBy:    job.UserID,
			TotalEvents:    total,
			Summary: ReportSummary{
				TotalActions:      total,
				UniqueUsers:       int64(len(uniqueUsers)),
				FailedActions:     failedCount,
				ActionBreakdown:   actionBreakdown,
				ResourceBreakdown: resourceBreakdown,
			},
		},
		Entries: entries,
	}
}

func (r *ComplianceReporter) generateChangeLogReport(job *ExportJob, events []*audit.AuditEvent, total int64, start, end time.Time) *ChangeLogReport {
	entries := make([]ChangeLogEntry, 0)
	actionBreakdown := make(map[string]int64)
	resourceBreakdown := make(map[string]int64)
	uniqueUsers := make(map[string]bool)

	// Filter to change events only
	changeActions := map[audit.ActionType]bool{
		audit.ActionCreate: true,
		audit.ActionUpdate: true,
		audit.ActionDelete: true,
	}

	for _, event := range events {
		if !changeActions[event.Action] {
			continue
		}

		entries = append(entries, ChangeLogEntry{
			Timestamp:    event.Timestamp,
			UserID:       event.ActorID,
			UserEmail:    event.ActorEmail,
			ChangeType:   string(event.Action),
			ResourceType: string(event.ResourceType),
			ResourceID:   event.ResourceID,
			ResourceName: event.ResourceName,
			BeforeState:  event.BeforeState,
			AfterState:   event.AfterState,
			IPAddress:    event.IPAddress,
		})

		actionBreakdown[string(event.Action)]++
		resourceBreakdown[string(event.ResourceType)]++
		uniqueUsers[event.ActorID] = true
	}

	return &ChangeLogReport{
		Metadata: ComplianceReportMetadata{
			ReportID:       job.ID,
			ReportType:     ReportTypeChangeLog,
			OrganizationID: job.OrgID,
			GeneratedAt:    time.Now().UTC(),
			PeriodStart:    start,
			PeriodEnd:      end,
			GeneratedBy:    job.UserID,
			TotalEvents:    int64(len(entries)),
			Summary: ReportSummary{
				TotalActions:      int64(len(entries)),
				UniqueUsers:       int64(len(uniqueUsers)),
				ActionBreakdown:   actionBreakdown,
				ResourceBreakdown: resourceBreakdown,
			},
		},
		Entries: entries,
	}
}

func (r *ComplianceReporter) generateSecurityLogReport(job *ExportJob, events []*audit.AuditEvent, total int64, start, end time.Time) *SecurityLogReport {
	entries := make([]SecurityLogEntry, 0)
	actionBreakdown := make(map[string]int64)
	uniqueUsers := make(map[string]bool)
	var failedCount int64

	// Filter to security events
	securityActions := map[audit.ActionType]bool{
		audit.ActionLogin:   true,
		audit.ActionLogout:  true,
		audit.ActionApprove: true,
		audit.ActionReject:  true,
	}

	// Also include failed access attempts
	for _, event := range events {
		isSecurityEvent := securityActions[event.Action] || !event.Success

		if !isSecurityEvent {
			continue
		}

		entry := SecurityLogEntry{
			Timestamp:   event.Timestamp,
			EventType:   string(event.Action),
			UserID:      event.ActorID,
			UserEmail:   event.ActorEmail,
			IPAddress:   event.IPAddress,
			UserAgent:   event.UserAgent,
			Success:     event.Success,
			ErrorReason: event.ErrorMessage,
			Details:     event.Metadata,
		}

		if !event.Success && event.Action == audit.ActionLogin {
			entry.EventType = "failed_login"
		}

		entries = append(entries, entry)
		actionBreakdown[entry.EventType]++
		uniqueUsers[event.ActorID] = true
		if !event.Success {
			failedCount++
		}
	}

	return &SecurityLogReport{
		Metadata: ComplianceReportMetadata{
			ReportID:       job.ID,
			ReportType:     ReportTypeSecurityLog,
			OrganizationID: job.OrgID,
			GeneratedAt:    time.Now().UTC(),
			PeriodStart:    start,
			PeriodEnd:      end,
			GeneratedBy:    job.UserID,
			TotalEvents:    int64(len(entries)),
			Summary: ReportSummary{
				TotalActions:    int64(len(entries)),
				UniqueUsers:     int64(len(uniqueUsers)),
				FailedActions:   failedCount,
				ActionBreakdown: actionBreakdown,
			},
		},
		Entries: entries,
	}
}

func (r *ComplianceReporter) generateUserActivityReport(job *ExportJob, events []*audit.AuditEvent, total int64, start, end time.Time) *UserActivityReport {
	userActivity := make(map[string]*UserActivityEntry)
	userIPAddresses := make(map[string]map[string]bool)

	for _, event := range events {
		if event.ActorID == "system" {
			continue
		}

		activity, exists := userActivity[event.ActorID]
		if !exists {
			activity = &UserActivityEntry{
				UserID:        event.ActorID,
				UserEmail:     event.ActorEmail,
				FirstActivity: event.Timestamp,
				LastActivity:  event.Timestamp,
				ActionTypes:   make(map[string]int64),
				ResourceTypes: make(map[string]int64),
			}
			userActivity[event.ActorID] = activity
			userIPAddresses[event.ActorID] = make(map[string]bool)
		}

		activity.TotalActions++
		activity.ActionTypes[string(event.Action)]++
		activity.ResourceTypes[string(event.ResourceType)]++

		if event.Timestamp.Before(activity.FirstActivity) {
			activity.FirstActivity = event.Timestamp
		}
		if event.Timestamp.After(activity.LastActivity) {
			activity.LastActivity = event.Timestamp
		}
		if event.IPAddress != "" {
			userIPAddresses[event.ActorID][event.IPAddress] = true
		}
	}

	users := make([]UserActivityEntry, 0, len(userActivity))
	for userID, activity := range userActivity {
		// Collect unique IP addresses
		ips := make([]string, 0, len(userIPAddresses[userID]))
		for ip := range userIPAddresses[userID] {
			ips = append(ips, ip)
		}
		activity.IPAddresses = ips
		users = append(users, *activity)
	}

	return &UserActivityReport{
		Metadata: ComplianceReportMetadata{
			ReportID:       job.ID,
			ReportType:     ReportTypeUserActivity,
			OrganizationID: job.OrgID,
			GeneratedAt:    time.Now().UTC(),
			PeriodStart:    start,
			PeriodEnd:      end,
			GeneratedBy:    job.UserID,
			TotalEvents:    total,
			Summary: ReportSummary{
				TotalActions: total,
				UniqueUsers:  int64(len(users)),
			},
		},
		Users: users,
	}
}

// generateSOC2Report creates a SOC2-formatted compliance report
func (r *ComplianceReporter) generateSOC2Report(job *ExportJob, events []*audit.AuditEvent, total int64, start, end time.Time) *ComplianceReport {
	// SOC2 focuses on: Security, Availability, Processing Integrity, Confidentiality, Privacy
	// We'll organize the audit log data into these categories

	type SOC2Section struct {
		Principle   string                 `json:"principle"`
		Description string                 `json:"description"`
		Events      []map[string]interface{} `json:"events"`
		Statistics  map[string]int64       `json:"statistics"`
	}

	sections := map[string]*SOC2Section{
		"security": {
			Principle:   "Security",
			Description: "System is protected against unauthorized access",
			Events:      make([]map[string]interface{}, 0),
			Statistics:  make(map[string]int64),
		},
		"availability": {
			Principle:   "Availability",
			Description: "System is available for operation and use",
			Events:      make([]map[string]interface{}, 0),
			Statistics:  make(map[string]int64),
		},
		"confidentiality": {
			Principle:   "Confidentiality",
			Description: "Information designated as confidential is protected",
			Events:      make([]map[string]interface{}, 0),
			Statistics:  make(map[string]int64),
		},
	}

	for _, event := range events {
		eventMap := map[string]interface{}{
			"timestamp":     event.Timestamp,
			"actor":         event.ActorID,
			"action":        event.Action,
			"resource_type": event.ResourceType,
			"success":       event.Success,
		}

		// Categorize events into SOC2 principles
		switch event.Action {
		case audit.ActionLogin, audit.ActionLogout:
			sections["security"].Events = append(sections["security"].Events, eventMap)
			sections["security"].Statistics[string(event.Action)]++
		case audit.ActionAccess, audit.ActionRead:
			sections["confidentiality"].Events = append(sections["confidentiality"].Events, eventMap)
			sections["confidentiality"].Statistics[string(event.Action)]++
		default:
			sections["availability"].Events = append(sections["availability"].Events, eventMap)
			sections["availability"].Statistics[string(event.Action)]++
		}
	}

	// Count failed events as security concerns
	var failedCount int64
	for _, event := range events {
		if !event.Success {
			failedCount++
		}
	}

	return &ComplianceReport{
		Metadata: ComplianceReportMetadata{
			ReportID:       job.ID,
			ReportType:     ReportTypeSOC2,
			OrganizationID: job.OrgID,
			GeneratedAt:    time.Now().UTC(),
			PeriodStart:    start,
			PeriodEnd:      end,
			GeneratedBy:    job.UserID,
			TotalEvents:    total,
			Summary: ReportSummary{
				TotalActions:  total,
				FailedActions: failedCount,
			},
		},
		Data: sections,
	}
}

// generateHIPAAReport creates a HIPAA-formatted compliance report
func (r *ComplianceReporter) generateHIPAAReport(job *ExportJob, events []*audit.AuditEvent, total int64, start, end time.Time) *ComplianceReport {
	// HIPAA focuses on: Access Controls, Audit Controls, Integrity Controls, Transmission Security

	type HIPAASection struct {
		Control     string                   `json:"control"`
		Standard    string                   `json:"standard"`
		Events      []map[string]interface{} `json:"events"`
		Findings    []string                 `json:"findings"`
		Statistics  map[string]int64         `json:"statistics"`
	}

	sections := map[string]*HIPAASection{
		"access_controls": {
			Control:    "Access Control (§164.312(a)(1))",
			Standard:   "Unique User Identification, Automatic Logoff, Encryption",
			Events:     make([]map[string]interface{}, 0),
			Findings:   make([]string, 0),
			Statistics: make(map[string]int64),
		},
		"audit_controls": {
			Control:    "Audit Controls (§164.312(b))",
			Standard:   "Record and examine activity in systems containing PHI",
			Events:     make([]map[string]interface{}, 0),
			Findings:   make([]string, 0),
			Statistics: make(map[string]int64),
		},
		"integrity": {
			Control:    "Integrity (§164.312(c)(1))",
			Standard:   "Protect ePHI from improper alteration or destruction",
			Events:     make([]map[string]interface{}, 0),
			Findings:   make([]string, 0),
			Statistics: make(map[string]int64),
		},
	}

	uniqueUsers := make(map[string]bool)
	var failedLoginCount int64

	for _, event := range events {
		eventMap := map[string]interface{}{
			"timestamp":     event.Timestamp,
			"actor":         event.ActorID,
			"action":        event.Action,
			"resource_type": event.ResourceType,
			"ip_address":    event.IPAddress,
			"success":       event.Success,
		}

		uniqueUsers[event.ActorID] = true

		switch event.Action {
		case audit.ActionLogin, audit.ActionLogout:
			sections["access_controls"].Events = append(sections["access_controls"].Events, eventMap)
			sections["access_controls"].Statistics[string(event.Action)]++
			if event.Action == audit.ActionLogin && !event.Success {
				failedLoginCount++
			}
		case audit.ActionCreate, audit.ActionUpdate, audit.ActionDelete:
			sections["integrity"].Events = append(sections["integrity"].Events, eventMap)
			sections["integrity"].Statistics[string(event.Action)]++
		default:
			sections["audit_controls"].Events = append(sections["audit_controls"].Events, eventMap)
			sections["audit_controls"].Statistics[string(event.Action)]++
		}
	}

	// Generate findings
	if failedLoginCount > 10 {
		sections["access_controls"].Findings = append(
			sections["access_controls"].Findings,
			fmt.Sprintf("High number of failed login attempts detected: %d", failedLoginCount),
		)
	}

	return &ComplianceReport{
		Metadata: ComplianceReportMetadata{
			ReportID:       job.ID,
			ReportType:     ReportTypeHIPAA,
			OrganizationID: job.OrgID,
			GeneratedAt:    time.Now().UTC(),
			PeriodStart:    start,
			PeriodEnd:      end,
			GeneratedBy:    job.UserID,
			TotalEvents:    total,
			Summary: ReportSummary{
				TotalActions:  total,
				UniqueUsers:   int64(len(uniqueUsers)),
				FailedActions: failedLoginCount,
			},
		},
		Data: sections,
	}
}
