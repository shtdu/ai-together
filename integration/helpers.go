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

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	// TestDataDir is the base directory for test fixtures
	TestDataDir = "testdata"
	// LicenseFixturesDir contains license PEM files
	LicenseFixturesDir = "testdata/licenses"

	// License fixture names
	LicenseOpenSource       = "opensource.pem"
	LicenseCommercial       = "commercial.pem"
	LicenseExpired          = "expired.pem"
	LicenseZeroSeats        = "zero_seats.pem"
	LicenseImmediateExpiry  = "immediate_expiry.pem"
	LicenseInvalidSignature = "invalid_signature.pem"
)

var (
	providerCounter int64
	providerMutex   sync.Mutex
)

// generateUniqueProviderName creates a unique provider name by appending a counter
// This prevents duplicate name errors when tests run in sequence
func generateUniqueProviderName(baseName string) string {
	providerMutex.Lock()
	defer providerMutex.Unlock()
	providerCounter++
	return fmt.Sprintf("%s-%d", baseName, providerCounter)
}

// TestContext holds all the context needed for running integration tests.
type TestContext struct {
	ServerURL string
	TestDBURL string
	Logger    *slog.Logger
}

// LoadLicenseFixture reads a license PEM file from disk and returns its contents.
func LoadLicenseFixture(fixtureName string) (string, error) {
	path := filepath.Join(LicenseFixturesDir, fixtureName)
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

// setupTestContext initializes the test environment:
// 1. Loads environment from .env.test
// 2. Creates/updates test database
// 3. Connects to already-running test server (expects server on TEST_SERVER_URL)
// 4. Performs initial setup if needed
//
// Note: The test server must be started separately using ./test-server.sh
func setupTestContext() (*TestContext, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load environment from .env.test file
	loadTestEnv()

	// Get server URL from environment (default: localhost:8088)
	serverURL := os.Getenv("TEST_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8088"
	}

	// Setup test database
	testDBURL := setupTestDB(logger)
	logger.Info("Test database ready", "url", testDBURL)

	// Wait for server to be ready
	if err := waitForServer(serverURL, logger); err != nil {
		return nil, fmt.Errorf("test server not ready: %w", err)
	}

	ctx := &TestContext{
		ServerURL: serverURL,
		TestDBURL: testDBURL,
		Logger:    logger,
	}

	logger.Info("Test context ready", "server_url", serverURL)
	return ctx, nil
}

// waitForServer waits for the test server to be ready by attempting to connect.
func waitForServer(serverURL string, logger *slog.Logger) error {
	// Extract host and port from URL
	// Expected format: http://localhost:8088
	host := "localhost:8088" // Default for test server

	// Parse URL to get host:port
	if strings.HasPrefix(serverURL, "http://") {
		host = strings.TrimPrefix(serverURL, "http://")
	} else if strings.HasPrefix(serverURL, "https://") {
		host = strings.TrimPrefix(serverURL, "https://")
	}

	for i := 0; i < 100; i++ { // 10 seconds timeout
		conn, err := net.DialTimeout("tcp", host, 500*time.Millisecond)
		if err == nil {
			conn.Close()
			logger.Info("Server is ready", "url", serverURL)
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("server did not become ready within timeout")
}

// performInitialSetup checks if setup is needed and performs it via the setup endpoint.
func performInitialSetup(serverURL string, logger *slog.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to register admin user first (will fail if already exists)
	registerURL := serverURL + "/auth/register"
	registerData := map[string]string{
		"email":    "admin@example.com",
		"password": "AdminPassword123!",
		"name":     "Test Admin",
	}

	jsonData, err := json.Marshal(registerData)
	if err != nil {
		return fmt.Errorf("failed to marshal register request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", registerURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create register request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to register admin: %w", err)
	}
	defer resp.Body.Close()

	// If registration succeeds, great! If it fails (user already exists), that's also fine
	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusBadRequest {
		logger.Info("Admin user ready (registered or already exists)")
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("admin registration failed with status %d: %s", resp.StatusCode, string(body))
}

// loadTestEnv loads environment variables from .env.test file
func loadTestEnv() {
	data, err := os.ReadFile(".env.test")
	if err != nil {
		return
	}

	lines := bytes.Split(data, []byte{'\n'})
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 || bytes.HasPrefix(line, []byte{'#'}) {
			continue
		}

		parts := bytes.SplitN(line, []byte{'='}, 2)
		if len(parts) == 2 {
			key := string(bytes.TrimSpace(parts[0]))
			value := string(bytes.TrimSpace(parts[1]))
			os.Setenv(key, value)
		}
	}
}

// cleanupTestContext performs cleanup.
func cleanupTestContext(ctx *TestContext) {
	// Note: Test server is not stopped here as it runs in a separate terminal
	ctx.Logger.Info("Test cleanup complete")
}
