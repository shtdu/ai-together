<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'

interface Props {
  onMenuClick?: () => void
}

defineProps<Props>()

const router = useRouter()
const authStore = useAuthStore()
const showMenu = ref(false)

function handleProfile() {
  showMenu.value = false
  router.push('/profile')
}

async function handleLogout() {
  showMenu.value = false
  await authStore.logout()
  router.push('/login')
}

function getInitials(name: string): string {
  return name?.charAt(0).toUpperCase() || 'U'
}
</script>

<template>
  <header class="fixed top-0 left-0 right-0 h-16 bg-blue-600 dark:bg-blue-800 text-white flex items-center px-4 z-50">
    <!-- Mobile menu button -->
    <button
      v-if="onMenuClick"
      @click="onMenuClick"
      class="sm:hidden p-2 mr-2 hover:bg-blue-700 dark:hover:bg-blue-900 rounded"
    >
      <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
      </svg>
    </button>

    <!-- Logo -->
    <div class="flex items-center gap-2 mr-4">
      <svg class="w-6 h-6" fill="currentColor" viewBox="0 0 24 24">
        <path d="M12 2L2 7v10c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V7l-10-5z" />
      </svg>
      <span class="text-xl font-bold">AI Together</span>
    </div>

    <div class="h-8 w-px bg-white/30 mx-2 hidden sm:block"></div>

    <span class="flex-1 text-blue-100 hidden sm:block">Token Analysis Dashboard</span>

    <!-- User menu -->
    <div class="relative">
      <button
        @click="showMenu = !showMenu"
        class="flex items-center justify-center w-10 h-10 rounded-full bg-blue-500 dark:bg-blue-700 hover:bg-blue-400 dark:hover:bg-blue-600 transition-colors"
      >
        <span class="text-lg font-bold">{{ getInitials(authStore.user?.name || '') }}</span>
      </button>

      <div
        v-if="showMenu"
        class="absolute right-0 mt-2 w-64 bg-white dark:bg-gray-800 rounded-lg shadow-lg py-2 z-50"
      >
        <div class="px-4 py-2 border-b border-gray-200 dark:border-gray-700">
          <p class="text-sm font-medium text-gray-900 dark:text-gray-100">{{ authStore.user?.name }}</p>
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ authStore.user?.email }}</p>
          <p class="text-xs text-blue-600 dark:text-blue-400 mt-1">
            {{ authStore.user?.role === 'manager' ? 'Admin' : 'Member' }}
          </p>
        </div>
        <button
          @click="handleProfile"
          class="w-full px-4 py-2 text-left text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
          Profile
        </button>
        <button
          @click="handleLogout"
          class="w-full px-4 py-2 text-left text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
          </svg>
          Logout
        </button>
      </div>
    </div>

    <!-- Click outside to close menu -->
    <div
      v-if="showMenu"
      @click="showMenu = false"
      class="fixed inset-0 z-40"
    ></div>
  </header>
</template>
