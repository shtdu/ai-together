<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import ListItem from '../Setting/ListRow.vue'
import BaseButton from '../common/BaseButton.vue'
import BaseInput from '../common/BaseInput.vue'
import BaseModal from '../common/BaseModal.vue'
import {
  getCurrentUser,
  isAuthenticated,
  login,
  logout,
  getUserProfile,
  type User,
} from '../../services/auth'
import {
  syncProviders,
  syncUsageStats,
  getLastSyncTime,
  getPendingRecordCount,
  pushProvidersToServer,
  hasLocalChanges,
  deleteAllLocalUsageData,
  deleteAllProviders,
  clearSyncState,
} from '../../services/sync'
import {
  isAdmin,
  canEditSettings,
} from '../../services/permission'
import {
  isBackgroundSyncRunning,
  syncNow,
} from '../../services/backgroundSync'
import {
  fetchProxyStatus,
  disableProxy,
  type Platform,
} from '../../services/claudeSettings'
import {
  getServerURL,
  setServerURL,
  testConnection,
} from '../../services/server'
import { restartApp } from '../../services/app'
import { showToast } from '../../utils/toast'

const router = useRouter()
const { t } = useI18n()

const authStatus = ref<'loading' | 'authenticated' | 'unauthenticated'>('loading')
const currentUser = ref<User | null>(null)

const showLoginModal = ref(false)
const loginEmail = ref('')
const loginPassword = ref('')
const authBusy = ref(false)

const lastSyncTime = ref('')
const pendingRecords = ref(0)
const syncBusy = ref(false)

const backgroundSyncRunning = ref(false)
const configSyncInterval = ref(30) // minutes
const usageSyncInterval = ref(5) // minutes

const userIsAdmin = ref(false)
const userCanEdit = ref(false)
const pushBusy = ref(false)
const showPushConfirm = ref(false)
const hasPendingLocalChanges = ref(false)
const logoutBusy = ref(false)

// Server connection state
const serverURL = ref('')
const initialServerURL = ref('')
const connectionStatus = ref<'unknown' | 'testing' | 'success' | 'error'>('unknown')
const connectionError = ref('')
const showRestartNotice = ref(false)

let refreshInterval: ReturnType<typeof setInterval> | null = null

const goBack = () => {
  router.push('/')
}

const loadAuthStatus = async () => {
  authStatus.value = 'loading'
  try {
    const authenticated = await isAuthenticated()
    if (authenticated) {
      const user = await getCurrentUser()
      currentUser.value = user
      authStatus.value = 'authenticated'
      await loadPermissions()
    } else {
      currentUser.value = null
      authStatus.value = 'unauthenticated'
      userIsAdmin.value = false
      userCanEdit.value = false
    }
  } catch (error) {
    console.error('Failed to load auth status', error)
    authStatus.value = 'unauthenticated'
    currentUser.value = null
    userIsAdmin.value = false
    userCanEdit.value = false
  }
}

const loadPermissions = async () => {
  if (authStatus.value !== 'authenticated') {
    userIsAdmin.value = false
    userCanEdit.value = false
    return
  }

  try {
    const [admin, canEdit] = await Promise.all([isAdmin(), canEditSettings()])
    userIsAdmin.value = admin
    userCanEdit.value = canEdit
  } catch (error) {
    console.error('Failed to load permissions', error)
    userIsAdmin.value = false
    userCanEdit.value = false
  }
}

const handlePushToServer = async () => {
  if (!userIsAdmin.value) {
    showToast(t('components.server.push.onlyManagers'), 'error')
    return
  }

  pushBusy.value = true
  try {
    await pushProvidersToServer()
    showToast(t('components.server.push.success'), 'success')
    showPushConfirm.value = false
    await loadSyncStatus()
    await checkPendingLocalChanges()
  } catch (error) {
    console.error('Push to server failed', error)
    const message = error instanceof Error ? error.message : 'Push to server failed'
    showToast(message, 'error')
  } finally {
    pushBusy.value = false
  }
}

const handleLogin = async () => {
  if (!loginEmail.value || !loginPassword.value) {
    showToast(t('components.server.validation.fillAllFields'), 'error')
    return
  }

  authBusy.value = true
  try {
    const result = await login(loginEmail.value, loginPassword.value)
    currentUser.value = result.user
    authStatus.value = 'authenticated'
    showLoginModal.value = false
    loginEmail.value = ''
    loginPassword.value = ''
    showToast(`Welcome, ${result.user.name}!`, 'success')
    // Trigger provider sync after successful login
    try {
      await syncProviders()
      await loadSyncStatus()
    } catch (syncError) {
      console.error('Provider sync failed after login:', syncError)
      // Don't show error toast for sync failure - user is logged in
      await loadSyncStatus()
    }
  } catch (error) {
    showToast(t('components.server.auth.loginFailed'), 'error')
  } finally {
    authBusy.value = false
  }
}

const handleLogout = async () => {
  if (logoutBusy.value) return

  logoutBusy.value = true
  const errors: string[] = []

  try {
    // Step 1: Force sync usage logs (with timeout)
    try {
      const pendingCount = await getPendingRecordCount()
      if (pendingCount > 0) {
        console.log(`Syncing ${pendingCount} pending usage records before logout...`)
        const synced = await syncUsageStats()
        console.log(`Synced ${synced} usage records`)
      }
    } catch (error) {
      console.error('Failed to sync usage stats:', error)
      errors.push('Failed to sync usage stats')
      // Continue with cleanup even if sync fails
    }

    // Step 2: Delete all local usage data
    try {
      const deleted = await deleteAllLocalUsageData()
      console.log(`Deleted ${deleted} local usage records`)
    } catch (error) {
      console.error('Failed to delete local usage data:', error)
      errors.push('Failed to delete local usage data')
    }

    // Step 3: Disable all proxies if enabled (check status first, then disable)
    for (const platform of ['claude', 'codex', 'opencode'] as Platform[]) {
      try {
        const status = await fetchProxyStatus(platform)
        if (status.enabled) {
          console.log(`Disabling ${platform} proxy...`)
          await disableProxy(platform)
          console.log(`Disabled ${platform} proxy`)
        }
      } catch (error) {
        console.error(`Failed to disable ${platform} proxy:`, error)
        errors.push(`Failed to disable ${platform} proxy`)
      }
    }

    // Step 4: Delete all provider settings
    try {
      await deleteAllProviders()
      console.log('Deleted all provider settings')
    } catch (error) {
      console.error('Failed to delete providers:', error)
      errors.push('Failed to delete providers')
    }

    // Step 5: Clear sync state
    try {
      await clearSyncState()
      console.log('Cleared sync state')
    } catch (error) {
      console.error('Failed to clear sync state:', error)
      errors.push('Failed to clear sync state')
    }

    // Step 6: Delete auth information (current step)
    try {
      await logout()
      console.log('Logged out successfully')
    } catch (error) {
      console.error('Failed to logout:', error)
      errors.push('Failed to logout')
    }

    // Step 7: Update frontend state
    currentUser.value = null
    authStatus.value = 'unauthenticated'
    userIsAdmin.value = false
    userCanEdit.value = false
    hasPendingLocalChanges.value = false
    lastSyncTime.value = ''
    pendingRecords.value = 0

    if (errors.length > 0) {
      showToast(`Logout completed with warnings: ${errors.join(', ')}`, 'error')
    } else {
      showToast(t('components.server.auth.logoutSuccess'), 'success')
    }
  } catch (error) {
    console.error('Unexpected error during logout:', error)
    showToast(t('components.server.auth.logoutFailed'), 'error')
  } finally {
    logoutBusy.value = false
  }
}

const loadSyncStatus = async () => {
  if (authStatus.value !== 'authenticated') {
    lastSyncTime.value = ''
    pendingRecords.value = 0
    hasPendingLocalChanges.value = false
    return
  }

  try {
    const [syncTime, pending] = await Promise.all([getLastSyncTime(), getPendingRecordCount()])
    lastSyncTime.value = syncTime
    pendingRecords.value = pending
  } catch (error) {
    console.error('Failed to load sync status', error)
  }
}

const checkPendingLocalChanges = async () => {
  if (!userIsAdmin.value || !userCanEdit.value) {
    hasPendingLocalChanges.value = false
    return
  }
  try {
    hasPendingLocalChanges.value = await hasLocalChanges()
  } catch (error) {
    console.error('Failed to check for pending local changes', error)
    hasPendingLocalChanges.value = false
  }
}

const handleSyncNow = async () => {
  if (authStatus.value !== 'authenticated') {
    showToast(t('components.server.auth.loginRequired'), 'error')
    return
  }

  syncBusy.value = true
  try {
    const result = await syncNow()
    if (result.error) {
      showToast(`Sync error: ${result.error}`, 'error')
    } else {
      let message = t('components.server.sync.completed')
      if (result.configSynced) message += ', providers synced from server'
      if (result.usageSynced > 0) message += `, ${result.usageSynced} usage records synced`
      showToast(message, 'success')
      await loadSyncStatus()
    }
  } catch (error) {
    showToast(t('components.server.sync.failed'), 'error')
  } finally {
    syncBusy.value = false
  }
}

const loadBackgroundSyncStatus = async () => {
  try {
    backgroundSyncRunning.value = await isBackgroundSyncRunning()
  } catch (error) {
    console.error('Failed to load background sync status', error)
  }
}

const formatDate = (dateStr: string): string => {
  if (!dateStr) return t('components.server.formats.never')
  // Check for zero time (Go's time.Time{} zero value serializes to "0001-01-01T00:00:00Z")
  if (dateStr.startsWith('0001-01-01') || dateStr.startsWith('0000-01-01')) {
    return t('components.server.formats.never')
  }
  try {
    const date = new Date(dateStr)
    const diffMs = Date.now() - date.getTime()
    if (Number.isNaN(diffMs) || diffMs < 0) {
      return t('components.server.formats.invalidDate')
    }
    const diffMinutes = Math.floor(diffMs / 60000)
    if (diffMinutes < 1) return t('components.server.formats.justNow')
    return t('components.server.formats.minutesAgo', { minutes: diffMinutes })
  } catch {
    return t('components.server.formats.invalidDate')
  }
}

const refreshStatus = async () => {
  await loadAuthStatus()
  await Promise.all([
    loadSyncStatus(),
    loadBackgroundSyncStatus(),
    checkPendingLocalChanges(),
  ])
}

const loadServerConfig = async () => {
  try {
    serverURL.value = await getServerURL()
    initialServerURL.value = serverURL.value
  } catch (error) {
    console.error('Failed to load server config', error)
  }
}

const handleTestConnection = async () => {
  connectionStatus.value = 'testing'
  connectionError.value = ''

  try {
    await testConnection(serverURL.value)
    connectionStatus.value = 'success'
    showToast(t('components.server.notifications.connectionSuccess'), 'success')
  } catch (error) {
    connectionStatus.value = 'error'
    connectionError.value = error instanceof Error ? error.message : String(error)
    showToast(t('components.server.notifications.connectionFailed'), 'error')
  }
}

const handleSaveServerURL = async () => {
  if (!serverURL.value) {
    showToast(t('components.server.notifications.urlRequired'), 'error')
    return
  }

  // Test connection first before saving
  connectionStatus.value = 'testing'
  connectionError.value = ''

  try {
    // Test the connection with the URL from input field
    await testConnection(serverURL.value)

    // Only save if test succeeds
    await setServerURL(serverURL.value)
    connectionStatus.value = 'success'
    showToast(t('components.server.notifications.urlSaved'), 'success')

    if (serverURL.value !== initialServerURL.value) {
      showRestartNotice.value = true
    }
  } catch (error) {
    connectionStatus.value = 'error'
    connectionError.value = error instanceof Error ? error.message : String(error)
    showToast(t('components.server.notifications.testFailed'), 'error')
  }
}

const handleRestartConfirm = async () => {
  showRestartNotice.value = false
  await restartApp()
}

const serverStatusColor = computed(() => {
  switch (connectionStatus.value) {
    case 'success':
      return 'server-status-success'
    case 'error':
      return 'server-status-error'
    case 'testing':
      return 'server-status-testing'
    default:
      return 'server-status-unknown'
  }
})

const serverStatusText = computed(() => {
  switch (connectionStatus.value) {
    case 'success':
      return t('components.server.serverStatus.connected')
    case 'error':
      return t('components.server.serverStatus.connectionFailed')
    case 'testing':
      return 'Testing...'
    default:
      return t('components.server.serverStatus.notTested')
  }
})

onMounted(async () => {
  await refreshStatus()
  void loadServerConfig()

  // Refresh status every 30 seconds
  refreshInterval = setInterval(refreshStatus, 30000)
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})

</script>

<template>
  <div class="main-shell general-shell">
    <div class="global-actions">
      <p class="global-eyebrow">{{ t('components.main.controls.serverIntegration') }}</p>
      <button class="ghost-icon" :aria-label="t('components.server.actions.back')" @click="goBack">
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

    <div class="general-page">
      <section>
        <h2 class="mac-section-title">{{ t('components.server.sections.authentication') }}</h2>
        <div class="mac-panel">
          <div v-if="authStatus === 'loading'" class="auth-loading">
            <ListItem :label="t('components.server.labels.status')">{{ t('components.logs.loading') }}</ListItem>
          </div>

          <div v-else-if="authStatus === 'authenticated' && currentUser" class="auth-authenticated">
            <ListItem :label="currentUser.name" :subLabel="`${currentUser.email} · ${t('components.server.labels.role')}: ${currentUser.role}`">
              <BaseButton @click="handleLogout" :disabled="logoutBusy" variant="outline" size="sm">
  {{ logoutBusy ? t('components.server.auth.loggingOut') : t('components.server.auth.logout') }}
</BaseButton>
            </ListItem>
          </div>

          <div v-else-if="authStatus === 'unauthenticated'" class="auth-unauthenticated">
            <div class="auth-content">
              <p class="auth-description">{{ t('components.server.description') }}</p>
              <div class="auth-actions">
                <BaseButton @click="showLoginModal = true" size="sm">{{ t('components.server.auth.login') }}</BaseButton>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section>
        <h2 class="mac-section-title">{{ t('components.server.sections.syncStatus') }}</h2>
        <div class="mac-panel">
          <ListItem :label="t('components.server.labels.lastConfigSync')">{{ formatDate(lastSyncTime) }}</ListItem>
          <ListItem :label="t('components.server.labels.pendingUsageRecords')">{{ String(pendingRecords) }}</ListItem>
          <ListItem :label="t('components.server.labels.backgroundSync')">{{ backgroundSyncRunning ? t('components.server.status.running') : t('components.server.status.stopped') }}</ListItem>

          <div class="sync-actions">
            <BaseButton
              @click="handleSyncNow"
              :disabled="authStatus !== 'authenticated' || syncBusy"
              size="sm"
              class="sync-now-button"
            >
              {{ syncBusy ? t('components.server.sync.syncing') : t('components.server.sync.syncNow') }}
            </BaseButton>
            <div
              v-if="userIsAdmin"
              class="push-button-wrapper"
            >
              <BaseButton
                @click="showPushConfirm = true"
                :disabled="!hasPendingLocalChanges || authStatus !== 'authenticated' || pushBusy"
                :variant="hasPendingLocalChanges ? 'primary' : 'outline'"
                size="sm"
                class="push-button"
              >
                {{ hasPendingLocalChanges ? t('components.server.push.pushWithChanges') : t('components.server.push.push') }}
              </BaseButton>
            </div>
          </div>
        </div>
      </section>

      <section>
        <h2 class="mac-section-title">{{ t('components.server.serverConnection.title') }}</h2>
        <div class="mac-panel">
          <div class="server-connection-content">
            <div v-if="authStatus === 'authenticated'" class="server-url-locked-hint">
              <svg class="lock-icon" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd" />
              </svg>
              <span>{{ t('components.server.serverConnection.lockedHint') }}</span>
            </div>
            <div class="server-url-row">
              <label class="server-url-label">{{ t('components.server.serverConnection.serverUrl') }}</label>
              <div class="server-url-controls">
                <BaseInput
                  v-model="serverURL"
                  :placeholder="t('components.server.serverConnection.placeholder')"
                  class="server-url-input"
                  :class="{ 'is-locked': authStatus === 'authenticated' }"
                  :disabled="authStatus === 'authenticated'"
                />
                <BaseButton
                  @click="handleSaveServerURL"
                  :disabled="!serverURL || authStatus === 'authenticated'"
                  size="sm"
                >
                  {{ t('components.server.serverConnection.save') }}
                </BaseButton>
                <BaseButton
                  @click="handleTestConnection"
                  :disabled="!serverURL || connectionStatus === 'testing' || authStatus === 'authenticated'"
                  variant="outline"
                  size="sm"
                >
                  {{ t('components.server.serverConnection.test') }}
                </BaseButton>
              </div>
            </div>

            <div v-if="connectionStatus !== 'unknown'" class="server-status-row">
              <div :class="['server-status', serverStatusColor]">
                <svg class="server-status-icon" fill="currentColor" viewBox="0 0 20 20">
                  <circle cx="10" cy="10" r="10" />
                </svg>
                <span class="server-status-text">{{ serverStatusText }}</span>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- Restart Notice Modal -->
    <BaseModal
      :open="showRestartNotice"
      :title="t('components.server.notifications.restartTitle')"
      @close="handleRestartConfirm"
    >
      <div class="confirm-body">
        <p>{{ t('components.server.notifications.restartToApply') }}</p>
      </div>
      <footer class="form-actions confirm-actions">
        <BaseButton type="button" @click="handleRestartConfirm">
          {{ t('components.general.buttons.restart') }}
        </BaseButton>
      </footer>
    </BaseModal>

    <!-- Push to Server Confirmation Modal -->
    <BaseModal :open="showPushConfirm" :title="t('components.server.confirmPush.title')" @close="showPushConfirm = false">
      <div class="space-y-4 p-6">
        <p>{{ t('components.server.confirmPush.message') }}</p>
        <p class="text-sm text-gray-500 dark:text-gray-400">
          {{ t('components.server.confirmPush.warning') }}
        </p>
        <div class="form-actions">
          <BaseButton @click="showPushConfirm = false" variant="outline">{{ t('components.main.form.actions.cancel') }}</BaseButton>
          <BaseButton @click="handlePushToServer" :disabled="pushBusy">
            {{ pushBusy ? t('components.server.push.pushing') : t('components.server.push.push') }}
          </BaseButton>
        </div>
      </div>
    </BaseModal>

    <!-- Login Modal -->
    <BaseModal :open="showLoginModal" :title="t('components.server.auth.login')" @close="showLoginModal = false">
      <div class="space-y-4 p-6">
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            {{ t('components.server.auth.formLabels.email') }}
          </label>
          <BaseInput v-model="loginEmail" type="email" :placeholder="t('components.server.auth.placeholders.email')" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            {{ t('components.server.auth.formLabels.password') }}
          </label>
          <BaseInput v-model="loginPassword" type="password" :placeholder="t('components.server.auth.placeholders.password')" />
        </div>
        <div class="form-actions">
          <BaseButton @click="showLoginModal = false" variant="outline">{{ t('components.main.form.actions.cancel') }}</BaseButton>
          <BaseButton @click="handleLogin" :disabled="authBusy">{{ t('components.server.auth.login') }}</BaseButton>
        </div>
      </div>
    </BaseModal>
  </div>
</template>

<style scoped>
.auth-loading {
  padding: 18px 20px;
}

.auth-authenticated {
  /* ListItem handles its own padding */
}

.auth-unauthenticated {
  padding: 18px 20px;
}

.auth-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.auth-description {
  font-size: 0.85rem;
  color: var(--mac-text-secondary);
  margin: 0;
  line-height: 1.5;
}

.auth-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-start;
}

.sync-actions {
  padding: 18px 20px;
  border-top: 1px solid var(--mac-divider);
  display: flex;
  justify-content: flex-start;
  gap: 8px;
}

.sync-now-button {
  min-width: 100px;
}

.push-button {
  min-width: 120px;
}

.push-button-wrapper {
  display: inline-block;
  position: relative;
}

/* Push button wrapper uses global tooltip styles from style.css */

.server-connection-content {
  padding: 18px 20px;
}

.server-url-row {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 12px;
}

.server-url-label {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--mac-text);
  white-space: nowrap;
  min-width: fit-content;
}

.server-url-controls {
  display: flex;
  gap: 8px;
  align-items: center;
  flex: 1;
}

.server-url-input {
  flex: 1;
  min-width: 0;
}

.server-status-row {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--mac-divider);
}

.server-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.85rem;
  font-weight: 500;
}

.server-status-icon {
  width: 12px;
  height: 12px;
}

.server-status-text {
  line-height: 1.2;
}

.server-status-success {
  color: #22c55e;
}

.server-status-error {
  color: #ef4444;
}

.server-status-testing {
  color: #eab308;
}

.server-status-unknown {
  color: var(--mac-text-secondary);
}

.server-url-locked-hint {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  padding: 12px 16px;
  background-color: var(--mac-bg-secondary, rgba(0, 0, 0, 0.03));
  border-radius: 8px;
  border: 1px solid var(--mac-divider, rgba(0, 0, 0, 0.1));
  font-size: 0.85rem;
  color: var(--mac-text-secondary, #6b7280);
  line-height: 1.4;
}

.lock-icon {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
  color: var(--mac-text-tertiary, #9ca3af);
}

.server-url-input.is-locked {
  opacity: 0.6;
  cursor: not-allowed;
  background-color: var(--mac-bg-secondary, rgba(0, 0, 0, 0.03));
}
</style>
