export type Status =
  | 'REQUESTED'
  | 'PLANNING'
  | 'PROVISIONING'
  | 'HEALTH_CHECK'
  | 'READY'
  | 'DESTROYING'
  | 'DESTROYED'
  | 'FAILED'
  | 'DESTROY_FAILED'
  | 'ORPHANED'
  | 'FAILED_CLEANED'

export type DesiredState = 'READY' | 'DESTROYED'

export interface Resources {
  cpu: number
  memory: number
  database_mode: 'shared-schema' | 'none'
}

export interface CloudResource {
  kind: string
  id: string
}

export interface EnvironmentEvent {
  at: string
  status: Status
  message: string
}

export interface PreviewEnvironment {
  id: string
  idempotency_key: string
  repository: string
  branch: string
  commit_sha: string
  pull_request: number
  owner: string
  image: string
  status: Status
  desired_state: DesiredState
  status_message: string
  terraform_workspace: string
  preview_url?: string
  database_schema?: string
  routing_priority: number
  resources: Resources
  cloud_resources: CloudResource[]
  estimated_hourly_cost: number
  created_at: string
  updated_at: string
  expires_at: string
  ready_at?: string
  destroyed_at?: string
  last_error?: string
  provision_attempts: number
  destroy_attempts: number
  events: EnvironmentEvent[]
}

export interface CreateEnvironmentInput {
  repository: string
  branch: string
  commit_sha: string
  pull_request: number
  owner: string
  image?: string
  ttl: string
  resources: Resources
}
