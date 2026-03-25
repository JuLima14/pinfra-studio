import { useState, useEffect, useRef, useCallback } from 'react'
import type { StreamChunk } from '@/lib/api'
import { getSSEUrl } from '@/lib/api'

interface UseSSEOptions {
  projectId: string
  chatId: string
  enabled: boolean
  onChunk?: (chunk: StreamChunk) => void
  onDone?: () => void
  onError?: (error: string) => void
}

interface UseSSEReturn {
  chunks: StreamChunk[]
  isConnected: boolean
  error: string | null
  reset: () => void
}

export function useSSE({
  projectId,
  chatId,
  enabled,
  onChunk,
  onDone,
  onError,
}: UseSSEOptions): UseSSEReturn {
  const [chunks, setChunks] = useState<StreamChunk[]>([])
  const [isConnected, setIsConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const esRef = useRef<EventSource | null>(null)
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const reconnectCountRef = useRef(0)
  const mountedRef = useRef(true)

  const reset = useCallback(() => {
    setChunks([])
    setError(null)
  }, [])

  const cleanup = useCallback(() => {
    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current)
      reconnectTimerRef.current = null
    }
    if (esRef.current) {
      esRef.current.close()
      esRef.current = null
    }
  }, [])

  const connect = useCallback(() => {
    if (!mountedRef.current || !enabled || !projectId || !chatId) return

    cleanup()

    const url = getSSEUrl(projectId, chatId)
    const es = new EventSource(url)
    esRef.current = es

    es.onopen = () => {
      if (!mountedRef.current) return
      setIsConnected(true)
      setError(null)
      reconnectCountRef.current = 0
    }

    es.onmessage = (event) => {
      if (!mountedRef.current) return
      try {
        const chunk = JSON.parse(event.data) as StreamChunk
        setChunks((prev) => [...prev, chunk])
        onChunk?.(chunk)

        if (chunk.type === 'done') {
          onDone?.()
          cleanup()
          setIsConnected(false)
        } else if (chunk.type === 'error') {
          const errMsg = chunk.error ?? 'Stream error'
          setError(errMsg)
          onError?.(errMsg)
          cleanup()
          setIsConnected(false)
        }
      } catch {
        // ignore parse errors
      }
    }

    es.onerror = () => {
      if (!mountedRef.current) return
      cleanup()
      setIsConnected(false)

      // Exponential backoff reconnect (max 30s)
      const delay = Math.min(1000 * 2 ** reconnectCountRef.current, 30000)
      reconnectCountRef.current += 1

      reconnectTimerRef.current = setTimeout(() => {
        if (mountedRef.current && enabled) {
          connect()
        }
      }, delay)
    }
  }, [projectId, chatId, enabled, cleanup, onChunk, onDone, onError])

  useEffect(() => {
    mountedRef.current = true

    if (enabled) {
      connect()
    } else {
      cleanup()
      setIsConnected(false)
    }

    return () => {
      mountedRef.current = false
      cleanup()
    }
  }, [enabled, connect, cleanup])

  return { chunks, isConnected, error, reset }
}
