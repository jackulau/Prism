package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

// WorkflowExecutor orchestrates the complete agent execution workflow
type WorkflowExecutor struct {
	config       ExecutorConfig
	llmManager   *llm.Manager
	sandbox      *sandbox.Service
	repos        *WorkflowRepositories
	eventBus     *EventBus
	toolExecutor ToolExecutor
	gitClient    GitClient

	// Step handlers
	loadStep    LoadStepHandler
	setupStep   SetupStepHandler
	gitStep     GitStepHandler
	llmStep     LLMStepHandler
	persistStep PersistStepHandler
	cleanupStep CleanupStepHandler
}

// Step handler interfaces for dependency injection
type LoadStepHandler interface {
	LoadAgent(ctx context.Context, agentID string) (*AgentData, error)
	ValidateGitHubToken(ctx context.Context, userID string) (string, error)
	ValidateGitHubTokenOptional(ctx context.Context, userID string) (string, error)
}

type SetupStepHandler interface {
	CreateSandbox(ctx context.Context, userID string, repoURL string, token string, cloneDepth int) (*SandboxContext, error)
	LoadConversationHistory(ctx context.Context, conversationID string) ([]llm.Message, error)
	CreateConversation(ctx context.Context, userID, provider, model, systemPrompt string) (*repository.Conversation, error)
}

type GitStepHandler interface {
	CreateBranchIfNeeded(ctx context.Context, sandboxCtx *SandboxContext, agentID, agentName string) (string, error)
	CommitChanges(ctx context.Context, sandboxCtx *SandboxContext, message string, prefix string) (string, error)
	PushChanges(ctx context.Context, sandboxCtx *SandboxContext, token string) error
}

type LLMStepHandler interface {
	RunLLMLoop(ctx context.Context, config LLMLoopConfig) (*LLMResult, error)
}

type PersistStepHandler interface {
	SaveResponse(ctx context.Context, conversationID string, result *LLMResult) error
	TrackTokens(ctx context.Context, executionID, userID, provider, model string, usage *llm.Usage) error
}

type CleanupStepHandler interface {
	CleanupSandbox(ctx context.Context, sandboxCtx *SandboxContext, keepFiles bool) error
	MarkComplete(ctx context.Context, executionID, agentID, status, commitSHA string, errorMsg string, iterations int) error
	MarkFailed(ctx context.Context, executionID, agentID string, err error) error
}

// NewWorkflowExecutor creates a new workflow executor
func NewWorkflowExecutor(
	config ExecutorConfig,
	llmManager *llm.Manager,
	sandbox *sandbox.Service,
	repos *WorkflowRepositories,
	toolExecutor ToolExecutor,
	gitClient GitClient,
) *WorkflowExecutor {
	return &WorkflowExecutor{
		config:       config,
		llmManager:   llmManager,
		sandbox:      sandbox,
		repos:        repos,
		eventBus:     NewEventBus(),
		toolExecutor: toolExecutor,
		gitClient:    gitClient,
	}
}

// SetStepHandlers sets custom step handlers (for testing or custom implementations)
func (e *WorkflowExecutor) SetStepHandlers(
	loadStep LoadStepHandler,
	setupStep SetupStepHandler,
	gitStep GitStepHandler,
	llmStep LLMStepHandler,
	persistStep PersistStepHandler,
	cleanupStep CleanupStepHandler,
) {
	e.loadStep = loadStep
	e.setupStep = setupStep
	e.gitStep = gitStep
	e.llmStep = llmStep
	e.persistStep = persistStep
	e.cleanupStep = cleanupStep
}

// GetEventBus returns the event bus for subscribing to workflow events
func (e *WorkflowExecutor) GetEventBus() *EventBus {
	return e.eventBus
}

// Execute runs the complete 10-step workflow for an agent
func (e *WorkflowExecutor) Execute(ctx context.Context, agentID, userID string) (*ExecutionContext, error) {
	// Create execution context
	execCtx := &ExecutionContext{
		ExecutionID: uuid.New().String(),
		AgentID:     agentID,
		UserID:      userID,
		StartedAt:   time.Now(),
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, e.config.DefaultTimeout)
	execCtx.Context = ctx
	execCtx.Cancel = cancel
	defer cancel()

	// Create event emitter
	emitter := NewEventEmitter(e.eventBus, execCtx.ExecutionID, agentID)

	// Create recovery manager
	recovery := NewRecoveryManager(nil, emitter)

	// Create rollback manager
	rollback := NewRollbackManager(emitter)

	// Track completed steps for rollback
	var completedSteps []WorkflowStep

	// Create execution record
	if e.repos.AgentExecution != nil {
		_, err := e.repos.AgentExecution.Create(ctx, agentID, userID, "")
		if err != nil {
			return nil, fmt.Errorf("failed to create execution record: %w", err)
		}
	}

	// Define cleanup function
	cleanup := func(err error) {
		if e.cleanupStep != nil {
			_ = e.cleanupStep.CleanupSandbox(ctx, execCtx.SandboxCtx, false)
		}
		if err != nil && e.cleanupStep != nil {
			_ = e.cleanupStep.MarkFailed(ctx, execCtx.ExecutionID, agentID, err)
		}
	}

	// Step 1: Load agent
	execCtx.CurrentStep = StepLoadAgent
	emitter.EmitStepStarted(StepLoadAgent, nil)
	if e.repos.AgentExecution != nil {
		if err := e.repos.AgentExecution.UpdateStep(ctx, execCtx.ExecutionID, StepLoadAgent); err != nil {
			// Log but continue
		}
	}

	agent, err := e.executeLoadAgent(ctx, agentID, recovery, emitter)
	if err != nil {
		cleanup(err)
		return execCtx, err
	}
	execCtx.AgentData = agent
	completedSteps = append(completedSteps, StepLoadAgent)
	emitter.EmitStepCompleted(StepLoadAgent, map[string]interface{}{"agent_name": agent.Name})

	// Step 2: Validate GitHub token
	execCtx.CurrentStep = StepValidateGitHub
	emitter.EmitStepStarted(StepValidateGitHub, nil)
	if e.repos.AgentExecution != nil {
		if err := e.repos.AgentExecution.UpdateStep(ctx, execCtx.ExecutionID, StepValidateGitHub); err != nil {
			// Log but continue
		}
	}

	token, err := e.executeValidateGitHub(ctx, userID, agent.RepoURL != "", recovery, emitter)
	if err != nil {
		cleanup(err)
		return execCtx, err
	}
	execCtx.GitHubToken = token
	completedSteps = append(completedSteps, StepValidateGitHub)
	emitter.EmitStepCompleted(StepValidateGitHub, nil)

	// Step 3: Create sandbox
	execCtx.CurrentStep = StepCreateSandbox
	emitter.EmitStepStarted(StepCreateSandbox, nil)
	if e.repos.AgentExecution != nil {
		if err := e.repos.AgentExecution.UpdateStep(ctx, execCtx.ExecutionID, StepCreateSandbox); err != nil {
			// Log but continue
		}
	}

	sandboxCtx, err := e.executeCreateSandbox(ctx, userID, agent.RepoURL, token, recovery, emitter)
	if err != nil {
		cleanup(err)
		return execCtx, err
	}
	execCtx.SandboxCtx = sandboxCtx
	completedSteps = append(completedSteps, StepCreateSandbox)
	emitter.EmitStepCompleted(StepCreateSandbox, map[string]interface{}{"work_dir": sandboxCtx.WorkDir})

	// Register sandbox cleanup for rollback
	rollback.RegisterHandler(StepCreateSandbox, func(ctx context.Context, ec *ExecutionContext) error {
		if e.cleanupStep != nil {
			return e.cleanupStep.CleanupSandbox(ctx, ec.SandboxCtx, false)
		}
		return nil
	})

	// Step 4: Load conversation history
	execCtx.CurrentStep = StepLoadHistory
	emitter.EmitStepStarted(StepLoadHistory, nil)
	if e.repos.AgentExecution != nil {
		if err := e.repos.AgentExecution.UpdateStep(ctx, execCtx.ExecutionID, StepLoadHistory); err != nil {
			// Log but continue
		}
	}

	messages, err := e.executeLoadHistory(ctx, agent.ConversationID, recovery, emitter)
	if err != nil {
		// History loading is not critical, log and continue
		emitter.EmitStepFailed(StepLoadHistory, err, nil)
		messages = []llm.Message{}
	} else {
		completedSteps = append(completedSteps, StepLoadHistory)
		emitter.EmitStepCompleted(StepLoadHistory, map[string]interface{}{"message_count": len(messages)})
	}
	execCtx.Messages = messages

	// Step 5: Create branch if first run
	execCtx.CurrentStep = StepCreateBranch
	emitter.EmitStepStarted(StepCreateBranch, nil)
	if e.repos.AgentExecution != nil {
		if err := e.repos.AgentExecution.UpdateStep(ctx, execCtx.ExecutionID, StepCreateBranch); err != nil {
			// Log but continue
		}
	}

	branchName, err := e.executeCreateBranch(ctx, sandboxCtx, agentID, agent.Name, recovery, emitter)
	if err != nil {
		// Branch creation failure is not always critical
		emitter.EmitStepFailed(StepCreateBranch, err, nil)
	} else {
		sandboxCtx.BranchName = branchName
		completedSteps = append(completedSteps, StepCreateBranch)
		emitter.EmitStepCompleted(StepCreateBranch, map[string]interface{}{"branch": branchName})
	}

	// Update execution with branch name
	if branchName != "" && e.repos.AgentExecution != nil {
		_ = e.repos.AgentExecution.UpdateBranch(ctx, execCtx.ExecutionID, branchName)
	}

	// Step 6: Run LLM loop
	execCtx.CurrentStep = StepRunLLM
	emitter.EmitStepStarted(StepRunLLM, nil)
	if e.repos.AgentExecution != nil {
		if err := e.repos.AgentExecution.UpdateStep(ctx, execCtx.ExecutionID, StepRunLLM); err != nil {
			// Log but continue
		}
	}

	llmResult, err := e.executeRunLLM(ctx, agent, messages, sandboxCtx, emitter)
	if err != nil {
		cleanup(err)
		return execCtx, err
	}
	if llmResult == nil {
		llmResult = &LLMResult{
			Output:       "",
			Iterations:   0,
			Messages:     messages,
			FilesChanged: []string{},
		}
	}
	execCtx.LLMResult = llmResult
	completedSteps = append(completedSteps, StepRunLLM)
	emitter.EmitStepCompleted(StepRunLLM, map[string]interface{}{
		"iterations":    llmResult.Iterations,
		"files_changed": len(llmResult.FilesChanged),
	})

	// Step 7: Save response and track tokens
	execCtx.CurrentStep = StepSaveResponse
	emitter.EmitStepStarted(StepSaveResponse, nil)
	if e.repos.AgentExecution != nil {
		if err := e.repos.AgentExecution.UpdateStep(ctx, execCtx.ExecutionID, StepSaveResponse); err != nil {
			// Log but continue
		}
	}

	if err := e.executeSaveResponse(ctx, execCtx, agent, llmResult, recovery, emitter); err != nil {
		// Persistence failure is not critical, log and continue
		emitter.EmitStepFailed(StepSaveResponse, err, nil)
	} else {
		completedSteps = append(completedSteps, StepSaveResponse)
		emitter.EmitStepCompleted(StepSaveResponse, nil)
	}

	// Step 8: Commit changes
	execCtx.CurrentStep = StepCommitChanges
	emitter.EmitStepStarted(StepCommitChanges, nil)
	if e.repos.AgentExecution != nil {
		if err := e.repos.AgentExecution.UpdateStep(ctx, execCtx.ExecutionID, StepCommitChanges); err != nil {
			// Log but continue
		}
	}

	var commitSHA string
	if e.config.AutoCommit && len(llmResult.FilesChanged) > 0 {
		commitSHA, err = e.executeCommitChanges(ctx, sandboxCtx, llmResult.Output, token, recovery, emitter)
		if err != nil {
			// Commit failure is logged but doesn't fail the workflow
			emitter.EmitStepFailed(StepCommitChanges, err, nil)
		} else {
			completedSteps = append(completedSteps, StepCommitChanges)
			emitter.EmitStepCompleted(StepCommitChanges, map[string]interface{}{"commit_sha": commitSHA})
		}
	} else {
		emitter.EmitStepCompleted(StepCommitChanges, map[string]interface{}{"skipped": true})
	}

	// Step 9: Cleanup sandbox
	execCtx.CurrentStep = StepCleanupSandbox
	emitter.EmitStepStarted(StepCleanupSandbox, nil)
	if e.repos.AgentExecution != nil {
		if err := e.repos.AgentExecution.UpdateStep(ctx, execCtx.ExecutionID, StepCleanupSandbox); err != nil {
			// Log but continue
		}
	}

	if err := e.executeCleanupSandbox(ctx, sandboxCtx, emitter); err != nil {
		// Cleanup failure is not critical
		emitter.EmitStepFailed(StepCleanupSandbox, err, nil)
	} else {
		completedSteps = append(completedSteps, StepCleanupSandbox)
		emitter.EmitStepCompleted(StepCleanupSandbox, nil)
	}

	// Step 10: Mark complete
	execCtx.CurrentStep = StepMarkComplete
	emitter.EmitStepStarted(StepMarkComplete, nil)

	if err := e.executeMarkComplete(ctx, execCtx.ExecutionID, agentID, commitSHA, llmResult.Iterations, emitter); err != nil {
		emitter.EmitStepFailed(StepMarkComplete, err, nil)
		return execCtx, err
	}
	completedSteps = append(completedSteps, StepMarkComplete)
	emitter.EmitStepCompleted(StepMarkComplete, nil)

	// Set completion time
	now := time.Now()
	execCtx.CompletedAt = &now

	return execCtx, nil
}

// Step execution methods

func (e *WorkflowExecutor) executeLoadAgent(ctx context.Context, agentID string, recovery *RecoveryManager, emitter *EventEmitter) (*AgentData, error) {
	if e.loadStep != nil {
		return e.loadStep.LoadAgent(ctx, agentID)
	}

	// Default implementation using repository
	if e.repos.Agent == nil {
		return nil, ErrAgentNotFound
	}

	return e.repos.Agent.GetByID(ctx, agentID)
}

func (e *WorkflowExecutor) executeValidateGitHub(ctx context.Context, userID string, required bool, recovery *RecoveryManager, emitter *EventEmitter) (string, error) {
	if e.loadStep != nil {
		if required {
			return e.loadStep.ValidateGitHubToken(ctx, userID)
		}
		return e.loadStep.ValidateGitHubTokenOptional(ctx, userID)
	}

	// Default implementation
	if e.repos.GitHub == nil {
		if required {
			return "", ErrGitHubTokenMissing
		}
		return "", nil
	}

	token, err := e.repos.GitHub.GetToken(ctx, userID)
	if err != nil && required {
		return "", err
	}

	return token, nil
}

func (e *WorkflowExecutor) executeCreateSandbox(ctx context.Context, userID, repoURL, token string, recovery *RecoveryManager, emitter *EventEmitter) (*SandboxContext, error) {
	if e.setupStep != nil {
		return e.setupStep.CreateSandbox(ctx, userID, repoURL, token, e.config.CloneDepth)
	}

	// Default implementation
	workDir, err := e.sandbox.GetOrCreateWorkDir(userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSandboxCreationFailed, err)
	}

	sandboxCtx := &SandboxContext{
		UserID:     userID,
		WorkDir:    workDir,
		RepoURL:    repoURL,
		Service:    e.sandbox,
		GitClient:  e.gitClient,
		CloneDepth: e.config.CloneDepth,
	}

	// Clone repository if URL provided
	if repoURL != "" && e.gitClient != nil {
		if err := e.gitClient.Clone(ctx, repoURL, workDir, token); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrCloneFailed, err)
		}
	}

	return sandboxCtx, nil
}

func (e *WorkflowExecutor) executeLoadHistory(ctx context.Context, conversationID string, recovery *RecoveryManager, emitter *EventEmitter) ([]llm.Message, error) {
	if e.setupStep != nil {
		return e.setupStep.LoadConversationHistory(ctx, conversationID)
	}

	// Default implementation
	if conversationID == "" || e.repos.Message == nil {
		return []llm.Message{}, nil
	}

	messages, err := e.repos.Message.ListByConversationID(conversationID)
	if err != nil {
		return nil, err
	}

	// Convert to LLM messages
	llmMessages := make([]llm.Message, 0, len(messages))
	for _, msg := range messages {
		llmMsg := llm.Message{
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
		}
		llmMessages = append(llmMessages, llmMsg)
	}

	return llmMessages, nil
}

func (e *WorkflowExecutor) executeCreateBranch(ctx context.Context, sandboxCtx *SandboxContext, agentID, agentName string, recovery *RecoveryManager, emitter *EventEmitter) (string, error) {
	if e.gitStep != nil {
		return e.gitStep.CreateBranchIfNeeded(ctx, sandboxCtx, agentID, agentName)
	}

	// Default implementation - skip if no git client
	if sandboxCtx == nil || sandboxCtx.GitClient == nil {
		return "", nil
	}

	// Use existing branch if set
	if sandboxCtx.BranchName != "" {
		return sandboxCtx.BranchName, nil
	}

	return "", nil
}

func (e *WorkflowExecutor) executeRunLLM(ctx context.Context, agent *AgentData, messages []llm.Message, sandboxCtx *SandboxContext, emitter *EventEmitter) (*LLMResult, error) {
	if e.llmStep != nil {
		config := LLMLoopConfig{
			Agent:         agent,
			Messages:      messages,
			Tools:         agent.Tools,
			MaxIterations: e.config.MaxLLMIterations,
			SandboxCtx:    sandboxCtx,
			ToolExecutor:  e.toolExecutor,
		}
		return e.llmStep.RunLLMLoop(ctx, config)
	}

	// Minimal default implementation
	return &LLMResult{
		Output:     "",
		Iterations: 0,
		Messages:   messages,
	}, nil
}

func (e *WorkflowExecutor) executeSaveResponse(ctx context.Context, execCtx *ExecutionContext, agent *AgentData, result *LLMResult, recovery *RecoveryManager, emitter *EventEmitter) error {
	if e.persistStep != nil {
		if err := e.persistStep.SaveResponse(ctx, agent.ConversationID, result); err != nil {
			return err
		}
		return e.persistStep.TrackTokens(ctx, execCtx.ExecutionID, execCtx.UserID, agent.Provider, agent.Model, result.TokenUsage)
	}

	// Default: no persistence
	return nil
}

func (e *WorkflowExecutor) executeCommitChanges(ctx context.Context, sandboxCtx *SandboxContext, message, token string, recovery *RecoveryManager, emitter *EventEmitter) (string, error) {
	if e.gitStep != nil {
		commitSHA, err := e.gitStep.CommitChanges(ctx, sandboxCtx, message, e.config.CommitMessagePrefix)
		if err != nil {
			return "", err
		}

		// Push if token available
		if token != "" {
			if err := e.gitStep.PushChanges(ctx, sandboxCtx, token); err != nil {
				return commitSHA, err // Return SHA even if push fails
			}
		}

		return commitSHA, nil
	}

	return "", nil
}

func (e *WorkflowExecutor) executeCleanupSandbox(ctx context.Context, sandboxCtx *SandboxContext, emitter *EventEmitter) error {
	if e.cleanupStep != nil {
		return e.cleanupStep.CleanupSandbox(ctx, sandboxCtx, false)
	}
	return nil
}

func (e *WorkflowExecutor) executeMarkComplete(ctx context.Context, executionID, agentID, commitSHA string, iterations int, emitter *EventEmitter) error {
	if e.cleanupStep != nil {
		return e.cleanupStep.MarkComplete(ctx, executionID, agentID, "completed", commitSHA, "", iterations)
	}

	// Default: update execution repo directly
	if e.repos.AgentExecution != nil {
		return e.repos.AgentExecution.Complete(ctx, executionID, commitSHA, iterations)
	}

	return nil
}

// Cancel cancels a running workflow execution
func (e *WorkflowExecutor) Cancel(executionID string) {
	// This would need to be implemented with execution tracking
	// For now, it's a placeholder
}

// GetExecution retrieves the status of an execution
func (e *WorkflowExecutor) GetExecution(ctx context.Context, executionID string) (*AgentExecution, error) {
	if e.repos.AgentExecution == nil {
		return nil, fmt.Errorf("execution repository not configured")
	}
	return e.repos.AgentExecution.GetByID(ctx, executionID)
}
