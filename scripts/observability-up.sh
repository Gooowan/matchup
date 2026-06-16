#!/usr/bin/env bash
# Start the observability stack (Prometheus, Grafana, Loki, Tempo, Promtail).
#
# Usage:
#   ./scripts/observability-up.sh              # dev (compose.yml only)
#   ./scripts/observability-up.sh --prod       # prod (+ compose.prod.yml)
#
# Ensures METRICS_TOKEN exists in .env and writes secrets/metrics_token for
# Prometheus bearer auth (Prometheus does not expand env vars in its config).

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

PROD=false
for arg in "$@"; do
  case "$arg" in
    --prod) PROD=true ;;
    -h|--help)
      echo "Usage: $0 [--prod]"
      exit 0
      ;;
  esac
done

ENV_FILE="$ROOT/.env"
if [[ ! -f "$ENV_FILE" ]]; then
  echo "ERROR: $ENV_FILE not found. Copy .example.env to .env first." >&2
  exit 1
fi

# Read a single key from .env without sourcing (avoids shell metacharacters in values).
dotenv_get() {
  local key="$1"
  grep -E "^${key}=" "$ENV_FILE" 2>/dev/null | head -1 | cut -d= -f2- \
    | sed 's/^[[:space:]]*//;s/[[:space:]]*$//' | tr -d '"'"'"
}

METRICS_TOKEN="${METRICS_TOKEN:-$(dotenv_get METRICS_TOKEN)}"
GRAFANA_PASSWORD="${GRAFANA_PASSWORD:-$(dotenv_get GRAFANA_PASSWORD)}"

if [[ -z "${METRICS_TOKEN:-}" ]]; then
  METRICS_TOKEN="$(openssl rand -hex 32)"
  if grep -q '^METRICS_TOKEN=' "$ENV_FILE"; then
    sed -i.bak "s/^METRICS_TOKEN=.*/METRICS_TOKEN=$METRICS_TOKEN/" "$ENV_FILE"
    rm -f "$ENV_FILE.bak"
  else
    echo "METRICS_TOKEN=$METRICS_TOKEN" >> "$ENV_FILE"
  fi
  echo "Generated METRICS_TOKEN and saved to .env"
fi

mkdir -p "$ROOT/secrets"
printf '%s' "$METRICS_TOKEN" > "$ROOT/secrets/metrics_token"
# Prometheus runs as non-root (nobody); 600 root-only blocks the container from reading it.
chmod 644 "$ROOT/secrets/metrics_token"

COMPOSE=(docker compose -f compose.yml)
if $PROD; then
  COMPOSE+=(-f compose.prod.yml)
fi
COMPOSE+=(-f compose.observability.yml)

echo "Starting observability stack..."
"${COMPOSE[@]}" up -d prometheus grafana loki tempo promtail

echo "Restarting API to pick up METRICS_TOKEN / OTEL_ENDPOINT..."
"${COMPOSE[@]}" up -d api

echo ""
echo "Grafana:  http://$(hostname -I 2>/dev/null | awk '{print $1}' || echo localhost):3001"
echo "Login:    admin / ${GRAFANA_PASSWORD:-matchup_dev}"
echo ""
echo "Verify Prometheus scrape target:"
echo "  ${COMPOSE[*]} exec prometheus wget -qO- http://localhost:9090/api/v1/targets | grep -o '\"health\":\"[^\"]*\"' | head -5"
