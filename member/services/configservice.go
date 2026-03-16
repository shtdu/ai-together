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

package services

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	configDir  = ".code-together"
	configFile = "config.json"
)

// ConfigService manages config.json file containing app, server, and sync settings
// Thread-safe with mutex for concurrent access
type ConfigService struct {
	path string
	mu   sync.RWMutex
}

// NewConfigService creates a new config service
func NewConfigService() *ConfigService {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	path := filepath.Join(home, configDir, configFile)

	return &ConfigService{
		path: path,
	}
}

// GetApp returns the app settings from config
func (c *ConfigService) GetApp() (AppSettings, error) {
	config, err := c.Load()
	if err != nil {
		return AppSettings{}, err
	}
	return config.App, nil
}

// SetApp saves the app settings to config
func (c *ConfigService) SetApp(settings AppSettings) error {
	config, err := c.Load()
	if err != nil {
		// If config doesn't exist, create a new one
		if os.IsNotExist(err) {
			config = c.defaultConfig()
		} else {
			return err
		}
	}
	config.App = settings
	return c.Save(config)
}

// GetServer returns the server config from config
func (c *ConfigService) GetServer() (ServerConfig, error) {
	config, err := c.Load()
	if err != nil {
		return ServerConfig{}, err
	}
	return config.Server, nil
}

// SetServer saves the server config to config
func (c *ConfigService) SetServer(server ServerConfig) error {
	config, err := c.Load()
	if err != nil {
		// If config doesn't exist, create a new one
		if os.IsNotExist(err) {
			config = c.defaultConfig()
		} else {
			return err
		}
	}
	config.Server = server
	return c.Save(config)
}

// GetSync returns the sync state from config
func (c *ConfigService) GetSync() (SyncState, error) {
	config, err := c.Load()
	if err != nil {
		return SyncState{}, err
	}
	return config.Sync, nil
}

// SetSync saves the sync state to config
func (c *ConfigService) SetSync(sync SyncState) error {
	config, err := c.Load()
	if err != nil {
		// If config doesn't exist, create a new one
		if os.IsNotExist(err) {
			config = c.defaultConfig()
		} else {
			return err
		}
	}
	config.Sync = sync
	return c.Save(config)
}

// ClearSyncState resets the sync state to zero time (never synced)
// This is used during logout to ensure fresh sync state on next login
func (c *ConfigService) ClearSyncState() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	config, err := c.loadLocked()
	if err != nil {
		return err
	}

	// Reset sync state to zero time
	config.Sync = SyncState{
		LastSyncTime: (time.Time{}).UTC(), // Zero time means never synced
	}

	if err := c.saveLocked(config); err != nil {
		return err
	}

	slog.Info("cleared sync state", "last_sync_time", config.Sync.LastSyncTime)
	return nil
}

// Load loads the entire config from disk
func (c *ConfigService) Load() (AppConfig, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.loadLocked()
}

// Save saves the entire config to disk atomically
func (c *ConfigService) Save(config AppConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.saveLocked(config)
}

// defaultConfig returns default values for a new config
func (c *ConfigService) defaultConfig() AppConfig {
	return AppConfig{
		App: AppSettings{
			ShowHeatmap:   true,
			ShowHomeTitle: true,
			Show24hStats:  false,
			AutoStart:     false,
		},
		Server: ServerConfig{
			ServerURL: "",
		},
		Sync: SyncState{
			LastSyncTime: (time.Time{}).UTC(), // Zero time means never synced
		},
	}
}

// loadLocked loads config from disk (must hold read lock)
func (c *ConfigService) loadLocked() (AppConfig, error) {
	config := c.defaultConfig()

	data, err := os.ReadFile(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return config, nil
		}
		return config, err
	}

	if len(data) == 0 {
		return config, nil
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return config, err
	}

	return config, nil
}

// saveLocked saves config to disk (must hold write lock)
func (c *ConfigService) saveLocked(config AppConfig) error {
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.path, data, 0o644)
}
