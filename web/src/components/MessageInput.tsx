import { useRef, useCallback, type KeyboardEvent } from 'react'
import { Button } from '@/components/ui/button'
import { ArrowUp, Square } from 'lucide-react'
import { cn } from '@/lib/utils'

interface MessageInputProps {
  onSend: (content: string) => void
  onCancel: () => void
  isGenerating: boolean
  disabled?: boolean
  value: string
  onChange: (value: string) => void
}

export function MessageInput({
  onSend,
  onCancel,
  isGenerating,
  disabled,
  value,
  onChange,
}: MessageInputProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const handleSend = useCallback(() => {
    const trimmed = value.trim()
    if (!trimmed || isGenerating || disabled) return
    onSend(trimmed)
    onChange('')
    // Reset height
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
    }
  }, [value, isGenerating, disabled, onSend, onChange])

  const handleKeyDown = useCallback(
    (e: KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault()
        handleSend()
      }
    },
    [handleSend]
  )

  const handleInput = useCallback(
    (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      onChange(e.target.value)
      // Auto-resize
      const el = e.target
      el.style.height = 'auto'
      const lineHeight = 24
      const maxHeight = lineHeight * 6
      el.style.height = `${Math.min(el.scrollHeight, maxHeight)}px`
    },
    [onChange]
  )

  return (
    <div className="flex items-end gap-2 rounded-xl border border-border bg-card p-3">
      <textarea
        ref={textareaRef}
        value={value}
        onChange={handleInput}
        onKeyDown={handleKeyDown}
        placeholder="Ask Claude to build something..."
        disabled={disabled}
        rows={1}
        className={cn(
          'min-h-6 flex-1 resize-none bg-transparent text-sm text-foreground placeholder:text-muted-foreground',
          'outline-none focus:outline-none',
          'disabled:cursor-not-allowed disabled:opacity-50'
        )}
        style={{ lineHeight: '24px' }}
      />
      {isGenerating ? (
        <Button
          size="icon-sm"
          variant="secondary"
          onClick={onCancel}
          className="shrink-0"
        >
          <Square className="size-3.5" />
          <span className="sr-only">Stop generation</span>
        </Button>
      ) : (
        <Button
          size="icon-sm"
          onClick={handleSend}
          disabled={!value.trim() || disabled}
          className="shrink-0"
        >
          <ArrowUp className="size-3.5" />
          <span className="sr-only">Send message</span>
        </Button>
      )}
    </div>
  )
}
