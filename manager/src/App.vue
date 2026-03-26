<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'
import { useSetupStore } from './stores/setup'

const router = useRouter()
const authStore = useAuthStore()
const setupStore = useSetupStore()

onMounted(async () => {
  // Initialize stores
  await setupStore.initialize()
  await authStore.initialize()
})

// Navigation guard for auth and setup (using return instead of deprecated next())
router.beforeEach(async (to) => {
  // If setup is required, only allow setup route
  if (setupStore.setupRequired === true && to.name !== 'setup') {
    return { name: 'setup' }
  }

  // Public routes
  if (to.meta.public) {
    // If authenticated and trying to access login, redirect to dashboard
    if (to.name === 'login' && authStore.isAuthenticated) {
      return { name: 'dashboard' }
    }
    return
  }

  // Protected routes
  if (!authStore.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  // Admin-only routes
  if (to.meta.requiresAdmin && !authStore.isAdmin) {
    return { name: 'dashboard' }
  }

  return
})
</script>

<template>
  <RouterView />
</template>
