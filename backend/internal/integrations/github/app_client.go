package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AppClient is an authenticated client for GitHub API calls using App installation tokens
type AppClient struct {
	app            *GitHubApp
	installationID int64
	httpClient     *http.Client
}

// NewAppClient creates a new authenticated client for a specific installation
func NewAppClient(app *GitHubApp, installationID int64) *AppClient {
	return &AppClient{
		app:            app,
		installationID: installationID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest makes an authenticated request to the GitHub API
// Automatically refreshes token on 401 Unauthorized
func (c *AppClient) doRequest(method, url string, body io.Reader) (*http.Response, error) {
	return c.doRequestWithRetry(method, url, body, true)
}

func (c *AppClient) doRequestWithRetry(method, url string, body io.Reader, retry bool) (*http.Response, error) {
	token, err := c.app.GetInstallationToken(c.installationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get installation token: %w", err)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// If we get 401, invalidate the cached token and retry once
	if resp.StatusCode == http.StatusUnauthorized && retry {
		resp.Body.Close()
		c.app.InvalidateToken(c.installationID)
		return c.doRequestWithRetry(method, url, body, false)
	}

	return resp, nil
}

// RepositoryListResponse represents the response from listing installation repositories
type RepositoryListResponse struct {
	TotalCount   int          `json:"total_count"`
	Repositories []Repository `json:"repositories"`
}

// ListRepositories lists all repositories accessible to this installation
func (c *AppClient) ListRepositories() ([]Repository, error) {
	var allRepos []Repository
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("https://api.github.com/installation/repositories?per_page=%d&page=%d", perPage, page)
		resp, err := c.doRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
		}

		var repoResp RepositoryListResponse
		if err := json.NewDecoder(resp.Body).Decode(&repoResp); err != nil {
			return nil, fmt.Errorf("failed to parse repositories: %w", err)
		}

		allRepos = append(allRepos, repoResp.Repositories...)

		// Check if there are more pages
		if len(repoResp.Repositories) < perPage {
			break
		}
		page++
	}

	return allRepos, nil
}

// GetRepository retrieves a specific repository
func (c *AppClient) GetRepository(owner, repo string) (*Repository, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("repository %s/%s not found or not accessible", owner, repo)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var repository Repository
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return nil, fmt.Errorf("failed to parse repository: %w", err)
	}

	return &repository, nil
}

// CommitInfo represents a commit from the GitHub API
type CommitInfo struct {
	SHA     string `json:"sha"`
	NodeID  string `json:"node_id"`
	HTMLURL string `json:"html_url"`
	Commit  struct {
		Author struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"author"`
		Committer struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"committer"`
		Message string `json:"message"`
	} `json:"commit"`
	Author    *User `json:"author"`
	Committer *User `json:"committer"`
}

// GetCommits retrieves commits for a repository
func (c *AppClient) GetCommits(owner, repo string, options *CommitListOptions) ([]CommitInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?per_page=%d", owner, repo, 30)

	if options != nil {
		if options.SHA != "" {
			url += "&sha=" + options.SHA
		}
		if options.Path != "" {
			url += "&path=" + options.Path
		}
		if options.Author != "" {
			url += "&author=" + options.Author
		}
		if options.PerPage > 0 {
			url = fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?per_page=%d", owner, repo, options.PerPage)
		}
	}

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var commits []CommitInfo
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, fmt.Errorf("failed to parse commits: %w", err)
	}

	return commits, nil
}

// CommitListOptions specifies options for listing commits
type CommitListOptions struct {
	SHA     string
	Path    string
	Author  string
	PerPage int
}

// PullRequestInfo represents a pull request from the GitHub API
type PullRequestInfo struct {
	ID        int64      `json:"id"`
	NodeID    string     `json:"node_id"`
	Number    int        `json:"number"`
	State     string     `json:"state"` // "open", "closed"
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	HTMLURL   string     `json:"html_url"`
	User      *User      `json:"user"`
	Head      *Branch    `json:"head"`
	Base      *Branch    `json:"base"`
	Merged    bool       `json:"merged"`
	Mergeable *bool      `json:"mergeable"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at"`
	MergedAt  *time.Time `json:"merged_at"`
}

// GetPullRequests retrieves pull requests for a repository
func (c *AppClient) GetPullRequests(owner, repo string, options *PullRequestListOptions) ([]PullRequestInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?per_page=30", owner, repo)

	if options != nil {
		if options.State != "" {
			url += "&state=" + options.State
		}
		if options.Head != "" {
			url += "&head=" + options.Head
		}
		if options.Base != "" {
			url += "&base=" + options.Base
		}
		if options.Sort != "" {
			url += "&sort=" + options.Sort
		}
		if options.Direction != "" {
			url += "&direction=" + options.Direction
		}
		if options.PerPage > 0 {
			url = fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?per_page=%d", owner, repo, options.PerPage)
			if options.State != "" {
				url += "&state=" + options.State
			}
		}
	}

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var prs []PullRequestInfo
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, fmt.Errorf("failed to parse pull requests: %w", err)
	}

	return prs, nil
}

// PullRequestListOptions specifies options for listing pull requests
type PullRequestListOptions struct {
	State     string // "open", "closed", "all"
	Head      string
	Base      string
	Sort      string // "created", "updated", "popularity", "long-running"
	Direction string // "asc", "desc"
	PerPage   int
}

// GetPullRequest retrieves a specific pull request
func (c *AppClient) GetPullRequest(owner, repo string, number int) (*PullRequestInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, number)
	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("pull request %s/%s#%d not found", owner, repo, number)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var pr PullRequestInfo
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, fmt.Errorf("failed to parse pull request: %w", err)
	}

	return &pr, nil
}

// IssueInfo represents an issue from the GitHub API
type IssueInfo struct {
	ID        int64      `json:"id"`
	NodeID    string     `json:"node_id"`
	Number    int        `json:"number"`
	State     string     `json:"state"` // "open", "closed"
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	HTMLURL   string     `json:"html_url"`
	User      *User      `json:"user"`
	Labels    []Label    `json:"labels"`
	Assignees []User     `json:"assignees"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at"`
}

// GetIssues retrieves issues for a repository
func (c *AppClient) GetIssues(owner, repo string, options *IssueListOptions) ([]IssueInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?per_page=30", owner, repo)

	if options != nil {
		if options.State != "" {
			url += "&state=" + options.State
		}
		if options.Labels != "" {
			url += "&labels=" + options.Labels
		}
		if options.Sort != "" {
			url += "&sort=" + options.Sort
		}
		if options.Direction != "" {
			url += "&direction=" + options.Direction
		}
		if options.PerPage > 0 {
			url = fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?per_page=%d", owner, repo, options.PerPage)
			if options.State != "" {
				url += "&state=" + options.State
			}
		}
	}

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var issues []IssueInfo
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, fmt.Errorf("failed to parse issues: %w", err)
	}

	return issues, nil
}

// IssueListOptions specifies options for listing issues
type IssueListOptions struct {
	State     string // "open", "closed", "all"
	Labels    string // comma-separated list of labels
	Sort      string // "created", "updated", "comments"
	Direction string // "asc", "desc"
	PerPage   int
}

// GetBranches retrieves branches for a repository
func (c *AppClient) GetBranches(owner, repo string) ([]BranchInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches?per_page=100", owner, repo)

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var branches []BranchInfo
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, fmt.Errorf("failed to parse branches: %w", err)
	}

	return branches, nil
}

// BranchInfo represents a branch from the GitHub API
type BranchInfo struct {
	Name      string `json:"name"`
	Protected bool   `json:"protected"`
	Commit    struct {
		SHA string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
}

// GetContents retrieves file contents from a repository
func (c *AppClient) GetContents(owner, repo, path string, ref string) (*ContentInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	if ref != "" {
		url += "?ref=" + ref
	}

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("file %s not found in %s/%s", path, owner, repo)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var content ContentInfo
	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	return &content, nil
}

// ContentInfo represents file content from the GitHub API
type ContentInfo struct {
	Type        string `json:"type"` // "file" or "dir"
	Encoding    string `json:"encoding"`
	Size        int    `json:"size"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Content     string `json:"content"` // base64-encoded for files
	SHA         string `json:"sha"`
	HTMLURL     string `json:"html_url"`
	DownloadURL string `json:"download_url"`
}

// InstallationID returns the installation ID for this client
func (c *AppClient) InstallationID() int64 {
	return c.installationID
}
