package builtin

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

const (
	maxBackgroundShells    = 10
	maxBackgroundShellTime = 30 * time.Minute
	maxBackgroundOutput    = 1024 * 1024 // 1MB
)

// BackgroundShell represents a shell process running in the background
type BackgroundShell struct {
	ID        string
	UserID    string
	Command   string
	Args      []string
	WorkDir   string
	StartTime time.Time
	Stdout    *strings.Builder
	Stderr    *strings.Builder
	Done      bool
	ExitCode  int
	Error     string
	cancel    context.CancelFunc
	mu        sync.Mutex
}

// BackgroundShellManager manages background shell processes
type BackgroundShellManager struct {
	shells  map[string]*BackgroundShell
	mu      sync.RWMutex
	sandbox *sandbox.Service
	config  ShellExecConfig
}

// NewBackgroundShellManager creates a new background shell manager
func NewBackgroundShellManager(sandbox *sandbox.Service, config ShellExecConfig) *BackgroundShellManager {
	return &BackgroundShellManager{
		shells:  make(map[string]*BackgroundShell),
		sandbox: sandbox,
		config:  config,
	}
}

// StartBackground starts a command in the background
func (m *BackgroundShellManager) StartBackground(ctx context.Context, userID, command string, args []string, workDir string) (*BackgroundShell, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Count user's shells
	userShellCount := 0
	for _, shell := range m.shells {
		if shell.UserID == userID && !shell.Done {
			userShellCount++
		}
	}
	if userShellCount >= maxBackgroundShells {
		return nil, fmt.Errorf("maximum number of background shells reached (%d)", maxBackgroundShells)
	}

	// Create shell
	shellID := uuid.New().String()[:8]
	shellCtx, cancel := context.WithTimeout(context.Background(), maxBackgroundShellTime)

	shell := &BackgroundShell{
		ID:        shellID,
		UserID:    userID,
		Command:   command,
		Args:      args,
		WorkDir:   workDir,
		StartTime: time.Now(),
		Stdout:    &strings.Builder{},
		Stderr:    &strings.Builder{},
		cancel:    cancel,
	}

	m.shells[shellID] = shell

	// Start the command
	go m.runCommand(shellCtx, shell)

	return shell, nil
}

func (m *BackgroundShellManager) runCommand(ctx context.Context, shell *BackgroundShell) {
	defer shell.cancel()

	cmd := exec.CommandContext(ctx, shell.Command, shell.Args...)
	cmd.Dir = shell.WorkDir

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		shell.mu.Lock()
		shell.Done = true
		shell.Error = fmt.Sprintf("failed to create stdout pipe: %v", err)
		shell.ExitCode = -1
		shell.mu.Unlock()
		return
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		shell.mu.Lock()
		shell.Done = true
		shell.Error = fmt.Sprintf("failed to create stderr pipe: %v", err)
		shell.ExitCode = -1
		shell.mu.Unlock()
		return
	}

	if err := cmd.Start(); err != nil {
		shell.mu.Lock()
		shell.Done = true
		shell.Error = fmt.Sprintf("failed to start command: %v", err)
		shell.ExitCode = -1
		shell.mu.Unlock()
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Read stdout
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			shell.mu.Lock()
			if shell.Stdout.Len() < maxBackgroundOutput {
				shell.Stdout.WriteString(scanner.Text())
				shell.Stdout.WriteString("\n")
			}
			shell.mu.Unlock()
		}
	}()

	// Read stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			shell.mu.Lock()
			if shell.Stderr.Len() < maxBackgroundOutput {
				shell.Stderr.WriteString(scanner.Text())
				shell.Stderr.WriteString("\n")
			}
			shell.mu.Unlock()
		}
	}()

	wg.Wait()
	err = cmd.Wait()

	shell.mu.Lock()
	shell.Done = true
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			shell.ExitCode = exitErr.ExitCode()
		} else if ctx.Err() == context.DeadlineExceeded {
			shell.Error = "command timed out"
			shell.ExitCode = -1
		} else if ctx.Err() == context.Canceled {
			shell.Error = "command was cancelled"
			shell.ExitCode = -1
		} else {
			shell.Error = err.Error()
			shell.ExitCode = -1
		}
	}
	shell.mu.Unlock()
}

// GetShell retrieves a background shell by ID
func (m *BackgroundShellManager) GetShell(shellID string) (*BackgroundShell, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	shell, ok := m.shells[shellID]
	return shell, ok
}

// KillShell terminates a background shell
func (m *BackgroundShellManager) KillShell(shellID string, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	shell, ok := m.shells[shellID]
	if !ok {
		return fmt.Errorf("shell not found: %s", shellID)
	}

	if shell.UserID != userID {
		return fmt.Errorf("shell not found: %s", shellID)
	}

	if shell.cancel != nil {
		shell.cancel()
	}

	return nil
}

// ListUserShells returns all shells for a user
func (m *BackgroundShellManager) ListUserShells(userID string) []*BackgroundShell {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var shells []*BackgroundShell
	for _, shell := range m.shells {
		if shell.UserID == userID {
			shells = append(shells, shell)
		}
	}
	return shells
}

// CleanupOldShells removes completed shells older than the TTL
func (m *BackgroundShellManager) CleanupOldShells(maxAge time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for id, shell := range m.shells {
		if shell.Done && now.Sub(shell.StartTime) > maxAge {
			delete(m.shells, id)
		}
	}
}

// BashOutputTool retrieves output from a background shell
type BashOutputTool struct {
	shellManager *BackgroundShellManager
}

// NewBashOutputTool creates a new bash output tool
func NewBashOutputTool(shellManager *BackgroundShellManager) *BashOutputTool {
	return &BashOutputTool{shellManager: shellManager}
}

func (t *BashOutputTool) Name() string {
	return "bash_output"
}

func (t *BashOutputTool) Description() string {
	return "Retrieves output from a running or completed background shell. Returns stdout, stderr, and status information. Use the shell_id from a background shell_execute command."
}

func (t *BashOutputTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"shell_id": {
				Type:        "string",
				Description: "The ID of the background shell to retrieve output from",
			},
			"filter": {
				Type:        "string",
				Description: "Optional regex pattern to filter output lines",
			},
		},
		Required: []string{"shell_id"},
	}
}

func (t *BashOutputTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	shellID, ok := params["shell_id"].(string)
	if !ok || shellID == "" {
		return nil, fmt.Errorf("shell_id parameter is required")
	}

	shell, ok := t.shellManager.GetShell(shellID)
	if !ok {
		return nil, fmt.Errorf("shell not found: %s", shellID)
	}

	if shell.UserID != userID {
		return nil, fmt.Errorf("shell not found: %s", shellID)
	}

	shell.mu.Lock()
	stdout := shell.Stdout.String()
	stderr := shell.Stderr.String()
	done := shell.Done
	exitCode := shell.ExitCode
	errorMsg := shell.Error
	shell.mu.Unlock()

	// Apply filter if provided
	filterPattern, _ := params["filter"].(string)
	if filterPattern != "" {
		re, err := regexp.Compile(filterPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid filter pattern: %w", err)
		}
		stdout = filterLines(stdout, re)
		stderr = filterLines(stderr, re)
	}

	result := map[string]interface{}{
		"shell_id":  shellID,
		"stdout":    stdout,
		"stderr":    stderr,
		"done":      done,
		"exit_code": exitCode,
		"command":   shell.Command + " " + strings.Join(shell.Args, " "),
		"duration":  time.Since(shell.StartTime).Milliseconds(),
	}

	if errorMsg != "" {
		result["error"] = errorMsg
	}

	return result, nil
}

func filterLines(content string, re *regexp.Regexp) string {
	lines := strings.Split(content, "\n")
	var filtered []string
	for _, line := range lines {
		if re.MatchString(line) {
			filtered = append(filtered, line)
		}
	}
	return strings.Join(filtered, "\n")
}

func (t *BashOutputTool) RequiresConfirmation() bool {
	return false
}

// KillShellTool terminates a running background shell
type KillShellTool struct {
	shellManager *BackgroundShellManager
}

// NewKillShellTool creates a new kill shell tool
func NewKillShellTool(shellManager *BackgroundShellManager) *KillShellTool {
	return &KillShellTool{shellManager: shellManager}
}

func (t *KillShellTool) Name() string {
	return "kill_shell"
}

func (t *KillShellTool) Description() string {
	return "Terminates a running background shell process. Use the shell_id from a background shell_execute command."
}

func (t *KillShellTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"shell_id": {
				Type:        "string",
				Description: "The ID of the background shell to terminate",
			},
		},
		Required: []string{"shell_id"},
	}
}

func (t *KillShellTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	shellID, ok := params["shell_id"].(string)
	if !ok || shellID == "" {
		return nil, fmt.Errorf("shell_id parameter is required")
	}

	err := t.shellManager.KillShell(shellID, userID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success":  true,
		"shell_id": shellID,
		"message":  "Shell terminated",
	}, nil
}

func (t *KillShellTool) RequiresConfirmation() bool {
	return false
}
