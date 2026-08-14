import { useState } from 'react'
import { createEnvironment, destroyEnvironment, extendEnvironment, retryEnvironment } from './api'
import { CreateDialog } from './components/CreateDialog'
import { DetailPanel } from './components/DetailPanel'
import { EnvironmentCard } from './components/EnvironmentCard'
import {
  AlertIcon,
  ChevronIcon,
  MetricsIcon,
  PlusIcon,
  RefreshIcon,
  SearchIcon,
} from './components/Icons'
import { Button, CheckerStrip, Logo, StatTicket } from './components/ui'
import { buttonClass, cn } from './components/design'
import {
  ACTIVE_STATUSES,
  FAILURE_STATUSES,
  STATUS_OPTIONS,
  statusLabel,
  type EnvironmentAction,
} from './lib/environment'
import type { CreateEnvironmentInput, PreviewEnvironment, Status } from './types'
import { useEnvironments } from './useEnvironments'

export function App() {
  const { items, loading, refreshing, error, lastUpdated, refresh } = useEnvironments()
  const [query, setQuery] = useState('')
  const [status, setStatus] = useState<Status | 'ALL'>('ALL')
  const [createOpen, setCreateOpen] = useState(false)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [busyKey, setBusyKey] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)

  const normalizedQuery = query.trim().toLowerCase()
  const filtered = items.filter((item) => {
    const matchesStatus = status === 'ALL' || item.status === status
    const haystack =
      `${item.id} ${item.repository} ${item.branch} ${item.owner} ${item.pull_request}`.toLowerCase()
    return matchesStatus && (normalizedQuery === '' || haystack.includes(normalizedQuery))
  })
  const selected =
    selectedId === null ? null : (items.find((item) => item.id === selectedId) ?? null)
  const readyCount = items.filter((item) => item.status === 'READY').length
  const provisioningCount = items.filter((item) =>
    ['REQUESTED', 'PLANNING', 'PROVISIONING', 'HEALTH_CHECK'].includes(item.status),
  ).length
  const failureCount = items.filter((item) => FAILURE_STATUSES.has(item.status)).length
  const activeCost = items
    .filter((item) => ACTIVE_STATUSES.has(item.status))
    .reduce((total, item) => total + item.estimated_hourly_cost, 0)

  async function onCreated(input: CreateEnvironmentInput) {
    const created = await createEnvironment(input)
    setCreateOpen(false)
    setNotice(`${created.id} was accepted and queued for provisioning.`)
    setSelectedId(created.id)
    await refresh(true)
  }

  async function runAction(item: PreviewEnvironment, action: EnvironmentAction) {
    if (action === 'destroy') {
      const confirmed = window.confirm(
        `Destroy ${item.id}? Its preview workload and isolated data will be removed.`,
      )
      if (!confirmed) return
    }

    const key = `${item.id}:${action}`
    setBusyKey(key)
    setNotice(null)
    try {
      if (action === 'destroy') {
        await destroyEnvironment(item.id)
        setNotice(`Cleanup was scheduled for ${item.id}.`)
      } else if (action === 'extend') {
        await extendEnvironment(item.id)
        setNotice(`${item.id} was extended by one hour.`)
      } else {
        await retryEnvironment(item.id)
        setNotice(`A lifecycle retry was scheduled for ${item.id}.`)
      }
      await refresh(true)
    } catch (caught) {
      setNotice(caught instanceof Error ? caught.message : 'The operation failed.')
    } finally {
      setBusyKey(null)
    }
  }

  return (
    <div className="stage-grain min-h-screen overflow-x-hidden bg-stage text-stage-ink">
      <header className="border-b border-stage-line bg-stage/90 backdrop-blur">
        <div className="mx-auto flex min-h-16 max-w-7xl items-center justify-between gap-4 px-4 sm:px-6 lg:px-8">
          <a className="flex items-center gap-3" href="#top" aria-label="Ephemera home">
            <Logo className="text-ember-500" />
            <span>
              <span className="block font-display text-xl uppercase leading-none tracking-[0.08em]">
                Ephemera
              </span>
              <span className="block font-label text-[0.6rem] font-semibold uppercase tracking-[0.24em] text-stage-ink-dim">
                Control plane
              </span>
            </span>
          </a>

          <div className="flex items-center gap-2">
            <div
              className="hidden items-center gap-2 border-r border-stage-line pr-4 font-label text-[0.68rem] font-semibold uppercase tracking-[0.16em] text-stage-ink-dim sm:flex"
              title={lastUpdated ? `Updated ${lastUpdated.toLocaleTimeString()}` : 'Connecting'}
            >
              <span
                className={cn(
                  'size-2',
                  error ? 'bg-ember-500' : 'animate-pulse-warm bg-amber-bright',
                )}
              />
              {error ? 'API unavailable' : 'Reconciler online'}
            </div>
            <a
              className={buttonClass('ghost', 'icon')}
              href="/metrics"
              target="_blank"
              rel="noreferrer"
              aria-label="Open platform metrics"
            >
              <MetricsIcon className="size-5" />
            </a>
            <Button className="hidden sm:inline-flex" onClick={() => setCreateOpen(true)}>
              <PlusIcon className="size-4" />
              New environment
            </Button>
            <Button
              className="sm:hidden"
              size="icon"
              onClick={() => setCreateOpen(true)}
              aria-label="New environment"
            >
              <PlusIcon className="size-5" />
            </Button>
          </div>
        </div>
        <CheckerStrip className="h-2" />
      </header>

      <main id="top">
        <section className="mx-auto grid max-w-7xl gap-8 px-4 pb-10 pt-12 sm:px-6 sm:pb-14 sm:pt-16 lg:grid-cols-[minmax(0,1fr)_17rem] lg:items-end lg:px-8 lg:pt-20">
          <div>
            <h1 className="max-w-4xl font-display text-5xl uppercase leading-[0.92] tracking-[0.015em] text-stage-ink sm:text-7xl lg:text-[5.8rem]">
              Ship the branch.
              <br />
              <span className="text-amber-bright">Inspect the real thing.</span>
            </h1>
            <p className="mt-6 max-w-2xl text-base leading-7 text-stage-ink-dim sm:text-lg">
              Provision, inspect, extend, retry, and clean up every pull-request environment from
              one warm, honest control surface.
            </p>
          </div>

          <div className="border border-stage-line bg-stage-raised/80 p-5">
            <div className="flex items-center justify-between gap-4">
              <span className="font-label text-xs font-semibold uppercase tracking-[0.2em] text-amber-bright">
                Desired state
              </span>
              <span className="h-px flex-1 bg-stage-line" />
              <span className="font-mono text-[0.65rem] text-stage-ink-dim">LOOP / 05s</span>
            </div>
            <p className="mt-4 text-sm leading-6 text-stage-ink-dim">
              The reconciler keeps every deployment moving toward ready—or safely toward gone.
            </p>
          </div>
        </section>

        <section
          className="paper-grain mx-auto grid max-w-7xl grid-cols-2 bg-paper text-paper-ink shadow-ticket lg:grid-cols-4"
          aria-label="Environment summary"
        >
          <StatTicket
            label="Ready"
            value={readyCount.toString().padStart(2, '0')}
            detail="Healthy previews"
          />
          <StatTicket
            label="In progress"
            value={provisioningCount.toString().padStart(2, '0')}
            detail="Planning and deploying"
          />
          <StatTicket
            label="Attention"
            value={failureCount.toString().padStart(2, '0')}
            detail="Failed or orphaned"
            accent
          />
          <StatTicket
            label="Hourly estimate"
            value={`$${activeCost.toFixed(2)}`}
            detail="Active environments"
          />
        </section>

        <section className="mx-auto max-w-7xl px-4 py-12 sm:px-6 sm:py-16 lg:px-8">
          <div className="mb-7 flex flex-col gap-5 border-b border-stage-line pb-6 lg:flex-row lg:items-end lg:justify-between">
            <div>
              <h2 className="font-display text-3xl uppercase leading-none tracking-wide sm:text-4xl">
                Deployment tickets
              </h2>
              <p className="mt-2 text-sm text-stage-ink-dim">
                {filtered.length} visible · {items.length} total
              </p>
            </div>

            <div className="grid gap-2 sm:grid-cols-[minmax(15rem,1fr)_12rem_auto]">
              <label className="flex min-h-11 items-center border border-stage-line bg-stage-raised text-stage-ink-dim focus-within:border-amber">
                <SearchIcon className="ml-3 size-4 shrink-0" />
                <span className="sr-only">Search environments</span>
                <input
                  className="min-w-0 flex-1 bg-transparent px-3 py-2 text-sm text-stage-ink outline-none placeholder:text-stage-ink-dim/70"
                  value={query}
                  onChange={(event) => setQuery(event.target.value)}
                  placeholder="Search repo, branch, owner…"
                />
                {query !== '' ? (
                  <button
                    className="h-full px-3 font-label text-lg text-stage-ink-dim hover:text-stage-ink"
                    type="button"
                    onClick={() => setQuery('')}
                    aria-label="Clear search"
                  >
                    ×
                  </button>
                ) : null}
              </label>

              <label className="relative flex min-h-11 items-center border border-stage-line bg-stage-raised focus-within:border-amber">
                <span className="sr-only">Filter by status</span>
                <select
                  className="h-full w-full appearance-none bg-transparent py-2 pl-3 pr-9 font-label text-xs font-semibold uppercase tracking-[0.12em] text-stage-ink outline-none"
                  value={status}
                  onChange={(event) => setStatus(event.target.value as Status | 'ALL')}
                >
                  {STATUS_OPTIONS.map((option) => (
                    <option className="bg-stage text-stage-ink" key={option} value={option}>
                      {statusLabel(option)}
                    </option>
                  ))}
                </select>
                <ChevronIcon className="pointer-events-none absolute right-3 size-4 text-stage-ink-dim" />
              </label>

              <Button
                intent="outline"
                size="icon"
                onClick={() => void refresh()}
                aria-label="Refresh environments"
              >
                <RefreshIcon className={cn('size-4', refreshing && 'animate-spin')} />
              </Button>
            </div>
          </div>

          {error ? (
            <div
              className="mb-6 flex flex-col gap-4 border border-ember-500/60 bg-ember-950/60 p-4 text-ember-200 sm:flex-row sm:items-center"
              role="alert"
            >
              <AlertIcon className="size-5 shrink-0 text-ember-300" />
              <div className="min-w-0 flex-1">
                <strong className="block font-label text-xs uppercase tracking-[0.16em]">
                  Control plane unreachable
                </strong>
                <span className="mt-1 block break-words text-sm text-ember-200/80">{error}</span>
              </div>
              <Button intent="outline" size="sm" onClick={() => void refresh()}>
                Try again
              </Button>
            </div>
          ) : null}

          {notice ? (
            <div
              className="mb-6 flex items-center justify-between gap-4 border border-amber/50 bg-amber/10 px-4 py-3 text-sm text-amber-bright"
              role="status"
            >
              <span>{notice}</span>
              <button
                className="px-2 text-lg text-amber-bright/70 hover:text-amber-bright"
                type="button"
                aria-label="Dismiss message"
                onClick={() => setNotice(null)}
              >
                ×
              </button>
            </div>
          ) : null}

          {loading ? (
            <div
              className="grid gap-6 md:grid-cols-2 xl:grid-cols-3"
              aria-label="Loading environments"
            >
              {[0, 1, 2].map((index) => (
                <SkeletonCard key={index} />
              ))}
            </div>
          ) : filtered.length > 0 ? (
            <div className="grid gap-6 md:grid-cols-2 xl:grid-cols-3">
              {filtered.map((item) => (
                <EnvironmentCard
                  key={item.id}
                  item={item}
                  busyKey={busyKey}
                  onSelect={() => setSelectedId(item.id)}
                  onAction={(action) => void runAction(item, action)}
                />
              ))}
            </div>
          ) : (
            <EmptyState filtered={items.length > 0} onCreate={() => setCreateOpen(true)} />
          )}
        </section>
      </main>

      <footer className="border-t border-stage-line bg-stage-raised/50">
        <CheckerStrip className="h-2 opacity-80" />
        <div className="mx-auto flex max-w-7xl flex-col gap-2 px-4 py-6 font-label text-[0.68rem] font-semibold uppercase tracking-[0.18em] text-stage-ink-dim sm:flex-row sm:items-center sm:justify-between sm:px-6 lg:px-8">
          <span>Ephemera control plane</span>
          <span>Desired state, continuously reconciled</span>
        </div>
      </footer>

      {createOpen ? (
        <CreateDialog onClose={() => setCreateOpen(false)} onCreate={onCreated} />
      ) : null}
      {selected ? (
        <DetailPanel
          item={selected}
          busyKey={busyKey}
          onClose={() => setSelectedId(null)}
          onAction={(action) => void runAction(selected, action)}
        />
      ) : null}
    </div>
  )
}

function EmptyState({ filtered, onCreate }: { filtered: boolean; onCreate: () => void }) {
  return (
    <div className="border border-stage-line bg-stage-raised/70 px-5 py-14 text-center sm:py-20">
      <div className="mx-auto grid size-20 place-items-center border border-stage-line bg-stage">
        <Logo className="size-11 text-ember-500" />
      </div>
      <h3 className="mt-6 font-display text-3xl uppercase tracking-wide">
        {filtered ? 'No tickets match' : 'The queue is clear'}
      </h3>
      <p className="mx-auto mt-2 max-w-md text-sm leading-6 text-stage-ink-dim">
        {filtered
          ? 'Try a broader search or clear the status filter.'
          : 'Create a preview environment and its deployment ticket will appear here.'}
      </p>
      {!filtered ? (
        <Button className="mt-6" onClick={onCreate}>
          <PlusIcon className="size-4" /> Create the first one
        </Button>
      ) : null}
    </div>
  )
}

function SkeletonCard() {
  return (
    <div className="paper-edge min-h-96 animate-pulse bg-paper p-6">
      <div className="flex gap-3">
        <span className="size-11 bg-paper-ink/15" />
        <span className="flex-1">
          <span className="block h-3 w-2/3 bg-paper-ink/10" />
          <span className="mt-3 block h-7 w-1/3 bg-paper-ink/15" />
        </span>
      </div>
      <span className="mt-8 block h-10 bg-paper-ink/10" />
      <span className="mt-5 block h-2 bg-paper-ink/10" />
      <span className="mt-8 block h-24 bg-paper-ink/10" />
    </div>
  )
}
