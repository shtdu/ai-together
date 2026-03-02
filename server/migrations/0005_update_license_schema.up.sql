-- Add new license type columns for 2-type license system
ALTER TABLE tenants ADD COLUMN license_type VARCHAR(20);
ALTER TABLE tenants ADD COLUMN license_max_seats INTEGER;
ALTER TABLE tenants ADD COLUMN license_max_teams INTEGER;
ALTER TABLE tenants ADD COLUMN license_data_retention_days INTEGER DEFAULT 7;

-- Migrate existing tier data to new type system
-- Tier 0.0 -> opensource, Tiers 1.0, 2.0, 3.0 -> commercial
UPDATE tenants SET license_type = CASE
    WHEN license_tier = '0.0' THEN 'opensource'
    ELSE 'commercial'
END WHERE license_tier IS NOT NULL;

-- Map seat limits (soft limits - informational only)
UPDATE tenants SET license_max_seats = CASE
    WHEN license_tier = '0.0' THEN 3      -- Community: 3 seats
    WHEN license_tier = '1.0' THEN 10     -- Standard: 10 seats
    WHEN license_tier = '2.0' THEN 25     -- Professional: 25 seats
    WHEN license_tier = '3.0' THEN 100    -- Enterprise: 100 seats
    ELSE 3                                -- Default: 3 seats
END WHERE license_tier IS NOT NULL;

-- Map team limits
UPDATE tenants SET license_max_teams = CASE
    WHEN license_tier = '0.0' THEN 1      -- Opensource: 1 team
    ELSE -1                               -- Commercial: unlimited
END WHERE license_tier IS NOT NULL;

-- Map data retention days
UPDATE tenants SET license_data_retention_days = CASE
    WHEN license_tier = '0.0' THEN 7      -- Opensource: 7 days
    ELSE 90                               -- Commercial: 90 days
END WHERE license_tier IS NOT NULL;

-- Old columns (license_tier, license_seats) will be deprecated but kept for backward compatibility
