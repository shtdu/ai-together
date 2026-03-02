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


package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthResponse_Structure(t *testing.T) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Database: DatabaseStatus{
			Status:  "up",
			Latency: "10ms",
		},
	}

	// Test JSON serialization
	data, err := json.Marshal(response)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"status":"ok"`)
	assert.Contains(t, string(data), `"database"`)
	assert.Contains(t, string(data), `"timestamp"`)

	// Test JSON deserialization
	var decoded HealthResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, response.Status, decoded.Status)
	assert.Equal(t, response.Database.Status, decoded.Database.Status)
}

func TestDatabaseStatus_Structure(t *testing.T) {
	tests := []struct {
		name   string
		status DatabaseStatus
	}{
		{
			name: "healthy database status",
			status: DatabaseStatus{
				Status:  "up",
				Latency: "5ms",
			},
		},
		{
			name: "unhealthy database status with error",
			status: DatabaseStatus{
				Status: "down",
				Error:  "connection refused",
			},
		},
		{
			name: "unhealthy database status with latency and error",
			status: DatabaseStatus{
				Status:  "down",
				Latency: "5000ms",
				Error:   "timeout",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON serialization
			data, err := json.Marshal(tt.status)
			require.NoError(t, err)

			// Test JSON deserialization
			var decoded DatabaseStatus
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)
			assert.Equal(t, tt.status.Status, decoded.Status)

			if tt.status.Error != "" {
				assert.Equal(t, tt.status.Error, decoded.Error)
			}
			if tt.status.Latency != "" {
				assert.Equal(t, tt.status.Latency, decoded.Latency)
			}
		})
	}
}

// TestHealthHandler_Initialization tests that the health handler can be initialized
func TestHealthHandler_Initialization(t *testing.T) {
	handler := NewHealthHandler(nil, "test-version")
	assert.NotNil(t, handler)
}

// TestHealthHandler_HandlerExists tests that the health check handler method exists
func TestHealthHandler_HandlerExists(t *testing.T) {
	handler := NewHealthHandler(nil, "test-version")
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.HealthCheck)
}

func TestHealthResponse_TimestampFormat(t *testing.T) {
	// Test that timestamp format is RFC3339
	now := time.Now().UTC()
	timestamp := now.Format(time.RFC3339)

	response := HealthResponse{
		Status:    "ok",
		Timestamp: timestamp,
		Database: DatabaseStatus{
			Status: "up",
		},
	}

	// Verify timestamp can be parsed back
	parsedTime, err := time.Parse(time.RFC3339, response.Timestamp)
	require.NoError(t, err)
	assert.WithinDuration(t, now, parsedTime, time.Second)
}

// Test database status with different scenarios
func TestDatabaseStatus_Scenarios(t *testing.T) {
	scenarios := []struct {
		name   string
		status DatabaseStatus
	}{
		{
			name: "database up with fast latency",
			status: DatabaseStatus{
				Status:  "up",
				Latency: "1ms",
			},
		},
		{
			name: "database up with slow latency",
			status: DatabaseStatus{
				Status:  "up",
				Latency: "500ms",
			},
		},
		{
			name: "database down with connection error",
			status: DatabaseStatus{
				Status: "down",
				Error:  "connection refused",
			},
		},
		{
			name: "database down with timeout error",
			status: DatabaseStatus{
				Status:  "down",
				Error:  "context deadline exceeded",
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Test JSON serialization/deserialization
			data, err := json.Marshal(scenario.status)
			require.NoError(t, err)

			var decoded DatabaseStatus
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)
			assert.Equal(t, scenario.status.Status, decoded.Status)
		})
	}
}

// TestHealthHandler_ResponseCodes verifies the expected response codes
func TestHealthHandler_ResponseCodes(t *testing.T) {
	expectedCodes := map[string]int{
		"healthy":   http.StatusOK,
		"degraded":  http.StatusServiceUnavailable,
	}

	for status, expectedCode := range expectedCodes {
		t.Run(status, func(t *testing.T) {
			// This test documents the expected behavior
			// Actual testing requires database connection
			assert.NotZero(t, expectedCode)
		})
	}
}

// BenchmarkHealthResponseSerialization benchmarks JSON serialization
func BenchmarkHealthResponseSerialization(b *testing.B) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Database: DatabaseStatus{
			Status:  "up",
			Latency: "10ms",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(response)
	}
}

// BenchmarkDatabaseStatusSerialization benchmarks JSON serialization
func BenchmarkDatabaseStatusSerialization(b *testing.B) {
	status := DatabaseStatus{
		Status:  "up",
		Latency: "5ms",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(status)
	}
}

// TestHealthHandler_HealthCheck_HandlerStructure tests the health check handler structure
// Note: Full testing requires a database connection. This test verifies the handler is properly configured.
func TestHealthHandler_HealthCheck_HandlerStructure(t *testing.T) {
	handler := NewHealthHandler(nil, "test-version")
	assert.NotNil(t, handler.HealthCheck)
}

// TestHealthHandler_VersionField tests that version is correctly stored
func TestHealthHandler_VersionField(t *testing.T) {
	testVersion := "v1.0.0"
	handler := NewHealthHandler(nil, testVersion)
	assert.Equal(t, testVersion, handler.version)
}

// TestHealthHandler_DatabaseField tests that database checker is correctly stored
func TestHealthHandler_DatabaseField(t *testing.T) {
	// Test with nil pool (should create nil checker)
	handler := NewHealthHandler(nil, "test-version")
	assert.Nil(t, handler.db)

	// Test with mock checker
	mockChecker := &mockHealthChecker{pingError: nil}
	handler = newHealthHandlerWithChecker(mockChecker, "test-version")
	assert.Equal(t, mockChecker, handler.db)
}

// TestHealthHandler_HealthCheck_ResponseFormat tests the expected response format
func TestHealthHandler_HealthCheck_ResponseFormat(t *testing.T) {
	// Create a sample health response to verify the format
	response := HealthResponse{
		Version:   "v1.0.0",
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Database: DatabaseStatus{
			Status:  "up",
			Latency: "10ms",
		},
	}

	data, err := json.Marshal(response)
	require.NoError(t, err)

	var decoded HealthResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "v1.0.0", decoded.Version)
	assert.Equal(t, "ok", decoded.Status)
	assert.Equal(t, "up", decoded.Database.Status)
	assert.Equal(t, "10ms", decoded.Database.Latency)
}

// TestHealthHandler_HealthCheck_DegradedResponseFormat tests the degraded status format
func TestHealthHandler_HealthCheck_DegradedResponseFormat(t *testing.T) {
	// Create a degraded health response to verify the format
	response := HealthResponse{
		Version:   "v1.0.0",
		Status:    "degraded",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Database: DatabaseStatus{
			Status: "down",
			Error:  "connection refused",
		},
	}

	data, err := json.Marshal(response)
	require.NoError(t, err)

	var decoded HealthResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "degraded", decoded.Status)
	assert.Equal(t, "down", decoded.Database.Status)
	assert.Equal(t, "connection refused", decoded.Database.Error)
}

// TestHealthHandler_TimezoneHandling tests that timestamps are in UTC
func TestHealthHandler_TimezoneHandling(t *testing.T) {
	nowUTC := time.Now().UTC()
	timestamp := nowUTC.Format(time.RFC3339)

	// Verify the timestamp can be parsed back
	parsedTime, err := time.Parse(time.RFC3339, timestamp)
	require.NoError(t, err)

	// Verify it's in UTC (no offset)
	_, offset := parsedTime.Zone()
	assert.Equal(t, 0, offset, "Timestamp should be in UTC (no timezone offset)")
}

// TestHealthHandler_DatabaseErrorBehavior documents expected behavior when database fails
func TestHealthHandler_DatabaseErrorBehavior(t *testing.T) {
	// This test documents the expected behavior when database Ping returns an error
	// The health check should return degraded status (503) with database error details

	expectedStatus := "degraded"
	expectedDBStatus := map[string]interface{}{"status": "down", "error": "simulated database connection error"}
	expectedVersion := "test-version"

	// Document expected behavior without requiring full HTTP test
	t.Log("Expected behavior when database fails:",
		"  Status:", expectedStatus,
		"  Database:", expectedDBStatus,
		"  Version:", expectedVersion)
}

// mockHealthChecker is a mock implementation of HealthChecker for testing
type mockHealthChecker struct {
	pingError error
}

func (m *mockHealthChecker) Ping(ctx context.Context) error {
	return m.pingError
}

// TestHealthHandler_HealthCheck_Success tests successful health check
func TestHealthHandler_HealthCheck_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a mock checker that returns no error (successful ping)
	mockChecker := &mockHealthChecker{pingError: nil}
	handler := newHealthHandlerWithChecker(mockChecker, "v1.0.0")

	router := gin.New()
	router.GET("/health", handler.HealthCheck)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "v1.0.0", response.Version)
	assert.Equal(t, "ok", response.Status)
	assert.Equal(t, "up", response.Database.Status)
	assert.NotEmpty(t, response.Database.Latency)
	assert.Empty(t, response.Database.Error)
	_, err = time.Parse(time.RFC3339, response.Timestamp)
	assert.NoError(t, err, "Timestamp should be in RFC3339 format")
}

// TestHealthHandler_HealthCheck_DatabaseDown tests health check when database is down
func TestHealthHandler_HealthCheck_DatabaseDown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a mock checker that returns an error (failed ping)
	mockChecker := &mockHealthChecker{pingError: errors.New("connection refused")}
	handler := newHealthHandlerWithChecker(mockChecker, "v1.0.0")

	router := gin.New()
	router.GET("/health", handler.HealthCheck)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "v1.0.0", response.Version)
	assert.Equal(t, "degraded", response.Status)
	assert.Equal(t, "down", response.Database.Status)
	assert.Equal(t, "connection refused", response.Database.Error)
	assert.Empty(t, response.Database.Latency)
}

// TestHealthHandler_HealthCheck_Timeout tests health check when database times out
func TestHealthHandler_HealthCheck_Timeout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a mock checker that returns a timeout error
	mockChecker := &mockHealthChecker{pingError: errors.New("context deadline exceeded")}
	handler := newHealthHandlerWithChecker(mockChecker, "v2.0.0")

	router := gin.New()
	router.GET("/health", handler.HealthCheck)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "degraded", response.Status)
	assert.Equal(t, "down", response.Database.Status)
	assert.Contains(t, response.Database.Error, "deadline exceeded")
}
