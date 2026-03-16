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

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv_WithValue(t *testing.T) {
	// Set environment variable
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	result := getEnv("TEST_VAR", "default")
	assert.Equal(t, "test_value", result)
}

func TestGetEnv_WithDefault(t *testing.T) {
	// Unset environment variable to ensure default is used
	os.Unsetenv("NONEXISTENT_VAR")

	result := getEnv("NONEXISTENT_VAR", "default_value")
	assert.Equal(t, "default_value", result)
}

func TestGetEnv_EmptyString(t *testing.T) {
	// Set environment variable to empty string
	os.Setenv("EMPTY_VAR", "")
	defer os.Unsetenv("EMPTY_VAR")

	result := getEnv("EMPTY_VAR", "default")
	// Empty string should return default
	assert.Equal(t, "default", result)
}

func TestLoadConfig_DefaultValues(t *testing.T) {
	// Ensure environment variables are not set
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("DEBUG")

	cfg := LoadConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, 9080, cfg.Port)
	assert.Equal(t, "default_secret_key_for_development", cfg.JWTSecret)
	assert.Equal(t, "postgres://localhost:5432/codetogether?sslmode=disable", cfg.DatabaseURL)
	assert.False(t, cfg.Debug)
}

func TestLoadConfig_WithEnvVariables(t *testing.T) {
	os.Setenv("PORT", "3000")
	os.Setenv("DATABASE_URL", "postgres://test:1234@localhost:5432/testdb")
	os.Setenv("JWT_SECRET", "test_secret")
	os.Setenv("DEBUG", "true")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("DEBUG")
	}()

	cfg := LoadConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, 3000, cfg.Port)
	assert.Equal(t, "postgres://test:1234@localhost:5432/testdb", cfg.DatabaseURL)
	assert.Equal(t, "test_secret", cfg.JWTSecret)
	assert.True(t, cfg.Debug)
}

func TestLoadConfig_InvalidPort(t *testing.T) {
	os.Setenv("PORT", "invalid")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("DEBUG")
	defer os.Unsetenv("PORT")

	cfg := LoadConfig()

	// Invalid port should fall back to default
	assert.Equal(t, 9080, cfg.Port)
}

func TestLoadConfig_DebugTrueVariations(t *testing.T) {
	debugValues := []string{"true"}

	for _, val := range debugValues {
		t.Run("debug_"+val, func(t *testing.T) {
			os.Setenv("DEBUG", val)
			defer os.Unsetenv("DEBUG")

			cfg := LoadConfig()
			assert.True(t, cfg.Debug)
		})
	}
}

func TestLoadConfig_DebugFalseValues(t *testing.T) {
	debugValues := []string{"false", "FALSE", "False", "no", "0", ""}

	for _, val := range debugValues {
		t.Run("debug_"+val, func(t *testing.T) {
			os.Setenv("DEBUG", val)
			defer os.Unsetenv("DEBUG")

			cfg := LoadConfig()
			assert.False(t, cfg.Debug)
		})
	}
}

func TestConfig_Structure(t *testing.T) {
	cfg := &Config{
		Port:        8080,
		JWTSecret:   "my_secret",
		DatabaseURL: "postgres://localhost/mydb",
		Debug:       true,
	}

	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "my_secret", cfg.JWTSecret)
	assert.Equal(t, "postgres://localhost/mydb", cfg.DatabaseURL)
	assert.True(t, cfg.Debug)
}

// TestLoadConfig_CustomEnvVars tests loading config with custom environment variables
func TestLoadConfig_CustomEnvVars(t *testing.T) {
	// Set custom environment variables
	os.Setenv("PORT", "8080")
	os.Setenv("JWT_SECRET", "custom_secret")
	os.Setenv("DATABASE_URL", "postgres://custom:5432/test")
	os.Setenv("DEBUG", "true")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("DEBUG")
	}()

	cfg := LoadConfig()

	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "custom_secret", cfg.JWTSecret)
	assert.Equal(t, "postgres://custom:5432/test", cfg.DatabaseURL)
	assert.True(t, cfg.Debug)
}

// TestLoadConfig_LicensePublicKeyDefault tests license public key default (production mode)
func TestLoadConfig_LicensePublicKeyDefault(t *testing.T) {
	// Ensure debug mode is off
	os.Unsetenv("DEBUG")
	defer os.Setenv("DEBUG", "")

	cfg := LoadConfig()

	assert.NotNil(t, cfg.LicensePublicKey)
	assert.Equal(t, ProductionLicensePublicKey, cfg.LicensePublicKey, "Should use hardcoded production key in non-debug mode")
}

// TestLoadConfig_LicensePublicKeyCustomDebug tests custom license key in debug mode
func TestLoadConfig_LicensePublicKeyCustomDebug(t *testing.T) {
	// Enable debug mode and set custom key
	os.Setenv("DEBUG", "true")
	os.Setenv("LICENSE_PUBLIC_KEY", "custom_debug_key_123")

	defer func() {
		os.Unsetenv("DEBUG")
		os.Unsetenv("LICENSE_PUBLIC_KEY")
	}()

	cfg := LoadConfig()

	assert.NotNil(t, cfg.LicensePublicKey)
	assert.Equal(t, "custom_debug_key_123", cfg.LicensePublicKey, "Should use custom key in debug mode")
}

// TestLoadConfig_LicensePublicKeyCustomRelease tests custom license key ignored in release mode
func TestLoadConfig_LicensePublicKeyCustomRelease(t *testing.T) {
	// Disable debug mode but set custom key (should be ignored)
	os.Unsetenv("DEBUG")
	os.Setenv("LICENSE_PUBLIC_KEY", "custom_key_ignored")

	defer func() {
		os.Setenv("DEBUG", "")
		os.Unsetenv("LICENSE_PUBLIC_KEY")
	}()

	cfg := LoadConfig()

	assert.NotNil(t, cfg.LicensePublicKey)
	assert.Equal(t, ProductionLicensePublicKey, cfg.LicensePublicKey, "Should ignore custom key in release mode")
}
