package cloudprovider

import (
	"errors"
	"fmt"
)

// Standard errors for cloud providers
var (
	// ErrAgentNotFound is returned when an agent cannot be found
	ErrAgentNotFound = errors.New("agent not found")

	// ErrUnauthorized is returned when credentials are invalid
	ErrUnauthorized = errors.New("unauthorized: invalid credentials")

	// ErrRateLimited is returned when the API rate limit is exceeded
	ErrRateLimited = errors.New("rate limited: too many requests")

	// ErrInvalidRequest is returned when the request is malformed
	ErrInvalidRequest = errors.New("invalid request")

	// ErrProviderUnavailable is returned when the provider service is unavailable
	ErrProviderUnavailable = errors.New("provider unavailable")

	// ErrContextCanceled is returned when the context is canceled
	ErrContextCanceled = errors.New("operation canceled")

	// ErrNoCredentials is returned when credentials are not configured
	ErrNoCredentials = errors.New("no credentials configured")

	// ErrStreamClosed is returned when a stream is unexpectedly closed
	ErrStreamClosed = errors.New("stream closed unexpectedly")
)

// APIError represents an error response from a cloud provider API
type APIError struct {
	// StatusCode is the HTTP status code
	StatusCode int

	// Code is the provider-specific error code
	Code string

	// Message is the error message
	Message string

	// Retryable indicates whether the request can be retried
	Retryable bool
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("API error %d (%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// NewAPIError creates a new APIError
func NewAPIError(statusCode int, code, message string) *APIError {
	retryable := statusCode == 429 || statusCode >= 500
	return &APIError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Retryable:  retryable,
	}
}

// IsRetryable returns true if the error can be retried
func IsRetryable(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Retryable
	}
	return errors.Is(err, ErrRateLimited) || errors.Is(err, ErrProviderUnavailable)
}

// IsNotFound returns true if the error indicates a resource was not found
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return errors.Is(err, ErrAgentNotFound)
}

// IsUnauthorized returns true if the error indicates an authentication failure
func IsUnauthorized(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401 || apiErr.StatusCode == 403
	}
	return errors.Is(err, ErrUnauthorized)
}
