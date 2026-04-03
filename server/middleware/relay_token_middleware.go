package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"switch-server/models"
	"switch-server/services"
)

// RelayTokenAuthMiddleware creates a middleware that authenticates requests
// using relay tokens from the Authorization: Bearer header.
// It falls through to the next auth middleware (JWT session auth) if no
// relay token is found or the token is invalid.
func RelayTokenAuthMiddleware(relayTokenService services.RelayTokenServiceInterface, rateLimiter *services.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			// No bearer token, fall through to session auth
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Quick check: JWT tokens have dots (header.payload.signature)
		// Relay tokens are hex strings without dots
		if strings.Contains(tokenString, ".") {
			// This looks like a JWT, fall through to session auth
			c.Next()
			return
		}

		requestID := c.GetString("request_id")

		// Try relay token validation
		token, err := relayTokenService.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			// Not a valid relay token, fall through to session auth
			c.Next()
			return
		}

		// Rate limiting check
		if rateLimiter != nil && !rateLimiter.Allow(token.ID) {
			slog.LogAttrs(c.Request.Context(), slog.LevelWarn, "relay token rate limited",
				slog.String("request_id", requestID),
				slog.Int64("token_id", token.ID),
				slog.Int64("user_id", token.UserID),
				slog.String("remote_addr", c.ClientIP()),
			)
			c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
				Error:   "Rate limit exceeded",
				Code:    models.ErrCodeRateLimited,
				Request: requestID,
			})
			c.Abort()
			return
		}

		// Set user context (same as JWT auth middleware)
		user := &models.User{
			ID:       token.UserID,
			TenantID: token.TenantID,
		}

		c.Set("user", user)
		c.Set("user_id", token.UserID)
		c.Set("tenant_id", token.TenantID)
		c.Set("role", "member")
		c.Set("auth_method", "relay_token")
		c.Set("userID", token.UserID)
		c.Set("tenantID", token.TenantID)

		slog.LogAttrs(c.Request.Context(), slog.LevelInfo, "relay token authentication successful",
			slog.String("request_id", requestID),
			slog.Int64("user_id", token.UserID),
			slog.Int64("tenant_id", token.TenantID),
			slog.String("remote_addr", c.ClientIP()),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		)

		c.Next()
	}
}
