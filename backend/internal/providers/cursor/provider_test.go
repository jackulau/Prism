package cursor

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jacklau/prism/internal/providers"
)

func TestProvider_Name(t *testing.T) {
	provider := NewProvider("test-key")
	if provider.Name() != "cursor" {
		t.Errorf("expected name 'cursor', got '%s'", provider.Name())
	}
}

func TestProvider_SupportsStreaming(t *testing.T) {
	provider := NewProvider("test-key")
	if !provider.SupportsStreaming() {
		t.Error("expected SupportsStreaming to return true")
	}
}

func TestProvider_HasConfiguredKey(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected bool
	}{
		{
			name:     "with key",
			apiKey:   "test-key",
			expected: true,
		},
		{
			name:     "without key",
			apiKey:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewProvider(tt.apiKey)
			if provider.HasConfiguredKey() != tt.expected {
				t.Errorf("expected HasConfiguredKey to return %v", tt.expected)
			}
		})
	}
}

func TestProvider_CreateAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/agents" {
			t.Errorf("expected path /agents, got %s", r.URL.Path)
		}

		// Verify auth header
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Basic ") {
			t.Errorf("expected Basic auth, got %s", auth)
		}

		// Parse request body
		var req CursorCreateAgentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if req.Prompt != "test prompt" {
			t.Errorf("expected prompt 'test prompt', got '%s'", req.Prompt)
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CursorAgentResponse{
			ID:        "agent-123",
			Status:    "running",
			CreatedAt: time.Now().Format(time.RFC3339),
			Model:     "gpt-4",
		})
	}))
	defer server.Close()

	provider := NewProviderWithBaseURL("test-key", server.URL)

	agent, err := provider.CreateAgent(context.Background(), providers.CreateAgentRequest{
		Prompt: "test prompt",
	})

	if err != nil {
		t.Fatalf("CreateAgent failed: %v", err)
	}

	if agent.ID != "agent-123" {
		t.Errorf("expected agent ID 'agent-123', got '%s'", agent.ID)
	}
	if agent.Status != providers.AgentStatusRunning {
		t.Errorf("expected status Running, got %s", agent.Status)
	}
	if agent.Provider != "cursor" {
		t.Errorf("expected provider 'cursor', got '%s'", agent.Provider)
	}
}

func TestProvider_GetAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/agents/agent-123" {
			t.Errorf("expected path /agents/agent-123, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CursorAgentResponse{
			ID:        "agent-123",
			Status:    "completed",
			CreatedAt: time.Now().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	provider := NewProviderWithBaseURL("test-key", server.URL)

	agent, err := provider.GetAgent(context.Background(), "agent-123")

	if err != nil {
		t.Fatalf("GetAgent failed: %v", err)
	}

	if agent.ID != "agent-123" {
		t.Errorf("expected agent ID 'agent-123', got '%s'", agent.ID)
	}
	if agent.Status != providers.AgentStatusCompleted {
		t.Errorf("expected status Completed, got %s", agent.Status)
	}
}

func TestProvider_GetMessages(t *testing.T) {
	now := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/agents/agent-123/messages" {
			t.Errorf("expected path /agents/agent-123/messages, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CursorMessagesResponse{
			Messages: []CursorMessageResponse{
				{ID: "msg-1", Role: "user", Content: "Hello", CreatedAt: now},
				{ID: "msg-2", Role: "assistant", Content: "Hi there!", CreatedAt: now.Add(time.Second)},
			},
		})
	}))
	defer server.Close()

	provider := NewProviderWithBaseURL("test-key", server.URL)

	messages, err := provider.GetMessages(context.Background(), "agent-123")

	if err != nil {
		t.Fatalf("GetMessages failed: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}

	if messages[0].Role != "user" {
		t.Errorf("expected first message role 'user', got '%s'", messages[0].Role)
	}
	if messages[1].Content != "Hi there!" {
		t.Errorf("expected second message content 'Hi there!', got '%s'", messages[1].Content)
	}
}

func TestProvider_SendMessage_Streaming(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/agents/agent-123/followup" {
			t.Errorf("expected path /agents/agent-123/followup, got %s", r.URL.Path)
		}

		// Parse request body
		var req CursorFollowupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if req.Message != "test message" {
			t.Errorf("expected message 'test message', got '%s'", req.Message)
		}

		// Send SSE response
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("streaming not supported")
		}

		// Send events
		events := []string{
			"event: content\ndata: {\"delta\": \"Hello\"}\n\n",
			"event: content\ndata: {\"delta\": \" world\"}\n\n",
			"event: done\ndata: {\"finish_reason\": \"stop\", \"message_id\": \"msg-456\"}\n\n",
		}

		for _, event := range events {
			io.WriteString(w, event)
			flusher.Flush()
		}
	}))
	defer server.Close()

	provider := NewProviderWithBaseURL("test-key", server.URL)

	chunks, err := provider.SendMessage(context.Background(), "agent-123", "test message")
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	var content strings.Builder
	var finishReason string
	var messageID string

	for chunk := range chunks {
		if chunk.Error != nil {
			t.Errorf("unexpected error: %v", chunk.Error)
		}
		if chunk.Delta != "" {
			content.WriteString(chunk.Delta)
		}
		if chunk.FinishReason != "" {
			finishReason = chunk.FinishReason
		}
		if chunk.MessageID != "" {
			messageID = chunk.MessageID
		}
	}

	if content.String() != "Hello world" {
		t.Errorf("expected content 'Hello world', got '%s'", content.String())
	}
	if finishReason != "stop" {
		t.Errorf("expected finish reason 'stop', got '%s'", finishReason)
	}
	if messageID != "msg-456" {
		t.Errorf("expected message ID 'msg-456', got '%s'", messageID)
	}
}

func TestProvider_StopAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/agents/agent-123/stop" {
			t.Errorf("expected path /agents/agent-123/stop, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	provider := NewProviderWithBaseURL("test-key", server.URL)

	err := provider.StopAgent(context.Background(), "agent-123")

	if err != nil {
		t.Fatalf("StopAgent failed: %v", err)
	}
}

func TestProvider_ErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		responseBody string
		expectedCode string
	}{
		{
			name:         "unauthorized",
			statusCode:   http.StatusUnauthorized,
			responseBody: `{"error": {"code": "unauthorized", "message": "Invalid API key"}}`,
			expectedCode: "unauthorized",
		},
		{
			name:         "not found",
			statusCode:   http.StatusNotFound,
			responseBody: `{"error": {"code": "not_found", "message": "Agent not found"}}`,
			expectedCode: "not_found",
		},
		{
			name:         "rate limited",
			statusCode:   http.StatusTooManyRequests,
			responseBody: `{"error": {"code": "rate_limited", "message": "Rate limit exceeded"}}`,
			expectedCode: "rate_limited",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				io.WriteString(w, tt.responseBody)
			}))
			defer server.Close()

			provider := NewProviderWithBaseURL("test-key", server.URL)

			_, err := provider.GetAgent(context.Background(), "agent-123")

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			providerErr, ok := err.(*providers.ProviderError)
			if !ok {
				t.Fatalf("expected ProviderError, got %T", err)
			}

			if providerErr.Code != tt.expectedCode {
				t.Errorf("expected code '%s', got '%s'", tt.expectedCode, providerErr.Code)
			}
			if providerErr.StatusCode != tt.statusCode {
				t.Errorf("expected status code %d, got %d", tt.statusCode, providerErr.StatusCode)
			}
		})
	}
}

func TestMapCursorStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected providers.AgentStatus
	}{
		{"pending", providers.AgentStatusPending},
		{"queued", providers.AgentStatusPending},
		{"running", providers.AgentStatusRunning},
		{"in_progress", providers.AgentStatusRunning},
		{"active", providers.AgentStatusRunning},
		{"completed", providers.AgentStatusCompleted},
		{"done", providers.AgentStatusCompleted},
		{"finished", providers.AgentStatusCompleted},
		{"failed", providers.AgentStatusFailed},
		{"error", providers.AgentStatusFailed},
		{"stopped", providers.AgentStatusStopped},
		{"cancelled", providers.AgentStatusStopped},
		{"canceled", providers.AgentStatusStopped},
		{"unknown", providers.AgentStatusPending},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapCursorStatus(tt.input)
			if result != tt.expected {
				t.Errorf("mapCursorStatus(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
