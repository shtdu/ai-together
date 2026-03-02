import { CurrentVersion } from '../../bindings/codeswitch/versionservice'

export const fetchCurrentVersion = async (): Promise<string> => {
  const version = await CurrentVersion()
  return version ?? ''
}
