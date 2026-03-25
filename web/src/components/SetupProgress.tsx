import { Button } from '@/components/ui/button'
import { Check, Loader2, AlertCircle } from 'lucide-react'
import { cn } from '@/lib/utils'

interface SetupStep {
  id: string
  label: string
  description?: string
}

const STEPS: SetupStep[] = [
  { id: 'scaffolding', label: 'Scaffolding', description: 'Creating project structure' },
  { id: 'installing', label: 'Installing', description: 'Installing dependencies' },
  { id: 'starting', label: 'Starting', description: 'Starting development server' },
  { id: 'ready', label: 'Ready', description: 'Project is ready' },
]

type StepStatus = 'pending' | 'active' | 'done' | 'error'

function getStepStatus(setupStatus: string, stepId: string): StepStatus {
  const order = STEPS.map((s) => s.id)
  const currentIndex = order.indexOf(setupStatus)
  const stepIndex = order.indexOf(stepId)

  if (setupStatus === 'error' && stepIndex === currentIndex) return 'error'
  if (stepIndex < currentIndex) return 'done'
  if (stepIndex === currentIndex) return 'active'
  return 'pending'
}

interface SetupProgressProps {
  setupStatus: string
  projectName?: string
  onRetry?: () => void
}

export function SetupProgress({ setupStatus, projectName, onRetry }: SetupProgressProps) {
  const isError = setupStatus === 'error'

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/95 backdrop-blur-sm">
      <div className="w-full max-w-sm px-6">
        <div className="mb-8 text-center">
          <h2 className="text-lg font-semibold">
            {isError ? 'Setup Failed' : 'Setting up your project'}
          </h2>
          {projectName && (
            <p className="mt-1 text-sm text-muted-foreground">{projectName}</p>
          )}
        </div>

        <div className="space-y-4">
          {STEPS.filter((s) => s.id !== 'ready' || setupStatus === 'ready').map((step) => {
            const status = getStepStatus(setupStatus, step.id)

            return (
              <div key={step.id} className="flex items-center gap-4">
                {/* Step indicator */}
                <div
                  className={cn(
                    'flex size-8 shrink-0 items-center justify-center rounded-full border-2 transition-all',
                    status === 'done' && 'border-primary bg-primary text-primary-foreground',
                    status === 'active' && 'border-primary bg-transparent text-primary',
                    status === 'pending' && 'border-border bg-transparent text-muted-foreground',
                    status === 'error' && 'border-destructive bg-destructive/10 text-destructive'
                  )}
                >
                  {status === 'done' && <Check className="size-4" />}
                  {status === 'active' && <Loader2 className="size-4 animate-spin" />}
                  {status === 'error' && <AlertCircle className="size-4" />}
                  {status === 'pending' && (
                    <div className="size-2 rounded-full bg-current opacity-30" />
                  )}
                </div>

                {/* Step info */}
                <div className="flex-1">
                  <p
                    className={cn(
                      'text-sm font-medium transition-colors',
                      status === 'active' && 'text-foreground',
                      status === 'done' && 'text-foreground',
                      status === 'pending' && 'text-muted-foreground',
                      status === 'error' && 'text-destructive'
                    )}
                  >
                    {step.label}
                  </p>
                  {(status === 'active' || status === 'error') && step.description && (
                    <p className="mt-0.5 text-xs text-muted-foreground">{step.description}</p>
                  )}
                </div>

                {/* Connector line */}
              </div>
            )
          })}
        </div>

        {isError && onRetry && (
          <div className="mt-8 text-center">
            <Button onClick={onRetry} size="sm">
              Retry Setup
            </Button>
          </div>
        )}

        {!isError && (
          <div className="mt-8">
            <div className="h-1 w-full overflow-hidden rounded-full bg-muted">
              <div
                className="h-full rounded-full bg-primary transition-all duration-500"
                style={{
                  width: `${
                    setupStatus === 'scaffolding'
                      ? 25
                      : setupStatus === 'installing'
                      ? 50
                      : setupStatus === 'starting'
                      ? 75
                      : 100
                  }%`,
                }}
              />
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
