import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { licenseApi } from '../api/license'
import { getErrorMessage } from '../api/client'

export const useLicenseStore = defineStore('license', () => {
  const hasActiveLicense = ref(false)
  const isCommercial = computed(() => {
    return hasActiveLicense.value && licenseType.value === 'commercial'
  })
  const isInitialized = ref(false)
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const licenseType = ref<string>('open_source')
  const expiresAt = ref<string | null>(null)
  const daysRemaining = ref(0)
  const canCreateTeam = ref(true)
  const dismissedExpiryBanner = ref(false)

  const isExpired = computed(() => {
    // Active license is not expired
    if (hasActiveLicense.value) return false
    // If there's an expiry date and license is not active, it's expired
    return !!expiresAt.value
  })

  async function fetchLicense() {
    isLoading.value = true
    error.value = null
    try {
      const response = await licenseApi.getLicense()
      hasActiveLicense.value = response.data.status.has_active_license
      licenseType.value = response.data.license.type
      expiresAt.value = response.data.license.expires_at as string | null
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
    isCommercial,
    isInitialized,
    isLoading,
    error,
    licenseType,
    expiresAt,
    isExpired,
    daysRemaining,
    canCreateTeam,
    dismissedExpiryBanner,
    fetchLicense,
    initialize,
  }
})
