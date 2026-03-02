<template>
  <div class="summary-cards-group">
    <h3 class="summary-cards-title">{{ t('components.reports.kpi.title') }}</h3>
    <article v-for="card in cards" :key="card.key" class="summary-card">
      <div class="summary-card__label">{{ card.label }}</div>
      <div class="summary-card__value">{{ card.value }}</div>
      <div class="summary-card__hint">{{ card.hint }}</div>
    </article>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ToolReport } from '../../services/reports'
import { formatDuration } from '../../services/reports'

const props = defineProps<{
  data: ToolReport
}>()

const { t } = useI18n()

const formatNumber = (value: number) => {
  return value.toLocaleString()
}

const cards = computed(() => {
  const details = props.data.details
  const eventTypeCounts = details.event_type_counts

  const postToolUse = eventTypeCounts.PostToolUse || 0
  const userPrompts = eventTypeCounts.UserPromptSubmit || 0
  const permissions = eventTypeCounts.PermissionRequest || 0

  // Calculate total session time
  const totalSessionTimeMs = details.session_stats.unique_sessions.reduce(
    (sum, session) => sum + session.duration_ms,
    0
  )

  // Calculate autonomy ratio
  let autonomyRatio = '—'
  let autonomyMode = ''
  if (userPrompts > 0) {
    autonomyRatio = `${(postToolUse / userPrompts).toFixed(1)}x`
    if (permissions > userPrompts) {
      autonomyMode = t('components.reports.kpi.semiAutonomous')
    } else if (permissions > userPrompts / 2) {
      autonomyMode = t('components.reports.kpi.guided')
    } else {
      autonomyMode = t('components.reports.kpi.autonomous')
    }
  }

  return [
    {
      key: 'sessions',
      label: t('components.reports.kpi.totalSessions'),
      value: formatNumber(details.session_stats.session_count),
      hint: t('components.reports.kpi.activeSessions', { count: details.session_stats.session_count }),
      insight: null
    },
    {
      key: 'actions',
      label: t('components.reports.kpi.agentActions'),
      value: formatNumber(postToolUse),
      hint: t('components.reports.kpi.actionsPerPrompt'),
      insight: postToolUse > 500 ? t('components.reports.kpi.highActivity') : null
    },
    {
      key: 'totalTime',
      label: t('components.reports.kpi.totalTime'),
      value: formatDuration(totalSessionTimeMs),
      hint: t('components.reports.kpi.totalTimeHint'),
      insight: null
    },
    {
      key: 'autonomy',
      label: t('components.reports.kpi.autonomyRatio'),
      value: autonomyRatio,
      hint: t('components.reports.kpi.autonomyHint'),
      insight: autonomyMode || null
    }
  ]
})
</script>

<style scoped>
.summary-cards-group {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(190px, 1fr));
  gap: 1rem;
  padding: 1.25rem;
  border-radius: 20px;
  background: rgba(248, 250, 252, 0.5);
  border: 1px solid rgba(15, 23, 42, 0.06);
  backdrop-filter: blur(8px);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.summary-card {
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 16px;
  padding: 1rem 1.25rem;
  background: radial-gradient(circle at top, rgba(148, 163, 184, 0.1), rgba(15, 23, 42, 0));
  backdrop-filter: blur(6px);
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.summary-card__label {
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #475569;
}

.summary-card__value {
  font-size: 1.85rem;
  font-weight: 600;
  color: #0f172a;
}

.summary-card__hint {
  font-size: 0.85rem;
  color: #94a3b8;
}

/* Dark Mode */
html.dark .summary-card {
  border-color: rgba(255, 255, 255, 0.12);
  background: radial-gradient(circle at top, rgba(148, 163, 184, 0.2), rgba(15, 23, 42, 0.35));
}

html.dark .summary-card__label {
  color: rgba(248, 250, 252, 0.75);
}

html.dark .summary-card__value {
  color: rgba(248, 250, 252, 0.95);
}

html.dark .summary-card__hint {
  color: rgba(186, 194, 210, 0.8);
}

html.dark .summary-cards-group {
  background: rgba(15, 23, 42, 0.4);
  border-color: rgba(255, 255, 255, 0.1);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

.summary-cards-title {
  grid-column: 1 / -1;
  font-size: 0.9rem;
  font-weight: 600;
  color: #334155;
  margin-bottom: 0.25rem;
}

html.dark .summary-cards-title {
  color: #e2e8f0;
}

/* Responsive */
@media (max-width: 768px) {
  .summary-cards-group {
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    padding: 1rem;
  }

  .summary-card__value {
    font-size: 1.5rem;
  }
}
</style>
