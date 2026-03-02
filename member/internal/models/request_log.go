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


package models

import (
	"time"
)

type RequestLog struct {
	ID                int64     `db:"id" json:"id"`
	Platform          string    `db:"platform" json:"platform"` // claude code or codex
	Model             string    `db:"model" json:"model"`
	Provider          string    `db:"provider" json:"provider"` // provider name
	HttpCode          int       `db:"http_code" json:"http_code"`
	InputTokens       int       `db:"input_tokens" json:"input_tokens"`
	OutputTokens      int       `db:"output_tokens" json:"output_tokens"`
	CacheCreateTokens int       `db:"cache_create_tokens" json:"cache_create_tokens"`
	CacheReadTokens   int       `db:"cache_read_tokens" json:"cache_read_tokens"`
	ReasoningTokens   int       `db:"reasoning_tokens" json:"reasoning_tokens"`
	IsStream          bool      `db:"is_stream" json:"is_stream"`
	DurationSec       float64   `db:"duration_sec" json:"duration_sec"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	SyncedToServer    bool      `db:"synced_to_server" json:"synced_to_server"` // whether synced to server
	InputCost         float64   `json:"input_cost"`
	OutputCost        float64   `json:"output_cost"`
	CacheCreateCost   float64   `json:"cache_create_cost"`
	CacheReadCost     float64   `json:"cache_read_cost"`
	Ephemeral5mCost   float64   `json:"ephemeral_5m_cost"`
	Ephemeral1hCost   float64   `json:"ephemeral_1h_cost"`
	TotalCost         float64   `json:"total_cost"`
	HasPricing        bool      `json:"has_pricing"`
}

type HeatmapStat struct {
	Day             string  `json:"day"`
	TotalRequests   int64   `json:"total_requests"`
	InputTokens     int64   `json:"input_tokens"`
	OutputTokens    int64   `json:"output_tokens"`
	ReasoningTokens int64   `json:"reasoning_tokens"`
	TotalCost       float64 `json:"total_cost"`
}

type LogStats struct {
	TotalRequests     int64            `json:"total_requests"`
	InputTokens       int64            `json:"input_tokens"`
	OutputTokens      int64            `json:"output_tokens"`
	ReasoningTokens   int64            `json:"reasoning_tokens"`
	CacheCreateTokens int64            `json:"cache_create_tokens"`
	CacheReadTokens   int64            `json:"cache_read_tokens"`
	CostTotal         float64          `json:"cost_total"`
	CostInput         float64          `json:"cost_input"`
	CostOutput        float64          `json:"cost_output"`
	CostCacheCreate   float64          `json:"cost_cache_create"`
	CostCacheRead     float64          `json:"cost_cache_read"`
	Series            []LogStatsSeries `json:"series"`
}

type ProviderDailyStat struct {
	Provider           string  `json:"provider"`
	TotalRequests      int64   `json:"total_requests"`
	SuccessfulRequests int64   `json:"successful_requests"`
	FailedRequests     int64   `json:"failed_requests"`
	SuccessRate        float64 `json:"success_rate"`
	InputTokens        int64   `json:"input_tokens"`
	OutputTokens       int64   `json:"output_tokens"`
	ReasoningTokens    int64   `json:"reasoning_tokens"`
	CacheCreateTokens  int64   `json:"cache_create_tokens"`
	CacheReadTokens    int64   `json:"cache_read_tokens"`
	CostTotal          float64 `json:"cost_total"`
}

type LogStatsSeries struct {
	Day               string  `json:"day"`
	TotalRequests     int64   `json:"total_requests"`
	InputTokens       int64   `json:"input_tokens"`
	OutputTokens      int64   `json:"output_tokens"`
	ReasoningTokens   int64   `json:"reasoning_tokens"`
	CacheCreateTokens int64   `json:"cache_create_tokens"`
	CacheReadTokens   int64   `json:"cache_read_tokens"`
	TotalCost         float64 `json:"total_cost"`
}
