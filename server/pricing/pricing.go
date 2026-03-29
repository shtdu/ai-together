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

// Package pricing provides per-model token pricing for cost estimation.
// Prices are expressed as USD per 1 million tokens.
// Update this table when providers change their pricing.
package pricing

// Price represents per-million-token pricing for a model.
type Price struct {
	Input  float64 // USD per 1M input tokens
	Output float64 // USD per 1M output tokens
}

// modelPrices maps model names to their per-million-token pricing.
// Prices sourced from provider pricing pages (as of 2025-03).
// Unknown models fall back to DefaultPrice.
var modelPrices = map[string]Price{
	// Anthropic Claude
	"claude-sonnet-4-20250514":  {Input: 3.0, Output: 15.0},
	"claude-3-5-sonnet-20241022": {Input: 3.0, Output: 15.0},
	"claude-3-5-sonnet-latest":   {Input: 3.0, Output: 15.0},
	"claude-3-5-haiku-20241022":  {Input: 0.8, Output: 4.0},
	"claude-3-5-haiku-latest":    {Input: 0.8, Output: 4.0},
	"claude-3-opus-20240229":     {Input: 15.0, Output: 75.0},
	"claude-3-opus-latest":       {Input: 15.0, Output: 75.0},
	"claude-3-sonnet-20240229":   {Input: 3.0, Output: 15.0},
	"claude-3-haiku-20240307":    {Input: 0.25, Output: 1.25},

	// OpenAI GPT
	"gpt-4.1":                    {Input: 2.0, Output: 8.0},
	"gpt-4.1-mini":               {Input: 0.4, Output: 1.6},
	"gpt-4.1-nano":               {Input: 0.1, Output: 0.4},
	"gpt-4o":                     {Input: 2.5, Output: 10.0},
	"gpt-4o-mini":                {Input: 0.15, Output: 0.6},
	"gpt-4-turbo":                {Input: 10.0, Output: 30.0},
	"gpt-4":                      {Input: 30.0, Output: 60.0},
	"gpt-3.5-turbo":              {Input: 0.5, Output: 1.5},
	"o1":                         {Input: 15.0, Output: 60.0},
	"o1-mini":                    {Input: 1.10, Output: 4.40},
	"o1-preview":                 {Input: 15.0, Output: 60.0},
	"o3-mini":                    {Input: 1.10, Output: 4.40},
	"o3":                         {Input: 10.0, Output: 40.0},
	"o4-mini":                    {Input: 1.10, Output: 4.40},

	// Google Gemini
	"gemini-2.5-pro":             {Input: 1.25, Output: 10.0},
	"gemini-2.5-flash":           {Input: 0.15, Output: 0.60},
	"gemini-2.0-flash":           {Input: 0.10, Output: 0.40},
	"gemini-2.0-flash-lite":      {Input: 0.075, Output: 0.30},
	"gemini-1.5-pro":             {Input: 1.25, Output: 5.0},
	"gemini-1.5-flash":           {Input: 0.075, Output: 0.30},

	// DeepSeek
	"deepseek-chat":              {Input: 0.14, Output: 0.28},
	"deepseek-reasoner":          {Input: 0.55, Output: 2.19},

	// Meta Llama (via providers like Together, Fireworks, etc.)
	"llama-3.3-70b-instruct":     {Input: 0.39, Output: 0.39},
	"llama-3.1-405b-instruct":    {Input: 2.00, Output: 2.00},
	"llama-3.1-70b-instruct":     {Input: 0.39, Output: 0.39},
	"llama-3.1-8b-instruct":      {Input: 0.06, Output: 0.06},

	// Mistral
	"mistral-large-latest":       {Input: 2.0, Output: 6.0},
	"mistral-medium-latest":      {Input: 2.7, Output: 8.1},
	"mistral-small-latest":       {Input: 0.20, Output: 0.60},
	"codestral-latest":           {Input: 0.30, Output: 0.90},

	// Qwen
	"qwen-2.5-coder-32b":        {Input: 0.18, Output: 0.18},
	"qwen-2.5-72b-instruct":     {Input: 0.40, Output: 0.40},
}

// DefaultPrice is used when a model is not found in the price table.
// A conservative middle-ground estimate for unknown models.
var DefaultPrice = Price{Input: 3.0, Output: 15.0}

// GetPrice returns the per-million-token pricing for a given model name.
// Falls back to DefaultPrice for unknown models.
func GetPrice(model string) Price {
	if p, ok := modelPrices[model]; ok {
		return p
	}
	return DefaultPrice
}

// EstimateCost calculates the estimated cost in USD for a single request.
// inputTokens and outputTokens are raw token counts (not per-million).
func EstimateCost(model string, inputTokens, outputTokens int64) float64 {
	p := GetPrice(model)
	inputCost := float64(inputTokens) * p.Input / 1_000_000
	outputCost := float64(outputTokens) * p.Output / 1_000_000
	return inputCost + outputCost
}
