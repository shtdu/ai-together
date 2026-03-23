import { GenerateDailyReport } from '../../bindings/codeswitch/services/reportservice'
import type {
  DailyReport,
  SessionInfo,
  ToolReport,
  DayDetails,
  SessionStats,
  ToolUsageItem,
  PromptStats
} from '../../bindings/codeswitch/services/models'

// Data retention window in days (includes today)
export const RETENTION_DAYS = 14

// Re-export types for convenience
export type { DailyReport, SessionInfo, ToolReport, DayDetails, SessionStats, ToolUsageItem, PromptStats }

export interface ReportQuery {
  date: string          // YYYY-MM-DD format
  toolName?: string     // Optional: claude, codex, opencode (default: empty = all tools)
}

/**
 * Fetches a daily activity report for the specified date and tool
 * @param query - Report query parameters
 * @returns Daily report data
 */
export const fetchDailyReport = async (query: ReportQuery): Promise<DailyReport> => {
  const date = query.date ?? new Date().toISOString().split('T')[0]
  const toolName = query.toolName ?? ''

  const result = await GenerateDailyReport(date, toolName)
  if (!result) {
    throw new Error('No report data returned')
  }
  return result
}

/**
 * Helper function to format duration in milliseconds to human-readable string
 * @param durationMs - Duration in milliseconds
 * @returns Formatted duration string (e.g., "6h 44m", "22m 9s", "45s")
 */
export const formatDuration = (durationMs: number): string => {
  const hours = Math.floor(durationMs / (1000 * 60 * 60))
  const minutes = Math.floor((durationMs % (1000 * 60 * 60)) / (1000 * 60))
  const seconds = Math.floor((durationMs % (1000 * 60)) / 1000)

  if (hours > 0) {
    return `${hours}h ${minutes}m`
  } else if (minutes > 0) {
    return `${minutes}m ${seconds}s`
  } else {
    return `${seconds}s`
  }
}

/**
 * Helper function to format time from ISO timestamp to local time string
 * @param isoTimestamp - ISO timestamp string
 * @returns Formatted time string (HH:MM:SS)
 */
export const formatTime = (isoTimestamp: string): string => {
  const date = new Date(isoTimestamp)
  return date.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  })
}

/**
 * Helper function to calculate intensity level based on events and duration
 * @param eventCount - Number of events in the session
 * @param durationMs - Session duration in milliseconds
 * @returns Intensity level (1-5)
 */
export const calculateIntensity = (eventCount: number, durationMs: number): number => {
  if (durationMs === 0) return 0

  const durationMinutes = durationMs / (1000 * 60)
  const eventsPerMinute = eventCount / durationMinutes

  // Intensity thresholds (events per minute)
  if (eventsPerMinute >= 10) return 5
  if (eventsPerMinute >= 5) return 4
  if (eventsPerMinute >= 2) return 3
  if (eventsPerMinute >= 1) return 2
  return 1
}

/**
 * Helper function to get today's date in YYYY-MM-DD format
 * @returns Today's date string
 */
export const getTodayDate = (): string => {
  return new Date().toISOString().split('T')[0]
}

/**
 * Helper function to get yesterday's date in YYYY-MM-DD format
 * @returns Yesterday's date string
 */
export const getYesterdayDate = (): string => {
  const date = new Date()
  date.setDate(date.getDate() - 1)
  return date.toISOString().split('T')[0]
}

/**
 * Helper function to get previous day's date in YYYY-MM-DD format
 * @param currentDate - Current date string in YYYY-MM-DD format
 * @returns Previous day's date string
 */
export const getPreviousDay = (currentDate: string): string => {
  const date = new Date(currentDate)
  date.setDate(date.getDate() - 1)
  return date.toISOString().split('T')[0]
}

/**
 * Helper function to get next day's date in YYYY-MM-DD format
 * @param currentDate - Current date string in YYYY-MM-DD format
 * @returns Next day's date string
 */
export const getNextDay = (currentDate: string): string => {
  const date = new Date(currentDate)
  date.setDate(date.getDate() + 1)
  return date.toISOString().split('T')[0]
}

/**
 * Helper function to get the minimum date in YYYY-MM-DD format
 * This matches the client's data retention policy (RETENTION_DAYS including today)
 * @returns Minimum date string (RETENTION_DAYS - 1 days ago)
 */
export const getMinDate = (): string => {
  const date = new Date()
  date.setDate(date.getDate() - (RETENTION_DAYS - 1))
  return date.toISOString().split('T')[0]
}
