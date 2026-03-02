import { Call } from '@wailsio/runtime'

export const restartApp = async (): Promise<void> => {
  await Call.ByName('main.AppService.RestartApp')
}
