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
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	modelpricing "codeswitch/resources/model-pricing"

	"codeswitch/internal/db"
	"codeswitch/internal/models"
)

const timeLayout = "2006-01-02 15:04:05"

type LogService struct {
	pricing *modelpricing.Service
}

func NewLogService() *LogService {
	svc, err := modelpricing.DefaultService()
	if err != nil {
		log.Printf("pricing service init failed: %v", err)
	}
	return &LogService{pricing: svc}
}

func (ls *LogService) ListRequestLogs(platform string, provider string, limit int) ([]models.RequestLog, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	// Use read-only connection to avoid lock contention with write operations
	roDB, err := db.GetReadOnlyDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get read-only database: %w", err)
	}

	query := `SELECT * FROM request_log WHERE 1=1`
	args := make([]interface{}, 0)
	argCount := 0

	if platform != "" {
		argCount++
		query += fmt.Sprintf(" AND platform = $%d", argCount)
		args = append(args, platform)
	}
	if provider != "" {
		argCount++
		query += fmt.Sprintf(" AND provider = $%d", argCount)
		args = append(args, provider)
	}

	query += fmt.Sprintf(" ORDER BY id DESC LIMIT $%d", argCount+1)
	args = append(args, limit)

	var logs []models.RequestLog
	err = roDB.Select(&logs, query, args...)
	if err != nil {
		return nil, err
	}

	// Decorate with cost information
	for i := range logs {
		ls.decorateCost(&logs[i])
	}

	return logs, nil
}

func (ls *LogService) ListProviders(platform string) ([]string, error) {
	// Use read-only connection to avoid lock contention with write operations
	roDB, err := db.GetReadOnlyDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get read-only database: %w", err)
	}

	query := `SELECT DISTINCT provider FROM request_log WHERE provider != ''`
	args := make([]interface{}, 0)
	argCount := 0

	if platform != "" {
		argCount++
		query += fmt.Sprintf(" AND platform = $%d", argCount)
		args = append(args, platform)
	}

	query += " ORDER BY provider ASC"

	var providers []string
	err = roDB.Select(&providers, query, args...)
	if err != nil {
		return nil, err
	}

	// Clean up and filter empty strings
	result := make([]string, 0, len(providers))
	for _, name := range providers {
		name = strings.TrimSpace(name)
		if name != "" {
			result = append(result, name)
		}
	}

	return result, nil
}

func (ls *LogService) HeatmapStats(days int) ([]models.HeatmapStat, error) {
	if days <= 0 {
		days = 30
	}
	totalHours := days * 24
	if totalHours <= 0 {
		totalHours = 24
	}
	rangeStart := startOfHour(time.Now())
	if totalHours > 1 {
		rangeStart = rangeStart.Add(-time.Duration(totalHours-1) * time.Hour)
	}

	// Use read-only connection to avoid lock contention with write operations
	roDB, err := db.GetReadOnlyDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get read-only database: %w", err)
	}

	query := `SELECT model, input_tokens, output_tokens, reasoning_tokens,
			  cache_create_tokens, cache_read_tokens, created_at
			  FROM request_log
			  WHERE created_at >= $1
			  ORDER BY created_at DESC`

	var logs []models.RequestLog
	err = roDB.Select(&logs, query, rangeStart.Format(timeLayout))
	if err != nil {
		if isNoSuchTableErr(err) {
			return []models.HeatmapStat{}, nil
		}
		return nil, err
	}

	hourBuckets := map[int64]*models.HeatmapStat{}
	for _, log := range logs {
		createdAt, ok := parseCreatedAt(log.CreatedAt)
		if !ok {
			continue
		}
		hourStart := startOfHour(createdAt)
		hourKey := hourStart.Unix()
		bucket := hourBuckets[hourKey]
		if bucket == nil {
			bucket = &models.HeatmapStat{Day: hourStart.Format("01-02 15")}
			hourBuckets[hourKey] = bucket
		}
		bucket.TotalRequests++
		bucket.InputTokens += int64(log.InputTokens)
		bucket.OutputTokens += int64(log.OutputTokens)
		bucket.ReasoningTokens += int64(log.ReasoningTokens)

		usage := modelpricing.UsageSnapshot{
			InputTokens:       log.InputTokens,
			OutputTokens:      log.OutputTokens,
			CacheCreateTokens: log.CacheCreateTokens,
			CacheReadTokens:   log.CacheReadTokens,
		}
		cost := ls.calculateCost(log.Model, usage)
		bucket.TotalCost += cost.TotalCost
	}

	if len(hourBuckets) == 0 {
		return []models.HeatmapStat{}, nil
	}

	hourKeys := make([]int64, 0, len(hourBuckets))
	for key := range hourBuckets {
		hourKeys = append(hourKeys, key)
	}
	sort.Slice(hourKeys, func(i, j int) bool {
		return hourKeys[i] < hourKeys[j]
	})

	stats := make([]models.HeatmapStat, 0, min(len(hourKeys), totalHours))
	for i := len(hourKeys) - 1; i >= 0 && len(stats) < totalHours; i-- {
		stats = append(stats, *hourBuckets[hourKeys[i]])
	}
	return stats, nil
}

func (ls *LogService) StatsSince(platform string) (models.LogStats, error) {
	const seriesHours = 24

	stats := models.LogStats{
		Series: make([]models.LogStatsSeries, 0, seriesHours),
	}
	now := time.Now()
	seriesStart := startOfHour(now).Add(-time.Duration(seriesHours-1) * time.Hour)
	seriesEnd := seriesStart.Add(seriesHours * time.Hour)
	queryStart := seriesStart
	summaryStart := seriesStart

	// Use read-only connection to avoid lock contention with write operations
	roDB, err := db.GetReadOnlyDB()
	if err != nil {
		return stats, fmt.Errorf("failed to get read-only database: %w", err)
	}

	query := `SELECT model, input_tokens, output_tokens, reasoning_tokens,
			  cache_create_tokens, cache_read_tokens, created_at
			  FROM request_log
			  WHERE created_at >= $1`
	args := []interface{}{queryStart.Format(timeLayout)}

	if platform != "" {
		query += " AND platform = $2"
		args = append(args, platform)
	}

	query += " ORDER BY created_at ASC"

	var logs []models.RequestLog
	err = roDB.Select(&logs, query, args...)
	if err != nil {
		if isNoSuchTableErr(err) {
			return stats, nil
		}
		return stats, err
	}

	seriesBuckets := make([]*models.LogStatsSeries, seriesHours)
	for i := 0; i < seriesHours; i++ {
		bucketTime := seriesStart.Add(time.Duration(i) * time.Hour)
		seriesBuckets[i] = &models.LogStatsSeries{
			Day: bucketTime.Format(timeLayout),
		}
	}

	for _, log := range logs {
		createdAt := log.CreatedAt

		if createdAt.IsZero() {
			createdAt = seriesStart
		}

		if !createdAt.IsZero() && (createdAt.Before(seriesStart) || !createdAt.Before(seriesEnd)) {
			continue
		}

		bucketIndex := 0
		if !createdAt.IsZero() {
			bucketIndex = int(createdAt.Sub(seriesStart) / time.Hour)
			if bucketIndex < 0 {
				bucketIndex = 0
			}
			if bucketIndex >= seriesHours {
				bucketIndex = seriesHours - 1
			}
		}
		bucket := seriesBuckets[bucketIndex]

		usage := modelpricing.UsageSnapshot{
			InputTokens:       log.InputTokens,
			OutputTokens:      log.OutputTokens,
			CacheCreateTokens: log.CacheCreateTokens,
			CacheReadTokens:   log.CacheReadTokens,
		}
		cost := ls.calculateCost(log.Model, usage)

		bucket.TotalRequests++
		bucket.InputTokens += int64(log.InputTokens)
		bucket.OutputTokens += int64(log.OutputTokens)
		bucket.ReasoningTokens += int64(log.ReasoningTokens)
		bucket.CacheCreateTokens += int64(log.CacheCreateTokens)
		bucket.CacheReadTokens += int64(log.CacheReadTokens)
		bucket.TotalCost += cost.TotalCost

		if createdAt.IsZero() || createdAt.Before(summaryStart) {
			continue
		}
		stats.TotalRequests++
		stats.InputTokens += int64(log.InputTokens)
		stats.OutputTokens += int64(log.OutputTokens)
		stats.ReasoningTokens += int64(log.ReasoningTokens)
		stats.CacheCreateTokens += int64(log.CacheCreateTokens)
		stats.CacheReadTokens += int64(log.CacheReadTokens)
		stats.CostInput += cost.InputCost
		stats.CostOutput += cost.OutputCost
		stats.CostCacheCreate += cost.CacheCreateCost
		stats.CostCacheRead += cost.CacheReadCost
		stats.CostTotal += cost.TotalCost
	}

	for i := 0; i < seriesHours; i++ {
		if bucket := seriesBuckets[i]; bucket != nil {
			stats.Series = append(stats.Series, *bucket)
		} else {
			bucketTime := seriesStart.Add(time.Duration(i) * time.Hour)
			stats.Series = append(stats.Series, models.LogStatsSeries{
				Day: bucketTime.Format(timeLayout),
			})
		}
	}

	return stats, nil
}

func (ls *LogService) ProviderDailyStats(platform string) ([]models.ProviderDailyStat, error) {
	start := startOfDay(time.Now())
	end := start.Add(24 * time.Hour)
	queryStart := start.Add(-24 * time.Hour)

	// Use read-only connection to avoid lock contention with write operations
	roDB, err := db.GetReadOnlyDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get read-only database: %w", err)
	}

	query := `SELECT provider, model, http_code, input_tokens, output_tokens,
			  reasoning_tokens, cache_create_tokens, cache_read_tokens, created_at
			  FROM request_log
			  WHERE created_at >= $1`
	args := []interface{}{queryStart.Format(timeLayout)}

	if platform != "" {
		query += " AND platform = $2"
		args = append(args, platform)
	}

	var logs []models.RequestLog
	err = roDB.Select(&logs, query, args...)
	if err != nil {
		if isNoSuchTableErr(err) {
			return []models.ProviderDailyStat{}, nil
		}
		return nil, err
	}

	statMap := map[string]*models.ProviderDailyStat{}
	for _, log := range logs {
		provider := strings.TrimSpace(log.Provider)
		if provider == "" {
			provider = "(unknown)"
		}
		createdAt := log.CreatedAt
		if !createdAt.IsZero() {
			if createdAt.Before(start) || !createdAt.Before(end) {
				continue
			}
		} else {
			dayKey := dayFromTimestamp(log.CreatedAt.Format(timeLayout))
			if dayKey != start.Format("2006-01-02") {
				continue
			}
		}
		stat := statMap[provider]
		if stat == nil {
			stat = &models.ProviderDailyStat{Provider: provider}
			statMap[provider] = stat
		}

		usage := modelpricing.UsageSnapshot{
			InputTokens:       log.InputTokens,
			OutputTokens:      log.OutputTokens,
			CacheCreateTokens: log.CacheCreateTokens,
			CacheReadTokens:   log.CacheReadTokens,
		}
		cost := ls.calculateCost(log.Model, usage)

		stat.TotalRequests++
		// Only HTTP 200-299 counts as success, others (including 0) count as failure
		if log.HttpCode >= 200 && log.HttpCode < 300 {
			stat.SuccessfulRequests++
		} else {
			stat.FailedRequests++
		}
		stat.InputTokens += int64(log.InputTokens)
		stat.OutputTokens += int64(log.OutputTokens)
		stat.ReasoningTokens += int64(log.ReasoningTokens)
		stat.CacheCreateTokens += int64(log.CacheCreateTokens)
		stat.CacheReadTokens += int64(log.CacheReadTokens)
		stat.CostTotal += cost.TotalCost
	}

	stats := make([]models.ProviderDailyStat, 0, len(statMap))
	for _, stat := range statMap {
		if stat.TotalRequests > 0 {
			stat.SuccessRate = float64(stat.SuccessfulRequests) / float64(stat.TotalRequests)
		}
		stats = append(stats, *stat)
	}
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].TotalRequests == stats[j].TotalRequests {
			return stats[i].Provider < stats[j].Provider
		}
		return stats[i].TotalRequests > stats[j].TotalRequests
	})
	return stats, nil
}

func (ls *LogService) decorateCost(logEntry *models.RequestLog) {
	if ls == nil || ls.pricing == nil || logEntry == nil {
		return
	}
	usage := modelpricing.UsageSnapshot{
		InputTokens:       logEntry.InputTokens,
		OutputTokens:      logEntry.OutputTokens,
		CacheCreateTokens: logEntry.CacheCreateTokens,
		CacheReadTokens:   logEntry.CacheReadTokens,
	}
	cost := ls.pricing.CalculateCost(logEntry.Model, usage)
	logEntry.HasPricing = cost.HasPricing
	logEntry.InputCost = cost.InputCost
	logEntry.OutputCost = cost.OutputCost
	logEntry.CacheCreateCost = cost.CacheCreateCost
	logEntry.CacheReadCost = cost.CacheReadCost
	logEntry.Ephemeral5mCost = cost.Ephemeral5mCost
	logEntry.Ephemeral1hCost = cost.Ephemeral1hCost
	logEntry.TotalCost = cost.TotalCost
}

func (ls *LogService) calculateCost(model string, usage modelpricing.UsageSnapshot) modelpricing.CostBreakdown {
	if ls == nil || ls.pricing == nil {
		return modelpricing.CostBreakdown{}
	}
	return ls.pricing.CalculateCost(model, usage)
}

func parseCreatedAt(t time.Time) (time.Time, bool) {
	if !t.IsZero() {
		return t.In(time.Local), true
	}
	return time.Time{}, false
}

func dayFromTimestamp(value string) string {
	if len(value) >= len("2006-01-02") {
		if t, err := time.ParseInLocation(timeLayout, value, time.Local); err == nil {
			return t.Format("2006-01-02")
		}
		return value[:10]
	}
	return value
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func startOfHour(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, t.Hour(), 0, 0, 0, t.Location())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func isNoSuchTableErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "no such table")
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
