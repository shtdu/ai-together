-- name: CreateProvider :one
INSERT INTO providers (name, api_url, api_key, team_id, enabled, kind, model_mapping, supported_models, level)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, name, api_url, api_key, team_id, kind, enabled, model_mapping, supported_models, level, created_at, updated_at;

-- name: GetProviderByID :one
SELECT id, name, api_url, api_key, team_id, kind, enabled, model_mapping, supported_models, level, created_at, updated_at
FROM providers
WHERE id = $1
LIMIT 1;

-- name: UpdateProvider :exec
UPDATE providers
SET name = $2, api_url = $3, api_key = $4, enabled = $5, kind = $6, model_mapping = $7, supported_models = $8, level = $9, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteProvider :exec
DELETE FROM providers WHERE id = $1;

-- name: GetProvidersByTeamID :many
SELECT id, name, api_url, api_key, team_id, kind, enabled, model_mapping, supported_models, level, created_at, updated_at
FROM providers
WHERE team_id = $1;

-- name: EnableProvider :exec
UPDATE providers SET enabled = true WHERE id = $1;

-- name: DisableProvider :exec
UPDATE providers SET enabled = false WHERE id = $1;