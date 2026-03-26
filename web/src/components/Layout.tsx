import { useState, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { Group as PanelGroup, Panel, Separator as PanelResizeHandle } from 'react-resizable-panels'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Loader2, GitBranch, Monitor, ArrowLeft } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { Chat } from '@/components/Chat'
import { Preview } from '@/components/Preview'
import { CodeView } from '@/components/CodeView'
import { FileTree } from '@/components/FileTree'
import { SetupProgress } from '@/components/SetupProgress'
import { useProject } from '@/hooks/useProject'
import { useChat } from '@/hooks/useChat'
import { pushGitHub } from '@/lib/api'

function SandboxStatusBadge({ status }: { status?: string }) {
  if (!status) return null
  const variantMap: Record<string, 'default' | 'secondary' | 'destructive' | 'outline'> = {
    running: 'secondary',
    starting: 'outline',
    stopped: 'outline',
    error: 'destructive',
  }
  return (
    <Badge variant={variantMap[status] ?? 'outline'} className="text-xs">
      {status === 'starting' && <Loader2 className="mr-1 size-2.5 animate-spin" />}
      {status}
    </Badge>
  )
}

export function Layout() {
  const { id: projectId } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState<string>('preview')
  const [selectedFile, setSelectedFile] = useState<string | null>(null)
  const [isPushing, setIsPushing] = useState(false)

  const { project, sandbox, isLoading: projectLoading, refresh } = useProject(projectId ?? '')
  const {
    messages,
    chats,
    activeChat,
    isGenerating,
    sendMessage,
    cancelGeneration,
    switchChat,
    createChat,
    isLoading: chatLoading,
  } = useChat(projectId ?? '')

  const handleFileSelect = useCallback(
    (path: string) => {
      setSelectedFile(path)
      setActiveTab('code')
    },
    []
  )

  const handleGitHubPush = useCallback(async () => {
    if (!projectId) return
    setIsPushing(true)
    try {
      await pushGitHub(projectId)
    } catch {
      // ignore
    } finally {
      setIsPushing(false)
    }
  }, [projectId])

  if (projectLoading) {
    return (
      <div className="flex h-screen items-center justify-center bg-background">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  const setupStatus = project?.setupStatus
  const isSetupDone = !setupStatus || setupStatus === 'ready'

  return (
    <div className="flex h-screen flex-col bg-background">
      {/* Top bar */}
      <header className="flex h-11 shrink-0 items-center gap-3 border-b border-border px-4">
        <Button
          size="icon-sm"
          variant="ghost"
          onClick={() => navigate('/')}
          className="shrink-0"
        >
          <ArrowLeft className="size-3.5" />
          <span className="sr-only">Back to projects</span>
        </Button>

        <div className="flex items-center gap-1.5">
          <div className="flex size-5 items-center justify-center rounded bg-primary">
            <span className="text-[10px] font-bold text-primary-foreground">P</span>
          </div>
          <Monitor className="size-3.5 text-muted-foreground" />
          <span className="text-sm font-medium">{project?.name ?? 'Loading...'}</span>
        </div>

        <SandboxStatusBadge status={sandbox?.status} />

        <div className="ml-auto flex items-center gap-2">
          <Button
            size="sm"
            variant="outline"
            onClick={handleGitHubPush}
            disabled={isPushing || !isSetupDone}
          >
            {isPushing ? (
              <Loader2 className="mr-1.5 size-3.5 animate-spin" />
            ) : (
              <GitBranch className="mr-1.5 size-3.5" />
            )}
            Push
          </Button>
        </div>
      </header>

      {/* Main content */}
      <div className="flex-1 overflow-hidden">
        <PanelGroup orientation="horizontal" className="h-full">
          {/* Left panel - Chat */}
          <Panel defaultSize="35%" minSize={280} maxSize={600}>
            <div className="flex h-full flex-col border-r border-border">
              <Chat
                messages={messages}
                chats={chats}
                activeChat={activeChat}
                isGenerating={isGenerating}
                isLoading={chatLoading}
                onSend={sendMessage}
                onCancel={cancelGeneration}
                onSwitchChat={switchChat}
                onCreateChat={createChat}
              />
            </div>
          </Panel>

          <PanelResizeHandle className="w-1 cursor-col-resize bg-border transition-colors hover:bg-primary/30 active:bg-primary/50" />

          {/* Right panel - Preview/Code/Files */}
          <Panel defaultSize="65%" minSize={400}>
            <Tabs
              value={activeTab}
              onValueChange={setActiveTab}
              className="flex h-full flex-col"
            >
              <div className="flex items-center border-b border-border px-3 py-1.5">
                <TabsList>
                  <TabsTrigger value="preview">Preview</TabsTrigger>
                  <TabsTrigger value="code">Code</TabsTrigger>
                  <TabsTrigger value="files">Files</TabsTrigger>
                </TabsList>
              </div>

              <TabsContent value="preview" className="flex-1 overflow-hidden">
                <Preview
                  projectId={projectId ?? ''}
                  sandbox={sandbox}
                  onSandboxUpdate={refresh}
                />
              </TabsContent>

              <TabsContent value="code" className="flex-1 overflow-hidden">
                <CodeView
                  projectId={projectId ?? ''}
                  messages={messages}
                  selectedFile={selectedFile}
                  onFileSelect={setSelectedFile}
                />
              </TabsContent>

              <TabsContent value="files" className="flex-1 overflow-hidden">
                <FileTree
                  projectId={projectId ?? ''}
                  onFileSelect={handleFileSelect}
                  selectedFile={selectedFile}
                />
              </TabsContent>
            </Tabs>
          </Panel>
        </PanelGroup>
      </div>

      {/* Setup progress overlay */}
      {!isSetupDone && setupStatus && (
        <SetupProgress
          setupStatus={setupStatus}
          projectName={project?.name}
          onRetry={refresh}
        />
      )}
    </div>
  )
}
