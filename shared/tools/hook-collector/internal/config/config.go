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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/code-together/shared/hook-common/config"
)

// Config holds the configuration for the hook collector
type Config struct {
	DatabasePath  string `json:"database_path"`
	RetentionDays int    `json:"retention_days"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	dbPath, err := config.GetDefaultDatabasePath()
	if err != nil {
		// Fallback to relative path if home directory cannot be determined
		dbPath = "./.code-together/hook-events.sqlite"
	}
	return &Config{
		DatabasePath:  dbPath,
		RetentionDays: 90,
	}
}

// LoadConfig loads the configuration from a file
func LoadConfig(path string) (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	configStruct := DefaultConfig()
	if err := json.Unmarshal(data, configStruct); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return configStruct, nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(path string, configStruct *Config) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(configStruct, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// FindWorkspaceRoot finds the workspace root by using the shared config package
func FindWorkspaceRoot() (string, error) {
	return config.FindWorkspaceRoot()
}

// GetConfigPath returns the path to the config file in $HOME/.code-together
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".code-together", "hook-collector.json"), nil
}

// GetDatabasePath returns the path to the database file
// If the path is relative, it's treated as relative to the config directory
func GetDatabasePath(configPath string) string {
	// If already absolute, return as-is
	if filepath.IsAbs(configPath) {
		return configPath
	}

	// For relative paths, resolve relative to $HOME/.code-together
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return configPath
	}
	return filepath.Join(homeDir, ".code-together", configPath)
}

// EnsureConfigDir ensures the $HOME/.code-together directory exists
func EnsureConfigDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".code-together")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create .code-together directory: %w", err)
	}

	return nil
}

// InitConfig initializes the configuration by loading from file or creating default
func InitConfig() (*Config, string, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get config path: %w", err)
	}

	// Ensure .code-together directory exists
	if err := EnsureConfigDir(); err != nil {
		return nil, "", err
	}

	// Load config (returns default if file doesn't exist)
	configStruct, err := LoadConfig(configPath)
	if err != nil {
		return nil, "", err
	}

	// Resolve database path (make it absolute relative to workspace root)
	dbPath := GetDatabasePath(configStruct.DatabasePath)

	// Ensure database directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, "", fmt.Errorf("failed to create database directory: %w", err)
	}

	return configStruct, dbPath, nil
}
