#!/usr/bin/env bash
# Generate sustained API traffic to populate Grafana dashboards.
# Run ON THE SERVER while watching Grafana (Last 15 minutes).
#
# Usage:
#   ./scripts/grafana-fill.sh
#   LOOPS=120 ./scripts/grafana-fill.sh
#   EMAIL=your@email.com PASS=secret ./scripts/grafana-fill.sh

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

LOOPS="${LOOPS:-120}"
SLEEP="${SLEEP:-0.5}"
EMAIL="${EMAIL:-seed-00001@matchup.local}"
PASS="${PASS:-password123}"

API="${API:-http://$(docker compose -f compose.yml -f compose.prod.yml ps -q api \
  | xargs docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}'):8000}"

JAR=$(mktemp)
HDR=$(mktemp)
trap 'rm -f "$JAR" "$HDR"' EXIT

echo "=== Grafana fill ==="
echo "API:   $API"
echo "User:  $EMAIL"
echo "Loops: $LOOPS (~$((LOOPS * 8)) requests + ~$((LOOPS * 5)) swipes)"
echo ""

echo "Login..."
HTTP=$(curl -sS -m 15 -D "$HDR" -o /dev/null -w '%{http_code}' -c "$JAR" \
  -X POST "$API/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASS\"}")

if [[ "$HTTP" != "200" ]]; then
  echo "Login failed (HTTP $HTTP)" >&2
  exit 1
fi

# Secure cookies are not stored in the jar over plain HTTP — parse Set-Cookie header.
TOKEN=$(grep -i '^set-cookie: auth_token=' "$HDR" | head -1 | sed -E 's/.*auth_token=([^;]+).*/\1/i')
if [[ -z "$TOKEN" ]]; then
  TOKEN=$(awk '$6 == "auth_token" { print $7 }' "$JAR" | tail -1)
fi
if [[ -z "$TOKEN" ]]; then
  echo "No auth_token after login (check EMAIL/PASS)" >&2
  exit 1
fi
echo "Login OK"

A=(-H "Authorization: Bearer $TOKEN")

# Pre-load swipe targets (faster than DB query every loop).
mapfile -t TARGETS < <(docker compose -f compose.yml -f compose.prod.yml exec -T db \
  psql -U matchup -d matchup -tA \
  -c "SELECT id::text FROM users WHERE email LIKE 'seed-%' ORDER BY email LIMIT 50")
echo "Loaded ${#TARGETS[@]} swipe targets"
TIDX=0

next_target() {
  if [[ ${#TARGETS[@]} -eq 0 ]]; then
    echo ""
    return
  fi
  echo "${TARGETS[$((TIDX % ${#TARGETS[@]}))]}"
  TIDX=$((TIDX + 1))
}

swipe() {
  local action="$1"
  local source="${2:-feed}"
  local tid
  tid="$(next_target)"
  [[ -z "$tid" ]] && return 0
  curl -sS -o /dev/null -X POST "${A[@]}" -H 'Content-Type: application/json' \
    -d "{\"target_user_id\":\"$tid\",\"action\":\"$action\",\"source\":\"$source\"}" \
    "$API/matchup/swipe" || true
}

for i in $(seq 1 "$LOOPS"); do
  # health
  curl -sS -o /dev/null "$API/health" || true
  curl -sS -o /dev/null -I "$API/health" || true

  # read endpoints
  curl -sS -o /dev/null "${A[@]}" "$API/matchup/feed?limit=20" || true
  curl -sS -o /dev/null "${A[@]}" "$API/user/profile" || true
  curl -sS -o /dev/null "${A[@]}" "$API/me/profile" || true
  curl -sS -o /dev/null "${A[@]}" "$API/chats" || true
  curl -sS -o /dev/null -X POST "${A[@]}" -H 'Content-Type: application/json' \
    -d '{"latitude":50.45,"longitude":30.52,"count":20}' "$API/map/nearby/count" || true

  # swipes — 5 per loop (fills Swipes/min + Matches/min panels)
  swipe LIKE feed
  swipe LIKE tier1
  swipe PASS feed
  swipe LIKE feed
  swipe PASS tier1

  # extra feed/profile between swipes
  curl -sS -o /dev/null "${A[@]}" "$API/matchup/feed?limit=10" || true
  swipe LIKE feed
  swipe LIKE tier2

  [[ $((i % 10)) -eq 0 ]] && echo "  batch $i/$LOOPS done"
  sleep "$SLEEP"
done

echo ""
echo "Done — refresh Grafana (Last 15 minutes)"
