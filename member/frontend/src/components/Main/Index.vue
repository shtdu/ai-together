<template>
  <div class="main-shell">
    <div class="global-actions">
      <p class="global-eyebrow">{{ t('components.main.hero.eyebrow') }}</p>
      <button
        class="ghost-icon github-icon"
        :class="{ 'github-upgrade': hasUpdateAvailable }"
        :data-tooltip="hasUpdateAvailable ? t('components.main.controls.githubUpdate') : t('components.main.controls.github')"
        @click="openGitHub"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" fill="currentColor"/>
        </svg>
      </button>
      <button
        class="ghost-icon"
        :data-tooltip="t('components.main.controls.serverIntegration')"
        @click="goToServer"
      >
        <span class="material-symbols-outlined">cloud</span>
      </button>
      <button
        class="ghost-icon"
        :data-tooltip="t('components.main.controls.theme')"
        @click="toggleTheme"
      >
        <span v-if="themeIcon === 'sun'" class="material-symbols-outlined">light_mode</span>
        <span v-else class="material-symbols-outlined">dark_mode</span>
      </button>
      <button
        class="ghost-icon"
        :data-tooltip="t('components.main.controls.settings')"
        @click="goToSettings"
      >
        <span class="material-symbols-outlined">settings</span>
      </button>
    </div>
    <div class="contrib-page">
      <section class="contrib-hero">
        <h1 v-if="showHomeTitle">{{ t('components.main.hero.title') }}</h1>
        <!-- <p class="lead">
          {{ t('components.main.hero.lead') }}
        </p> -->
      </section>

      <section
        v-if="settingsLoaded && showHeatmap"
        ref="heatmapContainerRef"
        class="contrib-wall"
        :aria-label="t('components.main.heatmap.ariaLabel')"
      >
        <div class="contrib-legend">
          <span>{{ t('components.main.heatmap.legendLow') }}</span>
          <div class="legend-gradient-bar"></div>
          <span>{{ t('components.main.heatmap.legendHigh') }}</span>
        </div>

        <div class="contrib-grid">
          <div
            v-for="(week, weekIndex) in usageHeatmap"
            :key="weekIndex"
            class="contrib-column"
          >
            <div
              v-for="(day, dayIndex) in week"
              :key="dayIndex"
              class="contrib-cell"
              :class="intensityClass(day.intensity)"
              @mouseenter="showUsageTooltip(day, $event)"
              @mousemove="showUsageTooltip(day, $event)"
              @mouseleave="hideUsageTooltip"
            />
          </div>
        </div>
        <div
          v-if="usageTooltip.visible"
          ref="tooltipRef"
          class="contrib-tooltip"
          :class="usageTooltip.placement"
          :style="{ left: `${usageTooltip.left}px`, top: `${usageTooltip.top}px` }"
        >
          <p class="tooltip-heading">{{ formattedTooltipLabel }}</p>
          <ul class="tooltip-metrics">
            <li v-for="metric in usageTooltipMetrics" :key="metric.key">
              <span class="metric-label">{{ metric.label }}</span>
              <span class="metric-value">{{ metric.value }}</span>
            </li>
          </ul>
        </div>
      </section>

      <section class="logs-summary" v-if="settingsLoaded && show24hStats && stats24h" style="margin-bottom: 2rem;">
        <SummaryStats :stats="stats24h" />
      </section>

      <section class="automation-section">
      <div class="section-header">
        <div class="tab-group" role="tablist" :aria-label="t('components.main.tabs.ariaLabel')">
          <button
            v-for="(tab, idx) in visibleTabs"
            :key="tab.id"
            class="tab-pill"
            :class="{ active: selectedIndex === idx }"
            role="tab"
            :aria-selected="selectedIndex === idx"
            type="button"
            @click="onTabChange(idx)"
          >
            {{ t(`components.main.platforms.${tab.id}`) }}
          </button>
        </div>
        <div class="section-controls">
          <button
            class="ghost-icon disabled"
            :data-tooltip="t('components.main.controls.mcp')"
            disabled
          >
          <!-- TODO: Intended to disabled it by Gene // @click="goToMcp"  -->
            <span class="material-symbols-outlined">extension</span>
          </button>
          <button
            class="ghost-icon disabled"
            :data-tooltip="t('components.main.controls.skill')"
            disabled
          >
          <!-- TODO: Intended to disabled it by Gene // @click="goToSkill"  -->
            <span class="material-symbols-outlined">psychology</span>
          </button>
          <button
            class="ghost-icon"
            :data-tooltip="t('components.main.reports.view')"
            @click="goToReports"
          >
            <span class="material-symbols-outlined">analytics</span>
          </button>
          <button
            class="ghost-icon"
            :data-tooltip="t('components.main.logs.view')"
            @click="goToLogs"
          >
            <span class="material-symbols-outlined">history</span>
          </button>

        </div>
      </div>
      <div class="automation-list" @dragover.prevent>
        <article class="automation-card">
          <div class="card-leading">
            <div class="card-icon" style="background-color: var(--mac-surface-strong); color: var(--mac-text);">
              <span class="material-symbols-outlined">hub</span>
            </div>
            <div class="card-text">
              <div class="card-title-row">
                <p class="card-title">{{ currentProxyLabel }}</p>
              </div>
              <p class="card-metrics" style="color: var(--mac-text-secondary);">
                {{ t('components.main.relayToggle.tooltip') }}
              </p>
            </div>
          </div>
          <div class="card-actions">
             <label
               class="mac-switch sm"
               :class="{ disabled: relayToggleDisabled }"
             >
                <input
                  type="checkbox"
                  :checked="activeProxyState"
                  :disabled="relayToggleDisabled"
                  @click.prevent="onProxyToggle"
                />
                <span></span>
              </label>
              <div class="card-action-divider"></div>
              <button
                class="ghost-icon"
                :data-tooltip="canEditProviders ? t('components.main.tabs.addCard') : t('components.main.errors.noAddPermission')"
                :class="{ disabled: !canEditProviders }"
                @click="canEditProviders ? openCreateModal() : null"
              >
                <span class="material-symbols-outlined">add</span>
              </button>
              <button
                class="ghost-icon disabled"
                :data-tooltip="t('components.main.controls.launchTool')"
                @click="launchTool()"
              >
                <span class="material-symbols-outlined">rocket_launch</span>
              </button>
              
          </div>
        </article>
        <article
          v-for="card in activeCards"
          :key="card.id"
          :class="['automation-card', { dragging: draggingId === card.id }]"
          draggable="true"
          @dragstart="onDragStart(card.id)"
          @dragend="onDragEnd"
          @drop="onDrop(card.id)"
        >
          <div class="card-leading">
            <div class="card-icon" :style="{ backgroundColor: card.tint, color: card.accent }">
              <span class="icon-fallback">
                {{ vendorInitials(card.name) }}
              </span>
            </div>
            <div class="card-text">
              <div class="card-title-row">
                <p class="card-title">{{ card.name }}</p>
              </div>
              <!-- <p class="card-subtitle">{{ card.apiUrl }}</p> -->
              <p
                v-for="(stats, idx) in [providerStatDisplay(card.name)]"
                :key="`metrics-${card.id}-${idx}`"
                class="card-metrics"
              >
                <template v-if="stats.state !== 'ready'">
                  {{ stats.message }}
                </template>
                <template v-else>
                  <span
                    v-if="stats.successRateLabel"
                    class="card-success-rate"
                    :class="stats.successRateClass"
                  >
                    {{ stats.successRateLabel }}
                  </span>
                  <span class="card-metric-separator" aria-hidden="true">·</span>
                  <span >{{ stats.requests }}</span>
                  <span class="card-metric-separator" aria-hidden="true">·</span>
                  <span>{{ stats.tokens }}</span>
                  <span class="card-metric-separator" aria-hidden="true">·</span>
                  <span>{{ stats.cost }}</span>
                </template>
              </p>
            </div>
          </div>
          <div class="card-actions">
            <label
              class="mac-switch sm"
              :class="{ disabled: providerToggleDisabled }"
            >
              <input
                type="checkbox"
                v-model="card.enabled"
                :disabled="providerToggleDisabled"
                @change="persistProviders(activeTab)"
              />
              <span></span>
            </label>
            <div class="card-action-divider"></div>
            <button
              class="ghost-icon"
              :class="{ disabled: !canEditProviders }"
              :data-tooltip="canEditProviders ? t('components.main.form.labels.name') : t('components.main.errors.noEditPermission')"
              @click="canEditProviders ? configure(card) : null"
            >
              <span class="material-symbols-outlined">edit</span>
            </button>
            <button
              class="ghost-icon"
              :class="{ disabled: !canEditProviders }"
              :data-tooltip="canEditProviders ? t('components.main.form.actions.delete') : t('components.main.errors.noDeletePermission')"
              @click="canEditProviders ? requestRemove(card) : null"
            >
              <span class="material-symbols-outlined">delete</span>
            </button>
          </div>
        </article>
      </div>
      </section>

      <BaseModal
      :open="modalState.open"
      :title="modalState.editingId ? t('components.main.form.editTitle') : t('components.main.form.createTitle')"
      @close="closeModal"
    >
      <form class="vendor-form" @submit.prevent="submitModal">
                <!-- General error message -->
                <div v-if="modalState.errors.general" class="form-alert form-alert-error">
                  <svg viewBox="0 0 16 16" width="16" height="16" aria-hidden="true">
                    <path
                      d="M8 1a7 7 0 100 14A7 7 0 008 1zm0 13A6 6 0 118 2a6 6 0 010 12zm0-9a.75.75 0 01.75.75v4a.75.75 0 01-1.5 0v-4A.75.75 0 018 4.5zm0 7.5a1 1 0 100-2 1 1 0 000 2z"
                      fill="currentColor"
                    />
                  </svg>
                  <span>{{ modalState.errors.general }}</span>
                </div>

                <label class="form-field">
                  <span>{{ t('components.main.form.labels.name') }}</span>
                  <BaseInput
                    v-model="modalState.form.name"
                    type="text"
                    :placeholder="t('components.main.form.placeholders.name')"
                    required
                    :disabled="Boolean(modalState.editingId)"
                    :class="{
                      'has-error': !!modalState.errors.name,
                      'is-disabled': Boolean(modalState.editingId)
                    }"
                  />
                  <span v-if="modalState.editingId && !modalState.errors.name" class="field-hint">
                    {{ t('components.main.form.hints.nameCannotBeChanged') }}
                  </span>
                  <div v-if="modalState.errors.name" class="form-alert form-alert-error form-alert-field">
                    <svg viewBox="0 0 16 16" width="16" height="16" aria-hidden="true">
                      <path
                        d="M8 1a7 7 0 100 14A7 7 0 008 1zm0 13A6 6 0 118 2a6 6 0 010 12zm0-9a.75.75 0 01.75.75v4a.75.75 0 01-1.5 0v-4A.75.75 0 018 4.5zm0 7.5a1 1 0 100-2 1 1 0 000 2z"
                        fill="currentColor"
                      />
                    </svg>
                    <span>{{ modalState.errors.name }}</span>
                  </div>
                </label>

                <label class="form-field">
                  <span class="label-row">
                    {{ t('components.main.form.labels.apiUrl') }}
                    <span v-if="modalState.errors.apiUrl" class="field-error">
                      {{ modalState.errors.apiUrl }}
                    </span>
                  </span>
                  <BaseInput
                    v-model="modalState.form.apiUrl"
                    type="text"
                    :placeholder="t('components.main.form.placeholders.apiUrl')"
                    required
                    :class="{ 'has-error': !!modalState.errors.apiUrl }"
                  />
                </label>

                <label class="form-field">
                  <span>{{ t('components.main.form.labels.apiKey') }}</span>
                  <BaseInput
                    v-model="modalState.form.apiKey"
                    type="password"
                    :placeholder="t('components.main.form.placeholders.apiKey')"
                  />
                </label>

                <div class="form-field">
                  <ModelWhitelistEditor v-model="modalState.form.supportedModels" />
                </div>

                <div class="form-field">
                  <ModelMappingEditor v-model="modalState.form.modelMapping" />
                </div>

                <div class="form-footer-row">
                  <div class="form-field switch-field">
                    <span>{{ t('components.main.form.labels.enabled') }}</span>
                    <div class="switch-inline">
                      <label class="mac-switch">
                        <input type="checkbox" v-model="modalState.form.enabled" />
                        <span></span>
                      </label>
                      <span class="switch-text">
                        {{ modalState.form.enabled ? t('components.main.form.switch.on') : t('components.main.form.switch.off') }}
                      </span>
                    </div>
                  </div>

                  <footer class="form-actions">
                    <BaseButton variant="outline" type="button" @click="closeModal">
                      {{ t('components.main.form.actions.cancel') }}
                    </BaseButton>
                    <BaseButton type="submit">
                      {{ t('components.main.form.actions.save') }}
                    </BaseButton>
                  </footer>
                </div>
      </form>
      </BaseModal>
      <BaseModal
      :open="confirmState.open"
      :title="t('components.main.form.confirmDeleteTitle')"
      variant="confirm"
      @close="closeConfirm"
    >
      <div class="confirm-body">
        <p>
          {{ t('components.main.form.confirmDeleteMessage', { name: confirmState.card?.name ?? '' }) }}
        </p>
      </div>
      <footer class="form-actions confirm-actions">
        <BaseButton variant="outline" type="button" @click="closeConfirm">
          {{ t('components.main.form.actions.cancel') }}
        </BaseButton>
        <BaseButton variant="danger" type="button" @click="confirmRemove">
          {{ t('components.main.form.actions.delete') }}
        </BaseButton>
      </footer>
      </BaseModal>
      <!-- <footer v-if="appVersion" class="main-version">
        {{ t('components.main.versionLabel', { version: appVersion }) }}
      </footer> -->
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Browser } from '@wailsio/runtime'
import {
	buildUsageHeatmapMatrix,
	generateFallbackUsageHeatmap,
	DEFAULT_HEATMAP_DAYS,
	calculateHeatmapDayRange,
	type UsageHeatmapWeek,
	type UsageHeatmapDay,
} from '../../data/usageHeatmap'
import { automationCardGroups, createAutomationCards, type AutomationCard } from '../../data/cards'
import BaseButton from '../common/BaseButton.vue'
import BaseModal from '../common/BaseModal.vue'
import BaseInput from '../common/BaseInput.vue'
import ModelWhitelistEditor from '../common/ModelWhitelistEditor.vue'
import ModelMappingEditor from '../common/ModelMappingEditor.vue'
import SummaryStats from '../Shared/SummaryStats.vue'
import { LoadProviders, SaveProviders } from '../../../bindings/codeswitch/services/providerservice'
import { fetchProxyStatus, enableProxy, disableProxy } from '../../services/claudeSettings'
import { fetchHeatmapStats, fetchProviderDailyStats, fetchLogStats, type LogStats, type ProviderDailyStat } from '../../services/logs'
import { fetchCurrentVersion } from '../../services/version'
import { fetchAppSettings, type AppSettings } from '../../services/appSettings'
import { getCurrentTheme, setTheme, type ThemeMode } from '../../utils/ThemeManager'
import { useRouter } from 'vue-router'
import { isAuthenticated, getCurrentUser } from '../../services/auth'
import { canEditSettings } from '../../services/permission'
import { showToast } from '../../utils/toast'

const { t, locale } = useI18n()
const router = useRouter()
const themeMode = ref<ThemeMode>(getCurrentTheme())
const resolvedTheme = computed(() => {
  if (themeMode.value === 'systemdefault') {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }
  return themeMode.value
})
const themeIcon = computed(() => (resolvedTheme.value === 'dark' ? 'moon' : 'sun'))
const releasePageUrl = 'https://github.com/shtdu/together/releases'
const releaseApiUrl = 'https://api.github.com/repos/shtdu/together/releases/latest'

const HEATMAP_DAYS = DEFAULT_HEATMAP_DAYS
const usageHeatmap = ref<UsageHeatmapWeek[]>(generateFallbackUsageHeatmap(HEATMAP_DAYS))
const heatmapContainerRef = ref<HTMLElement | null>(null)
const tooltipRef = ref<HTMLElement | null>(null)
const proxyStates = reactive<Record<ProviderTab, boolean>>({
  claude: false,
  codex: false,
  opencode: false,
})
const proxyBusy = reactive<Record<ProviderTab, boolean>>({
  claude: false,
  codex: false,
  opencode: false,
})

const providerStatsMap = reactive<Record<ProviderTab, Record<string, ProviderDailyStat>>>({
  claude: {},
  codex: {},
  opencode: {},
} as Record<ProviderTab, Record<string, ProviderDailyStat>>)
const providerStatsLoading = reactive<Record<ProviderTab, boolean>>({
  claude: false,
  codex: false,
  opencode: false,
} as Record<ProviderTab, boolean>)
const providerStatsLoaded = reactive<Record<ProviderTab, boolean>>({
  claude: false,
  codex: false,
  opencode: false,
} as Record<ProviderTab, boolean>)
let providerStatsTimer: number | undefined
let updateTimer: number | undefined
let authStatusTimer: number | undefined
const settingsLoaded = ref(false)
const showHeatmap = ref(true)
const showHomeTitle = ref(true)
const show24hStats = ref(false)
const stats24h = ref<LogStats | null>(null)
const appVersion = ref('')
const hasUpdateAvailable = ref(false)
const isLoggedIn = ref(false)
const canEditProviders = ref(false)

const intensityClass = (value: number) => `gh-level-${value}`

type TooltipPlacement = 'above' | 'below'

const usageTooltip = reactive({
  visible: false,
  label: '',
  dateKey: '',
  left: 0,
  top: 0,
  placement: 'above' as TooltipPlacement,
  requests: 0,
  inputTokens: 0,
  outputTokens: 0,
  reasoningTokens: 0,
  cost: 0,
})

const formatMetric = (value: number) => value.toLocaleString()

const tooltipDateFormatter = computed(() =>
  new Intl.DateTimeFormat(locale.value || 'en', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
)

const currencyFormatter = computed(() =>
  new Intl.NumberFormat(locale.value || 'en', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
)

const formattedTooltipLabel = computed(() => {
  if (!usageTooltip.dateKey) return usageTooltip.label
  const date = new Date(usageTooltip.dateKey)
  if (Number.isNaN(date.getTime())) {
    return usageTooltip.label
  }
  return tooltipDateFormatter.value.format(date)
})

const formattedTooltipAmount = computed(() =>
  currencyFormatter.value.format(Math.max(usageTooltip.cost, 0))
)

const usageTooltipMetrics = computed(() => [
  {
    key: 'cost',
    label: t('components.main.heatmap.metrics.cost'),
    value: formattedTooltipAmount.value,
  },
  {
    key: 'requests',
    label: t('components.main.heatmap.metrics.requests'),
    value: formatMetric(usageTooltip.requests),
  },
  {
    key: 'inputTokens',
    label: t('components.main.heatmap.metrics.inputTokens'),
    value: formatMetric(usageTooltip.inputTokens),
  },
  {
    key: 'outputTokens',
    label: t('components.main.heatmap.metrics.outputTokens'),
    value: formatMetric(usageTooltip.outputTokens),
  },
  {
    key: 'reasoningTokens',
    label: t('components.main.heatmap.metrics.reasoningTokens'),
    value: formatMetric(usageTooltip.reasoningTokens),
  },
])

const clamp = (value: number, min: number, max: number) => {
  if (max <= min) return min
  return Math.min(Math.max(value, min), max)
}

const TOOLTIP_DEFAULT_WIDTH = 220
const TOOLTIP_DEFAULT_HEIGHT = 120
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
  return {
    width: heatmapContainerRef.value?.clientWidth ?? 0,
    height: heatmapContainerRef.value?.clientHeight ?? 0,
  }
}

const showUsageTooltip = (day: UsageHeatmapDay, event: MouseEvent) => {
  const target = event.currentTarget as HTMLElement | null
  const cellRect = target?.getBoundingClientRect()
  if (!cellRect) return
  usageTooltip.label = day.label
  usageTooltip.dateKey = day.dateKey
  usageTooltip.requests = day.requests
  usageTooltip.inputTokens = day.inputTokens
  usageTooltip.outputTokens = day.outputTokens
  usageTooltip.reasoningTokens = day.reasoningTokens
  usageTooltip.cost = day.cost
  const { width: tooltipWidth, height: tooltipHeight } = getTooltipSize()
  const { width: viewportWidth, height: viewportHeight } = viewportSize()
  const centerX = cellRect.left + cellRect.width / 2
  const halfWidth = tooltipWidth / 2
  const minLeft = TOOLTIP_HORIZONTAL_MARGIN + halfWidth
  const maxLeft = viewportWidth > 0 ? viewportWidth - halfWidth - TOOLTIP_HORIZONTAL_MARGIN : centerX
  usageTooltip.left = clamp(centerX, minLeft, maxLeft)

  const anchorTop = cellRect.top
  const anchorBottom = cellRect.bottom
  const canShowAbove = anchorTop - tooltipHeight - TOOLTIP_VERTICAL_OFFSET >= TOOLTIP_VERTICAL_MARGIN
  const viewportBottomLimit = viewportHeight > 0 ? viewportHeight - tooltipHeight - TOOLTIP_VERTICAL_MARGIN : anchorBottom
  const shouldPlaceBelow = !canShowAbove
  usageTooltip.placement = shouldPlaceBelow ? 'below' : 'above'
  const desiredTop = shouldPlaceBelow
    ? anchorBottom + TOOLTIP_VERTICAL_OFFSET
    : anchorTop - tooltipHeight - TOOLTIP_VERTICAL_OFFSET
  usageTooltip.top = clamp(desiredTop, TOOLTIP_VERTICAL_MARGIN, viewportBottomLimit)
  usageTooltip.visible = true
}

const hideUsageTooltip = () => {
  usageTooltip.visible = false
}

const loadAppSettings = async () => {
  try {
    const data: AppSettings = await fetchAppSettings()
    showHeatmap.value = data?.show_heatmap ?? true
    showHomeTitle.value = data?.show_home_title ?? true
    show24hStats.value = data?.show_24h_stats ?? false

    if (show24hStats.value) {
      void load24hStats()
    }
  } catch (error) {
    console.error('failed to load app settings', error)
    showHeatmap.value = true
    showHomeTitle.value = true
    show24hStats.value = false
  } finally {
    settingsLoaded.value = true
  }
}

const load24hStats = async () => {
    try {
        const data = await fetchLogStats('')
        stats24h.value = data ?? null
    } catch(err) {
        console.error('failed to load 24h stats', err)
    }
}

const checkForUpdates = async () => {
  try {
    const version = await fetchCurrentVersion()
    appVersion.value = version || ''
  } catch (error) {
    console.error('failed to load app version', error)
  }

  try {
    const resp = await fetch(releaseApiUrl, {
      headers: {
        Accept: 'application/vnd.github+json',
      },
    })
    if (!resp.ok) {
      return
    }
    const data = await resp.json()
    const latestTag = data?.tag_name ?? ''
    if (latestTag && compareVersions(appVersion.value || '0.0.0', latestTag) < 0) {
      hasUpdateAvailable.value = true
    }
  } catch (error) {
    console.error('failed to fetch release info', error)
  }
}

const handleAppSettingsUpdated = () => {
  void loadAppSettings()
}

const startUpdateTimer = () => {
  stopUpdateTimer()
  updateTimer = window.setInterval(() => {
    void checkForUpdates()
  }, 60 * 60 * 1000)
}

const stopUpdateTimer = () => {
  if (updateTimer) {
    clearInterval(updateTimer)
    updateTimer = undefined
  }
}

const loadAuthStatus = async () => {
  try {
    isLoggedIn.value = await isAuthenticated()
    if (isLoggedIn.value) {
      canEditProviders.value = await canEditSettings()
    } else {
      canEditProviders.value = false
    }
  } catch (error) {
    console.error('Failed to load auth status', error)
    isLoggedIn.value = false
    canEditProviders.value = false
  }
}

const normalizeProviderKey = (value: string) => value?.trim().toLowerCase() ?? ''

const normalizeVersion = (value: string) => value.replace(/^v/i, '').trim()

const compareVersions = (current: string, remote: string) => {
  const curParts = normalizeVersion(current).split('.').map((part) => parseInt(part, 10) || 0)
  const remoteParts = normalizeVersion(remote).split('.').map((part) => parseInt(part, 10) || 0)
  const maxLen = Math.max(curParts.length, remoteParts.length)
  for (let i = 0; i < maxLen; i++) {
    const cur = curParts[i] ?? 0
    const rem = remoteParts[i] ?? 0
    if (cur === rem) continue
    return cur < rem ? -1 : 1
  }
  return 0
}

const loadUsageHeatmap = async () => {
	try {
		const rangeDays = calculateHeatmapDayRange(HEATMAP_DAYS)
		const stats = await fetchHeatmapStats(rangeDays)
		usageHeatmap.value = buildUsageHeatmapMatrix(stats, HEATMAP_DAYS)
	} catch (error) {
		console.error('Failed to load usage heatmap stats', error)
	}
}

const tabs = [
  { id: 'opencode' },
  { id: 'claude' },
  { id: 'codex', hidden: true },
] as const
type ProviderTab = (typeof tabs)[number]['id']
const providerTabIds = tabs.map((tab) => tab.id) as ProviderTab[]

const cards = reactive<Record<ProviderTab, AutomationCard[]>>({
  claude: createAutomationCards(automationCardGroups.claude),
  codex: createAutomationCards(automationCardGroups.codex),
  opencode: createAutomationCards(automationCardGroups.opencode),
})
const draggingId = ref<number | null>(null)

// Ensure modelMapping and supportedModels are always serialized (even if empty)
// to prevent data loss during save/load cycles
const serializeProviders = (providers: AutomationCard[]) =>
  providers.map((provider) => ({
    ...provider,
    modelMapping: provider.modelMapping || {},
    supportedModels: provider.supportedModels || [],
  }))

const persistProviders = async (tabId: ProviderTab) => {
  try {
    await SaveProviders(tabId, serializeProviders(cards[tabId]))
  } catch (error) {
    console.error('Failed to save providers', error)
  }
}

const replaceProviders = (tabId: ProviderTab, data: AutomationCard[]) => {
  cards[tabId].splice(0, cards[tabId].length, ...createAutomationCards(data))
}

const loadProvidersFromDisk = async () => {
  for (const tab of providerTabIds) {
    try {
      const saved = await LoadProviders(tab)
      if (Array.isArray(saved)) {
        replaceProviders(tab, saved as AutomationCard[])
      } else {
        await persistProviders(tab)
      }
    } catch (error) {
      console.error('Failed to load providers', error)
    }
  }
}

const refreshProxyState = async (tab: ProviderTab) => {
  try {
    const status = await fetchProxyStatus(tab)
    proxyStates[tab] = Boolean(status?.enabled)
  } catch (error) {
    console.error(`Failed to fetch proxy status for ${tab}`, error)
    proxyStates[tab] = false
  }
}

const launchTool = () => {
  console.log('Launch tool clicked')
}

const onProxyToggle = async () => {
  const tab = activeTab.value
  if (proxyBusy[tab]) return
  proxyBusy[tab] = true
  const nextState = !proxyStates[tab]
  try {
    if (nextState) {
      // Validate: at least one enabled provider is required
      const hasEnabledProvider = cards[tab].some((p) => p.enabled)
      if (!hasEnabledProvider) {
        const toolName = t(`components.main.platforms.${tab}`)
        showToast(t('components.main.relayToggle.noProviderError', { tool: toolName }), 'error')
        return
      }
      await enableProxy(tab)
    } else {
      await disableProxy(tab)
    }
    proxyStates[tab] = nextState
  } catch (error) {
    console.error(`Failed to toggle proxy for ${tab}`, error)
    // Display error message to user
    const errorMessage = error instanceof Error ? error.message : String(error)
    showToast(errorMessage, 'error')
    // Re-fetch the actual proxy state to sync UI with backend
    await refreshProxyState(tab)
  } finally {
    proxyBusy[tab] = false
  }
}

const loadProviderStats = async (tab: ProviderTab) => {
  providerStatsLoading[tab] = true
  try {
    const stats = await fetchProviderDailyStats(tab)
    const mapped: Record<string, ProviderDailyStat> = {}
    ;(stats ?? []).forEach((stat) => {
      mapped[normalizeProviderKey(stat.provider)] = stat
    })
    const hadExistingStats = Object.keys(providerStatsMap[tab] ?? {}).length > 0
    if ((stats?.length ?? 0) > 0) {
      providerStatsMap[tab] = mapped
    } else if (!hadExistingStats) {
      providerStatsMap[tab] = mapped
    }
    providerStatsLoaded[tab] = true
  } catch (error) {
    console.error(`Failed to load provider stats for ${tab}`, error)
    if (!providerStatsLoaded[tab]) {
      providerStatsLoaded[tab] = true
    }
  } finally {
    providerStatsLoading[tab] = false
  }
}

type ProviderStatDisplay =
  | { state: 'loading' | 'empty'; message: string }
  | {
      state: 'ready'
      requests: string
      tokens: string
      cost: string
      successRateLabel: string
      successRateClass: string
    }

const SUCCESS_RATE_THRESHOLDS = {
  healthy: 0.95,
  warning: 0.8,
} as const

const formatSuccessRateLabel = (value: number) => {
  const percent = clamp(value, 0, 1) * 100
  const decimals = percent >= 99.5 || percent === 0 ? 0 : 1
  return `${t('components.main.providers.successRate')}: ${percent.toFixed(decimals)}%`
}

const successRateClassName = (value: number) => {
  const rate = clamp(value, 0, 1)
  if (rate >= SUCCESS_RATE_THRESHOLDS.healthy) {
    return 'success-good'
  }
  if (rate >= SUCCESS_RATE_THRESHOLDS.warning) {
    return 'success-warn'
  }
  return 'success-bad'
}

const providerStatDisplay = (providerName: string): ProviderStatDisplay => {
  const tab = activeTab.value
  if (!providerStatsLoaded[tab]) {
    return { state: 'loading', message: t('components.main.providers.loading') }
  }
  const stat = providerStatsMap[tab]?.[normalizeProviderKey(providerName)]
  if (!stat) {
    return { state: 'empty', message: t('components.main.providers.noData') }
  }
  const totalTokens = stat.input_tokens + stat.output_tokens
  const successRateValue = Number.isFinite(stat.success_rate) ? clamp(stat.success_rate, 0, 1) : null
  const successRateLabel = successRateValue !== null ? formatSuccessRateLabel(successRateValue) : ''
  const successRateClass = successRateValue !== null ? successRateClassName(successRateValue) : ''
  return {
    state: 'ready',
    requests: `${t('components.main.providers.requests')}: ${formatMetric(stat.total_requests)}`,
    tokens: `${t('components.main.providers.tokens')}: ${formatMetric(totalTokens)}`,
    cost: `${t('components.main.providers.cost')}: ${currencyFormatter.value.format(Math.max(stat.cost_total, 0))}`,
    successRateLabel,
    successRateClass,
  }
}

const startProviderStatsTimer = () => {
  stopProviderStatsTimer()
  providerStatsTimer = window.setInterval(() => {
    providerTabIds.forEach((tab) => {
      void loadProviderStats(tab)
    })
    if (show24hStats.value) {
      void load24hStats()
    }
    if (showHeatmap.value) {
      void loadUsageHeatmap()
    }
  }, 60_000)
}

const stopProviderStatsTimer = () => {
  if (providerStatsTimer) {
    clearInterval(providerStatsTimer)
    providerStatsTimer = undefined
  }
}

onMounted(async () => {
  await loadAppSettings()
  void loadUsageHeatmap()
  await loadProvidersFromDisk()
  await Promise.all(providerTabIds.map(refreshProxyState))
  await Promise.all(providerTabIds.map((tab) => loadProviderStats(tab)))
  await checkForUpdates()
  await loadAuthStatus()
  startProviderStatsTimer()
  startUpdateTimer()
  window.addEventListener('app-settings-updated', handleAppSettingsUpdated)
  // Refresh auth status periodically
  authStatusTimer = window.setInterval(() => {
    void loadAuthStatus()
  }, 30000)
})

onUnmounted(() => {
  stopProviderStatsTimer()
  window.removeEventListener('app-settings-updated', handleAppSettingsUpdated)
  stopUpdateTimer()
  if (authStatusTimer) {
    clearInterval(authStatusTimer)
    authStatusTimer = undefined
  }
})

const selectedIndex = ref(0)
const visibleTabs = computed(() => tabs.filter(tab => !('hidden' in tab && tab.hidden)))
const activeTab = computed<ProviderTab>(() => tabs[selectedIndex.value]?.id ?? tabs[0].id)
const activeCards = computed(() => cards[activeTab.value] ?? [])
const currentProxyLabel = computed(() =>
  activeTab.value === 'claude'
    ? t('components.main.relayToggle.hostClaude')
    : activeTab.value === 'codex'
      ? t('components.main.relayToggle.hostCodex')
      : t('components.main.relayToggle.hostOpenCode')
)
const activeProxyState = computed(() => proxyStates[activeTab.value])
const activeProxyBusy = computed(() => proxyBusy[activeTab.value])

// Member is license-unaware; license enforcement is on server/Manager only
const relayToggleDisabled = computed(() => activeProxyBusy.value)
const providerToggleDisabled = computed(() => false)

const goToLogs = () => {
  router.push('/logs')
}

const goToReports = () => {
  router.push('/reports')
}

const goToServer = () => {
  router.push('/server')
}

const goToSettings = () => {
  router.push('/settings')
}

const toggleTheme = () => {
  const next = resolvedTheme.value === 'dark' ? 'light' : 'dark'
  themeMode.value = next
  setTheme(next)
}

const openGitHub = () => {
  Browser.OpenURL(releasePageUrl).catch(() => {
    console.error('failed to open github')
  })
}

type VendorForm = {
  name: string
  apiUrl: string
  apiKey: string
  enabled: boolean
  supportedModels?: string[]
  modelMapping?: Record<string, string>
}

const defaultFormValues = (): VendorForm => ({
  name: '',
  apiUrl: '',
  apiKey: '',
  enabled: true,
  supportedModels: [],
  modelMapping: {},
})

const modalState = reactive({
  open: false,
  tabId: tabs[0].id as ProviderTab,
  editingId: null as number | null,
  form: defaultFormValues(),
  errors: {
    apiUrl: '',
    name: '',
    general: '',
  },
})

// Format backend validation errors into user-friendly messages
const formatSaveError = (error: unknown): string => {
  if (error instanceof Error) {
    const message = error.message
    // Check for model mapping validation errors
    if (message.includes('model mapping') || message.includes('Invalid model mapping')) {
      return t('components.main.form.errors.invalidModelMapping')
    }
    // Check for supported models validation errors
    if (message.includes('supportedModels') || message.includes('modelMapping')) {
      return t('components.main.form.errors.modelMappingValidation')
    }
    // Return original error message for other cases
    return message
  }
  return t('components.main.form.errors.saveFailed')
}

const editingCard = ref<AutomationCard | null>(null)
const confirmState = reactive({ open: false, card: null as AutomationCard | null, tabId: tabs[0].id as ProviderTab })

const isProviderNameAvailable = (name: string, excludeId?: number): boolean => {
  const normalizedName = name.trim().toLowerCase()

  if (!normalizedName) {
    return false
  }

  for (const tab of providerTabIds) {
    const list = cards[tab]
    if (!list) continue

    for (const card of list) {
      if (excludeId !== undefined && card.id === excludeId) {
        continue
      }

      const existingName = card.name.trim().toLowerCase()
      if (existingName === normalizedName) {
        return false
      }
    }
  }

  return true
}

const openCreateModal = () => {
  modalState.tabId = activeTab.value
  modalState.editingId = null
  editingCard.value = null
  Object.assign(modalState.form, defaultFormValues())
  modalState.errors.apiUrl = ''
  modalState.errors.name = ''
  modalState.errors.general = ''
  modalState.open = true
}

const openEditModal = (card: AutomationCard) => {
  modalState.tabId = activeTab.value
  modalState.editingId = card.id
  editingCard.value = card
  Object.assign(modalState.form, {
    name: card.name,
    apiUrl: card.apiUrl,
    apiKey: card.apiKey,
    enabled: card.enabled,
    supportedModels: card.supportedModels || [],
    modelMapping: card.modelMapping || {},
  })
  modalState.errors.apiUrl = ''
  modalState.errors.name = ''
  modalState.errors.general = ''
  modalState.open = true
}

const closeModal = () => {
  modalState.open = false
}

const closeConfirm = () => {
  confirmState.open = false
  confirmState.card = null
}

const submitModal = async () => {
  const list = cards[modalState.tabId]
  if (!list) return
  const name = modalState.form.name.trim()
  const apiUrl = modalState.form.apiUrl.trim()
  const apiKey = modalState.form.apiKey.trim()

  modalState.errors.apiUrl = ''
  modalState.errors.name = ''
  modalState.errors.general = ''

  const excludeId = editingCard.value ? editingCard.value.id : undefined
  if (!isProviderNameAvailable(name, excludeId)) {
    modalState.errors.name = t('components.main.form.errors.duplicateName')
    return
  }

  try {
    const parsed = new URL(apiUrl)
    if (!/^https?:/.test(parsed.protocol)) throw new Error('protocol')
  } catch {
    modalState.errors.apiUrl = t('components.main.form.errors.invalidUrl')
    return
  }

  if (editingCard.value) {
    Object.assign(editingCard.value, {
      apiUrl: apiUrl || editingCard.value.apiUrl,
      apiKey,
      enabled: modalState.form.enabled,
      supportedModels: modalState.form.supportedModels || [],
      modelMapping: modalState.form.modelMapping || {},
    })
    try {
      await persistProviders(modalState.tabId)
      closeModal()
    } catch (error: unknown) {
      modalState.errors.general = formatSaveError(error)
    }
  } else {
    // Get current user to set teamId
    let teamId: number | undefined
    try {
      const user = await getCurrentUser()
      teamId = user.tenant_id
    } catch {
      // User not authenticated, teamId remains undefined
    }

    const newCard: AutomationCard = {
      id: Date.now(),
      name: name || 'Untitled vendor',
      apiUrl,
      apiKey,
      accent: '#0a84ff',
      tint: 'rgba(15, 23, 42, 0.12)',
      enabled: modalState.form.enabled,
      supportedModels: modalState.form.supportedModels || [],
      modelMapping: modalState.form.modelMapping || {},
      level: list.length + 1,
      teamId,
    }
    list.push(newCard)
    try {
      await persistProviders(modalState.tabId)
      closeModal()
    } catch (error: unknown) {
      // Remove the newly added card if save failed
      const index = list.findIndex((c) => c.id === newCard.id)
      if (index > -1) {
        list.splice(index, 1)
      }
      modalState.errors.general = formatSaveError(error)
    }
  }
}

const configure = (card: AutomationCard) => {
  openEditModal(card)
}

const remove = (id: number, tabId: ProviderTab = activeTab.value) => {
  const list = cards[tabId]
  if (!list) return
  const index = list.findIndex((card) => card.id === id)
  if (index > -1) {
    list.splice(index, 1)
    void persistProviders(tabId)
  }
}

const requestRemove = (card: AutomationCard) => {
  confirmState.card = card
  confirmState.tabId = activeTab.value
  confirmState.open = true
}

const confirmRemove = () => {
  if (!confirmState.card) return
  remove(confirmState.card.id, confirmState.tabId)
  closeConfirm()
}

const onDragStart = (id: number) => {
  draggingId.value = id
}

const onDrop = (targetId: number) => {
  // Permission check disabled - allow all users to reorder providers
  // if (!canEditProviders.value) return
  if (draggingId.value === null || draggingId.value === targetId) return
  const currentTab = activeTab.value
  const list = cards[currentTab]
  if (!list) return
  const fromIndex = list.findIndex((card) => card.id === draggingId.value)
  const toIndex = list.findIndex((card) => card.id === targetId)
  if (fromIndex === -1 || toIndex === -1) return
  const [moved] = list.splice(fromIndex, 1)
  const newIndex = fromIndex < toIndex ? toIndex - 1 : toIndex
  list.splice(newIndex, 0, moved)
  draggingId.value = null
  // Update level values based on new order (1-indexed)
  list.forEach((card, index) => {
    card.level = index + 1
  })
  void persistProviders(currentTab)
}

const onDragEnd = () => {
  draggingId.value = null
}

const vendorInitials = (name: string) => {
  if (!name) return 'AI'
  return name
    .split(/\s+/)
    .filter(Boolean)
    .map((word) => word[0])
    .join('')
    .slice(0, 2)
    .toUpperCase()
}

const onTabChange = (visibleIdx: number) => {
  const visibleTab = visibleTabs.value[visibleIdx]
  if (!visibleTab) return
  // Find the actual index in the full tabs array
  const actualIdx = tabs.findIndex(tab => tab.id === visibleTab.id)
  if (actualIdx >= 0) {
    selectedIndex.value = actualIdx
    void refreshProxyState(visibleTab.id as ProviderTab)
    void loadProviderStats(visibleTab.id as ProviderTab)
  }
}

</script>

<style scoped>
/* ==================== Version Info ==================== */
.main-version {
  margin: 32px auto 12px;
  text-align: center;
  color: var(--mac-text-secondary);
  font-size: 0.85rem;
}

/* ==================== Alert/Message Boxes ==================== */
.form-alert {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-radius: 8px;
  font-size: 0.875rem;
  line-height: 1.5;
}

/* Variant: Page-level alerts (top of form) */
.form-alert:not(.form-alert-field) {
  margin-bottom: 16px;
}

/* Variant: Field-level alerts (below inputs) */
.form-alert-field {
  margin-top: 8px;
  padding: 10px 12px;
}

/* Modifier: Error state */
.form-alert-error {
  background-color: var(--error-bg);
  color: var(--error);
  border: 1px solid var(--error);
}

.form-alert svg {
  flex-shrink: 0;
}

.form-alert span {
  flex: 1;
  word-break: break-word;
}

/* ==================== Form Fields ==================== */
.field-hint {
  display: block;
  margin-top: 4px;
  font-size: 0.75rem;
  color: var(--mac-text-secondary);
  font-style: italic;
}

/* ==================== Disabled States ==================== */

/* Disabled buttons */
.ghost-icon.disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.ghost-icon.disabled:hover {
  background-color: transparent;
}

/* Disabled inputs */
:deep(.is-disabled) {
  opacity: 0.6;
  background-color: var(--mac-surface-strong) !important;
  cursor: not-allowed;
}

:deep(.is-disabled:focus) {
  outline: none;
  box-shadow: none;
}
</style>
