# Lifecycle and reconciliation

```mermaid
stateDiagram-v2
    [*] --> REQUESTED
    REQUESTED --> PLANNING
    PLANNING --> PROVISIONING
    PROVISIONING --> HEALTH_CHECK
    HEALTH_CHECK --> READY
    READY --> DESTROYING: PR close / API / TTL
    REQUESTED --> DESTROYING: cancelled before start
    PLANNING --> FAILED
    PROVISIONING --> FAILED
    HEALTH_CHECK --> FAILED
    FAILED --> REQUESTED: retry
    FAILED --> DESTROYING: cleanup
    DESTROYING --> DESTROYED
    DESTROYING --> DESTROY_FAILED
    DESTROY_FAILED --> DESTROYING: bounded retry
    PLANNING --> ORPHANED: operation timeout
    PROVISIONING --> ORPHANED: operation timeout
    HEALTH_CHECK --> ORPHANED: operation timeout
    DESTROYING --> ORPHANED: operation timeout
    ORPHANED --> DESTROYING
    DESTROYED --> [*]
```

## Desired state

- `READY`: the pull request exists and its TTL is valid.
- `DESTROYED`: the PR closed, a user requested cleanup, or TTL expired.

The status describes current reality; desired state describes the target. Cleanup requests therefore remain recoverable even if the webhook sender or API client disappears immediately afterward.

## Idempotency

The default key is `lower(repository):pull_request:commit_sha`. A unique database constraint guarantees one environment for a webhook identity. A replay returns HTTP 200 with `Idempotency-Replayed: true`; the first request returns HTTP 201.

Once an environment is `DESTROYED` or `FAILED_CLEANED`, its key is released on the next request: the old record's key is renamed to `<key>#released:<id>`, an event records the release, and the request provisions a fresh environment (HTTP 201). This lets a CI re-run for the same commit recover after the TTL reaper removed its preview. An environment that is still being torn down keeps its key, because its cloud resources and listener priority still exist; the request returns it with `desired_state: DESTROYED`.

## TTL reaper

Every environment has `expires_at`. Each reconciliation pass marks expired active environments with desired state `DESTROYED`. A subsequent worker claim performs provider cleanup. This is independent of PR-close delivery.

## Crash recovery

Planning, provisioning, health-check, and destroy operations have an `updated_at` lease. If that lease exceeds `OPERATION_TIMEOUT`, reconciliation marks the environment `ORPHANED` and schedules cleanup. Destroy failures use exponential backoff capped at 64 seconds. Errors are truncated to 4,000 characters before persistence, and the latest 100 lifecycle events are retained.

## Database schema lifecycle

When `database_mode=shared-schema` and `PREVIEW_DATABASE_URL` is configured, the provider wrapper creates the validated schema before workload provisioning. If provisioning fails, it makes a bounded best-effort cleanup. Normal destruction removes the workload first and then drops the schema with `CASCADE`; a schema cleanup error moves the whole operation into the retryable cleanup path.
