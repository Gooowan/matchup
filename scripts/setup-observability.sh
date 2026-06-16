#!/usr/bin/env bash
# ============================================================
# MatchUp — Observability Stack Setup Script
# ============================================================
# Usage:
#   ./scripts/setup-observability.sh [DOMAIN] [EMAIL]
#
# Arguments:
#   DOMAIN   (optional) Domain for Grafana HTTPS proxy, e.g. grafana.yourdomain.com
#   EMAIL    (optional) Email for Let's Encrypt cert, e.g. admin@yourdomain.com
#
# Examples:
#   # Minimal setup (no HTTPS, access Grafana at http://SERVER_IP:3001)
#   ./scripts/setup-observability.sh
#
#   # Full setup with HTTPS via certbot
#   ./scripts/setup-observability.sh grafana.myapp.com admin@myapp.com
#
# Environment variables (optional overrides):
#   REPO_DIR          Path to MatchUp repo (default: /opt/matchup)
#   GRAFANA_PASSWORD  Custom Grafana admin password (default: auto-generated)
# ============================================================

set -euo pipefail

DOMAIN="${1:-}"
EMAIL="${2:-admin@example.com}"
REPO_DIR="${REPO_DIR:-/opt/matchup}"
GRAFANA_PASSWORD="${GRAFANA_PASSWORD:-}"

# ── Colours ─────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
info()    { echo -e "${GREEN}[INFO]${NC}  $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC}  $*"; }
error()   { echo -e "${RED}[ERROR]${NC} $*" >&2; exit 1; }

# ── Checks ──────────────────────────────────────────────────
[[ "$(id -u)" -eq 0 ]] || error "Run as root (or with sudo)"
[[ -d "$REPO_DIR" ]] || error "Repo not found at $REPO_DIR — set REPO_DIR env var"

info "MatchUp observability setup starting..."
info "Repo: $REPO_DIR"

# ── 1. Install Docker ────────────────────────────────────────
if ! command -v docker &>/dev/null; then
    info "Installing Docker..."
    curl -fsSL https://get.docker.com | sh
    systemctl enable docker
    systemctl start docker
    info "Docker installed: $(docker --version)"
else
    info "Docker already installed: $(docker --version)"
fi

# ── 2. Install Docker Compose plugin ────────────────────────
if ! docker compose version &>/dev/null 2>&1; then
    info "Installing Docker Compose plugin..."
    apt-get update -qq
    apt-get install -y docker-compose-plugin
fi
info "Docker Compose: $(docker compose version)"

# ── 3. Generate Grafana password ────────────────────────────
if [[ -z "$GRAFANA_PASSWORD" ]]; then
    GRAFANA_PASSWORD="$(openssl rand -base64 24 | tr -d '/+' | head -c 32)"
    info "Generated Grafana admin password"
fi

# Write / update in .env file
ENV_FILE="$REPO_DIR/.env"
if [[ -f "$ENV_FILE" ]]; then
    if grep -q "^GRAFANA_PASSWORD=" "$ENV_FILE"; then
        sed -i "s/^GRAFANA_PASSWORD=.*/GRAFANA_PASSWORD=$GRAFANA_PASSWORD/" "$ENV_FILE"
    else
        echo "GRAFANA_PASSWORD=$GRAFANA_PASSWORD" >> "$ENV_FILE"
    fi
else
    echo "GRAFANA_PASSWORD=$GRAFANA_PASSWORD" > "$ENV_FILE"
fi
info "GRAFANA_PASSWORD written to $ENV_FILE"

# ── 3b. Generate METRICS_TOKEN (required for /metrics + Prometheus scrape) ──
if [[ -z "${METRICS_TOKEN:-}" ]] && [[ -f "$ENV_FILE" ]]; then
    METRICS_TOKEN="$(grep -E '^METRICS_TOKEN=' "$ENV_FILE" 2>/dev/null | head -1 | cut -d= -f2- || true)"
fi
if [[ -z "${METRICS_TOKEN:-}" ]]; then
    METRICS_TOKEN="$(openssl rand -hex 32)"
    info "Generated METRICS_TOKEN"
    if grep -q "^METRICS_TOKEN=" "$ENV_FILE" 2>/dev/null; then
        sed -i "s/^METRICS_TOKEN=.*/METRICS_TOKEN=$METRICS_TOKEN/" "$ENV_FILE"
    else
        echo "METRICS_TOKEN=$METRICS_TOKEN" >> "$ENV_FILE"
    fi
fi
mkdir -p "$REPO_DIR/secrets"
printf '%s' "$METRICS_TOKEN" > "$REPO_DIR/secrets/metrics_token"
chmod 644 "$REPO_DIR/secrets/metrics_token"
info "METRICS_TOKEN written to $ENV_FILE and secrets/metrics_token"

# ── 4. Ensure network exists ─────────────────────────────────
if ! docker network inspect matchup-network &>/dev/null; then
    docker network create \
        --driver bridge \
        --subnet 172.71.0.0/24 \
        matchup-network
    info "Created Docker network: matchup-network"
else
    info "Docker network matchup-network already exists"
fi

# ── 5. Optional: Nginx + HTTPS via Certbot ───────────────────
if [[ -n "$DOMAIN" ]]; then
    info "Setting up Nginx reverse proxy for Grafana at $DOMAIN..."

    if ! command -v nginx &>/dev/null; then
        apt-get update -qq
        apt-get install -y nginx
    fi

    cat > /etc/nginx/sites-available/grafana <<NGINX
server {
    listen 80;
    server_name $DOMAIN;

    location / {
        proxy_pass         http://127.0.0.1:3001;
        proxy_set_header   Host              \$host;
        proxy_set_header   X-Real-IP         \$remote_addr;
        proxy_set_header   X-Forwarded-For   \$proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto \$scheme;
    }
}
NGINX

    ln -sf /etc/nginx/sites-available/grafana /etc/nginx/sites-enabled/grafana
    nginx -t
    systemctl reload nginx
    info "Nginx configured for $DOMAIN"

    if ! command -v certbot &>/dev/null; then
        apt-get install -y certbot python3-certbot-nginx
    fi

    certbot --nginx \
        -d "$DOMAIN" \
        --email "$EMAIL" \
        --agree-tos \
        --non-interactive \
        --redirect
    info "HTTPS certificate issued for $DOMAIN"
else
    warn "No DOMAIN specified — Grafana accessible at http://SERVER_IP:3001"
fi

# ── 6. Create systemd service ────────────────────────────────
info "Creating systemd service: matchup-observability..."

COMPOSE_FLAGS="-f compose.yml"
if [[ -f "$REPO_DIR/compose.prod.yml" ]]; then
    COMPOSE_FLAGS="$COMPOSE_FLAGS -f compose.prod.yml"
fi
COMPOSE_FLAGS="$COMPOSE_FLAGS -f compose.observability.yml"

cat > /etc/systemd/system/matchup-observability.service <<SYSTEMD
[Unit]
Description=MatchUp Observability Stack (Prometheus + Grafana + Loki + Tempo)
After=docker.service network-online.target
Requires=docker.service
Wants=network-online.target

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=$REPO_DIR
EnvironmentFile=-$ENV_FILE
ExecStart=$REPO_DIR/scripts/observability-up.sh --prod
ExecStop=/usr/bin/docker compose $COMPOSE_FLAGS stop prometheus grafana loki tempo promtail
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
SYSTEMD

systemctl daemon-reload
systemctl enable matchup-observability
info "Systemd service created and enabled"

# ── 7. Print summary ─────────────────────────────────────────
echo ""
echo -e "${GREEN}============================================================${NC}"
echo -e "${GREEN} MatchUp Observability Stack — Setup Complete${NC}"
echo -e "${GREEN}============================================================${NC}"
echo ""
if [[ -n "$DOMAIN" ]]; then
    echo -e "  Grafana URL:       ${YELLOW}https://$DOMAIN${NC}"
else
    SERVER_IP="$(hostname -I | awk '{print $1}')"
    echo -e "  Grafana URL:       ${YELLOW}http://$SERVER_IP:3001${NC}"
fi
echo -e "  Grafana user:      ${YELLOW}admin${NC}"
echo -e "  Grafana password:  ${YELLOW}$GRAFANA_PASSWORD${NC}"
echo ""
echo -e "  Start the stack:"
echo -e "    ${YELLOW}systemctl start matchup-observability${NC}"
echo ""
echo -e "  Or manually:"
echo -e "    ${YELLOW}cd $REPO_DIR && ./scripts/observability-up.sh --prod${NC}"
echo ""
echo -e "  Load test (before/after seed):"
echo -e "    ${YELLOW}./scripts/benchmark.sh user@example.com password https://your-api-url 2m${NC}"
echo ""
echo -e "${GREEN}  IMPORTANT: Save the Grafana password shown above!${NC}"
echo ""
