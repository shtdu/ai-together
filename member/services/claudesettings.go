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
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	claudeSettingsDir      = ".claude"
	claudeSettingsFileName = "settings.json"
	claudeBackupFileName   = "ct-studio.back.settings.json"
	claudeAuthTokenValue   = "code-together"
)

type ClaudeProxyStatus struct {
	Enabled bool   `json:"enabled"`
	BaseURL string `json:"base_url"`
}

type ClaudeSettingsService struct {
	relayAddr string
}

func NewClaudeSettingsService(relayAddr string) *ClaudeSettingsService {
	return &ClaudeSettingsService{relayAddr: relayAddr}
}

func (css *ClaudeSettingsService) ProxyStatus() (ClaudeProxyStatus, error) {
	status := ClaudeProxyStatus{Enabled: false, BaseURL: css.baseURL()}
	settingsPath, _, err := css.paths()
	if err != nil {
		return status, err
	}
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return status, nil
		}
		return status, err
	}
	var settings claudeSettingsFile
	if err := json.Unmarshal(data, &settings); err != nil {
		return status, nil
	}

	env := getEnv(settings)
	authToken := getEnvString(env, "ANTHROPIC_AUTH_TOKEN")
	baseURL := getEnvString(env, "ANTHROPIC_BASE_URL")
	expectedBaseURL := css.baseURL()

	enabled := strings.EqualFold(authToken, claudeAuthTokenValue) &&
		strings.EqualFold(baseURL, expectedBaseURL)
	status.Enabled = enabled
	return status, nil
}

func (css *ClaudeSettingsService) EnableProxy() error {
	settingsPath, backupPath, err := css.paths()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return err
	}

	// Read existing config if it exists, otherwise start with empty map
	settings := make(claudeSettingsFile)
	if _, err := os.Stat(settingsPath); err == nil {
		content, readErr := os.ReadFile(settingsPath)
		if readErr != nil {
			return readErr
		}
		// Backup existing config before modifying
		if err := os.WriteFile(backupPath, content, 0o600); err != nil {
			return err
		}
		// Unmarshal existing config to preserve ALL fields
		if err := json.Unmarshal(content, &settings); err != nil {
			// If unmarshal fails, start with empty map (will overwrite corrupted file)
			settings = make(claudeSettingsFile)
		}
	}

	// Merge Code Together hooks into settings
	settings = mergeCTHooks(settings)

	// Get or create env map
	env := getEnv(settings)

	// Set proxy settings (merges with existing env vars)
	env["ANTHROPIC_AUTH_TOKEN"] = claudeAuthTokenValue
	env["ANTHROPIC_BASE_URL"] = css.baseURL()

	payload, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, payload, 0o600)
}

func (css *ClaudeSettingsService) DisableProxy() error {
	settingsPath, backupPath, err := css.paths()
	if err != nil {
		return err
	}
	if err := os.Remove(settingsPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if _, err := os.Stat(backupPath); err == nil {
		if err := os.Rename(backupPath, settingsPath); err != nil {
			return err
		}
	} else if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return nil
}

func (css *ClaudeSettingsService) paths() (settingsPath string, backupPath string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	dir := filepath.Join(home, claudeSettingsDir)
	return filepath.Join(dir, claudeSettingsFileName), filepath.Join(dir, claudeBackupFileName), nil
}

func (css *ClaudeSettingsService) baseURL() string {
	addr := strings.TrimSpace(css.relayAddr)
	if addr == "" {
		addr = ":18100"
	}
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return addr
	}
	host := addr
	if strings.HasPrefix(host, ":") {
		host = "127.0.0.1" + host
	}
	if !strings.Contains(host, "://") {
		host = "http://" + host
	}
	return host
}

// getEnv retrieves the env map from settings, creating it if needed
func getEnv(settings claudeSettingsFile) map[string]any {
	if settings["env"] == nil {
		settings["env"] = make(map[string]any)
	}
	return settings["env"].(map[string]any)
}

// getEnvString retrieves a string value from env
func getEnvString(env map[string]any, key string) string {
	if val, ok := env[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// claudeSettingsFile represents the structure of Claude's settings.json
// Using map[string]any to preserve all fields (permissions, hooks, statusLine, enabledPlugins, etc.)
type claudeSettingsFile map[string]any

// getCTHookConfig returns the Code Together hook configuration template
// The hook command sends event data to the local proxy for collection
func getCTHookConfig() map[string]any {
	return map[string]any{
		"command": "jq -cM | curl -X POST http://localhost:18100/collect/claude -s -o /dev/null --show-error --fail -H \"Content-Type: application/json\" -d @-",
		"type":    "command",
	}
}

// shouldAddHook checks if a hook of the given type already exists
// Returns true if the hook should be added (no duplicate CT hook)
func shouldAddHook(settings map[string]any, hookType string) bool {
	hooks, ok := settings["hooks"]
	if !ok {
		return true
	}

	hooksMap, ok := hooks.(map[string]any)
	if !ok {
		return true
	}

	existingHooks, ok := hooksMap[hookType]
	if !ok {
		return true
	}

	// Check if there's already a CT hook for this event type
	hookList, ok := existingHooks.([]any)
	if !ok {
		return true
	}

	ctHookConfig := getCTHookConfig()
	ctCommand := ctHookConfig["command"].(string)

	for _, h := range hookList {
		hookEntry, ok := h.(map[string]any)
		if !ok {
			continue
		}

		// Handle Claude's wrapped hook format: {"matcher": "*", "hooks": [...]}
		if nestedHooks, hasNested := hookEntry["hooks"]; hasNested {
			nestedList, ok := nestedHooks.([]any)
			if !ok {
				continue
			}
			for _, nestedHook := range nestedList {
				hookMap, ok := nestedHook.(map[string]any)
				if !ok {
					continue
				}
				command, ok := hookMap["command"].(string)
				if ok && command == ctCommand {
					// Already have a CT hook with this command
					return false
				}
			}
		} else {
			// Handle simple format: {"command": "...", "type": "command"}
			command, ok := hookEntry["command"].(string)
			if ok && command == ctCommand {
				// Already have a CT hook with this command
				return false
			}
		}
	}

	return true
}

// mergeCTHooks merges Code Together hooks into existing settings
// Preserves all existing user hooks, only adds hooks that don't already exist
func mergeCTHooks(settings map[string]any) map[string]any {
	// Create a copy to avoid modifying the original
	result := make(claudeSettingsFile)
	for k, v := range settings {
		result[k] = v
	}

	// Initialize hooks map if it doesn't exist
	if result["hooks"] == nil {
		result["hooks"] = make(map[string]any)
	}

	hooks := result["hooks"].(map[string]any)

	// Hook event types that Code Together collects
	hookEvents := []string{
		"Notification",
		"PermissionRequest",
		"PostToolUse",
		"PostToolUseFailure",
		"PreCompact",
		"PreToolUse",
		"SessionEnd",
		"SessionStart",
		"Stop",
		"SubagentStart",
		"SubagentStop",
		"TaskCompleted",
		"TeammateIdle",
		"UserPromptSubmit",
	}

	ctHookConfig := getCTHookConfig()

	// Claude's hook format wraps each hook entry with matcher and hooks array
	ctHookEntry := []any{
		map[string]any{
			"matcher": "*",
			"hooks":   []any{ctHookConfig},
		},
	}

	for _, event := range hookEvents {
		if shouldAddHook(result, event) {
			// Get existing hooks for this event type
			existing, ok := hooks[event]
			if !ok {
				// No existing hooks, add CT hook
				hooks[event] = ctHookEntry
			} else {
				// Existing hooks present, append CT hook
				existingList, ok := existing.([]any)
				if !ok {
					// Not a list, replace with CT hook
					hooks[event] = ctHookEntry
				} else {
					// Append CT hook to existing list in wrapped format
					wrappedCTHook := map[string]any{
						"matcher": "*",
						"hooks":   []any{ctHookConfig},
					}
					newList := make([]any, len(existingList))
					copy(newList, existingList)
					newList = append(newList, wrappedCTHook)
					hooks[event] = newList
				}
			}
		}
	}

	return result
}
