import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  Plus,
  Loader2,
  FolderOpen,
  Trash2,
  AlertCircle,
} from 'lucide-react'
import type { Project } from '@/lib/api'
import { getProjects, createProject, deleteProject } from '@/lib/api'

function formatDate(dateStr: string): string {
  try {
    return new Intl.RelativeTimeFormat('en', { numeric: 'auto' }).format(
      Math.round((new Date(dateStr).getTime() - Date.now()) / (1000 * 60 * 60 * 24)),
      'day'
    )
  } catch {
    return new Date(dateStr).toLocaleDateString()
  }
}

function getStatusVariant(status?: string): 'default' | 'secondary' | 'destructive' | 'outline' {
  if (!status || status === 'ready') return 'secondary'
  if (status === 'error') return 'destructive'
  return 'outline'
}

export function ProjectList() {
  const navigate = useNavigate()
  const [projects, setProjects] = useState<Project[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showNewDialog, setShowNewDialog] = useState(false)
  const [newProjectName, setNewProjectName] = useState('')
  const [isCreating, setIsCreating] = useState(false)
  const [deletingId, setDeletingId] = useState<string | null>(null)

  const loadProjects = useCallback(async () => {
    try {
      const list = await getProjects()
      setProjects(list)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load projects')
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    loadProjects()
  }, [loadProjects])

  const handleCreateProject = useCallback(async () => {
    const name = newProjectName.trim()
    if (!name) return
    setIsCreating(true)
    try {
      const project = await createProject(name)
      setProjects((prev) => [project, ...prev])
      setShowNewDialog(false)
      setNewProjectName('')
      navigate(`/projects/${project.id}`)
    } catch {
      // ignore
    } finally {
      setIsCreating(false)
    }
  }, [newProjectName, navigate])

  const handleDeleteProject = useCallback(
    async (e: React.MouseEvent, projectId: string) => {
      e.stopPropagation()
      if (!confirm('Delete this project? This cannot be undone.')) return
      setDeletingId(projectId)
      try {
        await deleteProject(projectId)
        setProjects((prev) => prev.filter((p) => p.id !== projectId))
      } catch {
        // ignore
      } finally {
        setDeletingId(null)
      }
    },
    []
  )

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b border-border">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
          <div className="flex items-center gap-2.5">
            <div className="flex size-7 items-center justify-center rounded-lg bg-primary">
              <span className="text-xs font-bold text-primary-foreground">P</span>
            </div>
            <span className="text-sm font-semibold">Pinfra Studio</span>
          </div>
          <Button size="sm" onClick={() => setShowNewDialog(true)}>
            <Plus className="mr-1.5 size-3.5" />
            New Project
          </Button>
        </div>
      </header>

      {/* Content */}
      <main className="mx-auto max-w-6xl px-6 py-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold">Projects</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Build and manage your applications with Claude
          </p>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="size-6 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div className="flex flex-col items-center justify-center gap-3 py-16 text-muted-foreground">
            <AlertCircle className="size-6" />
            <p className="text-sm">{error}</p>
            <Button size="sm" variant="outline" onClick={loadProjects}>
              Retry
            </Button>
          </div>
        ) : projects.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-4 rounded-xl border border-dashed border-border py-24 text-center">
            <div className="rounded-full bg-muted p-4">
              <FolderOpen className="size-6 text-muted-foreground" />
            </div>
            <div>
              <p className="font-medium">Create your first project</p>
              <p className="mt-1 text-sm text-muted-foreground">
                Start building with Claude's help
              </p>
            </div>
            <Button size="sm" onClick={() => setShowNewDialog(true)}>
              <Plus className="mr-1.5 size-3.5" />
              New Project
            </Button>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {projects.map((project) => (
              <Card
                key={project.id}
                className="group cursor-pointer transition-all hover:ring-2 hover:ring-primary/30"
                onClick={() => navigate(`/projects/${project.id}`)}
              >
                <CardHeader>
                  <div className="flex items-start justify-between gap-2">
                    <CardTitle className="truncate">{project.name}</CardTitle>
                    <div className="flex items-center gap-1.5 opacity-0 transition-opacity group-hover:opacity-100">
                      <button
                        type="button"
                        onClick={(e) => handleDeleteProject(e, project.id)}
                        disabled={deletingId === project.id}
                        className="flex size-6 items-center justify-center rounded-md text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
                      >
                        {deletingId === project.id ? (
                          <Loader2 className="size-3 animate-spin" />
                        ) : (
                          <Trash2 className="size-3" />
                        )}
                      </button>
                    </div>
                  </div>
                  {project.template && (
                    <CardDescription>{project.template}</CardDescription>
                  )}
                </CardHeader>
                <CardContent>
                  <div className="flex items-center justify-between">
                    <Badge variant={getStatusVariant(project.setupStatus)}>
                      {project.setupStatus ?? 'ready'}
                    </Badge>
                    <span className="text-xs text-muted-foreground">
                      {formatDate(project.updatedAt)}
                    </span>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </main>

      {/* New Project Dialog */}
      <Dialog open={showNewDialog} onOpenChange={setShowNewDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create New Project</DialogTitle>
          </DialogHeader>
          <div className="py-2">
            <Input
              placeholder="Project name"
              value={newProjectName}
              onChange={(e) => setNewProjectName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') handleCreateProject()
              }}
              autoFocus
            />
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setShowNewDialog(false)
                setNewProjectName('')
              }}
            >
              Cancel
            </Button>
            <Button
              onClick={handleCreateProject}
              disabled={!newProjectName.trim() || isCreating}
            >
              {isCreating && <Loader2 className="mr-1.5 size-3.5 animate-spin" />}
              Create
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
