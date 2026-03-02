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


package services

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	license "github.com/vitalvas/go-license/license"
	"switch-server/models"
	"switch-server/repository"
)

const (
	// DefaultLicenseType is the fallback license type when no valid license is present
	DefaultLicenseType = "opensource"
	// DefaultMaxSeats is the fallback seat limit (unlimited/soft limit)
	DefaultMaxSeats = -1
	// DefaultMaxTeams is the fallback team limit (1 for opensource)
	DefaultMaxTeams = 1
	// DefaultDataRetentionDays is the fallback data retention period (7 days for opensource)
	DefaultDataRetentionDays = 7
)

type LicenseService struct {
	licenseRepo   repository.LicenseRepositoryInterface
	publicKey     ed25519.PublicKey
	licenses      map[int64]*models.License // Pre-loaded cache of decoded licenses
	licensesMutex sync.RWMutex              // Thread-safe access to licenses cache
}

func NewLicenseService(licenseRepo repository.LicenseRepositoryInterface, publicKeyBase64 string) (*LicenseService, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}
	if len(keyBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: expected %d, got %d", ed25519.PublicKeySize, len(keyBytes))
	}

	service := &LicenseService{
		licenseRepo: licenseRepo,
		publicKey:   ed25519.PublicKey(keyBytes),
		licenses:    make(map[int64]*models.License),
	}

	// Pre-load all licenses on initialization
	if err := service.loadAllLicenses(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to load licenses: %w", err)
	}

	return service, nil
}

// loadAllLicenses loads all tenant licenses from database, decodes and validates them
// Called once during service initialization
func (s *LicenseService) loadAllLicenses(ctx context.Context) error {
	// Load all licenses from database
	allLicenses, err := s.licenseRepo.GetAllLicenses(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve licenses: %w", err)
	}

	// Decode and validate each license
	loadedCount := 0
	for _, storedLicense := range allLicenses {
		if storedLicense.LicenseKey == "" {
			// No license key for this tenant - skip
			continue
		}

		// Decode the license key
		decoded, err := license.Decode([]byte(storedLicense.LicenseKey), s.publicKey)
		if err != nil {
			// Invalid license key - log and skip
			fmt.Printf("WARN: Invalid license key for tenant %d: %v\n", storedLicense.TenantID, err)
			continue
		}

		// Check if expired - still cache it but GetEffectiveLicense will filter it out
		if decoded.Expired() {
			fmt.Printf("WARN: Expired license for tenant %d\n", storedLicense.TenantID)
		}

		// Extract tier and seats from decoded data (source of truth)
		var licenseData models.LicenseData
		if err := json.Unmarshal(decoded.Data, &licenseData); err != nil {
			fmt.Printf("WARN: Invalid license data for tenant %d: %v\n", storedLicense.TenantID, err)
			continue
		}

		// Create license object with DECODED values (not DB values)
		license := &models.License{
			TenantID:          storedLicense.TenantID,
			CustomerName:      storedLicense.CustomerName,
			LicenseID:         decoded.ID,
			LicenseType:       licenseData.LicenseType,
			MaxSeats:          licenseData.MaxSeats,
			MaxTeams:          licenseData.MaxTeams,
			DataRetentionDays: licenseData.DataRetentionDays,
			LicenseKey:        storedLicense.LicenseKey,
			IssuedAt:          time.Unix(decoded.IssuedAt, 0),
			ExpiresAt:         time.Unix(decoded.ExpiredAt, 0),
		}

		// Store in cache
		s.licensesMutex.Lock()
		s.licenses[storedLicense.TenantID] = license
		s.licensesMutex.Unlock()

		loadedCount++
	}

	fmt.Printf("INFO: License loading complete: loaded %d licenses for %d tenants\n", loadedCount, len(s.licenses))
	return nil
}

// GetLicense retrieves the license for a tenant
// NOTE: This queries the database directly. For cached license access, use GetEffectiveLicense().
// GetLicense() should only be used when you need the raw database state (e.g., for integrity verification)
func (s *LicenseService) GetLicense(ctx context.Context, tenantID int64) (*models.License, error) {
	return s.licenseRepo.GetTenantLicense(ctx, tenantID)
}

// GetLicenseUsage retrieves the current usage against license limits
// Uses cached license for performance
func (s *LicenseService) GetLicenseUsage(ctx context.Context, tenantID int64) (*models.LicenseUsage, error) {
	license := s.GetEffectiveLicense(ctx, tenantID)

	currentUsers, err := s.licenseRepo.CountTenantUsers(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	providerCounts, err := s.licenseRepo.GetProviderCountsByKind(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Calculate seats remaining (soft limit - for display/analytics only)
	maxSeats := license.GetMaxSeats()
	seatsRemaining := -1 // -1 indicates unlimited
	if maxSeats != -1 {
		seatsRemaining = maxSeats - currentUsers
		if seatsRemaining < 0 {
			seatsRemaining = 0
		}
	}

	// Get team count for team limit checking (hard enforced)
	currentTeams, err := s.licenseRepo.CountTenantTeams(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	teamsRemaining := -1 // unlimited by default
	maxTeams := license.GetMaxTeams()
	if maxTeams != -1 {
		teamsRemaining = maxTeams - currentTeams
		if teamsRemaining < 0 {
			teamsRemaining = 0
		}
	}

	return &models.LicenseUsage{
		License:          *license,
		CurrentUsers:     currentUsers,
		SeatsRemaining:   seatsRemaining, // Deprecated but kept for backward compatibility
		CurrentTeams:     currentTeams,
		TeamsRemaining:   teamsRemaining,
		ProviderCounts:   providerCounts,
	}, nil
}

// CanCreateTeam checks if a new team can be created based on license type
// Open Source: max 1 team
// Commercial: unlimited teams
func (s *LicenseService) CanCreateTeam(ctx context.Context, tenantID int64) (bool, error) {
	license := s.GetEffectiveLicense(ctx, tenantID)

	maxTeams := license.GetMaxTeams()
	if maxTeams == -1 {
		// Unlimited teams
		return true, nil
	}

	// Check current team count
	currentTeams, err := s.licenseRepo.CountTenantTeams(ctx, tenantID)
	if err != nil {
		return false, err
	}

	return currentTeams < maxTeams, nil
}

// IsLicenseValid checks if the tenant has a valid (non-expired) license
// Uses cached license for performance
func (s *LicenseService) IsLicenseValid(ctx context.Context, tenantID int64) (bool, string, error) {
	license := s.GetEffectiveLicense(ctx, tenantID)

	if !license.IsValid() {
		return false, "No valid license found", nil
	}

	if license.IsExpired() {
		return false, "License has expired", nil
	}

	return true, "", nil
}

// ActivateLicense decodes and validates a license key, then stores it
func (s *LicenseService) ActivateLicense(ctx context.Context, tenantID int64, licenseKeyPEM string) error {
	// Decode license using go-license with public key
	decoded, err := license.Decode([]byte(licenseKeyPEM), s.publicKey)
	if err != nil {
		return fmt.Errorf("invalid license key: %w", err)
	}

	// Check expiration
	if decoded.Expired() {
		return fmt.Errorf("license has expired")
	}

	// Extract custom data (license_type, max_seats, max_teams, data_retention_days)
	var licenseData models.LicenseData
	if err := json.Unmarshal(decoded.Data, &licenseData); err != nil {
		return fmt.Errorf("invalid license data: %w", err)
	}

	// Validate license type
	if licenseData.LicenseType == "" {
		return fmt.Errorf("license type is required")
	}
	if licenseData.LicenseType != "opensource" && licenseData.LicenseType != "commercial" {
		return fmt.Errorf("license type must be 'opensource' or 'commercial'")
	}
	// MaxSeats is informational (soft limit), default to -1 (unlimited) if not set
	if licenseData.MaxSeats == 0 {
		licenseData.MaxSeats = -1
	}
	if licenseData.MaxTeams == 0 {
		licenseData.MaxTeams = 1 // Default to 1 team
	}
	if licenseData.DataRetentionDays == 0 {
		if licenseData.LicenseType == "opensource" {
			licenseData.DataRetentionDays = 7
		} else {
			licenseData.DataRetentionDays = 90
		}
	}

	// Create license object
	license := &models.License{
		TenantID:           tenantID,
		LicenseID:          decoded.ID,
		LicenseType:        licenseData.LicenseType,
		MaxSeats:           licenseData.MaxSeats,
		MaxTeams:           licenseData.MaxTeams,
		DataRetentionDays:  licenseData.DataRetentionDays,
		LicenseKey:         licenseKeyPEM,
		IssuedAt:           time.Unix(decoded.IssuedAt, 0),
		ExpiresAt:          time.Unix(decoded.ExpiredAt, 0),
	}

	// Store decoded values in database
	if err := s.licenseRepo.UpdateTenantLicense(ctx, tenantID, license); err != nil {
		return err
	}

	// Update cache immediately
	s.licensesMutex.Lock()
	defer s.licensesMutex.Unlock()
	s.licenses[tenantID] = license

	fmt.Printf("INFO: License activated for tenant %d: type=%s, max_teams=%d\n", tenantID, license.LicenseType, license.MaxTeams)
	return nil
}

// GetEffectiveLicense returns the cached license or default fallback
// NEVER returns nil - always returns a valid license object
func (s *LicenseService) GetEffectiveLicense(ctx context.Context, tenantID int64) *models.License {
	// Read from cache (concurrent-safe)
	s.licensesMutex.RLock()
	license, found := s.licenses[tenantID]
	s.licensesMutex.RUnlock()

	if found && license.IsValid() && !license.IsExpired() {
		return license
	}

	// Return default license (no DB query needed)
	return &models.License{
		TenantID:           tenantID,
		LicenseType:        DefaultLicenseType,
		MaxSeats:           DefaultMaxSeats,
		MaxTeams:           DefaultMaxTeams,
		DataRetentionDays:  DefaultDataRetentionDays,
	}
}

// HasActiveLicense checks if tenant has a valid (non-default) license
// Uses cached license for performance
func (s *LicenseService) HasActiveLicense(ctx context.Context, tenantID int64) bool {
	license := s.GetEffectiveLicense(ctx, tenantID)
	// Check if it's the default license (no active license)
	return license.IsValid() && !license.IsExpired() && license.LicenseID != ""
}

// GetTiers returns information about available license types
func (s *LicenseService) GetTiers() map[string]interface{} {
	return map[string]interface{}{
		"license_types": []map[string]interface{}{
			{
				"type":                 "opensource",
				"name":                 "Open Source",
				"max_seats":            -1, // unlimited (soft limit)
				"max_providers":        -1, // unlimited
				"max_teams":            1,  // hard limit
				"data_retention_days":  7,
				"features":             []string{"Basic provider management", "Personal usage tracking", "Unlimited team members"},
			},
			{
				"type":                 "commercial",
				"name":                 "Commercial",
				"max_seats":            100, // soft limit (informational)
				"max_providers":        -1,  // unlimited
				"max_teams":            -1,  // unlimited (hard limit)
				"data_retention_days":  90,
				"features":             []string{"All Open Source features", "Team analytics dashboard", "Usage data export", "Multi-team support", "90-day data retention"},
			},
		},
	}
}

// VerifyLicenseIntegrity verifies that the stored license values match the encoded license key.
// This prevents database tampering where users modify tier/seats directly in the database.
// If a mismatch is detected, the license is automatically re-imported to correct the values.
//
// NOTE: This method intentionally queries the database (not cache) to compare stored values
// against the trusted decoded PEM values. This is the correct use case for GetTenantLicense().
//
// Returns (valid, mismatchDescription, error).
// - valid: true if no tampering detected (or after auto-correction)
// - mismatchDescription: describes which fields didn't match and were corrected (empty if valid)
// - error: only for technical failures (e.g., invalid license key format)
func (s *LicenseService) VerifyLicenseIntegrity(ctx context.Context, tenantID int64) (bool, string, error) {
	// Load license from database to compare with cached values
	storedLicense, err := s.licenseRepo.GetTenantLicense(ctx, tenantID)
	if err != nil {
		// No license found is not a tampering issue - return valid
		if err.Error() == "sql: no rows in result set" {
			return true, "", nil
		}
		// Other errors should be reported
		return false, "", fmt.Errorf("failed to retrieve license: %w", err)
	}

	// No license key means no active license to verify
	if storedLicense.LicenseKey == "" {
		return true, "", nil
	}

	// Decode the stored license_key PEM using public key
	decoded, err := license.Decode([]byte(storedLicense.LicenseKey), s.publicKey)
	if err != nil {
		// Invalid PEM format indicates technical corruption
		return false, "", fmt.Errorf("invalid license key format: %w", err)
	}

	// Extract embedded tier and seats from decoded license data
	var licenseData models.LicenseData
	if err := json.Unmarshal(decoded.Data, &licenseData); err != nil {
		return false, "", fmt.Errorf("failed to decode license data: %w", err)
	}

	// Compare stored values with embedded values
	var mismatches []string

	if storedLicense.LicenseType != licenseData.LicenseType {
		mismatches = append(mismatches, fmt.Sprintf("license_type mismatch: stored=%s, actual=%s", storedLicense.LicenseType, licenseData.LicenseType))
	}

	if storedLicense.MaxSeats != licenseData.MaxSeats {
		mismatches = append(mismatches, fmt.Sprintf("max_seats mismatch: stored=%d, actual=%d", storedLicense.MaxSeats, licenseData.MaxSeats))
	}

	if storedLicense.MaxTeams != licenseData.MaxTeams {
		mismatches = append(mismatches, fmt.Sprintf("max_teams mismatch: stored=%d, actual=%d", storedLicense.MaxTeams, licenseData.MaxTeams))
	}

	if storedLicense.DataRetentionDays != licenseData.DataRetentionDays {
		mismatches = append(mismatches, fmt.Sprintf("data_retention_days mismatch: stored=%d, actual=%d", storedLicense.DataRetentionDays, licenseData.DataRetentionDays))
	}

	// Build mismatch description
	var mismatchDesc string
	if len(mismatches) > 0 {
		for i, m := range mismatches {
			if i > 0 {
				mismatchDesc += "; "
			}
			mismatchDesc += m
		}
	}

	// If mismatches detected, re-import the license to correct the values
	if len(mismatches) > 0 {
		// Create corrected license object with values from the decoded license key
		correctedLicense := &models.License{
			TenantID:           tenantID,
			LicenseID:          decoded.ID,
			LicenseType:        licenseData.LicenseType,
			MaxSeats:           licenseData.MaxSeats,
			MaxTeams:           licenseData.MaxTeams,
			DataRetentionDays:  licenseData.DataRetentionDays,
			LicenseKey:         storedLicense.LicenseKey,
			IssuedAt:           time.Unix(decoded.IssuedAt, 0),
			ExpiresAt:          time.Unix(decoded.ExpiredAt, 0),
		}

		// Update database with corrected values
		if err := s.licenseRepo.UpdateTenantLicense(ctx, tenantID, correctedLicense); err != nil {
			return false, mismatchDesc, fmt.Errorf("failed to auto-correct license: %w", err)
		}

		// Return with mismatch description indicating it was corrected
		mismatchDesc += " (auto-corrected)"
		return true, mismatchDesc, nil
	}

	return true, "", nil
}
