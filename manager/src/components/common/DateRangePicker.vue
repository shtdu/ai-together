<script setup lang="ts">
import { ref } from 'vue'
import dayjs from 'dayjs'

interface Props {
  startDate: string
  endDate: string
}

interface Emits {
  (e: 'change', startDate: string, endDate: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const showPresets = ref(false)

const presets = [
  { label: 'Today', days: 0 },
  { label: 'Last 7 days', days: 7 },
  { label: 'Last 14 days', days: 14 },
  { label: 'Last 30 days', days: 30 },
  { label: 'Last 90 days', days: 90 },
]

function handlePresetClick(days: number) {
  const end = dayjs().format('YYYY-MM-DD')
  const start = dayjs().subtract(days, 'day').format('YYYY-MM-DD')
  emit('change', start, end)
  showPresets.value = false
}

function handleStartDateChange(e: Event) {
  const target = e.target as HTMLInputElement
  emit('change', target.value, props.endDate)
}

function handleEndDateChange(e: Event) {
  const target = e.target as HTMLInputElement
  emit('change', props.startDate, target.value)
}
</script>

<template>
  <div class="flex flex-wrap gap-2 items-center">
    <div class="flex flex-col">
      <label class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Start Date</label>
      <input
        :value="startDate"
        type="date"
        @input="handleStartDateChange"
        class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
      />
    </div>
    <div class="flex flex-col">
      <label class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">End Date</label>
      <input
        :value="endDate"
        type="date"
        @input="handleEndDateChange"
        class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
      />
    </div>
    <div class="relative self-end">
      <button
        @click="showPresets = !showPresets"
        class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
      >
        Presets ▼
      </button>
      <div
        v-if="showPresets"
        class="absolute right-0 mt-1 w-40 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md shadow-lg z-10"
      >
        <div
          v-for="preset in presets"
          :key="preset.label"
          @click="handlePresetClick(preset.days)"
          class="px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer"
        >
          {{ preset.label }}
        </div>
      </div>
    </div>
  </div>
</template>
