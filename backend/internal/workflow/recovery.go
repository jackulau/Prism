package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Common workflow errors
var (
	ErrAgentNotFound        = errors.New("agent not found")
	ErrInvalidAgentConfig   = errors.New("invalid agent configuration")
	ErrGitHubTokenInvalid   = errors.New("github token is invalid or expired")
	ErrGitHubTokenMissing   = errors.New("github token not configured")
	ErrSandboxCreationFailed = errors.New("failed to create sandbox environment")
	ErrCloneFailed          = errors.New("failed to clone repository")
	ErrBranchCreationFailed = errors.New("failed to create branch")
	ErrLLMExecutionFailed   = errors.New("llm execution failed")
	ErrMaxIterationsReached = errors.New("maximum LLM iterations reached")
	ErrToolExecutionFailed  = errors.New("tool execution failed")
	ErrCommitFailed         = errors.New("failed to commit changes")
	ErrPushFailed           = errors.New("failed to push changes")
	ErrWorkflowCancelled    = errors.New("workflow was cancelled")
	ErrExecutionTimeout     = errors.New("workflow execution timed out")
)

// RecoveryAction defines an action to take during error recovery
type RecoveryAction string

const (
	RecoveryRetry    RecoveryAction = "retry"
	RecoverySkip     RecoveryAction = "skip"
	RecoveryRollback RecoveryAction = "rollback"
	RecoveryAbort    RecoveryAction = "abort"
)

// RecoveryStrategy defines how to handle errors for each step
type RecoveryStrategy struct {
	MaxRetries    int
	RetryDelay    time.Duration
	DefaultAction RecoveryAction
	StepActions   map[WorkflowStep]RecoveryAction
}

// DefaultRecoveryStrategy returns the default recovery strategy
func DefaultRecoveryStrategy() *RecoveryStrategy {
	return &RecoveryStrategy{
		MaxRetries:    3,
		RetryDelay:    time.Second * 2,
		DefaultAction: RecoveryAbort,
		StepActions: map[WorkflowStep]RecoveryAction{
			StepLoadAgent:      RecoveryAbort,      // Can't recover from missing agent
			StepValidateGitHub: RecoveryAbort,      // Can't proceed without valid token
			StepCreateSandbox:  RecoveryRetry,      // Sandbox creation can be retried
			StepLoadHistory:    RecoverySkip,       // Can proceed without history
			StepCreateBranch:   RecoveryRetry,      // Branch creation can be retried
			StepRunLLM:         RecoveryAbort,      // LLM failures need investigation
			StepSaveResponse:   RecoveryRetry,      // Database saves can be retried
			StepCommitChanges:  RecoverySkip,       // Can skip commit if no changes
			StepCleanupSandbox: RecoverySkip,       // Cleanup failures are non-fatal
			StepMarkComplete:   RecoveryRetry,      // Status updates can be retried
		},
	}
}

// RecoveryManager handles error recovery during workflow execution
type RecoveryManager struct {
	strategy    *RecoveryStrategy
	emitter     *EventEmitter
	retryCounts map[WorkflowStep]int
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager(strategy *RecoveryStrategy, emitter *EventEmitter) *RecoveryManager {
	if strategy == nil {
		strategy = DefaultRecoveryStrategy()
	}
	return &RecoveryManager{
		strategy:    strategy,
		emitter:     emitter,
		retryCounts: make(map[WorkflowStep]int),
	}
}

// HandleError determines how to handle an error for a given step
func (rm *RecoveryManager) HandleError(ctx context.Context, step WorkflowStep, err error) (RecoveryAction, error) {
	// Check for context cancellation
	if ctx.Err() != nil {
		return RecoveryAbort, ErrWorkflowCancelled
	}

	// Get the action for this step
	action, ok := rm.strategy.StepActions[step]
	if !ok {
		action = rm.strategy.DefaultAction
	}

	// Handle retry logic
	if action == RecoveryRetry {
		rm.retryCounts[step]++
		if rm.retryCounts[step] > rm.strategy.MaxRetries {
			return RecoveryAbort, fmt.Errorf("max retries exceeded for step %s: %w", step, err)
		}

		// Emit retry event
		if rm.emitter != nil {
			rm.emitter.EmitProgress(step, 0, fmt.Sprintf("Retrying step (attempt %d/%d)", rm.retryCounts[step], rm.strategy.MaxRetries))
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			return RecoveryAbort, ErrWorkflowCancelled
		case <-time.After(rm.strategy.RetryDelay):
		}

		return RecoveryRetry, nil
	}

	return action, err
}

// ShouldRetry checks if a step should be retried
func (rm *RecoveryManager) ShouldRetry(step WorkflowStep) bool {
	action, ok := rm.strategy.StepActions[step]
	if !ok {
		action = rm.strategy.DefaultAction
	}
	return action == RecoveryRetry && rm.retryCounts[step] < rm.strategy.MaxRetries
}

// ResetRetryCount resets the retry count for a step
func (rm *RecoveryManager) ResetRetryCount(step WorkflowStep) {
	delete(rm.retryCounts, step)
}

// RollbackHandler defines a function that can rollback a step
type RollbackHandler func(ctx context.Context, execCtx *ExecutionContext) error

// RollbackManager manages rollback operations
type RollbackManager struct {
	handlers map[WorkflowStep]RollbackHandler
	emitter  *EventEmitter
}

// NewRollbackManager creates a new rollback manager
func NewRollbackManager(emitter *EventEmitter) *RollbackManager {
	return &RollbackManager{
		handlers: make(map[WorkflowStep]RollbackHandler),
		emitter:  emitter,
	}
}

// RegisterHandler registers a rollback handler for a step
func (rm *RollbackManager) RegisterHandler(step WorkflowStep, handler RollbackHandler) {
	rm.handlers[step] = handler
}

// Rollback executes rollback for a specific step
func (rm *RollbackManager) Rollback(ctx context.Context, step WorkflowStep, execCtx *ExecutionContext) error {
	handler, ok := rm.handlers[step]
	if !ok {
		// No rollback handler registered, skip
		return nil
	}

	if rm.emitter != nil {
		rm.emitter.EmitProgress(step, 0, "Rolling back step")
	}

	if err := handler(ctx, execCtx); err != nil {
		if rm.emitter != nil {
			rm.emitter.EmitStepFailed(step, err, map[string]interface{}{
				"operation": "rollback",
			})
		}
		return fmt.Errorf("rollback failed for step %s: %w", step, err)
	}

	return nil
}

// RollbackAll executes rollback for all steps in reverse order
func (rm *RollbackManager) RollbackAll(ctx context.Context, completedSteps []WorkflowStep, execCtx *ExecutionContext) []error {
	var errors []error

	// Rollback in reverse order
	for i := len(completedSteps) - 1; i >= 0; i-- {
		step := completedSteps[i]
		if err := rm.Rollback(ctx, step, execCtx); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// ErrorClassifier classifies errors into categories for handling
type ErrorClassifier struct{}

// ErrorCategory represents a category of errors
type ErrorCategory string

const (
	ErrorCategoryTransient   ErrorCategory = "transient"   // Can be retried
	ErrorCategoryPermanent   ErrorCategory = "permanent"   // Cannot be retried
	ErrorCategoryUserError   ErrorCategory = "user_error"  // User configuration issue
	ErrorCategorySystemError ErrorCategory = "system"      // System/infrastructure issue
)

// Classify determines the category of an error
func (c *ErrorClassifier) Classify(err error) ErrorCategory {
	if err == nil {
		return ""
	}

	// Check for specific error types
	switch {
	case errors.Is(err, ErrWorkflowCancelled):
		return ErrorCategoryPermanent
	case errors.Is(err, ErrAgentNotFound):
		return ErrorCategoryUserError
	case errors.Is(err, ErrInvalidAgentConfig):
		return ErrorCategoryUserError
	case errors.Is(err, ErrGitHubTokenInvalid):
		return ErrorCategoryUserError
	case errors.Is(err, ErrGitHubTokenMissing):
		return ErrorCategoryUserError
	case errors.Is(err, ErrSandboxCreationFailed):
		return ErrorCategoryTransient
	case errors.Is(err, ErrCloneFailed):
		return ErrorCategoryTransient
	case errors.Is(err, ErrLLMExecutionFailed):
		return ErrorCategoryTransient
	case errors.Is(err, ErrMaxIterationsReached):
		return ErrorCategoryPermanent
	case errors.Is(err, ErrExecutionTimeout):
		return ErrorCategoryTransient
	case errors.Is(err, context.Canceled):
		return ErrorCategoryPermanent
	case errors.Is(err, context.DeadlineExceeded):
		return ErrorCategoryTransient
	default:
		return ErrorCategorySystemError
	}
}

// IsRetryable returns true if the error can potentially be retried
func (c *ErrorClassifier) IsRetryable(err error) bool {
	category := c.Classify(err)
	return category == ErrorCategoryTransient
}

// WrapStepError wraps an error with step context
func WrapStepError(step WorkflowStep, err error) error {
	if err == nil {
		return nil
	}
	return &StepError{
		Step:  step,
		Cause: err,
	}
}

// StepError represents an error that occurred during a specific workflow step
type StepError struct {
	Step  WorkflowStep
	Cause error
}

func (e *StepError) Error() string {
	return fmt.Sprintf("step %s failed: %v", e.Step, e.Cause)
}

func (e *StepError) Unwrap() error {
	return e.Cause
}
