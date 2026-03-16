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
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"switch-server/models"
	"switch-server/services"
)

type LicenseHandler struct {
	licenseService services.LicenseServiceInterface
}

func NewLicenseHandler(licenseService services.LicenseServiceInterface) *LicenseHandler {
	return &LicenseHandler{
		licenseService: licenseService,
	}
}

// GetLicense returns the current license status and usage for the tenant
func (h *LicenseHandler) GetLicense(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	authenticatedUser := user.(*models.User)

	license, err := h.licenseService.GetLicense(c.Request.Context(), authenticatedUser.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get license"})
		return
	}

	usage, err := h.licenseService.GetLicenseUsage(c.Request.Context(), authenticatedUser.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get license usage"})
		return
	}

	// Calculate days remaining
	daysRemaining := int64(0)
	if !license.ExpiresAt.IsZero() {
		expiresAt := license.ExpiresAt
		daysRemaining = int64(time.Until(expiresAt).Hours() / 24)
		if daysRemaining < 0 {
			daysRemaining = 0
		}
	}

	// No seat or provider limits - all users/providers can be added
	canAddUser := true
	canAddProvider := map[string]bool{"claude": true, "codex": true, "opencode": true}

	// Check team limit
	canCreateTeam, _ := h.licenseService.CanCreateTeam(c.Request.Context(), authenticatedUser.TenantID)

	// Check for active license
	hasActiveLicense := h.licenseService.HasActiveLicense(c.Request.Context(), authenticatedUser.TenantID)

	response := gin.H{
		"license": gin.H{
			"customer_name":       license.CustomerName,
			"license_id":          nilOrString(license.LicenseID),
			"type":                license.GetLicenseType(),
			"type_name":           license.LicenseTypeName(),
			"max_seats":           license.GetMaxSeats(),
			"max_teams":           license.GetMaxTeams(),
			"data_retention_days": license.GetDataRetentionDays(),
			"issued_at":           nilOrTime(license.IssuedAt),
			"expires_at":          nilOrTime(license.ExpiresAt),
		},
		"usage": gin.H{
			"current_users":   usage.CurrentUsers,
			"current_teams":   usage.CurrentTeams,
			"teams_remaining": usage.TeamsRemaining,
			"provider_counts": usage.ProviderCounts,
		},
		"status": gin.H{
			"has_active_license": hasActiveLicense,
			"days_remaining":     daysRemaining,
			"can_add_user":       canAddUser,
			"can_add_provider":   canAddProvider,
			"can_create_team":    canCreateTeam,
		},
	}

	// Add is_using_defaults flag if no active license
	if !hasActiveLicense {
		response["status"].(gin.H)["is_using_defaults"] = true
	}

	c.JSON(http.StatusOK, response)
}

// ActivateLicenseRequest is the request body for activating a license
type ActivateLicenseRequest struct {
	LicenseKey string `json:"license_key" binding:"required"`
}

// ActivateLicense activates or upgrades a license with a license key
func (h *LicenseHandler) ActivateLicense(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	authenticatedUser := user.(*models.User)

	var req ActivateLicenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	err := h.licenseService.ActivateLicense(c.Request.Context(), authenticatedUser.TenantID, req.LicenseKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "License activated successfully"})
}

// GetTiers returns information about available license tiers
func (h *LicenseHandler) GetTiers(c *gin.Context) {
	tiers := h.licenseService.GetTiers()
	c.JSON(http.StatusOK, tiers)
}

// Helper functions for handling nil values
func nilOrString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nilOrTime(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t
}
