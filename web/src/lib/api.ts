// API client library for pinfra-studio

const BASE = '/api/v1'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...options?.headers },
    ...options,
  })
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText)
    throw new Error(`API error ${res.status}: ${text}`)
  }
  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}

// ─── Types ────────────────────────────────────────────────────────────────────

export interface Project {
  id: string
  name: string
  template?: string
  setupStatus?: string
  createdAt: string
  updatedAt: string
}

export interface Chat {
  id: string
  projectId: string
  title?: string
  isActive: boolean
  createdAt: string
  updatedAt: string
  messages?: Message[]
}

export interface Message {
  id: string
  chatId: string
  role: 'user' | 'assistant' | 'tool'
  content: string
  toolName?: string
  toolInput?: string
  createdAt: string
}

export interface SandboxStatus {
  status: 'stopped' | 'starting' | 'running' | 'error'
  port?: number
  url?: string
  error?: string
}

export interface FileEntry {
  name: string
  path: string
  isDir: boolean
  children?: FileEntry[]
  size?: number
}

export interface StreamChunk {
  type: 'text' | 'tool_use' | 'tool_result' | 'done' | 'error'
  text?: string
  toolName?: string
  toolInput?: Record<string, unknown>
  toolResult?: string
  error?: string
}

// ─── Projects ─────────────────────────────────────────────────────────────────

export const getProjects = () => request<Project[]>('/projects')
export const createProject = (name: string) =>
  request<Project>('/projects', { method: 'POST', body: JSON.stringify({ name }) })
export const getProject = (id: string) => request<Project>(`/projects/${id}`)
export const deleteProject = (id: string) =>
  request<void>(`/projects/${id}`, { method: 'DELETE' })

// ─── Chats ────────────────────────────────────────────────────────────────────

export const getChats = (projectId: string) =>
  request<Chat[]>(`/projects/${projectId}/chats`)
export const createChat = (projectId: string) =>
  request<Chat>(`/projects/${projectId}/chats`, { method: 'POST' })
export const getChat = (projectId: string, chatId: string) =>
  request<Chat>(`/projects/${projectId}/chats/${chatId}`)
export const activateChat = (projectId: string, chatId: string) =>
  request<void>(`/projects/${projectId}/chats/${chatId}/activate`, { method: 'POST' })
export const deleteChat = (projectId: string, chatId: string) =>
  request<void>(`/projects/${projectId}/chats/${chatId}`, { method: 'DELETE' })

// ─── Messages ─────────────────────────────────────────────────────────────────

export const sendMessage = (projectId: string, chatId: string, content: string) =>
  request<Message>(`/projects/${projectId}/chats/${chatId}/messages`, {
    method: 'POST',
    body: JSON.stringify({ content }),
  })
export const cancelMessage = (projectId: string, chatId: string) =>
  request<void>(`/projects/${projectId}/chats/${chatId}/messages/cancel`, { method: 'POST' })

export const getSSEUrl = (projectId: string, chatId: string) =>
  `${BASE}/projects/${projectId}/chats/${chatId}/events`

// ─── Sandbox ──────────────────────────────────────────────────────────────────

export const getSandboxStatus = (projectId: string) =>
  request<SandboxStatus>(`/projects/${projectId}/sandbox/status`)
export const startSandbox = (projectId: string) =>
  request<SandboxStatus>(`/projects/${projectId}/sandbox/start`, { method: 'POST' })
export const stopSandbox = (projectId: string) =>
  request<void>(`/projects/${projectId}/sandbox/stop`, { method: 'POST' })
export const getSandboxUrl = (projectId: string) =>
  request<{ url: string }>(`/projects/${projectId}/sandbox/url`)

// ─── Files ────────────────────────────────────────────────────────────────────

export const getFiles = (projectId: string) =>
  request<FileEntry[]>(`/projects/${projectId}/files`)
export const getFileContent = (projectId: string, filePath: string) =>
  request<{ content: string }>(`/projects/${projectId}/files/${filePath}`)

// ─── GitHub ───────────────────────────────────────────────────────────────────

export const connectGitHub = (projectId: string, repoUrl: string) =>
  request<void>(`/projects/${projectId}/github/connect`, {
    method: 'POST',
    body: JSON.stringify({ repo_url: repoUrl }),
  })
export const pushGitHub = (projectId: string) =>
  request<void>(`/projects/${projectId}/github/push`, { method: 'POST' })
export const createPR = (projectId: string, title: string, body: string) =>
  request<{ url: string }>(`/projects/${projectId}/github/pr`, {
    method: 'POST',
    body: JSON.stringify({ title, body }),
  })
