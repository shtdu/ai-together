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
	"time"

	"github.com/gin-gonic/gin"
	"switch-server/models"
	"switch-server/services"
)

type DashboardHandler struct {
	usageService services.UsageServiceInterface
	userService  services.UserServiceInterface
}

func NewDashboardHandler(usageService services.UsageServiceInterface, userService services.UserServiceInterface) *DashboardHandler {
	return &DashboardHandler{
		usageService: usageService,
		userService:  userService,
	}
}

// MetricDataPoint represents a single data point in the time series
type MetricDataPoint struct {
	Timestamp         time.Time `json:"timestamp"`
	ActiveTimeSeconds float64   `json:"active_time_seconds"`
	TotalTokens       int64     `json:"total_tokens"`
	InputTokens       int64     `json:"input_tokens"`
	OutputTokens      int64     `json:"output_tokens"`
	RequestCount      int64     `json:"request_count"`
}

// DashboardMetricsResponse is the response for the metrics endpoint
type DashboardMetricsResponse struct {
	Period struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	} `json:"period"`
	Interval   string            `json:"interval"`
	DataPoints []MetricDataPoint `json:"data_points"`
	Summary    struct {
		TotalActiveTimeHours float64 `json:"total_active_time_hours"`
		TotalTokens          int64   `json:"total_tokens"`
		TotalRequests        int64   `json:"total_requests"`
		EstimatedCost        float64 `json:"estimated_cost"`
	} `json:"summary"`
}

// ProviderRanking represents a provider's ranking
type ProviderRanking struct {
	Rank         int     `json:"rank"`
	Provider     string  `json:"provider"`
	TotalTokens  int64   `json:"total_tokens"`
	Percentage   float64 `json:"percentage"`
	RequestCount int64   `json:"request_count"`
}

// DashboardRankingsResponse is the response for the rankings endpoint
type DashboardRankingsResponse struct {
	Period   string            `json:"period"`
	Rankings []ProviderRanking `json:"rankings"`
}

// MemberStats represents statistics for a team member
type MemberStats struct {
	UserID          int64     `json:"user_id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	ActiveTimeHours float64   `json:"active_time_hours"`
	AvgTokensPerDay float64   `json:"avg_tokens_per_day"`
	TotalTokens     int64     `json:"total_tokens"`
	LastActive      time.Time `json:"last_active"`
}

// DashboardMembersResponse is the response for the members endpoint
type DashboardMembersResponse struct {
	Members []MemberStats `json:"members"`
}

// GetMetrics returns time-series metrics for the dashboard
func (h *DashboardHandler) GetMetrics(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	authenticatedUser := user.(*models.User)

	// Parse query parameters
	rangeParam := c.DefaultQuery("range", "7d")
	interval := c.DefaultQuery("interval", "hour")
	startDateParam := c.Query("start_date")
	endDateParam := c.Query("end_date")

	// Calculate time range
	var startTime time.Time
	endTime := time.Now()

	if startDateParam != "" && endDateParam != "" {
		// Custom date range takes priority over preset range
		parsed, err := time.Parse("2006-01-02", startDateParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format, use YYYY-MM-DD"})
			return
		}
		startTime = parsed

		parsed, err = time.Parse("2006-01-02", endDateParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format, use YYYY-MM-DD"})
			return
		}
		endTime = parsed.Add(24*time.Hour - time.Second) // End of the day
	} else {
		switch rangeParam {
		case "24h":
			startTime = endTime.Add(-24 * time.Hour)
		case "7d":
			startTime = endTime.Add(-7 * 24 * time.Hour)
		case "30d":
			startTime = endTime.Add(-30 * 24 * time.Hour)
		default:
			startTime = endTime.Add(-7 * 24 * time.Hour)
		}
	}

	// Get metrics from service
	metrics, err := h.usageService.GetDashboardMetrics(c.Request.Context(), authenticatedUser.TenantID, authenticatedUser.ID, authenticatedUser.Role, startTime, endTime, interval)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metrics: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetRankings returns provider rankings for the last 24 hours
func (h *DashboardHandler) GetRankings(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	authenticatedUser := user.(*models.User)

	rankings, err := h.usageService.GetProviderRankings(c.Request.Context(), authenticatedUser.TenantID, authenticatedUser.ID, authenticatedUser.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rankings: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, rankings)
}

// GetMembers returns member statistics
func (h *DashboardHandler) GetMembers(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	authenticatedUser := user.(*models.User)

	// Only managers can see all members
	if authenticatedUser.Role != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	members, err := h.usageService.GetMemberStats(c.Request.Context(), authenticatedUser.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get member stats: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, members)
}
