// Package support tests test utilities and context management for BDD tests
package support

import (
	"testing"
)

// TestLoadFixtureData tests loading fixture data from JSON files
func TestLoadFixtureData(t *testing.T) {
	// Test 1: Load from testdata/fixtures
	t.Run("loads from testdata/fixtures", func(t *testing.T) {
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

		// Verify admin user exists by key
		adminUser, exists := fixtures.Users["admin"]
		if !exists {
			t.Error("expected to find admin user by key 'admin'")
		} else if adminUser.Role != "admin" {
			t.Errorf("expected admin role to be 'admin', got '%s'", adminUser.Role)
		}

		// Verify member user exists by key
		memberUser, exists := fixtures.Users["member"]
		if !exists {
			t.Error("expected to find member user by key 'member'")
		} else if memberUser.Role != "member" {
			t.Errorf("expected member role to be 'member', got '%s'", memberUser.Role)
		}

		// Verify claude provider exists by key
		claudeProvider, exists := fixtures.Providers["claude"]
		if !exists {
			t.Error("expected to find claude provider by key 'claude'")
		} else if claudeProvider.Kind != "claude" {
			t.Errorf("expected provider kind 'claude', got '%s'", claudeProvider.Kind)
		}
	})

	// Test 2: Teams and licenses are optional
	t.Run("teams and licenses are optional", func(t *testing.T) {
		fixtures, err := LoadFixtureData()
		if err != nil {
			t.Fatalf("failed to load fixture data: %v", err)
		}

		// Teams and licenses may or may not exist depending on fixture files
		_ = fixtures.Teams
		_ = fixtures.Licenses
	})

	// Test 3: LoadLicenseFixture loads specific license by tier
	t.Run("LoadLicenseFixture loads by tier", func(t *testing.T) {
		fixtures, err := LoadFixtureData()
		if err != nil {
			t.Fatalf("failed to load fixture data: %v", err)
		}

		// Skip if no licenses in fixtures
		if len(fixtures.Licenses) == 0 {
			t.Skip("no licenses in fixtures")
			return
		}

		// Test loading trial license
		trialLicense, exists := fixtures.Licenses["trial"]
		if !exists {
			t.Error("expected to find trial license by key 'trial'")
		} else if trialLicense.Tier != "trial" {
			t.Errorf("expected tier 'trial', got '%s'", trialLicense.Tier)
		}

		// Test loading professional license
		profLicense, exists := fixtures.Licenses["professional"]
		if !exists {
			t.Error("expected to find professional license by key 'professional'")
		} else if profLicense.Tier != "professional" {
			t.Errorf("expected tier 'professional', got '%s'", profLicense.Tier)
		}
	})
}
