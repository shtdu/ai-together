import { apiClient } from './client'
import type {
  LoginRequest,
  LoginResponse,
  RefreshRequest,
  RefreshResponse,
  UserProfile,
  MessageResponse,
} from '../types/api'

export const authApi = {
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    const response = await apiClient.post<LoginResponse>('/auth/login', data)
    return response.data
  },

  logout: async (): Promise<MessageResponse> => {
    const response = await apiClient.post<MessageResponse>('/api/v1/logout')
    return response.data
  },

  refresh: async (data: RefreshRequest): Promise<RefreshResponse> => {
    const response = await apiClient.post<RefreshResponse>('/auth/refresh', data)
    return response.data
  },

  verify: async (): Promise<{ valid: boolean }> => {
    const response = await apiClient.post<{ valid: boolean }>('/auth/verify')
    return response.data
  },

  getProfile: async (): Promise<UserProfile> => {
    const response = await apiClient.get<UserProfile>('/api/v1/user/profile')
    return response.data
  },
}
