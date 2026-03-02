-- Add license fields to tenants table
ALTER TABLE tenants ADD COLUMN license_id VARCHAR(255);
ALTER TABLE tenants ADD COLUMN license_tier VARCHAR(10);
ALTER TABLE tenants ADD COLUMN license_seats INTEGER;
ALTER TABLE tenants ADD COLUMN license_key TEXT;
ALTER TABLE tenants ADD COLUMN license_issued_at TIMESTAMP;
ALTER TABLE tenants ADD COLUMN license_expires_at TIMESTAMP;

-- Create index for license lookups
CREATE INDEX idx_tenants_license_id ON tenants(license_id);
