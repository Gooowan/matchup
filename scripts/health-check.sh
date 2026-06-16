#!/usr/bin/env bash
# Run on the production server to verify MatchUp stack health.
# Bypasses Cloudflare — tests containers and internal endpoints directly.
#
# Usage:
#   cd ~/matchup && ./scripts/health-check.sh
#   ./scripts/health-check.sh --load   # also hit API endpoints (needs EMAIL/PASS in env or args)

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

COMPOSE=(docker compose -f compose.yml -f compose.prod.yml)
OBS=(docker compose -f compose.yml -f compose.prod.yml -f compose.observability.yml)

RUN_LOAD=false
EMAIL="${EMAIL:-}"
PASS="${PASS:-}"

for arg in "$@"; do
  case "$arg" in
    --load) RUN_LOAD=true ;;
    -h|--help)
      cat <<EOF
Usage: $0 [--load]

Checks Docker services, DB, API /health, Redis, Prometheus scrape, optional public URL.

With --load, also tests authenticated API endpoints:
  EMAIL=you@example.com PASS=secret $0 --load
EOF
      exit 0
      ;;
  esac
done

dotenv_get() {
  local key="$1"
  grep -E "^${key}=" .env 2>/dev/null | head -1 | cut -d= -f2- \
    | sed 's/^[[:space:]]*//;s/[[:space:]]*$//' | tr -d '"'"'"
}

ok()   { echo "  ✓ $*"; }
fail() { echo "  ✗ $*"; FAIL=1; }
warn() { echo "  ! $*"; }

FAIL=0

echo "=== MatchUp health check ==="
echo "Time: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
echo ""

echo "--- Docker services ---"
for svc in db redis api web cloudflared; do
  status="$("${COMPOSE[@]}" ps "$svc" --format '{{.Status}}' 2>/dev/null || echo missing)"
  if echo "$status" | grep -qiE 'up|running'; then
    ok "$svc: $status"
  else
    fail "$svc: $status"
  fi
done

echo ""
echo "--- Database ---"
if "${COMPOSE[@]}" exec -T db pg_isready -U "$(dotenv_get POSTGRES_USER)" -d "$(dotenv_get POSTGRES_DB)" >/dev/null 2>&1; then
  ok "Postgres accepting connections"
  counts="$("${COMPOSE[@]}" exec -T db psql -U "$(dotenv_get POSTGRES_USER)" -d "$(dotenv_get POSTGRES_DB)" -tA -c "
    SELECT 'users=' || count(*) FROM users;
    SELECT 'seed_users=' || count(*) FROM users WHERE email LIKE 'seed-%';
    SELECT 'clubs=' || count(*) FROM clubs WHERE is_active = true;
    SELECT 'swipes=' || count(*) FROM matches;
  " 2>/dev/null | tr '\n' ' ')"
  ok "Counts: $counts"
else
  fail "Postgres not ready"
fi

echo ""
echo "--- API (internal) ---"
health="$("${COMPOSE[@]}" exec -T api wget -qO- http://127.0.0.1:8000/health 2>/dev/null || true)"
if echo "$health" | grep -q '"status":"ok"'; then
  ok "GET /health → ok"
else
  fail "GET /health failed: ${health:-no response}"
fi

TOKEN="$(dotenv_get METRICS_TOKEN)"
if [[ -n "$TOKEN" ]]; then
  metrics="$("${COMPOSE[@]}" exec -T api wget -qO- --header="Authorization: Bearer $TOKEN" \
    http://127.0.0.1:8000/metrics 2>/dev/null | head -1 || true)"
  if echo "$metrics" | grep -q 'matchup_'; then
    ok "/metrics reachable (Prometheus token OK)"
  else
    fail "/metrics not returning metrics (check METRICS_TOKEN)"
  fi
else
  warn "METRICS_TOKEN not set — /metrics disabled"
fi

echo ""
echo "--- Redis ---"
if "${COMPOSE[@]}" exec -T redis valkey-cli PING 2>/dev/null | grep -q PONG; then
  ok "Redis PING → PONG"
else
  fail "Redis not responding"
fi

echo ""
echo "--- Observability (optional) ---"
if "${OBS[@]}" ps prometheus --format '{{.Status}}' 2>/dev/null | grep -qi up; then
  ok "Prometheus running"
  target_health="$("${OBS[@]}" exec -T prometheus wget -qO- http://localhost:9090/api/v1/targets 2>/dev/null \
    | grep -o '"job":"matchup-api"[^}]*"health":"[^"]*"' | grep -o '"health":"[^"]*"' | head -1 || true)"
  if echo "$target_health" | grep -q '"health":"up"'; then
    ok "Prometheus scrape matchup-api → up"
  else
    fail "Prometheus scrape matchup-api not up ($target_health)"
  fi
else
  warn "Prometheus not running (start with ./scripts/observability-up.sh --prod)"
fi

echo ""
echo "--- Public URL (via Cloudflare) ---"
PUBLIC_API="${PUBLIC_API:-https://matchup-api.potuzhno.in.ua}"
pub="$(curl -sS -m 10 -o /tmp/mu-health-body.txt -w '%{http_code}' "$PUBLIC_API/health" 2>/dev/null || echo err)"
if [[ "$pub" == "200" ]] && grep -q ok /tmp/mu-health-body.txt 2>/dev/null; then
  ok "Public $PUBLIC_API/health → 200"
else
  fail "Public $PUBLIC_API/health → $pub (Cloudflare 1033 = tunnel/origin issue; use internal checks above)"
  warn "Try: ${COMPOSE[*]} logs cloudflared --tail 20"
fi
rm -f /tmp/mu-health-body.txt

if $RUN_LOAD; then
  echo ""
  echo "--- Authenticated API load (internal via api container) ---"
  if [[ -z "$EMAIL" || -z "$PASS" ]]; then
    warn "Set EMAIL and PASS to run load tests, e.g.: EMAIL=you@example.com PASS=secret $0 --load"
  elif [[ -x "$ROOT/scripts/benchmark.sh" ]]; then
    # Hit API from host through cloudflare if up, else skip
    if [[ "$pub" == "200" ]]; then
      "$ROOT/scripts/benchmark.sh" "$EMAIL" "$PASS" "$PUBLIC_API" 30s
    else
      warn "Skipping benchmark — public API unreachable"
    fi
  else
    warn "scripts/benchmark.sh not found"
  fi
fi

echo ""
if [[ "$FAIL" -eq 0 ]]; then
  echo "=== All critical checks passed ==="
else
  echo "=== Some checks FAILED ==="
  exit 1
fi
