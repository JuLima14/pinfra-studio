import { useState, useEffect } from 'react'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Button } from '@/components/ui/button'
import { Loader2, FileCode } from 'lucide-react'
import { cn } from '@/lib/utils'
import { getFileContent } from '@/lib/api'
import type { UIMessage } from '@/hooks/useChat'

interface CodeViewProps {
  projectId: string
  messages: UIMessage[]
  selectedFile?: string | null
  onFileSelect?: (path: string) => void
}

interface ModifiedFile {
  path: string
  content?: string
}

function extractModifiedFiles(messages: UIMessage[]): ModifiedFile[] {
  const fileMap = new Map<string, ModifiedFile>()
  for (const msg of messages) {
    if (msg.role !== 'assistant' || !msg.toolUses) continue
    for (const tool of msg.toolUses) {
      if (tool.toolName === 'Write' || tool.toolName === 'Edit') {
        const path =
          (tool.toolInput?.path as string) ??
          (tool.toolInput?.file_path as string) ??
          ''
        if (path) {
          fileMap.set(path, {
            path,
            content: tool.toolInput?.content as string | undefined,
          })
        }
      }
    }
  }
  return Array.from(fileMap.values())
}

function getLanguageClass(path: string): string {
  const ext = path.split('.').pop()?.toLowerCase() ?? ''
  const map: Record<string, string> = {
    ts: 'language-typescript',
    tsx: 'language-tsx',
    js: 'language-javascript',
    jsx: 'language-jsx',
    py: 'language-python',
    go: 'language-go',
    rs: 'language-rust',
    json: 'language-json',
    css: 'language-css',
    html: 'language-html',
    md: 'language-markdown',
    yml: 'language-yaml',
    yaml: 'language-yaml',
    sh: 'language-bash',
    toml: 'language-toml',
  }
  return map[ext] ?? 'language-text'
}

function CodeDisplay({ content, path }: { content: string; path: string }) {
  const lines = content.split('\n')
  const langClass = getLanguageClass(path)

  return (
    <div className="flex h-full overflow-hidden font-mono text-xs">
      {/* Line numbers */}
      <div className="select-none border-r border-border bg-muted/30 py-4 pr-3 pl-4 text-right text-muted-foreground/50">
        {lines.map((_, i) => (
          <div key={i} className="leading-5">
            {i + 1}
          </div>
        ))}
      </div>
      {/* Code content */}
      <ScrollArea className="flex-1">
        <pre className={cn('p-4 leading-5 text-foreground', langClass)}>
          <code>{content}</code>
        </pre>
      </ScrollArea>
    </div>
  )
}

export function CodeView({
  projectId,
  messages,
  selectedFile,
  onFileSelect,
}: CodeViewProps) {
  const modifiedFiles = extractModifiedFiles(messages)
  const [activeFile, setActiveFile] = useState<string | null>(
    selectedFile ?? modifiedFiles[0]?.path ?? null
  )
  const [fileContent, setFileContent] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  // Sync external selectedFile
  useEffect(() => {
    if (selectedFile) setActiveFile(selectedFile)
  }, [selectedFile])

  // Auto-select first modified file if none selected
  useEffect(() => {
    if (!activeFile && modifiedFiles.length > 0) {
      setActiveFile(modifiedFiles[0]?.path ?? null)
    }
  }, [modifiedFiles, activeFile])

  // Load file content when active file changes
  useEffect(() => {
    if (!activeFile) return

    // Check if we have inline content from tool_use
    const modified = modifiedFiles.find((f) => f.path === activeFile)
    if (modified?.content !== undefined) {
      setFileContent(modified.content)
      return
    }

    // Fetch from API
    setIsLoading(true)
    setFileContent(null)
    getFileContent(projectId, activeFile)
      .then((res) => setFileContent(res.content))
      .catch(() => setFileContent(null))
      .finally(() => setIsLoading(false))
  }, [activeFile, projectId]) // eslint-disable-line react-hooks/exhaustive-deps

  const handleTabClick = (path: string) => {
    setActiveFile(path)
    onFileSelect?.(path)
  }

  if (modifiedFiles.length === 0 && !selectedFile) {
    return (
      <div className="flex h-full flex-col items-center justify-center gap-3 text-muted-foreground">
        <FileCode className="size-8 opacity-40" />
        <p className="text-sm">
          Modified files will appear here as Claude edits them
        </p>
      </div>
    )
  }

  const displayFiles =
    selectedFile && !modifiedFiles.find((f) => f.path === selectedFile)
      ? [...modifiedFiles, { path: selectedFile }]
      : modifiedFiles

  return (
    <div className="flex h-full flex-col">
      {/* Tab bar */}
      {displayFiles.length > 0 && (
        <div className="flex items-center gap-1 overflow-x-auto border-b border-border px-2 py-1.5">
          {displayFiles.map((f) => (
            <button
              key={f.path}
              type="button"
              onClick={() => handleTabClick(f.path)}
              className={cn(
                'flex shrink-0 items-center gap-1.5 rounded-md px-2.5 py-1 text-xs transition-colors',
                activeFile === f.path
                  ? 'bg-muted text-foreground'
                  : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground'
              )}
            >
              <FileCode className="size-3" />
              <span className="max-w-[120px] truncate">{f.path.split('/').pop()}</span>
            </button>
          ))}
        </div>
      )}

      {/* Code content */}
      <div className="flex-1 overflow-hidden">
        {isLoading ? (
          <div className="flex h-full items-center justify-center">
            <Loader2 className="size-5 animate-spin text-muted-foreground" />
          </div>
        ) : fileContent !== null ? (
          <ScrollArea className="h-full">
            <CodeDisplay content={fileContent} path={activeFile ?? ''} />
          </ScrollArea>
        ) : activeFile ? (
          <div className="flex h-full flex-col items-center justify-center gap-2 text-muted-foreground">
            <FileCode className="size-6 opacity-40" />
            <p className="text-xs">Unable to load file</p>
            <Button
              size="xs"
              variant="outline"
              onClick={() => {
                setIsLoading(true)
                getFileContent(projectId, activeFile)
                  .then((res) => setFileContent(res.content))
                  .catch(() => {})
                  .finally(() => setIsLoading(false))
              }}
            >
              Retry
            </Button>
          </div>
        ) : null}
      </div>
    </div>
  )
}
