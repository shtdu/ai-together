-- name: GetTenantLicense :one
SELECT license_tier, license_seats, license_key, license_issued_at, license_expires_at, license_id, name,
       license_type, license_max_seats, license_max_teams, license_data_retention_days
FROM tenants WHERE id = $1;

-- name: UpdateTenantLicense :exec
UPDATE tenants SET
    license_id = $2,
    license_tier = $3,
    license_seats = $4,
    license_key = $5,
    license_issued_at = $6,
    license_expires_at = $7,
    license_type = $8,
    license_max_seats = $9,
    license_max_teams = $10,
    license_data_retention_days = $11,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: CountTenantUsers :one
SELECT COUNT(*) FROM users WHERE tenant_id = $1;

-- name: CountTenantProvidersByKind :one
SELECT COUNT(*) FROM providers p
JOIN teams t ON p.team_id = t.id
WHERE t.tenant_id = $1 AND p.kind = $2;

-- name: GetProviderCountsByKind :many
SELECT p.kind, COUNT(*) as count FROM providers p
JOIN teams t ON p.team_id = t.id
WHERE t.tenant_id = $1
GROUP BY p.kind;

-- name: GetAllLicenses :many
SELECT id, license_tier, license_seats, license_key, license_issued_at, license_expires_at, license_id, name,
       license_type, license_max_seats, license_max_teams, license_data_retention_days
FROM tenants;

-- name: CountTenantTeams :one
SELECT COUNT(*) FROM teams WHERE tenant_id = $1;
