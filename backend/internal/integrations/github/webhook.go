package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidSignature = errors.New("invalid webhook signature")
	ErrMissingSignature = errors.New("missing webhook signature")
	ErrUnsupportedEvent = errors.New("unsupported webhook event")
)

// WebhookHandler handles incoming GitHub webhooks
type WebhookHandler struct {
	processors map[string]EventProcessor
}

// EventProcessor processes a specific type of webhook event
type EventProcessor interface {
	Process(event interface{}, config *WebhookConfig) error
	EventType() string
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{
		processors: make(map[string]EventProcessor),
	}
}

// RegisterProcessor registers an event processor for a specific event type
func (h *WebhookHandler) RegisterProcessor(processor EventProcessor) {
	h.processors[processor.EventType()] = processor
}

// VerifySignature verifies the GitHub webhook signature
func VerifySignature(payload []byte, signature, secret string) error {
	if signature == "" {
		return ErrMissingSignature
	}

	// GitHub sends the signature as "sha256=<hex>"
	if !strings.HasPrefix(signature, "sha256=") {
		return ErrInvalidSignature
	}

	expectedSig := signature[7:] // Remove "sha256=" prefix

	// Calculate HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	calculatedSig := hex.EncodeToString(mac.Sum(nil))

	// Use constant-time comparison to prevent timing attacks
	if !hmac.Equal([]byte(expectedSig), []byte(calculatedSig)) {
		return ErrInvalidSignature
	}

	return nil
}

// ParseEvent parses a webhook event based on the event type
func ParseEvent(eventType string, payload []byte) (interface{}, error) {
	switch eventType {
	case "issues":
		var event IssueEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			return nil, fmt.Errorf("failed to parse issue event: %w", err)
		}
		return &event, nil

	case "issue_comment":
		var event IssueCommentEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			return nil, fmt.Errorf("failed to parse issue comment event: %w", err)
		}
		return &event, nil

	case "pull_request":
		var event PullRequestEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			return nil, fmt.Errorf("failed to parse pull request event: %w", err)
		}
		return &event, nil

	case "push":
		var event PushEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			return nil, fmt.Errorf("failed to parse push event: %w", err)
		}
		return &event, nil

	default:
		// For unsupported events, return a generic map
		var event map[string]interface{}
		if err := json.Unmarshal(payload, &event); err != nil {
			return nil, fmt.Errorf("failed to parse event: %w", err)
		}
		return event, nil
	}
}

// HandleWebhook processes an incoming webhook
func (h *WebhookHandler) HandleWebhook(eventType string, event interface{}, config *WebhookConfig) error {
	processor, ok := h.processors[eventType]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnsupportedEvent, eventType)
	}

	return processor.Process(event, config)
}

// GetEventAction extracts the action from an event
func GetEventAction(event interface{}) string {
	switch e := event.(type) {
	case *IssueEvent:
		return e.Action
	case *IssueCommentEvent:
		return e.Action
	case *PullRequestEvent:
		return e.Action
	case *PushEvent:
		return "push"
	default:
		if m, ok := event.(map[string]interface{}); ok {
			if action, ok := m["action"].(string); ok {
				return action
			}
		}
		return ""
	}
}

// GetRepoFullName extracts the repository full name from an event
func GetRepoFullName(event interface{}) string {
	switch e := event.(type) {
	case *IssueEvent:
		if e.Repo != nil {
			return e.Repo.FullName
		}
	case *IssueCommentEvent:
		if e.Repo != nil {
			return e.Repo.FullName
		}
	case *PullRequestEvent:
		if e.Repo != nil {
			return e.Repo.FullName
		}
	case *PushEvent:
		if e.Repository != nil {
			return e.Repository.FullName
		}
	default:
		if m, ok := event.(map[string]interface{}); ok {
			if repo, ok := m["repository"].(map[string]interface{}); ok {
				if fullName, ok := repo["full_name"].(string); ok {
					return fullName
				}
			}
		}
	}
	return ""
}
