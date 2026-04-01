#!/usr/bin/env bash
set -e

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT_DIR"

# Use Node 22 via nvm (required by Capacitor/Vite)
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
nvm use 22 2>/dev/null || { echo "Node 22 not found, installing..."; nvm install 22; }

# 1. Start Docker services (db, redis, minio, api, web)
echo "==> Starting Docker services..."
docker compose up -d

# Wait for API to be healthy
echo "==> Waiting for API to be healthy..."
until docker compose ps api --format '{{.Status}}' | grep -q healthy; do
  sleep 2
done
echo "    API is healthy."

# 2. Build frontend for production (served by nginx container via tunnel)
echo "==> Building frontend..."
cd "$ROOT_DIR/services/frontend"
pnpm install --frozen-lockfile
pnpm build
cd "$ROOT_DIR"

# Restart web container to pick up fresh build
echo "==> Restarting web container..."
docker compose restart web

# 3. Start Vite dev server in background for local development
echo "==> Starting Vite dev server..."
cd "$ROOT_DIR/services/frontend"
pnpm dev &
DEV_PID=$!
cd "$ROOT_DIR"

echo ""
echo "============================================"
echo "  MatchUp is running!"
echo "  Local dev:  http://localhost:5173"
echo "  Tunnel:     https://matchup.potuzhno.in.ua"
echo "  API:        https://matchup-api.potuzhno.in.ua"
echo "============================================"
echo ""
echo "Press Ctrl+C to stop the dev server."

trap "kill $DEV_PID 2>/dev/null; echo 'Dev server stopped.'" EXIT
wait $DEV_PID
