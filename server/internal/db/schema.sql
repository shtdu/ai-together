-- Create tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(255) UNIQUE,
    license_id VARCHAR(255),
    license_tier VARCHAR(10),
    license_seats INTEGER,
    license_key TEXT,
    license_issued_at TIMESTAMP,
    license_expires_at TIMESTAMP,
    license_type VARCHAR(20),
    license_max_seats INTEGER,
    license_max_teams INTEGER,
    license_data_retention_days INTEGER DEFAULT 7,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member', -- 'manager' or 'member'
    tenant_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

-- Create teams table
CREATE TABLE IF NOT EXISTS teams (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES users(id),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

-- Create team_members table
CREATE TABLE IF NOT EXISTS team_members (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member', -- 'manager' or 'member'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, user_id),
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create providers table
CREATE TABLE IF NOT EXISTS providers (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    api_url VARCHAR(255) NOT NULL,
    api_key VARCHAR(255) NOT NULL,
    team_id BIGINT NOT NULL,
    kind VARCHAR(50) NOT NULL DEFAULT 'codex',
    enabled BOOLEAN DEFAULT FALSE,
    model_mapping JSONB DEFAULT '{}',
    supported_models JSONB DEFAULT '[]',
    level INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
);

-- Create request_log table with multi-tenant support
CREATE TABLE IF NOT EXISTS request_log (
    id BIGSERIAL PRIMARY KEY,
    platform VARCHAR(100),
    model VARCHAR(255),
    provider VARCHAR(255),
    http_code INTEGER,
    input_tokens INTEGER,
    output_tokens INTEGER,
    cache_create_tokens INTEGER,
    cache_read_tokens INTEGER,
    reasoning_tokens INTEGER,
    is_stream BOOLEAN DEFAULT FALSE,
    duration_sec REAL DEFAULT 0,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Create team_usage_summary table
CREATE TABLE IF NOT EXISTS team_usage_summary (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    total_input INTEGER DEFAULT 0,
    total_output INTEGER DEFAULT 0,
    total_cost DECIMAL(10, 4) DEFAULT 0.0000,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, period_start, period_end),
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
);

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

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_tenants_license_id ON tenants(license_id);
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_teams_tenant_id ON teams(tenant_id);
CREATE INDEX IF NOT EXISTS idx_teams_owner_id ON teams(owner_id);
CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members(team_id);
CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members(user_id);
CREATE INDEX IF NOT EXISTS idx_providers_team_id ON providers(team_id);
CREATE INDEX IF NOT EXISTS idx_providers_kind ON providers(kind);
CREATE INDEX IF NOT EXISTS idx_request_log_tenant_id ON request_log(tenant_id);
CREATE INDEX IF NOT EXISTS idx_request_log_user_id ON request_log(user_id);
CREATE INDEX IF NOT EXISTS idx_request_log_created_at ON request_log(created_at);
CREATE INDEX IF NOT EXISTS idx_team_usage_summary_team_id ON team_usage_summary(team_id);
CREATE INDEX IF NOT EXISTS idx_team_usage_summary_period ON team_usage_summary(period_start, period_end);