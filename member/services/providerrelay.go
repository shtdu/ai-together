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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"codeswitch/internal/db"
	"codeswitch/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"

	"github.com/code-together/shared/streaming"
)

type ProviderRelayService struct {
	providerService *ProviderService
	hookService     *HookService
	server          *http.Server
	addr            string
	verbose         bool // Controls Gin debug mode and logging verbosity
}

func NewProviderRelayService(providerService *ProviderService, hookService *HookService, addr string, verbose bool) *ProviderRelayService {
	if addr == "" {
		addr = ":18100"
	}

	if err := db.Init(); err != nil {
		slog.Error("failed to initialize database", "error", err)
	}

	return &ProviderRelayService{
		providerService: providerService,
		hookService:     hookService,
		addr:            addr,
		verbose:         verbose,
	}
}

func (prs *ProviderRelayService) Start() error {
	// Set Gin mode based on verbose flag
	if prs.verbose {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	if warnings := prs.validateConfig(); len(warnings) > 0 {
		slog.Info("======== Provider Configuration Validation Warnings ========")
		for _, warn := range warnings {
			slog.Warn(warn, "warning", true)
		}
		slog.Info("===========================================================")
	}

	// Create router without default middleware
	// Only add Logger middleware in verbose mode
	router := gin.New()
	router.Use(gin.Recovery())
	if prs.verbose {
		router.Use(gin.Logger())
	}
	prs.registerRoutes(router)

	prs.server = &http.Server{
		Addr:    prs.addr,
		Handler: router,
	}

	slog.Info("provider relay server listening", "address", prs.addr)

	go func() {
		if err := prs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("provider relay server error", "error", err)
		}
	}()
	return nil
}

func (prs *ProviderRelayService) validateConfig() []string {
	warnings := make([]string, 0)

	for _, kind := range []string{"claude", "codex", "opencode"} {
		providers, err := prs.providerService.LoadProviders(kind)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("[%s] failed to load configuration: %v", kind, err))
			continue
		}

		enabledCount := 0
		for _, p := range providers {
			if !p.Enabled {
				continue
			}
			enabledCount++

			if errs := p.ValidateConfiguration(); len(errs) > 0 {
				for _, errMsg := range errs {
					warnings = append(warnings, fmt.Sprintf("[%s/%s] %s", kind, p.Name, errMsg))
				}
			}

			if len(p.SupportedModels) == 0 &&
				len(p.ModelMapping) == 0 {
				warnings = append(warnings, fmt.Sprintf(
					"[%s/%s] no supportedModels or modelMapping configured, assuming all models are supported (may cause fallback failure)",
					kind, p.Name))
			}
		}

		if enabledCount == 0 {
			warnings = append(warnings, fmt.Sprintf("[%s] no enabled providers", kind))
		}
	}

	return warnings
}

func (prs *ProviderRelayService) Stop() error {
	if prs.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return prs.server.Shutdown(ctx)
}

func (prs *ProviderRelayService) Addr() string {
	return prs.addr
}

func (prs *ProviderRelayService) registerRoutes(router gin.IRouter) {
	router.POST("/v1/messages", prs.proxyHandler("claude", "/v1/messages"))
	router.POST("/responses", prs.proxyHandler("codex", "/responses"))
	router.POST("/chat/completions", prs.proxyHandler("opencode", "/chat/completions"))

	// Hook collection endpoint (delegated to HookService)
	if prs.hookService != nil {
		router.POST("/collect/:tool_name", prs.hookService.CollectEvent)
	}
}

func (prs *ProviderRelayService) proxyHandler(kind string, endpoint string) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		if requestedModel == "" {
			slog.Warn("request does not specify model name, cannot perform model fallback")
		}

		slog.Debug("relay request received",
			"request_id", requestID,
			"kind", kind,
			"endpoint", endpoint,
			"model", requestedModel,
			"is_stream", isStream,
			"client_ip", c.ClientIP(),
		)

		providers, err := prs.providerService.LoadProviders(kind)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load providers"})
			return
		}

		active := make([]Provider, 0, len(providers))
		skippedCount := 0
		for _, provider := range providers {
			if !provider.Enabled || provider.APIURL == "" || provider.APIKey == "" {
				continue
			}

			if errs := provider.ValidateConfiguration(); len(errs) > 0 {
				slog.Warn("provider configuration validation failed, automatically skipped", "provider", provider.Name, "errors", errs)
				skippedCount++
				continue
			}

			if requestedModel != "" && !provider.IsModelSupported(requestedModel) {
				slog.Debug("provider does not support model, skipped", "provider", provider.Name, "model", requestedModel)
				skippedCount++
				continue
			}

			active = append(active, provider)
		}

		// Sort providers by level (lower number = higher priority)
		// Use stable sort to preserve original order within same level
		sort.SliceStable(active, func(i, j int) bool {
			levelI := active[i].Level
			levelJ := active[j].Level
			if levelI <= 0 {
				levelI = 1 // Default to level 1
			}
			if levelJ <= 0 {
				levelJ = 1 // Default to level 1
			}
			return levelI < levelJ // Lower level = higher priority
		})

		if len(active) == 0 {
			if requestedModel != "" {
				c.JSON(http.StatusNotFound, gin.H{
					"error": fmt.Sprintf("no available provider supports model '%s' (skipped %d incompatible providers)", requestedModel, skippedCount),
				})
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": "no providers available"})
			}
			return
		}

		providerNames := make([]string, len(active))
		for i, p := range active {
			providerNames[i] = p.Name
		}
		slog.Debug("found available providers",
			"request_id", requestID,
			"count", len(active),
			"filtered", skippedCount,
			"providers", providerNames,
		)

		query := flattenQuery(c.Request.URL.Query())
		clientHeaders := cloneHeaders(c.Request.Header)

		var lastErr error
		attemptCount := 0
		for i, provider := range active {
			attemptCount++

			effectiveModel := provider.GetEffectiveModel(requestedModel)

			currentBodyBytes := bodyBytes
			if effectiveModel != requestedModel && requestedModel != "" {
				slog.Debug("provider model mapping", "provider", provider.Name, "from", requestedModel, "to", effectiveModel)

				modifiedBody, err := streaming.ReplaceModelInRequestBody(bodyBytes, effectiveModel)
				if err != nil {
					slog.Error("failed to replace model name", "provider", provider.Name, "error", err)
					lastErr = err
					continue
				}
				currentBodyBytes = modifiedBody
			}

			slog.Debug("trying provider", "index", i+1, "total", len(active), "provider", provider.Name, "model", effectiveModel)

			startTime := time.Now()
			ok, err := prs.forwardRequest(
				c,
				requestID,
				kind,
				provider,
				endpoint,
				query,
				clientHeaders,
				currentBodyBytes,
				isStream,
				effectiveModel,
			)
			duration := time.Since(startTime)

			if ok {
				slog.Debug("provider succeeded",
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
}

func (prs *ProviderRelayService) forwardRequest(
	c *gin.Context,
	requestID string,
	kind string,
	provider Provider,
	endpoint string,
	query map[string]string,
	clientHeaders map[string]string,
	bodyBytes []byte,
	isStream bool,
	model string,
) (bool, error) {
	targetURL := joinURL(provider.APIURL, endpoint)
	headers := cloneMap(clientHeaders)
	headers["Authorization"] = fmt.Sprintf("Bearer %s", provider.APIKey)
	if _, ok := headers["Accept"]; !ok {
		headers["Accept"] = "application/json"
	}

	slog.Debug("forwarding request",
		"request_id", requestID,
		"provider", provider.Name,
		"target_url", targetURL,
		"model", model,
		"is_stream", isStream,
	)

	requestLog := &models.RequestLog{
		Platform: kind,
		Provider: provider.Name,
		Model:    model,
		IsStream: isStream,
	}
	start := time.Now()
	defer func() {
		requestLog.DurationSec = time.Since(start).Seconds()
		if _, err := db.DB.NamedExec(`INSERT INTO request_log
			(platform, model, provider, http_code, input_tokens, output_tokens,
			 cache_create_tokens, cache_read_tokens, reasoning_tokens, is_stream, duration_sec)
			VALUES (:platform, :model, :provider, :http_code, :input_tokens, :output_tokens,
					:cache_create_tokens, :cache_read_tokens, :reasoning_tokens, :is_stream, :duration_sec)`,
			requestLog); err != nil {
			slog.Error("failed to write to request_log", "error", err)
		}
	}()

	client := streaming.NewClient()
	resp, err := client.PostWithHeaders(targetURL, headers, query, bodyBytes)
	if err != nil {
		return false, err
	}

	if resp == nil {
		return false, fmt.Errorf("empty response")
	}

	status := resp.StatusCode
	requestLog.HttpCode = status

	if streaming.IsSuccess(status) {
		// Use hook for streaming responses to extract token usage
		var captured []byte
		const captureLimit = 16 * 1024
		var streamTail []string
		const streamTailLimit = 20
		streamBuf := ""
		appendStreamLine := func(line string) {
			line = strings.TrimSpace(line)
			if line == "" {
				return
			}
			if strings.HasPrefix(line, "data:") {
				line = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			}
			if line == "" {
				return
			}
			if len(streamTail) < streamTailLimit {
				streamTail = append(streamTail, line)
				return
			}
			copy(streamTail, streamTail[1:])
			streamTail[len(streamTail)-1] = line
		}
		appendStreamChunk := func(data []byte) {
			streamBuf += string(data)
			lines := strings.Split(streamBuf, "\n")
			if len(lines) == 0 {
				return
			}
			streamBuf = lines[len(lines)-1]
			for _, line := range lines[:len(lines)-1] {
				appendStreamLine(line)
			}
		}

		baseHook := ReqeustLogHook(c, kind, requestLog)
		hook := func(data []byte) (bool, []byte) {
			if !isStream && len(captured) < captureLimit {
				remaining := captureLimit - len(captured)
				if len(data) > remaining {
					captured = append(captured, data[:remaining]...)
				} else {
					captured = append(captured, data...)
				}
			}
			if isStream {
				appendStreamChunk(data)
			}
			return baseHook(data)
		}
		copyErr := streaming.CopyResponseWithHook(resp, c.Writer, hook)
		if copyErr == nil {
			slog.Debug("upstream response",
				"request_id", requestID,
				"provider", provider.Name,
				"status", status,
				"content_type", resp.Header.Get("Content-Type"),
			)
			if !isStream && len(captured) > 0 {
				payload := formatPrettyPayload(string(captured))
				slog.Debug(fmt.Sprintf("upstream response body request_id=%s provider=%s\n%s", requestID, provider.Name, payload))
			}
			if isStream && len(streamTail) > 0 {
				formattedLines := make([]string, 0, len(streamTail))
				for _, line := range streamTail {
					formattedLines = append(formattedLines, formatPrettyPayload(line))
				}
				slog.Debug(fmt.Sprintf("upstream stream tail request_id=%s provider=%s\n%s", requestID, provider.Name, strings.Join(formattedLines, "\n")))
			}
		}
		return copyErr == nil, copyErr
	}

	upstreamBody, readErr := readResponseSnippet(resp, 16*1024)
	if readErr != nil {
		slog.Warn("failed to read upstream response body",
			"request_id", requestID,
			"provider", provider.Name,
			"status", status,
			"error", readErr,
		)
	}
	slog.Warn("upstream response error",
		"request_id", requestID,
		"provider", provider.Name,
		"status", status,
		"content_type", resp.Header.Get("Content-Type"),
		"body", upstreamBody,
	)
	return false, fmt.Errorf("upstream status %d", status)
}

func cloneHeaders(header http.Header) map[string]string {
	cloned := make(map[string]string, len(header))
	for key, values := range header {
		if len(values) > 0 {
			cloned[key] = values[len(values)-1]
		}
	}
	return cloned
}

func cloneMap(m map[string]string) map[string]string {
	cloned := make(map[string]string, len(m))
	for k, v := range m {
		cloned[k] = v
	}
	return cloned
}

func flattenQuery(values map[string][]string) map[string]string {
	query := make(map[string]string, len(values))
	for key, items := range values {
		if len(items) > 0 {
			query[key] = items[len(items)-1]
		}
	}
	return query
}

func joinURL(base string, endpoint string) string {
	base = strings.TrimSuffix(base, "/")
	endpoint = "/" + strings.TrimPrefix(endpoint, "/")
	return base + endpoint
}

func readResponseSnippet(resp *http.Response, limit int64) (string, error) {
	if resp == nil || resp.Body == nil {
		return "", nil
	}
	defer resp.Body.Close()

	reader := io.LimitReader(resp.Body, limit)
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", nil
	}
	return sanitizeLogPayload(string(data)), nil
}

func formatPrettyPayload(payload string) string {
	payload = strings.TrimSpace(strings.ReplaceAll(payload, "\r", ""))
	if payload == "" {
		return payload
	}

	var pretty bytes.Buffer
	if strings.HasPrefix(payload, "{") || strings.HasPrefix(payload, "[") {
		if err := json.Indent(&pretty, []byte(payload), "", "  "); err == nil {
			return pretty.String()
		}
	}

	return payload
}

func sanitizeLogPayload(payload string) string {
	payload = strings.TrimSpace(strings.ReplaceAll(payload, "\r", ""))
	if payload == "" {
		return payload
	}

	if strings.HasPrefix(payload, "{") || strings.HasPrefix(payload, "[") {
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, []byte(payload), "", "  "); err == nil {
			return pretty.String()
		}
	}

	payload = strings.ReplaceAll(payload, "\n", "\\n")
	return payload
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ReqeustLogHook creates a hook function for parsing token usage from streaming responses
func ReqeustLogHook(c *gin.Context, kind string, usage *models.RequestLog) func(data []byte) (bool, []byte) {
	// Create a shared TokenUsage wrapper for the RequestLog
	tokenUsage := &streaming.TokenUsage{}

	return func(data []byte) (bool, []byte) {
		payload := strings.TrimSpace(string(data))

		// Parse token usage into shared TokenUsage struct
		streaming.ParseEventPayload(payload, func(data string, tu *streaming.TokenUsage) {
			switch kind {
			case "codex":
				streaming.CodexParseTokenUsageFromResponse(data, tu)
			case "opencode":
				streaming.OpenCodeParseTokenUsageFromResponse(data, tu)
			default:
				streaming.ClaudeCodeParseTokenUsageFromResponse(data, tu)
			}
		}, tokenUsage)

		// Copy shared TokenUsage to RequestLog
		usage.InputTokens = tokenUsage.InputTokens
		usage.OutputTokens = tokenUsage.OutputTokens
		usage.CacheCreateTokens = tokenUsage.CacheCreateTokens
		usage.CacheReadTokens = tokenUsage.CacheReadTokens
		usage.ReasoningTokens = tokenUsage.ReasoningTokens

		return true, data
	}
}
