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

package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"switch-server/models"
	"switch-server/services"

	"github.com/code-together/shared/provider"
	"github.com/code-together/shared/streaming"
)

type RelayHandler struct {
	providerService services.ProviderServiceInterface
	usageService    services.UsageServiceInterface
}

func NewRelayHandler(providerService services.ProviderServiceInterface, usageService services.UsageServiceInterface) *RelayHandler {
	return &RelayHandler{
		providerService: providerService,
		usageService:    usageService,
	}
}

// RelayMessages godoc
// @Summary      Relay Anthropic messages request
// @Description  Forward Anthropic-style /v1/messages requests to configured providers
// @Tags         Relay,Member
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        tool path string true "Provider kind (claude, codex, opencode)" Enums(claude, codex, opencode)
// @Param        request body object true "Request body (forwarded to provider)"
// @Success      200  {object}  object "Provider response"
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/relay/{tool}/v1/messages [post]
func (h *RelayHandler) RelayMessages(c *gin.Context) {
	tool := c.Param("tool")
	if !isValidProviderKind(tool) {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid tool kind: %s", tool)})
		return
	}

	h.relayForward(c, tool, "/v1/messages")
}

// RelayChatCompletions godoc
// @Summary      Relay OpenAI chat completions request
// @Description  Forward OpenAI-style /v1/chat/completions requests to configured providers
// @Tags         Relay,Member
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        tool path string true "Provider kind (claude, codex, opencode)" Enums(claude, codex, opencode)
// @Param        request body object true "Request body (forwarded to provider)"
// @Success      200  {object}  object "Provider response"
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/relay/{tool}/v1/chat/completions [post]
func (h *RelayHandler) RelayChatCompletions(c *gin.Context) {
	tool := c.Param("tool")
	if !isValidProviderKind(tool) {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid tool kind: %s", tool)})
		return
	}

	h.relayForward(c, tool, "/v1/chat/completions")
}

func (h *RelayHandler) relayForward(c *gin.Context, tool, endpoint string) {
	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}
	currentUser := user.(*models.User)

	// Read request body
	var bodyBytes []byte
	if c.Request.Body != nil {
		data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		bodyBytes = data
		c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	isStream := gjson.GetBytes(bodyBytes, "stream").Bool()
	requestedModel := gjson.GetBytes(bodyBytes, "model").String()
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())

	slog.Info("relay request received",
		"request_id", requestID,
		"tool", tool,
		"endpoint", endpoint,
		"model", requestedModel,
		"is_stream", isStream,
		"client_ip", c.ClientIP(),
		"tenant_id", currentUser.TenantID,
	)

	// Load providers for the user's team
	providers, err := h.providerService.GetProvidersByTeamID(currentUser.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load providers"})
		return
	}

	// Filter and validate providers
	active := h.filterProvidersByKind(providers, tool, requestedModel)

	if len(active) == 0 {
		if requestedModel != "" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("no available provider supports model '%s'", requestedModel),
			})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "no providers available"})
		}
		return
	}

	// Sort by level (lower = higher priority)
	sort.SliceStable(active, func(i, j int) bool {
		levelI := active[i].Level
		levelJ := active[j].Level
		if levelI <= 0 {
			levelI = 1
		}
		if levelJ <= 0 {
			levelJ = 1
		}
		return levelI < levelJ
	})

	providerNames := make([]string, len(active))
	for i, p := range active {
		providerNames[i] = p.Name
	}
	slog.Info("found available providers",
		"request_id", requestID,
		"count", len(active),
		"providers", providerNames,
	)

	// Try each provider in order
	var lastErr error
	attemptCount := 0
	for i, provider := range active {
		attemptCount++

		providerAdapter := provider.GetProviderAdapter()
		effectiveModel := provider.GetEffectiveModel(providerAdapter, requestedModel)

		currentBodyBytes := bodyBytes
		if effectiveModel != requestedModel && requestedModel != "" {
			slog.Info("provider model mapping", "provider", provider.Name, "from", requestedModel, "to", effectiveModel)

			modifiedBody, err := streaming.ReplaceModelInRequestBody(bodyBytes, effectiveModel)
			if err != nil {
				slog.Error("failed to replace model name", "provider", provider.Name, "error", err)
				lastErr = err
				continue
			}
			currentBodyBytes = modifiedBody
		}

		slog.Info("trying provider", "index", i+1, "total", len(active), "provider", provider.Name, "model", effectiveModel)

		startTime := time.Now()
		ok, err := h.forwardRequest(
			c,
			requestID,
			tool,
			&provider,
			endpoint,
			currentBodyBytes,
			isStream,
			effectiveModel,
		)
		duration := time.Since(startTime)

		if ok {
			slog.Info("provider succeeded",
				"request_id", requestID,
				"provider", provider.Name,
				"duration_sec", duration.Seconds(),
			)
			return
		}

		errorMsg := "unknown error"
		if err != nil {
			errorMsg = err.Error()
		}
		slog.Warn("provider failed",
			"request_id", requestID,
			"provider", provider.Name,
			"error", errorMsg,
			"duration_sec", duration.Seconds(),
		)
		lastErr = err
	}

	message := fmt.Sprintf("all %d providers failed (%d attempts total)", len(active), attemptCount)
	if lastErr != nil {
		message = fmt.Sprintf("%s: %s", message, lastErr.Error())
	}
	slog.Error("all providers failed", "message", message)
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}

// filterProvidersByKind filters providers by kind, enabled status, and model support
func (h *RelayHandler) filterProvidersByKind(providers []models.Provider, kind, requestedModel string) []models.Provider {
	active := make([]models.Provider, 0, len(providers))

	for _, p := range providers {
		// Skip if not enabled
		if !p.Enabled {
			continue
		}

		// Skip if kind doesn't match
		if p.Kind != kind {
			continue
		}

		// Skip if missing required fields
		if p.APIURL == "" || p.APIKey == "" {
			continue
		}

		// Validate configuration
		providerAdapter := p.GetProviderAdapter()
		if errs := provider.ValidateConfiguration(providerAdapter); len(errs) > 0 {
			slog.Warn("provider configuration validation failed, automatically skipped", "provider", p.Name, "errors", errs)
			continue
		}

		// Check model support
		if requestedModel != "" && !provider.IsModelSupported(providerAdapter, requestedModel) {
			slog.Info("provider does not support model, skipped", "provider", p.Name, "model", requestedModel)
			continue
		}

		active = append(active, p)
	}

	return active
}

// forwardRequest forwards the request to a single provider
func (h *RelayHandler) forwardRequest(
	c *gin.Context,
	requestID string,
	kind string,
	provider *models.Provider,
	endpoint string,
	bodyBytes []byte,
	isStream bool,
	model string,
) (bool, error) {
	targetURL := joinURL(provider.APIURL, endpoint)

	// Prepare headers
	headers := make(map[string]string)
	for k, v := range c.Request.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}
	headers["Authorization"] = fmt.Sprintf("Bearer %s", provider.APIKey)
	if _, ok := headers["Accept"]; !ok {
		headers["Accept"] = "application/json"
	}

	slog.Info("forwarding request",
		"request_id", requestID,
		"provider", provider.Name,
		"target_url", targetURL,
		"model", model,
		"is_stream", isStream,
	)

	usage := &streaming.TokenUsage{}
	start := time.Now()

	// Record usage on defer
	defer func() {
		duration := time.Since(start)

		usageRecord := models.UsageRecord{
			Platform:          kind,
			Model:             model,
			Provider:          provider.Name,
			HttpCode:          0, // Will be set if we get a response
			InputTokens:       usage.InputTokens,
			OutputTokens:      usage.OutputTokens,
			CacheCreateTokens: usage.CacheCreateTokens,
			CacheReadTokens:   usage.CacheReadTokens,
			ReasoningTokens:   usage.ReasoningTokens,
			TenantID:          provider.TeamID,
			UserID:            0, // Will be set from context if available
			DurationSec:       duration.Seconds(),
		}

		// Get user ID from context
		if user, exists := c.Get("user"); exists {
			if u, ok := user.(*models.User); ok {
				usageRecord.UserID = u.ID
			}
		}

		ctx := c.Request.Context()
		if err := h.usageService.CreateUsageRecord(ctx, usageRecord); err != nil {
			slog.Error("failed to record usage", "error", err, "request_id", requestID)
		}
	}()

	client := streaming.NewClient()
	resp, err := client.PostWithHeaders(targetURL, headers, flattenQuery(c.Request.URL.Query()), bodyBytes)
	if err != nil {
		return false, err
	}

	if resp == nil {
		return false, fmt.Errorf("empty response")
	}

	status := resp.StatusCode

	// For successful responses, use streaming hook to extract token usage
	if streaming.IsSuccess(status) {
		baseHook := tokenUsageHook(kind, usage)
		hook := func(data []byte) (bool, []byte) {
			return baseHook(data)
		}

		copyErr := streaming.CopyResponseWithHook(resp, c.Writer, hook)
		if copyErr == nil {
			slog.Info("upstream response",
				"request_id", requestID,
				"provider", provider.Name,
				"status", status,
				"content_type", resp.Header.Get("Content-Type"),
			)
		}
		return copyErr == nil, copyErr
	}

	// For error responses, read body and return
	upstreamBody, _ := io.ReadAll(resp.Body)
	slog.Warn("upstream response error",
		"request_id", requestID,
		"provider", provider.Name,
		"status", status,
		"body", string(upstreamBody),
	)

	// Return error to trigger failover
	return false, fmt.Errorf("provider returned status %d: %s", status, string(upstreamBody))
}

// tokenUsageHook creates a hook function for parsing token usage from streaming responses
func tokenUsageHook(kind string, usage *streaming.TokenUsage) func([]byte) (bool, []byte) {
	return func(data []byte) (bool, []byte) {
		payload := strings.TrimSpace(string(data))

		var parserFn func(string, *streaming.TokenUsage)
		switch kind {
		case "codex":
			parserFn = streaming.CodexParseTokenUsageFromResponse
		case "opencode":
			parserFn = streaming.OpenCodeParseTokenUsageFromResponse
		default:
			parserFn = streaming.ClaudeCodeParseTokenUsageFromResponse
		}

		streaming.ParseEventPayload(payload, parserFn, usage)
		return true, data
	}
}

// joinURL joins base URL and path properly
func joinURL(base, path string) string {
	base = strings.TrimSuffix(base, "/")
	path = strings.TrimPrefix(path, "/")
	return base + "/" + path
}

// flattenQuery converts url.Values to map[string]string
func flattenQuery(q map[string][]string) map[string]string {
	result := make(map[string]string)
	for k, v := range q {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}
