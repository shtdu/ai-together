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

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"testing"
	"time"

	integrationclient "github.com/code-together/shared/integration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// RelayTokenTestSuite tests relay token functionality.
// This is a standalone suite that does NOT embed IntegrationTestSuite,
// so it doesn't inherit unrelated test methods (permissions, analytics, etc.).
// It connects to the test server independently with minimal setup.
type RelayTokenTestSuite struct {
	suite.Suite

	ServerURL  string
	TestDBURL  string
	AdminToken string
	Client     *integrationclient.ClientWithResponses
	ctx        *TestContext
	logger     *slog.Logger
}

// SetupSuite connects to the test server and authenticates as admin.
func (s *RelayTokenTestSuite) SetupSuite() {
	ctx, err := setupTestContext()
	s.Require().NoError(err, "Failed to setup test context")
	s.ctx = ctx
	s.ServerURL = ctx.ServerURL
	s.TestDBURL = ctx.TestDBURL
	s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	// Login as admin
	s.loginAdminUser()

	s.T().Logf("RelayTokenTestSuite: connected to %s", s.ServerURL)
}

// TearDownSuite cleans up the test context.
func (s *RelayTokenTestSuite) TearDownSuite() {
	if s.ctx != nil {
		cleanupTestContext(s.ctx)
	}
}

// loginAdminUser authenticates as the admin user and stores the token + client.
func (s *RelayTokenTestSuite) loginAdminUser() {
	ctx := context.Background()

	// Login with admin credentials
	loginBody := map[string]string{
		"email":    "admin@example.com",
		"password": "AdminPassword123!",
	}
	jsonBody, err := json.Marshal(loginBody)
	s.Require().NoError(err)

	req, err := http.NewRequestWithContext(ctx, "POST", s.ServerURL+"/auth/login", bytes.NewReader(jsonBody))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode, "Admin login failed")

	var loginResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	s.Require().NoError(err)

	s.AdminToken = loginResp["access_token"].(string)
	s.Require().NotEmpty(s.AdminToken, "Admin token is empty")

	// Create API client
	apiClient, err := integrationclient.NewAuthenticatedClient(
		s.ServerURL,
		func() (string, error) { return s.AdminToken, nil },
		s.logger,
		false,
	)
	s.Require().NoError(err)
	s.Client = apiClient
}

// --- Helpers ---

// adminRequest performs an authenticated admin API request using the admin token.
func (s *RelayTokenTestSuite) adminRequest(method, path string, body interface{}) (*http.Response, map[string]interface{}) {
	ctx := context.Background()

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, s.ServerURL+path, reqBody)
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.AdminToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	var jsonBody map[string]interface{}
	if resp.Body != nil {
		_ = json.NewDecoder(resp.Body).Decode(&jsonBody)
	}

	return resp, jsonBody
}

// relayTokenRequest performs a request using a relay token as Bearer auth.
func (s *RelayTokenTestSuite) relayTokenRequest(method, path, token string) (*http.Response, map[string]interface{}) {
	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, method, s.ServerURL+path, nil)
	require.NoError(s.T(), err)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	var jsonBody map[string]interface{}
	if resp.Body != nil {
		_ = json.NewDecoder(resp.Body).Decode(&jsonBody)
	}

	return resp, jsonBody
}

// --- Generate Relay Token Tests ---

// TestGenerateRelayTokenSuccess tests that a manager can generate a relay token for a user.
func (s *RelayTokenTestSuite) TestGenerateRelayTokenSuccess() {
	adminID := int64(1)

	resp, body := s.adminRequest("POST", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)

	require.Equal(s.T(), http.StatusCreated, resp.StatusCode, "Generate should return 201")
	assert.NotEmpty(s.T(), body["token"], "Response should contain raw token")
	assert.NotEmpty(s.T(), body["prefix"], "Response should contain token prefix")
	assert.NotEmpty(s.T(), body["created_at"], "Response should contain created_at")
	assert.Contains(s.T(), body["message"], "Save this token securely")

	s.T().Logf("Generated relay token prefix: %v", body["prefix"])

	// Clean up
	s.adminRequest("DELETE", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
}

// TestGenerateRelayTokenForMember tests generating a relay token for the current admin user.
func (s *RelayTokenTestSuite) TestGenerateRelayTokenForMember() {
	ctx := context.Background()

	// Get admin user ID from profile
	profileResp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, profileResp.StatusCode())
	adminID := profileResp.JSON200.User.Id

	// Generate relay token
	resp, body := s.adminRequest("POST", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)

	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	assert.NotEmpty(s.T(), body["token"])
	assert.NotEmpty(s.T(), body["prefix"])

	// Clean up
	s.adminRequest("DELETE", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
}

// TestGenerateRelayTokenDuplicate tests that generating a second token replaces the first.
func (s *RelayTokenTestSuite) TestGenerateRelayTokenDuplicate() {
	ctx := context.Background()

	profileResp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	adminID := profileResp.JSON200.User.Id

	// Generate first token
	resp1, body1 := s.adminRequest("POST", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusCreated, resp1.StatusCode)
	token1 := body1["token"].(string)
	prefix1 := body1["prefix"].(string)

	// Generate second token (should replace first)
	resp2, body2 := s.adminRequest("POST", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusCreated, resp2.StatusCode)
	token2 := body2["token"].(string)
	prefix2 := body2["prefix"].(string)

	// Tokens should be different
	assert.NotEqual(s.T(), token1, token2, "New token should differ from previous")
	assert.NotEqual(s.T(), prefix1, prefix2, "Prefixes should differ")

	// Clean up
	s.adminRequest("DELETE", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
}

// --- Get Relay Token Info Tests ---

// TestGetRelayTokenInfoSuccess tests retrieving masked token info.
func (s *RelayTokenTestSuite) TestGetRelayTokenInfoSuccess() {
	ctx := context.Background()

	profileResp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	adminID := profileResp.JSON200.User.Id

	// Generate a token first
	s.adminRequest("POST", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)

	// Get token info
	resp, body := s.adminRequest("GET", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)

	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	assert.NotEmpty(s.T(), body["prefix"], "Info should contain token prefix")
	assert.NotEmpty(s.T(), body["created_at"], "Info should contain created_at")
	// Raw token should NOT be in info response
	_, hasRawToken := body["token"]
	assert.False(s.T(), hasRawToken, "Info response should not contain raw token")

	// Clean up
	s.adminRequest("DELETE", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
}

// TestGetRelayTokenInfoNoToken tests getting info when no token exists for a valid user.
func (s *RelayTokenTestSuite) TestGetRelayTokenInfoNoToken() {
	ctx := context.Background()

	// Get admin user ID — should have no token if tests clean up properly
	profileResp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	adminID := profileResp.JSON200.User.Id

	// Make sure no token exists
	s.adminRequest("DELETE", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)

	// Get token info — should return 404
	resp, _ := s.adminRequest("GET", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)

	// 404 means no active token found
	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode, "Should return 404 when no token exists")
}

// --- Revoke Relay Token Tests ---

// TestRevokeRelayTokenSuccess tests revoking a relay token.
func (s *RelayTokenTestSuite) TestRevokeRelayTokenSuccess() {
	ctx := context.Background()

	profileResp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	adminID := profileResp.JSON200.User.Id

	// Generate token
	resp1, _ := s.adminRequest("POST", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusCreated, resp1.StatusCode)

	// Revoke token
	resp2, body2 := s.adminRequest("DELETE", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusOK, resp2.StatusCode)
	assert.Contains(s.T(), body2["message"], "revoked")

	// Verify token info is gone
	resp3, _ := s.adminRequest("GET", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	assert.Equal(s.T(), http.StatusNotFound, resp3.StatusCode, "Token info should be gone after revocation")
}

// TestRevokeRelayTokenNonExistent tests revoking a token that doesn't exist.
func (s *RelayTokenTestSuite) TestRevokeRelayTokenNonExistent() {
	// Use a non-existent user ID
	resp, _ := s.adminRequest("DELETE", "/api/v1/users/99999/relay-token", nil)
	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)
}

// --- Authorization Tests ---

// TestRelayTokenEndpointUnauthorized tests that unauthenticated requests are rejected.
func (s *RelayTokenTestSuite) TestRelayTokenEndpointUnauthorized() {
	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, "POST", s.ServerURL+"/api/v1/users/1/relay-token", nil)
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode, "Should require authentication")
}

// TestRelayTokenInvalidUserID tests that invalid user IDs return 400.
func (s *RelayTokenTestSuite) TestRelayTokenInvalidUserID() {
	resp, _ := s.adminRequest("POST", "/api/v1/users/invalid/relay-token", nil)
	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode, "Invalid user ID should return 400")
}

// --- Relay Token Authentication Flow Tests ---

// TestRelayTokenAuthFlow tests the full flow: generate token → use it for auth → revoke.
func (s *RelayTokenTestSuite) TestRelayTokenAuthFlow() {
	ctx := context.Background()

	profileResp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	adminID := profileResp.JSON200.User.Id

	// Step 1: Generate relay token
	resp, body := s.adminRequest("POST", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	rawToken := body["token"].(string)
	require.NotEmpty(s.T(), rawToken)
	s.T().Logf("Generated relay token for auth flow test")

	// Step 2: Use relay token on a relay endpoint (the middleware is only on relay routes)
	// The relay endpoint will proxy to a provider, so it will return an error about
	// missing provider — but we just care that auth succeeds (not 401).
	req, err := http.NewRequestWithContext(ctx, "POST", s.ServerURL+"/api/v1/relay/claude/v1/messages", nil)
	require.NoError(s.T(), err)
	req.Header.Set("Authorization", "Bearer "+rawToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	authResp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer authResp.Body.Close()

	// Should NOT be 401 (unauthorized) — relay token auth should work
	assert.NotEqual(s.T(), http.StatusUnauthorized, authResp.StatusCode, "Relay token should authenticate (not 401)")

	// Step 3: Revoke the token
	resp, _ = s.adminRequest("DELETE", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Step 4: Verify revoked token no longer works
	req2, err := http.NewRequestWithContext(ctx, "POST", s.ServerURL+"/api/v1/relay/claude/v1/messages", nil)
	require.NoError(s.T(), err)
	req2.Header.Set("Authorization", "Bearer "+rawToken)
	req2.Header.Set("Content-Type", "application/json")

	authResp2, err := client.Do(req2)
	require.NoError(s.T(), err)
	defer authResp2.Body.Close()

	// Revoked token falls through to session auth which rejects it → 401
	assert.Equal(s.T(), http.StatusUnauthorized, authResp2.StatusCode, "Revoked token should be rejected")
}

// TestRelayTokenJWTFallback tests that JWT tokens still work on relay endpoints.
func (s *RelayTokenTestSuite) TestRelayTokenJWTFallback() {
	ctx := context.Background()

	// Use admin JWT token on a relay endpoint.
	// The relay token middleware should detect dots in JWT and fall through to session auth.
	req, err := http.NewRequestWithContext(ctx, "POST", s.ServerURL+"/api/v1/relay/claude/v1/messages", nil)
	require.NoError(s.T(), err)
	req.Header.Set("Authorization", "Bearer "+s.AdminToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	// JWT should pass relay token middleware (detected by dots) and proceed to handler.
	// The handler will return a non-401 status (likely 400/500 since no actual provider is configured).
	assert.NotEqual(s.T(), http.StatusUnauthorized, resp.StatusCode, "JWT should not be rejected by relay token middleware")
}

// TestRelayTokenInvalidToken tests that an invalid token is rejected on relay endpoints.
func (s *RelayTokenTestSuite) TestRelayTokenInvalidToken() {
	// Use a completely invalid token on a relay endpoint
	resp, _ := s.relayTokenRequest("POST", "/api/v1/relay/claude/v1/messages", "invalid_token_not_valid_hex_64chars_aaaaaaaaaaaaaaaaaaaaaaaaaa")

	// Should fall through to session auth and fail (not a valid JWT either)
	assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode, "Invalid token should be rejected")
}

// TestRelayTokenNoAuthHeader tests that requests without Authorization header are rejected on relay endpoints.
func (s *RelayTokenTestSuite) TestRelayTokenNoAuthHeader() {
	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, "POST", s.ServerURL+"/api/v1/relay/claude/v1/messages", nil)
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header at all

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode, "Request without auth header should be rejected")
}

// --- Token Format & Validation Tests ---

// TestRelayTokenFormat tests that generated tokens have the expected format.
func (s *RelayTokenTestSuite) TestRelayTokenFormat() {
	ctx := context.Background()

	profileResp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	adminID := profileResp.JSON200.User.Id

	resp, body := s.adminRequest("POST", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	rawToken := body["token"].(string)
	prefix := body["prefix"].(string)

	// Token should be 64 hex characters (32 bytes = SHA256 input length for display)
	assert.Len(s.T(), rawToken, 64, "Raw token should be 64 hex characters")
	assert.Regexp(s.T(), `^[0-9a-f]{64}$`, rawToken, "Token should be lowercase hex only")

	// Prefix should be first 8 chars of token
	assert.Len(s.T(), prefix, 8, "Prefix should be 8 characters")
	assert.Equal(s.T(), rawToken[:8], prefix, "Prefix should match first 8 chars of token")

	// Token should NOT contain dots (not a JWT)
	assert.NotContains(s.T(), rawToken, ".", "Relay token should not contain dots (JWT separator)")

	// Clean up
	s.adminRequest("DELETE", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
}

// TestRelayTokenTooShort tests that a token shorter than the minimum length is rejected.
func (s *RelayTokenTestSuite) TestRelayTokenTooShort() {
	// Token shorter than 8 chars (TokenPrefixLength) should be rejected
	resp, _ := s.relayTokenRequest("POST", "/api/v1/relay/claude/v1/messages", "abc123")
	assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode, "Short token should be rejected")
}

// --- Lifecycle Tests ---

// TestRelayTokenRevokeThenRegenerate tests revoking a token then generating a new one.
func (s *RelayTokenTestSuite) TestRelayTokenRevokeThenRegenerate() {
	ctx := context.Background()

	profileResp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	adminID := profileResp.JSON200.User.Id

	// Generate first token
	resp1, body1 := s.adminRequest("POST", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusCreated, resp1.StatusCode)
	token1 := body1["token"].(string)

	// Revoke it
	resp2, _ := s.adminRequest("DELETE", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusOK, resp2.StatusCode)

	// Generate new token
	resp3, body3 := s.adminRequest("POST", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusCreated, resp3.StatusCode)
	token2 := body3["token"].(string)

	// New token should be different from the revoked one
	assert.NotEqual(s.T(), token1, token2, "New token after revoke should differ")

	// Old token should not work
	resp4, _ := s.relayTokenRequest("POST", "/api/v1/relay/claude/v1/messages", token1)
	assert.Equal(s.T(), http.StatusUnauthorized, resp4.StatusCode, "Revoked token should not work")

	// New token should work
	resp5, _ := s.relayTokenRequest("POST", "/api/v1/relay/claude/v1/messages", token2)
	assert.NotEqual(s.T(), http.StatusUnauthorized, resp5.StatusCode, "New token should authenticate")

	// Clean up
	s.adminRequest("DELETE", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
}

// TestRelayTokenReuseAfterRegenerate tests that generating a new token invalidates the old one.
func (s *RelayTokenTestSuite) TestRelayTokenReuseAfterRegenerate() {
	ctx := context.Background()

	profileResp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	adminID := profileResp.JSON200.User.Id

	// Generate first token
	resp1, body1 := s.adminRequest("POST", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusCreated, resp1.StatusCode)
	token1 := body1["token"].(string)

	// Verify it works
	resp2, _ := s.relayTokenRequest("POST", "/api/v1/relay/claude/v1/messages", token1)
	assert.NotEqual(s.T(), http.StatusUnauthorized, resp2.StatusCode, "First token should work")

	// Generate second token (replaces first)
	resp3, _ := s.adminRequest("POST", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusCreated, resp3.StatusCode)

	// First token should no longer work (replaced by second)
	resp4, _ := s.relayTokenRequest("POST", "/api/v1/relay/claude/v1/messages", token1)
	assert.Equal(s.T(), http.StatusUnauthorized, resp4.StatusCode, "Old token should be invalid after regeneration")

	// Clean up
	s.adminRequest("DELETE", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
}

// --- Relay Endpoint Tests ---

// TestRelayTokenChatCompletionsEndpoint tests that relay tokens work on chat completions endpoint too.
func (s *RelayTokenTestSuite) TestRelayTokenChatCompletionsEndpoint() {
	ctx := context.Background()

	profileResp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	adminID := profileResp.JSON200.User.Id

	// Generate token
	resp, body := s.adminRequest("POST", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	rawToken := body["token"].(string)

	// Use on chat completions endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", s.ServerURL+"/api/v1/relay/openai/v1/chat/completions", nil)
	require.NoError(s.T(), err)
	req.Header.Set("Authorization", "Bearer "+rawToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	authResp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer authResp.Body.Close()

	// Should NOT be 401 — auth should work
	assert.NotEqual(s.T(), http.StatusUnauthorized, authResp.StatusCode, "Relay token should work on chat completions endpoint")

	// Clean up
	s.adminRequest("DELETE", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
}

// TestRelayTokenWithDifferentTools tests that the :tool parameter is properly routed.
func (s *RelayTokenTestSuite) TestRelayTokenWithDifferentTools() {
	ctx := context.Background()

	profileResp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	adminID := profileResp.JSON200.User.Id

	// Generate token
	resp, body := s.adminRequest("POST", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	rawToken := body["token"].(string)

	client := &http.Client{}

	// Test multiple tool names — all should authenticate (not 401)
	tools := []string{"claude", "openai", "anthropic", "gemini"}
	for _, tool := range tools {
		req, err := http.NewRequestWithContext(ctx, "POST", s.ServerURL+"/api/v1/relay/"+tool+"/v1/messages", nil)
		require.NoError(s.T(), err)
		req.Header.Set("Authorization", "Bearer "+rawToken)
		req.Header.Set("Content-Type", "application/json")

		authResp, err := client.Do(req)
		require.NoError(s.T(), err)
		defer authResp.Body.Close()

		assert.NotEqual(s.T(), http.StatusUnauthorized, authResp.StatusCode,
			"Relay token should work for tool=%s", tool)
	}

	// Clean up
	s.adminRequest("DELETE", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
}

// --- Edge Case Tests ---

// TestRelayTokenNegativeUserID tests that negative user IDs are handled.
func (s *RelayTokenTestSuite) TestRelayTokenNegativeUserID() {
	resp, _ := s.adminRequest("POST", "/api/v1/users/-1/relay-token", nil)
	// Negative IDs should either be 400 (invalid) or 404 (user not found)
	assert.True(s.T(), resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound,
		"Negative user ID should return 400 or 404, got %d", resp.StatusCode)
}

// TestRelayTokenZeroUserID tests that zero user ID is handled.
func (s *RelayTokenTestSuite) TestRelayTokenZeroUserID() {
	resp, _ := s.adminRequest("POST", "/api/v1/users/0/relay-token", nil)
	// Zero ID should either be 400 or 404
	assert.True(s.T(), resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound,
		"Zero user ID should return 400 or 404, got %d", resp.StatusCode)
}

// TestRelayTokenWrongHTTPMethod tests that only allowed HTTP methods work on token endpoints.
func (s *RelayTokenTestSuite) TestRelayTokenWrongHTTPMethod() {
	adminID := int64(1)

	// PUT should not be allowed on generate endpoint (Gin returns 404 for unregistered method routes)
	resp, _ := s.adminRequest("PUT", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	assert.True(s.T(), resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusNotFound,
		"PUT should return 405 or 404, got %d", resp.StatusCode)

	// PATCH should not be allowed
	resp, _ = s.adminRequest("PATCH", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	assert.True(s.T(), resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusNotFound,
		"PATCH should return 405 or 404, got %d", resp.StatusCode)
}

// TestRelayTokenDoubleRevoke tests that revoking an already-revoked token is idempotent.
func (s *RelayTokenTestSuite) TestRelayTokenDoubleRevoke() {
	ctx := context.Background()

	profileResp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	adminID := profileResp.JSON200.User.Id

	// Generate token
	resp, _ := s.adminRequest("POST", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	// Revoke once
	resp2, _ := s.adminRequest("DELETE", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusOK, resp2.StatusCode)

	// Revoke again — should be idempotent (200 OK)
	resp3, _ := s.adminRequest("DELETE", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	assert.Equal(s.T(), http.StatusOK, resp3.StatusCode, "Double revoke should be idempotent (200 OK)")
}

// TestRelayTokenLastUsedAtField tests that last_used_at is returned after token usage.
func (s *RelayTokenTestSuite) TestRelayTokenLastUsedAtField() {
	ctx := context.Background()

	profileResp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	adminID := profileResp.JSON200.User.Id

	// Generate token
	resp, body := s.adminRequest("POST", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	rawToken := body["token"].(string)

	// Get info before use — last_used_at may or may not be present initially
	s.adminRequest("GET", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)

	// Use the token on a relay endpoint
	s.relayTokenRequest("POST", "/api/v1/relay/claude/v1/messages", rawToken)

	// Wait for async last-used update to complete (runs in a goroutine)
	time.Sleep(200 * time.Millisecond)

	// Get info after use — last_used_at should now be present
	infoResp2, infoBody2 := s.adminRequest("GET", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusOK, infoResp2.StatusCode)
	_, hasLastUsed2 := infoBody2["last_used_at"]
	assert.True(s.T(), hasLastUsed2, "last_used_at should be set after token usage")

	// Clean up
	s.adminRequest("DELETE", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
}

// TestRelayTokenResponseStructure tests that all response fields have the correct types.
func (s *RelayTokenTestSuite) TestRelayTokenResponseStructure() {
	ctx := context.Background()

	profileResp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	adminID := profileResp.JSON200.User.Id

	// Generate — check response structure
	resp, body := s.adminRequest("POST", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	_, isString := body["token"].(string)
	assert.True(s.T(), isString, "token should be a string")

	_, isString = body["prefix"].(string)
	assert.True(s.T(), isString, "prefix should be a string")

	_, isString = body["message"].(string)
	assert.True(s.T(), isString, "message should be a string")

	_, isString = body["created_at"].(string)
	assert.True(s.T(), isString, "created_at should be a string")

	// Info — check response structure
	infoResp, infoBody := s.adminRequest("GET", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusOK, infoResp.StatusCode)

	// Raw token should NOT be in info response
	_, hasToken := infoBody["token"]
	assert.False(s.T(), hasToken, "Info response should not contain raw token")

	_, isString = infoBody["prefix"].(string)
	assert.True(s.T(), isString, "info prefix should be a string")

	_, isString = infoBody["created_at"].(string)
	assert.True(s.T(), isString, "info created_at should be a string")

	// Revoke — check response structure
	revokeResp, revokeBody := s.adminRequest("DELETE", "/api/v1/users/"+strconv.FormatInt(adminID, 10)+"/relay-token", nil)
	require.Equal(s.T(), http.StatusOK, revokeResp.StatusCode)

	_, isString = revokeBody["message"].(string)
	assert.True(s.T(), isString, "revoke message should be a string")
}

// TestRelayTokenTestSuite is the testify entry point.
func TestRelayTokenTestSuite(t *testing.T) {
	suite.Run(t, new(RelayTokenTestSuite))
}
