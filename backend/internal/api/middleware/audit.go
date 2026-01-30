package middleware

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/audit"
)

// AuditConfig configures the audit middleware
type AuditConfig struct {
	// Logger is the audit logger to use
	Logger *audit.Logger

	// SkipPaths are paths that should not be audited
	SkipPaths []string

	// SkipMethods are HTTP methods that should not be audited (e.g., OPTIONS)
	SkipMethods []string

	// IncludeRequestBody determines if request bodies should be logged
	IncludeRequestBody bool

	// IncludeResponseBody determines if response bodies should be logged
	IncludeResponseBody bool

	// MaxBodySize is the maximum size of body to capture (0 = unlimited)
	MaxBodySize int

	// SensitiveHeaders are headers that should be redacted
	SensitiveHeaders []string

	// SensitiveFields are JSON fields that should be redacted from request/response bodies
	SensitiveFields []string
}

// DefaultAuditConfig returns a default audit configuration
func DefaultAuditConfig(logger *audit.Logger) AuditConfig {
	return AuditConfig{
		Logger: logger,
		SkipPaths: []string{
			"/health",
			"/api/v1/sse",
			"/api/v1/ws",
		},
		SkipMethods: []string{"OPTIONS"},
		IncludeRequestBody:  true,
		IncludeResponseBody: false,
		MaxBodySize:         10240, // 10KB
		SensitiveHeaders: []string{
			"authorization",
			"cookie",
			"x-api-key",
		},
		SensitiveFields: []string{
			"password",
			"password_hash",
			"token",
			"access_token",
			"refresh_token",
			"api_key",
			"secret",
			"private_key",
		},
	}
}

// AuditMiddleware creates middleware that logs all API requests to the audit log
func AuditMiddleware(config AuditConfig) fiber.Handler {
	if config.Logger == nil {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	return func(c *fiber.Ctx) error {
		// Check if path should be skipped
		path := c.Path()
		for _, skipPath := range config.SkipPaths {
			if strings.HasPrefix(path, skipPath) {
				return c.Next()
			}
		}

		// Check if method should be skipped
		method := c.Method()
		for _, skipMethod := range config.SkipMethods {
			if method == skipMethod {
				return c.Next()
			}
		}

		// Capture request info before processing
		startTime := time.Now()
		requestBody := captureBody(c.Body(), config.MaxBodySize, config.SensitiveFields)

		// Process request
		err := c.Next()

		// Capture response info after processing
		duration := time.Since(startTime)
		statusCode := c.Response().StatusCode()
		success := statusCode >= 200 && statusCode < 400

		// Get user info from context
		userID := GetUserID(c)
		email := GetEmail(c)
		actorType := "user"
		if userID == "" {
			userID = "anonymous"
			actorType = "anonymous"
		}

		// Determine action and resource type from the request
		action := determineAction(method)
		resourceType, resourceID := determineResource(path)

		// Build metadata
		metadata := map[string]interface{}{
			"method":       method,
			"path":         path,
			"status_code":  statusCode,
			"duration_ms":  duration.Milliseconds(),
			"content_type": c.Get("Content-Type"),
		}

		// Add query parameters (redacted)
		if queryString := string(c.Request().URI().QueryString()); queryString != "" {
			metadata["query"] = redactQueryString(queryString, config.SensitiveFields)
		}

		// Add request body if configured
		if config.IncludeRequestBody && len(requestBody) > 0 {
			metadata["request_body"] = requestBody
		}

		// Capture headers (redacted)
		headers := make(map[string]string)
		c.Request().Header.VisitAll(func(key, value []byte) {
			headerName := strings.ToLower(string(key))
			if isSensitiveHeader(headerName, config.SensitiveHeaders) {
				headers[headerName] = "[REDACTED]"
			} else {
				headers[headerName] = string(value)
			}
		})
		metadata["headers"] = headers

		// Get organization ID if available
		orgID := ""
		if orgIDVal := c.Locals("organizationID"); orgIDVal != nil {
			if oid, ok := orgIDVal.(string); ok {
				orgID = oid
			}
		}

		// Create audit log options
		opts := []audit.AuditEventOption{
			audit.WithActorEmail(email),
			audit.WithIPAddress(c.IP()),
			audit.WithUserAgent(c.Get("User-Agent")),
			audit.WithMetadata(metadata),
		}

		if orgID != "" {
			opts = append(opts, audit.WithOrgID(orgID))
		}

		// Add session ID if available
		if sessionID := c.Locals("sessionID"); sessionID != nil {
			if sid, ok := sessionID.(string); ok {
				opts = append(opts, audit.WithSessionID(sid))
			}
		}

		// Handle errors
		if !success {
			var errorMsg string
			if err != nil {
				errorMsg = err.Error()
			} else {
				errorMsg = string(c.Response().Body())
			}
			opts = append(opts, audit.WithError(&auditError{message: errorMsg}))
		}

		// Log the event asynchronously to not block the response
		go func() {
			_ = config.Logger.Log(userID, actorType, action, resourceType, resourceID, opts...)
		}()

		return err
	}
}

// determineAction maps HTTP methods to audit actions
func determineAction(method string) audit.ActionType {
	switch method {
	case "GET":
		return audit.ActionRead
	case "POST":
		return audit.ActionCreate
	case "PUT", "PATCH":
		return audit.ActionUpdate
	case "DELETE":
		return audit.ActionDelete
	default:
		return audit.ActionAccess
	}
}

// determineResource extracts resource type and ID from the request path
func determineResource(path string) (audit.ResourceType, string) {
	// Remove /api/v1/ prefix
	path = strings.TrimPrefix(path, "/api/v1/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		return audit.ResourceType("unknown"), ""
	}

	// Map path prefixes to resource types
	resourceMap := map[string]audit.ResourceType{
		"users":         audit.ResourceUser,
		"agents":        audit.ResourceAgent,
		"workflows":     audit.ResourceWorkflow,
		"conversations": audit.ResourceConversation,
		"messages":      audit.ResourceMessage,
		"settings":      audit.ResourceSettings,
		"integrations":  audit.ResourceIntegration,
		"tools":         audit.ResourceTool,
		"exports":       audit.ResourceExport,
		"audit":         audit.ResourceAuditLog,
		"sessions":      audit.ResourceSession,
		"auth":          audit.ResourceSession,
		"organizations": audit.ResourceOrganization,
		"workspaces":    audit.ResourceWorkspace,
	}

	resourceType := audit.ResourceType(parts[0])
	if mapped, ok := resourceMap[parts[0]]; ok {
		resourceType = mapped
	}

	resourceID := ""
	if len(parts) > 1 {
		resourceID = parts[1]
	}

	return resourceType, resourceID
}

// captureBody captures and redacts sensitive fields from a request body
func captureBody(body []byte, maxSize int, sensitiveFields []string) map[string]interface{} {
	if len(body) == 0 {
		return nil
	}

	if maxSize > 0 && len(body) > maxSize {
		return map[string]interface{}{
			"_truncated": true,
			"_size":      len(body),
		}
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		// Not JSON, return as string (truncated if needed)
		s := string(body)
		if maxSize > 0 && len(s) > maxSize {
			s = s[:maxSize] + "..."
		}
		return map[string]interface{}{
			"_raw": s,
		}
	}

	return redactFields(data, sensitiveFields)
}

// redactFields recursively redacts sensitive fields from a map
func redactFields(data map[string]interface{}, sensitiveFields []string) map[string]interface{} {
	result := make(map[string]interface{})
	sensitiveSet := make(map[string]bool)
	for _, f := range sensitiveFields {
		sensitiveSet[strings.ToLower(f)] = true
	}

	for key, value := range data {
		if sensitiveSet[strings.ToLower(key)] {
			result[key] = "[REDACTED]"
			continue
		}

		switch v := value.(type) {
		case map[string]interface{}:
			result[key] = redactFields(v, sensitiveFields)
		case []interface{}:
			arr := make([]interface{}, len(v))
			for i, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					arr[i] = redactFields(m, sensitiveFields)
				} else {
					arr[i] = item
				}
			}
			result[key] = arr
		default:
			result[key] = value
		}
	}

	return result
}

// redactQueryString redacts sensitive parameters from a query string
func redactQueryString(query string, sensitiveFields []string) string {
	sensitiveSet := make(map[string]bool)
	for _, f := range sensitiveFields {
		sensitiveSet[strings.ToLower(f)] = true
	}

	pairs := strings.Split(query, "&")
	result := make([]string, 0, len(pairs))

	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 && sensitiveSet[strings.ToLower(parts[0])] {
			result = append(result, parts[0]+"=[REDACTED]")
		} else {
			result = append(result, pair)
		}
	}

	return strings.Join(result, "&")
}

// isSensitiveHeader checks if a header name should be redacted
func isSensitiveHeader(name string, sensitiveHeaders []string) bool {
	name = strings.ToLower(name)
	for _, h := range sensitiveHeaders {
		if strings.ToLower(h) == name {
			return true
		}
	}
	return false
}

// auditError is a simple error wrapper for audit logging
type auditError struct {
	message string
}

func (e *auditError) Error() string {
	return e.message
}
