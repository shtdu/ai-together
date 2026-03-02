import { Call } from '@wailsio/runtime'

/**
 * PermissionService provides permission checking methods for the frontend
 */

/**
 * Checks if the current user has admin (manager) role
 */
export const isAdmin = async (): Promise<boolean> => {
  try {
    return await Call.ByName('codeswitch/services.PermissionService.IsAdmin')
  } catch (error) {
    console.error('Failed to check admin status', error)
    return false
  }
}

/**
 * Gets the current user's role
 */
export const getUserRole = async (): Promise<string> => {
  try {
    return await Call.ByName('codeswitch/services.PermissionService.GetUserRole')
  } catch (error) {
    console.error('Failed to get user role', error)
    return 'unknown'
  }
}

/**
 * Checks if the user can edit settings
 */
export const canEditSettings = async (): Promise<boolean> => {
  try {
    return await Call.ByName('codeswitch/services.PermissionService.CanEditSettings')
  } catch (error) {
    console.error('Failed to check edit permissions', error)
    return false
  }
}

/**
 * Checks if the user can push master copy to server
 */
export const canPushMasterCopy = async (): Promise<boolean> => {
  try {
    return await Call.ByName('codeswitch/services.PermissionService.CanPushMasterCopy')
  } catch (error) {
    console.error('Failed to check push permissions', error)
    return false
  }
}
