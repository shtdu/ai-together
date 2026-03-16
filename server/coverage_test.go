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

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

// TestCoverageServer runs the server for coverage collection.
// Usage: TEST_COVERAGE_SERVER=1 go test -v -cover -coverpkg=./... -run TestCoverageServer
// This test keeps the server running until interrupted (Ctrl+C).
// Set GOCOVERDIR environment variable to collect coverage data.
func TestCoverageServer(t *testing.T) {
	// Check if this is the coverage mode
	if os.Getenv("TEST_COVERAGE_SERVER") != "1" {
		t.Skip("Skipping coverage server test - set TEST_COVERAGE_SERVER=1 to run")
		return
	}

	// Start the server in a goroutine
	serverDone := make(chan bool)
	go func() {
		startServer()
		serverDone <- true
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until signal received
	sig := <-sigChan
	fmt.Printf("\nReceived signal: %v, shutting down coverage server...\n", sig)

	// Note: The server goroutine will exit when the process exits
	// We don't wait for serverDone because router.Run() blocks forever
	_ = serverDone
}
