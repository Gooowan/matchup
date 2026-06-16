#!/usr/bin/env bash
# Generate API load and print latency percentiles (via hey or curl).
#
# Usage:
#   ./scripts/benchmark.sh EMAIL PASSWORD [API_URL] [DURATION]
#
# Examples:
#   ./scripts/benchmark.sh user@example.com secret https://matchup-api.example.com 2m
#   ./scripts/benchmark.sh user@example.com secret http://localhost:8000 30s
#
# Requires: curl. Optional: hey (brew install hey / go install github.com/rakyll/hey@latest)

set -euo pipefail

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" || $# -lt 2 ]]; then
  cat <<EOF
Usage: $0 EMAIL PASSWORD [API_URL] [DURATION]

Runs load against key endpoints and prints latency percentiles.
Use the same args before and after seed-all for a fair comparison.

Endpoints: /matchup/feed, /user/profile, /chats, /map/nearby/count
EOF
  exit 0
fi

EMAIL="$1"
PASS="$2"
API="${3:-http://localhost:8000}"
DURATION="${4:-2m}"

COOKIE_JAR="$(mktemp)"
HDR="$(mktemp)"
trap 'rm -f "$COOKIE_JAR" "$HDR"' EXIT

echo "=== Login ==="
HTTP="$(curl -sS -o /dev/null -D "$HDR" -w '%{http_code}' -c "$COOKIE_JAR" \
  -X POST "$API/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASS\"}")"
if [[ "$HTTP" != "200" ]]; then
  echo "ERROR: login failed (HTTP $HTTP)" >&2
  exit 1
fi

TOKEN="$(grep -i '^set-cookie: auth_token=' "$HDR" | head -1 | sed -E 's/.*auth_token=([^;]+).*/\1/i')"
if [[ -z "$TOKEN" ]]; then
  TOKEN="$(awk '$6 == "auth_token" { print $7 }' "$COOKIE_JAR" | tail -1)"
fi
if [[ -z "$TOKEN" ]]; then
  echo "ERROR: auth_token not found after login" >&2
  exit 1
fi

AUTH=(-H "Authorization: Bearer $TOKEN")

run_hey() {
  local name="$1"
  shift
  echo ""
  echo "=== $name (${DURATION}, hey) ==="
  if command -v hey >/dev/null 2>&1; then
    hey -z "$DURATION" -c 20 "$@" | tee "/tmp/mu-bench-${name// /-}.txt" | tail -20
  else
    echo "hey not installed — falling back to curl loop (100 requests)"
    run_curl "$name" "$@"
  fi
}

run_curl() {
  local name="$1"
  shift
  local url="$1"
  local method="${2:-GET}"
  local body="${3:-}"

  echo ""
  echo "=== $name (100 requests, curl) ==="
  local tmp
  tmp="$(mktemp)"
  for _ in $(seq 1 100); do
    if [[ "$method" == "POST" && -n "$body" ]]; then
      curl -o /dev/null -s -w '%{time_total}\n' -X POST "${AUTH[@]}" \
        -H 'Content-Type: application/json' -d "$body" "$url" >> "$tmp"
    else
      curl -o /dev/null -s -w '%{time_total}\n' "${AUTH[@]}" "$url" >> "$tmp"
    fi
  done
  sort -n "$tmp" | awk -v n="$(wc -l < "$tmp")" '
    { a[NR]=$1 }
    END {
      printf "  p50: %.0f ms\n", a[int(n*0.50)]*1000
      printf "  p95: %.0f ms\n", a[int(n*0.95)]*1000
      printf "  p99: %.0f ms\n", a[int(n*0.99)]*1000
    }'
  rm -f "$tmp"
}

echo "API: $API  duration: $DURATION"

run_hey "feed" "${AUTH[@]}" "$API/matchup/feed?limit=20"
run_hey "profile" "${AUTH[@]}" "$API/user/profile"
run_hey "chats" "${AUTH[@]}" "$API/chats"
run_hey "map nearby" "${AUTH[@]}" -m POST \
  -H "Content-Type: application/json" \
  -d '{"latitude":50.45,"longitude":30.52,"count":20}' \
  "$API/map/nearby/count"

echo ""
echo "Done. Re-run after seed-all with the same command for comparison."
echo "For Grafana percentiles: Dashboards → MatchUp → API Overview (during load window)."
