-- Add kind column to providers table
ALTER TABLE providers ADD COLUMN IF NOT EXISTS kind VARCHAR(50) NOT NULL DEFAULT 'codex';

-- Add index for kind column
CREATE INDEX IF NOT EXISTS idx_providers_kind ON providers(kind);
