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
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// RelayTokenTestSuite tests relay token functionality.
// Extends IntegrationTestSuite but skips bootstrapStandardFixture since
// relay token tests only need admin auth, not providers/teams/license.
type RelayTokenTestSuite struct {
	IntegrationTestSuite
}

// SetupTest runs before each test. Skips bootstrap fixture setup.
func (s *RelayTokenTestSuite) SetupTest() {
	s.loginAdminUser()
	// Skip bootstrap — relay token tests don't need providers/teams
}

// TearDownSuite delegates to parent.
func (s *RelayTokenTestSuite) TearDownSuite() {
	s.IntegrationTestSuite.TearDownSuite()
}

// TearDownTest delegates to parent.
func (s *RelayTokenTestSuite) TearDownTest() {
	s.IntegrationTestSuite.TearDownTest()
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

// TestRelayTokenTestSuite is the testify entry point.
func TestRelayTokenTestSuite(t *testing.T) {
	suite.Run(t, new(RelayTokenTestSuite))
}
