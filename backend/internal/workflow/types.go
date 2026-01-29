package workflow

import (
	"context"
	"time"

	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

// WorkflowStep represents a step in the workflow execution
type WorkflowStep string

const (
	StepLoadAgent         WorkflowStep = "load_agent"
	StepValidateGitHub    WorkflowStep = "validate_github"
	StepCreateSandbox     WorkflowStep = "create_sandbox"
	StepLoadHistory       WorkflowStep = "load_history"
	StepCreateBranch      WorkflowStep = "create_branch"
	StepRunLLM            WorkflowStep = "run_llm"
	StepSaveResponse      WorkflowStep = "save_response"
	StepCommitChanges     WorkflowStep = "commit_changes"
	StepCleanupSandbox    WorkflowStep = "cleanup_sandbox"
	StepMarkComplete      WorkflowStep = "mark_complete"
)

// WorkflowStatus represents the status of a workflow execution
type WorkflowStatus string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
)

// WorkflowRepositories holds all required repositories for workflow execution
type WorkflowRepositories struct {
	Agent          AgentRepository
	Conversation   *repository.ConversationRepository
	Message        *repository.MessageRepository
	GitHub         GitHubRepository
	AgentExecution *AgentExecutionRepository
	TokenUsage     *TokenUsageRepository
}

// AgentRepository defines the interface for agent data access
type AgentRepository interface {
	GetByID(ctx context.Context, id string) (*AgentData, error)
	UpdateStatus(ctx context.Context, id string, status string, errorMsg string) error
	UpdateBranch(ctx context.Context, id string, branchName string) error
}

// AgentData represents agent data from the database
type AgentData struct {
	ID             string            `json:"id"`
	UserID         string            `json:"user_id"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	RepoURL        string            `json:"repo_url"`
	BranchName     string            `json:"branch_name"`
	ConversationID string            `json:"conversation_id"`
	Provider       string            `json:"provider"`
	Model          string            `json:"model"`
	SystemPrompt   string            `json:"system_prompt"`
	Tools          []llm.ToolDefinition `json:"tools"`
	Status         string            `json:"status"`
	CurrentTask    string            `json:"current_task"`
	Metadata       map[string]string `json:"metadata"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// GitHubRepository defines the interface for GitHub token management
type GitHubRepository interface {
	GetToken(ctx context.Context, userID string) (string, error)
	ValidateToken(ctx context.Context, token string) (bool, error)
}

// SandboxContext holds the context for a sandbox environment
type SandboxContext struct {
	UserID     string
	WorkDir    string
	RepoURL    string
	BranchName string
	Service    *sandbox.Service
	GitClient  GitClient
	CloneDepth int
}

// GitClient defines the interface for git operations
type GitClient interface {
	Clone(ctx context.Context, url, dest, token string) error
	CreateBranch(ctx context.Context, workDir, branchName string) error
	Checkout(ctx context.Context, workDir, branchName string) error
	Add(ctx context.Context, workDir string, paths ...string) error
	Commit(ctx context.Context, workDir, message string) error
	Push(ctx context.Context, workDir, remote, branch, token string) error
	HasChanges(ctx context.Context, workDir string) (bool, error)
	GetCurrentBranch(ctx context.Context, workDir string) (string, error)
}

// LLMLoopConfig configures the LLM execution loop
type LLMLoopConfig struct {
	Agent         *AgentData
	Messages      []llm.Message
	Tools         []llm.ToolDefinition
	MaxIterations int
	SandboxCtx    *SandboxContext
	ToolExecutor  ToolExecutor
}

// ToolExecutor defines the interface for executing tools
type ToolExecutor interface {
	Execute(ctx context.Context, toolCall llm.ToolCall, sandboxCtx *SandboxContext) (string, error)
	GetAvailableTools() []llm.ToolDefinition
}

// LLMResult holds the results from an LLM execution loop
type LLMResult struct {
	Output       string                 `json:"output"`
	ToolCalls    []ToolCallResult       `json:"tool_calls,omitempty"`
	TokenUsage   *llm.Usage             `json:"token_usage,omitempty"`
	Iterations   int                    `json:"iterations"`
	Messages     []llm.Message          `json:"messages"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	FilesChanged []string               `json:"files_changed,omitempty"`
}

// ToolCallResult represents the result of a tool execution
type ToolCallResult struct {
	ToolCallID string                 `json:"tool_call_id"`
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Output     string                 `json:"output"`
	Error      string                 `json:"error,omitempty"`
	Duration   time.Duration          `json:"duration"`
}

// WorkflowEvent represents an event during workflow execution
type WorkflowEvent struct {
	ExecutionID string                 `json:"execution_id"`
	AgentID     string                 `json:"agent_id"`
	Step        WorkflowStep           `json:"step"`
	Status      string                 `json:"status"` // "started", "completed", "failed"
	Data        map[string]interface{} `json:"data,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// WorkflowEventHandler is a function that handles workflow events
type WorkflowEventHandler func(event WorkflowEvent)

// ExecutorConfig holds configuration for the workflow executor
type ExecutorConfig struct {
	MaxLLMIterations   int           `json:"max_llm_iterations"`
	DefaultTimeout     time.Duration `json:"default_timeout"`
	CloneDepth         int           `json:"clone_depth"`
	AutoCommit         bool          `json:"auto_commit"`
	CommitMessagePrefix string       `json:"commit_message_prefix"`
}

// DefaultExecutorConfig returns the default executor configuration
func DefaultExecutorConfig() ExecutorConfig {
	return ExecutorConfig{
		MaxLLMIterations:   50,
		DefaultTimeout:     30 * time.Minute,
		CloneDepth:         1,
		AutoCommit:         true,
		CommitMessagePrefix: "[agent] ",
	}
}

// ExecutionContext holds the context for a workflow execution
type ExecutionContext struct {
	Context       context.Context
	Cancel        context.CancelFunc
	ExecutionID   string
	AgentID       string
	UserID        string
	AgentData     *AgentData
	SandboxCtx    *SandboxContext
	GitHubToken   string
	Messages      []llm.Message
	LLMResult     *LLMResult
	CurrentStep   WorkflowStep
	StartedAt     time.Time
	CompletedAt   *time.Time
	Error         error
}

// AgentExecution represents a tracked agent execution in the database
type AgentExecution struct {
	ID             string     `json:"id"`
	AgentID        string     `json:"agent_id"`
	UserID         string     `json:"user_id"`
	ConversationID string     `json:"conversation_id"`
	Status         string     `json:"status"`
	CurrentStep    string     `json:"current_step"`
	StartedAt      *time.Time `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	Error          string     `json:"error"`
	BranchName     string     `json:"branch_name"`
	CommitSHA      string     `json:"commit_sha"`
	Iterations     int        `json:"iterations"`
}

// TokenUsageRecord represents a token usage record in the database
type TokenUsageRecord struct {
	ID               string    `json:"id"`
	ExecutionID      string    `json:"execution_id"`
	UserID           string    `json:"user_id"`
	Provider         string    `json:"provider"`
	Model            string    `json:"model"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	CostUSD          float64   `json:"cost_usd"`
	CreatedAt        time.Time `json:"created_at"`
}

// StepResult represents the result of a workflow step
type StepResult struct {
	Step      WorkflowStep
	Success   bool
	Error     error
	Data      map[string]interface{}
	Duration  time.Duration
}
