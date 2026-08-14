import { useEffect, useState, type FormEvent, type ReactNode } from 'react'
import type { CreateEnvironmentInput } from '../types'
import { CloseIcon, PlusIcon, ShieldIcon } from './Icons'
import { fieldClass } from './design'
import { Button, CheckerStrip, FieldLabel, Spinner } from './ui'

interface CreateDialogProps {
  onClose: () => void
  onCreate: (input: CreateEnvironmentInput) => Promise<void>
}

export function CreateDialog({ onClose, onCreate }: CreateDialogProps) {
  const [submitting, setSubmitting] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  useEffect(() => {
    function closeOnEscape(event: KeyboardEvent) {
      if (event.key === 'Escape') onClose()
    }

    document.addEventListener('keydown', closeOnEscape)
    return () => document.removeEventListener('keydown', closeOnEscape)
  }, [onClose])

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const data = new FormData(event.currentTarget)
    setSubmitting(true)
    setFormError(null)
    try {
      await onCreate({
        repository: String(data.get('repository')),
        branch: String(data.get('branch')),
        commit_sha: String(data.get('commit_sha')),
        pull_request: Number(data.get('pull_request')),
        owner: String(data.get('owner')),
        image: String(data.get('image')),
        ttl: String(data.get('ttl')),
        resources: {
          cpu: Number(data.get('cpu')),
          memory: Number(data.get('memory')),
          database_mode: data.get('database_mode') as 'shared-schema' | 'none',
        },
      })
    } catch (caught) {
      setFormError(caught instanceof Error ? caught.message : 'Could not create the environment.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 grid place-items-center overflow-y-auto bg-stage/90 p-3 backdrop-blur-sm sm:p-6"
      onMouseDown={(event) => {
        if (event.currentTarget === event.target) onClose()
      }}
    >
      <section
        className="paper-edge paper-grain my-auto flex max-h-[calc(100dvh-1.5rem)] w-full max-w-3xl flex-col overflow-hidden bg-paper pb-3 text-paper-ink shadow-ticket sm:max-h-[calc(100dvh-3rem)]"
        role="dialog"
        aria-modal="true"
        aria-labelledby="create-title"
      >
        <CheckerStrip className="shrink-0" />
        <header className="flex shrink-0 items-start justify-between gap-6 border-b border-paper-ink/20 px-5 py-5 sm:px-8 sm:py-6">
          <div>
            <h2
              id="create-title"
              className="font-display text-3xl uppercase leading-none tracking-wide sm:text-4xl"
            >
              New preview environment
            </h2>
            <p className="mt-2 max-w-xl text-sm leading-6 text-paper-ink-dim">
              Issue a short-lived deployment ticket. The control plane will reconcile it into
              compute, routing, and isolated data.
            </p>
          </div>
          <Button intent="paper" size="icon" onClick={onClose} aria-label="Close create form">
            <CloseIcon className="size-5" />
          </Button>
        </header>

        <form className="flex min-h-0 flex-1 flex-col" onSubmit={(event) => void submit(event)}>
          <div className="min-h-0 overflow-y-auto">
            <div className="grid gap-x-4 gap-y-5 px-5 py-6 sm:grid-cols-2 sm:px-8">
              <FormField id="repository" label="Repository" wide>
                <input
                  className={fieldClass}
                  id="repository"
                  name="repository"
                  required
                  autoFocus
                  placeholder="acme/billing-api"
                  pattern="[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+"
                />
              </FormField>
              <FormField id="pull_request" label="Pull request">
                <input
                  className={fieldClass}
                  id="pull_request"
                  name="pull_request"
                  type="number"
                  min="1"
                  required
                  placeholder="127"
                />
              </FormField>
              <FormField id="owner" label="Owner">
                <input
                  className={fieldClass}
                  id="owner"
                  name="owner"
                  required
                  placeholder="bilal"
                />
              </FormField>
              <FormField id="branch" label="Branch" wide>
                <input
                  className={fieldClass}
                  id="branch"
                  name="branch"
                  required
                  placeholder="feature/new-checkout"
                />
              </FormField>
              <FormField id="commit_sha" label="Commit SHA" wide>
                <input
                  className={fieldClass}
                  id="commit_sha"
                  name="commit_sha"
                  required
                  minLength={7}
                  maxLength={64}
                  pattern="[A-Fa-f0-9]{7,64}"
                  placeholder="a81f2aa9b32c"
                />
              </FormField>
              <FormField id="image" label="Container image (optional)" wide>
                <input
                  className={fieldClass}
                  id="image"
                  name="image"
                  placeholder="123456789.dkr.ecr.us-east-1.amazonaws.com/app:sha"
                />
              </FormField>
            </div>

            <div className="mx-5 border-y border-dashed border-paper-ink/30 py-3 font-label text-xs font-semibold uppercase tracking-[0.2em] text-ember-700 sm:mx-8">
              Resource policy
            </div>

            <div className="grid gap-x-4 gap-y-5 px-5 py-6 sm:grid-cols-2 sm:px-8 lg:grid-cols-4">
              <FormField id="ttl" label="TTL">
                <select className={fieldClass} id="ttl" name="ttl" defaultValue="6h">
                  <option value="1h">1 hour</option>
                  <option value="3h">3 hours</option>
                  <option value="6h">6 hours</option>
                  <option value="12h">12 hours</option>
                  <option value="24h">24 hours</option>
                </select>
              </FormField>
              <FormField id="cpu" label="CPU">
                <select className={fieldClass} id="cpu" name="cpu" defaultValue="512">
                  <option value="256">0.25 vCPU</option>
                  <option value="512">0.5 vCPU</option>
                  <option value="1024">1 vCPU</option>
                  <option value="2048">2 vCPU</option>
                </select>
              </FormField>
              <FormField id="memory" label="Memory">
                <select className={fieldClass} id="memory" name="memory" defaultValue="1024">
                  <option value="512">512 MiB</option>
                  <option value="1024">1 GiB</option>
                  <option value="2048">2 GiB</option>
                  <option value="4096">4 GiB</option>
                </select>
              </FormField>
              <FormField id="database_mode" label="Database">
                <select
                  className={fieldClass}
                  id="database_mode"
                  name="database_mode"
                  defaultValue="shared-schema"
                >
                  <option value="shared-schema">Isolated schema</option>
                  <option value="none">No database</option>
                </select>
              </FormField>
            </div>

            {formError ? (
              <p
                className="mx-5 mb-6 border border-ember-700/40 bg-ember-100 px-4 py-3 text-sm text-ember-800 sm:mx-8"
                role="alert"
              >
                {formError}
              </p>
            ) : null}
          </div>

          <footer className="flex shrink-0 flex-col gap-4 border-t border-paper-ink/20 bg-paper-raised px-5 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-8">
            <p className="flex items-center gap-2 text-xs text-paper-ink-dim">
              <ShieldIcon className="size-4 shrink-0 text-amber-deep" />
              Policy limits are enforced by the control plane.
            </p>
            <div className="flex items-center justify-end gap-2">
              <Button intent="paper" onClick={onClose}>
                Cancel
              </Button>
              <Button type="submit" disabled={submitting}>
                {submitting ? <Spinner /> : <PlusIcon className="size-4" />}
                {submitting ? 'Queuing…' : 'Create environment'}
              </Button>
            </div>
          </footer>
        </form>
      </section>
    </div>
  )
}

function FormField({
  id,
  label,
  wide = false,
  children,
}: {
  id: string
  label: string
  wide?: boolean
  children: ReactNode
}) {
  return (
    <div className={wide ? 'sm:col-span-2' : undefined}>
      <FieldLabel htmlFor={id}>{label}</FieldLabel>
      {children}
    </div>
  )
}
