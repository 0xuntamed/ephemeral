import type { Status } from '../types'

export const STATUS_OPTIONS: Array<Status | 'ALL'> = [
  'ALL',
  'READY',
  'PROVISIONING',
  'FAILED',
  'DESTROYING',
  'DESTROYED',
]

export const FAILURE_STATUSES = new Set<Status>(['FAILED', 'DESTROY_FAILED', 'ORPHANED'])

export const ACTIVE_STATUSES = new Set<Status>([
  'REQUESTED',
  'PLANNING',
  'PROVISIONING',
  'HEALTH_CHECK',
  'READY',
])

export type EnvironmentAction = 'destroy' | 'extend' | 'retry'

export function statusLabel(status: Status | 'ALL') {
  if (status === 'ALL') return 'All statuses'
  return status
    .toLowerCase()
    .replaceAll('_', ' ')
    .replace(/\b\w/g, (letter) => letter.toUpperCase())
}

export function lifecycleProgress(status: Status) {
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

export function initials(repository: string) {
  const name = repository.split('/').at(-1) ?? repository
  return name
    .split(/[-_.]/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join('')
    .toUpperCase()
}

export function shortSHA(sha: string) {
  return sha.slice(0, 7)
}

export function formatMemory(memory: number) {
  return memory >= 1024 ? `${memory / 1024} GiB` : `${memory} MiB`
}

export function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(
    new Date(value),
  )
}

export function formatRemaining(value: string, status: Status) {
  if (status === 'DESTROYED' || status === 'FAILED_CLEANED') return 'Cleaned up'
  const milliseconds = new Date(value).getTime() - Date.now()
  if (milliseconds <= 0) return 'Expiring now'
  const minutes = Math.ceil(milliseconds / 60_000)
  if (minutes < 60) return `${minutes}m left`
  const hours = Math.floor(minutes / 60)
  const remainder = minutes % 60
  return remainder === 0 ? `${hours}h left` : `${hours}h ${remainder}m left`
}
