<script setup lang="ts">
import { ref, computed } from 'vue'

interface UserOption {
  id: number
  name: string
  email: string
}

interface Props {
  label: string
  options: UserOption[]
  modelValue: number[]
  width?: number | string
}

interface Emits {
  (e: 'update:modelValue', value: number[]): void
}

const props = withDefaults(defineProps<Props>(), {
  width: 250
})

const emit = defineEmits<Emits>()

const isOpen = ref(false)
const searchQuery = ref('')

const selectedOptions = computed(() => {
  return props.options.filter(opt => props.modelValue.includes(opt.id))
})

const filteredOptions = computed(() => {
  if (!searchQuery.value) return props.options
  const query = searchQuery.value.toLowerCase()
  return props.options.filter(opt =>
    opt.name.toLowerCase().includes(query) ||
    opt.email.toLowerCase().includes(query)
  )
})

function toggleUser(user: UserOption) {
  const newValue = props.modelValue.includes(user.id)
    ? props.modelValue.filter(id => id !== user.id)
    : [...props.modelValue, user.id]
  emit('update:modelValue', newValue)
}

function isSelected(user: UserOption): boolean {
  return props.modelValue.includes(user.id)
}

function removeUser(userId: number) {
  emit('update:modelValue', props.modelValue.filter(id => id !== userId))
}

function handleClickOutside() {
  isOpen.value = false
}
</script>

<template>
  <div class="relative" :style="{ minWidth: typeof width === 'number' ? `${width}px` : width }">
    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ label }}</label>
    <button
      @click="isOpen = !isOpen"
      class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white text-left flex items-center justify-between"
    >
      <span class="truncate">
        <template v-if="modelValue.length === 0">Search users...</template>
        <template v-else>{{ modelValue.length }} selected</template>
      </span>
      <span class="ml-2">{{ isOpen ? '▲' : '▼' }}</span>
    </button>

    <!-- Selected chips -->
    <div v-if="selectedOptions.length > 0 && !isOpen" class="flex flex-wrap gap-1 mt-1">
      <span
        v-for="user in selectedOptions"
        :key="user.id"
        class="inline-flex items-center px-2 py-1 bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 text-xs rounded"
      >
        {{ user.name }}
        <button
          @click.stop="removeUser(user.id)"
          class="ml-1 hover:text-red-600 dark:hover:text-red-400"
        >
          ×
        </button>
      </span>
    </div>

    <!-- Dropdown -->
    <div
      v-if="isOpen"
      class="absolute z-10 w-full mt-1 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md shadow-lg max-h-60 overflow-auto"
    >
      <!-- Search input -->
      <div class="p-2 border-b border-gray-200 dark:border-gray-700">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search users..."
          class="w-full px-2 py-1 border border-gray-300 dark:border-gray-600 rounded text-sm dark:bg-gray-700 dark:text-white"
        />
      </div>

      <!-- Options -->
      <div
        v-for="user in filteredOptions"
        :key="user.id"
        @click="toggleUser(user)"
        class="px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer flex items-center"
        :class="{ 'bg-blue-50 dark:bg-blue-900/30': isSelected(user) }"
      >
        <input
          type="checkbox"
          :checked="isSelected(user)"
          class="mr-2"
          @click.stop
        />
        <div class="flex-1 min-w-0">
          <div class="text-sm font-medium dark:text-gray-200 truncate">{{ user.name }}</div>
          <div class="text-xs text-gray-500 dark:text-gray-400 truncate">{{ user.email }}</div>
        </div>
      </div>

      <div v-if="filteredOptions.length === 0" class="px-3 py-2 text-gray-500 dark:text-gray-400">
        No users found
      </div>
    </div>

    <!-- Click outside handler -->
    <div
      v-if="isOpen"
      @click="handleClickOutside"
      class="fixed inset-0 z-0"
    ></div>
  </div>
</template>
