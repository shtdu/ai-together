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
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/code-together/shared/integration"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const (
	authConfigDir  = ".code-together"
	authConfigFile = "auth.json"
)

// User represents a user from the server
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	TenantID  int64     `json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuthTokens stores JWT tokens
type AuthTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// AuthConfig stores authentication data
type AuthConfig struct {
	Tokens AuthTokens `json:"tokens"`
	User   User       `json:"user"`
}

// AuthResponse represents a successful auth response from server
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	User         User      `json:"user"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// AuthService handles authentication with the server
type AuthService struct {
	configPath           string
	mu                   sync.Mutex
	apiClient            integration.ClientWithResponsesInterface // Anonymous client for login/register/refresh
	serverConfigService  *ServerConfigService
}

// wrapAPIError converts an HTTP response into an error, parsing the response body
// for structured error information from the server
func (as *AuthService) wrapAPIError(httpResp *http.Response, action string) error {
	if httpResp == nil {
		return fmt.Errorf("nil HTTP response from server")
	}

	// Successful response
	if httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
		return nil
	}

	// Parse the error response
	apiErr := integration.UnwrapJSONResponse(httpResp)
	if apiErr != nil {
		// Add context about what action was being performed
		return fmt.Errorf("%s: %w", action, apiErr)
	}

	// Fallback to status code if parsing failed
	return fmt.Errorf("%s: server returned status %d", action, httpResp.StatusCode)
}

// NewAuthService creates a new auth service
func NewAuthService(apiClient integration.ClientWithResponsesInterface, serverConfigService *ServerConfigService) *AuthService {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	configPath := filepath.Join(home, authConfigDir, authConfigFile)

	return &AuthService{
		configPath:           configPath,
		apiClient:            apiClient,
		serverConfigService:  serverConfigService,
	}
}

// Login authenticates a user with email and password
func (a *AuthService) Login(email, password string) (*AuthResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := a.apiClient.PostAuthLoginWithResponse(ctx, integration.PostAuthLoginJSONRequestBody{
		Email:    openapi_types.Email(email),
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}

	if err := a.wrapAPIError(resp.HTTPResponse, "login"); err != nil {
		return nil, err
	}

	authResp := resp.JSON200

	// Convert API response to local format
	result := &AuthResponse{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		ExpiresAt:    authResp.ExpiresAt,
		User: User{
			ID:        authResp.User.Id,
			Email:     string(authResp.User.Email),
			Name:      authResp.User.Name,
			Role:      string(authResp.User.Role),
			TenantID:  authResp.User.TenantId,
			CreatedAt: authResp.User.CreatedAt,
			UpdatedAt: time.Time{}, // Will be set below if not nil
		},
	}
	if authResp.User.UpdatedAt != nil {
		result.User.UpdatedAt = *authResp.User.UpdatedAt
	}

	// Save tokens and user info
	if err := a.saveAuthConfig(result); err != nil {
		return nil, fmt.Errorf("failed to save auth config: %w", err)
	}

	return result, nil
}

// Register creates a new user account
func (a *AuthService) Register(email, password, name string) (*AuthResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := a.apiClient.PostAuthRegisterWithResponse(ctx, integration.PostAuthRegisterJSONRequestBody{
		Email:    openapi_types.Email(email),
		Password: password,
		Name:     name,
	})
	if err != nil {
		return nil, fmt.Errorf("register request failed: %w", err)
	}

	if err := a.wrapAPIError(resp.HTTPResponse, "register"); err != nil {
		return nil, err
	}

	authResp := resp.JSON201

	// Convert API response to local format
	result := &AuthResponse{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		ExpiresAt:    authResp.ExpiresAt,
		User: User{
			ID:        authResp.User.Id,
			Email:     string(authResp.User.Email),
			Name:      authResp.User.Name,
			Role:      string(authResp.User.Role),
			TenantID:  authResp.User.TenantId,
			CreatedAt: authResp.User.CreatedAt,
			UpdatedAt: time.Time{}, // Will be set below if not nil
		},
	}
	if authResp.User.UpdatedAt != nil {
		result.User.UpdatedAt = *authResp.User.UpdatedAt
	}

	// Save tokens and user info
	if err := a.saveAuthConfig(result); err != nil {
		return nil, fmt.Errorf("failed to save auth config: %w", err)
	}

	return result, nil
}

// GetCurrentUser returns the current authenticated user
func (a *AuthService) GetCurrentUser() (*User, error) {
	config, err := a.loadAuthConfig()
	if err != nil {
		return nil, err
	}

	return &config.User, nil
}

// IsAuthenticated checks if the user is authenticated
// Attempts to refresh token if expired before returning false
func (a *AuthService) IsAuthenticated() bool {
	config, err := a.loadAuthConfig()
	if err != nil {
		return false
	}

	// Check if access token exists and is not expired
	if config.Tokens.AccessToken == "" {
		return false
	}

	// Check if token is expired
	if time.Now().After(config.Tokens.ExpiresAt) {
		// Attempt refresh before returning false
		if err := a.refreshToken(); err != nil {
			slog.Error("token refresh failed in IsAuthenticated", "error", err)
			return false
		}
		// Reload config after successful refresh
		config, err = a.loadAuthConfig()
		if err != nil {
			return false
		}
	}

	return true
}

// GetTokenExpirationTime returns when the current access token will expire
func (a *AuthService) GetTokenExpirationTime() (time.Time, error) {
	config, err := a.loadAuthConfig()
	if err != nil {
		return time.Time{}, err
	}
	return config.Tokens.ExpiresAt, nil
}

// RefreshIfNeeded checks if token needs refresh and refreshes if necessary
// Returns true if refresh was attempted, regardless of success
func (a *AuthService) RefreshIfNeeded() bool {
	config, err := a.loadAuthConfig()
	if err != nil {
		return false
	}

	// Refresh if token will expire in the next hour
	if time.Now().Add(1*time.Hour).After(config.Tokens.ExpiresAt) {
		slog.Info("token expiring soon, refreshing now", "expires_at", config.Tokens.ExpiresAt)
		if err := a.refreshToken(); err != nil {
			slog.Error("token refresh failed", "error", err)
			return false
		}

		// Get new expiration time for logging
		newConfig, err := a.loadAuthConfig()
		if err == nil {
			slog.Info("token refreshed successfully", "new_expires_at", newConfig.Tokens.ExpiresAt)
		}
		return true
	}

	return false
}

// Logout clears the authentication tokens
func (a *AuthService) Logout() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Simply delete the auth file
	if err := os.Remove(a.configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove auth config: %w", err)
	}

	return nil
}

// GetAccessToken returns the current access token
func (a *AuthService) GetAccessToken() (string, error) {
	config, err := a.loadAuthConfig()
	if err != nil {
		return "", err
	}

	// Check if token is expired
	if time.Now().After(config.Tokens.ExpiresAt) {
		// Try to refresh
		if err := a.refreshToken(); err != nil {
			return "", fmt.Errorf("token expired and refresh failed: %w", err)
		}

		// Reload config after refresh
		config, err = a.loadAuthConfig()
		if err != nil {
			return "", err
		}
	}

	return config.Tokens.AccessToken, nil
}

// refreshToken attempts to refresh the access token
func (a *AuthService) refreshToken() error {
	config, err := a.loadAuthConfig()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := a.apiClient.PostAuthRefreshWithResponse(ctx, integration.PostAuthRefreshJSONRequestBody{
		RefreshToken: config.Tokens.RefreshToken,
	})
	if err != nil {
		return fmt.Errorf("refresh request failed: %w", err)
	}

	// Handle 401 errors specifically - refresh token is expired
	if resp.StatusCode() == http.StatusUnauthorized {
		// Clear auth config since refresh token is no longer valid
		a.mu.Lock()
		os.Remove(a.configPath)
		a.mu.Unlock()
		slog.Warn("refresh token expired, clearing auth config")
		return fmt.Errorf("refresh token expired, please login again")
	}

	if err := a.wrapAPIError(resp.HTTPResponse, "refresh token"); err != nil {
		return err
	}

	authResp := resp.JSON200

	// Convert API response to local format
	result := &AuthResponse{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		ExpiresAt:    authResp.ExpiresAt,
		User: User{
			ID:        authResp.User.Id,
			Email:     string(authResp.User.Email),
			Name:      authResp.User.Name,
			Role:      string(authResp.User.Role),
			TenantID:  authResp.User.TenantId,
			CreatedAt: authResp.User.CreatedAt,
			UpdatedAt: time.Time{}, // Will be set below if not nil
		},
	}
	if authResp.User.UpdatedAt != nil {
		result.User.UpdatedAt = *authResp.User.UpdatedAt
	}

	// Save new tokens
	return a.saveAuthConfig(result)
}

// loadAuthConfig loads the auth configuration from disk
func (a *AuthService) loadAuthConfig() (*AuthConfig, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	config := &AuthConfig{}

	data, err := os.ReadFile(a.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not authenticated")
		}
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("not authenticated")
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// saveAuthConfig saves the auth configuration to disk
func (a *AuthService) saveAuthConfig(authResp *AuthResponse) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	config := AuthConfig{
		Tokens: AuthTokens{
			AccessToken:  authResp.AccessToken,
			RefreshToken: authResp.RefreshToken,
			ExpiresAt:    authResp.ExpiresAt,
		},
		User: authResp.User,
	}

	dir := filepath.Dir(a.configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(a.configPath, data, 0o600)
}

// GetUserProfile fetches the current user profile from the server
// This allows refreshing the user info (including role) without re-authentication
// Note: This method requires an authenticated client to be set via SetAuthenticatedClient
func (a *AuthService) GetUserProfile() (*User, error) {
	serverURL, err := a.getServerURL()
	if err != nil {
		return nil, err
	}

	token, err := a.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Create an authenticated client for this request
	authClient, err := integration.NewAuthenticatedClient(serverURL, func() (string, error) {
		return token, nil
	}, slog.Default(), false) // Minimal logging for background operations
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticated client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := authClient.GetApiV1UserProfileWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user profile: %w", err)
	}

	if err := a.wrapAPIError(resp.HTTPResponse, "fetch user profile"); err != nil {
		return nil, err
	}

	userResp := resp.JSON200

	// Convert API response to local format
	user := &User{
		ID:        userResp.User.Id,
		Email:     string(userResp.User.Email),
		Name:      userResp.User.Name,
		Role:      string(userResp.User.Role),
		TenantID:  userResp.User.TenantId,
		CreatedAt: userResp.User.CreatedAt,
		UpdatedAt: time.Time{}, // Will be set below if not nil
	}
	if userResp.User.UpdatedAt != nil {
		user.UpdatedAt = *userResp.User.UpdatedAt
	}

	// Update the cached user info
	if err := a.updateUserInConfig(user); err != nil {
		slog.Warn("failed to update cached user profile", "error", err)
	}

	return user, nil
}

// updateUserInConfig updates only the user portion of the auth config
func (a *AuthService) updateUserInConfig(user *User) error {
	config, err := a.loadAuthConfig()
	if err != nil {
		return err
	}

	config.User = *user

	dir := filepath.Dir(a.configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(a.configPath, data, 0o600)
}

// getServerURL gets the current server URL from ServerConfigService
func (a *AuthService) getServerURL() (string, error) {
	if a.serverConfigService == nil {
		return "", fmt.Errorf("server config service not initialized")
	}

	url, err := a.serverConfigService.GetServerURL()
	if err != nil {
		return "", fmt.Errorf("failed to get server URL: %w", err)
	}

	if url == "" {
		return "", fmt.Errorf("server URL not configured")
	}

	return url, nil
}
