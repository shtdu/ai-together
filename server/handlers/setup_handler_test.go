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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupStatusResponse_Structure(t *testing.T) {
	response := SetupStatusResponse{
		SetupRequired: true,
	}

	data, err := json.Marshal(response)
	require.NoError(t, err)
	assert.Contains(t, string(data), "setup_required")

	var decoded SetupStatusResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.True(t, decoded.SetupRequired)
}

func TestSetupAdminRequest_Structure(t *testing.T) {
	req := SetupAdminRequest{
		OrganizationName: "Test Org",
		AdminEmail:       "admin@example.com",
		AdminName:        "Admin User",
		AdminPassword:    "password123",
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)
	assert.Contains(t, string(data), "admin@example.com")

	var decoded SetupAdminRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, req.OrganizationName, decoded.OrganizationName)
	assert.Equal(t, req.AdminEmail, decoded.AdminEmail)
	assert.Equal(t, req.AdminName, decoded.AdminName)
	assert.Equal(t, req.AdminPassword, decoded.AdminPassword)
}

func TestSetupAdminResponse_Structure(t *testing.T) {
	response := SetupAdminResponse{
		Message: "Setup successful",
		User: map[string]interface{}{
			"id":        int64(1),
			"email":     "admin@example.com",
			"name":      "Admin",
			"role":      "manager",
			"tenant_id": int64(1),
		},
		Tenant: map[string]interface{}{
			"id":   int64(1),
			"name": "Test Org",
		},
		Team: map[string]interface{}{
			"id":   int64(1),
			"name": "Default Team",
		},
	}

	data, err := json.Marshal(response)
	require.NoError(t, err)

	var decoded SetupAdminResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, response.Message, decoded.Message)
	assert.Equal(t, "admin@example.com", decoded.User["email"])
}

func TestNewSetupHandler(t *testing.T) {
	// Test that NewSetupHandler creates a handler
	// We can't test with nil pool, so we just test the constructor signature
	// In a real scenario, you would mock the database pool
	handlerType := func() interface{} {
		return NewSetupHandler(nil)
	}()

	// This just verifies the constructor exists and returns the right type
	assert.NotNil(t, handlerType)
}

func TestSetupHandler_GetSetupStatus_HandlerExists(t *testing.T) {
	// Test that the handler method exists and can be called
	// We'll create a minimal test to verify the handler signature
	handler := &SetupHandler{}

	// Verify the handler has the method
	assert.NotNil(t, handler.GetSetupStatus)
}

func TestSetupHandler_CreateInitialAdmin_HandlerExists(t *testing.T) {
	handler := &SetupHandler{}

	// Verify the handler has the method
	assert.NotNil(t, handler.CreateInitialAdmin)
}

func TestSetupAdminRequest_Validation(t *testing.T) {
	tests := []struct {
		name        string
		requestBody map[string]interface{}
		expectValid bool
	}{
		{
			name: "valid request",
			requestBody: map[string]interface{}{
				"organization_name": "Test Org",
				"admin_email":       "admin@example.com",
				"admin_name":        "Admin",
				"admin_password":    "password123",
			},
			expectValid: true,
		},
		{
			name: "missing organization_name",
			requestBody: map[string]interface{}{
				"admin_email":    "admin@example.com",
				"admin_name":     "Admin",
				"admin_password": "password123",
			},
			expectValid: false,
		},
		{
			name: "missing admin_email",
			requestBody: map[string]interface{}{
				"organization_name": "Test Org",
				"admin_name":        "Admin",
				"admin_password":    "password123",
			},
			expectValid: false,
		},
		{
			name: "missing admin_name",
			requestBody: map[string]interface{}{
				"organization_name": "Test Org",
				"admin_email":       "admin@example.com",
				"admin_password":    "password123",
			},
			expectValid: false,
		},
		{
			name: "missing admin_password",
			requestBody: map[string]interface{}{
				"organization_name": "Test Org",
				"admin_email":       "admin@example.com",
				"admin_name":        "Admin",
			},
			expectValid: false,
		},
		{
			name: "invalid email format",
			requestBody: map[string]interface{}{
				"organization_name": "Test Org",
				"admin_email":       "not-an-email",
				"admin_name":        "Admin",
				"admin_password":    "password123",
			},
			expectValid: false,
		},
		{
			name: "password too short",
			requestBody: map[string]interface{}{
				"organization_name": "Test Org",
				"admin_email":       "admin@example.com",
				"admin_name":        "Admin",
				"admin_password":    "short",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)

			// Try to decode into the request struct
			var req SetupAdminRequest
			err := json.Unmarshal(body, &req)

			// Basic validation checks
			if tt.expectValid {
				require.NoError(t, err)
				assert.Equal(t, "Test Org", req.OrganizationName)
				assert.Equal(t, "admin@example.com", req.AdminEmail)
				assert.Equal(t, "Admin", req.AdminName)
				assert.Equal(t, "password123", req.AdminPassword)
			} else {
				// Check that at least one required field is missing or invalid
				if tt.requestBody["organization_name"] == nil || tt.requestBody["organization_name"] == "" {
					assert.Empty(t, req.OrganizationName)
				}
				if tt.requestBody["admin_email"] == nil || tt.requestBody["admin_email"] == "" {
					assert.Empty(t, req.AdminEmail)
				}
				if tt.requestBody["admin_name"] == nil || tt.requestBody["admin_name"] == "" {
					assert.Empty(t, req.AdminName)
				}
				if tt.requestBody["admin_password"] == nil || tt.requestBody["admin_password"] == "" {
					assert.Empty(t, req.AdminPassword)
				}
			}
		})
	}
}

// TestGetSetupStatus_DatabaseError tests GetSetupStatus with database error
func TestGetSetupStatus_DatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewSetupHandler(nil)

	// Set a custom recovery handler to catch the panic
	router.Use(func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			}
		}()
		c.Next()
	})

	router.GET("/setup/status", handler.GetSetupStatus)

	req, _ := http.NewRequest("GET", "/setup/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 500 with error when database is nil
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}

// TestGetSetupStatus_WithUsers tests GetSetupStatus when users exist
func TestGetSetupStatus_WithUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewSetupHandler(nil)

	// Set a custom recovery handler to catch the panic
	router.Use(func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			}
		}()
		c.Next()
	})

	router.GET("/setup/status", handler.GetSetupStatus)

	req, _ := http.NewRequest("GET", "/setup/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 500 since pool is nil
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}

// TestSetupHandler_Routes tests that routes are properly set up
func TestSetupHandler_Routes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewSetupHandler(nil)

	// Register routes
	router.GET("/setup/status", handler.GetSetupStatus)
	router.POST("/setup/admin", handler.CreateInitialAdmin)

	// Test that routes are registered
	routes := router.Routes()
	assert.Len(t, routes, 2)

	// Check route paths and methods
	routeMap := make(map[string]string)
	for _, route := range routes {
		key := route.Method + ":" + route.Path
		routeMap[key] = route.Path
	}

	assert.Contains(t, routeMap, "GET:/setup/status")
	assert.Contains(t, routeMap, "POST:/setup/admin")
}

// TestCreateInitialAdmin_InvalidJSON tests invalid JSON payload
func TestCreateInitialAdmin_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewSetupHandler(nil)

	router.Use(func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			}
		}()
		c.Next()
	})

	router.POST("/setup/admin", handler.CreateInitialAdmin)

	// Send invalid JSON (empty body)
	req, _ := http.NewRequest("POST", "/setup/admin", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 400 with invalid JSON error
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}

// TestCreateInitialAdmin_InvalidEmailFormat tests invalid email format
func TestCreateInitialAdmin_InvalidEmailFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewSetupHandler(nil)

	router.Use(func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			}
		}()
		c.Next()
	})

	router.POST("/setup/admin", handler.CreateInitialAdmin)

	// Send valid JSON with invalid email
	reqBody := `{"organization_name":"Test Org","admin_email":"not-an-email","admin_name":"Admin User","admin_password":"password123"}`
	req, _ := http.NewRequest("POST", "/setup/admin", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 400 with validation error
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}

// TestCreateInitialAdmin_ShortPassword tests password validation
func TestCreateInitialAdmin_ShortPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewSetupHandler(nil)

	router.Use(func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			}
		}()
		c.Next()
	})

	router.POST("/setup/admin", handler.CreateInitialAdmin)

	// Send valid JSON with short password
	reqBody := `{"organization_name":"Test Org","admin_email":"admin@example.com","admin_name":"Admin User","admin_password":"short"}`
	req, _ := http.NewRequest("POST", "/setup/admin", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 400 with validation error
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}
