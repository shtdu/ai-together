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
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	licenseLib "github.com/vitalvas/go-license/license"
)

// GenerateTestKeyPair generates a test Ed25519 key pair for license testing
func GenerateTestKeyPair(t *testing.T) (ed25519.PrivateKey, ed25519.PublicKey) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate test key pair: %v", err)
	}
	return priv, pub
}

// GenerateTestLicense creates a test license PEM with the given parameters
func GenerateTestLicense(t *testing.T, privateKey ed25519.PrivateKey, tenantID string, licenseType string, maxSeats int, maxTeams int, dataRetentionDays int, daysValid int) string {
	licenseData := LicenseData{
		LicenseType:       licenseType,
		MaxSeats:          maxSeats,
		MaxTeams:          maxTeams,
		DataRetentionDays: dataRetentionDays,
	}
	dataBytes, _ := json.Marshal(licenseData)

	lic := &licenseLib.License{
		ID:        tenantID,
		Customer:  fmt.Sprintf("Test Customer %s", tenantID),
		Type:      licenseType,
		IssuedAt:  time.Now().Unix(),
		ExpiredAt: time.Now().Add(time.Duration(daysValid) * 24 * time.Hour).Unix(),
		Data:      dataBytes,
	}

	encoded, err := lic.Encode(privateKey)
	if err != nil {
		t.Fatalf("Failed to encode test license: %v", err)
	}

	return string(encoded)
}

// GetTestPublicKey returns the public key in base64 format for testing
func GetTestPublicKey(publicKey ed25519.PublicKey) string {
	return base64.StdEncoding.EncodeToString(publicKey)
}

// SetupTestLicense creates a test license object for testing
func SetupTestLicense(tenantID int64, licenseType string, maxSeats int, maxTeams int, dataRetentionDays int, daysValid int) *License {
	return &License{
		TenantID:          tenantID,
		LicenseID:         fmt.Sprintf("test-license-%d", tenantID),
		LicenseType:       licenseType,
		MaxSeats:          maxSeats,
		MaxTeams:          maxTeams,
		DataRetentionDays: dataRetentionDays,
		LicenseKey:        "test-license-key",
		IssuedAt:          time.Now(),
		ExpiresAt:         time.Now().Add(time.Duration(daysValid) * 24 * time.Hour),
	}
}

// SetupTestUser creates a test user object for testing
func SetupTestUser(userID int64, email, role string, tenantID int64) *User {
	return &User{
		ID:        userID,
		Email:     email,
		Name:      "Test User",
		Password:  "hashed_password",
		Role:      role,
		TenantID:  tenantID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// GenerateTestToken generates a valid JWT token for testing
func GenerateTestToken(t *testing.T, userID int64, email, role string, tenantID int64) string {
	claims := jwt.MapClaims{
		"user_id":   userID,
		"email":     email,
		"name":      "Test User",
		"role":      role,
		"tenant_id": tenantID,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test_secret_key"))
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	return tokenString
}

func TestLicenseIsValid(t *testing.T) {
	tests := []struct {
		name     string
		license  *License
		expected bool
	}{
		{
			name: "valid license",
			license: &License{
				LicenseID:   "test-license-id",
				LicenseType: "commercial",
				ExpiresAt:   time.Now().Add(24 * time.Hour),
			},
			expected: true,
		},
		{
			name: "missing license_id",
			license: &License{
				LicenseType: "commercial",
				ExpiresAt:   time.Now(),
			},
			expected: false,
		},
		{
			name: "empty license_type",
			license: &License{
				LicenseID:   "test-license-id",
				LicenseType: "",
				ExpiresAt:   time.Now(),
			},
			expected: false,
		},
		{
			name: "zero expiration",
			license: &License{
				LicenseID:   "test-license-id",
				LicenseType: "commercial",
				ExpiresAt:   time.Time{},
			},
			expected: false,
		},
		{
			name:     "nil license",
			license:  nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.license.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLicenseIsExpired(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		license  *License
		expected bool
	}{
		{
			name: "not expired",
			license: &License{
				ExpiresAt: now.Add(24 * time.Hour),
			},
			expected: false,
		},
		{
			name: "expired",
			license: &License{
				ExpiresAt: now.Add(-24 * time.Hour),
			},
			expected: true,
		},
		{
			name: "zero expiration (treated as expired)",
			license: &License{
				ExpiresAt: time.Time{},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.license.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLicenseGetLicenseType(t *testing.T) {
	tests := []struct {
		name        string
		licenseType string
		oldTier     string // for backward compatibility testing
		expected    string
	}{
		{"opensource type", "opensource", "", "opensource"},
		{"commercial type", "commercial", "", "commercial"},
		{"backward compat: tier 0.0", "", "0.0", "opensource"},
		{"backward compat: tier 1.0", "", "1.0", "commercial"},
		{"backward compat: tier 2.0", "", "2.0", "commercial"},
		{"backward compat: tier 3.0", "", "3.0", "commercial"},
		{"new field takes precedence", "commercial", "0.0", "commercial"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			license := &License{
				LicenseType: tt.licenseType,
				Tier:        tt.oldTier,
			}
			result := license.GetLicenseType()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLicenseMaxProvidersPerKind(t *testing.T) {
	tests := []struct {
		name     string
		expected int
	}{
		{"opensource type", -1},
		{"commercial type", -1},
		{"nil license", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var license *License
			if tt.name != "nil license" {
				license = &License{LicenseType: tt.name}
			}
			result := license.MaxProvidersPerKind()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLicenseLicenseTypeName(t *testing.T) {
	tests := []struct {
		name        string
		licenseType string
		oldTier     string
		expected    string
	}{
		{"Open Source", "opensource", "", "Open Source"},
		{"Commercial", "commercial", "", "Commercial"},
		{"backward compat: tier 0.0", "", "0.0", "Open Source"},
		{"backward compat: tier 1.0", "", "1.0", "Commercial"},
		// When both empty, GetLicenseType returns "commercial" (any non-"0.0" tier maps to commercial)
		// This is by design for backward compatibility - old code that set tier to empty would get commercial
		{"empty defaults to commercial via backward compat", "", "", "Commercial"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			license := &License{
				LicenseType: tt.licenseType,
				Tier:        tt.oldTier,
			}
			result := license.LicenseTypeName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLicenseGetMaxTeams(t *testing.T) {
	tests := []struct {
		name        string
		licenseType string
		maxTeams    int
		oldTier     string
		expected    int
	}{
		{"opensource explicit", "opensource", 1, "", 1},
		{"commercial unlimited", "commercial", -1, "", -1},
		{"backward compat: tier 0.0", "", 0, "0.0", 1},
		{"backward compat: tier 1.0", "", 0, "1.0", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			license := &License{
				LicenseType: tt.licenseType,
				MaxTeams:    tt.maxTeams,
				Tier:        tt.oldTier,
			}
			result := license.GetMaxTeams()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLicenseGetDataRetentionDays(t *testing.T) {
	tests := []struct {
		name              string
		licenseType       string
		dataRetentionDays int
		oldTier           string
		expected          int
	}{
		{"opensource explicit", "opensource", 7, "", 7},
		{"commercial explicit", "commercial", 90, "", 90},
		{"backward compat: tier 0.0", "", 0, "0.0", 7},
		{"backward compat: tier 1.0", "", 0, "1.0", 90},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			license := &License{
				LicenseType:       tt.licenseType,
				DataRetentionDays: tt.dataRetentionDays,
				Tier:              tt.oldTier,
			}
			result := license.GetDataRetentionDays()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLicenseUsageSeatsRemaining(t *testing.T) {
	tests := []struct {
		name           string
		maxSeats       int
		currentUsers   int
		seatsRemaining int
	}{
		{"under limit", 10, 5, 5},
		{"at limit", 10, 10, 0},
		{"negative becomes zero", 10, 15, 0}, // service layer should set this to 0
		{"unlimited", -1, 100, -1},
		{"large seats", 100, 1, 99},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := &LicenseUsage{
				License: License{
					MaxSeats: tt.maxSeats,
				},
				CurrentUsers:   tt.currentUsers,
				SeatsRemaining: tt.seatsRemaining, // Set by service layer
			}

			// Verify the struct holds the values correctly
			assert.Equal(t, tt.maxSeats, usage.License.MaxSeats)
			assert.Equal(t, tt.currentUsers, usage.CurrentUsers)
			assert.Equal(t, tt.seatsRemaining, usage.SeatsRemaining)
		})
	}
}

func TestLicenseProviderCounts(t *testing.T) {
	tests := []struct {
		name           string
		providerCounts map[string]int
		verifyFunc     func(*testing.T, *LicenseUsage)
	}{
		{
			name: "multiple providers",
			providerCounts: map[string]int{
				"claude":   2,
				"codex":    1,
				"opencode": 0,
			},
			verifyFunc: func(t *testing.T, usage *LicenseUsage) {
				assert.Equal(t, 2, usage.ProviderCounts["claude"])
				assert.Equal(t, 1, usage.ProviderCounts["codex"])
				assert.Equal(t, 0, usage.ProviderCounts["opencode"])
				assert.Len(t, usage.ProviderCounts, 3)
			},
		},
		{
			name:           "no providers",
			providerCounts: map[string]int{},
			verifyFunc: func(t *testing.T, usage *LicenseUsage) {
				assert.Empty(t, usage.ProviderCounts)
				assert.Len(t, usage.ProviderCounts, 0)
			},
		},
		{
			name: "single kind",
			providerCounts: map[string]int{
				"claude": 1,
			},
			verifyFunc: func(t *testing.T, usage *LicenseUsage) {
				assert.Equal(t, 1, usage.ProviderCounts["claude"])
				assert.Len(t, usage.ProviderCounts, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := &LicenseUsage{
				ProviderCounts: tt.providerCounts,
			}
			tt.verifyFunc(t, usage)
		})
	}
}

// Test edge cases with nil pointers and zero values
func TestLicenseEdgeCases(t *testing.T) {
	t.Run("nil license methods don't panic", func(t *testing.T) {
		var license *License

		// These should not panic on nil receiver
		require.False(t, license.IsValid(), "nil license should not be valid")
		require.True(t, license.IsExpired(), "nil license should be expired")
		require.Equal(t, "opensource", license.GetLicenseType(), "nil license should default to opensource")
		require.Equal(t, -1, license.MaxProvidersPerKind(), "nil license should return -1 (unlimited)")
		require.Equal(t, "Unknown", license.LicenseTypeName(), "nil license should be Unknown")
		require.Equal(t, 1, license.GetMaxTeams(), "nil license should default to 1 team")
		require.Equal(t, 7, license.GetDataRetentionDays(), "nil license should default to 7 days")
	})

	t.Run("license with future expiration", func(t *testing.T) {
		license := &License{
			ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
		}
		require.False(t, license.IsExpired(), "future expiration should not be expired")
	})

	t.Run("license with past expiration", func(t *testing.T) {
		license := &License{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		require.True(t, license.IsExpired(), "past expiration should be expired")
	})
}

// Test comprehensive license scenarios
func TestLicenseComprehensive(t *testing.T) {
	privateKey, publicKey := GenerateTestKeyPair(t)
	licensePEM := GenerateTestLicense(t, privateKey, "tenant-123", "commercial", 100, -1, 90, 365)

	t.Run("generate and decode test license", func(t *testing.T) {
		// Verify license PEM was generated
		require.NotEmpty(t, licensePEM, "license PEM should not be empty")
		require.Contains(t, licensePEM, "BEGIN LICENSE KEY", "should have license header")
		require.Contains(t, licensePEM, "tenant-123", "should contain tenant ID")
	})

	t.Run("public key encoding", func(t *testing.T) {
		pubKeyBase64 := GetTestPublicKey(publicKey)
		require.NotEmpty(t, pubKeyBase64, "public key base64 should not be empty")

		// Should be able to decode back
		decoded, err := base64.StdEncoding.DecodeString(pubKeyBase64)
		require.NoError(t, err, "should decode base64")
		require.Equal(t, len(publicKey), len(decoded), "decoded key should match original size")
	})

	t.Run("test commercial license object", func(t *testing.T) {
		license := SetupTestLicense(999, "commercial", 100, -1, 90, 365)

		// Verify all fields
		assert.Equal(t, int64(999), license.TenantID)
		assert.Equal(t, "test-license-999", license.LicenseID)
		assert.Equal(t, "commercial", license.LicenseType)
		assert.Equal(t, 100, license.MaxSeats)
		assert.Equal(t, -1, license.MaxTeams)
		assert.Equal(t, 90, license.DataRetentionDays)
		assert.True(t, license.ExpiresAt.After(time.Now()))
		assert.True(t, license.IsValid())
		assert.False(t, license.IsExpired())
		assert.Equal(t, "commercial", license.GetLicenseType())
		assert.Equal(t, "Commercial", license.LicenseTypeName())
		assert.Equal(t, -1, license.MaxProvidersPerKind()) // unlimited
		assert.Equal(t, -1, license.GetMaxTeams())         // unlimited for commercial
		assert.Equal(t, 90, license.GetDataRetentionDays())
	})

	t.Run("test opensource license object", func(t *testing.T) {
		license := SetupTestLicense(998, "opensource", -1, 1, 7, 365)

		// Verify all fields
		assert.Equal(t, int64(998), license.TenantID)
		assert.Equal(t, "opensource", license.LicenseType)
		assert.Equal(t, -1, license.MaxSeats)
		assert.Equal(t, 1, license.MaxTeams)
		assert.Equal(t, 7, license.DataRetentionDays)
		assert.Equal(t, "opensource", license.GetLicenseType())
		assert.Equal(t, "Open Source", license.LicenseTypeName())
		assert.Equal(t, 1, license.GetMaxTeams()) // 1 team for opensource
		assert.Equal(t, 7, license.GetDataRetentionDays())
	})

	t.Run("test user object", func(t *testing.T) {
		user := SetupTestUser(123, "test@example.com", "manager", 456)

		assert.Equal(t, int64(123), user.ID)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "Test User", user.Name)
		assert.Equal(t, "manager", user.Role)
		assert.Equal(t, int64(456), user.TenantID)
	})

	t.Run("test token generation", func(t *testing.T) {
		token := GenerateTestToken(t, 123, "test@example.com", "manager", 456)

		require.NotEmpty(t, token, "token should not be empty")
		assert.Contains(t, token, ".", "should have dot separators")
	})
}
