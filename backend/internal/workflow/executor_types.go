package workflow

import (
	"context"
	"time"

	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

// WorkflowStep represents a step in the agent workflow
type WorkflowStep string

const (
	StepLoadAgent      WorkflowStep = "load_agent"
	StepValidateGitHub WorkflowStep = "validate_github"
	StepCreateSandbox  WorkflowStep = "create_sandbox"
	StepLoadHistory    WorkflowStep = "load_history"
	StepCreateBranch   WorkflowStep = "create_branch"
	StepRunLLM         WorkflowStep = "run_llm"
	StepSaveResponse   WorkflowStep = "save_response"
	StepCommitChanges  WorkflowStep = "commit_changes"
	StepCleanupSandbox WorkflowStep = "cleanup_sandbox"
	StepMarkComplete   WorkflowStep = "mark_complete"
)

// WorkflowEventHandler is a function that handles executor events
type WorkflowEventHandler func(event ExecutorEvent)

// ExecutorConfig holds configuration for the workflow executor
type ExecutorConfig struct {
	// DefaultTimeout is the default timeout for workflow execution
	DefaultTimeout time.Duration
	// MaxLLMIterations is the maximum number of LLM iterations
	MaxLLMIterations int
	// AutoCommit indicates whether to auto-commit changes
	AutoCommit bool
	// CommitMessagePrefix is the prefix for commit messages
	CommitMessagePrefix string
	// CloneDepth is the depth for git clone operations
	CloneDepth int
}

// DefaultExecutorConfig returns default executor configuration
func DefaultExecutorConfig() ExecutorConfig {
	return ExecutorConfig{
		DefaultTimeout:      30 * time.Minute,
		MaxLLMIterations:    50,
		AutoCommit:          true,
		CommitMessagePrefix: "[prism]",
		CloneDepth:          1,
	}
}

// WorkflowRepositories holds repository dependencies for the executor
type WorkflowRepositories struct {
	Agent          AgentRepository
	AgentExecution *AgentExecutionRepository
	Message        *repository.MessageRepository
	GitHub         GitHubTokenRepository
}

// AgentRepository defines the interface for agent storage
type AgentRepository interface {
	GetByID(ctx context.Context, id string) (*AgentData, error)
	UpdateStatus(ctx context.Context, agentID, status, errorMsg string) error
	UpdateBranch(ctx context.Context, agentID, branchName string) error
}

// GitHubTokenRepository defines the interface for GitHub token storage
type GitHubTokenRepository interface {
	GetToken(ctx context.Context, userID string) (string, error)
}

// GitHubRepository defines the interface for GitHub token storage and validation
type GitHubRepository interface {
	GetToken(ctx context.Context, userID string) (string, error)
	ValidateToken(ctx context.Context, token string) (bool, error)
}

// ToolExecutor defines the interface for executing tools
type ToolExecutor interface {
	Execute(ctx context.Context, toolCall llm.ToolCall, sandboxCtx *SandboxContext) (string, error)
}

// GitClient defines the interface for git operations
type GitClient interface {
	Clone(ctx context.Context, url, dir, token string) error
	Checkout(ctx context.Context, dir, branch string) error
	CreateBranch(ctx context.Context, dir, branch string) error
	Add(ctx context.Context, dir string, paths ...string) error
	Commit(ctx context.Context, dir, message string) error
	Push(ctx context.Context, dir, remote, branch, token string) error
	HasChanges(ctx context.Context, dir string) (bool, error)
	GetCurrentBranch(ctx context.Context, dir string) (string, error)
}

// AgentData represents agent configuration and state
type AgentData struct {
	ID             string
	Name           string
	UserID         string
	Provider       string
	Model          string
	SystemPrompt   string
	RepoURL        string
	ConversationID string
	Tools          []llm.ToolDefinition
	Metadata       map[string]interface{}
	CurrentTask    string
}

// SandboxContext holds sandbox state during workflow execution
type SandboxContext struct {
	UserID     string
	WorkDir    string
	RepoURL    string
	BranchName string
	Service    *sandbox.Service
	GitClient  GitClient
	CloneDepth int
}

// ExecutionContext holds state during workflow execution
type ExecutionContext struct {
	// Identifiers
	ExecutionID string
	AgentID     string
	UserID      string

	// Timing
	StartedAt   time.Time
	CompletedAt *time.Time

	// State
	CurrentStep WorkflowStep
	Context     context.Context
	Cancel      context.CancelFunc

	// Data loaded during execution
	AgentData   *AgentData
	GitHubToken string
	SandboxCtx  *SandboxContext
	Messages    []llm.Message
	LLMResult   *LLMResult
}

// LLMLoopConfig holds configuration for the LLM loop
type LLMLoopConfig struct {
	Agent         *AgentData
	Messages      []llm.Message
	Tools         []llm.ToolDefinition
	MaxIterations int
	SandboxCtx    *SandboxContext
	ToolExecutor  ToolExecutor
}

// LLMResult holds the result of LLM execution
type LLMResult struct {
	Output       string
	Iterations   int
	Messages     []llm.Message
	ToolCalls    []ToolCallResult
	FilesChanged []string
	TokenUsage   *llm.Usage
	Metadata     map[string]interface{}
}

// AgentExecution represents an execution record
type AgentExecution struct {
	ID             string
	AgentID        string
	UserID         string
	ConversationID string
	Status         string
	CurrentStep    string
	StartedAt      *time.Time
	CompletedAt    *time.Time
	Error          string
	BranchName     string
	CommitSHA      string
	Iterations     int
}

// TokenUsageRecord represents a token usage record
type TokenUsageRecord struct {
	ID               string
	ExecutionID      string
	UserID           string
	Provider         string
	Model            string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	CostUSD          float64
	CreatedAt        time.Time
}

// ToolCallResult represents the result of a tool call
type ToolCallResult struct {
	ID         string
	ToolCallID string
	Name       string
	Parameters map[string]interface{}
	Output     string
	Error      string
	Duration   time.Duration
}

// ExecutorEvent represents an event during workflow execution
type ExecutorEvent struct {
	ExecutionID string
	AgentID     string
	Step        WorkflowStep
	Status      string
	Data        map[string]interface{}
	Error       string
	Timestamp   time.Time
}
