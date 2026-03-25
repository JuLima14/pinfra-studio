import { useEffect, useRef, useState } from 'react'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { MessageInput } from '@/components/MessageInput'
import {
  MessageSquare,
  Plus,
  ChevronDown,
  Wrench,
  Loader2,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import type { UIMessage, ToolUse } from '@/hooks/useChat'
import type { Chat as ChatType } from '@/lib/api'

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

function ToolUseCard({ toolUse }: { toolUse: ToolUse }) {
  const [expanded, setExpanded] = useState(false)
  const filePath =
    (toolUse.toolInput?.path as string) ??
    (toolUse.toolInput?.file_path as string) ??
    ''

  return (
    <div className="mt-2 rounded-lg border border-border bg-muted/30 text-xs">
      <button
        type="button"
        onClick={() => setExpanded((v) => !v)}
        className="flex w-full items-center gap-2 px-3 py-2 text-left text-muted-foreground hover:text-foreground"
      >
        <Wrench className="size-3 shrink-0" />
        <span className="font-medium">{toolUse.toolName}</span>
        {filePath && (
          <span className="truncate text-muted-foreground/70">{filePath}</span>
        )}
        <ChevronDown
          className={cn(
            'ml-auto size-3 shrink-0 transition-transform',
            expanded && 'rotate-180'
          )}
        />
      </button>
      {expanded && toolUse.toolInput && (
        <div className="border-t border-border px-3 py-2">
          <pre className="whitespace-pre-wrap break-all font-mono text-[11px] text-muted-foreground">
            {JSON.stringify(toolUse.toolInput, null, 2)}
          </pre>
        </div>
      )}
    </div>
  )
}

function MessageBubble({ message }: { message: UIMessage }) {
  const isUser = message.role === 'user'

  return (
    <div className={cn('flex', isUser ? 'justify-end' : 'justify-start')}>
      <div
        className={cn(
          'max-w-[85%] rounded-2xl px-4 py-2.5 text-sm',
          isUser
            ? 'bg-primary text-primary-foreground'
            : 'bg-muted text-foreground'
        )}
      >
        <p className="whitespace-pre-wrap leading-relaxed">
          {message.content}
          {message.isStreaming && (
            <span className="ml-0.5 inline-block h-3.5 w-0.5 animate-pulse bg-current align-middle" />
          )}
        </p>
        {message.toolUses && message.toolUses.length > 0 && (
          <div>
            {message.toolUses.map((tu) => (
              <ToolUseCard key={tu.id} toolUse={tu} />
            ))}
          </div>
        )}
      </div>
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
      <ScrollArea className="flex-1">
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
      </ScrollArea>

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
