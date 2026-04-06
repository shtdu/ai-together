-- Create relay_tokens table for per-member relay authentication
CREATE TABLE IF NOT EXISTS relay_tokens (
    id BIGSERIAL PRIMARY KEY,
    token_hash VARCHAR(255) NOT NULL,
    user_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL,
    token_prefix VARCHAR(8) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP,
    revoked_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- One active token per user
CREATE UNIQUE INDEX IF NOT EXISTS idx_relay_tokens_user_active ON relay_tokens (user_id) WHERE revoked_at IS NULL;
-- Fast lookup by hash
CREATE INDEX IF NOT EXISTS idx_relay_tokens_hash ON relay_tokens (token_hash) WHERE revoked_at IS NULL;
-- List by tenant
CREATE INDEX IF NOT EXISTS idx_relay_tokens_tenant_id ON relay_tokens (tenant_id);
