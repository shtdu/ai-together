// Copyright (c) 2025 Code Together
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


// Package integration_manager provides HTTP transport middleware for the manager API client.
// It includes logging, authentication, and retry functionality.
package integration_manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

// LoggingTransport logs HTTP requests and responses.
type LoggingTransport struct {
	base   http.RoundTripper
	logger *slog.Logger
}

// NewLoggingTransport creates a new logging transport.
// If base is nil, http.DefaultTransport is used.
func NewLoggingTransport(base http.RoundTripper, logger *slog.Logger) *LoggingTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &LoggingTransport{
		base:   base,
		logger: logger,
	}
}

// RoundTrip implements http.RoundTripper.
func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Log request
	t.logRequest(req)

	// Make the request
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		t.logger.Error("API request failed",
			"method", req.Method,
			"url", req.URL.String(),
			"duration", time.Since(start).String(),
			"error", err.Error(),
		)
		return nil, err
	}

	// Log response
	duration := time.Since(start)
	t.logResponse(resp, duration)

	return resp, nil
}

func (t *LoggingTransport) logRequest(req *http.Request) {
	// Sanitize headers - don't log auth tokens
	headers := make(map[string]string)
	for k, v := range req.Header {
		if k == "Authorization" || k == "Cookie" {
			headers[k] = "***REDACTED***"
		} else {
			headers[k] = strings.Join(v, ", ")
		}
	}

	// Capture request body if present
	var bodyStr string
	var bodyLogged bool
	if req.Body != nil && req.Body != http.NoBody {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {
			// Restore the body for the actual request
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			// Try to prettify JSON
			bodyStr = t.prettifyJSON(bodyBytes)
			bodyLogged = true
		}
		// If we can't read the body, continue without logging it
	}

	// Log main request info
	t.logger.Info("API request",
		"method", req.Method,
		"url", req.URL.String(),
		"headers", headers,
	)

	// Log body separately if present (directly to avoid escaping)
	if bodyLogged && bodyStr != "" {
		fmt.Fprintf(os.Stderr, "┌─ API request body:\n%s\n└────────\n", bodyStr)
	}
}

func (t *LoggingTransport) logResponse(resp *http.Response, duration time.Duration) {
	// Capture response body if present
	var bodyStr string
	var bodyLogged bool
	if resp.Body != nil && resp.Body != http.NoBody {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil {
			// Restore the body for the caller
			resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			// Try to prettify JSON
			bodyStr = t.prettifyJSON(bodyBytes)
			bodyLogged = true
		}
		// If we can't read the body, continue without logging it
	}

	// Log main response info
	t.logger.Info("API response",
		"status", resp.StatusCode,
		"status_text", http.StatusText(resp.StatusCode),
		"duration", duration.String(),
		"content_length", resp.ContentLength,
	)

	// Log body separately if present (directly to avoid escaping)
	if bodyLogged && bodyStr != "" {
		fmt.Fprintf(os.Stderr, "┌─ API response body:\n%s\n└────────\n", bodyStr)
	}
}

// prettifyJSON attempts to parse the input as JSON and return a prettified string.
// If the input is not valid JSON, it returns the original string.
func (t *LoggingTransport) prettifyJSON(data []byte) string {
	// Check if the content might be JSON by looking at the first byte
	if len(data) == 0 {
		return ""
	}

	// Trim whitespace and check if it starts with { or [
	trimmed := strings.TrimSpace(string(data))
	if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		// Not JSON, return as-is
		return string(data)
	}

	// Try to parse as JSON
	var iface interface{}
	if err := json.Unmarshal(data, &iface); err != nil {
		// Not valid JSON, return original string
		return string(data)
	}

	// Prettify the JSON
	prettified, err := json.MarshalIndent(iface, "", "  ")
	if err != nil {
		// If marshaling fails, return original
		return string(data)
	}

	return string(prettified)
}

// AuthTransport injects JWT tokens into requests.
type AuthTransport struct {
	base     http.RoundTripper
	getToken TokenGetter
}

// NewAuthTransport creates a new auth transport.
// If base is nil, http.DefaultTransport is used.
func NewAuthTransport(base http.RoundTripper, getToken TokenGetter) *AuthTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &AuthTransport{
		base:     base,
		getToken: getToken,
	}
}

// RoundTrip implements http.RoundTripper.
func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Skip auth if token getter is nil
	if t.getToken == nil {
		return t.base.RoundTrip(req)
	}

	// Get the token
	token, err := t.getToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	// Add Authorization header if token is present
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return t.base.RoundTrip(req)
}

// RetryTransport retries failed requests with exponential backoff.
type RetryTransport struct {
	base       http.RoundTripper
	maxRetries int
	retryOn    []int // Status codes to retry on
	backoff    time.Duration
	logger     *slog.Logger
}

// NewRetryTransport creates a new retry transport.
// If base is nil, http.DefaultTransport is used.
func NewRetryTransport(base http.RoundTripper, logger *slog.Logger) *RetryTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &RetryTransport{
		base:       base,
		maxRetries: 3,
		retryOn:    []int{408, 429, 500, 502, 503, 504},
		backoff:    500 * time.Millisecond,
		logger:     logger,
	}
}

// RoundTrip implements http.RoundTripper.
func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		// Clone the request for retries (body can only be read once)
		reqClone := cloneRequest(req)

		resp, err = t.base.RoundTrip(reqClone)
		if err != nil {
			if attempt < t.maxRetries && isRetriableError(err) {
				wait := t.backoff * time.Duration(attempt+1)
				t.logger.Warn("Request failed due to error, retrying",
					"attempt", attempt+1,
					"max_retries", t.maxRetries,
					"wait", wait.String(),
					"error", err.Error(),
				)
				time.Sleep(wait)
				continue
			}
			return nil, err
		}

		// Check if we should retry based on status code
		if !t.shouldRetry(resp.StatusCode) {
			return resp, nil
		}

		// This is a retryable status code
		if attempt < t.maxRetries {
			// Drain and close the response body before retry
			drainBody(resp.Body)

			wait := t.backoff * time.Duration(attempt+1)
			t.logger.Warn("Request failed with retryable status, retrying",
				"attempt", attempt+1,
				"max_retries", t.maxRetries,
				"status", resp.StatusCode,
				"wait", wait.String(),
			)
			time.Sleep(wait)
		}
	}

	return resp, nil
}

func (t *RetryTransport) shouldRetry(statusCode int) bool {
	for _, code := range t.retryOn {
		if code == statusCode {
			return true
		}
	}
	return false
}

// isRetriableError determines if an error is retryable (e.g., network errors).
func isRetriableError(err error) bool {
	// Network errors are typically retryable
	// This is a simple check - you could make this more sophisticated
	return err != nil
}

// cloneRequest creates a shallow copy of the request with a new body.
func cloneRequest(req *http.Request) *http.Request {
	r := req.Clone(req.Context())

	// Copy the body if it exists
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			// If we can't read the body, we can't retry
			return req
		}
		req.Body.Close()

		// Create a new reader for the cloned request
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		// Also update the original request's body so it can be retried
		req.Body = io.NopCloser(strings.NewReader(string(body)))
	}

	return r
}

// drainBody reads and closes the response body to allow reusing the connection.
func drainBody(body io.ReadCloser) {
	if body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, body)
	body.Close()
}

// NewTransportChain creates a transport chain with retry, auth, and logging.
// The order is: retry -> auth -> logging -> base transport.
func NewTransportChain(getToken TokenGetter, logger *slog.Logger) http.RoundTripper {
	base := http.DefaultTransport
	logging := NewLoggingTransport(base, logger)
	auth := NewAuthTransport(logging, getToken)
	retry := NewRetryTransport(auth, logger)
	return retry
}
