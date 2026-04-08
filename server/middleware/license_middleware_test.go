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

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"switch-server/models"
)

// MockLicenseServiceForMiddleware is a mock for LicenseServiceInterface
type MockLicenseServiceForMiddleware struct {
	mock.Mock
}

func (m *MockLicenseServiceForMiddleware) GetLicense(ctx context.Context, tenantID int64) (*models.License, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.License), args.Error(1)
}

func (m *MockLicenseServiceForMiddleware) GetLicenseUsage(ctx context.Context, tenantID int64) (*models.LicenseUsage, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LicenseUsage), args.Error(1)
}

func (m *MockLicenseServiceForMiddleware) CanCreateTeam(ctx context.Context, tenantID int64) (bool, error) {
	args := m.Called(ctx, tenantID)
	return args.Bool(0), args.Error(1)
}

func (m *MockLicenseServiceForMiddleware) IsLicenseValid(ctx context.Context, tenantID int64) (bool, string, error) {
	args := m.Called(ctx, tenantID)
	return args.Bool(0), args.String(1), args.Error(2)
}

func (m *MockLicenseServiceForMiddleware) ActivateLicense(ctx context.Context, tenantID int64, licenseKeyPEM string) error {
	args := m.Called(ctx, tenantID, licenseKeyPEM)
	return args.Error(0)
}

func (m *MockLicenseServiceForMiddleware) GetEffectiveLicense(ctx context.Context, tenantID int64) *models.License {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.License)
}

func (m *MockLicenseServiceForMiddleware) HasActiveLicense(ctx context.Context, tenantID int64) bool {
	args := m.Called(ctx, tenantID)
	return args.Bool(0)
}

func (m *MockLicenseServiceForMiddleware) GetTiers() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockLicenseServiceForMiddleware) VerifyLicenseIntegrity(ctx context.Context, tenantID int64) (bool, string, error) {
	args := m.Called(ctx, tenantID)
	return args.Bool(0), args.String(1), args.Error(2)
}

func TestLicenseMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func(*gin.Context)
		mockLicense   *models.License
		mockHasActive bool
		verifyContext func(*testing.T, *gin.Context)
	}{
		{
			name: "user in context - sets license and has_active_license",
			setupContext: func(c *gin.Context) {
				user := &models.User{
					ID:       1,
					Email:    "test@example.com",
					Role:     "manager",
					TenantID: 123,
				}
				c.Set("user", user)
			},
			mockLicense: &models.License{
				TenantID:  123,
				LicenseID: "lic-123",
				Tier:      "2.0",
				Seats:     25,
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
			},
			mockHasActive: true,
			verifyContext: func(t *testing.T, c *gin.Context) {
				license, exists := c.Get("license")
				require.True(t, exists, "license should be set in context")
				require.NotNil(t, license)

				lic := license.(*models.License)
				assert.Equal(t, int64(123), lic.TenantID)
				assert.Equal(t, "lic-123", lic.LicenseID)
				assert.Equal(t, "2.0", lic.Tier)
				assert.Equal(t, 25, lic.Seats)

				hasActive, exists := c.Get("has_active_license")
				require.True(t, exists, "has_active_license should be set")
				assert.True(t, hasActive.(bool))
			},
		},
		{
			name: "user in context - default license (no active license)",
			setupContext: func(c *gin.Context) {
				user := &models.User{
					ID:       1,
					Email:    "test@example.com",
					Role:     "manager",
					TenantID: 123,
				}
				c.Set("user", user)
			},
			mockLicense: &models.License{
				TenantID: 123,
				Tier:     "0.0", // Default tier
				Seats:    3,     // Default seats
			},
			mockHasActive: false,
			verifyContext: func(t *testing.T, c *gin.Context) {
				license, exists := c.Get("license")
				require.True(t, exists, "license should be set in context")
				require.NotNil(t, license)

				lic := license.(*models.License)
				assert.Equal(t, "0.0", lic.Tier)
				assert.Equal(t, 3, lic.Seats)

				hasActive, exists := c.Get("has_active_license")
				require.True(t, exists, "has_active_license should be set")
				assert.False(t, hasActive.(bool))
			},
		},
		{
			name: "user not in context - continues without setting license",
			setupContext: func(c *gin.Context) {
				// Don't set user
			},
			mockLicense:   nil,
			mockHasActive: false,
			verifyContext: func(t *testing.T, c *gin.Context) {
				license, exists := c.Get("license")
				assert.False(t, exists, "license should not be set when no user in context")
				assert.Nil(t, license)

				hasActive, exists := c.Get("has_active_license")
				assert.False(t, exists, "has_active_license should not be set when no user in context")
				assert.Nil(t, hasActive)
			},
		},
		{
			name: "invalid user type - continues without setting license",
			setupContext: func(c *gin.Context) {
				c.Set("user", "not a user object")
			},
			mockLicense:   nil,
			mockHasActive: false,
			verifyContext: func(t *testing.T, c *gin.Context) {
				license, exists := c.Get("license")
				assert.False(t, exists, "license should not be set for invalid user type")
				assert.Nil(t, license)

				hasActive, exists := c.Get("has_active_license")
				assert.False(t, exists, "has_active_license should not be set for invalid user type")
				assert.Nil(t, hasActive)
			},
		},
		{
			name: "service returns nil license - continues without crashing",
			setupContext: func(c *gin.Context) {
				user := &models.User{
					ID:       1,
					Email:    "test@example.com",
					Role:     "manager",
					TenantID: 123,
				}
				c.Set("user", user)
			},
			mockLicense:   nil,
			mockHasActive: false,
			verifyContext: func(t *testing.T, c *gin.Context) {
				// Should set license even if nil (from service)
				license, exists := c.Get("license")
				require.True(t, exists, "license key should exist in context")
				assert.Nil(t, license, "license value should be nil when service returns nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			mockService := new(MockLicenseServiceForMiddleware)
			if tt.mockLicense != nil || tt.setupContext == nil || tt.setupContext != nil {
				// Only set up mock if we expect it to be called
				user, _ := getTestUserFromSetup(tt.setupContext)
				if user != nil {
					mockService.On("GetEffectiveLicense", mock.Anything, user.TenantID).Return(tt.mockLicense)
					mockService.On("HasActiveLicense", mock.Anything, user.TenantID).Return(tt.mockHasActive)
				}
			}

			router := gin.New()
			router.Use(LicenseMiddleware(mockService))

			// Test handler that verifies context and calls Next
			testHandler := func(c *gin.Context) {
				if tt.verifyContext != nil {
					tt.verifyContext(t, c)
				}
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			}

			router.GET("/test", testHandler)

			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			// Set up context before middleware if needed
			if tt.setupContext != nil {
				// We'll set up the context through the router, so this is handled in the handler
				// Actually, we need to set it before the middleware runs
				// Let's modify the approach
			}

			// Better approach: create a custom handler to set up context
			setupHandler := func(c *gin.Context) {
				if tt.setupContext != nil {
					tt.setupContext(c)
				}
				c.Next()
			}

			routerWithSetup := gin.New()
			routerWithSetup.Use(setupHandler)
			routerWithSetup.Use(LicenseMiddleware(mockService))
			routerWithSetup.GET("/test", testHandler)

			routerWithSetup.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "request should succeed (middleware is non-blocking)")
			mockService.AssertExpectations(t)
		})
	}
}

// Helper to extract user from setup function
func getTestUserFromSetup(setupFunc func(*gin.Context)) (*models.User, bool) {
	if setupFunc == nil {
		return nil, false
	}

	// Create a test context and run the setup
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	setupFunc(c)

	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*models.User); ok {
			return u, true
		}
	}
	return nil, false
}

func TestLicenseMiddleware_NonBlocking(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                  string
		setupContext          func(*gin.Context)
		setupMockExpectations func(*MockLicenseServiceForMiddleware)
	}{
		{
			name: "continues when user not found",
			setupContext: func(c *gin.Context) {
				// Don't set user
			},
			setupMockExpectations: func(m *MockLicenseServiceForMiddleware) {
				// No expectations - service shouldn't be called
			},
		},
		{
			name: "continues when invalid user type",
			setupContext: func(c *gin.Context) {
				c.Set("user", "invalid")
			},
			setupMockExpectations: func(m *MockLicenseServiceForMiddleware) {
				// No expectations - service shouldn't be called
			},
		},
		{
			name: "continues when valid user",
			setupContext: func(c *gin.Context) {
				user := &models.User{
					ID:       1,
					Email:    "test@example.com",
					TenantID: 123,
				}
				c.Set("user", user)
			},
			setupMockExpectations: func(m *MockLicenseServiceForMiddleware) {
				license := &models.License{
					TenantID: 123,
					Tier:     "1.0",
					Seats:    10,
				}
				m.On("GetEffectiveLicense", mock.Anything, int64(123)).Return(license)
				m.On("HasActiveLicense", mock.Anything, int64(123)).Return(true)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockLicenseServiceForMiddleware)
			tt.setupMockExpectations(mockService)

			router := gin.New()

			// Setup handler to set context
			setupHandler := func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			}

			// Test handler that should always be called
			handlerCalled := false
			testHandler := func(c *gin.Context) {
				handlerCalled = true
				c.JSON(http.StatusOK, gin.H{"called": true})
			}

			router.Use(setupHandler)
			router.Use(LicenseMiddleware(mockService))
			router.GET("/test", testHandler)

			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.True(t, handlerCalled, "test handler should always be called (non-blocking)")
			assert.Equal(t, http.StatusOK, w.Code, "request should succeed")
			mockService.AssertExpectations(t)
		})
	}
}

func TestLicenseMiddleware_WithRealUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockLicenseServiceForMiddleware)
	user := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     "manager",
		TenantID: 456,
	}

	activeLicense := &models.License{
		TenantID:  456,
		LicenseID: "lic-active-456",
		Tier:      "2.0",
		Seats:     50,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
	}

	mockService.On("GetEffectiveLicense", mock.Anything, int64(456)).Return(activeLicense)
	mockService.On("HasActiveLicense", mock.Anything, int64(456)).Return(true)

	router := gin.New()

	// Handler to set user in context
	setupHandler := func(c *gin.Context) {
		c.Set("user", user)
		c.Next()
	}

	// Test handler to verify license is available
	testHandler := func(c *gin.Context) {
		license := c.MustGet("license")
		hasActive := c.MustGet("has_active_license")

		lic := license.(*models.License)
		c.JSON(http.StatusOK, gin.H{
			"license_id":         lic.LicenseID,
			"tier":               lic.Tier,
			"seats":              lic.Seats,
			"has_active_license": hasActive,
		})
	}

	router.Use(setupHandler)
	router.Use(LicenseMiddleware(mockService))
	router.GET("/test", testHandler)

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "lic-active-456", response["license_id"])
	assert.Equal(t, "2.0", response["tier"])
	assert.Equal(t, float64(50), response["seats"])
	assert.True(t, response["has_active_license"].(bool), "has_active_license should be true")

	mockService.AssertExpectations(t)
}

func TestLicenseMiddleware_DifferentTenants(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tenantIDs := []int64{1, 123, 456, 999}

	for _, tenantID := range tenantIDs {
		t.Run("tenant_"+string(rune(tenantID)), func(t *testing.T) {
			mockService := new(MockLicenseServiceForMiddleware)
			user := &models.User{
				ID:       1,
				Email:    "test@example.com",
				TenantID: tenantID,
			}

			license := &models.License{
				TenantID: tenantID,
				Tier:     "1.0",
				Seats:    10,
			}

			mockService.On("GetEffectiveLicense", mock.Anything, tenantID).Return(license)
			mockService.On("HasActiveLicense", mock.Anything, tenantID).Return(true)

			router := gin.New()

			setupHandler := func(c *gin.Context) {
				c.Set("user", user)
				c.Next()
			}

			testHandler := func(c *gin.Context) {
				license := c.MustGet("license")
				lic := license.(*models.License)
				assert.Equal(t, tenantID, lic.TenantID)
				c.JSON(http.StatusOK, gin.H{"tenant_id": tenantID})
			}

			router.Use(setupHandler)
			router.Use(LicenseMiddleware(mockService))
			router.GET("/test", testHandler)

			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestLicenseMiddleware_ContextIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockLicenseServiceForMiddleware)

	router := gin.New()

	// Handler to set user in context
	setupHandler := func(c *gin.Context) {
		user := &models.User{
			ID:       1,
			Email:    "test@example.com",
			TenantID: 123,
		}
		c.Set("user", user)
		c.Next()
	}

	license := &models.License{
		TenantID: 123,
		Tier:     "2.0",
		Seats:    25,
	}

	mockService.On("GetEffectiveLicense", mock.Anything, int64(123)).Return(license)
	mockService.On("HasActiveLicense", mock.Anything, int64(123)).Return(true)

	router.Use(setupHandler)
	router.Use(LicenseMiddleware(mockService))

	// Multiple test endpoints
	testHandler1 := func(c *gin.Context) {
		license := c.MustGet("license")
		lic := license.(*models.License)
		c.JSON(http.StatusOK, gin.H{"endpoint": 1, "tier": lic.Tier})
	}

	testHandler2 := func(c *gin.Context) {
		license := c.MustGet("license")
		lic := license.(*models.License)
		c.JSON(http.StatusOK, gin.H{"endpoint": 2, "seats": lic.Seats})
	}

	router.GET("/test1", testHandler1)
	router.GET("/test2", testHandler2)

	// Test first endpoint
	req1, _ := http.NewRequest("GET", "/test1", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)

	var response1 map[string]interface{}
	json.Unmarshal(w1.Body.Bytes(), &response1)
	assert.Equal(t, float64(1), response1["endpoint"])
	assert.Equal(t, "2.0", response1["tier"])

	// Test second endpoint
	req2, _ := http.NewRequest("GET", "/test2", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var response2 map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &response2)
	assert.Equal(t, float64(2), response2["endpoint"])
	assert.Equal(t, float64(25), response2["seats"])

	mockService.AssertExpectations(t)
}

func TestRequireCommercial_ActiveLicense(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("has_active_license", true)
		c.Next()
	})
	router.Use(RequireCommercial())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireCommercial_NoLicense(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("has_active_license", false)
		c.Next()
	})
	router.Use(RequireCommercial())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "license_required", response["code"])
}

func TestRequireCommercial_NoFlagSet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	// Don't set has_active_license at all (missing flag)
	router.Use(RequireCommercial())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
