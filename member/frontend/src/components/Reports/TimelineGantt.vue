<template>
  <section class="timeline-section">
    <div class="section-header">
      <h3>{{ t('components.reports.timeline.title') }}</h3>
      <button
        class="ghost-icon filter-toggle"
        :data-tooltip="filterShort ? t('components.reports.timeline.showAllTooltip') : t('components.reports.timeline.filterTooltip')"
        @click="toggleFilter"
      >
        <span class="material-symbols-outlined">{{ filterShort ? 'filter_alt_off' : 'filter_list' }}</span>
      </button>
    </div>

    <div class="timeline-container">
      <!-- Hour markers -->
      <div class="timeline-markers">
        <span v-for="hour in hourMarkers" :key="hour" class="hour-mark">
          {{ String(hour).padStart(2, '0') }}
        </span>
      </div>

      <!-- Session blocks -->
      <div class="timeline-track" :style="{ height: `${trackHeight}px` }">
        <div
          v-for="block in filteredBlocks"
          :key="block.sessionId"
          class="time-block"
          :class="`duration-level-${block.durationLevel}`"
          :style="{
            left: `${block.leftPercent}%`,
            width: `${block.widthPercent}%`,
            top: `${block.lane * (LANE_HEIGHT + LANE_GAP)}px`,
            height: `${LANE_HEIGHT}px`
          }"
          @mouseenter="showTooltip(block, $event)"
          @mouseleave="hideTooltip"
        ></div>
      </div>

      <!-- Tooltip -->
      <div
        v-if="activeTooltip"
        class="contrib-tooltip"
        :style="{ left: `${tooltipX}px`, top: `${tooltipY}px` }"
      >
        <p class="tooltip-heading">{{ activeTooltip.sessionId }}</p>
        <ul class="tooltip-metrics">
          <li>
            <span class="metric-label">{{ t('components.reports.timeline.start') }}</span>
            <span class="metric-value">{{ activeTooltip.startTime }}</span>
          </li>
          <li>
            <span class="metric-label">{{ t('components.reports.timeline.duration') }}</span>
            <span class="metric-value">{{ activeTooltip.durationLabel }}</span>
          </li>
        </ul>
      </div>
    </div>

    <div v-if="deepWorkBlock" class="deep-work-highlight">
      <span class="material-symbols-outlined">psychology</span>
      <span>
        {{ t('components.reports.timeline.deepWork') }}: {{ deepWorkBlock.start }} - {{ deepWorkBlock.end }}
        ({{ deepWorkBlock.duration }})
      </span>
    </div>

    <div v-if="filteredBlocks.length === 0 && sessions.length > 0" class="no-sessions">
      <p>{{ t('components.reports.timeline.noSessions') }}</p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SessionInfo } from '../../services/reports'
import { formatDuration } from '../../services/reports'

// TimeBlock interface with lane property
interface TimeBlock {
  sessionId: string
  startHour: number
  durationHours: number
  leftPercent: number
  widthPercent: number
  lane: number
  intensity: 'high' | 'normal'
  startTime: string
  durationMs: number
  durationLabel: string
  durationLevel: DurationLevel
}

const props = defineProps<{
  sessions: SessionInfo[]
}>()

const { t } = useI18n()

// State
const filterShort = ref(true) // Filter sessions < 5 minutes by default
const activeTooltip = ref<TimeBlock | null>(null)
const tooltipX = ref(0)
const tooltipY = ref(0)

// Lane configuration constants
const LANE_HEIGHT = 12 // pixels per lane
const LANE_GAP = 4 // pixels between lanes

// Duration color levels (6 levels based on session length)
type DurationLevel = 1 | 2 | 3 | 4 | 5 | 6

function getDurationLevel(durationHours: number): DurationLevel {
  if (durationHours < 0.25) return 1  // < 15 min
  if (durationHours < 0.5) return 2   // < 30 min
  if (durationHours < 1) return 3     // < 1H
  if (durationHours < 2) return 4     // < 2H
  if (durationHours < 4) return 5     // < 4H
  return 6                            // >= 4H
}

// Show every 4th hour marker for cleaner display
const hourMarkers = computed(() => {
  const hours: number[] = []
  for (let i = 0; i <= 24; i += 4) {
    hours.push(i)
  }
  return hours
})

// Helper function to check if two sessions overlap
function sessionsOverlap(a: TimeBlock, b: TimeBlock): boolean {
  const aEnd = a.startHour + a.durationHours
  const bEnd = b.startHour + b.durationHours
  return a.startHour < bEnd && b.startHour < aEnd
}

// Lane assignment algorithm using greedy first-fit
function assignLanes(blocks: TimeBlock[]): TimeBlock[] {
  const lanes: TimeBlock[][] = []
  const sortedBlocks = [...blocks].sort((a, b) => a.startHour - b.startHour)

  for (const block of sortedBlocks) {
    let assignedLane = -1
    for (let laneIndex = 0; laneIndex < lanes.length; laneIndex++) {
      const lane = lanes[laneIndex]
      const canFitInLane = lane.every(existingBlock => !sessionsOverlap(existingBlock, block))

      if (canFitInLane) {
        assignedLane = laneIndex
        lane.push(block)
        break
      }
    }

    if (assignedLane === -1) {
      lanes.push([block])
      assignedLane = lanes.length - 1
    }

    block.lane = assignedLane
  }

  return sortedBlocks
}

// Calculate time blocks from sessions
const timeBlocks = computed(() => {
  return props.sessions
    .filter(s => s.duration_ms >= (filterShort.value ? 300000 : 0)) // Filter: >= 5 minutes if enabled
    .map(session => {
      const startDate = new Date(session.begin_time)
      const startHour = startDate.getHours() + startDate.getMinutes() / 60
      const durationHours = session.duration_ms / (1000 * 60 * 60)

      // Calculate position and width as percentages
      const leftPercent = (startHour / 24) * 100
      const widthPercent = (durationHours / 24) * 100

      // Format duration
      const durationLabel = formatDuration(session.duration_ms)

      // Format start time
      const startTime = startDate.toLocaleTimeString('en-US', {
        hour: '2-digit',
        minute: '2-digit',
        hour12: false
      })

      return {
        sessionId: session.session_id.substring(0, 8), // Short hash
        startHour,
        durationHours,
        leftPercent,
        widthPercent,
        intensity: (durationHours >= 1 ? 'high' : 'normal') as 'high' | 'normal',
        durationLabel: widthPercent > 3 ? durationLabel : '',
        startTime,
        durationMs: session.duration_ms,
        lane: 0, // Will be assigned by assignLanes
        durationLevel: getDurationLevel(durationHours)
      }
    })
    .sort((a, b) => a.startHour - b.startHour)
})

// Lane-assigned blocks (stacked)
const stackedBlocks = computed(() => {
  return assignLanes(timeBlocks.value)
})

// Dynamic track height based on number of lanes
const trackHeight = computed(() => {
  if (stackedBlocks.value.length === 0) return 60
  const maxLane = Math.max(...stackedBlocks.value.map(b => b.lane))
  return (maxLane + 1) * (LANE_HEIGHT + LANE_GAP)
})

// Filtered blocks (use stacked version)
const filteredBlocks = computed(() => stackedBlocks.value)

// Find longest session as "Deep Work Block"
const deepWorkBlock = computed(() => {
  const blocks = timeBlocks.value
  if (blocks.length === 0) return null

  const longest = blocks.reduce((max, block) =>
    block.durationHours > max.durationHours ? block : max
  )

  if (longest.durationHours < 1) return null // Only show if >= 1 hour

  const startDate = new Date()
  const [startHour, startMin] = longest.startTime.split(':').map(Number)
  startDate.setHours(startHour, startMin, 0)

  const endDate = new Date(startDate.getTime() + longest.durationMs)
  const endTime = endDate.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false
  })

  return {
    start: longest.startTime,
    end: endTime,
    duration: longest.durationLabel || formatDuration(longest.durationMs)
  }
})

// Toggle filter
const toggleFilter = () => {
  filterShort.value = !filterShort.value
}

// Show tooltip
const showTooltip = (block: any, event: MouseEvent) => {
  activeTooltip.value = block
  tooltipX.value = event.clientX + 10
  tooltipY.value = event.clientY - 10
}

// Hide tooltip
const hideTooltip = () => {
  activeTooltip.value = null
}
</script>

<style scoped>
.timeline-section {
  border: 1px solid var(--mac-border);
  border-radius: 20px;
  padding: 20px;
  background: var(--mac-surface);
  box-shadow: 0 15px 40px rgba(0, 0, 0, 0.08);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.section-header h3 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--mac-text);
  margin: 0;
}

.filter-toggle {
  flex-shrink: 0;
}

.filter-toggle .material-symbols-outlined {
  font-size: 20px;
}

.timeline-container {
  position: relative;
  min-height: 100px;
  max-height: 200px;
  height: auto;
  margin-top: 1rem;
}

.timeline-markers {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.5rem;
  padding: 0 0.5rem;
}

.hour-mark {
  font-size: 0.65rem;
  color: var(--mac-text-secondary);
  font-family: 'SF Mono', 'Monaco', monospace;
  opacity: 0.6;
}

.timeline-track {
  position: relative;
  background: linear-gradient(
    90deg,
    rgba(15, 23, 42, 0.02) 0%,
    rgba(15, 23, 42, 0.06) 50%,
    rgba(15, 23, 42, 0.02) 100%
  );
  border-radius: 12px;
  overflow: hidden;
}

.time-block {
  position: absolute;
  border-radius: 8px;
  border: 1px solid;
  transition: all 0.2s ease-out;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Duration level colors - light mode (distinct color gradient) */
.time-block.duration-level-1 {
  background: linear-gradient(135deg, #a5f3fc, #67e8f9);
  border-color: #22d3ee;
}

.time-block.duration-level-2 {
  background: linear-gradient(135deg, #67e8f9, #38bdf8);
  border-color: #0ea5e9;
}

.time-block.duration-level-3 {
  background: linear-gradient(135deg, #38bdf8, #3b82f6);
  border-color: #2563eb;
}

.time-block.duration-level-4 {
  background: linear-gradient(135deg, #3b82f6, #6366f1);
  border-color: #4f46e5;
}

.time-block.duration-level-5 {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border-color: #7c3aed;
}

.time-block.duration-level-6 {
  background: linear-gradient(135deg, #8b5cf6, #a855f7);
  border-color: #9333ea;
}

.time-block:hover {
  transform: scaleY(1.1);
  filter: brightness(1.1);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.contrib-tooltip {
  position: fixed;
  z-index: 1000;
  padding: 0.75rem;
  border-radius: 12px;
  background: var(--contrib-tooltip-bg, rgba(6, 11, 19, 0.96));
  border: 1px solid var(--contrib-tooltip-border, rgba(255, 255, 255, 0.08));
  box-shadow: var(--contrib-tooltip-shadow, 0 8px 24px rgba(0, 0, 0, 0.3));
  backdrop-filter: blur(12px);
  min-width: 180px;
  pointer-events: none;
}

.tooltip-heading {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--contrib-tooltip-heading, #f8fafc);
  margin: 0 0 0.5rem 0;
  font-family: 'SF Mono', 'Monaco', monospace;
}

.tooltip-metrics {
  list-style: none;
  padding: 0;
  margin: 0;
}

.tooltip-metrics li {
  display: flex;
  justify-content: space-between;
  font-size: 0.7rem;
  margin-bottom: 0.25rem;
}

.metric-label {
  color: var(--contrib-tooltip-text, rgba(248, 250, 252, 0.75));
}

.metric-value {
  color: var(--contrib-tooltip-value, #f8fafc);
  font-weight: 500;
}

.deep-work-highlight {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
  color: var(--mac-accent);
  padding: 0.75rem 1rem;
  background: rgba(10, 132, 255, 0.08);
  border-radius: 10px;
  border: 1px solid rgba(10, 132, 255, 0.15);
}

.deep-work-highlight .material-symbols-outlined {
  font-size: 1.2rem;
}

.no-sessions {
  margin-top: 1rem;
  text-align: center;
  color: var(--mac-text-secondary);
  font-size: 0.9rem;
}

/* Dark Mode */
html.dark .timeline-section {
  background: rgba(15, 23, 42, 0.4);
  border-color: rgba(255, 255, 255, 0.1);
}

html.dark .timeline-track {
  background: linear-gradient(
    90deg,
    rgba(255, 255, 255, 0.02) 0%,
    rgba(255, 255, 255, 0.05) 50%,
    rgba(255, 255, 255, 0.02) 100%
  );
}

/* Duration level colors - dark mode (distinct color gradient) */
html.dark .time-block.duration-level-1 {
  background: linear-gradient(135deg, #155e75, #0e7490);
  border-color: #06b6d4;
}

html.dark .time-block.duration-level-2 {
  background: linear-gradient(135deg, #0e7490, #0369a1);
  border-color: #0ea5e9;
}

html.dark .time-block.duration-level-3 {
  background: linear-gradient(135deg, #0369a1, #1d4ed8);
  border-color: #2563eb;
}

html.dark .time-block.duration-level-4 {
  background: linear-gradient(135deg, #1d4ed8, #4338ca);
  border-color: #4f46e5;
}

html.dark .time-block.duration-level-5 {
  background: linear-gradient(135deg, #4338ca, #6d28d9);
  border-color: #7c3aed;
}

html.dark .time-block.duration-level-6 {
  background: linear-gradient(135deg, #6d28d9, #7e22ce);
  border-color: #9333ea;
}

/* Responsive */
@media (max-width: 768px) {
  .timeline-section {
    padding: 1rem;
  }

  .hour-mark {
    font-size: 0.55rem;
  }
}
</style>
