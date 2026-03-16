// Copyright (c) 2025 AI Together
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


package collector

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLogError(t *testing.T) {
	// Create a temporary log file for testing
	tempDir := t.TempDir()
	testLogPath := filepath.Join(tempDir, "test-collector.log")

	// Set custom log file path
	SetLogFile(testLogPath)

	// Clean up any existing log file
	os.Remove(testLogPath)

	// Test 1: Log an error with payload
	testErr := errors.New("test error message")
	testPayload := []byte(`{"test": "payload", "data": 123}`)

	LogError(testErr, testPayload)

	// Verify log file was created and contains expected content
	content, err := os.ReadFile(testLogPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)

	// Check for error message
	if !strings.Contains(contentStr, "test error message") {
		t.Errorf("Log file does not contain error message. Got:\n%s", contentStr)
	}

	// Check for payload (pretty-printed format)
	if !strings.Contains(contentStr, `"test": "payload"`) || !strings.Contains(contentStr, `"data": 123`) {
		t.Errorf("Log file does not contain payload. Got:\n%s", contentStr)
	}

	// Check for timestamp format
	if !strings.Contains(contentStr, "[20") || !strings.Contains(contentStr, "T") {
		t.Errorf("Log file does not contain proper timestamp. Got:\n%s", contentStr)
	}

	// Test 2: Log multiple errors (should append)
	testErr2 := errors.New("second error message")
	testPayload2 := []byte(`{"another": "payload"}`)

	LogError(testErr2, testPayload2)

	// Verify both entries are in the log
	content, err = os.ReadFile(testLogPath)
	if err != nil {
		t.Fatalf("Failed to read log file after second log: %v", err)
	}

	contentStr = string(content)

	if !strings.Contains(contentStr, "test error message") {
		t.Errorf("First error message missing after second log")
	}

	if !strings.Contains(contentStr, "second error message") {
		t.Errorf("Second error message missing")
	}

	if !strings.Contains(contentStr, `"test": "payload"`) || !strings.Contains(contentStr, `"data": 123`) {
		t.Errorf("First payload missing after second log")
	}

	if !strings.Contains(contentStr, `"another": "payload"`) {
		t.Errorf("Second payload missing")
	}

	// Test 3: Log with nil/empty payload
	LogError(errors.New("empty payload test"), nil)

	content, err = os.ReadFile(testLogPath)
	if err != nil {
		t.Fatalf("Failed to read log file after empty payload log: %v", err)
	}

	contentStr = string(content)

	if !strings.Contains(contentStr, "empty payload test") {
		t.Errorf("Empty payload error message missing")
	}

	if !strings.Contains(contentStr, "<empty or nil payload>") {
		t.Errorf("Empty payload marker missing")
	}

	// Test 4: Log with nil error (should be no-op)
	initialSize, err := os.Stat(testLogPath)
	if err != nil {
		t.Fatalf("Failed to stat log file: %v", err)
	}

	LogError(nil, []byte("should not log"))

	afterSize, err := os.Stat(testLogPath)
	if err != nil {
		t.Fatalf("Failed to stat log file after nil error: %v", err)
	}

	if initialSize.Size() != afterSize.Size() {
		t.Errorf("Log file size changed when logging nil error")
	}
}

func TestLogErrorInvalidPath(t *testing.T) {
	// Test that logging to an invalid path doesn't panic
	// This should silently fail
	invalidPath := "/root/nonexistent/path/that/cannot/be/created/collector.log"
	SetLogFile(invalidPath)

	// This should not panic
	LogError(errors.New("test error"), []byte("test payload"))

	// Reset to a valid path
	tempDir := t.TempDir()
	SetLogFile(filepath.Join(tempDir, "test.log"))
}

func TestLogErrorConcurrent(t *testing.T) {
	// Test concurrent logging (simulating multiple goroutines)
	tempDir := t.TempDir()
	testLogPath := filepath.Join(tempDir, "concurrent-test.log")
	SetLogFile(testLogPath)

	// Clean up
	os.Remove(testLogPath)

	// Log from multiple goroutines
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			for j := 0; j < 10; j++ {
				LogError(errors.New("concurrent error"), []byte(`{"index": `+string(rune('0'+idx))+`}`))
				time.Sleep(time.Millisecond)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify log file exists
	content, err := os.ReadFile(testLogPath)
	if err != nil {
		t.Fatalf("Failed to read concurrent log file: %v", err)
	}

	// Should have 100 entries
	contentStr := string(content)
	count := strings.Count(contentStr, "concurrent error")
	if count < 100 {
		t.Errorf("Expected at least 100 log entries, got %d", count)
	}
}
