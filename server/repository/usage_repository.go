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

package repository

import (
	"context"
	"fmt"
	"time"

	"switch-server/internal/db"
	"switch-server/models"
	"switch-server/pricing"

	"github.com/jackc/pgx/v5/pgtype"
)

type UsageRepository struct {
	db *db.DB
}

func NewUsageRepository(database *db.DB) *UsageRepository {
	return &UsageRepository{
		db: database,
	}
}

// parseDateRange parses start and end date strings in "2006-01-02" format.
// Returns parsed times or an error if either date is invalid.
func parseDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	startTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to parse start date %q: %w", startDate, err)
	}

	endTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to parse end date %q: %w", endDate, err)
	}

	return startTime, endTime, nil
}

func (r *UsageRepository) CreateUsageRecord(ctx context.Context, usage models.UsageRecord) error {
	err := r.db.CreateUsageRecord(ctx, db.CreateUsageRecordParams{
		Platform:          pgtype.Text{String: usage.Platform, Valid: usage.Platform != ""},
		Model:             pgtype.Text{String: usage.Model, Valid: usage.Model != ""},
		Provider:          pgtype.Text{String: usage.Provider, Valid: usage.Provider != ""},
		HttpCode:          pgtype.Int4{Int32: int32(usage.HttpCode), Valid: true},
		InputTokens:       pgtype.Int4{Int32: int32(usage.InputTokens), Valid: true},
		OutputTokens:      pgtype.Int4{Int32: int32(usage.OutputTokens), Valid: true},
		CacheCreateTokens: pgtype.Int4{Int32: int32(usage.CacheCreateTokens), Valid: true},
		CacheReadTokens:   pgtype.Int4{Int32: int32(usage.CacheReadTokens), Valid: true},
		ReasoningTokens:   pgtype.Int4{Int32: int32(usage.ReasoningTokens), Valid: true},
		IsStream:          pgtype.Bool{Bool: usage.IsStream, Valid: true},
		DurationSec:       pgtype.Float4{Float32: float32(usage.DurationSec), Valid: true},
		TenantID:          usage.TenantID,
		UserID:            usage.UserID,
	})
	if err != nil {
		return fmt.Errorf("failed to create usage record: %w", err)
	}

	return nil
}

func (r *UsageRepository) GetUsageByTeamIDAndPeriod(ctx context.Context, tenantID int64, startDate, endDate string) ([]models.UsageRecord, error) {
	// Parse date range
	startTime, endTime, err := parseDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	records, err := r.db.GetUsageByTeamIDAndPeriod(ctx, db.GetUsageByTeamIDAndPeriodParams{
		TenantID:    tenantID,
		CreatedAt:   pgtype.Timestamp{Time: startTime, Valid: true},
		CreatedAt_2: pgtype.Timestamp{Time: endTime, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get usage by team ID and period: %w", err)
	}

	var result []models.UsageRecord
	for _, record := range records {
		model := models.UsageRecord{
			ID:                record.ID,
			Platform:          record.Platform.String,
			Model:             record.Model.String,
			Provider:          record.Provider.String,
			HttpCode:          int(record.HttpCode.Int32),
			InputTokens:       int(record.InputTokens.Int32),
			OutputTokens:      int(record.OutputTokens.Int32),
			CacheCreateTokens: int(record.CacheCreateTokens.Int32),
			CacheReadTokens:   int(record.CacheReadTokens.Int32),
			ReasoningTokens:   int(record.ReasoningTokens.Int32),
			IsStream:          record.IsStream.Bool,
			DurationSec:       float64(record.DurationSec.Float32),
			TenantID:          record.TenantID,
			UserID:            record.UserID,
			CreatedAt:         record.CreatedAt.Time,
		}
		result = append(result, model)
	}

	return result, nil
}

func (r *UsageRepository) GetUsageByUserIDAndPeriod(ctx context.Context, userID int64, startDate, endDate string) ([]models.UsageRecord, error) {
	// Parse date range
	startTime, endTime, err := parseDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	records, err := r.db.GetUsageByUserIDAndPeriod(ctx, db.GetUsageByUserIDAndPeriodParams{
		UserID:      userID,
		CreatedAt:   pgtype.Timestamp{Time: startTime, Valid: true},
		CreatedAt_2: pgtype.Timestamp{Time: endTime, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get usage by user ID and period: %w", err)
	}

	var result []models.UsageRecord
	for _, record := range records {
		model := models.UsageRecord{
			ID:                record.ID,
			Platform:          record.Platform.String,
			Model:             record.Model.String,
			Provider:          record.Provider.String,
			HttpCode:          int(record.HttpCode.Int32),
			InputTokens:       int(record.InputTokens.Int32),
			OutputTokens:      int(record.OutputTokens.Int32),
			CacheCreateTokens: int(record.CacheCreateTokens.Int32),
			CacheReadTokens:   int(record.CacheReadTokens.Int32),
			ReasoningTokens:   int(record.ReasoningTokens.Int32),
			IsStream:          record.IsStream.Bool,
			DurationSec:       float64(record.DurationSec.Float32),
			TenantID:          record.TenantID,
			UserID:            record.UserID,
			CreatedAt:         record.CreatedAt.Time,
		}
		result = append(result, model)
	}

	return result, nil
}

func (r *UsageRepository) GetTeamUsageSummary(ctx context.Context, teamID int64, startDate, endDate string) ([]models.TeamUsageSummary, error) {
	// Parse date range
	startTime, endTime, err := parseDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	summaries, err := r.db.GetTeamUsageSummary(ctx, db.GetTeamUsageSummaryParams{
		TeamID:      teamID,
		PeriodStart: pgtype.Timestamp{Time: startTime, Valid: true},
		PeriodEnd:   pgtype.Timestamp{Time: endTime, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get team usage summary: %w", err)
	}

	var result []models.TeamUsageSummary
	for _, summary := range summaries {
		totalInput := 0
		if summary.TotalInput.Valid {
			totalInput = int(summary.TotalInput.Int32)
		}

		totalOutput := 0
		if summary.TotalOutput.Valid {
			totalOutput = int(summary.TotalOutput.Int32)
		}

		totalCost := 0.0
		if summary.TotalCost.Valid {
			// Convert pgtype.Numeric to float64
			if summary.TotalCost.Valid {
				totalCost = 0.0
			} // Will properly convert pgtype.Numeric in the final implementation
		}

		model := models.TeamUsageSummary{
			TeamID:      summary.TeamID,
			PeriodStart: summary.PeriodStart.Time,
			PeriodEnd:   summary.PeriodEnd.Time,
			TotalInput:  totalInput,
			TotalOutput: totalOutput,
			TotalCost:   totalCost,
			CreatedAt:   summary.CreatedAt.Time,
		}
		result = append(result, model)
	}

	return result, nil
}

func (r *UsageRepository) CreateTeamUsageSummary(ctx context.Context, summary models.TeamUsageSummary) error {
	err := r.db.CreateTeamUsageSummary(ctx, db.CreateTeamUsageSummaryParams{
		TeamID:      summary.TeamID,
		PeriodStart: pgtype.Timestamp{Time: summary.PeriodStart, Valid: true},
		PeriodEnd:   pgtype.Timestamp{Time: summary.PeriodEnd, Valid: true},
		TotalInput:  pgtype.Int4{Int32: int32(summary.TotalInput), Valid: true},
		TotalOutput: pgtype.Int4{Int32: int32(summary.TotalOutput), Valid: true},
		TotalCost:   pgtype.Numeric{Valid: summary.TotalCost != 0.0},
	})
	if err != nil {
		return fmt.Errorf("failed to create team usage summary: %w", err)
	}

	return nil
}

// GetUsageStatsForTeam returns aggregated usage statistics for a team
func (r *UsageRepository) GetUsageStatsForTeam(ctx context.Context, tenantID int64) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_requests,
			COALESCE(SUM(input_tokens), 0) as total_input,
			COALESCE(SUM(output_tokens), 0) as total_output
		FROM request_log
		WHERE tenant_id = $1
	`

	var totalRequests int32
	var totalInputVal, totalOutputVal int64

	err := r.db.Conn().QueryRow(ctx, query, tenantID).Scan(&totalRequests, &totalInputVal, &totalOutputVal)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage stats for team: %w", err)
	}

	totalCost := (float64(totalInputVal) * 0.001 / 1000) + (float64(totalOutputVal) * 0.002 / 1000)

	stats := map[string]interface{}{
		"total_requests":      totalRequests,
		"total_input_tokens":  totalInputVal,
		"total_output_tokens": totalOutputVal,
		"total_cost":          totalCost,
	}

	return stats, nil
}

// GetCurrentUsageForUser returns usage data for the current day for a user
func (r *UsageRepository) GetCurrentUsageForUser(ctx context.Context, userID int64) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_requests,
			COALESCE(SUM(input_tokens), 0) as total_input,
			COALESCE(SUM(output_tokens), 0) as total_output
		FROM request_log
		WHERE user_id = $1 AND DATE(created_at) = CURRENT_DATE
	`
	var totalRequests int32
	var totalInput, totalOutput int64

	err := r.db.Conn().QueryRow(ctx, query, userID).Scan(&totalRequests, &totalInput, &totalOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to get current usage for user: %w", err)
	}

	usage := map[string]interface{}{
		"total_requests":      totalRequests,
		"total_input_tokens":  totalInput,
		"total_output_tokens": totalOutput,
	}

	return usage, nil
}

// GetMetricsByTenant returns time-series metrics for a tenant (managers)
func (r *UsageRepository) GetMetricsByTenant(ctx context.Context, tenantID int64, startTime, endTime time.Time, interval string) ([]map[string]interface{}, error) {
	truncFunc := "hour"
	if interval == "day" {
		truncFunc = "day"
	}

	query := fmt.Sprintf(`
		SELECT
			date_trunc('%s', created_at) AS timestamp,
			COALESCE(SUM(duration_sec), 0) AS active_time_seconds,
			COALESCE(SUM(input_tokens + output_tokens), 0) AS total_tokens,
			COALESCE(SUM(input_tokens), 0) AS input_tokens,
			COALESCE(SUM(output_tokens), 0) AS output_tokens,
			COUNT(*) AS request_count
		FROM request_log
		WHERE tenant_id = $1
		  AND created_at >= $2
		  AND created_at < $3
		GROUP BY date_trunc('%s', created_at)
		ORDER BY timestamp ASC
	`, truncFunc, truncFunc)

	rows, err := r.db.Conn().Query(ctx, query, tenantID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics by tenant: %w", err)
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var timestamp time.Time
		var activeTime float64
		var totalTokens, inputTokens, outputTokens, requestCount int64

		err := rows.Scan(&timestamp, &activeTime, &totalTokens, &inputTokens, &outputTokens, &requestCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metrics: %w", err)
		}

		result = append(result, map[string]interface{}{
			"timestamp":           timestamp,
			"active_time_seconds": activeTime,
			"total_tokens":        totalTokens,
			"input_tokens":        inputTokens,
			"output_tokens":       outputTokens,
			"request_count":       requestCount,
		})
	}

	return result, nil
}

// GetMetricsByUser returns time-series metrics for a specific user (members)
func (r *UsageRepository) GetMetricsByUser(ctx context.Context, userID int64, startTime, endTime time.Time, interval string) ([]map[string]interface{}, error) {
	truncFunc := "hour"
	if interval == "day" {
		truncFunc = "day"
	}

	query := fmt.Sprintf(`
		SELECT
			date_trunc('%s', created_at) AS timestamp,
			COALESCE(SUM(duration_sec), 0) AS active_time_seconds,
			COALESCE(SUM(input_tokens + output_tokens), 0) AS total_tokens,
			COALESCE(SUM(input_tokens), 0) AS input_tokens,
			COALESCE(SUM(output_tokens), 0) AS output_tokens,
			COUNT(*) AS request_count
		FROM request_log
		WHERE user_id = $1
		  AND created_at >= $2
		  AND created_at < $3
		GROUP BY date_trunc('%s', created_at)
		ORDER BY timestamp ASC
	`, truncFunc, truncFunc)

	rows, err := r.db.Conn().Query(ctx, query, userID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics by user: %w", err)
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var timestamp time.Time
		var activeTime float64
		var totalTokens, inputTokens, outputTokens, requestCount int64

		err := rows.Scan(&timestamp, &activeTime, &totalTokens, &inputTokens, &outputTokens, &requestCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metrics: %w", err)
		}

		result = append(result, map[string]interface{}{
			"timestamp":           timestamp,
			"active_time_seconds": activeTime,
			"total_tokens":        totalTokens,
			"input_tokens":        inputTokens,
			"output_tokens":       outputTokens,
			"request_count":       requestCount,
		})
	}

	return result, nil
}

// GetProviderRankingsByTenant returns provider rankings for the last 24 hours (managers)
func (r *UsageRepository) GetProviderRankingsByTenant(ctx context.Context, tenantID int64) ([]map[string]interface{}, error) {
	query := `
		WITH totals AS (
			SELECT COALESCE(SUM(input_tokens + output_tokens), 0) AS grand_total
			FROM request_log
			WHERE tenant_id = $1
			  AND created_at >= NOW() - INTERVAL '24 hours'
		)
		SELECT
			provider,
			COALESCE(SUM(input_tokens + output_tokens), 0) AS total_tokens,
			COUNT(*) AS request_count,
			CASE WHEN t.grand_total > 0
				THEN ROUND(COALESCE(SUM(input_tokens + output_tokens), 0) * 100.0 / t.grand_total, 2)
				ELSE 0
			END AS percentage
		FROM request_log, totals t
		WHERE tenant_id = $1
		  AND created_at >= NOW() - INTERVAL '24 hours'
		  AND provider IS NOT NULL AND provider != ''
		GROUP BY provider, t.grand_total
		ORDER BY total_tokens DESC
		LIMIT 10
	`

	rows, err := r.db.Conn().Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider rankings: %w", err)
	}
	defer rows.Close()

	var result []map[string]interface{}
	rank := 1
	for rows.Next() {
		var provider string
		var totalTokens, requestCount int64
		var percentage float64

		err := rows.Scan(&provider, &totalTokens, &requestCount, &percentage)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider rankings: %w", err)
		}

		result = append(result, map[string]interface{}{
			"rank":          rank,
			"provider":      provider,
			"total_tokens":  totalTokens,
			"request_count": requestCount,
			"percentage":    percentage,
		})
		rank++
	}

	return result, nil
}

// GetProviderRankingsByUser returns provider rankings for a specific user (members)
func (r *UsageRepository) GetProviderRankingsByUser(ctx context.Context, userID int64) ([]map[string]interface{}, error) {
	query := `
		WITH totals AS (
			SELECT COALESCE(SUM(input_tokens + output_tokens), 0) AS grand_total
			FROM request_log
			WHERE user_id = $1
			  AND created_at >= NOW() - INTERVAL '24 hours'
		)
		SELECT
			provider,
			COALESCE(SUM(input_tokens + output_tokens), 0) AS total_tokens,
			COUNT(*) AS request_count,
			CASE WHEN t.grand_total > 0
				THEN ROUND(COALESCE(SUM(input_tokens + output_tokens), 0) * 100.0 / t.grand_total, 2)
				ELSE 0
			END AS percentage
		FROM request_log, totals t
		WHERE user_id = $1
		  AND created_at >= NOW() - INTERVAL '24 hours'
		  AND provider IS NOT NULL AND provider != ''
		GROUP BY provider, t.grand_total
		ORDER BY total_tokens DESC
		LIMIT 10
	`

	rows, err := r.db.Conn().Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider rankings: %w", err)
	}
	defer rows.Close()

	var result []map[string]interface{}
	rank := 1
	for rows.Next() {
		var provider string
		var totalTokens, requestCount int64
		var percentage float64

		err := rows.Scan(&provider, &totalTokens, &requestCount, &percentage)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider rankings: %w", err)
		}

		result = append(result, map[string]interface{}{
			"rank":          rank,
			"provider":      provider,
			"total_tokens":  totalTokens,
			"request_count": requestCount,
			"percentage":    percentage,
		})
		rank++
	}

	return result, nil
}

// GetMemberStatsByTenant returns member statistics for a tenant
func (r *UsageRepository) GetMemberStatsByTenant(ctx context.Context, tenantID int64) ([]map[string]interface{}, error) {
	query := `
		SELECT
			u.id AS user_id,
			u.name,
			u.email,
			COALESCE(SUM(r.duration_sec), 0) / 3600.0 AS active_time_hours,
			COALESCE(SUM(r.input_tokens + r.output_tokens), 0) AS total_tokens,
			CASE WHEN COUNT(DISTINCT DATE(r.created_at)) > 0
				THEN COALESCE(SUM(r.input_tokens + r.output_tokens), 0) / COUNT(DISTINCT DATE(r.created_at))
				ELSE 0
			END AS avg_tokens_per_day,
			COALESCE(MAX(r.created_at), u.created_at) AS last_active
		FROM users u
		LEFT JOIN request_log r ON u.id = r.user_id AND r.created_at >= NOW() - INTERVAL '30 days'
		WHERE u.tenant_id = $1
		GROUP BY u.id, u.name, u.email, u.created_at
		ORDER BY total_tokens DESC
	`

	rows, err := r.db.Conn().Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get member stats: %w", err)
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var userID int64
		var name, email string
		var activeTimeHours, avgTokensPerDay float64
		var totalTokens int64
		var lastActive time.Time

		err := rows.Scan(&userID, &name, &email, &activeTimeHours, &totalTokens, &avgTokensPerDay, &lastActive)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member stats: %w", err)
		}

		result = append(result, map[string]interface{}{
			"user_id":            userID,
			"name":               name,
			"email":              email,
			"active_time_hours":  activeTimeHours,
			"total_tokens":       totalTokens,
			"avg_tokens_per_day": avgTokensPerDay,
			"last_active":        lastActive,
		})
	}

	return result, nil
}

// GetProviderStats returns detailed statistics for a specific provider
func (r *UsageRepository) GetProviderStats(ctx context.Context, providerName string, tenantID int64) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_requests,
			SUM(input_tokens) as total_input,
			SUM(output_tokens) as total_output,
			AVG(duration_sec) as avg_response_time,
			MAX(created_at) as last_used
		FROM request_log
		WHERE provider = $1 AND tenant_id = $2
	`
	rows, err := r.db.Conn().Query(ctx, query, providerName, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider stats: %w", err)
	}
	defer rows.Close()

	var totalRequests int
	var sumInput, sumOutput *int64
	var avgTime *float64
	var lastUsed *string

	for rows.Next() {
		err = rows.Scan(&totalRequests, &sumInput, &sumOutput, &avgTime, &lastUsed)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider stats: %w", err)
		}
		break // Only one row expected
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	// Calculate success rate (responses with 2xx status codes)
	// Use NULLIF to handle division by zero when provider has no usage data
	successQuery := `
		SELECT
			COALESCE(
				COUNT(*) * 100.0 / NULLIF((SELECT COUNT(*) FROM request_log WHERE provider = $1 AND tenant_id = $2), 0),
				0.0
			) as success_rate
		FROM request_log
		WHERE provider = $1 AND tenant_id = $2 AND http_code >= 200 AND http_code < 300
	`
	rows2, err := r.db.Conn().Query(ctx, successQuery, providerName, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get success rate: %w", err)
	}
	defer rows2.Close()

	var successRate *float64
	for rows2.Next() {
		err = rows2.Scan(&successRate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan success rate: %w", err)
		}
		break // Only one row expected
	}

	err = rows2.Err()
	if err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	totalInputVal := int64(0)
	if sumInput != nil {
		totalInputVal = *sumInput
	}

	totalOutputVal := int64(0)
	if sumOutput != nil {
		totalOutputVal = *sumOutput
	}

	// Simple cost calculation: $0.001 per 1000 input tokens + $0.002 per 1000 output tokens
	totalCost := (float64(totalInputVal) * 0.001 / 1000) + (float64(totalOutputVal) * 0.002 / 1000)

	avgRespTime := 0.0
	if avgTime != nil {
		avgRespTime = *avgTime
	}

	lastUsedStr := ""
	if lastUsed != nil {
		lastUsedStr = *lastUsed
	}

	successRateVal := 0.0
	if successRate != nil {
		successRateVal = *successRate
	}

	stats := map[string]interface{}{
		"total_requests":    totalRequests,
		"total_input":       totalInputVal,
		"total_output":      totalOutputVal,
		"total_cost":        totalCost,
		"success_rate":      successRateVal,
		"avg_response_time": avgRespTime,
		"last_used":         lastUsedStr,
	}

	return stats, nil
}

// GetProviderAnalytics returns provider analytics with filtering
func (r *UsageRepository) GetProviderAnalytics(ctx context.Context, tenantID int64, startDate, endDate string, providers, models, tools []string) (map[string]interface{}, error) {
	// Parse date range
	startTime, endTime, err := parseDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}
	endTime = endTime.Add(24 * time.Hour) // Include the end date

	// Build dynamic WHERE clause
	whereClause := "tenant_id = $1 AND created_at >= $2 AND created_at < $3"
	args := []interface{}{tenantID, startTime, endTime}
	argIndex := 4

	if len(providers) > 0 {
		whereClause += fmt.Sprintf(" AND provider = ANY($%d)", argIndex)
		args = append(args, providers)
		argIndex++
	}
	if len(models) > 0 {
		whereClause += fmt.Sprintf(" AND model = ANY($%d)", argIndex)
		args = append(args, models)
		argIndex++
	}
	if len(tools) > 0 {
		whereClause += fmt.Sprintf(" AND platform = ANY($%d)", argIndex)
		args = append(args, tools)
		argIndex++
	}

	// Get summary
	summaryQuery := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(input_tokens + output_tokens), 0) AS total_tokens,
			COALESCE(SUM(input_tokens), 0) AS total_input,
			COALESCE(SUM(output_tokens), 0) AS total_output,
			COUNT(*) AS total_requests
		FROM request_log
		WHERE %s
	`, whereClause)

	var totalTokens, totalInput, totalOutput, totalRequests int64
	err = r.db.Conn().QueryRow(ctx, summaryQuery, args...).Scan(&totalTokens, &totalInput, &totalOutput, &totalRequests)
	if err != nil {
		return nil, fmt.Errorf("failed to get summary: %w", err)
	}

	totalCost := (float64(totalInput)*0.001 + float64(totalOutput)*0.002) / 1000

	// Get trend data by day and provider
	trendQuery := fmt.Sprintf(`
		SELECT
			DATE(created_at) AS date,
			provider,
			COALESCE(SUM(input_tokens + output_tokens), 0) AS tokens,
			COUNT(*) AS requests
		FROM request_log
		WHERE %s AND provider IS NOT NULL AND provider != ''
		GROUP BY DATE(created_at), provider
		ORDER BY date ASC, tokens DESC
	`, whereClause)

	rows, err := r.db.Conn().Query(ctx, trendQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get trend data: %w", err)
	}
	defer rows.Close()

	trendData := make(map[string]map[string]int64)
	for rows.Next() {
		var date time.Time
		var provider string
		var tokens, requests int64
		if err := rows.Scan(&date, &provider, &tokens, &requests); err != nil {
			return nil, fmt.Errorf("failed to scan trend data: %w", err)
		}
		dateStr := date.Format("2006-01-02")
		if trendData[dateStr] == nil {
			trendData[dateStr] = make(map[string]int64)
		}
		trendData[dateStr][provider] = tokens
	}

	// Convert trend data to array format
	var trendArray []map[string]interface{}
	for date, byProvider := range trendData {
		trendArray = append(trendArray, map[string]interface{}{
			"date":        date,
			"by_provider": byProvider,
		})
	}

	// Get distribution by provider
	distQuery := fmt.Sprintf(`
		SELECT
			provider,
			COALESCE(SUM(input_tokens + output_tokens), 0) AS tokens,
			COUNT(*) AS requests
		FROM request_log
		WHERE %s AND provider IS NOT NULL AND provider != ''
		GROUP BY provider
		ORDER BY tokens DESC
	`, whereClause)

	rows2, err := r.db.Conn().Query(ctx, distQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get distribution: %w", err)
	}
	defer rows2.Close()

	var byProvider []map[string]interface{}
	for rows2.Next() {
		var provider string
		var tokens, requests int64
		if err := rows2.Scan(&provider, &tokens, &requests); err != nil {
			return nil, fmt.Errorf("failed to scan distribution: %w", err)
		}
		percentage := float64(0)
		if totalTokens > 0 {
			percentage = float64(tokens) * 100 / float64(totalTokens)
		}
		byProvider = append(byProvider, map[string]interface{}{
			"name":       provider,
			"tokens":     tokens,
			"requests":   requests,
			"percentage": percentage,
			"cost":       (float64(tokens) * 0.0015) / 1000,
		})
	}

	// Get distribution by model
	modelDistQuery := fmt.Sprintf(`
		SELECT
			model,
			COALESCE(SUM(input_tokens + output_tokens), 0) AS tokens,
			COUNT(*) AS requests
		FROM request_log
		WHERE %s AND model IS NOT NULL AND model != ''
		GROUP BY model
		ORDER BY tokens DESC
		LIMIT 10
	`, whereClause)

	rows3, err := r.db.Conn().Query(ctx, modelDistQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get model distribution: %w", err)
	}
	defer rows3.Close()

	var byModel []map[string]interface{}
	for rows3.Next() {
		var model string
		var tokens, requests int64
		if err := rows3.Scan(&model, &tokens, &requests); err != nil {
			return nil, fmt.Errorf("failed to scan model distribution: %w", err)
		}
		percentage := float64(0)
		if totalTokens > 0 {
			percentage = float64(tokens) * 100 / float64(totalTokens)
		}
		byModel = append(byModel, map[string]interface{}{
			"name":       model,
			"tokens":     tokens,
			"requests":   requests,
			"percentage": percentage,
			"cost":       (float64(tokens) * 0.0015) / 1000,
		})
	}

	// Get distribution by tool (platform)
	toolDistQuery := fmt.Sprintf(`
		SELECT
			platform,
			COALESCE(SUM(input_tokens + output_tokens), 0) AS tokens,
			COUNT(*) AS requests
		FROM request_log
		WHERE %s AND platform IS NOT NULL AND platform != ''
		GROUP BY platform
		ORDER BY tokens DESC
	`, whereClause)

	rows4, err := r.db.Conn().Query(ctx, toolDistQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get tool distribution: %w", err)
	}
	defer rows4.Close()

	var byTool []map[string]interface{}
	for rows4.Next() {
		var tool string
		var tokens, requests int64
		if err := rows4.Scan(&tool, &tokens, &requests); err != nil {
			return nil, fmt.Errorf("failed to scan tool distribution: %w", err)
		}
		percentage := float64(0)
		if totalTokens > 0 {
			percentage = float64(tokens) * 100 / float64(totalTokens)
		}
		byTool = append(byTool, map[string]interface{}{
			"name":       tool,
			"tokens":     tokens,
			"requests":   requests,
			"percentage": percentage,
			"cost":       (float64(tokens) * 0.0015) / 1000,
		})
	}

	return map[string]interface{}{
		"period": map[string]interface{}{
			"start": startDate,
			"end":   endDate,
		},
		"summary": map[string]interface{}{
			"total_tokens":   totalTokens,
			"total_input":    totalInput,
			"total_output":   totalOutput,
			"total_cost":     totalCost,
			"total_requests": totalRequests,
		},
		"trend_data": trendArray,
		"distribution": map[string]interface{}{
			"by_provider": byProvider,
			"by_model":    byModel,
			"by_tool":     byTool,
		},
	}, nil
}

// GetUserAnalytics returns user analytics with filtering
func (r *UsageRepository) GetUserAnalytics(ctx context.Context, tenantID int64, startDate, endDate string, userIDs []int64, providers, tools []string) (map[string]interface{}, error) {
	// Parse date range
	startTime, endTime, err := parseDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}
	endTime = endTime.Add(24 * time.Hour)

	whereClause := "r.tenant_id = $1 AND r.created_at >= $2 AND r.created_at < $3"
	args := []interface{}{tenantID, startTime, endTime}
	argIndex := 4

	if len(userIDs) > 0 {
		whereClause += fmt.Sprintf(" AND r.user_id = ANY($%d)", argIndex)
		args = append(args, userIDs)
		argIndex++
	}
	if len(providers) > 0 {
		whereClause += fmt.Sprintf(" AND r.provider = ANY($%d)", argIndex)
		args = append(args, providers)
		argIndex++
	}
	if len(tools) > 0 {
		whereClause += fmt.Sprintf(" AND r.platform = ANY($%d)", argIndex)
		args = append(args, tools)
	}

	// Get leaderboard
	leaderboardQuery := fmt.Sprintf(`
		SELECT
			u.id AS user_id,
			u.name,
			COALESCE(SUM(r.input_tokens + r.output_tokens), 0) AS total_tokens,
			COUNT(*) AS total_requests
		FROM users u
		LEFT JOIN request_log r ON u.id = r.user_id AND %s
		WHERE u.tenant_id = $1
		GROUP BY u.id, u.name
		ORDER BY total_tokens DESC
		LIMIT 20
	`, whereClause)

	rows, err := r.db.Conn().Query(ctx, leaderboardQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}
	defer rows.Close()

	var totalTokensAll int64
	var leaderboard []map[string]interface{}
	for rows.Next() {
		var userID int64
		var name string
		var totalTokens, totalRequests int64
		if err := rows.Scan(&userID, &name, &totalTokens, &totalRequests); err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard: %w", err)
		}
		totalTokensAll += totalTokens
		leaderboard = append(leaderboard, map[string]interface{}{
			"user_id":        userID,
			"name":           name,
			"total_tokens":   totalTokens,
			"total_requests": totalRequests,
			"total_cost":     (float64(totalTokens) * 0.0015) / 1000,
		})
	}

	// Calculate percentages and ranks
	for i, item := range leaderboard {
		tokens := item["total_tokens"].(int64)
		percentage := float64(0)
		if totalTokensAll > 0 {
			percentage = float64(tokens) * 100 / float64(totalTokensAll)
		}
		leaderboard[i]["rank"] = i + 1
		leaderboard[i]["percentage"] = percentage
	}

	// Get detailed stats
	detailsQuery := fmt.Sprintf(`
		SELECT
			r.user_id,
			u.name AS user_name,
			DATE(r.created_at) AS date,
			r.provider,
			r.model,
			COALESCE(SUM(r.input_tokens + r.output_tokens), 0) AS total_tokens,
			AVG(r.duration_sec) * 1000 AS avg_latency_ms
		FROM request_log r
		JOIN users u ON r.user_id = u.id
		WHERE %s
		GROUP BY r.user_id, u.name, DATE(r.created_at), r.provider, r.model
		ORDER BY date DESC, total_tokens DESC
		LIMIT 100
	`, whereClause)

	rows2, err := r.db.Conn().Query(ctx, detailsQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get details: %w", err)
	}
	defer rows2.Close()

	var details []map[string]interface{}
	for rows2.Next() {
		var userID int64
		var userName string
		var date time.Time
		var provider, model *string
		var totalTokens int64
		var avgLatency float64

		if err := rows2.Scan(&userID, &userName, &date, &provider, &model, &totalTokens, &avgLatency); err != nil {
			return nil, fmt.Errorf("failed to scan details: %w", err)
		}

		providerStr := ""
		if provider != nil {
			providerStr = *provider
		}
		modelStr := ""
		if model != nil {
			modelStr = *model
		}

		details = append(details, map[string]interface{}{
			"user_id":        userID,
			"user_name":      userName,
			"date":           date.Format("2006-01-02"),
			"provider":       providerStr,
			"model":          modelStr,
			"total_tokens":   totalTokens,
			"total_cost":     (float64(totalTokens) * 0.0015) / 1000,
			"avg_latency_ms": avgLatency,
		})
	}

	return map[string]interface{}{
		"period": map[string]interface{}{
			"start": startDate,
			"end":   endDate,
		},
		"leaderboard": leaderboard,
		"details":     details,
	}, nil
}

// GetHistory returns paginated request logs
func (r *UsageRepository) GetHistory(ctx context.Context, tenantID int64, startDate, endDate string, page, limit int, userIDs []int64, providers, models, tools []string, sortBy, sortOrder string) (map[string]interface{}, error) {
	// Parse date range
	startTime, endTime, err := parseDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}
	endTime = endTime.Add(24 * time.Hour)

	whereClause := "r.tenant_id = $1 AND r.created_at >= $2 AND r.created_at < $3"
	args := []interface{}{tenantID, startTime, endTime}
	argIndex := 4

	if len(userIDs) > 0 {
		whereClause += fmt.Sprintf(" AND r.user_id = ANY($%d)", argIndex)
		args = append(args, userIDs)
		argIndex++
	}
	if len(providers) > 0 {
		whereClause += fmt.Sprintf(" AND r.provider = ANY($%d)", argIndex)
		args = append(args, providers)
		argIndex++
	}
	if len(models) > 0 {
		whereClause += fmt.Sprintf(" AND r.model = ANY($%d)", argIndex)
		args = append(args, models)
		argIndex++
	}
	if len(tools) > 0 {
		whereClause += fmt.Sprintf(" AND r.platform = ANY($%d)", argIndex)
		args = append(args, tools)
	}

	// Validate sort column
	validSortColumns := map[string]bool{
		"created_at": true, "input_tokens": true, "output_tokens": true, "duration_sec": true,
	}
	if !validSortColumns[sortBy] {
		sortBy = "created_at"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Get total count
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM request_log r WHERE %s`, whereClause)
	var totalCount int64
	err = r.db.Conn().QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get count: %w", err)
	}

	totalPages := (totalCount + int64(limit) - 1) / int64(limit)
	offset := (page - 1) * limit

	// Get records
	recordsQuery := fmt.Sprintf(`
		SELECT
			r.id,
			r.created_at,
			r.user_id,
			u.name AS user_name,
			r.provider,
			r.model,
			r.platform,
			r.input_tokens,
			r.output_tokens,
			r.http_code,
			r.duration_sec,
			r.is_stream
		FROM request_log r
		JOIN users u ON r.user_id = u.id
		WHERE %s
		ORDER BY r.%s %s
		LIMIT %d OFFSET %d
	`, whereClause, sortBy, sortOrder, limit, offset)

	rows, err := r.db.Conn().Query(ctx, recordsQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get records: %w", err)
	}
	defer rows.Close()

	var records []map[string]interface{}
	for rows.Next() {
		var id, userID int64
		var createdAt time.Time
		var userName string
		var provider, model, platform *string
		var inputTokens, outputTokens, httpCode *int32
		var durationSec *float32
		var isStream *bool

		if err := rows.Scan(&id, &createdAt, &userID, &userName, &provider, &model, &platform, &inputTokens, &outputTokens, &httpCode, &durationSec, &isStream); err != nil {
			return nil, fmt.Errorf("failed to scan record: %w", err)
		}

		// Calculate cost using per-model pricing table
		modelName := stringOrEmpty(model)
		cost := pricing.EstimateCost(modelName, int64(intOrZero(inputTokens)), int64(intOrZero(outputTokens)))

		record := map[string]interface{}{
			"id":             id,
			"timestamp":      createdAt,
			"user_id":        userID,
			"user_name":      userName,
			"provider":       stringOrEmpty(provider),
			"model":          modelName,
			"platform":       stringOrEmpty(platform),
			"input_tokens":   intOrZero(inputTokens),
			"output_tokens":  intOrZero(outputTokens),
			"http_code":      intOrZero(httpCode),
			"duration_sec":   floatOrZero(durationSec),
			"is_stream":      boolOrFalse(isStream),
			"estimated_cost": cost,
		}
		records = append(records, record)
	}

	return map[string]interface{}{
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total_count": totalCount,
			"total_pages": totalPages,
		},
		"records": records,
	}, nil
}

// GetFilterOptions returns available filter options
func (r *UsageRepository) GetFilterOptions(ctx context.Context, tenantID int64) (map[string]interface{}, error) {
	// Get unique providers
	providersQuery := `
		SELECT DISTINCT provider FROM request_log
		WHERE tenant_id = $1 AND provider IS NOT NULL AND provider != ''
		ORDER BY provider
	`
	rows, err := r.db.Conn().Query(ctx, providersQuery, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}
	defer rows.Close()

	var providers []string
	for rows.Next() {
		var provider string
		if err := rows.Scan(&provider); err != nil {
			return nil, fmt.Errorf("failed to scan provider: %w", err)
		}
		providers = append(providers, provider)
	}

	// Get unique models
	modelsQuery := `
		SELECT DISTINCT model FROM request_log
		WHERE tenant_id = $1 AND model IS NOT NULL AND model != ''
		ORDER BY model
	`
	rows2, err := r.db.Conn().Query(ctx, modelsQuery, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %w", err)
	}
	defer rows2.Close()

	var modelsList []string
	for rows2.Next() {
		var model string
		if err := rows2.Scan(&model); err != nil {
			return nil, fmt.Errorf("failed to scan model: %w", err)
		}
		modelsList = append(modelsList, model)
	}

	// Get unique tools (platforms)
	toolsQuery := `
		SELECT DISTINCT platform FROM request_log
		WHERE tenant_id = $1 AND platform IS NOT NULL AND platform != ''
		ORDER BY platform
	`
	rows4, err := r.db.Conn().Query(ctx, toolsQuery, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tools: %w", err)
	}
	defer rows4.Close()

	var tools []string
	for rows4.Next() {
		var tool string
		if err := rows4.Scan(&tool); err != nil {
			return nil, fmt.Errorf("failed to scan tool: %w", err)
		}
		tools = append(tools, tool)
	}

	// Get users
	usersQuery := `
		SELECT id, name, email FROM users WHERE tenant_id = $1 ORDER BY name
	`
	rows3, err := r.db.Conn().Query(ctx, usersQuery, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows3.Close()

	var users []map[string]interface{}
	for rows3.Next() {
		var id int64
		var name, email string
		if err := rows3.Scan(&id, &name, &email); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, map[string]interface{}{
			"id":    id,
			"name":  name,
			"email": email,
		})
	}

	return map[string]interface{}{
		"providers": providers,
		"models":    modelsList,
		"tools":     tools,
		"users":     users,
	}, nil
}

// GetPersonalAnalytics returns personal analytics for a specific user with filtering
func (r *UsageRepository) GetPersonalAnalytics(ctx context.Context, userID, tenantID int64, startDate, endDate string, providers, models, tools []string) (map[string]interface{}, error) {
	startTime, endTime, err := parseDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}
	endTime = endTime.Add(24 * time.Hour)

	whereClause := "tenant_id = $1 AND user_id = $2 AND created_at >= $3 AND created_at < $4"
	args := []interface{}{tenantID, userID, startTime, endTime}
	argIndex := 5

	if len(providers) > 0 {
		whereClause += fmt.Sprintf(" AND provider = ANY($%d)", argIndex)
		args = append(args, providers)
		argIndex++
	}
	if len(models) > 0 {
		whereClause += fmt.Sprintf(" AND model = ANY($%d)", argIndex)
		args = append(args, models)
		argIndex++
	}
	if len(tools) > 0 {
		whereClause += fmt.Sprintf(" AND platform = ANY($%d)", argIndex)
		args = append(args, tools)
	}

	// Get summary
	summaryQuery := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(input_tokens + output_tokens), 0) AS total_tokens,
			COALESCE(SUM(input_tokens), 0) AS total_input,
			COALESCE(SUM(output_tokens), 0) AS total_output,
			COUNT(*) AS total_requests
		FROM request_log
		WHERE %s
	`, whereClause)

	var totalTokens, totalInput, totalOutput, totalRequests int64
	err = r.db.Conn().QueryRow(ctx, summaryQuery, args...).Scan(&totalTokens, &totalInput, &totalOutput, &totalRequests)
	if err != nil {
		return nil, fmt.Errorf("failed to get summary: %w", err)
	}

	totalCost := (float64(totalInput)*0.001 + float64(totalOutput)*0.002) / 1000

	// Get success rate
	successQuery := fmt.Sprintf(`
		SELECT
			COALESCE(
				COUNT(*) * 100.0 / NULLIF((SELECT COUNT(*) FROM request_log WHERE %s), 0),
				0.0
			)
		FROM request_log
		WHERE %s AND http_code >= 200 AND http_code < 300
	`, whereClause, whereClause)

	var successRate float64
	err = r.db.Conn().QueryRow(ctx, successQuery, append(args, args...)...).Scan(&successRate)
	if err != nil {
		successRate = 0
	}

	// Get trend data by day
	trendQuery := fmt.Sprintf(`
		SELECT
			DATE(created_at) AS date,
			COALESCE(SUM(input_tokens + output_tokens), 0) AS tokens,
			COALESCE(SUM(input_tokens), 0) AS input_tokens,
			COALESCE(SUM(output_tokens), 0) AS output_tokens,
			COUNT(*) AS requests
		FROM request_log
		WHERE %s
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`, whereClause)

	rows, err := r.db.Conn().Query(ctx, trendQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get trend data: %w", err)
	}
	defer rows.Close()

	var trendArray []map[string]interface{}
	for rows.Next() {
		var date time.Time
		var tokens, inputTokens, outputTokens, requests int64
		if err := rows.Scan(&date, &tokens, &inputTokens, &outputTokens, &requests); err != nil {
			return nil, fmt.Errorf("failed to scan trend data: %w", err)
		}
		trendArray = append(trendArray, map[string]interface{}{
			"date":          date.Format("2006-01-02"),
			"tokens":        tokens,
			"input_tokens":  inputTokens,
			"output_tokens": outputTokens,
			"requests":      requests,
		})
	}

	// Get distribution by provider
	distQuery := fmt.Sprintf(`
		SELECT
			provider,
			COALESCE(SUM(input_tokens + output_tokens), 0) AS tokens,
			COUNT(*) AS requests
		FROM request_log
		WHERE %s AND provider IS NOT NULL AND provider != ''
		GROUP BY provider
		ORDER BY tokens DESC
	`, whereClause)

	rows2, err := r.db.Conn().Query(ctx, distQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get distribution: %w", err)
	}
	defer rows2.Close()

	var byProvider []map[string]interface{}
	for rows2.Next() {
		var provider string
		var tokens, requests int64
		if err := rows2.Scan(&provider, &tokens, &requests); err != nil {
			return nil, fmt.Errorf("failed to scan distribution: %w", err)
		}
		percentage := float64(0)
		if totalTokens > 0 {
			percentage = float64(tokens) * 100 / float64(totalTokens)
		}
		byProvider = append(byProvider, map[string]interface{}{
			"name":       provider,
			"tokens":     tokens,
			"requests":   requests,
			"percentage": percentage,
			"cost":       (float64(tokens) * 0.0015) / 1000,
		})
	}

	// Get distribution by model
	modelDistQuery := fmt.Sprintf(`
		SELECT
			model,
			COALESCE(SUM(input_tokens + output_tokens), 0) AS tokens,
			COUNT(*) AS requests
		FROM request_log
		WHERE %s AND model IS NOT NULL AND model != ''
		GROUP BY model
		ORDER BY tokens DESC
		LIMIT 10
	`, whereClause)

	rows3, err := r.db.Conn().Query(ctx, modelDistQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get model distribution: %w", err)
	}
	defer rows3.Close()

	var byModel []map[string]interface{}
	for rows3.Next() {
		var model string
		var tokens, requests int64
		if err := rows3.Scan(&model, &tokens, &requests); err != nil {
			return nil, fmt.Errorf("failed to scan model distribution: %w", err)
		}
		percentage := float64(0)
		if totalTokens > 0 {
			percentage = float64(tokens) * 100 / float64(totalTokens)
		}
		byModel = append(byModel, map[string]interface{}{
			"name":       model,
			"tokens":     tokens,
			"requests":   requests,
			"percentage": percentage,
			"cost":       (float64(tokens) * 0.0015) / 1000,
		})
	}

	return map[string]interface{}{
		"period": map[string]interface{}{
			"start": startDate,
			"end":   endDate,
		},
		"summary": map[string]interface{}{
			"total_tokens":   totalTokens,
			"total_input":    totalInput,
			"total_output":   totalOutput,
			"total_cost":     totalCost,
			"total_requests": totalRequests,
			"success_rate":   successRate,
		},
		"trend_data": trendArray,
		"distribution": map[string]interface{}{
			"by_provider": byProvider,
			"by_model":    byModel,
		},
	}, nil
}

// GetPersonalHistory returns paginated request logs for a specific user
func (r *UsageRepository) GetPersonalHistory(ctx context.Context, userID, tenantID int64, startDate, endDate string, page, limit int, providers, models, tools []string, sortBy, sortOrder string) (map[string]interface{}, error) {
	startTime, endTime, err := parseDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}
	endTime = endTime.Add(24 * time.Hour)

	whereClause := "r.tenant_id = $1 AND r.user_id = $2 AND r.created_at >= $3 AND r.created_at < $4"
	args := []interface{}{tenantID, userID, startTime, endTime}
	argIndex := 5

	if len(providers) > 0 {
		whereClause += fmt.Sprintf(" AND r.provider = ANY($%d)", argIndex)
		args = append(args, providers)
		argIndex++
	}
	if len(models) > 0 {
		whereClause += fmt.Sprintf(" AND r.model = ANY($%d)", argIndex)
		args = append(args, models)
		argIndex++
	}
	if len(tools) > 0 {
		whereClause += fmt.Sprintf(" AND r.platform = ANY($%d)", argIndex)
		args = append(args, tools)
	}

	validSortColumns := map[string]bool{
		"created_at": true, "input_tokens": true, "output_tokens": true, "duration_sec": true,
	}
	if !validSortColumns[sortBy] {
		sortBy = "created_at"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Get total count
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM request_log r WHERE %s`, whereClause)
	var totalCount int64
	err = r.db.Conn().QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get count: %w", err)
	}

	totalPages := (totalCount + int64(limit) - 1) / int64(limit)
	offset := (page - 1) * limit

	recordsQuery := fmt.Sprintf(`
		SELECT
			r.id,
			r.created_at,
			r.provider,
			r.model,
			r.platform,
			r.input_tokens,
			r.output_tokens,
			r.http_code,
			r.duration_sec,
			r.is_stream
		FROM request_log r
		WHERE %s
		ORDER BY r.%s %s
		LIMIT %d OFFSET %d
	`, whereClause, sortBy, sortOrder, limit, offset)

	rows, err := r.db.Conn().Query(ctx, recordsQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get records: %w", err)
	}
	defer rows.Close()

	var records []map[string]interface{}
	for rows.Next() {
		var id int64
		var createdAt time.Time
		var provider, model, platform *string
		var inputTokens, outputTokens, httpCode *int32
		var durationSec *float32
		var isStream *bool

		if err := rows.Scan(&id, &createdAt, &provider, &model, &platform, &inputTokens, &outputTokens, &httpCode, &durationSec, &isStream); err != nil {
			return nil, fmt.Errorf("failed to scan record: %w", err)
		}

		modelName := stringOrEmpty(model)
		cost := pricing.EstimateCost(modelName, int64(intOrZero(inputTokens)), int64(intOrZero(outputTokens)))

		record := map[string]interface{}{
			"id":            id,
			"timestamp":     createdAt,
			"provider":      stringOrEmpty(provider),
			"model":         modelName,
			"platform":      stringOrEmpty(platform),
			"input_tokens":  intOrZero(inputTokens),
			"output_tokens": intOrZero(outputTokens),
			"http_code":     intOrZero(httpCode),
			"duration_sec":  floatOrZero(durationSec),
			"is_stream":     boolOrFalse(isStream),
			"cost":          cost,
		}
		records = append(records, record)
	}

	return map[string]interface{}{
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total_count": totalCount,
			"total_pages": totalPages,
		},
		"records": records,
	}, nil
}

// GetPersonalFilterOptions returns filter options scoped to a specific user
func (r *UsageRepository) GetPersonalFilterOptions(ctx context.Context, userID, tenantID int64) (map[string]interface{}, error) {
	// Get unique providers for user
	providersQuery := `
		SELECT DISTINCT provider FROM request_log
		WHERE tenant_id = $1 AND user_id = $2 AND provider IS NOT NULL AND provider != ''
		ORDER BY provider
	`
	rows, err := r.db.Conn().Query(ctx, providersQuery, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}
	defer rows.Close()

	var providers []string
	for rows.Next() {
		var provider string
		if err := rows.Scan(&provider); err != nil {
			return nil, fmt.Errorf("failed to scan provider: %w", err)
		}
		providers = append(providers, provider)
	}

	// Get unique models for user
	modelsQuery := `
		SELECT DISTINCT model FROM request_log
		WHERE tenant_id = $1 AND user_id = $2 AND model IS NOT NULL AND model != ''
		ORDER BY model
	`
	rows2, err := r.db.Conn().Query(ctx, modelsQuery, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %w", err)
	}
	defer rows2.Close()

	var modelsList []string
	for rows2.Next() {
		var model string
		if err := rows2.Scan(&model); err != nil {
			return nil, fmt.Errorf("failed to scan model: %w", err)
		}
		modelsList = append(modelsList, model)
	}

	// Get unique tools (platforms) for user
	toolsQuery := `
		SELECT DISTINCT platform FROM request_log
		WHERE tenant_id = $1 AND user_id = $2 AND platform IS NOT NULL AND platform != ''
		ORDER BY platform
	`
	rows3, err := r.db.Conn().Query(ctx, toolsQuery, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tools: %w", err)
	}
	defer rows3.Close()

	var tools []string
	for rows3.Next() {
		var tool string
		if err := rows3.Scan(&tool); err != nil {
			return nil, fmt.Errorf("failed to scan tool: %w", err)
		}
		tools = append(tools, tool)
	}

	return map[string]interface{}{
		"providers": providers,
		"models":    modelsList,
		"tools":     tools,
	}, nil
}

// Helper functions
func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func intOrZero(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func floatOrZero(f *float32) float32 {
	if f == nil {
		return 0
	}
	return *f
}

func boolOrFalse(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func float64OrZero(f *float64) float64 {
	if f == nil {
		return 0.0
	}
	return *f
}
