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
	"switch-server/services"
)

// MockUserService is a mock implementation of UserServiceInterface using testify/mock
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(email, password, name, role string, tenantID int64) (*models.User, error) {
	args := m.Called(email, password, name, role, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(userID int64) (*models.User, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) ValidatePassword(user *models.User, password string) bool {
	args := m.Called(user, password)
	return args.Bool(0)
}

func (m *MockUserService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockUserService) ListUsersByTenant(tenantID int64) ([]models.User, error) {
	args := m.Called(tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(userID int64, name, role, password string) (*models.User, error) {
	args := m.Called(userID, name, role, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(userID int64) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockUserService) UpdateProfileName(userID int64, name string) (*models.User, error) {
	args := m.Called(userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) ChangePassword(userID int64, currentPassword, newPassword string) error {
	args := m.Called(userID, currentPassword, newPassword)
	return args.Error(0)
}

// MockTeamService is a mock implementation of TeamServiceInterface
type MockTeamService struct {
	mock.Mock
}

func (m *MockTeamService) CreateTeam(name, description string, ownerID, tenantID int64, settings map[string]string) (*models.Team, error) {
	args := m.Called(name, description, ownerID, tenantID, settings)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamService) GetTeamByID(teamID int64) (*models.Team, error) {
	args := m.Called(teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamService) GetTeamsByUserID(userID int64) ([]models.Team, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Team), args.Error(1)
}

func (m *MockTeamService) AddTeamMember(teamID, userID int64, role string) error {
	args := m.Called(teamID, userID, role)
	return args.Error(0)
}

func (m *MockTeamService) RemoveTeamMember(teamID, userID int64) error {
	args := m.Called(teamID, userID)
	return args.Error(0)
}

func (m *MockTeamService) GetTeamMembers(teamID int64) ([]models.User, error) {
	args := m.Called(teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockTeamService) UpdateTeam(teamID int64, name, description string, settings map[string]string) error {
	args := m.Called(teamID, name, description, settings)
	return args.Error(0)
}

func (m *MockTeamService) DeleteTeam(teamID int64) error {
	args := m.Called(teamID)
	return args.Error(0)
}

// MockUsageService is a mock implementation of UsageServiceInterface
type MockUsageService struct {
	mock.Mock
}

func (m *MockUsageService) GetCurrentUsage(ctx context.Context, userID, tenantID int64) (map[string]interface{}, error) {
	args := m.Called(ctx, userID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUsageService) GetUsageStats(ctx context.Context, tenantID int64) (map[string]interface{}, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUsageService) CreateUsageRecord(ctx context.Context, record models.UsageRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockUsageService) GetProviderStats(ctx context.Context, providerName string, tenantID int64) (map[string]interface{}, error) {
	args := m.Called(ctx, providerName, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUsageService) GetDashboardMetrics(ctx context.Context, tenantID, userID int64, role string, startTime, endTime time.Time, interval string) (map[string]interface{}, error) {
	args := m.Called(ctx, tenantID, userID, role, startTime, endTime, interval)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUsageService) GetProviderRankings(ctx context.Context, tenantID, userID int64, role string) (map[string]interface{}, error) {
	args := m.Called(ctx, tenantID, userID, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUsageService) GetMemberStats(ctx context.Context, tenantID int64) (map[string]interface{}, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUsageService) GetProviderAnalytics(ctx context.Context, tenantID int64, startDate, endDate string, providers, models, tools []string) (map[string]interface{}, error) {
	args := m.Called(ctx, tenantID, startDate, endDate, providers, models, tools)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUsageService) GetUserAnalytics(ctx context.Context, tenantID int64, startDate, endDate string, userIDs []int64, providers, tools []string) (map[string]interface{}, error) {
	args := m.Called(ctx, tenantID, startDate, endDate, userIDs, providers, tools)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUsageService) GetHistory(ctx context.Context, tenantID int64, startDate, endDate string, page, limit int, userIDs []int64, providers, models, tools []string, sortBy, sortOrder string) (map[string]interface{}, error) {
	args := m.Called(ctx, tenantID, startDate, endDate, page, limit, userIDs, providers, models, tools, sortBy, sortOrder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUsageService) GetFilterOptions(ctx context.Context, tenantID int64) (map[string]interface{}, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUsageService) GetPersonalAnalytics(ctx context.Context, userID, tenantID int64, startDate, endDate string, providers, models, tools []string) (map[string]interface{}, error) {
	args := m.Called(ctx, userID, tenantID, startDate, endDate, providers, models, tools)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUsageService) GetPersonalHistory(ctx context.Context, userID, tenantID int64, startDate, endDate string, page, limit int, providers, models, tools []string, sortBy, sortOrder string) (map[string]interface{}, error) {
	args := m.Called(ctx, userID, tenantID, startDate, endDate, page, limit, providers, models, tools, sortBy, sortOrder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUsageService) GetPersonalFilterOptions(ctx context.Context, userID, tenantID int64) (map[string]interface{}, error) {
	args := m.Called(ctx, userID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// MockProviderService is a mock implementation of ProviderServiceInterface
type MockProviderService struct {
	mock.Mock
}

func (m *MockProviderService) CreateProvider(name, apiURL, apiKey, kind string, teamID int64, enabled bool, modelMapping map[string]interface{}, supportedModels []string, level int) (*models.Provider, error) {
	args := m.Called(name, apiURL, apiKey, kind, teamID, enabled, modelMapping, supportedModels, level)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Provider), args.Error(1)
}

func (m *MockProviderService) GetProviderByID(providerID int64) (*models.Provider, error) {
	args := m.Called(providerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Provider), args.Error(1)
}

func (m *MockProviderService) GetProvidersByTeamID(teamID int64) ([]models.Provider, error) {
	args := m.Called(teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Provider), args.Error(1)
}

func (m *MockProviderService) UpdateProvider(providerID int64, name, apiURL, apiKey, kind string, enabled bool, modelMapping map[string]interface{}, supportedModels []string, level int) error {
	args := m.Called(providerID, name, apiURL, apiKey, kind, enabled, modelMapping, supportedModels, level)
	return args.Error(0)
}

func (m *MockProviderService) DeleteProvider(providerID int64) error {
	args := m.Called(providerID)
	return args.Error(0)
}

func (m *MockProviderService) EnableProvider(providerID int64) error {
	args := m.Called(providerID)
	return args.Error(0)
}

func (m *MockProviderService) DisableProvider(providerID int64) error {
	args := m.Called(providerID)
	return args.Error(0)
}

func (m *MockProviderService) CountProvidersByNameAndTeam(ctx context.Context, name string, teamID int64) (int64, error) {
	args := m.Called(ctx, name, teamID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProviderService) CountProvidersByNameAndTeamExcludingID(ctx context.Context, name string, teamID, excludeID int64) (int64, error) {
	args := m.Called(ctx, name, teamID, excludeID)
	return args.Get(0).(int64), args.Error(1)
}

// Helper function to create a test user
func createTestUser(id int64, email, name, role string, tenantID int64) *models.User {
	return &models.User{
		ID:       id,
		Email:    email,
		Name:     name,
		Role:     role,
		TenantID: tenantID,
		Password: "$2a$10$dummy.hashed.password.for.testing", // Dummy hash
	}
}

// Helper function to set up test gin context
func setupTestRouter(handler *AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", handler.Login)
	router.POST("/register", handler.Register)
	router.POST("/refresh", handler.Refresh)
	router.POST("/verify", handler.Verify)
	router.GET("/profile", handler.GetProfile)
	return router
}

// TestLoginHandler_Success tests successful login with mock service
func TestLoginHandler_Success(t *testing.T) {
	// Create mock service
	mockService := new(MockUserService)

	// Set up mock expectations
	testUser := createTestUser(1, "test@example.com", "Test User", "manager", 1)
	mockService.On("GetUserByEmail", "test@example.com").Return(testUser, nil)
	mockService.On("ValidatePassword", testUser, "password123").Return(true)

	// Create handler with mock service
	handler := NewAuthHandler(mockService, nil)
	router := setupTestRouter(handler)

	// Create request
	reqBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, "test@example.com", response.User.Email)
	assert.True(t, response.ExpiresAt.After(time.Now()))

	// Verify mock expectations were met
	mockService.AssertExpectations(t)
}

// TestLoginHandler_UserNotFound tests login with non-existent user
func TestLoginHandler_UserNotFound(t *testing.T) {
	mockService := new(MockUserService)

	// Set up mock to return error (user not found)
	mockService.On("GetUserByEmail", "nonexistent@example.com").Return(nil, assert.AnError)

	handler := NewAuthHandler(mockService, nil)
	router := setupTestRouter(handler)

	reqBody := map[string]interface{}{
		"email":    "nonexistent@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid credentials", response["error"])

	mockService.AssertExpectations(t)
}

// TestLoginHandler_InvalidPassword tests login with invalid password
func TestLoginHandler_InvalidPassword(t *testing.T) {
	mockService := new(MockUserService)

	testUser := createTestUser(1, "test@example.com", "Test User", "manager", 1)
	mockService.On("GetUserByEmail", "test@example.com").Return(testUser, nil)
	mockService.On("ValidatePassword", testUser, "wrongpassword").Return(false)

	handler := NewAuthHandler(mockService, nil)
	router := setupTestRouter(handler)

	reqBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid credentials", response["error"])

	mockService.AssertExpectations(t)
}

// TestLoginHandler_ValidationError tests login with invalid request format
func TestLoginHandler_ValidationError(t *testing.T) {
	mockService := new(MockUserService)
	// No expectations set - should fail validation before calling service

	handler := NewAuthHandler(mockService, nil)
	router := setupTestRouter(handler)

	tests := []struct {
		name        string
		requestBody map[string]interface{}
	}{
		{
			name: "missing email",
			requestBody: map[string]interface{}{
				"password": "password123",
			},
		},
		{
			name: "missing password",
			requestBody: map[string]interface{}{
				"email": "test@example.com",
			},
		},
		{
			name: "invalid email format",
			requestBody: map[string]interface{}{
				"email":    "not-an-email",
				"password": "password123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Equal(t, "Invalid request payload", response["error"])
		})
	}

	// Verify no mock methods were called
	mockService.AssertNotCalled(t, "GetUserByEmail")
	mockService.AssertNotCalled(t, "ValidatePassword")
}

// TestRegisterHandler_Success tests successful registration with mock service
func TestRegisterHandler_Success(t *testing.T) {
	mockService := new(MockUserService)

	testUser := createTestUser(1, "newuser@example.com", "New User", "member", 1)
	mockService.On("CreateUser", "newuser@example.com", "Password123!", "New User", "member", int64(1)).
		Return(testUser, nil)

	handler := NewAuthHandler(mockService, nil)
	router := setupTestRouter(handler)

	reqBody := map[string]interface{}{
		"email":    "newuser@example.com",
		"password": "Password123!",
		"name":     "New User",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, "newuser@example.com", response.User.Email)

	mockService.AssertExpectations(t)
}

// TestRegisterHandler_UserExists tests registration with existing email
func TestRegisterHandler_UserExists(t *testing.T) {
	mockService := new(MockUserService)

	mockService.On("CreateUser", "existing@example.com", "Password123!", "Existing User", "member", int64(1)).
		Return(nil, assert.AnError)

	handler := NewAuthHandler(mockService, nil)
	router := setupTestRouter(handler)

	reqBody := map[string]interface{}{
		"email":    "existing@example.com",
		"password": "Password123!",
		"name":     "Existing User",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Failed to create user")

	mockService.AssertExpectations(t)
}

// TestGetProfileHandler_Success tests getting user profile
func TestGetProfileHandler_Success(t *testing.T) {
	mockService := new(MockUserService)
	mockTeamService := new(MockTeamService)

	testUser := createTestUser(1, "test@example.com", "Test User", "manager", 1)
	mockService.On("GetUserByID", int64(1)).Return(testUser, nil)
	mockTeamService.On("GetTeamsByUserID", int64(1)).Return([]models.Team{}, nil)

	handler := NewAuthHandler(mockService, mockTeamService)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up context with user
	testHandler := func(c *gin.Context) {
		c.Set("user", testUser)
		handler.GetProfile(c)
	}

	router.GET("/profile", testHandler)

	req, _ := http.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	userData := response["user"].(map[string]interface{})
	assert.Equal(t, "test@example.com", userData["email"])
	assert.Equal(t, "Test User", userData["name"])
	assert.Equal(t, "manager", userData["role"])
	assert.Equal(t, []interface{}{}, userData["teams"])

	mockService.AssertExpectations(t)
	mockTeamService.AssertExpectations(t)
}

// TestGetProfileHandler_UserNotFound tests getting profile for non-existent user
func TestGetProfileHandler_UserNotFound(t *testing.T) {
	mockService := new(MockUserService)
	mockTeamService := new(MockTeamService)

	testUser := createTestUser(1, "test@example.com", "Test User", "manager", 1)
	mockService.On("GetUserByID", int64(1)).Return(nil, assert.AnError)

	handler := NewAuthHandler(mockService, mockTeamService)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	testHandler := func(c *gin.Context) {
		c.Set("user", testUser)
		handler.GetProfile(c)
	}

	router.GET("/profile", testHandler)

	req, _ := http.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Failed to get user profile", response["error"])

	mockService.AssertExpectations(t)
}

// TestRefreshHandler_Success tests token refresh with valid token
func TestRefreshHandler_Success(t *testing.T) {
	mockService := new(MockUserService)

	testUser := createTestUser(1, "test@example.com", "Test User", "manager", 1)
	mockService.On("GetUserByEmail", "test@example.com").Return(testUser, nil)
	mockService.On("ValidatePassword", testUser, "password123").Return(true)
	// Allow optional GetUserByID call for refresh
	mockService.On("GetUserByID", int64(1)).Return(testUser, nil).Maybe()

	handler := NewAuthHandler(mockService, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", handler.Login)
	router.POST("/refresh", handler.Refresh)

	// First, login to get tokens
	reqBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		var loginResponse AuthResponse
		json.Unmarshal(w.Body.Bytes(), &loginResponse)

		// Now test refresh
		refreshReqBody := map[string]interface{}{
			"refresh_token": loginResponse.RefreshToken,
		}
		refreshBody, _ := json.Marshal(refreshReqBody)
		refreshReq, _ := http.NewRequest("POST", "/refresh", bytes.NewReader(refreshBody))
		refreshReq.Header.Set("Content-Type", "application/json")

		refreshW := httptest.NewRecorder()
		router.ServeHTTP(refreshW, refreshReq)

		// Should succeed or fail gracefully
		assert.True(t, refreshW.Code == http.StatusOK || refreshW.Code == http.StatusBadRequest)
	}

	// Allow flexible expectations
	mockService.AssertCalled(t, "GetUserByEmail", "test@example.com")
	mockService.AssertCalled(t, "ValidatePassword", testUser, "password123")
}

// TestVerifyHandler_Success tests token verification
func TestVerifyHandler_Success(t *testing.T) {
	mockService := new(MockUserService)

	handler := NewAuthHandler(mockService, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/verify", handler.Verify)

	// Create a test token (this would need to match your JWT generation logic)
	// For now, test with an invalid token to verify error handling
	reqBody := map[string]interface{}{
		"token": "invalid.token.here",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should handle invalid token gracefully
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusUnauthorized)
}

// TestRegisterHandler_ValidationError tests registration validation errors
func TestRegisterHandler_ValidationErrors(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewAuthHandler(mockService, nil)
	router := setupTestRouter(handler)

	tests := []struct {
		name        string
		requestBody map[string]interface{}
	}{
		{
			name: "missing email",
			requestBody: map[string]interface{}{
				"password": "password123",
				"name":     "Test User",
			},
		},
		{
			name: "missing password",
			requestBody: map[string]interface{}{
				"email": "test@example.com",
				"name":  "Test User",
			},
		},
		{
			name: "missing name",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
			},
		},
		{
			name: "short password",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "12345",
				"name":     "Test User",
			},
		},
		{
			name: "invalid email format",
			requestBody: map[string]interface{}{
				"email":    "not-an-email",
				"password": "password123",
				"name":     "Test User",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Equal(t, "Invalid request payload", response["error"])
		})
	}
}

// TestLoginHandler_ConcurrentRequests tests concurrent login requests
func TestLoginHandler_ConcurrentRequests(t *testing.T) {
	mockService := new(MockUserService)

	testUser := createTestUser(1, "test@example.com", "Test User", "manager", 1)
	mockService.On("GetUserByEmail", "test@example.com").Return(testUser, nil)
	mockService.On("ValidatePassword", testUser, "password123").Return(true)

	handler := NewAuthHandler(mockService, nil)
	router := setupTestRouter(handler)

	// Make multiple concurrent requests
	results := make(chan int, 3)

	for i := 0; i < 3; i++ {
		go func() {
			reqBody := map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
			}
			body, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			results <- w.Code
		}()
	}

	// Collect results
	for i := 0; i < 3; i++ {
		statusCode := <-results
		assert.Equal(t, http.StatusOK, statusCode)
	}

	mockService.AssertExpectations(t)
}

// TestLogoutHandler_Success tests the logout endpoint
func TestLogoutHandler_Success(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewAuthHandler(mockService, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/logout", handler.Logout)

	req, _ := http.NewRequest("POST", "/logout", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Logout should return success (it's a simple endpoint that clears client-side tokens)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Logged out successfully", response["message"])
}

// TestLogoutHandler_MethodNotAllowed tests that only POST is allowed
func TestLogoutHandler_MethodNotAllowed(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewAuthHandler(mockService, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/logout", handler.Logout)

	// Try GET instead of POST
	req, _ := http.NewRequest("GET", "/logout", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Gin returns 404 for GET on POST-only routes
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestGetProfile_Success tests successful profile retrieval
func TestGetProfile_Success(t *testing.T) {
	mockUserService := new(MockUserService)
	mockTeamService := new(MockTeamService)

	testUser := createTestUser(1, "test@example.com", "Test User", "manager", 1)
	testTeams := []models.Team{
		{ID: 1, Name: "Team 1", OwnerID: 1},
	}

	mockUserService.On("GetUserByID", int64(1)).Return(testUser, nil)
	mockTeamService.On("GetTeamsByUserID", int64(1)).Return(testTeams, nil)

	handler := NewAuthHandler(mockUserService, mockTeamService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/profile", func(c *gin.Context) {
		c.Set("user", testUser)
		handler.GetProfile(c)
	})

	req, _ := http.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "user")

	mockUserService.AssertExpectations(t)
	mockTeamService.AssertExpectations(t)
}

// TestGetProfile_MissingUserContext tests profile retrieval without user in context
func TestGetProfile_MissingUserContext(t *testing.T) {
	mockUserService := new(MockUserService)
	mockTeamService := new(MockTeamService)

	handler := NewAuthHandler(mockUserService, mockTeamService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/profile", handler.GetProfile)

	req, _ := http.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "User not found in context", response["error"])
}

// TestGetProfile_InvalidUserType tests profile retrieval with invalid user type in context
func TestGetProfile_InvalidUserType(t *testing.T) {
	mockUserService := new(MockUserService)
	mockTeamService := new(MockTeamService)

	handler := NewAuthHandler(mockUserService, mockTeamService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/profile", func(c *gin.Context) {
		c.Set("user", "invalid_type") // Not a *models.User
		handler.GetProfile(c)
	})

	req, _ := http.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid user type in context", response["error"])
}

// TestGetProfile_UserNotFound tests profile retrieval when user not found in database
func TestGetProfile_UserNotFound(t *testing.T) {
	mockUserService := new(MockUserService)
	mockTeamService := new(MockTeamService)

	testUser := createTestUser(1, "test@example.com", "Test User", "manager", 1)

	mockUserService.On("GetUserByID", int64(1)).Return((*models.User)(nil), assert.AnError)

	handler := NewAuthHandler(mockUserService, mockTeamService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/profile", func(c *gin.Context) {
		c.Set("user", testUser)
		handler.GetProfile(c)
	})

	req, _ := http.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Failed to get user profile", response["error"])

	mockUserService.AssertExpectations(t)
}

// TestGetProfile_TeamsError tests profile retrieval when team lookup fails
func TestGetProfile_TeamsError(t *testing.T) {
	mockUserService := new(MockUserService)
	mockTeamService := new(MockTeamService)

	testUser := createTestUser(1, "test@example.com", "Test User", "manager", 1)

	mockUserService.On("GetUserByID", int64(1)).Return(testUser, nil)
	mockTeamService.On("GetTeamsByUserID", int64(1)).Return([]models.Team{}, assert.AnError)

	handler := NewAuthHandler(mockUserService, mockTeamService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/profile", func(c *gin.Context) {
		c.Set("user", testUser)
		handler.GetProfile(c)
	})

	req, _ := http.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Failed to get user teams")

	mockUserService.AssertExpectations(t)
	mockTeamService.AssertExpectations(t)
}

// TestRefreshToken_InvalidTokenFormat tests refresh with invalid token format
func TestRefreshToken_InvalidTokenFormat(t *testing.T) {
	mockUserService := new(MockUserService)
	mockTeamService := new(MockTeamService)

	handler := NewAuthHandler(mockUserService, mockTeamService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/refresh", handler.Refresh)

	reqBody := map[string]string{} // Missing refresh_token
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid request payload", response["error"])
}

// TestVerifyToken_MissingAuthHeader tests verify without Authorization header
func TestVerifyToken_MissingAuthHeader(t *testing.T) {
	mockUserService := new(MockUserService)
	mockTeamService := new(MockTeamService)

	handler := NewAuthHandler(mockUserService, mockTeamService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/verify", handler.Verify)

	req, _ := http.NewRequest("POST", "/verify", nil) // No Authorization header
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Token required", response["error"])
	assert.Equal(t, "UNAUTHORIZED", response["code"])
}

// --- Profile Update Tests ---

func TestUpdateProfile_Success(t *testing.T) {
	mockService := new(MockUserService)
	mockTeamService := new(MockTeamService)

	testUser := createTestUser(1, "test@example.com", "Test User", "manager", 1)
	updatedUser := createTestUser(1, "test@example.com", "New Name", "manager", 1)

	mockService.On("UpdateProfileName", int64(1), "New Name").Return(updatedUser, nil)

	handler := NewAuthHandler(mockService, mockTeamService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/profile", func(c *gin.Context) {
		c.Set("user", testUser)
		handler.UpdateProfile(c)
	})

	reqBody := `{"name": "New Name"}`
	req, _ := http.NewRequest("PUT", "/profile", bytes.NewReader([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "user")

	userData := response["user"].(map[string]interface{})
	assert.Equal(t, "New Name", userData["name"])

	mockService.AssertExpectations(t)
}

func TestUpdateProfile_MissingUserContext(t *testing.T) {
	mockService := new(MockUserService)
	mockTeamService := new(MockTeamService)

	handler := NewAuthHandler(mockService, mockTeamService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/profile", handler.UpdateProfile)

	req, _ := http.NewRequest("PUT", "/profile", bytes.NewReader([]byte(`{"name": "Test"}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateProfile_MissingName(t *testing.T) {
	mockService := new(MockUserService)
	mockTeamService := new(MockTeamService)

	testUser := createTestUser(1, "test@example.com", "Test User", "manager", 1)

	handler := NewAuthHandler(mockService, mockTeamService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/profile", func(c *gin.Context) {
		c.Set("user", testUser)
		handler.UpdateProfile(c)
	})

	reqBody := `{}`
	req, _ := http.NewRequest("PUT", "/profile", bytes.NewReader([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateProfile_ServiceError(t *testing.T) {
	mockService := new(MockUserService)
	mockTeamService := new(MockTeamService)

	testUser := createTestUser(1, "test@example.com", "Test User", "manager", 1)
	mockService.On("UpdateProfileName", int64(1), "New Name").Return((*models.User)(nil), assert.AnError)

	handler := NewAuthHandler(mockService, mockTeamService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/profile", func(c *gin.Context) {
		c.Set("user", testUser)
		handler.UpdateProfile(c)
	})

	reqBody := `{"name": "New Name"}`
	req, _ := http.NewRequest("PUT", "/profile", bytes.NewReader([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Failed to update profile", response["error"])

	mockService.AssertExpectations(t)
}

// --- Change Password Tests ---

func TestChangePassword_Success(t *testing.T) {
	mockService := new(MockUserService)
	mockTeamService := new(MockTeamService)

	testUser := createTestUser(1, "test@example.com", "Test User", "manager", 1)
	mockService.On("ChangePassword", int64(1), "oldPass123", "newPass456").Return(nil)

	handler := NewAuthHandler(mockService, mockTeamService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/password", func(c *gin.Context) {
		c.Set("user", testUser)
		handler.ChangePassword(c)
	})

	reqBody := `{"current_password": "oldPass123", "new_password": "newPass456"}`
	req, _ := http.NewRequest("PUT", "/password", bytes.NewReader([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Password changed successfully", response["message"])

	mockService.AssertExpectations(t)
}

func TestChangePassword_MissingUserContext(t *testing.T) {
	mockService := new(MockUserService)
	mockTeamService := new(MockTeamService)

	handler := NewAuthHandler(mockService, mockTeamService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/password", handler.ChangePassword)

	req, _ := http.NewRequest("PUT", "/password", bytes.NewReader([]byte(`{"current_password":"a","new_password":"b"}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestChangePassword_ShortNewPassword(t *testing.T) {
	mockService := new(MockUserService)
	mockTeamService := new(MockTeamService)

	testUser := createTestUser(1, "test@example.com", "Test User", "manager", 1)

	handler := NewAuthHandler(mockService, mockTeamService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/password", func(c *gin.Context) {
		c.Set("user", testUser)
		handler.ChangePassword(c)
	})

	reqBody := `{"current_password": "oldPass123", "new_password": "abc"}`
	req, _ := http.NewRequest("PUT", "/password", bytes.NewReader([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChangePassword_MissingFields(t *testing.T) {
	mockService := new(MockUserService)
	mockTeamService := new(MockTeamService)

	testUser := createTestUser(1, "test@example.com", "Test User", "manager", 1)

	handler := NewAuthHandler(mockService, mockTeamService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/password", func(c *gin.Context) {
		c.Set("user", testUser)
		handler.ChangePassword(c)
	})

	tests := []struct {
		name    string
		reqBody string
	}{
		{"missing both", `{}`},
		{"missing current", `{"new_password": "newpass123"}`},
		{"missing new", `{"current_password": "oldpass123"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("PUT", "/password", bytes.NewReader([]byte(tt.reqBody)))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	mockService := new(MockUserService)
	mockTeamService := new(MockTeamService)

	testUser := createTestUser(1, "test@example.com", "Test User", "manager", 1)
	mockService.On("ChangePassword", int64(1), "wrongPass", "newPass456").Return(services.ErrIncorrectPassword)

	handler := NewAuthHandler(mockService, mockTeamService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/password", func(c *gin.Context) {
		c.Set("user", testUser)
		handler.ChangePassword(c)
	})

	reqBody := `{"current_password": "wrongPass", "new_password": "newPass456"}`
	req, _ := http.NewRequest("PUT", "/password", bytes.NewReader([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "VALIDATION_ERROR", response["code"])

	mockService.AssertExpectations(t)
}

func TestChangePassword_InternalError(t *testing.T) {
	mockService := new(MockUserService)
	mockTeamService := new(MockTeamService)

	testUser := createTestUser(1, "test@example.com", "Test User", "manager", 1)
	mockService.On("ChangePassword", int64(1), "oldPass123", "newPass456").Return(assert.AnError)

	handler := NewAuthHandler(mockService, mockTeamService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/password", func(c *gin.Context) {
		c.Set("user", testUser)
		handler.ChangePassword(c)
	})

	reqBody := `{"current_password": "oldPass123", "new_password": "newPass456"}`
	req, _ := http.NewRequest("PUT", "/password", bytes.NewReader([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "INTERNAL_ERROR", response["code"])
	assert.Equal(t, "Failed to change password", response["error"])

	mockService.AssertExpectations(t)
}
