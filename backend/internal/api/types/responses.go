// Package types provides standardized API response types for the Prism API.
// These types are designed to be exported to TypeScript for type-safe client usage.
package types

import "time"

// APIResponse is the standard wrapper for all API responses.
// Success responses will have Data populated, error responses will have Error populated.
type APIResponse[T any] struct {
	Data  *T     `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// PaginatedResponse wraps paginated data with metadata.
type PaginatedResponse[T any] struct {
	Items      []T `json:"items"`
	Total      int `json:"total"`
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
	HasMore    bool `json:"has_more"`
}

// SuccessResponse is a simple success message response.
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// --- Auth Types ---

// UserDTO represents a user data transfer object.
type UserDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthResponse represents an authentication response with tokens.
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserDTO   `json:"user"`
}

// RegisterRequest represents a registration request.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents a login request.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest represents a token refresh request.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// --- Conversation Types ---

// ConversationDTO represents a conversation response.
type ConversationDTO struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Provider     string    `json:"provider"`
	Model        string    `json:"model"`
	SystemPrompt string    `json:"system_prompt,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// MessageDTO represents a message in a conversation.
type MessageDTO struct {
	ID         string                   `json:"id"`
	Role       string                   `json:"role"`
	Content    string                   `json:"content"`
	ToolCalls  []map[string]interface{} `json:"tool_calls,omitempty"`
	ToolCallID string                   `json:"tool_call_id,omitempty"`
	CreatedAt  time.Time                `json:"created_at"`
}

// CreateConversationRequest represents a request to create a conversation.
type CreateConversationRequest struct {
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	SystemPrompt string `json:"system_prompt,omitempty"`
}

// UpdateConversationRequest represents a request to update a conversation.
type UpdateConversationRequest struct {
	Title string `json:"title"`
}

// ConversationListResponse represents a list of conversations.
type ConversationListResponse struct {
	Conversations []ConversationDTO `json:"conversations"`
}

// MessageListResponse represents a list of messages.
type MessageListResponse struct {
	Messages []MessageDTO `json:"messages"`
}

// --- Provider Types ---

// SetKeyRequest represents a request to set an API key.
type SetKeyRequest struct {
	APIKey string `json:"api_key"`
}

// ValidateKeyRequest represents a request to validate an API key.
type ValidateKeyRequest struct {
	APIKey string `json:"api_key"`
}

// KeyStatusResponse represents the status of an API key.
type KeyStatusResponse struct {
	HasKey   bool   `json:"has_key"`
	Provider string `json:"provider"`
}

// KeyValidationResponse represents the result of key validation.
type KeyValidationResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
}

// ProviderListResponse represents a list of configured providers.
type ProviderListResponse struct {
	Providers []string `json:"providers"`
}

// --- Workspace Types ---

// DirectoryResponse represents a workspace directory.
type DirectoryResponse struct {
	Path string `json:"path"`
}

// SetDirectoryRequest represents a request to set workspace directory.
type SetDirectoryRequest struct {
	Directory string `json:"directory"`
}

// BrowseDirectoryEntry represents a directory entry when browsing.
type BrowseDirectoryEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// BrowseDirectoriesResponse represents the response from browsing directories.
type BrowseDirectoriesResponse struct {
	CurrentPath string                 `json:"current_path"`
	ParentPath  string                 `json:"parent_path"`
	Directories []BrowseDirectoryEntry `json:"directories"`
}

// WorkspaceInfo represents workspace information.
type WorkspaceInfo struct {
	ID             string  `json:"id"`
	Path           string  `json:"path"`
	Name           string  `json:"name"`
	IsCurrent      bool    `json:"is_current"`
	LastAccessedAt *string `json:"last_accessed_at,omitempty"`
}

// WorkspaceListResponse represents a list of workspaces.
type WorkspaceListResponse struct {
	Workspaces []WorkspaceInfo `json:"workspaces"`
}

// CloneRepoRequest represents a request to clone a repository.
type CloneRepoRequest struct {
	RepoURL string `json:"repo_url"`
	Branch  string `json:"branch,omitempty"`
}

// CloneRepoResponse represents the response from cloning a repository.
type CloneRepoResponse struct {
	Success bool   `json:"success"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

// --- File Types ---

// FileNode represents a file or directory in the file tree.
type FileNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Type     string     `json:"type"` // "file" or "directory"
	Children []FileNode `json:"children,omitempty"`
}

// FileListResponse represents a list of files.
type FileListResponse struct {
	Files []FileNode `json:"files"`
}

// FileContentResponse represents file content.
type FileContentResponse struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// WriteFileRequest represents a request to write a file.
type WriteFileRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// RenameFileRequest represents a request to rename a file.
type RenameFileRequest struct {
	SourcePath string `json:"source_path"`
	DestPath   string `json:"dest_path"`
}

// CreateDirectoryRequest represents a request to create a directory.
type CreateDirectoryRequest struct {
	Path string `json:"path"`
}

// --- Integration Types ---

// IntegrationStatus represents the status of a single integration.
type IntegrationStatus struct {
	Enabled   bool   `json:"enabled"`
	Connected bool   `json:"connected"`
	ChannelID string `json:"channel_id,omitempty"`
}

// IntegrationStatusResponse represents the status of all integrations.
type IntegrationStatusResponse struct {
	Discord IntegrationStatus `json:"discord"`
	Slack   IntegrationStatus `json:"slack"`
	PostHog IntegrationStatus `json:"posthog"`
}

// SetIntegrationRequest represents a request to set integration settings.
type SetIntegrationRequest struct {
	WebhookURL string `json:"webhook_url,omitempty"`
	ChannelID  string `json:"channel_id,omitempty"`
	Enabled    bool   `json:"enabled"`
}

// --- Build Types ---

// BuildInfo represents build information.
type BuildInfo struct {
	ID         string  `json:"id"`
	Status     string  `json:"status"` // "pending", "running", "success", "failed"
	Command    string  `json:"command"`
	StartTime  string  `json:"start_time"`
	EndTime    *string `json:"end_time,omitempty"`
	DurationMs *int64  `json:"duration_ms,omitempty"`
	Error      string  `json:"error,omitempty"`
	PreviewURL string  `json:"preview_url,omitempty"`
}

// --- GitHub Types ---

// GitHubStatusResponse represents GitHub connection status.
type GitHubStatusResponse struct {
	Connected bool   `json:"connected"`
	Username  string `json:"username,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// GitHubRepoInfo represents a GitHub repository.
type GitHubRepoInfo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description,omitempty"`
	Private     bool   `json:"private"`
	HTMLURL     string `json:"html_url"`
	CloneURL    string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
}

// GitHubReposResponse represents a list of GitHub repositories.
type GitHubReposResponse struct {
	Repos []GitHubRepoInfo `json:"repos"`
}
