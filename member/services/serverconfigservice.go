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


package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/code-together/shared/integration"
)

// Note: ServerConfig struct is now defined in config.go
// This service uses ConfigService to manage settings in config.json

// ServerConfig stores the server connection configuration
type ServerConfig struct {
	ServerURL string `json:"server_url"`
}

// ServerConfigService manages server connection settings
type ServerConfigService struct {
	configService *ConfigService
	mu            sync.Mutex
	logger        *slog.Logger
	apiClient     integration.ClientWithResponsesInterface // Optional API client for health checks
}

// NewServerConfigService creates a new server config service
func NewServerConfigService(configService *ConfigService, logger *slog.Logger) *ServerConfigService {
	return &ServerConfigService{
		configService: configService,
		logger:        logger,
	}
}

// SetAPIClient sets the API client for server communication
//wails:ignore
func (s *ServerConfigService) SetAPIClient(client integration.ClientWithResponsesInterface) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.apiClient = client
}

// GetServerURL returns the configured server URL
func (s *ServerConfigService) GetServerURL() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	config, err := s.configService.GetServer()
	if err != nil {
		return "", err
	}

	return config.ServerURL, nil
}

// SetServerURL saves the server URL
func (s *ServerConfigService) SetServerURL(url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	config := ServerConfig{
		ServerURL: url,
	}

	if err := s.configService.SetServer(config); err != nil {
		return err
	}

	return nil
}

// TestConnection attempts to connect to the configured server
// If url is provided, it will be used; otherwise, the saved URL will be used
func (s *ServerConfigService) TestConnection(url string) error {
	testURL := url
	if testURL == "" {
		var err error
		testURL, err = s.GetServerURL()
		if err != nil {
			return fmt.Errorf("failed to get server URL: %w", err)
		}
	}

	if testURL == "" {
		return fmt.Errorf("server URL not configured")
	}

	// If testing a custom URL (not the saved one), use direct HTTP request
	// to ensure we test the exact URL provided
	if url != "" {
		return s.testConnectionDirect(testURL)
	}

	// Use API client if available for testing saved URL
	s.mu.Lock()
	client := s.apiClient
	s.mu.Unlock()

	if client != nil {
		return s.testConnectionWithClient(client)
	}

	// Fallback to direct HTTP request for backward compatibility
	return s.testConnectionDirect(testURL)
}

// testConnectionWithClient tests the connection using the API client
func (s *ServerConfigService) testConnectionWithClient(client integration.ClientWithResponsesInterface) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.GetHealthWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	// Check if the response indicates an error
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("server returned error status: %d", resp.StatusCode())
	}

	s.logger.Info("Server health check successful",
		"status", resp.JSON200.Status,
		"version", resp.JSON200.Version,
	)

	return nil
}

// testConnectionDirect tests the connection using direct HTTP request (fallback)
func (s *ServerConfigService) testConnectionDirect(serverURL string) error {
	// Remove trailing slash if present
	if len(serverURL) > 0 && serverURL[len(serverURL)-1] == '/' {
		serverURL = serverURL[:len(serverURL)-1]
	}
	// Ensure /api/v1 base path is included
	testURL := serverURL + "/api/v1/health"

	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set a timeout for this specific request
	ctx := req.Context()
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(timeoutCtx)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	// Any response from the server means it's up and running
	return nil
}
