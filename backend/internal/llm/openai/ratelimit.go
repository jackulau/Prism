package openai

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimiter implements rate limiting for API requests
type RateLimiter struct {
	requestsPerMinute int
	tokensPerMinute   int

	// Sliding window counters
	requestCount  int
	tokenCount    int
	windowStart   time.Time
	windowSize    time.Duration

	// Rate limit headers from API
	remainingRequests int
	remainingTokens   int
	resetRequests     time.Time
	resetTokens       time.Time

	mu sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute, tokensPerMinute int) *RateLimiter {
	return &RateLimiter{
		requestsPerMinute: requestsPerMinute,
		tokensPerMinute:   tokensPerMinute,
		windowSize:        time.Minute,
		windowStart:       time.Now(),
		remainingRequests: requestsPerMinute,
		remainingTokens:   tokensPerMinute,
	}
}

// DefaultRateLimiter returns a rate limiter with OpenAI's default limits
func DefaultRateLimiter() *RateLimiter {
	// Default limits for GPT-4 tier
	return NewRateLimiter(500, 30000)
}

// Wait waits until the request can proceed, respecting rate limits
func (rl *RateLimiter) Wait(ctx context.Context, estimatedTokens int) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Reset window if needed
	if now.Sub(rl.windowStart) >= rl.windowSize {
		rl.windowStart = now
		rl.requestCount = 0
		rl.tokenCount = 0
	}

	// Check if we need to wait
	waitDuration := rl.calculateWait(estimatedTokens, now)
	if waitDuration > 0 {
		rl.mu.Unlock()
		select {
		case <-time.After(waitDuration):
		case <-ctx.Done():
			rl.mu.Lock()
			return ctx.Err()
		}
		rl.mu.Lock()
		// Reset after waiting
		rl.windowStart = time.Now()
		rl.requestCount = 0
		rl.tokenCount = 0
	}

	// Record this request
	rl.requestCount++
	rl.tokenCount += estimatedTokens

	return nil
}

// calculateWait calculates how long to wait before making a request
func (rl *RateLimiter) calculateWait(tokens int, now time.Time) time.Duration {
	var maxWait time.Duration

	// Check request limit
	if rl.requestCount >= rl.requestsPerMinute {
		waitUntil := rl.windowStart.Add(rl.windowSize)
		if wait := waitUntil.Sub(now); wait > maxWait {
			maxWait = wait
		}
	}

	// Check token limit
	if rl.tokenCount+tokens > rl.tokensPerMinute {
		waitUntil := rl.windowStart.Add(rl.windowSize)
		if wait := waitUntil.Sub(now); wait > maxWait {
			maxWait = wait
		}
	}

	// Check API-reported limits
	if rl.remainingRequests <= 0 && now.Before(rl.resetRequests) {
		if wait := rl.resetRequests.Sub(now); wait > maxWait {
			maxWait = wait
		}
	}
	if rl.remainingTokens <= tokens && now.Before(rl.resetTokens) {
		if wait := rl.resetTokens.Sub(now); wait > maxWait {
			maxWait = wait
		}
	}

	return maxWait
}

// UpdateFromHeaders updates rate limit info from API response headers
func (rl *RateLimiter) UpdateFromHeaders(headers http.Header) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Parse remaining requests
	if remaining := headers.Get("X-RateLimit-Remaining-Requests"); remaining != "" {
		if val, err := strconv.Atoi(remaining); err == nil {
			rl.remainingRequests = val
		}
	}

	// Parse remaining tokens
	if remaining := headers.Get("X-RateLimit-Remaining-Tokens"); remaining != "" {
		if val, err := strconv.Atoi(remaining); err == nil {
			rl.remainingTokens = val
		}
	}

	// Parse reset times
	if reset := headers.Get("X-RateLimit-Reset-Requests"); reset != "" {
		if t, err := time.Parse(time.RFC3339, reset); err == nil {
			rl.resetRequests = t
		}
	}

	if reset := headers.Get("X-RateLimit-Reset-Tokens"); reset != "" {
		if t, err := time.Parse(time.RFC3339, reset); err == nil {
			rl.resetTokens = t
		}
	}
}

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  time.Second,
		MaxDelay:      time.Minute,
		BackoffFactor: 2.0,
	}
}

// isRetryableError checks if an error/status should be retried
func isRetryableError(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

// isRetryableNetworkError checks if a network error is retryable
func isRetryableNetworkError(err error) bool {
	if err == nil {
		return false
	}
	// Network errors are generally retryable
	return true
}

// RetryableRequest wraps a request function with retry logic
type RetryableRequest struct {
	config      *RetryConfig
	rateLimiter *RateLimiter
}

// NewRetryableRequest creates a new retryable request handler
func NewRetryableRequest(config *RetryConfig, rateLimiter *RateLimiter) *RetryableRequest {
	if config == nil {
		config = DefaultRetryConfig()
	}
	return &RetryableRequest{
		config:      config,
		rateLimiter: rateLimiter,
	}
}

// Do executes a request with retry logic
func (rr *RetryableRequest) Do(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
	var lastErr error
	delay := rr.config.InitialDelay

	for attempt := 0; attempt <= rr.config.MaxRetries; attempt++ {
		// Clone request for retry (body needs special handling)
		reqCopy := req.Clone(ctx)

		resp, err := client.Do(reqCopy)
		if err != nil {
			if !isRetryableNetworkError(err) {
				return nil, err
			}
			lastErr = err
		} else {
			// Update rate limiter from response headers
			if rr.rateLimiter != nil {
				rr.rateLimiter.UpdateFromHeaders(resp.Header)
			}

			// Check if we should retry
			if !isRetryableError(resp.StatusCode) {
				return resp, nil
			}

			// Read and close body before retry
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))

			// Check for Retry-After header
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					delay = time.Duration(seconds) * time.Second
				}
			}
		}

		// Don't retry if this was the last attempt
		if attempt >= rr.config.MaxRetries {
			break
		}

		// Wait before retry
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * rr.config.BackoffFactor)
		if delay > rr.config.MaxDelay {
			delay = rr.config.MaxDelay
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// CalculateBackoff calculates the backoff duration for a given attempt
func CalculateBackoff(attempt int, config *RetryConfig) time.Duration {
	if config == nil {
		config = DefaultRetryConfig()
	}

	delay := float64(config.InitialDelay) * math.Pow(config.BackoffFactor, float64(attempt))
	if time.Duration(delay) > config.MaxDelay {
		return config.MaxDelay
	}
	return time.Duration(delay)
}
