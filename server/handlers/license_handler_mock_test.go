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
	"context"
	"encoding/json"
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

// MockLicenseService is a mock implementation of LicenseService for testing
type MockLicenseService struct {
	mock.Mock
}

func (m *MockLicenseService) GetLicense(ctx context.Context, tenantID int64) (*models.License, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.License), args.Error(1)
}

func (m *MockLicenseService) GetLicenseUsage(ctx context.Context, tenantID int64) (*models.LicenseUsage, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LicenseUsage), args.Error(1)
}

func (m *MockLicenseService) CanCreateTeam(ctx context.Context, tenantID int64) (bool, error) {
	args := m.Called(ctx, tenantID)
	return args.Bool(0), args.Error(1)
}

func (m *MockLicenseService) IsLicenseValid(ctx context.Context, tenantID int64) (bool, string, error) {
	args := m.Called(ctx, tenantID)
	return args.Bool(0), args.String(1), args.Error(2)
}

func (m *MockLicenseService) ActivateLicense(ctx context.Context, tenantID int64, licenseKeyPEM string) error {
	args := m.Called(ctx, tenantID, licenseKeyPEM)
	return args.Error(0)
}

func (m *MockLicenseService) GetEffectiveLicense(ctx context.Context, tenantID int64) *models.License {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.License)
}

func (m *MockLicenseService) HasActiveLicense(ctx context.Context, tenantID int64) bool {
	args := m.Called(ctx, tenantID)
	return args.Bool(0)
}

func (m *MockLicenseService) GetTiers() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockLicenseService) VerifyLicenseIntegrity(ctx context.Context, tenantID int64) (bool, string, error) {
	args := m.Called(ctx, tenantID)
	return args.Bool(0), args.String(1), args.Error(2)
}

// Helper function to create a test user
func createTestLicenseUser(id int64, email, role string, tenantID int64) *models.User {
	return &models.User{
		ID:       id,
		Email:    email,
		Name:     "Test User",
		Role:     role,
		TenantID: tenantID,
	}
}

// Helper function to set up test gin context with user
func setupTestGinContext(user *models.User) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user", user)
	return c, w
}

// Helper to set up test router
func setupLicenseTestRouter(handler *LicenseHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/license", func(c *gin.Context) {
		// Mock authentication middleware - set user in context
		user := createTestLicenseUser(1, "test@example.com", "manager", 123)
		c.Set("user", user)
		handler.GetLicense(c)
	})
	router.POST("/api/v1/license/activate", func(c *gin.Context) {
		user := createTestLicenseUser(1, "test@example.com", "manager", 123)
		c.Set("user", user)
		handler.ActivateLicense(c)
	})
	router.GET("/api/v1/license/tiers", handler.GetTiers)
	return router
}

func TestLicenseHandler_GetLicense_Success(t *testing.T) {
	mockService := new(MockLicenseService)
	handler := NewLicenseHandler(mockService)

	license := &models.License{
		TenantID:           123,
		LicenseID:          "lic-123",
		LicenseType:        "commercial",
		MaxSeats:           100,
		MaxTeams:           -1,
		DataRetentionDays:  90,
		IssuedAt:           time.Now(),
		ExpiresAt:          time.Now().Add(30 * 24 * time.Hour),
	}

	usage := &models.LicenseUsage{
		License:        *license,
		CurrentUsers:   10,
		CurrentTeams:   2,
		SeatsRemaining: 90,
		TeamsRemaining: -1, // unlimited
		ProviderCounts: map[string]int{
			"claude": 2,
			"codex":  1,
		},
	}

	mockService.On("GetLicense", mock.Anything, int64(123)).Return(license, nil)
	mockService.On("GetLicenseUsage", mock.Anything, int64(123)).Return(usage, nil)
	mockService.On("CanCreateTeam", mock.Anything, int64(123)).Return(true, nil)
	mockService.On("HasActiveLicense", mock.Anything, int64(123)).Return(true)

	router := setupLicenseTestRouter(handler)
	req, _ := http.NewRequest("GET", "/api/v1/license", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify license structure
	licenseData := response["license"].(map[string]interface{})
	assert.Equal(t, "lic-123", licenseData["license_id"])
	assert.Equal(t, "commercial", licenseData["type"])
	assert.Equal(t, "Commercial", licenseData["type_name"])
	assert.Equal(t, float64(100), licenseData["max_seats"])
	assert.Equal(t, float64(-1), licenseData["max_teams"]) // unlimited

	// Verify usage structure
	usageData := response["usage"].(map[string]interface{})
	assert.Equal(t, float64(10), usageData["current_users"])
	assert.Equal(t, float64(2), usageData["current_teams"])
	providerCounts := usageData["provider_counts"].(map[string]interface{})
	assert.Equal(t, float64(2), providerCounts["claude"])
	assert.Equal(t, float64(1), providerCounts["codex"])

	// Verify status structure
	statusData := response["status"].(map[string]interface{})
	assert.True(t, statusData["has_active_license"].(bool))
	assert.True(t, statusData["can_add_user"].(bool))
	canAddProvider := statusData["can_add_provider"].(map[string]interface{})
	assert.True(t, canAddProvider["claude"].(bool))
	assert.True(t, canAddProvider["codex"].(bool))
	assert.True(t, statusData["can_create_team"].(bool))

	mockService.AssertExpectations(t)
}

func TestLicenseHandler_GetLicense_DefaultLicense(t *testing.T) {
	mockService := new(MockLicenseService)
	handler := NewLicenseHandler(mockService)

	// Return invalid license (triggers default fallback)
	invalidLicense := &models.License{
		TenantID: 123,
		// Missing required fields
	}

	usage := &models.LicenseUsage{
		License:        *invalidLicense,
		CurrentUsers:   2,
		SeatsRemaining: 1,
		ProviderCounts: map[string]int{
			"claude": 1,
		},
	}

	mockService.On("GetLicense", mock.Anything, int64(123)).Return(invalidLicense, nil)
	mockService.On("GetLicenseUsage", mock.Anything, int64(123)).Return(usage, nil)
	mockService.On("CanCreateTeam", mock.Anything, int64(123)).Return(true, nil)
	mockService.On("HasActiveLicense", mock.Anything, int64(123)).Return(false)

	router := setupLicenseTestRouter(handler)
	req, _ := http.NewRequest("GET", "/api/v1/license", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	statusData := response["status"].(map[string]interface{})
	assert.False(t, statusData["has_active_license"].(bool))
	assert.True(t, statusData["is_using_defaults"].(bool), "should have is_using_defaults flag")

	mockService.AssertExpectations(t)
}

func TestLicenseHandler_GetLicense_UserNotInContext(t *testing.T) {
	mockService := new(MockLicenseService)
	handler := NewLicenseHandler(mockService)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// Don't set user in context

	handler.GetLicense(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "User not found in context")
}

func TestLicenseHandler_GetLicense_ServiceError(t *testing.T) {
	mockService := new(MockLicenseService)
	handler := NewLicenseHandler(mockService)

	mockService.On("GetLicense", mock.Anything, int64(123)).Return(nil, assert.AnError)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/license", nil)
	user := createTestLicenseUser(1, "test@example.com", "manager", 123)
	c.Set("user", user)

	handler.GetLicense(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Failed to get license")

	mockService.AssertExpectations(t)
}

func TestLicenseHandler_ActivateLicense_Success(t *testing.T) {
	mockService := new(MockLicenseService)
	handler := NewLicenseHandler(mockService)

	mockService.On("ActivateLicense", mock.Anything, int64(123), "test-license-key-pem").Return(nil)

	router := setupLicenseTestRouter(handler)
	reqBody := map[string]string{
		"license_key": "test-license-key-pem",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/license/activate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "License activated successfully", response["message"])

	mockService.AssertExpectations(t)
}

func TestLicenseHandler_ActivateLicense_MissingLicenseKey(t *testing.T) {
	mockService := new(MockLicenseService)
	handler := NewLicenseHandler(mockService)

	router := setupLicenseTestRouter(handler)
	reqBody := map[string]string{} // Missing license_key
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/license/activate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid request")
}

func TestLicenseHandler_ActivateLicense_EmptyLicenseKey(t *testing.T) {
	mockService := new(MockLicenseService)
	handler := NewLicenseHandler(mockService)

	router := setupLicenseTestRouter(handler)
	reqBody := map[string]string{
		"license_key": "", // Empty license_key
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/license/activate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid request")
}

func TestLicenseHandler_ActivateLicense_ServiceError(t *testing.T) {
	mockService := new(MockLicenseService)
	handler := NewLicenseHandler(mockService)

	mockService.On("ActivateLicense", mock.Anything, int64(123), "invalid-key").Return(assert.AnError)

	router := setupLicenseTestRouter(handler)
	reqBody := map[string]string{
		"license_key": "invalid-key",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/license/activate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "assert.AnError")

	mockService.AssertExpectations(t)
}

func TestLicenseHandler_ActivateLicense_UserNotInContext(t *testing.T) {
	mockService := new(MockLicenseService)
	handler := NewLicenseHandler(mockService)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// Don't set user in context

	reqBody := map[string]string{
		"license_key": "test-key",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/license/activate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.ActivateLicense(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "User not found in context")
}

func TestLicenseHandler_GetTiers_Success(t *testing.T) {
	mockService := new(MockLicenseService)
	handler := NewLicenseHandler(mockService)

	expectedTiers := map[string]interface{}{
		"tiers": []interface{}{
			map[string]interface{}{
				"tier":                  "0.0",
				"name":                  "Community",
				"max_providers_per_kind": 2,
				"features":              []string{"Basic provider management"},
			},
			map[string]interface{}{
				"tier":                  "1.0",
				"name":                  "Standard",
				"max_providers_per_kind": -1,
				"features":              []string{"Unlimited providers", "Basic analytics"},
			},
		},
	}

	mockService.On("GetTiers").Return(expectedTiers)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/license/tiers", nil)

	handler.GetTiers(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response has tiers key
	assert.Contains(t, response, "tiers")
	tiers := response["tiers"].([]interface{})
	assert.Len(t, tiers, 2)

	// Verify first tier (Community)
	tier0 := tiers[0].(map[string]interface{})
	assert.Equal(t, "0.0", tier0["tier"])
	assert.Equal(t, "Community", tier0["name"])
	assert.Equal(t, float64(2), tier0["max_providers_per_kind"])

	// Verify second tier (Standard)
	tier1 := tiers[1].(map[string]interface{})
	assert.Equal(t, "1.0", tier1["tier"])
	assert.Equal(t, "Standard", tier1["name"])
	assert.Equal(t, float64(-1), tier1["max_providers_per_kind"])

	mockService.AssertExpectations(t)
}

func TestLicenseHandler_GetTiers_NoAuthRequired(t *testing.T) {
	mockService := new(MockLicenseService)
	handler := NewLicenseHandler(mockService)

	expectedTiers := map[string]interface{}{
		"tiers": []interface{}{
			map[string]interface{}{
				"tier":                  "0.0",
				"name":                  "Community",
				"max_providers_per_kind": 2,
				"features":              []string{"Basic provider management"},
			},
		},
	}

	mockService.On("GetTiers").Return(expectedTiers)

	router := gin.New()
	router.GET("/api/v1/license/tiers", handler.GetTiers)

	req, _ := http.NewRequest("GET", "/api/v1/license/tiers", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should succeed without authentication
	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestLicenseHandler_DaysRemaining_Calculation(t *testing.T) {
	mockService := new(MockLicenseService)
	handler := NewLicenseHandler(mockService)

	tests := []struct {
		name           string
		expiresAt      time.Time
		expectedMin    int64
		expectedMax    int64
	}{
		{
			name:        "expires in 5 days",
			expiresAt:   time.Now().Add(5 * 24 * time.Hour),
			expectedMin: 4, // Allow 1 day variance
			expectedMax: 6,
		},
		{
			name:        "expires in 30 days",
			expiresAt:   time.Now().Add(30 * 24 * time.Hour),
			expectedMin: 29,
			expectedMax: 31,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			license := &models.License{
				TenantID:  123,
				LicenseID: "lic-123",
				Tier:      "2.0",
				Seats:     25,
				IssuedAt:  time.Now(),
				ExpiresAt: tt.expiresAt,
			}

			usage := &models.LicenseUsage{
				License:        *license,
				CurrentUsers:   10,
				SeatsRemaining: 15,
				ProviderCounts: map[string]int{},
			}

			mockService.On("GetLicense", mock.Anything, int64(123)).Return(license, nil).Once()
			mockService.On("GetLicenseUsage", mock.Anything, int64(123)).Return(usage, nil).Once()
			mockService.On("CanAddUser", mock.Anything, int64(123)).Return(true, nil).Once()
			mockService.On("CanCreateTeam", mock.Anything, int64(123)).Return(true, nil)
	mockService.On("HasActiveLicense", mock.Anything, int64(123)).Return(true).Once()

			router := setupLicenseTestRouter(handler)
			req, _ := http.NewRequest("GET", "/api/v1/license", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			statusData := response["status"].(map[string]interface{})
			daysRemaining := int64(statusData["days_remaining"].(float64))
			assert.GreaterOrEqual(t, daysRemaining, tt.expectedMin)
			assert.LessOrEqual(t, daysRemaining, tt.expectedMax)
		})
	}
}


func TestLicenseHandler_NilOrString_Helper(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"non-empty string", "test-value", "test-value"},
		{"empty string", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nilOrString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLicenseHandler_NilOrTime_Helper(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		input    time.Time
		expected interface{}
	}{
		{"valid time", now, now},
		{"zero time", time.Time{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nilOrTime(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
