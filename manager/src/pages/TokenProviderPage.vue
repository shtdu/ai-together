<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import dayjs from 'dayjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend
} from 'chart.js'
import { Bar, Pie } from 'vue-chartjs'
import DateRangePicker from '../components/common/DateRangePicker.vue'
import MultiSelect from '../components/common/MultiSelect.vue'
import { analyticsApi, type ProviderAnalyticsResponse, type FilterOptions } from '../api/analytics'
import { exportCSV, exportJSON } from '../utils/export'

ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend
)

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8', '#82CA9D', '#FF6B6B', '#4ECDC4']

const startDate = ref(dayjs().subtract(7, 'day').format('YYYY-MM-DD'))
const endDate = ref(dayjs().format('YYYY-MM-DD'))
const selectedProviders = ref<string[]>([])
const selectedModels = ref<string[]>([])
const selectedTools = ref<string[]>([])

const searchParams = ref({
  startDate: dayjs().subtract(7, 'day').format('YYYY-MM-DD'),
  endDate: dayjs().format('YYYY-MM-DD'),
  providers: [] as string[],
  models: [] as string[],
  tools: [] as string[],
})

// Helper function to format tool names for display
function formatToolName(platform: string): string {
  if (!platform) return ''
  const lower = platform.toLowerCase()
  if (lower === 'claude') return 'Claude'
  if (lower === 'codex') return 'Codex'
  if (lower === 'opencode') return 'OpenCode'
  return platform
}

const filterOptions = ref<FilterOptions | null>(null)
const data = ref<ProviderAnalyticsResponse | null>(null)
const isLoading = ref(false)
const filterLoading = ref(false)
const error = ref<string | null>(null)
const showExportMenu = ref(false)

function formatNumber(num: number): string {
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`
  if (num >= 1000) return `${(num / 1000).toFixed(1)}K`
  return num.toString()
}

function formatCost(cost: number): string {
  return `$${cost.toFixed(2)}`
}

async function fetchFilterOptions() {
  filterLoading.value = true
  try {
    filterOptions.value = await analyticsApi.getFilterOptions()
  } finally {
    filterLoading.value = false
  }
}

async function fetchData() {
  isLoading.value = true
  error.value = null
  try {
    data.value = await analyticsApi.getProviderAnalytics({
      start_date: searchParams.value.startDate,
      end_date: searchParams.value.endDate,
      providers: searchParams.value.providers.length > 0 ? searchParams.value.providers : undefined,
      models: searchParams.value.models.length > 0 ? searchParams.value.models : undefined,
      tools: searchParams.value.tools.length > 0 ? searchParams.value.tools : undefined,
    })
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load provider analytics'
  } finally {
    isLoading.value = false
  }
}

function handleDateChange(start: string, end: string) {
  startDate.value = start
  endDate.value = end
}

function handleSearch() {
  searchParams.value = {
    startDate: startDate.value,
    endDate: endDate.value,
    providers: selectedProviders.value,
    models: selectedModels.value,
    tools: selectedTools.value,
  }
  fetchData()
}

function handleRefresh() {
  fetchData()
}

function handleExportCSV() {
  if (!data.value?.distribution.by_provider) return

  const headers = ['Provider', 'Requests', 'Total Tokens', 'Cost', 'Percentage']
  const rows = data.value.distribution.by_provider.map(p => [
    p.name,
    p.requests,
    p.tokens,
    p.cost.toFixed(4),
    `${p.percentage.toFixed(2)}%`
  ])

  exportCSV(headers, rows, `provider-analytics-${searchParams.value.startDate}-${searchParams.value.endDate}`)
}

function handleExportJSON() {
  if (!data.value) return

  exportJSON(data.value.distribution.by_provider, `provider-analytics-${searchParams.value.startDate}-${searchParams.value.endDate}`)
}

// Chart data
const trendChartData = computed(() => {
  if (!data.value?.trend_data) return { labels: [], datasets: [] }

  const labels = data.value.trend_data.map(item => dayjs(item.date).format('MM/DD'))
  const providers = new Set<string>()
  data.value.trend_data.forEach(item => {
    Object.keys(item.by_provider).forEach(p => providers.add(p))
  })

  const datasets = Array.from(providers).map((provider, idx) => ({
    label: provider,
    data: data.value!.trend_data!.map(item => item.by_provider[provider] || 0),
    backgroundColor: COLORS[idx % COLORS.length],
    stack: 'Stack 0'
  }))

  return { labels, datasets }
})

const pieChartData = computed(() => {
  if (!data.value?.distribution.by_provider) return { labels: [], datasets: [] }
  return {
    labels: data.value.distribution.by_provider.map(p => p.name),
    datasets: [{
      data: data.value.distribution.by_provider.map(p => p.tokens),
      backgroundColor: COLORS.slice(0, data.value.distribution.by_provider.length)
    }]
  }
})

const toolPieChartData = computed(() => {
  if (!data.value?.distribution.by_tool) return { labels: [], datasets: [] }
  return {
    labels: data.value.distribution.by_tool.map(t => formatToolName(t.name)),
    datasets: [{
      data: data.value.distribution.by_tool.map(t => t.tokens),
      backgroundColor: COLORS.slice(0, data.value.distribution.by_tool.length)
    }]
  }
})

const modelPieChartData = computed(() => {
  if (!data.value?.distribution.by_model) return { labels: [], datasets: [] }
  return {
    labels: data.value.distribution.by_model.map(m => m.name),
    datasets: [{
      data: data.value.distribution.by_model.map(m => m.tokens),
      backgroundColor: COLORS.slice(0, data.value.distribution.by_model.length)
    }]
  }
})

const barOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { position: 'top' as const }
  },
  scales: {
    x: { stacked: true },
    y: { stacked: true }
  }
}

const pieOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false }
  }
}

onMounted(() => {
  fetchFilterOptions()
  fetchData()
})
</script>

<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-3xl font-bold">Provider Analytics</h1>
      <div class="flex gap-2">
        <button
          @click="handleRefresh"
          class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
        >
          ⟳ Refresh
        </button>
        <div class="relative">
          <button
            @click="showExportMenu = !showExportMenu"
            :disabled="!data || !data.distribution.by_provider"
            class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            📥 Export ▼
          </button>
          <div
            v-if="showExportMenu"
            class="absolute right-0 mt-1 w-40 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md shadow-lg z-10"
          >
            <button
              @click="handleExportCSV(); showExportMenu = false"
              class="w-full text-left px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-700"
            >
              CSV
            </button>
            <button
              @click="handleExportJSON(); showExportMenu = false"
              class="w-full text-left px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-700"
            >
              JSON
            </button>
          </div>
        </div>
      </div>
    </div>

    <div
      v-if="error"
      class="mb-4 p-4 rounded-lg bg-red-100 dark:bg-red-900 text-red-700 dark:text-red-200"
    >
      {{ error }}
      <button @click="error = null" class="float-right text-current opacity-70 hover:opacity-100">×</button>
    </div>

    <!-- Filters -->
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
      <div class="flex flex-wrap gap-4 items-center">
        <DateRangePicker
          :start-date="startDate"
          :end-date="endDate"
          @change="handleDateChange"
        />
        <MultiSelect
          label="Tools"
          :options="filterOptions?.tools || []"
          v-model="selectedTools"
          :labels="{ claude: 'Claude', codex: 'Codex', opencode: 'OpenCode' }"
        />
        <MultiSelect
          label="Providers"
          :options="filterOptions?.providers || []"
          v-model="selectedProviders"
        />
        <MultiSelect
          label="Models"
          :options="filterOptions?.models || []"
          v-model="selectedModels"
        />
        <button
          @click="handleSearch"
          :disabled="filterLoading"
          class="bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-medium py-2 px-4 rounded-md transition-colors"
        >
          🔍 Search
        </button>
      </div>
    </div>

    <div v-if="isLoading" class="flex justify-center items-center py-20">
      <div class="animate-spin text-4xl">⟳</div>
    </div>

    <div v-else-if="data">
      <!-- Summary Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <p class="text-sm text-gray-600 dark:text-gray-400">Total Requests</p>
          <p class="text-2xl font-bold">{{ formatNumber(data.summary.total_requests) }}</p>
        </div>
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <p class="text-sm text-gray-600 dark:text-gray-400">Total Tokens</p>
          <p class="text-2xl font-bold">{{ formatNumber(data.summary.total_tokens) }}</p>
        </div>
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <p class="text-sm text-gray-600 dark:text-gray-400">Input / Output</p>
          <p class="text-2xl font-bold">{{ formatNumber(data.summary.total_input) }} / {{ formatNumber(data.summary.total_output) }}</p>
        </div>
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <p class="text-sm text-gray-600 dark:text-gray-400">Total Cost</p>
          <p class="text-2xl font-bold">{{ formatCost(data.summary.total_cost) }}</p>
        </div>
      </div>

      <!-- Charts -->
      <div v-if="data.distribution.by_provider && data.distribution.by_provider.length > 0" class="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
        <!-- Daily Trend Chart -->
        <div class="lg:col-span-2 bg-white dark:bg-gray-800 rounded-lg shadow p-4">
          <h3 class="text-lg font-semibold mb-4">Daily Token Usage by Provider</h3>
          <div class="h-[330px]">
            <Bar :data="trendChartData" :options="barOptions" />
          </div>
        </div>

        <!-- Provider Distribution Pie Chart -->
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
          <h3 class="text-lg font-semibold mb-4">Provider Distribution</h3>
          <div class="h-[330px]">
            <Pie :data="pieChartData" :options="pieOptions" />
          </div>
        </div>
      </div>

      <div v-else class="mb-6 p-4 bg-blue-50 dark:bg-blue-900 text-blue-700 dark:text-blue-200 rounded-lg">
        No usage data available for the selected time period. Start using AI providers to see analytics here.
      </div>

      <!-- Provider Details Table -->
      <div v-if="data.distribution.by_provider && data.distribution.by_provider.length > 0" class="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
        <h3 class="text-lg font-semibold mb-4">Provider Details</h3>
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead class="bg-gray-50 dark:bg-gray-900">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Provider</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Requests</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Total Tokens</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Share</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Cost</th>
              </tr>
            </thead>
            <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
              <tr v-for="provider in data.distribution.by_provider" :key="provider.name">
                <td class="px-6 py-4 text-sm font-medium text-gray-900 dark:text-gray-100">{{ provider.name }}</td>
                <td class="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">{{ provider.requests.toLocaleString() }}</td>
                <td class="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">{{ formatNumber(provider.tokens) }}</td>
                <td class="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">{{ provider.percentage.toFixed(1) }}%</td>
                <td class="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">{{ formatCost(provider.cost) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Tool Distribution Section -->
      <div v-if="data.distribution.by_tool && data.distribution.by_tool.length > 0" class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Tool Distribution Pie Chart -->
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
          <h3 class="text-lg font-semibold mb-4">Tool Distribution</h3>
          <div class="h-[250px]">
            <Pie :data="toolPieChartData" :options="pieOptions" />
          </div>
        </div>

        <!-- Tool Details Table -->
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
          <h3 class="text-lg font-semibold mb-4">Tool Details</h3>
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
              <thead class="bg-gray-50 dark:bg-gray-900">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Tool</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Requests</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Total Tokens</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Share</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Cost</th>
                </tr>
              </thead>
              <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                <tr v-for="tool in data.distribution.by_tool" :key="tool.name">
                  <td class="px-6 py-4 text-sm font-medium text-gray-900 dark:text-gray-100">{{ formatToolName(tool.name) }}</td>
                  <td class="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">{{ tool.requests.toLocaleString() }}</td>
                  <td class="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">{{ formatNumber(tool.tokens) }}</td>
                  <td class="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">{{ tool.percentage.toFixed(1) }}%</td>
                  <td class="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">{{ formatCost(tool.cost) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Model Distribution Section -->
      <div v-if="data.distribution.by_model && data.distribution.by_model.length > 0" class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Model Distribution Pie Chart -->
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
          <h3 class="text-lg font-semibold mb-4">Model Distribution</h3>
          <div class="h-[250px]">
            <Pie :data="modelPieChartData" :options="pieOptions" />
          </div>
        </div>

        <!-- Model Details Table -->
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
          <h3 class="text-lg font-semibold mb-4">Model Details</h3>
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
              <thead class="bg-gray-50 dark:bg-gray-900">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Model</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Requests</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Total Tokens</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Share</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Cost</th>
                </tr>
              </thead>
              <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                <tr v-for="model in data.distribution.by_model" :key="model.name">
                  <td class="px-6 py-4 text-sm font-medium text-gray-900 dark:text-gray-100">{{ model.name }}</td>
                  <td class="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">{{ model.requests.toLocaleString() }}</td>
                  <td class="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">{{ formatNumber(model.tokens) }}</td>
                  <td class="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">{{ model.percentage.toFixed(1) }}%</td>
                  <td class="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">{{ formatCost(model.cost) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
