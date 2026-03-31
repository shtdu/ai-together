import { apiClient } from './client';

export interface ProviderAnalyticsParams {
  start_date: string;
  end_date: string;
  providers?: string[];
  models?: string[];
  tools?: string[];
}

export interface UserAnalyticsParams {
  start_date: string;
  end_date: string;
  user_ids?: number[];
  providers?: string[];
  tools?: string[];
}

export interface HistoryParams {
  start_date: string;
  end_date: string;
  page?: number;
  limit?: number;
  user_ids?: number[];
  providers?: string[];
  models?: string[];
  tools?: string[];
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface FilterOptions {
  providers: string[];
  models: string[];
  tools: string[];
  users: Array<{ id: number; name: string; email: string }>;
}

export interface ProviderAnalyticsResponse {
  summary: {
    total_requests: number;
    total_tokens: number;
    total_input: number;
    total_output: number;
    total_cost: number;
  };
  distribution: {
    by_provider: Array<{
      name: string;
      requests: number;
      tokens: number;
      cost: number;
      percentage: number;
    }>;
    by_model: Array<{
      name: string;
      requests: number;
      tokens: number;
      cost: number;
      percentage: number;
    }>;
    by_tool: Array<{
      name: string;
      requests: number;
      tokens: number;
      cost: number;
      percentage: number;
    }>;
  };
  trend_data: Array<{
    date: string;
    tokens?: number;
    input_tokens?: number;
    output_tokens?: number;
    requests?: number;
    by_provider?: Record<string, number>;
  }>;
  period: {
    start: string;
    end: string;
  };
}

export interface UserAnalyticsResponse {
  leaderboard: Array<{
    user_id: number;
    name: string;
    total_tokens: number;
    total_requests: number;
    total_cost: number;
    rank: number;
    percentage: number;
  }>;
  details: Array<{
    user_id: number;
    user_name: string;
    date: string;
    provider: string;
    model: string;
    total_tokens: number;
    avg_latency_ms: number;
    total_cost: number;
  }>;
  period: {
    start: string;
    end: string;
  };
}

export interface HistoryResponse {
  records: Array<{
    id: number;
    user_id: number;
    user_name: string;
    provider: string;
    model: string;
    platform: string;
    input_tokens: number;
    output_tokens: number;
    duration_sec: number;
    estimated_cost: number;
    timestamp: string;
  }>;
  pagination: {
    page: number;
    limit: number;
    total_count: number;
    total_pages: number;
  };
}

export const analyticsApi = {
  getProviderAnalytics: async (params: ProviderAnalyticsParams): Promise<ProviderAnalyticsResponse> => {
    const queryParams = new URLSearchParams({
      start_date: params.start_date,
      end_date: params.end_date,
    });
    if (params.providers?.length) {
      queryParams.set('providers', params.providers.join(','));
    }
    if (params.models?.length) {
      queryParams.set('models', params.models.join(','));
    }
    if (params.tools?.length) {
      queryParams.set('tools', params.tools.join(','));
    }
    const response = await apiClient.get(`/api/v1/analytics/providers?${queryParams}`);
    return response.data;
  },

  getUserAnalytics: async (params: UserAnalyticsParams): Promise<UserAnalyticsResponse> => {
    const queryParams = new URLSearchParams({
      start_date: params.start_date,
      end_date: params.end_date,
    });
    if (params.user_ids?.length) {
      queryParams.set('user_ids', params.user_ids.join(','));
    }
    if (params.providers?.length) {
      queryParams.set('providers', params.providers.join(','));
    }
    if (params.tools?.length) {
      queryParams.set('tools', params.tools.join(','));
    }
    const response = await apiClient.get(`/api/v1/analytics/users?${queryParams}`);
    return response.data;
  },

  getHistory: async (params: HistoryParams): Promise<HistoryResponse> => {
    const queryParams = new URLSearchParams({
      start_date: params.start_date,
      end_date: params.end_date,
    });
    if (params.page) queryParams.set('page', params.page.toString());
    if (params.limit) queryParams.set('limit', params.limit.toString());
    if (params.user_ids?.length) {
      queryParams.set('user_ids', params.user_ids.join(','));
    }
    if (params.providers?.length) {
      queryParams.set('providers', params.providers.join(','));
    }
    if (params.models?.length) {
      queryParams.set('models', params.models.join(','));
    }
    if (params.tools?.length) {
      queryParams.set('tools', params.tools.join(','));
    }
    if (params.sort_by) queryParams.set('sort_by', params.sort_by);
    if (params.sort_order) queryParams.set('sort_order', params.sort_order);
    const response = await apiClient.get(`/api/v1/analytics/history?${queryParams}`);
    return response.data;
  },

  getFilterOptions: async (): Promise<FilterOptions> => {
    const response = await apiClient.get('/api/v1/analytics/filters');
    return response.data;
  },

  getPersonalAnalytics: async (params: ProviderAnalyticsParams): Promise<ProviderAnalyticsResponse> => {
    const queryParams = new URLSearchParams({
      start_date: params.start_date,
      end_date: params.end_date,
    });
    if (params.providers?.length) {
      queryParams.set('providers', params.providers.join(','));
    }
    if (params.models?.length) {
      queryParams.set('models', params.models.join(','));
    }
    if (params.tools?.length) {
      queryParams.set('tools', params.tools.join(','));
    }
    const response = await apiClient.get(`/api/v1/analytics/personal?${queryParams}`);
    return response.data;
  },

  getPersonalHistory: async (params: HistoryParams): Promise<HistoryResponse> => {
    const queryParams = new URLSearchParams({
      start_date: params.start_date,
      end_date: params.end_date,
    });
    if (params.page) queryParams.set('page', params.page.toString());
    if (params.limit) queryParams.set('limit', params.limit.toString());
    if (params.providers?.length) {
      queryParams.set('providers', params.providers.join(','));
    }
    if (params.models?.length) {
      queryParams.set('models', params.models.join(','));
    }
    if (params.tools?.length) {
      queryParams.set('tools', params.tools.join(','));
    }
    if (params.sort_by) queryParams.set('sort_by', params.sort_by);
    if (params.sort_order) queryParams.set('sort_order', params.sort_order);
    const response = await apiClient.get(`/api/v1/analytics/personal/history?${queryParams}`);
    return response.data;
  },

  getPersonalFilterOptions: async (): Promise<FilterOptions> => {
    const response = await apiClient.get('/api/v1/analytics/personal/filters');
    return response.data;
  },
};
