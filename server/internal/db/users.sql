-- name: CreateUser :exec
INSERT INTO users (email, name, password, role, tenant_id)
VALUES ($1, $2, $3, $4, $5);

-- name: GetUserByEmail :one
SELECT id, email, name, password, role, tenant_id, created_at, updated_at
FROM users
WHERE email = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT id, email, name, password, role, tenant_id, created_at, updated_at
FROM users
WHERE id = $1
LIMIT 1;

-- name: UpdateUser :exec
UPDATE users
SET name = $2, role = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;