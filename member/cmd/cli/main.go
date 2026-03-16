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

package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"codeswitch/services"
)

func main() {
	// Parse CLI flags
	var (
		toolName  = flag.String("tool-name", "", "Tool to configure (claude, codex, opencode)")
		authToken = flag.String("auth", "", "JWT auth token (optional if .code-together/auth.json exists)")
		serverURL = flag.String("server", "", "Server URL (optional if .code-together/config.json exists)")
		verbose   = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Validate required flags before setting up logger
	if *toolName == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Setup logger
	setupLogger(*verbose)

	// Resolve configuration (CLI flags > .code-together settings)
	config, err := resolveConfig(*toolName, *authToken, *serverURL, *verbose)
	if err != nil {
		slog.Error("configuration resolution failed", "error", err)
		os.Exit(1)
	}

	slog.Info("configuration resolved",
		"tool", config.ToolName,
		"server", config.ServerURL,
		"source", config.Source)

	// Load providers from server
	slog.Info("loading providers from server", "server", config.ServerURL)
	providersByType, err := loadProvidersFromServer(config.ServerURL, config.AuthToken, config.Verbose)
	if err != nil {
		slog.Error("failed to load providers from server", "error", err)
		os.Exit(1)
	}

	// Create provider service and pre-load providers from server
	providerService := services.NewProviderService()
	for kind, providers := range providersByType {
		if err := providerService.ReplaceProviders(kind, providers); err != nil {
			slog.Error("failed to load providers", "kind", kind, "error", err)
			os.Exit(1)
		}
	}

	// Initialize hook service
	hookService := services.NewHookService()
	defer hookService.Close()

	// Initialize provider relay
	providerRelay := services.NewProviderRelayService(providerService, hookService, ":18100", config.Verbose)

	// Get settings service based on tool name
	settingsService := getSettingsService(config.ToolName, providerRelay.Addr())
	if settingsService == nil {
		slog.Error("unknown tool name", "tool", config.ToolName)
		os.Exit(1)
	}

	// Check and log current status
	enabled := settingsService.IsProxyEnabled()
	slog.Info("current proxy status", "enabled", enabled)

	// Enable proxy (this also backs up existing settings internally)
	if err := settingsService.EnableProxy(); err != nil {
		slog.Error("failed to enable proxy", "error", err)
		os.Exit(1)
	}
	slog.Info("proxy enabled successfully")

	// Start services
	if err := providerRelay.Start(); err != nil {
		slog.Error("failed to start relay", "error", err)
		settingsService.DisableProxy() // Restore from backup
		os.Exit(1)
	}

	slog.Info("CLI tool started successfully",
		"tool", config.ToolName,
		"proxy", providerRelay.Addr(),
		"config_source", config.Source)

	// Setup signal handling for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	// Block until signal received
	sig := <-sigs
	slog.Info("received shutdown signal", "signal", sig)

	// Stop accepting new requests
	if err := providerRelay.Stop(); err != nil {
		slog.Warn("failed to stop relay cleanly", "error", err)
	}

	// Give pending requests a moment to complete
	time.Sleep(100 * time.Millisecond)

	// Disable proxy (this also restores from backup internally)
	if err := settingsService.DisableProxy(); err != nil {
		slog.Warn("failed to disable proxy", "error", err)
	}

	// Close hook service
	hookService.Close()

	slog.Info("shutdown complete")
}

// settingsService provides a unified interface for tool-specific settings operations
type settingsService interface {
	IsProxyEnabled() bool
	EnableProxy() error
	DisableProxy() error
}

// claudeSettingsWrapper wraps ClaudeSettingsService
type claudeSettingsWrapper struct {
	service *services.ClaudeSettingsService
}

func (w *claudeSettingsWrapper) IsProxyEnabled() bool {
	status, err := w.service.ProxyStatus()
	return err == nil && status.Enabled
}

func (w *claudeSettingsWrapper) EnableProxy() error {
	return w.service.EnableProxy()
}

func (w *claudeSettingsWrapper) DisableProxy() error {
	return w.service.DisableProxy()
}

// codexSettingsWrapper wraps CodexSettingsService
type codexSettingsWrapper struct {
	service *services.CodexSettingsService
}

func (w *codexSettingsWrapper) IsProxyEnabled() bool {
	status, err := w.service.ProxyStatus()
	return err == nil && status.Enabled
}

func (w *codexSettingsWrapper) EnableProxy() error {
	return w.service.EnableProxy()
}

func (w *codexSettingsWrapper) DisableProxy() error {
	return w.service.DisableProxy()
}

// opencodeSettingsWrapper wraps OpenCodeSettingsService
type opencodeSettingsWrapper struct {
	service *services.OpenCodeSettingsService
}

func (w *opencodeSettingsWrapper) IsProxyEnabled() bool {
	status, err := w.service.ProxyStatus()
	return err == nil && status.Enabled
}

func (w *opencodeSettingsWrapper) EnableProxy() error {
	return w.service.EnableProxy()
}

func (w *opencodeSettingsWrapper) DisableProxy() error {
	return w.service.DisableProxy()
}

// getSettingsService returns the appropriate settings wrapper for the tool
func getSettingsService(toolName string, proxyURL string) settingsService {
	switch toolName {
	case "claude":
		return &claudeSettingsWrapper{service: services.NewClaudeSettingsService(proxyURL)}
	case "codex":
		return &codexSettingsWrapper{service: services.NewCodexSettingsService(proxyURL)}
	case "opencode":
		return &opencodeSettingsWrapper{service: services.NewOpenCodeSettingsService(proxyURL)}
	default:
		return nil
	}
}

// setupLogger configures the global logger based on verbose flag
func setupLogger(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{
		Level: level,
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(logger)
}
