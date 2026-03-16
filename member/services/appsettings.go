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
	"sync"
)

type AppSettings struct {
	ShowHeatmap   bool `json:"show_heatmap"`
	ShowHomeTitle bool `json:"show_home_title"`
	Show24hStats  bool `json:"show_24h_stats"`
	AutoStart     bool `json:"auto_start"`
}

type AppSettingsService struct {
	configService    *ConfigService
	mu               sync.Mutex
	autoStartService *AutoStartService
}

func NewAppSettingsService(configService *ConfigService, autoStartService *AutoStartService) *AppSettingsService {
	return &AppSettingsService{
		configService:    configService,
		autoStartService: autoStartService,
	}
}

func (as *AppSettingsService) defaultSettings() AppSettings {
	// 检查当前开机自启动状态
	autoStartEnabled := false
	if as.autoStartService != nil {
		if enabled, err := as.autoStartService.IsEnabled(); err == nil {
			autoStartEnabled = enabled
		}
	}

	return AppSettings{
		ShowHeatmap:   true,
		ShowHomeTitle: true,
		Show24hStats:  false,
		AutoStart:     autoStartEnabled,
	}
}

// GetAppSettings returns the persisted app settings or defaults if the file does not exist.
func (as *AppSettingsService) GetAppSettings() (AppSettings, error) {
	as.mu.Lock()
	defer as.mu.Unlock()

	settings, err := as.configService.GetApp()
	if err != nil {
		return as.defaultSettings(), err
	}

	// Return the settings as-is. If the config file doesn't exist,
	// ConfigService.GetApp() will return default values, so we don't need
	// to check for "empty" states here. All-false is a valid user preference.
	return settings, nil
}

// SaveAppSettings persists the provided settings to disk.
func (as *AppSettingsService) SaveAppSettings(settings AppSettings) (AppSettings, error) {
	as.mu.Lock()
	defer as.mu.Unlock()

	// 同步开机自启动状态
	if as.autoStartService != nil {
		if settings.AutoStart {
			if err := as.autoStartService.Enable(); err != nil {
				return settings, err
			}
		} else {
			if err := as.autoStartService.Disable(); err != nil {
				return settings, err
			}
		}
	}

	if err := as.configService.SetApp(settings); err != nil {
		return settings, err
	}
	return settings, nil
}
