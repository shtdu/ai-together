<template>
  <div class="summary-cards-group">
    <h3 class="summary-cards-title">{{ t('components.logs.summary.stats24h') }}</h3>
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
import type { LogStats } from '../../services/logs'

const props = defineProps<{
  stats: LogStats | null
}>()

const { t } = useI18n()

const formatNumber = (value?: number) => {
  if (value === undefined || value === null) return '—'
  return value.toLocaleString()
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

const cards = computed(() => {
  const data = props.stats
  const totalTokens =
    (data?.input_tokens ?? 0) + (data?.output_tokens ?? 0) + (data?.reasoning_tokens ?? 0)
  return [
    {
      key: 'requests',
      label: t('components.logs.summary.total'),
      hint: t('components.logs.summary.requests'),
      value: data ? formatNumber(data.total_requests) : '—',
    },
    {
      key: 'tokens',
      label: t('components.logs.summary.tokens'),
      hint: t('components.logs.summary.tokenHint'),
      value: data ? formatNumber(totalTokens) : '—',
    },
    {
      key: 'cacheReads',
      label: t('components.logs.summary.cache'),
      hint: t('components.logs.summary.cacheHint'),
      value: data ? formatNumber(data.cache_read_tokens) : '—',
    },
    {
      key: 'cost',
      label: t('components.logs.tokenLabels.cost'),
      hint: t('components.logs.summary.last24HoursScope'),
      value: formatCurrency(data?.cost_total ?? 0),
    },
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

@media (max-width: 768px) {
  .summary-cards-group {
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    padding: 1rem;
  }
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
</style>
