import {
  createContext,
  useState,
  useEffect,
  type ReactNode,
} from 'react'
import { setupApi } from '../api/setup'

interface SetupState {
  setupRequired: boolean | null
  isLoading: boolean
  error: string | null
}

interface SetupContextValue extends SetupState {
  checkSetupStatus: () => Promise<void>
}

// eslint-disable-next-line react-refresh/only-export-components
export const SetupContext = createContext<SetupContextValue | undefined>(undefined)

interface SetupProviderProps {
  children: ReactNode
}

export function SetupProvider({ children }: SetupProviderProps) {
  const [state, setState] = useState<SetupState>({
    setupRequired: null,
    isLoading: true,
    error: null,
  })

  const checkSetupStatus = async () => {
    try {
      setState((prev) => ({ ...prev, isLoading: true, error: null }))
      const response = await setupApi.getStatus()
      setState({
        setupRequired: response.setup_required,
        isLoading: false,
        error: null,
      })
    } catch (error) {
      setState({
        setupRequired: null,
        isLoading: false,
        error: error instanceof Error ? error.message : 'Failed to check setup status',
      })
    }
  }

  useEffect(() => {
    checkSetupStatus()
  }, [])

  return (
    <SetupContext.Provider
      value={{
        ...state,
        checkSetupStatus,
      }}
    >
      {children}
    </SetupContext.Provider>
  )
}
