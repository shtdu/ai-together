import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '../api/auth'
import { getErrorMessage } from '../api/client'
import type { User, Team } from '../types/models'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const teams = ref<Team[]>([])
  const isLoading = ref(false)

  const isAuthenticated = computed(() => !!user.value)
  const isAdmin = computed(() => user.value?.role === 'manager')

  async function login(email: string, password: string) {
    try {
      const response = await authApi.login({ email, password })
      localStorage.setItem('access_token', response.access_token)
      localStorage.setItem('refresh_token', response.refresh_token)

      // Fetch full profile after login
      const profile = await authApi.getProfile()
      user.value = profile.user
      teams.value = profile.teams || []
    } catch (error) {
      throw new Error(getErrorMessage(error))
    }
  }

  async function logout() {
    try {
      await authApi.logout()
    } catch {
      // Ignore logout API errors
    } finally {
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      user.value = null
      teams.value = []
    }
  }

  async function refreshUser() {
    const token = localStorage.getItem('access_token')
    if (!token) {
      user.value = null
      teams.value = []
      return
    }

    try {
      const profile = await authApi.getProfile()
      user.value = profile.user
      teams.value = profile.teams || []
    } catch {
      // Token invalid, clear storage
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      user.value = null
      teams.value = []
    }
  }

  async function initialize() {
    isLoading.value = true
    await refreshUser()
    isLoading.value = false
  }

  return {
    user,
    teams,
    isLoading,
    isAuthenticated,
    isAdmin,
    login,
    logout,
    refreshUser,
    initialize,
  }
})
