import { defineStore } from 'pinia'
import { ref } from 'vue'
import { setupApi } from '../api/setup'

export const useSetupStore = defineStore('setup', () => {
  const setupRequired = ref<boolean | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const isInitialized = ref(false)

  async function checkSetupStatus() {
    try {
      isLoading.value = true
      error.value = null
      const response = await setupApi.getStatus()
      setupRequired.value = response.setup_required
    } catch (err) {
      setupRequired.value = null
      error.value = err instanceof Error ? err.message : 'Failed to check setup status'
    } finally {
      isLoading.value = false
    }
  }

  async function initialize() {
    await checkSetupStatus()
    isInitialized.value = true
  }

  return {
    setupRequired,
    isLoading,
    error,
    isInitialized,
    checkSetupStatus,
    initialize,
  }
})
