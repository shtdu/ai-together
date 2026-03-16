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

package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_JSONSerialization(t *testing.T) {
	user := User{
		ID:        1,
		Email:     "test@example.com",
		Name:      "Test User",
		Password:  "hashed_password",
		Role:      "manager",
		TenantID:  1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(user)
	require.NoError(t, err)

	// Password should not be exposed in JSON
	assert.NotContains(t, string(data), "hashed_password")
	assert.Contains(t, string(data), "test@example.com")

	// Unmarshal back
	var decoded User
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, user.Email, decoded.Email)
	assert.Equal(t, "", decoded.Password) // Password is empty in JSON
}

func TestUser_Fields(t *testing.T) {
	now := time.Now()
	user := User{
		ID:        123,
		Email:     "user@test.com",
		Name:      "Test Name",
		Password:  "secret",
		Role:      "member",
		TenantID:  456,
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, int64(123), user.ID)
	assert.Equal(t, "user@test.com", user.Email)
	assert.Equal(t, "Test Name", user.Name)
	assert.Equal(t, "secret", user.Password)
	assert.Equal(t, "member", user.Role)
	assert.Equal(t, int64(456), user.TenantID)
	assert.Equal(t, now, user.CreatedAt)
	assert.Equal(t, now, user.UpdatedAt)
}

func TestTeam_JSONSerialization(t *testing.T) {
	team := Team{
		ID:          1,
		Name:        "Test Team",
		Description: "A test team",
		OwnerID:     1,
		TenantID:    1,
		Settings:    map[string]string{"key": "value"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	data, err := json.Marshal(team)
	require.NoError(t, err)

	// Unmarshal back
	var decoded Team
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, team.Name, decoded.Name)
	assert.Equal(t, team.Settings, decoded.Settings)
}

func TestTeamMember_Fields(t *testing.T) {
	now := time.Now()
	member := TeamMember{
		ID:        1,
		TeamID:    10,
		UserID:    20,
		Role:      "manager",
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, int64(1), member.ID)
	assert.Equal(t, int64(10), member.TeamID)
	assert.Equal(t, int64(20), member.UserID)
	assert.Equal(t, "manager", member.Role)
}

func TestProvider_JSONSerialization(t *testing.T) {
	provider := Provider{
		ID:              1,
		Name:            "Test Provider",
		APIURL:          "https://api.example.com",
		APIKey:          "test_key",
		TeamID:          1,
		Kind:            "anthropic",
		Enabled:         true,
		ModelMapping:    map[string]string{"model": "claude-3"},
		SupportedModels: []string{"claude-3", "claude-2"},
		Level:           1,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	data, err := json.Marshal(provider)
	require.NoError(t, err)

	// Unmarshal back
	var decoded Provider
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, provider.Name, decoded.Name)
	assert.Equal(t, provider.APIKey, decoded.APIKey)
	assert.Equal(t, provider.Kind, decoded.Kind)
	assert.Equal(t, provider.Enabled, decoded.Enabled)
	assert.Len(t, decoded.SupportedModels, 2)
}

func TestProvider_Adapter(t *testing.T) {
	provider := &Provider{
		ID:              1,
		Name:            "Test Provider",
		APIURL:          "https://api.example.com",
		APIKey:          "test_key",
		Kind:            "anthropic",
		Enabled:         true,
		Level:           1,
		ModelMapping:    map[string]string{"model": "claude-3"},
		SupportedModels: []string{"claude-3", "claude-2"},
	}

	adapter := provider.GetProviderAdapter()

	assert.Equal(t, int64(1), adapter.GetID())
	assert.Equal(t, "Test Provider", adapter.GetName())
	assert.Equal(t, "https://api.example.com", adapter.GetAPIURL())
	assert.Equal(t, "test_key", adapter.GetAPIKey())
	assert.Equal(t, "anthropic", adapter.GetKind())
	assert.True(t, adapter.IsEnabled())
	assert.Equal(t, 1, adapter.GetLevel())
	assert.Equal(t, map[string]string{"model": "claude-3"}, adapter.GetModelMapping())
	assert.Equal(t, []string{"claude-3", "claude-2"}, adapter.GetSupportedModels())
}

func TestProvider_Adapter_NilFields(t *testing.T) {
	provider := &Provider{
		ID:      1,
		Name:    "Test",
		APIURL:  "https://api.example.com",
		APIKey:  "key",
		Kind:    "anthropic",
		Enabled: true,
		Level:   1,
		// ModelMapping and SupportedModels are nil
	}

	adapter := provider.GetProviderAdapter()

	// Should return empty collections instead of nil
	modelMapping := adapter.GetModelMapping()
	assert.NotNil(t, modelMapping)
	assert.Equal(t, map[string]string{}, modelMapping)

	supportedModels := adapter.GetSupportedModels()
	assert.NotNil(t, supportedModels)
	assert.Equal(t, []string{}, supportedModels)
}

func TestUsageRecord_Fields(t *testing.T) {
	now := time.Now()
	record := UsageRecord{
		ID:                1,
		Platform:          "anthropic",
		Model:             "claude-3-5-sonnet",
		Provider:          "provider1",
		HttpCode:          200,
		InputTokens:       100,
		OutputTokens:      50,
		CacheCreateTokens: 10,
		CacheReadTokens:   5,
		ReasoningTokens:   20,
		IsStream:          true,
		DurationSec:       1.5,
		TenantID:          1,
		UserID:            1,
		CreatedAt:         now,
	}

	assert.Equal(t, int64(1), record.ID)
	assert.Equal(t, "anthropic", record.Platform)
	assert.Equal(t, "claude-3-5-sonnet", record.Model)
	assert.Equal(t, 200, record.HttpCode)
	assert.Equal(t, 100, record.InputTokens)
	assert.Equal(t, 50, record.OutputTokens)
	assert.Equal(t, 10, record.CacheCreateTokens)
	assert.Equal(t, 5, record.CacheReadTokens)
	assert.Equal(t, 20, record.ReasoningTokens)
	assert.True(t, record.IsStream)
	assert.Equal(t, 1.5, record.DurationSec)
}

func TestUsageRecord_JSONSerialization(t *testing.T) {
	record := UsageRecord{
		ID:           1,
		Platform:     "anthropic",
		Model:        "claude-3",
		Provider:     "provider1",
		HttpCode:     200,
		InputTokens:  100,
		OutputTokens: 50,
		IsStream:     true,
		DurationSec:  1.5,
		TenantID:     1,
		UserID:       1,
		CreatedAt:    time.Now(),
	}

	data, err := json.Marshal(record)
	require.NoError(t, err)

	var decoded UsageRecord
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, record.Platform, decoded.Platform)
	assert.Equal(t, record.Model, decoded.Model)
	assert.Equal(t, record.InputTokens, decoded.InputTokens)
	assert.Equal(t, record.OutputTokens, decoded.OutputTokens)
}

func TestTeamUsageSummary_Fields(t *testing.T) {
	now := time.Now()
	summary := TeamUsageSummary{
		TeamID:      1,
		PeriodStart: now,
		PeriodEnd:   now.Add(24 * time.Hour),
		TotalInput:  1000,
		TotalOutput: 500,
		TotalCost:   0.05,
		CreatedAt:   now,
	}

	assert.Equal(t, int64(1), summary.TeamID)
	assert.Equal(t, 1000, summary.TotalInput)
	assert.Equal(t, 500, summary.TotalOutput)
	assert.Equal(t, 0.05, summary.TotalCost)
}

func TestTeamUsageSummary_JSONSerialization(t *testing.T) {
	now := time.Now()
	summary := TeamUsageSummary{
		TeamID:      1,
		PeriodStart: now,
		PeriodEnd:   now.Add(24 * time.Hour),
		TotalInput:  1000,
		TotalOutput: 500,
		TotalCost:   0.05,
		CreatedAt:   now,
	}

	data, err := json.Marshal(summary)
	require.NoError(t, err)

	var decoded TeamUsageSummary
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, summary.TeamID, decoded.TeamID)
	assert.Equal(t, summary.TotalInput, decoded.TotalInput)
	assert.Equal(t, summary.TotalOutput, decoded.TotalOutput)
}
