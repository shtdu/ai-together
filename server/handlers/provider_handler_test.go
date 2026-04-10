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

func TestProviderHandler_ListProviders_Manager(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	providers := []models.Provider{
		{ID: 1, Name: "Provider 1", TeamID: 1, Enabled: true, APIKey: "secret-key-1"},
		{ID: 2, Name: "Provider 2", TeamID: 1, Enabled: false, APIKey: "secret-key-2"},
	}

	mockProviderService.On("GetProvidersByTeamID", int64(1)).Return(providers, nil)

	handler := NewProviderHandler(mockProviderService, mockUsageService)
	router := setupProviderRouter(handler)

	testUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/providers", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Provider
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response, 2)
	// Managers should see API keys
	assert.Equal(t, "secret-key-1", response[0].APIKey)
	assert.Equal(t, "secret-key-2", response[1].APIKey)

	mockProviderService.AssertExpectations(t)
}

func TestProviderHandler_ListProviders_Member(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	providers := []models.Provider{
		{ID: 1, Name: "Provider 1", TeamID: 1, Enabled: true, APIURL: "https://api.anthropic.com", APIKey: "secret-key-1", Kind: "claude"},
		{ID: 2, Name: "Provider 2", TeamID: 1, Enabled: false, APIURL: "https://api.openai.com", APIKey: "secret-key-2", Kind: "codex"},
	}

	mockProviderService.On("GetProvidersByTeamID", int64(1)).Return(providers, nil)

	handler := NewProviderHandler(mockProviderService, mockUsageService)
	router := setupProviderRouter(handler)

	testUser := createTestUser(1, "member@example.com", "Member", "member", 1)

	req, _ := http.NewRequest("GET", "/providers", nil)
	req.Host = "server.example.com"
	req.Header.Set("X-Forwarded-Proto", "https")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Provider
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	// Members should only see enabled providers
	assert.Len(t, response, 1)
	assert.Equal(t, "Provider 1", response[0].Name)
	assert.Empty(t, response[0].APIKey)
	assert.Equal(t, "https://server.example.com/api/v1/relay/claude", response[0].APIURL)

	mockProviderService.AssertExpectations(t)
}

// Regression test: member users should NOT receive the real provider api_url
// or api_key. Instead, they should get the server relay endpoint so they
// route requests through the server relay, which holds the real credentials.
// See: https://github.com/shtdu/ai-together/issues/148
func TestProviderHandler_ListProviders_Member_NoRealProviderURL(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	providers := []models.Provider{
		{
			ID:      1,
			Name:    "Claude Direct",
			TeamID:  1,
			Enabled: true,
			APIURL:  "https://api.anthropic.com",
			APIKey:  "sk-ant-secret-key-12345",
			Kind:    "claude",
		},
	}

	mockProviderService.On("GetProvidersByTeamID", int64(1)).Return(providers, nil)

	handler := NewProviderHandler(mockProviderService, mockUsageService)
	router := setupProviderRouter(handler)

	testUser := createTestUser(1, "member@example.com", "Member", "member", 1)

	req, _ := http.NewRequest("GET", "/providers", nil)
	req.Host = "server.example.com"
	req.Header.Set("X-Forwarded-Proto", "https")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Provider
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Len(t, response, 1)

	assert.Empty(t, response[0].APIKey, "provider API key should not be exposed to member users")
	assert.NotEqual(t, "https://api.anthropic.com", response[0].APIURL,
		"Real provider API URL should not be exposed to member users; server relay URL should be used instead")
	assert.Equal(t, "https://server.example.com/api/v1/relay/claude", response[0].APIURL)

	mockProviderService.AssertExpectations(t)
}

func TestProviderHandler_CreateProvider_Success(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	testProvider := &models.Provider{ID: 1, Name: "New Provider", APIURL: "https://api.example.com", TeamID: 1}

	mockProviderService.On("CountProvidersByNameAndTeam", mock.Anything, "New Provider", int64(1)).Return(int64(0), nil)
	mockProviderService.On("CreateProvider", "New Provider", "https://api.example.com", "key123", "claude", int64(1), false, mock.AnythingOfType("map[string]interface {}"), []string(nil), 0).
		Return(testProvider, nil)

	handler := NewProviderHandler(mockProviderService, mockUsageService)
	router := setupProviderRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	reqBody := CreateProviderRequest{
		Name:   "New Provider",
		APIURL: "https://api.example.com",
		APIKey: "key123",
		Kind:   "claude",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/providers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Provider
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "New Provider", response.Name)

	mockProviderService.AssertExpectations(t)
}

func TestProviderHandler_CreateProvider_ValidationError(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	handler := NewProviderHandler(mockProviderService, mockUsageService)
	router := setupProviderRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	reqBody := map[string]interface{}{
		"name": "Provider",
		// Missing api_url
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/providers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProviderHandler_UpdateProvider_Success(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	existingProvider := &models.Provider{ID: 1, Name: "Old Name", APIURL: "https://old.example.com", APIKey: "oldkey", Kind: "claude", Enabled: true, TeamID: 1, Level: 0, ModelMapping: map[string]string{}, SupportedModels: []string{}}
	updatedProvider := &models.Provider{ID: 1, Name: "New Name", APIURL: "https://new.example.com", APIKey: "newkey", Kind: "claude", Enabled: true, TeamID: 1, Level: 0, ModelMapping: map[string]string{}, SupportedModels: []string{}}

	mockProviderService.On("GetProviderByID", int64(1)).Once().Return(existingProvider, nil)
	mockProviderService.On("CountProvidersByNameAndTeamExcludingID", mock.Anything, "New Name", int64(1), int64(1)).Once().Return(int64(0), nil)
	mockProviderService.On("UpdateProvider", int64(1), "New Name", "https://new.example.com", "newkey", "claude", true, mock.AnythingOfType("map[string]interface {}"), mock.AnythingOfType("[]string"), 0).Return(nil)
	mockProviderService.On("GetProviderByID", int64(1)).Once().Return(updatedProvider, nil)

	handler := NewProviderHandler(mockProviderService, mockUsageService)
	router := setupProviderRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	reqBody := UpdateProviderRequest{
		Name:   "New Name",
		APIURL: "https://new.example.com",
		APIKey: "newkey",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/providers/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Provider
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "New Name", response.Name)

	mockProviderService.AssertExpectations(t)
}

func TestProviderHandler_UpdateProvider_AccessDenied(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	existingProvider := &models.Provider{ID: 1, TeamID: 999} // Different team

	mockProviderService.On("GetProviderByID", int64(1)).Return(existingProvider, nil)

	handler := NewProviderHandler(mockProviderService, mockUsageService)
	router := setupProviderRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	reqBody := UpdateProviderRequest{
		Name: "New Name",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/providers/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "You don't have access to this provider", response["error"])

	mockProviderService.AssertExpectations(t)
}

func TestProviderHandler_DeleteProvider_Success(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	existingProvider := &models.Provider{ID: 1, TeamID: 1}

	mockProviderService.On("GetProviderByID", int64(1)).Return(existingProvider, nil)
	mockProviderService.On("DeleteProvider", int64(1)).Return(nil)

	handler := NewProviderHandler(mockProviderService, mockUsageService)
	router := setupProviderRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	req, _ := http.NewRequest("DELETE", "/providers/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Provider deleted successfully", response["message"])

	mockProviderService.AssertExpectations(t)
}

func TestProviderHandler_DeleteProvider_AccessDenied(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	existingProvider := &models.Provider{ID: 1, TeamID: 999} // Different tenant

	mockProviderService.On("GetProviderByID", int64(1)).Return(existingProvider, nil)
	// DeleteProvider should NOT be called for cross-tenant requests

	handler := NewProviderHandler(mockProviderService, mockUsageService)
	router := setupProviderRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1) // TenantID: 1

	req, _ := http.NewRequest("DELETE", "/providers/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusForbidden, w.Code)

	mockProviderService.AssertExpectations(t)
}

func TestProviderHandler_EnableProvider_Success(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	provider := &models.Provider{ID: 1, TeamID: 1, Enabled: false}

	mockProviderService.On("GetProviderByID", int64(1)).Return(provider, nil)
	mockProviderService.On("EnableProvider", int64(1)).Return(nil)

	handler := NewProviderHandler(mockProviderService, mockUsageService)
	router := setupProviderRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	req, _ := http.NewRequest("POST", "/providers/1/enable", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Provider enabled successfully", response["message"])
	assert.Equal(t, true, response["enabled"])

	mockProviderService.AssertExpectations(t)
}

func TestProviderHandler_EnableProvider_AccessDenied(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	provider := &models.Provider{ID: 1, TeamID: 999} // Different team

	mockProviderService.On("GetProviderByID", int64(1)).Return(provider, nil)

	handler := NewProviderHandler(mockProviderService, mockUsageService)
	router := setupProviderRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	req, _ := http.NewRequest("POST", "/providers/1/enable", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusForbidden, w.Code)

	mockProviderService.AssertExpectations(t)
}

func TestProviderHandler_DisableProvider_Success(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	provider := &models.Provider{ID: 1, TeamID: 1, Enabled: true}

	mockProviderService.On("GetProviderByID", int64(1)).Return(provider, nil)
	mockProviderService.On("DisableProvider", int64(1)).Return(nil)

	handler := NewProviderHandler(mockProviderService, mockUsageService)
	router := setupProviderRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	req, _ := http.NewRequest("POST", "/providers/1/disable", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Provider disabled successfully", response["message"])
	assert.Equal(t, false, response["enabled"])

	mockProviderService.AssertExpectations(t)
}

func TestProviderHandler_TestProvider_Success(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	provider := &models.Provider{ID: 1, Name: "Test Provider", TeamID: 1, APIURL: "https://api.example.com"}

	mockProviderService.On("GetProviderByID", int64(1)).Return(provider, nil)

	handler := NewProviderHandler(mockProviderService, mockUsageService)
	router := setupProviderRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	req, _ := http.NewRequest("POST", "/providers/1/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Provider connectivity test passed", response["message"])

	mockProviderService.AssertExpectations(t)
}

func TestProviderHandler_TestProvider_AccessDenied(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	provider := &models.Provider{ID: 1, TeamID: 999} // Different team

	mockProviderService.On("GetProviderByID", int64(1)).Return(provider, nil)

	handler := NewProviderHandler(mockProviderService, mockUsageService)
	router := setupProviderRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	req, _ := http.NewRequest("POST", "/providers/1/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusForbidden, w.Code)

	mockProviderService.AssertExpectations(t)
}

func TestProviderHandler_GetProviderStats_Success(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	provider := &models.Provider{ID: 1, Name: "Test Provider", TeamID: 1}
	stats := map[string]interface{}{
		"total_requests": 100,
		"total_tokens":   10000,
	}

	mockProviderService.On("GetProviderByID", int64(1)).Return(provider, nil)
	mockUsageService.On("GetProviderStats", mock.Anything, "Test Provider", int64(1)).Return(stats, nil)

	handler := NewProviderHandler(mockProviderService, mockUsageService)
	router := setupProviderRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	req, _ := http.NewRequest("GET", "/providers/1/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(1), response["provider_id"])
	assert.Equal(t, "Test Provider", response["provider_name"])
	assert.Equal(t, float64(100), response["total_requests"])

	mockProviderService.AssertExpectations(t)
	mockUsageService.AssertExpectations(t)
}

func TestProviderHandler_GetProviderStats_AccessDenied(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	provider := &models.Provider{ID: 1, TeamID: 999} // Different team

	mockProviderService.On("GetProviderByID", int64(1)).Return(provider, nil)

	handler := NewProviderHandler(mockProviderService, mockUsageService)
	router := setupProviderRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	req, _ := http.NewRequest("GET", "/providers/1/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusForbidden, w.Code)

	mockProviderService.AssertExpectations(t)
}

func TestProviderHandler_UpdateProvider_InvalidID(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	handler := NewProviderHandler(mockProviderService, mockUsageService)
	router := setupProviderRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	reqBody := UpdateProviderRequest{
		Name: "New Name",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/providers/invalid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid provider ID", response["error"])
}

func TestProviderHandler_DeleteProvider_InvalidID(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	handler := NewProviderHandler(mockProviderService, mockUsageService)
	router := setupProviderRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	req, _ := http.NewRequest("DELETE", "/providers/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid provider ID", response["error"])
}

// Helper function to set up provider router
func setupProviderRouter(handler *ProviderHandler) *gin.Engine {
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

	router.GET("/providers", handler.ListProviders)
	router.POST("/providers", handler.CreateProvider)
	router.PUT("/providers/:id", handler.UpdateProvider)
	router.DELETE("/providers/:id", handler.DeleteProvider)
	router.POST("/providers/:id/enable", handler.EnableProvider)
	router.POST("/providers/:id/disable", handler.DisableProvider)
	router.POST("/providers/:id/test", handler.TestProvider)
	router.GET("/providers/:id/stats", handler.GetProviderStats)

	return router
}
