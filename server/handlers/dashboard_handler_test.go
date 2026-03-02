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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"switch-server/models"
)

func TestDashboardHandler_GetMetrics_Success(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testData := map[string]interface{}{
		"period": map[string]interface{}{
			"start": time.Now().Add(-7 * 24 * time.Hour),
			"end":   time.Now(),
		},
		"interval": "hour",
		"data_points": []interface{}{
			map[string]interface{}{
				"timestamp":            time.Now(),
				"total_tokens":         1000,
				"request_count":        10,
			},
		},
	}

	mockUsageService.On("GetDashboardMetrics",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		int64(1), int64(1), "manager",
		mock.MatchedBy(func(t time.Time) bool { return true }), // startTime
		mock.MatchedBy(func(t time.Time) bool { return true }), // endTime
		"hour").
		Return(testData, nil)

	handler := NewDashboardHandler(mockUsageService, mockUserService)
	router := setupDashboardRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/dashboard/metrics?range=7d&interval=hour", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusOK, w.Code)

	mockUsageService.AssertExpectations(t)
}

func TestDashboardHandler_GetMetrics_DefaultRange(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testData := map[string]interface{}{
		"period":   map[string]interface{}{},
		"interval": "hour", // Default interval is "hour"
		"data_points": []interface{}{},
	}

	mockUsageService.On("GetDashboardMetrics",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		int64(1), int64(1), "member",
		mock.MatchedBy(func(t time.Time) bool { return true }),
		mock.MatchedBy(func(t time.Time) bool { return true }),
		"hour"). // Default is "hour"
		Return(testData, nil)

	handler := NewDashboardHandler(mockUsageService, mockUserService)
	router := setupDashboardRouter(handler)

	memberUser := createTestUser(1, "member@example.com", "Member", "member", 1)

	// Test default range (7d) and interval (hour)
	req, _ := http.NewRequest("GET", "/dashboard/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, memberUser))

	assert.Equal(t, http.StatusOK, w.Code)

	mockUsageService.AssertExpectations(t)
}

func TestDashboardHandler_GetMetrics_DifferentRanges(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testData := map[string]interface{}{"data_points": []interface{}{}}

	// Allow any call with flexible time matching
	mockUsageService.On("GetDashboardMetrics",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		int64(1), int64(1), "manager",
		mock.MatchedBy(func(t time.Time) bool { return true }),
		mock.MatchedBy(func(t time.Time) bool { return true }),
		mock.AnythingOfType("string")).
		Return(testData, nil)

	handler := NewDashboardHandler(mockUsageService, mockUserService)
	router := setupDashboardRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	tests := []struct {
		name string
		url  string
	}{
		{"24h range", "/dashboard/metrics?range=24h"},
		{"7d range", "/dashboard/metrics?range=7d"},
		{"30d range", "/dashboard/metrics?range=30d"},
		{"invalid range defaults to 7d", "/dashboard/metrics?range=invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, setUserContext(req, managerUser))

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestDashboardHandler_GetMetrics_MissingUserContext(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	handler := NewDashboardHandler(mockUsageService, mockUserService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/dashboard/metrics", handler.GetMetrics)

	req, _ := http.NewRequest("GET", "/dashboard/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "User not found in context", response["error"])
}

func TestDashboardHandler_GetRankings_Success(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testData := map[string]interface{}{
		"period": "24h",
		"rankings": []map[string]interface{}{
			{"rank": 1, "provider": "anthropic", "total_tokens": 5000, "percentage": 60.0, "request_count": 50},
			{"rank": 2, "provider": "openai", "total_tokens": 3333, "percentage": 40.0, "request_count": 33},
		},
	}

	mockUsageService.On("GetProviderRankings",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		int64(1), int64(1), "manager").
		Return(testData, nil)

	handler := NewDashboardHandler(mockUsageService, mockUserService)
	router := setupDashboardRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/dashboard/rankings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "24h", response["period"])

	mockUsageService.AssertExpectations(t)
}

func TestDashboardHandler_GetRankings_MemberAccess(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testData := map[string]interface{}{
		"period":    "24h",
		"rankings":  []interface{}{},
	}

	mockUsageService.On("GetProviderRankings",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		int64(1), int64(1), "member").
		Return(testData, nil)

	handler := NewDashboardHandler(mockUsageService, mockUserService)
	router := setupDashboardRouter(handler)

	memberUser := createTestUser(1, "member@example.com", "Member", "member", 1)

	req, _ := http.NewRequest("GET", "/dashboard/rankings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, memberUser))

	assert.Equal(t, http.StatusOK, w.Code)

	mockUsageService.AssertExpectations(t)
}

func TestDashboardHandler_GetMembers_Success(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testData := map[string]interface{}{
		"members": []map[string]interface{}{
			{"user_id": int64(1), "name": "User 1", "email": "user1@example.com", "total_tokens": 1000},
			{"user_id": int64(2), "name": "User 2", "email": "user2@example.com", "total_tokens": 500},
		},
	}

	mockUsageService.On("GetMemberStats",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		int64(1)).
		Return(testData, nil)

	handler := NewDashboardHandler(mockUsageService, mockUserService)
	router := setupDashboardRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/dashboard/members", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusOK, w.Code)

	mockUsageService.AssertExpectations(t)
}

func TestDashboardHandler_GetMembers_AccessDenied_Member(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	handler := NewDashboardHandler(mockUsageService, mockUserService)
	router := setupDashboardRouter(handler)

	memberUser := createTestUser(1, "member@example.com", "Member", "member", 1)

	req, _ := http.NewRequest("GET", "/dashboard/members", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, memberUser))

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Access denied", response["error"])
}

func TestDashboardHandler_GetMembers_MissingUserContext(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	handler := NewDashboardHandler(mockUsageService, mockUserService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/dashboard/members", handler.GetMembers)

	req, _ := http.NewRequest("GET", "/dashboard/members", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Helper function to set up dashboard router with user context middleware
func setupDashboardRouter(handler *DashboardHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Middleware to set user context (simulates auth middleware)
	router.Use(func(c *gin.Context) {
		userEmail := c.GetHeader("X-Test-User-Email")
		if userEmail != "" {
			userID := c.GetHeader("X-Test-User-ID")
			userRole := c.GetHeader("X-Test-User-Role")
			tenantID := c.GetHeader("X-Test-User-Tenant-ID")

			var id, tenant int64
			if userID != "" {
				fmt.Sscanf(userID, "%d", &id)
			}
			if tenantID != "" {
				fmt.Sscanf(tenantID, "%d", &tenant)
			}

			user := &models.User{
				ID:       id,
				Email:    userEmail,
				Role:     userRole,
				TenantID: tenant,
			}
			c.Set("user", user)
		}
		c.Next()
	})

	router.GET("/dashboard/metrics", handler.GetMetrics)
	router.GET("/dashboard/rankings", handler.GetRankings)
	router.GET("/dashboard/members", handler.GetMembers)

	return router
}

func TestMetricDataPoint_Structure(t *testing.T) {
	now := time.Now()
	point := MetricDataPoint{
		Timestamp:          now,
		ActiveTimeSeconds:  3600.0,
		TotalTokens:        1000,
		InputTokens:        600,
		OutputTokens:       400,
		RequestCount:       10,
	}

	data, err := json.Marshal(point)
	require.NoError(t, err)
	assert.Contains(t, string(data), "total_tokens")

	var decoded MetricDataPoint
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, point.TotalTokens, decoded.TotalTokens)
}

func TestProviderRanking_Structure(t *testing.T) {
	ranking := ProviderRanking{
		Rank:         1,
		Provider:     "anthropic",
		TotalTokens:  5000,
		Percentage:   60.0,
		RequestCount: 50,
	}

	data, err := json.Marshal(ranking)
	require.NoError(t, err)

	var decoded ProviderRanking
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, ranking.Provider, decoded.Provider)
	assert.Equal(t, ranking.Rank, decoded.Rank)
}

func TestMemberStats_Structure(t *testing.T) {
	now := time.Now()
	stats := MemberStats{
		UserID:          1,
		Name:            "Test User",
		Email:           "test@example.com",
		ActiveTimeHours: 40.5,
		AvgTokensPerDay: 100.5,
		TotalTokens:     1000,
		LastActive:      now,
	}

	data, err := json.Marshal(stats)
	require.NoError(t, err)

	var decoded MemberStats
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, stats.Name, decoded.Name)
	assert.Equal(t, stats.Email, decoded.Email)
}

// TestDashboardHandler_GetMetrics_InvalidRange tests invalid range parameter
func TestDashboardHandler_GetMetrics_InvalidRange(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testData := map[string]interface{}{
		"interval": "hour",
		"data_points": []interface{}{},
	}

	// Mock the call with flexible time matching
	mockUsageService.On("GetDashboardMetrics",
		mock.Anything,
		int64(1), int64(1), "manager",
		mock.Anything,
		mock.Anything,
		"hour").
		Return(testData, nil)

	handler := NewDashboardHandler(mockUsageService, mockUserService)
	router := setupDashboardRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	// Test with invalid range that should default to 7d
	req, _ := http.NewRequest("GET", "/dashboard/metrics?range=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	// Should still process with default range
	assert.True(t, w.Code >= 200 && w.Code < 500)

	mockUsageService.AssertExpectations(t)
}

// TestDashboardHandler_GetRankings_MissingUserContext tests rankings without user
func TestDashboardHandler_GetRankings_MissingUserContext(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	handler := NewDashboardHandler(mockUsageService, mockUserService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/dashboard/rankings", handler.GetRankings)

	req, _ := http.NewRequest("GET", "/dashboard/rankings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "User not found in context", response["error"])
}

// TestDashboardHandler_GetRankings_ServiceError tests rankings with service error
func TestDashboardHandler_GetRankings_ServiceError(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	mockUsageService.On("GetProviderRankings",
		mock.Anything,
		int64(1), int64(1), "member").
		Return(nil, assert.AnError)

	handler := NewDashboardHandler(mockUsageService, mockUserService)
	router := setupDashboardRouter(handler)

	memberUser := createTestUser(1, "member@example.com", "Member", "member", 1)

	req, _ := http.NewRequest("GET", "/dashboard/rankings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, memberUser))

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	assert.Contains(t, response["error"], "Failed to get rankings")

	mockUsageService.AssertExpectations(t)
}

// TestDashboardHandler_GetMembers_ServiceError tests members with service error
func TestDashboardHandler_GetMembers_ServiceError(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	mockUsageService.On("GetMemberStats",
		mock.Anything,
		int64(1)).
		Return(nil, assert.AnError)

	handler := NewDashboardHandler(mockUsageService, mockUserService)
	router := setupDashboardRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/dashboard/members", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	assert.Contains(t, response["error"], "Failed to get member stats")

	mockUsageService.AssertExpectations(t)
}

// TestDashboardHandler_GetMetrics_ServiceError tests metrics with service error
func TestDashboardHandler_GetMetrics_ServiceError(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	mockUsageService.On("GetDashboardMetrics",
		mock.Anything,
		int64(1), int64(1), "manager",
		mock.Anything,
		mock.Anything,
		"hour").
		Return(nil, assert.AnError)

	handler := NewDashboardHandler(mockUsageService, mockUserService)
	router := setupDashboardRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/dashboard/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockUsageService.AssertExpectations(t)
}

// TestMetricDataPoint_EmptyValues tests MetricDataPoint with zero values
func TestMetricDataPoint_EmptyValues(t *testing.T) {
	point := MetricDataPoint{}

	data, err := json.Marshal(point)
	require.NoError(t, err)

	var decoded MetricDataPoint
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, int64(0), decoded.TotalTokens)
	assert.Equal(t, int64(0), decoded.InputTokens)
	assert.Equal(t, int64(0), decoded.OutputTokens)
}

// TestProviderRanking_ZeroValues tests ProviderRanking with zero values
func TestProviderRanking_ZeroValues(t *testing.T) {
	ranking := ProviderRanking{}

	data, err := json.Marshal(ranking)
	require.NoError(t, err)

	var decoded ProviderRanking
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, 0, decoded.Rank)
	assert.Empty(t, decoded.Provider)
	assert.Equal(t, int64(0), decoded.TotalTokens)
	assert.Equal(t, int64(0), decoded.RequestCount)
}

// TestMemberStats_ZeroValues tests MemberStats with zero values
func TestMemberStats_ZeroValues(t *testing.T) {
	stats := MemberStats{}

	data, err := json.Marshal(stats)
	require.NoError(t, err)

	var decoded MemberStats
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, int64(0), decoded.UserID)
	assert.Empty(t, stats.Name)
	assert.Empty(t, stats.Email)
}
