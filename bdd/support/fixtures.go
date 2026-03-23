// Package support provides test utilities and context management for BDD tests
package support

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// FixtureData holds test fixture data loaded from testdata/fixtures
type FixtureData struct {
	Users     map[string]UserFixture     `json:"users"`
	Providers map[string]ProviderFixture `json:"providers"`
	Teams     map[string]TeamFixture     `json:"teams"`
	Licenses  map[string]LicenseFixture  `json:"licenses"`
}

// UserFixture represents a user test fixture (matches integration format)
// BDD extension: includes Role for convenience (not in integration JSON)
type UserFixture struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Role     string `json:"role,omitempty"` // Optional: defaults inferred from fixture key
}

// ProviderFixture represents a provider test fixture (matches integration format)
type ProviderFixture struct {
	Name            string   `json:"name"`
	Kind            string   `json:"kind"`
	APIKey          string   `json:"api_key"`
	APIURL          string   `json:"api_url"`
	Enabled         bool     `json:"enabled"`
	Level           int      `json:"level"`
	SupportedModels []string `json:"supported_models"`
}

// TeamFixture represents a team test fixture
type TeamFixture struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	TenantID    int64  `json:"tenant_id"`
}

// LicenseFixture represents a license test fixture
type LicenseFixture struct {
	ID        string `json:"id"`
	Key       string `json:"key"`
	Tier      string `json:"tier"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

// LoadFixtureData loads all fixture data from JSON files in testdata/fixtures
// Matches integration test format: separate files for users, providers, teams, licenses
func LoadFixtureData() (*FixtureData, error) {
	fixtures := &FixtureData{
		Users:     make(map[string]UserFixture),
		Providers: make(map[string]ProviderFixture),
		Teams:     make(map[string]TeamFixture),
		Licenses:  make(map[string]LicenseFixture),
	}

	// Get the directory of this file (support package)
	_, currentFilePath, _, _ := runtime.Caller(0)
	supportDir := filepath.Dir(currentFilePath)
	fixturesDir := filepath.Join(supportDir, "..", "testdata", "fixtures")

	// Load users fixture
	usersFile := filepath.Join(fixturesDir, "users.json")
	if data, err := os.ReadFile(usersFile); err == nil {
		if err := json.Unmarshal(data, &fixtures.Users); err != nil {
			return nil, fmt.Errorf("failed to parse users.json: %w", err)
		}
		// Infer roles from fixture key names if not explicitly set
		for key, user := range fixtures.Users {
			if user.Role == "" {
				user.Role = key // "admin" -> Role: "admin", "member" -> Role: "member"
				fixtures.Users[key] = user
			}
		}
	} else {
		return nil, fmt.Errorf("failed to read users.json: %w", err)
	}

	// Load providers fixture
	providersFile := filepath.Join(fixturesDir, "providers.json")
	if data, err := os.ReadFile(providersFile); err == nil {
		if err := json.Unmarshal(data, &fixtures.Providers); err != nil {
			return nil, fmt.Errorf("failed to parse providers.json: %w", err)
		}
	} else {
		return nil, fmt.Errorf("failed to read providers.json: %w", err)
	}

	// Load teams fixture (optional)
	teamsFile := filepath.Join(fixturesDir, "teams.json")
	if data, err := os.ReadFile(teamsFile); err == nil {
		if err := json.Unmarshal(data, &fixtures.Teams); err != nil {
			return nil, fmt.Errorf("failed to parse teams.json: %w", err)
		}
	}

	// Load licenses fixture (optional)
	licensesFile := filepath.Join(fixturesDir, "licenses.json")
	if data, err := os.ReadFile(licensesFile); err == nil {
		if err := json.Unmarshal(data, &fixtures.Licenses); err != nil {
			return nil, fmt.Errorf("failed to parse licenses.json: %w", err)
		}
	}

	return fixtures, nil
}

// LoadLicenseFixture loads a specific license fixture by tier key
func LoadLicenseFixture(tier string) (*LicenseFixture, error) {
	fixtures, err := LoadFixtureData()
	if err != nil {
		return nil, err
	}

	license, exists := fixtures.Licenses[tier]
	if !exists {
		return nil, fmt.Errorf("license fixture with tier '%s' not found", tier)
	}

	return &license, nil
}

// LoadLicensePEM loads a license PEM file from testdata/licenses
func LoadLicensePEM(fixtureName string) (string, error) {
	// Get the directory of this file (support package)
	_, currentFilePath, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}
	supportDir := filepath.Dir(currentFilePath)
	// Navigate from bdd/support/ to bdd/testdata/licenses/
	licensePath := filepath.Join(supportDir, "..", "testdata", "licenses", fixtureName+".pem")

	pemContent, err := os.ReadFile(licensePath)
	if err != nil {
		return "", fmt.Errorf("failed to read license fixture %s: %w", licensePath, err)
	}
	return string(pemContent), nil
}
