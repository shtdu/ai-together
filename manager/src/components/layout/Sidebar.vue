<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../../stores/auth'

interface NavItem {
  label: string
  path: string
  icon: string
  adminOnly?: boolean
}

const navItems: NavItem[] = [
  { label: 'Dashboard', path: '/', icon: 'dashboard' },
  { label: 'Personal Analytics', path: '/analytics/personal', icon: 'personal' },
  { label: 'Provider Stats', path: '/analytics/providers', icon: 'storage', adminOnly: true },
  { label: 'User Stats', path: '/analytics/users', icon: 'chart', adminOnly: true },
  { label: 'History', path: '/analytics/history', icon: 'history', adminOnly: true },
  { label: 'Users', path: '/users', icon: 'users', adminOnly: true },
]

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const filteredItems = computed(() => {
  return navItems.filter(item => !item.adminOnly || authStore.isAdmin)
})

function navigate(path: string) {
  router.push(path)
}

function isActive(path: string): boolean {
  return route.path === path
}

// Icon components as strings (will be rendered with heroicons-style paths)
const icons: Record<string, string> = {
  dashboard: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />',
  personal: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />',
  storage: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />',
  chart: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />',
  history: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />',
  users: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />',
}
</script>

<template>
  <nav class="h-full overflow-y-auto py-4">
    <div class="space-y-1 px-2">
      <button
        v-for="item in filteredItems"
        :key="item.path"
        @click="navigate(item.path)"
        class="w-full flex items-center px-4 py-3 rounded-lg transition-colors"
        :class="{
          'bg-blue-600 text-white': isActive(item.path),
          'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800': !isActive(item.path)
        }"
      >
        <svg class="w-5 h-5 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" v-html="icons[item.icon]"></svg>
        <span class="font-medium">{{ item.label }}</span>
      </button>
    </div>
  </nav>
</template>
