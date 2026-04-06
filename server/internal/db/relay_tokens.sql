-- name: CreateRelayToken :one
INSERT INTO relay_tokens (token_hash, user_id, tenant_id, token_prefix)
VALUES ($1, $2, $3, $4)
RETURNING id, token_hash, user_id, tenant_id, token_prefix, created_at, last_used_at, revoked_at;

-- name: GetRelayTokenByHash :one
SELECT id, token_hash, user_id, tenant_id, token_prefix, created_at, last_used_at, revoked_at
FROM relay_tokens
WHERE token_hash = $1 AND revoked_at IS NULL;

-- name: GetRelayTokenByUserID :one
SELECT id, token_hash, user_id, tenant_id, token_prefix, created_at, last_used_at, revoked_at
FROM relay_tokens
WHERE user_id = $1 AND revoked_at IS NULL;

-- name: RevokeRelayTokenByUserID :exec
UPDATE relay_tokens
SET revoked_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND revoked_at IS NULL;

-- name: UpdateRelayTokenLastUsed :exec
UPDATE relay_tokens
SET last_used_at = CURRENT_TIMESTAMP
WHERE id = $1;
