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

	integration_manager "github.com/code-together/integration_manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Dashboard Tests (6 tests)
// ============================================================================

func (s *IntegrationTestSuite) TestMgrGetDashboardMetrics() {
	ctx := context.Background()

	// Get dashboard metrics with default range
	resp, err := s.ManagerClient.GetApiV1DashboardMetricsWithResponse(ctx, &integration_manager.GetApiV1DashboardMetricsParams{})
	require.NoError(s.T(), err, "Failed to get dashboard metrics")
	require.Equal(s.T(), 200, resp.StatusCode(), "Get dashboard metrics should return 200")
	require.NotNil(s.T(), resp.JSON200)

	// Verify response structure
	assert.NotNil(s.T(), resp.JSON200.Period)
	assert.NotNil(s.T(), resp.JSON200.Interval)
	assert.NotNil(s.T(), resp.JSON200.DataPoints)
}

func (s *IntegrationTestSuite) TestMgrGetDashboardMetrics7d() {
	ctx := context.Background()

	// Get dashboard metrics for 7 days
	sevenDays := integration_manager.GetApiV1DashboardMetricsParamsRange("7d")
	resp, err := s.ManagerClient.GetApiV1DashboardMetricsWithResponse(ctx, &integration_manager.GetApiV1DashboardMetricsParams{
		Range:    &sevenDays,
		Interval: nil,
	})
	require.NoError(s.T(), err, "Failed to get dashboard metrics")
	require.Equal(s.T(), 200, resp.StatusCode())
	require.NotNil(s.T(), resp.JSON200)
}

func (s *IntegrationTestSuite) TestMgrGetDashboardMetrics24h() {
	ctx := context.Background()

	// Get dashboard metrics for 24 hours
	twentyFourHours := integration_manager.GetApiV1DashboardMetricsParamsRange("24h")
	hourInterval := integration_manager.GetApiV1DashboardMetricsParamsInterval("hour")
	resp, err := s.ManagerClient.GetApiV1DashboardMetricsWithResponse(ctx, &integration_manager.GetApiV1DashboardMetricsParams{
		Range:    &twentyFourHours,
		Interval: &hourInterval,
	})
	require.NoError(s.T(), err, "Failed to get dashboard metrics")
	require.Equal(s.T(), 200, resp.StatusCode())
	require.NotNil(s.T(), resp.JSON200)
}

func (s *IntegrationTestSuite) TestMgrGetDashboardRankings() {
	ctx := context.Background()

	// Get provider rankings
	resp, err := s.ManagerClient.GetApiV1DashboardRankingsWithResponse(ctx)
	require.NoError(s.T(), err, "Failed to get dashboard rankings")
	require.Equal(s.T(), 200, resp.StatusCode(), "Get dashboard rankings should return 200")
	require.NotNil(s.T(), resp.JSON200)

	// Verify response structure
	assert.NotEmpty(s.T(), resp.JSON200.Period)
	assert.NotNil(s.T(), resp.JSON200.Rankings)
}

func (s *IntegrationTestSuite) TestMgrGetDashboardMembers() {
	ctx := context.Background()

	// Get member statistics (manager-only)
	resp, err := s.ManagerClient.GetApiV1DashboardMembersWithResponse(ctx)
	require.NoError(s.T(), err, "Failed to get dashboard members")
	require.Equal(s.T(), 200, resp.StatusCode(), "Get dashboard members should return 200")
	require.NotNil(s.T(), resp.JSON200)

	// Verify response structure
	assert.NotNil(s.T(), resp.JSON200.Members)
}

func (s *IntegrationTestSuite) TestMgrGetDashboardMetrics30d() {
	ctx := context.Background()

	// Get dashboard metrics for 30 days
	thirtyDays := integration_manager.GetApiV1DashboardMetricsParamsRange("30d")
	dayInterval := integration_manager.GetApiV1DashboardMetricsParamsInterval("day")
	resp, err := s.ManagerClient.GetApiV1DashboardMetricsWithResponse(ctx, &integration_manager.GetApiV1DashboardMetricsParams{
		Range:    &thirtyDays,
		Interval: &dayInterval,
	})
	require.NoError(s.T(), err, "Failed to get dashboard metrics")
	require.Equal(s.T(), 200, resp.StatusCode())
	require.NotNil(s.T(), resp.JSON200)
}
