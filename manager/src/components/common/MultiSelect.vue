<script setup lang="ts">
import { ref, computed } from 'vue'

interface Props {
  label: string
  options: string[]
  modelValue: string[]
  width?: number | string
  /** Optional map of value -> display label. If provided, shows labels in UI but emits raw values. */
  labels?: Record<string, string>
}

interface Emits {
  (e: 'update:modelValue', value: string[]): void
}

const props = withDefaults(defineProps<Props>(), {
  width: 200,
  labels: undefined,
})

const emit = defineEmits<Emits>()

const isOpen = ref(false)
const searchQuery = ref('')

function getLabel(value: string): string {
  return props.labels?.[value] ?? value
}

const filteredOptions = computed(() => {
  if (!searchQuery.value) return props.options
  const query = searchQuery.value.toLowerCase()
  return props.options.filter(opt => {
    const display = getLabel(opt).toLowerCase()
    return display.includes(query) || opt.toLowerCase().includes(query)
  })
})

function toggleOption(option: string) {
  const newValue = props.modelValue.includes(option)
    ? props.modelValue.filter(v => v !== option)
    : [...props.modelValue, option]
  emit('update:modelValue', newValue)
}

function isSelected(option: string): boolean {
  return props.modelValue.includes(option)
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
        <template v-if="modelValue.length === 0">Select options...</template>
        <template v-else>{{ modelValue.length }} selected</template>
      </span>
      <span class="ml-2">{{ isOpen ? '▲' : '▼' }}</span>
    </button>

    <!-- Selected chips -->
    <div v-if="modelValue.length > 0 && !isOpen" class="flex flex-wrap gap-1 mt-1">
      <span
        v-for="val in modelValue"
        :key="val"
        class="inline-flex items-center px-2 py-1 bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 text-xs rounded"
      >
        {{ getLabel(val) }}
        <button
          @click.stop="toggleOption(val)"
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
          placeholder="Search..."
          class="w-full px-2 py-1 border border-gray-300 dark:border-gray-600 rounded text-sm dark:bg-gray-700 dark:text-white"
        />
      </div>

      <!-- Options -->
      <div
        v-for="option in filteredOptions"
        :key="option"
        @click="toggleOption(option)"
        class="px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer flex items-center"
        :class="{ 'bg-blue-50 dark:bg-blue-900/30': isSelected(option) }"
      >
        <input
          type="checkbox"
          :checked="isSelected(option)"
          class="mr-2"
          @click.stop
        />
        <span class="dark:text-gray-200">{{ getLabel(option) }}</span>
      </div>

      <div v-if="filteredOptions.length === 0" class="px-3 py-2 text-gray-500 dark:text-gray-400">
        No options found
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
