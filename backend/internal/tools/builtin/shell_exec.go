package builtin

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

// ShellExecConfig holds configuration for shell command execution
type ShellExecConfig struct {
	AllowedCommands []string
	BlockedPatterns []string
	MaxTimeout      time.Duration
	MaxOutputSize   int
}

// DefaultShellExecConfig returns the default configuration
func DefaultShellExecConfig() ShellExecConfig {
	return ShellExecConfig{
		AllowedCommands: []string{
			// Package Managers
			"npm", "npx", "yarn", "pnpm", "pip", "pip3", "poetry", "cargo", "go",
			// Version Control
			"git",
			// Build Tools
			"make", "cmake", "gradle", "mvn",
			// Language Runtimes
			"node", "python", "python3", "ruby", "php",
			// Common Utilities
			"ls", "cat", "head", "tail", "grep", "find", "pwd", "echo", "mkdir", "cp", "mv", "touch", "rm", "wc", "sort", "uniq", "diff",
			// Container Tools
			"docker", "docker-compose", "kubectl",
			// Other Dev Tools
			"curl", "wget", "tar", "unzip", "zip", "ssh", "scp",
		},
		BlockedPatterns: []string{
			`rm\s+(-rf?|--recursive)\s+/\s*$`,
			`rm\s+(-rf?|--recursive)\s+/\*`,
			`rm\s+(-rf?|--recursive)\s+~`,
			`dd\s+if=`,
			`:(){:|:&};:`,
			`>\s*/dev/sd`,
			`mkfs\.`,
			`chmod\s+(-R\s+)?777\s+/`,
			`curl\s+.*\|\s*(ba)?sh`,
			`wget\s+.*\|\s*(ba)?sh`,
			`\bsudo\b`,
			`\bsu\b`,
			`>\s*/etc/`,
			`rm\s+.*\.(bashrc|profile|zshrc)`,
		},
		MaxTimeout:    30 * time.Minute,
		MaxOutputSize: 1024 * 1024, // 1MB
	}
}

// ShellOutputCallback is called with streaming output
type ShellOutputCallback func(stream string, content string)

// ShellExecTool executes shell commands in the user's workspace
type ShellExecTool struct {
	sandbox        *sandbox.Service
	config         ShellExecConfig
	activeCmds     map[string]context.CancelFunc
	mu             sync.Mutex
	outputCallback ShellOutputCallback
	backgroundMgr  *BackgroundShellManager
}

// NewShellExecTool creates a new shell execution tool
func NewShellExecTool(sandbox *sandbox.Service, config ShellExecConfig) *ShellExecTool {
	return &ShellExecTool{
		sandbox:    sandbox,
		config:     config,
		activeCmds: make(map[string]context.CancelFunc),
	}
}

// SetBackgroundManager sets the background shell manager for background execution support
func (t *ShellExecTool) SetBackgroundManager(mgr *BackgroundShellManager) {
	t.backgroundMgr = mgr
}

// SetOutputCallback sets a callback for streaming output
func (t *ShellExecTool) SetOutputCallback(cb ShellOutputCallback) {
	t.outputCallback = cb
}

func (t *ShellExecTool) Name() string {
	return "shell_execute"
}

func (t *ShellExecTool) Description() string {
	return `Execute shell commands in the user's workspace directory. Supports common development tools like npm, git, pip, python, node, docker, and more. Commands run directly on the host system within the workspace directory. Use this for installing dependencies, running builds, git operations, and other command-line tasks.`
}

func (t *ShellExecTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"command": {
				Type:        "string",
				Description: "The command to execute (e.g., 'npm', 'git', 'python')",
			},
			"args": {
				Type:        "array",
				Description: "Command arguments as an array (e.g., ['install', 'express'] for 'npm install express')",
			},
			"cwd": {
				Type:        "string",
				Description: "Optional working directory (relative to workspace). Defaults to workspace root.",
			},
			"timeout": {
				Type:        "number",
				Description: "Timeout in seconds (max 1800 = 30 minutes). Defaults to 300 (5 minutes).",
			},
			"run_in_background": {
				Type:        "boolean",
				Description: "Run the command in the background. Returns a shell_id for later retrieval with bash_output.",
			},
		},
		Required: []string{"command"},
	}
}

func (t *ShellExecTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	// Parse command
	command, ok := params["command"].(string)
	if !ok || command == "" {
		return nil, fmt.Errorf("command parameter is required")
	}

	// Validate command is in whitelist
	if !t.isCommandAllowed(command) {
		return nil, fmt.Errorf("command '%s' is not allowed. Allowed commands: %s", command, strings.Join(t.config.AllowedCommands, ", "))
	}

	// Parse arguments
	var args []string
	if argsRaw, ok := params["args"].([]interface{}); ok {
		for _, arg := range argsRaw {
			if argStr, ok := arg.(string); ok {
				args = append(args, argStr)
			}
		}
	}

	// Build full command string for pattern checking
	fullCommand := command + " " + strings.Join(args, " ")

	// Check for blocked patterns
	if blocked, pattern := t.isCommandBlocked(fullCommand); blocked {
		return nil, fmt.Errorf("command contains blocked pattern: %s", pattern)
	}

	// Get workspace directory
	workDir, err := t.sandbox.GetOrCreateWorkDir(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Handle custom working directory
	if cwdParam, ok := params["cwd"].(string); ok && cwdParam != "" {
		// Ensure cwd is relative and doesn't escape workspace
		cwdParam = filepath.Clean(cwdParam)
		if strings.HasPrefix(cwdParam, "..") || filepath.IsAbs(cwdParam) {
			return nil, fmt.Errorf("cwd must be a relative path within the workspace")
		}
		workDir = filepath.Join(workDir, cwdParam)
	}

	// Check for background execution
	runInBackground := false
	if bg, ok := params["run_in_background"].(bool); ok {
		runInBackground = bg
	}

	if runInBackground {
		if t.backgroundMgr == nil {
			return nil, fmt.Errorf("background execution not available")
		}

		shell, err := t.backgroundMgr.StartBackground(ctx, userID, command, args, workDir)
		if err != nil {
			return nil, fmt.Errorf("failed to start background shell: %w", err)
		}

		return map[string]interface{}{
			"shell_id":   shell.ID,
			"background": true,
			"command":    fullCommand,
			"message":    "Command started in background. Use bash_output to retrieve results.",
		}, nil
	}

	// Parse timeout (default 5 minutes, max 30 minutes)
	timeout := 5 * time.Minute
	if timeoutSec, ok := params["timeout"].(float64); ok {
		timeout = time.Duration(timeoutSec) * time.Second
		if timeout > t.config.MaxTimeout {
			timeout = t.config.MaxTimeout
		}
		if timeout < time.Second {
			timeout = time.Second
		}
	}

	// Create command with timeout context
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Store cancel function for potential cancellation
	commandID := fmt.Sprintf("%s-%d", userID, time.Now().UnixNano())
	t.mu.Lock()
	t.activeCmds[commandID] = cancel
	t.mu.Unlock()
	defer func() {
		t.mu.Lock()
		delete(t.activeCmds, commandID)
		t.mu.Unlock()
	}()

	// Create the command
	cmd := exec.CommandContext(execCtx, command, args...)
	cmd.Dir = workDir

	// Set up pipes for stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start command
	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Collect output
	var stdout, stderr strings.Builder
	var outputMu sync.Mutex
	var wg sync.WaitGroup

	// Helper to read from pipe and optionally stream
	readPipe := func(pipe *bufio.Scanner, builder *strings.Builder, streamName string) {
		defer wg.Done()
		for pipe.Scan() {
			line := pipe.Text()
			outputMu.Lock()
			if builder.Len() < t.config.MaxOutputSize {
				builder.WriteString(line)
				builder.WriteString("\n")
			}
			outputMu.Unlock()

			// Call output callback if set
			if t.outputCallback != nil {
				t.outputCallback(streamName, line)
			}
		}
	}

	wg.Add(2)
	go readPipe(bufio.NewScanner(stdoutPipe), &stdout, "stdout")
	go readPipe(bufio.NewScanner(stderrPipe), &stderr, "stderr")

	// Wait for output collection
	wg.Wait()

	// Wait for command to finish
	err = cmd.Wait()
	duration := time.Since(startTime)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else if execCtx.Err() == context.DeadlineExceeded {
			return map[string]interface{}{
				"stdout":    stdout.String(),
				"stderr":    stderr.String(),
				"exit_code": -1,
				"success":   false,
				"error":     "command timed out",
				"duration":  duration.Milliseconds(),
				"command":   fullCommand,
			}, nil
		} else if execCtx.Err() == context.Canceled {
			return map[string]interface{}{
				"stdout":    stdout.String(),
				"stderr":    stderr.String(),
				"exit_code": -1,
				"success":   false,
				"error":     "command was cancelled",
				"duration":  duration.Milliseconds(),
				"command":   fullCommand,
			}, nil
		}
	}

	return map[string]interface{}{
		"stdout":    stdout.String(),
		"stderr":    stderr.String(),
		"exit_code": exitCode,
		"success":   exitCode == 0,
		"duration":  duration.Milliseconds(),
		"command":   fullCommand,
		"cwd":       workDir,
	}, nil
}

func (t *ShellExecTool) RequiresConfirmation() bool {
	return true // Shell execution should require confirmation for safety
}

// isCommandAllowed checks if a command is in the whitelist
func (t *ShellExecTool) isCommandAllowed(command string) bool {
	// Extract just the command name (in case path is provided)
	command = filepath.Base(command)

	for _, allowed := range t.config.AllowedCommands {
		if command == allowed {
			return true
		}
	}
	return false
}

// isCommandBlocked checks if the full command matches any blocked patterns
func (t *ShellExecTool) isCommandBlocked(fullCommand string) (bool, string) {
	for _, pattern := range t.config.BlockedPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		if re.MatchString(fullCommand) {
			return true, pattern
		}
	}
	return false, ""
}

// CancelCommand cancels a running command by its ID
func (t *ShellExecTool) CancelCommand(commandID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if cancel, ok := t.activeCmds[commandID]; ok {
		cancel()
		delete(t.activeCmds, commandID)
		return nil
	}
	return fmt.Errorf("command not found: %s", commandID)
}

// GetAllowedCommands returns the list of allowed commands
func (t *ShellExecTool) GetAllowedCommands() []string {
	return t.config.AllowedCommands
}
