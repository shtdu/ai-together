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

// LoadFixtureData loads test fixture data from JSON files
// Priority: 1) local fixtures.json, 2) integration testdata, 3) error
// Note: Falls back to integration testdata for compatibility during transition
func LoadFixtureData() (*FixtureData, error) {
	// First, try local fixtures.json (in bdd/support/)
	localFixturePath := filepath.Join(".", "fixtures.json")

	if data, err := os.ReadFile(localFixturePath); err == nil {
		var fixtures FixtureData
		if err := json.Unmarshal(data, &fixtures); err != nil {
			return nil, fmt.Errorf("failed to parse local fixture JSON: %w", err)
		}
		return &fixtures, nil
	}

	// Fall back to integration testdata (for compatibility)
	integrationFixturePath := filepath.Join("..", "..", "integration", "testdata", "fixtures.json")

	if _, err := os.Stat(integrationFixturePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no fixtures file found: tried %s and %s", localFixturePath, integrationFixturePath)
	}

	data, err := os.ReadFile(integrationFixturePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read integration fixture file: %w", err)
	}

	var fixtures FixtureData
	if err := json.Unmarshal(data, &fixtures); err != nil {
		return nil, fmt.Errorf("failed to parse integration fixture JSON: %w", err)
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

// GetStandardFixtureData returns fixture data from fixtures.json
// Returns error if file doesn't exist (no hardcoded fallback)
func GetStandardFixtureData() (*FixtureData, error) {
	localFixturePath := filepath.Join(".", "fixtures.json")

	if _, err := os.Stat(localFixturePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("fixtures.json not found in %s", localFixturePath)
	}

	data, err := os.ReadFile(localFixturePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixtures.json: %w", err)
	}

	var fixtures FixtureData
	if err := json.Unmarshal(data, &fixtures); err != nil {
		return nil, fmt.Errorf("failed to parse fixtures.json: %w", err)
	}

	return &fixtures, nil
}
