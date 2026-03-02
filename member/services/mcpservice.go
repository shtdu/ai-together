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
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/pelletier/go-toml/v2"
)

const (
	mcpStoreDir        = ".code-together"
	mcpStoreFile       = "mcp.json"
	claudeMcpFile      = ".claude.json"
	codexDirName       = ".codex"
	codexConfigFile    = "config.toml"
	opencodeDirName    = ".config/opencode"
	opencodeConfigFile = "opencode.json"
	platClaudeCode     = "claude-code"
	platCodex          = "codex"
	platOpenCode       = "opencode"
)

var builtInServers = map[string]rawMCPServer{
	"reftools": {
		Type:    "http",
		URL:     "https://api.ref.tools/mcp?apiKey={apiKey}",
		Website: "https://ref.tools",
		Tips:    "Visit ref.tools to claim your API key.",
	},
	"chrome-devtools": {
		Type:    "stdio",
		Command: "npx",
		Args:    []string{"-y", "chrome-devtools-mcp@latest"},
		Tips:    "Needs Node.js. Run once to install dependencies.",
	},
}

var placeholderPattern = regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)

type MCPService struct {
	mu sync.Mutex
}

func NewMCPService() *MCPService {
	return &MCPService{}
}

type MCPServer struct {
	Name                string            `json:"name"`
	Type                string            `json:"type"`
	Command             string            `json:"command,omitempty"`
	Args                []string          `json:"args,omitempty"`
	Env                 map[string]string `json:"env,omitempty"`
	URL                 string            `json:"url,omitempty"`
	Website             string            `json:"website,omitempty"`
	Tips                string            `json:"tips,omitempty"`
	Headers             map[string]string `json:"headers,omitempty"`
	EnablePlatform      []string          `json:"enable_platform"`
	EnabledInClaude     bool              `json:"enabled_in_claude"`
	EnabledInCodex      bool              `json:"enabled_in_codex"`
	EnabledInOpenCode   bool              `json:"enabled_in_opencode"`
	MissingPlaceholders []string          `json:"missing_placeholders"`
}

type rawMCPServer struct {
	Type           string            `json:"type"`
	Command        string            `json:"command,omitempty"`
	Args           []string          `json:"args,omitempty"`
	Env            map[string]string `json:"env,omitempty"`
	URL            string            `json:"url,omitempty"`
	Website        string            `json:"website,omitempty"`
	Tips           string            `json:"tips,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	EnablePlatform []string          `json:"enable_platform"`
}

type claudeMcpFilePayload struct {
	Servers map[string]json.RawMessage `json:"mcpServers"`
}

type codexMcpFilePayload struct {
	Servers map[string]map[string]any `toml:"mcp_servers"`
}

type claudeDesktopServer struct {
	Type    string            `json:"type,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
}

func (ms *MCPService) ListServers() ([]MCPServer, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	config, err := ms.loadConfig()
	if err != nil {
		return nil, err
	}

	claudeEnabled := loadClaudeEnabledServers()
	codexEnabled := loadCodexEnabledServers()
	opencodeEnabled := loadOpenCodeEnabledServers()

	names := make([]string, 0, len(config))
	for name := range config {
		names = append(names, name)
	}
	sort.Strings(names)

	servers := make([]MCPServer, 0, len(names))
	for _, name := range names {
		entry := config[name]
		typ := normalizeServerType(entry.Type)
		platforms := normalizePlatforms(entry.EnablePlatform)
		server := MCPServer{
			Name:              name,
			Type:              typ,
			Command:           strings.TrimSpace(entry.Command),
			Args:              cloneArgs(entry.Args),
			Env:               cloneEnv(entry.Env),
			URL:               strings.TrimSpace(entry.URL),
			Website:           strings.TrimSpace(entry.Website),
			Tips:              strings.TrimSpace(entry.Tips),
			Headers:           cloneEnv(entry.Headers),
			EnablePlatform:    platforms,
			EnabledInClaude:   containsNormalized(claudeEnabled, name),
			EnabledInCodex:    containsNormalized(codexEnabled, name),
			EnabledInOpenCode: containsNormalized(opencodeEnabled, name),
		}
		server.MissingPlaceholders = detectPlaceholders(server.URL, server.Args)
		servers = append(servers, server)
	}

	return servers, nil
}

func (ms *MCPService) SaveServers(servers []MCPServer) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	normalized := make([]MCPServer, len(servers))
	raw := make(map[string]rawMCPServer, len(servers))
	for i := range servers {
		server := servers[i]
		name := strings.TrimSpace(server.Name)
		if name == "" {
			return fmt.Errorf("server name 不能为空")
		}
		typ := normalizeServerType(server.Type)
		platforms := normalizePlatforms(server.EnablePlatform)
		args := cleanArgs(server.Args)
		env := cleanEnv(server.Env)
		command := strings.TrimSpace(server.Command)
		url := strings.TrimSpace(server.URL)
		if typ == "stdio" && command == "" {
			return fmt.Errorf("%s 需要提供 command", name)
		}
		if typ == "http" && url == "" {
			return fmt.Errorf("%s 需要提供 url", name)
		}
		normalized[i] = MCPServer{
			Name:              name,
			Type:              typ,
			Command:           command,
			Args:              args,
			Env:               env,
			URL:               url,
			Website:           strings.TrimSpace(server.Website),
			Tips:              strings.TrimSpace(server.Tips),
			Headers:           cleanEnv(server.Headers),
			EnablePlatform:    platforms,
			EnabledInClaude:   server.EnabledInClaude,
			EnabledInCodex:    server.EnabledInCodex,
			EnabledInOpenCode: server.EnabledInOpenCode,
		}
		raw[name] = rawMCPServer{
			Type:           typ,
			Command:        command,
			Args:           args,
			Env:            env,
			URL:            url,
			Website:        normalized[i].Website,
			Tips:           normalized[i].Tips,
			Headers:        normalized[i].Headers,
			EnablePlatform: platforms,
		}
		placeholders := detectPlaceholders(url, args)
		normalized[i].MissingPlaceholders = placeholders
		if len(placeholders) > 0 {
			normalized[i].EnablePlatform = []string{}
			rawEntry := raw[name]
			rawEntry.EnablePlatform = []string{}
			raw[name] = rawEntry
		}
	}

	if err := ms.saveConfig(raw); err != nil {
		return err
	}
	if err := ms.syncClaudeServers(normalized); err != nil {
		return err
	}
	if err := ms.syncCodexServers(normalized); err != nil {
		return err
	}
	if err := ms.syncOpenCodeServers(normalized); err != nil {
		return err
	}
	return nil
}

func (ms *MCPService) configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, mcpStoreDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, mcpStoreFile), nil
}

func (ms *MCPService) loadConfig() (map[string]rawMCPServer, error) {
	path, err := ms.configPath()
	if err != nil {
		return nil, err
	}
	payload := map[string]rawMCPServer{}
	if data, err := os.ReadFile(path); err == nil {
		if len(data) > 0 {
			if err := json.Unmarshal(data, &payload); err != nil {
				return nil, err
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	for name, entry := range payload {
		payload[name] = normalizeRawEntry(entry)
	}

	changed := false
	if imported, err := ms.importFromClaude(payload); err == nil {
		if ms.mergeImportedServers(payload, imported) {
			changed = true
		}
	} else {
		return nil, err
	}

	if imported, err := ms.importFromOpenCode(payload); err == nil {
		if ms.mergeImportedServers(payload, imported) {
			changed = true
		}
	}
	// Ignore OpenCode import errors - it may not exist yet

	if ensureBuiltInServers(payload) {
		changed = true
	}

	if changed {
		if err := ms.saveConfig(payload); err != nil {
			return payload, err
		}
	}

	return payload, nil
}

func (ms *MCPService) importFromClaude(existing map[string]rawMCPServer) (map[string]rawMCPServer, error) {
	path, err := claudeConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]rawMCPServer{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return map[string]rawMCPServer{}, nil
	}
	var payload struct {
		Servers map[string]claudeDesktopServer `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	result := make(map[string]rawMCPServer, len(payload.Servers))
	for name, entry := range payload.Servers {
		trimmedName := strings.TrimSpace(name)
		if trimmedName == "" {
			continue
		}
		if _, exists := existing[trimmedName]; exists {
			continue
		}
		typeHint := entry.Type
		if strings.TrimSpace(typeHint) == "" {
			if strings.TrimSpace(entry.URL) != "" {
				typeHint = "http"
			}
		}
		if strings.TrimSpace(typeHint) == "" {
			typeHint = "stdio"
		}
		typ := normalizeServerType(typeHint)
		if typ == "http" && entry.URL == "" {
			continue
		}
		if typ == "stdio" && entry.Command == "" {
			continue
		}
		result[trimmedName] = rawMCPServer{
			Type:           typ,
			Command:        strings.TrimSpace(entry.Command),
			Args:           cleanArgs(entry.Args),
			Env:            cleanEnv(entry.Env),
			URL:            strings.TrimSpace(entry.URL),
			EnablePlatform: []string{platClaudeCode},
		}
	}
	return result, nil
}

func (ms *MCPService) importFromOpenCode(existing map[string]rawMCPServer) (map[string]rawMCPServer, error) {
	path, err := opencodeConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]rawMCPServer{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return map[string]rawMCPServer{}, nil
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	result := make(map[string]rawMCPServer)
	if mcpSection, ok := payload["mcp"].(map[string]interface{}); ok {
		for name, entry := range mcpSection {
			trimmedName := strings.TrimSpace(name)
			if trimmedName == "" {
				continue
			}
			// Don't skip existing servers - we need to import headers too
			serverEntry, ok := entry.(map[string]interface{})
			if !ok {
				continue
			}
			typStr, _ := serverEntry["type"].(string)
			var typ string
			switch strings.ToLower(strings.TrimSpace(typStr)) {
			case "local":
				typ = "stdio"
			case "remote":
				typ = "http"
			default:
				typ = normalizeServerType(typStr)
			}
			var command string
			var args []string
			if cmdArray, ok := serverEntry["command"].([]interface{}); ok && len(cmdArray) > 0 {
				command = strings.TrimSpace(getStringValue(cmdArray[0]))
				for i := 1; i < len(cmdArray); i++ {
					arg := strings.TrimSpace(getStringValue(cmdArray[i]))
					if arg != "" {
						args = append(args, arg)
					}
				}
			}
			var url string
			if u, ok := serverEntry["url"].(string); ok {
				url = strings.TrimSpace(u)
			}
			if typ == "http" && url == "" {
				continue
			}
			if typ == "stdio" && command == "" {
				continue
			}
			var env map[string]string
			if envMap, ok := serverEntry["environment"].(map[string]interface{}); ok {
				env = make(map[string]string, len(envMap))
				for k, v := range envMap {
					if strVal, ok := v.(string); ok {
						env[strings.TrimSpace(k)] = strVal
					}
				}
			}
			var headers map[string]string
			if headerMap, ok := serverEntry["headers"].(map[string]interface{}); ok {
				headers = make(map[string]string, len(headerMap))
				for k, v := range headerMap {
					if strVal, ok := v.(string); ok {
						headers[strings.TrimSpace(k)] = strVal
					}
				}
			}
			result[trimmedName] = rawMCPServer{
				Type:           typ,
				Command:        command,
				Args:           args,
				Env:            env,
				URL:            url,
				Headers:        headers,
				EnablePlatform: []string{platOpenCode},
			}
		}
	}
	return result, nil
}

func getStringValue(v interface{}) string {
	if str, ok := v.(string); ok {
		return str
	}
	if num, ok := v.(float64); ok {
		return fmt.Sprintf("%v", num)
	}
	return ""
}

func (ms *MCPService) saveConfig(payload map[string]rawMCPServer) error {
	path, err := ms.configPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func normalizeServerType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "http":
		return "http"
	default:
		return "stdio"
	}
}

func normalizePlatforms(values []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(values))
	for _, raw := range values {
		if platform, ok := normalizePlatform(raw); ok {
			if _, exists := seen[platform]; exists {
				continue
			}
			seen[platform] = struct{}{}
			result = append(result, platform)
		}
	}
	return result
}

func normalizePlatform(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "claude", "claude_code", "claude-code":
		return "claude-code", true
	case "codex":
		return "codex", true
	case "opencode", "open-code", "open_code":
		return "opencode", true
	default:
		return "", false
	}
}

func unionPlatforms(primary, secondary []string) []string {
	combined := append([]string{}, primary...)
	combined = append(combined, secondary...)
	return normalizePlatforms(combined)
}

func normalizeRawEntry(entry rawMCPServer) rawMCPServer {
	entry.Type = normalizeServerType(entry.Type)
	entry.Command = strings.TrimSpace(entry.Command)
	entry.URL = strings.TrimSpace(entry.URL)
	entry.Website = strings.TrimSpace(entry.Website)
	entry.Tips = strings.TrimSpace(entry.Tips)
	entry.Args = cleanArgs(entry.Args)
	entry.Env = cleanEnv(entry.Env)
	entry.EnablePlatform = normalizePlatforms(entry.EnablePlatform)
	return entry
}

func cloneArgs(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	dup := make([]string, len(values))
	copy(dup, values)
	return dup
}

func cloneEnv(values map[string]string) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}
	dup := make(map[string]string, len(values))
	for k, v := range values {
		dup[k] = v
	}
	return dup
}

func cleanArgs(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	result := make([]string, 0, len(values))
	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func cleanEnv(values map[string]string) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}
	result := make(map[string]string, len(values))
	for key, value := range values {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		result[trimmedKey] = strings.TrimSpace(value)
	}
	return result
}

func containsNormalized(pool map[string]struct{}, value string) bool {
	if len(pool) == 0 {
		return false
	}
	_, ok := pool[strings.ToLower(strings.TrimSpace(value))]
	return ok
}

func loadClaudeEnabledServers() map[string]struct{} {
	result := map[string]struct{}{}
	home, err := os.UserHomeDir()
	if err != nil {
		return result
	}
	path := filepath.Join(home, claudeMcpFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return result
	}
	var payload claudeMcpFilePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return result
	}
	for name := range payload.Servers {
		result[strings.ToLower(strings.TrimSpace(name))] = struct{}{}
	}
	return result
}

func loadCodexEnabledServers() map[string]struct{} {
	result := map[string]struct{}{}
	home, err := os.UserHomeDir()
	if err != nil {
		return result
	}
	path := filepath.Join(home, codexDirName, codexConfigFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return result
	}
	var payload codexMcpFilePayload
	if err := toml.Unmarshal(data, &payload); err != nil {
		return result
	}
	for name := range payload.Servers {
		result[strings.ToLower(strings.TrimSpace(name))] = struct{}{}
	}
	return result
}

func loadOpenCodeEnabledServers() map[string]struct{} {
	result := map[string]struct{}{}
	home, err := os.UserHomeDir()
	if err != nil {
		return result
	}
	path := filepath.Join(home, opencodeDirName, opencodeConfigFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return result
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return result
	}
	// OpenCode stores MCP servers under "mcp" key
	if servers, ok := payload["mcp"].(map[string]interface{}); ok {
		for name := range servers {
			result[strings.ToLower(strings.TrimSpace(name))] = struct{}{}
		}
	}
	return result
}

func (ms *MCPService) mergeImportedServers(target, imported map[string]rawMCPServer) bool {
	changed := false
	for name, entry := range imported {
		entry = normalizeRawEntry(entry)
		if existing, ok := target[name]; ok {
			entry.EnablePlatform = unionPlatforms(existing.EnablePlatform, entry.EnablePlatform)
			// Prefer imported values, fall back to existing if empty
			if entry.Website == "" {
				entry.Website = existing.Website
			}
			if entry.Tips == "" {
				entry.Tips = existing.Tips
			}
			if entry.Command == "" {
				entry.Command = existing.Command
			}
			if len(entry.Args) == 0 {
				entry.Args = existing.Args
			}
			if len(entry.Env) == 0 {
				entry.Env = existing.Env
			}
			if entry.URL == "" {
				entry.URL = existing.URL
			}
			// For headers, prefer imported values (from OpenCode config)
			// but merge with any existing headers that aren't in imported
			if len(existing.Headers) > 0 {
				if entry.Headers == nil {
					entry.Headers = make(map[string]string)
				}
				for k, v := range existing.Headers {
					if _, ok := entry.Headers[k]; !ok {
						entry.Headers[k] = v
					}
				}
			}
		}
		if existing, ok := target[name]; !ok || !reflect.DeepEqual(existing, entry) {
			target[name] = entry
			changed = true
		}
	}
	return changed
}

func ensureBuiltInServers(target map[string]rawMCPServer) bool {
	changed := false
	for name, builtIn := range builtInServers {
		builtIn = normalizeRawEntry(builtIn)
		if existing, ok := target[name]; ok {
			merged := existing
			merged.EnablePlatform = unionPlatforms(existing.EnablePlatform, builtIn.EnablePlatform)
			if merged.Command == "" {
				merged.Command = builtIn.Command
			}
			if len(merged.Args) == 0 {
				merged.Args = builtIn.Args
			}
			if len(merged.Env) == 0 {
				merged.Env = builtIn.Env
			}
			if merged.URL == "" {
				merged.URL = builtIn.URL
			}
			if merged.Website == "" {
				merged.Website = builtIn.Website
			}
			if merged.Tips == "" {
				merged.Tips = builtIn.Tips
			}
			merged = normalizeRawEntry(merged)
			if !reflect.DeepEqual(existing, merged) {
				target[name] = merged
				changed = true
			}
			continue
		}
		target[name] = builtIn
		changed = true
	}
	return changed
}

func (ms *MCPService) syncClaudeServers(servers []MCPServer) error {
	path, err := claudeConfigPath()
	if err != nil {
		return err
	}
	desired := make(map[string]claudeDesktopServer)
	for _, server := range servers {
		if !platformContains(server.EnablePlatform, platClaudeCode) {
			continue
		}
		desired[server.Name] = buildClaudeDesktopEntry(server)
	}
	payload := make(map[string]any)
	if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &payload); err != nil {
			payload = make(map[string]any)
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	payload["mcpServers"] = desired
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func (ms *MCPService) syncCodexServers(servers []MCPServer) error {
	path, err := codexConfigPath()
	if err != nil {
		return err
	}
	desired := make(map[string]map[string]any)
	for _, server := range servers {
		if !platformContains(server.EnablePlatform, platCodex) {
			continue
		}
		desired[server.Name] = buildCodexEntry(server)
	}
	payload := make(map[string]any)
	if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
		if err := toml.Unmarshal(data, &payload); err != nil {
			payload = make(map[string]any)
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	payload["mcp_servers"] = desired
	data, err := toml.Marshal(payload)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (ms *MCPService) syncOpenCodeServers(servers []MCPServer) error {
	path, err := opencodeConfigPath()
	if err != nil {
		return err
	}

	// Create a set of server names managed by our system
	managedServers := make(map[string]struct{})
	for _, server := range servers {
		managedServers[strings.ToLower(strings.TrimSpace(server.Name))] = struct{}{}
	}

	// Read existing config to preserve other fields AND manually configured MCP servers
	payload := make(map[string]any)
	if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &payload); err != nil {
			payload = make(map[string]any)
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	// Get or create mcp section, preserving ONLY manually configured entries
	mcpServers := make(map[string]map[string]any)
	if existingMcp, ok := payload["mcp"].(map[string]any); ok {
		for name, entry := range existingMcp {
			// Only preserve if it's NOT managed by our system (manually configured)
			nameLower := strings.ToLower(strings.TrimSpace(name))
			if _, isManaged := managedServers[nameLower]; !isManaged {
				if serverEntry, ok := entry.(map[string]any); ok {
					mcpServers[name] = serverEntry
				}
			}
		}
	}

	// Add/update servers that have OpenCode enabled in our system
	for _, server := range servers {
		if !platformContains(server.EnablePlatform, platOpenCode) {
			continue
		}
		mcpServers[server.Name] = buildOpenCodeEntry(server)
	}

	// Set MCP servers in config (OpenCode uses "mcp" key, not "mcpServers")
	payload["mcp"] = mcpServers
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func platformContains(platforms []string, target string) bool {
	for _, value := range platforms {
		if value == target {
			return true
		}
	}
	return false
}

func buildClaudeDesktopEntry(server MCPServer) claudeDesktopServer {
	entry := claudeDesktopServer{Type: server.Type}
	if server.Type == "http" {
		entry.URL = server.URL
	} else {
		entry.Command = server.Command
		if len(server.Args) > 0 {
			entry.Args = server.Args
		}
		if len(server.Env) > 0 {
			entry.Env = server.Env
		}
	}
	return entry
}

func buildCodexEntry(server MCPServer) map[string]any {
	entry := make(map[string]any)
	entry["type"] = server.Type
	if server.Type == "http" {
		entry["url"] = server.URL
	} else {
		entry["command"] = server.Command
		if len(server.Args) > 0 {
			entry["args"] = server.Args
		}
		if len(server.Env) > 0 {
			entry["env"] = server.Env
		}
	}
	return entry
}

func buildOpenCodeEntry(server MCPServer) map[string]any {
	entry := make(map[string]any)
	// Convert types: stdio → local, http → remote
	if server.Type == "http" {
		entry["type"] = "remote"
		entry["url"] = server.URL
		if len(server.Headers) > 0 {
			entry["headers"] = server.Headers
		}
	} else {
		entry["type"] = "local"
		// Build command array: [command, arg1, arg2, ...]
		cmdArray := []string{server.Command}
		cmdArray = append(cmdArray, server.Args...)
		entry["command"] = cmdArray
		if len(server.Env) > 0 {
			entry["environment"] = server.Env
		}
	}
	return entry
}

func claudeConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, claudeMcpFile), nil
}

func codexConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, codexDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, codexConfigFile), nil
}

func opencodeConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, opencodeDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, opencodeConfigFile), nil
}

func detectPlaceholders(url string, args []string) []string {
	set := make(map[string]struct{})
	collectPlaceholders(set, url)
	for _, arg := range args {
		collectPlaceholders(set, arg)
	}
	if len(set) == 0 {
		return []string{}
	}
	result := make([]string, 0, len(set))
	for key := range set {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func collectPlaceholders(set map[string]struct{}, value string) {
	if value == "" {
		return
	}
	matches := placeholderPattern.FindAllStringSubmatch(value, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		set[match[1]] = struct{}{}
	}
}
