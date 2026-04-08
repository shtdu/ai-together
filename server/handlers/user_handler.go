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
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"switch-server/models"
	"switch-server/services"
)

type UserHandler struct {
	userService services.UserServiceInterface
}

func NewUserHandler(userService services.UserServiceInterface) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// ListUsers godoc
// @Summary      List users
// @Description  Get all users in the tenant (managers only)
// @Tags         Users,Manager
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{} "Users array"
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeUnauthorized,
		})
		return
	}
	authenticatedUser := user.(*models.User)

	// Only managers can list users
	if authenticatedUser.Role != "manager" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Access denied",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	users, err := h.userService.ListUsersByTenant(authenticatedUser.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// CreateUserRequest is the request body for creating a user
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=manager member"`
}

// CreateUser godoc
// @Summary      Create a user
// @Description  Create a new user (managers only)
// @Tags         Users,Manager
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body handlers.CreateUserRequest true "User data"
// @Success      201  {object}  models.User
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeUnauthorized,
		})
		return
	}
	authenticatedUser := user.(*models.User)

	// Only managers can create users
	if authenticatedUser.Role != "manager" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Access denied",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Code:    models.ErrCodeValidation,
			Details: err.Error(),
		})
		return
	}

	// No seat limits - users can always be added
	// License check removed per new 2-type design

	newUser, err := h.userService.CreateUser(req.Email, req.Password, req.Name, req.Role, authenticatedUser.TenantID)
	if err != nil {
		requestID := c.GetString("request_id")
		log.Printf("[ERROR] [%s] Failed to create user - email=%s tenant_id=%d error=%v", requestID, req.Email, authenticatedUser.TenantID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to create user",
			Code:    models.ErrCodeInternal,
			Details: err.Error(),
			Request: requestID,
		})
		return
	}

	c.JSON(http.StatusCreated, newUser)
}

// UpdateUserRequest is the request body for updating a user
type UpdateUserRequest struct {
	Name     string `json:"name"`
	Role     string `json:"role" binding:"omitempty,oneof=manager member"`
	Password string `json:"password" binding:"omitempty,min=6"`
}

// UpdateUser godoc
// @Summary      Update a user
// @Description  Update user information (managers only)
// @Tags         Users,Manager
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "User ID"  example(1)
// @Param        request body handlers.UpdateUserRequest true "User data"
// @Success      200  {object}  models.User
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeUnauthorized,
		})
		return
	}
	authenticatedUser := user.(*models.User)

	// Only managers can update users
	if authenticatedUser.Role != "manager" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Access denied",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Code:    models.ErrCodeValidation,
			Details: err.Error(),
		})
		return
	}

	// Verify the user belongs to the same tenant
	targetUser, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if targetUser.TenantID != authenticatedUser.TenantID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Access denied",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	updatedUser, err := h.userService.UpdateUser(userID, req.Name, req.Role, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

// DeleteUser godoc
// @Summary      Delete a user
// @Description  Delete a user (managers only, cannot delete yourself)
// @Tags         Users,Manager
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "User ID"  example(1)
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeUnauthorized,
		})
		return
	}
	authenticatedUser := user.(*models.User)

	// Only managers can delete users
	if authenticatedUser.Role != "manager" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Access denied",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Prevent self-deletion
	if userID == authenticatedUser.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete yourself"})
		return
	}

	// Verify the user belongs to the same tenant
	targetUser, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if targetUser.TenantID != authenticatedUser.TenantID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Access denied",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	err = h.userService.DeleteUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
