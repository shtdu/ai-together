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

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"switch-server/models"
)

func TestAnalyticsHandler_GetProviderAnalytics_Success(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testResult := map[string]interface{}{
		"providers": []map[string]interface{}{
			{"name": "anthropic", "total_tokens": 5000, "request_count": 50},
		},
		"summary": map[string]interface{}{
			"total_tokens":   5000,
			"total_requests": 50,
		},
	}

	mockUsageService.On("GetProviderAnalytics",
		mock.Anything,
		int64(1), "2024-01-01", "2024-01-31",
		[]string(nil), []string(nil), []string(nil)).
		Return(testResult, nil)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/analytics/providers?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.NotNil(t, response["providers"])

	mockUsageService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetProviderAnalytics_MissingDates(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	tests := []struct {
		name        string
		url         string
		expectError string
	}{
		{
			name:        "missing start_date",
			url:         "/analytics/providers?end_date=2024-01-31",
			expectError: "start_date and end_date are required",
		},
		{
			name:        "missing end_date",
			url:         "/analytics/providers?start_date=2024-01-01",
			expectError: "start_date and end_date are required",
		},
		{
			name:        "missing both dates",
			url:         "/analytics/providers",
			expectError: "start_date and end_date are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, setUserContext(req, managerUser))

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			json.NewDecoder(w.Body).Decode(&response)
			assert.Equal(t, tt.expectError, response["error"])
		})
	}
}

func TestAnalyticsHandler_GetProviderAnalytics_AccessDenied_Member(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	memberUser := createTestUser(1, "member@example.com", "Member", "member", 1)

	req, _ := http.NewRequest("GET", "/analytics/providers?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, memberUser))

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Access denied", response["error"])
}

func TestAnalyticsHandler_GetProviderAnalytics_WithFilters(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testResult := map[string]interface{}{
		"providers": []map[string]interface{}{
			{"name": "anthropic", "total_tokens": 3000},
		},
	}

	mockUsageService.On("GetProviderAnalytics",
		mock.Anything,
		int64(1), "2024-01-01", "2024-01-31",
		[]string{"anthropic"}, []string{"claude-3-opus"}, []string(nil)).
		Return(testResult, nil)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/analytics/providers?start_date=2024-01-01&end_date=2024-01-31&providers=anthropic&models=claude-3-opus", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusOK, w.Code)

	mockUsageService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetUserAnalytics_Success(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testResult := map[string]interface{}{
		"users": []map[string]interface{}{
			{"user_id": int64(1), "name": "User 1", "total_tokens": 1000},
		},
	}

	mockUsageService.On("GetUserAnalytics",
		mock.Anything,
		int64(1), "2024-01-01", "2024-01-31",
		[]int64{1, 2}, []string(nil), []string(nil)).
		Return(testResult, nil)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/analytics/users?start_date=2024-01-01&end_date=2024-01-31&user_ids=1,2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusOK, w.Code)

	mockUsageService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetHistory_Success(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testResult := map[string]interface{}{
		"records": []map[string]interface{}{
			{"id": int64(1), "model": "claude-3-opus", "total_tokens": 1000},
		},
		"pagination": map[string]interface{}{
			"page":  1,
			"limit": 50,
			"total": 1,
		},
	}

	mockUsageService.On("GetHistory",
		mock.Anything,
		int64(1), "2024-01-01", "2024-01-31",
		1, 50,
		[]int64(nil), []string(nil), []string(nil), []string(nil),
		"created_at", "desc").
		Return(testResult, nil)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/analytics/history?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusOK, w.Code)

	mockUsageService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetHistory_CustomPagination(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testResult := map[string]interface{}{
		"records": []map[string]interface{}{},
		"pagination": map[string]interface{}{
			"page":  2,
			"limit": 25,
			"total": 0,
		},
	}

	mockUsageService.On("GetHistory",
		mock.Anything,
		int64(1), "2024-01-01", "2024-01-31",
		2, 25,
		[]int64{1}, []string{"anthropic"}, []string{"claude-3-opus"}, []string(nil),
		"total_tokens", "asc").
		Return(testResult, nil)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/analytics/history?start_date=2024-01-01&end_date=2024-01-31&page=2&limit=25&user_ids=1&providers=anthropic&models=claude-3-opus&sort_by=total_tokens&sort_order=asc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusOK, w.Code)

	mockUsageService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetHistory_InvalidPage(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testResult := map[string]interface{}{
		"records": []map[string]interface{}{},
		"pagination": map[string]interface{}{
			"page":  1,
			"limit": 50,
			"total": 0,
		},
	}

	// Invalid page (0) should default to 1
	mockUsageService.On("GetHistory",
		mock.Anything,
		int64(1), "2024-01-01", "2024-01-31",
		1, 50, // Default page=1 when page<1
		[]int64(nil), []string(nil), []string(nil), []string(nil),
		"created_at", "desc").
		Return(testResult, nil)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/analytics/history?start_date=2024-01-01&end_date=2024-01-31&page=0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusOK, w.Code)

	mockUsageService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetFilterOptions_Success(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testResult := map[string]interface{}{
		"providers": []string{"anthropic", "openai"},
		"models":    []string{"claude-3-opus", "gpt-4"},
		"users": []map[string]interface{}{
			{"id": int64(1), "name": "User 1", "email": "user1@example.com"},
		},
	}

	mockUsageService.On("GetFilterOptions",
		mock.Anything,
		int64(1)).
		Return(testResult, nil)

	_ = NewAnalyticsHandler(mockUsageService, mockUserService)

	// Both manager and member can access filter options
	tests := []struct {
		name string
		user *models.User
	}{
		{
			name: "manager access",
			user: createTestUser(1, "manager@example.com", "Manager", "manager", 1),
		},
		{
			name: "member access",
			user: createTestUser(1, "member@example.com", "Member", "member", 1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsageService := new(MockUsageService)
			mockUserService := new(MockUserService)

			mockUsageService.On("GetFilterOptions",
				mock.Anything,
				int64(1)).
				Return(testResult, nil)

			handler := NewAnalyticsHandler(mockUsageService, mockUserService)
			router := setupAnalyticsRouter(handler)

			req, _ := http.NewRequest("GET", "/analytics/filter-options", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, setUserContext(req, tt.user))

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.NotNil(t, response["providers"])
			assert.NotNil(t, response["models"])

			mockUsageService.AssertExpectations(t)
		})
	}
}

func TestNewAnalyticsHandler(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.usageService)
	assert.NotNil(t, handler.userService)
}

// Helper function to set up analytics router
func setupAnalyticsRouter(handler *AnalyticsHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

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

	router.GET("/analytics/providers", handler.GetProviderAnalytics)
	router.GET("/analytics/users", handler.GetUserAnalytics)
	router.GET("/analytics/history", handler.GetHistory)
	router.GET("/analytics/filter-options", handler.GetFilterOptions)
	router.GET("/analytics/personal", handler.GetPersonalAnalytics)
	router.GET("/analytics/personal/history", handler.GetPersonalHistory)
	router.GET("/analytics/personal/filters", handler.GetPersonalFilterOptions)

	return router
}

// TestAnalyticsHandler_GetUserAnalytics_AccessDenied_Member tests member access to user analytics
func TestAnalyticsHandler_GetUserAnalytics_AccessDenied_Member(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	memberUser := createTestUser(1, "member@example.com", "Member", "member", 1)

	req, _ := http.NewRequest("GET", "/analytics/users?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, memberUser))

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Access denied", response["error"])
}

// TestAnalyticsHandler_GetUserAnalytics_MissingDates tests user analytics without dates
func TestAnalyticsHandler_GetUserAnalytics_MissingDates(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/analytics/users", nil) // Missing dates
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "start_date and end_date are required", response["error"])
}

// TestAnalyticsHandler_GetHistory_AccessDenied_Member tests member access to history
func TestAnalyticsHandler_GetHistory_AccessDenied_Member(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	memberUser := createTestUser(1, "member@example.com", "Member", "member", 1)

	req, _ := http.NewRequest("GET", "/analytics/history?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, memberUser))

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Access denied", response["error"])
}

// TestAnalyticsHandler_GetHistory_MissingDates tests history without dates
func TestAnalyticsHandler_GetHistory_MissingDates(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/analytics/history", nil) // Missing dates
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "start_date and end_date are required", response["error"])
}

// TestAnalyticsHandler_GetFilterOptions_MissingUserContext tests filter options without user
func TestAnalyticsHandler_GetFilterOptions_MissingUserContext(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/analytics/filter-options", handler.GetFilterOptions)

	req, _ := http.NewRequest("GET", "/analytics/filter-options", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "User not found in context", response["error"])
}

// TestNewAnalyticsHandler_VerifyInitialization tests handler initialization
func TestNewAnalyticsHandler_VerifyInitialization(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.usageService)
	assert.NotNil(t, handler.userService)
}

// --- Tool dimension tests (issue #104) ---

func TestAnalyticsHandler_GetProviderAnalytics_WithToolsFilter(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testResult := map[string]interface{}{
		"summary": map[string]interface{}{
			"total_tokens":   3000,
			"total_requests": 30,
		},
		"distribution": map[string]interface{}{
			"by_tool": []map[string]interface{}{
				{"name": "claude", "tokens": 3000, "requests": 30, "percentage": 100.0},
			},
		},
	}

	mockUsageService.On("GetProviderAnalytics",
		mock.Anything,
		int64(1), "2024-01-01", "2024-01-31",
		[]string(nil), []string(nil), []string{"claude"}).
		Return(testResult, nil)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/analytics/providers?start_date=2024-01-01&end_date=2024-01-31&tools=claude", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	dist := response["distribution"].(map[string]interface{})
	byTool := dist["by_tool"].([]interface{})
	assert.Len(t, byTool, 1)
	assert.Equal(t, "claude", byTool[0].(map[string]interface{})["name"])

	mockUsageService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetProviderAnalytics_WithMultipleToolsFilter(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testResult := map[string]interface{}{
		"summary": map[string]interface{}{
			"total_tokens": 9000,
		},
	}

	mockUsageService.On("GetProviderAnalytics",
		mock.Anything,
		int64(1), "2024-01-01", "2024-01-31",
		[]string(nil), []string(nil), []string{"claude", "codex"}).
		Return(testResult, nil)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/analytics/providers?start_date=2024-01-01&end_date=2024-01-31&tools=claude,codex", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusOK, w.Code)
	mockUsageService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetUserAnalytics_WithToolsFilter(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testResult := map[string]interface{}{
		"leaderboard": []map[string]interface{}{
			{"user_id": int64(1), "name": "User 1", "total_tokens": 1000},
		},
	}

	mockUsageService.On("GetUserAnalytics",
		mock.Anything,
		int64(1), "2024-01-01", "2024-01-31",
		[]int64(nil), []string(nil), []string{"opencode"}).
		Return(testResult, nil)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/analytics/users?start_date=2024-01-01&end_date=2024-01-31&tools=opencode", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusOK, w.Code)
	mockUsageService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetHistory_WithToolsFilter(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testResult := map[string]interface{}{
		"records": []map[string]interface{}{
			{"id": int64(1), "model": "claude-3-opus", "platform": "claude"},
		},
		"pagination": map[string]interface{}{
			"page":  1,
			"limit": 50,
			"total": 1,
		},
	}

	mockUsageService.On("GetHistory",
		mock.Anything,
		int64(1), "2024-01-01", "2024-01-31",
		1, 50,
		[]int64(nil), []string(nil), []string(nil), []string{"codex"},
		"created_at", "desc").
		Return(testResult, nil)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/analytics/history?start_date=2024-01-01&end_date=2024-01-31&tools=codex", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	records := response["records"].([]interface{})
	assert.Len(t, records, 1)
	assert.Equal(t, "claude", records[0].(map[string]interface{})["platform"])

	mockUsageService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetFilterOptions_IncludesTools(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testResult := map[string]interface{}{
		"providers": []string{"anthropic"},
		"models":    []string{"claude-3-opus"},
		"tools":     []string{"claude", "codex", "opencode"},
		"users":     []interface{}{},
	}

	mockUsageService.On("GetFilterOptions",
		mock.Anything,
		int64(1)).
		Return(testResult, nil)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/analytics/filter-options", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	tools := response["tools"].([]interface{})
	assert.Len(t, tools, 3)
	assert.Equal(t, "claude", tools[0])
	assert.Equal(t, "codex", tools[1])
	assert.Equal(t, "opencode", tools[2])

	mockUsageService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetHistory_WithAllFiltersIncludingTools(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testResult := map[string]interface{}{
		"records": []map[string]interface{}{},
		"pagination": map[string]interface{}{
			"page":  1,
			"limit": 25,
			"total": 0,
		},
	}

	mockUsageService.On("GetHistory",
		mock.Anything,
		int64(1), "2024-01-01", "2024-01-31",
		1, 25,
		[]int64{1}, []string{"anthropic"}, []string{"claude-3-opus"}, []string{"claude"},
		"total_tokens", "asc").
		Return(testResult, nil)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/analytics/history?start_date=2024-01-01&end_date=2024-01-31&page=1&limit=25&user_ids=1&providers=anthropic&models=claude-3-opus&tools=claude&sort_by=total_tokens&sort_order=asc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusOK, w.Code)
	mockUsageService.AssertExpectations(t)
}

// --- Personal analytics tests (issue #101) ---

func TestAnalyticsHandler_GetPersonalAnalytics_Success(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testResult := map[string]interface{}{
		"summary": map[string]interface{}{
			"total_tokens":   5000,
			"total_requests": 50,
		},
		"distribution": map[string]interface{}{
			"by_provider": []map[string]interface{}{
				{"name": "anthropic", "tokens": 5000, "requests": 50, "percentage": 100.0},
			},
			"by_model": []map[string]interface{}{
				{"name": "claude-3-opus", "tokens": 5000, "requests": 50, "percentage": 100.0},
			},
		},
	}

	mockUsageService.On("GetPersonalAnalytics",
		mock.Anything,
		int64(42), int64(1), "2024-01-01", "2024-01-31",
		[]string(nil), []string(nil), []string(nil)).
		Return(testResult, nil)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	// Both member and manager can access personal analytics
	memberUser := createTestUser(42, "member@example.com", "Member", "member", 1)

	req, _ := http.NewRequest("GET", "/analytics/personal?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, memberUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.NotNil(t, response["summary"])

	mockUsageService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetPersonalAnalytics_ManagerCanAccess(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testResult := map[string]interface{}{
		"summary": map[string]interface{}{
			"total_tokens":   1000,
			"total_requests": 10,
		},
	}

	mockUsageService.On("GetPersonalAnalytics",
		mock.Anything,
		int64(1), int64(1), "2024-01-01", "2024-01-31",
		[]string(nil), []string(nil), []string(nil)).
		Return(testResult, nil)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/analytics/personal?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusOK, w.Code)
	mockUsageService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetPersonalAnalytics_MissingDates(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	memberUser := createTestUser(42, "member@example.com", "Member", "member", 1)

	req, _ := http.NewRequest("GET", "/analytics/personal", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, memberUser))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "start_date and end_date are required", response["error"])
}

func TestAnalyticsHandler_GetPersonalAnalytics_WithFilters(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testResult := map[string]interface{}{
		"summary": map[string]interface{}{"total_tokens": 3000},
	}

	mockUsageService.On("GetPersonalAnalytics",
		mock.Anything,
		int64(42), int64(1), "2024-01-01", "2024-01-31",
		[]string{"anthropic"}, []string{"claude-3-opus"}, []string{"claude"}).
		Return(testResult, nil)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	memberUser := createTestUser(42, "member@example.com", "Member", "member", 1)

	req, _ := http.NewRequest("GET", "/analytics/personal?start_date=2024-01-01&end_date=2024-01-31&providers=anthropic&models=claude-3-opus&tools=claude", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, memberUser))

	assert.Equal(t, http.StatusOK, w.Code)
	mockUsageService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetPersonalAnalytics_Unauthorized(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/analytics/personal", handler.GetPersonalAnalytics)

	req, _ := http.NewRequest("GET", "/analytics/personal?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAnalyticsHandler_GetPersonalHistory_Success(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testResult := map[string]interface{}{
		"records": []map[string]interface{}{
			{"id": int64(1), "model": "claude-3-opus", "total_tokens": 1000},
		},
		"pagination": map[string]interface{}{
			"page":  1,
			"limit": 100,
			"total": 1,
		},
	}

	mockUsageService.On("GetPersonalHistory",
		mock.Anything,
		int64(42), int64(1), "2024-01-01", "2024-01-31",
		1, 100,
		[]string(nil), []string(nil), []string(nil),
		"created_at", "desc").
		Return(testResult, nil)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	memberUser := createTestUser(42, "member@example.com", "Member", "member", 1)

	req, _ := http.NewRequest("GET", "/analytics/personal/history?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, memberUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.NotNil(t, response["records"])

	mockUsageService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetPersonalHistory_MissingDates(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	memberUser := createTestUser(42, "member@example.com", "Member", "member", 1)

	req, _ := http.NewRequest("GET", "/analytics/personal/history", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, memberUser))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyticsHandler_GetPersonalHistory_WithPagination(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testResult := map[string]interface{}{
		"records": []map[string]interface{}{},
		"pagination": map[string]interface{}{
			"page":  2,
			"limit": 25,
			"total": 100,
		},
	}

	mockUsageService.On("GetPersonalHistory",
		mock.Anything,
		int64(42), int64(1), "2024-01-01", "2024-01-31",
		2, 25,
		[]string{"anthropic"}, []string(nil), []string(nil),
		"total_tokens", "asc").
		Return(testResult, nil)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	memberUser := createTestUser(42, "member@example.com", "Member", "member", 1)

	req, _ := http.NewRequest("GET", "/analytics/personal/history?start_date=2024-01-01&end_date=2024-01-31&page=2&limit=25&providers=anthropic&sort_by=total_tokens&sort_order=asc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, memberUser))

	assert.Equal(t, http.StatusOK, w.Code)
	mockUsageService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetPersonalFilterOptions_Success(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	testResult := map[string]interface{}{
		"providers": []string{"anthropic"},
		"models":    []string{"claude-3-opus"},
		"tools":     []string{"claude"},
	}

	mockUsageService.On("GetPersonalFilterOptions",
		mock.Anything,
		int64(42), int64(1)).
		Return(testResult, nil)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)
	router := setupAnalyticsRouter(handler)

	memberUser := createTestUser(42, "member@example.com", "Member", "member", 1)

	req, _ := http.NewRequest("GET", "/analytics/personal/filters", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, memberUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.NotNil(t, response["providers"])
	assert.NotNil(t, response["models"])
	assert.NotNil(t, response["tools"])

	mockUsageService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetPersonalFilterOptions_Unauthorized(t *testing.T) {
	mockUsageService := new(MockUsageService)
	mockUserService := new(MockUserService)

	handler := NewAnalyticsHandler(mockUsageService, mockUserService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/analytics/personal/filters", handler.GetPersonalFilterOptions)

	req, _ := http.NewRequest("GET", "/analytics/personal/filters", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
