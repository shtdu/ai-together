import {
  createContext,
  useState,
  useEffect,
  useCallback,
  type ReactNode,
} from 'react'
import { authApi } from '../api/auth'
import { getErrorMessage } from '../api/client'
import type { User, Team } from '../types/models'

interface AuthState {
  user: User | null
  teams: Team[]
  isAuthenticated: boolean
  isLoading: boolean
  isAdmin: boolean
}

interface AuthContextValue extends AuthState {
  login: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
}

// eslint-disable-next-line react-refresh/only-export-components
export const AuthContext = createContext<AuthContextValue | undefined>(undefined)

interface AuthProviderProps {
  children: ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [state, setState] = useState<AuthState>({
    user: null,
    teams: [],
    isAuthenticated: false,
    isLoading: true,
    isAdmin: false,
  })

  const refreshUser = useCallback(async () => {
    const token = localStorage.getItem('access_token')
    if (!token) {
      setState({
        user: null,
        teams: [],
        isAuthenticated: false,
        isLoading: false,
        isAdmin: false,
      })
      return
    }

    try {
      const profile = await authApi.getProfile()
      setState({
        user: profile.user,
        teams: profile.teams || [],
        isAuthenticated: true,
        isLoading: false,
        isAdmin: profile.user.role === 'manager',
      })
    } catch {
      // Token invalid, clear storage
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      setState({
        user: null,
        teams: [],
        isAuthenticated: false,
        isLoading: false,
        isAdmin: false,
      })
    }
  }, [])

  useEffect(() => {
    refreshUser()
  }, [refreshUser])

  const login = useCallback(async (email: string, password: string) => {
    try {
      const response = await authApi.login({ email, password })
      localStorage.setItem('access_token', response.access_token)
      localStorage.setItem('refresh_token', response.refresh_token)

      // Fetch full profile after login
      const profile = await authApi.getProfile()
      setState({
        user: profile.user,
        teams: profile.teams || [],
        isAuthenticated: true,
        isLoading: false,
        isAdmin: profile.user.role === 'manager',
      })
    } catch (error) {
      throw new Error(getErrorMessage(error))
    }
  }, [])

  const logout = useCallback(async () => {
    try {
      await authApi.logout()
    } catch {
      // Ignore logout API errors
    } finally {
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      setState({
        user: null,
        teams: [],
        isAuthenticated: false,
        isLoading: false,
        isAdmin: false,
      })
    }
  }, [])

  return (
    <AuthContext.Provider
      value={{
        ...state,
        login,
        logout,
        refreshUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}
