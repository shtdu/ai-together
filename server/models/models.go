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
	"time"

	"github.com/code-together/shared/provider"
)

// User represents a user in the system with authentication and authorization properties
// Each user belongs to a tenant and has a role (manager or member) that determines permissions
type User struct {
	ID        int64     `json:"id" db:"id"`                 // Unique identifier for the user
	Email     string    `json:"email" db:"email"`           // User's email address, used for login
	Name      string    `json:"name" db:"name"`             // User's display name
	Password  string    `json:"-" db:"password"`            // Hashed password (not exposed in JSON responses)
	Role      string    `json:"role" db:"role"`             // User's role: "manager" or "member"
	TenantID  int64     `json:"tenant_id" db:"tenant_id"`   // ID of the tenant the user belongs to
	CreatedAt time.Time `json:"created_at" db:"created_at"` // Timestamp when the user was created
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // Timestamp when the user was last updated
}

// Team represents a team entity in the multi-tenant system
// Teams contain members and have configurations specific to their tenant
type Team struct {
	ID          int64             `json:"id" db:"id"`                   // Unique identifier for the team
	Name        string            `json:"name" db:"name"`               // Name of the team
	Description string            `json:"description" db:"description"` // Optional description of the team
	OwnerID     int64             `json:"owner_id" db:"owner_id"`       // ID of the user who owns/created the team
	TenantID    int64             `json:"tenant_id" db:"tenant_id"`     // ID of the tenant this team belongs to
	Settings    map[string]string `json:"settings" db:"settings"`       // JSON object containing team-specific settings
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`   // Timestamp when the team was created
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`   // Timestamp when the team was last updated
}

// TeamMember represents the relationship between a user and a team
// This enables many-to-many relationships between users and teams
type TeamMember struct {
	ID        int64     `json:"id" db:"id"`                 // Unique identifier for the team member relationship
	TeamID    int64     `json:"team_id" db:"team_id"`       // ID of the team the user belongs to
	UserID    int64     `json:"user_id" db:"user_id"`       // ID of the user in the team
	Role      string    `json:"role" db:"role"`             // Role of the user in this team: "manager" or "member"
	CreatedAt time.Time `json:"created_at" db:"created_at"` // Timestamp when the user joined the team
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // Timestamp when the relationship was last updated
}

// Provider represents an AI service provider configuration
// Each provider is owned by a team and can be used by team members
type Provider struct {
	ID              int64             `json:"id" db:"id"`                             // Unique identifier for the provider
	Name            string            `json:"name" db:"name"`                         // Display name for the provider
	APIURL          string            `json:"api_url" db:"api_url"`                   // Base URL for the provider's API
	APIKey          string            `json:"api_key" db:"api_key"`                   // Authentication key for the provider API
	TeamID          int64             `json:"team_id" db:"team_id"`                   // ID of the team that owns this provider
	Kind            string            `json:"kind" db:"kind"`                         // Provider kind: "claude", "codex", or "opencode"
	Enabled         bool              `json:"enabled" db:"enabled"`                   // Whether this provider is currently enabled
	ModelMapping    map[string]string `json:"model_mapping" db:"model_mapping"`       // Mapping of local model names to provider model names
	SupportedModels []string          `json:"supported_models" db:"supported_models"` // Models supported by this provider
	Level           int               `json:"level" db:"level"`                       // Priority level for provider selection
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`             // Timestamp when the provider was added
	UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`             // Timestamp when the provider was last updated
}

// UsageRecord represents a single usage event for an AI service request
// This is used for tracking, analytics, and billing purposes
type UsageRecord struct {
	ID                int64     `json:"id" db:"id"`                                   // Unique identifier for the usage record
	Platform          string    `json:"platform" db:"platform"`                       // The platform that handled the request (e.g. "openai", "anthropic")
	Model             string    `json:"model" db:"model"`                             // The model that was used for the request
	Provider          string    `json:"provider" db:"provider"`                       // The provider that fulfilled the request
	HttpCode          int       `json:"http_code" db:"http_code"`                     // HTTP status code returned by the provider
	InputTokens       int       `json:"input_tokens" db:"input_tokens"`               // Number of tokens in the input to the model
	OutputTokens      int       `json:"output_tokens" db:"output_tokens"`             // Number of tokens in the output from the model
	CacheCreateTokens int       `json:"cache_create_tokens" db:"cache_create_tokens"` // Number of cache creation tokens (for models that support caching)
	CacheReadTokens   int       `json:"cache_read_tokens" db:"cache_read_tokens"`     // Number of cache read tokens (for models that support caching)
	ReasoningTokens   int       `json:"reasoning_tokens" db:"reasoning_tokens"`       // Number of tokens used for reasoning (for models that support reasoning)
	IsStream          bool      `json:"is_stream" db:"is_stream"`                     // Whether the response was streamed
	DurationSec       float64   `json:"duration_sec" db:"duration_sec"`               // Duration of the request in seconds
	TenantID          int64     `json:"tenant_id" db:"tenant_id"`                     // ID of the tenant associated with this usage
	UserID            int64     `json:"user_id" db:"user_id"`                         // ID of the user who made this request
	CreatedAt         time.Time `json:"created_at" db:"created_at"`                   // Timestamp when the request was made
}

// TeamUsageSummary provides aggregated usage statistics for a team
// This enables efficient retrieval of usage metrics without complex queries
type TeamUsageSummary struct {
	TeamID      int64     `json:"team_id" db:"team_id"`           // ID of the team this summary is for
	PeriodStart time.Time `json:"period_start" db:"period_start"` // Start of the aggregation period
	PeriodEnd   time.Time `json:"period_end" db:"period_end"`     // End of the aggregation period
	TotalInput  int       `json:"total_input" db:"total_input"`   // Total input tokens used during the period
	TotalOutput int       `json:"total_output" db:"total_output"` // Total output tokens generated during the period
	TotalCost   float64   `json:"total_cost" db:"total_cost"`     // Total cost of usage during the period
	CreatedAt   time.Time `json:"created_at" db:"created_at"`     // When this summary was created
}

// GetProviderAdapter converts a Provider to implement the shared.Provider interface
func (p *Provider) GetProviderAdapter() provider.Provider {
	return &providerAdapter{p: p}
}

// GetEffectiveModel gets the actual model name that should be used
// This is a convenience method that delegates to the shared provider package
func (p *Provider) GetEffectiveModel(sharedProvider provider.Provider, requestedModel string) string {
	return provider.GetEffectiveModel(sharedProvider, requestedModel)
}

// providerAdapter implements the shared.Provider interface for models.Provider
type providerAdapter struct {
	p *Provider
}

func (a *providerAdapter) GetID() int64      { return a.p.ID }
func (a *providerAdapter) GetName() string   { return a.p.Name }
func (a *providerAdapter) GetAPIURL() string { return a.p.APIURL }
func (a *providerAdapter) GetAPIKey() string { return a.p.APIKey }
func (a *providerAdapter) GetKind() string   { return a.p.Kind }
func (a *providerAdapter) IsEnabled() bool   { return a.p.Enabled }
func (a *providerAdapter) GetLevel() int     { return a.p.Level }

func (a *providerAdapter) GetModelMapping() map[string]string {
	if a.p.ModelMapping == nil {
		return map[string]string{}
	}
	return a.p.ModelMapping
}

func (a *providerAdapter) GetSupportedModels() []string {
	if a.p.SupportedModels == nil {
		return []string{}
	}
	return a.p.SupportedModels
}
