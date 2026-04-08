import { defineStore } from 'pinia'
import { ref } from 'vue'
import { licenseApi } from '../api/license'
import { getErrorMessage } from '../api/client'

export const useLicenseStore = defineStore('license', () => {
  const hasActiveLicense = ref(false)
  const isInitialized = ref(false)
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const licenseType = ref<string>('open_source')
  const daysRemaining = ref(0)
  const canCreateTeam = ref(true)

  async function fetchLicense() {
    isLoading.value = true
    error.value = null
    try {
      const response = await licenseApi.getLicense()
      hasActiveLicense.value = response.data.status.has_active_license
      licenseType.value = response.data.license.type
      daysRemaining.value = response.data.status.days_remaining
      canCreateTeam.value = response.data.status.can_create_team
    } catch (err) {
      error.value = getErrorMessage(err)
      // Default to open source on error
      hasActiveLicense.value = false
    } finally {
      isLoading.value = false
    }
  }

  async function initialize() {
    await fetchLicense()
    isInitialized.value = true
  }

  return {
    hasActiveLicense,
    isInitialized,
    isLoading,
    error,
    licenseType,
    daysRemaining,
    canCreateTeam,
    fetchLicense,
    initialize,
  }
})
