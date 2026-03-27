<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from './stores/auth'
import { useSetupStore } from './stores/setup'

const authStore = useAuthStore()
const setupStore = useSetupStore()

// Initialize stores on app mount
onMounted(async () => {
  await setupStore.initialize()
  await authStore.initialize()
})
</script>

<template>
  <RouterView />
</template>
