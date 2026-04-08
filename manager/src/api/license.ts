import { apiClient } from './client'

export interface LicenseInfo {
  customer_name: string | null
  license_id: string | null
  type: string
  type_name: string
  max_seats: number
  max_teams: number
  data_retention_days: number
  issued_at: string | null
  expires_at: string | null
}

export interface LicenseUsage {
  current_users: number
  current_teams: number
  teams_remaining: number
  provider_counts: Record<string, number>
}

export interface LicenseStatus {
  has_active_license: boolean
  days_remaining: number
  is_using_defaults?: boolean
  can_add_user: boolean
  can_add_provider: Record<string, boolean>
  can_create_team: boolean
}

export interface LicenseResponse {
  license: LicenseInfo
  usage: LicenseUsage
  status: LicenseStatus
}

export const licenseApi = {
  getLicense() {
    return apiClient.get<LicenseResponse>('/api/v1/license')
  },
}
