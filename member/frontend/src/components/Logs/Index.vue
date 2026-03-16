<template>
  <div class="main-shell logs-shell">
    <div class="global-actions">
      <p class="global-eyebrow">{{ t('components.logs.eyebrow') }}</p>
      <div class="refresh-indicator">
        <span>{{ t('components.logs.nextRefresh', { seconds: countdown }) }}</span>
        <BaseButton size="sm" :disabled="loading" @click="manualRefresh">
          {{ t('components.logs.refresh') }}
        </BaseButton>
      </div>
      <button class="ghost-icon" :aria-label="$t('components.logs.back')" @click="backToHome">
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path
            d="M15 18l-6-6 6-6"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
      </button>
    </div>

    <div class="logs-page">
      <section class="logs-summary" v-if="stats">
        <SummaryStats :stats="stats" />
      </section>

      <section class="logs-chart">
        <Line :data="chartData" :options="chartOptions" />
      </section>

      <form class="logs-filter-row" @submit.prevent="applyFilters">
        <div class="filter-fields">
          <label class="filter-field">
            <span>{{ t('components.logs.filters.platform') }}</span>
            <select v-model="filters.platform" class="mac-select">
              <option value="">{{ t('components.logs.filters.allPlatforms') }}</option>
              <option value="claude">Claude</option>
              <option value="opencode">OpenCode</option>
            </select>
          </label>
          <label class="filter-field">
            <span>{{ t('components.logs.filters.provider') }}</span>
            <select v-model="filters.provider" class="mac-select">
              <option value="">{{ t('components.logs.filters.allProviders') }}</option>
              <option v-for="provider in providerOptions" :key="provider" :value="provider">
                {{ provider }}
              </option>
            </select>
          </label>
        </div>
        <div class="filter-actions">
          <BaseButton type="submit" :disabled="loading">
            {{ t('components.logs.query') }}
          </BaseButton>
        </div>
      </form>

      <section class="logs-table-wrapper">
        <table class="logs-table">
          <thead>
            <tr>
              <th class="col-time">{{ t('components.logs.table.time') }}</th>
              <th class="col-platform">{{ t('components.logs.table.platform') }}</th>
              <th class="col-provider">{{ t('components.logs.table.provider') }}</th>
              <th class="col-model">{{ t('components.logs.table.model') }}</th>
              <th class="col-status">{{ t('components.logs.table.status') }}</th>
              <th class="col-duration">{{ t('components.logs.table.duration') }}</th>
              <th class="col-tokens">{{ t('components.logs.table.tokens') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in pagedLogs" :key="item.id">
              <td>{{ formatTime(item.created_at) }}</td>
              <td>{{ item.platform || '—' }}</td>
              <td>{{ item.provider || '—' }}</td>
              <td :title="item.model ?? ''">{{ formatModel(item.model) }}</td>
              <td class="status-cell">
                <span :class="['code', httpCodeClass(item.http_code)]">{{ item.http_code }}</span>
                <span :class="['stream-tag', item.is_stream ? 'on' : 'off']">{{ formatStream(item.is_stream) }}</span>
              </td>
              <td><span :class="['duration-tag', durationColor(item.duration_sec)]">{{ formatDuration(item.duration_sec) }}</span></td>
              <td>
                <span
                  class="price-tag"
                  @mouseenter="showLogTooltip(item, $event)"
                  @mousemove="showLogTooltip(item, $event)"
                  @mouseleave="hideLogTooltip"
                >
                  {{ formatCost(item) }}
                </span>
              </td>
            </tr>
            <tr v-if="!pagedLogs.length && !loading">
              <td colspan="7" class="empty">{{ t('components.logs.empty') }}</td>
            </tr>
          </tbody>
        </table>
        <div
          v-if="logTooltip.visible"
          ref="tooltipRef"
          class="contrib-tooltip"
          :class="logTooltip.placement"
          :style="{ left: `${logTooltip.left}px`, top: `${logTooltip.top}px` }"
        >
          <p class="tooltip-heading">{{ logTooltip.label }}</p>
          <ul class="tooltip-metrics">
            <li v-for="metric in logTooltipMetrics" :key="metric.key">
              <span class="metric-label">{{ metric.label }}</span>
              <span class="metric-value">{{ metric.value }}</span>
            </li>
          </ul>
        </div>
        <p v-if="loading" class="empty">{{ t('components.logs.loading') }}</p>
      </section>

      <div class="logs-pagination">
        <span>{{ page }} / {{ totalPages }}</span>
        <div class="pagination-actions">
          <BaseButton variant="outline" size="sm" :disabled="page === 1 || loading" @click="prevPage">
            ‹
          </BaseButton>
          <BaseButton variant="outline" size="sm" :disabled="page >= totalPages || loading" @click="nextPage">
            ›
          </BaseButton>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, onMounted, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import BaseButton from '../common/BaseButton.vue'
import {
  fetchRequestLogs,
  fetchEnabledProviders,
  fetchLogStats,
  type RequestLog,
  type LogStats,
  type LogStatsSeries,
} from '../../services/logs'
import {
  Chart,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'
import type { ChartOptions } from 'chart.js'
import { Line } from 'vue-chartjs'
import SummaryStats from '../Shared/SummaryStats.vue'

Chart.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend, Filler)

const { t } = useI18n()
const router = useRouter()

const logs = ref<RequestLog[]>([])
const stats = ref<LogStats | null>(null)
const loading = ref(false)
const filters = reactive({ platform: '', provider: '' })
const page = ref(1)
const PAGE_SIZE = 15
const providerOptions = ref<string[]>([])
const statsSeries = computed<LogStatsSeries[]>(() => stats.value?.series ?? [])

const isBrowser = typeof window !== 'undefined' && typeof document !== 'undefined'
const readDarkMode = () => (isBrowser ? document.documentElement.classList.contains('dark') : false)
const isDarkMode = ref(readDarkMode())
let themeObserver: MutationObserver | null = null

const getCssVarValue = (name: string, fallback: string) => {
  if (!isBrowser) return fallback
  const value = getComputedStyle(document.documentElement).getPropertyValue(name)
  return value?.trim() || fallback
}

const syncThemeState = () => {
  isDarkMode.value = readDarkMode()
}

const setupThemeObserver = () => {
  if (!isBrowser || themeObserver) return
  syncThemeState()
  themeObserver = new MutationObserver((mutations) => {
    if (mutations.some((mutation) => mutation.attributeName === 'class')) {
      syncThemeState()
    }
  })
  themeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class'],
  })
}

const teardownThemeObserver = () => {
  if (!themeObserver) return
  themeObserver.disconnect()
  themeObserver = null
}

const parseLogDate = (value?: string) => {
  if (!value) return null
  const normalize = value.replace(' ', 'T')
  const attempts = [value, `${normalize}`, `${normalize}Z`]
  for (const candidate of attempts) {
    const parsed = new Date(candidate)
    if (!Number.isNaN(parsed.getTime())) {
      return parsed
    }
  }
  const match = value.match(/^(\d{4}-\d{2}-\d{2}) (\d{2}:\d{2}:\d{2}) ([+-]\d{4}) UTC$/)
  if (match) {
    const [, day, time, zone] = match
    const zoneFormatted = `${zone.slice(0, 3)}:${zone.slice(3)}`
    const parsed = new Date(`${day}T${time}${zoneFormatted}`)
    if (!Number.isNaN(parsed.getTime())) {
      return parsed
    }
  }
  return null
}

const chartData = computed(() => {
  const series = statsSeries.value
  return {
    labels: series.map((item) => formatSeriesLabel(item.day)),
    datasets: [
      {
        label: t('components.logs.tokenLabels.cost'),
        data: series.map((item) => Number(((item.total_cost ?? 0)).toFixed(4))),
        borderColor: '#f97316',
        backgroundColor: 'rgba(249, 115, 22, 0.2)',
        tension: 0.3,
        fill: false,
        yAxisID: 'yCost',
      },
      {
        label: t('components.logs.tokenLabels.input'),
        data: series.map((item) => item.input_tokens ?? 0),
        borderColor: '#34d399',
        backgroundColor: 'rgba(52, 211, 153, 0.25)',
        tension: 0.35,
        fill: true,
      },
      {
        label: t('components.logs.tokenLabels.output'),
        data: series.map((item) => item.output_tokens ?? 0),
        borderColor: '#60a5fa',
        backgroundColor: 'rgba(96, 165, 250, 0.2)',
        tension: 0.35,
        fill: true,
      },
      {
        label: t('components.logs.tokenLabels.reasoning'),
        data: series.map((item) => item.reasoning_tokens ?? 0),
        borderColor: '#f472b6',
        backgroundColor: 'rgba(244, 114, 182, 0.2)',
        tension: 0.35,
        fill: true,
      },
      {
        label: t('components.logs.tokenLabels.cacheWrite'),
        data: series.map((item) => item.cache_create_tokens ?? 0),
        borderColor: '#fbbf24',
        backgroundColor: 'rgba(251, 191, 36, 0.2)',
        tension: 0.35,
        fill: false,
      },
      {
        label: t('components.logs.tokenLabels.cacheRead'),
        data: series.map((item) => item.cache_read_tokens ?? 0),
        borderColor: '#38bdf8',
        backgroundColor: 'rgba(56, 189, 248, 0.15)',
        tension: 0.35,
        fill: false,
      },
    ],
  }
})

const chartOptions = computed<ChartOptions<'line'>>(() => {
  const legendColor = getCssVarValue('--mac-text', isDarkMode.value ? '#f8fafc' : '#0f172a')
  const axisColor = getCssVarValue(
    '--mac-text-secondary',
    isDarkMode.value ? '#cbd5f5' : '#94a3b8',
  )
  const axisStrongColor = getCssVarValue('--mac-text', isDarkMode.value ? '#e2e8f0' : '#475569')
  const gridColor = isDarkMode.value ? 'rgba(148, 163, 184, 0.35)' : 'rgba(148, 163, 184, 0.2)'

  return {
    responsive: true,
    maintainAspectRatio: false,
    interaction: {
      mode: 'index',
      intersect: false,
    },
    plugins: {
      legend: {
        labels: {
          color: legendColor,
          font: {
            size: 12,
            weight: 500,
          },
        },
      },
    },
    scales: {
      x: {
        grid: { display: false },
        ticks: { color: axisColor },
      },
      y: {
        beginAtZero: true,
        ticks: { color: axisColor },
        grid: { color: gridColor },
      },
      yCost: {
        position: 'right',
        beginAtZero: true,
        grid: { drawOnChartArea: false },
        ticks: {
          color: axisStrongColor,
          callback: (value: string | number) => {
            const numeric = typeof value === 'number' ? value : Number(value)
            if (Number.isNaN(numeric)) return '$0'
            if (numeric >= 1) return `$${numeric.toFixed(2)}`
            return `$${numeric.toFixed(4)}`
          },
        },
      },
    },
  }
})
const formatSeriesLabel = (value?: string) => {
  if (!value) return ''
  const parsed = parseLogDate(value)
  if (parsed) {
    return `${padHour(parsed.getHours())}:00`
  }
  const match = value.match(/(\d{2}):(\d{2})/)
  if (match) {
    return `${match[1]}:${match[2]}`
  }
  return value
}

const tooltipRef = ref<HTMLElement | null>(null)

type TooltipPlacement = 'above' | 'below'

const logTooltip = reactive({
  visible: false,
  label: '',
  left: 0,
  top: 0,
  placement: 'above' as TooltipPlacement,
  hasPricing: true,
  cost: 0,
  inputTokens: 0,
  outputTokens: 0,
  reasoningTokens: 0,
  cacheCreateTokens: 0,
  cacheReadTokens: 0,
})

const formatMetric = (value: number) => value.toLocaleString()

const formatCost = (item: RequestLog) => {
  if (item.has_pricing === false) {
    return '0'
  }
  return formatCurrency(item.total_cost ?? 0)
}

const logTooltipMetrics = computed(() => [
  {
    key: 'cost',
    label: t('components.logs.tokenLabels.cost'),
    value: logTooltip.hasPricing ? formatCurrency(logTooltip.cost) : '—',
  },
  {
    key: 'inputTokens',
    label: t('components.logs.tokenLabels.input'),
    value: formatMetric(logTooltip.inputTokens),
  },
  {
    key: 'outputTokens',
    label: t('components.logs.tokenLabels.output'),
    value: formatMetric(logTooltip.outputTokens),
  },
  {
    key: 'reasoningTokens',
    label: t('components.logs.tokenLabels.reasoning'),
    value: formatMetric(logTooltip.reasoningTokens),
  },
  {
    key: 'cacheWrite',
    label: t('components.logs.tokenLabels.cacheWrite'),
    value: formatMetric(logTooltip.cacheCreateTokens),
  },
  {
    key: 'cacheRead',
    label: t('components.logs.tokenLabels.cacheRead'),
    value: formatMetric(logTooltip.cacheReadTokens),
  },
])

const clamp = (value: number, min: number, max: number) => {
  if (max <= min) return min
  return Math.min(Math.max(value, min), max)
}

const TOOLTIP_DEFAULT_WIDTH = 220
const TOOLTIP_DEFAULT_HEIGHT = 150
const TOOLTIP_VERTICAL_OFFSET = 12
const TOOLTIP_HORIZONTAL_MARGIN = 20
const TOOLTIP_VERTICAL_MARGIN = 24

const getTooltipSize = () => {
  const rect = tooltipRef.value?.getBoundingClientRect()
  return {
    width: rect?.width ?? TOOLTIP_DEFAULT_WIDTH,
    height: rect?.height ?? TOOLTIP_DEFAULT_HEIGHT,
  }
}

const viewportSize = () => {
  if (typeof window !== 'undefined') {
    return { width: window.innerWidth, height: window.innerHeight }
  }
  if (typeof document !== 'undefined' && document.documentElement) {
    return {
      width: document.documentElement.clientWidth,
      height: document.documentElement.clientHeight,
    }
  }
  return { width: 0, height: 0 }
}

const showLogTooltip = (log: RequestLog, event: MouseEvent) => {
  const target = event.currentTarget as HTMLElement | null
  const cellRect = target?.getBoundingClientRect()
  if (!cellRect) return
  logTooltip.label = formatTime(log.created_at)
  logTooltip.hasPricing = log.has_pricing ?? true
  logTooltip.cost = log.total_cost ?? 0
  logTooltip.inputTokens = log.input_tokens ?? 0
  logTooltip.outputTokens = log.output_tokens ?? 0
  logTooltip.reasoningTokens = log.reasoning_tokens ?? 0
  logTooltip.cacheCreateTokens = log.cache_create_tokens ?? 0
  logTooltip.cacheReadTokens = log.cache_read_tokens ?? 0
  const { width: tooltipWidth, height: tooltipHeight } = getTooltipSize()
  const { width: viewportWidth, height: viewportHeight } = viewportSize()
  const centerX = cellRect.left + cellRect.width / 2
  const halfWidth = tooltipWidth / 2
  const minLeft = TOOLTIP_HORIZONTAL_MARGIN + halfWidth
  const maxLeft =
    viewportWidth > 0 ? viewportWidth - halfWidth - TOOLTIP_HORIZONTAL_MARGIN : centerX
  logTooltip.left = clamp(centerX, minLeft, maxLeft)

  const anchorTop = cellRect.top
  const anchorBottom = cellRect.bottom
  const canShowAbove = anchorTop - tooltipHeight - TOOLTIP_VERTICAL_OFFSET >= TOOLTIP_VERTICAL_MARGIN
  const viewportBottomLimit =
    viewportHeight > 0 ? viewportHeight - tooltipHeight - TOOLTIP_VERTICAL_MARGIN : anchorBottom
  const shouldPlaceBelow = !canShowAbove
  logTooltip.placement = shouldPlaceBelow ? 'below' : 'above'
  const desiredTop = shouldPlaceBelow
    ? anchorBottom + TOOLTIP_VERTICAL_OFFSET
    : anchorTop - tooltipHeight - TOOLTIP_VERTICAL_OFFSET
  logTooltip.top = clamp(desiredTop, TOOLTIP_VERTICAL_MARGIN, viewportBottomLimit)
  logTooltip.visible = true
}

const hideLogTooltip = () => {
  logTooltip.visible = false
}

const REFRESH_INTERVAL = 30
const countdown = ref(REFRESH_INTERVAL)
let timer: number | undefined

const resetTimer = () => {
  countdown.value = REFRESH_INTERVAL
}

const startCountdown = () => {
  stopCountdown()
  timer = window.setInterval(() => {
    if (countdown.value <= 1) {
      countdown.value = REFRESH_INTERVAL
      void loadDashboard()
    } else {
      countdown.value -= 1
    }
  }, 1000)
}

const stopCountdown = () => {
  if (timer) {
    clearInterval(timer)
    timer = undefined
  }
}

const loadLogs = async () => {
  loading.value = true
  try {
    const data = await fetchRequestLogs({
      platform: filters.platform,
      provider: filters.provider,
      limit: 200,
    })
    logs.value = data ?? []
    page.value = Math.min(page.value, totalPages.value)
  } catch (error) {
    console.error('failed to load request logs', error)
  } finally {
    loading.value = false
  }
}

const loadStats = async () => {
  try {
    const data = await fetchLogStats(filters.platform)
    stats.value = data ?? null
  } catch (error) {
    console.error('failed to load log stats', error)
  }
}

const loadDashboard = async () => {
  await Promise.all([loadLogs(), loadStats()])
}

const pagedLogs = computed(() => {
  const start = (page.value - 1) * PAGE_SIZE
  return logs.value.slice(start, start + PAGE_SIZE)
})

const totalPages = computed(() => Math.max(1, Math.ceil(logs.value.length / PAGE_SIZE)))

const applyFilters = async () => {
  page.value = 1
  await loadDashboard()
  resetTimer()
}

const manualRefresh = () => {
  resetTimer()
  void loadDashboard()
}

const nextPage = () => {
  if (page.value < totalPages.value) {
    page.value += 1
  }
}

const prevPage = () => {
  if (page.value > 1) {
    page.value -= 1
  }
}

const backToHome = () => {
  router.push('/')
}

const padHour = (num: number) => num.toString().padStart(2, '0')

const formatTime = (value?: string) => {
  const date = parseLogDate(value)
  if (!date) return value || '—'
  return `${padHour(date.getMonth() + 1)}/${padHour(date.getDate())} ${padHour(date.getHours())}:${padHour(date.getMinutes())}:${padHour(date.getSeconds())}`
}

const formatStream = (value?: boolean | number) => {
  const isOn = value === true || value === 1
  return isOn ? t('components.logs.streamOn') : t('components.logs.streamOff')
}

const formatDuration = (value?: number) => {
  if (!value || Number.isNaN(value)) return '—'
  return `${value.toFixed(2)}s`
}

const formatModel = (value?: string) => {
  if (!value) return '—'
  if (value.length > 20) {
    return `${value.slice(0, 17)}...`
  }
  return value
}

const httpCodeClass = (code: number) => {
  if (code >= 500) return 'http-server-error'
  if (code >= 400) return 'http-client-error'
  if (code >= 300) return 'http-redirect'
  if (code >= 200) return 'http-success'
  return 'http-info'
}

const durationColor = (value?: number) => {
  if (!value || Number.isNaN(value)) return 'neutral'
  if (value < 2) return 'fast'
  if (value < 5) return 'medium'
  return 'slow'
}

const formatCurrency = (value?: number) => {
  if (value === undefined || value === null || Number.isNaN(value)) {
    return '$0.0000'
  }
  if (value >= 1) {
    return `$${value.toFixed(2)}`
  }
  if (value >= 0.01) {
    return `$${value.toFixed(3)}`
  }
  return `$${value.toFixed(4)}`
}



const loadProviderOptions = async () => {
  try {
    const list = await fetchEnabledProviders(filters.platform)
    providerOptions.value = list ?? []
    if (filters.provider && !providerOptions.value.includes(filters.provider)) {
      filters.provider = ''
    }
  } catch (error) {
    console.error('failed to load provider options', error)
  }
}

watch(
  () => filters.platform,
  async () => {
    await loadProviderOptions()
  },
)

onMounted(async () => {
  await Promise.all([loadDashboard(), loadProviderOptions()])
  startCountdown()
  setupThemeObserver()
})

onUnmounted(() => {
  stopCountdown()
  teardownThemeObserver()
})
</script>

<style scoped>


@media (max-width: 768px) {
}
</style>
