<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import Header from './Header.vue'
import Sidebar from './Sidebar.vue'

const MIN_DRAWER_WIDTH = 180
const MAX_DRAWER_WIDTH = 400
const DEFAULT_DRAWER_WIDTH = 240

const mobileOpen = ref(false)
const drawerWidth = ref(DEFAULT_DRAWER_WIDTH)
const isResizing = ref(false)

// Initialize drawer width from localStorage
const savedWidth = localStorage.getItem('sidebarWidth')
if (savedWidth) {
  drawerWidth.value = parseInt(savedWidth, 10)
}

function handleDrawerToggle() {
  mobileOpen.value = !mobileOpen.value
}

function handleMouseDown(e: MouseEvent) {
  e.preventDefault()
  isResizing.value = true
}

function handleMouseMove(e: MouseEvent) {
  if (!isResizing.value) return

  const newWidth = e.clientX
  if (newWidth >= MIN_DRAWER_WIDTH && newWidth <= MAX_DRAWER_WIDTH) {
    drawerWidth.value = newWidth
  }
}

function handleMouseUp() {
  if (isResizing.value) {
    isResizing.value = false
    localStorage.setItem('sidebarWidth', drawerWidth.value.toString())
  }
}

onMounted(() => {
  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', handleMouseUp)
})

onUnmounted(() => {
  document.removeEventListener('mousemove', handleMouseMove)
  document.removeEventListener('mouseup', handleMouseUp)
})
</script>

<template>
  <div class="flex min-h-screen bg-gray-50 dark:bg-gray-900">
    <Header @menu-click="handleDrawerToggle" />

    <!-- Mobile drawer -->
    <div
      v-if="mobileOpen"
      class="fixed inset-0 z-40 sm:hidden"
    >
      <div class="absolute inset-0 bg-black/50" @click="handleDrawerToggle"></div>
      <div
        class="absolute left-0 top-0 bottom-0 w-64 bg-white dark:bg-gray-800 shadow-lg"
        :style="{ width: `${drawerWidth}px` }"
      >
        <Sidebar />
      </div>
    </div>

    <!-- Desktop drawer -->
    <aside
      class="hidden sm:flex bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 flex-shrink-0"
      :style="{ width: `${drawerWidth}px` }"
    >
      <div class="flex-1 overflow-hidden">
        <Sidebar />
      </div>

      <!-- Resize handle -->
      <div
        @mousedown="handleMouseDown"
        class="w-1 cursor-col-resize hover:bg-blue-500 transition-colors"
        :class="{ 'bg-blue-500': isResizing }"
      ></div>
    </aside>

    <!-- Main content -->
    <main
      class="flex-1 p-6 bg-gray-50 dark:bg-gray-900 min-h-screen"
      :style="{ width: `calc(100% - ${drawerWidth}px)` }"
      :class="{ 'sm:ml-0': mobileOpen, 'sm:hidden': mobileOpen }"
    >
      <!-- Spacer for fixed header -->
      <div class="h-16"></div>
      <RouterView />
    </main>
  </div>
</template>
