package workflow

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/jacklau/prism/internal/agent"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/tools"
)

// StepExecutor defines the interface for step execution
type StepExecutor interface {
	Execute(ctx context.Context, step *Step, state map[string]interface{}) (*StepResult, error)
}

// AgentExecutor executes agent steps
type AgentExecutor struct {
	agentMgr   *agent.Manager
	llmManager *llm.Manager
}

// NewAgentExecutor creates a new agent executor
func NewAgentExecutor(agentMgr *agent.Manager, llmManager *llm.Manager) *AgentExecutor {
	return &AgentExecutor{
		agentMgr:   agentMgr,
		llmManager: llmManager,
	}
}

// Execute runs an agent step
func (e *AgentExecutor) Execute(ctx context.Context, step *Step, state map[string]interface{}) (*StepResult, error) {
	config := step.Config.AgentConfig
	if config == nil {
		return nil, fmt.Errorf("agent config is required for agent step")
	}

	// Interpolate prompt with state values
	prompt := interpolateState(config.Prompt, state)

	// Create agent config
	agentConfig := agent.AgentConfig{
		Provider:     config.Provider,
		Model:        config.Model,
		SystemPrompt: config.SystemPrompt,
		Temperature:  config.Temperature,
		MaxTokens:    config.MaxTokens,
	}

	// Create task
	task := agent.NewTask(prompt)

	// Run the agent
	execution, err := e.agentMgr.RunTask(ctx, task, agentConfig)
	if err != nil {
		return &StepResult{
			StepID:   step.ID,
			StepName: step.Name,
			Status:   StepStatusFailed,
			Error:    err.Error(),
		}, err
	}

	// Wait for completion
	execution.Wait()

	results := execution.GetResults()
	if len(results) == 0 {
		return &StepResult{
			StepID:   step.ID,
			StepName: step.Name,
			Status:   StepStatusFailed,
			Error:    "no results from agent",
		}, fmt.Errorf("no results from agent")
	}

	result := results[0]
	if !result.Success {
		return &StepResult{
			StepID:   step.ID,
			StepName: step.Name,
			Status:   StepStatusFailed,
			Error:    result.Error,
			Output:   result.Output,
		}, fmt.Errorf("agent failed: %s", result.Error)
	}

	return &StepResult{
		StepID:   step.ID,
		StepName: step.Name,
		Status:   StepStatusCompleted,
		Output:   result.Output,
		Metadata: map[string]interface{}{
			"usage":    result.Usage,
			"duration": result.Duration.Milliseconds(),
		},
	}, nil
}

// ToolExecutor executes tool steps
type ToolExecutor struct {
	registry *tools.Registry
}

// NewToolExecutor creates a new tool executor
func NewToolExecutor(registry *tools.Registry) *ToolExecutor {
	return &ToolExecutor{
		registry: registry,
	}
}

// Execute runs a tool step
func (e *ToolExecutor) Execute(ctx context.Context, step *Step, state map[string]interface{}) (*StepResult, error) {
	config := step.Config.ToolConfig
	if config == nil {
		return nil, fmt.Errorf("tool config is required for tool step")
	}

	// Interpolate parameters with state values
	params := interpolateParams(config.Parameters, state)

	// Get the tool from registry
	if e.registry == nil {
		return nil, fmt.Errorf("tool registry not available")
	}

	tool, ok := e.registry.Get(config.ToolName)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", config.ToolName)
	}

	// Execute the tool
	output, err := tool.Execute(ctx, params)
	if err != nil {
		return &StepResult{
			StepID:   step.ID,
			StepName: step.Name,
			Status:   StepStatusFailed,
			Error:    err.Error(),
		}, err
	}

	return &StepResult{
		StepID:   step.ID,
		StepName: step.Name,
		Status:   StepStatusCompleted,
		Output:   output,
	}, nil
}

// ConditionExecutor evaluates conditions
type ConditionExecutor struct{}

// NewConditionExecutor creates a new condition executor
func NewConditionExecutor() *ConditionExecutor {
	return &ConditionExecutor{}
}

// Execute evaluates a condition step
func (e *ConditionExecutor) Execute(ctx context.Context, step *Step, state map[string]interface{}) (*StepResult, error) {
	config := step.Config.ConditionConfig
	if config == nil {
		return nil, fmt.Errorf("condition config is required for condition step")
	}

	// Interpolate expression
	expr := interpolateState(config.Expression, state)

	// Evaluate the expression
	result := evaluateExpression(expr)

	branch := config.FalseBranch
	if result {
		branch = config.TrueBranch
	}

	return &StepResult{
		StepID:   step.ID,
		StepName: step.Name,
		Status:   StepStatusCompleted,
		Output: map[string]interface{}{
			"result":     result,
			"next_step":  branch,
			"expression": expr,
		},
		Metadata: map[string]interface{}{
			"condition_result": result,
			"next_branch":      branch,
		},
	}, nil
}

// ParallelExecutor executes steps in parallel
type ParallelExecutor struct {
	engine *Engine
}

// NewParallelExecutor creates a new parallel executor
func NewParallelExecutor(engine *Engine) *ParallelExecutor {
	return &ParallelExecutor{
		engine: engine,
	}
}

// Execute runs steps in parallel
func (e *ParallelExecutor) Execute(ctx context.Context, step *Step, state map[string]interface{}) (*StepResult, error) {
	config := step.Config.ParallelConfig
	if config == nil {
		return nil, fmt.Errorf("parallel config is required for parallel step")
	}

	if len(config.Steps) == 0 {
		return &StepResult{
			StepID:   step.ID,
			StepName: step.Name,
			Status:   StepStatusCompleted,
			Output:   map[string]interface{}{},
		}, nil
	}

	var wg sync.WaitGroup
	results := make([]*StepResult, len(config.Steps))
	errors := make([]error, len(config.Steps))
	errChan := make(chan int, len(config.Steps))

	// Create a shared state copy for parallel execution
	stateCopy := make(map[string]interface{})
	for k, v := range state {
		stateCopy[k] = v
	}

	for i, subStep := range config.Steps {
		wg.Add(1)
		go func(idx int, s Step) {
			defer wg.Done()

			executor, ok := e.engine.executors[s.Type]
			if !ok {
				errors[idx] = fmt.Errorf("no executor for step type: %s", s.Type)
				if config.FailOnFirst {
					errChan <- idx
				}
				return
			}

			result, err := executor.Execute(ctx, &s, stateCopy)
			results[idx] = result
			if err != nil {
				errors[idx] = err
				if config.FailOnFirst {
					errChan <- idx
				}
			}
		}(i, subStep)
	}

	// Wait for completion or first error
	if config.FailOnFirst {
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case idx := <-errChan:
			return &StepResult{
				StepID:   step.ID,
				StepName: step.Name,
				Status:   StepStatusFailed,
				Error:    errors[idx].Error(),
			}, errors[idx]
		case <-done:
			// All completed
		case <-ctx.Done():
			return &StepResult{
				StepID:   step.ID,
				StepName: step.Name,
				Status:   StepStatusFailed,
				Error:    ctx.Err().Error(),
			}, ctx.Err()
		}
	} else {
		wg.Wait()
	}

	// Collect results
	outputMap := make(map[string]interface{})
	var firstError error
	for i, result := range results {
		if result != nil {
			outputMap[config.Steps[i].ID] = result.Output
		}
		if errors[i] != nil && firstError == nil {
			firstError = errors[i]
		}
	}

	status := StepStatusCompleted
	var errMsg string
	if firstError != nil && !config.WaitForAll {
		status = StepStatusFailed
		errMsg = firstError.Error()
	}

	return &StepResult{
		StepID:   step.ID,
		StepName: step.Name,
		Status:   status,
		Output:   outputMap,
		Error:    errMsg,
	}, firstError
}

// WaitExecutor handles wait steps
type WaitExecutor struct {
	inputChan chan interface{}
}

// NewWaitExecutor creates a new wait executor
func NewWaitExecutor() *WaitExecutor {
	return &WaitExecutor{
		inputChan: make(chan interface{}, 1),
	}
}

// Execute waits for external input or timeout
func (e *WaitExecutor) Execute(ctx context.Context, step *Step, state map[string]interface{}) (*StepResult, error) {
	config := step.Config.WaitConfig
	if config == nil {
		return nil, fmt.Errorf("wait config is required for wait step")
	}

	// Check if input already exists in state
	inputKey := fmt.Sprintf("input_%s", step.ID)
	if val, exists := state[inputKey]; exists {
		return &StepResult{
			StepID:   step.ID,
			StepName: step.Name,
			Status:   StepStatusCompleted,
			Output:   val,
		}, nil
	}

	// Create timeout context if specified
	waitCtx := ctx
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	switch config.WaitType {
	case "timeout":
		// Just wait for the timeout
		<-waitCtx.Done()
		if waitCtx.Err() == context.DeadlineExceeded {
			return &StepResult{
				StepID:   step.ID,
				StepName: step.Name,
				Status:   StepStatusCompleted,
				Output:   nil,
			}, nil
		}
		return nil, waitCtx.Err()

	case "user_input", "webhook":
		// Poll state for input
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-waitCtx.Done():
				if waitCtx.Err() == context.DeadlineExceeded {
					return &StepResult{
						StepID:   step.ID,
						StepName: step.Name,
						Status:   StepStatusFailed,
						Error:    "timeout waiting for input",
					}, fmt.Errorf("timeout waiting for input")
				}
				return nil, waitCtx.Err()

			case <-ticker.C:
				// Check if input has been provided
				if val, exists := state[inputKey]; exists {
					return &StepResult{
						StepID:   step.ID,
						StepName: step.Name,
						Status:   StepStatusCompleted,
						Output:   val,
					}, nil
				}
			}
		}

	default:
		return nil, fmt.Errorf("unknown wait type: %s", config.WaitType)
	}
}

// TransformExecutor handles data transformation steps
type TransformExecutor struct{}

// NewTransformExecutor creates a new transform executor
func NewTransformExecutor() *TransformExecutor {
	return &TransformExecutor{}
}

// Execute transforms data according to the configuration
func (e *TransformExecutor) Execute(ctx context.Context, step *Step, state map[string]interface{}) (*StepResult, error) {
	config := step.Config.TransformConfig
	if config == nil {
		return nil, fmt.Errorf("transform config is required for transform step")
	}

	// Get input data
	var input interface{} = state
	if config.InputKey != "" {
		if val, exists := state[config.InputKey]; exists {
			input = val
		} else {
			input = nil
		}
	}

	var output interface{}
	var err error

	switch config.Type {
	case "template":
		output, err = e.executeTemplate(config.Template, state)

	case "mapping":
		output, err = e.executeMapping(config.Mapping, state)

	case "jq":
		// JQ-like transformation (simplified)
		output, err = e.executeJQ(config.Template, input)

	case "script":
		// Script execution (placeholder for future implementation)
		return nil, fmt.Errorf("script transformation not yet implemented")

	default:
		return nil, fmt.Errorf("unknown transform type: %s", config.Type)
	}

	if err != nil {
		return &StepResult{
			StepID:   step.ID,
			StepName: step.Name,
			Status:   StepStatusFailed,
			Error:    err.Error(),
		}, err
	}

	return &StepResult{
		StepID:   step.ID,
		StepName: step.Name,
		Status:   StepStatusCompleted,
		Output:   output,
	}, nil
}

// executeTemplate processes a Go template
func (e *TransformExecutor) executeTemplate(tmplStr string, state map[string]interface{}) (string, error) {
	tmpl, err := template.New("transform").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, state); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// executeMapping applies a simple key mapping
func (e *TransformExecutor) executeMapping(mapping map[string]string, state map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for newKey, sourceKey := range mapping {
		if val, exists := state[sourceKey]; exists {
			result[newKey] = val
		}
	}

	return result, nil
}

// executeJQ applies a simplified JQ-like transformation
func (e *TransformExecutor) executeJQ(query string, input interface{}) (interface{}, error) {
	// Simplified JQ - just support basic path access like ".field" or ".field.subfield"
	if query == "." {
		return input, nil
	}

	if !strings.HasPrefix(query, ".") {
		return nil, fmt.Errorf("JQ query must start with '.'")
	}

	parts := strings.Split(query[1:], ".")
	current := input

	for _, part := range parts {
		if part == "" {
			continue
		}

		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, nil
			}
		default:
			return nil, fmt.Errorf("cannot access field '%s' on non-object", part)
		}
	}

	return current, nil
}

// Helper functions

// interpolateState replaces {{state.key}} placeholders with actual values
func interpolateState(s string, state map[string]interface{}) string {
	re := regexp.MustCompile(`\{\{state\.([^}]+)\}\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		key := re.FindStringSubmatch(match)[1]
		if val, ok := state[key]; ok {
			return fmt.Sprintf("%v", val)
		}
		return match
	})
}

// interpolateParams recursively interpolates state values in parameters
func interpolateParams(params map[string]interface{}, state map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, val := range params {
		switch v := val.(type) {
		case string:
			result[key] = interpolateState(v, state)
		case map[string]interface{}:
			result[key] = interpolateParams(v, state)
		default:
			result[key] = val
		}
	}

	return result
}

// evaluateExpression evaluates a simple expression
func evaluateExpression(expr string) bool {
	expr = strings.TrimSpace(strings.ToLower(expr))

	// Handle common boolean expressions
	if expr == "true" || expr == "1" || expr == "yes" {
		return true
	}
	if expr == "false" || expr == "0" || expr == "no" || expr == "" {
		return false
	}

	// Handle equality checks like "value == other"
	if strings.Contains(expr, "==") {
		parts := strings.SplitN(expr, "==", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]) == strings.TrimSpace(parts[1])
		}
	}

	// Handle inequality checks like "value != other"
	if strings.Contains(expr, "!=") {
		parts := strings.SplitN(expr, "!=", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]) != strings.TrimSpace(parts[1])
		}
	}

	// Non-empty strings are truthy
	return len(expr) > 0
}
