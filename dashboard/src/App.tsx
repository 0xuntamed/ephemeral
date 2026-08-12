import { useState, type FormEvent, type ReactNode } from 'react'
import { createEnvironment, destroyEnvironment, extendEnvironment, retryEnvironment } from './api'
import type { CreateEnvironmentInput, PreviewEnvironment, Status } from './types'
import { useEnvironments } from './useEnvironments'

const STATUS_OPTIONS: Array<Status | 'ALL'> = [
  'ALL',
  'READY',
  'PROVISIONING',
  'FAILED',
  'DESTROYING',
  'DESTROYED',
]

const FAILURE_STATUSES = new Set<Status>(['FAILED', 'DESTROY_FAILED', 'ORPHANED'])
const ACTIVE_STATUSES = new Set<Status>([
  'REQUESTED',
  'PLANNING',
  'PROVISIONING',
  'HEALTH_CHECK',
  'READY',
])

type Action = 'destroy' | 'extend' | 'retry'

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

  async function runAction(item: PreviewEnvironment, action: Action) {
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
    <div className="app-shell">
      <div className="ambient ambient-one" />
      <div className="ambient ambient-two" />
      <header className="topbar">
        <a className="brand" href="#top" aria-label="Ephemera home">
          <Logo />
          <span>ephemera</span>
          <span className="brand-edition">control plane</span>
        </a>
        <div className="topbar-actions">
          <div
            className="connection-state"
            title={lastUpdated ? `Updated ${lastUpdated.toLocaleTimeString()}` : 'Connecting'}
          >
            <span className={`connection-dot ${error ? 'is-offline' : ''}`} />
            {error ? 'API unavailable' : 'Reconciler online'}
          </div>
          <a
            className="icon-button"
            href="/metrics"
            target="_blank"
            rel="noreferrer"
            aria-label="Open platform metrics"
          >
            <MetricsIcon />
          </a>
          <button
            className="primary-button compact"
            type="button"
            onClick={() => setCreateOpen(true)}
          >
            <PlusIcon />
            New environment
          </button>
        </div>
      </header>

      <main id="top">
        <section className="hero">
          <div className="hero-copy">
            <p className="eyebrow">
              <span /> Preview operations
            </p>
            <h1>
              Every pull request gets
              <br />
              <em>a place to breathe.</em>
            </h1>
            <p className="hero-description">
              Isolated workloads, deterministic routing, and automatic expiry—without handing
              developers the keys to the cloud account.
            </p>
          </div>
          <div className="hero-mark" aria-hidden="true">
            <span className="orbit orbit-one" />
            <span className="orbit orbit-two" />
            <Logo large />
          </div>
        </section>

        <section className="stats-grid" aria-label="Environment summary">
          <StatCard
            label="Ready now"
            value={readyCount.toString().padStart(2, '0')}
            tone="mint"
            detail="Healthy previews"
          />
          <StatCard
            label="In motion"
            value={provisioningCount.toString().padStart(2, '0')}
            tone="amber"
            detail="Planning or deploying"
          />
          <StatCard
            label="Need attention"
            value={failureCount.toString().padStart(2, '0')}
            tone="coral"
            detail="Failed or orphaned"
          />
          <StatCard
            label="Estimated burn"
            value={`$${activeCost.toFixed(2)}`}
            tone="blue"
            detail="Per active hour"
          />
        </section>

        <section className="workspace">
          <div className="section-heading">
            <div>
              <p className="section-kicker">Fleet</p>
              <h2>Preview environments</h2>
            </div>
            <div className="toolbar">
              <label className="search-field">
                <SearchIcon />
                <span className="sr-only">Search environments</span>
                <input
                  value={query}
                  onChange={(event) => setQuery(event.target.value)}
                  placeholder="Search repo, branch, owner…"
                />
                {query !== '' ? (
                  <button type="button" onClick={() => setQuery('')} aria-label="Clear search">
                    ×
                  </button>
                ) : null}
              </label>
              <label className="select-field">
                <span className="sr-only">Filter by status</span>
                <select
                  value={status}
                  onChange={(event) => setStatus(event.target.value as Status | 'ALL')}
                >
                  {STATUS_OPTIONS.map((option) => (
                    <option key={option} value={option}>
                      {statusLabel(option)}
                    </option>
                  ))}
                </select>
                <ChevronIcon />
              </label>
              <button
                className="icon-button refresh-button"
                type="button"
                onClick={() => void refresh()}
                aria-label="Refresh environments"
              >
                <span className={refreshing ? 'spinning' : ''}>
                  <RefreshIcon />
                </span>
              </button>
            </div>
          </div>

          {error ? (
            <div className="error-banner" role="alert">
              <AlertIcon />
              <div>
                <strong>Control plane unreachable</strong>
                <span>{error}</span>
              </div>
              <button type="button" onClick={() => void refresh()}>
                Try again
              </button>
            </div>
          ) : null}

          {notice ? (
            <div className="notice-banner" role="status">
              <span>{notice}</span>
              <button type="button" aria-label="Dismiss message" onClick={() => setNotice(null)}>
                ×
              </button>
            </div>
          ) : null}

          {loading ? (
            <div className="environment-grid" aria-label="Loading environments">
              {[0, 1, 2].map((index) => (
                <SkeletonCard key={index} />
              ))}
            </div>
          ) : filtered.length > 0 ? (
            <div className="environment-grid">
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

      <footer>
        <span>Ephemera control plane</span>
        <span className="footer-rule" />
        <span>Desired state, continuously reconciled</span>
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

function StatCard({
  label,
  value,
  tone,
  detail,
}: {
  label: string
  value: string
  tone: string
  detail: string
}) {
  return (
    <article className={`stat-card tone-${tone}`}>
      <div className="stat-top">
        <span>{label}</span>
        <span className="stat-signal" />
      </div>
      <strong>{value}</strong>
      <p>{detail}</p>
    </article>
  )
}

function EnvironmentCard({
  item,
  busyKey,
  onSelect,
  onAction,
}: {
  item: PreviewEnvironment
  busyKey: string | null
  onSelect: () => void
  onAction: (action: Action) => void
}) {
  const progress = lifecycleProgress(item.status)
  const canExtend = item.status === 'READY' && item.desired_state === 'READY'
  const canDestroy =
    item.status !== 'DESTROYED' && item.status !== 'DESTROYING' && item.status !== 'FAILED_CLEANED'
  const canRetry = FAILURE_STATUSES.has(item.status)

  return (
    <article
      className={`environment-card status-${item.status.toLowerCase().replace('_', '-')}`}
      onClick={onSelect}
    >
      <div className="card-accent" />
      <header className="environment-header">
        <div className="repo-mark">{initials(item.repository)}</div>
        <div className="environment-title">
          <p>{item.repository}</p>
          <h3>PR #{item.pull_request}</h3>
        </div>
        <StatusBadge status={item.status} />
      </header>

      <div className="branch-row">
        <BranchIcon />
        <span>{item.branch}</span>
        <code>{shortSHA(item.commit_sha)}</code>
      </div>

      <div
        className="lifecycle-progress"
        aria-label={`${statusLabel(item.status)}: ${progress}% complete`}
      >
        <div className="progress-track">
          <span style={{ width: `${progress}%` }} />
        </div>
        <div className="progress-labels">
          <span>{item.status_message}</span>
          <span>{progress}%</span>
        </div>
      </div>

      <dl className="environment-facts">
        <div>
          <dt>Compute</dt>
          <dd>
            {item.resources.cpu / 1024} vCPU · {formatMemory(item.resources.memory)}
          </dd>
        </div>
        <div>
          <dt>Expires</dt>
          <dd>{formatRemaining(item.expires_at, item.status)}</dd>
        </div>
        <div>
          <dt>Owner</dt>
          <dd>@{item.owner}</dd>
        </div>
        <div>
          <dt>Est. cost</dt>
          <dd>${item.estimated_hourly_cost.toFixed(3)}/hr</dd>
        </div>
      </dl>

      <div className="environment-actions" onClick={(event) => event.stopPropagation()}>
        {item.preview_url && item.status === 'READY' ? (
          <a
            className="card-primary-action"
            href={item.preview_url}
            target="_blank"
            rel="noreferrer"
          >
            Open preview <ArrowUpRightIcon />
          </a>
        ) : (
          <button className="card-primary-action" type="button" onClick={onSelect}>
            View activity <ArrowRightIcon />
          </button>
        )}
        <div className="card-secondary-actions">
          {canRetry ? (
            <ActionButton
              label="Retry"
              busy={busyKey === `${item.id}:retry`}
              onClick={() => onAction('retry')}
            >
              <RetryIcon />
            </ActionButton>
          ) : null}
          {canExtend ? (
            <ActionButton
              label="Extend TTL"
              busy={busyKey === `${item.id}:extend`}
              onClick={() => onAction('extend')}
            >
              <ClockIcon />
            </ActionButton>
          ) : null}
          {canDestroy ? (
            <ActionButton
              label="Destroy"
              danger
              busy={busyKey === `${item.id}:destroy`}
              onClick={() => onAction('destroy')}
            >
              <TrashIcon />
            </ActionButton>
          ) : null}
        </div>
      </div>
    </article>
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
    <button
      className={danger ? 'mini-button danger' : 'mini-button'}
      type="button"
      title={label}
      aria-label={label}
      disabled={busy}
      onClick={onClick}
    >
      {busy ? <Spinner /> : children}
    </button>
  )
}

function StatusBadge({ status }: { status: Status }) {
  return (
    <span className={`status-badge badge-${status.toLowerCase().replace('_', '-')}`}>
      <i />
      {statusLabel(status)}
    </span>
  )
}

function CreateDialog({
  onClose,
  onCreate,
}: {
  onClose: () => void
  onCreate: (input: CreateEnvironmentInput) => Promise<void>
}) {
  const [submitting, setSubmitting] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

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
      className="dialog-backdrop"
      onMouseDown={(event) => {
        if (event.currentTarget === event.target) onClose()
      }}
    >
      <section
        className="create-dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="create-title"
      >
        <header className="dialog-header">
          <div>
            <p className="section-kicker">Provision</p>
            <h2 id="create-title">New preview environment</h2>
          </div>
          <button
            className="icon-button"
            type="button"
            onClick={onClose}
            aria-label="Close create form"
          >
            ×
          </button>
        </header>
        <form onSubmit={(event) => void submit(event)}>
          <div className="form-grid">
            <label className="field wide">
              <span>Repository</span>
              <input
                name="repository"
                required
                placeholder="acme/billing-api"
                pattern="[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+"
              />
            </label>
            <label className="field">
              <span>Pull request</span>
              <input name="pull_request" type="number" min="1" required placeholder="127" />
            </label>
            <label className="field">
              <span>Owner</span>
              <input name="owner" required placeholder="bilal" />
            </label>
            <label className="field wide">
              <span>Branch</span>
              <input name="branch" required placeholder="feature/new-checkout" />
            </label>
            <label className="field wide">
              <span>Commit SHA</span>
              <input
                name="commit_sha"
                required
                minLength={7}
                maxLength={64}
                pattern="[A-Fa-f0-9]{7,64}"
                placeholder="a81f2aa9b32c"
              />
            </label>
            <label className="field wide">
              <span>
                Container image <small>optional</small>
              </span>
              <input name="image" placeholder="123456789.dkr.ecr.us-east-1.amazonaws.com/app:sha" />
            </label>
          </div>

          <div className="form-divider">
            <span>Resource policy</span>
          </div>
          <div className="form-grid resource-grid">
            <label className="field">
              <span>TTL</span>
              <select name="ttl" defaultValue="6h">
                <option value="1h">1 hour</option>
                <option value="3h">3 hours</option>
                <option value="6h">6 hours</option>
                <option value="12h">12 hours</option>
                <option value="24h">24 hours</option>
              </select>
            </label>
            <label className="field">
              <span>CPU</span>
              <select name="cpu" defaultValue="512">
                <option value="256">0.25 vCPU</option>
                <option value="512">0.5 vCPU</option>
                <option value="1024">1 vCPU</option>
                <option value="2048">2 vCPU</option>
              </select>
            </label>
            <label className="field">
              <span>Memory</span>
              <select name="memory" defaultValue="1024">
                <option value="512">512 MiB</option>
                <option value="1024">1 GiB</option>
                <option value="2048">2 GiB</option>
                <option value="4096">4 GiB</option>
              </select>
            </label>
            <label className="field">
              <span>Database</span>
              <select name="database_mode" defaultValue="shared-schema">
                <option value="shared-schema">Isolated schema</option>
                <option value="none">No database</option>
              </select>
            </label>
          </div>

          {formError ? (
            <p className="form-error" role="alert">
              {formError}
            </p>
          ) : null}
          <div className="dialog-footer">
            <p>
              <ShieldIcon /> Policy limits are enforced by the control plane.
            </p>
            <div>
              <button className="secondary-button" type="button" onClick={onClose}>
                Cancel
              </button>
              <button className="primary-button" type="submit" disabled={submitting}>
                {submitting ? (
                  <>
                    <Spinner /> Queuing…
                  </>
                ) : (
                  <>
                    <PlusIcon /> Create environment
                  </>
                )}
              </button>
            </div>
          </div>
        </form>
      </section>
    </div>
  )
}

function DetailPanel({
  item,
  busyKey,
  onClose,
  onAction,
}: {
  item: PreviewEnvironment
  busyKey: string | null
  onClose: () => void
  onAction: (action: Action) => void
}) {
  const events = [...(item.events ?? [])].reverse()
  const cloudResources = item.cloud_resources ?? []
  const canRetry = FAILURE_STATUSES.has(item.status)
  const canDestroy = !['DESTROYED', 'DESTROYING', 'FAILED_CLEANED'].includes(item.status)
  const canExtend = item.status === 'READY' && item.desired_state === 'READY'
  return (
    <div
      className="detail-backdrop"
      onMouseDown={(event) => {
        if (event.currentTarget === event.target) onClose()
      }}
    >
      <aside className="detail-panel" aria-label={`${item.id} details`}>
        <header className="detail-header">
          <div>
            <StatusBadge status={item.status} />
            <h2>{item.id}</h2>
            <p>
              {item.repository} · PR #{item.pull_request}
            </p>
          </div>
          <button
            className="icon-button"
            type="button"
            onClick={onClose}
            aria-label="Close details"
          >
            ×
          </button>
        </header>
        <div className="detail-actions">
          {item.preview_url && item.status === 'READY' ? (
            <a className="primary-button" href={item.preview_url} target="_blank" rel="noreferrer">
              Open preview <ArrowUpRightIcon />
            </a>
          ) : null}
          {canExtend ? (
            <button
              className="secondary-button"
              type="button"
              disabled={busyKey === `${item.id}:extend`}
              onClick={() => onAction('extend')}
            >
              <ClockIcon /> Extend 1h
            </button>
          ) : null}
          {canRetry ? (
            <button
              className="secondary-button"
              type="button"
              disabled={busyKey === `${item.id}:retry`}
              onClick={() => onAction('retry')}
            >
              <RetryIcon /> Retry
            </button>
          ) : null}
          {canDestroy ? (
            <button
              className="secondary-button danger-text"
              type="button"
              disabled={busyKey === `${item.id}:destroy`}
              onClick={() => onAction('destroy')}
            >
              <TrashIcon /> Destroy
            </button>
          ) : null}
        </div>
        {item.last_error ? (
          <div className="detail-error">
            <AlertIcon />
            <div>
              <strong>Last lifecycle error</strong>
              <code>{item.last_error}</code>
            </div>
          </div>
        ) : null}
        <section className="detail-section">
          <h3>Environment</h3>
          <dl className="detail-list">
            <div>
              <dt>Branch</dt>
              <dd>{item.branch}</dd>
            </div>
            <div>
              <dt>Commit</dt>
              <dd>
                <code>{item.commit_sha}</code>
              </dd>
            </div>
            <div>
              <dt>Image</dt>
              <dd>
                <code>{item.image}</code>
              </dd>
            </div>
            <div>
              <dt>Expires</dt>
              <dd>
                {formatDate(item.expires_at)} · {formatRemaining(item.expires_at, item.status)}
              </dd>
            </div>
            <div>
              <dt>Terraform state</dt>
              <dd>
                <code>{item.terraform_workspace}</code>
              </dd>
            </div>
            <div>
              <dt>Database schema</dt>
              <dd>
                <code>{item.database_schema || 'none'}</code>
              </dd>
            </div>
          </dl>
        </section>
        <section className="detail-section">
          <h3>
            Lifecycle activity <span>{events.length}</span>
          </h3>
          <ol className="event-list">
            {events.map((event, index) => (
              <li key={`${event.at}-${index}`}>
                <span className={`event-dot event-${event.status.toLowerCase()}`} />
                <div>
                  <strong>{statusLabel(event.status)}</strong>
                  <p>{event.message}</p>
                  <time>{formatDate(event.at)}</time>
                </div>
              </li>
            ))}
          </ol>
        </section>
        <section className="detail-section">
          <h3>
            Cloud resources <span>{cloudResources.length}</span>
          </h3>
          {cloudResources.length > 0 ? (
            <ul className="resource-list">
              {cloudResources.map((resource) => (
                <li key={`${resource.kind}:${resource.id}`}>
                  <span>{resource.kind}</span>
                  <code>{resource.id}</code>
                </li>
              ))}
            </ul>
          ) : (
            <p className="muted">Resources will appear after provisioning.</p>
          )}
        </section>
      </aside>
    </div>
  )
}

function EmptyState({ filtered, onCreate }: { filtered: boolean; onCreate: () => void }) {
  return (
    <div className="empty-state">
      <div className="empty-orbit">
        <Logo />
      </div>
      <h3>{filtered ? 'No environments match' : 'Nothing ephemeral—yet'}</h3>
      <p>
        {filtered
          ? 'Try a broader search or clear the status filter.'
          : 'Turn the next pull request into a safe, observable preview.'}
      </p>
      {!filtered ? (
        <button className="primary-button" type="button" onClick={onCreate}>
          <PlusIcon /> Create the first one
        </button>
      ) : null}
    </div>
  )
}

function SkeletonCard() {
  return (
    <div className="environment-card skeleton">
      <span />
      <span />
      <span />
      <span />
    </div>
  )
}

function statusLabel(status: Status | 'ALL') {
  if (status === 'ALL') return 'All statuses'
  return status
    .toLowerCase()
    .replaceAll('_', ' ')
    .replace(/\b\w/g, (letter) => letter.toUpperCase())
}

function lifecycleProgress(status: Status) {
  const progressByStatus: Record<Status, number> = {
    REQUESTED: 8,
    PLANNING: 24,
    PROVISIONING: 58,
    HEALTH_CHECK: 84,
    READY: 100,
    DESTROYING: 64,
    DESTROYED: 100,
    FAILED: 100,
    DESTROY_FAILED: 100,
    ORPHANED: 100,
    FAILED_CLEANED: 100,
  }

  return progressByStatus[status]
}

function initials(repository: string) {
  const name = repository.split('/').at(-1) ?? repository
  return name
    .split(/[-_.]/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join('')
    .toUpperCase()
}

function shortSHA(sha: string) {
  return sha.slice(0, 7)
}

function formatMemory(memory: number) {
  return memory >= 1024 ? `${memory / 1024} GiB` : `${memory} MiB`
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(
    new Date(value),
  )
}

function formatRemaining(value: string, status: Status) {
  if (status === 'DESTROYED' || status === 'FAILED_CLEANED') return 'Cleaned up'
  const milliseconds = new Date(value).getTime() - Date.now()
  if (milliseconds <= 0) return 'Expiring now'
  const minutes = Math.ceil(milliseconds / 60_000)
  if (minutes < 60) return `${minutes}m left`
  const hours = Math.floor(minutes / 60)
  const remainder = minutes % 60
  return remainder === 0 ? `${hours}h left` : `${hours}h ${remainder}m left`
}

function Spinner() {
  return (
    <span className="spinner">
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <circle cx="12" cy="12" r="9" />
      </svg>
    </span>
  )
}

function Logo({ large = false }: { large?: boolean }) {
  return (
    <svg className={large ? 'logo large' : 'logo'} viewBox="0 0 36 36" aria-hidden="true">
      <path d="M18 3 31 10.5v15L18 33 5 25.5v-15L18 3Z" />
      <path d="m11 14 7-4 7 4v8l-7 4-7-4v-8Z" />
      <circle cx="18" cy="18" r="2.3" />
    </svg>
  )
}

function PlusIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M12 5v14M5 12h14" />
    </svg>
  )
}

function SearchIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="11" cy="11" r="7" />
      <path d="m16 16 4 4" />
    </svg>
  )
}

function ChevronIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="m8 10 4 4 4-4" />
    </svg>
  )
}

function RefreshIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M20 6v5h-5M4 18v-5h5" />
      <path d="M6.1 9A7 7 0 0 1 18.5 6.5L20 11M4 13l1.5 4.5A7 7 0 0 0 18 15" />
    </svg>
  )
}

function BranchIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="6" cy="5" r="2" />
      <circle cx="18" cy="6" r="2" />
      <circle cx="6" cy="19" r="2" />
      <path d="M6 7v10M8 13h3a7 7 0 0 0 7-5" />
    </svg>
  )
}

function ArrowRightIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M5 12h14m-5-5 5 5-5 5" />
    </svg>
  )
}

function ArrowUpRightIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M7 17 17 7M8 7h9v9" />
    </svg>
  )
}

function ClockIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="12" cy="12" r="8" />
      <path d="M12 8v5l3 2" />
    </svg>
  )
}

function TrashIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M5 7h14M9 7V4h6v3m2 0-1 13H8L7 7m3 4v5m4-5v5" />
    </svg>
  )
}

function RetryIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M20 7v5h-5" />
      <path d="M18.5 8.5A7 7 0 1 0 19 16" />
    </svg>
  )
}

function ShieldIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="m12 3 8 3v5c0 5-3.4 8.3-8 10-4.6-1.7-8-5-8-10V6l8-3Z" />
      <path d="m9 12 2 2 4-4" />
    </svg>
  )
}

function AlertIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M12 3 2.5 20h19L12 3Z" />
      <path d="M12 9v5m0 3h.01" />
    </svg>
  )
}

function MetricsIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M5 19V9m7 10V5m7 14v-7" />
    </svg>
  )
}
