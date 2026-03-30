import { ref, onMounted, onUnmounted } from 'vue'

/**
 * Auto-refresh composable for analytics pages.
 * Refreshes data at a fixed interval, pauses when the tab is hidden,
 * and guards against stacking requests during loading.
 *
 * Per spec TA-04-102: team analytics should update every 30 seconds.
 */
export function useAutoRefresh(fetchFn: () => Promise<void>, isLoading: () => boolean, intervalMs = 30_000) {
  const lastUpdated = ref<Date | null>(null)
  let timer: ReturnType<typeof setInterval> | null = null

  async function tick() {
    if (isLoading()) return
    try {
      await fetchFn()
      lastUpdated.value = new Date()
    } catch {
      // errors are handled inside fetchFn already
    }
  }

  function start() {
    stop()
    timer = setInterval(tick, intervalMs)
  }

  function stop() {
    if (timer) {
      clearInterval(timer)
      timer = null
    }
  }

  function handleVisibility() {
    if (document.visibilityState === 'visible') {
      // Tab became visible — refresh immediately then resume interval
      tick()
      start()
    } else {
      stop()
    }
  }

  onMounted(() => {
    // First data load sets the initial timestamp
    lastUpdated.value = new Date()
    start()
    document.addEventListener('visibilitychange', handleVisibility)
  })

  onUnmounted(() => {
    stop()
    document.removeEventListener('visibilitychange', handleVisibility)
  })

  return { lastUpdated, start, stop }
}
