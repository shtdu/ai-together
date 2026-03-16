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

package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEnforcer_Success(t *testing.T) {
	enforcer, err := NewEnforcer()
	require.NoError(t, err)
	assert.NotNil(t, enforcer)
}

func TestNewEnforcer_Enforce(t *testing.T) {
	enforcer, err := NewEnforcer()
	require.NoError(t, err)

	tests := []struct {
		name    string
		subject string
		object  string
		action  string
		allowed bool
	}{
		{
			name:    "member can read providers",
			subject: "member",
			object:  "providers",
			action:  "read",
			allowed: true,
		},
		{
			name:    "member cannot write providers",
			subject: "member",
			object:  "providers",
			action:  "write",
			allowed: false,
		},
		{
			name:    "member can read config",
			subject: "member",
			object:  "config",
			action:  "read",
			allowed: true,
		},
		{
			name:    "member cannot write config",
			subject: "member",
			object:  "config",
			action:  "write",
			allowed: false,
		},
		{
			name:    "member can sync config",
			subject: "member",
			object:  "config",
			action:  "sync",
			allowed: true,
		},
		{
			name:    "manager can read providers",
			subject: "manager",
			object:  "providers",
			action:  "read",
			allowed: true,
		},
		{
			name:    "manager can write providers",
			subject: "manager",
			object:  "providers",
			action:  "write",
			allowed: true,
		},
		{
			name:    "manager can read config",
			subject: "manager",
			object:  "config",
			action:  "read",
			allowed: true,
		},
		{
			name:    "manager can write config",
			subject: "manager",
			object:  "config",
			action:  "write",
			allowed: true,
		},
		{
			name:    "manager can read analytics",
			subject: "manager",
			object:  "analytics",
			action:  "read",
			allowed: true,
		},
		{
			name:    "manager can read users",
			subject: "manager",
			object:  "users",
			action:  "read",
			allowed: true,
		},
		{
			name:    "manager can write users",
			subject: "manager",
			object:  "users",
			action:  "write",
			allowed: true,
		},
		{
			name:    "member cannot access analytics",
			subject: "member",
			object:  "analytics",
			action:  "read",
			allowed: false,
		},
		{
			name:    "member cannot access users",
			subject: "member",
			object:  "users",
			action:  "read",
			allowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := enforcer.Enforce(tt.subject, tt.object, tt.action)
			require.NoError(t, err)
			assert.Equal(t, tt.allowed, allowed)
		})
	}
}

func TestRequirePermission_Success(t *testing.T) {
	enforcer, err := NewEnforcer()
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up middleware with permission check
	router.Use(func(c *gin.Context) {
		c.Set("role", "manager")
		c.Next()
	})

	testHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}

	router.Use(RequirePermission(enforcer, "providers", "read"))
	router.GET("/test", testHandler)

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequirePermission_MissingRole(t *testing.T) {
	enforcer, err := NewEnforcer()
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	testHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}

	router.Use(RequirePermission(enforcer, "providers", "read"))
	router.GET("/test", testHandler)

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "User role not found in context", response["error"])
}

func TestRequirePermission_InsufficientPermissions(t *testing.T) {
	enforcer, err := NewEnforcer()
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("role", "member")
		c.Next()
	})

	testHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}

	router.Use(RequirePermission(enforcer, "providers", "write"))
	router.GET("/test", testHandler)

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Insufficient permissions")
}

func TestRequirePermission_InvalidRoleType(t *testing.T) {
	enforcer, err := NewEnforcer()
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("role", 123) // Invalid type
		c.Next()
	})

	testHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}

	router.Use(RequirePermission(enforcer, "providers", "read"))
	router.GET("/test", testHandler)

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Invalid role type in context", response["error"])
}

func TestGetRequiredRole(t *testing.T) {
	tests := []struct {
		name     string
		object   string
		action   string
		expected string
	}{
		{
			name:     "write action requires manager",
			object:   "providers",
			action:   "write",
			expected: "manager",
		},
		{
			name:     "read action requires member",
			object:   "providers",
			action:   "read",
			expected: "member",
		},
		{
			name:     "sync action requires member",
			object:   "config",
			action:   "sync",
			expected: "member",
		},
		{
			name:     "unknown action defaults to member",
			object:   "unknown",
			action:   "unknown",
			expected: "member",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getRequiredRole(tt.object, tt.action)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRequirePermission_MemberCanSyncConfig(t *testing.T) {
	enforcer, err := NewEnforcer()
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("role", "member")
		c.Next()
	})

	testHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}

	router.Use(RequirePermission(enforcer, "config", "sync"))
	router.GET("/test", testHandler)

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Casbin Error Scenario Tests

func TestEnforce_EmptyStrings(t *testing.T) {
	enforcer, err := NewEnforcer()
	require.NoError(t, err)

	tests := []struct {
		name    string
		subject string
		object  string
		action  string
	}{
		{"empty subject", "", "providers", "read"},
		{"empty object", "member", "", "read"},
		{"empty action", "member", "providers", ""},
		{"all empty", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Casbin should handle empty strings gracefully
			allowed, err := enforcer.Enforce(tt.subject, tt.object, tt.action)
			require.NoError(t, err)
			// Empty strings should not match any policy
			assert.False(t, allowed, "Empty strings should not match any policy")
		})
	}
}

func TestEnforce_NonExistentRole(t *testing.T) {
	enforcer, err := NewEnforcer()
	require.NoError(t, err)

	tests := []struct {
		name    string
		subject string
		object  string
		action  string
	}{
		{"non-existent role - admin", "admin", "providers", "read"},
		{"non-existent role - superuser", "superuser", "users", "write"},
		{"non-existent role - guest", "guest", "config", "read"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := enforcer.Enforce(tt.subject, tt.object, tt.action)
			require.NoError(t, err)
			assert.False(t, allowed, "Non-existent roles should not have permissions")
		})
	}
}

func TestEnforce_NonExistentObject(t *testing.T) {
	enforcer, err := NewEnforcer()
	require.NoError(t, err)

	tests := []struct {
		name    string
		subject string
		object  string
		action  string
	}{
		{"non-existent object - settings", "manager", "settings", "write"},
		{"non-existent object - audit", "manager", "audit", "read"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := enforcer.Enforce(tt.subject, tt.object, tt.action)
			require.NoError(t, err)
			assert.False(t, allowed, "Non-existent objects should not have permissions")
		})
	}
}

func TestEnforce_NonExistentAction(t *testing.T) {
	enforcer, err := NewEnforcer()
	require.NoError(t, err)

	tests := []struct {
		name    string
		subject string
		object  string
		action  string
	}{
		{"non-existent action - delete", "manager", "providers", "delete"},
		{"non-existent action - admin", "manager", "users", "admin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := enforcer.Enforce(tt.subject, tt.object, tt.action)
			require.NoError(t, err)
			assert.False(t, allowed, "Non-existent actions should not have permissions")
		})
	}
}

func TestEnforce_SpecialCharactersInRole(t *testing.T) {
	enforcer, err := NewEnforcer()
	require.NoError(t, err)

	tests := []struct {
		name    string
		subject string
		object  string
		action  string
	}{
		{"role with spaces", "member ", "providers", "read"},
		{"role with special chars", "member@admin", "providers", "read"},
		{"role with unicode", "mëmbér", "providers", "read"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := enforcer.Enforce(tt.subject, tt.object, tt.action)
			require.NoError(t, err)
			// Roles with special chars should not match exact policy
			assert.False(t, allowed, "Roles with special characters should not match exact policy")
		})
	}
}

func TestEnforce_ConcurrentPermissionChecks(t *testing.T) {
	enforcer, err := NewEnforcer()
	require.NoError(t, err)

	// Test concurrent permission checks
	concurrency := 10
	results := make(chan bool, concurrency)
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			allowed, err := enforcer.Enforce("member", "providers", "read")
			results <- allowed
			errors <- err
		}()
	}

	// Collect results
	passCount := 0
	errCount := 0
	for i := 0; i < concurrency; i++ {
		if <-results {
			passCount++
		}
		if <-errors != nil {
			errCount++
		}
	}

	assert.Equal(t, 0, errCount, "No errors should occur during concurrent checks")
	assert.Equal(t, concurrency, passCount, "All concurrent checks should succeed")
}

func TestEnforce_CaseSensitivity(t *testing.T) {
	enforcer, err := NewEnforcer()
	require.NoError(t, err)

	tests := []struct {
		name    string
		subject string
		object  string
		action  string
		allowed bool
	}{
		{"uppercase role", "MEMBER", "providers", "read", false},
		{"lowercase role first", "Member", "providers", "read", false},
		{"mixed case role", "MeMbEr", "providers", "read", false},
		{"uppercase object", "member", "PROVIDERS", "read", false},
		{"uppercase action", "member", "providers", "READ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := enforcer.Enforce(tt.subject, tt.object, tt.action)
			require.NoError(t, err)
			// Casbin is case-sensitive by default
			assert.False(t, allowed, "Case-sensitive matching should not match different cases")
		})
	}
}

func TestGetRequiredRole_CaseSensitivity(t *testing.T) {
	tests := []struct {
		name     string
		object   string
		action   string
		expected string
		allowed  bool
	}{
		{
			name:     "write action requires manager (uppercase)",
			object:   "PROVIDERS",
			action:   "WRITE",
			expected: "member", // getRequiredRole is case-sensitive, WRITE != write
			allowed:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getRequiredRole(tt.object, tt.action)
			// getRequiredRole checks action == "write", which is case-sensitive
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRequirePermission_UnknownRolePolicy(t *testing.T) {
	enforcer, err := NewEnforcer()
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("role", "unknown_role")
		c.Next()
	})

	testHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}

	router.Use(RequirePermission(enforcer, "providers", "read"))
	router.GET("/test", testHandler)

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Unknown role should be forbidden
	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Insufficient permissions")
}

// TestNewEnforcer_PolicyLoading tests that policies are correctly loaded
func TestNewEnforcer_PolicyLoading(t *testing.T) {
	enforcer, err := NewEnforcer()
	require.NoError(t, err)

	// Verify that some expected policies exist
	tests := []struct {
		subject string
		object  string
		action  string
		allowed bool
	}{
		// Member policies
		{"member", "providers", "read", true},
		{"member", "providers", "write", false},
		{"member", "config", "read", true},
		{"member", "config", "write", false},
		{"member", "config", "sync", true},
		{"member", "license", "read", true},
		{"member", "license", "write", false},
		// Manager policies
		{"manager", "providers", "read", true},
		{"manager", "providers", "write", true},
		{"manager", "config", "read", true},
		{"manager", "config", "write", true},
		{"manager", "analytics", "read", true},
		{"manager", "users", "read", true},
		{"manager", "users", "write", true},
		{"manager", "license", "read", true},
		{"manager", "license", "write", true},
		// Negative cases
		{"member", "analytics", "read", false},
		{"member", "users", "read", false},
	}

	for _, tt := range tests {
		t.Run(tt.subject+"_"+tt.object+"_"+tt.action, func(t *testing.T) {
			allowed, err := enforcer.Enforce(tt.subject, tt.object, tt.action)
			require.NoError(t, err)
			assert.Equal(t, tt.allowed, allowed,
				fmt.Sprintf("Expected %s to %s %s: %v", tt.subject, tt.action, tt.object, tt.allowed))
		})
	}
}

// TestRequirePermission_WriteRequiresManager tests that write operations require manager role
func TestRequirePermission_WriteRequiresManager(t *testing.T) {
	enforcer, err := NewEnforcer()
	require.NoError(t, err)

	tests := []struct {
		name    string
		role    string
		object  string
		action  string
		allowed bool
	}{
		{"member cannot write providers", "member", "providers", "write", false},
		{"member cannot write config", "member", "config", "write", false},
		{"manager can write providers", "manager", "providers", "write", true},
		{"manager can write config", "manager", "config", "write", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			router.Use(func(c *gin.Context) {
				c.Set("role", tt.role)
				c.Next()
			})

			router.Use(RequirePermission(enforcer, tt.object, tt.action))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if tt.allowed {
				assert.Equal(t, http.StatusOK, w.Code)
			} else {
				assert.Equal(t, http.StatusForbidden, w.Code)
			}
		})
	}
}
