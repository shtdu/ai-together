// Copyright (c) 2025 Code Together
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
	"fmt"
	"log/slog"
)

// PermissionService handles permission checks for the member client
type PermissionService struct {
	authService *AuthService
}

// NewPermissionService creates a new permission service
func NewPermissionService(authService *AuthService) *PermissionService {
	return &PermissionService{
		authService: authService,
	}
}

// Start initializes the permission service
func (p *PermissionService) Start() error {
	slog.Info("PermissionService started")
	return nil
}

// Stop stops the permission service
func (p *PermissionService) Stop() error {
	slog.Info("PermissionService stopped")
	return nil
}

// IsAdmin checks if the current user has admin (manager) role
func (p *PermissionService) IsAdmin() (bool, error) {
	user, err := p.authService.GetCurrentUser()
	if err != nil {
		return false, fmt.Errorf("failed to get current user: %w", err)
	}

	// Manager role has admin permissions
	return user.Role == "manager", nil
}

// GetUserRole returns the current user's role
func (p *PermissionService) GetUserRole() (string, error) {
	user, err := p.authService.GetCurrentUser()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	return user.Role, nil
}

// CanEditSettings checks if the user can edit provider settings
// Only managers/admins can edit settings
func (p *PermissionService) CanEditSettings() (bool, error) {
	return p.IsAdmin()
}

// CanPushMasterCopy checks if the user can push local settings to server as master copy
// Only managers/admins can push master copy
func (p *PermissionService) CanPushMasterCopy() (bool, error) {
	return p.IsAdmin()
}

// EnsureAdmin returns an error if the user is not an admin
// This is a convenience method for use in other services
func (p *PermissionService) EnsureAdmin() error {
	isAdmin, err := p.IsAdmin()
	if err != nil {
		return err
	}

	if !isAdmin {
		return fmt.Errorf("permission denied: this action requires manager role")
	}

	return nil
}
