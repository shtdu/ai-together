import { Call } from '@wailsio/runtime'

export type ConfigImportStatus = {
  config_exists: boolean
  config_path: string
  pending_providers: boolean
  pending_mcp: boolean
  pending_provider_count: number
  pending_mcp_count: number
}

export type ConfigImportResult = {
  status: ConfigImportStatus
  imported_providers: number
  imported_mcp: number
}

const emptyStatus: ConfigImportStatus = {
  config_exists: false,
  config_path: '',
  pending_providers: false,
  pending_mcp: false,
  pending_provider_count: 0,
  pending_mcp_count: 0
}

export const fetchConfigImportStatus = async (): Promise<ConfigImportStatus> => {
  const response = await Call.ByName('codeswitch/services.ImportService.GetStatus')
  return (response as ConfigImportStatus) ?? emptyStatus
}

export const fetchConfigImportStatusForFile = async (
  path: string,
): Promise<ConfigImportStatus> => {
  const response = await Call.ByName('codeswitch/services.ImportService.GetStatusForFile', path)
  return response as ConfigImportStatus
}

export const importFromCcSwitch = async (): Promise<ConfigImportResult> => {
  const response = await Call.ByName('codeswitch/services.ImportService.ImportAll')
  return response as ConfigImportResult
}

export const importFromCustomFile = async (path: string): Promise<ConfigImportResult> => {
  const response = await Call.ByName('codeswitch/services.ImportService.ImportFromFile', path)
  return response as ConfigImportResult
}
