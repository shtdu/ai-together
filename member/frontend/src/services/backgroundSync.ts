import { Call } from '@wailsio/runtime'

export type SyncResult = {
  configSynced: boolean
  usageSynced: number
  error?: string
}

export const isBackgroundSyncRunning = async (): Promise<boolean> => {
  return Call.ByName('codeswitch/services.BackgroundSyncService.IsRunning')
}

export const syncNow = async (): Promise<SyncResult> => {
  try {
    const result = await Call.ByName('codeswitch/services.BackgroundSyncService.SyncNow')
    return result
  } catch (error) {
    console.error('Failed to sync now', error)
    return {
      configSynced: false,
      usageSynced: 0,
      error: error instanceof Error ? error.message : String(error),
    }
  }
}

export const setConfigSyncInterval = async (minutes: number): Promise<void> => {
  // Convert minutes to nanoseconds (Go time.Duration)
  const nanos = minutes * 60 * 1000 * 1000 * 1000
  return Call.ByName('codeswitch/services.BackgroundSyncService.SetConfigSyncInterval', nanos)
}

export const setUsageSyncInterval = async (minutes: number): Promise<void> => {
  // Convert minutes to nanoseconds (Go time.Duration)
  const nanos = minutes * 60 * 1000 * 1000 * 1000
  return Call.ByName('codeswitch/services.BackgroundSyncService.SetUsageSyncInterval', nanos)
}
