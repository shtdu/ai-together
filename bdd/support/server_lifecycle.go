// Package support provides test utilities and context management for BDD tests
package support

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	// DefaultTestServerPort is the default port for the test server
	DefaultTestServerPort = 8088

	// DefaultHealthCheckPath is the default path for health checks
	DefaultHealthCheckPath = "/health"

	// DefaultTestServerTimeout is the default timeout for waiting for the server
	DefaultTestServerTimeout = 30 * time.Second

	// HealthCheckInterval is the interval between health checks
	HealthCheckInterval = 500 * time.Millisecond
)

// WaitForTestServer waits for the test server to be ready with a timeout
func WaitForTestServer(serverURL string, timeout time.Duration) error {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	healthURL := serverURL + DefaultHealthCheckPath
	start := time.Now()

	for time.Since(start) < timeout {
		resp, err := client.Get(healthURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}

		if resp != nil {
			resp.Body.Close()
		}

		time.Sleep(HealthCheckInterval)
	}

	return fmt.Errorf("test server not ready after %v", timeout)
}

// IsTestServerRunning checks if the test server is currently running
func IsTestServerRunning(serverURL string) bool {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	healthURL := serverURL + DefaultHealthCheckPath
	resp, err := client.Get(healthURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GetTestServerLogs retrieves the test server logs from the default location
// Returns the log file contents or an error
func GetTestServerLogs() ([]byte, error) {
	// Default log location for test server
	logPath := "/tmp/test-server.log"

	// Check if file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("test server log file not found: %s", logPath)
	}

	// Read file
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read test server logs: %w", err)
	}

	return data, nil
}

// GetTestServerLogPath returns the path to the test server log file
func GetTestServerLogPath() string {
	// Check environment variable first
	if path := os.Getenv("TEST_SERVER_LOG_PATH"); path != "" {
		return path
	}
	return "/tmp/test-server.log"
}

// StartTestServer starts the test server using the integration test script
// This is typically called from shell scripts, but can be called programmatically
func StartTestServer() error {
	// The test server is usually started via the integration/test-server.sh script
	// This function is provided for programmatic control if needed

	// Check if server is already running
	serverURL := GetTestServerURL()
	if IsTestServerRunning(serverURL) {
		return fmt.Errorf("test server is already running at %s", serverURL)
	}

	// In most cases, the test server should be started manually:
	// cd ../integration && ./test-server.sh

	return fmt.Errorf("test server should be started via shell script: cd ../integration && ./test-server.sh")
}

// StopTestServer stops the test server by PID
func StopTestServer(pidFile string) error {
	// Read PID from file
	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	// Parse PID (simplified - in production use proper PID parsing)
	var pid int
	_, err = fmt.Sscanf(string(pidData), "%d", &pid)
	if err != nil {
		return fmt.Errorf("failed to parse PID: %w", err)
	}

	// Send SIGTERM to process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	err = process.Signal(os.Interrupt)
	if err != nil {
		return fmt.Errorf("failed to send signal to process: %w", err)
	}

	return nil
}

// GetDefaultTestServerURL returns the default test server URL
func GetDefaultTestServerURL() string {
	return fmt.Sprintf("http://localhost:%d", DefaultTestServerPort)
}
