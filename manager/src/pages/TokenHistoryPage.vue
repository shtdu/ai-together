<script setup lang="ts">
import { ref, onMounted } from 'vue'
import dayjs from 'dayjs'
import DateRangePicker from '../components/common/DateRangePicker.vue'
import MultiSelect from '../components/common/MultiSelect.vue'
import UserSelect from '../components/common/UserSelect.vue'
import { analyticsApi, type HistoryResponse, type FilterOptions } from '../api/analytics'
import { exportCSV, exportJSON } from '../utils/export'

const startDate = ref(dayjs().subtract(7, 'day').format('YYYY-MM-DD'))
const endDate = ref(dayjs().format('YYYY-MM-DD'))
const selectedProviders = ref<string[]>([])
const selectedModels = ref<string[]>([])
const selectedTools = ref<string[]>([])
const selectedUsers = ref<number[]>([])

const paginationModel = ref({ page: 0, pageSize: 25 })
const sortModel = ref({ field: 'created_at', order: 'desc' as 'asc' | 'desc' })

const searchParams = ref({
  startDate: dayjs().subtract(7, 'day').format('YYYY-MM-DD'),
  endDate: dayjs().format('YYYY-MM-DD'),
  providers: [] as string[],
  models: [] as string[],
  tools: [] as string[],
  userIds: [] as number[],
})

const filterOptions = ref<FilterOptions | null>(null)
const data = ref<HistoryResponse | null>(null)
const isLoading = ref(false)
const filterLoading = ref(false)
const error = ref<string | null>(null)
const showExportMenu = ref(false)

function formatNumber(num: number): string {
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`
  if (num >= 1000) return `${(num / 1000).toFixed(1)}K`
  return num.toString()
}

function formatDuration(sec: number): string {
  if (sec >= 60) return `${(sec / 60).toFixed(1)}m`
  if (sec >= 1) return `${sec.toFixed(2)}s`
  return `${sec * 1000}ms`
}

function formatToolName(platform: string): string {
  // Convert platform to display name: claude -> Claude, codex -> Codex, opencode -> OpenCode
  if (!platform) return ''
  const lower = platform.toLowerCase()
  if (lower === 'claude') return 'Claude'
  if (lower === 'codex') return 'Codex'
  if (lower === 'opencode') return 'OpenCode'
  return platform
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
    data.value = await analyticsApi.getHistory({
      start_date: searchParams.value.startDate,
      end_date: searchParams.value.endDate,
      page: paginationModel.value.page + 1,
      limit: paginationModel.value.pageSize,
      providers: searchParams.value.providers.length > 0 ? searchParams.value.providers : undefined,
      models: searchParams.value.models.length > 0 ? searchParams.value.models : undefined,
      tools: searchParams.value.tools.length > 0 ? searchParams.value.tools : undefined,
      user_ids: searchParams.value.userIds.length > 0 ? searchParams.value.userIds : undefined,
      sort_by: sortModel.value.field,
      sort_order: sortModel.value.order,
    })
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load history'
  } finally {
    isLoading.value = false
  }
}

function handleDateChange(start: string, end: string) {
  startDate.value = start
  endDate.value = end
}

function handleSearch() {
  paginationModel.value.page = 0
  searchParams.value = {
    startDate: startDate.value,
    endDate: endDate.value,
    providers: selectedProviders.value,
    models: selectedModels.value,
    tools: selectedTools.value,
    userIds: selectedUsers.value,
  }
  fetchData()
}

function handleRefresh() {
  fetchData()
}

function handleSort(field: string) {
  if (sortModel.value.field === field) {
    sortModel.value.order = sortModel.value.order === 'asc' ? 'desc' : 'asc'
  } else {
    sortModel.value.field = field
    sortModel.value.order = 'desc'
  }
  fetchData()
}

function handlePageChange(newPage: number) {
  paginationModel.value.page = newPage
  fetchData()
}

function handlePageSizeChange(newSize: number) {
  paginationModel.value.pageSize = newSize
  paginationModel.value.page = 0
  fetchData()
}

function handleExportCSV() {
  if (!data.value) return

  const headers = ['ID', 'User', 'Tool', 'Provider', 'Model', 'Input Tokens', 'Output Tokens', 'Duration', 'Est. Cost', 'Timestamp']
  const rows = data.value.records.map(r => [
    r.id,
    r.user_name,
    formatToolName(r.platform),
    r.provider,
    r.model,
    r.input_tokens,
    r.output_tokens,
    r.duration_sec,
    r.estimated_cost?.toFixed(4) || '0.0000',
    r.timestamp
  ])

  exportCSV(headers, rows, `token-history-${searchParams.value.startDate}-${searchParams.value.endDate}`)
}

function handleExportJSON() {
  if (!data.value) return

  exportJSON(data.value.records, `token-history-${searchParams.value.startDate}-${searchParams.value.endDate}`)
}

onMounted(() => {
  fetchFilterOptions()
  fetchData()
})
</script>

<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-3xl font-bold">Token History</h1>
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

    <!-- Data Table -->
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-900">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                ID
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                User
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                Tool
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                Provider
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                Model
              </th>
              <th
                @click="handleSort('input_tokens')"
                class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800"
              >
                Input
                <span v-if="sortModel.field === 'input_tokens'">{{ sortModel.order === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th
                @click="handleSort('output_tokens')"
                class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800"
              >
                Output
                <span v-if="sortModel.field === 'output_tokens'">{{ sortModel.order === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th
                @click="handleSort('duration_sec')"
                class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800"
              >
                Duration
                <span v-if="sortModel.field === 'duration_sec'">{{ sortModel.order === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                Cost
              </th>
              <th
                @click="handleSort('created_at')"
                class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800"
              >
                Created At
                <span v-if="sortModel.field === 'created_at'">{{ sortModel.order === 'asc' ? '↑' : '↓' }}</span>
              </th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-if="isLoading">
              <td colspan="10" class="px-6 py-20 text-center">
                <div class="animate-spin text-4xl inline-block">⟳</div>
              </td>
            </tr>
            <tr v-else-if="!data?.records || data.records.length === 0">
              <td colspan="10" class="px-6 py-20 text-center text-gray-500 dark:text-gray-400">
                No records found
              </td>
            </tr>
            <tr v-else v-for="record in data.records" :key="record.id">
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">{{ record.id }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">{{ record.user_name }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">{{ formatToolName(record.platform) }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">{{ record.provider }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">{{ record.model }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">{{ formatNumber(record.input_tokens) }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">{{ formatNumber(record.output_tokens) }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">{{ formatDuration(record.duration_sec) }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">${{ record.estimated_cost?.toFixed(4) || '0.0000' }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">{{ dayjs(record.timestamp).format('YYYY-MM-DD HH:mm:ss') }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="data?.pagination" class="bg-gray-50 dark:bg-gray-900 px-6 py-3 flex items-center justify-between border-t border-gray-200 dark:border-gray-700">
        <div class="text-sm text-gray-700 dark:text-gray-300">
          Showing {{ ((data.pagination.page - 1) * data.pagination.limit) + 1 }} to {{ Math.min(data.pagination.page * data.pagination.limit, data.pagination.total_count) }} of {{ data.pagination.total_count.toLocaleString() }} results
        </div>
        <div class="flex items-center gap-2">
          <select
            :value="paginationModel.pageSize"
            @change="handlePageSizeChange(parseInt(($event.target as HTMLSelectElement).value))"
            class="px-2 py-1 border border-gray-300 dark:border-gray-600 rounded text-sm dark:bg-gray-800 dark:text-white"
          >
            <option :value="10">10 per page</option>
            <option :value="25">25 per page</option>
            <option :value="50">50 per page</option>
            <option :value="100">100 per page</option>
          </select>
          <div class="flex gap-1">
            <button
              @click="handlePageChange(paginationModel.page - 1)"
              :disabled="paginationModel.page === 0"
              class="px-3 py-1 border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-100 dark:hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed dark:text-white"
            >
              Previous
            </button>
            <span class="px-3 py-1 text-gray-700 dark:text-gray-300">
              Page {{ paginationModel.page + 1 }} of {{ data.pagination.total_pages }}
            </span>
            <button
              @click="handlePageChange(paginationModel.page + 1)"
              :disabled="paginationModel.page >= data.pagination.total_pages - 1"
              class="px-3 py-1 border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-100 dark:hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed dark:text-white"
            >
              Next
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
