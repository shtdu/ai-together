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
	"testing"

	"codeswitch/internal/models"

	"github.com/code-together/shared/streaming"
	"github.com/tidwall/gjson"
)

// Test helper functions to convert between RequestLog and TokenUsage
func parseTokenUsageToRequestLog(data string, usage *models.RequestLog, parser func(string, *streaming.TokenUsage)) {
	tokenUsage := &streaming.TokenUsage{
		InputTokens:       usage.InputTokens,
		OutputTokens:      usage.OutputTokens,
		CacheCreateTokens: usage.CacheCreateTokens,
		CacheReadTokens:   usage.CacheReadTokens,
		ReasoningTokens:   usage.ReasoningTokens,
	}
	parser(data, tokenUsage)
	usage.InputTokens = tokenUsage.InputTokens
	usage.OutputTokens = tokenUsage.OutputTokens
	usage.CacheCreateTokens = tokenUsage.CacheCreateTokens
	usage.CacheReadTokens = tokenUsage.CacheReadTokens
	usage.ReasoningTokens = tokenUsage.ReasoningTokens
}

func parseEventPayloadToRequestLog(payload string, usage *models.RequestLog, parser func(string, *models.RequestLog)) {
	var combinedInput, combinedOutput int

	wrappedParser := func(data string, tu *streaming.TokenUsage) {
		// Create a temporary RequestLog for the legacy parser
		tempLog := &models.RequestLog{}
		parser(data, tempLog)

		// Accumulate in local vars
		combinedInput += tempLog.InputTokens
		combinedOutput += tempLog.OutputTokens
	}

	streaming.ParseEventPayload(payload, wrappedParser, &streaming.TokenUsage{})

	// Update usage with accumulated values
	usage.InputTokens += combinedInput
	usage.OutputTokens += combinedOutput
}

// ==================== ReplaceModelInRequestBody Test ====================

func TestReplaceModelInRequestBody(t *testing.T) {
	tests := []struct {
		name          string
		inputJSON     string
		newModel      string
		expectError   bool
		expectedModel string
	}{
		// Success scenarios
		{
			name: "Simple replacement",
			inputJSON: `{
				"model": "claude-sonnet-4",
				"messages": [{"role": "user", "content": "Hello"}]
			}`,
			newModel:      "anthropic/claude-sonnet-4",
			expectError:   false,
			expectedModel: "anthropic/claude-sonnet-4",
		},
		{
			name: "Complex nested JSON",
			inputJSON: `{
				"model": "claude-opus-4",
				"messages": [
					{
						"role": "user",
						"content": "Test"
					}
				],
				"temperature": 0.7,
				"max_tokens": 1000,
				"metadata": {
					"user_id": "12345"
				}
			}`,
			newModel:      "gpt-4",
			expectError:   false,
			expectedModel: "gpt-4",
		},
		{
			name: "Model name with special characters",
			inputJSON: `{
				"model": "claude-sonnet-4",
				"messages": []
			}`,
			newModel:      "anthropic/claude-3.5-sonnet@20241022",
			expectError:   false,
			expectedModel: "anthropic/claude-3.5-sonnet@20241022",
		},

		// Error scenarios
		{
			name: "Missing model field - no error (sjson sets it)",
			inputJSON: `{
				"messages": [{"role": "user", "content": "Hello"}]
			}`,
			newModel:      "any-model",
			expectError:   false, // sjson will set the field
			expectedModel: "any-model",
		},
		{
			name:          "Empty JSON - no error (sjson sets it)",
			inputJSON:     `{}`,
			newModel:      "any-model",
			expectError:   false, // sjson will set the field
			expectedModel: "any-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes := []byte(tt.inputJSON)
			result, err := streaming.ReplaceModelInRequestBody(bodyBytes, tt.newModel)

			// Check error expectations
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// If no error expected, verify result
			if !tt.expectError {
				// Verify returned JSON is valid
				if !json.Valid(result) {
					t.Errorf("Returned JSON is invalid")
				}

				// Verify model name is correctly replaced
				actualModel := gjson.GetBytes(result, "model").String()
				if actualModel != tt.expectedModel {
					t.Errorf("Replaced model name = %q, expected %q", actualModel, tt.expectedModel)
				}

				// Verify other fields are not modified
				if gjson.GetBytes(bodyBytes, "messages").Exists() {
					originalMessages := gjson.GetBytes(bodyBytes, "messages").Raw
					resultMessages := gjson.GetBytes(result, "messages").Raw
					if originalMessages != resultMessages {
						t.Errorf("messages field was unexpectedly modified")
					}
				}
			}
		})
	}
}

// ==================== End-to-End Scenario Tests ====================

func TestModelMappingEndToEnd(t *testing.T) {
	// Simulate real scenario: User requests claude-sonnet-4, needs to map to OpenRouter format
	provider := Provider{
		Name: "OpenRouter",
		SupportedModels: []string{
			"anthropic/claude-sonnet-4",
			"anthropic/claude-opus-4",
			"openai/gpt-4",
			"google/gemini-pro",
			"meta-llama/llama-3.1-405b",
			"anthropic/claude-3.5-sonnet",
			"anthropic/claude-3.5-haiku",
		},
		ModelMapping: map[string]string{
			"claude-*": "anthropic/claude-*",
			"gpt-*":    "openai/gpt-*",
			"gemini-*": "google/gemini-*",
			"llama-*":  "meta-llama/llama-*",
		},
	}

	scenarios := []struct {
		requestedModel string
		shouldSupport  bool
		effectiveModel string
	}{
		// Wildcard mapping scenarios
		{"claude-sonnet-4", true, "anthropic/claude-sonnet-4"},
		{"claude-opus-4", true, "anthropic/claude-opus-4"},
		{"claude-3.5-sonnet", true, "anthropic/claude-3.5-sonnet"},
		{"gpt-4", true, "openai/gpt-4"},
		{"gpt-4-turbo", true, "openai/gpt-4-turbo"},
		{"gemini-pro", true, "google/gemini-pro"},
		{"llama-3.1-405b", true, "meta-llama/llama-3.1-405b"},

		// Unsupported models
		{"deepseek-v3", false, "deepseek-v3"},
		{"qwen-max", false, "qwen-max"},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.requestedModel, func(t *testing.T) {
			// 1. Check if supported
			supported := provider.IsModelSupported(scenario.requestedModel)
			if supported != scenario.shouldSupport {
				t.Errorf("IsModelSupported(%q) = %v, expected %v",
					scenario.requestedModel, supported, scenario.shouldSupport)
			}

			// 2. Get effective model name
			effectiveModel := provider.GetEffectiveModel(scenario.requestedModel)
			if effectiveModel != scenario.effectiveModel {
				t.Errorf("GetEffectiveModel(%q) = %q, expected %q",
					scenario.requestedModel, effectiveModel, scenario.effectiveModel)
			}

			// 3. If supported, test request body replacement
			if scenario.shouldSupport {
				requestBody := `{"model": "` + scenario.requestedModel + `", "messages": []}`
				result, err := streaming.ReplaceModelInRequestBody([]byte(requestBody), effectiveModel)
				if err != nil {
					t.Fatalf("streaming.ReplaceModelInRequestBody failed: %v", err)
				}

				actualModel := gjson.GetBytes(result, "model").String()
				if actualModel != scenario.effectiveModel {
					t.Errorf("Model in request body = %q, expected %q", actualModel, scenario.effectiveModel)
				}
			}
		})
	}
}

// ==================== Configuration Validation Integration Tests ====================

func TestProviderConfigValidation(t *testing.T) {
	// Scenario 1: Perfect configuration
	validProvider := Provider{
		Name: "ValidProvider",
		SupportedModels: []string{
			"anthropic/claude-sonnet-4",
			"anthropic/claude-opus-4",
		},
		ModelMapping: map[string]string{
			"claude-sonnet-4": "anthropic/claude-sonnet-4",
			"claude-opus-4":   "anthropic/claude-opus-4",
		},
	}

	errors := validProvider.ValidateConfiguration()
	if len(errors) != 0 {
		t.Errorf("Perfect configuration should have no errors, but got: %v", errors)
	}

	// Scenario 2: Invalid configuration - Mapping target does not exist
	invalidProvider := Provider{
		Name: "InvalidProvider",
		SupportedModels: []string{
			"model-a",
		},
		ModelMapping: map[string]string{
			"external": "non-existent-model",
		},
	}

	errors = invalidProvider.ValidateConfiguration()
	if len(errors) == 0 {
		t.Errorf("Invalid configuration should return validation errors")
	}

	// Scenario 3: Wildcard configuration
	wildcardProvider := Provider{
		Name: "WildcardProvider",
		SupportedModels: []string{
			"anthropic/claude-*",
			"openai/gpt-*",
		},
		ModelMapping: map[string]string{
			"claude-*": "anthropic/claude-*",
			"gpt-*":    "openai/gpt-*",
		},
	}

	errors = wildcardProvider.ValidateConfiguration()
	if len(errors) != 0 {
		t.Errorf("Wildcard configuration should have no errors, but got: %v", errors)
	}
}

// ==================== Performance Tests ====================

func BenchmarkIsModelSupported(b *testing.B) {
	provider := Provider{
		SupportedModels: []string{
			"claude-sonnet-4",
			"claude-opus-4",
			"gpt-4",
			"gpt-4-turbo",
		},
		ModelMapping: map[string]string{
			"claude-*": "anthropic/claude-*",
			"gpt-*":    "openai/gpt-*",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = provider.IsModelSupported("claude-sonnet-4")
	}
}

func BenchmarkGetEffectiveModel(b *testing.B) {
	provider := Provider{
		ModelMapping: map[string]string{
			"claude-*": "anthropic/claude-*",
			"gpt-*":    "openai/gpt-*",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = provider.GetEffectiveModel("claude-sonnet-4")
	}
}

func BenchmarkReplaceModelInRequestBody(b *testing.B) {
	bodyBytes := []byte(`{
		"model": "claude-sonnet-4",
		"messages": [{"role": "user", "content": "Hello"}],
		"temperature": 0.7,
		"max_tokens": 1000
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = streaming.ReplaceModelInRequestBody(bodyBytes, "anthropic/claude-sonnet-4")
	}
}

// ==================== Token Parsing Tests ====================

func TestClaudeCodeParseTokenUsageFromResponse(t *testing.T) {
	tests := []struct {
		name           string
		data           string
		expectedInput  int
		expectedOutput int
		expectedCache  int
		expectedRead   int
	}{
		{
			name:           "message.usage format",
			data:           `{"message":{"usage":{"input_tokens":100,"output_tokens":200,"cache_creation_input_tokens":10,"cache_read_input_tokens":5}}}`,
			expectedInput:  105,
			expectedOutput: 200,
			expectedCache:  10,
			expectedRead:   5,
		},
		{
			name:           "usage cache_read_input_tokens format",
			data:           `{"usage":{"input_tokens":50,"output_tokens":75,"cache_read_input_tokens":12}}`,
			expectedInput:  62,
			expectedOutput: 75,
			expectedCache:  0,
			expectedRead:   12,
		},
		{
			name:           "usage format",
			data:           `{"usage":{"input_tokens":50,"output_tokens":75}}`,
			expectedInput:  50,
			expectedOutput: 75,
			expectedCache:  0,
			expectedRead:   0,
		},
		{
			name:           "both formats (should accumulate)",
			data:           `{"message":{"usage":{"input_tokens":100,"output_tokens":200}},"usage":{"input_tokens":50,"output_tokens":75}}`,
			expectedInput:  150, // 100 + 50
			expectedOutput: 275, // 200 + 75
			expectedCache:  0,
			expectedRead:   0,
		},
		{
			name:           "missing fields",
			data:           `{"message":{}}`,
			expectedInput:  0,
			expectedOutput: 0,
			expectedCache:  0,
			expectedRead:   0,
		},
		{
			name:           "empty JSON",
			data:           `{}`,
			expectedInput:  0,
			expectedOutput: 0,
			expectedCache:  0,
			expectedRead:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := &models.RequestLog{}
			parseTokenUsageToRequestLog(tt.data, usage, streaming.ClaudeCodeParseTokenUsageFromResponse)

			if usage.InputTokens != tt.expectedInput {
				t.Errorf("InputTokens = %d, want %d", usage.InputTokens, tt.expectedInput)
			}
			if usage.OutputTokens != tt.expectedOutput {
				t.Errorf("OutputTokens = %d, want %d", usage.OutputTokens, tt.expectedOutput)
			}
			if usage.CacheCreateTokens != tt.expectedCache {
				t.Errorf("CacheCreateTokens = %d, want %d", usage.CacheCreateTokens, tt.expectedCache)
			}
			if usage.CacheReadTokens != tt.expectedRead {
				t.Errorf("CacheReadTokens = %d, want %d", usage.CacheReadTokens, tt.expectedRead)
			}
		})
	}
}

func TestCodexParseTokenUsageFromResponse(t *testing.T) {
	tests := []struct {
		name              string
		data              string
		expectedInput     int
		expectedOutput    int
		expectedCacheRead int
		expectedReasoning int
	}{
		{
			name:              "full usage format",
			data:              `{"response":{"usage":{"input_tokens":150,"output_tokens":300,"input_tokens_details":{"cached_tokens":20},"output_tokens_details":{"reasoning_tokens":50}}}}`,
			expectedInput:     150,
			expectedOutput:    300,
			expectedCacheRead: 20,
			expectedReasoning: 50,
		},
		{
			name:              "minimal usage format",
			data:              `{"response":{"usage":{"input_tokens":100,"output_tokens":200}}}`,
			expectedInput:     100,
			expectedOutput:    200,
			expectedCacheRead: 0,
			expectedReasoning: 0,
		},
		{
			name:              "missing details",
			data:              `{"response":{"usage":{"input_tokens":50,"output_tokens":75}}}`,
			expectedInput:     50,
			expectedOutput:    75,
			expectedCacheRead: 0,
			expectedReasoning: 0,
		},
		{
			name:              "empty JSON",
			data:              `{}`,
			expectedInput:     0,
			expectedOutput:    0,
			expectedCacheRead: 0,
			expectedReasoning: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := &models.RequestLog{}
			parseTokenUsageToRequestLog(tt.data, usage, streaming.CodexParseTokenUsageFromResponse)

			if usage.InputTokens != tt.expectedInput {
				t.Errorf("InputTokens = %d, want %d", usage.InputTokens, tt.expectedInput)
			}
			if usage.OutputTokens != tt.expectedOutput {
				t.Errorf("OutputTokens = %d, want %d", usage.OutputTokens, tt.expectedOutput)
			}
			if usage.CacheReadTokens != tt.expectedCacheRead {
				t.Errorf("CacheReadTokens = %d, want %d", usage.CacheReadTokens, tt.expectedCacheRead)
			}
			if usage.ReasoningTokens != tt.expectedReasoning {
				t.Errorf("ReasoningTokens = %d, want %d", usage.ReasoningTokens, tt.expectedReasoning)
			}
		})
	}
}

func TestParseEventPayload(t *testing.T) {
	tests := []struct {
		name           string
		payload        string
		kind           string
		expectedInput  int
		expectedOutput int
	}{
		{
			name: "claude SSE format",
			payload: "data: {\"message\":{\"usage\":{\"input_tokens\":100,\"output_tokens\":200}}}\n\n" +
				"data: {\"message\":{\"usage\":{\"input_tokens\":50,\"output_tokens\":75}}}\n\n",
			kind:           "claude",
			expectedInput:  150, // 100 + 50
			expectedOutput: 275, // 200 + 75
		},
		{
			name: "codex SSE format",
			payload: "data: {\"response\":{\"usage\":{\"input_tokens\":200,\"output_tokens\":400}}}\n\n" +
				"data: {\"response\":{\"usage\":{\"input_tokens\":100,\"output_tokens\":200}}}\n\n",
			kind:           "codex",
			expectedInput:  300, // 200 + 100
			expectedOutput: 600, // 400 + 200
		},
		{
			name: "mixed lines (only data: lines parsed)",
			payload: "event: message\n" +
				"data: {\"message\":{\"usage\":{\"input_tokens\":100,\"output_tokens\":200}}}\n" +
				"id: 123\n" +
				"data: {\"message\":{\"usage\":{\"input_tokens\":50,\"output_tokens\":75}}}\n\n",
			kind:           "claude",
			expectedInput:  150,
			expectedOutput: 275,
		},
		{
			name:           "no data: lines",
			payload:        "event: message\nid: 123\n\n",
			kind:           "claude",
			expectedInput:  0,
			expectedOutput: 0,
		},
		{
			name:           "empty payload",
			payload:        "",
			kind:           "claude",
			expectedInput:  0,
			expectedOutput: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := &models.RequestLog{}

			var parser func(string, *models.RequestLog)
			if tt.kind == "claude" {
				parser = func(data string, u *models.RequestLog) {
					parseTokenUsageToRequestLog(data, u, streaming.ClaudeCodeParseTokenUsageFromResponse)
				}
			} else {
				parser = func(data string, u *models.RequestLog) {
					parseTokenUsageToRequestLog(data, u, streaming.CodexParseTokenUsageFromResponse)
				}
			}

			parseEventPayloadToRequestLog(tt.payload, usage, parser)

			if usage.InputTokens != tt.expectedInput {
				t.Errorf("InputTokens = %d, want %d", usage.InputTokens, tt.expectedInput)
			}
			if usage.OutputTokens != tt.expectedOutput {
				t.Errorf("OutputTokens = %d, want %d", usage.OutputTokens, tt.expectedOutput)
			}
		})
	}
}

func TestReqeustLogHook(t *testing.T) {
	// Test that the hook function is created and works correctly
	usage := &models.RequestLog{}

	// Simulate SSE data chunks
	sseChunk1 := []byte("data: {\"message\":{\"usage\":{\"input_tokens\":100,\"output_tokens\":200}}}\n\n")
	sseChunk2 := []byte("data: {\"message\":{\"usage\":{\"input_tokens\":50,\"output_tokens\":75}}}\n\n")

	// Create a mock gin context (we'll just test the hook function directly)
	hook := func(data []byte) (bool, []byte) {
		payload := string(data)
		parseTokenUsageToRequestLog(payload, usage, streaming.ClaudeCodeParseTokenUsageFromResponse)
		return true, data
	}

	// Process chunks
	shouldContinue, modified := hook(sseChunk1)
	if !shouldContinue {
		t.Error("hook should continue after first chunk")
	}
	if string(modified) != string(sseChunk1) {
		t.Error("hook should return original data unchanged")
	}

	shouldContinue, modified = hook(sseChunk2)
	if !shouldContinue {
		t.Error("hook should continue after second chunk")
	}

	// Verify accumulated tokens
	if usage.InputTokens != 150 {
		t.Errorf("InputTokens = %d, want 150", usage.InputTokens)
	}
	if usage.OutputTokens != 275 {
		t.Errorf("OutputTokens = %d, want 275", usage.OutputTokens)
	}
}

func TestTokenParsingAccumulation(t *testing.T) {
	// Test that tokens accumulate correctly across multiple calls
	usage := &models.RequestLog{}

	// First call
	parseTokenUsageToRequestLog(
		`{"message":{"usage":{"input_tokens":100,"output_tokens":200}}}`,
		usage,
		streaming.ClaudeCodeParseTokenUsageFromResponse,
	)

	// Second call (should accumulate)
	parseTokenUsageToRequestLog(
		`{"message":{"usage":{"input_tokens":50,"output_tokens":75}}}`,
		usage,
		streaming.ClaudeCodeParseTokenUsageFromResponse,
	)

	// Third call (should accumulate)
	parseTokenUsageToRequestLog(
		`{"usage":{"input_tokens":25,"output_tokens":50}}`,
		usage,
		streaming.ClaudeCodeParseTokenUsageFromResponse,
	)

	if usage.InputTokens != 175 {
		t.Errorf("InputTokens = %d, want 175 (100+50+25)", usage.InputTokens)
	}
	if usage.OutputTokens != 325 {
		t.Errorf("OutputTokens = %d, want 325 (200+75+50)", usage.OutputTokens)
	}
}

func BenchmarkClaudeCodeParseTokenUsageFromResponse(b *testing.B) {
	data := `{"message":{"usage":{"input_tokens":100,"output_tokens":200,"cache_creation_input_tokens":10,"cache_read_input_tokens":5}},"usage":{"input_tokens":50,"output_tokens":75}}`
	usage := &models.RequestLog{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		usage.InputTokens = 0
		usage.OutputTokens = 0
		usage.CacheCreateTokens = 0
		usage.CacheReadTokens = 0
		parseTokenUsageToRequestLog(data, usage, streaming.ClaudeCodeParseTokenUsageFromResponse)
	}
}

func BenchmarkCodexParseTokenUsageFromResponse(b *testing.B) {
	data := `{"response":{"usage":{"input_tokens":150,"output_tokens":300,"input_tokens_details":{"cached_tokens":20},"output_tokens_details":{"reasoning_tokens":50}}}}`
	usage := &models.RequestLog{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		usage.InputTokens = 0
		usage.OutputTokens = 0
		usage.CacheReadTokens = 0
		usage.ReasoningTokens = 0
		parseTokenUsageToRequestLog(data, usage, streaming.CodexParseTokenUsageFromResponse)
	}
}

func BenchmarkParseEventPayload(b *testing.B) {
	payload := "data: {\"message\":{\"usage\":{\"input_tokens\":100,\"output_tokens\":200}}}\n\n" +
		"data: {\"message\":{\"usage\":{\"input_tokens\":50,\"output_tokens\":75}}}\n\n"
	usage := &models.RequestLog{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		usage.InputTokens = 0
		usage.OutputTokens = 0
		parseEventPayloadToRequestLog(payload, usage, func(data string, u *models.RequestLog) {
			parseTokenUsageToRequestLog(data, u, streaming.ClaudeCodeParseTokenUsageFromResponse)
		})
	}
}
