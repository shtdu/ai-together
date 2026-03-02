<template>
  <section class="interaction-section">
    <h3>{{ t('components.reports.interaction.title') }}</h3>

    <div class="interaction-grid">
      <!-- Left: Donut Chart -->
      <div class="interaction-chart">
        <Doughnut :data="chartData" :options="chartOptions" />
        <div class="chart-center-text">
          <span class="total-count">{{ totalInteractions }}</span>
          <span class="total-label">{{ t('components.reports.interaction.events') }}</span>
        </div>
      </div>

      <!-- Right: Metrics Grid -->
      <div class="metrics-grid">
        <div class="metric-row">
          <span class="metric-label">{{ t('components.reports.interaction.avgLength') }}</span>
          <span class="metric-value">{{ avgLength }} {{ t('components.reports.interaction.chars') }}</span>
        </div>
        <div class="metric-row">
          <span class="metric-label">{{ t('components.reports.interaction.maxLength') }}</span>
          <span class="metric-value">{{ maxLength }} {{ t('components.reports.interaction.chars') }}</span>
        </div>
        <div class="metric-row">
          <span class="metric-label">{{ t('components.reports.interaction.minLength') }}</span>
          <span class="metric-value">{{ minLength }} {{ t('components.reports.interaction.chars') }}</span>
        </div>
        <div class="metric-row">
          <span class="metric-label">{{ t('components.reports.interaction.permissionRatio') }}</span>
          <span class="metric-value">{{ permissionRatio }}</span>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Doughnut } from 'vue-chartjs'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import type { ToolReport } from '../../services/reports'

ChartJS.register(ArcElement, Tooltip, Legend)

const props = defineProps<{
  data: ToolReport
}>()

const { t } = useI18n()

// Event type counts
const eventTypeCounts = computed(() => props.data.details.event_type_counts)

// Total interactions (only interactive events, excluding notifications)
const totalInteractions = computed(() => {
  const events = eventTypeCounts.value
  return (events.PermissionRequest || 0) +
         (events.UserPromptSubmit || 0)
})

// Prompt statistics
const promptStats = computed(() => props.data.details.prompt_stats)

const avgLength = computed(() => Math.round(promptStats.value.average_prompt_length || 0))
const maxLength = computed(() => promptStats.value.longest_prompt || 0)
const minLength = computed(() => promptStats.value.shortest_prompt || 0)

// Permission ratio
const permissionRatio = computed(() => {
  const events = eventTypeCounts.value
  const permissions = events.PermissionRequest || 0
  const prompts = events.UserPromptSubmit || 0

  if (prompts === 0) return '—'
  return (permissions / prompts).toFixed(2)
})

// Chart data
const chartData = computed(() => {
  const events = eventTypeCounts.value

  return {
    labels: [
      t('components.reports.interaction.permissions'),
      t('components.reports.interaction.prompts')
    ],
    datasets: [{
      data: [
        events.PermissionRequest || 0,
        events.UserPromptSubmit || 0
      ],
      backgroundColor: [
        'rgba(10, 132, 255, 0.7)',   // Blue for permissions
        'rgba(16, 185, 129, 0.7)'   // Green for prompts
      ],
      borderColor: [
        'rgba(10, 132, 255, 1)',
        'rgba(16, 185, 129, 1)'
      ],
      borderWidth: 2
    }]
  }
})

// Chart options
const chartOptions = {
  responsive: true,
  maintainAspectRatio: true,
  cutout: '70%',
  plugins: {
    legend: {
      display: true,
      position: 'bottom' as const,
      labels: {
        padding: 15,
        usePointStyle: true,
        pointStyle: 'circle',
        font: {
          size: 11,
          family: '-apple-system, BlinkMacSystemFont, "SF Pro Text"'
        },
        color: 'rgba(15, 23, 42, 0.75)',
        generateLabels: (chart: any) => {
          const data = chart.data
          if (data.labels.length === 0) return []

          const isDark = document.documentElement.classList.contains('dark')
          const textColor = isDark ? 'rgba(248, 250, 252, 0.75)' : 'rgba(15, 23, 42, 0.75)'

          return data.labels.map((label: string, i: number) => {
            const value = data.datasets[0].data[i]
            const total = data.datasets[0].data.reduce((a: number, b: number) => a + b, 0)
            const percentage = total > 0 ? ((value / total) * 100).toFixed(0) : 0

            return {
              text: `${label}: ${value} (${percentage}%)`,
              fillStyle: data.datasets[0].backgroundColor[i],
              hidden: false,
              index: i,
              color: textColor
            }
          })
        }
      }
    },
    tooltip: {
      backgroundColor: 'rgba(6, 11, 19, 0.96)',
      titleColor: '#f8fafc',
      bodyColor: 'rgba(248, 250, 252, 0.75)',
      borderColor: 'rgba(255, 255, 255, 0.08)',
      borderWidth: 1,
      padding: 12,
      cornerRadius: 12,
      displayColors: true,
      callbacks: {
        label: (context: any) => {
          const value = context.raw || 0
          const total = context.dataset.data.reduce((a: number, b: number) => a + b, 0)
          const percentage = total > 0 ? ((value / total) * 100).toFixed(0) : 0
          return `${context.label}: ${value} (${percentage}%)`
        }
      }
    }
  }
}
</script>

<style scoped>
.interaction-section {
  border: 1px solid var(--mac-border);
  border-radius: 20px;
  padding: 20px;
  background: var(--mac-surface);
  box-shadow: 0 15px 40px rgba(0, 0, 0, 0.08);
}

.interaction-section h3 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--mac-text);
  margin: 0 0 1rem 0;
}

.interaction-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2.5rem;
}

.interaction-chart {
  position: relative;
  width: 100%;
  max-width: 280px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.chart-center-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  pointer-events: none;
}

.total-count {
  display: block;
  font-size: 2rem;
  font-weight: 600;
  color: var(--mac-text);
  line-height: 1;
}

.total-label {
  display: block;
  font-size: 0.75rem;
  color: var(--mac-text-secondary);
  margin-top: 0.25rem;
}

.metrics-grid {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.metric-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background: rgba(255, 255, 255, 0.5);
  border: 1px solid rgba(15, 23, 42, 0.06);
  border-radius: 10px;
}

.metric-label {
  font-size: 0.85rem;
  color: var(--mac-text-secondary);
}

.metric-value {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--mac-text);
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'SF Mono', 'Monaco', monospace;
}

/* Dark Mode */
html.dark .interaction-section {
  background: rgba(15, 23, 42, 0.4);
  border-color: rgba(255, 255, 255, 0.1);
}

html.dark .metric-row {
  background: rgba(0, 0, 0, 0.2);
  border-color: rgba(255, 255, 255, 0.06);
}

html.dark .total-count {
  color: rgba(248, 250, 252, 0.95);
}

html.dark .total-label {
  color: rgba(186, 194, 210, 0.8);
}

/* Responsive */
@media (max-width: 768px) {
  .interaction-section {
    padding: 1rem;
  }

  .interaction-grid {
    grid-template-columns: 1fr;
    gap: 1.5rem;
  }

  .interaction-chart {
    max-width: 220px;
  }
}
</style>
