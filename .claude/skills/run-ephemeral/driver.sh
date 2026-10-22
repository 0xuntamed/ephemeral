#!/usr/bin/env bash
# Drive the Ephemera stack (Go control-plane API + Vite dashboard) for agents.
#
#   .claude/skills/run-ephemeral/driver.sh start        build API, launch API + dashboard, wait until healthy
#   .claude/skills/run-ephemeral/driver.sh status       what is listening on :8080 / :5173
#   .claude/skills/run-ephemeral/driver.sh smoke        create -> READY -> preview -> idempotency -> delete -> re-create -> delete
#   .claude/skills/run-ephemeral/driver.sh screenshot [out.png] [WxH]   headless Chrome/Edge screenshot of the dashboard
#   .claude/skills/run-ephemeral/driver.sh logs         tail API + dashboard logs
#   .claude/skills/run-ephemeral/driver.sh stop         kill whatever listens on :8080 / :5173
#
# State, logs and screenshots go to .cache/run/ (gitignored). The API uses an isolated
# JSON store there, wiped on every `start`, so runs never touch data/environments.json.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
RUN_REL=".cache/run"
RUN="$ROOT/$RUN_REL"
API_PORT=8080
WEB_PORT=5173
API="http://localhost:$API_PORT"
WEB="http://localhost:$WEB_PORT"
WINDOWS=false; [ "${OS:-}" = "Windows_NT" ] && WINDOWS=true
EXE=""; $WINDOWS && EXE=".exe"
API_BIN_REL="bin/ephemeral-api$EXE"

mkdir -p "$RUN"

log() { printf '[driver] %s\n' "$*"; }
die() { printf '[driver] FAIL: %s\n' "$*" >&2; exit 1; }

# PIDs listening on a TCP port. Killing by port is deliberate: `go run` and `npm run dev`
# leave the real server in a child process, so killing the launcher does not free the port.
port_pids() {
  if $WINDOWS; then
    netstat -ano | awk -v p=":$1" '$1=="TCP" && substr($2, length($2)-length(p)+1)==p && $4=="LISTENING" {print $5}' | sort -u
  else
    lsof -ti "tcp:$1" -sTCP:LISTEN 2>/dev/null || true
  fi
}

kill_port() {
  local pid
  for pid in $(port_pids "$1"); do
    if $WINDOWS; then taskkill //F //T //PID "$pid" >/dev/null; else kill "$pid"; fi
    log "stopped PID $pid on :$1"
  done
}

http_code() { curl -s -o /dev/null -w '%{http_code}' "$1" || true; }

wait_http() { # url timeout_seconds
  local _
  for _ in $(seq 1 "$2"); do
    [ "$(http_code "$1")" = "200" ] && return 0
    sleep 1
  done
  return 1
}

# Evaluate a JS expression against JSON on stdin (no jq on this machine; node is required anyway).
json() { # json_string expression_using_d
  node -e 'let s="";process.stdin.on("data",c=>s+=c).on("end",()=>console.log(new Function("d","return "+process.argv[1])(JSON.parse(s))))' "$2" <<<"$1"
}

start() {
  [ -z "$(port_pids $API_PORT)" ] || die ":$API_PORT already in use - run '$0 stop' first"
  [ -z "$(port_pids $WEB_PORT)" ] || die ":$WEB_PORT already in use - run '$0 stop' first"
  cd "$ROOT"

  log "building API -> $API_BIN_REL"
  go build -o "$API_BIN_REL" ./controller/cmd/api

  rm -f "$RUN/environments.json"
  log "starting API on :$API_PORT (demo provider, JSON store $RUN_REL/environments.json)"
  # Relative paths on purpose: a native Windows exe cannot resolve Git Bash /d/... paths.
  DATABASE_URL="" FILE_STORE_PATH="$RUN_REL/environments.json" PROVIDER=demo RUN_WORKER=true \
    WORKER_INTERVAL=1s DEMO_STAGE_DELAY=500ms \
    nohup "./$API_BIN_REL" </dev/null >"$RUN/api.log" 2>&1 &
  wait_http "$API/readyz" 30 || die "API not ready after 30s; see $RUN_REL/api.log"

  if [ ! -d dashboard/node_modules ]; then
    log "dashboard/node_modules missing - running npm ci"
    (cd dashboard && npm ci --no-audit --no-fund >/dev/null)
  fi
  log "starting dashboard on :$WEB_PORT"
  # --strictPort: without it Vite silently hops to 5174 when 5173 is taken.
  (cd dashboard && nohup npx vite --port "$WEB_PORT" --strictPort </dev/null >"$RUN/dashboard.log" 2>&1 &)
  wait_http "$WEB/" 60 || die "dashboard not up after 60s; see $RUN_REL/dashboard.log"

  log "up: API $API  dashboard $WEB"
}

status() {
  local p pids
  for p in $API_PORT $WEB_PORT; do
    pids="$(port_pids "$p" | tr '\n' ' ')"
    if [ -n "$pids" ]; then echo ":$p listening (PID $pids)"; else echo ":$p down"; fi
  done
  [ "$(http_code "$API/readyz")" = "200" ] && echo "API readyz: $(curl -s "$API/readyz")" || true
}

wait_status() { # id wanted timeout_seconds
  local s="" _
  for _ in $(seq 1 "$3"); do
    s="$(json "$(curl -s "$API/api/v1/environments/$1")" d.status)"
    [ "$s" = "$2" ] && { log "$1 -> $2"; return 0; }
    sleep 1
  done
  die "$1 stuck at '$s' (wanted $2 within $3s)"
}

smoke() {
  wait_http "$API/readyz" 5 || die "API not running - run '$0 start' first"
  # Fresh SHA per run so each smoke starts from its own environment, never one an earlier
  # run left mid-teardown under the same key.
  local sha key body created again id preview replacement
  sha="$(node -e 'console.log(require("crypto").randomBytes(6).toString("hex"))')"
  key="acme/billing-api:127:$sha"
  body='{"repository":"acme/billing-api","branch":"feature/new-checkout","commit_sha":"'"$sha"'","pull_request":127,"owner":"smoke","image":"demo-service:latest","ttl":"6h","resources":{"cpu":512,"memory":1024,"database_mode":"shared-schema"}}'

  created="$(curl -sf -X POST "$API/api/v1/environments" -H 'Content-Type: application/json' -H "Idempotency-Key: $key" -d "$body")" \
    || die "POST /api/v1/environments failed"
  id="$(json "$created" d.id)"
  log "created $id (status $(json "$created" d.status))"
  wait_status "$id" READY 30

  preview="$(json "$(curl -s "$API/api/v1/environments/$id")" d.preview_url)"
  [ "$(http_code "$preview")" = "200" ] || die "preview $preview did not return 200"
  log "preview page 200: $preview"

  again="$(curl -sf -X POST "$API/api/v1/environments" -H 'Content-Type: application/json' -H "Idempotency-Key: $key" -d "$body")" \
    || die "idempotent re-POST failed"
  [ "$(json "$again" d.id)" = "$id" ] || die "idempotent re-POST returned a different environment"
  log "idempotent re-POST returned the same environment"

  [ "$(http_code "$API/metrics")" = "200" ] || die "/metrics did not return 200"
  log "/metrics 200"

  if [ "$(http_code "$WEB/")" = "200" ]; then
    [ "$(http_code "$WEB/api/v1/environments")" = "200" ] || die "dashboard proxy $WEB/api -> $API broken"
    log "dashboard proxy $WEB/api -> $API OK"
  else
    log "dashboard not running - skipped proxy check"
  fi

  curl -sf -X DELETE "$API/api/v1/environments/$id" >/dev/null || die "DELETE failed"
  wait_status "$id" DESTROYED 30

  # A destroyed environment releases its key, so the same request must provision a fresh one.
  again="$(curl -sf -X POST "$API/api/v1/environments" -H 'Content-Type: application/json' -H "Idempotency-Key: $key" -d "$body")" \
    || die "re-POST after destroy failed"
  replacement="$(json "$again" d.id)"
  [ "$replacement" != "$id" ] || die "re-POST after destroy replayed the destroyed environment"
  log "re-POST after destroy created $replacement"
  wait_status "$replacement" READY 30
  curl -sf -X DELETE "$API/api/v1/environments/$replacement" >/dev/null || die "DELETE failed"
  wait_status "$replacement" DESTROYED 30
  log "PASS"
}

screenshot() {
  local out="${1:-$RUN/dashboard.png}" size="${2:-1440,1300}" browser="" b out_arg profile_arg
  for b in "/c/Program Files/Google/Chrome/Application/chrome.exe" \
           "/c/Program Files (x86)/Microsoft/Edge/Application/msedge.exe" \
           "$(command -v chromium 2>/dev/null || true)" "$(command -v google-chrome 2>/dev/null || true)"; do
    [ -n "$b" ] && [ -f "$b" ] && { browser="$b"; break; }
  done
  [ -n "$browser" ] || die "no Chrome/Edge/Chromium found"
  [ "$(http_code "$WEB/")" = "200" ] || die "dashboard not running - run '$0 start' first"

  # Chrome must get an absolute --screenshot path: given a relative one on Windows it
  # exits 0, prints nothing, and the PNG never appears where you expect it.
  case "$out" in /*) ;; *) out="$PWD/$out" ;; esac
  mkdir -p "$(dirname "$out")"; rm -f "$out"
  out_arg="$out"; profile_arg="$RUN/chrome-profile"
  if $WINDOWS; then out_arg="$(cygpath -wa "$out")"; profile_arg="$(cygpath -wa "$profile_arg")"; fi
  # Throwaway profile: never touches the user's real browser profile, and does not
  # collide with a Chrome window the user already has open.
  "$browser" --headless=new --disable-gpu --hide-scrollbars --no-first-run \
    --user-data-dir="$profile_arg" --window-size="$size" --virtual-time-budget=8000 \
    --screenshot="$out_arg" "$WEB/" >/dev/null 2>&1 || true
  [ -s "$out" ] || die "screenshot not written"
  log "screenshot -> $out"
}

logs() {
  local f
  for f in api dashboard; do
    echo "=== $RUN_REL/$f.log ==="
    tail -n "${1:-20}" "$RUN/$f.log" 2>/dev/null || echo "(none)"
  done
}

stop() {
  kill_port "$API_PORT"
  kill_port "$WEB_PORT"
  [ -z "$(port_pids $API_PORT)$(port_pids $WEB_PORT)" ] || die "ports still in use"
  log "stopped"
}

cmd="${1:-}"; shift || true
case "$cmd" in
  start|status|smoke|screenshot|logs|stop) "$cmd" "$@" ;;
  *) sed -n '2,12p' "$0"; exit 2 ;;
esac
