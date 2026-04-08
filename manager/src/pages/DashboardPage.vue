<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import dayjs from 'dayjs'
import { useAuthStore } from '../stores/auth'
import { useLicenseStore } from '../stores/license'
import { dashboardApi } from '../api/dashboard'
import type { DashboardMetricsResponse, DashboardRankingsResponse, DashboardMembersResponse } from '../types/api'
import ActivityChart from '../components/charts/ActivityChart.vue'
import ProviderRankingChart from '../components/charts/ProviderRankingChart.vue'
import MemberStatsTable from '../components/charts/MemberStatsTable.vue'
import DateRangePicker from '../components/common/DateRangePicker.vue'

const authStore = useAuthStore()
const licenseStore = useLicenseStore()

type TimeRange = '24h' | '7d' | '30d' | 'custom'
const timeRange = ref<TimeRange>('7d')
const customStartDate = ref(dayjs().subtract(7, 'day').format('YYYY-MM-DD'))
const customEndDate = ref(dayjs().format('YYYY-MM-DD'))

const metricsData = ref<DashboardMetricsResponse | null>(null)
const metricsLoading = ref(false)

const rankingsData = ref<DashboardRankingsResponse | null>(null)
const rankingsLoading = ref(false)

const membersData = ref<DashboardMembersResponse | null>(null)
const membersLoading = ref(false)

const summary = computed(() => metricsData.value?.summary)

async function fetchMetrics() {
  metricsLoading.value = true
  try {
    metricsData.value = await dashboardApi.getMetrics(
      timeRange.value,
      'hour',
      timeRange.value === 'custom' ? customStartDate.value : undefined,
      timeRange.value === 'custom' ? customEndDate.value : undefined
    )
  } finally {
    metricsLoading.value = false
  }
}

async function fetchRankings() {
  if (!licenseStore.isCommercial) return
  rankingsLoading.value = true
  try {
    rankingsData.value = await dashboardApi.getRankings()
  } finally {
    rankingsLoading.value = false
  }
}

async function fetchMembers() {
  if (!authStore.isAdmin) return
  membersLoading.value = true
  try {
    membersData.value = await dashboardApi.getMembers()
  } finally {
    membersLoading.value = false
  }
}

function handleRefresh() {
  fetchMetrics()
  if (licenseStore.hasActiveLicense) {
    fetchRankings()
  }
  if (authStore.isAdmin) {
    fetchMembers()
  }
}

function handleTimeRangeChange(newRange: TimeRange) {
  timeRange.value = newRange
  fetchMetrics()
}

function handleCustomDateChange(start: string, end: string) {
  customStartDate.value = start
  customEndDate.value = end
  fetchMetrics()
}

function formatTokens(value: number): string {
  if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`
  if (value >= 1000) return `${(value / 1000).toFixed(0)}K`
  return value.toString()
}

onMounted(() => {
  fetchMetrics()
  if (licenseStore.hasActiveLicense) {
    fetchRankings()
  }
  if (authStore.isAdmin) {
    fetchMembers()
  }
})
</script>

<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-3xl font-bold">Dashboard</h1>
      <div class="flex gap-4 items-center">
        <div class="inline-flex rounded-md shadow-sm" role="group">
          <button
            v-for="range in ['24h', '7d', '30d', 'custom'] as TimeRange[]"
            :key="range"
            @click="handleTimeRangeChange(range)"
            class="px-4 py-2 text-sm font-medium"
            :class="{
              'bg-blue-600 text-white': timeRange === range,
              'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700': timeRange !== range
            }"
          >
            {{ range === '24h' ? '24h' : range === '7d' ? '7 Days' : range === '30d' ? '30 Days' : 'Custom' }}
          </button>
        </div>
        <button
          @click="handleRefresh"
          class="p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
          title="Refresh data"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Custom Date Range Picker -->
    <div v-if="timeRange === 'custom'" class="mb-6">
      <DateRangePicker
        :start-date="customStartDate"
        :end-date="customEndDate"
        @change="handleCustomDateChange"
      />
    </div>

    <!-- Summary Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <p class="text-sm text-gray-600 dark:text-gray-400">Total Requests</p>
        <p class="text-2xl font-bold">{{ summary?.total_requests?.toLocaleString() || '0' }}</p>
      </div>
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <p class="text-sm text-gray-600 dark:text-gray-400">Total Tokens</p>
        <p class="text-2xl font-bold">{{ summary?.total_tokens ? formatTokens(summary.total_tokens) : '0' }}</p>
      </div>
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <p class="text-sm text-gray-600 dark:text-gray-400">AI Active Time</p>
        <p class="text-2xl font-bold">{{ summary?.total_active_time_hours ? `${summary.total_active_time_hours.toFixed(1)}h` : '0h' }}</p>
      </div>
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <p class="text-sm text-gray-600 dark:text-gray-400">Est. Cost</p>
        <p class="text-2xl font-bold">${{ summary?.estimated_cost?.toFixed(2) || '0.00' }}</p>
      </div>
    </div>

    <!-- Charts Row -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
      <div class="lg:col-span-2 bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <ActivityChart
          :data="metricsData?.data_points || []"
          :is-loading="metricsLoading"
        />
      </div>
      <!-- Provider Rankings (Commercial only) -->
      <div v-if="licenseStore.hasActiveLicense" class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <ProviderRankingChart
          :rankings="rankingsData?.rankings || []"
          :is-loading="rankingsLoading"
        />
      </div>
    </div>

    <!-- Member Stats Table (Admin + Commercial only) -->
    <div v-if="authStore.isAdmin && licenseStore.hasActiveLicense" class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
      <MemberStatsTable
        :members="membersData?.members || []"
        :is-loading="membersLoading"
      />
    </div>
  </div>
</template>
