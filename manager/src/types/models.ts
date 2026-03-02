export interface User {
  id: number
  email: string
  name: string
  role: 'manager' | 'member'
  tenant_id: number
  created_at: string
  updated_at?: string
}

export interface Team {
  id: number
  name: string
  description?: string
  owner_id: number
  tenant_id: number
  created_at: string
  updated_at?: string
}

export interface Provider {
  id: number
  name: string
  api_url: string
  api_key?: string
  team_id: number
  kind: 'claude' | 'codex' | 'opencode'
  enabled: boolean
  level: number
  created_at: string
  updated_at?: string
}

export interface UsageRecord {
  id: number
  platform: string
  model: string
  provider: string
  http_code: number
  input_tokens: number
  output_tokens: number
  cache_create_tokens: number
  cache_read_tokens: number
  reasoning_tokens: number
  is_stream: boolean
  duration_sec: number
  tenant_id: number
  user_id: number
  created_at: string
}
