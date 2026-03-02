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
	"github.com/stretchr/testify/require"
	"switch-server/models"
)

func TestUserHandler_ListUsers_Success(t *testing.T) {
	mockService := new(MockUserService)
	testUsers := []models.User{
		*createTestUser(1, "user1@example.com", "User 1", "manager", 1),
		*createTestUser(2, "user2@example.com", "User 2", "member", 1),
	}

	mockService.On("ListUsersByTenant", int64(1)).Return(testUsers, nil)

	handler := NewUserHandler(mockService)
	router := setupUserRouter(handler)

	// Create a manager user context
	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()

	// Set up the request with user context
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "users")

	mockService.AssertExpectations(t)
}

func TestUserHandler_ListUsers_AccessDenied_Member(t *testing.T) {
	mockService := new(MockUserService)
	
	// Should not be called due to access denial

	handler := NewUserHandler(mockService)
	router := setupUserRouter(handler)

	memberUser := createTestUser(1, "member@example.com", "Member", "member", 1)

	req, _ := http.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, memberUser))

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Access denied", response["error"])
}

func TestUserHandler_CreateUser_Success(t *testing.T) {
	mockService := new(MockUserService)
	
	testUser := createTestUser(2, "newuser@example.com", "New User", "member", 1)

	mockService.On("CreateUser", "newuser@example.com", "password123", "New User", "member", int64(1)).
		Return(testUser, nil)

	handler := NewUserHandler(mockService)
	router := setupUserRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	reqBody := CreateUserRequest{
		Email:    "newuser@example.com",
		Name:     "New User",
		Password: "password123",
		Role:     "member",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "newuser@example.com", response.Email)

	mockService.AssertExpectations(t)
}

func TestUserHandler_CreateUser_ValidationError(t *testing.T) {
	mockService := new(MockUserService)
	
	handler := NewUserHandler(mockService)
	router := setupUserRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	tests := []struct {
		name        string
		requestBody map[string]interface{}
	}{
		{
			name: "missing email",
			requestBody: map[string]interface{}{
				"name":     "Test User",
				"password": "password123",
				"role":     "member",
			},
		},
		{
			name: "missing name",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
				"role":     "member",
			},
		},
		{
			name: "missing password",
			requestBody: map[string]interface{}{
				"email": "test@example.com",
				"name":  "Test User",
				"role":  "member",
			},
		},
		{
			name: "invalid email",
			requestBody: map[string]interface{}{
				"email":    "not-an-email",
				"name":     "Test User",
				"password": "password123",
				"role":     "member",
			},
		},
		{
			name: "short password",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"name":     "Test User",
				"password": "12345",
				"role":     "member",
			},
		},
		{
			name: "invalid role",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"name":     "Test User",
				"password": "password123",
				"role":     "admin",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/users", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, setUserContext(req, managerUser))

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestUserHandler_UpdateUser_Success(t *testing.T) {
	mockService := new(MockUserService)
	
	testUser := createTestUser(2, "user@example.com", "Updated Name", "member", 1)

	mockService.On("GetUserByID", int64(2)).Return(testUser, nil)
	mockService.On("UpdateUser", int64(2), "Updated Name", "manager", "").Return(testUser, nil)

	handler := NewUserHandler(mockService)
	router := setupUserRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	reqBody := UpdateUserRequest{
		Name: "Updated Name",
		Role: "manager",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/users/2", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestUserHandler_UpdateUser_AccessDenied_Member(t *testing.T) {
	mockService := new(MockUserService)
	
	handler := NewUserHandler(mockService)
	router := setupUserRouter(handler)

	memberUser := createTestUser(1, "member@example.com", "Member", "member", 1)

	reqBody := UpdateUserRequest{
		Name: "Updated Name",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/users/2", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, memberUser))

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUserHandler_DeleteUser_Success(t *testing.T) {
	mockService := new(MockUserService)
	
	testUser := createTestUser(2, "user@example.com", "User", "member", 1)

	mockService.On("GetUserByID", int64(2)).Return(testUser, nil)
	mockService.On("DeleteUser", int64(2)).Return(nil)

	handler := NewUserHandler(mockService)
	router := setupUserRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("DELETE", "/users/2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "User deleted successfully", response["message"])

	mockService.AssertExpectations(t)
}

func TestUserHandler_DeleteUser_PreventSelfDeletion(t *testing.T) {
	mockService := new(MockUserService)
	
	handler := NewUserHandler(mockService)
	router := setupUserRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("DELETE", "/users/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Cannot delete yourself", response["error"])
}

func TestUserHandler_DeleteUser_AccessDenied_Member(t *testing.T) {
	mockService := new(MockUserService)
	
	handler := NewUserHandler(mockService)
	router := setupUserRouter(handler)

	memberUser := createTestUser(1, "member@example.com", "Member", "member", 1)

	req, _ := http.NewRequest("DELETE", "/users/2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, memberUser))

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUserHandler_DeleteUser_UserNotFound(t *testing.T) {
	mockService := new(MockUserService)
	
	mockService.On("GetUserByID", int64(999)).Return(nil, assert.AnError)

	handler := NewUserHandler(mockService)
	router := setupUserRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("DELETE", "/users/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockService.AssertExpectations(t)
}

// Helper function to set up user router with user context middleware
func setupUserRouter(handler *UserHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Middleware to set user context (simulates auth middleware)
	router.Use(func(c *gin.Context) {
		userEmail := c.GetHeader("X-Test-User-Email")
		if userEmail != "" {
			userID := c.GetHeader("X-Test-User-ID")
			userRole := c.GetHeader("X-Test-User-Role")
			tenantID := c.GetHeader("X-Test-User-Tenant-ID")

			// Parse IDs from headers
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

	router.GET("/users", handler.ListUsers)
	router.POST("/users", handler.CreateUser)
	router.PUT("/users/:id", handler.UpdateUser)
	router.DELETE("/users/:id", handler.DeleteUser)
	return router
}

// Helper function to set user context in request
func setUserContext(req *http.Request, user *models.User) *http.Request {
	req.Header.Set("X-Test-User-Email", user.Email)
	req.Header.Set("X-Test-User-Role", user.Role)
	req.Header.Set("X-Test-User-ID", fmt.Sprintf("%d", user.ID))
	req.Header.Set("X-Test-User-Tenant-ID", fmt.Sprintf("%d", user.TenantID))
	return req
}

// Test helper that simulates auth middleware setting user context
func testHandlerWithUser(handler gin.HandlerFunc, user *models.User) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user", user)
		handler(c)
	}
}

// TestUserHandler_MissingUserContext tests handler behavior when user is not in context
func TestUserHandler_MissingUserContext(t *testing.T) {
	mockService := new(MockUserService)
	
	handler := NewUserHandler(mockService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/users", func(c *gin.Context) {
		// Don't set user in context
		handler.ListUsers(c)
	})

	req, _ := http.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestCreateUserRequest_Structure tests the request structure
func TestCreateUserRequest_Structure(t *testing.T) {
	req := CreateUserRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
		Role:     "manager",
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)
	assert.Contains(t, string(data), "test@example.com")

	var decoded CreateUserRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, req.Email, decoded.Email)
}

// TestUpdateUserRequest_Structure tests the update request structure
func TestUpdateUserRequest_Structure(t *testing.T) {
	req := UpdateUserRequest{
		Name:     "Updated Name",
		Role:     "member",
		Password: "newpassword",
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var decoded UpdateUserRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, req.Name, decoded.Name)
}

// TestUserHandler_CreateUser_InvalidRole tests user creation with invalid role
func TestUserHandler_CreateUser_InvalidRole(t *testing.T) {
	mockService := new(MockUserService)
	
	handler := NewUserHandler(mockService)
	router := setupUserRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	reqBody := map[string]interface{}{
		"email":    "newuser@example.com",
		"name":     "New User",
		"password": "password123",
		"role":     "invalid_role",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestUserHandler_CreateUser_InvalidEmail tests user creation with invalid email
func TestUserHandler_CreateUser_InvalidEmail(t *testing.T) {
	mockService := new(MockUserService)
	
	handler := NewUserHandler(mockService)
	router := setupUserRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	reqBody := map[string]interface{}{
		"email":    "not-an-email",
		"name":     "New User",
		"password": "password123",
		"role":     "member",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestUserHandler_UpdateUser_UserNotFound tests updating a non-existent user
func TestUserHandler_UpdateUser_UserNotFound(t *testing.T) {
	mockService := new(MockUserService)
	

	// User not found
	mockService.On("GetUserByID", int64(999)).Return((*models.User)(nil), assert.AnError)

	handler := NewUserHandler(mockService)
	router := setupUserRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	reqBody := UpdateUserRequest{
		Name: "Updated Name",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/users/999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockService.AssertExpectations(t)
}

// TestUserHandler_UpdateUser_DifferentTenant tests updating user from different tenant
func TestUserHandler_UpdateUser_DifferentTenant(t *testing.T) {
	mockService := new(MockUserService)
	
	targetUser := createTestUser(2, "other@example.com", "Other", "member", 999) // Different tenant

	mockService.On("GetUserByID", int64(2)).Return(targetUser, nil)

	handler := NewUserHandler(mockService)
	router := setupUserRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	reqBody := UpdateUserRequest{
		Name: "Updated Name",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/users/2", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Access denied", response["error"])

	mockService.AssertExpectations(t)
}

// TestUserHandler_DeleteUser_SelfDeletion tests that users cannot delete themselves
func TestUserHandler_DeleteUser_SelfDeletion(t *testing.T) {
	mockService := new(MockUserService)
	
	handler := NewUserHandler(mockService)
	router := setupUserRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("DELETE", "/users/1", nil) // Try to delete self
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Cannot delete yourself", response["error"])
}

// TestUserHandler_DeleteUser_DifferentTenant tests deleting user from different tenant
func TestUserHandler_DeleteUser_DifferentTenant(t *testing.T) {
	mockService := new(MockUserService)
	
	targetUser := createTestUser(2, "other@example.com", "Other", "member", 999) // Different tenant

	mockService.On("GetUserByID", int64(2)).Return(targetUser, nil)

	handler := NewUserHandler(mockService)
	router := setupUserRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("DELETE", "/users/2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Access denied", response["error"])

	mockService.AssertExpectations(t)
}

// TestUserHandler_ListUsers_MissingUserContext tests listing users without authentication
func TestUserHandler_ListUsers_MissingUserContext(t *testing.T) {
	mockService := new(MockUserService)
	
	handler := NewUserHandler(mockService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/users", handler.ListUsers)

	req, _ := http.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "User not found in context", response["error"])
}

// TestUserHandler_UpdateUser_InvalidUserID tests updating with invalid user ID
func TestUserHandler_UpdateUser_InvalidUserID(t *testing.T) {
	mockService := new(MockUserService)
	
	handler := NewUserHandler(mockService)
	router := setupUserRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	reqBody := UpdateUserRequest{
		Name: "Updated Name",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/users/invalid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid user ID", response["error"])
}

// TestUserHandler_DeleteUser_InvalidUserID tests deleting with invalid user ID
func TestUserHandler_DeleteUser_InvalidUserID(t *testing.T) {
	mockService := new(MockUserService)
	
	handler := NewUserHandler(mockService)
	router := setupUserRouter(handler)

	managerUser := createTestUser(1, "manager@example.com", "Manager", "manager", 1)

	req, _ := http.NewRequest("DELETE", "/users/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, managerUser))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid user ID", response["error"])
}
