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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"switch-server/models"
	"switch-server/services"
)

type TeamHandler struct {
	teamService    services.TeamServiceInterface
	userService    services.UserServiceInterface
	licenseService services.LicenseServiceInterface
}

func NewTeamHandler(teamService services.TeamServiceInterface, userService services.UserServiceInterface, licenseService services.LicenseServiceInterface) *TeamHandler {
	return &TeamHandler{
		teamService:    teamService,
		userService:    userService,
		licenseService: licenseService,
	}
}

type CreateTeamRequest struct {
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	Settings    map[string]string `json:"settings"`
}

type UpdateTeamRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Settings    map[string]string `json:"settings"`
}

type AddMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"oneof=manager member"`
}

// ListTeams godoc
// @Summary      List teams
// @Description  Get all teams for the authenticated user
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  models.Team
// @Failure      401  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/teams [get]
func (h *TeamHandler) ListTeams(c *gin.Context) {
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

	// Get teams for the user from the service
	teams, err := h.teamService.GetTeamsByUserID(currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get teams: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, teams)
}

// CreateTeam godoc
// @Summary      Create a team
// @Description  Create a new team for the authenticated user
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body handlers.CreateTeamRequest true "Team data"
// @Success      201  {object}  models.Team
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/teams [post]
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request payload",
			Code:  models.ErrCodeValidation,
		})
		return
	}

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

	// Check team limit based on license type
	canCreate, err := h.licenseService.CanCreateTeam(c.Request.Context(), currentUser.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check license: " + err.Error()})
		return
	}
	if !canCreate {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Team limit reached. Commercial license required for multiple teams.",
			"code":  "TEAM_LIMIT_EXCEEDED",
		})
		return
	}

	// Create team via service
	team, err := h.teamService.CreateTeam(req.Name, req.Description, currentUser.ID, currentUser.TenantID, req.Settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, team)
}

// UpdateTeam godoc
// @Summary      Update a team
// @Description  Update team information (only team owner can update)
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Team ID"  example(1)
// @Param        request body handlers.UpdateTeamRequest true "Team data"
// @Success      200  {object}  models.Team
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/teams/{id} [put]
func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid team ID",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	var req UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request payload",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Get user from context to check permissions
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeInternal,
		})
		return
	}
	currentUser := user.(*models.User)

	// Check if user has permission to update this team
	// Verify that the user is a manager of this team
	team, err := h.teamService.GetTeamByID(teamID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Team not found",
			Code:  models.ErrCodeNotFound,
		})
		return
	}

	// Only allow updating if user is the owner of the team
	if team.OwnerID != currentUser.ID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "You don't have permission to update this team",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	// Update the team
	err = h.teamService.UpdateTeam(teamID, req.Name, req.Description, req.Settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team: " + err.Error()})
		return
	}

	// Get the updated team
	updatedTeam, err := h.teamService.GetTeamByID(teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated team: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedTeam)
}

// DeleteTeam godoc
// @Summary      Delete a team
// @Description  Delete a team (only team owner can delete)
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Team ID"  example(1)
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/teams/{id} [delete]
func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid team ID",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Get user from context to check permissions
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeInternal,
		})
		return
	}
	currentUser := user.(*models.User)

	// Check if user has permission to delete this team
	// Verify that the user is the owner of this team
	team, err := h.teamService.GetTeamByID(teamID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Team not found",
			Code:  models.ErrCodeNotFound,
		})
		return
	}

	// Only allow deletion if user is the owner of the team
	if team.OwnerID != currentUser.ID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "You don't have permission to delete this team",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	// Delete the team
	err = h.teamService.DeleteTeam(teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete team: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team deleted successfully", "team_id": teamID})
}

// ListTeamMembers godoc
// @Summary      List team members
// @Description  Get all members of a specific team
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Team ID"  example(1)
// @Success      200  {object}  map[string]interface{} "team_id and members array"
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/teams/{id}/members [get]
func (h *TeamHandler) ListTeamMembers(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid team ID",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Get team members from the service
	members, err := h.teamService.GetTeamMembers(teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get team members: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"team_id": teamID, "members": members})
}

// AddTeamMember godoc
// @Summary      Add team member
// @Description  Add a member to a team (creates user account if doesn't exist)
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Team ID"  example(1)
// @Param        request body handlers.AddMemberRequest true "Member data"
// @Success      201  {object}  map[string]interface{} "User and role info"
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/teams/{id}/members [post]
func (h *TeamHandler) AddTeamMember(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid team ID",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request payload",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Get user from context to check permissions
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeInternal,
		})
		return
	}
	currentUser := user.(*models.User)

	// Check if user has permission to add a member to this team
	// Only allow if user is the owner of the team
	team, err := h.teamService.GetTeamByID(teamID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Team not found",
			Code:  models.ErrCodeNotFound,
		})
		return
	}

	if team.OwnerID != currentUser.ID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "You don't have permission to add members to this team",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	// First, find the user by email
	existingUser, err := h.userService.GetUserByEmail(req.Email)
	if err != nil {
		// If user doesn't exist, create an invitation token
		// Generate a temporary password for the invited user
		tempPassword := generateRandomPassword(12) // 12-character random password

		// Create a new user account with the temp password
		// For invitations, we'll create the user as a "member" by default
		newUser, err := h.userService.CreateUser(req.Email, tempPassword, req.Email, "member", currentUser.TenantID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user account: " + err.Error()})
			return
		}

		// Add the new user to the team
		err = h.teamService.AddTeamMember(teamID, newUser.ID, req.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add team member: " + err.Error()})
			return
		}

		// Generate an invitation refresh token that can be used to set a permanent password
		invToken, _, err := h.generateToken(newUser.ID, req.Email, req.Role, currentUser.TenantID, 7*24*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate invitation token"})
			return
		}

		// In a real application, we would send an email invitation with the token
		// For now, return the invitation token
		c.JSON(http.StatusCreated, gin.H{
			"user":             newUser,
			"team_id":          teamID,
			"role":             req.Role,
			"message":          "Invitation sent. User account created and added to team.",
			"invitation_token": invToken, // This would typically be sent via email in a real application
		})
		return
	}

	// If user already exists, add them to the team
	err = h.teamService.AddTeamMember(teamID, existingUser.ID, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add team member: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":    existingUser,
		"team_id": teamID,
		"role":    req.Role,
		"message": "Member added to team successfully",
	})
}

// Helper function to generate a random password
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%&*"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}

// Helper function to generate JWT tokens
func (h *TeamHandler) generateToken(userID int64, email, role string, tenantID int64, duration time.Duration) (string, time.Time, error) {
	expirationTime := time.Now().Add(duration)
	claims := &jwt.MapClaims{
		"user_id":   userID,
		"email":     email,
		"role":      role,
		"tenant_id": tenantID,
		"exp":       expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("default_secret_key_for_development"))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}

// RemoveTeamMember godoc
// @Summary      Remove team member
// @Description  Remove a member from a team (only team owner can remove)
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Team ID"  example(1)
// @Param        memberId path int true "Member User ID" example(2)
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/teams/{id}/members/{memberId} [delete]
func (h *TeamHandler) RemoveTeamMember(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid team ID",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	memberID, err := strconv.ParseInt(c.Param("memberId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid member ID",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Get user from context to check permissions
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeInternal,
		})
		return
	}
	currentUser := user.(*models.User)

	// Check if user has permission to remove a member from this team
	// Only allow if user is the owner of the team
	team, err := h.teamService.GetTeamByID(teamID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Team not found",
			Code:  models.ErrCodeNotFound,
		})
		return
	}

	if team.OwnerID != currentUser.ID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "You don't have permission to remove members from this team",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	// Remove the member from the team
	err = h.teamService.RemoveTeamMember(teamID, memberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove team member: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Member removed from team successfully",
		"team_id":   teamID,
		"member_id": memberID,
	})
}

// GetTeamSettings godoc
// @Summary      Get team settings
// @Description  Get settings for a specific team
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Team ID"  example(1)
// @Success      200  {object}  map[string]interface{} "team_id and settings"
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/teams/{id}/settings [get]
func (h *TeamHandler) GetTeamSettings(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid team ID",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// In a real application, get team settings from the database
	// For now, return empty settings
	settings := make(map[string]string)
	c.JSON(http.StatusOK, gin.H{"team_id": teamID, "settings": settings})
}

// UpdateTeamSettings godoc
// @Summary      Update team settings
// @Description  Update settings for a specific team (only team owner can update)
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Team ID"  example(1)
// @Param        settings body map[string]string true "Settings object"
// @Success      200  {object}  map[string]interface{} "team_id and settings"
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/teams/{id}/settings [put]
func (h *TeamHandler) UpdateTeamSettings(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid team ID",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	var settings map[string]string
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid settings format",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Get user from context to check permissions
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeInternal,
		})
		return
	}
	currentUser := user.(*models.User)

	// Check if user has permission to update settings for this team
	// Only allow if user is the owner of the team
	team, err := h.teamService.GetTeamByID(teamID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Team not found",
			Code:  models.ErrCodeNotFound,
		})
		return
	}

	if team.OwnerID != currentUser.ID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "You don't have permission to update settings for this team",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	// Update the team settings
	err = h.teamService.UpdateTeam(teamID, "", "", settings) // Only update settings, not name or description
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team settings: " + err.Error()})
		return
	}

	// Get the updated team to return settings
	updatedTeam, err := h.teamService.GetTeamByID(teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated team: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"team_id": teamID, "settings": updatedTeam.Settings})
}
