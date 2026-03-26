import { useState, useEffect, useCallback, useRef } from 'react'
import type { Chat, StreamChunk } from '@/lib/api'
import {
  getChats,
  getChat as apiGetChat,
  createChat as apiCreateChat,
  activateChat,
  sendMessage as apiSendMessage,
  cancelMessage as apiCancelMessage,
  getSSEUrl,
} from '@/lib/api'

export interface UIMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  isStreaming?: boolean
  toolUses?: ToolUse[]
}

export interface ToolUse {
  id: string
  toolName: string
  toolInput?: Record<string, unknown>
}

interface UseChatReturn {
  messages: UIMessage[]
  chats: Chat[]
  activeChat: Chat | null
  isGenerating: boolean
  sendMessage: (content: string) => Promise<void>
  cancelGeneration: () => Promise<void>
  switchChat: (chatId: string) => Promise<void>
  createChat: () => Promise<void>
  isLoading: boolean
}

export function useChat(projectId: string): UseChatReturn {
  const [chats, setChats] = useState<Chat[]>([])
  const [activeChat, setActiveChat] = useState<Chat | null>(null)
  const [messages, setMessages] = useState<UIMessage[]>([])
  const [isGenerating, setIsGenerating] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const esRef = useRef<EventSource | null>(null)
  const mountedRef = useRef(true)

  const closeSSE = useCallback(() => {
    if (esRef.current) {
      esRef.current.close()
      esRef.current = null
    }
  }, [])

  const startSSE = useCallback(
    (chatId: string) => {
      closeSSE()
      const url = getSSEUrl(projectId, chatId)
      const es = new EventSource(url)
      esRef.current = es

      // Streaming assistant message id
      const streamingId = `stream-${Date.now()}`

      // Add placeholder
      setMessages((prev) => [
        ...prev,
        { id: streamingId, role: 'assistant', content: '', isStreaming: true, toolUses: [] },
      ])

      es.onmessage = (event: MessageEvent) => {
        if (!mountedRef.current) return
        try {
          const chunk = JSON.parse(event.data as string) as StreamChunk

          if (chunk.type === 'text' && chunk.text) {
            setMessages((prev) =>
              prev.map((m) =>
                m.id === streamingId ? { ...m, content: m.content + chunk.text } : m
              )
            )
          } else if (chunk.type === 'tool_use') {
            setMessages((prev) =>
              prev.map((m) => {
                if (m.id !== streamingId) return m
                const toolUse: ToolUse = {
                  id: `tool-${Date.now()}`,
                  toolName: chunk.toolName ?? 'unknown',
                  toolInput: chunk.toolInput,
                }
                return { ...m, toolUses: [...(m.toolUses ?? []), toolUse] }
              })
            )
          } else if (chunk.type === 'done' || chunk.type === 'error') {
            setMessages((prev) =>
              prev.map((m) =>
                m.id === streamingId ? { ...m, isStreaming: false } : m
              )
            )
            setIsGenerating(false)
            closeSSE()
          }
        } catch {
          // ignore parse errors
        }
      }

      es.onerror = () => {
        if (!mountedRef.current) return
        setMessages((prev) =>
          prev.map((m) =>
            m.id === streamingId ? { ...m, isStreaming: false } : m
          )
        )
        setIsGenerating(false)
        closeSSE()
      }
    },
    [projectId, closeSSE]
  )

  const loadMessagesForChat = useCallback(
    async (chatId: string) => {
      try {
        const chat = await apiGetChat(projectId, chatId)
        if (!mountedRef.current) return
        // Convert API messages to UIMessages, merging tool messages into assistant
        const msgs: UIMessage[] = []
        for (const m of chat.messages ?? []) {
          if (m.role === 'user') {
            msgs.push({ id: m.id, role: 'user', content: m.content })
          } else if (m.role === 'assistant') {
            msgs.push({ id: m.id, role: 'assistant', content: m.content, toolUses: [] })
          } else if (m.role === 'tool') {
            // Attach tool use to the last assistant message
            const lastAssistant = [...msgs].reverse().find((msg) => msg.role === 'assistant')
            if (lastAssistant) {
              let toolInput: Record<string, unknown> | undefined
              try {
                if (m.toolInput) toolInput = JSON.parse(m.toolInput) as Record<string, unknown>
              } catch { /* ignore */ }
              lastAssistant.toolUses = [
                ...(lastAssistant.toolUses ?? []),
                { id: m.id, toolName: m.toolName ?? 'unknown', toolInput },
              ]
            }
          }
        }
        setMessages(msgs)
      } catch {
        setMessages([])
      }
    },
    [projectId]
  )

  const loadChats = useCallback(async () => {
    if (!projectId) return
    try {
      const list = await getChats(projectId)
      if (!mountedRef.current) return
      setChats(list)
      const active = list.find((c) => c.isActive) ?? list[0] ?? null
      setActiveChat(active)
      // Load messages for the active chat
      if (active) {
        await loadMessagesForChat(active.id)
      }
    } catch {
      // ignore
    } finally {
      if (mountedRef.current) setIsLoading(false)
    }
  }, [projectId, loadMessagesForChat])

  useEffect(() => {
    mountedRef.current = true
    loadChats()
    return () => {
      mountedRef.current = false
      closeSSE()
    }
  }, [loadChats, closeSSE])

  const sendMessage = useCallback(
    async (content: string) => {
      if (!activeChat || !content.trim() || isGenerating) return

      const userMsg: UIMessage = {
        id: `user-${Date.now()}`,
        role: 'user',
        content,
      }
      setMessages((prev) => [...prev, userMsg])
      setIsGenerating(true)

      try {
        await apiSendMessage(projectId, activeChat.id, content)
        startSSE(activeChat.id)
      } catch {
        setIsGenerating(false)
      }
    },
    [activeChat, isGenerating, projectId, startSSE]
  )

  const cancelGeneration = useCallback(async () => {
    if (!activeChat) return
    try {
      await apiCancelMessage(projectId, activeChat.id)
    } catch {
      // ignore
    }
    closeSSE()
    setIsGenerating(false)
    setMessages((prev) =>
      prev.map((m) => (m.isStreaming ? { ...m, isStreaming: false } : m))
    )
  }, [activeChat, projectId, closeSSE])

  const switchChat = useCallback(
    async (chatId: string) => {
      const chat = chats.find((c) => c.id === chatId)
      if (!chat) return
      closeSSE()
      setIsGenerating(false)
      setMessages([])
      await activateChat(projectId, chatId)
      setActiveChat(chat)
      setChats((prev) => prev.map((c) => ({ ...c, isActive: c.id === chatId })))
      // Load persisted messages for this chat
      await loadMessagesForChat(chatId)
    },
    [chats, projectId, closeSSE, loadMessagesForChat]
  )

  const createChat = useCallback(async () => {
    try {
      const chat = await apiCreateChat(projectId)
      setChats((prev) => [...prev, chat])
      await switchChat(chat.id)
    } catch {
      // ignore
    }
  }, [projectId, switchChat])

  return {
    messages,
    chats,
    activeChat,
    isGenerating,
    sendMessage,
    cancelGeneration,
    switchChat,
    createChat,
    isLoading,
  }
}
