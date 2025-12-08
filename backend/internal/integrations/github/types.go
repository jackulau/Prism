package github

import "time"

// WebhookEvent represents the common fields in all GitHub webhook events
type WebhookEvent struct {
	Action string      `json:"action"`
	Sender *User       `json:"sender"`
	Repo   *Repository `json:"repository"`
}

// User represents a GitHub user
type User struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
	Type      string `json:"type"`
}

// Repository represents a GitHub repository
type Repository struct {
	ID          int64  `json:"id"`
	NodeID      string `json:"node_id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Private     bool   `json:"private"`
	HTMLURL     string `json:"html_url"`
	Description string `json:"description"`
	CloneURL    string `json:"clone_url"`
	SSHURL      string `json:"ssh_url"`
	Owner       *User  `json:"owner"`
}

// Issue represents a GitHub issue
type Issue struct {
	ID        int64      `json:"id"`
	NodeID    string     `json:"node_id"`
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	State     string     `json:"state"`
	HTMLURL   string     `json:"html_url"`
	User      *User      `json:"user"`
	Labels    []Label    `json:"labels"`
	Assignees []User     `json:"assignees"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at"`
}

// Label represents a GitHub label
type Label struct {
	ID          int64  `json:"id"`
	NodeID      string `json:"node_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

// IssueEvent represents a GitHub issue webhook event
type IssueEvent struct {
	WebhookEvent
	Issue   *Issue `json:"issue"`
	Changes *struct {
		Title *struct {
			From string `json:"from"`
		} `json:"title"`
		Body *struct {
			From string `json:"from"`
		} `json:"body"`
	} `json:"changes"`
}

// IssueCommentEvent represents a GitHub issue comment webhook event
type IssueCommentEvent struct {
	WebhookEvent
	Issue   *Issue   `json:"issue"`
	Comment *Comment `json:"comment"`
}

// Comment represents a GitHub comment
type Comment struct {
	ID        int64     `json:"id"`
	NodeID    string    `json:"node_id"`
	Body      string    `json:"body"`
	User      *User     `json:"user"`
	HTMLURL   string    `json:"html_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PullRequestEvent represents a GitHub pull request webhook event
type PullRequestEvent struct {
	WebhookEvent
	Number      int          `json:"number"`
	PullRequest *PullRequest `json:"pull_request"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	ID        int64      `json:"id"`
	NodeID    string     `json:"node_id"`
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	State     string     `json:"state"`
	HTMLURL   string     `json:"html_url"`
	User      *User      `json:"user"`
	Head      *Branch    `json:"head"`
	Base      *Branch    `json:"base"`
	Merged    bool       `json:"merged"`
	MergedBy  *User      `json:"merged_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at"`
	MergedAt  *time.Time `json:"merged_at"`
}

// Branch represents a GitHub branch reference
type Branch struct {
	Ref  string      `json:"ref"`
	SHA  string      `json:"sha"`
	Repo *Repository `json:"repo"`
}

// PushEvent represents a GitHub push webhook event
type PushEvent struct {
	Ref        string      `json:"ref"`
	Before     string      `json:"before"`
	After      string      `json:"after"`
	Repository *Repository `json:"repository"`
	Pusher     *struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"pusher"`
	Sender  *User    `json:"sender"`
	Commits []Commit `json:"commits"`
}

// Commit represents a GitHub commit
type Commit struct {
	ID        string   `json:"id"`
	Message   string   `json:"message"`
	Timestamp string   `json:"timestamp"`
	URL       string   `json:"url"`
	Author    *Author  `json:"author"`
	Added     []string `json:"added"`
	Removed   []string `json:"removed"`
	Modified  []string `json:"modified"`
}

// Author represents a commit author
type Author struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// WebhookConfig represents the configuration for a GitHub webhook
type WebhookConfig struct {
	ID              string            `json:"id"`
	UserID          string            `json:"user_id"`
	RepoFullName    string            `json:"repo_full_name"`
	WebhookSecret   string            `json:"webhook_secret"`
	Events          []string          `json:"events"`
	AutoRunEnabled  bool              `json:"auto_run_enabled"`
	AutoRunTriggers []AutoRunTrigger  `json:"auto_run_triggers"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// AutoRunTrigger defines when to automatically run code
type AutoRunTrigger struct {
	Event       string            `json:"event"`       // e.g., "issues", "issue_comment"
	Action      string            `json:"action"`      // e.g., "opened", "created"
	Labels      []string          `json:"labels"`      // Filter by labels (optional)
	Command     string            `json:"command"`     // Command to run
	Environment string            `json:"environment"` // e.g., "node", "python", "shell"
	WorkDir     string            `json:"work_dir"`    // Working directory
	EnvVars     map[string]string `json:"env_vars"`    // Environment variables
}

// WebhookDelivery represents a record of a webhook delivery
type WebhookDelivery struct {
	ID            string                 `json:"id"`
	WebhookID     string                 `json:"webhook_id"`
	Event         string                 `json:"event"`
	Action        string                 `json:"action"`
	Payload       map[string]interface{} `json:"payload"`
	Status        string                 `json:"status"` // "pending", "processing", "completed", "failed"
	ErrorMessage  string                 `json:"error_message,omitempty"`
	ProcessedAt   *time.Time             `json:"processed_at"`
	CreatedAt     time.Time              `json:"created_at"`
}

// CodeExecutionResult represents the result of an automatic code execution
type CodeExecutionResult struct {
	ID           string    `json:"id"`
	DeliveryID   string    `json:"delivery_id"`
	Command      string    `json:"command"`
	Environment  string    `json:"environment"`
	ExitCode     int       `json:"exit_code"`
	Stdout       string    `json:"stdout"`
	Stderr       string    `json:"stderr"`
	Duration     int64     `json:"duration_ms"`
	StartedAt    time.Time `json:"started_at"`
	CompletedAt  time.Time `json:"completed_at"`
}
