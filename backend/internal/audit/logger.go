package audit

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ActionType represents the type of action being audited
type ActionType string

const (
	ActionCreate   ActionType = "create"
	ActionRead     ActionType = "read"
	ActionUpdate   ActionType = "update"
	ActionDelete   ActionType = "delete"
	ActionAccess   ActionType = "access"
	ActionExport   ActionType = "export"
	ActionLogin    ActionType = "login"
	ActionLogout   ActionType = "logout"
	ActionApprove  ActionType = "approve"
	ActionReject   ActionType = "reject"
	ActionExecute  ActionType = "execute"
	ActionDownload ActionType = "download"
)

// ResourceType represents the type of resource being acted upon
type ResourceType string

const (
	ResourceUser         ResourceType = "user"
	ResourceAgent        ResourceType = "agent"
	ResourceWorkflow     ResourceType = "workflow"
	ResourceConversation ResourceType = "conversation"
	ResourceMessage      ResourceType = "message"
	ResourceSettings     ResourceType = "settings"
	ResourceIntegration  ResourceType = "integration"
	ResourceTool         ResourceType = "tool"
	ResourceExport       ResourceType = "export"
	ResourceAuditLog     ResourceType = "audit_log"
	ResourceSession      ResourceType = "session"
	ResourceAPIKey       ResourceType = "api_key"
	ResourceOrganization ResourceType = "organization"
	ResourceWorkspace    ResourceType = "workspace"
)

// AuditEvent represents a single audit log entry
type AuditEvent struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	ActorID      string                 `json:"actor_id"`
	ActorEmail   string                 `json:"actor_email,omitempty"`
	ActorType    string                 `json:"actor_type"` // "user", "system", "api_key"
	Action       ActionType             `json:"action"`
	ResourceType ResourceType           `json:"resource_type"`
	ResourceID   string                 `json:"resource_id,omitempty"`
	ResourceName string                 `json:"resource_name,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	SessionID    string                 `json:"session_id,omitempty"`
	OrgID        string                 `json:"organization_id,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	BeforeState  json.RawMessage        `json:"before_state,omitempty"`
	AfterState   json.RawMessage        `json:"after_state,omitempty"`
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"error_message,omitempty"`
}

// AuditEventOption is a functional option for configuring audit events
type AuditEventOption func(*AuditEvent)

// WithActorEmail sets the actor email
func WithActorEmail(email string) AuditEventOption {
	return func(e *AuditEvent) {
		e.ActorEmail = email
	}
}

// WithResourceName sets the resource name
func WithResourceName(name string) AuditEventOption {
	return func(e *AuditEvent) {
		e.ResourceName = name
	}
}

// WithIPAddress sets the IP address
func WithIPAddress(ip string) AuditEventOption {
	return func(e *AuditEvent) {
		e.IPAddress = ip
	}
}

// WithUserAgent sets the user agent
func WithUserAgent(ua string) AuditEventOption {
	return func(e *AuditEvent) {
		e.UserAgent = ua
	}
}

// WithSessionID sets the session ID
func WithSessionID(sessionID string) AuditEventOption {
	return func(e *AuditEvent) {
		e.SessionID = sessionID
	}
}

// WithOrgID sets the organization ID
func WithOrgID(orgID string) AuditEventOption {
	return func(e *AuditEvent) {
		e.OrgID = orgID
	}
}

// WithMetadata sets additional metadata
func WithMetadata(meta map[string]interface{}) AuditEventOption {
	return func(e *AuditEvent) {
		e.Metadata = meta
	}
}

// WithBeforeState sets the state before the action
func WithBeforeState(state interface{}) AuditEventOption {
	return func(e *AuditEvent) {
		if data, err := json.Marshal(state); err == nil {
			e.BeforeState = data
		}
	}
}

// WithAfterState sets the state after the action
func WithAfterState(state interface{}) AuditEventOption {
	return func(e *AuditEvent) {
		if data, err := json.Marshal(state); err == nil {
			e.AfterState = data
		}
	}
}

// WithError marks the event as failed with an error message
func WithError(err error) AuditEventOption {
	return func(e *AuditEvent) {
		e.Success = false
		if err != nil {
			e.ErrorMessage = err.Error()
		}
	}
}

// AuditRepository is the interface for storing audit events
type AuditRepository interface {
	Create(event *AuditEvent) error
	List(filter AuditFilter) ([]*AuditEvent, int64, error)
	GetByID(id string) (*AuditEvent, error)
	DeleteBefore(timestamp time.Time, excludeLegalHold bool) (int64, error)
}

// AuditFilter represents filtering options for audit queries
type AuditFilter struct {
	ActorID      string
	OrgID        string
	Action       ActionType
	ResourceType ResourceType
	ResourceID   string
	StartTime    *time.Time
	EndTime      *time.Time
	Success      *bool
	IPAddress    string
	Limit        int
	Offset       int
}

// Logger is the audit logging service
type Logger struct {
	repo AuditRepository
	mu   sync.Mutex

	// Sensitive fields that should be redacted
	sensitiveFields map[string]bool
}

// NewLogger creates a new audit logger
func NewLogger(repo AuditRepository) *Logger {
	return &Logger{
		repo: repo,
		sensitiveFields: map[string]bool{
			"password":      true,
			"password_hash": true,
			"secret":        true,
			"token":         true,
			"api_key":       true,
			"access_token":  true,
			"refresh_token": true,
			"private_key":   true,
			"credit_card":   true,
			"ssn":           true,
		},
	}
}

// Log creates and stores a new audit event
func (l *Logger) Log(
	actorID string,
	actorType string,
	action ActionType,
	resourceType ResourceType,
	resourceID string,
	opts ...AuditEventOption,
) error {
	event := &AuditEvent{
		ID:           uuid.New().String(),
		Timestamp:    time.Now().UTC(),
		ActorID:      actorID,
		ActorType:    actorType,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Success:      true,
	}

	for _, opt := range opts {
		opt(event)
	}

	// Redact sensitive fields from metadata
	if event.Metadata != nil {
		event.Metadata = l.redactSensitive(event.Metadata)
	}

	return l.repo.Create(event)
}

// LogUserAction is a convenience method for logging user actions
func (l *Logger) LogUserAction(
	userID string,
	userEmail string,
	action ActionType,
	resourceType ResourceType,
	resourceID string,
	opts ...AuditEventOption,
) error {
	opts = append(opts, WithActorEmail(userEmail))
	return l.Log(userID, "user", action, resourceType, resourceID, opts...)
}

// LogSystemAction logs an action performed by the system
func (l *Logger) LogSystemAction(
	action ActionType,
	resourceType ResourceType,
	resourceID string,
	opts ...AuditEventOption,
) error {
	return l.Log("system", "system", action, resourceType, resourceID, opts...)
}

// Query retrieves audit events based on filter criteria
func (l *Logger) Query(filter AuditFilter) ([]*AuditEvent, int64, error) {
	if filter.Limit <= 0 {
		filter.Limit = 100
	}
	if filter.Limit > 1000 {
		filter.Limit = 1000
	}
	return l.repo.List(filter)
}

// GetByID retrieves a single audit event by ID
func (l *Logger) GetByID(id string) (*AuditEvent, error) {
	return l.repo.GetByID(id)
}

// redactSensitive removes sensitive fields from a map
func (l *Logger) redactSensitive(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		if l.sensitiveFields[key] {
			result[key] = "[REDACTED]"
		} else if nested, ok := value.(map[string]interface{}); ok {
			result[key] = l.redactSensitive(nested)
		} else {
			result[key] = value
		}
	}
	return result
}

// SensitiveFieldPatterns returns a copy of sensitive field names
func (l *Logger) SensitiveFieldPatterns() []string {
	patterns := make([]string, 0, len(l.sensitiveFields))
	for field := range l.sensitiveFields {
		patterns = append(patterns, field)
	}
	return patterns
}

// AddSensitiveField adds a new field to the sensitive fields list
func (l *Logger) AddSensitiveField(field string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.sensitiveFields[field] = true
}
