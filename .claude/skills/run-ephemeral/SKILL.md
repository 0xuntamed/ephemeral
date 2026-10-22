---
name: run-ephemeral
description: Build, run, start, stop, smoke-test and screenshot the Ephemera control plane — the Go API server on :8080 and the React ops dashboard on :5173. Use when asked to run/launch/start/stop the server or dashboard, check that a change works end-to-end in the real app, create or destroy a preview environment against the local API, take a screenshot of the dashboard, or run the Go and dashboard test suites.
---

# Run Ephemera (API + dashboard)

Ephemera is a Go control plane (`controller/cmd/api`) that turns PRs into expiring preview
environments, plus a Vite/React dashboard (`dashboard/`) that proxies `/api` to it. Locally it
runs with the safe `demo` provider and a JSON-file store, so no Docker, Postgres or AWS is needed.

Agents drive it with **`.claude/skills/run-ephemeral/driver.sh`** (bash; verified in Git Bash on
Windows 11). All paths here are relative to the repo root. Logs, state and screenshots go to
`.cache/run/` (gitignored).

## Prerequisites

Go ≥ 1.24 (verified 1.26.2), Node ≥ 22 (verified 22.13.0), `curl`, and Chrome or Edge for
screenshots. No `jq` needed — the driver parses JSON with `node`.

## Run (agent path)

```bash
.claude/skills/run-ephemeral/driver.sh start
```
Builds `bin/ephemeral-api.exe`, wipes and uses `.cache/run/environments.json` as the store,
starts the API (embedded worker, 500 ms demo stages), runs `npm ci` if `dashboard/node_modules`
is missing, starts Vite with `--strictPort`, and blocks until both answer. Cold ≈ 50 s (Go
build), warm ≈ 7 s; both processes keep running after the command returns.

```bash
.claude/skills/run-ephemeral/driver.sh smoke
```
Full lifecycle in ~5 s, exits non-zero on the first failure: create (fresh commit SHA) →
`READY` → preview page 200 → idempotent re-POST returns same id → `/metrics` 200 →
dashboard proxy 200 → `DELETE` → `DESTROYED` → `PASS`. Safe to re-run.

```bash
.claude/skills/run-ephemeral/driver.sh screenshot
```
Headless Chrome/Edge → `.cache/run/dashboard.png` (1440×1300: header, stat tiles and the first
row of environment cards). Optional args: output path and `W,H`, e.g.
`driver.sh screenshot .cache/run/full.png 1440,2400`. **Open the PNG and look at it.**

```bash
.claude/skills/run-ephemeral/driver.sh status
```
```bash
.claude/skills/run-ephemeral/driver.sh logs
```
```bash
.claude/skills/run-ephemeral/driver.sh stop
```

To make a screenshot show a live environment, create one with a commit SHA not used before in
this run (new → HTTP 201; a replayed key → HTTP 200 with the old record):

```bash
curl -s -o /dev/null -w "create: %{http_code}\n" -X POST http://localhost:8080/api/v1/environments \
  -H 'Content-Type: application/json' -H 'Idempotency-Key: acme/billing-api:128:5e1ec7ed0001' \
  -d '{"repository":"acme/billing-api","branch":"feature/fresh-sha","commit_sha":"5e1ec7ed0001","pull_request":128,"owner":"bilal","image":"demo-service:latest","ttl":"6h","resources":{"cpu":512,"memory":1024,"database_mode":"shared-schema"}}'
```

For interactive flows (clicking **+ New Environment**, extend/delete buttons), open
`http://localhost:5173` in Claude's built-in browser pane after `start`; the driver's
screenshot is static.

## Test / direct invocation

Most changes land in `internal/controlplane`, `internal/store`, `internal/environment` — test
those directly without launching anything:

```bash
go test ./internal/... ./controller/...
```
```bash
cd dashboard && npm run check
```
Go ≈ 15 s; `npm run check` (tsc + eslint + prettier) ≈ 55 s.

## Run (human path)

`go run ./controller/cmd/api` in one terminal, `cd dashboard && npm run dev` in another, open
http://localhost:5173. This uses the real `data/environments.json`. Stop with the driver's `stop`.
`docker compose up --build` exists (Postgres + API +
nginx dashboard on :3000) but was **not verified** — the Docker daemon was not running.

## Gotchas

- **Idempotency on repo + PR + commit SHA is permanent.** After `DELETE`, POSTing the same key
  returns the old `DESTROYED` record (HTTP 200), not a new environment. Change `commit_sha` to
  reprovision. The driver's `smoke` generates a fresh SHA per run for this reason.
- **Kill by port, never by process name.** `go run` runs a temp-dir child exe, and `npm run dev`
  → `npx` → `node`, so killing the launcher leaves the port bound. On this machine ~10 unrelated
  `node.exe` processes belong to Claude Code and its tooling — `taskkill //IM node.exe` would take
  those down. `driver.sh stop` resolves PIDs from `netstat` for :8080/:5173 and kills just those trees.
- **Chrome `--screenshot` needs an absolute Windows path.** Given a relative path it exits 0,
  prints nothing, and no PNG appears. The driver converts with `cygpath -wa`, and uses a throwaway
  profile in `.cache/run/chrome-profile` so it never touches (or collides with) your real Chrome.
- **Native Windows exes can't read Git Bash paths** (`/d/devops/...`) in env vars. The driver
  passes `FILE_STORE_PATH` as a repo-relative path with cwd at the repo root.
- **Vite silently moves to 5174** if 5173 is taken; the driver passes `--strictPort` so it fails
  loudly instead.
- **The JSON store file is created lazily** on the first write — `.cache/run/environments.json`
  won't exist right after `start`.
- **The preview URL is served by the API itself** (`http://localhost:8080/preview/<id>`, from the
  `DEMO_PREVIEW_BASE_URL` default) — a simulator page, not a deployment.
- **`go test ./...` walks into `dashboard/node_modules`** (the `flatted` npm package ships Go
  code). Harmless today; scope to `./internal/... ./controller/...`.
- **`go test -race` does not work here** — no C toolchain, cgo disabled. The README's
  `go test -race ./...` verification line fails on this machine; run without `-race`.

## Troubleshooting

| Symptom | Fix |
|---|---|
| `smoke`: `stuck at 'DESTROYED' (wanted READY)` | Reused idempotency key — see first gotcha. Fixed in `smoke`; in hand-written curl, change `commit_sha`. |
| `screenshot`: `FAIL: screenshot not written` | Relative `--screenshot` path — fixed in the driver. If it recurs, run the Chrome command with `--enable-logging=stderr --v=0`. |
| `start`: `:8080 already in use` / `:5173 already in use` | `driver.sh stop`, then `start` again. |
| `go: -race requires cgo; enable cgo by setting CGO_ENABLED=1` | Run `go test` without `-race`. |
| Dashboard shows `READY 00` right after a POST | The POST replayed a destroyed environment (HTTP 200, not 201). Use a fresh SHA. |
