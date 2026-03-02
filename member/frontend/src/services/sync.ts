import { Call } from '@wailsio/runtime'

export type SyncState = {
  last_sync_time: string
}

export const syncProviders = async (): Promise<void> => {
  return Call.ByName('codeswitch/services.ConfigSyncService.SyncProviders')
}

export const syncProviderType = async (kind: string): Promise<void> => {
  return Call.ByName('codeswitch/services.ConfigSyncService.SyncProviderType', kind)
}

export const getLastSyncTime = async (): Promise<string> => {
  try {
    const data = await Call.ByName('codeswitch/services.ConfigSyncService.GetLastSyncTime')
    return data ?? ''
  } catch (error) {
    console.error('Failed to get last sync time', error)
    return ''
  }
}

export const syncUsageStats = async (): Promise<number> => {
  return Call.ByName('codeswitch/services.UsageSyncService.SyncUsageStats')
}

export const getPendingRecordCount = async (): Promise<number> => {
  return Call.ByName('codeswitch/services.UsageSyncService.GetPendingRecordCount')
}

/**
 * Pushes local provider configurations to the server as master copy
 * Requires admin/manager role on the server
 */
export const pushProvidersToServer = async (): Promise<void> => {
  return Call.ByName('codeswitch/services.ConfigSyncService.PushProvidersToServer')
}

/**
 * Checks if any provider type has local changes pending push
 */
export const hasLocalChanges = async (): Promise<boolean> => {
  try {
    return await Call.ByName('codeswitch/services.ProviderService.HasAnyLocalChanges') ?? false
  } catch (error) {
    console.error('Failed to check local changes', error)
    return false
  }
}

/**
 * Checks if a specific provider type has local changes pending push
 * @param kind - Provider type: "claude", "codex", or "opencode"
 */
export const hasLocalChangesByKind = async (kind: string): Promise<boolean> => {
  try {
    return await Call.ByName('codeswitch/services.ProviderService.HasLocalChanges', kind) ?? false
  } catch (error) {
    console.error(`Failed to check local changes for ${kind}`, error)
    return false
  }
}

/**
 * Deletes all local usage data from the database
 * Used during logout to clean up all usage records
 */
export const deleteAllLocalUsageData = async (): Promise<number> => {
  return Call.ByName('codeswitch/services.UsageSyncService.DeleteAllLocalUsageData')
}

/**
 * Deletes all provider configuration files
 * Used during logout to clean up all provider settings
 */
export const deleteAllProviders = async (): Promise<void> => {
  return Call.ByName('codeswitch/services.ProviderService.DeleteAllProviders')
}

/**
 * Clears the sync state (resets last sync time to zero)
 * Used during logout to ensure fresh sync state on next login
 */
export const clearSyncState = async (): Promise<void> => {
  return Call.ByName('codeswitch/services.ConfigService.ClearSyncState')
}
