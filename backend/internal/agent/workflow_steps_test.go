package agent

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/jacklau/prism/internal/llm"
)

func TestIsValidRepoFormat(t *testing.T) {
	tests := []struct {
		name     string
		repoName string
		want     bool
	}{
		{"valid format", "owner/repo", true},
		{"valid with dots", "my.org/my.repo", true},
		{"valid with hyphens", "my-org/my-repo", true},
		{"valid with underscores", "my_org/my_repo", true},
		{"valid with numbers", "owner123/repo456", true},
		{"empty string", "", false},
		{"no slash", "ownerrepo", false},
		{"multiple slashes", "owner/repo/extra", false},
		{"empty owner", "/repo", false},
		{"empty repo", "owner/", false},
		{"invalid chars in owner", "owner@/repo", false},
		{"invalid chars in repo", "owner/repo!", false},
		{"spaces", "owner name/repo name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidRepoFormat(tt.repoName)
			if got != tt.want {
				t.Errorf("isValidRepoFormat(%q) = %v, want %v", tt.repoName, got, tt.want)
			}
		})
	}
}

func TestGenerateBranchName(t *testing.T) {
	tests := []struct {
		name     string
		agentID  string
		prompt   string
		expected string
	}{
		{
			name:     "simple prompt",
			agentID:  "abc12345-6789-abcd-ef01-234567890123",
			prompt:   "Fix the bug",
			expected: "prism/abc12345-fix-the-bug",
		},
		{
			name:     "long prompt gets truncated",
			agentID:  "abc12345-6789-abcd-ef01-234567890123",
			prompt:   "This is a very long prompt that should be truncated to a reasonable length",
			expected: "prism/abc12345-this-is-a-very-long-prompt-tha",
		},
		{
			name:     "special characters removed",
			agentID:  "abc12345-6789-abcd-ef01-234567890123",
			prompt:   "Fix bug #123 (important!)",
			expected: "prism/abc12345-fix-bug-123-important",
		},
		{
			name:     "uppercase converted to lowercase",
			agentID:  "abc12345-6789-abcd-ef01-234567890123",
			prompt:   "FIX THE BUG",
			expected: "prism/abc12345-fix-the-bug",
		},
		{
			name:     "short agent ID",
			agentID:  "abc",
			prompt:   "Test",
			expected: "prism/abc-test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateBranchName(tt.agentID, tt.prompt)
			if got != tt.expected {
				t.Errorf("generateBranchName(%q, %q) = %q, want %q", tt.agentID, tt.prompt, got, tt.expected)
			}
		})
	}
}

func TestGenerateCommitMessage(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		expected string
	}{
		{
			name:     "short prompt",
			prompt:   "fix the bug",
			expected: "Fix the bug",
		},
		{
			name:     "already capitalized",
			prompt:   "Fix the bug",
			expected: "Fix the bug",
		},
		{
			name:     "long prompt truncated",
			prompt:   "This is a very long commit message that exceeds the 50 character limit",
			expected: "This is a very long commit message that exceeds...",
		},
		{
			name:     "exactly 50 chars",
			prompt:   "12345678901234567890123456789012345678901234567890",
			expected: "12345678901234567890123456789012345678901234567890",
		},
		{
			name:     "whitespace trimmed",
			prompt:   "  Fix the bug  ",
			expected: "Fix the bug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateCommitMessage(tt.prompt)
			if got != tt.expected {
				t.Errorf("generateCommitMessage(%q) = %q, want %q", tt.prompt, got, tt.expected)
			}
		})
	}
}

func TestValidateAndGetToken(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		repoName string
		wantErr  bool
	}{
		{"valid repo", "owner/repo", false},
		{"empty repo", "", true},
		{"invalid format", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateAndGetToken(ctx, tt.repoName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndGetToken(%q) error = %v, wantErr %v", tt.repoName, err, tt.wantErr)
			}
		})
	}
}

func TestPrepareSandbox(t *testing.T) {
	ctx := context.Background()

	t.Run("nil sandbox service returns error", func(t *testing.T) {
		wfCtx := &WorkflowContext{}
		_, _, err := PrepareSandbox(ctx, wfCtx, "agent-123", "owner/repo", "")
		if err == nil {
			t.Error("expected error for nil sandbox service")
		}
	})
}

func TestLoadPreviousMessages(t *testing.T) {
	ctx := context.Background()

	t.Run("nil message repo returns error", func(t *testing.T) {
		wfCtx := &WorkflowContext{}
		_, err := LoadPreviousMessages(ctx, wfCtx, "conv-123")
		if err == nil {
			t.Error("expected error for nil message repo")
		}
	})

	t.Run("empty conversation ID returns empty slice", func(t *testing.T) {
		// When conversation ID is empty, the function should return early
		// with an empty slice before checking MessageRepo
		wfCtx := &WorkflowContext{
			MessageRepo: nil,
		}
		messages, err := LoadPreviousMessages(ctx, wfCtx, "")
		if err != nil {
			t.Errorf("expected no error for empty conversation ID, got %v", err)
		}
		if len(messages) != 0 {
			t.Errorf("expected empty message slice, got %d messages", len(messages))
		}
	})
}

func TestSaveAgentResponse(t *testing.T) {
	ctx := context.Background()

	t.Run("nil message repo returns error", func(t *testing.T) {
		wfCtx := &WorkflowContext{}
		err := SaveAgentResponse(ctx, wfCtx, "conv-123", "assistant", "Hello", nil, nil)
		if err == nil {
			t.Error("expected error for nil message repo")
		}
	})

	t.Run("empty conversation ID returns error", func(t *testing.T) {
		wfCtx := &WorkflowContext{
			// Even if we had a message repo, empty conversation ID should error
		}
		err := SaveAgentResponse(ctx, wfCtx, "", "assistant", "Hello", nil, nil)
		if err == nil {
			t.Error("expected error for empty conversation ID")
		}
	})
}

func TestLoadAgent(t *testing.T) {
	llmManager := llm.NewManager()

	t.Run("agent not found", func(t *testing.T) {
		manager := NewManager(llmManager, DefaultManagerConfig())
		_, err := LoadAgent(manager, "nonexistent-agent")
		if err != ErrAgentNotFound {
			t.Errorf("expected ErrAgentNotFound, got %v", err)
		}
	})
}

func TestCleanupSandbox(t *testing.T) {
	ctx := context.Background()

	t.Run("nil sandbox service returns nil", func(t *testing.T) {
		err := CleanupSandbox(ctx, nil, "build-123")
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("empty build ID returns nil", func(t *testing.T) {
		err := CleanupSandbox(ctx, nil, "")
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})
}

// Integration tests for git operations
func TestCreateBranchIfNeeded_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create a temporary git repository
	tempDir, err := os.MkdirTemp("", "workflow-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo
	initCmd := exec.Command("git", "init")
	initCmd.Dir = tempDir
	if err := initCmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git user for commits
	configEmailCmd := exec.Command("git", "config", "user.email", "test@test.com")
	configEmailCmd.Dir = tempDir
	configEmailCmd.Run()

	configNameCmd := exec.Command("git", "config", "user.name", "Test User")
	configNameCmd.Dir = tempDir
	configNameCmd.Run()

	// Create initial commit
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = tempDir
	addCmd.Run()

	commitCmd := exec.Command("git", "commit", "-m", "Initial commit")
	commitCmd.Dir = tempDir
	commitCmd.Run()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("creates new branch", func(t *testing.T) {
		branchName, err := CreateBranchIfNeeded(ctx, nil, tempDir, "agent123", "fix bug")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := "prism/agent123-fix-bug"
		if branchName != expected {
			t.Errorf("expected branch name %q, got %q", expected, branchName)
		}

		// Verify branch was created
		listCmd := exec.Command("git", "branch", "--list", branchName)
		listCmd.Dir = tempDir
		output, err := listCmd.Output()
		if err != nil {
			t.Fatalf("failed to list branches: %v", err)
		}
		if len(output) == 0 {
			t.Error("branch was not created")
		}
	})

	t.Run("checks out existing branch", func(t *testing.T) {
		// First go back to master/main
		exec.Command("git", "checkout", "master").Run()

		// Try to create the same branch again
		branchName, err := CreateBranchIfNeeded(ctx, nil, tempDir, "agent123", "fix bug")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := "prism/agent123-fix-bug"
		if branchName != expected {
			t.Errorf("expected branch name %q, got %q", expected, branchName)
		}
	})
}

func TestCommitChangesIfNeeded_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create a temporary git repository
	tempDir, err := os.MkdirTemp("", "workflow-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo
	initCmd := exec.Command("git", "init")
	initCmd.Dir = tempDir
	if err := initCmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git user for commits
	configEmailCmd := exec.Command("git", "config", "user.email", "test@test.com")
	configEmailCmd.Dir = tempDir
	configEmailCmd.Run()

	configNameCmd := exec.Command("git", "config", "user.name", "Test User")
	configNameCmd.Dir = tempDir
	configNameCmd.Run()

	// Create initial commit
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = tempDir
	addCmd.Run()

	commitCmd := exec.Command("git", "commit", "-m", "Initial commit")
	commitCmd.Dir = tempDir
	commitCmd.Run()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("no changes returns ErrNoChangesToCommit", func(t *testing.T) {
		_, err := CommitChangesIfNeeded(ctx, tempDir, "agent123", "test commit")
		if err != ErrNoChangesToCommit {
			t.Errorf("expected ErrNoChangesToCommit, got %v", err)
		}
	})

	t.Run("commits changes successfully", func(t *testing.T) {
		// Make a change
		if err := os.WriteFile(testFile, []byte("updated content"), 0644); err != nil {
			t.Fatalf("failed to update test file: %v", err)
		}

		sha, err := CommitChangesIfNeeded(ctx, tempDir, "agent123", "update test file")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if sha == "" {
			t.Error("expected non-empty commit SHA")
		}

		// Verify commit was created
		logCmd := exec.Command("git", "log", "-1", "--pretty=format:%s")
		logCmd.Dir = tempDir
		output, err := logCmd.Output()
		if err != nil {
			t.Fatalf("failed to get log: %v", err)
		}

		if string(output) != "Update test file" {
			t.Errorf("unexpected commit message: %s", output)
		}
	})
}

func TestNewWorkflowContext(t *testing.T) {
	wfCtx := NewWorkflowContext(nil, nil, nil, nil)
	if wfCtx == nil {
		t.Error("expected non-nil workflow context")
	}
}
