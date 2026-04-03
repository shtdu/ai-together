package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"switch-server/models"
	"switch-server/services"
)

// RelayTokenHandler handles admin relay token management endpoints
type RelayTokenHandler struct {
	relayTokenService services.RelayTokenServiceInterface
	userService       services.UserServiceInterface
}

func NewRelayTokenHandler(relayTokenService services.RelayTokenServiceInterface, userService services.UserServiceInterface) *RelayTokenHandler {
	return &RelayTokenHandler{
		relayTokenService: relayTokenService,
		userService:       userService,
	}
}

// GenerateRelayToken generates a new relay token for a user (admin only)
// POST /api/v1/admin/users/:id/relay-token
func (h *RelayTokenHandler) GenerateRelayToken(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid user ID",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Get the target user to verify they exist and get tenant info
	targetUser, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "User not found",
			Code:  models.ErrCodeNotFound,
		})
		return
	}

	// Verify the admin user belongs to the same tenant
	adminUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeUnauthorized,
		})
		return
	}
	currentUser := adminUser.(*models.User)

	if currentUser.TenantID != targetUser.TenantID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Cannot manage relay tokens for users in other tenants",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	// Only managers can generate relay tokens
	if currentUser.Role != "manager" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Only managers can manage relay tokens",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	token, rawToken, err := h.relayTokenService.GenerateToken(context.Background(), userID, targetUser.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to generate relay token",
			Code:  models.ErrCodeInternal,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token":      rawToken,
		"prefix":     token.TokenPrefix,
		"created_at": token.CreatedAt,
		"message":    "Save this token securely. It will not be shown again.",
	})
}

// RevokeRelayToken revokes a user's relay token (admin only)
// DELETE /api/v1/admin/users/:id/relay-token
func (h *RelayTokenHandler) RevokeRelayToken(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid user ID",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Get the target user to verify they exist and get tenant info
	targetUser, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "User not found",
			Code:  models.ErrCodeNotFound,
		})
		return
	}

	// Verify same tenant
	adminUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeUnauthorized,
		})
		return
	}
	currentUser := adminUser.(*models.User)

	if currentUser.TenantID != targetUser.TenantID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Cannot manage relay tokens for users in other tenants",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	if currentUser.Role != "manager" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Only managers can manage relay tokens",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	err = h.relayTokenService.RevokeToken(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to revoke relay token",
			Code:  models.ErrCodeInternal,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Relay token revoked successfully",
	})
}

// GetRelayTokenInfo returns masked token info for a user (admin only)
// GET /api/v1/admin/users/:id/relay-token
func (h *RelayTokenHandler) GetRelayTokenInfo(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid user ID",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Get the target user to verify they exist and get tenant info
	targetUser, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "User not found",
			Code:  models.ErrCodeNotFound,
		})
		return
	}

	// Verify same tenant
	adminUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeUnauthorized,
		})
		return
	}
	currentUser := adminUser.(*models.User)

	if currentUser.TenantID != targetUser.TenantID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Cannot view relay tokens for users in other tenants",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	if currentUser.Role != "manager" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Only managers can view relay tokens",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	token, err := h.relayTokenService.GetTokenInfo(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "No active relay token found for this user",
			Code:  models.ErrCodeNotFound,
		})
		return
	}

	response := gin.H{
		"prefix":     token.TokenPrefix,
		"created_at": token.CreatedAt,
	}
	if token.LastUsedAt != nil {
		response["last_used_at"] = token.LastUsedAt
	}

	c.JSON(http.StatusOK, response)
}
