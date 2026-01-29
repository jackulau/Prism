package workflow

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
)

// Mock implementations for testing

type mockAgentRepository struct {
	agent *AgentData
	err   error
}

func (m *mockAgentRepository) GetByID(ctx context.Context, id string) (*AgentData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.agent, nil
}

func (m *mockAgentRepository) UpdateStatus(ctx context.Context, id string, status string, errorMsg string) error {
	return nil
}

func (m *mockAgentRepository) UpdateBranch(ctx context.Context, id string, branchName string) error {
	return nil
}

type mockGitHubRepository struct {
	token string
	valid bool
	err   error
}

func (m *mockGitHubRepository) GetToken(ctx context.Context, userID string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.token, nil
}

func (m *mockGitHubRepository) ValidateToken(ctx context.Context, token string) (bool, error) {
	return m.valid, nil
}

type mockLoadStep struct {
	agent       *AgentData
	token       string
	loadErr     error
	validateErr error
}

func (m *mockLoadStep) LoadAgent(ctx context.Context, agentID string) (*AgentData, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	return m.agent, nil
}

func (m *mockLoadStep) ValidateGitHubToken(ctx context.Context, userID string) (string, error) {
	if m.validateErr != nil {
		return "", m.validateErr
	}
	return m.token, nil
}

func (m *mockLoadStep) ValidateGitHubTokenOptional(ctx context.Context, userID string) (string, error) {
	return m.token, nil
}

type mockSetupStep struct {
	sandboxCtx *SandboxContext
	messages   []llm.Message
	conv       *repository.Conversation
	err        error
}

func (m *mockSetupStep) CreateSandbox(ctx context.Context, userID string, repoURL string, token string, cloneDepth int) (*SandboxContext, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.sandboxCtx, nil
}

func (m *mockSetupStep) LoadConversationHistory(ctx context.Context, conversationID string) ([]llm.Message, error) {
	return m.messages, nil
}

func (m *mockSetupStep) CreateConversation(ctx context.Context, userID, provider, model, systemPrompt string) (*repository.Conversation, error) {
	return m.conv, nil
}

type mockGitStep struct {
	branchName string
	commitSHA  string
	err        error
}

func (m *mockGitStep) CreateBranchIfNeeded(ctx context.Context, sandboxCtx *SandboxContext, agentID, agentName string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.branchName, nil
}

func (m *mockGitStep) CommitChanges(ctx context.Context, sandboxCtx *SandboxContext, message string, prefix string) (string, error) {
	return m.commitSHA, nil
}

func (m *mockGitStep) PushChanges(ctx context.Context, sandboxCtx *SandboxContext, token string) error {
	return nil
}

type mockLLMStep struct {
	result *LLMResult
	err    error
}

func (m *mockLLMStep) RunLLMLoop(ctx context.Context, config LLMLoopConfig) (*LLMResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

type mockPersistStep struct {
	err error
}

func (m *mockPersistStep) SaveResponse(ctx context.Context, conversationID string, result *LLMResult) error {
	return m.err
}

func (m *mockPersistStep) TrackTokens(ctx context.Context, executionID, userID, provider, model string, usage *llm.Usage) error {
	return nil
}

type mockCleanupStep struct {
	err error
}

func (m *mockCleanupStep) CleanupSandbox(ctx context.Context, sandboxCtx *SandboxContext, keepFiles bool) error {
	return nil
}

func (m *mockCleanupStep) MarkComplete(ctx context.Context, executionID, agentID, status, commitSHA string, errorMsg string, iterations int) error {
	return m.err
}

func (m *mockCleanupStep) MarkFailed(ctx context.Context, executionID, agentID string, err error) error {
	return nil
}

// Tests

func TestWorkflowExecutor_Execute_Success(t *testing.T) {
	// Setup mocks
	agent := &AgentData{
		ID:       "agent-1",
		UserID:   "user-1",
		Name:     "Test Agent",
		Provider: "openai",
		Model:    "gpt-4",
	}

	sandboxCtx := &SandboxContext{
		UserID:  "user-1",
		WorkDir: "/tmp/test",
	}

	llmResult := &LLMResult{
		Output:     "Task completed",
		Iterations: 1,
		Messages:   []llm.Message{},
	}

	// Create executor with mock repos
	config := DefaultExecutorConfig()
	config.DefaultTimeout = 5 * time.Second

	executor := NewWorkflowExecutor(config, nil, nil, &WorkflowRepositories{
		AgentExecution: nil, // Skip execution tracking for this test
	}, nil, nil)

	// Set mock step handlers
	executor.SetStepHandlers(
		&mockLoadStep{agent: agent, token: "gh-token"},
		&mockSetupStep{sandboxCtx: sandboxCtx, messages: []llm.Message{}},
		&mockGitStep{branchName: "agent/test-123"},
		&mockLLMStep{result: llmResult},
		&mockPersistStep{},
		&mockCleanupStep{},
	)

	// Execute
	ctx := context.Background()
	execCtx, err := executor.Execute(ctx, "agent-1", "user-1")

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if execCtx.AgentData != agent {
		t.Error("Agent data not set correctly")
	}

	if execCtx.LLMResult != llmResult {
		t.Error("LLM result not set correctly")
	}

	if execCtx.CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}
}

func TestWorkflowExecutor_Execute_AgentNotFound(t *testing.T) {
	config := DefaultExecutorConfig()
	config.DefaultTimeout = 5 * time.Second

	executor := NewWorkflowExecutor(config, nil, nil, &WorkflowRepositories{}, nil, nil)

	executor.SetStepHandlers(
		&mockLoadStep{loadErr: ErrAgentNotFound},
		nil, nil, nil, nil,
		&mockCleanupStep{},
	)

	ctx := context.Background()
	_, err := executor.Execute(ctx, "agent-1", "user-1")

	if !errors.Is(err, ErrAgentNotFound) {
		t.Errorf("Expected ErrAgentNotFound, got: %v", err)
	}
}

func TestWorkflowExecutor_Execute_LLMFailure(t *testing.T) {
	agent := &AgentData{
		ID:       "agent-1",
		Provider: "openai",
		Model:    "gpt-4",
	}

	sandboxCtx := &SandboxContext{
		UserID:  "user-1",
		WorkDir: "/tmp/test",
	}

	config := DefaultExecutorConfig()
	config.DefaultTimeout = 5 * time.Second

	executor := NewWorkflowExecutor(config, nil, nil, &WorkflowRepositories{}, nil, nil)

	executor.SetStepHandlers(
		&mockLoadStep{agent: agent, token: ""},
		&mockSetupStep{sandboxCtx: sandboxCtx},
		&mockGitStep{},
		&mockLLMStep{err: ErrLLMExecutionFailed},
		nil,
		&mockCleanupStep{},
	)

	ctx := context.Background()
	_, err := executor.Execute(ctx, "agent-1", "user-1")

	if !errors.Is(err, ErrLLMExecutionFailed) {
		t.Errorf("Expected ErrLLMExecutionFailed, got: %v", err)
	}
}

func TestWorkflowExecutor_Execute_Cancellation(t *testing.T) {
	agent := &AgentData{
		ID:       "agent-1",
		Provider: "openai",
		Model:    "gpt-4",
	}

	sandboxCtx := &SandboxContext{
		UserID:  "user-1",
		WorkDir: "/tmp/test",
	}

	// Create an LLM step that returns context.Canceled
	cancelledLLMStep := &mockLLMStep{err: context.Canceled}

	config := DefaultExecutorConfig()
	config.DefaultTimeout = 5 * time.Second

	executor := NewWorkflowExecutor(config, nil, nil, &WorkflowRepositories{}, nil, nil)

	executor.SetStepHandlers(
		&mockLoadStep{agent: agent, token: ""},
		&mockSetupStep{sandboxCtx: sandboxCtx},
		&mockGitStep{},
		cancelledLLMStep,
		nil,
		&mockCleanupStep{},
	)

	ctx := context.Background()
	_, err := executor.Execute(ctx, "agent-1", "user-1")

	// Should get context cancelled error
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}
}

func TestEventBus_PublishSubscribe(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	// Subscribe to events
	ch, unsubscribe := bus.Subscribe("exec-1")
	defer unsubscribe()

	// Publish an event
	event := WorkflowEvent{
		ExecutionID: "exec-1",
		AgentID:     "agent-1",
		Step:        StepLoadAgent,
		Status:      "started",
		Timestamp:   time.Now(),
	}

	go bus.Publish(event)

	// Receive the event
	select {
	case received := <-ch:
		if received.ExecutionID != event.ExecutionID {
			t.Error("Received event doesn't match")
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for event")
	}
}

func TestEventBus_WildcardSubscription(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	// Subscribe to all events
	ch, unsubscribe := bus.SubscribeAll()
	defer unsubscribe()

	// Publish an event for a different execution
	event := WorkflowEvent{
		ExecutionID: "exec-2",
		AgentID:     "agent-1",
		Step:        StepLoadAgent,
		Status:      "started",
		Timestamp:   time.Now(),
	}

	go bus.Publish(event)

	// Should still receive the event
	select {
	case received := <-ch:
		if received.ExecutionID != event.ExecutionID {
			t.Error("Received event doesn't match")
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for event")
	}
}

func TestRecoveryManager_RetryLogic(t *testing.T) {
	strategy := DefaultRecoveryStrategy()
	strategy.MaxRetries = 3
	strategy.RetryDelay = 1 * time.Millisecond

	rm := NewRecoveryManager(strategy, nil)

	ctx := context.Background()
	testErr := errors.New("test error")

	// First attempt should return retry
	action, err := rm.HandleError(ctx, StepCreateSandbox, testErr)
	if action != RecoveryRetry {
		t.Errorf("Expected RecoveryRetry, got %v", action)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Retry until max retries exceeded
	for i := 0; i < 3; i++ {
		rm.HandleError(ctx, StepCreateSandbox, testErr)
	}

	// Should now return abort
	action, err = rm.HandleError(ctx, StepCreateSandbox, testErr)
	if action != RecoveryAbort {
		t.Errorf("Expected RecoveryAbort after max retries, got %v", action)
	}
}

func TestErrorClassifier(t *testing.T) {
	classifier := &ErrorClassifier{}

	tests := []struct {
		err      error
		expected ErrorCategory
	}{
		{ErrAgentNotFound, ErrorCategoryUserError},
		{ErrGitHubTokenMissing, ErrorCategoryUserError},
		{ErrSandboxCreationFailed, ErrorCategoryTransient},
		{ErrWorkflowCancelled, ErrorCategoryPermanent},
		{ErrMaxIterationsReached, ErrorCategoryPermanent},
		{context.Canceled, ErrorCategoryPermanent},
		{context.DeadlineExceeded, ErrorCategoryTransient},
		{errors.New("unknown error"), ErrorCategorySystemError},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			category := classifier.Classify(tt.err)
			if category != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, category)
			}
		})
	}
}

func TestStepError(t *testing.T) {
	cause := errors.New("original error")
	stepErr := WrapStepError(StepLoadAgent, cause)

	if stepErr == nil {
		t.Fatal("Expected non-nil error")
	}

	// Check error message
	expected := "step load_agent failed: original error"
	if stepErr.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, stepErr.Error())
	}

	// Check unwrap
	if !errors.Is(stepErr, cause) {
		t.Error("Unwrap should return original cause")
	}
}

func TestDefaultExecutorConfig(t *testing.T) {
	config := DefaultExecutorConfig()

	if config.MaxLLMIterations <= 0 {
		t.Error("MaxLLMIterations should be positive")
	}

	if config.DefaultTimeout <= 0 {
		t.Error("DefaultTimeout should be positive")
	}

	if config.CommitMessagePrefix == "" {
		t.Error("CommitMessagePrefix should not be empty")
	}
}
