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


package streaming

import (
	"testing"
)

// ==================== ClaudeCodeParseTokenUsageFromResponse Tests ====================

func TestClaudeCodeParseTokenUsageFromResponse(t *testing.T) {
	tests := []struct {
		name                string
		data                string
		expectedInput       int
		expectedOutput      int
		expectedCacheCreate int
		expectedCacheRead   int
	}{
		{
			name: "message.usage format",
			data: `{"message":{"usage":{"input_tokens":100,"output_tokens":200,"cache_creation_input_tokens":10,"cache_read_input_tokens":5}}}`,
			expectedInput:       105, // 100 + 5
			expectedOutput:      200,
			expectedCacheCreate: 10,
			expectedCacheRead:   5,
		},
		{
			name: "top-level usage format",
			data: `{"usage":{"input_tokens":50,"output_tokens":75,"cache_read_input_tokens":12}}`,
			expectedInput:       62, // 50 + 12
			expectedOutput:      75,
			expectedCacheCreate: 0,
			expectedCacheRead:   12,
		},
		{
			name: "both formats (accumulates)",
			data: `{"message":{"usage":{"input_tokens":100,"output_tokens":200}},"usage":{"input_tokens":50,"output_tokens":75}}`,
			expectedInput:       150, // 100 + 50
			expectedOutput:      275, // 200 + 75
			expectedCacheCreate: 0,
			expectedCacheRead:   0,
		},
		{
			name: "missing fields",
			data: `{"message":{}}`,
			expectedInput:       0,
			expectedOutput:      0,
			expectedCacheCreate: 0,
			expectedCacheRead:   0,
		},
		{
			name: "empty JSON",
			data: `{}`,
			expectedInput:       0,
			expectedOutput:      0,
			expectedCacheCreate: 0,
			expectedCacheRead:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := &TokenUsage{}
			ClaudeCodeParseTokenUsageFromResponse(tt.data, usage)

			if usage.InputTokens != tt.expectedInput {
				t.Errorf("InputTokens = %d, want %d", usage.InputTokens, tt.expectedInput)
			}
			if usage.OutputTokens != tt.expectedOutput {
				t.Errorf("OutputTokens = %d, want %d", usage.OutputTokens, tt.expectedOutput)
			}
			if usage.CacheCreateTokens != tt.expectedCacheCreate {
				t.Errorf("CacheCreateTokens = %d, want %d", usage.CacheCreateTokens, tt.expectedCacheCreate)
			}
			if usage.CacheReadTokens != tt.expectedCacheRead {
				t.Errorf("CacheReadTokens = %d, want %d", usage.CacheReadTokens, tt.expectedCacheRead)
			}
		})
	}
}

// ==================== CodexParseTokenUsageFromResponse Tests ====================

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
			name: "full usage format",
			data: `{"response":{"usage":{"input_tokens":150,"output_tokens":300,"input_tokens_details":{"cached_tokens":20},"output_tokens_details":{"reasoning_tokens":50}}}}`,
			expectedInput:     150,
			expectedOutput:    300,
			expectedCacheRead: 20,
			expectedReasoning: 50,
		},
		{
			name: "minimal usage format",
			data: `{"response":{"usage":{"input_tokens":100,"output_tokens":200}}}`,
			expectedInput:     100,
			expectedOutput:    200,
			expectedCacheRead: 0,
			expectedReasoning: 0,
		},
		{
			name: "missing details",
			data: `{"response":{"usage":{"input_tokens":50,"output_tokens":75}}}`,
			expectedInput:     50,
			expectedOutput:    75,
			expectedCacheRead: 0,
			expectedReasoning: 0,
		},
		{
			name: "empty JSON",
			data: `{}`,
			expectedInput:     0,
			expectedOutput:    0,
			expectedCacheRead: 0,
			expectedReasoning: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := &TokenUsage{}
			CodexParseTokenUsageFromResponse(tt.data, usage)

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

// ==================== OpenCodeParseTokenUsageFromResponse Tests ====================

func TestOpenCodeParseTokenUsageFromResponse(t *testing.T) {
	tests := []struct {
		name              string
		data              string
		expectedInput     int
		expectedOutput    int
		expectedCacheRead int
	}{
		{
			name: "standard format with prompt_tokens",
			data: `{"usage":{"prompt_tokens":100,"completion_tokens":200}}`,
			expectedInput:     100,
			expectedOutput:    200,
			expectedCacheRead: 0,
		},
		{
			name: "with prompt_cache_hit_tokens",
			data: `{"usage":{"prompt_tokens":100,"completion_tokens":200,"prompt_cache_hit_tokens":20}}`,
			expectedInput:     100,
			expectedOutput:    200,
			expectedCacheRead: 20,
		},
		{
			name: "with prompt_tokens_details.cached_tokens",
			data: `{"usage":{"prompt_tokens":100,"completion_tokens":200,"prompt_tokens_details":{"cached_tokens":15}}`,
			expectedInput:     100,
			expectedOutput:    200,
			expectedCacheRead: 15,
		},
		{
			name: "cache_hit takes precedence over cached_tokens",
			data: `{"usage":{"prompt_tokens":100,"completion_tokens":200,"prompt_cache_hit_tokens":20,"prompt_tokens_details":{"cached_tokens":15}}`,
			expectedInput:     100,
			expectedOutput:    200,
			expectedCacheRead: 20, // prompt_cache_hit_tokens takes precedence
		},
		{
			name: "empty JSON",
			data: `{}`,
			expectedInput:     0,
			expectedOutput:    0,
			expectedCacheRead: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := &TokenUsage{}
			OpenCodeParseTokenUsageFromResponse(tt.data, usage)

			if usage.InputTokens != tt.expectedInput {
				t.Errorf("InputTokens = %d, want %d", usage.InputTokens, tt.expectedInput)
			}
			if usage.OutputTokens != tt.expectedOutput {
				t.Errorf("OutputTokens = %d, want %d", usage.OutputTokens, tt.expectedOutput)
			}
			if usage.CacheReadTokens != tt.expectedCacheRead {
				t.Errorf("CacheReadTokens = %d, want %d", usage.CacheReadTokens, tt.expectedCacheRead)
			}
		})
	}
}

// ==================== ParseEventPayload Tests ====================

func TestParseEventPayload(t *testing.T) {
	tests := []struct {
		name                string
		payload             string
		expectedInput       int
		expectedOutput      int
		expectedCacheRead   int
		useClaudeParser     bool
	}{
		{
			name: "single SSE event - claude",
			payload: "data: {\"message\":{\"usage\":{\"input_tokens\":100,\"output_tokens\":200}}}\n\n",
			expectedInput:     100,
			expectedOutput:    200,
			expectedCacheRead: 0,
			useClaudeParser:   true,
		},
		{
			name: "multiple SSE events - claude (accumulates)",
			payload: "data: {\"message\":{\"usage\":{\"input_tokens\":100,\"output_tokens\":200}}}\n\n" +
				"data: {\"message\":{\"usage\":{\"input_tokens\":50,\"output_tokens\":75}}}\n\n",
			expectedInput:     150, // 100 + 50
			expectedOutput:    275, // 200 + 75
			expectedCacheRead: 0,
			useClaudeParser:   true,
		},
		{
			name: "mixed lines (only data: parsed)",
			payload: "event: message\n" +
				"data: {\"message\":{\"usage\":{\"input_tokens\":100,\"output_tokens\":200}}}\n" +
				"id: 123\n" +
				"data: {\"message\":{\"usage\":{\"input_tokens\":50,\"output_tokens\":75}}}\n\n",
			expectedInput:     150, // 100 + 50
			expectedOutput:    275, // 200 + 75
			expectedCacheRead: 0,
			useClaudeParser:   true,
		},
		{
			name: "no data: lines",
			payload: "event: message\nid: 123\n\n",
			expectedInput:     0,
			expectedOutput:    0,
			expectedCacheRead: 0,
			useClaudeParser:   true,
		},
		{
			name: "empty payload",
			payload: "",
			expectedInput:     0,
			expectedOutput:    0,
			expectedCacheRead: 0,
			useClaudeParser:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := &TokenUsage{}

			var parser func(string, *TokenUsage)
			if tt.useClaudeParser {
				parser = ClaudeCodeParseTokenUsageFromResponse
			} else {
				parser = CodexParseTokenUsageFromResponse
			}

			ParseEventPayload(tt.payload, parser, usage)

			if usage.InputTokens != tt.expectedInput {
				t.Errorf("InputTokens = %d, want %d", usage.InputTokens, tt.expectedInput)
			}
			if usage.OutputTokens != tt.expectedOutput {
				t.Errorf("OutputTokens = %d, want %d", usage.OutputTokens, tt.expectedOutput)
			}
			if usage.CacheReadTokens != tt.expectedCacheRead {
				t.Errorf("CacheReadTokens = %d, want %d", usage.CacheReadTokens, tt.expectedCacheRead)
			}
		})
	}
}

// ==================== ReplaceModelInRequestBody Tests ====================

func TestReplaceModelInRequestBody(t *testing.T) {
	tests := []struct {
		name          string
		inputJSON     string
		newModel      string
		expectError   bool
		expectedModel string
	}{
		{
			name: "simple replacement",
			inputJSON: `{"model":"claude-sonnet-4","messages":[{"role":"user","content":"Hello"}]}`,
			newModel:      "anthropic/claude-sonnet-4",
			expectError:   false,
			expectedModel: "anthropic/claude-sonnet-4",
		},
		{
			name: "complex nested JSON",
			inputJSON: `{"model":"claude-opus-4","messages":[{"role":"user","content":"Test"}],"temperature":0.7}`,
			newModel:      "gpt-4",
			expectError:   false,
			expectedModel: "gpt-4",
		},
		{
			name:           "missing model field - sjson sets it",
			inputJSON:      `{"messages":[{"role":"user","content":"Hello"}]}`,
			newModel:       "any-model",
			expectError:    false,
			expectedModel:  "any-model",
		},
		{
			name:           "empty JSON - sjson sets model",
			inputJSON:      `{}`,
			newModel:       "any-model",
			expectError:    false,
			expectedModel:  "any-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes := []byte(tt.inputJSON)
			result, err := ReplaceModelInRequestBody(bodyBytes, tt.newModel)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				// Check model was replaced by checking if newModel value appears in result
				if !containsString(string(result), `"`+tt.newModel+`"`) {
					t.Errorf("Model %q not found in result: %s", tt.newModel, string(result))
				}
			}
		})
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
