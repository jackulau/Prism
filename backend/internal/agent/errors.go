package agent

import "errors"

// Agent errors
var (
	ErrAgentNotFound       = errors.New("agent not found")
	ErrAgentAlreadyRunning = errors.New("agent is already running")
	ErrAgentPanicked       = errors.New("agent panicked during execution")
	ErrAgentTimeout        = errors.New("agent execution timed out")
)

// Pool errors
var (
	ErrPoolNotRunning = errors.New("agent pool is not running")
	ErrQueueFull      = errors.New("task queue is full")
)

// Task errors
var (
	ErrTaskNotFound = errors.New("task not found")
	ErrTaskTimeout  = errors.New("task execution timed out")
)

// Manager errors
var (
	ErrManagerNotInitialized = errors.New("agent manager not initialized")
	ErrInvalidAgentConfig    = errors.New("invalid agent configuration")
)

// Workflow step errors
var (
	ErrSandboxNotReady       = errors.New("sandbox is not ready")
	ErrInvalidRepository     = errors.New("invalid repository")
	ErrTokenValidationFailed = errors.New("token validation failed")
	ErrBranchCreationFailed  = errors.New("branch creation failed")
	ErrCommitFailed          = errors.New("commit failed")
	ErrMessageLoadFailed     = errors.New("failed to load messages")
	ErrNoChangesToCommit     = errors.New("no changes to commit")
)
