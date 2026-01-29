package cloudprovider

import (
	"errors"
	"fmt"
)

// Standard errors for cloud provider operations.
// These errors can be checked using errors.Is() for error handling.
var (
	// ErrAgentNotFound is returned when an agent cannot be found.
	ErrAgentNotFound = errors.New("agent not found")

	// ErrUnauthorized is returned when authentication fails.
	ErrUnauthorized = errors.New("unauthorized: invalid or expired credentials")

	// ErrRateLimited is returned when the provider's rate limit is exceeded.
	ErrRateLimited = errors.New("rate limited: too many requests")

	// ErrContextCanceled is returned when the operation is canceled via context.
	ErrContextCanceled = errors.New("operation canceled")

	// ErrInvalidInput is returned when the input parameters are invalid.
	ErrInvalidInput = errors.New("invalid input parameters")

	// ErrProviderUnavailable is returned when the provider's service is unavailable.
	ErrProviderUnavailable = errors.New("provider service unavailable")

	// ErrQuotaExceeded is returned when the account's quota is exceeded.
	ErrQuotaExceeded = errors.New("quota exceeded")

	// ErrAgentBusy is returned when trying to interact with a busy agent.
	ErrAgentBusy = errors.New("agent is busy processing another request")

	// ErrStreamClosed is returned when trying to read from a closed stream.
	ErrStreamClosed = errors.New("stream closed")

	// ErrNoCredentials is returned when credentials are required but not configured.
	ErrNoCredentials = errors.New("no credentials configured")

	// ErrProviderNotFound is returned when a provider is not registered with the manager.
	ErrProviderNotFound = errors.New("provider not found")

	// ErrProviderAlreadyRegistered is returned when trying to register a provider that already exists.
	ErrProviderAlreadyRegistered = errors.New("provider already registered")
)

// ProviderError wraps an error with provider-specific context.
type ProviderError struct {
	// Provider is the name of the provider that returned the error
	Provider string
	// Operation is the operation that failed (e.g., "CreateAgent", "SendMessage")
	Operation string
	// Err is the underlying error
	Err error
	// StatusCode is the HTTP status code if applicable
	StatusCode int
	// Message is an optional human-readable message
	Message string
}

// Error implements the error interface.
func (e *ProviderError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s.%s: %s: %v", e.Provider, e.Operation, e.Message, e.Err)
	}
	return fmt.Sprintf("%s.%s: %v", e.Provider, e.Operation, e.Err)
}

// Unwrap returns the underlying error for errors.Is/As support.
func (e *ProviderError) Unwrap() error {
	return e.Err
}

// NewProviderError creates a new ProviderError.
func NewProviderError(provider, operation string, err error) *ProviderError {
	return &ProviderError{
		Provider:  provider,
		Operation: operation,
		Err:       err,
	}
}

// WithStatusCode adds an HTTP status code to the error.
func (e *ProviderError) WithStatusCode(code int) *ProviderError {
	e.StatusCode = code
	return e
}

// WithMessage adds a human-readable message to the error.
func (e *ProviderError) WithMessage(msg string) *ProviderError {
	e.Message = msg
	return e
}

// IsRetryable returns true if the error suggests the operation could succeed if retried.
func IsRetryable(err error) bool {
	if errors.Is(err, ErrRateLimited) {
		return true
	}
	if errors.Is(err, ErrProviderUnavailable) {
		return true
	}
	if errors.Is(err, ErrAgentBusy) {
		return true
	}
	var providerErr *ProviderError
	if errors.As(err, &providerErr) {
		// 429 (Too Many Requests) and 503 (Service Unavailable) are retryable
		return providerErr.StatusCode == 429 || providerErr.StatusCode == 503
	}
	return false
}
