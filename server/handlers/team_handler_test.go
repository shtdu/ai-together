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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"switch-server/models"
)

func TestTeamHandler_ListTeams_Success(t *testing.T) {
	mockTeamService := new(MockTeamService)
	mockLicenseService := new(MockLicenseService)
	mockUserService := new(MockUserService)

	testTeams := []models.Team{
		{ID: 1, Name: "Team 1", OwnerID: 1, TenantID: 1},
		{ID: 2, Name: "Team 2", OwnerID: 1, TenantID: 1},
	}

	mockTeamService.On("GetTeamsByUserID", int64(1)).Return(testTeams, nil)

	handler := NewTeamHandler(mockTeamService, mockUserService, mockLicenseService)
	router := setupTeamRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	req, _ := http.NewRequest("GET", "/teams", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Team
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response, 2)

	mockTeamService.AssertExpectations(t)
}

func TestTeamHandler_CreateTeam_Success(t *testing.T) {
	mockTeamService := new(MockTeamService)
	mockLicenseService := new(MockLicenseService)
	mockUserService := new(MockUserService)

	testTeam := &models.Team{ID: 1, Name: "New Team", Description: "A new team", OwnerID: 1, TenantID: 1}
	mockTeamService.On("CreateTeam", "New Team", "A new team", int64(1), int64(1), map[string]string(nil)).
		Return(testTeam, nil)
	mockLicenseService.On("CanCreateTeam", mock.Anything, int64(1)).Return(true, nil)

	handler := NewTeamHandler(mockTeamService, mockUserService, mockLicenseService)
	router := setupTeamRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	reqBody := CreateTeamRequest{
		Name:        "New Team",
		Description: "A new team",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/teams", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Team
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "New Team", response.Name)

	mockTeamService.AssertExpectations(t)
}

func TestTeamHandler_CreateTeam_ValidationError(t *testing.T) {
	mockTeamService := new(MockTeamService)
	mockLicenseService := new(MockLicenseService)
	mockUserService := new(MockUserService)

	handler := NewTeamHandler(mockTeamService, mockUserService, mockLicenseService)
	router := setupTeamRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	reqBody := map[string]interface{}{
		"description": "Missing name",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/teams", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTeamHandler_UpdateTeam_Success(t *testing.T) {
	mockTeamService := new(MockTeamService)
	mockLicenseService := new(MockLicenseService)
	mockUserService := new(MockUserService)

	existingTeam := &models.Team{ID: 1, Name: "Old Name", OwnerID: 1, TenantID: 1}
	updatedTeam := &models.Team{ID: 1, Name: "Updated Name", OwnerID: 1, TenantID: 1}

	mockTeamService.On("GetTeamByID", int64(1)).Once().Return(existingTeam, nil)
	mockTeamService.On("UpdateTeam", int64(1), "Updated Name", "", map[string]string(nil)).Return(nil)
	mockTeamService.On("GetTeamByID", int64(1)).Once().Return(updatedTeam, nil)

	handler := NewTeamHandler(mockTeamService, mockUserService, mockLicenseService)
	router := setupTeamRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	reqBody := UpdateTeamRequest{
		Name: "Updated Name",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/teams/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Team
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", response.Name)

	mockTeamService.AssertExpectations(t)
}

func TestTeamHandler_UpdateTeam_PermissionDenied(t *testing.T) {
	mockTeamService := new(MockTeamService)
	mockLicenseService := new(MockLicenseService)
	mockUserService := new(MockUserService)

	existingTeam := &models.Team{ID: 1, Name: "Team", OwnerID: 999, TenantID: 1} // Different owner

	mockTeamService.On("GetTeamByID", int64(1)).Return(existingTeam, nil)

	handler := NewTeamHandler(mockTeamService, mockUserService, mockLicenseService)
	router := setupTeamRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	reqBody := UpdateTeamRequest{
		Name: "Updated Name",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/teams/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "You don't have permission to update this team", response["error"])

	mockTeamService.AssertExpectations(t)
}

func TestTeamHandler_DeleteTeam_Success(t *testing.T) {
	mockTeamService := new(MockTeamService)
	mockLicenseService := new(MockLicenseService)
	mockUserService := new(MockUserService)

	existingTeam := &models.Team{ID: 1, Name: "Team", OwnerID: 1, TenantID: 1}

	mockTeamService.On("GetTeamByID", int64(1)).Return(existingTeam, nil)
	mockTeamService.On("DeleteTeam", int64(1)).Return(nil)

	handler := NewTeamHandler(mockTeamService, mockUserService, mockLicenseService)
	router := setupTeamRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	req, _ := http.NewRequest("DELETE", "/teams/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Team deleted successfully", response["message"])

	mockTeamService.AssertExpectations(t)
}

func TestTeamHandler_ListTeamMembers_Success(t *testing.T) {
	mockTeamService := new(MockTeamService)
	mockLicenseService := new(MockLicenseService)
	mockUserService := new(MockUserService)

	members := []models.User{
		*createTestUser(2, "member1@example.com", "Member 1", "member", 1),
		*createTestUser(3, "member2@example.com", "Member 2", "member", 1),
	}

	mockTeamService.On("GetTeamMembers", int64(1)).Return(members, nil)

	handler := NewTeamHandler(mockTeamService, mockUserService, mockLicenseService)
	router := setupTeamRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	req, _ := http.NewRequest("GET", "/teams/1/members", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(1), response["team_id"])
	assert.Contains(t, response, "members")

	mockTeamService.AssertExpectations(t)
}

func TestTeamHandler_AddTeamMember_Success(t *testing.T) {
	mockTeamService := new(MockTeamService)
	mockLicenseService := new(MockLicenseService)
	mockUserService := new(MockUserService)

	team := &models.Team{ID: 1, Name: "Team", OwnerID: 1, TenantID: 1}
	existingUser := createTestUser(2, "member@example.com", "Member", "member", 1)

	mockTeamService.On("GetTeamByID", int64(1)).Return(team, nil)
	mockUserService.On("GetUserByEmail", "member@example.com").Return(existingUser, nil)
	mockTeamService.On("AddTeamMember", int64(1), int64(2), "member").Return(nil)

	handler := NewTeamHandler(mockTeamService, mockUserService, mockLicenseService)
	router := setupTeamRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	reqBody := AddMemberRequest{
		Email: "member@example.com",
		Role:  "member",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/teams/1/members", bytes.NewReader(body))
	req.Header.Set("Content-type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusCreated, w.Code)

	mockTeamService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestTeamHandler_AddTeamMember_CreateNewUser(t *testing.T) {
	mockTeamService := new(MockTeamService)
	mockLicenseService := new(MockLicenseService)
	mockUserService := new(MockUserService)

	team := &models.Team{ID: 1, Name: "Team", OwnerID: 1, TenantID: 1}
	newUser := createTestUser(2, "newmember@example.com", "newmember@example.com", "member", 1)

	mockTeamService.On("GetTeamByID", int64(1)).Return(team, nil)
	mockUserService.On("GetUserByEmail", "newmember@example.com").Return((*models.User)(nil), assert.AnError)
	mockUserService.On("CreateUser", "newmember@example.com", mock.AnythingOfType("string"), "newmember@example.com", "member", int64(1)).
		Return(newUser, nil)
	mockTeamService.On("AddTeamMember", int64(1), int64(2), "member").Return(nil)

	handler := NewTeamHandler(mockTeamService, mockUserService, mockLicenseService)
	router := setupTeamRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	reqBody := AddMemberRequest{
		Email: "newmember@example.com",
		Role:  "member",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/teams/1/members", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusCreated, w.Code)

	mockTeamService.AssertExpectations(t)
	// Note: We don't assert mockUserService expectations because CreateUser uses a random password
}

func TestTeamHandler_AddTeamMember_ValidationError(t *testing.T) {
	mockTeamService := new(MockTeamService)
	mockLicenseService := new(MockLicenseService)
	mockUserService := new(MockUserService)

	handler := NewTeamHandler(mockTeamService, mockUserService, mockLicenseService)
	router := setupTeamRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	reqBody := map[string]interface{}{
		"email": "not-an-email",
		"role":  "member",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/teams/1/members", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTeamHandler_AddTeamMember_PermissionDenied(t *testing.T) {
	mockTeamService := new(MockTeamService)
	mockLicenseService := new(MockLicenseService)
	mockUserService := new(MockUserService)

	team := &models.Team{ID: 1, Name: "Team", OwnerID: 999, TenantID: 1} // Different owner

	mockTeamService.On("GetTeamByID", int64(1)).Return(team, nil)

	handler := NewTeamHandler(mockTeamService, mockUserService, mockLicenseService)
	router := setupTeamRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	reqBody := AddMemberRequest{
		Email: "member@example.com",
		Role:  "member",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/teams/1/members", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusForbidden, w.Code)

	mockTeamService.AssertExpectations(t)
}

func TestTeamHandler_RemoveTeamMember_Success(t *testing.T) {
	mockTeamService := new(MockTeamService)
	mockLicenseService := new(MockLicenseService)
	mockUserService := new(MockUserService)

	team := &models.Team{ID: 1, Name: "Team", OwnerID: 1, TenantID: 1}

	mockTeamService.On("GetTeamByID", int64(1)).Return(team, nil)
	mockTeamService.On("RemoveTeamMember", int64(1), int64(2)).Return(nil)

	handler := NewTeamHandler(mockTeamService, mockUserService, mockLicenseService)
	router := setupTeamRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	req, _ := http.NewRequest("DELETE", "/teams/1/members/2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Member removed from team successfully", response["message"])

	mockTeamService.AssertExpectations(t)
}

func TestTeamHandler_RemoveTeamMember_PermissionDenied(t *testing.T) {
	mockTeamService := new(MockTeamService)
	mockLicenseService := new(MockLicenseService)
	mockUserService := new(MockUserService)

	team := &models.Team{ID: 1, Name: "Team", OwnerID: 999, TenantID: 1} // Different owner

	mockTeamService.On("GetTeamByID", int64(1)).Return(team, nil)

	handler := NewTeamHandler(mockTeamService, mockUserService, mockLicenseService)
	router := setupTeamRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	req, _ := http.NewRequest("DELETE", "/teams/1/members/2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusForbidden, w.Code)

	mockTeamService.AssertExpectations(t)
}

func TestTeamHandler_GetTeamSettings_Success(t *testing.T) {
	mockTeamService := new(MockTeamService)
	mockLicenseService := new(MockLicenseService)
	mockUserService := new(MockUserService)

	handler := NewTeamHandler(mockTeamService, mockUserService, mockLicenseService)
	router := setupTeamRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	req, _ := http.NewRequest("GET", "/teams/1/settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(1), response["team_id"])
	assert.Contains(t, response, "settings")
}

func TestTeamHandler_UpdateTeamSettings_Success(t *testing.T) {
	mockTeamService := new(MockTeamService)
	mockLicenseService := new(MockLicenseService)
	mockUserService := new(MockUserService)

	team := &models.Team{ID: 1, Name: "Team", OwnerID: 1, TenantID: 1, Settings: map[string]string{"key": "value"}}
	updatedTeam := &models.Team{ID: 1, Name: "Team", OwnerID: 1, TenantID: 1, Settings: map[string]string{"newKey": "newValue"}}

	mockTeamService.On("GetTeamByID", int64(1)).Return(team, nil)
	mockTeamService.On("UpdateTeam", int64(1), "", "", map[string]string{"newKey": "newValue"}).Return(nil)
	mockTeamService.On("GetTeamByID", int64(1)).Return(updatedTeam, nil)

	handler := NewTeamHandler(mockTeamService, mockUserService, mockLicenseService)
	router := setupTeamRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	settings := map[string]string{"newKey": "newValue"}
	body, _ := json.Marshal(settings)
	req, _ := http.NewRequest("PUT", "/teams/1/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusOK, w.Code)

	mockTeamService.AssertExpectations(t)
}

func TestTeamHandler_UpdateTeamSettings_PermissionDenied(t *testing.T) {
	mockTeamService := new(MockTeamService)
	mockLicenseService := new(MockLicenseService)
	mockUserService := new(MockUserService)

	team := &models.Team{ID: 1, Name: "Team", OwnerID: 999, TenantID: 1} // Different owner

	mockTeamService.On("GetTeamByID", int64(1)).Return(team, nil)

	handler := NewTeamHandler(mockTeamService, mockUserService, mockLicenseService)
	router := setupTeamRouter(handler)

	testUser := createTestUser(1, "user@example.com", "User", "manager", 1)

	settings := map[string]string{"key": "value"}
	body, _ := json.Marshal(settings)
	req, _ := http.NewRequest("PUT", "/teams/1/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, setUserContext(req, testUser))

	assert.Equal(t, http.StatusForbidden, w.Code)

	mockTeamService.AssertExpectations(t)
}

// Helper function to set up team router
func setupTeamRouter(handler *TeamHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		userEmail := c.GetHeader("X-Test-User-Email")
		if userEmail != "" {
			userID := c.GetHeader("X-Test-User-ID")
			userRole := c.GetHeader("X-Test-User-Role")
			tenantID := c.GetHeader("X-Test-User-Tenant-ID")

			var id, tenant int64
			if userID != "" {
				fmt.Sscanf(userID, "%d", &id)
			}
			if tenantID != "" {
				fmt.Sscanf(tenantID, "%d", &tenant)
			}

			user := &models.User{
				ID:       id,
				Email:    userEmail,
				Role:     userRole,
				TenantID: tenant,
			}
			c.Set("user", user)
		}
		c.Next()
	})

	router.GET("/teams", handler.ListTeams)
	router.POST("/teams", handler.CreateTeam)
	router.PUT("/teams/:id", handler.UpdateTeam)
	router.DELETE("/teams/:id", handler.DeleteTeam)
	router.GET("/teams/:id/members", handler.ListTeamMembers)
	router.POST("/teams/:id/members", handler.AddTeamMember)
	router.DELETE("/teams/:id/members/:memberId", handler.RemoveTeamMember)
	router.GET("/teams/:id/settings", handler.GetTeamSettings)
	router.PUT("/teams/:id/settings", handler.UpdateTeamSettings)

	return router
}
