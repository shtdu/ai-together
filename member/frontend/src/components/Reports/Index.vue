<template>
  <div class="main-shell reports-shell">
    <div class="global-actions">
      <p class="global-eyebrow">{{ t('components.reports.eyebrow') }}</p>
      <div class="date-picker">
        <button
          class="nav-button"
          :aria-label="t('components.reports.previousDay')"
          :disabled="!canGoBack"
          @click="goToPreviousDay"
        >
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path
              d="M15 18l-6-6 6-6"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
        </button>
        <input
          v-model="selectedDate"
          type="date"
          class="mac-select"
          :min="minDate"
          :max="maxDate"
          @change="onDateChange"
        />
        <button
          class="nav-button"
          :aria-label="t('components.reports.nextDay')"
          :disabled="!canGoForward"
          @click="goToNextDay"
        >
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path
              d="M9 18l6-6-6-6"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
        </button>
      </div>
      <button class="ghost-icon" :aria-label="t('components.reports.back')" @click="backToHome">
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

    <div class="reports-page">
      <!-- Loading State -->
      <div v-if="loading" class="loading-state">
        <div class="spinner"></div>
        <p>{{ t('components.reports.loading') }}</p>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="error-state">
        <span class="material-symbols-outlined error-icon">error</span>
        <p>{{ error }}</p>
        <BaseButton @click="fetchReport">{{ t('components.reports.retry') }}</BaseButton>
      </div>

      <!-- Empty State -->
      <div v-else-if="!reportData || reportData.tools.length === 0" class="empty-state">
        <span class="material-symbols-outlined empty-icon">event_busy</span>
        <p>{{ t('components.reports.noData') }}</p>
        <p class="empty-hint">{{ t('components.reports.noDataHint') }}</p>
      </div>

      <!-- Report Content -->
      <div v-else class="report-content">
        <KPICards v-if="currentToolData" :data="currentToolData" />
        <TimelineGantt v-if="currentToolData" :sessions="currentToolData.details.session_stats.unique_sessions" />
        <InteractionSection v-if="currentToolData" :data="currentToolData" />
        <SessionTable v-if="currentToolData" :sessions="currentToolData.details.session_stats.unique_sessions" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import BaseButton from '../common/BaseButton.vue'
import KPICards from './KPICards.vue'
import TimelineGantt from './TimelineGantt.vue'
import InteractionSection from './InteractionSection.vue'
import SessionTable from './SessionTable.vue'
import { fetchDailyReport, getTodayDate, getPreviousDay, getNextDay, getMinDate, type DailyReport } from '../../services/reports'

const router = useRouter()
const { t } = useI18n()

// State
const loading = ref(false)
const error = ref<string | null>(null)
const reportData = ref<DailyReport | null>(null)
const selectedDate = ref(getTodayDate())
const minDate = ref(getMinDate()) // Calculated from RETENTION_DAYS constant
const maxDate = ref(getTodayDate())

// Debounce timer for date changes
let dateChangeTimer: ReturnType<typeof setTimeout> | null = null

onUnmounted(() => {
  if (dateChangeTimer !== null) {
    clearTimeout(dateChangeTimer)
    dateChangeTimer = null
  }
})

// Get the first tool's data (default to claude if available)
const currentToolData = computed(() => {
  if (!reportData.value || reportData.value.tools.length === 0) return null

  // Prefer claude data, otherwise use first tool
  const claudeTool = reportData.value.tools.find(t => t.tool_name === 'claude')
  return claudeTool || reportData.value.tools[0]
})

// Check if navigation is possible
const canGoBack = computed(() => {
  return selectedDate.value > minDate.value // Can go back if after RETENTION_DAYS limit
})

const canGoForward = computed(() => {
  return selectedDate.value < maxDate.value
})

// Fetch report data
const fetchReport = async () => {
  loading.value = true
  error.value = null

  try {
    const data = await fetchDailyReport({
      date: selectedDate.value,
      toolName: '' // Get all tools
    })
    reportData.value = data
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('components.reports.errorLoading')
    console.error('Failed to fetch report:', err)
  } finally {
    loading.value = false
  }
}

// Handle date change with debounce to prevent rapid API calls
const onDateChange = () => {
  // Clear any pending fetch
  if (dateChangeTimer) {
    clearTimeout(dateChangeTimer)
  }

  // Debounce fetch to 300ms
  dateChangeTimer = setTimeout(() => {
    fetchReport()
    dateChangeTimer = null
  }, 300)
}

// Navigate back to home
const backToHome = () => {
  router.push('/')
}

// Navigate to previous day
const goToPreviousDay = () => {
  selectedDate.value = getPreviousDay(selectedDate.value)
  fetchReport()
}

// Navigate to next day
const goToNextDay = () => {
  if (selectedDate.value < maxDate.value) {
    selectedDate.value = getNextDay(selectedDate.value)
    fetchReport()
  }
}

// Load report on mount
onMounted(() => {
  fetchReport()
})
</script>

<style scoped>
.reports-shell {
  min-height: 100vh;
  background: var(--app-background);
}

.global-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  background: rgba(248, 250, 252, 0.5);
  border-bottom: 1px solid rgba(15, 23, 42, 0.06);
  backdrop-filter: blur(8px);
}

html.dark .global-actions {
  background: rgba(15, 23, 42, 0.4);
  border-bottom-color: rgba(255, 255, 255, 0.1);
}

.global-eyebrow {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--mac-text-secondary);
  margin: 0;
}

.date-picker {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.date-picker input {
  border-radius: 8px;
  padding: 0.4rem 0.75rem;
  font-size: 0.85rem;
  border: 1px solid var(--mac-border);
  background: var(--mac-surface);
  color: var(--mac-text);
  min-width: 140px;
}

.nav-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: 1px solid var(--mac-border);
  border-radius: 6px;
  background: var(--mac-surface);
  color: var(--mac-text);
  cursor: pointer;
  transition: all 0.2s ease;
  padding: 0;
  margin: 0;
}

.nav-button:hover:not(:disabled) {
  background: var(--mac-hover);
  border-color: var(--mac-accent);
}

.nav-button:active:not(:disabled) {
  transform: scale(0.95);
}

.nav-button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.nav-button svg {
  width: 16px;
  height: 16px;
}

.reports-page {
  flex: 1;
  width: 100%;
  max-width: 1280px;
  margin: 0 auto;
  padding: 40px 48px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.loading-state,
.error-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  gap: 1rem;
  text-align: center;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--mac-border);
  border-top-color: var(--mac-accent);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-icon,
.empty-icon {
  font-size: 3rem;
  color: var(--mac-text-secondary);
}

.error-icon {
  color: #ef4444;
}

.error-state p,
.empty-state p {
  color: var(--mac-text);
  margin: 0;
}

.empty-hint {
  color: var(--mac-text-secondary);
  font-size: 0.9rem;
}

.report-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* Responsive */
@media (max-width: 768px) {
  .reports-page {
    padding: 24px;
  }

  .global-actions {
    padding: 0.75rem 1rem;
  }
}
</style>
