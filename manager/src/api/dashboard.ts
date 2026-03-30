import { apiClient } from './client'
import type {
  DashboardMetricsResponse,
  DashboardRankingsResponse,
  DashboardMembersResponse,
} from '../types/api'

export const dashboardApi = {
  getMetrics: async (
    range: '24h' | '7d' | '30d' | 'custom' = '7d',
    interval: 'hour' | 'day' = 'hour',
    startDate?: string,
    endDate?: string
  ): Promise<DashboardMetricsResponse> => {
    const params: Record<string, string> = { range, interval }
    if (range === 'custom' && startDate && endDate) {
      params.start_date = startDate
      params.end_date = endDate
    }
    const response = await apiClient.get<DashboardMetricsResponse>(
      `/api/v1/dashboard/metrics`,
      { params }
    )
    return response.data
  },

  getRankings: async (): Promise<DashboardRankingsResponse> => {
    const response = await apiClient.get<DashboardRankingsResponse>(
      `/api/v1/dashboard/rankings`
    )
    return response.data
  },

  getMembers: async (): Promise<DashboardMembersResponse> => {
    const response = await apiClient.get<DashboardMembersResponse>(
      `/api/v1/dashboard/members`
    )
    return response.data
  },
}
