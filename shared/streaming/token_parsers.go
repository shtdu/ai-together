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


package streaming

import (
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// TokenUsage represents token usage information extracted from API responses
type TokenUsage struct {
	InputTokens       int
	OutputTokens      int
	CacheCreateTokens int
	CacheReadTokens   int
	ReasoningTokens   int
}

// ParseTokenUsageFromResponse parses token usage from an API response based on the provider kind
// kind: "claude", "codex", or "opencode"
func ParseTokenUsageFromResponse(data string, kind string, usage *TokenUsage) {
	switch kind {
	case "codex":
		CodexParseTokenUsageFromResponse(data, usage)
	case "opencode":
		OpenCodeParseTokenUsageFromResponse(data, usage)
	default:
		ClaudeCodeParseTokenUsageFromResponse(data, usage)
	}
}

// ClaudeCodeParseTokenUsageFromResponse parses token usage from Anthropic API responses
func ClaudeCodeParseTokenUsageFromResponse(data string, usage *TokenUsage) {
	messageInput := int(gjson.Get(data, "message.usage.input_tokens").Int())
	messageOutput := int(gjson.Get(data, "message.usage.output_tokens").Int())
	messageCacheCreate := int(gjson.Get(data, "message.usage.cache_creation_input_tokens").Int())
	messageCacheRead := int(gjson.Get(data, "message.usage.cache_read_input_tokens").Int())

	usage.InputTokens += messageInput + messageCacheRead
	usage.OutputTokens += messageOutput
	usage.CacheCreateTokens += messageCacheCreate
	usage.CacheReadTokens += messageCacheRead

	topInput := int(gjson.Get(data, "usage.input_tokens").Int())
	topOutput := int(gjson.Get(data, "usage.output_tokens").Int())
	topCacheRead := int(gjson.Get(data, "usage.cache_read_input_tokens").Int())

	usage.InputTokens += topInput + topCacheRead
	usage.OutputTokens += topOutput
	usage.CacheReadTokens += topCacheRead
}

// CodexParseTokenUsageFromResponse parses token usage from Codex API responses
func CodexParseTokenUsageFromResponse(data string, usage *TokenUsage) {
	usage.InputTokens += int(gjson.Get(data, "response.usage.input_tokens").Int())
	usage.OutputTokens += int(gjson.Get(data, "response.usage.output_tokens").Int())
	usage.CacheReadTokens += int(gjson.Get(data, "response.usage.input_tokens_details.cached_tokens").Int())
	usage.ReasoningTokens += int(gjson.Get(data, "response.usage.output_tokens_details.reasoning_tokens").Int())
}

// OpenCodeParseTokenUsageFromResponse parses token usage from OpenAI-compatible API responses
func OpenCodeParseTokenUsageFromResponse(data string, usage *TokenUsage) {
	usage.InputTokens += int(gjson.Get(data, "usage.prompt_tokens").Int())
	usage.OutputTokens += int(gjson.Get(data, "usage.completion_tokens").Int())
	cacheHit := gjson.Get(data, "usage.prompt_cache_hit_tokens")
	if cacheHit.Exists() && cacheHit.Type != gjson.Null {
		usage.CacheReadTokens += int(cacheHit.Int())
	} else {
		usage.CacheReadTokens += int(gjson.Get(data, "usage.prompt_tokens_details.cached_tokens").Int())
	}
}

// ParseEventPayload parses SSE event payload and calls parser for each data line
func ParseEventPayload(payload string, parser func(string, *TokenUsage), usage *TokenUsage) {
	lines := strings.Split(payload, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data:") {
			parser(strings.TrimPrefix(line, "data: "), usage)
		}
	}
}

// ReplaceModelInRequestBody replaces the model field in a JSON request body
func ReplaceModelInRequestBody(body []byte, newModel string) ([]byte, error) {
	modified, err := sjson.SetBytes(body, "model", newModel)
	if err != nil {
		return body, err
	}

	return modified, nil
}
