-- name: CreateTeam :one
INSERT INTO teams (name, description, owner_id, tenant_id, settings)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, description, owner_id, tenant_id, settings, created_at, updated_at;

-- name: GetTeamByID :one
SELECT id, name, description, owner_id, tenant_id, settings, created_at, updated_at
FROM teams
WHERE id = $1
LIMIT 1;

-- name: GetTeamsByUserID :many
SELECT t.id, t.name, t.description, t.owner_id, t.tenant_id, t.settings
FROM teams t
INNER JOIN team_members tm ON t.id = tm.team_id
WHERE tm.user_id = $1
ORDER BY t.id DESC;

-- name: UpdateTeam :exec
UPDATE teams
SET name = $2, description = $3, settings = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteTeam :exec
DELETE FROM teams WHERE id = $1;

-- name: AddTeamMember :exec
INSERT INTO team_members (team_id, user_id, role)
VALUES ($1, $2, $3);

-- name: RemoveTeamMember :exec
DELETE FROM team_members
WHERE team_id = $1 AND user_id = $2;

-- name: GetTeamMembers :many
SELECT u.id, u.email, u.name, u.role, u.tenant_id
FROM users u
INNER JOIN team_members tm ON u.id = tm.user_id
WHERE tm.team_id = $1
ORDER BY u.id DESC;