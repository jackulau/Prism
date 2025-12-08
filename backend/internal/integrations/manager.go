package integrations

import (
	"log"
	"sync"
)

// EventType represents the type of event
type EventType string

const (
	EventConversationCreated EventType = "conversation.created"
	EventMessageSent         EventType = "message.sent"
	EventChatCompleted       EventType = "chat.completed"
	EventChatStopped         EventType = "chat.stopped"
	EventToolStarted         EventType = "tool.started"
	EventToolCompleted       EventType = "tool.completed"
	EventToolApproved        EventType = "tool.approved"
	EventToolRejected        EventType = "tool.rejected"
	EventError               EventType = "error"
	EventUserLogin           EventType = "user.login"
	EventUserRegister        EventType = "user.register"
)

// Event represents an event to be tracked or notified
type Event struct {
	Type           EventType              `json:"type"`
	UserID         string                 `json:"user_id"`
	ConversationID string                 `json:"conversation_id,omitempty"`
	MessageID      string                 `json:"message_id,omitempty"`
	Data           map[string]interface{} `json:"data,omitempty"`
}

// NotificationProvider interface for sending notifications
type NotificationProvider interface {
	Name() string
	Enabled() bool
	Send(event *Event) error
}

// AnalyticsProvider interface for tracking analytics
type AnalyticsProvider interface {
	Name() string
	Enabled() bool
	Track(event *Event) error
	Flush() error
	Close() error
}

// Manager manages all integrations
type Manager struct {
	notifications []NotificationProvider
	analytics     []AnalyticsProvider
	mu            sync.RWMutex
}

// NewManager creates a new integrations manager
func NewManager() *Manager {
	return &Manager{
		notifications: make([]NotificationProvider, 0),
		analytics:     make([]AnalyticsProvider, 0),
	}
}

// RegisterNotification registers a notification provider
func (m *Manager) RegisterNotification(provider NotificationProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notifications = append(m.notifications, provider)
	log.Printf("Registered notification provider: %s (enabled: %v)", provider.Name(), provider.Enabled())
}

// RegisterAnalytics registers an analytics provider
func (m *Manager) RegisterAnalytics(provider AnalyticsProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.analytics = append(m.analytics, provider)
	log.Printf("Registered analytics provider: %s (enabled: %v)", provider.Name(), provider.Enabled())
}

// Notify sends a notification to all enabled providers
func (m *Manager) Notify(event *Event) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, provider := range m.notifications {
		if provider.Enabled() {
			go func(p NotificationProvider) {
				if err := p.Send(event); err != nil {
					log.Printf("Failed to send notification via %s: %v", p.Name(), err)
				}
			}(provider)
		}
	}
}

// Track sends an event to all enabled analytics providers
func (m *Manager) Track(event *Event) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, provider := range m.analytics {
		if provider.Enabled() {
			go func(p AnalyticsProvider) {
				if err := p.Track(event); err != nil {
					log.Printf("Failed to track event via %s: %v", p.Name(), err)
				}
			}(provider)
		}
	}
}

// TrackAndNotify tracks an event and sends notifications
func (m *Manager) TrackAndNotify(event *Event) {
	m.Track(event)
	m.Notify(event)
}

// TrackMessageSent is a convenience method for tracking message sent events
func (m *Manager) TrackMessageSent(userID, conversationID, messageID string) {
	m.Track(&Event{
		Type:           EventMessageSent,
		UserID:         userID,
		ConversationID: conversationID,
		MessageID:      messageID,
	})
}

// TrackChatCompleted is a convenience method for tracking chat completed events
func (m *Manager) TrackChatCompleted(userID, conversationID, messageID, finishReason string) {
	m.Track(&Event{
		Type:           EventChatCompleted,
		UserID:         userID,
		ConversationID: conversationID,
		MessageID:      messageID,
		Data: map[string]interface{}{
			"finish_reason": finishReason,
		},
	})
}

// TrackError is a convenience method for tracking error events
func (m *Manager) TrackError(userID, conversationID, code, message string) {
	m.TrackAndNotify(&Event{
		Type:           EventError,
		UserID:         userID,
		ConversationID: conversationID,
		Data: map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
}

// Flush flushes all analytics providers
func (m *Manager) Flush() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, provider := range m.analytics {
		if provider.Enabled() {
			if err := provider.Flush(); err != nil {
				log.Printf("Failed to flush analytics provider %s: %v", provider.Name(), err)
			}
		}
	}
	return nil
}

// Close closes all integrations
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, provider := range m.analytics {
		if err := provider.Close(); err != nil {
			log.Printf("Failed to close analytics provider %s: %v", provider.Name(), err)
		}
	}
	return nil
}
