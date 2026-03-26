import { useEffect, useRef, useState } from 'react'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { MessageInput } from '@/components/MessageInput'
import {
  MessageSquare,
  Plus,
  ChevronDown,
  ChevronRight,
  Search,
  FileEdit,
  Terminal,
  FileCode,
  Loader2,
  Brain,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import type { UIMessage, ToolUse } from '@/hooks/useChat'
import type { Chat as ChatType } from '@/lib/api'

// Tools that add noise to the chat — hide from UI
const HIDDEN_TOOLS = new Set([
  'TodoWrite',
  'TodoRead',
  'Agent',
  'Task',
  'WebSearch',
  'WebFetch',
  'LSP',
  'NotebookEdit',
  'EnterPlanMode',
  'ExitPlanMode',
])

interface ChatProps {
  messages: UIMessage[]
  chats: ChatType[]
  activeChat: ChatType | null
  isGenerating: boolean
  isLoading: boolean
  onSend: (content: string) => void
  onCancel: () => void
  onSwitchChat: (chatId: string) => void
  onCreateChat: () => void
}

function getToolIcon(toolName: string) {
  const name = toolName.toLowerCase()
  if (name === 'read' || name === 'grep' || name === 'glob') return Search
  if (name === 'write' || name === 'edit') return FileEdit
  if (name === 'bash') return Terminal
  if (name.includes('notebook')) return FileCode
  return FileCode
}

function getToolLabel(toolUse: ToolUse): string {
  const filePath =
    (toolUse.toolInput?.path as string) ??
    (toolUse.toolInput?.file_path as string) ??
    (toolUse.toolInput?.command as string) ??
    ''

  if (!filePath) return toolUse.toolName

  // Show just filename for Read/Write/Edit
  const name = toolUse.toolName.toLowerCase()
  if (name === 'read' || name === 'write' || name === 'edit' || name === 'grep' || name === 'glob') {
    const shortPath = filePath.split('/').slice(-2).join('/')
    return `${toolUse.toolName} ${shortPath}`
  }
  // For Bash, truncate command
  if (name === 'bash') {
    const cmd = filePath.length > 60 ? filePath.slice(0, 60) + '…' : filePath
    return cmd
  }
  return `${toolUse.toolName} ${filePath}`
}

function ToolUseItem({ toolUse }: { toolUse: ToolUse }) {
  const [expanded, setExpanded] = useState(false)
  const Icon = getToolIcon(toolUse.toolName)
  const label = getToolLabel(toolUse)

  return (
    <div>
      <button
        type="button"
        onClick={() => setExpanded((v) => !v)}
        className="flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-left text-sm text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground"
      >
        <Icon className="size-4 shrink-0 text-muted-foreground/70" />
        <span className="flex-1 truncate">{label}</span>
        <ChevronRight
          className={cn(
            'size-3.5 shrink-0 text-muted-foreground/50 transition-transform',
            expanded && 'rotate-90'
          )}
        />
      </button>
      {expanded && toolUse.toolInput && (
        <div className="mb-1 ml-9 mr-3 rounded-md bg-muted/30 px-3 py-2">
          <pre className="whitespace-pre-wrap break-all font-mono text-[11px] text-muted-foreground">
            {JSON.stringify(toolUse.toolInput, null, 2)}
          </pre>
        </div>
      )}
    </div>
  )
}

function ThinkingIndicator({ text }: { text: string }) {
  return (
    <button
      type="button"
      className="flex items-center gap-2.5 rounded-lg px-3 py-2 text-left text-sm text-muted-foreground"
    >
      <Brain className="size-4 shrink-0 text-purple-400/70" />
      <span>{text}</span>
    </button>
  )
}

function MessageBubble({ message }: { message: UIMessage }) {
  const isUser = message.role === 'user'
  const hasToolUses = message.toolUses && message.toolUses.length > 0

  if (isUser) {
    return (
      <div className="flex justify-end">
        <div className="max-w-[85%] rounded-2xl bg-primary px-4 py-2.5 text-sm text-primary-foreground">
          <p className="whitespace-pre-wrap leading-relaxed">{message.content}</p>
        </div>
      </div>
    )
  }

  // Assistant message: tool uses rendered as separate items OUTSIDE the bubble
  return (
    <div className="space-y-1">
      {/* Tool uses as v0-style list items (filter out noise) */}
      {hasToolUses && (
        <div className="-mx-1">
          {message.toolUses!
            .filter((tu) => !HIDDEN_TOOLS.has(tu.toolName))
            .map((tu) => (
              <ToolUseItem key={tu.id} toolUse={tu} />
            ))}
        </div>
      )}

      {/* Text content — no background, plain text like v0 */}
      {message.content && (
        <div className="px-3 text-sm text-foreground">
          <p className="whitespace-pre-wrap leading-relaxed">
            {message.content}
            {message.isStreaming && (
              <span className="ml-0.5 inline-block h-3.5 w-0.5 animate-pulse bg-current align-middle" />
            )}
          </p>
        </div>
      )}

      {/* Streaming with no content yet — show thinking */}
      {message.isStreaming && !message.content && !hasToolUses && (
        <ThinkingIndicator text="Thinking..." />
      )}
    </div>
  )
}

export function Chat({
  messages,
  chats,
  activeChat,
  isGenerating,
  isLoading,
  onSend,
  onCancel,
  onSwitchChat,
  onCreateChat,
}: ChatProps) {
  const [inputValue, setInputValue] = useState('')
  const [showChatMenu, setShowChatMenu] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)

  // Auto-scroll to bottom on new messages
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  return (
    <div className="flex h-full flex-col">
      {/* Chat selector header */}
      <div className="flex items-center gap-2 border-b border-border px-3 py-2">
        <div className="relative flex-1">
          <button
            type="button"
            onClick={() => setShowChatMenu((v) => !v)}
            className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-muted"
          >
            <MessageSquare className="size-3.5 text-muted-foreground" />
            <span className="flex-1 truncate text-left font-medium">
              {activeChat?.title ?? 'Chat'}
            </span>
            <ChevronDown className="size-3.5 text-muted-foreground" />
          </button>

          {showChatMenu && chats.length > 0 && (
            <div className="absolute left-0 top-full z-50 mt-1 w-full rounded-lg border border-border bg-popover py-1 shadow-lg">
              {chats.map((chat) => (
                <button
                  key={chat.id}
                  type="button"
                  onClick={() => {
                    onSwitchChat(chat.id)
                    setShowChatMenu(false)
                  }}
                  className={cn(
                    'flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:bg-muted',
                    chat.id === activeChat?.id && 'bg-muted'
                  )}
                >
                  <MessageSquare className="size-3.5 text-muted-foreground" />
                  <span className="flex-1 truncate">{chat.title ?? 'Chat'}</span>
                  {chat.isActive && (
                    <Badge variant="secondary" className="text-xs">
                      Active
                    </Badge>
                  )}
                </button>
              ))}
            </div>
          )}
        </div>

        <Button size="icon-sm" variant="ghost" onClick={onCreateChat}>
          <Plus className="size-3.5" />
          <span className="sr-only">New chat</span>
        </Button>
      </div>

      {/* Messages */}
      <div className="min-h-0 flex-1 overflow-y-auto overscroll-contain">
        <div className="flex flex-col gap-3 p-4">
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="size-5 animate-spin text-muted-foreground" />
            </div>
          ) : messages.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <MessageSquare className="mb-3 size-8 text-muted-foreground/40" />
              <p className="text-sm text-muted-foreground">
                Start a conversation to build something
              </p>
            </div>
          ) : (
            messages.map((msg) => <MessageBubble key={msg.id} message={msg} />)
          )}
          <div ref={bottomRef} />
        </div>
      </div>

      {/* Input */}
      <div className="border-t border-border p-3">
        <MessageInput
          value={inputValue}
          onChange={setInputValue}
          onSend={onSend}
          onCancel={onCancel}
          isGenerating={isGenerating}
          disabled={isLoading || !activeChat}
        />
      </div>
    </div>
  )
}
