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

package main

import (
	"os"
	"path/filepath"
	"testing"

	"codeswitch/services"
)

// TestIsValidTool tests the tool name validation
func TestIsValidTool(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		want     bool
	}{
		{"valid claude", "claude", true},
		{"valid codex", "codex", true},
		{"valid opencode", "opencode", true},
		{"invalid tool", "invalid", false},
		{"empty tool", "", false},
		{"uppercase CLAUDE", "CLAUDE", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidTool(tt.toolName)
			if got != tt.want {
				t.Errorf("isValidTool(%q) = %v, want %v", tt.toolName, got, tt.want)
			}
		})
	}
}

// TestResolveConfig_AuthServerPairing tests that auth and server must be specified together
func TestResolveConfig_AuthServerPairing(t *testing.T) {
	tests := []struct {
		name      string
		cliAuth   string
		cliServer string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "both provided - valid",
			cliAuth:   "test_token",
			cliServer: "http://localhost:9080",
			wantErr:   false,
		},
		{
			name:      "neither provided - will load from settings",
			cliAuth:   "",
			cliServer: "",
			wantErr:   false, // If settings exist, will succeed
		},
		{
			name:      "only auth provided - invalid",
			cliAuth:   "test_token",
			cliServer: "",
			wantErr:   true,
			errMsg:    "both -auth and -server must be specified together",
		},
		{
			name:      "only server provided - invalid",
			cliAuth:   "",
			cliServer: "http://localhost:9080",
			wantErr:   true,
			errMsg:    "both -auth and -server must be specified together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock settings files for the "neither provided" test case
			if tt.name == "neither provided - will load from settings" {
				// Use t.TempDir() and set HOME to it for isolated test environment
				tmpDir := t.TempDir()
				oldHome := os.Getenv("HOME")
				os.Setenv("HOME", tmpDir)
				t.Cleanup(func() {
					os.Setenv("HOME", oldHome)
				})

				// Create .code-together directory in temp directory
				configDir := filepath.Join(tmpDir, ".code-together")
				if err := os.MkdirAll(configDir, 0o755); err != nil {
					t.Fatalf("failed to create config directory: %v", err)
				}

				// Create mock config.json
				configContent := `{
  "app": {
    "show_heatmap": false,
    "show_home_title": true,
    "show_24h_stats": false,
    "auto_start": false
  },
  "server": {
    "server_url": "http://localhost:8080"
  },
  "sync": {
    "last_sync_time": "2026-03-16T00:00:00Z"
  }
}`
				configPath := filepath.Join(configDir, "config.json")
				if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
					t.Fatalf("failed to create mock config.json: %v", err)
				}

				// Create mock auth.json
				authContent := `{
  "tokens": {
    "access_token": "test_mock_token_for_cli_testing",
    "refresh_token": "test_refresh_token",
    "expires_at": "2026-12-31T23:59:59Z"
  }
}`
				authPath := filepath.Join(configDir, "auth.json")
				if err := os.WriteFile(authPath, []byte(authContent), 0o600); err != nil {
					t.Fatalf("failed to create mock auth.json: %v", err)
				}
			}

			_, err := resolveConfig("claude", tt.cliAuth, tt.cliServer, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Only check error message if we expect an error and have one
			if tt.wantErr && err != nil && tt.errMsg != "" {
				errMsg := err.Error()
				if len(errMsg) < len(tt.errMsg) || errMsg[:len(tt.errMsg)] != tt.errMsg {
					t.Errorf("resolveConfig() error = %q, want prefix %q", errMsg, tt.errMsg)
				}
			}
		})
	}
}

// TestResolveConfig_InvalidTool tests that invalid tool names are rejected
func TestResolveConfig_InvalidTool(t *testing.T) {
	invalidTools := []string{"invalid", "test", "", "CLAUDE", "Codex"}

	for _, tool := range invalidTools {
		t.Run(tool, func(t *testing.T) {
			_, err := resolveConfig(tool, "token", "http://localhost:9080", false)
			if err == nil {
				t.Errorf("resolveConfig() with invalid tool %q should return error", tool)
			}
			if errMsg := err.Error(); errMsg[:len("invalid tool name")] != "invalid tool name" {
				t.Errorf("resolveConfig() error = %q, want prefix 'invalid tool name'", errMsg)
			}
		})
	}
}

// TestGetSettingsService tests the settings service factory
// Note: InMemoryProviderService tests are in providers_test.go
func TestGetSettingsService(t *testing.T) {
	proxyURL := "http://localhost:18100"

	t.Run("claude settings", func(t *testing.T) {
		service := getSettingsService("claude", proxyURL)
		if service == nil {
			t.Fatal("getSettingsService() returned nil for claude")
		}
		_, ok := service.(*claudeSettingsWrapper)
		if !ok {
			t.Error("getSettingsService() did not return claudeSettingsWrapper for claude")
		}
	})

	t.Run("codex settings", func(t *testing.T) {
		service := getSettingsService("codex", proxyURL)
		if service == nil {
			t.Fatal("getSettingsService() returned nil for codex")
		}
		_, ok := service.(*codexSettingsWrapper)
		if !ok {
			t.Error("getSettingsService() did not return codexSettingsWrapper for codex")
		}
	})

	t.Run("opencode settings", func(t *testing.T) {
		service := getSettingsService("opencode", proxyURL)
		if service == nil {
			t.Fatal("getSettingsService() returned nil for opencode")
		}
		_, ok := service.(*opencodeSettingsWrapper)
		if !ok {
			t.Error("getSettingsService() did not return opencodeSettingsWrapper for opencode")
		}
	})

	t.Run("invalid tool", func(t *testing.T) {
		service := getSettingsService("invalid", proxyURL)
		if service != nil {
			t.Error("getSettingsService() should return nil for invalid tool")
		}
	})
}

// TestSettingsWrappers tests the settings wrapper implementations
func TestSettingsWrappers(t *testing.T) {
	t.Run("claudeSettingsWrapper", func(t *testing.T) {
		wrapper := &claudeSettingsWrapper{
			service: services.NewClaudeSettingsService("http://localhost:18100"),
		}

		// Test methods don't panic
		_ = wrapper.IsProxyEnabled()
		_ = wrapper.EnableProxy()
		_ = wrapper.DisableProxy()
	})

	t.Run("codexSettingsWrapper", func(t *testing.T) {
		wrapper := &codexSettingsWrapper{
			service: services.NewCodexSettingsService("http://localhost:18100"),
		}

		// Test methods don't panic
		_ = wrapper.IsProxyEnabled()
		_ = wrapper.EnableProxy()
		_ = wrapper.DisableProxy()
	})

	t.Run("opencodeSettingsWrapper", func(t *testing.T) {
		wrapper := &opencodeSettingsWrapper{
			service: services.NewOpenCodeSettingsService("http://localhost:18100"),
		}

		// Test methods don't panic
		_ = wrapper.IsProxyEnabled()
		_ = wrapper.EnableProxy()
		_ = wrapper.DisableProxy()
	})
}

// TestSetupLogger tests that setupLogger doesn't panic
func TestSetupLogger(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
	}{
		{"normal logging", false},
		{"verbose logging", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just ensure it doesn't panic
			setupLogger(tt.verbose)
		})
	}
}

// Helper function to create a temporary directory for tests
func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "cli-test-")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

// Helper function to cleanup temp directory
func cleanupTempDir(t *testing.T, dir string) {
	t.Helper()
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
}

// Helper function to create test settings files
func createTestSettings(t *testing.T, dir string, filename string, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
