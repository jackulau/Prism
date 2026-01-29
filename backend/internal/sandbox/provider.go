package sandbox

import (
	"context"
	"time"
)

// Framework represents the type of framework for a sandbox
type Framework string

const (
	FrameworkNextJS Framework = "nextjs"
	FrameworkReact  Framework = "react"
	FrameworkVue    Framework = "vue"
	FrameworkVite   Framework = "vite"
	FrameworkStatic Framework = "static"
)

// SandboxStatus represents the status of a sandbox
type SandboxStatus string

const (
	SandboxStatusCreating  SandboxStatus = "creating"
	SandboxStatusReady     SandboxStatus = "ready"
	SandboxStatusDeploying SandboxStatus = "deploying"
	SandboxStatusDeployed  SandboxStatus = "deployed"
	SandboxStatusFailed    SandboxStatus = "failed"
	SandboxStatusDeleted   SandboxStatus = "deleted"
)

// DeploymentStatus represents the status of a deployment
type DeploymentStatus string

const (
	DeploymentStatusQueued    DeploymentStatus = "queued"
	DeploymentStatusBuilding  DeploymentStatus = "building"
	DeploymentStatusReady     DeploymentStatus = "ready"
	DeploymentStatusError     DeploymentStatus = "error"
	DeploymentStatusCancelled DeploymentStatus = "cancelled"
)

// CreateOptions contains options for creating a sandbox
type CreateOptions struct {
	Framework    Framework `json:"framework"`
	NodeVersion  string    `json:"node_version,omitempty"`
	BuildCommand string    `json:"build_command,omitempty"`
	OutputDir    string    `json:"output_dir,omitempty"`
	EnvVars      map[string]string `json:"env_vars,omitempty"`
}

// Sandbox represents a sandbox environment
type Sandbox struct {
	ID         string        `json:"id"`
	Provider   string        `json:"provider"`
	Status     SandboxStatus `json:"status"`
	PreviewURL string        `json:"preview_url,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// DeployResult contains the result of a deployment
type DeployResult struct {
	DeploymentID string           `json:"deployment_id"`
	PreviewURL   string           `json:"preview_url"`
	Status       DeploymentStatus `json:"status"`
	BuildOutput  string           `json:"build_output,omitempty"`
	Error        string           `json:"error,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	ReadyAt      *time.Time       `json:"ready_at,omitempty"`
}

// LogEntry represents a log entry from the sandbox
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Level     string    `json:"level"` // "info", "warn", "error"
	Source    string    `json:"source,omitempty"` // "build", "runtime", etc.
}

// Provider defines the interface for sandbox providers
type Provider interface {
	// Name returns the name of the provider (e.g., "vercel", "docker")
	Name() string

	// CreateSandbox creates a new sandbox environment
	CreateSandbox(ctx context.Context, opts *CreateOptions) (*Sandbox, error)

	// DeploySandbox deploys files to an existing sandbox
	DeploySandbox(ctx context.Context, sandboxID string, files map[string][]byte) (*DeployResult, error)

	// GetSandbox retrieves the current state of a sandbox
	GetSandbox(ctx context.Context, sandboxID string) (*Sandbox, error)

	// GetLogs returns a channel of log entries for the sandbox
	GetLogs(ctx context.Context, sandboxID string) (<-chan LogEntry, error)

	// DeleteSandbox deletes a sandbox and its associated resources
	DeleteSandbox(ctx context.Context, sandboxID string) error

	// GetPreviewURL returns the preview URL for a sandbox
	GetPreviewURL(sandboxID string) string
}

// DeployOptions contains options for deploying to a sandbox
type DeployOptions struct {
	Files        map[string][]byte `json:"files"`
	BuildCommand string            `json:"build_command,omitempty"`
	OutputDir    string            `json:"output_dir,omitempty"`
	EnvVars      map[string]string `json:"env_vars,omitempty"`
}
