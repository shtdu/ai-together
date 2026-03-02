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
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// BackgroundSyncService manages background synchronization tasks
type BackgroundSyncService struct {
	configSyncService  *ConfigSyncService
	usageSyncService   *UsageSyncService
	authService        *AuthService
	serverConfigService *ServerConfigService
	// Member is license-unaware; license is enforced on server/Manager only

	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	running    bool
	mu         sync.Mutex

	// Configurable intervals
	configSyncInterval   time.Duration
	usageSyncInterval    time.Duration
	tokenRefreshInterval time.Duration
}

// NewBackgroundSyncService creates a new background sync service
func NewBackgroundSyncService(
	configSyncService *ConfigSyncService,
	usageSyncService *UsageSyncService,
	authService *AuthService,
	serverConfigService *ServerConfigService,
) *BackgroundSyncService {
	return &BackgroundSyncService{
		configSyncService:    configSyncService,
		usageSyncService:     usageSyncService,
		authService:          authService,
		serverConfigService:  serverConfigService,
		configSyncInterval:   30 * time.Minute, // Sync config every 30 minutes
		usageSyncInterval:    5 * time.Minute,  // Sync usage every 5 minutes
		tokenRefreshInterval: 30 * time.Minute, // Check token refresh every 30 minutes
	}
}

// Start begins the background sync processes
func (bs *BackgroundSyncService) Start() error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if bs.running {
		return nil // Already running
	}

	bs.ctx, bs.cancel = context.WithCancel(context.Background())
	bs.running = true

	// Start config sync goroutine
	bs.wg.Add(1)
	go bs.configSyncWorker()

	// Start usage sync goroutine
	bs.wg.Add(1)
	go bs.usageSyncWorker()

	// Start token refresh goroutine
	bs.wg.Add(1)
	go bs.tokenRefreshWorker()

	slog.Info("started background sync service")
	return nil
}

// Stop gracefully stops all background sync processes
func (bs *BackgroundSyncService) Stop() {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if !bs.running {
		return
	}

	slog.Info("stopping background sync service")
	bs.cancel()
	bs.wg.Wait()
	bs.running = false

	slog.Info("background sync service stopped")
}

// IsRunning returns whether the background sync service is running
func (bs *BackgroundSyncService) IsRunning() bool {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	return bs.running
}

// SetConfigSyncInterval sets the interval for config sync
func (bs *BackgroundSyncService) SetConfigSyncInterval(interval time.Duration) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.configSyncInterval = interval
}

// SetUsageSyncInterval sets the interval for usage sync
func (bs *BackgroundSyncService) SetUsageSyncInterval(interval time.Duration) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.usageSyncInterval = interval
}

// SyncNow triggers an immediate sync of both config and usage
func (bs *BackgroundSyncService) SyncNow() (configSynced bool, usageSynced int, err error) {
	// Sync config
	configErr := bs.syncConfig()
	if configErr != nil {
		slog.Error("config sync failed", "error", configErr)
	} else {
		configSynced = true
		slog.Info("config sync completed successfully")
	}

	// Sync usage
	count, usageErr := bs.syncUsage()
	if usageErr != nil {
		slog.Error("usage sync failed", "error", usageErr)
	} else {
		usageSynced = count
		slog.Info("usage sync completed", "count", count)
	}

	// Return combined error if all failed
	if configErr != nil && usageErr != nil {
		err = fmt.Errorf("config sync: %w, usage sync: %w", configErr, usageErr)
	} else if configErr != nil {
		err = fmt.Errorf("config sync: %w", configErr)
	} else if usageErr != nil {
		err = fmt.Errorf("usage sync: %w", usageErr)
	}

	return
}

// configSyncWorker runs periodic config synchronization
func (bs *BackgroundSyncService) configSyncWorker() {
	defer bs.wg.Done()

	// Initial sync on startup
	bs.trySyncConfig("initial")

	ticker := time.NewTicker(bs.getConfigSyncInterval())
	defer ticker.Stop()

	for {
		select {
		case <-bs.ctx.Done():
			slog.Info("config sync worker stopped")
			return
		case <-ticker.C:
			bs.trySyncConfig("periodic")
		}
	}
}

// usageSyncWorker runs periodic usage synchronization
func (bs *BackgroundSyncService) usageSyncWorker() {
	defer bs.wg.Done()

	ticker := time.NewTicker(bs.getUsageSyncInterval())
	defer ticker.Stop()

	for {
		select {
		case <-bs.ctx.Done():
			slog.Info("usage sync worker stopped")
			return
		case <-ticker.C:
			bs.trySyncUsage()
		}
	}
}

// trySyncConfig attempts to sync config with error handling
func (bs *BackgroundSyncService) trySyncConfig(trigger string) {
	// Ensure token is fresh before syncing
	if !bs.ensureAuthenticated() {
		slog.Debug("skipping config sync", "trigger", trigger, "reason", "not authenticated")
		return
	}

	if err := bs.syncConfig(); err != nil {
		slog.Error("config sync error", "trigger", trigger, "error", err)
	}
}

// trySyncUsage attempts to sync usage with error handling
func (bs *BackgroundSyncService) trySyncUsage() {
	// Ensure token is fresh before syncing
	if !bs.ensureAuthenticated() {
		slog.Debug("skipping usage sync", "reason", "not authenticated")
		return
	}

	count, err := bs.syncUsage()
	if err != nil {
		slog.Error("usage sync error", "error", err)
	} else if count > 0 {
		slog.Info("synced usage records", "count", count)
	}
}

// syncConfig performs the actual config sync
func (bs *BackgroundSyncService) syncConfig() error {
	return bs.configSyncService.SyncProviders()
}

// syncUsage performs the actual usage sync
func (bs *BackgroundSyncService) syncUsage() (int, error) {
	return bs.usageSyncService.SyncUsageStats()
}

// tokenRefreshWorker runs periodic token refresh checks
func (bs *BackgroundSyncService) tokenRefreshWorker() {
	defer bs.wg.Done()

	// Initial check on startup
	bs.tryRefreshToken()

	ticker := time.NewTicker(bs.getTokenRefreshInterval())
	defer ticker.Stop()

	for {
		select {
		case <-bs.ctx.Done():
			slog.Info("token refresh worker stopped")
			return
		case <-ticker.C:
			bs.tryRefreshToken()
		}
	}
}

// tryRefreshToken attempts to refresh the token if needed
func (bs *BackgroundSyncService) tryRefreshToken() {
	if !bs.authService.IsAuthenticated() {
		return // Not authenticated, nothing to refresh
	}

	// Check and refresh if needed
	refreshed := bs.authService.RefreshIfNeeded()
	if refreshed {
		slog.Info("token refreshed successfully")
	}
}

// ensureAuthenticated checks if authenticated and attempts refresh if needed
func (bs *BackgroundSyncService) ensureAuthenticated() bool {
	if !bs.authService.IsAuthenticated() {
		return false
	}
	// Refresh if needed
	bs.authService.RefreshIfNeeded()
	return true
}

// getConfigSyncInterval returns the current config sync interval
func (bs *BackgroundSyncService) getConfigSyncInterval() time.Duration {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	return bs.configSyncInterval
}

// getUsageSyncInterval returns the current usage sync interval
func (bs *BackgroundSyncService) getUsageSyncInterval() time.Duration {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	return bs.usageSyncInterval
}

// getTokenRefreshInterval returns the current token refresh interval
func (bs *BackgroundSyncService) getTokenRefreshInterval() time.Duration {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	return bs.tokenRefreshInterval
}

// SetTokenRefreshInterval sets the interval for token refresh
func (bs *BackgroundSyncService) SetTokenRefreshInterval(interval time.Duration) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.tokenRefreshInterval = interval
}
