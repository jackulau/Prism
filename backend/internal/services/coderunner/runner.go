package coderunner

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/integrations/github"
)

// Runner executes code in sandbox environments
type Runner struct {
	config *Config
}

// Config contains configuration for the code runner
type Config struct {
	// Docker configuration
	DockerEnabled bool   `json:"docker_enabled"`
	DockerNetwork string `json:"docker_network"`

	// Resource limits
	MemoryLimit string        `json:"memory_limit"` // e.g., "512m"
	CPULimit    string        `json:"cpu_limit"`    // e.g., "0.5"
	Timeout     time.Duration `json:"timeout"`

	// Sandbox images
	NodeImage   string `json:"node_image"`
	PythonImage string `json:"python_image"`
	ShellImage  string `json:"shell_image"`

	// Working directory for non-Docker execution
	WorkDir string `json:"work_dir"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		DockerEnabled: false, // Default to local execution for simplicity
		MemoryLimit:   "512m",
		CPULimit:      "0.5",
		Timeout:       5 * time.Minute,
		NodeImage:     "node:18-alpine",
		PythonImage:   "python:3.11-alpine",
		ShellImage:    "alpine:latest",
		WorkDir:       "/tmp/coderunner",
	}
}

// NewRunner creates a new code runner
func NewRunner(config *Config) *Runner {
	if config == nil {
		config = DefaultConfig()
	}
	return &Runner{config: config}
}

// Run executes code based on the request
func (r *Runner) Run(request *github.CodeRunRequest) (*github.CodeExecutionResult, error) {
	startTime := time.Now()
	resultID := uuid.New().String()

	log.Printf("Starting code execution: id=%s, env=%s, cmd=%s",
		resultID, request.Environment, request.Command)

	var result *github.CodeExecutionResult
	var err error

	if r.config.DockerEnabled {
		result, err = r.runInDocker(request, resultID, startTime)
	} else {
		result, err = r.runLocally(request, resultID, startTime)
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

// runLocally executes code locally (for development or when Docker is not available)
func (r *Runner) runLocally(request *github.CodeRunRequest, resultID string, startTime time.Time) (*github.CodeExecutionResult, error) {
	timeout := r.config.Timeout
	if request.Timeout > 0 {
		timeout = time.Duration(request.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Determine the shell and command based on environment
	var shell, shellFlag string
	switch request.Environment {
	case "node":
		shell = "node"
		shellFlag = "-e"
	case "python":
		shell = "python3"
		shellFlag = "-c"
	case "shell", "bash", "":
		shell = "bash"
		shellFlag = "-c"
	default:
		shell = "bash"
		shellFlag = "-c"
	}

	var cmd *exec.Cmd
	if request.Environment == "node" || request.Environment == "python" {
		// For interpreted languages, pass the command directly
		cmd = exec.CommandContext(ctx, shell, shellFlag, request.Command)
	} else {
		// For shell commands
		cmd = exec.CommandContext(ctx, shell, shellFlag, request.Command)
	}

	// Set working directory
	workDir := request.WorkDir
	if workDir == "" {
		workDir = r.config.WorkDir
	}
	cmd.Dir = workDir

	// Set environment variables
	for k, v := range request.EnvVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	completedAt := time.Now()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return &github.CodeExecutionResult{
		ID:          resultID,
		Command:     request.Command,
		Environment: request.Environment,
		ExitCode:    exitCode,
		Stdout:      stdout.String(),
		Stderr:      stderr.String(),
		Duration:    completedAt.Sub(startTime).Milliseconds(),
		StartedAt:   startTime,
		CompletedAt: completedAt,
	}, nil
}

// runInDocker executes code in a Docker container
func (r *Runner) runInDocker(request *github.CodeRunRequest, resultID string, startTime time.Time) (*github.CodeExecutionResult, error) {
	timeout := r.config.Timeout
	if request.Timeout > 0 {
		timeout = time.Duration(request.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Select the appropriate Docker image
	var image string
	switch request.Environment {
	case "node":
		image = r.config.NodeImage
	case "python":
		image = r.config.PythonImage
	case "shell", "bash", "":
		image = r.config.ShellImage
	default:
		image = r.config.ShellImage
	}

	// Build docker run command
	args := []string{
		"run",
		"--rm",
		"--memory", r.config.MemoryLimit,
		"--cpus", r.config.CPULimit,
		"--network", "none", // Disable network by default for security
	}

	// Add environment variables
	for k, v := range request.EnvVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// Add the image and command
	args = append(args, image)

	// Add the appropriate shell command based on environment
	switch request.Environment {
	case "node":
		args = append(args, "node", "-e", request.Command)
	case "python":
		args = append(args, "python3", "-c", request.Command)
	default:
		args = append(args, "sh", "-c", request.Command)
	}

	cmd := exec.CommandContext(ctx, "docker", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	completedAt := time.Now()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return &github.CodeExecutionResult{
		ID:          resultID,
		Command:     request.Command,
		Environment: request.Environment,
		ExitCode:    exitCode,
		Stdout:      stdout.String(),
		Stderr:      stderr.String(),
		Duration:    completedAt.Sub(startTime).Milliseconds(),
		StartedAt:   startTime,
		CompletedAt: completedAt,
	}, nil
}

// RunScript executes a script file in the appropriate environment
func (r *Runner) RunScript(script, environment string, envVars map[string]string) (*github.CodeExecutionResult, error) {
	request := &github.CodeRunRequest{
		Command:     script,
		Environment: environment,
		EnvVars:     envVars,
	}
	return r.Run(request)
}

// SupportedEnvironments returns a list of supported execution environments
func (r *Runner) SupportedEnvironments() []string {
	return []string{"node", "python", "shell", "bash"}
}

// ValidateCommand performs basic validation on a command
func ValidateCommand(command string) error {
	// Check for obviously dangerous patterns
	dangerousPatterns := []string{
		"rm -rf /",
		"dd if=",
		":(){:|:&};:",  // Fork bomb
		"mkfs",
		"> /dev/sd",
		"chmod -R 777 /",
	}

	lowercaseCmd := strings.ToLower(command)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowercaseCmd, pattern) {
			return fmt.Errorf("potentially dangerous command detected: %s", pattern)
		}
	}

	return nil
}
