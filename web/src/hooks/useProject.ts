import { useState, useEffect, useCallback, useRef } from 'react'
import type { Project, SandboxStatus } from '@/lib/api'
import { getProject, getSandboxStatus } from '@/lib/api'

interface UseProjectReturn {
  project: Project | null
  sandbox: SandboxStatus | null
  isLoading: boolean
  error: string | null
  refresh: () => void
}

export function useProject(projectId: string): UseProjectReturn {
  const [project, setProject] = useState<Project | null>(null)
  const [sandbox, setSandbox] = useState<SandboxStatus | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const pollTimerRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const mountedRef = useRef(true)

  const fetchData = useCallback(async () => {
    if (!projectId) return
    try {
      const [proj, sb] = await Promise.all([
        getProject(projectId),
        getSandboxStatus(projectId).catch(() => null),
      ])
      if (!mountedRef.current) return
      setProject(proj)
      setSandbox(sb)
      setError(null)
    } catch (err) {
      if (!mountedRef.current) return
      setError(err instanceof Error ? err.message : 'Failed to load project')
    } finally {
      if (mountedRef.current) setIsLoading(false)
    }
  }, [projectId])

  const fetchSandbox = useCallback(async () => {
    if (!projectId) return
    try {
      const sb = await getSandboxStatus(projectId)
      if (!mountedRef.current) return
      setSandbox(sb)
    } catch {
      // ignore sandbox poll errors
    }
  }, [projectId])

  // Poll sandbox every 5s when status is "starting"
  useEffect(() => {
    if (sandbox?.status === 'starting') {
      pollTimerRef.current = setInterval(fetchSandbox, 5000)
    } else {
      if (pollTimerRef.current) {
        clearInterval(pollTimerRef.current)
        pollTimerRef.current = null
      }
    }
    return () => {
      if (pollTimerRef.current) {
        clearInterval(pollTimerRef.current)
        pollTimerRef.current = null
      }
    }
  }, [sandbox?.status, fetchSandbox])

  useEffect(() => {
    mountedRef.current = true
    fetchData()
    return () => {
      mountedRef.current = false
    }
  }, [fetchData])

  return { project, sandbox, isLoading, error, refresh: fetchData }
}
