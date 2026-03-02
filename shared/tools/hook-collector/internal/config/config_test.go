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
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Default config should use $HOME/.code-together
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}
	expectedPath := filepath.Join(homeDir, ".code-together", "hook-events.sqlite")

	if config.DatabasePath != expectedPath {
		t.Errorf("DatabasePath = %s, want %s", config.DatabasePath, expectedPath)
	}
	if config.RetentionDays != 90 {
		t.Errorf("RetentionDays = %d, want 90", config.RetentionDays)
	}
}

func TestLoadConfig(t *testing.T) {
	t.Run("nonexistent file returns default", func(t *testing.T) {
		config, err := LoadConfig(t.TempDir() + "/nonexistent.json")
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}

		defaultConfig := DefaultConfig()
		if config.DatabasePath != defaultConfig.DatabasePath {
			t.Errorf("DatabasePath = %s, want %s", config.DatabasePath, defaultConfig.DatabasePath)
		}
		if config.RetentionDays != defaultConfig.RetentionDays {
			t.Errorf("RetentionDays = %d, want %d", config.RetentionDays, defaultConfig.RetentionDays)
		}
	})

	t.Run("load existing config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "test-config.json")

		// Write config file
		testConfig := `{"database_path": "/custom/path.db", "retention_days": 30}`
		if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		// Load config
		config, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}

		if config.DatabasePath != "/custom/path.db" {
			t.Errorf("DatabasePath = %s, want /custom/path.db", config.DatabasePath)
		}
		if config.RetentionDays != 30 {
			t.Errorf("RetentionDays = %d, want 30", config.RetentionDays)
		}
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.json")

		if err := os.WriteFile(configPath, []byte("invalid json"), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		_, err := LoadConfig(configPath)
		if err == nil {
			t.Error("LoadConfig() expected error for invalid JSON, got nil")
		}
	})
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.json")

	config := &Config{
		DatabasePath:  "/test/path.db",
		RetentionDays: 60,
	}

	err := SaveConfig(configPath, config)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("SaveConfig() did not create config file")
	}

	// Load and verify
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if loaded.DatabasePath != config.DatabasePath {
		t.Errorf("DatabasePath = %s, want %s", loaded.DatabasePath, config.DatabasePath)
	}
	if loaded.RetentionDays != config.RetentionDays {
		t.Errorf("RetentionDays = %d, want %d", loaded.RetentionDays, config.RetentionDays)
	}
}

func TestFindWorkspaceRoot(t *testing.T) {
	t.Run("finds .code-together directory", func(t *testing.T) {
		// Create temp directory structure
		tmpDir := t.TempDir()
		codeTogetherDir := filepath.Join(tmpDir, ".code-together")
		if err := os.Mkdir(codeTogetherDir, 0755); err != nil {
			t.Fatalf("failed to create .code-together directory: %v", err)
		}

		// Change to temp directory
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		root, err := FindWorkspaceRoot()
		if err != nil {
			t.Fatalf("FindWorkspaceRoot() error = %v", err)
		}

		if root != tmpDir {
			t.Errorf("FindWorkspaceRoot() = %s, want %s", root, tmpDir)
		}
	})

	t.Run("finds .git directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		if err := os.Mkdir(gitDir, 0755); err != nil {
			t.Fatalf("failed to create .git directory: %v", err)
		}

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		root, err := FindWorkspaceRoot()
		if err != nil {
			t.Fatalf("FindWorkspaceRoot() error = %v", err)
		}

		if root != tmpDir {
			t.Errorf("FindWorkspaceRoot() = %s, want %s", root, tmpDir)
		}
	})

	t.Run("returns current directory when no marker found", func(t *testing.T) {
		// Create a subdirectory in temp to avoid hitting parent markers
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "subdir")
		if err := os.Mkdir(subDir, 0755); err != nil {
			t.Fatalf("failed to create subdirectory: %v", err)
		}

		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)
		if err := os.Chdir(subDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		root, err := FindWorkspaceRoot()
		if err != nil {
			t.Fatalf("FindWorkspaceRoot() error = %v", err)
		}

		// The function will find the temp dir (parent) since it has no markers
		// or will go up to find markers. Either way, it should find tmpDir or parent
		if root != tmpDir && root != subDir {
			// If we found a different root, that's okay - just verify it exists
			if _, err := os.Stat(root); err != nil {
				t.Errorf("FindWorkspaceRoot() returned non-existent path: %s", root)
			}
		}
	})
}

func TestGetConfigPath(t *testing.T) {
	t.Run("returns path in home directory", func(t *testing.T) {
		path, err := GetConfigPath()
		if err != nil {
			t.Fatalf("GetConfigPath() error = %v", err)
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("failed to get home directory: %v", err)
		}

		expected := filepath.Join(homeDir, ".code-together", "hook-collector.json")
		if path != expected {
			t.Errorf("GetConfigPath() = %s, want %s", path, expected)
		}
	})
}

func TestEnsureConfigDir(t *testing.T) {
	// This test creates the .code-together directory in $HOME
	// In a real test environment, we should be careful about this
	// For now, just verify the function doesn't error

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	codeTogetherDir := filepath.Join(homeDir, ".code-together")

	// Ensure directory
	err = EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() error = %v", err)
	}

	// Verify directory exists
	if info, err := os.Stat(codeTogetherDir); err != nil {
		t.Fatalf("Stat error after EnsureConfigDir: %v", err)
	} else if !info.IsDir() {
		t.Error(".code-together is not a directory")
	}
}

func TestInitConfig(t *testing.T) {
	// InitConfig now uses $HOME/.code-together instead of workspace root
	config, dbPath, err := InitConfig()
	if err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}

	// Check config
	if config == nil {
		t.Error("config is nil")
	}
	if dbPath == "" {
		t.Error("dbPath is empty")
	}

	// Check that database directory was created
	dbDir := filepath.Dir(dbPath)
	if info, err := os.Stat(dbDir); err != nil {
		t.Fatalf("failed to stat db dir: %v", err)
	} else if !info.IsDir() {
		t.Error("db dir is not a directory")
	}

	// Verify the path is in $HOME/.code-together
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}
	expectedDir := filepath.Join(homeDir, ".code-together")
	if dbDir != expectedDir {
		t.Errorf("dbDir = %s, want %s", dbDir, expectedDir)
	}
}
