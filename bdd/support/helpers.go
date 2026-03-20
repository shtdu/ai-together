// Package support provides test utilities and context management for BDD tests
package support

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

// GenerateUniqueProviderName creates a unique provider name using timestamp
func GenerateUniqueProviderName(base string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s-%d", base, timestamp)
}

// GenerateUniqueEmail creates a unique email address using UUID
func GenerateUniqueEmail(base string) string {
	uniqueID := uuid.New().String()[:8]
	return fmt.Sprintf("%s-%s@example.com", base, uniqueID)
}

// GenerateUniqueTeamName creates a unique team name using timestamp
func GenerateUniqueTeamName(base string) string {
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s-%s", base, timestamp)
}

// GenerateUniqueLicenseKey creates a unique license key for testing
func GenerateUniqueLicenseKey() string {
	return fmt.Sprintf("test-license-%s", uuid.New().String())
}

// GenerateUniqueID creates a unique integer ID for testing
func GenerateUniqueID() int64 {
	return time.Now().UnixNano()
}

// Test data constants for consistent test values
const (
	TestAPIKeyClaude   = "sk-test-claude-123"
	TestAPIKeyCodex    = "sk-test-codex-123"
	TestAPIKeyOpenCode = "sk-test-opencode-123"

	TestPasswordStrong = "TestPassword123!"
	TestPasswordWeak   = "weak"

	ProviderKindClaude   = "claude"
	ProviderKindCodex    = "codex"
	ProviderKindOpenCode = "opencode"

	RoleAdmin  = "admin"
	RoleMember = "member"

	TierTrial        = "trial"
	TierStarter      = "starter"
	TierProfessional = "professional"
	TierEnterprise   = "enterprise"
)

// IsValidProviderKind checks if a provider kind is valid
func IsValidProviderKind(kind string) bool {
	switch kind {
	case ProviderKindClaude, ProviderKindCodex, ProviderKindOpenCode:
		return true
	default:
		return false
	}
}

// IsValidRole checks if a role is valid
func IsValidRole(role string) bool {
	switch role {
	case RoleAdmin, RoleMember:
		return true
	default:
		return false
	}
}

// IsValidTier checks if a license tier is valid
func IsValidTier(tier string) bool {
	switch tier {
	case TierTrial, TierStarter, TierProfessional, TierEnterprise:
		return true
	default:
		return false
	}
}

// GetTestServerURL returns the default test server URL
// Can be overridden via environment variable
func GetTestServerURL() string {
	// Check for environment variable
	if url := getEnv("TEST_SERVER_URL", ""); url != "" {
		return url
	}
	return "http://localhost:8088"
}

// GetTestDatabaseURL returns the default test database URL
// Can be overridden via environment variable
func GetTestDatabaseURL() string {
	if url := getEnv("TEST_DATABASE_URL", ""); url != "" {
		return url
	}
	return "postgres://test:test@localhost:5432/codetogether_test?sslmode=disable"
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
