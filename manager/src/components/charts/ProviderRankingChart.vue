<script setup lang="ts">
import { computed } from 'vue'
import {
  Chart as ChartJS,
  ArcElement,
  Tooltip,
  Legend
} from 'chart.js'
import { Pie } from 'vue-chartjs'
import type { ProviderRanking } from '../../types/api'

ChartJS.register(
  ArcElement,
  Tooltip,
  Legend
)

interface Props {
  rankings: ProviderRanking[]
  isLoading?: boolean
}

const props = defineProps<Props>()

const COLORS = ['#1976d2', '#2196f3', '#4caf50', '#ff9800', '#f44336', '#9c27b0', '#00bcd4', '#795548']

const chartData = computed(() => ({
  labels: props.rankings.map(r => r.provider),
  datasets: [{
    data: props.rankings.map(r => r.total_tokens),
    backgroundColor: COLORS.slice(0, props.rankings.length),
    borderWidth: 0
  }]
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false
    },
    tooltip: {
      callbacks: {
        label: (context: unknown) => {
          const ctx = context as { dataIndex: number }
          const idx = ctx.dataIndex
          const ranking = props.rankings[idx]
          return `${ranking.provider}: ${ranking.percentage}% (${ranking.total_tokens.toLocaleString()} tokens)`
        }
      }
    }
  }
}
</script>

<template>
  <div>
    <div v-if="isLoading" class="h-[250px] flex items-center justify-center">
      <p class="text-gray-500 dark:text-gray-400">Loading...</p>
    </div>

    <div v-else-if="rankings.length === 0" class="h-[250px] flex items-center justify-center">
      <p class="text-gray-500 dark:text-gray-400">No provider data available</p>
    </div>

    <!-- Show pie chart if <= 5 providers -->
    <div v-else-if="rankings.length <= 5">
      <h3 class="text-lg font-semibold mb-4">Provider Rankings (24h)</h3>
      <div class="h-[200px]">
        <Pie
          :data="chartData"
          :options="chartOptions"
        />
      </div>
    </div>

    <!-- Show list for more providers -->
    <div v-else>
      <h3 class="text-lg font-semibold mb-4">Provider Rankings (24h)</h3>
      <div class="max-h-[220px] overflow-y-auto">
        <div
          v-for="(ranking, index) in rankings"
          :key="ranking.provider"
          class="flex items-center justify-between py-2 px-3 border-b border-gray-200 dark:border-gray-700 last:border-0"
        >
          <div class="flex-1">
            <p class="font-medium">{{ ranking.rank }}. {{ ranking.provider }}</p>
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ ranking.total_tokens.toLocaleString() }} tokens</p>
          </div>
          <span
            class="px-2 py-1 text-xs font-semibold rounded-full text-white"
            :style="{ backgroundColor: COLORS[index % COLORS.length] }"
          >
            {{ ranking.percentage }}%
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
