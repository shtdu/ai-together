import { Call } from '@wailsio/runtime'

export type User = {
  id: number
  email: string
  name: string
  role: string
  tenant_id: number
  created_at: string
  updated_at: string
}

export type AuthTokens = {
  access_token: string
  refresh_token: string
  expires_at: string
}

export type AuthResponse = {
  access_token: string
  refresh_token: string
  user: User
  expires_at: string
}

export type LoginRequest = {
  email: string
  password: string
}

export const getCurrentUser = async (): Promise<User> => {
  return Call.ByName('codeswitch/services.AuthService.GetCurrentUser')
}

export const isAuthenticated = async (): Promise<boolean> => {
  return Call.ByName('codeswitch/services.AuthService.IsAuthenticated')
}

export const login = async (email: string, password: string): Promise<AuthResponse> => {
  return Call.ByName('codeswitch/services.AuthService.Login', email, password)
}

export const logout = async (): Promise<void> => {
  return Call.ByName('codeswitch/services.AuthService.Logout')
}

/**
 * Fetches the current user's profile from the server
 * This refreshes the user info (including role) without re-authentication
 */
export const getUserProfile = async (): Promise<User> => {
  return Call.ByName('codeswitch/services.AuthService.GetUserProfile')
}
