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


// Package integration_manager provides a client factory for the Code Together Manager API.
package integration_manager

import (
	"log/slog"
	"net/http"
	"time"
)

// APIError represents an error response from the server.
type APIError struct {
	StatusCode int
	Message    string `json:"error"`
	// ErrorCode is the machine-readable error code (e.g., "VALIDATION_ERROR")
	ErrorCode string `json:"code,omitempty"`
	// Details contains additional context (only in debug mode)
	Details string `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return http.StatusText(e.StatusCode)
}

// IsNotFound returns true if the error is a 404.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsUnauthorized returns true if the error is a 401.
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsForbidden returns true if the error is a 403.
func (e *APIError) IsForbidden() bool {
	return e.StatusCode == http.StatusForbidden
}

// IsClientError returns true if the error is a 4xx.
func (e *APIError) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// IsServerError returns true if the error is a 5xx.
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500
}

// TokenGetter is a function that returns the current auth token.
type TokenGetter func() (string, error)

// NewAnonymousClient creates a client without authentication (for public endpoints).
func NewAnonymousClient(baseURL string, logger *slog.Logger) (*ClientWithResponses, error) {
	// Create transport with only logging and retry (no auth)
	base := http.DefaultTransport
	logging := NewLoggingTransport(base, logger)
	retry := NewRetryTransport(logging, logger)

	httpClient := &http.Client{
		Transport: retry,
		Timeout:   30 * time.Second,
	}

	return NewClientWithResponses(baseURL, WithHTTPClient(httpClient))
}

// NewAuthenticatedClient creates a client with automatic token injection.
func NewAuthenticatedClient(baseURL string, getToken TokenGetter, logger *slog.Logger) (*ClientWithResponses, error) {
	// Create full transport chain
	transport := NewTransportChain(getToken, logger)

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return NewClientWithResponses(baseURL, WithHTTPClient(httpClient))
}

// NewAnonymousClientImpl creates a client implementation without authentication.
func NewAnonymousClientImpl(baseURL string, logger *slog.Logger) (ClientWithResponsesInterface, error) {
	return NewAnonymousClient(baseURL, logger)
}

// NewAuthenticatedClientImpl creates a client implementation with authentication.
func NewAuthenticatedClientImpl(baseURL string, getToken TokenGetter, logger *slog.Logger) (ClientWithResponsesInterface, error) {
	return NewAuthenticatedClient(baseURL, getToken, logger)
}
