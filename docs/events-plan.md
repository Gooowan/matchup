# Events module — implementation plan

## Current state

There is **no** events module in the codebase as of this writing. Specifically:

- No `modules/events/` directory, schema, or service.
- No `/events` routes in `services/api/main.go`.
- No `/events` page in the SvelteKit frontend.
- The only occurrences of the word "event" in the codebase refer to
  unrelated things (JS DOM events, Sentry breadcrumbs, Prometheus events).

The `goal` field on profiles has a `competition` label visible on the profile
preview screen (`services/frontend/src/routes/(app)/profiles/[userId]/+page.svelte`),
but no actual competition/event entity is persisted anywhere.

## Goal

Add a "competitions/events" feature so dancers can:

- discover upcoming tournaments and social events,
- mark interest / RSVP,
- see who else is going (for matchmaking around a specific event),
- have a club (host) attach events to its club page.

This is also the natural hook for: trainer feed (task 4 in the original
request), photo galleries from past events, and paid promotion.

## Phased delivery

### Phase 1 — Read-only event catalog (1–2 days)

The minimum we need to ship the navigation entry without it being empty.

1. **DB schema** (`modules/events/schema.sql`):

   ```sql
   CREATE TABLE events (
     id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
     slug        text NOT NULL UNIQUE,
     name        text NOT NULL,
     description text,
     kind        text NOT NULL,                -- 'competition' | 'workshop' | 'social' | 'concert'
     starts_at   timestamptz NOT NULL,
     ends_at     timestamptz,
     country     text,
     city        text,
     address     text,
     latitude    double precision,
     longitude   double precision,
     cover_url   text,
     host_club_id uuid REFERENCES clubs(id) ON DELETE SET NULL,
     organizer   text,                         -- free-form when host_club is null
     external_url text,                        -- ticketing / official site
     program     text,                         -- 'standard'|'latina'|'both'|null
     categories  text[] NOT NULL DEFAULT '{}',  -- juvenile1, junior2, etc.
     created_at  timestamptz NOT NULL DEFAULT now(),
     updated_at  timestamptz NOT NULL DEFAULT now()
   );

   CREATE INDEX idx_events_starts_at ON events(starts_at);
   CREATE INDEX idx_events_city ON events(country, city);
   CREATE INDEX idx_events_host_club ON events(host_club_id);
   ```

2. **Module skeleton** mirroring `modules/clubs/`:

   - `modules/events/queries/events.sql` — `ListUpcomingEvents`,
     `GetEventBySlug`, `GetEventsByClub`.
   - `modules/events/service.go` — pure data access; no auth logic yet.
   - `modules/events/controller.go` — public-read routes:
     - `GET /events` (paged + filterable by city/program/category/start window)
     - `GET /events/:slug`
     - `GET /clubs/:slug/events` (events hosted by club, used on club page)
   - Wire into `services/api/main.go` alongside `clubCtrl`.

3. **Seeding**:

   - Extend `cmd/seed-profiles` (or add `cmd/seed-events`) to insert a
     handful of realistic Ukrainian tournaments (e.g. "Kyiv Open 2026",
     "Lviv Spring Cup") with diverse `program`/`categories` so the
     frontend has data to render.

4. **Frontend**:

   - Add `services/frontend/src/lib/api/events.ts` (mirror of
     `services/frontend/src/lib/api/chats.ts`).
   - Add `services/frontend/src/routes/(app)/events/+page.svelte`
     (list/grid of upcoming events, filter pill for city/program).
   - Add `services/frontend/src/routes/(app)/events/[slug]/+page.svelte`
     (event detail: cover, dates, location map, host, "I'm going" button
     stub).
   - Optional: add an "events" entry to `BottomNav` or surface inside
     the existing `Map` screen (events shown as map pins).

### Phase 2 — RSVP + social signal (2–3 days)

Lets users mark intent and lets us power "find a partner for X event".

1. **DB additions**:

   ```sql
   CREATE TABLE event_attendees (
     event_id   uuid NOT NULL REFERENCES events(id) ON DELETE CASCADE,
     user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     status     text NOT NULL,            -- 'going'|'interested'|'cancelled'
     role       text,                     -- 'solo'|'partner_lead'|'partner_follow'
     partner_id uuid REFERENCES users(id) ON DELETE SET NULL,
     created_at timestamptz NOT NULL DEFAULT now(),
     PRIMARY KEY (event_id, user_id)
   );
   CREATE INDEX idx_event_attendees_user ON event_attendees(user_id);
   ```

2. **Endpoints**:

   - `POST   /events/:slug/rsvp` body `{ status, role?, partner_id? }`
   - `DELETE /events/:slug/rsvp`
   - `GET    /events/:slug/attendees?status=going` (paged)
   - `GET    /me/events` (events the caller has RSVP'd to)

3. **Frontend**:

   - "Я іду" / "Цікаво" / "Скасувати" buttons on the event detail page.
   - Grid of attending dancers (re-use `SwipeCard` thumbnail style)
     with a "Swipe partner for this event" CTA which opens a filtered
     feed (`/feed?event=<slug>`) — the feed candidate query gets an
     optional `event_id` filter that ranks attendees first.

4. **Recommendation**:

   - In `modules/recommendation/tier2/provider.go`, add an
     "attending same event" signal so co-attendees score higher.
   - In `modules/feed/recommender.go`, honor `?event_id=` to filter to
     only attendees if set.

### Phase 3 — Club-owned events (2 days)

Lets verified clubs create their own events, mirroring how
`modules/clubs/controller.go` already supports `parse-gmaps` + claim.

1. **Endpoints** (gated behind "is owner of `host_club_id`" check that
   already lives in `clubs.IsClubOwner`):

   - `POST   /clubs/:slug/events`           (create)
   - `PUT    /clubs/:slug/events/:eventSlug` (update)
   - `DELETE /clubs/:slug/events/:eventSlug`
   - `POST   /clubs/:slug/events/:eventSlug/cover` (upload to S3/MinIO via
     the existing `files` module — `events` bucket).

2. **Admin**:

   - Add `GET /admin/events` and `POST /admin/events/:id/verify` so the
     admin panel can spotlight curated events on the homepage
     (re-use `modules/users/admin_controller.go` patterns).

3. **Frontend (Business Panel)**:

   - Add an "Events" tab to `routes/(app)/business/+page.svelte` for
     club owners to CRUD their events; reuse the working-hours date
     picker pattern.

### Phase 4 — Notifications + polish (1–2 days)

1. **Push** (`modules/push`): "Event you RSVP'd to starts tomorrow" job
   in `cmd/cron` (currently commented out in `compose.yml`; uncomment
   when we add the cron service).
2. **Chat hooks**: when two matched users have both RSVP'd to the same
   event, surface a small banner in their chat ("You're both going to
   Kyiv Open 2026!").
3. **Map integration**: render event pins in `modules/map` with the
   existing recommender query.
4. **Analytics**: capture `event_view`, `event_rsvp` via the existing
   PostHog hook (`services/frontend/src/lib/analytics/posthog.ts`).
5. **Localization**: add `services/frontend/src/lib/locale/{uk,en}/events.json`.

## Open product questions before starting

- Do amateurs see events at all, or only competitions matching their
  level? (Suggest: amateurs default to "workshop"/"social", pros
  default to "competition".)
- Ticket purchases — out of scope or external link only? (Suggest:
  external `external_url` for v1.)
- Photo gallery for past events — separate feature or part of event
  detail?

## Storage / deployment note

Event cover images follow the same dual-storage pattern as the rest of
the app (see `.env.example`): MinIO inside Docker for local/dev, AWS S3
in production, swapped only via env vars. No code changes required to
support that.
