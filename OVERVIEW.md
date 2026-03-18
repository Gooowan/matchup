# MatchUp — Project Overview

## What It Is
A **dancer dating/matching app** — think Tinder but specifically for finding dance partners. Users match based on dance styles, roles (leader/follower), skill levels, and physical proximity.

---

## Architecture

**Modular Monolith** with two services:

```
services/
├── api/        → REST API (Gin, port 8000)
└── cron/       → Scheduled jobs (port 8001, health only)

modules/
├── core/           → DB connection, shared types
├── users/          → Auth + user management
├── recommendation/ → Profiles + preferences
├── feed/           → Swiping + matching engine
├── chat/           → Direct messaging
├── files/          → Avatar + media uploads
├── map/            → Geolocation
├── moderation/     → Block/report
├── subscriptions/  → Premium plans (NEW)
├── email/          → Email abstraction
├── ratelimit/      → Redis rate limiting
└── otp/            → OTP stub (unused)
```

**Stack**: Go 1.25 · Gin · PostgreSQL 18 · Valkey (Redis) · MinIO · sqlc (code gen) · JWT · Docker

---

## Database Tables

| Table | Module | Purpose |
|---|---|---|
| `users` | users | Auth credentials, roles, tokens |
| `profiles` | recommendation | Dance styles[], lat/lng, JSONB metadata |
| `user_preferences` | recommendation | Matching filters (JSONB) |
| `matches` | feed | Swipe records (LIKE/PASS) |
| `chats` | chat | Mutual match conversations |
| `messages` | chat | Chat messages |
| `user_locations` | map | Current lat/lng per user |
| `media` | files | File records (avatars, marketing) |
| `blocks` | moderation | Block relationships |
| `reports` | moderation | User reports |
| `subscriptions` | subscriptions | Plan definitions (duration, price) |
| `user_subscriptions` | subscriptions | User plan assignments (active/finished) |

Views: `user_subscriptions_expiring_1d`, `user_subscriptions_expiring_1w` (used by cron)

---

## Key Business Logic Flows

### Matching
1. User requests feed → fetch visible profiles within 100km
2. Exclude: already-swiped + blocked users
3. Apply preference filters in-memory (role, styles, level, height, age, gender)
4. Return up to 20 candidates — falls back to random profiles if empty
5. User swipes LIKE → if mutual → auto-create chat

### Authentication
- `POST /auth/register` → bcrypt password + email verification token
- `POST /auth/email-verify` → clear token
- `POST /auth/login` → JWT (8hr expiry), returned in header or cookie
- JWT middleware injects user into Gin context; role check for admin routes

### Subscriptions
1. Admin creates plan (name, duration_days, price_cents)
2. Admin assigns plan to user → `expired_at = now + duration_days`
3. Cron every 5min → marks expired records as `finished`
4. Cron hourly → logs expiring-soon subscriptions (email TODO)

---

## API Surface

| Group | Auth | Purpose |
|---|---|---|
| `/auth/*` | Public | Register, login, verify, password reset |
| `/me/*` | User | Profile, preferences, media |
| `/matchup/*` | User | Feed, swipe, hide |
| `/chats/*` | User | List chats, messages, send |
| `/map/*` | User | Location CRUD, nearby search |
| `/users/:id/*` | User | Block, report |
| `/subscriptions/plans` | User | View active plans |
| `/subscriptions/my` | User | Own subscriptions |
| `/admin/*` | Admin | Users, stats, marketing, subscription plans |
| `/media/*` | User | Public marketing materials |

---

## Module Details

### users
- Registration with bcrypt + email verification flow
- JWT login (8hr expiry), auth nonce for invalidation
- Password reset via secure token
- Role-based access: `USER`, `ADMIN`
- Admin: search users, update metadata, system stats

### recommendation
- Profile: dance_styles[], lat/lng, JSONB (role, level, height, bio, birth_date, gender, city, media_urls)
- Preferences: JSONB (preferred_role, styles, level range, height range, age range, gender, max_distance_km)
- Other users can view profiles via `/profiles/:userId/preview`

### feed
- **NearestCandidatesProvider**: sorts by distance, applies preference filters
- **RandomFallbackProvider**: fallback when primary returns empty
- Swipe records: LIKE or PASS
- Mutual LIKE → creates chat automatically

### chat
- Chats created only on mutual match
- Blocked users cannot send messages
- Message history: newest-first, time-cursor pagination

### files
- Avatar upload: 2MB max, JPG/PNG/WebP, async old-file deletion
- Marketing materials: 50MB max, admin-only, visibility toggle
- Storage: MinIO (S3-compatible), presigned URLs (15min expiry)

### map
- Upsert lat/lng per user
- Find N closest users or users within X km radius
- Distance via PostGIS

### moderation
- Block/unblock users (affects feed visibility + messaging)
- Report users with category + comment

### subscriptions
- Admin manages plans (name, duration_days, price_cents, is_active)
- Admin assigns plans to users
- User views active plans and own subscription history
- Cron handles expiration and expiry-soon logging

---

## Implementation Status

| Feature | Status |
|---|---|
| JWT auth + email verification | Done |
| Profile + preference management | Done |
| Feed generation + swiping | Done |
| Mutual match → chat creation | Done |
| Chat with cursor pagination | Done |
| Geolocation + nearby search | Done |
| Avatar + marketing file uploads | Done |
| Block / report | Done |
| Rate limiting | Done |
| Admin user management | Done |
| Subscription plan CRUD | Done |
| Subscription assignment + expiry cron | Done |
| Subscription expiry notifications | Stub (TODO) |
| Payment processing | Not implemented |
| Real-time chat (WebSocket) | Not implemented |
| OTP service | Stub (unused) |
| Push notifications | Not implemented |

---

## Key Files

| File | Purpose |
|---|---|
| [services/api/main.go](services/api/main.go) | API service entry, route registration |
| [services/cron/main.go](services/cron/main.go) | Cron service entry, job definitions |
| [build/combined_schema.sql](build/combined_schema.sql) | Full database schema |
| [modules/users/auth/service.go](modules/users/auth/service.go) | Auth logic (JWT, bcrypt, email) |
| [modules/users/auth/middleware.go](modules/users/auth/middleware.go) | JWT middleware |
| [modules/feed/recommender.go](modules/feed/recommender.go) | Feed recommendation engine |
| [modules/subscriptions/service.go](modules/subscriptions/service.go) | Subscriptions service |
| [modules/subscriptions/controller.go](modules/subscriptions/controller.go) | Subscriptions API |
| [compose.yml](compose.yml) | Docker Compose (Postgres, Valkey, MinIO) |
| [config.yaml](config.yaml) | Module configuration |
