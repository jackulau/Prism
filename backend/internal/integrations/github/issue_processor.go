package github

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

// IssueProcessor processes GitHub issue events
type IssueProcessor struct {
	codeRunner CodeRunner
}

// CodeRunner interface for executing code in response to events
type CodeRunner interface {
	Run(request *CodeRunRequest) (*CodeExecutionResult, error)
}

// CodeRunRequest represents a request to run code
type CodeRunRequest struct {
	Command     string            `json:"command"`
	Environment string            `json:"environment"` // "node", "python", "shell"
	WorkDir     string            `json:"work_dir"`
	EnvVars     map[string]string `json:"env_vars"`
	Timeout     int               `json:"timeout_seconds"`
	Context     *EventContext     `json:"context"`
}

// EventContext provides context about the triggering event
type EventContext struct {
	EventType    string `json:"event_type"`
	Action       string `json:"action"`
	RepoFullName string `json:"repo_full_name"`
	RepoURL      string `json:"repo_url"`
	IssueNumber  int    `json:"issue_number,omitempty"`
	IssueTitle   string `json:"issue_title,omitempty"`
	IssueBody    string `json:"issue_body,omitempty"`
	IssueURL     string `json:"issue_url,omitempty"`
	SenderLogin  string `json:"sender_login"`
}

// NewIssueProcessor creates a new issue processor
func NewIssueProcessor(runner CodeRunner) *IssueProcessor {
	return &IssueProcessor{
		codeRunner: runner,
	}
}

// EventType returns the event type this processor handles
func (p *IssueProcessor) EventType() string {
	return "issues"
}

// Process processes an issue event
func (p *IssueProcessor) Process(event interface{}, config *WebhookConfig) error {
	issueEvent, ok := event.(*IssueEvent)
	if !ok {
		return fmt.Errorf("expected IssueEvent, got %T", event)
	}

	log.Printf("Processing issue event: %s for issue #%d in %s",
		issueEvent.Action, issueEvent.Issue.Number, config.RepoFullName)

	// Check if auto-run is enabled
	if !config.AutoRunEnabled {
		log.Printf("Auto-run disabled for webhook %s", config.ID)
		return nil
	}

	// Find matching triggers
	triggers := p.findMatchingTriggers(issueEvent, config.AutoRunTriggers)
	if len(triggers) == 0 {
		log.Printf("No matching triggers for action %s", issueEvent.Action)
		return nil
	}

	// Build event context
	ctx := p.buildEventContext(issueEvent, config)

	// Execute each matching trigger
	for _, trigger := range triggers {
		if err := p.executeTrigger(trigger, ctx); err != nil {
			log.Printf("Failed to execute trigger: %v", err)
			// Continue with other triggers
		}
	}

	return nil
}

// findMatchingTriggers finds triggers that match the event
func (p *IssueProcessor) findMatchingTriggers(event *IssueEvent, triggers []AutoRunTrigger) []AutoRunTrigger {
	var matching []AutoRunTrigger

	for _, trigger := range triggers {
		// Check event type
		if trigger.Event != "issues" {
			continue
		}

		// Check action
		if trigger.Action != "" && trigger.Action != event.Action {
			continue
		}

		// Check labels if specified
		if len(trigger.Labels) > 0 {
			if !p.hasMatchingLabels(event.Issue.Labels, trigger.Labels) {
				continue
			}
		}

		matching = append(matching, trigger)
	}

	return matching
}

// hasMatchingLabels checks if the issue has any of the required labels
func (p *IssueProcessor) hasMatchingLabels(issueLabels []Label, requiredLabels []string) bool {
	labelSet := make(map[string]bool)
	for _, label := range issueLabels {
		labelSet[strings.ToLower(label.Name)] = true
	}

	for _, required := range requiredLabels {
		if labelSet[strings.ToLower(required)] {
			return true
		}
	}

	return false
}

// buildEventContext builds the context for code execution
func (p *IssueProcessor) buildEventContext(event *IssueEvent, config *WebhookConfig) *EventContext {
	ctx := &EventContext{
		EventType:    "issues",
		Action:       event.Action,
		RepoFullName: config.RepoFullName,
	}

	if event.Repo != nil {
		ctx.RepoURL = event.Repo.HTMLURL
	}

	if event.Issue != nil {
		ctx.IssueNumber = event.Issue.Number
		ctx.IssueTitle = event.Issue.Title
		ctx.IssueBody = event.Issue.Body
		ctx.IssueURL = event.Issue.HTMLURL
	}

	if event.Sender != nil {
		ctx.SenderLogin = event.Sender.Login
	}

	return ctx
}

// executeTrigger executes a single trigger
func (p *IssueProcessor) executeTrigger(trigger AutoRunTrigger, ctx *EventContext) error {
	// Expand variables in command
	command := p.expandVariables(trigger.Command, ctx)

	// Expand variables in env vars
	envVars := make(map[string]string)
	for k, v := range trigger.EnvVars {
		envVars[k] = p.expandVariables(v, ctx)
	}

	// Add context as environment variables
	envVars["GITHUB_EVENT"] = ctx.EventType
	envVars["GITHUB_ACTION"] = ctx.Action
	envVars["GITHUB_REPOSITORY"] = ctx.RepoFullName
	envVars["GITHUB_ISSUE_NUMBER"] = fmt.Sprintf("%d", ctx.IssueNumber)
	envVars["GITHUB_ISSUE_TITLE"] = ctx.IssueTitle
	envVars["GITHUB_SENDER"] = ctx.SenderLogin

	request := &CodeRunRequest{
		Command:     command,
		Environment: trigger.Environment,
		WorkDir:     trigger.WorkDir,
		EnvVars:     envVars,
		Timeout:     300, // 5 minute default
		Context:     ctx,
	}

	log.Printf("Executing trigger: %s (environment: %s)", command, trigger.Environment)

	result, err := p.codeRunner.Run(request)
	if err != nil {
		return fmt.Errorf("code execution failed: %w", err)
	}

	log.Printf("Execution completed: exit_code=%d, duration=%dms", result.ExitCode, result.Duration)
	return nil
}

// expandVariables expands template variables in the string
func (p *IssueProcessor) expandVariables(s string, ctx *EventContext) string {
	replacements := map[string]string{
		"{{event}}":        ctx.EventType,
		"{{action}}":       ctx.Action,
		"{{repo}}":         ctx.RepoFullName,
		"{{repo_url}}":     ctx.RepoURL,
		"{{issue_number}}": fmt.Sprintf("%d", ctx.IssueNumber),
		"{{issue_title}}":  ctx.IssueTitle,
		"{{issue_body}}":   ctx.IssueBody,
		"{{issue_url}}":    ctx.IssueURL,
		"{{sender}}":       ctx.SenderLogin,
	}

	result := s
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// ExtractCodeBlocks extracts code blocks from issue body
func ExtractCodeBlocks(body string) []CodeBlock {
	var blocks []CodeBlock

	// Match fenced code blocks with optional language
	re := regexp.MustCompile("```(\\w*)\\n([\\s\\S]*?)```")
	matches := re.FindAllStringSubmatch(body, -1)

	for _, match := range matches {
		lang := match[1]
		code := strings.TrimSpace(match[2])
		if code != "" {
			blocks = append(blocks, CodeBlock{
				Language: lang,
				Code:     code,
			})
		}
	}

	return blocks
}

// CodeBlock represents a code block extracted from text
type CodeBlock struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

// IssueCommentProcessor processes GitHub issue comment events
type IssueCommentProcessor struct {
	codeRunner CodeRunner
}

// NewIssueCommentProcessor creates a new issue comment processor
func NewIssueCommentProcessor(runner CodeRunner) *IssueCommentProcessor {
	return &IssueCommentProcessor{
		codeRunner: runner,
	}
}

// EventType returns the event type this processor handles
func (p *IssueCommentProcessor) EventType() string {
	return "issue_comment"
}

// Process processes an issue comment event
func (p *IssueCommentProcessor) Process(event interface{}, config *WebhookConfig) error {
	commentEvent, ok := event.(*IssueCommentEvent)
	if !ok {
		return fmt.Errorf("expected IssueCommentEvent, got %T", event)
	}

	log.Printf("Processing issue comment event: %s on issue #%d in %s",
		commentEvent.Action, commentEvent.Issue.Number, config.RepoFullName)

	// Check if auto-run is enabled
	if !config.AutoRunEnabled {
		return nil
	}

	// Find matching triggers for comments
	for _, trigger := range config.AutoRunTriggers {
		if trigger.Event != "issue_comment" {
			continue
		}

		if trigger.Action != "" && trigger.Action != commentEvent.Action {
			continue
		}

		// Build context
		ctx := &EventContext{
			EventType:    "issue_comment",
			Action:       commentEvent.Action,
			RepoFullName: config.RepoFullName,
		}

		if commentEvent.Repo != nil {
			ctx.RepoURL = commentEvent.Repo.HTMLURL
		}

		if commentEvent.Issue != nil {
			ctx.IssueNumber = commentEvent.Issue.Number
			ctx.IssueTitle = commentEvent.Issue.Title
			ctx.IssueURL = commentEvent.Issue.HTMLURL
		}

		if commentEvent.Comment != nil {
			ctx.IssueBody = commentEvent.Comment.Body
		}

		if commentEvent.Sender != nil {
			ctx.SenderLogin = commentEvent.Sender.Login
		}

		// Execute trigger
		command := expandVariables(trigger.Command, ctx)
		envVars := make(map[string]string)
		for k, v := range trigger.EnvVars {
			envVars[k] = expandVariables(v, ctx)
		}

		envVars["GITHUB_EVENT"] = ctx.EventType
		envVars["GITHUB_ACTION"] = ctx.Action
		envVars["GITHUB_REPOSITORY"] = ctx.RepoFullName
		envVars["GITHUB_ISSUE_NUMBER"] = fmt.Sprintf("%d", ctx.IssueNumber)
		envVars["GITHUB_SENDER"] = ctx.SenderLogin

		request := &CodeRunRequest{
			Command:     command,
			Environment: trigger.Environment,
			WorkDir:     trigger.WorkDir,
			EnvVars:     envVars,
			Timeout:     300,
			Context:     ctx,
		}

		if _, err := p.codeRunner.Run(request); err != nil {
			log.Printf("Failed to execute trigger: %v", err)
		}
	}

	return nil
}

// expandVariables is a package-level function for expanding template variables
func expandVariables(s string, ctx *EventContext) string {
	replacements := map[string]string{
		"{{event}}":        ctx.EventType,
		"{{action}}":       ctx.Action,
		"{{repo}}":         ctx.RepoFullName,
		"{{repo_url}}":     ctx.RepoURL,
		"{{issue_number}}": fmt.Sprintf("%d", ctx.IssueNumber),
		"{{issue_title}}":  ctx.IssueTitle,
		"{{issue_body}}":   ctx.IssueBody,
		"{{issue_url}}":    ctx.IssueURL,
		"{{sender}}":       ctx.SenderLogin,
	}

	result := s
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}
