package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"switch-server/models"
	"switch-server/services"
)

// RelayTokenHandler handles relay token management endpoints.
type RelayTokenHandler struct {
	relayTokenService services.RelayTokenServiceInterface
}

func NewRelayTokenHandler(relayTokenService services.RelayTokenServiceInterface) *RelayTokenHandler {
	return &RelayTokenHandler{
		relayTokenService: relayTokenService,
	}
}

// MyRelayToken generates a new relay token for the current user.
// POST /api/v1/user/relay-token
func (h *RelayTokenHandler) MyRelayToken(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeUnauthorized,
		})
		return
	}
	currentUser := user.(*models.User)

	token, rawToken, err := h.relayTokenService.GenerateToken(context.Background(), currentUser.ID, currentUser.TenantID)
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

// GetMyRelayTokenInfo returns masked token info for the current user.
// GET /api/v1/user/relay-token
func (h *RelayTokenHandler) GetMyRelayTokenInfo(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeUnauthorized,
		})
		return
	}
	currentUser := user.(*models.User)

	token, err := h.relayTokenService.GetTokenInfo(context.Background(), currentUser.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "No active relay token found",
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

// RevokeMyRelayToken revokes the current user's relay token.
// DELETE /api/v1/user/relay-token
func (h *RelayTokenHandler) RevokeMyRelayToken(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeUnauthorized,
		})
		return
	}
	currentUser := user.(*models.User)

	err := h.relayTokenService.RevokeToken(context.Background(), currentUser.ID)
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
