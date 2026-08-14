import { useEffect, type ReactNode } from 'react'
import type { PreviewEnvironment } from '../types'
import {
  FAILURE_STATUSES,
  formatDate,
  formatRemaining,
  statusLabel,
  type EnvironmentAction,
} from '../lib/environment'
import { AlertIcon, ArrowUpRightIcon, ClockIcon, CloseIcon, RetryIcon, TrashIcon } from './Icons'
import { buttonClass, cn } from './design'
import { Button, CheckerStrip, Spinner, StatusBadge } from './ui'

interface DetailPanelProps {
  item: PreviewEnvironment
  busyKey: string | null
  onClose: () => void
  onAction: (action: EnvironmentAction) => void
}

export function DetailPanel({ item, busyKey, onClose, onAction }: DetailPanelProps) {
  const events = [...(item.events ?? [])].reverse()
  const cloudResources = item.cloud_resources ?? []
  const canRetry = FAILURE_STATUSES.has(item.status)
  const canDestroy = !['DESTROYED', 'DESTROYING', 'FAILED_CLEANED'].includes(item.status)
  const canExtend = item.status === 'READY' && item.desired_state === 'READY'

  useEffect(() => {
    function closeOnEscape(event: KeyboardEvent) {
      if (event.key === 'Escape') onClose()
    }

    document.addEventListener('keydown', closeOnEscape)
    return () => document.removeEventListener('keydown', closeOnEscape)
  }, [onClose])

  return (
    <div
      className="fixed inset-0 z-40 flex justify-end bg-stage/80 backdrop-blur-sm"
      onMouseDown={(event) => {
        if (event.currentTarget === event.target) onClose()
      }}
    >
      <aside
        className="paper-grain h-full w-full max-w-2xl overflow-y-auto bg-paper text-paper-ink shadow-[-20px_0_60px_rgb(0_0_0/35%)]"
        role="dialog"
        aria-modal="true"
        aria-label={`${item.id} details`}
      >
        <CheckerStrip />
        <header className="sticky top-0 z-10 flex items-start justify-between gap-5 border-b border-paper-ink/20 bg-paper/95 px-5 py-5 backdrop-blur sm:px-8 sm:py-7">
          <div className="min-w-0">
            <StatusBadge status={item.status} />
            <h2 className="mt-4 truncate font-display text-3xl uppercase leading-none tracking-wide sm:text-4xl">
              {item.id}
            </h2>
            <p className="mt-2 text-sm text-paper-ink-dim">
              {item.repository} · PR #{item.pull_request}
            </p>
          </div>
          <Button intent="paper" size="icon" onClick={onClose} aria-label="Close details">
            <CloseIcon className="size-5" />
          </Button>
        </header>

        <div className="p-5 sm:p-8">
          <div className="flex flex-wrap gap-2 border-b border-dashed border-paper-ink/30 pb-6">
            {item.preview_url && item.status === 'READY' ? (
              <a
                className={cn(
                  buttonClass('danger'),
                  'border-ember-700 bg-ember-700 text-ember-50 hover:bg-ember-800',
                )}
                href={item.preview_url}
                target="_blank"
                rel="noreferrer"
              >
                Open preview <ArrowUpRightIcon className="size-4" />
              </a>
            ) : null}
            {canExtend ? (
              <Button
                intent="paper"
                disabled={busyKey === `${item.id}:extend`}
                onClick={() => onAction('extend')}
              >
                {busyKey === `${item.id}:extend` ? <Spinner /> : <ClockIcon className="size-4" />}
                Extend 1h
              </Button>
            ) : null}
            {canRetry ? (
              <Button
                intent="paper"
                disabled={busyKey === `${item.id}:retry`}
                onClick={() => onAction('retry')}
              >
                {busyKey === `${item.id}:retry` ? <Spinner /> : <RetryIcon className="size-4" />}
                Retry
              </Button>
            ) : null}
            {canDestroy ? (
              <Button
                intent="danger"
                disabled={busyKey === `${item.id}:destroy`}
                onClick={() => onAction('destroy')}
              >
                {busyKey === `${item.id}:destroy` ? <Spinner /> : <TrashIcon className="size-4" />}
                Destroy
              </Button>
            ) : null}
          </div>

          {item.last_error ? (
            <div className="mt-6 flex gap-3 border border-ember-700/35 bg-ember-100 p-4 text-ember-800">
              <AlertIcon className="mt-0.5 size-5 shrink-0" />
              <div className="min-w-0">
                <strong className="font-label text-xs uppercase tracking-[0.16em]">
                  Last lifecycle error
                </strong>
                <code className="mt-2 block whitespace-pre-wrap break-words text-xs">
                  {item.last_error}
                </code>
              </div>
            </div>
          ) : null}

          <DetailSection title="Environment">
            <dl className="divide-y divide-paper-ink/15 border-y border-paper-ink/15">
              <DetailRow label="Branch">{item.branch}</DetailRow>
              <DetailRow label="Commit">
                <code className="break-all text-xs">{item.commit_sha}</code>
              </DetailRow>
              <DetailRow label="Image">
                <code className="break-all text-xs">{item.image || 'default image'}</code>
              </DetailRow>
              <DetailRow label="Expires">
                {formatDate(item.expires_at)} · {formatRemaining(item.expires_at, item.status)}
              </DetailRow>
              <DetailRow label="Terraform state">
                <code className="break-all text-xs">{item.terraform_workspace}</code>
              </DetailRow>
              <DetailRow label="Database schema">
                <code className="break-all text-xs">{item.database_schema || 'none'}</code>
              </DetailRow>
            </dl>
          </DetailSection>

          <DetailSection title="Lifecycle activity" count={events.length}>
            {events.length > 0 ? (
              <ol className="relative border-l border-paper-ink/25">
                {events.map((event, index) => (
                  <li className="relative ml-5 pb-6 last:pb-0" key={`${event.at}-${index}`}>
                    <span
                      className={cn(
                        'absolute -left-[1.42rem] top-1 size-2.5 border-2 border-paper',
                        FAILURE_STATUSES.has(event.status) ? 'bg-ember-700' : 'bg-amber-deep',
                      )}
                    />
                    <strong className="font-label text-xs uppercase tracking-[0.15em]">
                      {statusLabel(event.status)}
                    </strong>
                    <p className="mt-1 text-sm leading-6 text-paper-ink-dim">{event.message}</p>
                    <time className="mt-1 block font-mono text-[0.68rem] text-paper-ink-dim">
                      {formatDate(event.at)}
                    </time>
                  </li>
                ))}
              </ol>
            ) : (
              <p className="text-sm text-paper-ink-dim">No lifecycle events recorded yet.</p>
            )}
          </DetailSection>

          <DetailSection title="Cloud resources" count={cloudResources.length}>
            {cloudResources.length > 0 ? (
              <ul className="divide-y divide-paper-ink/15 border-y border-paper-ink/15">
                {cloudResources.map((resource) => (
                  <li
                    className="grid gap-1 py-3 sm:grid-cols-[8rem_1fr] sm:gap-4"
                    key={`${resource.kind}:${resource.id}`}
                  >
                    <span className="font-label text-xs uppercase tracking-[0.14em] text-paper-ink-dim">
                      {resource.kind}
                    </span>
                    <code className="break-all text-xs">{resource.id}</code>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-sm text-paper-ink-dim">
                Resources will appear after provisioning.
              </p>
            )}
          </DetailSection>
        </div>
      </aside>
    </div>
  )
}

function DetailSection({
  title,
  count,
  children,
}: {
  title: string
  count?: number
  children: ReactNode
}) {
  return (
    <section className="mt-8">
      <h3 className="mb-4 flex items-center gap-2 font-display text-xl uppercase tracking-wide">
        {title}
        {count !== undefined ? (
          <span className="border border-paper-ink/25 px-1.5 py-0.5 font-label text-[0.65rem] font-semibold text-paper-ink-dim">
            {count}
          </span>
        ) : null}
      </h3>
      {children}
    </section>
  )
}

function DetailRow({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="grid gap-1 py-3 sm:grid-cols-[8rem_1fr] sm:gap-4">
      <dt className="font-label text-xs font-semibold uppercase tracking-[0.14em] text-paper-ink-dim">
        {label}
      </dt>
      <dd className="min-w-0 text-sm">{children}</dd>
    </div>
  )
}
