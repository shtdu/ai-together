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
)

// LicenseData is the custom data embedded in go-license's Data field
type LicenseData struct {
	LicenseType       string `json:"license_type"`        // "opensource" or "commercial"
	MaxSeats          int    `json:"max_seats"`           // Soft limit for display/analytics, -1 for unlimited
	MaxTeams          int    `json:"max_teams"`           // 1 for opensource, -1 for unlimited
	DataRetentionDays int    `json:"data_retention_days"` // 7 for opensource, 90 for commercial
}

// License represents the decoded license stored in the database
type License struct {
	TenantID          int64     `json:"tenant_id"`
	CustomerName      string    `json:"customer_name"`         // Tenant/customer name
	LicenseID         string    `json:"license_id"`            // From go-license ID field
	LicenseType       string    `json:"license_type"`          // "opensource" or "commercial"
	MaxSeats          int       `json:"max_seats"`             // Soft limit (informational), -1 for unlimited
	MaxTeams          int       `json:"max_teams"`             // 1 for opensource, -1 for unlimited
	DataRetentionDays int       `json:"data_retention_days"`   // 7 for opensource, 90 for commercial
	LicenseKey        string    `json:"license_key,omitempty"` // Full PEM key
	IssuedAt          time.Time `json:"issued_at"`             // From go-license IssuedAt
	ExpiresAt         time.Time `json:"expires_at"`            // From go-license ExpiredAt

	// Deprecated: kept for backward compatibility during migration
	Tier  string `json:"tier,omitempty"`
	Seats int    `json:"seats,omitempty"`
}

// IsValid checks if the license has required fields set
func (l *License) IsValid() bool {
	if l == nil {
		return false
	}
	return l.LicenseID != "" && l.LicenseType != "" && !l.ExpiresAt.IsZero()
	// Note: MaxSeats is informational (soft limit), so 0 or -1 is valid
}

// IsExpired checks if the license has expired
func (l *License) IsExpired() bool {
	if l == nil || l.ExpiresAt.IsZero() {
		return true
	}
	return time.Now().After(l.ExpiresAt)
}

// GetLicenseType returns the license type ("opensource" or "commercial")
// For backward compatibility, also handles old tier format
func (l *License) GetLicenseType() string {
	if l == nil {
		return "opensource" // default
	}

	// If new field is set, use it
	if l.LicenseType != "" {
		return l.LicenseType
	}

	// Backward compatibility: map old tiers to new types
	if l.Tier == "0.0" {
		return "opensource"
	}
	return "commercial" // Tiers 1.0, 2.0, 3.0 all map to commercial
}

// MaxProvidersPerKind returns the provider limit per kind
// All license types now have unlimited providers
func (l *License) MaxProvidersPerKind() int {
	return -1 // unlimited for all license types
}

// LicenseTypeName returns the human-readable license type name
func (l *License) LicenseTypeName() string {
	if l == nil {
		return "Unknown"
	}
	switch l.GetLicenseType() {
	case "opensource":
		return "Open Source"
	case "commercial":
		return "Commercial"
	default:
		return "Unknown"
	}
}

// LicenseUsage represents current usage against license limits
type LicenseUsage struct {
	License        License        `json:"license"`
	CurrentUsers   int            `json:"current_users"`
	CurrentTeams   int            `json:"current_teams"`
	TeamsRemaining int            `json:"teams_remaining"`
	ProviderCounts map[string]int `json:"provider_counts"` // count per kind: {"claude": 1, "codex": 2}

	// Deprecated: kept for backward compatibility
	SeatsRemaining int `json:"seats_remaining,omitempty"`
}

// GetMaxSeats returns the seat limit (soft limit - informational only)
// Returns -1 for unlimited
func (l *License) GetMaxSeats() int {
	if l == nil {
		return -1 // unlimited by default
	}

	// If new field is set, use it
	if l.MaxSeats != 0 {
		return l.MaxSeats
	}

	// Backward compatibility: map old tiers to seat counts
	switch l.Tier {
	case "0.0":
		return 3
	case "1.0":
		return 10
	case "2.0":
		return 25
	case "3.0":
		return 100
	default:
		return -1 // unlimited
	}
}

// GetMaxTeams returns the team limit based on license type (hard enforced)
// Returns -1 for unlimited
func (l *License) GetMaxTeams() int {
	if l == nil {
		return 1 // default to opensource (1 team)
	}

	// If new field is set, use it
	if l.MaxTeams != 0 {
		return l.MaxTeams
	}

	// Backward compatibility: map old tiers
	if l.Tier == "0.0" {
		return 1
	}
	return -1 // unlimited for tiers 1.0+
}

// GetDataRetentionDays returns the data retention period in days
func (l *License) GetDataRetentionDays() int {
	if l == nil {
		return 7 // default to opensource (7 days)
	}

	// If new field is set, use it
	if l.DataRetentionDays != 0 {
		return l.DataRetentionDays
	}

	// Backward compatibility: map old tiers
	if l.Tier == "0.0" {
		return 7
	}
	return 90 // 90 days for tiers 1.0+
}

// ProviderKindUsage represents provider count for a specific kind
type ProviderKindUsage struct {
	Kind      string `json:"kind"`
	Count     int    `json:"count"`
	Limit     int    `json:"limit"`     // -1 for unlimited
	Remaining int    `json:"remaining"` // -1 for unlimited
}
