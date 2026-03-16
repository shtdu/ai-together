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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"codeswitch/services"
)

// CLIConfig holds the resolved configuration for the CLI tool
type CLIConfig struct {
	ToolName  string
	AuthToken string
	ServerURL string
	Verbose   bool
	Source    string // "cli" or "settings"
}

// resolveConfig resolves configuration from CLI flags and .code-together settings
// Priority: If both -auth and -server are provided, use CLI flags
// Otherwise, load from .code-together settings
func resolveConfig(toolName, cliAuth, cliServer string, verbose bool) (*CLIConfig, error) {
	// Validate tool name first
	if !isValidTool(toolName) {
		return nil, fmt.Errorf("invalid tool name: %s (must be claude, codex, or opencode)", toolName)
	}

	config := &CLIConfig{
		ToolName: toolName,
		Verbose:  verbose,
		Source:   "cli",
	}

	// Validate: auth and server must be specified together
	hasCLIAuth := cliAuth != ""
	hasCLIServer := cliServer != ""

	if hasCLIAuth != hasCLIServer {
		return nil, fmt.Errorf("both -auth and -server must be specified together (got auth=%v, server=%v)",
			hasCLIAuth, hasCLIServer)
	}

	// If both CLI flags provided, use them
	if hasCLIAuth && hasCLIServer {
		config.AuthToken = cliAuth
		config.ServerURL = cliServer
		config.Source = "cli"
		return config, nil
	}

	// Otherwise, load from .code-together settings
	configService := services.NewConfigService()

	// Load server URL from config.json
	serverConfig, err := configService.GetServer()
	if err != nil {
		return nil, fmt.Errorf("server URL not found in settings: %w\nProvide both -auth and -server flags", err)
	}
	if serverConfig.ServerURL == "" {
		return nil, fmt.Errorf("server URL not configured in settings\nProvide both -auth and -server flags")
	}
	config.ServerURL = serverConfig.ServerURL

	// Load auth token from auth.json
	home, _ := os.UserHomeDir()
	authPath := filepath.Join(home, ".code-together", "auth.json")
	data, err := os.ReadFile(authPath)
	if err != nil {
		return nil, fmt.Errorf("auth token not found in settings: %w\nProvide both -auth and -server flags", err)
	}

	var authConfig struct {
		Tokens struct {
			AccessToken string `json:"access_token"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(data, &authConfig); err != nil {
		return nil, fmt.Errorf("failed to parse auth settings: %w\nProvide both -auth and -server flags", err)
	}
	if authConfig.Tokens.AccessToken == "" {
		return nil, fmt.Errorf("auth token is empty in settings\nProvide both -auth and -server flags")
	}
	config.AuthToken = authConfig.Tokens.AccessToken
	config.Source = "settings"

	return config, nil
}

// isValidTool checks if the tool name is valid
func isValidTool(toolName string) bool {
	switch toolName {
	case "claude", "codex", "opencode":
		return true
	default:
		return false
	}
}
