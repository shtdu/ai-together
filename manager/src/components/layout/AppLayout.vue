<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import Header from './Header.vue'
import Sidebar from './Sidebar.vue'
import { useLicenseStore } from '../../stores/license'

const sidebarOpen = ref(false)
const licenseStore = useLicenseStore()

onMounted(() => {
  window.addEventListener('resize', handleResize)
  handleResize()
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})

function handleResize() {
  if (window.innerWidth >= 768) {
    sidebarOpen.value = false
  }
}

function toggleSidebar() {
  sidebarOpen.value = !sidebarOpen.value
}
</script>

<template>
  <div class="h-screen flex flex-col overflow-hidden bg-gray-50 dark:bg-gray-900">
    <!-- License Expiry Banner (#107) -->
    <div
      v-if="licenseStore.isExpired && !licenseStore.dismissedExpiryBanner"
      class="bg-amber-50 dark:bg-amber-900/20 border-b border-amber-200 dark:border-amber-800 px-4 py-2.5 flex items-center justify-between"
    >
      <div class="flex items-center gap-2 text-sm">
        <svg class="w-5 h-5 text-amber-600 dark:text-amber-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
        </svg>
        <span class="text-amber-800 dark:text-amber-200">
          Your commercial license has expired.
          <a
            href="https://ai-together.dev/pricing"
            target="_blank"
            rel="noopener noreferrer"
            class="font-medium underline hover:text-amber-900 dark:hover:text-amber-100"
          >Renew now</a>
          to restore advanced features.
        </span>
      </div>
      <button
        @click="licenseStore.dismissedExpiryBanner = true"
        class="text-amber-600 dark:text-amber-400 hover:text-amber-800 dark:hover:text-amber-200 transition-colors ml-4 shrink-0"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <div class="flex flex-1 overflow-hidden">
      <!-- Mobile sidebar overlay -->
      <div
        v-if="sidebarOpen"
        class="fixed inset-0 z-40 bg-black/50 md:hidden"
        @click="sidebarOpen = false"
      ></div>

      <!-- Sidebar -->
      <aside
        class="fixed inset-y-0 left-0 z-50 w-64 bg-white dark:bg-gray-800 shadow-lg transform transition-transform duration-200 ease-in-out md:relative md:translate-x-0 md:shadow-none"
        :class="{ '-translate-x-full': !sidebarOpen }"
      >
        <Sidebar />
      </aside>

      <!-- Main content -->
      <div class="flex-1 flex flex-col overflow-hidden">
        <Header @toggle-sidebar="toggleSidebar" />
        <main class="flex-1 overflow-y-auto p-6">
          <router-view />
        </main>
      </div>
    </div>
  </div>
</template>
