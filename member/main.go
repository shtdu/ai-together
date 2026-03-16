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

package main

import (
	"codeswitch/internal/hookdb"
	"codeswitch/services"
	"context"
	"embed"
	_ "embed"
	"fmt"
	"log"
	"log/slog"
	"runtime"
	"time"

	"github.com/code-together/shared/integration"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/services/dock"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

//go:embed assets/icon.png assets/icon-dark.png
var trayIcons embed.FS

type AppService struct {
	App *application.App
}

func (a *AppService) SetApp(app *application.App) {
	a.App = app
}

func (a *AppService) OpenSecondWindow() {
	if a.App == nil {
		fmt.Println("[ERROR] app not initialized")
		return
	}
	name := fmt.Sprintf("logs-%d", time.Now().UnixNano())
	win := a.App.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "Logs",
		Name:      name,
		Width:     1024,
		Height:    800,
		MinWidth:  600,
		MinHeight: 300,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			TitleBar:                application.MacTitleBarHidden,
			Backdrop:                application.MacBackdropTransparent,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/#/logs",
	})
	win.Center()
}

func (a *AppService) RestartApp() {
	if a.App == nil {
		fmt.Println("[ERROR] app not initialized")
		return
	}
	a.App.Quit()
}

// main function serves as the application's entry point. It initializes the application, creates a window,
// and starts a goroutine that emits a time-based event every second. It subsequently runs the application and
// logs any error that might occur.
func main() {
	appservice := &AppService{}

	// Create config service first (other services depend on it)
	configService := services.NewConfigService()

	suiService, errt := services.NewSuiStore()
	if errt != nil {
		// 处理错误，比如日志或退出
	}

	providerService := services.NewProviderService()
	hookService := services.NewHookService()
	providerRelay := services.NewProviderRelayService(providerService, hookService, ":18100", false) // GUI mode defaults to non-verbose
	claudeSettings := services.NewClaudeSettingsService(providerRelay.Addr())
	codexSettings := services.NewCodexSettingsService(providerRelay.Addr())
	opencodeSettings := services.NewOpenCodeSettingsService(providerRelay.Addr())
	opencodeSettings.SetProviderService(providerService)
	logService := services.NewLogService()
	autoStartService := services.NewAutoStartService()
	appSettings := services.NewAppSettingsService(configService, autoStartService)
	mcpService := services.NewMCPService()
	skillService := services.NewSkillService()
	importService := services.NewImportService(providerService, mcpService)
	reportService := &services.ReportService{}
	dockService := dock.New()
	versionService := NewVersionService()

	// Server integration services
	logger := slog.Default()

	// Create server config service (now uses configService)
	serverConfigService := services.NewServerConfigService(configService, logger)

	// Get server URL for client creation
	serverURL, err := serverConfigService.GetServerURL()
	if err != nil {
		log.Printf("Warning: failed to get server URL: %v (server integration will be disabled)", err)
		// Continue without server integration
		serverURL = ""
	}

	var authService *services.AuthService
	var configSyncService *services.ConfigSyncService
	var usageSyncService *services.UsageSyncService
	var permissionService *services.PermissionService

	// Only create server integration services if server URL is configured
	if serverURL != "" {
		// Create anonymous API client for auth operations (login/register/refresh)
		anonClient, err := integration.NewAnonymousClientImpl(serverURL, logger)
		if err != nil {
			log.Printf("Warning: failed to create anonymous API client: %v", err)
		} else {
			authService = services.NewAuthService(anonClient, serverConfigService)

			// Create authenticated API client for other operations
			// The client will get tokens from AuthService
			authClient, err := integration.NewAuthenticatedClientImpl(serverURL, authService.GetAccessToken, logger)
			if err != nil {
				log.Printf("Warning: failed to create authenticated API client: %v", err)
			} else {
				// Set the anonymous API client on server config service for health checks
				// Health endpoint doesn't require authentication
				serverConfigService.SetAPIClient(anonClient)

				// Create other services that need the authenticated client
				permissionService = services.NewPermissionService(authService)
				providerService.SetPermissionService(permissionService)
				configSyncService = services.NewConfigSyncService(authService, authClient, providerService, importService, configService)
				usageSyncService = services.NewUsageSyncService(authService, authClient)
			}
		}
	}

	// Create background sync service (will be nil if server integration is disabled)
	var backgroundSyncService *services.BackgroundSyncService
	if configSyncService != nil && usageSyncService != nil && authService != nil {
		backgroundSyncService = services.NewBackgroundSyncService(configSyncService, usageSyncService, authService, serverConfigService)
	}

	go func() {
		if err := providerRelay.Start(); err != nil {
			log.Printf("provider relay start error: %v", err)
		}
	}()

	// Initialize hook events database for reports
	if err := hookdb.Init(); err != nil {
		log.Printf("Warning: failed to initialize hook database: %v (reports will be disabled)", err)
	}

	// Create service list (conditionally add server integration services)
	servicesList := []application.Service{
		application.NewService(appservice),
		application.NewService(suiService),
		application.NewService(providerService),
		application.NewService(claudeSettings),
		application.NewService(codexSettings),
		application.NewService(opencodeSettings),
		application.NewService(logService),
		application.NewService(appSettings),
		application.NewService(mcpService),
		application.NewService(skillService),
		application.NewService(importService),
		application.NewService(reportService),
		application.NewService(dockService),
		application.NewService(versionService),
		application.NewService(configService),
		application.NewService(serverConfigService),
	}

	// Add server integration services if they were created
	if authService != nil {
		servicesList = append(servicesList, application.NewService(authService))
	}
	if permissionService != nil {
		servicesList = append(servicesList, application.NewService(permissionService))
	}
	if configSyncService != nil {
		servicesList = append(servicesList, application.NewService(configSyncService))
	}
	if usageSyncService != nil {
		servicesList = append(servicesList, application.NewService(usageSyncService))
	}
	if backgroundSyncService != nil {
		servicesList = append(servicesList, application.NewService(backgroundSyncService))
	}

	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.
	app := application.New(application.Options{
		Name:        "Code Together",
		Description: "Claude Code and Codex provier manager",
		Services:    servicesList,
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	// Set minimal menu to show only app name (Code Together) on macOS
	menu := application.NewMenu()
	if runtime.GOOS == "darwin" {
		menu.AddRole(application.AppMenu)  // Adds "Code Together" menu with About, Preferences, Quit
		menu.AddRole(application.EditMenu) // Adds Edit menu with Undo, Redo, Cut, Copy, Paste, Select All
	}
	app.Menu.Set(menu)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app.OnShutdown(func() {
		cancel() // Cancel context to stop all goroutines
		_ = providerRelay.Stop()
		if backgroundSyncService != nil {
			backgroundSyncService.Stop()
		}
		// Close hook service (including database)
		_ = hookService.Close()
	})

	// Start background sync service (only if it was created)
	if backgroundSyncService != nil {
		if err := backgroundSyncService.Start(); err != nil {
			log.Printf("Failed to start background sync service: %v", err)
		}
	}

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	mainWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "Code Together",
		Width:     1024,
		Height:    800,
		MinWidth:  600,
		MinHeight: 300,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})
	var mainWindowCentered bool
	focusMainWindow := func() {
		if runtime.GOOS == "windows" {
			mainWindow.SetAlwaysOnTop(true)
			mainWindow.Focus()
			go func() {
				time.Sleep(150 * time.Millisecond)
				mainWindow.SetAlwaysOnTop(false)
			}()
			return
		}
		mainWindow.Focus()
	}
	showMainWindow := func(withFocus bool) {
		if !mainWindowCentered {
			mainWindow.Center()
			mainWindowCentered = true
		}
		if mainWindow.IsMinimised() {
			mainWindow.UnMinimise()
		}
		mainWindow.Show()
		if withFocus {
			focusMainWindow()
		}
		handleDockVisibility(dockService, true)
	}
	showMainWindow(false)

	mainWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		mainWindow.Hide()
		handleDockVisibility(dockService, false)
		e.Cancel()
	})

	app.Event.OnApplicationEvent(events.Mac.ApplicationShouldHandleReopen, func(event *application.ApplicationEvent) {
		showMainWindow(true)
	})

	app.Event.OnApplicationEvent(events.Mac.ApplicationDidBecomeActive, func(event *application.ApplicationEvent) {
		if mainWindow.IsVisible() {
			mainWindow.Focus()
			return
		}
		showMainWindow(true)
	})

	systray := app.SystemTray.New()
	// systray.SetLabel("Code Together")
	systray.SetTooltip("Code Together")
	if lightIcon := loadTrayIcon("assets/icon.png"); len(lightIcon) > 0 {
		systray.SetIcon(lightIcon)
	}
	if darkIcon := loadTrayIcon("assets/icon-dark.png"); len(darkIcon) > 0 {
		systray.SetDarkModeIcon(darkIcon)
	}

	trayMenu := application.NewMenu()
	trayMenu.Add("显示主窗口").OnClick(func(ctx *application.Context) {
		showMainWindow(true)
	})
	trayMenu.AddSeparator()

	// Tool Toggles
	claudeItem := trayMenu.AddCheckbox("Claude Code", false)
	claudeItem.OnClick(func(ctx *application.Context) {
		status, _ := claudeSettings.ProxyStatus()
		if status.Enabled {
			_ = claudeSettings.DisableProxy()
		} else {
			_ = claudeSettings.EnableProxy()
		}
		// Read the actual state after toggle to handle errors correctly
		if newStatus, err := claudeSettings.ProxyStatus(); err == nil {
			claudeItem.SetChecked(newStatus.Enabled)
		}
	})

	codexItem := trayMenu.AddCheckbox("Codex", false)
	codexItem.OnClick(func(ctx *application.Context) {
		status, _ := codexSettings.ProxyStatus()
		if status.Enabled {
			_ = codexSettings.DisableProxy()
		} else {
			_ = codexSettings.EnableProxy()
		}
		// Read the actual state after toggle to handle errors correctly
		if newStatus, err := codexSettings.ProxyStatus(); err == nil {
			codexItem.SetChecked(newStatus.Enabled)
		}
	})

	openCodeItem := trayMenu.AddCheckbox("OpenCode", false)
	openCodeItem.OnClick(func(ctx *application.Context) {
		status, _ := opencodeSettings.ProxyStatus()
		if status.Enabled {
			_ = opencodeSettings.DisableProxy()
		} else {
			_ = opencodeSettings.EnableProxy()
		}
		// Read the actual state after toggle to handle errors correctly
		if newStatus, err := opencodeSettings.ProxyStatus(); err == nil {
			openCodeItem.SetChecked(newStatus.Enabled)
		}
	})

	trayMenu.AddSeparator()
	trayMenu.Add("退出").OnClick(func(ctx *application.Context) {
		app.Quit()
	})
	systray.SetMenu(trayMenu)

	systray.OnClick(func() {
		if !mainWindow.IsVisible() {
			showMainWindow(true)
			return
		}
		if !mainWindow.IsFocused() {
			focusMainWindow()
		}
	})

	appservice.SetApp(app)

	// Create a goroutine that emits an event containing the current time every second.
	// The frontend can listen to this event and update the UI accordingly.
	// Also poling tray menu status
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return // Exit goroutine when context is cancelled
			case <-ticker.C:
				// Update Claude Status
				if s, err := claudeSettings.ProxyStatus(); err == nil {
					claudeItem.SetChecked(s.Enabled)
				}
				// Update Codex Status
				if s, err := codexSettings.ProxyStatus(); err == nil {
					codexItem.SetChecked(s.Enabled)
				}
				// Update OpenCode Status
				if s, err := opencodeSettings.ProxyStatus(); err == nil {
					openCodeItem.SetChecked(s.Enabled)
				}
			}
		}
	}()

	// Run the application. This blocks until the application has been exited.
	err = app.Run()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}

func loadTrayIcon(path string) []byte {
	data, err := trayIcons.ReadFile(path)
	if err != nil {
		log.Printf("failed to load tray icon %s: %v", path, err)
		return nil
	}
	return data
}

func handleDockVisibility(service *dock.DockService, show bool) {
	if runtime.GOOS != "darwin" || service == nil {
		return
	}
	if show {
		service.ShowAppIcon()
	} else {
		service.HideAppIcon()
	}
}
