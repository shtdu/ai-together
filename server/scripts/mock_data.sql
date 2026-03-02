-- Mock Data for Testing Configuration Sync
-- This script creates test data to demonstrate the server-to-member config sync
--
-- Usage:
--   PGPASSWORD=your_password psql -h localhost -U your_username -d code_together -f scripts/mock_data.sql
--
-- Or use the environment variable from .env:
--   PGPASSWORD=pgsql psql -h localhost -U admin -d codetogether -f scripts/mock_data.sql

-- Clear existing test data (optional - comment out if you want to keep existing data)
-- DELETE FROM providers WHERE team_id IN (SELECT id FROM teams WHERE name LIKE 'Test Team%');
-- DELETE FROM team_members WHERE team_id IN (SELECT id FROM teams WHERE name LIKE 'Test Team%');
-- DELETE FROM teams WHERE name LIKE 'Test Team%';
-- DELETE FROM users WHERE email LIKE 'test%@example.com';
-- DELETE FROM tenants WHERE name LIKE 'Test Tenant%';

-- Insert test tenant
INSERT INTO tenants (name, subdomain) VALUES
('Test Organization', 'test-org')
ON CONFLICT (subdomain) DO NOTHING;

-- Get the tenant ID (we'll use this in subsequent inserts)
-- Assuming this is the first or only tenant, ID should be 1, but let's use a CTE for safety

-- Insert test users (password is 'password123' for both users)
INSERT INTO users (email, name, password, role, tenant_id) VALUES
('test-member@example.com', 'Test Member', '$2a$10$6y2mLB13m8xqd7j47xWi9.5ZnNbctm9bgIVVDt1WWXNsN6V3u1Xle', 'member', 1),
('test-admin@example.com', 'Test Admin', '$2a$10$6y2mLB13m8xqd7j47xWi9.5ZnNbctm9bgIVVDt1WWXNsN6V3u1Xle', 'manager', 1)
ON CONFLICT (email) DO NOTHING;

-- Insert test team
INSERT INTO teams (name, description, owner_id, tenant_id) VALUES
('Test Team', 'Team for testing configuration sync', 2, 1)
ON CONFLICT DO NOTHING;

-- Add users to the test team
INSERT INTO team_members (team_id, user_id, role) VALUES
(1, 1, 'member'),  -- test-member as member
(1, 2, 'manager')  -- test-admin as manager
ON CONFLICT (team_id, user_id) DO NOTHING;

-- Insert test providers with different kinds
INSERT INTO providers (name, api_url, api_key, team_id, enabled, kind, level, model_mapping, supported_models) VALUES
-- Claude providers with model mapping
('Qiniu Claude', 'https://api.qnaigc.com', 'sk-e8a42688946d68de9f69d3bc4a0be543989d09d8d13b203b902de1a4feef7ad0', 1, true, 'claude', 2,
 '{"claude-haiku-4-*": "glm-4.5-air", "claude-sonnet-4-*": "z-ai/glm-4.7"}'::jsonb,
 '["glm-4.5-air", "z-ai/glm-4.7"]'::jsonb),
('Zhipu AI Claude', 'https://open.bigmodel.cn/api/anthropic', '1ac3b6c0bc6842188e4ff7e2a24fdf84.n1gbblIDojJHy0UX', 1, true, 'claude', 1,
 '{}'::jsonb,
 '[]'::jsonb),

-- Codex providers with model mapping
('GitHub Copilot', 'https://api.githubcopilot.com', 'ghu-test-key-abcdef', 1, true, 'codex', 1,
 '{}'::jsonb,
 '[]'::jsonb),
('Azure OpenAI', 'https://azure-openai.openai.azure.com', 'sk-azure-test-key-xyz', 1, true, 'codex', 2,
 '{}'::jsonb,
 '[]'::jsonb),

-- OpenCode providers
('Zhipu AI', 'https://open.bigmodel.cn/api/coding/paas/v4', '1ac3b6c0bc6842188e4ff7e2a24fdf84.n1gbblIDojJHy0UX', 1, true, 'opencode', 1,
 '{}'::jsonb,
 '[]'::jsonb),

-- Disabled provider (should not sync)
('Disabled Provider', 'https://api.example.com', 'test-key-disabled', 1, false, 'claude', 3,
 '{}'::jsonb,
 '[]'::jsonb)

ON CONFLICT DO NOTHING;

-- Verify the data
SELECT 'Tenants:' as info;
SELECT id, name, subdomain FROM tenants;

SELECT 'Users:' as info;
SELECT id, email, name, role, tenant_id FROM users;

SELECT 'Teams:' as info;
SELECT id, name, owner_id, tenant_id FROM teams;

SELECT 'Team Members:' as info;
SELECT tm.team_id, t.name as team_name, tm.user_id, u.email, tm.role
FROM team_members tm
JOIN teams t ON tm.team_id = t.id
JOIN users u ON tm.user_id = u.id;

SELECT 'Providers:' as info;
SELECT id, name, api_url, kind, enabled, level, team_id
FROM providers
ORDER BY kind, level;

-- Summary
SELECT 'Summary:' as info;
SELECT
    (SELECT COUNT(*) FROM tenants) as tenants,
    (SELECT COUNT(*) FROM users) as users,
    (SELECT COUNT(*) FROM teams) as teams,
    (SELECT COUNT(*) FROM team_members) as team_members,
    (SELECT COUNT(*) FROM providers) as total_providers,
    (SELECT COUNT(*) FROM providers WHERE enabled = true) as enabled_providers,
    (SELECT COUNT(*) FROM providers WHERE kind = 'claude') as claude_providers,
    (SELECT COUNT(*) FROM providers WHERE kind = 'codex') as codex_providers;
