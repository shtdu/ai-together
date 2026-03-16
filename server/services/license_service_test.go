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
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	licenseLib "github.com/vitalvas/go-license/license"
	"switch-server/models"
)

// MockLicenseRepository is a mock for testing
type MockLicenseRepository struct {
	mock.Mock
}

func (m *MockLicenseRepository) GetTenantLicense(ctx context.Context, tenantID int64) (*models.License, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.License), args.Error(1)
}

func (m *MockLicenseRepository) UpdateTenantLicense(ctx context.Context, tenantID int64, license *models.License) error {
	args := m.Called(ctx, tenantID, license)
	return args.Error(0)
}

func (m *MockLicenseRepository) CountTenantUsers(ctx context.Context, tenantID int64) (int, error) {
	args := m.Called(ctx, tenantID)
	return args.Int(0), args.Error(1)
}

func (m *MockLicenseRepository) CountTenantProvidersByKind(ctx context.Context, tenantID int64, kind string) (int, error) {
	args := m.Called(ctx, tenantID, kind)
	return args.Int(0), args.Error(1)
}

func (m *MockLicenseRepository) GetProviderCountsByKind(ctx context.Context, tenantID int64) (map[string]int, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *MockLicenseRepository) CountTenantTeams(ctx context.Context, tenantID int64) (int, error) {
	args := m.Called(ctx, tenantID)
	return args.Int(0), args.Error(1)
}

func (m *MockLicenseRepository) GetAllLicenses(ctx context.Context) ([]*models.License, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.License), args.Error(1)
}

// newMockService creates a new LicenseService with a mock repository for testing
func newMockService(t *testing.T, pubKey ed25519.PublicKey, licenses ...*models.License) (*LicenseService, *MockLicenseRepository) {
	mockRepo := new(MockLicenseRepository)
	// Mock license list
	mockRepo.On("GetAllLicenses", mock.Anything).Return(licenses, nil)

	service, err := NewLicenseService(mockRepo, GetTestPublicKey(pubKey))
	require.NoError(t, err)
	return service, mockRepo
}

// newMockServiceWithLicenses creates a new LicenseService with actual encoded license keys
// This is the preferred way for tests that need valid licenses in the cache
func newMockServiceWithLicenses(t *testing.T, privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey, tenantID int64, licenseType string, maxSeats int, maxTeams int, dataRetentionDays int, daysValid int) (*LicenseService, *MockLicenseRepository, *models.License) {
	mockRepo := new(MockLicenseRepository)

	// Generate a real license PEM
	licensePEM := GenerateTestLicense(t, privateKey, fmt.Sprintf("tenant-%d", tenantID), licenseType, maxSeats, maxTeams, dataRetentionDays, daysValid)

	// Decode it to get the values
	decoded, err := licenseLib.Decode([]byte(licensePEM), publicKey)
	require.NoError(t, err)

	var licenseData models.LicenseData
	require.NoError(t, json.Unmarshal(decoded.Data, &licenseData))

	// Create the license object with decoded values
	license := &models.License{
		TenantID:          tenantID,
		LicenseID:         decoded.ID,
		LicenseType:       licenseData.LicenseType,
		MaxSeats:          licenseData.MaxSeats,
		MaxTeams:          licenseData.MaxTeams,
		DataRetentionDays: licenseData.DataRetentionDays,
		LicenseKey:        licensePEM,
		IssuedAt:          time.Unix(decoded.IssuedAt, 0),
		ExpiresAt:         time.Unix(decoded.ExpiredAt, 0),
	}

	// Mock license list with our properly encoded license
	mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{license}, nil)

	service, err := NewLicenseService(mockRepo, GetTestPublicKey(publicKey))
	require.NoError(t, err)

	return service, mockRepo, license
}

// Test helper functions (copied from models/license_test.go since they can't be shared across packages)

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
	licenseData := models.LicenseData{
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

// Helper to create a valid test public key
func generateTestPublicKey(t *testing.T) ed25519.PublicKey {
	_, pub := GenerateTestKeyPair(t)
	return pub
}

func TestNewLicenseService(t *testing.T) {
	t.Run("valid public key", func(t *testing.T) {
		pubKey := generateTestPublicKey(t)
		pubKeyBase64 := GetTestPublicKey(pubKey)

		mockRepo := new(MockLicenseRepository)
		// Mock empty license list
		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{}, nil)
		service, err := NewLicenseService(mockRepo, pubKeyBase64)

		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, mockRepo, service.licenseRepo)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid base64", func(t *testing.T) {
		mockRepo := new(MockLicenseRepository)
		// Don't mock GetAllLicenses since it won't be called (fails before that)
		service, err := NewLicenseService(mockRepo, "invalid-base64!!!")

		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "invalid public key")
	})

	t.Run("wrong key size", func(t *testing.T) {
		mockRepo := new(MockLicenseRepository)
		// Don't mock GetAllLicenses since it won't be called (fails before that)
		shortKey := "AQIDBAUG" // base64 of 4 bytes

		service, err := NewLicenseService(mockRepo, shortKey)

		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "invalid public key size")
	})
}

func TestLicenseService_GetLicense(t *testing.T) {
	ctx := context.Background()
	tenantID := int64(123)

	t.Run("success - returns license from repository", func(t *testing.T) {
		pubKey := generateTestPublicKey(t)
		service, mockRepo := newMockService(t, pubKey)

		expectedLicense := &models.License{
			TenantID:  tenantID,
			LicenseID: "lic-123",
			Tier:      "2.0",
			Seats:     25,
			ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
		}

		mockRepo.On("GetTenantLicense", ctx, tenantID).Return(expectedLicense, nil)

		license, err := service.GetLicense(ctx, tenantID)

		assert.NoError(t, err)
		assert.Equal(t, expectedLicense, license)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - repository error propagation", func(t *testing.T) {
		pubKey := generateTestPublicKey(t)
		service, mockRepo := newMockService(t, pubKey)

		expectedError := errors.New("database connection failed")
		mockRepo.On("GetTenantLicense", ctx, tenantID).Return(nil, expectedError)

		license, err := service.GetLicense(ctx, tenantID)

		assert.Error(t, err)
		assert.Nil(t, license)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestLicenseService_GetLicenseUsage(t *testing.T) {
	ctx := context.Background()
	tenantID := int64(123)

	t.Run("success - valid license with usage", func(t *testing.T) {
		privateKey, publicKey := GenerateTestKeyPair(t)
		service, mockRepo, license := newMockServiceWithLicenses(t, privateKey, publicKey, tenantID, "commercial", 100, -1, 90, 365)

		mockRepo.On("CountTenantUsers", ctx, tenantID).Return(10, nil)
		mockRepo.On("CountTenantTeams", ctx, tenantID).Return(3, nil)
		mockRepo.On("GetProviderCountsByKind", ctx, tenantID).Return(map[string]int{
			"claude": 2,
			"codex":  1,
		}, nil)

		usage, err := service.GetLicenseUsage(ctx, tenantID)

		assert.NoError(t, err)
		assert.Equal(t, *license, usage.License)
		assert.Equal(t, 10, usage.CurrentUsers)
		assert.Equal(t, 90, usage.SeatsRemaining) // 100 - 10
		assert.Equal(t, 3, usage.CurrentTeams)
		assert.Equal(t, -1, usage.TeamsRemaining) // unlimited
		assert.Equal(t, 2, usage.ProviderCounts["claude"])
		assert.Equal(t, 1, usage.ProviderCounts["codex"])
		mockRepo.AssertExpectations(t)
	})

	t.Run("success - at seat limit", func(t *testing.T) {
		privateKey, publicKey := GenerateTestKeyPair(t)
		service, mockRepo, _ := newMockServiceWithLicenses(t, privateKey, publicKey, tenantID, "commercial", 100, -1, 90, 365)

		mockRepo.On("CountTenantUsers", ctx, tenantID).Return(10, nil)
		mockRepo.On("CountTenantTeams", ctx, tenantID).Return(3, nil)
		mockRepo.On("GetProviderCountsByKind", ctx, tenantID).Return(map[string]int{}, nil)

		usage, err := service.GetLicenseUsage(ctx, tenantID)

		assert.NoError(t, err)
		assert.Equal(t, 100, usage.License.MaxSeats)
		assert.Equal(t, 90, usage.SeatsRemaining) // 100 - 10
		mockRepo.AssertExpectations(t)
	})

	t.Run("success - over seat limit (negative becomes zero)", func(t *testing.T) {
		privateKey, publicKey := GenerateTestKeyPair(t)
		service, mockRepo, _ := newMockServiceWithLicenses(t, privateKey, publicKey, tenantID, "commercial", 100, -1, 90, 365)

		mockRepo.On("CountTenantUsers", ctx, tenantID).Return(10, nil) // More users than seats
		mockRepo.On("CountTenantTeams", ctx, tenantID).Return(3, nil)
		mockRepo.On("GetProviderCountsByKind", ctx, tenantID).Return(map[string]int{}, nil)

		usage, err := service.GetLicenseUsage(ctx, tenantID)

		assert.NoError(t, err)
		assert.Equal(t, 90, usage.SeatsRemaining) // 100 - 10
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - CountTenantUsers fails", func(t *testing.T) {
		privateKey, publicKey := GenerateTestKeyPair(t)
		service, mockRepo, _ := newMockServiceWithLicenses(t, privateKey, publicKey, tenantID, "commercial", 100, -1, 90, 365)

		expectedError := errors.New("count error")
		mockRepo.On("CountTenantUsers", ctx, tenantID).Return(0, expectedError)
		// CountTenantTeams won't be called because CountTenantUsers fails first

		usage, err := service.GetLicenseUsage(ctx, tenantID)

		assert.Error(t, err)
		assert.Nil(t, usage)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - CountTenantTeams fails", func(t *testing.T) {
		privateKey, publicKey := GenerateTestKeyPair(t)
		service, mockRepo, _ := newMockServiceWithLicenses(t, privateKey, publicKey, tenantID, "commercial", 100, -1, 90, 365)

		mockRepo.On("CountTenantUsers", ctx, tenantID).Return(5, nil)
		mockRepo.On("GetProviderCountsByKind", ctx, tenantID).Return(map[string]int{}, nil)
		expectedError := errors.New("teams count error")
		mockRepo.On("CountTenantTeams", ctx, tenantID).Return(0, expectedError)

		usage, err := service.GetLicenseUsage(ctx, tenantID)

		assert.Error(t, err)
		assert.Nil(t, usage)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - GetProviderCountsByKind fails", func(t *testing.T) {
		privateKey, publicKey := GenerateTestKeyPair(t)
		service, mockRepo, _ := newMockServiceWithLicenses(t, privateKey, publicKey, tenantID, "commercial", 100, -1, 90, 365)

		mockRepo.On("CountTenantUsers", ctx, tenantID).Return(5, nil)
		expectedError := errors.New("provider count error")
		mockRepo.On("GetProviderCountsByKind", ctx, tenantID).Return(nil, expectedError)
		// CountTenantTeams won't be called because GetProviderCountsByKind fails first

		usage, err := service.GetLicenseUsage(ctx, tenantID)

		assert.Error(t, err)
		assert.Nil(t, usage)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestLicenseService_IsLicenseValid(t *testing.T) {
	ctx := context.Background()
	tenantID := int64(123)

	t.Run("valid and not expired", func(t *testing.T) {
		privateKey, publicKey := GenerateTestKeyPair(t)
		service, mockRepo, _ := newMockServiceWithLicenses(t, privateKey, publicKey, tenantID, "commercial", 100, -1, 90, 365)

		isValid, message, err := service.IsLicenseValid(ctx, tenantID)

		assert.NoError(t, err)
		assert.True(t, isValid)
		assert.Empty(t, message)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid - missing fields (no license)", func(t *testing.T) {
		pubKey := generateTestPublicKey(t)
		// No license preloaded - will use default
		service, mockRepo := newMockService(t, pubKey)

		isValid, message, err := service.IsLicenseValid(ctx, tenantID)

		assert.NoError(t, err)
		assert.False(t, isValid)
		assert.Equal(t, "No valid license found", message)
		mockRepo.AssertExpectations(t)
	})

	t.Run("expired license - falls back to default", func(t *testing.T) {
		privateKey, publicKey := GenerateTestKeyPair(t)
		mockRepo := new(MockLicenseRepository)

		// Create an expired license by issuing it in the past
		licenseData := models.LicenseData{
			LicenseType:       "commercial",
			MaxSeats:          100,
			MaxTeams:          -1,
			DataRetentionDays: 90,
		}
		dataBytes, _ := json.Marshal(licenseData)

		// Issue license 10 days ago, expire 5 days ago
		expiredLicense := &licenseLib.License{
			ID:        fmt.Sprintf("tenant-%d", tenantID),
			Customer:  fmt.Sprintf("Test Customer %d", tenantID),
			Type:      "commercial",
			IssuedAt:  time.Now().Add(-10 * 24 * time.Hour).Unix(),
			ExpiredAt: time.Now().Add(-5 * 24 * time.Hour).Unix(),
			Data:      dataBytes,
		}

		encoded, err := expiredLicense.Encode(privateKey)
		require.NoError(t, err)

		// Decode it to get the values
		decoded, err := licenseLib.Decode(encoded, publicKey)
		require.NoError(t, err)

		var storedLicenseData models.LicenseData
		require.NoError(t, json.Unmarshal(decoded.Data, &storedLicenseData))

		license := &models.License{
			TenantID:          tenantID,
			LicenseID:         decoded.ID,
			LicenseType:       storedLicenseData.LicenseType,
			MaxSeats:          storedLicenseData.MaxSeats,
			MaxTeams:          storedLicenseData.MaxTeams,
			DataRetentionDays: storedLicenseData.DataRetentionDays,
			LicenseKey:        string(encoded),
			IssuedAt:          time.Unix(decoded.IssuedAt, 0),
			ExpiresAt:         time.Unix(decoded.ExpiredAt, 0),
		}

		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{license}, nil)

		service, err := NewLicenseService(mockRepo, GetTestPublicKey(publicKey))
		require.NoError(t, err)

		isValid, message, err := service.IsLicenseValid(ctx, tenantID)

		assert.NoError(t, err)
		assert.False(t, isValid)
		// Expired licenses fall back to default tier, which is valid but not an active license
		assert.Equal(t, "No valid license found", message)
		mockRepo.AssertExpectations(t)
	})
}

func TestLicenseService_ActivateLicense(t *testing.T) {
	ctx := context.Background()
	tenantID := int64(123)

	privateKey, publicKey := GenerateTestKeyPair(t)
	pubKeyBase64 := GetTestPublicKey(publicKey)

	t.Run("success - valid license", func(t *testing.T) {
		mockRepo := new(MockLicenseRepository)
		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{}, nil)
		service, _ := NewLicenseService(mockRepo, pubKeyBase64)

		licensePEM := GenerateTestLicense(t, privateKey, "tenant-123", "commercial", 100, -1, 90, 365)

		mockRepo.On("UpdateTenantLicense", ctx, tenantID, mock.AnythingOfType("*models.License")).Return(nil)

		err := service.ActivateLicense(ctx, tenantID, licensePEM)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - invalid PEM format", func(t *testing.T) {
		mockRepo := new(MockLicenseRepository)
		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{}, nil)
		service, _ := NewLicenseService(mockRepo, pubKeyBase64)

		err := service.ActivateLicense(ctx, tenantID, "not-a-valid-pem")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid license key")
	})

	t.Run("error - expired license", func(t *testing.T) {
		mockRepo := new(MockLicenseRepository)
		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{}, nil)
		service, _ := NewLicenseService(mockRepo, pubKeyBase64)

		// Create an expired license by issuing it in the past
		licenseData := models.LicenseData{
			LicenseType:       "commercial",
			MaxSeats:          100,
			MaxTeams:          -1,
			DataRetentionDays: 90,
		}
		dataBytes, _ := json.Marshal(licenseData)
		expiredLicense := &licenseLib.License{
			ID:        "tenant-123",
			Customer:  "Test Customer tenant-123",
			Type:      "commercial",
			IssuedAt:  time.Now().Add(-10 * 24 * time.Hour).Unix(),
			ExpiredAt: time.Now().Add(-5 * 24 * time.Hour).Unix(),
			Data:      dataBytes,
		}

		encoded, err := expiredLicense.Encode(privateKey)
		if err != nil {
			t.Fatalf("Failed to encode expired license: %v", err)
		}

		err = service.ActivateLicense(ctx, tenantID, string(encoded))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "license has expired")
	})

	t.Run("error - wrong signature", func(t *testing.T) {
		mockRepo := new(MockLicenseRepository)
		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{}, nil)

		// Create a different key pair for signing
		wrongPrivateKey, _ := GenerateTestKeyPair(t)
		service, _ := NewLicenseService(mockRepo, pubKeyBase64) // Public key doesn't match wrongPrivateKey

		// Sign with wrong private key
		wrongLicensePEM := GenerateTestLicense(t, wrongPrivateKey, "tenant-123", "commercial", 100, -1, 90, 365)

		err := service.ActivateLicense(ctx, tenantID, wrongLicensePEM)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid license key")
	})

	t.Run("error - repository update failure", func(t *testing.T) {
		mockRepo := new(MockLicenseRepository)
		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{}, nil)
		service, _ := NewLicenseService(mockRepo, pubKeyBase64)

		licensePEM := GenerateTestLicense(t, privateKey, "tenant-123", "commercial", 100, -1, 90, 365)

		expectedError := errors.New("database update failed")
		mockRepo.On("UpdateTenantLicense", ctx, tenantID, mock.AnythingOfType("*models.License")).Return(expectedError)

		err := service.ActivateLicense(ctx, tenantID, licensePEM)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestLicenseService_GetEffectiveLicense(t *testing.T) {
	ctx := context.Background()
	tenantID := int64(123)

	t.Run("valid license - returns as-is", func(t *testing.T) {
		privateKey, publicKey := GenerateTestKeyPair(t)
		service, mockRepo, license := newMockServiceWithLicenses(t, privateKey, publicKey, tenantID, "commercial", 100, -1, 90, 365)

		effective := service.GetEffectiveLicense(ctx, tenantID)

		assert.Equal(t, license.LicenseID, effective.LicenseID)
		assert.Equal(t, "commercial", effective.LicenseType)
		assert.Equal(t, 100, effective.MaxSeats)
		assert.Equal(t, -1, effective.MaxTeams)
		assert.Equal(t, 90, effective.DataRetentionDays)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid license - returns default", func(t *testing.T) {
		pubKey := generateTestPublicKey(t)
		// No license preloaded - will use default
		service, mockRepo := newMockService(t, pubKey)

		effective := service.GetEffectiveLicense(ctx, tenantID)

		assert.Equal(t, DefaultLicenseType, effective.LicenseType)
		assert.Equal(t, DefaultMaxSeats, effective.MaxSeats)
		assert.Equal(t, DefaultMaxTeams, effective.MaxTeams)
		assert.Equal(t, DefaultDataRetentionDays, effective.DataRetentionDays)
		mockRepo.AssertExpectations(t)
	})

	t.Run("expired license - returns default", func(t *testing.T) {
		privateKey, publicKey := GenerateTestKeyPair(t)
		mockRepo := new(MockLicenseRepository)

		// Create an expired license by issuing it in the past
		licenseData := models.LicenseData{
			LicenseType:       "commercial",
			MaxSeats:          100,
			MaxTeams:          -1,
			DataRetentionDays: 90,
		}
		dataBytes, _ := json.Marshal(licenseData)

		// Issue license 10 days ago, expire 5 days ago
		expiredLicense := &licenseLib.License{
			ID:        fmt.Sprintf("tenant-%d", tenantID),
			Customer:  fmt.Sprintf("Test Customer %d", tenantID),
			Type:      "commercial",
			IssuedAt:  time.Now().Add(-10 * 24 * time.Hour).Unix(),
			ExpiredAt: time.Now().Add(-5 * 24 * time.Hour).Unix(),
			Data:      dataBytes,
		}

		encoded, err := expiredLicense.Encode(privateKey)
		require.NoError(t, err)

		// Decode it to get the values
		decoded, err := licenseLib.Decode(encoded, publicKey)
		require.NoError(t, err)

		var storedLicenseData models.LicenseData
		require.NoError(t, json.Unmarshal(decoded.Data, &storedLicenseData))

		license := &models.License{
			TenantID:          tenantID,
			LicenseID:         decoded.ID,
			LicenseType:       storedLicenseData.LicenseType,
			MaxSeats:          storedLicenseData.MaxSeats,
			MaxTeams:          storedLicenseData.MaxTeams,
			DataRetentionDays: storedLicenseData.DataRetentionDays,
			LicenseKey:        string(encoded),
			IssuedAt:          time.Unix(decoded.IssuedAt, 0),
			ExpiresAt:         time.Unix(decoded.ExpiredAt, 0),
		}

		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{license}, nil)

		service, err := NewLicenseService(mockRepo, GetTestPublicKey(publicKey))
		require.NoError(t, err)

		effective := service.GetEffectiveLicense(ctx, tenantID)

		assert.Equal(t, DefaultLicenseType, effective.LicenseType)
		assert.Equal(t, DefaultMaxSeats, effective.MaxSeats)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error - returns default", func(t *testing.T) {
		pubKey := generateTestPublicKey(t)
		// No license preloaded - will use default
		service, mockRepo := newMockService(t, pubKey)

		effective := service.GetEffectiveLicense(ctx, tenantID)

		assert.Equal(t, DefaultLicenseType, effective.LicenseType)
		assert.Equal(t, DefaultMaxSeats, effective.MaxSeats)
		mockRepo.AssertExpectations(t)
	})
}

func TestLicenseService_HasActiveLicense(t *testing.T) {
	ctx := context.Background()
	tenantID := int64(123)

	t.Run("valid non-expired license - true", func(t *testing.T) {
		privateKey, publicKey := GenerateTestKeyPair(t)
		service, mockRepo, _ := newMockServiceWithLicenses(t, privateKey, publicKey, tenantID, "commercial", 100, -1, 90, 365)

		hasActive := service.HasActiveLicense(ctx, tenantID)

		assert.True(t, hasActive)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid license - false (no license)", func(t *testing.T) {
		pubKey := generateTestPublicKey(t)
		// No license preloaded - will use default
		service, mockRepo := newMockService(t, pubKey)

		hasActive := service.HasActiveLicense(ctx, tenantID)

		assert.False(t, hasActive)
		mockRepo.AssertExpectations(t)
	})

	t.Run("expired license - false", func(t *testing.T) {
		privateKey, publicKey := GenerateTestKeyPair(t)
		mockRepo := new(MockLicenseRepository)

		// Create an expired commercial license by issuing it in the past
		licenseData := models.LicenseData{
			LicenseType:       "commercial",
			MaxSeats:          25,
			MaxTeams:          -1,
			DataRetentionDays: 90,
		}
		dataBytes, _ := json.Marshal(licenseData)

		// Issue license 10 days ago, expire 5 days ago
		expiredLicense := &licenseLib.License{
			ID:        fmt.Sprintf("tenant-%d", tenantID),
			Customer:  fmt.Sprintf("Test Customer %d", tenantID),
			Type:      "commercial",
			IssuedAt:  time.Now().Add(-10 * 24 * time.Hour).Unix(),
			ExpiredAt: time.Now().Add(-5 * 24 * time.Hour).Unix(),
			Data:      dataBytes,
		}

		encoded, err := expiredLicense.Encode(privateKey)
		require.NoError(t, err)

		// Decode it to get the values
		decoded, err := licenseLib.Decode(encoded, publicKey)
		require.NoError(t, err)

		var storedLicenseData models.LicenseData
		require.NoError(t, json.Unmarshal(decoded.Data, &storedLicenseData))

		license := &models.License{
			TenantID:          tenantID,
			LicenseID:         decoded.ID,
			LicenseType:       storedLicenseData.LicenseType,
			MaxSeats:          storedLicenseData.MaxSeats,
			MaxTeams:          storedLicenseData.MaxTeams,
			DataRetentionDays: storedLicenseData.DataRetentionDays,
			LicenseKey:        string(encoded),
			IssuedAt:          time.Unix(decoded.IssuedAt, 0),
			ExpiresAt:         time.Unix(decoded.ExpiredAt, 0),
		}

		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{license}, nil)

		service, err := NewLicenseService(mockRepo, GetTestPublicKey(publicKey))
		require.NoError(t, err)

		hasActive := service.HasActiveLicense(ctx, tenantID)

		assert.False(t, hasActive)
		mockRepo.AssertExpectations(t)
	})
}

func TestLicenseService_GetTiers(t *testing.T) {
	pubKey := generateTestPublicKey(t)
	service, _ := newMockService(t, pubKey)

	tiers := service.GetTiers()

	assert.NotNil(t, tiers)
	assert.Contains(t, tiers, "license_types")

	tiersList, ok := tiers["license_types"].([]map[string]interface{})
	assert.True(t, ok, "license_types should be a list of maps")
	assert.Len(t, tiersList, 2, "should have 2 license types")

	// Verify opensource type
	assert.Equal(t, "opensource", tiersList[0]["type"])
	assert.Equal(t, "Open Source", tiersList[0]["name"])
	assert.Equal(t, -1, tiersList[0]["max_seats"]) // unlimited
	assert.Equal(t, 1, tiersList[0]["max_teams"])
	assert.Equal(t, -1, tiersList[0]["max_providers"]) // unlimited

	// Verify commercial type
	assert.Equal(t, "commercial", tiersList[1]["type"])
	assert.Equal(t, "Commercial", tiersList[1]["name"])
	assert.Equal(t, 100, tiersList[1]["max_seats"])
	assert.Equal(t, -1, tiersList[1]["max_teams"])     // unlimited
	assert.Equal(t, -1, tiersList[1]["max_providers"]) // unlimited
}

func TestLicenseService_VerifyLicenseIntegrity(t *testing.T) {
	ctx := context.Background()
	tenantID := int64(123)
	privateKey, publicKey := GenerateTestKeyPair(t)
	pubKeyBase64 := GetTestPublicKey(publicKey)

	t.Run("valid integrity - stored values match decoded license key", func(t *testing.T) {
		mockRepo := new(MockLicenseRepository)
		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{}, nil)
		service, _ := NewLicenseService(mockRepo, pubKeyBase64)

		// Generate a valid commercial license
		licensePEM := GenerateTestLicense(t, privateKey, "tenant-123", "commercial", 100, -1, 90, 365)
		storedLicense := &models.License{
			TenantID:          tenantID,
			LicenseID:         "tenant-123",
			LicenseType:       "commercial",
			MaxSeats:          100,
			MaxTeams:          -1,
			DataRetentionDays: 90,
			LicenseKey:        licensePEM,
		}

		mockRepo.On("GetTenantLicense", ctx, tenantID).Return(storedLicense, nil)

		valid, mismatch, err := service.VerifyLicenseIntegrity(ctx, tenantID)

		assert.NoError(t, err)
		assert.True(t, valid)
		assert.Empty(t, mismatch)
		mockRepo.AssertExpectations(t)
	})

	t.Run("tier tampering - stored tier doesn't match (auto-corrected)", func(t *testing.T) {
		mockRepo := new(MockLicenseRepository)
		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{}, nil)
		service, _ := NewLicenseService(mockRepo, pubKeyBase64)

		// Generate a commercial license
		licensePEM := GenerateTestLicense(t, privateKey, "tenant-123", "commercial", 100, -1, 90, 365)
		storedLicense := &models.License{
			TenantID:          tenantID,
			LicenseID:         "tenant-123",
			LicenseType:       "opensource", // Tampered: stored as opensource instead of commercial
			MaxSeats:          100,
			MaxTeams:          -1,
			DataRetentionDays: 90,
			LicenseKey:        licensePEM,
		}

		mockRepo.On("GetTenantLicense", ctx, tenantID).Return(storedLicense, nil)
		// Should auto-correct by updating with the correct values from the license key
		mockRepo.On("UpdateTenantLicense", ctx, tenantID, mock.MatchedBy(func(l *models.License) bool {
			return l.LicenseType == "commercial" && l.MaxSeats == 100 && l.LicenseID == "tenant-123"
		})).Return(nil)

		valid, mismatch, err := service.VerifyLicenseIntegrity(ctx, tenantID)

		assert.NoError(t, err)
		assert.True(t, valid) // Now valid after auto-correction
		assert.Contains(t, mismatch, "license_type mismatch: stored=opensource, actual=commercial")
		assert.Contains(t, mismatch, "(auto-corrected)")
		mockRepo.AssertExpectations(t)
	})

	t.Run("seats tampering - stored seats don't match (auto-corrected)", func(t *testing.T) {
		mockRepo := new(MockLicenseRepository)
		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{}, nil)
		service, _ := NewLicenseService(mockRepo, pubKeyBase64)

		// Generate a commercial license with 100 seats
		licensePEM := GenerateTestLicense(t, privateKey, "tenant-123", "commercial", 100, -1, 90, 365)
		storedLicense := &models.License{
			TenantID:          tenantID,
			LicenseID:         "tenant-123",
			LicenseType:       "commercial",
			MaxSeats:          50, // Tampered: stored as 50 instead of 100
			MaxTeams:          -1,
			DataRetentionDays: 90,
			LicenseKey:        licensePEM,
		}

		mockRepo.On("GetTenantLicense", ctx, tenantID).Return(storedLicense, nil)
		// Should auto-correct by updating with the correct values from the license key
		mockRepo.On("UpdateTenantLicense", ctx, tenantID, mock.MatchedBy(func(l *models.License) bool {
			return l.LicenseType == "commercial" && l.MaxSeats == 100 && l.LicenseID == "tenant-123"
		})).Return(nil)

		valid, mismatch, err := service.VerifyLicenseIntegrity(ctx, tenantID)

		assert.NoError(t, err)
		assert.True(t, valid) // Now valid after auto-correction
		assert.NotContains(t, mismatch, "license_type mismatch")
		assert.Contains(t, mismatch, "max_seats mismatch: stored=50, actual=100")
		assert.Contains(t, mismatch, "(auto-corrected)")
		mockRepo.AssertExpectations(t)
	})

	t.Run("both tampered - type and seats don't match (auto-corrected)", func(t *testing.T) {
		mockRepo := new(MockLicenseRepository)
		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{}, nil)
		service, _ := NewLicenseService(mockRepo, pubKeyBase64)

		// Generate a commercial license
		licensePEM := GenerateTestLicense(t, privateKey, "tenant-123", "commercial", 100, -1, 90, 365)
		storedLicense := &models.License{
			TenantID:          tenantID,
			LicenseID:         "tenant-123",
			LicenseType:       "opensource", // Tampered
			MaxSeats:          50,           // Tampered
			MaxTeams:          1,            // Tampered
			DataRetentionDays: 7,            // Tampered
			LicenseKey:        licensePEM,
		}

		mockRepo.On("GetTenantLicense", ctx, tenantID).Return(storedLicense, nil)
		// Should auto-correct by updating with the correct values from the license key
		mockRepo.On("UpdateTenantLicense", ctx, tenantID, mock.MatchedBy(func(l *models.License) bool {
			return l.LicenseType == "commercial" && l.MaxSeats == 100 && l.LicenseID == "tenant-123"
		})).Return(nil)

		valid, mismatch, err := service.VerifyLicenseIntegrity(ctx, tenantID)

		assert.NoError(t, err)
		assert.True(t, valid) // Now valid after auto-correction
		assert.Contains(t, mismatch, "license_type mismatch: stored=opensource, actual=commercial")
		assert.Contains(t, mismatch, "max_seats mismatch: stored=50, actual=100")
		assert.Contains(t, mismatch, "(auto-corrected)")
		mockRepo.AssertExpectations(t)
	})

	t.Run("no license - tenant has no license (return valid, no mismatch)", func(t *testing.T) {
		mockRepo := new(MockLicenseRepository)
		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{}, nil)
		service, _ := NewLicenseService(mockRepo, pubKeyBase64)

		// Simulate "no rows in result set" error
		mockRepo.On("GetTenantLicense", ctx, tenantID).Return(nil, errors.New("sql: no rows in result set"))

		valid, mismatch, err := service.VerifyLicenseIntegrity(ctx, tenantID)

		assert.NoError(t, err)
		assert.True(t, valid)
		assert.Empty(t, mismatch)
		mockRepo.AssertExpectations(t)
	})

	t.Run("no license key - empty license key field", func(t *testing.T) {
		mockRepo := new(MockLicenseRepository)
		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{}, nil)
		service, _ := NewLicenseService(mockRepo, pubKeyBase64)

		storedLicense := &models.License{
			TenantID:   tenantID,
			LicenseID:  "",
			Tier:       "",
			Seats:      0,
			LicenseKey: "", // No license key
		}

		mockRepo.On("GetTenantLicense", ctx, tenantID).Return(storedLicense, nil)

		valid, mismatch, err := service.VerifyLicenseIntegrity(ctx, tenantID)

		assert.NoError(t, err)
		assert.True(t, valid)
		assert.Empty(t, mismatch)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid license key - corrupted PEM format (return error)", func(t *testing.T) {
		mockRepo := new(MockLicenseRepository)
		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{}, nil)
		service, _ := NewLicenseService(mockRepo, pubKeyBase64)

		storedLicense := &models.License{
			TenantID:   tenantID,
			LicenseID:  "tenant-123",
			Tier:       "2.0",
			Seats:      25,
			LicenseKey: "-----BEGIN LICENSE-----\nINVALID_BASE64_DATA!!!\n-----END LICENSE-----",
		}

		mockRepo.On("GetTenantLicense", ctx, tenantID).Return(storedLicense, nil)

		valid, mismatch, err := service.VerifyLicenseIntegrity(ctx, tenantID)

		assert.Error(t, err)
		assert.False(t, valid)
		assert.Empty(t, mismatch)
		assert.Contains(t, err.Error(), "invalid license key format")
		mockRepo.AssertExpectations(t)
	})

	t.Run("expired license - still verify integrity even if expired", func(t *testing.T) {
		mockRepo := new(MockLicenseRepository)
		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{}, nil)
		service, _ := NewLicenseService(mockRepo, pubKeyBase64)

		// Generate an expired license
		licenseData := models.LicenseData{
			LicenseType:       "commercial",
			MaxSeats:          100,
			MaxTeams:          -1,
			DataRetentionDays: 90,
		}
		dataBytes, _ := json.Marshal(licenseData)

		// Issue license 10 days ago, expire 5 days ago
		expiredLicense := &licenseLib.License{
			ID:        "tenant-123",
			Customer:  "Test Customer tenant-123",
			Type:      "commercial",
			IssuedAt:  time.Now().Add(-10 * 24 * time.Hour).Unix(),
			ExpiredAt: time.Now().Add(-5 * 24 * time.Hour).Unix(),
			Data:      dataBytes,
		}

		encoded, err := expiredLicense.Encode(privateKey)
		require.NoError(t, err)

		storedLicense := &models.License{
			TenantID:          tenantID,
			LicenseID:         "tenant-123",
			LicenseType:       "commercial",
			MaxSeats:          100,
			MaxTeams:          -1,
			DataRetentionDays: 90,
			LicenseKey:        string(encoded),
		}

		mockRepo.On("GetTenantLicense", ctx, tenantID).Return(storedLicense, nil)

		valid, mismatch, err := service.VerifyLicenseIntegrity(ctx, tenantID)

		// Integrity check should succeed even if license is expired
		assert.NoError(t, err)
		assert.True(t, valid)
		assert.Empty(t, mismatch)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error - database connection fails", func(t *testing.T) {
		mockRepo := new(MockLicenseRepository)
		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{}, nil)
		service, _ := NewLicenseService(mockRepo, pubKeyBase64)

		expectedError := errors.New("database connection failed")
		mockRepo.On("GetTenantLicense", ctx, tenantID).Return(nil, expectedError)

		valid, mismatch, err := service.VerifyLicenseIntegrity(ctx, tenantID)

		assert.Error(t, err)
		assert.False(t, valid)
		assert.Empty(t, mismatch)
		assert.Contains(t, err.Error(), "failed to retrieve license")
		mockRepo.AssertExpectations(t)
	})

	t.Run("auto-correction failure - update database fails", func(t *testing.T) {
		mockRepo := new(MockLicenseRepository)
		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{}, nil)
		service, _ := NewLicenseService(mockRepo, pubKeyBase64)

		// Generate a commercial license
		licensePEM := GenerateTestLicense(t, privateKey, "tenant-123", "commercial", 100, -1, 90, 365)
		storedLicense := &models.License{
			TenantID:          tenantID,
			LicenseID:         "tenant-123",
			LicenseType:       "opensource", // Tampered
			MaxSeats:          100,
			MaxTeams:          -1,
			DataRetentionDays: 90,
			LicenseKey:        licensePEM,
		}

		mockRepo.On("GetTenantLicense", ctx, tenantID).Return(storedLicense, nil)
		// Auto-correction fails
		expectedError := errors.New("database update failed")
		mockRepo.On("UpdateTenantLicense", ctx, tenantID, mock.AnythingOfType("*models.License")).Return(expectedError)

		valid, mismatch, err := service.VerifyLicenseIntegrity(ctx, tenantID)

		assert.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, mismatch, "license_type mismatch: stored=opensource, actual=commercial")
		assert.Contains(t, err.Error(), "failed to auto-correct license")
		mockRepo.AssertExpectations(t)
	})

	t.Run("expired license with tampering - auto-corrects even if expired", func(t *testing.T) {
		mockRepo := new(MockLicenseRepository)
		mockRepo.On("GetAllLicenses", mock.Anything).Return([]*models.License{}, nil)
		service, _ := NewLicenseService(mockRepo, pubKeyBase64)

		// Generate an expired commercial license
		licenseData := models.LicenseData{
			LicenseType:       "commercial",
			MaxSeats:          100,
			MaxTeams:          -1,
			DataRetentionDays: 90,
		}
		dataBytes, _ := json.Marshal(licenseData)

		// Issue license 10 days ago, expire 5 days ago
		expiredLicense := &licenseLib.License{
			ID:        "tenant-123",
			Customer:  "Test Customer tenant-123",
			Type:      "commercial",
			IssuedAt:  time.Now().Add(-10 * 24 * time.Hour).Unix(),
			ExpiredAt: time.Now().Add(-5 * 24 * time.Hour).Unix(),
			Data:      dataBytes,
		}

		encoded, err := expiredLicense.Encode(privateKey)
		require.NoError(t, err)

		storedLicense := &models.License{
			TenantID:          tenantID,
			LicenseID:         "tenant-123",
			LicenseType:       "opensource", // Tampered - should be commercial
			MaxSeats:          10,           // Tampered
			MaxTeams:          1,            // Tampered
			DataRetentionDays: 7,            // Tampered
			LicenseKey:        string(encoded),
		}

		mockRepo.On("GetTenantLicense", ctx, tenantID).Return(storedLicense, nil)
		// Should auto-correct even though license is expired
		mockRepo.On("UpdateTenantLicense", ctx, tenantID, mock.MatchedBy(func(l *models.License) bool {
			return l.LicenseType == "commercial" && l.MaxSeats == 100 && l.LicenseID == "tenant-123"
		})).Return(nil)

		valid, mismatch, err := service.VerifyLicenseIntegrity(ctx, tenantID)

		// Should auto-correct even if expired
		assert.NoError(t, err)
		assert.True(t, valid)
		assert.Contains(t, mismatch, "license_type mismatch: stored=opensource, actual=commercial")
		assert.Contains(t, mismatch, "(auto-corrected)")
		mockRepo.AssertExpectations(t)
	})
}
