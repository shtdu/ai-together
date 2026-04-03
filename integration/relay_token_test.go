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
	"testing"
	"time"

	integrationclient "github.com/code-together/shared/integration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// RelayTokenTestSuite tests relay token functionality.
// Standalone suite — does NOT embed IntegrationTestSuite to avoid inheriting unrelated tests.
type RelayTokenTestSuite struct {
	suite.Suite
	ServerURL  string
	AdminToken string
	Client     *integrationclient.ClientWithResponses
	ctx        *TestContext
	logger     *slog.Logger
}

func (s *RelayTokenTestSuite) SetupSuite() {
	ctx, err := setupTestContext()
	s.Require().NoError(err)
	s.ctx = ctx
	s.ServerURL = ctx.ServerURL
	s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	s.loginAdmin()
	s.T().Logf("RelayTokenTestSuite: connected to %s", s.ServerURL)
}

func (s *RelayTokenTestSuite) TearDownSuite() {
	if s.ctx != nil {
		cleanupTestContext(s.ctx)
	}
}

func (s *RelayTokenTestSuite) loginAdmin() {
	body, _ := json.Marshal(map[string]string{"email": "admin@example.com", "password": "AdminPassword123!"})
	req, _ := http.NewRequestWithContext(context.Background(), "POST", s.ServerURL+"/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{}).Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	s.AdminToken = result["access_token"].(string)
	s.Require().NotEmpty(s.AdminToken)

	client, err := integrationclient.NewAuthenticatedClient(s.ServerURL, func() (string, error) { return s.AdminToken, nil }, s.logger, false)
	s.Require().NoError(err)
	s.Client = client
}

// --- Helpers ---

func (s *RelayTokenTestSuite) authReq(method, path string, body interface{}) (*http.Response, map[string]interface{}) {
	var reqBody io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	}
	req, _ := http.NewRequestWithContext(context.Background(), method, s.ServerURL+path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.AdminToken)
	resp, err := (&http.Client{}).Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	var m map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&m)
	return resp, m
}

func (s *RelayTokenTestSuite) tokenReq(method, path, token string) (*http.Response, map[string]interface{}) {
	req, _ := http.NewRequestWithContext(context.Background(), method, s.ServerURL+path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := (&http.Client{}).Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	var m map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&m)
	return resp, m
}

// --- Generate Tests ---

func (s *RelayTokenTestSuite) TestGenerateSuccess() {
	resp, body := s.authReq("POST", "/api/v1/user/relay-token", nil)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	assert.NotEmpty(s.T(), body["token"])
	assert.NotEmpty(s.T(), body["prefix"])
	assert.NotEmpty(s.T(), body["created_at"])
	assert.Len(s.T(), body["token"].(string), 64)
	s.authReq("DELETE", "/api/v1/user/relay-token", nil)
}

func (s *RelayTokenTestSuite) TestGenerateDuplicate() {
	r1, b1 := s.authReq("POST", "/api/v1/user/relay-token", nil)
	require.Equal(s.T(), http.StatusCreated, r1.StatusCode)
	t1 := b1["token"].(string)

	r2, b2 := s.authReq("POST", "/api/v1/user/relay-token", nil)
	require.Equal(s.T(), http.StatusCreated, r2.StatusCode)
	t2 := b2["token"].(string)

	assert.NotEqual(s.T(), t1, t2)
	s.authReq("DELETE", "/api/v1/user/relay-token", nil)
}

// --- Get Info Tests ---

func (s *RelayTokenTestSuite) TestGetInfoSuccess() {
	s.authReq("POST", "/api/v1/user/relay-token", nil)

	resp, body := s.authReq("GET", "/api/v1/user/relay-token", nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	assert.NotEmpty(s.T(), body["prefix"])
	assert.NotEmpty(s.T(), body["created_at"])
	_, hasToken := body["token"]
	assert.False(s.T(), hasToken, "Info should not expose raw token")

	s.authReq("DELETE", "/api/v1/user/relay-token", nil)
}

func (s *RelayTokenTestSuite) TestGetInfoNoToken() {
	s.authReq("DELETE", "/api/v1/user/relay-token", nil)
	resp, _ := s.authReq("GET", "/api/v1/user/relay-token", nil)
	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)
}

// --- Revoke Tests ---

func (s *RelayTokenTestSuite) TestRevokeSuccess() {
	s.authReq("POST", "/api/v1/user/relay-token", nil)

	resp, body := s.authReq("DELETE", "/api/v1/user/relay-token", nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	assert.Contains(s.T(), body["message"], "revoked")

	resp2, _ := s.authReq("GET", "/api/v1/user/relay-token", nil)
	assert.Equal(s.T(), http.StatusNotFound, resp2.StatusCode)
}

func (s *RelayTokenTestSuite) TestDoubleRevoke() {
	s.authReq("POST", "/api/v1/user/relay-token", nil)
	s.authReq("DELETE", "/api/v1/user/relay-token", nil)
	resp, _ := s.authReq("DELETE", "/api/v1/user/relay-token", nil)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode, "Double revoke should be idempotent")
}

// --- Auth Tests ---

func (s *RelayTokenTestSuite) TestUnauthorized() {
	req, _ := http.NewRequestWithContext(context.Background(), "POST", s.ServerURL+"/api/v1/user/relay-token", nil)
	resp, _ := (&http.Client{}).Do(req)
	defer resp.Body.Close()
	assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)
}

func (s *RelayTokenTestSuite) TestAuthFlow() {
	// Generate
	_, b1 := s.authReq("POST", "/api/v1/user/relay-token", nil)
	token := b1["token"].(string)

	// Use on relay endpoint
	resp, _ := s.tokenReq("POST", "/api/v1/relay/claude/v1/messages", token)
	assert.NotEqual(s.T(), http.StatusUnauthorized, resp.StatusCode, "Token should authenticate")

	// Revoke
	s.authReq("DELETE", "/api/v1/user/relay-token", nil)

	// Token should no longer work
	resp2, _ := s.tokenReq("POST", "/api/v1/relay/claude/v1/messages", token)
	assert.Equal(s.T(), http.StatusUnauthorized, resp2.StatusCode, "Revoked token rejected")
}

func (s *RelayTokenTestSuite) TestJWTFallback() {
	req, _ := http.NewRequestWithContext(context.Background(), "POST", s.ServerURL+"/api/v1/relay/claude/v1/messages", nil)
	req.Header.Set("Authorization", "Bearer "+s.AdminToken)
	resp, _ := (&http.Client{}).Do(req)
	defer resp.Body.Close()
	assert.NotEqual(s.T(), http.StatusUnauthorized, resp.StatusCode, "JWT should still work on relay endpoints")
}

func (s *RelayTokenTestSuite) TestInvalidToken() {
	resp, _ := s.tokenReq("POST", "/api/v1/relay/claude/v1/messages", "invalid_token_64chars_aaaaaaaaaaaaaaaaaaaaaaaaaa")
	assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)
}

func (s *RelayTokenTestSuite) TestNoAuthHeader() {
	req, _ := http.NewRequestWithContext(context.Background(), "POST", s.ServerURL+"/api/v1/relay/claude/v1/messages", nil)
	resp, _ := (&http.Client{}).Do(req)
	defer resp.Body.Close()
	assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)
}

func (s *RelayTokenTestSuite) TestTokenTooShort() {
	resp, _ := s.tokenReq("POST", "/api/v1/relay/claude/v1/messages", "abc123")
	assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)
}

// --- Token Format ---

func (s *RelayTokenTestSuite) TestTokenFormat() {
	_, body := s.authReq("POST", "/api/v1/user/relay-token", nil)
	rawToken := body["token"].(string)
	prefix := body["prefix"].(string)

	assert.Len(s.T(), rawToken, 64)
	assert.Regexp(s.T(), `^[0-9a-f]{64}$`, rawToken)
	assert.Len(s.T(), prefix, 8)
	assert.Equal(s.T(), rawToken[:8], prefix)
	assert.NotContains(s.T(), rawToken, ".", "Token should not contain dots (JWT separator)")

	s.authReq("DELETE", "/api/v1/user/relay-token", nil)
}

// --- Lifecycle ---

func (s *RelayTokenTestSuite) TestRevokeThenRegenerate() {
	_, b1 := s.authReq("POST", "/api/v1/user/relay-token", nil)
	t1 := b1["token"].(string)

	s.authReq("DELETE", "/api/v1/user/relay-token", nil)

	_, b2 := s.authReq("POST", "/api/v1/user/relay-token", nil)
	t2 := b2["token"].(string)
	assert.NotEqual(s.T(), t1, t2)

	// Old token rejected
	resp, _ := s.tokenReq("POST", "/api/v1/relay/claude/v1/messages", t1)
	assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)

	// New token works
	resp2, _ := s.tokenReq("POST", "/api/v1/relay/claude/v1/messages", t2)
	assert.NotEqual(s.T(), http.StatusUnauthorized, resp2.StatusCode)

	s.authReq("DELETE", "/api/v1/user/relay-token", nil)
}

func (s *RelayTokenTestSuite) TestRegenerateInvalidatesOld() {
	_, b1 := s.authReq("POST", "/api/v1/user/relay-token", nil)
	t1 := b1["token"].(string)

	// Verify old works
	r1, _ := s.tokenReq("POST", "/api/v1/relay/claude/v1/messages", t1)
	assert.NotEqual(s.T(), http.StatusUnauthorized, r1.StatusCode)

	// Regenerate
	s.authReq("POST", "/api/v1/user/relay-token", nil)

	// Old should be invalid
	r2, _ := s.tokenReq("POST", "/api/v1/relay/claude/v1/messages", t1)
	assert.Equal(s.T(), http.StatusUnauthorized, r2.StatusCode)

	s.authReq("DELETE", "/api/v1/user/relay-token", nil)
}

// --- Relay Endpoints ---

func (s *RelayTokenTestSuite) TestChatCompletionsEndpoint() {
	_, body := s.authReq("POST", "/api/v1/user/relay-token", nil)
	token := body["token"].(string)

	req, _ := http.NewRequestWithContext(context.Background(), "POST", s.ServerURL+"/api/v1/relay/openai/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := (&http.Client{}).Do(req)
	defer resp.Body.Close()
	assert.NotEqual(s.T(), http.StatusUnauthorized, resp.StatusCode)

	s.authReq("DELETE", "/api/v1/user/relay-token", nil)
}

func (s *RelayTokenTestSuite) TestDifferentToolNames() {
	_, body := s.authReq("POST", "/api/v1/user/relay-token", nil)
	token := body["token"].(string)

	for _, tool := range []string{"claude", "openai", "anthropic", "gemini"} {
		req, _ := http.NewRequestWithContext(context.Background(), "POST", s.ServerURL+"/api/v1/relay/"+tool+"/v1/messages", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ := (&http.Client{}).Do(req)
		defer resp.Body.Close()
		assert.NotEqual(s.T(), http.StatusUnauthorized, resp.StatusCode, "tool=%s", tool)
	}

	s.authReq("DELETE", "/api/v1/user/relay-token", nil)
}

// --- Data & Structure ---

func (s *RelayTokenTestSuite) TestLastUsedAt() {
	_, body := s.authReq("POST", "/api/v1/user/relay-token", nil)
	token := body["token"].(string)

	// Use token
	s.tokenReq("POST", "/api/v1/relay/claude/v1/messages", token)
	time.Sleep(200 * time.Millisecond)

	// Check last_used_at populated
	_, infoBody := s.authReq("GET", "/api/v1/user/relay-token", nil)
	_, hasLastUsed := infoBody["last_used_at"]
	assert.True(s.T(), hasLastUsed, "last_used_at should be set after usage")

	s.authReq("DELETE", "/api/v1/user/relay-token", nil)
}

func (s *RelayTokenTestSuite) TestResponseStructure() {
	resp, body := s.authReq("POST", "/api/v1/user/relay-token", nil)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	_, isStr := body["token"].(string)
	assert.True(s.T(), isStr)
	_, isStr = body["prefix"].(string)
	assert.True(s.T(), isStr)
	_, isStr = body["message"].(string)
	assert.True(s.T(), isStr)

	infoResp, infoBody := s.authReq("GET", "/api/v1/user/relay-token", nil)
	require.Equal(s.T(), http.StatusOK, infoResp.StatusCode)
	_, hasToken := infoBody["token"]
	assert.False(s.T(), hasToken)

	revokeResp, revokeBody := s.authReq("DELETE", "/api/v1/user/relay-token", nil)
	require.Equal(s.T(), http.StatusOK, revokeResp.StatusCode)
	_, isStr = revokeBody["message"].(string)
	assert.True(s.T(), isStr)
}

// TestRelayTokenTestSuite is the testify entry point.
func TestRelayTokenTestSuite(t *testing.T) {
	suite.Run(t, new(RelayTokenTestSuite))
}
