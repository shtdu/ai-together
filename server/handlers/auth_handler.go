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
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"switch-server/models"
	"switch-server/services"
)

type AuthHandler struct {
	userService services.UserServiceInterface
	teamService services.TeamServiceInterface
	jwtSecret  string
}

func NewAuthHandler(userService services.UserServiceInterface, teamService services.TeamServiceInterface, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		teamService: teamService,
		jwtSecret:  jwtSecret,
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type VerifyRequest struct {
	AccessToken string `json:"access_token"`
}

type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         models.User `json:"user"`
	ExpiresAt    time.Time   `json:"expires_at"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request payload",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Get user details from the service
	user, err := h.userService.GetUserByEmail(req.Email)
	if err != nil {
		requestID := c.GetString("request_id")
		log.Printf("[LOGIN_FAILED] [%s] Email=%s error=user_not_found", requestID, req.Email)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Invalid credentials",
			Code:    models.ErrCodeUnauthorized,
			Request: requestID,
		})
		return
	}

	// Validate the provided password
	if !h.userService.ValidatePassword(user, req.Password) {
		requestID := c.GetString("request_id")
		log.Printf("[LOGIN_FAILED] [%s] Email=%s user_id=%d error=invalid_password", requestID, req.Email, user.ID)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Invalid credentials",
			Code:    models.ErrCodeUnauthorized,
			Request: requestID,
		})
		return
	}

	// Generate tokens
	accessToken, accessExp, err := generateToken(user.ID, user.Email, user.Role, user.TenantID, 24*time.Hour, h.jwtSecret)
	if err != nil {
		requestID := c.GetString("request_id")
		log.Printf("[ERROR] [%s] Failed to generate access token - user_id=%d error=%v", requestID, user.ID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to generate access token",
			Code:    models.ErrCodeInternal,
			Request: requestID,
		})
		return
	}

	refreshToken, _, err := generateToken(user.ID, user.Email, user.Role, user.TenantID, 7*24*time.Hour, h.jwtSecret)
	if err != nil {
		requestID := c.GetString("request_id")
		log.Printf("[ERROR] [%s] Failed to generate refresh token - user_id=%d error=%v", requestID, user.ID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to generate refresh token",
			Code:    models.ErrCodeInternal,
			Request: requestID,
		})
		return
	}

	response := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
		ExpiresAt:    accessExp,
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request payload",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// For now, assume a default tenant_id of 1 for new users
	// In a real application, you'd either create a new tenant or assign to an appropriate one
	tenantID := int64(1)

	// Create user via service
	user, err := h.userService.CreateUser(req.Email, req.Password, req.Name, "member", tenantID)
	if err != nil {
		requestID := c.GetString("request_id")
		log.Printf("[ERROR] [%s] Failed to create user - email=%s tenant_id=%d error=%v", requestID, req.Email, tenantID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to create user",
			Code:    models.ErrCodeInternal,
			Details: err.Error(),
			Request: requestID,
		})
		return
	}

	// Generate tokens
	accessToken, accessExp, err := generateToken(user.ID, user.Email, user.Role, user.TenantID, 24*time.Hour, h.jwtSecret)
	if err != nil {
		requestID := c.GetString("request_id")
		log.Printf("[ERROR] [%s] Failed to generate access token - user_id=%d error=%v", requestID, user.ID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to generate access token",
			Code:    models.ErrCodeInternal,
			Request: requestID,
		})
		return
	}

	refreshToken, _, err := generateToken(user.ID, user.Email, user.Role, user.TenantID, 7*24*time.Hour, h.jwtSecret)
	if err != nil {
		requestID := c.GetString("request_id")
		log.Printf("[ERROR] [%s] Failed to generate refresh token - user_id=%d error=%v", requestID, user.ID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to generate refresh token",
			Code:    models.ErrCodeInternal,
			Request: requestID,
		})
		return
	}

	response := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
		ExpiresAt:    accessExp,
	}

	c.JSON(http.StatusCreated, response)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request payload",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Parse and validate the refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		requestID := c.GetString("request_id")
		log.Printf("[REFRESH_FAILED] [%s] error=invalid_token", requestID)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Invalid refresh token",
			Code:    models.ErrCodeUnauthorized,
			Request: requestID,
		})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		requestID := c.GetString("request_id")
		log.Printf("[REFRESH_FAILED] [%s] error=invalid_claims", requestID)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Invalid token claims",
			Code:    models.ErrCodeUnauthorized,
			Request: requestID,
		})
		return
	}

	// Extract user info from claims
	userID := int64(claims["user_id"].(float64))
	email := claims["email"].(string)
	role := claims["role"].(string)
	tenantID := int64(claims["tenant_id"].(float64))

	// Generate new tokens
	accessToken, accessExp, err := generateToken(userID, email, role, tenantID, 24*time.Hour, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, _, err := generateToken(userID, email, role, tenantID, 7*24*time.Hour, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	user := models.User{
		ID:       userID,
		Email:    email,
		Name:     "Test User", // Would retrieve from DB in real app
		Role:     role,
		TenantID: tenantID,
	}

	response := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		ExpiresAt:    accessExp,
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Verify(c *gin.Context) {
	// Extract token from body or header
	tokenString := ""

	// Try request body first (OpenAPI spec compliance)
	var req VerifyRequest
	bodyErr := c.ShouldBindJSON(&req)
	if bodyErr == nil && req.AccessToken != "" {
		tokenString = req.AccessToken
	} else {
		// Fallback to Authorization header (backward compatibility)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenString = authHeader[7:]
			} else {
				c.JSON(http.StatusUnauthorized, models.ErrorResponse{
					Error: "Invalid authorization header format",
					Code:  models.ErrCodeUnauthorized,
				})
				return
			}
		}
	}

	// Validate token presence
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Token required",
			Code:  models.ErrCodeUnauthorized,
		})
		return
	}

	// Parse and validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		requestID := c.GetString("request_id")
		log.Printf("[VERIFY_FAILED] [%s] error=invalid_token", requestID)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Invalid token",
			Code:    models.ErrCodeUnauthorized,
			Request: requestID,
		})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		requestID := c.GetString("request_id")
		log.Printf("[VERIFY_FAILED] [%s] error=invalid_claims", requestID)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Invalid token claims",
			Code:    models.ErrCodeUnauthorized,
			Request: requestID,
		})
		return
	}

	// Extract user information from claims
	userID := int64(claims["user_id"].(float64))

	// Get user from database to return complete user info
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		requestID := c.GetString("request_id")
		log.Printf("[VERIFY_FAILED] [%s] user_id=%d error=user_not_found", requestID, userID)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "User not found",
			Code:    models.ErrCodeNotFound,
			Request: requestID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"name":       user.Name,
			"role":       user.Role,
			"tenant_id":  user.TenantID,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Since JWT is stateless, logout primarily happens on client-side
	// by removing stored tokens. This endpoint confirms the logout request.
	// In a production environment, you might want to implement token blacklisting
	// using Redis or a similar store.
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Get user from context (set by AuthMiddleware)
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeUnauthorized,
		})
		return
	}

	authenticatedUser, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Invalid user type in context",
			Code:  models.ErrCodeInternal,
		})
		return
	}

	// Get full user details from database
	fullUser, err := h.userService.GetUserByID(authenticatedUser.ID)
	if err != nil {
		requestID := c.GetString("request_id")
		log.Printf("[ERROR] [%s] Failed to get user profile - user_id=%d error=%v", requestID, authenticatedUser.ID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to get user profile",
			Code:    models.ErrCodeInternal,
			Request: requestID,
		})
		return
	}

	// Get teams for the user
	teams, err := h.teamService.GetTeamsByUserID(authenticatedUser.ID)
	if err != nil {
		requestID := c.GetString("request_id")
		log.Printf("[ERROR] [%s] Failed to get user teams - user_id=%d error=%v", requestID, authenticatedUser.ID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to get user teams",
			Code:    models.ErrCodeInternal,
			Request: requestID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":        fullUser.ID,
			"email":     fullUser.Email,
			"name":      fullUser.Name,
			"role":      fullUser.Role,
			"tenant_id": fullUser.TenantID,
			"teams":     teams,
		},
	})
}

// UpdateProfileRequest is the request body for updating the user's own profile
type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required,min=1"`
}

// UpdateProfile updates the authenticated user's profile (name)
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeUnauthorized,
		})
		return
	}
	authenticatedUser := user.(*models.User)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Code:    models.ErrCodeValidation,
			Details: err.Error(),
		})
		return
	}

	updatedUser, err := h.userService.UpdateProfileName(authenticatedUser.ID, req.Name)
	if err != nil {
		requestID := c.GetString("request_id")
		log.Printf("[ERROR] [%s] Failed to update profile - user_id=%d error=%v", requestID, authenticatedUser.ID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to update profile",
			Code:    models.ErrCodeInternal,
			Request: requestID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": updatedUser})
}

// ChangePasswordRequest is the request body for changing the user's password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword changes the authenticated user's password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeUnauthorized,
		})
		return
	}
	authenticatedUser := user.(*models.User)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Code:    models.ErrCodeValidation,
			Details: err.Error(),
		})
		return
	}

	err := h.userService.ChangePassword(authenticatedUser.ID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		requestID := c.GetString("request_id")
		if errors.Is(err, services.ErrIncorrectPassword) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   err.Error(),
				Code:    models.ErrCodeValidation,
				Request: requestID,
			})
			return
		}
		log.Printf("[ERROR] [%s] Failed to change password - user_id=%d error=%v", requestID, authenticatedUser.ID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to change password",
			Code:    models.ErrCodeInternal,
			Request: requestID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// Helper function to generate JWT tokens
func generateToken(userID int64, email, role string, tenantID int64, duration time.Duration, jwtSecret string) (string, time.Time, error) {
	expirationTime := time.Now().Add(duration)
	claims := &jwt.MapClaims{
		"user_id":   userID,
		"email":     email,
		"role":      role,
		"tenant_id": tenantID,
		"exp":       expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}
