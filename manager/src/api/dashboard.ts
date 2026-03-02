import { apiClient } from './client'
import type {
  DashboardMetricsResponse,
  DashboardRankingsResponse,
  DashboardMembersResponse,
} from '../types/api'

export const dashboardApi = {
  getMetrics: async (
    range: '24h' | '7d' | '30d' = '7d',
    interval: 'hour' | 'day' = 'hour'
  ): Promise<DashboardMetricsResponse> => {
    const response = await apiClient.get<DashboardMetricsResponse>(
      `/api/v1/dashboard/metrics`,
      { params: { range, interval } }
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
