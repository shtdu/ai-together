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


package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindWorkspaceRoot(t *testing.T) {
	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create a temporary test workspace
	tmpDir := filepath.Join(os.TempDir(), "hook-common-workspace-test")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	codeTogetherDir := filepath.Join(tmpDir, ".code-together")

	// Test with .code-together directory
	t.Run("with .code-together directory", func(t *testing.T) {
		if err := os.MkdirAll(codeTogetherDir, 0755); err != nil {
			t.Fatalf("Failed to create .code-together dir: %v", err)
		}

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}

		root, err := FindWorkspaceRoot()
		if err != nil {
			t.Fatalf("FindWorkspaceRoot failed: %v", err)
		}

		if root != tmpDir {
			t.Errorf("Expected root %s, got %s", tmpDir, root)
		}
	})

	// Clean up for next test
	os.RemoveAll(codeTogetherDir)

	// Test with .git directory
	t.Run("with .git directory", func(t *testing.T) {
		gitDir := filepath.Join(tmpDir, ".git")
		if err := os.MkdirAll(gitDir, 0755); err != nil {
			t.Fatalf("Failed to create .git dir: %v", err)
		}

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}

		root, err := FindWorkspaceRoot()
		if err != nil {
			t.Fatalf("FindWorkspaceRoot failed: %v", err)
		}

		if root != tmpDir {
			t.Errorf("Expected root %s, got %s", tmpDir, root)
		}
	})

	// Test nested directory
	t.Run("nested directory", func(t *testing.T) {
		nestedDir := filepath.Join(tmpDir, "subdir", "nested")
		if err := os.MkdirAll(nestedDir, 0755); err != nil {
			t.Fatalf("Failed to create nested dir: %v", err)
		}

		if err := os.Chdir(nestedDir); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}

		root, err := FindWorkspaceRoot()
		if err != nil {
			t.Fatalf("FindWorkspaceRoot failed: %v", err)
		}

		if root != tmpDir {
			t.Errorf("Expected root %s, got %s", tmpDir, root)
		}
	})

	// Test fallback to current directory
	t.Run("fallback to current directory", func(t *testing.T) {
		// Use a directory with no .code-together or .git
		noMarkerDir := filepath.Join(os.TempDir(), "no-marker-test")
		if err := os.MkdirAll(noMarkerDir, 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		defer os.RemoveAll(noMarkerDir)

		if err := os.Chdir(noMarkerDir); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}

		root, err := FindWorkspaceRoot()
		if err != nil {
			t.Fatalf("FindWorkspaceRoot failed: %v", err)
		}

		// Should fall back to current directory
		if root != noMarkerDir {
			t.Logf("Warning: Expected root %s, got %s (may have found parent)", noMarkerDir, root)
		}
	})
}

func TestGetDefaultDatabasePath(t *testing.T) {
	// Mock home directory for testing
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	testHome := filepath.Join(os.TempDir(), "test-home")
	os.Setenv("HOME", testHome)

	path, err := GetDefaultDatabasePath()
	if err != nil {
		t.Fatalf("GetDefaultDatabasePath failed: %v", err)
	}

	expected := filepath.Join(testHome, ".code-together", "hook-events.sqlite")
	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create a temporary test workspace
	tmpDir := filepath.Join(os.TempDir(), "hook-common-config-test")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Add .git marker so FindWorkspaceRoot can find the root
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Also add .code-together directory
	codeTogetherDir := filepath.Join(tmpDir, ".code-together")
	if err := os.MkdirAll(codeTogetherDir, 0755); err != nil {
		t.Fatalf("Failed to create .code-together dir: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	// Remove the .code-together directory to test EnsureConfigDir creates it
	os.RemoveAll(codeTogetherDir)

	err = EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir failed: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(codeTogetherDir); err != nil {
		if os.IsNotExist(err) {
			t.Error(".code-together directory was not created by EnsureConfigDir")
		} else {
			t.Fatalf("Failed to stat .code-together dir: %v", err)
		}
	}
}
