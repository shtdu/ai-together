<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from './stores/auth'
import { useSetupStore } from './stores/setup'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const setupStore = useSetupStore()

onMounted(async () => {
  // Initialize stores
  await setupStore.initialize()
  await authStore.initialize()
})

// Navigation guard for auth and setup
router.beforeEach(async (to, from, next) => {
  // If setup is required, only allow setup route
  if (setupStore.setupRequired === true && to.name !== 'setup') {
    next({ name: 'setup' })
    return
  }

  // Public routes
  if (to.meta.public) {
    // If authenticated and trying to access login, redirect to dashboard
    if (to.name === 'login' && authStore.isAuthenticated) {
      next({ name: 'dashboard' })
      return
    }
    next()
    return
  }

  // Protected routes
  if (!authStore.isAuthenticated) {
    next({ name: 'login', query: { redirect: to.fullPath } })
    return
  }

  // Admin-only routes
  if (to.meta.requiresAdmin && !authStore.isAdmin) {
    next({ name: 'dashboard' })
    return
  }

  next()
})
</script>

<template>
  <RouterView />
</template>
