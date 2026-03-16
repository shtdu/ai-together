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

package integration

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	integration_manager "github.com/code-together/integration_manager"
	integrationclient "github.com/code-together/shared/integration"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite is the main test suite for integration tests.
// It follows black-box testing principles - using only the API client,
// never importing server packages.
type IntegrationTestSuite struct {
	suite.Suite

	ServerURL       string
	TestDBURL       string
	Logger          *slog.Logger
	AnonymousClient *integrationclient.ClientWithResponses
	Client          *integrationclient.ClientWithResponses
	AdminToken      string
	MemberToken     string
	TestContext     *TestContext
	skipBootstrap   bool // Flag to skip bootstrapStandardFixture() for tests needing clean state
	// Manager client for testing manager endpoints
	ManagerClient *integration_manager.ClientWithResponses
}

// SetupSuite runs once before all tests in the suite.
// It initializes the test environment and connects to the running test server.
func (s *IntegrationTestSuite) SetupSuite() {
	// Initialize test context (connects to running server on localhost:8081)
	ctx, err := setupTestContext()
	s.Require().NoError(err, "Failed to setup test context")
	s.TestContext = ctx

	s.ServerURL = ctx.ServerURL
	s.TestDBURL = ctx.TestDBURL
	s.Logger = ctx.Logger

	// Create anonymous client
	anonClient, err := integrationclient.NewAnonymousClient(s.ServerURL, s.Logger, false)
	s.Require().NoError(err, "Failed to create anonymous client")
	s.AnonymousClient = anonClient

	s.Logger.Info("Test suite setup complete")
}

// TearDownSuite runs once after all tests in the suite.
func (s *IntegrationTestSuite) TearDownSuite() {
	s.Logger.Info("Tearing down test suite")

	if s.TestContext != nil {
		cleanupTestContext(s.TestContext)
	}

	s.Logger.Info("Test suite teardown complete")
}

// SetupTest runs before each test.
// Does NOT cleanup - tests should handle existing database state gracefully.
func (s *IntegrationTestSuite) SetupTest() {
	// Login admin user (created by test-server.sh)
	s.loginAdminUser()

	// Bootstrap standard fixture (creates license, providers, member user)
	// Skip if test requests clean state via s.skipBootstrap = true
	if !s.skipBootstrap {
		s.bootstrapStandardFixture()
	}
}

// TearDownTest runs after each test.
// It cleans up resources created during the test via API calls only.
func (s *IntegrationTestSuite) TearDownTest() {
	ctx := context.Background()

	// Clean up providers (delete all created during test)
	if s.Client != nil {
		// List all providers
		providersResp, err := s.Client.GetApiV1ProvidersWithResponse(ctx)
		if err == nil && providersResp.StatusCode() == 200 && providersResp.JSON200 != nil {
			for _, provider := range *providersResp.JSON200 {
				// Delete each provider via API
				s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, provider.Id)
			}
		}
	}

	// Clean up teams using manager client
	if s.ManagerClient != nil {
		// List all teams
		teamsResp, err := s.ManagerClient.GetApiV1TeamsWithResponse(ctx)
		if err == nil && teamsResp.StatusCode() == 200 && teamsResp.JSON200 != nil {
			for _, team := range *teamsResp.JSON200 {
				// Delete each team via API (skip default team with ID 1)
				if team.Id != 1 {
					s.ManagerClient.DeleteApiV1TeamsTeamIdWithResponse(ctx, team.Id)
				}
			}
		}
	}

	s.Logger.Info("Test cleanup complete")
}

// loginAdminUser logs in the admin user that was created by test-server.sh.
func (s *IntegrationTestSuite) loginAdminUser() {
	ctx := context.Background()

	// Load fixtures to get admin credentials
	fixtures, err := GetFixtureData()
	s.Require().NoError(err, "Failed to load fixtures")

	adminUser := fixtures.Users["admin"]
	s.Require().NotEmpty(adminUser.Email, "Admin user fixture not found")

	// Login admin via API
	email := openapi_types.Email(adminUser.Email)
	loginReq := integrationclient.PostAuthLoginJSONRequestBody{
		Email:    email,
		Password: adminUser.Password,
	}
	loginResp, err := s.AnonymousClient.PostAuthLoginWithResponse(ctx, loginReq)
	s.Require().NoError(err, "Failed to login admin user")
	s.Require().Equal(200, loginResp.StatusCode(), "Login should return 200")
	s.Require().NotNil(loginResp.JSON200, "Login response should have JSON200 body")

	s.AdminToken = loginResp.JSON200.AccessToken

	// Create authenticated client
	s.Client, err = integrationclient.NewAuthenticatedClient(
		s.ServerURL,
		func() (string, error) { return s.AdminToken, nil },
		s.Logger,
		false,
	)
	s.Require().NoError(err, "Failed to create authenticated client")

	// Create manager client for testing manager endpoints
	s.ManagerClient, err = integration_manager.NewAuthenticatedClient(
		s.ServerURL,
		func() (string, error) { return s.AdminToken, nil },
		s.Logger,
	)
	s.Require().NoError(err, "Failed to create manager client")
}

// createAuthenticatedClient creates a new authenticated client with the given token.
func (s *IntegrationTestSuite) createAuthenticatedClient(token string) *integrationclient.ClientWithResponses {
	client, err := integrationclient.NewAuthenticatedClient(
		s.ServerURL,
		func() (string, error) { return token, nil },
		s.Logger,
		false,
	)
	s.Require().NoError(err, "Failed to create authenticated client")
	return client
}

// registerAndLoginUserFromFixture registers a user from fixture and logs in, returning the access token.
// If user already exists, skips registration and just logs in.
func (s *IntegrationTestSuite) registerAndLoginUserFromFixture(fixtureName string) string {
	ctx := context.Background()

	// Load fixtures
	fixtures, err := GetFixtureData()
	s.Require().NoError(err, "Failed to load fixtures")

	user, ok := fixtures.Users[fixtureName]
	s.Require().True(ok, "User fixture not found: %s", fixtureName)

	// Convert email string to openapi_types.Email
	emailAddr := openapi_types.Email(user.Email)

	// Try to login first (user might already exist)
	loginReq := integrationclient.PostAuthLoginJSONRequestBody{
		Email:    emailAddr,
		Password: user.Password,
	}
	loginResp, err := s.AnonymousClient.PostAuthLoginWithResponse(ctx, loginReq)

	// If login fails, try to register
	if err != nil || loginResp.StatusCode() != 200 {
		regReq := integrationclient.PostAuthRegisterJSONRequestBody{
			Email:    emailAddr,
			Password: user.Password,
			Name:     user.Name,
		}
		regResp, err := s.AnonymousClient.PostAuthRegisterWithResponse(ctx, regReq)
		s.Require().NoError(err, "Failed to register user")

		// Accept 201 (created) or 500 (duplicate user - already exists)
		if regResp.StatusCode() == 500 {
			// User might already exist, try login again
			loginResp, err = s.AnonymousClient.PostAuthLoginWithResponse(ctx, loginReq)
			s.Require().NoError(err, "Failed to login after registration error")
		} else {
			s.Require().Equal(201, regResp.StatusCode(), "Registration should return 201")

			// Login after successful registration
			loginResp, err = s.AnonymousClient.PostAuthLoginWithResponse(ctx, loginReq)
			s.Require().NoError(err, "Failed to login user")
		}
	}

	s.Require().Equal(200, loginResp.StatusCode(), "Login should return 200")
	s.Require().NotNil(loginResp.JSON200, "Login response should have JSON200 body")

	return loginResp.JSON200.AccessToken
}

// activateLicenseFixture loads a license PEM file and activates it via API,
// then fetches and returns the license status.
func (s *IntegrationTestSuite) activateLicenseFixture(fixtureName string) *integrationclient.LicenseStatus {
	ctx := context.Background()

	// Load license PEM from file
	licensePEM, err := LoadLicenseFixture(fixtureName)
	s.Require().NoError(err, "Failed to load license fixture")

	// Activate license via API
	activateResp, err := s.Client.PostApiV1LicenseActivateWithResponse(ctx, integrationclient.PostApiV1LicenseActivateJSONRequestBody{
		LicenseKey: licensePEM,
	})
	s.Require().NoError(err, "Failed to activate license")
	s.Require().Equal(200, activateResp.StatusCode(), "License activation should return 200")
	s.Require().NotNil(activateResp.JSON200, "License activation response should have JSON200 body")

	// Fetch license status via API
	statusResp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	s.Require().NoError(err, "Failed to get license status")
	s.Require().Equal(200, statusResp.StatusCode(), "Get license should return 200")
	s.Require().NotNil(statusResp.JSON200, "License status response should have JSON200 body")
	s.Require().NotNil(statusResp.JSON200.License, "License status should contain license")

	return statusResp.JSON200.License
}

// createProviderFixtureFromFixture creates a provider from fixture data and returns its ID.
func (s *IntegrationTestSuite) createProviderFixtureFromFixture(fixtureName string) int64 {
	ctx := context.Background()

	// Load fixtures
	fixtures, err := GetFixtureData()
	s.Require().NoError(err, "Failed to load fixtures")

	provider, ok := fixtures.Providers[fixtureName]
	s.Require().True(ok, "Provider fixture not found: %s", fixtureName)

	// Convert string kind to CreateProviderRequestKind
	providerKind := integrationclient.CreateProviderRequestKind(provider.Kind)

	// Generate unique name to avoid duplicate name errors across test runs
	uniqueName := generateUniqueProviderName(provider.Name)

	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &providerKind,
		ApiKey:          provider.APIKey,
		ApiUrl:          provider.APIURL,
		Enabled:         &provider.Enabled,
		Level:           &provider.Level,
		SupportedModels: &provider.SupportedModels,
	}

	resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	s.Require().NoError(err, "Failed to create provider")
	s.Require().Equal(201, resp.StatusCode(), "Provider creation should return 201")
	s.Require().NotNil(resp.JSON201, "Provider creation response should have JSON201 body")

	return resp.JSON201.Id
}

// bootstrapStandardFixture creates a standard test environment with:
// - Admin user (already created in SetupTest)
// - Tier 1 license activated
// - 2 providers (Claude and OpenAI)
// - Member user
//
// This returns a TestFixture struct with all the created entities.
func (s *IntegrationTestSuite) bootstrapStandardFixture() *TestFixture {
	// Activate Commercial license
	license := s.activateLicenseFixture(LicenseCommercial)

	// Create Claude provider
	claudeProviderID := s.createProviderFixtureFromFixture("claude")

	// Create OpenCode provider
	opencodeProviderID := s.createProviderFixtureFromFixture("opencode")

	// Create member user
	memberToken := s.registerAndLoginUserFromFixture("member")
	s.MemberToken = memberToken

	return &TestFixture{
		License:            license,
		ClaudeProviderID:   claudeProviderID,
		OpenCodeProviderID: opencodeProviderID,
		MemberToken:        memberToken,
	}
}

// TestFixture holds pre-created test entities for convenience.
type TestFixture struct {
	License            *integrationclient.LicenseStatus
	ClaudeProviderID   int64
	OpenCodeProviderID int64
	MemberToken        string
}

// TestIntegrationSuite runs the integration test suite.
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// TestMain is the entry point for the test suite.
func TestMain(m *testing.M) {
	// Run tests
	exitCode := m.Run()
	os.Exit(exitCode)
}

// requireErrorResponse validates an error response has expected properties
func requireErrorResponse(s *IntegrationTestSuite, errResp *integrationclient.ErrorResponse, expectedKeyword string) {
	s.T().Helper()
	require.NotNil(s.T(), errResp, "Error response should not be nil")
	assert.NotEmpty(s.T(), errResp.Error, "Error message should not be empty")

	if expectedKeyword != "" {
		// Case-insensitive check for keyword
		assert.Contains(s.T(), strings.ToLower(errResp.Error), strings.ToLower(expectedKeyword),
			"Error message should contain expected keyword: "+expectedKeyword)
	}
}
