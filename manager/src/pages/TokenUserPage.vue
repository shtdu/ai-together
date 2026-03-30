<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import dayjs from 'dayjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend
} from 'chart.js'
import { Bar } from 'vue-chartjs'
import DateRangePicker from '../components/common/DateRangePicker.vue'
import MultiSelect from '../components/common/MultiSelect.vue'
import UserSelect from '../components/common/UserSelect.vue'
import { analyticsApi, type UserAnalyticsResponse, type FilterOptions } from '../api/analytics'
import { exportCSV, exportJSON } from '../utils/export'

ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend
)

const startDate = ref(dayjs().subtract(7, 'day').format('YYYY-MM-DD'))
const endDate = ref(dayjs().format('YYYY-MM-DD'))
const selectedUsers = ref<number[]>([])
const selectedProviders = ref<string[]>([])
const selectedTools = ref<string[]>([])

const searchParams = ref({
  startDate: dayjs().subtract(7, 'day').format('YYYY-MM-DD'),
  endDate: dayjs().format('YYYY-MM-DD'),
  userIds: [] as number[],
  providers: [] as string[],
  tools: [] as string[],
})

const filterOptions = ref<FilterOptions | null>(null)
const data = ref<UserAnalyticsResponse | null>(null)
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
  return `$${cost.toFixed(4)}`
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
    data.value = await analyticsApi.getUserAnalytics({
      start_date: searchParams.value.startDate,
      end_date: searchParams.value.endDate,
      user_ids: searchParams.value.userIds.length > 0 ? searchParams.value.userIds : undefined,
      providers: searchParams.value.providers.length > 0 ? searchParams.value.providers : undefined,
      tools: searchParams.value.tools.length > 0 ? searchParams.value.tools : undefined,
    })
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load user analytics'
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
    userIds: selectedUsers.value,
    providers: selectedProviders.value,
    tools: selectedTools.value,
  }
  fetchData()
}

function handleRefresh() {
  fetchData()
}

function handleExportCSV() {
  if (!data.value) return

  const headers = ['Rank', 'User', 'Total Tokens', 'Total Requests', 'Percentage', 'Estimated Cost']
  const rows = data.value.leaderboard.map(u => [
    u.rank,
    u.name,
    u.total_tokens,
    u.total_requests,
    `${u.percentage.toFixed(2)}%`,
    u.total_cost.toFixed(4)
  ])

  exportCSV(headers, rows, `user-analytics-${searchParams.value.startDate}-${searchParams.value.endDate}`)
}

function handleExportJSON() {
  if (!data.value) return

  exportJSON(data.value.leaderboard, `user-analytics-${searchParams.value.startDate}-${searchParams.value.endDate}`)
}

// Chart data (top 10 users)
const rankingChartData = computed(() => {
  if (!data.value?.leaderboard) return { labels: [], datasets: [] }

  const top10 = data.value.leaderboard.slice(0, 10)
  return {
    labels: top10.map(u => u.name),
    datasets: [{
      label: 'Tokens',
      data: top10.map(u => u.total_tokens),
      backgroundColor: '#8884d8'
    }]
  }
})

const horizontalBarOptions = {
  indexAxis: 'y' as const,
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
      <h1 class="text-3xl font-bold">User Analytics</h1>
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
            :disabled="!data"
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
        <UserSelect
          label="Users"
          :options="filterOptions?.users || []"
          v-model="selectedUsers"
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

    <div v-else-if="data" class="grid grid-cols-1 lg:grid-cols-5 gap-6 mb-6">
      <!-- User Ranking Chart -->
      <div class="lg:col-span-2 bg-white dark:bg-gray-800 rounded-lg shadow p-4">
        <h3 class="text-lg font-semibold mb-4">Top 10 Users by Token Usage</h3>
        <div class="h-[380px]">
          <Bar :data="rankingChartData" :options="horizontalBarOptions" />
        </div>
      </div>

      <!-- Leaderboard Table -->
      <div class="lg:col-span-3 bg-white dark:bg-gray-800 rounded-lg shadow p-4">
        <h3 class="text-lg font-semibold mb-4">User Leaderboard</h3>
        <div class="overflow-x-auto h-[380px]">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead class="bg-gray-50 dark:bg-gray-900 sticky top-0">
              <tr>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Rank</th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">User</th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Tokens</th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Requests</th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Share</th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Est. Cost</th>
              </tr>
            </thead>
            <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
              <tr v-for="user in data.leaderboard" :key="user.user_id">
                <td class="px-4 py-2 text-sm text-gray-900 dark:text-gray-100">{{ user.rank }}</td>
                <td class="px-4 py-2 text-sm font-medium text-gray-900 dark:text-gray-100">{{ user.name }}</td>
                <td class="px-4 py-2 text-sm text-gray-900 dark:text-gray-100">{{ formatNumber(user.total_tokens) }}</td>
                <td class="px-4 py-2 text-sm text-gray-900 dark:text-gray-100">{{ user.total_requests.toLocaleString() }}</td>
                <td class="px-4 py-2 text-sm text-gray-900 dark:text-gray-100">{{ user.percentage.toFixed(1) }}%</td>
                <td class="px-4 py-2 text-sm text-gray-900 dark:text-gray-100">{{ formatCost(user.total_cost) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Usage Details Table -->
      <div class="lg:col-span-5 bg-white dark:bg-gray-800 rounded-lg shadow p-4">
        <h3 class="text-lg font-semibold mb-4">Usage Details</h3>
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead class="bg-gray-50 dark:bg-gray-900">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">User</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Date</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Provider</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Model</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Tokens</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Avg Latency</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Cost</th>
              </tr>
            </thead>
            <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
              <tr v-for="(detail, i) in data.details" :key="i">
                <td class="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">{{ detail.user_name }}</td>
                <td class="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">{{ detail.date }}</td>
                <td class="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">{{ detail.provider }}</td>
                <td class="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">{{ detail.model }}</td>
                <td class="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">{{ formatNumber(detail.total_tokens) }}</td>
                <td class="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">{{ detail.avg_latency_ms.toFixed(0) }}ms</td>
                <td class="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">{{ formatCost(detail.total_cost) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
