package remote

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

// proxyMiddleware handles HTTP proxying to the main server
func (t *TunnelServer) proxyMiddleware(c *fiber.Ctx) error {
	conn := c.Locals("tunnel_connection").(*TunnelConnection)

	// Get the original request
	req := c.Request()

	// Create a new fasthttp request for the backend
	backendReq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(backendReq)

	// Copy the request
	req.CopyTo(backendReq)

	// Set the target URL
	backendReq.SetRequestURI(t.config.TargetURL + string(req.RequestURI()))

	// Add remote access headers
	backendReq.Header.Set("X-Remote-Access", "true")
	backendReq.Header.Set("X-Forwarded-For", conn.ClientIP)
	backendReq.Header.Set("X-Real-IP", conn.ClientIP)
	backendReq.Header.Set("X-Remote-Session-ID", conn.Session.ID)

	// Remove hop-by-hop headers
	removeHopByHopHeaders(&backendReq.Header)

	// Track bytes in
	conn.AddBytesIn(int64(len(req.Body())))

	// Create fasthttp client for the request
	client := &fasthttp.Client{}
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Send the request to the backend
	if err := client.Do(backendReq, resp); err != nil {
		log.Printf("Proxy error: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "failed to connect to backend server",
		})
	}

	// Copy response headers
	resp.Header.VisitAll(func(key, value []byte) {
		keyStr := string(key)
		// Skip hop-by-hop headers
		if isHopByHopHeader(keyStr) {
			return
		}
		c.Set(keyStr, string(value))
	})

	// Track bytes out
	conn.AddBytesOut(int64(len(resp.Body())))

	// Update last seen
	conn.UpdateLastSeen()

	// Set status code and body
	c.Status(resp.StatusCode())
	return c.Send(resp.Body())
}

// handleAuth handles authentication requests
func (t *TunnelServer) handleAuth(c *fiber.Ctx) error {
	type AuthRequest struct {
		Password string `json:"password"`
	}

	var req AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password is required",
		})
	}

	clientIP := getClientIP(c)
	userAgent := c.Get("User-Agent")

	session, token, err := t.authService.Authenticate(req.Password, clientIP, userAgent)
	if err != nil {
		status := fiber.StatusUnauthorized
		if err == ErrRemoteAccessDisabled {
			status = fiber.StatusForbidden
		} else if err == ErrConnectionLimitExceeded {
			status = fiber.StatusTooManyRequests
		}
		return c.Status(status).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"token":      token,
		"session_id": session.ID,
		"expires_at": session.ExpiresAt,
	})
}

// handleLogout handles logout requests
func (t *TunnelServer) handleLogout(c *fiber.Ctx) error {
	session := c.Locals("session").(*RemoteSession)

	// Close all connections for this session
	t.connManager.CloseBySession(session.ID)

	// Invalidate the session
	if err := t.authService.InvalidateSession(session.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to invalidate session",
		})
	}

	return c.JSON(fiber.Map{
		"message": "logged out successfully",
	})
}

// handleStatus handles status requests
func (t *TunnelServer) handleStatus(c *fiber.Ctx) error {
	session := c.Locals("session").(*RemoteSession)
	clientIP := getClientIP(c)

	// Get connection count for this IP
	connCount := t.connManager.CountByIP(clientIP)

	return c.JSON(fiber.Map{
		"session_id":    session.ID,
		"client_ip":     session.ClientIP,
		"created_at":    session.CreatedAt,
		"expires_at":    session.ExpiresAt,
		"last_activity": session.LastActivity,
		"connections":   connCount,
	})
}

// handleHealth handles health check requests
func (t *TunnelServer) handleHealth(c *fiber.Ctx) error {
	stats := t.GetStats()
	return c.JSON(fiber.Map{
		"status":             "healthy",
		"active_connections": stats.ActiveConnections,
		"uptime_seconds":     int64(stats.Uptime.Seconds()),
	})
}

// authMiddleware validates the authentication token
func (t *TunnelServer) authMiddleware(c *fiber.Ctx) error {
	token := extractToken(c)
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	session, err := t.authService.ValidateToken(token)
	if err != nil {
		status := fiber.StatusUnauthorized
		if err == ErrSessionExpired {
			status = fiber.StatusUnauthorized
		}
		return c.Status(status).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Store session in context
	c.Locals("session", session)
	return c.Next()
}

// connectionLimitMiddleware checks connection limits and creates a connection
func (t *TunnelServer) connectionLimitMiddleware(c *fiber.Ctx) error {
	session := c.Locals("session").(*RemoteSession)
	clientIP := getClientIP(c)

	// Check if we can accept a new connection
	if !t.connManager.CanAccept(clientIP) {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": "connection limit exceeded",
		})
	}

	// Create a new connection
	conn := NewTunnelConnection(clientIP, session)
	if err := t.connManager.Add(conn); err != nil {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Store connection in context
	c.Locals("tunnel_connection", conn)

	// Remove connection when request completes
	defer t.connManager.Remove(conn.ID)

	return c.Next()
}

// extractToken extracts the authentication token from the request
func extractToken(c *fiber.Ctx) string {
	// Check Authorization header
	auth := c.Get("Authorization")
	if auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}

	// Check X-Remote-Token header
	token := c.Get("X-Remote-Token")
	if token != "" {
		return token
	}

	// Check query parameter
	token = c.Query("token")
	if token != "" {
		return token
	}

	return ""
}

// getClientIP extracts the real client IP from the request
func getClientIP(c *fiber.Ctx) string {
	// Check X-Forwarded-For header
	xff := c.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in the list
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	realIP := c.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to connection IP
	return c.IP()
}

// removeHopByHopHeaders removes hop-by-hop headers from the request
func removeHopByHopHeaders(h *fasthttp.RequestHeader) {
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"TE",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	for _, header := range hopByHopHeaders {
		h.Del(header)
	}
}

// isHopByHopHeader checks if a header is a hop-by-hop header
func isHopByHopHeader(header string) bool {
	hopByHopHeaders := map[string]bool{
		"Connection":          true,
		"Keep-Alive":          true,
		"Proxy-Authenticate":  true,
		"Proxy-Authorization": true,
		"TE":                  true,
		"Trailers":            true,
		"Transfer-Encoding":   true,
		"Upgrade":             true,
	}
	return hopByHopHeaders[header]
}

// CountingReader wraps an io.Reader and counts bytes read
type CountingReader struct {
	Reader io.Reader
	Count  int64
}

// Read implements io.Reader
func (r *CountingReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.Count += int64(n)
	return
}

// CountingWriter wraps an http.ResponseWriter and counts bytes written
type CountingWriter struct {
	http.ResponseWriter
	Count int64
}

// Write implements io.Writer
func (w *CountingWriter) Write(p []byte) (n int, err error) {
	n, err = w.ResponseWriter.Write(p)
	w.Count += int64(n)
	return
}
