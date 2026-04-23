# MatchUp — Observability Guide

Complete guide for logging, metrics, tracing, error tracking, and analytics.

---

## Stack Overview

| Concern | Tool | Where |
|---|---|---|
| Structured logging | Go `log/slog` (JSON) | API + Cron services |
| Metrics | Prometheus + Grafana | `/metrics` on port 8000 / 8001 |
| Distributed tracing | OpenTelemetry → Grafana Tempo | OTLP HTTP on `tempo:4318` |
| Log aggregation | Grafana Loki + Promtail | Reads Docker container logs |
| Error tracking (BE) | Sentry Go SDK | `SENTRY_DSN` env var |
| Error tracking (FE/iOS) | `@sentry/capacitor` | `VITE_SENTRY_DSN` env var |
| Product analytics | PostHog | `VITE_POSTHOG_KEY` env var |

---

## Quick Start (Local Dev)

```bash
# Start all services + observability stack
docker compose -f compose.yml -f compose.observability.yml up -d

# Check all services are healthy
docker compose -f compose.yml -f compose.observability.yml ps

# Open Grafana
open http://localhost:3001
# Login: admin / matchup_dev  (or GRAFANA_PASSWORD from .env)

# View raw Prometheus metrics from the API
curl http://localhost:8000/metrics
```

---

## Environment Variables

All observability env vars are **optional** — the app runs fine without them. Add to `.env` to enable.

```bash
# Backend
SENTRY_DSN=https://...@sentry.io/...   # Sentry project DSN
OTEL_ENDPOINT=tempo:4318               # OTLP HTTP endpoint (auto-set in compose overlay)
GRAFANA_PASSWORD=changeme              # Grafana admin password (default: matchup_dev)

# Frontend (must be prefixed VITE_ to be exposed to the browser)
VITE_SENTRY_DSN=https://...@sentry.io/...
VITE_POSTHOG_KEY=phc_...
VITE_POSTHOG_HOST=https://app.posthog.com   # Or your self-hosted instance
VITE_APP_VERSION=1.0.0                       # Used in Sentry release tagging

# CI only (Sentry source map upload)
SENTRY_AUTH_TOKEN=sntrys_...
SENTRY_ORG=your-org-slug
SENTRY_PROJECT=matchup-frontend
```

---

## Viewing Logs

### In Terminal
```bash
# Follow API service logs (structured JSON)
docker compose logs -f api

# Filter for errors only
docker compose logs api | jq 'select(.level == "ERROR")'

# Find all logs for a specific request
docker compose logs api | jq 'select(.request_id == "abc-123")'
```

### In Grafana → Loki

1. Open Grafana → **Explore** → select **Loki** datasource
2. Filter by service: `{service="api"}`
3. Filter by log level: `{service="api"} | json | level="ERROR"`
4. Find a request: `{service="api"} | json | request_id="<uuid>"`
5. Find all logs for a trace: `{service="api"} | json | trace_id="<trace-id>"`

**Log-to-Trace correlation**: If a log line has a `trace_id` field, click the link icon next to it to jump directly to the Tempo trace view.

---

## Viewing Traces

### From a Log Line
Click the `trace_id` derived field link in any Loki log line → opens the Tempo trace.

### In Grafana → Tempo
1. Open Grafana → **Explore** → select **Tempo** datasource
2. Paste a `trace_id` in the search box
3. See the full trace: HTTP handler → DB spans → latency breakdown

### Trace Coverage
OTel spans are added to:
- Every HTTP request (via `otelgin` middleware — automatic)
- Feed generation (`GetFeed`) and swipe processing (`Swipe`)
- Recommendation profile fetch + scoring
- Chat message send + fetch

The trace ID is injected into every log entry, enabling seamless log↔trace correlation.

---

## Viewing Metrics

### Grafana Dashboards

**MatchUp → API Overview** dashboard includes:
- Requests/sec and error rate (5xx)
- P50 / P95 / P99 request latency
- DB connection pool (acquired / idle / total)
- Swipes per minute (LIKE vs PASS)
- Matches per minute
- Requests broken down by route
- Cron job success and failure counts

Open: Grafana → **Dashboards** → **MatchUp** → **API Overview**

### Adding a New Metric

1. Define it in [modules/core/metrics/prometheus.go](../modules/core/metrics/prometheus.go):

```go
var MyNewCounter = promauto.NewCounterVec(
    prometheus.CounterOpts{
        Name: "matchup_my_event_total",
        Help: "Description of what this counts",
    },
    []string{"label_one"},
)
```

2. Increment it at the relevant call site:

```go
import "github.com/Gooowan/matchup/modules/core/metrics"

metrics.MyNewCounter.WithLabelValues("value").Inc()
```

3. Query in Grafana: `rate(matchup_my_event_total[5m])`

**Label cardinality warning**: Never use high-cardinality values (user IDs, UUIDs) as Prometheus labels. Use fixed categories only.

---

## Adding a Trace Span

For critical code paths (not every function — only hot paths or operations with significant latency):

```go
import "github.com/Gooowan/matchup/modules/core/tracing"

func (s *MyService) DoSomething(ctx context.Context) error {
    ctx, span := tracing.StartDBSpan(ctx, "DoSomething", "my_table")
    defer span.End()

    // ... database operations ...
}
```

For non-DB spans, use the OTel API directly:

```go
import "go.opentelemetry.io/otel"

tracer := otel.Tracer("matchup/mymodule")
ctx, span := tracer.Start(ctx, "my-operation")
defer span.End()
```

---

## Using the Logger

Every HTTP request context has a child logger pre-loaded with `request_id`, `method`, `path`, `client_ip`, and `trace_id`. Access it anywhere that receives a `context.Context`:

```go
import "github.com/Gooowan/matchup/modules/core/logging"

func (s *MyService) DoSomething(ctx context.Context) error {
    log := logging.FromContext(ctx)

    log.Info("doing something", "key", "value")

    if err != nil {
        log.Error("something failed", "error", err, "extra_field", someValue)
        return err
    }
    return nil
}
```

Log levels:
- `Debug` — verbose detail, only visible in dev (`GIN_MODE != release`)
- `Info` — normal operation events
- `Warn` — recoverable problems (e.g., missing optional data, using fallback)
- `Error` — unexpected failures that affect the response

---

## Frontend Error Tracking (Sentry)

### Setup
1. Create a project at [sentry.io](https://sentry.io) (or self-host)
2. Set `VITE_SENTRY_DSN` in your `.env` file
3. For iOS source maps in CI: set `SENTRY_AUTH_TOKEN`, `SENTRY_ORG`, `SENTRY_PROJECT`

### iOS / Capacitor Notes
- The app uses `@sentry/capacitor` which wraps `@sentry/sveltekit`
- Native iOS crashes are captured by the Capacitor layer
- SvelteKit route errors are captured by `hooks.client.ts`
- Source maps match `capacitor://localhost/...` scheme URLs

### Identifying Users
After login, call `setSentryUser(userId)` to correlate errors to users. **Only the opaque user ID is sent — no email or name**.

---

## Product Analytics (PostHog)

### Setup
1. Create a project at [posthog.com](https://posthog.com) or self-host
2. Set `VITE_POSTHOG_KEY` (and optionally `VITE_POSTHOG_HOST`) in `.env`

### Events Tracked Automatically
- `$pageview` — on every SvelteKit page navigation
- `$pageleave` — when user leaves a page

### Custom Events to Add
The helpers in [services/frontend/src/lib/analytics/posthog.ts](../services/frontend/src/lib/analytics/posthog.ts):

```typescript
import { captureSwipe, captureMatch, captureAuthEvent, identifyUser } from '$lib/analytics/posthog';

// After a swipe
captureSwipe('LIKE', 'tier1');

// After a mutual match
captureMatch();

// After login
captureAuthEvent('login');
identifyUser(user.id);  // ties future events to this user
```

### iOS / Capacitor Notes
PostHog is initialised with `persistence: 'localStorage'` (not cookies). This is required for stable behaviour in Capacitor's WKWebView on iOS, where cookie persistence is unreliable. The PostHog `api_host` must be reachable from the device in production.

---

## Production Server Setup

```bash
# Clone / pull repo to server
git clone https://github.com/yourorg/matchup /opt/matchup

# Set required env vars in /opt/matchup/.env

# Run setup script (with optional Grafana HTTPS domain)
sudo ./scripts/setup-observability.sh grafana.yourdomain.com admin@yourdomain.com

# Start the full stack
sudo systemctl start matchup-observability

# Check status
sudo systemctl status matchup-observability
docker compose -f compose.yml -f compose.observability.yml ps
```

### What the setup script does
1. Installs Docker + Docker Compose if absent
2. Generates a random Grafana admin password (written to `.env`)
3. Creates the `matchup-network` Docker network if needed
4. Optionally: configures Nginx reverse proxy + HTTPS via Let's Encrypt certbot
5. Creates a systemd service that auto-starts the full stack on boot

### Accessing Grafana in Production
- With domain: `https://grafana.yourdomain.com`
- Without domain: `http://SERVER_IP:3001`

**Security note**: Port 3001 should be firewalled from the public internet. Use the Nginx HTTPS proxy or an SSH tunnel for remote access.

---

## Correlation Recipe: Finding a Production Bug

1. User reports error at ~14:32 UTC
2. In Grafana → Loki: filter `{service="api"} | json | level="ERROR"` around that time
3. Find the log line → note `request_id` and `trace_id`
4. Click `trace_id` link → Tempo shows the full request trace with DB span timings
5. Filter `{service="api"} | json | request_id="<id>"` to see ALL logs from that request (auth, handler, service)
6. Check Sentry (if configured) for the corresponding error event with full stack trace
