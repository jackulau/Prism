package builtin

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/security"
)

const (
	githubAPITimeout   = 30 * time.Second
	defaultCommitLimit = 10
	maxCommitLimit     = 100
)

// GitHubCommitsTool fetches commit history from GitHub repositories
type GitHubCommitsTool struct {
	userRepo      *repository.UserRepository
	encryptionSvc *security.EncryptionService
	httpClient    *http.Client
}

// NewGitHubCommitsTool creates a new GitHub commits tool
func NewGitHubCommitsTool(userRepo *repository.UserRepository, encryptionSvc *security.EncryptionService) *GitHubCommitsTool {
	return &GitHubCommitsTool{
		userRepo:      userRepo,
		encryptionSvc: encryptionSvc,
		httpClient: &http.Client{
			Timeout: githubAPITimeout,
		},
	}
}

func (t *GitHubCommitsTool) Name() string {
	return "github_commits"
}

func (t *GitHubCommitsTool) Description() string {
	return `Fetches commit history from a GitHub repository using the user's connected GitHub account. Returns commit SHA, message, author, date, and files changed. Use this to understand project history, recent changes, or for code review purposes.`
}

func (t *GitHubCommitsTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"repo": {
				Type:        "string",
				Description: "Repository in owner/name format (e.g., 'anthropics/claude-code')",
			},
			"branch": {
				Type:        "string",
				Description: "Branch name to fetch commits from (optional, defaults to repository's default branch)",
			},
			"limit": {
				Type:        "integer",
				Description: "Maximum number of commits to fetch (optional, defaults to 10, max 100)",
			},
		},
		Required: []string{"repo"},
	}
}

func (t *GitHubCommitsTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Get user ID from context
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	// Parse parameters
	repo, ok := params["repo"].(string)
	if !ok || repo == "" {
		return nil, fmt.Errorf("repo parameter is required (format: owner/name)")
	}

	// Validate repo format
	parts := strings.Split(repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("invalid repo format: expected 'owner/name', got '%s'", repo)
	}
	owner, repoName := parts[0], parts[1]

	branch, _ := params["branch"].(string)

	limit := defaultCommitLimit
	if limitVal, ok := params["limit"].(float64); ok {
		limit = int(limitVal)
		if limit < 1 {
			limit = 1
		}
		if limit > maxCommitLimit {
			limit = maxCommitLimit
		}
	}

	// Get user and their GitHub token
	user, err := t.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.GitHubToken == "" {
		return nil, fmt.Errorf("GitHub not connected. Please connect your GitHub account in Settings to use this tool.")
	}

	// Decrypt the GitHub token
	token, err := t.decryptGitHubToken(user.GitHubToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt GitHub token: %w", err)
	}

	// Fetch commits from GitHub API
	commits, err := t.fetchCommits(ctx, token, owner, repoName, branch, limit)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"repo":    repo,
		"branch":  branch,
		"commits": commits,
		"count":   len(commits),
	}, nil
}

// decryptGitHubToken decrypts a GitHub token stored as "nonce_hex:ciphertext_hex"
func (t *GitHubCommitsTool) decryptGitHubToken(encryptedToken string) (string, error) {
	colonIdx := strings.Index(encryptedToken, ":")
	if colonIdx == -1 {
		return "", fmt.Errorf("invalid encrypted token format")
	}

	nonceHex := encryptedToken[:colonIdx]
	ciphertextHex := encryptedToken[colonIdx+1:]

	nonce, err := hex.DecodeString(nonceHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode nonce: %w", err)
	}
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	plaintext, err := t.encryptionSvc.Decrypt(ciphertext, nonce)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// CommitInfo represents the information about a single commit
type CommitInfo struct {
	SHA      string   `json:"sha"`
	Message  string   `json:"message"`
	Author   string   `json:"author"`
	Email    string   `json:"email"`
	Date     string   `json:"date"`
	URL      string   `json:"url"`
	Added    []string `json:"added,omitempty"`
	Removed  []string `json:"removed,omitempty"`
	Modified []string `json:"modified,omitempty"`
}

// GitHubCommitResponse represents the GitHub API response for commits
type GitHubCommitResponse struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Date  string `json:"date"`
		} `json:"author"`
	} `json:"commit"`
	HTMLURL string `json:"html_url"`
	Files   []struct {
		Filename string `json:"filename"`
		Status   string `json:"status"` // "added", "removed", "modified", "renamed"
	} `json:"files,omitempty"`
}

func (t *GitHubCommitsTool) fetchCommits(ctx context.Context, token, owner, repo, branch string, limit int) ([]CommitInfo, error) {
	// Build URL
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?per_page=%d", owner, repo, limit)
	if branch != "" {
		url += "&sha=" + branch
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "Prism-GitHub-Commits-Tool")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch commits: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle error responses
	switch resp.StatusCode {
	case http.StatusOK:
		// Continue processing
	case http.StatusNotFound:
		return nil, fmt.Errorf("repository '%s/%s' not found or you don't have access to it", owner, repo)
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("GitHub authentication failed. Please reconnect your GitHub account in Settings.")
	case http.StatusForbidden:
		// Check if it's a rate limit issue
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			resetTime := resp.Header.Get("X-RateLimit-Reset")
			return nil, fmt.Errorf("GitHub API rate limit exceeded. Resets at: %s", resetTime)
		}
		return nil, fmt.Errorf("access forbidden to repository '%s/%s'", owner, repo)
	default:
		// Try to parse error message from response
		var errResp struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
			return nil, fmt.Errorf("GitHub API error (%d): %s", resp.StatusCode, errResp.Message)
		}
		return nil, fmt.Errorf("GitHub API error (status %d)", resp.StatusCode)
	}

	var ghCommits []GitHubCommitResponse
	if err := json.Unmarshal(body, &ghCommits); err != nil {
		return nil, fmt.Errorf("failed to parse commits response: %w", err)
	}

	// Convert to our format
	commits := make([]CommitInfo, 0, len(ghCommits))
	for _, ghc := range ghCommits {
		commit := CommitInfo{
			SHA:     ghc.SHA,
			Message: ghc.Commit.Message,
			Author:  ghc.Commit.Author.Name,
			Email:   ghc.Commit.Author.Email,
			Date:    ghc.Commit.Author.Date,
			URL:     ghc.HTMLURL,
		}

		// Categorize files by status
		for _, f := range ghc.Files {
			switch f.Status {
			case "added":
				commit.Added = append(commit.Added, f.Filename)
			case "removed":
				commit.Removed = append(commit.Removed, f.Filename)
			case "modified", "renamed":
				commit.Modified = append(commit.Modified, f.Filename)
			}
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

func (t *GitHubCommitsTool) RequiresConfirmation() bool {
	return false
}
