-- name: CreateUsageRecord :exec
INSERT INTO request_log (platform, model, provider, http_code, input_tokens, output_tokens, cache_create_tokens, cache_read_tokens, reasoning_tokens, is_stream, duration_sec, tenant_id, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);

-- name: GetUsageByTeamIDAndPeriod :many
SELECT id, platform, model, provider, http_code, input_tokens, output_tokens, cache_create_tokens, cache_read_tokens, reasoning_tokens, is_stream, duration_sec, tenant_id, user_id, created_at
FROM request_log
WHERE tenant_id = $1
  AND created_at >= $2
  AND created_at <= $3
ORDER BY created_at DESC;

-- name: GetUsageByUserIDAndPeriod :many
SELECT id, platform, model, provider, http_code, input_tokens, output_tokens, cache_create_tokens, cache_read_tokens, reasoning_tokens, is_stream, duration_sec, tenant_id, user_id, created_at
FROM request_log
WHERE user_id = $1
  AND created_at >= $2
  AND created_at <= $3
ORDER BY created_at DESC;

-- name: GetTeamUsageSummary :many
SELECT team_id, period_start, period_end, total_input, total_output, total_cost, created_at
FROM team_usage_summary
WHERE team_id = $1
  AND period_start >= $2
  AND period_end <= $3
ORDER BY period_start DESC;

-- name: CreateTeamUsageSummary :exec
INSERT INTO team_usage_summary (team_id, period_start, period_end, total_input, total_output, total_cost)
VALUES ($1, $2, $3, $4, $5, $6);