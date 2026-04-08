<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import { useLicenseStore } from '../../stores/license'

interface NavItem {
  label: string
  path: string
  icon: string
  adminOnly?: boolean
  requiresCommercial?: boolean
}

const navItems: NavItem[] = [
  { label: 'Dashboard', path: '/', icon: 'dashboard' },
  { label: 'Personal Analytics', path: '/analytics/personal', icon: 'personal' },
  { label: 'Provider Stats', path: '/analytics/providers', icon: 'storage', adminOnly: true, requiresCommercial: true },
  { label: 'User Stats', path: '/analytics/users', icon: 'chart', adminOnly: true, requiresCommercial: true },
  { label: 'History', path: '/analytics/history', icon: 'history', adminOnly: true, requiresCommercial: true },
  { label: 'Users', path: '/users', icon: 'users', adminOnly: true },
]

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const licenseStore = useLicenseStore()

const showUpgradeModal = ref(false)

// Show upgrade modal if navigated with licenseRequired query
onMounted(() => {
  if (route.query.licenseRequired) {
    showUpgradeModal.value = true
  }
})

const filteredItems = computed(() => {
  return navItems.filter(item => !item.adminOnly || authStore.isAdmin)
})

function navigate(item: NavItem) {
  if (item.requiresCommercial && !licenseStore.hasActiveLicense) {
    showUpgradeModal.value = true
    return
  }
  router.push(item.path)
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
  lock: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />',
}
</script>

<template>
  <nav class="h-full overflow-y-auto py-4">
    <div class="space-y-1 px-2">
      <button
        v-for="item in filteredItems"
        :key="item.path"
        @click="navigate(item)"
        class="w-full flex items-center px-4 py-3 rounded-lg transition-colors"
        :class="{
          'bg-blue-600 text-white': isActive(item.path),
          'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800': !isActive(item.path) && !(item.requiresCommercial && !licenseStore.hasActiveLicense),
          'text-gray-400 dark:text-gray-500 cursor-pointer': item.requiresCommercial && !licenseStore.hasActiveLicense && !isActive(item.path),
        }"
      >
        <svg class="w-5 h-5 mr-3 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" v-html="icons[item.icon]"></svg>
        <span class="font-medium flex-1 text-left">{{ item.label }}</span>
        <svg
          v-if="item.requiresCommercial && !licenseStore.hasActiveLicense"
          class="w-4 h-4 shrink-0"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          v-html="icons.lock"
        ></svg>
      </button>
    </div>

    <!-- Upgrade Modal -->
    <Teleport to="body">
      <div
        v-if="showUpgradeModal"
        class="fixed inset-0 z-50 flex items-center justify-center"
      >
        <div class="fixed inset-0 bg-black/50" @click="showUpgradeModal = false"></div>
        <div class="relative bg-white dark:bg-gray-800 rounded-xl shadow-xl max-w-md w-full mx-4 p-6">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-10 h-10 bg-amber-100 dark:bg-amber-900/30 rounded-full flex items-center justify-center">
              <svg class="w-5 h-5 text-amber-600 dark:text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Commercial License Required</h3>
          </div>
          <p class="text-gray-600 dark:text-gray-400 mb-4">
            This feature requires a commercial license. Upgrade to unlock team analytics, cost tracking, request history, and more.
          </p>
          <ul class="text-sm text-gray-500 dark:text-gray-400 mb-6 space-y-1">
            <li>✅ Team analytics dashboard</li>
            <li>✅ Provider performance metrics</li>
            <li>✅ Request history with pagination</li>
            <li>✅ 90-day data retention</li>
            <li>✅ Multi-team support</li>
          </ul>
          <div class="flex gap-3">
            <button
              @click="showUpgradeModal = false"
              class="flex-1 px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 dark:bg-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
            >
              Close
            </button>
            <a
              href="https://ai-together.dev/pricing"
              target="_blank"
              rel="noopener noreferrer"
              class="flex-1 px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors text-center"
            >
              Learn More
            </a>
          </div>
        </div>
      </div>
    </Teleport>
  </nav>
</template>
