import { apiClient } from './client'

export interface SetupStatusResponse {
  setup_required: boolean
}

export interface SetupAdminRequest {
  organization_name: string
  admin_email: string
  admin_name: string
  admin_password: string
}

export interface SetupAdminResponse {
  message: string
  user: {
    id: number
    email: string
    name: string
    role: string
    tenant_id: number
  }
  tenant: {
    id: number
    name: string
  }
  team: {
    id: number
    name: string
  }
}

export const setupApi = {
  getStatus: async (): Promise<SetupStatusResponse> => {
    const response = await apiClient.get<SetupStatusResponse>('/api/v1/setup/status')
    return response.data
  },

  createAdmin: async (data: SetupAdminRequest): Promise<SetupAdminResponse> => {
    const response = await apiClient.post<SetupAdminResponse>('/api/v1/setup/admin', data)
    return response.data
  },
}
