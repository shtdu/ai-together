// Package support provides test utilities and context management for BDD tests
package support

import (
	"log/slog"

	"github.com/code-together/shared/integration"
)

// NewAnonymousClient creates a client without authentication for public endpoints.
// Verbose parameter controls whether request/response bodies are logged.
// Returns *integration.ClientWithResponses for making API calls.
func NewAnonymousClient(serverURL string, verbose bool) (*integration.ClientWithResponses, error) {
	logger := slog.Default()
	return integration.NewAnonymousClient(serverURL, logger, verbose)
}

// NewAuthenticatedClient creates a client with automatic token injection.
// Verbose parameter controls whether request/response bodies are logged.
// The getToken function should return the current auth token.
// Returns *integration.ClientWithResponses for making authenticated API calls.
func NewAuthenticatedClient(serverURL string, getToken integration.TokenGetter, verbose bool) (*integration.ClientWithResponses, error) {
	logger := slog.Default()
	return integration.NewAuthenticatedClient(serverURL, getToken, logger, verbose)
}
