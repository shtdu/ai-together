import { apiClient } from './client'
import type { User } from '../types/models'
import type { MessageResponse } from '../types/api'

export interface UpdateProfileRequest {
  name: string
}

export interface ChangePasswordRequest {
  current_password: string
  new_password: string
}

export const profileApi = {
  updateProfile: async (data: UpdateProfileRequest): Promise<{ user: User }> => {
    const response = await apiClient.put<{ user: User }>('/api/v1/user/profile', data)
    return response.data
  },

  changePassword: async (data: ChangePasswordRequest): Promise<MessageResponse> => {
    const response = await apiClient.put<MessageResponse>('/api/v1/user/password', data)
    return response.data
  },
}
