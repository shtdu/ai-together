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
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"switch-server/models"
	"switch-server/repository"
)

// ErrIncorrectPassword is returned when the current password doesn't match.
var ErrIncorrectPassword = errors.New("current password is incorrect")

type UserService struct {
	userRepo repository.UserRepositoryInterface
}

// NewUserService creates and returns a new instance of UserService
func NewUserService(userRepo repository.UserRepositoryInterface) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user with the specified email, password, name, role, and tenant
// The password is automatically hashed before storing in the database
// It returns the created user object or an error if the operation fails
func (s *UserService) CreateUser(email, password, name, role string, tenantID int64) (*models.User, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Insert the user into the database
	user, err := s.userRepo.CreateUser(context.Background(), email, string(hashedPassword), name, role, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by their email address
// It returns the user object or an error if the user doesn't exist
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	user, err := s.userRepo.GetUserByEmail(context.Background(), email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by their unique identifier
// It returns the user object or an error if the user doesn't exist
func (s *UserService) GetUserByID(userID int64) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}

	return user, nil
}

// ValidatePassword checks if the provided password matches the hashed password stored for the user
// It returns true if the passwords match, false otherwise
func (s *UserService) ValidatePassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

// HashPassword takes a plaintext password and returns the bcrypt hash
// This can be used to hash passwords externally if needed
func (s *UserService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// ListUsersByTenant returns all users belonging to a specific tenant
func (s *UserService) ListUsersByTenant(tenantID int64) ([]models.User, error) {
	users, err := s.userRepo.ListUsersByTenant(context.Background(), tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list users by tenant: %w", err)
	}

	// Clear passwords before returning
	for i := range users {
		users[i].Password = ""
	}

	return users, nil
}

// UpdateUser updates a user's information
func (s *UserService) UpdateUser(userID int64, name, role, password string) (*models.User, error) {
	// If password is provided, update with password
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		err = s.userRepo.UpdateUserWithPassword(context.Background(), userID, name, role, string(hashedPassword))
		if err != nil {
			return nil, fmt.Errorf("failed to update user with password: %w", err)
		}
	} else {
		err := s.userRepo.UpdateUser(context.Background(), userID, name, role)
		if err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Return updated user
	user, err := s.userRepo.GetUserByID(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated user: %w", err)
	}

	user.Password = ""
	return user, nil
}

// DeleteUser removes a user from the database
func (s *UserService) DeleteUser(userID int64) error {
	err := s.userRepo.DeleteUser(context.Background(), userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// UpdateProfileName updates the authenticated user's display name
func (s *UserService) UpdateProfileName(userID int64, name string) (*models.User, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	err := s.userRepo.UpdateUserName(context.Background(), userID, name)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile name: %w", err)
	}
	user, err := s.userRepo.GetUserByID(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated user: %w", err)
	}
	user.Password = ""
	return user, nil
}

// ChangePassword validates the current password and sets a new one
func (s *UserService) ChangePassword(userID int64, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetUserByID(context.Background(), userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if !s.ValidatePassword(user, currentPassword) {
		return ErrIncorrectPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	err = s.userRepo.UpdateUserPassword(context.Background(), userID, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}
