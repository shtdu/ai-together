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

package services

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"switch-server/models"
)

func TestUserService_CreateUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	testUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     "member",
		TenantID: 1,
		Password: "$2a$10$dummy.hashed.password", // Will be set by mock
	}

	mockRepo.On("CreateUser", mock.Anything, "test@example.com", mock.AnythingOfType("string"), "Test User", "member", int64(1)).
		Return(testUser, nil)

	user, err := service.CreateUser("test@example.com", "password123", "Test User", "member", 1)

	require.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, "member", user.Role)
	assert.NotEmpty(t, user.Password) // Password should be hashed

	mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("CreateUser", mock.Anything, "test@example.com", mock.AnythingOfType("string"), "Test User", "member", int64(1)).
		Return(nil, assert.AnError)

	user, err := service.CreateUser("test@example.com", "password123", "Test User", "member", 1)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to create user")

	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserByEmail_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	testUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     "member",
		TenantID: 1,
	}

	mockRepo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(testUser, nil)

	user, err := service.GetUserByEmail("test@example.com")

	require.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)

	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserByEmail_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("GetUserByEmail", mock.Anything, "nonexistent@example.com").Return(nil, assert.AnError)

	user, err := service.GetUserByEmail("nonexistent@example.com")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to find user by email")

	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserByID_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	testUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     "member",
		TenantID: 1,
	}

	mockRepo.On("GetUserByID", mock.Anything, int64(1)).Return(testUser, nil)

	user, err := service.GetUserByID(1)

	require.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)

	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserByID_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("GetUserByID", mock.Anything, int64(999)).Return(nil, assert.AnError)

	user, err := service.GetUserByID(999)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to find user by ID")

	mockRepo.AssertExpectations(t)
}

func TestUserService_ValidatePassword_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	// Generate a hash for the password
	hashedPassword, err := service.HashPassword("password123")
	require.NoError(t, err)

	testUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     "member",
		Password: hashedPassword,
	}

	// Test with correct password
	isValid := service.ValidatePassword(testUser, "password123")
	assert.True(t, isValid)
}

func TestUserService_ValidatePassword_Failure(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	// Generate a hash for the password
	hashedPassword, err := service.HashPassword("password123")
	require.NoError(t, err)

	testUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     "member",
		Password: hashedPassword,
	}

	// Test with incorrect password
	isValid := service.ValidatePassword(testUser, "wrongpassword")
	assert.False(t, isValid)
}

func TestUserService_HashPassword_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	password := "testpassword123"
	hashedPassword, err := service.HashPassword(password)

	require.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)
	assert.Contains(t, hashedPassword, "$2a$") // bcrypt hash prefix
}

func TestUserService_ListUsersByTenant_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	testUsers := []models.User{
		{ID: 1, Email: "user1@example.com", Name: "User 1", Role: "member", TenantID: 1, Password: "hashed1"},
		{ID: 2, Email: "user2@example.com", Name: "User 2", Role: "manager", TenantID: 1, Password: "hashed2"},
	}

	mockRepo.On("ListUsersByTenant", mock.Anything, int64(1)).Return(testUsers, nil)

	users, err := service.ListUsersByTenant(1)

	require.NoError(t, err)
	assert.Len(t, users, 2)
	// Verify passwords are cleared
	assert.Empty(t, users[0].Password)
	assert.Empty(t, users[1].Password)

	mockRepo.AssertExpectations(t)
}

func TestUserService_ListUsersByTenant_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("ListUsersByTenant", mock.Anything, int64(1)).Return([]models.User{}, assert.AnError)

	users, err := service.ListUsersByTenant(1)

	assert.Error(t, err)
	assert.Nil(t, users)
	assert.Contains(t, err.Error(), "failed to list users by tenant")

	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_WithPassword_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	updatedUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "Updated Name",
		Role:     "manager",
		TenantID: 1,
		Password: "hashedpassword",
	}

	mockRepo.On("UpdateUserWithPassword", mock.Anything, int64(1), "Updated Name", "manager", mock.AnythingOfType("string")).Return(nil)
	mockRepo.On("GetUserByID", mock.Anything, int64(1)).Return(updatedUser, nil)

	user, err := service.UpdateUser(1, "Updated Name", "manager", "newpassword123")

	require.NoError(t, err)
	assert.Equal(t, "Updated Name", user.Name)
	assert.Equal(t, "manager", user.Role)
	assert.Empty(t, user.Password) // Password should be cleared

	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_WithoutPassword_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	updatedUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "Updated Name",
		Role:     "manager",
		TenantID: 1,
		Password: "hashedpassword",
	}

	mockRepo.On("UpdateUser", mock.Anything, int64(1), "Updated Name", "manager").Return(nil)
	mockRepo.On("GetUserByID", mock.Anything, int64(1)).Return(updatedUser, nil)

	user, err := service.UpdateUser(1, "Updated Name", "manager", "")

	require.NoError(t, err)
	assert.Equal(t, "Updated Name", user.Name)
	assert.Empty(t, user.Password)

	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("UpdateUser", mock.Anything, int64(1), "Updated Name", "manager").Return(assert.AnError)

	user, err := service.UpdateUser(1, "Updated Name", "manager", "")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to update user")

	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_UserNotFoundError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("UpdateUser", mock.Anything, int64(1), "Updated Name", "manager").Return(nil)
	mockRepo.On("GetUserByID", mock.Anything, int64(1)).Return(nil, assert.AnError)

	user, err := service.UpdateUser(1, "Updated Name", "manager", "")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to retrieve updated user")

	mockRepo.AssertExpectations(t)
}

func TestUserService_DeleteUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("DeleteUser", mock.Anything, int64(1)).Return(nil)

	err := service.DeleteUser(1)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestUserService_DeleteUser_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("DeleteUser", mock.Anything, int64(1)).Return(assert.AnError)

	err := service.DeleteUser(1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete user")

	mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_PasswordHashing(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	testUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     "member",
		TenantID: 1,
	}

	var capturedPassword string
	mockRepo.On("CreateUser", mock.Anything, "test@example.com", mock.AnythingOfType("string"), "Test User", "member", int64(1)).
		Run(func(args mock.Arguments) {
			capturedPassword = args.String(2) // Get the hashedPassword argument
		}).Return(testUser, nil)

	_, err := service.CreateUser("test@example.com", "password123", "Test User", "member", 1)

	require.NoError(t, err)
	// Verify password was hashed (bcrypt hashes start with $2a$)
	assert.Contains(t, capturedPassword, "$2a$")
	assert.NotEqual(t, "password123", capturedPassword)

	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_PasswordHashing(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	updatedUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "Updated Name",
		Role:     "manager",
		TenantID: 1,
	}

	var capturedPassword string
	mockRepo.On("UpdateUserWithPassword", mock.Anything, int64(1), "Updated Name", "manager", mock.AnythingOfType("string")).
		Run(func(args mock.Arguments) {
			capturedPassword = args.String(4) // Get the hashedPassword argument (index 4: ctx, userID, name, role, hashedPassword)
		}).Return(nil)
	mockRepo.On("GetUserByID", mock.Anything, int64(1)).Return(updatedUser, nil)

	_, err := service.UpdateUser(1, "Updated Name", "manager", "newpassword123")

	require.NoError(t, err)
	// Verify password was hashed
	assert.Contains(t, capturedPassword, "$2a$")
	assert.NotEqual(t, "newpassword123", capturedPassword)

	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateProfileName_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	updatedUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "New Name",
		Role:     "member",
		TenantID: 1,
		Password: "hashed",
	}

	mockRepo.On("UpdateUserName", mock.Anything, int64(1), "New Name").Return(nil)
	mockRepo.On("GetUserByID", mock.Anything, int64(1)).Return(updatedUser, nil)

	user, err := service.UpdateProfileName(1, "New Name")

	require.NoError(t, err)
	assert.Equal(t, "New Name", user.Name)
	assert.Empty(t, user.Password) // Password should be cleared in response

	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateProfileName_EmptyName(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	_, err := service.UpdateProfileName(1, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

func TestUserService_UpdateProfileName_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("UpdateUserName", mock.Anything, int64(1), "New Name").Return(assert.AnError)

	_, err := service.UpdateProfileName(1, "New Name")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update profile name")

	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateProfileName_GetUserError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("UpdateUserName", mock.Anything, int64(1), "New Name").Return(nil)
	mockRepo.On("GetUserByID", mock.Anything, int64(1)).Return(nil, assert.AnError)

	_, err := service.UpdateProfileName(1, "New Name")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to retrieve updated user")

	mockRepo.AssertExpectations(t)
}

func TestUserService_ChangePassword_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	// Generate a valid bcrypt hash for the current password
	hashedPassword, err := service.HashPassword("currentPassword123")
	require.NoError(t, err)

	testUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     "member",
		Password: hashedPassword,
	}

	var capturedHash string
	mockRepo.On("GetUserByID", mock.Anything, int64(1)).Return(testUser, nil)
	mockRepo.On("UpdateUserPassword", mock.Anything, int64(1), mock.AnythingOfType("string")).
		Run(func(args mock.Arguments) {
			capturedHash = args.String(2)
		}).Return(nil)

	err = service.ChangePassword(1, "currentPassword123", "newPassword456")

	require.NoError(t, err)
	// Verify the new password was hashed
	assert.Contains(t, capturedHash, "$2a$")
	assert.NotEqual(t, "newPassword456", capturedHash)

	mockRepo.AssertExpectations(t)
}

func TestUserService_ChangePassword_WrongCurrentPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	// Hash for "correctPassword"
	hashedPassword, err := service.HashPassword("correctPassword")
	require.NoError(t, err)

	testUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     "member",
		Password: hashedPassword,
	}

	mockRepo.On("GetUserByID", mock.Anything, int64(1)).Return(testUser, nil)

	err = service.ChangePassword(1, "wrongPassword", "newPassword456")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrIncorrectPassword), "expected ErrIncorrectPassword")

	mockRepo.AssertExpectations(t)
}

func TestUserService_ChangePassword_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("GetUserByID", mock.Anything, int64(1)).Return(nil, assert.AnError)

	err := service.ChangePassword(1, "currentPassword", "newPassword456")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find user")

	mockRepo.AssertExpectations(t)
}

func TestUserService_ChangePassword_UpdateError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	hashedPassword, err := service.HashPassword("currentPassword")
	require.NoError(t, err)

	testUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     "member",
		Password: hashedPassword,
	}

	mockRepo.On("GetUserByID", mock.Anything, int64(1)).Return(testUser, nil)
	mockRepo.On("UpdateUserPassword", mock.Anything, int64(1), mock.AnythingOfType("string")).Return(assert.AnError)

	err = service.ChangePassword(1, "currentPassword", "newPassword456")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update password")

	mockRepo.AssertExpectations(t)
}
