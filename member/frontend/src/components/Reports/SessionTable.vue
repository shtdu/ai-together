<template>
  <section class="table-section">
    <div class="section-header">
      <h3>{{ t('components.reports.table.title') }}</h3>
      <div class="table-actions">
        <select v-model="sortBy" class="mac-select" size="sm">
          <option value="duration">{{ t('components.reports.table.sortDuration') }}</option>
          <option value="start">{{ t('components.reports.table.sortStart') }}</option>
        </select>
      </div>
    </div>

    <div class="logs-table-wrapper">
      <table class="logs-table">
        <thead>
          <tr>
            <th class="col-id">{{ t('components.reports.table.id') }}</th>
            <th class="col-start">{{ t('components.reports.table.start') }}</th>
            <th class="col-repo">{{ t('components.reports.table.repository') }}</th>
            <th class="col-prompts">{{ t('components.reports.table.prompts') }}</th>
            <th class="col-permissions">{{ t('components.reports.table.permissions') }}</th>
            <th class="col-duration">{{ t('components.reports.table.duration') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="session in sortedSessions" :key="session.sessionId">
            <td class="col-id">
              <code>{{ session.shortId }}</code>
            </td>
            <td class="col-start">
              <span class="time-value">{{ session.startTime }}</span>
            </td>
            <td class="col-repo">
              <span v-if="session.repositoryName" class="repo-badge">{{ session.repositoryName }}</span>
              <span v-else class="repo-empty">-</span>
            </td>
            <td class="col-prompts">
              <span :class="['event-count', session.prompts > 0 ? 'has-events' : 'no-events']">
                {{ session.prompts }}
              </span>
            </td>
            <td class="col-permissions">
              <span :class="['event-count', session.permissions > 0 ? 'has-events' : 'no-events']">
                {{ session.permissions }}
              </span>
            </td>
            <td class="col-duration">
              <span :class="['duration-badge', session.durationClass]">
                {{ session.durationLabel }}
              </span>
            </td>
          </tr>
          <tr v-if="sortedSessions.length === 0">
            <td colspan="6" class="empty">{{ t('components.reports.table.noSessions') }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SessionInfo } from '../../services/reports'
import { formatDuration } from '../../services/reports'

const props = defineProps<{
  sessions: SessionInfo[]
}>()

const { t } = useI18n()

const sortBy = ref<'duration' | 'start'>('duration')

const sessions = computed(() => {
  return props.sessions.map(session => {
    const durationMs = session.duration_ms

    const durationLabel = formatDuration(durationMs)

    const startDate = new Date(session.begin_time)
    const startTime = startDate.toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false
    })

    const hours = durationMs / (1000 * 60 * 60)

    // Extract event type counts
    const eventTypeCounts = session.event_type_counts || {}
    const prompts = eventTypeCounts['UserPromptSubmit'] || 0
    const permissions = eventTypeCounts['PermissionRequest'] || 0

    // Extract repository information
    const repositoryName = session.repository_name || ''

    return {
      sessionId: session.session_id,
      shortId: session.session_id.substring(0, 8) + '...',
      startTime,
      durationLabel,
      durationMs,
      durationClass: hours >= 1 ? 'long' : 'short',
      prompts,
      permissions,
      repositoryName
    }
  })
})

const sortedSessions = computed(() => {
  const sorted = [...sessions.value]

  switch (sortBy.value) {
    case 'duration':
      return sorted.sort((a, b) => b.durationMs - a.durationMs)
    case 'start':
      return sorted.sort((a, b) => a.startTime.localeCompare(b.startTime))
    default:
      return sorted
  }
})
</script>

<style scoped>
.table-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 4px;
}

.section-header h3 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--mac-text);
  margin: 0;
}

.mac-select {
  border-radius: 8px;
  border: 1px solid var(--mac-border);
  background: var(--mac-surface);
  padding: 0.4rem 0.75rem;
  font-size: 0.85rem;
  color: var(--mac-text);
}

/* Column widths */
.col-id {
  width: 200px;
}

.col-start {
  width: 150px;
}

.col-repo {
  width: 250px;
}

.col-duration {
  width: 120px;
}

.col-prompts {
  width: 100px;
}

.col-permissions {
  width: 120px;
}

/* Custom styles for code cells */
.col-id code {
  font-family: 'SF Mono', 'Monaco', monospace;
  font-size: 0.8rem;
  background: rgba(15, 23, 42, 0.05);
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
  color: var(--mac-accent);
}

html.dark .col-id code {
  background: rgba(255, 255, 255, 0.05);
}

.time-value {
  font-family: 'SF Mono', 'Monaco', monospace;
}

/* Duration badges */
.duration-badge {
  display: inline-block;
  padding: 0.25rem 0.625rem;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
  font-family: 'SF Mono', 'Monaco', monospace;
}

.duration-badge.long {
  background: rgba(10, 132, 255, 0.12);
  color: var(--mac-accent);
}

.duration-badge.short {
  background: rgba(148, 163, 184, 0.15);
  color: var(--mac-text-secondary);
}

html.dark .duration-badge.short {
  background: rgba(255, 255, 255, 0.08);
  color: rgba(248, 250, 252, 0.6);
}

/* Event count badges */
.event-count {
  display: inline-block;
  padding: 0.25rem 0.625rem;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
  font-family: 'SF Mono', 'Monaco', monospace;
  min-width: 40px;
  text-align: center;
}

.event-count.has-events {
  background: rgba(34, 197, 94, 0.12);
  color: #16a34a;
}

html.dark .event-count.has-events {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
}

.event-count.no-events {
  background: rgba(148, 163, 184, 0.1);
  color: var(--mac-text-secondary);
}

html.dark .event-count.no-events {
  background: rgba(255, 255, 255, 0.05);
  color: rgba(248, 250, 252, 0.4);
}

/* Repository badge */
.repo-badge {
  display: inline-block;
  padding: 0.25rem 0.625rem;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
  font-family: 'SF Mono', 'Monaco', monospace;
  background: rgba(99, 102, 241, 0.12);
  color: #6366f1;
}

html.dark .repo-badge {
  background: rgba(99, 102, 241, 0.2);
  color: #818cf8;
}

.repo-empty {
  display: inline-block;
  padding: 0.25rem 0.625rem;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
  font-family: 'SF Mono', 'Monaco', monospace;
  background: rgba(148, 163, 184, 0.1);
  color: var(--mac-text-secondary);
}

html.dark .repo-empty {
  background: rgba(255, 255, 255, 0.05);
  color: rgba(248, 250, 252, 0.4);
}

.logs-table .empty {
  text-align: center;
  padding: 20px;
  color: var(--mac-text-secondary);
}
</style>
