import type { CreateEnvironmentInput, PreviewEnvironment } from './types'

const API_BASE = import.meta.env.VITE_API_URL ?? '/api/v1'

interface ListResponse {
  items: PreviewEnvironment[]
  total: number
}

interface ErrorResponse {
  error?: { message?: string }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  })
  if (!response.ok) {
    const body = (await response.json().catch(() => ({}))) as ErrorResponse
    throw new Error(body.error?.message ?? `Request failed with status ${response.status}`)
  }
  return (await response.json()) as T
}

export async function listEnvironments(signal?: AbortSignal): Promise<PreviewEnvironment[]> {
  const response = await request<ListResponse>('/environments', { signal })
  return response.items
}

export function createEnvironment(input: CreateEnvironmentInput): Promise<PreviewEnvironment> {
  return request<PreviewEnvironment>('/environments', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function destroyEnvironment(id: string): Promise<PreviewEnvironment> {
  return request<PreviewEnvironment>(`/environments/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  })
}

export function extendEnvironment(id: string, duration = '1h'): Promise<PreviewEnvironment> {
  return request<PreviewEnvironment>(`/environments/${encodeURIComponent(id)}/extend`, {
    method: 'POST',
    body: JSON.stringify({ duration }),
  })
}

export function retryEnvironment(id: string): Promise<PreviewEnvironment> {
  return request<PreviewEnvironment>(`/environments/${encodeURIComponent(id)}/retry`, {
    method: 'POST',
  })
}
