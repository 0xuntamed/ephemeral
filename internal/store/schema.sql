CREATE TABLE IF NOT EXISTS environments (
    id               text PRIMARY KEY,
    idempotency_key  text NOT NULL UNIQUE,
    status           text NOT NULL,
    desired_state    text NOT NULL,
    repository       text NOT NULL,
    owner            text NOT NULL,
    pull_request     integer NOT NULL,
    expires_at       timestamptz NOT NULL,
    created_at       timestamptz NOT NULL,
    updated_at       timestamptz NOT NULL,
    document         jsonb NOT NULL
);

CREATE INDEX IF NOT EXISTS environments_status_idx ON environments (status);
CREATE INDEX IF NOT EXISTS environments_expiry_idx ON environments (expires_at) WHERE status <> 'DESTROYED';
CREATE INDEX IF NOT EXISTS environments_repo_idx ON environments (lower(repository), created_at DESC);
