-- Rollback license schema changes
ALTER TABLE tenants DROP COLUMN IF EXISTS license_type;
ALTER TABLE tenants DROP COLUMN IF EXISTS license_max_seats;
ALTER TABLE tenants DROP COLUMN IF EXISTS license_max_teams;
ALTER TABLE tenants DROP COLUMN IF EXISTS license_data_retention_days;
