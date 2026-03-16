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
	"bytes"
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

func TestUsageHandler_GetCurrentUsage_Success(t *testing.T) {
	mockUsageService := new(MockUsageService)

	usageData := map[string]interface{}{
		"total_tokens":   1000,
		"total_cost":     0.50,
		"total_requests": 10,
	}

	mockUsageService.On("GetCurrentUsage", mock.Anything, int64(1), int64(1)).Return(usageData, nil)

	handler := NewUsageHandler(mockUsageService)
	router := setupUsageRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "member", 1)

	req, _ := http.NewRequest("GET", "/usage/current", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(1000), response["total_tokens"])

	mockUsageService.AssertExpectations(t)
}

func TestUsageHandler_GetCurrentUsage_MissingUserContext(t *testing.T) {
	mockUsageService := new(MockUsageService)

	handler := NewUsageHandler(mockUsageService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/usage/current", handler.GetCurrentUsage)

	req, _ := http.NewRequest("GET", "/usage/current", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "User not found in context", response["error"])
}

func TestUsageHandler_GetUsageStats_Success(t *testing.T) {
	mockUsageService := new(MockUsageService)

	statsData := map[string]interface{}{
		"total_tokens":   50000,
		"total_cost":     25.00,
		"total_requests": 500,
		"active_users":   10,
	}

	mockUsageService.On("GetUsageStats", mock.Anything, int64(1)).Return(statsData, nil)

	handler := NewUsageHandler(mockUsageService)
	router := setupUsageRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	req, _ := http.NewRequest("GET", "/usage/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(50000), response["total_tokens"])

	mockUsageService.AssertExpectations(t)
}

func TestUsageHandler_CreateBatchUsageRecords_Success(t *testing.T) {
	mockUsageService := new(MockUsageService)

	records := []models.UsageRecord{
		{ID: 1, Platform: "openai", Model: "gpt-4", TenantID: 1, UserID: 1},
		{ID: 2, Platform: "anthropic", Model: "claude-3", TenantID: 1, UserID: 1},
	}

	mockUsageService.On("CreateUsageRecord", mock.Anything, records[0]).Return(nil)
	mockUsageService.On("CreateUsageRecord", mock.Anything, records[1]).Return(nil)

	handler := NewUsageHandler(mockUsageService)
	router := setupUsageRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "member", 1)

	body, _ := json.Marshal(records)
	req, _ := http.NewRequest("POST", "/usage/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response BatchUsageResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 2, response.SyncedCount)
	assert.Len(t, response.Errors, 0)

	mockUsageService.AssertExpectations(t)
}

func TestUsageHandler_CreateBatchUsageRecords_EmptyRecords(t *testing.T) {
	mockUsageService := new(MockUsageService)

	handler := NewUsageHandler(mockUsageService)
	router := setupUsageRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "member", 1)

	records := []models.UsageRecord{}
	body, _ := json.Marshal(records)
	req, _ := http.NewRequest("POST", "/usage/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "No records provided", response["error"])
}

func TestUsageHandler_CreateBatchUsageRecords_ValidationError(t *testing.T) {
	mockUsageService := new(MockUsageService)

	handler := NewUsageHandler(mockUsageService)
	router := setupUsageRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "member", 1)

	req, _ := http.NewRequest("POST", "/usage/batch", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUsageHandler_CreateBatchUsageRecords_TenantIDMismatch(t *testing.T) {
	mockUsageService := new(MockUsageService)

	records := []models.UsageRecord{
		{ID: 1, Platform: "openai", Model: "gpt-4", TenantID: 999, UserID: 1}, // Different tenant
	}

	handler := NewUsageHandler(mockUsageService)
	router := setupUsageRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "member", 1)

	body, _ := json.Marshal(records)
	req, _ := http.NewRequest("POST", "/usage/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response BatchUsageResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 0, response.SyncedCount)
	assert.Len(t, response.Errors, 1)
	assert.Contains(t, response.Errors[0], "tenant_id mismatch")
}

func TestUsageHandler_CreateBatchUsageRecords_UserIDMismatch(t *testing.T) {
	mockUsageService := new(MockUsageService)

	records := []models.UsageRecord{
		{ID: 1, Platform: "openai", Model: "gpt-4", TenantID: 1, UserID: 999}, // Different user
	}

	handler := NewUsageHandler(mockUsageService)
	router := setupUsageRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "member", 1)

	body, _ := json.Marshal(records)
	req, _ := http.NewRequest("POST", "/usage/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response BatchUsageResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 0, response.SyncedCount)
	assert.Len(t, response.Errors, 1)
	assert.Contains(t, response.Errors[0], "user_id mismatch")
}

func TestUsageHandler_CreateBatchUsageRecords_PartialSuccess(t *testing.T) {
	mockUsageService := new(MockUsageService)

	records := []models.UsageRecord{
		{ID: 1, Platform: "openai", Model: "gpt-4", TenantID: 1, UserID: 1},
		{ID: 2, Platform: "anthropic", Model: "claude-3", TenantID: 1, UserID: 1},
		{ID: 3, Platform: "openai", Model: "gpt-3.5", TenantID: 999, UserID: 1}, // Wrong tenant
	}

	mockUsageService.On("CreateUsageRecord", mock.Anything, records[0]).Return(nil)
	mockUsageService.On("CreateUsageRecord", mock.Anything, records[1]).Return(nil)

	handler := NewUsageHandler(mockUsageService)
	router := setupUsageRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "member", 1)

	body, _ := json.Marshal(records)
	req, _ := http.NewRequest("POST", "/usage/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response BatchUsageResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 2, response.SyncedCount)
	assert.Len(t, response.Errors, 1)

	mockUsageService.AssertExpectations(t)
}

func TestUsageHandler_CreateBatchUsageRecords_CreateError(t *testing.T) {
	mockUsageService := new(MockUsageService)

	records := []models.UsageRecord{
		{ID: 1, Platform: "openai", Model: "gpt-4", TenantID: 1, UserID: 1},
		{ID: 2, Platform: "anthropic", Model: "claude-3", TenantID: 1, UserID: 1},
	}

	mockUsageService.On("CreateUsageRecord", mock.Anything, records[0]).Return(nil)
	mockUsageService.On("CreateUsageRecord", mock.Anything, records[1]).Return(assert.AnError)

	handler := NewUsageHandler(mockUsageService)
	router := setupUsageRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "member", 1)

	body, _ := json.Marshal(records)
	req, _ := http.NewRequest("POST", "/usage/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response BatchUsageResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 1, response.SyncedCount)
	assert.Len(t, response.Errors, 1)
	assert.Contains(t, response.Errors[0], "failed to create")

	mockUsageService.AssertExpectations(t)
}

func TestUsageHandler_CreateBatchUsageRecords_MissingUserContext(t *testing.T) {
	mockUsageService := new(MockUsageService)

	handler := NewUsageHandler(mockUsageService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/usage/batch", handler.CreateBatchUsageRecords)

	records := []models.UsageRecord{
		{ID: 1, Platform: "openai", Model: "gpt-4", TenantID: 1, UserID: 1},
	}
	body, _ := json.Marshal(records)
	req, _ := http.NewRequest("POST", "/usage/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "User not found in context", response["error"])
}

func TestUsageHandler_GetCurrentUsage_ServiceError(t *testing.T) {
	mockUsageService := new(MockUsageService)

	mockUsageService.On("GetCurrentUsage", mock.Anything, int64(1), int64(1)).Return(nil, assert.AnError)

	handler := NewUsageHandler(mockUsageService)
	router := setupUsageRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "member", 1)

	req, _ := http.NewRequest("GET", "/usage/current", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Failed to get current usage")

	mockUsageService.AssertExpectations(t)
}

func TestUsageHandler_GetUsageStats_ServiceError(t *testing.T) {
	mockUsageService := new(MockUsageService)

	mockUsageService.On("GetUsageStats", mock.Anything, int64(1)).Return(nil, assert.AnError)

	handler := NewUsageHandler(mockUsageService)
	router := setupUsageRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	req, _ := http.NewRequest("GET", "/usage/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Failed to get usage stats")

	mockUsageService.AssertExpectations(t)
}

// Helper function to set up usage router
func setupUsageRouter(handler *UsageHandler) *gin.Engine {
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

	router.GET("/usage/current", handler.GetCurrentUsage)
	router.GET("/usage/stats", handler.GetUsageStats)
	router.POST("/usage/batch", handler.CreateBatchUsageRecords)

	return router
}
