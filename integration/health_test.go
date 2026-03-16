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
	"context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthCheck tests the health endpoint.
func (s *IntegrationTestSuite) TestHealthCheck() {
	ctx := context.Background()

	// Health check should succeed (using anonymous client - no auth required)
	resp, err := s.AnonymousClient.GetHealthWithResponse(ctx)
	require.NoError(s.T(), err, "Health check request should succeed")
	assert.Equal(s.T(), 200, resp.StatusCode(), "Health check should return 200")

	// Verify response structure
	require.NotNil(s.T(), resp.JSON200, "Response should have JSON200 body")
	assert.NotEmpty(s.T(), resp.JSON200.Version, "Health check should return version")
	assert.NotEmpty(s.T(), resp.JSON200.Status, "Health check should return status")
	assert.NotEmpty(s.T(), resp.JSON200.Timestamp, "Health check should return timestamp")
	assert.NotNil(s.T(), resp.JSON200.Database, "Health check should return database status")
}

// TestHealthCheckDatabaseStatus tests that the database status is reported correctly.
func (s *IntegrationTestSuite) TestHealthCheckDatabaseStatus() {
	ctx := context.Background()

	resp, err := s.AnonymousClient.GetHealthWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, resp.StatusCode())

	// Database should be up
	assert.Equal(s.T(), "up", string(resp.JSON200.Database.Status), "Database should be up")
}
