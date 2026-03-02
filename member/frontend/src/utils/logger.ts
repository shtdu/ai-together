/**
 * Simple logger utility for the frontend.
 * Provides consistent logging across the application.
 */

type LogLevel = 'info' | 'warn' | 'error' | 'debug'

interface LoggerConfig {
  enabled: boolean
  level: LogLevel
  prefix: string
}

// Default configuration - can be modified at runtime
const config: LoggerConfig = {
  enabled: true,
  level: 'info',
  prefix: '[App]',
}

/**
 * Checks if a log level should be output based on current config
 */
function shouldLog(level: LogLevel): boolean {
  if (!config.enabled) return false

  const levels: LogLevel[] = ['debug', 'info', 'warn', 'error']
  const currentLevelIndex = levels.indexOf(config.level)
  const messageLevelIndex = levels.indexOf(level)

  return messageLevelIndex >= currentLevelIndex
}

/**
 * Formats a log message with prefix and additional context
 */
function formatMessage(level: LogLevel, message: string, ...args: unknown[]): string {
  const timestamp = new Date().toISOString().split('T')[1].split('.')[0]
  return `${config.prefix} [${level.toUpperCase()}] ${timestamp} ${message}`
}

export const logger = {
  /**
   * Logs an informational message
   */
  info(message: string, ...args: unknown[]): void {
    if (shouldLog('info')) {
      console.log(formatMessage('info', message), ...args)
    }
  },

  /**
   * Logs a warning message
   */
  warn(message: string, ...args: unknown[]): void {
    if (shouldLog('warn')) {
      console.warn(formatMessage('warn', message), ...args)
    }
  },

  /**
   * Logs an error message
   */
  error(message: string, ...args: unknown[]): void {
    if (shouldLog('error')) {
      console.error(formatMessage('error', message), ...args)
    }
  },

  /**
   * Logs a debug message (only shown when level is 'debug')
   */
  debug(message: string, ...args: unknown[]): void {
    if (shouldLog('debug')) {
      console.debug(formatMessage('debug', message), ...args)
    }
  },

  /**
   * Updates logger configuration
   */
  configure(newConfig: Partial<LoggerConfig>): void {
    Object.assign(config, newConfig)
  },

  /**
   * Gets current configuration
   */
  getConfig(): LoggerConfig {
    return { ...config }
  },
}

// Export default logger for convenience
export default logger
