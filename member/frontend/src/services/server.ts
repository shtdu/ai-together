import { Call } from '@wailsio/runtime'

export type ServerConfig = {
  server_url: string
}

const DEFAULT_CONFIG: ServerConfig = {
  server_url: '',
}

export const getServerURL = async (): Promise<string> => {
  try {
    const data = await Call.ByName('codeswitch/services.ServerConfigService.GetServerURL')
    return data ?? DEFAULT_CONFIG.server_url
  } catch (error) {
    console.error('Failed to get server URL', error)
    return DEFAULT_CONFIG.server_url
  }
}

export const setServerURL = async (url: string): Promise<void> => {
  return Call.ByName('codeswitch/services.ServerConfigService.SetServerURL', url)
}

export const testConnection = async (url?: string): Promise<void> => {
  return Call.ByName('codeswitch/services.ServerConfigService.TestConnection', url ?? '')
}
