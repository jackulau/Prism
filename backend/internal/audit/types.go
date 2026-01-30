package audit

// EventCategory represents the category of an audit event
type EventCategory string

const (
	CategoryAuth     EventCategory = "authentication"
	CategoryAPIKey   EventCategory = "api_key"
	CategorySession  EventCategory = "session"
	CategorySettings EventCategory = "settings"
	CategoryMFA      EventCategory = "mfa"
	CategoryProvider EventCategory = "provider"
	CategoryGitHub   EventCategory = "github"
)

// EventType represents the type of an audit event
type EventType string

const (
	// Auth events
	EventLogin          EventType = "login"
	EventLoginFailed    EventType = "login_failed"
	EventLogout         EventType = "logout"
	EventRegister       EventType = "register"
	EventPasswordChange EventType = "password_change"
	EventTokenRefresh   EventType = "token_refresh"

	// API Key events
	EventAPIKeyCreated EventType = "api_key_created"
	EventAPIKeyDeleted EventType = "api_key_deleted"
	EventAPIKeyUsed    EventType = "api_key_used"

	// Session events
	EventSessionCreated    EventType = "session_created"
	EventSessionTerminated EventType = "session_terminated"
	EventSessionExpired    EventType = "session_expired"

	// MFA events
	EventMFAEnabled     EventType = "mfa_enabled"
	EventMFADisabled    EventType = "mfa_disabled"
	EventMFAVerified    EventType = "mfa_verified"
	EventMFAFailed      EventType = "mfa_failed"
	EventBackupCodeUsed EventType = "backup_code_used"

	// Provider events
	EventProviderKeySet     EventType = "provider_key_set"
	EventProviderKeyDeleted EventType = "provider_key_deleted"

	// Settings events
	EventSettingsChanged EventType = "settings_changed"

	// GitHub events
	EventGitHubConnected    EventType = "github_connected"
	EventGitHubDisconnected EventType = "github_disconnected"
)

// Entry represents a single audit log entry to be recorded
type Entry struct {
	UserID       *string
	EventType    EventType
	Category     EventCategory
	Action       string
	ResourceType string
	ResourceID   string
	IPAddress    string
	UserAgent    string
	Details      map[string]interface{}
	Success      bool
}
