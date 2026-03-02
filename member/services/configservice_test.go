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
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigService_FilePermissions(t *testing.T) {
	// Create a temp directory for testing
	tmpDir := t.TempDir()

	// Create a config service with temp directory
	cs := &ConfigService{
		path: filepath.Join(tmpDir, "config.json"),
	}

	// Test default config is created with correct permissions
	config := AppConfig{
		App: AppSettings{
			ShowHeatmap:   true,
			ShowHomeTitle: true,
			Show24hStats:  false,
			AutoStart:     false,
		},
		Server: ServerConfig{
			ServerURL: "https://example.com",
		},
		Sync: SyncState{
			LastSyncTime: time.Now().UTC(),
		},
	}

	err := cs.Save(config)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file permissions (0o644 = -rw-r--r--)
	info, err := os.Stat(cs.path)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	expectedMode := os.FileMode(0o644)
	if info.Mode() != expectedMode {
		t.Errorf("Expected file permissions %v, got %v", expectedMode, info.Mode())
	}

	// Verify we can load the config back
	loaded, err := cs.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loaded.App.ShowHeatmap != config.App.ShowHeatmap {
		t.Errorf("Expected ShowHeatmap %v, got %v", config.App.ShowHeatmap, loaded.App.ShowHeatmap)
	}

	if loaded.Server.ServerURL != config.Server.ServerURL {
		t.Errorf("Expected ServerURL %v, got %v", config.Server.ServerURL, loaded.Server.ServerURL)
	}
}

func TestConfigService_GetSetSections(t *testing.T) {
	// Create a temp directory for testing
	tmpDir := t.TempDir()

	// Create a config service with temp directory
	cs := &ConfigService{
		path: filepath.Join(tmpDir, "config.json"),
	}

	// Test setting and getting app section
	appSettings := AppSettings{
		ShowHeatmap:   false,
		ShowHomeTitle: true,
		Show24hStats:  true,
		AutoStart:     true,
	}

	err := cs.SetApp(appSettings)
	if err != nil {
		t.Fatalf("Failed to set app settings: %v", err)
	}

	loaded, err := cs.GetApp()
	if err != nil {
		t.Fatalf("Failed to get app settings: %v", err)
	}

	if loaded.ShowHeatmap != appSettings.ShowHeatmap {
		t.Errorf("Expected ShowHeatmap %v, got %v", appSettings.ShowHeatmap, loaded.ShowHeatmap)
	}

	// Test setting and getting server section
	serverConfig := ServerConfig{
		ServerURL: "https://test.example.com",
	}

	err = cs.SetServer(serverConfig)
	if err != nil {
		t.Fatalf("Failed to set server config: %v", err)
	}

	loadedServer, err := cs.GetServer()
	if err != nil {
		t.Fatalf("Failed to get server config: %v", err)
	}

	if loadedServer.ServerURL != serverConfig.ServerURL {
		t.Errorf("Expected ServerURL %v, got %v", serverConfig.ServerURL, loadedServer.ServerURL)
	}

	// Verify both sections are preserved in the same file
	fullConfig, err := cs.Load()
	if err != nil {
		t.Fatalf("Failed to load full config: %v", err)
	}

	if fullConfig.App.ShowHeatmap != appSettings.ShowHeatmap {
		t.Errorf("App section not preserved: expected ShowHeatmap %v, got %v", appSettings.ShowHeatmap, fullConfig.App.ShowHeatmap)
	}

	if fullConfig.Server.ServerURL != serverConfig.ServerURL {
		t.Errorf("Server section not preserved: expected ServerURL %v, got %v", serverConfig.ServerURL, fullConfig.Server.ServerURL)
	}
}

func TestConfigService_DefaultValues(t *testing.T) {
	// Create a temp directory for testing
	tmpDir := t.TempDir()

	// Create a config service with temp directory
	cs := &ConfigService{
		path: filepath.Join(tmpDir, "config.json"),
	}

	// Load when file doesn't exist should return defaults
	config, err := cs.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check default app settings
	if config.App.ShowHeatmap != true {
		t.Errorf("Expected default ShowHeatmap true, got %v", config.App.ShowHeatmap)
	}

	if config.App.ShowHomeTitle != true {
		t.Errorf("Expected default ShowHomeTitle true, got %v", config.App.ShowHomeTitle)
	}

	if config.App.Show24hStats != false {
		t.Errorf("Expected default Show24hStats false, got %v", config.App.Show24hStats)
	}
}
