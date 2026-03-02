-- Recreate provider_templates table (for rollback)
CREATE TABLE IF NOT EXISTS provider_templates (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    api_url VARCHAR(255) NOT NULL,
    model_mapping JSONB NOT NULL,
    team_id BIGINT NOT NULL,
    level INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
);
