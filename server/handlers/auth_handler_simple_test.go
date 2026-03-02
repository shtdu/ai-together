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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"switch-server/models"
)

// Helper function to generate a valid test token
func generateTestToken(userID int64, email, role string, tenantID int64, duration time.Duration) string {
	expirationTime := time.Now().Add(duration)
	claims := &jwt.MapClaims{
		"user_id":   userID,
		"email":     email,
		"role":      role,
		"tenant_id": tenantID,
		"exp":       expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("default_secret_key_for_development"))
	return tokenString
}

// TestGenerateToken tests the token generation helper function
func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name        string
		userID      int64
		email       string
		role        string
		tenantID    int64
		duration    time.Duration
		expectError bool
	}{
		{
			name:        "generate access token",
			userID:      1,
			email:       "test@example.com",
			role:        "manager",
			tenantID:    1,
			duration:    24 * time.Hour,
			expectError: false,
		},
		{
			name:        "generate refresh token",
			userID:      1,
			email:       "test@example.com",
			role:        "member",
			tenantID:    2,
			duration:    7 * 24 * time.Hour,
			expectError: false,
		},
		{
			name:        "generate short-lived token",
			userID:      1,
			email:       "test@example.com",
			role:        "manager",
			tenantID:    1,
			duration:    time.Minute,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, expiration, err := generateToken(tt.userID, tt.email, tt.role, tt.tenantID, tt.duration)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				assert.True(t, expiration.After(time.Now()))
				assert.True(t, expiration.Before(time.Now().Add(tt.duration+time.Second)))

				// Verify token can be parsed and contains correct claims
				parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
					return []byte("default_secret_key_for_development"), nil
				})
				require.NoError(t, err)
				require.True(t, parsedToken.Valid)

				claims := parsedToken.Claims.(jwt.MapClaims)
				assert.Equal(t, tt.userID, int64(claims["user_id"].(float64)))
				assert.Equal(t, tt.email, claims["email"])
				assert.Equal(t, tt.role, claims["role"])
				assert.Equal(t, tt.tenantID, int64(claims["tenant_id"].(float64)))
			}
		})
	}
}

// TestLoginRequestValidation tests request validation for Login endpoint
func TestLoginRequestValidation(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "missing email",
			requestBody: map[string]interface{}{
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing password",
			requestBody: map[string]interface{}{
				"email": "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty request",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid email format",
			requestBody: map[string]interface{}{
				"email":    "not-an-email",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid json",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest, // Will panic due to nil service, but validation should pass first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			// Create a handler with nil service
			handler := &AuthHandler{userService: nil}

			// Create a wrapper that recovers from panics
			wrappedHandler := func(c *gin.Context) {
				defer func() {
					if r := recover(); r != nil {
						// If we panic due to nil service, return 500
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
					}
				}()
				handler.Login(c)
			}

			router.POST("/login", wrappedHandler)

			var body []byte
			if tt.name == "invalid json" {
				body = []byte("invalid json")
			} else {
				var err error
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req, _ := http.NewRequest("POST", "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestRegisterRequestValidation tests request validation for Register endpoint
func TestRegisterRequestValidation(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "missing email",
			requestBody: map[string]interface{}{
				"password": "password123",
				"name":     "Test User",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing password",
			requestBody: map[string]interface{}{
				"email": "test@example.com",
				"name":  "Test User",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing name",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "password too short",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "12345",
				"name":     "Test User",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid email format",
			requestBody: map[string]interface{}{
				"email":    "not-an-email",
				"password": "password123",
				"name":     "Test User",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			handler := &AuthHandler{userService: nil}

			// Create a wrapper that recovers from panics
			wrappedHandler := func(c *gin.Context) {
				defer func() {
					if r := recover(); r != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
					}
				}()
				handler.Register(c)
			}

			router.POST("/register", wrappedHandler)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestRefreshRequestValidation tests request validation for Refresh endpoint
func TestRefreshRequestValidation(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "missing refresh_token",
			requestBody: map[string]interface{}{
				"invalid": "data",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty request",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid token",
			requestBody: map[string]interface{}{
				"refresh_token": "invalid-token",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "valid token format",
			requestBody: map[string]interface{}{
				"refresh_token": generateTestToken(1, "test@example.com", "manager", 1, 7*24*time.Hour),
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			handler := &AuthHandler{userService: nil}
			router.POST("/refresh", handler.Refresh)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/refresh", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestVerifyEndpoint tests the Verify endpoint
func TestVerifyEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid authorization format",
			authHeader:     "InvalidFormat token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid token (but nil service)",
			authHeader:     "Bearer " + generateTestToken(1, "test@example.com", "manager", 1, 24*time.Hour),
			expectedStatus: http.StatusInternalServerError, // 500 because service is nil
		},
		{
			name:           "expired token",
			authHeader:     "Bearer " + generateTestToken(1, "test@example.com", "manager", 1, -1*time.Hour),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			handler := &AuthHandler{userService: nil}

			// Create a wrapper that recovers from panics
			wrappedHandler := func(c *gin.Context) {
				defer func() {
					if r := recover(); r != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
					}
				}()
				handler.Verify(c)
			}

			router.POST("/verify", wrappedHandler)

			req, _ := http.NewRequest("POST", "/verify", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestAuthResponseStruct tests the AuthResponse struct
func TestAuthResponseStruct(t *testing.T) {
	user := models.User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     "manager",
		TenantID: 1,
	}

	response := AuthResponse{
		AccessToken:  "test-token",
		RefreshToken: "test-refresh-token",
		User:         user,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	// Test JSON serialization
	data, err := json.Marshal(response)
	require.NoError(t, err)
	assert.Contains(t, string(data), "access_token")
	assert.Contains(t, string(data), "refresh_token")
	assert.Contains(t, string(data), "user")

	// Test JSON deserialization
	var decoded AuthResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, response.AccessToken, decoded.AccessToken)
	assert.Equal(t, response.RefreshToken, decoded.RefreshToken)
	assert.Equal(t, user.Email, decoded.User.Email)
}

// BenchmarkGenerateToken benchmarks the token generation
func BenchmarkGenerateToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = generateToken(1, "test@example.com", "manager", 1, 24*time.Hour)
	}
}
