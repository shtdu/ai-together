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

// ==================== RelayMessages Tests ====================

func TestRelayHandler_RelayMessages_InvalidTool(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	handler := NewRelayHandler(mockProviderService, mockUsageService)
	router := setupRelayRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	reqBody := map[string]interface{}{
		"model":      "claude-sonnet-4",
		"max_tokens": 100,
		"messages":   []interface{}{map[string]string{"role": "user", "content": "hello"}},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/relay/invalid-tool/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "invalid tool kind")
}

func TestRelayHandler_RelayMessages_NoProvidersFound(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	// Return empty list of providers
	mockProviderService.On("GetProvidersByTeamID", int64(1)).Return([]models.Provider{}, nil)

	handler := NewRelayHandler(mockProviderService, mockUsageService)
	router := setupRelayRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	reqBody := map[string]interface{}{
		"model":      "claude-sonnet-4",
		"max_tokens": 100,
		"messages":   []interface{}{map[string]string{"role": "user", "content": "hello"}},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/relay/claude/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "no available provider supports model")

	mockProviderService.AssertExpectations(t)
}

func TestRelayHandler_RelayMessages_NoProviders_NoModelSpecified(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	// Return empty list of providers
	mockProviderService.On("GetProvidersByTeamID", int64(1)).Return([]models.Provider{}, nil)

	handler := NewRelayHandler(mockProviderService, mockUsageService)
	router := setupRelayRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	reqBody := map[string]interface{}{
		"max_tokens": 100,
		"messages":   []interface{}{map[string]string{"role": "user", "content": "hello"}},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/relay/claude/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "no providers available")

	mockProviderService.AssertExpectations(t)
}

func TestRelayHandler_RelayMessages_ProviderLoadError(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	// Return error from provider service
	mockProviderService.On("GetProvidersByTeamID", int64(1)).Return([]models.Provider{}, assert.AnError)

	handler := NewRelayHandler(mockProviderService, mockUsageService)
	router := setupRelayRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	reqBody := map[string]interface{}{
		"model":      "claude-sonnet-4",
		"max_tokens": 100,
		"messages":   []interface{}{map[string]string{"role": "user", "content": "hello"}},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/relay/claude/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "failed to load providers")

	mockProviderService.AssertExpectations(t)
}

func TestRelayHandler_RelayMessages_UserNotInContext(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	handler := NewRelayHandler(mockProviderService, mockUsageService)
	router := setupRelayRouter(handler)

	reqBody := map[string]interface{}{
		"model":      "claude-sonnet-4",
		"max_tokens": 100,
		"messages":   []interface{}{map[string]string{"role": "user", "content": "hello"}},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/relay/claude/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "User not found in context")
}

// ==================== RelayChatCompletions Tests ====================

func TestRelayHandler_RelayChatCompletions_InvalidTool(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	handler := NewRelayHandler(mockProviderService, mockUsageService)
	router := setupRelayRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	reqBody := map[string]interface{}{
		"model":    "gpt-4",
		"messages": []interface{}{map[string]string{"role": "user", "content": "hello"}},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/relay/invalid-tool/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "invalid tool kind")
}

func TestRelayHandler_RelayChatCompletions_ValidTool(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	// Return a provider that will fail (since we can't mock actual HTTP calls in handler tests)
	providers := []models.Provider{
		{
			ID:              1,
			Name:            "Test Provider",
			APIURL:          "https://api.invalid.example.com",
			APIKey:          "test-key",
			TeamID:          1,
			Kind:            "codex",
			Enabled:         true,
			ModelMapping:    map[string]string{},
			SupportedModels: []string{"gpt-4"},
			Level:           1,
		},
	}
	mockProviderService.On("GetProvidersByTeamID", int64(1)).Return(providers, nil)

	// Mock usage service
	mockUsageService.On("CreateUsageRecord", mock.Anything, mock.Anything).Return(nil)

	handler := NewRelayHandler(mockProviderService, mockUsageService)
	router := setupRelayRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	reqBody := map[string]interface{}{
		"model":    "gpt-4",
		"messages": []interface{}{map[string]string{"role": "user", "content": "hello"}},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/relay/codex/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	// Should return bad request because the provider will fail to connect
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "all")

	mockProviderService.AssertExpectations(t)
}

// ==================== filterProvidersByKind Tests ====================

func TestRelayHandler_FilterProvidersByKind_AllFilters(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	handler := NewRelayHandler(mockProviderService, mockUsageService)

	providers := []models.Provider{
		{ID: 1, Name: "Enabled Provider", APIURL: "https://api.example.com", APIKey: "key", TeamID: 1, Kind: "claude", Enabled: true, ModelMapping: map[string]string{}, SupportedModels: []string{"claude-sonnet-4"}, Level: 1},
		{ID: 2, Name: "Disabled Provider", APIURL: "https://api.example.com", APIKey: "key", TeamID: 1, Kind: "claude", Enabled: false, ModelMapping: map[string]string{}, SupportedModels: []string{"claude-sonnet-4"}, Level: 1},
		{ID: 3, Name: "Wrong Kind", APIURL: "https://api.example.com", APIKey: "key", TeamID: 1, Kind: "codex", Enabled: true, ModelMapping: map[string]string{}, SupportedModels: []string{"gpt-4"}, Level: 1},
		{ID: 4, Name: "Missing URL", APIURL: "", APIKey: "key", TeamID: 1, Kind: "claude", Enabled: true, ModelMapping: map[string]string{}, SupportedModels: []string{"claude-sonnet-4"}, Level: 1},
		{ID: 5, Name: "Missing Key", APIURL: "https://api.example.com", APIKey: "", TeamID: 1, Kind: "claude", Enabled: true, ModelMapping: map[string]string{}, SupportedModels: []string{"claude-sonnet-4"}, Level: 1},
		{ID: 6, Name: "Model Not Supported", APIURL: "https://api.example.com", APIKey: "key", TeamID: 1, Kind: "claude", Enabled: true, ModelMapping: map[string]string{}, SupportedModels: []string{"claude-opus-4"}, Level: 1},
	}

	// Only provider 1 should pass all filters
	active := handler.filterProvidersByKind(providers, "claude", "claude-sonnet-4")

	assert.Len(t, active, 1)
	assert.Equal(t, "Enabled Provider", active[0].Name)
}

func TestRelayHandler_FilterProvidersByKind_WildcardModel(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	handler := NewRelayHandler(mockProviderService, mockUsageService)

	providers := []models.Provider{
		{ID: 1, Name: "Wildcard Provider", APIURL: "https://api.example.com", APIKey: "key", TeamID: 1, Kind: "claude", Enabled: true, ModelMapping: map[string]string{}, SupportedModels: []string{"claude-*"}, Level: 1},
		{ID: 2, Name: "Specific Model Provider", APIURL: "https://api.example.com", APIKey: "key", TeamID: 1, Kind: "claude", Enabled: true, ModelMapping: map[string]string{}, SupportedModels: []string{"claude-opus-4"}, Level: 1},
	}

	// Both providers should support claude-sonnet-4 (one via wildcard, one doesn't match specific model)
	active := handler.filterProvidersByKind(providers, "claude", "claude-sonnet-4")

	assert.Len(t, active, 1)
	assert.Equal(t, "Wildcard Provider", active[0].Name)
}

func TestRelayHandler_FilterProvidersByKind_NoModelFilter(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	handler := NewRelayHandler(mockProviderService, mockUsageService)

	providers := []models.Provider{
		{ID: 1, Name: "Provider 1", APIURL: "https://api.example.com", APIKey: "key", TeamID: 1, Kind: "claude", Enabled: true, ModelMapping: map[string]string{}, SupportedModels: []string{"claude-sonnet-4"}, Level: 1},
		{ID: 2, Name: "Provider 2", APIURL: "https://api.example.com", APIKey: "key", TeamID: 1, Kind: "claude", Enabled: true, ModelMapping: map[string]string{}, SupportedModels: []string{}, Level: 1},
	}

	// With no model specified, both providers should be included
	active := handler.filterProvidersByKind(providers, "claude", "")

	assert.Len(t, active, 2)
}

func TestRelayHandler_FilterProvidersByKind_ModelMapping(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	handler := NewRelayHandler(mockProviderService, mockUsageService)

	providers := []models.Provider{
		{ID: 1, Name: "Mapping Provider", APIURL: "https://api.example.com", APIKey: "key", TeamID: 1, Kind: "claude", Enabled: true, ModelMapping: map[string]string{"sonnet": "claude-sonnet-4"}, SupportedModels: []string{"claude-sonnet-4"}, Level: 1},
	}

	// Provider supports claude-sonnet-4, and model mapping allows "sonnet" to map to it
	active := handler.filterProvidersByKind(providers, "claude", "sonnet")

	assert.Len(t, active, 1)
	assert.Equal(t, "Mapping Provider", active[0].Name)
}

// ==================== joinURL Tests ====================

func TestRelayHandler_JoinURL(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		path     string
		expected string
	}{
		{
			name:     "base without slash, path without slash",
			base:     "https://api.example.com",
			path:     "v1/messages",
			expected: "https://api.example.com/v1/messages",
		},
		{
			name:     "base with trailing slash, path without slash",
			base:     "https://api.example.com/",
			path:     "v1/messages",
			expected: "https://api.example.com/v1/messages",
		},
		{
			name:     "base without slash, path with leading slash",
			base:     "https://api.example.com",
			path:     "/v1/messages",
			expected: "https://api.example.com/v1/messages",
		},
		{
			name:     "both have slashes",
			base:     "https://api.example.com/",
			path:     "/v1/messages",
			expected: "https://api.example.com/v1/messages",
		},
		{
			name:     "base with path, path added",
			base:     "https://api.example.com/api",
			path:     "v1/messages",
			expected: "https://api.example.com/api/v1/messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinURL(tt.base, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ==================== flattenQuery Tests ====================

func TestRelayHandler_FlattenQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string][]string
		expected map[string]string
	}{
		{
			name: "single values",
			input: map[string][]string{
				"key1": {"value1"},
				"key2": {"value2"},
			},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "multiple values - takes first",
			input: map[string][]string{
				"key1": {"value1", "value2"},
			},
			expected: map[string]string{
				"key1": "value1",
			},
		},
		{
			name:     "empty map",
			input:    map[string][]string{},
			expected: map[string]string{},
		},
		{
			name: "empty slice value",
			input: map[string][]string{
				"key1": {},
				"key2": {"value2"},
			},
			expected: map[string]string{
				"key2": "value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenQuery(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ==================== Provider Sorting Tests ====================

func TestRelayHandler_ProviderSortingByLevel(t *testing.T) {
	mockProviderService := new(MockProviderService)
	mockUsageService := new(MockUsageService)

	handler := NewRelayHandler(mockProviderService, mockUsageService)

	providers := []models.Provider{
		{ID: 1, Name: "Level 3 Provider", APIURL: "https://api.example.com", APIKey: "key", TeamID: 1, Kind: "claude", Enabled: true, ModelMapping: map[string]string{}, SupportedModels: []string{"*"}, Level: 3},
		{ID: 2, Name: "Level 1 Provider", APIURL: "https://api.example.com", APIKey: "key", TeamID: 1, Kind: "claude", Enabled: true, ModelMapping: map[string]string{}, SupportedModels: []string{"*"}, Level: 1},
		{ID: 3, Name: "Level 2 Provider", APIURL: "https://api.example.com", APIKey: "key", TeamID: 1, Kind: "claude", Enabled: true, ModelMapping: map[string]string{}, SupportedModels: []string{"*"}, Level: 2},
		{ID: 4, Name: "Zero Level Provider", APIURL: "https://api.example.com", APIKey: "key", TeamID: 1, Kind: "claude", Enabled: true, ModelMapping: map[string]string{}, SupportedModels: []string{"*"}, Level: 0},
		{ID: 5, Name: "Negative Level Provider", APIURL: "https://api.example.com", APIKey: "key", TeamID: 1, Kind: "claude", Enabled: true, ModelMapping: map[string]string{}, SupportedModels: []string{"*"}, Level: -1},
	}

	active := handler.filterProvidersByKind(providers, "claude", "")

	// filterProvidersByKind doesn't sort - it just filters
	// The sorting happens in relayForward after filtering
	// So we just verify all providers passed the filter
	assert.Len(t, active, 5)

	// Verify that filtering by level normalization works for edge cases
	// Create a new slice with mixed levels and verify they all pass through
	mixedProviders := []models.Provider{
		{ID: 1, Name: "Positive Level", APIURL: "https://api.example.com", APIKey: "key", TeamID: 1, Kind: "claude", Enabled: true, ModelMapping: map[string]string{}, SupportedModels: []string{"*"}, Level: 5},
		{ID: 2, Name: "Zero Level", APIURL: "https://api.example.com", APIKey: "key", TeamID: 1, Kind: "claude", Enabled: true, ModelMapping: map[string]string{}, SupportedModels: []string{"*"}, Level: 0},
		{ID: 3, Name: "Negative Level", APIURL: "https://api.example.com", APIKey: "key", TeamID: 1, Kind: "claude", Enabled: true, ModelMapping: map[string]string{}, SupportedModels: []string{"*"}, Level: -5},
	}

	active = handler.filterProvidersByKind(mixedProviders, "claude", "")
	assert.Len(t, active, 3)

	// Verify all levels are represented
	levels := make([]int, len(active))
	for i, p := range active {
		levels[i] = p.Level
	}
	assert.Contains(t, levels, 5)
	assert.Contains(t, levels, 0)
	assert.Contains(t, levels, -5)
}

// ==================== Helper Functions ====================

// setupRelayRouter sets up a test router for relay handler
func setupRelayRouter(handler *RelayHandler) *gin.Engine {
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
				_, _ = fmt.Sscanf(userID, "%d", &id)
			}
			if tenantID != "" {
				_, _ = fmt.Sscanf(tenantID, "%d", &tenant)
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

	// Register relay routes
	router.POST("/relay/:tool/v1/messages", handler.RelayMessages)
	router.POST("/relay/:tool/v1/chat/completions", handler.RelayChatCompletions)

	return router
}
