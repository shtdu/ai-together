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
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type SetupHandler struct {
	pool *pgxpool.Pool
}

func NewSetupHandler(pool *pgxpool.Pool) *SetupHandler {
	return &SetupHandler{
		pool: pool,
	}
}

type SetupStatusResponse struct {
	SetupRequired bool `json:"setup_required"`
}

type SetupAdminRequest struct {
	OrganizationName string `json:"organization_name" binding:"required"`
	AdminEmail       string `json:"admin_email" binding:"required,email"`
	AdminName        string `json:"admin_name" binding:"required"`
	AdminPassword    string `json:"admin_password" binding:"required,min=8"`
}

type SetupAdminResponse struct {
	Message string                 `json:"message"`
	User    map[string]interface{} `json:"user"`
	Tenant  map[string]interface{} `json:"tenant"`
	Team    map[string]interface{} `json:"team"`
}

// GetSetupStatus checks if setup is needed by counting users
func (h *SetupHandler) GetSetupStatus(c *gin.Context) {
	ctx := context.Background()

	var userCount int64
	err := h.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check setup status"})
		return
	}

	c.JSON(http.StatusOK, SetupStatusResponse{
		SetupRequired: userCount == 0,
	})
}

// CreateInitialAdmin creates the initial admin account, tenant, and team
func (h *SetupHandler) CreateInitialAdmin(c *gin.Context) {
	var req SetupAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.AdminEmail) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	// Validate password length (already checked by binding, but double-check)
	if len(req.AdminPassword) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters"})
		return
	}

	ctx := context.Background()

	// Begin transaction
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback(ctx)

	// Re-check user count inside transaction to prevent race conditions
	var userCount int64
	err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user count"})
		return
	}

	if userCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Setup has already been completed"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create tenant
	var tenantID int64
	err = tx.QueryRow(ctx,
		"INSERT INTO tenants (name, subdomain) VALUES ($1, $2) RETURNING id",
		req.OrganizationName, "default").Scan(&tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create tenant: %v", err)})
		return
	}

	// Create admin user
	var adminUserID int64
	err = tx.QueryRow(ctx,
		"INSERT INTO users (email, name, password, role, tenant_id) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		req.AdminEmail, req.AdminName, string(hashedPassword), "manager", tenantID).Scan(&adminUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create admin user: %v", err)})
		return
	}

	// Create default team
	var teamID int64
	err = tx.QueryRow(ctx,
		"INSERT INTO teams (name, description, owner_id, tenant_id, settings) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		"Default Team", "Default team for administrator", adminUserID, tenantID, "{}").Scan(&teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create default team: %v", err)})
		return
	}

	// Add admin to team as manager
	_, err = tx.Exec(ctx,
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3)",
		teamID, adminUserID, "manager")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to add admin to team: %v", err)})
		return
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, SetupAdminResponse{
		Message: "Initial admin account created successfully",
		User: map[string]interface{}{
			"id":        adminUserID,
			"email":     req.AdminEmail,
			"name":      req.AdminName,
			"role":      "manager",
			"tenant_id": tenantID,
		},
		Tenant: map[string]interface{}{
			"id":   tenantID,
			"name": req.OrganizationName,
		},
		Team: map[string]interface{}{
			"id":   teamID,
			"name": "Default Team",
		},
	})
}
