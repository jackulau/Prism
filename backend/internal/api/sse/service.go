package sse

import (
	"bufio"
	"context"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	// Heartbeat interval for keeping connections alive
	heartbeatInterval = 30 * time.Second

	// Client send buffer size
	sendBufferSize = 256
)

// Client represents an SSE client connection
type Client struct {
	ID       string
	UserID   string
	Writer   *bufio.Writer
	Ctx      *fiber.Ctx
	Done     chan struct{}
	Send     chan *Event
	mu       sync.Mutex
}

// Service manages SSE client connections and event broadcasting
type Service struct {
	// Registered clients by client ID
	clients map[string]*Client

	// Clients indexed by user ID for user-targeted broadcasts
	userClients map[string]map[string]*Client

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewService creates a new SSE service
func NewService() *Service {
	return &Service{
		clients:     make(map[string]*Client),
		userClients: make(map[string]map[string]*Client),
	}
}

// RegisterClient registers a new SSE client and returns it
func (s *Service) RegisterClient(userID string, c *fiber.Ctx) (*Client, error) {
	client := &Client{
		ID:     uuid.New().String(),
		UserID: userID,
		Ctx:    c,
		Done:   make(chan struct{}),
		Send:   make(chan *Event, sendBufferSize),
	}

	s.mu.Lock()
	s.clients[client.ID] = client

	if s.userClients[userID] == nil {
		s.userClients[userID] = make(map[string]*Client)
	}
	s.userClients[userID][client.ID] = client
	s.mu.Unlock()

	log.Printf("SSE client registered: id=%s user=%s", client.ID, userID)

	return client, nil
}

// UnregisterClient removes a client from the service
func (s *Service) UnregisterClient(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, ok := s.clients[clientID]
	if !ok {
		return
	}

	// Close the done channel to signal goroutines to stop
	close(client.Done)

	// Remove from user index
	if userClients, ok := s.userClients[client.UserID]; ok {
		delete(userClients, clientID)
		if len(userClients) == 0 {
			delete(s.userClients, client.UserID)
		}
	}

	// Remove from main index
	delete(s.clients, clientID)

	log.Printf("SSE client unregistered: id=%s user=%s", clientID, client.UserID)
}

// GetClient returns a client by ID
func (s *Service) GetClient(clientID string) (*Client, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[clientID]
	return client, ok
}

// SendEvent sends an event to a specific client
func (s *Service) SendEvent(clientID string, event *Event) error {
	s.mu.RLock()
	client, ok := s.clients[clientID]
	s.mu.RUnlock()

	if !ok {
		return nil // Client not found, silently ignore
	}

	select {
	case client.Send <- event:
		return nil
	default:
		log.Printf("SSE client buffer full, dropping event for client=%s", clientID)
		return nil
	}
}

// BroadcastToUser sends an event to all clients of a user
func (s *Service) BroadcastToUser(userID string, event *Event) error {
	s.mu.RLock()
	userClients, ok := s.userClients[userID]
	if !ok {
		s.mu.RUnlock()
		return nil
	}

	// Copy client IDs to avoid holding lock while sending
	clientIDs := make([]string, 0, len(userClients))
	for id := range userClients {
		clientIDs = append(clientIDs, id)
	}
	s.mu.RUnlock()

	for _, clientID := range clientIDs {
		s.SendEvent(clientID, event)
	}

	return nil
}

// StreamToClient streams events to a client using the Fiber context
// This should be called from an HTTP handler and blocks until the client disconnects
func (s *Service) StreamToClient(ctx context.Context, client *Client, c *fiber.Ctx) error {
	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no") // Disable nginx buffering
	c.Set("Transfer-Encoding", "chunked")

	// Create a ticker for heartbeats
	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()

	// Use Fiber's streaming context
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		client.mu.Lock()
		client.Writer = w
		client.mu.Unlock()

		// Send initial connection event
		connEvent := NewEvent(EventConnected, map[string]interface{}{
			"client_id": client.ID,
			"message":   "SSE connection established",
		})
		if _, err := w.WriteString(connEvent.Format()); err != nil {
			return
		}
		w.Flush()

		for {
			select {
			case <-ctx.Done():
				return
			case <-client.Done:
				return
			case event, ok := <-client.Send:
				if !ok {
					return
				}
				if _, err := w.WriteString(event.Format()); err != nil {
					log.Printf("Failed to write SSE event: %v", err)
					return
				}
				if err := w.Flush(); err != nil {
					log.Printf("Failed to flush SSE event: %v", err)
					return
				}
			case <-heartbeat.C:
				// Send heartbeat to keep connection alive
				hbEvent := NewHeartbeat()
				if _, err := w.WriteString(hbEvent.Format()); err != nil {
					return
				}
				if err := w.Flush(); err != nil {
					return
				}
			}
		}
	})

	return nil
}

// GetClientCount returns the total number of connected clients
func (s *Service) GetClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

// GetUserClientCount returns the number of clients for a specific user
func (s *Service) GetUserClientCount(userID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if userClients, ok := s.userClients[userID]; ok {
		return len(userClients)
	}
	return 0
}
