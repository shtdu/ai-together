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

	"github.com/gin-gonic/gin"
	"switch-server/models"
	"switch-server/services"
)

type UsageHandler struct {
	usageService services.UsageServiceInterface
}

func NewUsageHandler(usageService services.UsageServiceInterface) *UsageHandler {
	return &UsageHandler{
		usageService: usageService,
	}
}

func (h *UsageHandler) GetCurrentUsage(c *gin.Context) {
	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeInternal,
		})
		return
	}
	currentUser := user.(*models.User)

	// Get current usage from the service
	usage, err := h.usageService.GetCurrentUsage(c.Request.Context(), currentUser.ID, currentUser.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current usage: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, usage)
}

func (h *UsageHandler) GetUsageStats(c *gin.Context) {
	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeInternal,
		})
		return
	}
	currentUser := user.(*models.User)

	// Get usage stats from the service
	stats, err := h.usageService.GetUsageStats(c.Request.Context(), currentUser.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get usage stats: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// BatchUsageRequest represents a batch of usage records
type BatchUsageRequest struct {
	Records []models.UsageRecord `json:"records" binding:"required"`
}

// BatchUsageResponse represents the response from batch upload
type BatchUsageResponse struct {
	SyncedCount int      `json:"synced_count"`
	Errors      []string `json:"errors"`
}

// CreateBatchUsageRecords handles batch upload of usage records
func (h *UsageHandler) CreateBatchUsageRecords(c *gin.Context) {
	var records []models.UsageRecord
	if err := c.ShouldBindJSON(&records); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request payload",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	if len(records) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "No records provided",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Get user from context for validation
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeUnauthorized,
		})
		return
	}
	currentUser := user.(*models.User)

	// Validate and insert records
	ctx := context.Background()
	syncedCount := 0
	errors := make([]string, 0)

	for _, record := range records {
		// Validate that TenantID and UserID match the authenticated user
		if record.TenantID != currentUser.TenantID {
			errors = append(errors, fmt.Sprintf("record %d: tenant_id mismatch", record.ID))
			continue
		}
		if record.UserID != currentUser.ID {
			errors = append(errors, fmt.Sprintf("record %d: user_id mismatch", record.ID))
			continue
		}

		// Create the usage record
		if err := h.usageService.CreateUsageRecord(ctx, record); err != nil {
			errors = append(errors, fmt.Sprintf("record %d: failed to create: %v", record.ID, err))
			continue
		}

		syncedCount++
	}

	// Return response
	response := BatchUsageResponse{
		SyncedCount: syncedCount,
		Errors:      errors,
	}

	// Always return OK, even if all records failed validation
	// The errors array will contain details about what went wrong
	c.JSON(http.StatusOK, response)
}
