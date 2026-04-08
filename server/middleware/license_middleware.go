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
	"net/http"

	"switch-server/models"
	"switch-server/services"

	"github.com/gin-gonic/gin"
)

// LicenseMiddleware injects the effective license into the request context
// Does NOT block requests - falls back to default license if no valid license
func LicenseMiddleware(licenseService services.LicenseServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.Next()
			return
		}

		authenticatedUser, ok := user.(*models.User)
		if !ok {
			c.Next()
			return
		}

		// GetEffectiveLicense returns default if no valid license (never blocks)
		license := licenseService.GetEffectiveLicense(c.Request.Context(), authenticatedUser.TenantID)
		c.Set("license", license)

		// Optionally set a flag indicating if using default/fallback license
		c.Set("has_active_license", licenseService.HasActiveLicense(c.Request.Context(), authenticatedUser.TenantID))

		c.Next()
	}
}

// RequireCommercial returns a middleware that blocks requests from tenants
// without an active commercial license. Must be used after LicenseMiddleware.
func RequireCommercial() gin.HandlerFunc {
	return func(c *gin.Context) {
		hasActive, exists := c.Get("has_active_license")
		if !exists || hasActive == false {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "This feature requires a commercial license",
				Code:    "license_required",
				Details: "Please activate a commercial license to access this feature",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
