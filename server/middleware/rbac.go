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

package middleware

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/gin-gonic/gin"
)

//go:embed rbac_model.conf
var rbacModel []byte

// Enforcer wraps the Casbin enforcer
type Enforcer struct {
	enforcer *casbin.Enforcer
}

// NewEnforcer creates a new RBAC enforcer with predefined policies
func NewEnforcer() (*Enforcer, error) {
	// Load the model from embedded bytes
	m, err := model.NewModelFromString(string(rbacModel))
	if err != nil {
		return nil, fmt.Errorf("failed to load Casbin model: %w", err)
	}

	// Create an enforcer with an empty policy adapter (memory-based)
	enforcer, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	// Define role-based policies
	// Member role: read-only access
	policies := [][]string{
		// Member permissions
		{"member", "providers", "read"},
		{"member", "config", "read"},
		{"member", "config", "sync"},
		{"member", "license", "read"}, // Members can read license for feature toggles

		// Manager permissions: full access
		{"manager", "providers", "read"},
		{"manager", "providers", "write"},
		{"manager", "config", "read"},
		{"manager", "config", "write"},
		{"manager", "config", "sync"},
		{"manager", "analytics", "read"},
		{"manager", "users", "read"},
		{"manager", "users", "write"},
		{"manager", "license", "read"},  // Managers can read
		{"manager", "license", "write"}, // Only managers can write/activate
	}

	// Add policies to the enforcer
	if _, err := enforcer.AddPolicies(policies); err != nil {
		return nil, fmt.Errorf("failed to add policies: %w", err)
	}

	return &Enforcer{enforcer: enforcer}, nil
}

// Enforce checks if the subject has permission to perform action on object
func (e *Enforcer) Enforce(subject, object, action string) (bool, error) {
	return e.enforcer.Enforce(subject, object, action)
}

// RequirePermission returns a middleware that checks permissions using Casbin
func RequirePermission(enforcer *Enforcer, object, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (set by AuthMiddleware)
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found in context"})
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid role type in context"})
			c.Abort()
			return
		}

		// Check permission using Casbin enforcer
		allowed, err := enforcer.Enforce(roleStr, object, action)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Insufficient permissions: requires %s role for %s on %s", getRequiredRole(object, action), action, object),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getRequiredRole returns the role required for a given object and action
func getRequiredRole(object, action string) string {
	// Write operations require manager role
	if action == "write" {
		return "manager"
	}
	// Read and sync operations can be done by members
	return "member"
}
