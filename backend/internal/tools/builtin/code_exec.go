package builtin

import (
	"context"
	"fmt"

	"github.com/jacklau/prism/internal/integrations/github"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/services/coderunner"
)

// CodeExecutionTool executes code in a sandboxed environment
type CodeExecutionTool struct {
	runner *coderunner.Runner
}

// NewCodeExecutionTool creates a new code execution tool
func NewCodeExecutionTool(runner *coderunner.Runner) *CodeExecutionTool {
	return &CodeExecutionTool{runner: runner}
}

func (t *CodeExecutionTool) Name() string {
	return "execute_code"
}

func (t *CodeExecutionTool) Description() string {
	return "Execute code in a sandboxed environment. Supports Python, Node.js (JavaScript), and Shell scripts. Returns the output (stdout/stderr) and exit code. Use this for running code, scripts, or shell commands."
}

func (t *CodeExecutionTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"code": {
				Type:        "string",
				Description: "The code or command to execute",
			},
			"environment": {
				Type:        "string",
				Description: "The execution environment",
				Enum:        []string{"python", "node", "shell"},
			},
		},
		Required: []string{"code", "environment"},
	}
}

func (t *CodeExecutionTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	code, ok := params["code"].(string)
	if !ok || code == "" {
		return nil, fmt.Errorf("code parameter is required")
	}

	environment, ok := params["environment"].(string)
	if !ok || environment == "" {
		return nil, fmt.Errorf("environment parameter is required")
	}

	// Validate environment
	switch environment {
	case "python", "node", "shell":
		// valid
	default:
		return nil, fmt.Errorf("invalid environment: %s (must be python, node, or shell)", environment)
	}

	req := &github.CodeRunRequest{
		Command:     code,
		Environment: environment,
	}

	result, err := t.runner.Run(req)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	return map[string]interface{}{
		"stdout":    result.Stdout,
		"stderr":    result.Stderr,
		"exit_code": result.ExitCode,
		"success":   result.ExitCode == 0,
		"duration":  result.Duration, // Already in milliseconds (int64)
	}, nil
}

func (t *CodeExecutionTool) RequiresConfirmation() bool {
	return true // Code execution should require confirmation for safety
}
