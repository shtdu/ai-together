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

// GetProviderAnalytics godoc
// @Summary      Get provider analytics
// @Description  Get analytics data for providers with filtering options (managers only)
// @Tags         Analytics,Manager
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        start_date query string true "Start date (YYYY-MM-DD)" example(2025-01-01)
// @Param        end_date query string true "End date (YYYY-MM-DD)" example(2025-01-31)
// @Param        providers query string false "Comma-separated provider names" example(claude,codex)
// @Param        models query string false "Comma-separated model names" example(claude-3-5-sonnet,gpt-4)
// @Success      200  {object}  map[string]interface{} "Provider analytics data"
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/analytics/providers [get]
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

// GetUserAnalytics godoc
// @Summary      Get user analytics
// @Description  Get analytics data for users with filtering options (managers only)
// @Tags         Analytics,Manager
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        start_date query string true "Start date (YYYY-MM-DD)" example(2025-01-01)
// @Param        end_date query string true "End date (YYYY-MM-DD)" example(2025-01-31)
// @Param        user_ids query string false "Comma-separated user IDs" example(1,2,3)
// @Param        providers query string false "Comma-separated provider names" example(claude,codex)
// @Success      200  {object}  map[string]interface{} "User analytics data"
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/analytics/users [get]
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

// GetHistory godoc
// @Summary      Get request history
// @Description  Get paginated request logs with filtering options (managers only)
// @Tags         Analytics,Manager
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        start_date query string true "Start date (YYYY-MM-DD)" example(2025-01-01)
// @Param        end_date query string true "End date (YYYY-MM-DD)" example(2025-01-31)
// @Param        page query int false "Page number" default(1) minimum(1)
// @Param        limit query int false "Items per page" default(50) minimum(1) maximum(100)
// @Param        user_ids query string false "Comma-separated user IDs" example(1,2,3)
// @Param        providers query string false "Comma-separated provider names" example(claude,codex)
// @Param        models query string false "Comma-separated model names" example(claude-3-5-sonnet,gpt-4)
// @Param        sort_by query string false "Sort field" Enums(created_at,model,provider) default(created_at)
// @Param        sort_order query string false "Sort order" Enums(asc,desc) default(desc)
// @Success      200  {object}  map[string]interface{} "Paginated history data"
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/analytics/history [get]
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

// GetFilterOptions godoc
// @Summary      Get filter options
// @Description  Get available filter options for analytics (providers, models, users)
// @Tags         Analytics,Member
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{} "Filter options"
// @Failure      401  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/analytics/filters [get]
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
