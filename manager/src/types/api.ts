import type { User, Team } from './models'

// Auth types
export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  user: User
}

export interface RefreshRequest {
  refresh_token: string
}

export interface RefreshResponse {
  access_token: string
  refresh_token: string
}

export interface UserProfile {
  user: User
  teams: Team[]
}

// Dashboard types
export interface DashboardMetricsRequest {
  range: '24h' | '7d' | '30d'
  interval: 'hour' | 'day'
}

export interface MetricDataPoint {
  timestamp: string
  active_time_seconds: number
  total_tokens: number
  input_tokens: number
  output_tokens: number
  request_count: number
}

export interface DashboardMetricsResponse {
  period: {
    start: string
    end: string
  }
  interval: string
  data_points: MetricDataPoint[]
  summary: {
    total_active_time_hours: number
    total_tokens: number
    total_requests: number
    estimated_cost: number
  }
}

export interface ProviderRanking {
  rank: number
  provider: string
  total_tokens: number
  percentage: number
  request_count: number
}

export interface DashboardRankingsResponse {
  period: string
  rankings: ProviderRanking[]
}

export interface MemberStats {
  user_id: number
  name: string
  email: string
  active_time_hours: number
  avg_tokens_per_day: number
  total_tokens: number
  last_active: string
}

export interface DashboardMembersResponse {
  members: MemberStats[]
}

// Analytics types
export interface AnalyticsFilters {
  start_date: string
  end_date: string
  providers?: string[]
  models?: string[]
  user_ids?: number[]
}

export interface ProviderTrendData {
  date: string
  by_provider: Record<string, { tokens: number; cost: number }>
  by_model: Record<string, { tokens: number; cost: number }>
}

export interface DistributionItem {
  name: string
  tokens: number
  percentage: number
  cost: number
}

export interface ProviderDetailRow {
  date: string
  provider: string
  model: string
  request_count: number
  input_tokens: number
  output_tokens: number
  cost: number
}

export interface ProviderAnalyticsResponse {
  period: {
    start: string
    end: string
  }
  summary: {
    total_tokens: number
    total_input: number
    total_output: number
    total_cost: number
    total_requests: number
  }
  trend_data: ProviderTrendData[] | null
  distribution: {
    by_provider: DistributionItem[] | null
    by_model: DistributionItem[] | null
  }
  details?: ProviderDetailRow[]
}

export interface UserLeaderboardItem {
  rank: number
  user_id: number
  name: string
  total_tokens: number
  total_cost: number
  percentage: number
}

export interface UserDetailRow {
  user_id: number
  user_name: string
  date: string
  provider: string
  model: string
  total_tokens: number
  total_cost: number
  avg_latency_ms: number
}

export interface UserAnalyticsResponse {
  period: {
    start: string
    end: string
  }
  leaderboard: UserLeaderboardItem[]
  details: UserDetailRow[]
}

export interface HistoryRecord {
  id: number
  timestamp: string
  user_id: number
  user_name: string
  provider: string
  model: string
  platform: string
  input_tokens: number
  output_tokens: number
  http_code: number
  duration_sec: number
  is_stream: boolean
}

export interface PaginatedHistoryResponse {
  pagination: {
    page: number
    limit: number
    total_count: number
    total_pages: number
  }
  records: HistoryRecord[]
}

// User management types
export interface CreateUserRequest {
  email: string
  name: string
  password: string
  role: 'manager' | 'member'
}

export interface UpdateUserRequest {
  name?: string
  role?: 'manager' | 'member'
  password?: string
}

// Common types
export interface ApiError {
  error: string
  message?: string
}

export interface MessageResponse {
  message: string
}
