package remote

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	fhws "github.com/fasthttp/websocket"
)

// WebSocketProxy handles WebSocket connection proxying
type WebSocketProxy struct {
	targetURL *url.URL
	dialer    *fhws.Dialer
}

// NewWebSocketProxy creates a new WebSocket proxy
func NewWebSocketProxy(targetURL *url.URL) *WebSocketProxy {
	// Convert HTTP URL to WebSocket URL
	wsURL := *targetURL
	if wsURL.Scheme == "https" {
		wsURL.Scheme = "wss"
	} else {
		wsURL.Scheme = "ws"
	}

	return &WebSocketProxy{
		targetURL: &wsURL,
		dialer: &fhws.Dialer{
			HandshakeTimeout: 10 * time.Second,
		},
	}
}

// handleWebSocket handles WebSocket upgrade and proxying
func (t *TunnelServer) handleWebSocket(c *websocket.Conn) {
	session := c.Locals("session").(*RemoteSession)
	clientIP := c.RemoteAddr().String()

	// Create a tunnel connection for tracking
	conn := NewTunnelConnection(clientIP, session)
	if err := t.connManager.Add(conn); err != nil {
		log.Printf("Failed to add WebSocket connection: %v", err)
		c.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseTryAgainLater, "connection limit exceeded"))
		return
	}
	defer func() {
		conn.Close()
		t.connManager.Remove(conn.ID)
	}()

	// Connect to the backend WebSocket
	backendURL := t.wsProxy.targetURL.String() + "/ws"

	// Add authentication headers for the backend
	header := http.Header{}
	header.Set("X-Remote-Access", "true")
	header.Set("X-Forwarded-For", clientIP)
	header.Set("X-Real-IP", clientIP)
	header.Set("X-Remote-Session-ID", session.ID)

	backendConn, resp, err := t.wsProxy.dialer.Dial(backendURL, header)
	if err != nil {
		log.Printf("Failed to connect to backend WebSocket: %v", err)
		if resp != nil {
			log.Printf("Backend response status: %d", resp.StatusCode)
		}
		c.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "backend connection failed"))
		return
	}
	defer backendConn.Close()

	log.Printf("WebSocket proxy established: client=%s session=%s", clientIP, session.ID)

	// Proxy messages bidirectionally
	var wg sync.WaitGroup
	wg.Add(2)

	// Client -> Backend
	go func() {
		defer wg.Done()
		t.proxyWebSocketMessages(c, backendConn, conn, "client->backend")
	}()

	// Backend -> Client
	go func() {
		defer wg.Done()
		t.proxyWebSocketMessagesReverse(backendConn, c, conn, "backend->client")
	}()

	wg.Wait()
	log.Printf("WebSocket proxy closed: client=%s session=%s bytesIn=%d bytesOut=%d",
		clientIP, session.ID, conn.GetBytesIn(), conn.GetBytesOut())
}

// proxyWebSocketMessages proxies messages from Fiber WebSocket to fasthttp WebSocket
func (t *TunnelServer) proxyWebSocketMessages(src *websocket.Conn, dst *fhws.Conn, conn *TunnelConnection, direction string) {
	for {
		messageType, message, err := src.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket read error (%s): %v", direction, err)
			}
			// Send close message to the other end
			dst.WriteMessage(fhws.CloseMessage,
				fhws.FormatCloseMessage(fhws.CloseNormalClosure, ""))
			return
		}

		// Track bytes
		conn.AddBytesIn(int64(len(message)))
		conn.UpdateLastSeen()

		// Forward the message
		if err := dst.WriteMessage(messageType, message); err != nil {
			log.Printf("WebSocket write error (%s): %v", direction, err)
			return
		}
	}
}

// proxyWebSocketMessagesReverse proxies messages from fasthttp WebSocket to Fiber WebSocket
func (t *TunnelServer) proxyWebSocketMessagesReverse(src *fhws.Conn, dst *websocket.Conn, conn *TunnelConnection, direction string) {
	for {
		messageType, message, err := src.ReadMessage()
		if err != nil {
			if fhws.IsUnexpectedCloseError(err, fhws.CloseGoingAway, fhws.CloseNormalClosure) {
				log.Printf("WebSocket read error (%s): %v", direction, err)
			}
			// Send close message to the other end
			dst.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}

		// Track bytes
		conn.AddBytesOut(int64(len(message)))
		conn.UpdateLastSeen()

		// Forward the message
		if err := dst.WriteMessage(messageType, message); err != nil {
			log.Printf("WebSocket write error (%s): %v", direction, err)
			return
		}
	}
}

// WebSocketMessage represents a WebSocket message for logging/debugging
type WebSocketMessage struct {
	Type    int
	Payload []byte
}

// String returns a string representation of the message type
func (m *WebSocketMessage) TypeString() string {
	switch m.Type {
	case websocket.TextMessage:
		return "text"
	case websocket.BinaryMessage:
		return "binary"
	case websocket.CloseMessage:
		return "close"
	case websocket.PingMessage:
		return "ping"
	case websocket.PongMessage:
		return "pong"
	default:
		return fmt.Sprintf("unknown(%d)", m.Type)
	}
}
