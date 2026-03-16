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
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"switch-server/models"
	"switch-server/services"
)

type AnalyticsHandler struct {
	usageService services.UsageServiceInterface
	userService  services.UserServiceInterface
}

func NewAnalyticsHandler(usageService services.UsageServiceInterface, userService services.UserServiceInterface) *AnalyticsHandler {
	return &AnalyticsHandler{
		usageService: usageService,
		userService:  userService,
	}
}

// GetProviderAnalytics returns provider analytics with filtering
func (h *AnalyticsHandler) GetProviderAnalytics(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	authenticatedUser := user.(*models.User)

	// Only managers can access provider analytics
	if authenticatedUser.Role != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Parse query parameters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	// Parse optional filters
	var providers []string
	if p := c.Query("providers"); p != "" {
		providers = strings.Split(p, ",")
	}

	var modelsList []string
	if m := c.Query("models"); m != "" {
		modelsList = strings.Split(m, ",")
	}

	result, err := h.usageService.GetProviderAnalytics(c.Request.Context(), authenticatedUser.TenantID, startDate, endDate, providers, modelsList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get provider analytics: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetUserAnalytics returns user analytics with filtering
func (h *AnalyticsHandler) GetUserAnalytics(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	authenticatedUser := user.(*models.User)

	// Only managers can access user analytics
	if authenticatedUser.Role != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Parse query parameters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	// Parse optional filters
	var userIDs []int64
	if u := c.Query("user_ids"); u != "" {
		for _, id := range strings.Split(u, ",") {
			if parsed, err := strconv.ParseInt(id, 10, 64); err == nil {
				userIDs = append(userIDs, parsed)
			}
		}
	}

	var providers []string
	if p := c.Query("providers"); p != "" {
		providers = strings.Split(p, ",")
	}

	result, err := h.usageService.GetUserAnalytics(c.Request.Context(), authenticatedUser.TenantID, startDate, endDate, userIDs, providers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user analytics: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetHistory returns paginated request logs
func (h *AnalyticsHandler) GetHistory(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	authenticatedUser := user.(*models.User)

	// Only managers can access full history
	if authenticatedUser.Role != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Parse query parameters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// Parse optional filters
	var userIDs []int64
	if u := c.Query("user_ids"); u != "" {
		for _, id := range strings.Split(u, ",") {
			if parsed, err := strconv.ParseInt(id, 10, 64); err == nil {
				userIDs = append(userIDs, parsed)
			}
		}
	}

	var providers []string
	if p := c.Query("providers"); p != "" {
		providers = strings.Split(p, ",")
	}

	var modelsList []string
	if m := c.Query("models"); m != "" {
		modelsList = strings.Split(m, ",")
	}

	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	result, err := h.usageService.GetHistory(c.Request.Context(), authenticatedUser.TenantID, startDate, endDate, page, limit, userIDs, providers, modelsList, sortBy, sortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get history: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetFilterOptions returns available filter options (providers, models, users)
func (h *AnalyticsHandler) GetFilterOptions(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	authenticatedUser := user.(*models.User)

	result, err := h.usageService.GetFilterOptions(c.Request.Context(), authenticatedUser.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get filter options: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
