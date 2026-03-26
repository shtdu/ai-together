<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line } from 'vue-chartjs'
import type { MetricDataPoint } from '../../types/api'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

interface Props {
  data: MetricDataPoint[]
  isLoading?: boolean
}

type ChartView = 'active_time' | 'tokens'

const props = defineProps<Props>()
const view = ref<ChartView>('active_time')

const chartData = computed(() => {
  return props.data.map(point => ({
    timestamp: new Date(point.timestamp).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric'
    }),
    fullDate: new Date(point.timestamp).toLocaleString(),
    activeTimeHours: point.active_time_seconds / 3600,
    totalTokens: point.total_tokens,
    inputTokens: point.input_tokens,
    outputTokens: point.output_tokens,
    requestCount: point.request_count
  }))
})

const activeTimeChartData = computed(() => ({
  labels: chartData.value.map(d => d.timestamp),
  datasets: [{
    label: 'AI Active Time (hours)',
    data: chartData.value.map(d => d.activeTimeHours),
    borderColor: '#1976d2',
    backgroundColor: 'rgba(25, 118, 210, 0.1)',
    tension: 0.4,
    fill: true
  }]
}))

const tokensChartData = computed(() => ({
  labels: chartData.value.map(d => d.timestamp),
  datasets: [
    {
      label: 'Input Tokens',
      data: chartData.value.map(d => d.inputTokens),
      borderColor: '#2196f3',
      backgroundColor: 'rgba(33, 150, 243, 0.1)',
      tension: 0.4
    },
    {
      label: 'Output Tokens',
      data: chartData.value.map(d => d.outputTokens),
      borderColor: '#4caf50',
      backgroundColor: 'rgba(76, 175, 80, 0.1)',
      tension: 0.4
    }
  ]
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: true
    },
    tooltip: {
      mode: 'index' as const,
      intersect: false
    }
  },
  scales: {
    y: {
      beginAtZero: true
    }
  }
}

function handleViewChange(newView: ChartView) {
  view.value = newView
}
</script>

<template>
  <div>
    <div class="flex justify-between items-center mb-4">
      <h3 class="text-lg font-semibold">
        {{ view === 'active_time' ? 'AI Active Time' : 'Token Volume' }}
      </h3>
      <div class="flex rounded-md shadow-sm">
        <button
          @click="handleViewChange('active_time')"
          class="px-3 py-1 text-sm font-medium rounded-l-md"
          :class="{
            'bg-blue-600 text-white': view === 'active_time',
            'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300': view !== 'active_time'
          }"
        >
          Active Time
        </button>
        <button
          @click="handleViewChange('tokens')"
          class="px-3 py-1 text-sm font-medium rounded-r-md"
          :class="{
            'bg-blue-600 text-white': view === 'tokens',
            'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300': view !== 'tokens'
          }"
        >
          Tokens
        </button>
      </div>
    </div>

    <div v-if="isLoading" class="h-[300px] flex items-center justify-center">
      <p class="text-gray-500 dark:text-gray-400">Loading...</p>
    </div>

    <div v-else-if="data.length === 0" class="h-[300px] flex items-center justify-center">
      <p class="text-gray-500 dark:text-gray-400">No data available</p>
    </div>

    <div v-else class="h-[250px]">
      <Line
        v-if="view === 'active_time'"
        :data="activeTimeChartData"
        :options="chartOptions"
      />
      <Line
        v-else
        :data="tokensChartData"
        :options="chartOptions"
      />
    </div>
  </div>
</template>
