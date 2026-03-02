-- Rollback: Remove license fields from tenants table
DROP INDEX IF EXISTS idx_tenants_license_id;
ALTER TABLE tenants DROP COLUMN IF EXISTS license_expires_at;
ALTER TABLE tenants DROP COLUMN IF EXISTS license_issued_at;
ALTER TABLE tenants DROP COLUMN IF EXISTS license_key;
ALTER TABLE tenants DROP COLUMN IF EXISTS license_seats;
ALTER TABLE tenants DROP COLUMN IF EXISTS license_tier;
ALTER TABLE tenants DROP COLUMN IF EXISTS license_id;
