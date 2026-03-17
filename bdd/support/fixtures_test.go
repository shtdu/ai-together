// Package support tests test utilities and context management for BDD tests
package support

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadFixtureData tests loading fixture data from JSON files
func TestLoadFixtureData(t *testing.T) {
	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	// Change to bdd/support directory
	if err := os.Chdir(filepath.Join("..", "support")); err != nil {
		t.Fatalf("failed to change to support directory: %v", err)
	}

	// Test 1: Load from local fixtures.json
	t.Run("loads from local fixtures.json", func(t *testing.T) {
		fixtures, err := LoadFixtureData()
		if err != nil {
			t.Fatalf("failed to load fixture data: %v", err)
		}

		// Verify we have expected data
		if len(fixtures.Users) == 0 {
			t.Error("expected at least one user, got none")
		}
		if len(fixtures.Providers) == 0 {
			t.Error("expected at least one provider, got none")
		}
		if len(fixtures.Teams) == 0 {
			t.Error("expected at least one team, got none")
		}
		if len(fixtures.Licenses) == 0 {
			t.Error("expected at least one license, got none")
		}

		// Verify admin user exists
		var adminUser *UserFixture
		for _, user := range fixtures.Users {
			if user.Role == "admin" {
				adminUser = &user
				break
			}
		}
		if adminUser == nil {
			t.Error("expected to find admin user")
		}

		// Verify claude provider exists
		var claudeProvider *ProviderFixture
		for _, provider := range fixtures.Providers {
			if provider.Kind == "claude" {
				claudeProvider = &provider
				break
			}
		}
		if claudeProvider == nil {
			t.Error("expected to find claude provider")
		}
	})

	// Test 2: GetStandardFixtureData returns error if file missing
	t.Run("GetStandardFixtureData requires fixtures.json", func(t *testing.T) {
		// Rename fixtures.json temporarily
		originalPath := filepath.Join(".", "fixtures.json")
		tempPath := filepath.Join(".", "fixtures.json.bak")

		if err := os.Rename(originalPath, tempPath); err != nil {
			t.Fatalf("failed to rename fixtures.json: %v", err)
		}
		t.Cleanup(func() {
			if err := os.Rename(tempPath, originalPath); err != nil {
				t.Logf("Warning: failed to restore fixtures.json: %v", err)
			}
		})

		// Should return error
		_, err := GetStandardFixtureData()
		if err == nil {
			t.Error("expected error when fixtures.json is missing, got nil")
		}
	})

	// Test 2.5: LoadFixtureData falls back to integration testdata
	t.Run("falls back to integration testdata when local missing", func(t *testing.T) {
		// Temporarily rename local fixtures.json
		originalPath := filepath.Join(".", "fixtures.json")
		tempPath := filepath.Join(".", "fixtures.json.bak")

		if err := os.Rename(originalPath, tempPath); err != nil {
			t.Fatalf("failed to rename fixtures.json: %v", err)
		}
		t.Cleanup(func() {
			if err := os.Rename(tempPath, originalPath); err != nil {
				t.Logf("Warning: failed to restore fixtures.json: %v", err)
			}
		})

		// Should load from integration testdata
		fixtures, err := LoadFixtureData()
		if err != nil {
			// Expected if integration testdata doesn't exist yet
			t.Skipf("integration testdata fallback not available yet: %v", err)
			return
		}

		// Verify data loaded
		if len(fixtures.Users) == 0 {
			t.Error("expected to load users from integration testdata")
		}
	})

	// Test 3: GetStandardFixtureData loads successfully
	t.Run("GetStandardFixtureData loads successfully", func(t *testing.T) {
		fixtures, err := GetStandardFixtureData()
		if err != nil {
			t.Fatalf("failed to load standard fixture data: %v", err)
		}

		// Verify we have expected data
		if len(fixtures.Users) == 0 {
			t.Error("expected at least one user, got none")
		}
		if len(fixtures.Licenses) == 0 {
			t.Error("expected at least one license, got none")
		}
	})

	// Test 4: LoadLicenseFixture loads specific license
	t.Run("LoadLicenseFixture loads by tier", func(t *testing.T) {
		// Test loading trial license
		trialLicense, err := LoadLicenseFixture("trial")
		if err != nil {
			t.Fatalf("failed to load trial license: %v", err)
		}
		if trialLicense.Tier != "trial" {
			t.Errorf("expected tier 'trial', got '%s'", trialLicense.Tier)
		}

		// Test loading professional license
		profLicense, err := LoadLicenseFixture("professional")
		if err != nil {
			t.Fatalf("failed to load professional license: %v", err)
		}
		if profLicense.Tier != "professional" {
			t.Errorf("expected tier 'professional', got '%s'", profLicense.Tier)
		}

		// Test loading enterprise license
		entLicense, err := LoadLicenseFixture("enterprise")
		if err != nil {
			t.Fatalf("failed to load enterprise license: %v", err)
		}
		if entLicense.Tier != "enterprise" {
			t.Errorf("expected tier 'enterprise', got '%s'", entLicense.Tier)
		}

		// Test loading non-existent license
		_, err = LoadLicenseFixture("nonexistent")
		if err == nil {
			t.Error("expected error for non-existent license tier, got nil")
		}
	})
}
