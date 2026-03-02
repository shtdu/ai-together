<template>
  <div class="main-shell">
    <div class="global-actions">
      <p class="global-eyebrow">{{ t('components.mcp.hero.eyebrow') }}</p>
      <button class="ghost-icon" :aria-label="t('components.mcp.controls.back')" @click="goHome">
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
      <button class="ghost-icon" :aria-label="t('components.mcp.controls.settings')" @click="goToSettings">
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path
            d="M12 15a3 3 0 100-6 3 3 0 000 6z"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
            fill="none"
          />
          <path
            d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 01-2.83 2.83l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09a1.65 1.65 0 00-1-1.51 1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06a1.65 1.65 0 00.33-1.82 1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09a1.65 1.65 0 001.51-1 1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06a1.65 1.65 0 001.82.33H9a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06a1.65 1.65 0 00-.33 1.82V9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
            fill="none"
          />
        </svg>
      </button>
    </div>

    <div class="contrib-page">
      <section class="contrib-hero">
        <h1>{{ t('components.mcp.hero.title') }}</h1>
        <p class="lead">{{ t('components.mcp.hero.lead') }}</p>
      </section>

      <section class="automation-section">
        <div class="section-header section-header-solo">
          <div class="section-controls">
            <button
              class="ghost-icon"
              :aria-label="t('components.mcp.controls.refresh')"
              :disabled="loading"
              @click="reload"
            >
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <path
                  d="M20.5 8a8.5 8.5 0 10-2.38 7.41"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="1.5"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
                <path
                  d="M20.5 4v4h-4"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="1.5"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
              </svg>
            </button>
            <button class="ghost-icon" :aria-label="t('components.mcp.controls.create')" @click="openCreateModal">
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <path
                  d="M12 5v14M5 12h14"
                  stroke="currentColor"
                  stroke-width="1.5"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  fill="none"
                />
              </svg>
            </button>
          </div>
        </div>

        <div v-if="errorMessage" class="alert-error">{{ errorMessage }}</div>

        <div v-if="loading" class="empty-state">{{ t('components.mcp.list.loading') }}</div>

        <div v-else-if="!servers.length" class="empty-state">
          <p>{{ t('components.mcp.list.empty') }}</p>
          <BaseButton type="button" @click="openCreateModal">
            {{ t('components.mcp.controls.create') }}
          </BaseButton>
        </div>

        <div v-else class="automation-list">
          <article v-for="server in servers" :key="server.name" class="automation-card">
            <div class="card-leading">
              <div class="card-icon" :style="iconStyle(server.name)">
                <span class="icon-fallback">{{ serverInitials(server.name) }}</span>
              </div>
              <div class="card-text">
                <div class="card-title-row">
                  <p class="card-title">{{ server.name }}</p>
                  <span class="chip">{{ typeLabel(server.type) }}</span>
                </div>
                <p class="card-metrics">{{ serverSummary(server) }}</p>
                <p v-if="server.website" class="card-link">
                  <a :href="server.website" target="_blank" rel="noreferrer">{{ server.website }}</a>
                </p>
                <p v-if="server.tips" class="card-tip">{{ server.tips }}</p>
              </div>
            </div>
            <div class="card-platforms">
              <div v-for="option in platformOptions" :key="option.id" class="platform-row">
                <div class="platform-info">
                  <span class="platform-label">{{ option.label }}</span>
                  <div class="platform-controls">
                    <label class="mac-switch sm">
                      <input
                        type="checkbox"
                        :checked="platformActive(server, option.id)"
                        :disabled="saveBusy"
                        @change="onPlatformToggle(server, option.id, $event)"
                      />
                      <span></span>
                    </label>
                  </div>
                </div>
              </div>
            </div>
            <div class="card-actions">
              <button class="ghost-icon" :aria-label="t('components.mcp.list.edit')" @click="openEditModal(server)">
                <svg viewBox="0 0 24 24" aria-hidden="true">
                  <path
                    d="M4 21v-3.5L15.5 6.5a2 2 0 012.83 0l.67.67a2 2 0 010 2.83L7.5 21H4z"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="1.5"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                </svg>
              </button>
              <button class="ghost-icon" :aria-label="t('components.mcp.list.delete')" @click="requestDelete(server)">
                <svg viewBox="0 0 24 24" aria-hidden="true">
                  <path
                    d="M9 3h6m-7 4h8m-6 0v11m4-11v11M5 7h14l-.867 12.138A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.862L5 7z"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="1.5"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                </svg>
              </button>
            </div>
          </article>
        </div>
      </section>
    </div>

    <BaseModal
      :open="modalState.open"
      :title="modalState.editingName ? t('components.mcp.form.editTitle') : t('components.mcp.form.createTitle')"
      @close="closeModal"
    >
      <div class="modal-scroll">
      <form class="vendor-form" @submit.prevent="submitModal">
        <div class="form-row">
          <label class="form-field">
            <span>{{ t('components.mcp.form.name') }}</span>
            <BaseInput v-model="modalState.form.name" type="text" :disabled="saveBusy" />
          </label>
          <label class="form-field">
            <span>{{ t('components.mcp.form.website') }}</span>
            <BaseInput v-model="modalState.form.website" type="text" :disabled="saveBusy" placeholder="https://example.com" />
          </label>
        </div>
        <label class="form-field">
          <span>{{ t('components.mcp.form.type') }}</span>
          <select v-model="modalState.form.type" :disabled="saveBusy" class="base-input">
            <option value="stdio">{{ t('components.mcp.types.stdio') }}</option>
            <option value="http">{{ t('components.mcp.types.http') }}</option>
          </select>
        </label>
        <label v-if="modalState.form.type === 'stdio'" class="form-field">
          <span>{{ t('components.mcp.form.command') }}</span>
          <BaseInput v-model="modalState.form.command" type="text" :disabled="saveBusy" />
        </label>
        <label v-if="modalState.form.type === 'stdio'" class="form-field">
          <span>{{ t('components.mcp.form.args') }}</span>
          <BaseTextarea
            v-model="modalState.form.argsText"
            :placeholder="t('components.mcp.form.argsHint')"
            :disabled="saveBusy"
            rows="5"
          />
        </label>
        <label v-if="modalState.form.type === 'http'" class="form-field">
          <span>{{ t('components.mcp.form.url') }}</span>
          <BaseInput v-model="modalState.form.url" type="text" :disabled="saveBusy" />
        </label>
        <label class="form-field">
          <span>{{ t('components.mcp.form.tips') }}</span>
          <BaseTextarea
            v-model="modalState.form.tips"
            :placeholder="t('components.mcp.form.tipsHint')"
            :disabled="saveBusy"
            rows="4"
          />
        </label>
        <div v-if="modalState.form.type === 'stdio'" class="form-field">
          <span>{{ t('components.mcp.form.env') }}</span>
          <div class="env-table">
            <div v-for="entry in modalState.form.envEntries" :key="entry.id" class="env-row">
              <BaseInput v-model="entry.key" :placeholder="t('components.mcp.form.envKey')" :disabled="saveBusy" />
              <BaseInput v-model="entry.value" :placeholder="t('components.mcp.form.envValue')" :disabled="saveBusy" />
              <button
                class="ghost-icon"
                type="button"
                :aria-label="t('components.mcp.form.envRemove')"
                :disabled="modalState.form.envEntries.length === 1 || saveBusy"
                @click="removeEnvEntry(entry.id)"
              >
                ✕
              </button>
            </div>
          </div>
          <BaseButton variant="outline" type="button" class="env-add" :disabled="saveBusy" @click="addEnvEntry()">
            {{ t('components.mcp.form.envAdd') }}
          </BaseButton>
        </div>
        <div v-if="modalState.form.type === 'http'" class="form-field">
          <span>{{ t('components.mcp.form.headers') }}</span>
          <div class="env-table">
            <div v-for="entry in modalState.form.headerEntries" :key="entry.id" class="env-row">
              <BaseInput v-model="entry.key" :placeholder="t('components.mcp.form.headerKey')" :disabled="saveBusy" />
              <BaseInput v-model="entry.value" :placeholder="t('components.mcp.form.headerValue')" :disabled="saveBusy" />
              <button
                class="ghost-icon"
                type="button"
                :aria-label="t('components.mcp.form.headerRemove')"
                :disabled="modalState.form.headerEntries.length === 1 || saveBusy"
                @click="removeHeaderEntry(entry.id)"
              >
                ✕
              </button>
            </div>
          </div>
          <BaseButton variant="outline" type="button" class="env-add" :disabled="saveBusy" @click="addHeaderEntry()">
            {{ t('components.mcp.form.headerAdd') }}
          </BaseButton>
        </div>
        <div class="form-field">
          <span>{{ t('components.mcp.form.platforms.title') }}</span>
          <div class="platform-checkboxes">
            <label v-for="option in platformOptions" :key="option.id" class="platform-checkbox">
              <input
                type="checkbox"
                :checked="modalState.form.enablePlatform.includes(option.id)"
                :disabled="saveBusy"
                @change="onModalPlatformToggle(option.id, $event)"
              />
              <span>{{ option.label }}</span>
            </label>
          </div>
        </div>

        <p v-if="modalError" class="alert-error">{{ modalError }}</p>

        <footer class="form-actions">
          <BaseButton variant="outline" type="button" :disabled="saveBusy" @click="closeModal">
            {{ t('components.mcp.form.actions.cancel') }}
          </BaseButton>
          <BaseButton :disabled="saveBusy" type="submit">
            {{ t('components.mcp.form.actions.save') }}
          </BaseButton>
        </footer>
      </form>
      </div>
    </BaseModal>

    <BaseModal
      :open="confirmState.open"
      :title="t('components.mcp.form.deleteTitle')"
      variant="confirm"
      @close="closeConfirm"
    >
      <div class="confirm-body">
        <p>
          {{ t('components.mcp.form.deleteMessage', { name: confirmState.target?.name ?? '' }) }}
        </p>
      </div>
      <footer class="form-actions confirm-actions">
        <BaseButton variant="outline" type="button" :disabled="saveBusy" @click="closeConfirm">
          {{ t('components.mcp.form.actions.cancel') }}
        </BaseButton>
        <BaseButton variant="danger" type="button" :disabled="saveBusy" @click="confirmDelete">
          {{ t('components.mcp.form.actions.delete') }}
        </BaseButton>
      </footer>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import BaseButton from '../common/BaseButton.vue'
import BaseModal from '../common/BaseModal.vue'
import BaseInput from '../common/BaseInput.vue'
import BaseTextarea from '../common/BaseTextarea.vue'
import { fetchMcpServers, saveMcpServers, type McpPlatform, type McpServer, type McpServerType } from '../../services/mcp'
import { showToast } from '../../utils/toast'

type EnvEntry = {
  id: number
  key: string
  value: string
}

type McpForm = {
  name: string
  type: McpServerType
  command: string
  url: string
  website: string
  tips: string
  argsText: string
  envEntries: EnvEntry[]
  headerEntries: EnvEntry[]
  enablePlatform: McpPlatform[]
}

const { t } = useI18n()
const router = useRouter()

const servers = ref<McpServer[]>([])
const loading = ref(false)
const saveBusy = ref(false)
const errorMessage = ref('')
const modalError = ref('')
const placeholderRegex = /\{([a-zA-Z0-9_]+)\}/g

let envEntryId = 0

const createEnvEntry = (key = '', value = ''): EnvEntry => ({
  id: ++envEntryId,
  key,
  value,
})

const createEmptyForm = (): McpForm => ({
  name: '',
  type: 'stdio',
  command: '',
  url: '',
  website: '',
  tips: '',
  argsText: '',
  envEntries: [createEnvEntry()],
  headerEntries: [createEnvEntry()],
  enablePlatform: [],
})

const modalState = reactive({
  open: false,
  editingName: '',
  form: createEmptyForm(),
})

const confirmState = reactive<{ open: boolean; target: McpServer | null }>({
  open: false,
  target: null,
})

const platformOptions = computed(() => [
  { id: 'claude-code' as McpPlatform, label: t('components.mcp.platforms.claude') },
  { id: 'codex' as McpPlatform, label: t('components.mcp.platforms.codex') },
  { id: 'opencode' as McpPlatform, label: t('components.mcp.platforms.opencode') },
])

const formMissingPlaceholders = computed(() => detectPlaceholders(modalState.form.url, modalState.form.argsText))

const loadServers = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const data = await fetchMcpServers()
    servers.value = (data ?? []).map((item) => ({
      ...item,
      args: item.args ?? [],
      env: item.env ?? {},
      enable_platform: item.enable_platform ?? [],
      website: item.website ?? '',
      tips: item.tips ?? '',
      missing_placeholders: item.missing_placeholders ?? [],
    }))
  } catch (error) {
    console.error('failed to load mcp servers', error)
    errorMessage.value = t('components.mcp.list.loadError')
  } finally {
    loading.value = false
  }
}

const persistServers = async () => {
  saveBusy.value = true
  try {
    await saveMcpServers(servers.value)
    await loadServers()
  } catch (error) {
    console.error('failed to save mcp servers', error)
    errorMessage.value = t('components.mcp.list.saveError')
  } finally {
    saveBusy.value = false
  }
}

const iconStyle = (name: string) => ({
  backgroundColor: 'rgba(255,255,255,0.08)',
  color: 'var(--text-primary)',
})

const serverInitials = (name: string) => {
  if (!name) return 'MC'
  return name
    .split(/\s+/)
    .filter(Boolean)
    .map((word) => word[0])
    .join('')
    .slice(0, 2)
    .toUpperCase()
}

const serverSummary = (server: McpServer) => {
  let summary: string
  if (server.type === 'http' && server.url) {
    summary = `${t('components.mcp.types.httpShort')} · ${server.url}`
  } else if (server.command) {
    summary = `${t('components.mcp.types.stdioShort')} · ${server.command}`
  } else {
    summary = server.type === 'http' ? t('components.mcp.types.httpShort') : t('components.mcp.types.stdioShort')
  }

  // Limit to 20 characters, with "..." if truncated
  if (summary.length > 35) {
    return summary.slice(0, 32) + '...'
  }
  return summary
}

const typeLabel = (type: McpServerType) =>
  type === 'http' ? t('components.mcp.types.http') : t('components.mcp.types.stdio')

const platformEnabled = (server: McpServer, platform: McpPlatform) =>
  server.enable_platform?.includes(platform) ?? false

const platformActive = (server: McpServer, platform: McpPlatform) => {
  if (platform === 'claude-code') return server.enabled_in_claude
  if (platform === 'codex') return server.enabled_in_codex
  if (platform === 'opencode') return server.enabled_in_opencode
  return false
}

const hasMissingPlaceholders = (server: McpServer) => (server.missing_placeholders?.length ?? 0) > 0

const showPlaceholderWarning = (variables: string[]) => {
  const list = (variables ?? []).filter(Boolean)
  showToast(t('components.mcp.toast.placeholder', { vars: list.join(', ') || 'variables' }), 'error')
}

const onModalPlatformToggle = (platform: McpPlatform, event: Event) => {
  const targetInput = event.target as HTMLInputElement | null
  if (!targetInput) return

  if (formMissingPlaceholders.value.length > 0) {
    targetInput.checked = modalState.form.enablePlatform.includes(platform)
    showPlaceholderWarning(formMissingPlaceholders.value)
    return
  }

  const next = new Set<McpPlatform>(modalState.form.enablePlatform)
  if (targetInput.checked) {
    next.add(platform)
  } else {
    next.delete(platform)
  }
  modalState.form.enablePlatform = Array.from(next)
}

const onPlatformToggle = async (server: McpServer, platform: McpPlatform, event: Event) => {
  const targetInput = event.target as HTMLInputElement | null
  if (!targetInput) return

  if (hasMissingPlaceholders(server)) {
    targetInput.checked = platformEnabled(server, platform)
    showPlaceholderWarning(server.missing_placeholders ?? [])
    return
  }

  const target = servers.value.find((item) => item.name === server.name)
  if (!target) return

  const next = new Set<McpPlatform>(target.enable_platform ?? [])
  if (targetInput.checked) {
    next.add(platform)
  } else {
    next.delete(platform)
  }
  target.enable_platform = Array.from(next)
  await persistServers()
}

const openCreateModal = () => {
  modalState.open = true
  modalState.editingName = ''
  modalState.form = createEmptyForm()
  modalError.value = ''
}

const openEditModal = (server: McpServer) => {
  modalState.open = true
  modalState.editingName = server.name
  modalError.value = ''

  // Build enablePlatform list based on actual enabled state
  const enabledPlatforms: McpPlatform[] = []
  if (server.enabled_in_claude) enabledPlatforms.push('claude-code')
  if (server.enabled_in_codex) enabledPlatforms.push('codex')
  if (server.enabled_in_opencode) enabledPlatforms.push('opencode')

  modalState.form = {
    name: server.name,
    type: server.type,
    command: server.command ?? '',
    url: server.url ?? '',
    website: server.website ?? '',
    tips: server.tips ?? '',
    argsText: (server.args ?? []).join('\n'),
    envEntries: buildEnvEntries(server.env),
    headerEntries: buildEnvEntries(server.headers),
    enablePlatform: enabledPlatforms,
  }
}

const closeModal = () => {
  modalState.open = false
  modalState.editingName = ''
  modalState.form = createEmptyForm()
  modalError.value = ''
}

const buildEnvEntries = (env: Record<string, string> | undefined) => {
  const entries = Object.entries(env ?? {})
  if (!entries.length) {
    return [createEnvEntry()]
  }
  return entries.map(([key, value]) => createEnvEntry(key, value))
}

const addEnvEntry = () => {
  modalState.form.envEntries.push(createEnvEntry())
}

const removeEnvEntry = (id: number) => {
  if (modalState.form.envEntries.length === 1) return
  const index = modalState.form.envEntries.findIndex((entry) => entry.id === id)
  if (index !== -1) {
    modalState.form.envEntries.splice(index, 1)
  }
}

const addHeaderEntry = () => {
  modalState.form.headerEntries.push(createEnvEntry())
}

const removeHeaderEntry = (id: number) => {
  if (modalState.form.headerEntries.length === 1) return
  const index = modalState.form.headerEntries.findIndex((entry) => entry.id === id)
  if (index !== -1) {
    modalState.form.headerEntries.splice(index, 1)
  }
}

const closeConfirm = () => {
  confirmState.open = false
  confirmState.target = null
}

const requestDelete = (server: McpServer) => {
  confirmState.target = server
  confirmState.open = true
}

const confirmDelete = async () => {
  if (!confirmState.target) return
  servers.value = servers.value.filter((server) => server.name !== confirmState.target?.name)
  closeConfirm()
  await persistServers()
}

const submitModal = async () => {
  modalError.value = ''
  const form = modalState.form
  const trimmedName = form.name.trim()
  if (!trimmedName) {
    modalError.value = t('components.mcp.form.errors.name')
    return
  }
  if (form.type === 'stdio' && !form.command.trim()) {
    modalError.value = t('components.mcp.form.errors.command')
    return
  }
  if (form.type === 'http' && !form.url.trim()) {
    modalError.value = t('components.mcp.form.errors.url')
    return
  }

  const existing = servers.value.find((server) => server.name === trimmedName)
  if (!modalState.editingName && existing) {
    modalError.value = t('components.mcp.form.errors.duplicate')
    return
  }
  if (modalState.editingName && modalState.editingName !== trimmedName && existing) {
    modalError.value = t('components.mcp.form.errors.duplicate')
    return
  }

  const payload: McpServer = {
    name: trimmedName,
    type: form.type,
    command: form.type === 'stdio' ? form.command.trim() : '',
    args: parseArgs(form.argsText),
    env: parseEnv(form.envEntries),
    url: form.type === 'http' ? form.url.trim() : '',
    website: form.website.trim(),
    tips: form.tips.trim(),
    headers: parseEnv(form.headerEntries),
    enable_platform: [...form.enablePlatform],
    enabled_in_claude: form.enablePlatform.includes('claude-code'),
    enabled_in_codex: form.enablePlatform.includes('codex'),
    enabled_in_opencode: form.enablePlatform.includes('opencode'),
    missing_placeholders: [],
  }

  if (modalState.editingName) {
    const index = servers.value.findIndex((server) => server.name === modalState.editingName)
    if (index !== -1) {
      servers.value.splice(index, 1, payload)
    } else {
      servers.value.push(payload)
    }
  } else {
    servers.value.push(payload)
  }

  closeModal()
  await persistServers()
}

const parseArgs = (value: string) =>
  value
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)

const parseEnv = (entries: EnvEntry[]) => {
  return entries.reduce<Record<string, string>>((acc, entry) => {
    const key = entry.key.trim()
    if (!key) return acc
    acc[key] = entry.value
    return acc
  }, {})
}

const goHome = () => {
  router.push('/')
}

const goToSettings = () => {
  router.push('/settings')
}

const reload = async () => {
  await loadServers()
}

const detectPlaceholders = (url: string, argsText: string) => {
  const set = new Set<string>()
  collectPlaceholders(url, set)
  argsText
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)
    .forEach((line) => collectPlaceholders(line, set))
  return Array.from(set)
}

const collectPlaceholders = (value: string, set: Set<string>) => {
  if (!value) return
  const matches = value.matchAll(placeholderRegex)
  for (const match of matches) {
    const key = match[1]
    if (key) {
      set.add(key)
    }
  }
}

onMounted(() => {
  void loadServers()
})
</script>

<style scoped>
.chip {
  padding: 2px 8px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.08);
  font-size: 12px;
  text-transform: uppercase;
}

.card-platforms {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  width: 160px;
  margin-left: auto;
}

.platform-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.platform-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.75rem;
  width: 160px;
}

.platform-label {
  font-weight: 600;
}

.platform-controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.platform-status {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
}

.platform-status.active {
  color: #4ade80;
}

.card-actions {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  align-items: flex-end;
}

.empty-state {
  text-align: center;
  padding: 2rem;
  border: 1px dashed rgba(255, 255, 255, 0.2);
  border-radius: 16px;
}

.alert-error {
  margin-bottom: 1rem;
  padding: 0.75rem 1rem;
  border-radius: 12px;
  background: rgba(244, 67, 54, 0.15);
  color: #ff9b9b;
}

.vendor-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.modal-scroll {
  max-height: 65vh;
  overflow-y: auto;
  padding-right: 0.25rem;
  margin-right: -0.25rem;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
  width: 100%;
}

.env-table {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.env-row {
  display: grid;
  grid-template-columns: 1fr 1fr auto;
  gap: 0.5rem;
  align-items: center;
}

.env-add {
  align-self: flex-start;
}

.platform-checkboxes {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.platform-checkbox {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

.card-leading {
  display: flex;
  gap: 1rem;
}

.card-icon {
  display: inline-flex;
  justify-content: center;
  align-items: center;
  width: 48px;
  height: 48px;
  border-radius: 14px;
  overflow: hidden;
}

.card-text {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.card-link {
  margin-top: 0.25rem;
}

.card-link a {
  color: var(--link-color, #9acaff);
  text-decoration: none;
}

.card-link a:hover {
  text-decoration: underline;
}

.card-tip {
  margin-top: 0.25rem;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.7);
}

.confirm-body {
  margin-bottom: 1rem;
}
</style>
