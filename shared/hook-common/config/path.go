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
	"fmt"
	"os"
	"path/filepath"
)

// FindWorkspaceRoot finds the workspace root by looking for .code-together directory
// or .git directory, starting from the current working directory and going up.
func FindWorkspaceRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	dir := cwd
	for {
		// Check for .code-together directory
		codeTogetherDir := filepath.Join(dir, ".code-together")
		if info, err := os.Stat(codeTogetherDir); err == nil && info.IsDir() {
			return dir, nil
		}

		// Check for .git directory
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return dir, nil
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			return cwd, nil // Fall back to current directory
		}
		dir = parent
	}
}

// GetDefaultDatabasePath returns the default path to the SQLite database file
// in the user's home directory under .code-together
func GetDefaultDatabasePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".code-together", "hook-events.sqlite"), nil
}

// EnsureConfigDir ensures the .code-together directory exists in the workspace root
func EnsureConfigDir() error {
	root, err := FindWorkspaceRoot()
	if err != nil {
		return err
	}

	configDir := filepath.Join(root, ".code-together")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create .code-together directory: %w", err)
	}

	return nil
}
