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
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"switch-server/models"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip JWT auth if user is already authenticated (e.g., by relay token middleware)
		if _, exists := c.Get("user"); exists {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			requestID := c.GetString("request_id")
			slog.LogAttrs(c.Request.Context(), slog.LevelWarn, "authentication failed",
				slog.String("request_id", requestID),
				slog.String("error", "missing_auth_header"),
				slog.String("remote_addr", c.ClientIP()),
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "Authorization header required",
				Code:    models.ErrCodeUnauthorized,
				Request: requestID,
			})
			c.Abort()
			return
		}

		tokenString := ""
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			requestID := c.GetString("request_id")
			slog.LogAttrs(c.Request.Context(), slog.LevelWarn, "authentication failed",
				slog.String("request_id", requestID),
				slog.String("error", "invalid_header_format"),
				slog.String("remote_addr", c.ClientIP()),
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "Invalid authorization header format",
				Code:    models.ErrCodeUnauthorized,
				Request: requestID,
			})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			requestID := c.GetString("request_id")
			slog.LogAttrs(c.Request.Context(), slog.LevelWarn, "authentication failed",
				slog.String("request_id", requestID),
				slog.String("error", "invalid_token"),
				slog.String("remote_addr", c.ClientIP()),
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
				slog.String("token_error", err.Error()),
			)
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "Invalid token",
				Code:    models.ErrCodeUnauthorized,
				Request: requestID,
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			requestID := c.GetString("request_id")
			slog.LogAttrs(c.Request.Context(), slog.LevelWarn, "authentication failed",
				slog.String("request_id", requestID),
				slog.String("error", "invalid_claims"),
				slog.String("remote_addr", c.ClientIP()),
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "Invalid token claims",
				Code:    models.ErrCodeUnauthorized,
				Request: requestID,
			})
			c.Abort()
			return
		}

		// Check if token is expired
		if exp, ok := claims["exp"].(float64); ok {
			if int64(exp) < time.Now().Unix() {
				requestID := c.GetString("request_id")
				slog.LogAttrs(c.Request.Context(), slog.LevelWarn, "authentication failed",
					slog.String("request_id", requestID),
					slog.String("error", "token_expired"),
					slog.String("remote_addr", c.ClientIP()),
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.Int64("exp", int64(exp)),
				)
				c.JSON(http.StatusUnauthorized, models.ErrorResponse{
					Error:   "Token expired",
					Code:    models.ErrCodeUnauthorized,
					Request: requestID,
				})
				c.Abort()
				return
			}
		}

		// Extract user information from claims
		userID := int64(claims["user_id"].(float64))
		email := claims["email"].(string)
		role := claims["role"].(string)
		tenantID := int64(claims["tenant_id"].(float64))

		// Log successful authentication
		requestID := c.GetString("request_id")
		slog.LogAttrs(c.Request.Context(), slog.LevelInfo, "authentication successful",
			slog.String("request_id", requestID),
			slog.Int64("user_id", userID),
			slog.String("email", email),
			slog.String("role", role),
			slog.Int64("tenant_id", tenantID),
			slog.String("remote_addr", c.ClientIP()),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		)

		// Create user object to pass to handlers
		user := models.User{
			ID:       userID,
			Email:    email,
			Role:     role,
			TenantID: tenantID,
		}

		// Set user in context for handlers to access
		c.Set("user", &user)
		c.Set("user_id", userID)
		c.Set("tenant_id", tenantID)
		c.Set("role", role)

		// Update the ErrorMiddleware context
		c.Set("userID", userID)
		c.Set("tenantID", tenantID)

		c.Next()
	}
}

// TenantMiddleware adds tenant context to requests
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get tenant_id from context set by AuthMiddleware
		tenantID, exists := c.Get("tenant_id")
		if !exists {
			requestID := c.GetString("request_id")
			slog.LogAttrs(c.Request.Context(), slog.LevelError, "tenant ID not found in context",
				slog.String("request_id", requestID),
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Tenant ID not found in context",
				Code:    models.ErrCodeInternal,
				Request: requestID,
			})
			c.Abort()
			return
		}

		// Add tenant_id to context for database queries
		c.Set("tenant_id", tenantID)
		c.Next()
	}
}
