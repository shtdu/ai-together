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
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ==================== Helper Functions ====================

// setupTestEnv creates a temporary home directory for testing
// Follows the pattern from opencodesettings_test.go:211-215
func setupTestEnv(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	homeDir := os.Getenv("HOME")
	t.Cleanup(func() {
		os.Setenv("HOME", homeDir)
	})
	os.Setenv("HOME", tmpDir)
	return tmpDir
}

// createSettingsFile creates a test settings.json file with specified content
func createSettingsFile(t *testing.T, tmpDir, authToken, baseURL string) {
	t.Helper()
	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	settingsPath := filepath.Join(configDir, claudeSettingsFileName)
	settings := make(claudeSettingsFile)
	env := getEnv(settings)

	if authToken != "" || baseURL != "" {
		env["ANTHROPIC_AUTH_TOKEN"] = authToken
		env["ANTHROPIC_BASE_URL"] = baseURL
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal settings: %v", err)
	}

	if err := os.WriteFile(settingsPath, data, 0o600); err != nil {
		t.Fatalf("Failed to write settings file: %v", err)
	}
}

// createSettingsFileWithEnv creates a test settings.json with custom env vars
func createSettingsFileWithEnv(t *testing.T, tmpDir string, env map[string]string) {
	t.Helper()
	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	settingsPath := filepath.Join(configDir, claudeSettingsFileName)
	settings := make(claudeSettingsFile)
	settingsEnv := getEnv(settings)

	for k, v := range env {
		settingsEnv[k] = v
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal settings: %v", err)
	}

	if err := os.WriteFile(settingsPath, data, 0o600); err != nil {
		t.Fatalf("Failed to write settings file: %v", err)
	}
}

// createBackupFile creates a backup file for testing
func createBackupFile(t *testing.T, tmpDir string, env map[string]string) {
	t.Helper()
	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	backupPath := filepath.Join(configDir, claudeBackupFileName)

	settings := make(claudeSettingsFile)
	settingsEnv := getEnv(settings)

	for k, v := range env {
		settingsEnv[k] = v
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal backup settings: %v", err)
	}

	if err := os.WriteFile(backupPath, data, 0o600); err != nil {
		t.Fatalf("Failed to write backup file: %v", err)
	}
}

// assertFilePermissions checks if a file has specific permissions
func assertFilePermissions(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Failed to stat file %s: %v", path, err)
	}
	got := info.Mode().Perm()
	if got != want {
		t.Errorf("File permissions for %s = %v, want %v", path, got, want)
	}
}

// assertSettingsContent verifies settings.json content matches expected values
func assertSettingsContent(t *testing.T, path, authToken, baseURL string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read settings file %s: %v", path, err)
	}

	var settings claudeSettingsFile
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("Failed to unmarshal settings: %v", err)
	}

	env := getEnv(settings)

	if authToken != "" {
		gotToken := getEnvString(env, "ANTHROPIC_AUTH_TOKEN")
		if gotToken != authToken {
			t.Errorf("ANTHROPIC_AUTH_TOKEN = %v, want %v", gotToken, authToken)
		}
	}

	if baseURL != "" {
		gotURL := getEnvString(env, "ANTHROPIC_BASE_URL")
		if gotURL != baseURL {
			t.Errorf("ANTHROPIC_BASE_URL = %v, want %v", gotURL, baseURL)
		}
	}
}

// assertFileExists checks if a file exists
func assertFileExists(t *testing.T, path string, shouldExist bool) {
	t.Helper()
	_, err := os.Stat(path)
	exists := !os.IsNotExist(err)
	if exists != shouldExist {
		if shouldExist {
			t.Errorf("File %s should exist but doesn't", path)
		} else {
			t.Errorf("File %s should not exist but does", path)
		}
	}
}

// ==================== BaseURL Tests ====================

func TestClaudeBaseURL(t *testing.T) {
	tests := []struct {
		name      string
		relayAddr string
		want      string
	}{
		{
			name:      "Empty address",
			relayAddr: "",
			want:      "http://127.0.0.1:18100",
		},
		{
			name:      "Port only",
			relayAddr: ":18100",
			want:      "http://127.0.0.1:18100",
		},
		{
			name:      "Host only",
			relayAddr: "localhost:18100",
			want:      "http://localhost:18100",
		},
		{
			name:      "Full URL with http",
			relayAddr: "http://example.com:18100",
			want:      "http://example.com:18100",
		},
		{
			name:      "Full URL with https",
			relayAddr: "https://example.com:18100",
			want:      "https://example.com:18100",
		},
		{
			name:      "IP address",
			relayAddr: "192.168.1.1:18100",
			want:      "http://192.168.1.1:18100",
		},
		{
			name:      "IP address with port only",
			relayAddr: ":9080",
			want:      "http://127.0.0.1:9080",
		},
		{
			name:      "Whitespace trimmed",
			relayAddr: "  localhost:18100  ",
			want:      "http://localhost:18100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewClaudeSettingsService(tt.relayAddr)
			got := service.baseURL()
			if got != tt.want {
				t.Errorf("baseURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ==================== ProxyStatus Tests ====================

func TestProxyStatus_NoSettingsFile(t *testing.T) {
	tmpDir := setupTestEnv(t)
	_ = tmpDir // Create tmpDir but don't create any settings file

	service := NewClaudeSettingsService(":18100")
	status, err := service.ProxyStatus()

	if err != nil {
		t.Fatalf("ProxyStatus() error = %v", err)
	}

	if status.Enabled {
		t.Errorf("ProxyStatus() Enabled = %v, want false", status.Enabled)
	}

	if status.BaseURL != "http://127.0.0.1:18100" {
		t.Errorf("ProxyStatus() BaseURL = %v, want http://127.0.0.1:18100", status.BaseURL)
	}
}

func TestProxyStatus_ProxyEnabled(t *testing.T) {
	tmpDir := setupTestEnv(t)
	createSettingsFile(t, tmpDir, "code-together", "http://127.0.0.1:18100")

	service := NewClaudeSettingsService(":18100")
	status, err := service.ProxyStatus()

	if err != nil {
		t.Fatalf("ProxyStatus() error = %v", err)
	}

	if !status.Enabled {
		t.Errorf("ProxyStatus() Enabled = %v, want true", status.Enabled)
	}

	if status.BaseURL != "http://127.0.0.1:18100" {
		t.Errorf("ProxyStatus() BaseURL = %v, want http://127.0.0.1:18100", status.BaseURL)
	}
}

func TestProxyStatus_ProxyDisabled(t *testing.T) {
	tmpDir := setupTestEnv(t)
	createSettingsFile(t, tmpDir, "different-token", "http://127.0.0.1:18100")

	service := NewClaudeSettingsService(":18100")
	status, err := service.ProxyStatus()

	if err != nil {
		t.Fatalf("ProxyStatus() error = %v", err)
	}

	if status.Enabled {
		t.Errorf("ProxyStatus() Enabled = %v, want false (different token)", status.Enabled)
	}
}

func TestProxyStatus_InvalidJSON(t *testing.T) {
	tmpDir := setupTestEnv(t)
	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	settingsPath := filepath.Join(configDir, claudeSettingsFileName)
	if err := os.WriteFile(settingsPath, []byte("invalid json{{{"), 0o600); err != nil {
		t.Fatal(err)
	}

	service := NewClaudeSettingsService(":18100")
	status, err := service.ProxyStatus()

	// Should return gracefully with disabled status
	if err != nil {
		t.Fatalf("ProxyStatus() error = %v, want nil", err)
	}

	if status.Enabled {
		t.Errorf("ProxyStatus() Enabled = %v, want false (invalid JSON)", status.Enabled)
	}
}

func TestProxyStatus_BaseURLFormats(t *testing.T) {
	tests := []struct {
		name        string
		relayAddr   string
		settingsURL string
		expectedURL string
	}{
		{
			name:        "Default relay address",
			relayAddr:   ":18100",
			settingsURL: "http://127.0.0.1:18100",
			expectedURL: "http://127.0.0.1:18100",
		},
		{
			name:        "Custom relay address",
			relayAddr:   "localhost:9080",
			settingsURL: "http://localhost:9080",
			expectedURL: "http://localhost:9080",
		},
		{
			name:        "Full URL relay",
			relayAddr:   "https://proxy.example.com",
			settingsURL: "https://proxy.example.com",
			expectedURL: "https://proxy.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := setupTestEnv(t)
			createSettingsFile(t, tmpDir, "code-together", tt.settingsURL)

			service := NewClaudeSettingsService(tt.relayAddr)
			status, err := service.ProxyStatus()

			if err != nil {
				t.Fatalf("ProxyStatus() error = %v", err)
			}

			if status.BaseURL != tt.expectedURL {
				t.Errorf("ProxyStatus() BaseURL = %v, want %v", status.BaseURL, tt.expectedURL)
			}
		})
	}
}

// ==================== EnableProxy Tests ====================

func TestEnableProxy_NoExistingSettings(t *testing.T) {
	tmpDir := setupTestEnv(t)

	service := NewClaudeSettingsService(":18100")
	if err := service.EnableProxy(); err != nil {
		t.Fatalf("EnableProxy() error = %v", err)
	}

	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	settingsPath := filepath.Join(configDir, claudeSettingsFileName)

	// Verify settings file created
	assertFileExists(t, settingsPath, true)

	// Verify file permissions
	assertFilePermissions(t, settingsPath, 0o600)

	// Verify content
	assertSettingsContent(t, settingsPath, "code-together", "http://127.0.0.1:18100")

	// Verify backup was NOT created (no existing settings)
	backupPath := filepath.Join(configDir, claudeBackupFileName)
	assertFileExists(t, backupPath, false)
}

func TestEnableProxy_WithExistingSettings(t *testing.T) {
	tmpDir := setupTestEnv(t)
	createSettingsFileWithEnv(t, tmpDir, map[string]string{
		"OTHER_VAR":  "other_value",
		"CUSTOM_VAR": "custom_value",
	})

	service := NewClaudeSettingsService(":18100")
	if err := service.EnableProxy(); err != nil {
		t.Fatalf("EnableProxy() error = %v", err)
	}

	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	settingsPath := filepath.Join(configDir, claudeSettingsFileName)
	backupPath := filepath.Join(configDir, claudeBackupFileName)

	// Verify settings file updated
	assertFileExists(t, settingsPath, true)
	assertSettingsContent(t, settingsPath, "code-together", "http://127.0.0.1:18100")

	// Verify backup was created
	assertFileExists(t, backupPath, true)
	assertFilePermissions(t, backupPath, 0o600)

	// Verify backup contains original settings
	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}
	var backupSettings claudeSettingsFile
	if err := json.Unmarshal(data, &backupSettings); err != nil {
		t.Fatalf("Failed to unmarshal backup: %v", err)
	}

	backupEnv := getEnv(backupSettings)
	if getEnvString(backupEnv, "OTHER_VAR") != "other_value" {
		t.Errorf("Backup should preserve OTHER_VAR, got %v", getEnvString(backupEnv, "OTHER_VAR"))
	}
	if getEnvString(backupEnv, "CUSTOM_VAR") != "custom_value" {
		t.Errorf("Backup should preserve CUSTOM_VAR, got %v", getEnvString(backupEnv, "CUSTOM_VAR"))
	}
}

func TestEnableProxy_PreservesExistingEnvVars(t *testing.T) {
	tmpDir := setupTestEnv(t)
	createSettingsFileWithEnv(t, tmpDir, map[string]string{
		"ANTHROPIC_AUTH_TOKEN": "old-token",
		"ANTHROPIC_BASE_URL":   "http://old-url:8080",
		"API_TIMEOUT_MS":       "3000000",
		"CUSTOM_VAR":           "custom_value",
	})

	service := NewClaudeSettingsService(":18100")
	if err := service.EnableProxy(); err != nil {
		t.Fatalf("EnableProxy() error = %v", err)
	}

	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	settingsPath := filepath.Join(configDir, claudeSettingsFileName)

	// Read the new settings
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read settings: %v", err)
	}

	var settings claudeSettingsFile
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("Failed to unmarshal settings: %v", err)
	}

	env := getEnv(settings)

	// Verify ANTHROPIC_* vars are set correctly
	if getEnvString(env, "ANTHROPIC_AUTH_TOKEN") != "code-together" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %v, want code-together", getEnvString(env, "ANTHROPIC_AUTH_TOKEN"))
	}
	if getEnvString(env, "ANTHROPIC_BASE_URL") != "http://127.0.0.1:18100" {
		t.Errorf("ANTHROPIC_BASE_URL = %v, want http://127.0.0.1:18100", getEnvString(env, "ANTHROPIC_BASE_URL"))
	}

	// Verify other env vars are preserved
	if getEnvString(env, "API_TIMEOUT_MS") != "3000000" {
		t.Errorf("API_TIMEOUT_MS = %v, want 3000000 (should be preserved)", getEnvString(env, "API_TIMEOUT_MS"))
	}
	if getEnvString(env, "CUSTOM_VAR") != "custom_value" {
		t.Errorf("CUSTOM_VAR = %v, want custom_value (should be preserved)", getEnvString(env, "CUSTOM_VAR"))
	}
}

func TestEnableProxy_FilePermissions(t *testing.T) {
	tmpDir := setupTestEnv(t)

	service := NewClaudeSettingsService(":18100")
	if err := service.EnableProxy(); err != nil {
		t.Fatalf("EnableProxy() error = %v", err)
	}

	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	settingsPath := filepath.Join(configDir, claudeSettingsFileName)

	// Verify directory permissions
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("Failed to stat directory: %v", err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("Directory permissions = %v, want 0o755", info.Mode().Perm())
	}

	// Verify file permissions
	assertFilePermissions(t, settingsPath, 0o600)
}

func TestEnableProxy_DirectoryCreation(t *testing.T) {
	tmpDir := setupTestEnv(t)
	// Don't create .claude directory - test should create it

	service := NewClaudeSettingsService(":18100")
	if err := service.EnableProxy(); err != nil {
		t.Fatalf("EnableProxy() error = %v", err)
	}

	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	settingsPath := filepath.Join(configDir, claudeSettingsFileName)

	// Verify directory was created
	assertFileExists(t, settingsPath, true)

	// Verify directory permissions
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("Failed to stat directory: %v", err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("Directory permissions = %v, want 0o755", info.Mode().Perm())
	}
}

func TestEnableProxy_OverwritesExistingBackup(t *testing.T) {
	tmpDir := setupTestEnv(t)
	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create initial backup with old content
	backupPath := filepath.Join(configDir, claudeBackupFileName)
	createBackupFile(t, tmpDir, map[string]string{
		"OLD_BACKUP": "old_backup_value",
	})

	// Create settings that will be backed up
	createSettingsFileWithEnv(t, tmpDir, map[string]string{
		"CURRENT_VAR": "current_value",
	})

	service := NewClaudeSettingsService(":18100")
	if err := service.EnableProxy(); err != nil {
		t.Fatalf("EnableProxy() error = %v", err)
	}

	// Verify backup was overwritten with new content
	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}
	var backupSettings claudeSettingsFile
	if err := json.Unmarshal(data, &backupSettings); err != nil {
		t.Fatalf("Failed to unmarshal backup: %v", err)
	}

	backupEnv := getEnv(backupSettings)
	// Should have current settings, not old backup
	if getEnvString(backupEnv, "CURRENT_VAR") != "current_value" {
		t.Errorf("Backup should have current settings, got %v", getEnvString(backupEnv, "CURRENT_VAR"))
	}
	if _, exists := backupEnv["OLD_BACKUP"]; exists {
		t.Error("Old backup content should be overwritten")
	}
}

// ==================== DisableProxy Tests ====================

func TestDisableProxy_WithBackup(t *testing.T) {
	tmpDir := setupTestEnv(t)

	// Create backup file
	createBackupFile(t, tmpDir, map[string]string{
		"RESTORED_VAR": "restored_value",
		"OTHER_VAR":    "other_value",
	})

	// Create current settings with proxy config
	createSettingsFile(t, tmpDir, "code-together", "http://127.0.0.1:18100")

	service := NewClaudeSettingsService(":18100")
	if err := service.DisableProxy(); err != nil {
		t.Fatalf("DisableProxy() error = %v", err)
	}

	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	settingsPath := filepath.Join(configDir, claudeSettingsFileName)
	backupPath := filepath.Join(configDir, claudeBackupFileName)

	// Verify settings file was restored
	assertFileExists(t, settingsPath, true)

	// Verify restored content
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read settings: %v", err)
	}
	var settings claudeSettingsFile
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("Failed to unmarshal settings: %v", err)
	}

	env := getEnv(settings)
	if getEnvString(env, "RESTORED_VAR") != "restored_value" {
		t.Errorf("RESTORED_VAR = %v, want restored_value", getEnvString(env, "RESTORED_VAR"))
	}

	// Verify backup file was removed
	assertFileExists(t, backupPath, false)
}

func TestDisableProxy_WithoutBackup(t *testing.T) {
	tmpDir := setupTestEnv(t)
	createSettingsFile(t, tmpDir, "code-together", "http://127.0.0.1:18100")

	service := NewClaudeSettingsService(":18100")
	if err := service.DisableProxy(); err != nil {
		t.Fatalf("DisableProxy() error = %v", err)
	}

	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	settingsPath := filepath.Join(configDir, claudeSettingsFileName)

	// Verify settings file was removed
	assertFileExists(t, settingsPath, false)
}

func TestDisableProxy_NoSettingsFile(t *testing.T) {
	tmpDir := setupTestEnv(t)
	// Don't create any files

	service := NewClaudeSettingsService(":18100")
	if err := service.DisableProxy(); err != nil {
		t.Fatalf("DisableProxy() error = %v, want nil", err)
	}

	// Should complete without error
	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	settingsPath := filepath.Join(configDir, claudeSettingsFileName)
	assertFileExists(t, settingsPath, false)
}

func TestDisableProxy_BackupPermissions(t *testing.T) {
	tmpDir := setupTestEnv(t)

	// Create backup file
	createBackupFile(t, tmpDir, map[string]string{
		"RESTORED_VAR": "restored_value",
	})

	// Create current settings
	createSettingsFile(t, tmpDir, "code-together", "http://127.0.0.1:18100")

	service := NewClaudeSettingsService(":18100")
	if err := service.DisableProxy(); err != nil {
		t.Fatalf("DisableProxy() error = %v", err)
	}

	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	settingsPath := filepath.Join(configDir, claudeSettingsFileName)

	// Verify restored file has correct permissions
	assertFilePermissions(t, settingsPath, 0o600)
}

func TestDisableProxy_PartialRestoreFailure(t *testing.T) {
	tmpDir := setupTestEnv(t)
	configDir := filepath.Join(tmpDir, claudeSettingsDir)

	// Create backup but NO settings file
	createBackupFile(t, tmpDir, map[string]string{
		"BACKUP_ONLY": "backup_value",
	})

	service := NewClaudeSettingsService(":18100")
	if err := service.DisableProxy(); err != nil {
		t.Fatalf("DisableProxy() error = %v", err)
	}

	settingsPath := filepath.Join(configDir, claudeSettingsFileName)
	backupPath := filepath.Join(configDir, claudeBackupFileName)

	// Backup should be renamed to settings.json
	assertFileExists(t, settingsPath, true)
	assertFileExists(t, backupPath, false)

	// Verify restored content
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read settings: %v", err)
	}
	var settings claudeSettingsFile
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("Failed to unmarshal settings: %v", err)
	}

	env := getEnv(settings)
	if getEnvString(env, "BACKUP_ONLY") != "backup_value" {
		t.Errorf("BACKUP_ONLY = %v, want backup_value", getEnvString(env, "BACKUP_ONLY"))
	}
}

// ==================== Error Scenarios ====================

func TestEnableProxy_MkdirFailure(t *testing.T) {
	tmpDir := setupTestEnv(t)
	configDir := filepath.Join(tmpDir, claudeSettingsDir)

	// Create a file with the same name as the directory to prevent mkdir
	if err := os.WriteFile(configDir, []byte("blocking file"), 0o600); err != nil {
		t.Fatal(err)
	}

	service := NewClaudeSettingsService(":18100")
	err := service.EnableProxy()

	if err == nil {
		t.Error("EnableProxy() should return error when cannot create directory")
	}
}

// ==================== Integration Scenarios ====================

func TestEnableDisableCycle(t *testing.T) {
	tmpDir := setupTestEnv(t)

	service := NewClaudeSettingsService(":18100")

	// 1. Enable proxy
	if err := service.EnableProxy(); err != nil {
		t.Fatalf("First EnableProxy() error = %v", err)
	}

	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	settingsPath := filepath.Join(configDir, claudeSettingsFileName)

	// Verify enabled
	data, _ := os.ReadFile(settingsPath)
	var settings claudeSettingsFile
	json.Unmarshal(data, &settings)
	env := getEnv(settings)
	if getEnvString(env, "ANTHROPIC_AUTH_TOKEN") != "code-together" {
		t.Error("Proxy should be enabled after first EnableProxy")
	}

	// 2. Disable proxy
	if err := service.DisableProxy(); err != nil {
		t.Fatalf("First DisableProxy() error = %v", err)
	}

	// Verify disabled (settings removed or backup restored if exists)
	if _, err := os.Stat(settingsPath); err == nil {
		// File exists, should be from backup (if we had one)
		data, _ = os.ReadFile(settingsPath)
		json.Unmarshal(data, &settings)
		env = getEnv(settings)
		if getEnvString(env, "ANTHROPIC_AUTH_TOKEN") == "code-together" {
			t.Error("Proxy should be disabled after DisableProxy")
		}
	}

	// 3. Enable proxy again
	if err := service.EnableProxy(); err != nil {
		t.Fatalf("Second EnableProxy() error = %v", err)
	}

	// Verify enabled again
	data, _ = os.ReadFile(settingsPath)
	json.Unmarshal(data, &settings)
	env = getEnv(settings)
	if getEnvString(env, "ANTHROPIC_AUTH_TOKEN") != "code-together" {
		t.Error("Proxy should be enabled after second EnableProxy")
	}

	// 4. Disable proxy again
	if err := service.DisableProxy(); err != nil {
		t.Fatalf("Second DisableProxy() error = %v", err)
	}

	// Final state should be clean (no proxy enabled)
	if _, err := os.Stat(settingsPath); err == nil {
		data, _ = os.ReadFile(settingsPath)
		json.Unmarshal(data, &settings)
		env = getEnv(settings)
		if getEnvString(env, "ANTHROPIC_AUTH_TOKEN") == "code-together" {
			t.Error("Proxy should be disabled after final DisableProxy")
		}
	}
}

func TestMultipleEnableCalls(t *testing.T) {
	tmpDir := setupTestEnv(t)

	service := NewClaudeSettingsService(":18100")

	// First enable
	if err := service.EnableProxy(); err != nil {
		t.Fatalf("First EnableProxy() error = %v", err)
	}

	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	backupPath := filepath.Join(configDir, claudeBackupFileName)

	// Verify no backup created on first enable (no existing settings)
	assertFileExists(t, backupPath, false)

	// Create some settings to backup
	createSettingsFileWithEnv(t, tmpDir, map[string]string{
		"SECOND_RUN": "second_run_value",
	})

	// Second enable - should create backup
	if err := service.EnableProxy(); err != nil {
		t.Fatalf("Second EnableProxy() error = %v", err)
	}

	// Verify backup created
	assertFileExists(t, backupPath, true)

	// Read backup content
	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}
	var backupSettings claudeSettingsFile
	if err := json.Unmarshal(data, &backupSettings); err != nil {
		t.Fatalf("Failed to unmarshal backup: %v", err)
	}

	backupEnv := getEnv(backupSettings)
	// Verify backup contains the settings from second run
	if getEnvString(backupEnv, "SECOND_RUN") != "second_run_value" {
		t.Errorf("Backup should contain settings from second run, got %v", getEnvString(backupEnv, "SECOND_RUN"))
	}
}

func TestConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	tmpDir := setupTestEnv(t)

	service := NewClaudeSettingsService(":18100")

	// Run multiple operations concurrently
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func() {
			defer func() { done <- true }()
			_ = service.EnableProxy()
			_ = service.DisableProxy()
		}()
	}

	// Wait for all goroutines to complete (5 goroutines)
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify final state is consistent
	// The exact state depends on timing, but no errors should occur
	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	if _, err := os.Stat(configDir); err != nil && !os.IsNotExist(err) {
		t.Errorf("Unexpected error checking config directory: %v", err)
	}
}

// ==================== Case Sensitivity Tests ====================

func TestProxyStatus_CaseInsensitiveToken(t *testing.T) {
	tests := []struct {
		name            string
		authToken       string
		shouldBeEnabled bool
	}{
		{
			name:            "Exact match",
			authToken:       "code-together",
			shouldBeEnabled: true,
		},
		{
			name:            "Uppercase",
			authToken:       "CODE-TOGETHER",
			shouldBeEnabled: true,
		},
		{
			name:            "Mixed case",
			authToken:       "Code-Together",
			shouldBeEnabled: true,
		},
		{
			name:            "Different token",
			authToken:       "different-token",
			shouldBeEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean tmpDir for each subtest
			tmpDir := setupTestEnv(t)
			createSettingsFile(t, tmpDir, tt.authToken, "http://127.0.0.1:18100")

			service := NewClaudeSettingsService(":18100")
			status, err := service.ProxyStatus()

			if err != nil {
				t.Fatalf("ProxyStatus() error = %v", err)
			}

			if status.Enabled != tt.shouldBeEnabled {
				t.Errorf("ProxyStatus() Enabled = %v, want %v", status.Enabled, tt.shouldBeEnabled)
			}
		})
	}
}

func TestProxyStatus_CaseInsensitiveBaseURL(t *testing.T) {
	tests := []struct {
		name            string
		baseURL         string
		shouldBeEnabled bool
	}{
		{
			name:            "Exact match",
			baseURL:         "http://127.0.0.1:18100",
			shouldBeEnabled: true,
		},
		{
			name:            "Uppercase HTTP",
			baseURL:         "HTTP://127.0.0.1:18100",
			shouldBeEnabled: true,
		},
		{
			name:            "Different URL",
			baseURL:         "http://different.com:18100",
			shouldBeEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := setupTestEnv(t)
			_ = tmpDir // Used by createSettingsFile
			createSettingsFile(t, tmpDir, "code-together", tt.baseURL)

			service := NewClaudeSettingsService(":18100")
			status, err := service.ProxyStatus()

			if err != nil {
				t.Fatalf("ProxyStatus() error = %v", err)
			}

			if status.Enabled != tt.shouldBeEnabled {
				t.Errorf("ProxyStatus() Enabled = %v, want %v", status.Enabled, tt.shouldBeEnabled)
			}
		})
	}
}

// ==================== Comprehensive Preservation Tests ====================

func TestEnableProxy_PreservesAllFields(t *testing.T) {
	tmpDir := setupTestEnv(t)

	// Create comprehensive settings file matching real structure
	settingsJSON := `{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "dc63c889782e40db9cc62210632b8710.eQt1UXZb69d4UDE3",
    "ANTHROPIC_BASE_URL": "https://open.bigmodel.cn/api/anthropic",
    "API_TIMEOUT_MS": "3000000",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
    "CUSTOM_VAR": "custom_value"
  },
  "permissions": {
    "allow": []
  },
  "hooks": {
    "PreToolUse": [{
      "matcher": "*",
      "hooks": [{"type": "command", "command": "echo test"}]
    }],
    "SessionStart": [{
      "matcher": "*",
      "hooks": [{"type": "command", "command": "echo start"}]
    }]
  },
  "statusLine": {
    "type": "command",
    "command": "~/.claude/statusline.sh"
  },
  "enabledPlugins": {
    "context7@claude-plugins-official": true,
    "github@claude-plugins-official": true
  }
}`

	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(configDir, claudeSettingsFileName)
	if err := os.WriteFile(settingsPath, []byte(settingsJSON), 0o600); err != nil {
		t.Fatal(err)
	}

	service := NewClaudeSettingsService(":18100")
	if err := service.EnableProxy(); err != nil {
		t.Fatalf("EnableProxy() error = %v", err)
	}

	// Read updated settings
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}

	var settings claudeSettingsFile
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatal(err)
	}

	// Verify proxy settings are updated
	env := getEnv(settings)
	if getEnvString(env, "ANTHROPIC_AUTH_TOKEN") != "code-together" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %v, want 'code-together'", getEnvString(env, "ANTHROPIC_AUTH_TOKEN"))
	}
	if getEnvString(env, "ANTHROPIC_BASE_URL") != "http://127.0.0.1:18100" {
		t.Errorf("ANTHROPIC_BASE_URL = %v, want 'http://127.0.0.1:18100'", getEnvString(env, "ANTHROPIC_BASE_URL"))
	}

	// Verify other env vars are preserved
	if getEnvString(env, "API_TIMEOUT_MS") != "3000000" {
		t.Errorf("API_TIMEOUT_MS not preserved, got %v", getEnvString(env, "API_TIMEOUT_MS"))
	}
	if getEnvString(env, "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC") != "1" {
		t.Errorf("CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC not preserved")
	}
	if getEnvString(env, "CUSTOM_VAR") != "custom_value" {
		t.Errorf("CUSTOM_VAR not preserved, got %v", getEnvString(env, "CUSTOM_VAR"))
	}

	// Verify top-level fields are preserved
	if settings["permissions"] == nil {
		t.Error("permissions field lost")
	}
	if settings["hooks"] == nil {
		t.Error("hooks field lost")
	}
	if settings["statusLine"] == nil {
		t.Error("statusLine field lost")
	}
	if settings["enabledPlugins"] == nil {
		t.Error("enabledPlugins field lost")
	}

	// Verify hooks structure is preserved
	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		t.Fatal("hooks is not a map")
	}
	if hooks["PreToolUse"] == nil {
		t.Error("PreToolUse hook lost")
	}
	if hooks["SessionStart"] == nil {
		t.Error("SessionStart hook lost")
	}

	// Verify statusLine structure
	statusLine, ok := settings["statusLine"].(map[string]any)
	if !ok {
		t.Fatal("statusLine is not a map")
	}
	if statusLine["type"] != "command" {
		t.Errorf("statusLine.type = %v, want 'command'", statusLine["type"])
	}

	// Verify enabledPlugins structure
	enabledPlugins, ok := settings["enabledPlugins"].(map[string]any)
	if !ok {
		t.Fatal("enabledPlugins is not a map")
	}
	if enabledPlugins["context7@claude-plugins-official"] != true {
		t.Error("context7 plugin setting lost")
	}
}

func TestEnableDisablePreservesAllFields(t *testing.T) {
	tmpDir := setupTestEnv(t)

	// Create settings with all fields
	settingsJSON := `{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "original_token",
    "API_TIMEOUT_MS": "3000000"
  },
  "permissions": {"allow": ["test"]},
  "hooks": {
    "PreToolUse": [{"matcher": "*", "hooks": [{"type": "command", "command": "echo"}]}]
  },
  "statusLine": {"type": "command", "command": "test.sh"},
  "enabledPlugins": {"plugin1": true}
}`

	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(configDir, claudeSettingsFileName)
	if err := os.WriteFile(settingsPath, []byte(settingsJSON), 0o600); err != nil {
		t.Fatal(err)
	}

	service := NewClaudeSettingsService(":18100")

	// Enable proxy
	if err := service.EnableProxy(); err != nil {
		t.Fatalf("EnableProxy() error = %v", err)
	}

	// Verify all fields present after enable
	data, _ := os.ReadFile(settingsPath)
	var settingsAfterEnable claudeSettingsFile
	json.Unmarshal(data, &settingsAfterEnable)
	if settingsAfterEnable["permissions"] == nil || settingsAfterEnable["hooks"] == nil {
		t.Error("Fields lost after EnableProxy")
	}

	// Disable proxy
	if err := service.DisableProxy(); err != nil {
		t.Fatalf("DisableProxy() error = %v", err)
	}

	// Verify all fields restored from backup
	data, _ = os.ReadFile(settingsPath)
	var settingsAfterDisable claudeSettingsFile
	json.Unmarshal(data, &settingsAfterDisable)

	env := getEnv(settingsAfterDisable)
	if getEnvString(env, "ANTHROPIC_AUTH_TOKEN") != "original_token" {
		t.Error("Backup not restored correctly")
	}
	if getEnvString(env, "API_TIMEOUT_MS") != "3000000" {
		t.Error("API_TIMEOUT_MS lost after disable")
	}
	if settingsAfterDisable["permissions"] == nil {
		t.Error("permissions lost after disable")
	}
	if settingsAfterDisable["hooks"] == nil {
		t.Error("hooks lost after disable")
	}
	if settingsAfterDisable["statusLine"] == nil {
		t.Error("statusLine lost after disable")
	}
	if settingsAfterDisable["enabledPlugins"] == nil {
		t.Error("enabledPlugins lost after disable")
	}
}

// ==================== Hook Merge Tests ====================

func TestGetCTHookConfig(t *testing.T) {
	config := getCTHookConfig()

	// Verify hook config structure
	if config == nil {
		t.Fatal("getCTHookConfig() returned nil")
	}

	// Verify required fields
	hookType, ok := config["type"]
	if !ok || hookType != "command" {
		t.Errorf("hook type = %v, want 'command'", hookType)
	}

	command, ok := config["command"]
	if !ok {
		t.Fatal("hook config missing 'command' field")
	}

	commandStr, ok := command.(string)
	if !ok {
		t.Fatal("hook command is not a string")
	}

	// Verify curl command contains expected elements
	if !strings.Contains(commandStr, "curl -X POST") {
		t.Error("hook command missing 'curl -X POST'")
	}
	if !strings.Contains(commandStr, "http://localhost:18100/collect/claude") {
		t.Error("hook command missing correct endpoint URL")
	}
	if !strings.Contains(commandStr, "Content-Type: application/json") {
		t.Error("hook command missing Content-Type header")
	}
	if !strings.Contains(commandStr, "jq -cM") {
		t.Error("hook command missing jq preprocessing")
	}
}

func TestShouldAddHook_NoExistingHooks(t *testing.T) {
	settings := make(claudeSettingsFile)

	// No hooks section at all
	if !shouldAddHook(settings, "SessionStart") {
		t.Error("shouldAddHook should return true when no hooks exist")
	}
}

func TestShouldAddHook_WithExistingHook(t *testing.T) {
	settings := make(claudeSettingsFile)
	ctHookConfig := getCTHookConfig()

	settings["hooks"] = map[string]any{
		"SessionStart": []any{ctHookConfig},
	}

	// CT hook already exists
	if shouldAddHook(settings, "SessionStart") {
		t.Error("shouldAddHook should return false when CT hook already exists")
	}
}

func TestShouldAddHook_DifferentHookExists(t *testing.T) {
	settings := make(claudeSettingsFile)

	differentHook := map[string]any{
		"command": "echo 'different hook'",
		"type":    "command",
	}

	settings["hooks"] = map[string]any{
		"SessionStart": []any{differentHook},
	}

	// Different hook exists, should add CT hook
	if !shouldAddHook(settings, "SessionStart") {
		t.Error("shouldAddHook should return true when only non-CT hook exists")
	}
}

func TestMergeCTHooks_NoExistingHooks(t *testing.T) {
	settings := make(claudeSettingsFile)

	// Add some other fields
	settings["env"] = map[string]any{
		"ANTHROPIC_AUTH_TOKEN": "test-token",
	}
	settings["permissions"] = map[string]any{
		"allow": []string{},
	}

	result := mergeCTHooks(settings)

	// Verify all 14 hook events were added
	hooks, ok := result["hooks"].(map[string]any)
	if !ok {
		t.Fatal("result hooks is not a map")
	}

	expectedHooks := []string{
		"Notification", "PermissionRequest", "PostToolUse",
		"PostToolUseFailure", "PreCompact", "PreToolUse",
		"SessionEnd", "SessionStart", "Stop",
		"SubagentStart", "SubagentStop", "TaskCompleted",
		"TeammateIdle", "UserPromptSubmit",
	}

	if len(hooks) != len(expectedHooks) {
		t.Errorf("got %d hooks, want %d", len(hooks), len(expectedHooks))
	}

	for _, hookName := range expectedHooks {
		if _, exists := hooks[hookName]; !exists {
			t.Errorf("missing hook: %s", hookName)
		}
	}

	// Verify other fields are preserved
	if result["env"] == nil {
		t.Error("env field lost during merge")
	}
	if result["permissions"] == nil {
		t.Error("permissions field lost during merge")
	}
}

func TestMergeCTHooks_WithExistingCTHooks(t *testing.T) {
	settings := make(claudeSettingsFile)
	ctHookConfig := getCTHookConfig()

	// Pre-populate some CT hooks
	settings["hooks"] = map[string]any{
		"SessionStart":     []any{ctHookConfig},
		"UserPromptSubmit": []any{ctHookConfig},
	}

	result := mergeCTHooks(settings)

	hooks, ok := result["hooks"].(map[string]any)
	if !ok {
		t.Fatal("result hooks is not a map")
	}

	// Verify SessionStart and UserPromptSubmit still have only one CT hook (no duplicates)
	sessionStartHooks, ok := hooks["SessionStart"].([]any)
	if !ok || len(sessionStartHooks) != 1 {
		t.Errorf("SessionStart should have 1 hook, got %d", len(sessionStartHooks))
	}

	userPromptHooks, ok := hooks["UserPromptSubmit"].([]any)
	if !ok || len(userPromptHooks) != 1 {
		t.Errorf("UserPromptSubmit should have 1 hook, got %d", len(userPromptHooks))
	}

	// Verify other hooks were added
	if _, exists := hooks["PostToolUse"]; !exists {
		t.Error("PostToolUse hook should be added")
	}
}

func TestMergeCTHooks_PreservesUserHooks(t *testing.T) {
	settings := make(claudeSettingsFile)

	userHook := map[string]any{
		"command": "echo 'user custom hook'",
		"type":    "command",
	}

	settings["hooks"] = map[string]any{
		"SessionStart": []any{userHook},
	}

	result := mergeCTHooks(settings)

	hooks, ok := result["hooks"].(map[string]any)
	if !ok {
		t.Fatal("result hooks is not a map")
	}

	// Verify user hook is preserved and CT hook is added
	sessionStartHooks, ok := hooks["SessionStart"].([]any)
	if !ok {
		t.Fatal("SessionStart hooks is not a list")
	}

	if len(sessionStartHooks) != 2 {
		t.Errorf("SessionStart should have 2 hooks (user + CT), got %d", len(sessionStartHooks))
	}

	// Verify first hook is the user hook (preserved in its original format)
	firstHook, ok := sessionStartHooks[0].(map[string]any)
	if !ok {
		t.Fatal("first hook is not a map")
	}

	if firstHook["command"] != "echo 'user custom hook'" {
		t.Error("user hook was not preserved")
	}

	// Verify second hook is the CT hook in wrapped format
	secondHook, ok := sessionStartHooks[1].(map[string]any)
	if !ok {
		t.Fatal("second hook is not a map")
	}

	// CT hook should be in wrapped format with matcher and hooks
	if secondHook["matcher"] != "*" {
		t.Error("CT hook missing 'matcher' field")
	}

	ctHooks, ok := secondHook["hooks"].([]any)
	if !ok || len(ctHooks) != 1 {
		t.Fatal("CT hook 'hooks' field is not valid")
	}

	ctHookConfig, ok := ctHooks[0].(map[string]any)
	if !ok {
		t.Fatal("CT hook config is not a map")
	}

	ctCommand := getCTHookConfig()["command"].(string)
	if ctHookConfig["command"] != ctCommand {
		t.Error("CT hook was not added correctly")
	}
}

func TestMergeCTHooks_MixedScenario(t *testing.T) {
	settings := make(claudeSettingsFile)
	ctHookConfig := getCTHookConfig()

	// Mix of CT hooks and user hooks
	userHook1 := map[string]any{"command": "custom1", "type": "command"}
	userHook2 := map[string]any{"command": "custom2", "type": "command"}

	settings["hooks"] = map[string]any{
		"SessionStart":     []any{ctHookConfig},         // CT hook already exists
		"UserPromptSubmit": []any{userHook1},            // User hook
		"PostToolUse":      []any{userHook1, userHook2}, // Multiple user hooks
	}

	settings["env"] = map[string]any{
		"PRESERVE_ME": "true",
	}

	result := mergeCTHooks(settings)

	hooks, ok := result["hooks"].(map[string]any)
	if !ok {
		t.Fatal("result hooks is not a map")
	}

	// SessionStart: CT hook exists, no duplicate
	sessionStartHooks, _ := hooks["SessionStart"].([]any)
	if len(sessionStartHooks) != 1 {
		t.Errorf("SessionStart should have 1 hook (no duplicate), got %d", len(sessionStartHooks))
	}

	// UserPromptSubmit: user hook + CT hook added
	userPromptHooks, _ := hooks["UserPromptSubmit"].([]any)
	if len(userPromptHooks) != 2 {
		t.Errorf("UserPromptSubmit should have 2 hooks (user + CT), got %d", len(userPromptHooks))
	}

	// PostToolUse: 2 user hooks + CT hook added
	postToolHooks, _ := hooks["PostToolUse"].([]any)
	if len(postToolHooks) != 3 {
		t.Errorf("PostToolUse should have 3 hooks (2 user + CT), got %d", len(postToolHooks))
	}

	// Verify other fields preserved
	env, ok := result["env"].(map[string]any)
	if !ok || env["PRESERVE_ME"] != "true" {
		t.Error("env field not preserved")
	}
}

func TestEnableProxy_WithHookMerge(t *testing.T) {
	tmpDir := setupTestEnv(t)

	// Create settings without hooks
	createSettingsFileWithEnv(t, tmpDir, map[string]string{
		"ANTHROPIC_AUTH_TOKEN": "old-token",
		"API_TIMEOUT_MS":       "3000000",
	})

	service := NewClaudeSettingsService(":18100")
	if err := service.EnableProxy(); err != nil {
		t.Fatalf("EnableProxy() error = %v", err)
	}

	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	settingsPath := filepath.Join(configDir, claudeSettingsFileName)
	backupPath := filepath.Join(configDir, claudeBackupFileName)

	// Verify backup was created
	assertFileExists(t, backupPath, true)

	// Read the updated settings
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read settings: %v", err)
	}

	var settings claudeSettingsFile
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("Failed to unmarshal settings: %v", err)
	}

	// Verify proxy settings are set
	env := getEnv(settings)
	if getEnvString(env, "ANTHROPIC_AUTH_TOKEN") != "code-together" {
		t.Error("ANTHROPIC_AUTH_TOKEN not set correctly")
	}
	if getEnvString(env, "ANTHROPIC_BASE_URL") != "http://127.0.0.1:18100" {
		t.Error("ANTHROPIC_BASE_URL not set correctly")
	}

	// Verify hooks were merged
	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		t.Fatal("hooks not present in settings")
	}

	// Check that all 14 hook types are present
	expectedHooks := []string{
		"Notification", "PermissionRequest", "PostToolUse",
		"PostToolUseFailure", "PreCompact", "PreToolUse",
		"SessionEnd", "SessionStart", "Stop",
		"SubagentStart", "SubagentStop", "TaskCompleted",
		"TeammateIdle", "UserPromptSubmit",
	}

	for _, hookName := range expectedHooks {
		if _, exists := hooks[hookName]; !exists {
			t.Errorf("missing hook: %s", hookName)
		}
	}

	// Verify that hooks contain CT hook command
	sessionStartHooks, ok := hooks["SessionStart"].([]any)
	if !ok || len(sessionStartHooks) != 1 {
		t.Fatalf("SessionStart should have 1 hook, got %d", len(sessionStartHooks))
	}

	// Hook is in wrapped format
	hookEntry, ok := sessionStartHooks[0].(map[string]any)
	if !ok {
		t.Fatal("hook is not a map")
	}

	// Check wrapped structure
	if hookEntry["matcher"] != "*" {
		t.Error("hook missing 'matcher' field")
	}

	nestedHooks, ok := hookEntry["hooks"].([]any)
	if !ok || len(nestedHooks) != 1 {
		t.Fatal("hook 'hooks' field is not valid")
	}

	hookConfig, ok := nestedHooks[0].(map[string]any)
	if !ok {
		t.Fatal("hook config is not a map")
	}

	command, ok := hookConfig["command"].(string)
	if !ok || !strings.Contains(command, "localhost:18100/collect/claude") {
		t.Error("hook command does not contain CT collection endpoint")
	}
}

func TestEnableProxy_HooksWithExistingUserHooks(t *testing.T) {
	tmpDir := setupTestEnv(t)

	// Create settings with existing user hooks
	settingsJSON := `{
		"env": {
			"ANTHROPIC_AUTH_TOKEN": "old-token",
			"API_TIMEOUT_MS": "3000000"
		},
		"hooks": {
			"SessionStart": [{
				"type": "command",
				"command": "echo 'user custom hook'"
			}],
			"PreToolUse": [{
				"type": "command",
				"command": "echo 'another user hook'"
			}]
		},
		"permissions": {
			"allow": []
		}
	}`

	configDir := filepath.Join(tmpDir, claudeSettingsDir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(configDir, claudeSettingsFileName)
	if err := os.WriteFile(settingsPath, []byte(settingsJSON), 0o600); err != nil {
		t.Fatal(err)
	}

	service := NewClaudeSettingsService(":18100")
	if err := service.EnableProxy(); err != nil {
		t.Fatalf("EnableProxy() error = %v", err)
	}

	// Read the updated settings
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read settings: %v", err)
	}

	var settings claudeSettingsFile
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("Failed to unmarshal settings: %v", err)
	}

	// Verify hooks were merged correctly
	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		t.Fatal("hooks not present in settings")
	}

	// SessionStart should have 2 hooks: user hook + CT hook
	sessionStartHooks, ok := hooks["SessionStart"].([]any)
	if !ok {
		t.Fatal("SessionStart hooks is not a list")
	}
	if len(sessionStartHooks) != 2 {
		t.Errorf("SessionStart should have 2 hooks, got %d", len(sessionStartHooks))
	}

	// Verify user hook is first (preserved in simple format)
	firstHook, ok := sessionStartHooks[0].(map[string]any)
	if !ok {
		t.Fatal("first hook is not a map")
	}
	if firstHook["command"] != "echo 'user custom hook'" {
		t.Error("user hook not preserved")
	}

	// Verify CT hook is second (in wrapped format)
	secondHook, ok := sessionStartHooks[1].(map[string]any)
	if !ok {
		t.Fatal("second hook is not a map")
	}

	// CT hook should be in wrapped format
	if secondHook["matcher"] != "*" {
		t.Error("CT hook missing 'matcher' field")
	}

	ctHooks, ok := secondHook["hooks"].([]any)
	if !ok || len(ctHooks) != 1 {
		t.Fatal("CT hook 'hooks' field is not valid")
	}

	ctHookConfig, ok := ctHooks[0].(map[string]any)
	if !ok {
		t.Fatal("CT hook config is not a map")
	}

	ctCommand, ok := ctHookConfig["command"].(string)
	if !ok || !strings.Contains(ctCommand, "localhost:18100/collect/claude") {
		t.Error("CT hook not added")
	}

	// PreToolUse should also have 2 hooks
	preToolHooks, ok := hooks["PreToolUse"].([]any)
	if !ok || len(preToolHooks) != 2 {
		t.Errorf("PreToolUse should have 2 hooks, got %d", len(preToolHooks))
	}

	// Verify other fields preserved
	if settings["permissions"] == nil {
		t.Error("permissions field lost")
	}
}
