import { useState, useEffect } from 'react'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Folder,
  FolderOpen,
  File,
  FileCode,
  FileText,
  Image,
  Loader2,
  AlertCircle,
  ChevronRight,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import type { FileEntry } from '@/lib/api'
import { getFiles } from '@/lib/api'

interface FileTreeProps {
  projectId: string
  onFileSelect: (path: string) => void
  selectedFile?: string | null
}

function getFileIcon(name: string) {
  const ext = name.split('.').pop()?.toLowerCase() ?? ''
  const codeExts = ['ts', 'tsx', 'js', 'jsx', 'go', 'py', 'rs', 'java', 'c', 'cpp', 'h', 'cs', 'php', 'rb']
  const textExts = ['md', 'txt', 'json', 'yml', 'yaml', 'toml', 'env', 'sh', 'bash', 'css', 'html', 'xml']
  const imageExts = ['png', 'jpg', 'jpeg', 'gif', 'svg', 'webp', 'ico']

  if (codeExts.includes(ext)) return FileCode
  if (textExts.includes(ext)) return FileText
  if (imageExts.includes(ext)) return Image
  return File
}

interface FileNodeProps {
  entry: FileEntry
  depth: number
  onFileSelect: (path: string) => void
  selectedFile?: string | null
}

function FileNode({ entry, depth, onFileSelect, selectedFile }: FileNodeProps) {
  const [expanded, setExpanded] = useState(depth === 0)
  const isDir = entry.isDir === true
  const hasChildren = isDir && entry.children && entry.children.length > 0
  const FileIcon = isDir ? (expanded ? FolderOpen : Folder) : getFileIcon(entry.name)

  const handleClick = () => {
    if (isDir) {
      setExpanded((v) => !v)
    } else {
      onFileSelect(entry.path)
    }
  }

  return (
    <div>
      <button
        type="button"
        onClick={handleClick}
        className={cn(
          'flex w-full items-center gap-1.5 rounded-md px-2 py-1 text-left text-xs transition-colors',
          'hover:bg-muted/50',
          !isDir && selectedFile === entry.path && 'bg-muted text-foreground',
          !isDir && selectedFile !== entry.path && 'text-muted-foreground hover:text-foreground',
          isDir && 'font-medium text-foreground cursor-pointer'
        )}
        style={{ paddingLeft: `${8 + depth * 16}px` }}
      >
        <FileIcon
          className={cn(
            'size-3.5 shrink-0',
            isDir ? 'text-yellow-400/80' : 'text-muted-foreground'
          )}
        />
        <span className="truncate">{entry.name}</span>
        {isDir && hasChildren && (
          <ChevronRight
            className={cn(
              'ml-auto size-3 shrink-0 text-muted-foreground/50 transition-transform',
              expanded && 'rotate-90'
            )}
          />
        )}
      </button>

      {isDir && expanded && hasChildren && (
        <div>
          {entry.children!.map((child) => (
            <FileNode
              key={child.path}
              entry={child}
              depth={depth + 1}
              onFileSelect={onFileSelect}
              selectedFile={selectedFile}
            />
          ))}
        </div>
      )}
    </div>
  )
}

export function FileTree({ projectId, onFileSelect, selectedFile }: FileTreeProps) {
  const [files, setFiles] = useState<FileEntry[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setIsLoading(true)
    setError(null)
    getFiles(projectId)
      .then(setFiles)
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load files'))
      .finally(() => setIsLoading(false))
  }, [projectId])

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="size-5 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex h-full flex-col items-center justify-center gap-2 text-muted-foreground">
        <AlertCircle className="size-5" />
        <p className="text-xs">{error}</p>
      </div>
    )
  }

  if (files.length === 0) {
    return (
      <div className="flex h-full flex-col items-center justify-center gap-2 text-muted-foreground">
        <Folder className="size-6 opacity-40" />
        <p className="text-xs">No files yet</p>
      </div>
    )
  }

  return (
    <ScrollArea className="h-full">
      <div className="py-2">
        {files.map((entry) => (
          <FileNode
            key={entry.path}
            entry={entry}
            depth={0}
            onFileSelect={onFileSelect}
            selectedFile={selectedFile}
          />
        ))}
      </div>
    </ScrollArea>
  )
}
