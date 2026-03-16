// Package support provides test utilities and context management for BDD tests
package support

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// FixtureData holds test fixture data loaded from integration/testdata
type FixtureData struct {
	Users     []UserFixture     `json:"users"`
	Providers []ProviderFixture `json:"providers"`
	Teams     []TeamFixture     `json:"teams"`
	Licenses  []LicenseFixture  `json:"licenses"`
}

// UserFixture represents a user test fixture
type UserFixture struct {
	ID             string `json:"id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	Password       string `json:"password"`
	Role           string `json:"role"`
	TenantID       int64  `json:"tenant_id"`
	ProviderUserID string `json:"provider_user_id,omitempty"`
}

// ProviderFixture represents a provider test fixture
type ProviderFixture struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	APIKey    string `json:"api_key"`
	Priority  int    `json:"priority"`
	Enabled   bool   `json:"enabled"`
	TenantID  int64  `json:"tenant_id"`
	CreatedAt string `json:"created_at,omitempty"`
}

// TeamFixture represents a team test fixture
type TeamFixture struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	TenantID    int64  `json:"tenant_id"`
	CreatedAt   string `json:"created_at,omitempty"`
}

// LicenseFixture represents a license test fixture
type LicenseFixture struct {
	ID        string `json:"id"`
	Key       string `json:"key"`
	Tier      string `json:"tier"`
	ExpiresAt string `json:"expires_at,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// LoadFixtureData loads test fixture data from the integration testdata directory
// The fixture path is relative to the bdd directory
func LoadFixtureData() (*FixtureData, error) {
	// Navigate from bdd/ to ../integration/testdata/
	fixturePath := filepath.Join("..", "integration", "testdata", "fixtures.json")

	// Check if file exists
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("fixture file not found: %s", fixturePath)
	}

	// Read file
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixture file: %w", err)
	}

	// Parse JSON
	var fixtures FixtureData
	if err := json.Unmarshal(data, &fixtures); err != nil {
		return nil, fmt.Errorf("failed to parse fixture JSON: %w", err)
	}

	return &fixtures, nil
}

// LoadLicenseFixture loads a specific license fixture by tier
func LoadLicenseFixture(tier string) (*LicenseFixture, error) {
	fixtures, err := LoadFixtureData()
	if err != nil {
		return nil, err
	}

	for _, license := range fixtures.Licenses {
		if license.Tier == tier {
			return &license, nil
		}
	}

	return nil, fmt.Errorf("license fixture with tier '%s' not found", tier)
}

// GetStandardFixtureData returns a standard set of fixture data for testing
// This creates basic test data if the fixtures.json file is not available
func GetStandardFixtureData() *FixtureData {
	return &FixtureData{
		Users: []UserFixture{
			{
				ID:       "test-admin-1",
				Email:    "admin@example.com",
				Name:     "Test Admin",
				Password: "TestPassword123!",
				Role:     "admin",
				TenantID: 1,
			},
			{
				ID:       "test-member-1",
				Email:    "member@example.com",
				Name:     "Test Member",
				Password: "TestPassword123!",
				Role:     "member",
				TenantID: 1,
			},
		},
		Providers: []ProviderFixture{
			{
				ID:       1,
				Name:     "test-claude-provider",
				Kind:     "claude",
				APIKey:   "sk-test-claude-123",
				Priority: 1,
				Enabled:  true,
				TenantID: 1,
			},
		},
		Teams: []TeamFixture{
			{
				ID:          1,
				Name:        "Default Team",
				Description: "Default team for testing",
				TenantID:    1,
			},
		},
		Licenses: []LicenseFixture{
			{
				ID:        "license-1",
				Key:       "test-license-key",
				Tier:      "trial",
				ExpiresAt: "2025-12-31T23:59:59Z",
			},
		},
	}
}
