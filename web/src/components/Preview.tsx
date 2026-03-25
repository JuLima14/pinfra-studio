import { useState, useCallback } from 'react'
import { Button } from '@/components/ui/button'
import {
  RefreshCw,
  Smartphone,
  Tablet,
  Monitor,
  ExternalLink,
  Loader2,
  AlertCircle,
  Power,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import type { SandboxStatus } from '@/lib/api'
import { startSandbox } from '@/lib/api'

interface PreviewProps {
  projectId: string
  sandbox: SandboxStatus | null
  onSandboxUpdate?: () => void
}

type Viewport = 'mobile' | 'tablet' | 'desktop'

const viewportConfig: Record<Viewport, { width: string; label: string }> = {
  mobile: { width: '375px', label: 'Mobile' },
  tablet: { width: '768px', label: 'Tablet' },
  desktop: { width: '100%', label: 'Desktop' },
}

export function Preview({ projectId, sandbox, onSandboxUpdate }: PreviewProps) {
  const [viewport, setViewport] = useState<Viewport>('desktop')
  const [iframeKey, setIframeKey] = useState(0)
  const [isStarting, setIsStarting] = useState(false)

  const previewUrl =
    sandbox?.status === 'running' && sandbox.port
      ? `http://localhost:${sandbox.port}`
      : sandbox?.url ?? null

  const handleRefresh = useCallback(() => {
    setIframeKey((k) => k + 1)
  }, [])

  const handleOpenNewTab = useCallback(() => {
    if (previewUrl) window.open(previewUrl, '_blank')
  }, [previewUrl])

  const handleStartSandbox = useCallback(async () => {
    setIsStarting(true)
    try {
      await startSandbox(projectId)
      onSandboxUpdate?.()
    } catch {
      // ignore
    } finally {
      setIsStarting(false)
    }
  }, [projectId, onSandboxUpdate])

  const isRunning = sandbox?.status === 'running'
  const isLoading = sandbox?.status === 'starting' || isStarting
  const isStopped = !sandbox || sandbox.status === 'stopped' || sandbox.status === 'error'

  return (
    <div className="flex h-full flex-col bg-background">
      {/* Controls bar */}
      <div className="flex items-center gap-2 border-b border-border px-3 py-2">
        <div className="flex items-center gap-1">
          <Button
            size="icon-sm"
            variant={viewport === 'mobile' ? 'secondary' : 'ghost'}
            onClick={() => setViewport('mobile')}
            title="Mobile view"
          >
            <Smartphone className="size-3.5" />
          </Button>
          <Button
            size="icon-sm"
            variant={viewport === 'tablet' ? 'secondary' : 'ghost'}
            onClick={() => setViewport('tablet')}
            title="Tablet view"
          >
            <Tablet className="size-3.5" />
          </Button>
          <Button
            size="icon-sm"
            variant={viewport === 'desktop' ? 'secondary' : 'ghost'}
            onClick={() => setViewport('desktop')}
            title="Desktop view"
          >
            <Monitor className="size-3.5" />
          </Button>
        </div>

        <div className="mx-1 h-4 w-px bg-border" />

        <Button
          size="icon-sm"
          variant="ghost"
          onClick={handleRefresh}
          disabled={!isRunning}
          title="Refresh"
        >
          <RefreshCw className="size-3.5" />
        </Button>

        <Button
          size="icon-sm"
          variant="ghost"
          onClick={handleOpenNewTab}
          disabled={!isRunning}
          title="Open in new tab"
        >
          <ExternalLink className="size-3.5" />
        </Button>

        {previewUrl && (
          <div className="flex-1 overflow-hidden px-2">
            <div className="truncate rounded-md bg-muted px-2 py-0.5 text-center text-xs text-muted-foreground">
              {previewUrl}
            </div>
          </div>
        )}
      </div>

      {/* Preview content */}
      <div className="relative flex flex-1 items-center justify-center overflow-hidden bg-muted/20">
        {isLoading && (
          <div className="flex flex-col items-center gap-3 text-muted-foreground">
            <Loader2 className="size-8 animate-spin" />
            <p className="text-sm">Starting sandbox...</p>
          </div>
        )}

        {!isLoading && isStopped && (
          <div className="flex flex-col items-center gap-4 text-center">
            <div className="rounded-full bg-muted p-4">
              <Power className="size-6 text-muted-foreground" />
            </div>
            <div>
              <p className="text-sm font-medium">Sandbox is stopped</p>
              <p className="mt-1 text-xs text-muted-foreground">
                Start the sandbox to preview your application
              </p>
            </div>
            <Button size="sm" onClick={handleStartSandbox} disabled={isStarting}>
              {isStarting ? (
                <Loader2 className="mr-1.5 size-3.5 animate-spin" />
              ) : (
                <Power className="mr-1.5 size-3.5" />
              )}
              Start Sandbox
            </Button>
            {sandbox?.error && (
              <div className="flex items-center gap-1.5 text-xs text-destructive">
                <AlertCircle className="size-3.5" />
                {sandbox.error}
              </div>
            )}
          </div>
        )}

        {isRunning && previewUrl && (
          <div
            className={cn(
              'h-full overflow-hidden transition-all duration-200',
              viewport !== 'desktop' && 'shadow-xl'
            )}
            style={{ width: viewportConfig[viewport].width }}
          >
            <iframe
              key={iframeKey}
              src={previewUrl}
              className="h-full w-full border-0 bg-white"
              title="Preview"
              sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
            />
          </div>
        )}
      </div>
    </div>
  )
}
