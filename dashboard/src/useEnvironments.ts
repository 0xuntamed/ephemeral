import { useCallback, useEffect, useRef, useState } from 'react'
import { listEnvironments } from './api'
import type { PreviewEnvironment } from './types'

const POLL_INTERVAL_MS = 5_000

export function useEnvironments() {
  const [items, setItems] = useState<PreviewEnvironment[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null)
  const activeController = useRef<AbortController | null>(null)

  const refresh = useCallback(async (quiet = false) => {
    activeController.current?.abort()
    const controller = new AbortController()
    activeController.current = controller
    if (!quiet) setRefreshing(true)
    try {
      const nextItems = await listEnvironments(controller.signal)
      setItems(nextItems)
      setError(null)
      setLastUpdated(new Date())
    } catch (caught) {
      if (!controller.signal.aborted) {
        setError(caught instanceof Error ? caught.message : 'Could not load environments')
      }
    } finally {
      if (!controller.signal.aborted) {
        setLoading(false)
        setRefreshing(false)
      }
    }
  }, [])

  useEffect(() => {
    async function loadInitialEnvironments() {
      await refresh(true)
    }

    void loadInitialEnvironments()
    const interval = window.setInterval(() => {
      if (document.visibilityState === 'visible') void refresh(true)
    }, POLL_INTERVAL_MS)
    return () => {
      window.clearInterval(interval)
      activeController.current?.abort()
    }
  }, [refresh])

  return { items, loading, refreshing, error, lastUpdated, refresh }
}
