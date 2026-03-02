-- Remove index
DROP INDEX IF EXISTS idx_providers_kind;

-- Remove kind column from providers table
ALTER TABLE providers DROP COLUMN IF EXISTS kind;
