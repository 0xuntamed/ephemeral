import type { ReactNode } from 'react'
import type { PreviewEnvironment } from '../types'
import {
  FAILURE_STATUSES,
  formatMemory,
  formatRemaining,
  initials,
  lifecycleProgress,
  shortSHA,
  statusLabel,
  type EnvironmentAction,
} from '../lib/environment'
import {
  ArrowRightIcon,
  ArrowUpRightIcon,
  BranchIcon,
  ClockIcon,
  RetryIcon,
  TrashIcon,
} from './Icons'
import { buttonClass, cn } from './design'
import { Button, Spinner, StatusBadge } from './ui'

interface EnvironmentCardProps {
  item: PreviewEnvironment
  busyKey: string | null
  onSelect: () => void
  onAction: (action: EnvironmentAction) => void
}

export function EnvironmentCard({ item, busyKey, onSelect, onAction }: EnvironmentCardProps) {
  const progress = lifecycleProgress(item.status)
  const canExtend = item.status === 'READY' && item.desired_state === 'READY'
  const canDestroy =
    item.status !== 'DESTROYED' && item.status !== 'DESTROYING' && item.status !== 'FAILED_CLEANED'
  const canRetry = FAILURE_STATUSES.has(item.status)
  const isFailure = FAILURE_STATUSES.has(item.status)

  return (
    <article className="content-auto bg-black/25 shadow-ticket">
      <div className="paper-edge paper-grain relative bg-paper pb-3 text-paper-ink">
        <div className={cn('h-2', isFailure ? 'bg-ember-600' : 'bg-amber')} />

        <button
          className="block w-full px-5 pt-5 text-left outline-none focus-visible:ring-2 focus-visible:ring-amber-deep focus-visible:ring-inset sm:px-6"
          type="button"
          onClick={onSelect}
          aria-label={`View ${item.repository} pull request ${item.pull_request}`}
        >
          <header className="grid grid-cols-[auto_1fr] items-start gap-3">
            <span className="grid size-11 place-items-center border border-paper-ink/25 bg-paper-raised font-display text-lg tracking-wider text-ember-700">
              {initials(item.repository)}
            </span>
            <span className="min-w-0">
              <span className="flex flex-wrap items-start justify-between gap-2">
                <span className="min-w-0">
                  <span className="block truncate text-sm font-medium text-paper-ink-dim">
                    {item.repository}
                  </span>
                  <span className="mt-0.5 block font-display text-2xl uppercase tracking-wide">
                    PR #{item.pull_request}
                  </span>
                </span>
                <StatusBadge status={item.status} />
              </span>
            </span>
          </header>

          <span className="mt-5 flex min-w-0 items-center gap-2 border-y border-dashed border-paper-ink/25 py-3 font-mono text-xs text-paper-ink-dim">
            <BranchIcon className="size-4 shrink-0" />
            <span className="truncate">{item.branch}</span>
            <code className="ml-auto border-l border-paper-ink/20 pl-2 text-paper-ink">
              {shortSHA(item.commit_sha)}
            </code>
          </span>

          <span
            className="mt-4 block"
            aria-label={`${statusLabel(item.status)}: ${progress}% complete`}
          >
            <span className="block h-1.5 overflow-hidden bg-paper-ink/10">
              <span
                className={cn('block h-full', isFailure ? 'bg-ember-600' : 'bg-amber-deep')}
                style={{ width: `${progress}%` }}
              />
            </span>
            <span className="mt-2 flex items-start justify-between gap-4 text-xs text-paper-ink-dim">
              <span>{item.status_message}</span>
              <span className="shrink-0 font-mono">{progress}%</span>
            </span>
          </span>

          <dl className="mt-5 grid grid-cols-2 gap-x-4 gap-y-3 border-t border-paper-ink/15 pt-4 text-sm">
            <TicketFact label="Compute">
              {item.resources.cpu / 1024} vCPU · {formatMemory(item.resources.memory)}
            </TicketFact>
            <TicketFact label="Expires">{formatRemaining(item.expires_at, item.status)}</TicketFact>
            <TicketFact label="Owner">@{item.owner}</TicketFact>
            <TicketFact label="Estimate">${item.estimated_hourly_cost.toFixed(3)}/hr</TicketFact>
          </dl>
        </button>

        <div className="mx-5 mt-5 flex items-center justify-between gap-3 border-t border-dashed border-paper-ink/30 pt-4 sm:mx-6">
          {item.preview_url && item.status === 'READY' ? (
            <a
              className={cn(
                buttonClass('danger', 'sm'),
                'border-ember-700 bg-ember-700 text-ember-50 hover:bg-ember-800',
              )}
              href={item.preview_url}
              target="_blank"
              rel="noreferrer"
            >
              Open preview <ArrowUpRightIcon className="size-4" />
            </a>
          ) : (
            <Button intent="paper" size="sm" onClick={onSelect}>
              Activity <ArrowRightIcon className="size-4" />
            </Button>
          )}

          <div className="flex items-center gap-1.5">
            {canRetry ? (
              <ActionButton
                label="Retry"
                busy={busyKey === `${item.id}:retry`}
                onClick={() => onAction('retry')}
              >
                <RetryIcon className="size-4" />
              </ActionButton>
            ) : null}
            {canExtend ? (
              <ActionButton
                label="Extend TTL"
                busy={busyKey === `${item.id}:extend`}
                onClick={() => onAction('extend')}
              >
                <ClockIcon className="size-4" />
              </ActionButton>
            ) : null}
            {canDestroy ? (
              <ActionButton
                label="Destroy"
                danger
                busy={busyKey === `${item.id}:destroy`}
                onClick={() => onAction('destroy')}
              >
                <TrashIcon className="size-4" />
              </ActionButton>
            ) : null}
          </div>
        </div>
      </div>
    </article>
  )
}

function TicketFact({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="min-w-0">
      <dt className="font-label text-[0.65rem] font-semibold uppercase tracking-[0.16em] text-paper-ink-dim">
        {label}
      </dt>
      <dd className="mt-0.5 truncate font-medium text-paper-ink">{children}</dd>
    </div>
  )
}

function ActionButton({
  label,
  danger = false,
  busy,
  onClick,
  children,
}: {
  label: string
  danger?: boolean
  busy: boolean
  onClick: () => void
  children: ReactNode
}) {
  return (
    <Button
      intent={danger ? 'danger' : 'paper'}
      size="icon"
      className="size-9"
      title={label}
      aria-label={label}
      disabled={busy}
      onClick={onClick}
    >
      {busy ? <Spinner /> : children}
    </Button>
  )
}
