// ============================================================================
// MatchUp Capstone — Chapter 3: System Design and Architecture
// Drop-in Typst body content (no preamble — paste into the KSE template body).
// Conventions assumed:
//   * Heading numbering is ON in the template, so headings carry no manual numbers.
//   * Cross-references use Typst labels: <sec:...>, <fig:...>, <tbl:...> + @ref.
//   * Figures point at assets in figures/ — placeholders captioned in full so the
//     chapter reads complete; swap in real exports (offered separately).
//   * Citations use IEEE-style @keys; add the matching entries to your .bib/.yaml.
//     Stub bibliography entries are listed at the bottom of this file in a comment.
// ============================================================================

= System Design and Architecture <ch:design>

The MatchUp architecture is a direct, deliberate answer to the structural realities
established in the preceding domain analysis. Competitive and social dance operates as
a fragmented, word-of-mouth ecosystem with no purpose-built infrastructure and a
tightly bounded, geographically clustered community. Two consequences of that
reality shape every decision in this chapter. First, matching is a _network-density_
problem rather than a national-coverage one: value is created the moment a single city
reaches enough active dancers that the next person to open the feed finds a real
candidate, which places liquidity and a never-empty feed above raw scale. Second, the
product is intended to grow along a fixed social chain — dancers pull in parents,
parents and dancers anchor trainers, trainers anchor clubs, and clubs host the events
module that drives long-term retention and supply-side revenue — which means the
system must be cheap to replicate per city and structurally prepared for an
entity model far larger than dancer-to-dancer swiping. The design prioritises
maintainability, read performance, graceful degradation, and a low operational
footprint for a lean team, while keeping a clean extraction path to a larger
distributed system once volume justifies it.

== Architecture Overview and Requirements Alignment <sec:overview>

The architecture is derived from the functional and non-functional requirements
formulated in Chapter 2. Rather than treating design as an independent exercise,
each major structural choice traces back to a specific requirement it was selected
to satisfy.

=== Requirements-Driven Architecture Decisions

*Functional requirements drive:*

- _Two-sided partner discovery (FR2):_ a dedicated recommendation module implements
  a tiered candidate pipeline that evaluates compatibility in both directions and
  never returns an empty feed.
- _Authentication and profile management (FR1):_ a self-contained `users` module
  owns identity, role-based account types (including the first-class parent/junior
  relationship), and nonce-based session invalidation.
- _In-app communication (FR3):_ a `chat` module creates a conversation automatically
  on a mutual match and applies keyword-based safety filtering.
- _Club, trainer, and map discovery (FR4):_ `clubs` and `map` modules model the
  supply side and geocode it for spatial browsing.
- _Trust, safety, and monetisation (FR5, FR6):_ `moderation` and `subscriptions`
  modules isolate block/report/ban logic and entitlement management respectively.

*Non-functional requirements drive:*

- _Maintainability (NFR4):_ a schema-first, type-safe data pipeline carries type
  guarantees from the database to the API contract, and dependencies are wired
  explicitly in a single readable composition root.
- _Reliability (NFR2):_ every third-party integration is swappable and degrades
  safely, and the recommendation pipeline guarantees a non-empty feed by design.
- _Scalability (NFR3):_ a modular monolith preserves clean domain boundaries and a
  per-module extraction path while avoiding premature distribution; city expansion
  is driven entirely by geocoding and configuration, requiring no code changes.
- _Read performance (NFR1):_ the data model is tuned for the read-heavy swipe
  workload through targeted indexing.
- _Observability (NFR6):_ metrics, traces, logs, error reporting, and product
  analytics are treated as first-class cross-cutting concerns rather than
  afterthoughts.

== System Architecture and Major Decisions <sec:major>

=== Modular Monolith Decision

*Decision:* Implement the backend as a _modular monolith_ — a set of self-contained
domain modules that compile into a single deployable binary — rather than either a
conventional layered monolith or a microservices system.

*Justification:* All domain logic lives in self-contained modules — `users`,
`recommendation`, `feed`, `chat`, `clubs`, `map`, `files`, `moderation`,
`subscriptions`, `push`, `otp`, `ratelimit`, `email`, and a shared `core`. Each
module owns its own slice of the database schema and its own generated data-access
code, exactly as an independently deployable service would, yet all modules compile
into one binary. This yields the principal benefit of microservices — strict domain
boundaries and a clear path to extract any module into its own service later —
without the operational tax of running and observing many services, pipelines, and
inter-service transport on day one. For a pre-scale product whose load can be served
by a single vertically scaled instance for a long time, this is the correct trade:
the boundaries that make a future split cheap are present from the outset, but the
cost of the split is deferred until volume warrants it.

*Trade-offs considered:* A pure microservices architecture was rejected as premature
— it would impose distributed-systems complexity (network failure modes, eventual
consistency, deployment orchestration) on a single-developer effort with no
corresponding scale benefit. A conventional layered monolith was rejected in the
opposite direction: it would ship faster initially but blur domain boundaries,
making a later extraction expensive. The modular monolith captures most of the
upside of both while paying the cost of neither.

Each module follows an identical internal shape — a _controller_ (HTTP handlers and
route registration), a _service_ (business logic and transactions), a _repository_
(the generated, type-safe queries), and its own schema and SQL files. Module
dependencies are wired explicitly and visibly in a single composition root
(`services/api/main.go`) rather than concealed inside a dependency-injection
container, so the entire system's composition can be read in one file — a direct
service to maintainability (NFR4).

== System Context and External Interactions <sec:context>

The system context situates MatchUp within its surrounding ecosystem of clients and
external services. The platform is the central hub through which dancers, parents,
trainers, and clubs interact, mediating every external dependency behind its own API
so that no client ever contacts a third party directly.

#figure(
  image("figures/system-context.svg", width: 100%),
  caption: [System context — MatchUp platform and its external integrations.
  Clients (web SPA and the Capacitor iOS app) reach the API through nginx and a
  Cloudflare tunnel; the API mediates all external services: Google Places and
  Nominatim/OpenStreetMap for location, AWS Rekognition for image moderation,
  RevenueCat for subscriptions, Apple Push Notification service for delivery,
  and Resend/Mailgun for transactional email.],
) <fig:context>

*Key external integrations* (detailed in @sec:integrations):

- _Location:_ Google Places API for importing a club from a Google Maps link, and
  Nominatim/OpenStreetMap for address-to-coordinate geocoding.
- _Image moderation:_ AWS Rekognition for optional, fail-open NSFW scanning of
  avatar uploads.
- _Subscriptions:_ RevenueCat, via a secured webhook that maps store product IDs to
  internal plans.
- _Push:_ Apple Push Notification service (APNs).
- _Email:_ Resend or Mailgun, with a mock provider for local development.
- _Object storage:_ MinIO in development and AWS S3 in production, behind one client.

== Module and Process Decomposition <sec:decomposition>

The deployable system comprises two processes that share the same module code but
differ in entrypoint: the _API_ process on port 8000 serves client traffic, and a
separate _cron_ process on port 8001 runs scheduled jobs. Decoupling scheduled work
from the request-serving process prevents background jobs from competing with
user-facing latency and allows the two to be scaled or relocated independently.

#figure(
  image("figures/module-architecture.svg", width: 100%),
  caption: [Module and process decomposition. Self-contained domain modules
  (`users`, `recommendation`, `feed`, `chat`, `clubs`, `map`, `files`,
  `moderation`, `subscriptions`, `push`, `otp`, `ratelimit`, `email`, `core`)
  compile into one binary exposed through two entrypoints — the API (:8000) and
  cron (:8001). Each module owns a controller, service, repository, and schema.],
) <fig:modules>

The API is a REST/JSON interface with a single uniform response envelope — every
response takes the shape `{ data, error, error_code }`. This uniformity is what
allows the frontend to translate any backend error into a localised message
without special-casing individual endpoints, directly supporting localisation
(NFR7) and maintainability (NFR4).

```go
// Uniform response envelope returned by every endpoint.
type Response struct {
    Data      any    `json:"data"`
    Error     string `json:"error"`
    ErrorCode string `json:"error_code"`
}
```

The data flow for a typical request is a fixed, observable chain:

client → nginx → Gin middleware chain (panic recovery → Sentry → request ID →
OpenTelemetry tracing → Prometheus metrics → structured logging → CORS) →
per-route authentication and rate-limit guards → controller → service →
generated query → PostgreSQL.

The frontend is a pure client-rendered single-page application (statically built)
that communicates with the API over cookies — the same model the Capacitor iOS
WebView requires, which is what permits a single frontend bundle to serve both web
and native targets.

== Technology Stack Selection and Justification <sec:stack>

The stack was selected against four explicit criteria, each mapping to a
non-functional requirement: minimise future technical debt and maintenance burden
(NFR4), avoid performance bottlenecks (NFR1), enable rapid vertical scaling (NFR3),
and keep the operational and cost footprint low enough for a lean team. The unifying
principle is _one codebase, one language per tier, and type-safety carried from the
database all the way to the API contract_.

=== Backend: Go 1.25 with Gin

*Decision:* Standardise the backend on Go 1.25 using the Gin HTTP framework.

*Justification:* Go compiles to a single statically linked binary with no runtime
dependencies, which makes deployment, containerisation, and vertical scaling
straightforward and keeps the production image tiny. Its concurrency model and
predictable performance suit an API-heavy, read-dominated workload, and Gin adds a
mature middleware ecosystem without imposing a heavyweight framework. The result
directly serves NFR1 and NFR3 while keeping the maintenance surface small.

=== Data Access: PostgreSQL 18 with pgx and sqlc

*Decision:* Use PostgreSQL 18 accessed through `pgx/v5`, with `sqlc` generating
fully type-safe Go from hand-written SQL — deliberately rejecting an ORM.

*Justification:* Writing raw SQL and generating type-safe accessors from it removes
the two failure modes most associated with ORMs in a read-heavy product: opaque
query generation and accidental N+1 access patterns. Type-safety extends from the
database schema through the repository layer to the API contract, so an
incompatible schema change is caught at compile time rather than in production —
the strongest possible service to maintainability (NFR4). PostgreSQL's mature
indexing (notably GIN indexes for array-valued fields and partial indexes for hot
predicates) is also what makes the recommendation workload performant, as detailed
in @sec:data.

*Trade-offs considered:* An ORM would have accelerated early CRUD development but at
the cost of query transparency and predictable performance — an unacceptable trade
for a system whose core path is a latency-sensitive matching query.

=== Frontend and Mobile: SvelteKit, Capacitor

*Decision:* Build the client as a SvelteKit 2 / Svelte 5 application (using the
runes reactivity model) with Vite 8, Tailwind CSS v4, and `shadcn-svelte`
components on `bits-ui`, with Leaflet for maps; wrap the same web build into a
native iOS application with Capacitor, using native Apple Push Notifications.

*Justification:* A single web bundle that ships to both the browser and — through
Capacitor — the iOS App Store removes an entire class of duplication: there is one
UI codebase, one design system, and one set of flows to maintain across two
delivery targets. Svelte's compile-time reactivity produces a lean client suited to
the swipe-centric interaction model, and Tailwind with `shadcn-svelte` gives a
consistent, rapidly iterable component layer. This consolidation is a deliberate
maintainability decision (NFR4) and the mechanism by which a lean team supports two
platforms at once.

=== Caching, Ephemeral State, and Object Storage

*Decision:* Use Redis for rate-limiting state, Valkey (the open Redis fork) for
one-time-password storage, and a single `minio-go` client targeting MinIO in
development and AWS S3 in production, selected purely by environment variable.

*Justification:* Separating ephemeral, fast-expiring state (rate-limit counters,
OTP codes) from the durable relational store keeps both stores doing what they are
best at, and binding object storage behind one client with an environment-driven
endpoint makes the development and production storage backends interchangeable
without code changes — a small but characteristic example of the swappability
principle that recurs throughout the integration layer.

@tbl:stack summarises the stack with its justification and the principal trade-off
accepted for each choice.

#figure(
  table(
    columns: (auto, auto, 1fr, 1fr),
    align: (left, left, left, left),
    table.header([*Layer*], [*Technology*], [*Justification*], [*Trade-off accepted*]),
    [Backend], [Go 1.25 + Gin], [Single static binary, predictable performance, simple deploy], [Smaller library ecosystem than older runtimes],
    [Data access], [PostgreSQL 18, pgx/v5, sqlc], [Compile-time type-safety DB→API; no ORM N+1], [Hand-written SQL; more upfront query work],
    [Frontend / mobile], [SvelteKit 2, Svelte 5, Capacitor], [One bundle to web and iOS; lean compiled client], [WebView constraints on native capability],
    [Ephemeral state], [Redis, Valkey], [Right store for rate limits and OTP TTLs], [Additional moving parts in the stack],
    [Object storage], [MinIO (dev) / S3 (prod)], [Env-swappable, identical client], [Eventual-consistency semantics],
    [Maps], [Leaflet + Nominatim / Google Places], [Open geocoding with a budgeted import path], [Two location providers to maintain],
  ),
  caption: [Technology stack: justification and trade-off analysis.],
) <tbl:stack>

== Component Deep Dive: The Recommendation Engine <sec:reco>

The recommendation engine is the technical heart of the system, and its central
design property — that the feed is _never empty_ — is a direct response to the
low-density market established in Chapter 2. In a city that has not yet reached
liquidity, an empty feed is fatal: it signals to a new user that the product does
not work for them before the network has had a chance to form. The engine therefore
implements a three-tier candidate pipeline that degrades gracefully in candidate
quality so that there is always something to evaluate.

#figure(
  table(
    columns: (auto, 1fr, auto),
    align: (left, left, left),
    table.header([*Tier*], [*Candidate definition*], [*Ordering*]),
    [3 — Mutual], [Candidates who pass the user's filters _and_ whose stated
      preferences the user satisfies (genuine two-sided compatibility)], [Distance],
    [2 — Filter], [Everyone satisfying all of the user's preferences: gender, age
      and height ranges, skill goal, program (Standard/Latin), competitive category,
      city, relocation and co-financing willingness], [Distance],
    [1 — Proximity], [Safety net of gender + distance only, used to fill the deck
      when higher tiers are exhausted], [Distance],
  ),
  caption: [The three-tier recommendation pipeline, walked in quality order.],
) <tbl:tiers>

The recommender walks the tiers in quality order — 3, then 2, then 1 —
deduplicating by user until it has assembled a full page. Distance is computed in
SQL through a haversine expression anchored on the user's primary club location,
falling back to a city centroid when no club has been set. Users who have already
been swiped, and users who are blocked, are excluded at the query level rather than
filtered in application code, which keeps the candidate set correct and the query
the single source of truth.

A notable design economy underpins the whole pipeline: the filters are expressed as
_nullable_ query parameters, where a `NULL` value means "do not filter on this
dimension." Because of this, a single parameterised SQL query powers all three
tiers — the tiers differ only in which arguments they bind and which they leave
null. This avoids three near-duplicate queries and the maintenance hazard that
duplication would create, and it is a concrete expression of the maintainability
principle (NFR4) at the level of the most performance-critical code path.

The engine is also instrumented for its own future. Every time a user likes a
profile, a feature snapshot of that match is written asynchronously to a
`recommendation_likes_log` table. Ranking today is rule-based SQL, which is correct,
transparent, and fast at current scale; but the instrumentation means the training
data for a learned ranking model — collaborative filtering and a preference model —
is already being collected. When volume justifies it, the system can move from rules
to a learned model without retrofitting any instrumentation. The architecture
anticipates the transition rather than precluding it.

#figure(
  image("figures/recommendation-pipeline.svg", width: 90%),
  caption: [Recommendation candidate pipeline. A single nullable-parameter query is
  invoked in quality order (Tier 3 → 2 → 1), deduplicating by user and excluding
  already-swiped and blocked users at the query level, until a full page is
  assembled; likes are logged asynchronously for future learned ranking.],
) <fig:reco>

== Data Layer and Schema Design <sec:data>

The data model is _schema-first_. Each module declares its own `schema.sql`, and a
custom generator (`cmd/schema-gen`) stitches the per-module schemas together, in
dependency order, into one combined schema that `sqlc` reads. This is what allows a
`profile` to reference a `club` by foreign key across a module boundary while each
module's schema remains self-contained — the data-layer mechanism that makes the
modular-monolith boundaries (@sec:major) real rather than nominal.

The core entities are: `users` (identity, role, and the nonce that powers session
invalidation); `profiles` and `user_preferences` (the two-sided matching data);
`matches` (swipe history and mutual matches); `chats`, `messages`, and `chat_reads`;
`clubs` with `club_members`, `club_trainers`, and `trainer_students`;
`user_locations` for the map; `blocks` and `reports` for moderation; `media` for
uploads; `subscriptions` and `user_subscriptions`; and `user_push_tokens`.

#figure(
  image("figures/data-model.svg", width: 100%),
  caption: [Core data model. The `users`, `profiles`, and `user_preferences`
  entities form the matching core; `matches`, `chats`/`messages`/`chat_reads`,
  the `clubs` cluster, `user_locations`, moderation tables, media, subscriptions,
  and push tokens extend it along the supply-side and safety axes.],
) <fig:data>

Because a swipe application is read-dominated — the expensive path is assembling a
feed, not writing a like — the schema is tuned for read performance (NFR1). The
`profiles` table is deliberately over-indexed for the matching workload: GIN indexes
on the dance-styles and category arrays, btree indexes on every filterable field,
and partial indexes for the hot predicates such as "visible profiles only." Indexing
is treated as part of query design rather than a later optimisation — a lesson
recorded explicitly in the sprint retrospectives of Chapter 4.

Migrations follow a two-path strategy. A fresh database receives the full combined
schema in a single application; an existing database receives only the pending
incremental migrations from `build/migrations/`, tracked in a `schema_migrations`
table. Roughly fourteen migrations exist to date (account types, club tables,
preference columns, the likes log, push tokens, primary-club linkage, club chats,
and others). The migration logic runs as a one-shot migration container that the API
process waits on before booting, guaranteeing that the database is at the correct
version before any traffic is served.

== Integration Architecture <sec:integrations>

Every external dependency follows one pattern: integrate the real provider, but keep
it swappable and make it degrade safely, so that no single third party can take the
whole product down. This is the data-layer-level realisation of the reliability
requirement (NFR2).

- _Location:_ the Google Places API imports a club directly from a Google Maps link,
  governed by a configurable daily spend cap so the integration cannot exceed
  budget; Nominatim/OpenStreetMap performs address-to-coordinate geocoding; Leaflet
  renders the map client-side.
- _Image moderation:_ AWS Rekognition scans avatar uploads for NSFW content. The
  check is _optional and fail-open_ — a service hiccup does not block an upload —
  and can be disabled entirely by environment variable.
- _Subscriptions:_ RevenueCat is integrated through a secured webhook that maps
  store product identifiers to internal plans and updates entitlements, which is the
  standard, correct mechanism for App Store in-app purchases.
- _Push notifications:_ Apple Push Notification service via the `apns2` library; the
  iOS app registers its device token with the backend, and a mutual match fires a
  native push.
- _Email:_ transactional email through Resend or Mailgun, pluggable, with a mock
  provider for local development so real mail can never be sent accidentally.

#figure(
  table(
    columns: (auto, auto, 1fr),
    align: (left, left, left),
    table.header([*Concern*], [*Provider*], [*Degradation / swappability*]),
    [Club import], [Google Places], [Daily spend cap; import is optional],
    [Geocoding], [Nominatim / OSM], [Open provider; replaceable],
    [Avatar moderation], [AWS Rekognition], [Fail-open; toggled by env var],
    [Subscriptions], [RevenueCat], [Webhook-driven entitlement mapping],
    [Push], [APNs (`apns2`)], [Per-device token; non-blocking],
    [Email], [Resend / Mailgun], [Pluggable; mock provider in dev],
    [Object storage], [MinIO / S3], [Identical client, env-selected endpoint],
  ),
  caption: [Integration fallback and swappability matrix.],
) <tbl:integrations>

== Deployment and Infrastructure Architecture <sec:deploy>

The deployment posture is reported here honestly as current state versus planned
state, because the distinction is itself an architectural point: the system is
already built for the target topology, so reaching it is an infrastructure exercise
rather than a rewrite.

*Current state.* The application is live and reachable on the public internet. The
entire stack runs through Docker Compose — PostgreSQL, Valkey, MinIO, the migration
job, the Go API, and an nginx container serving the static frontend — on a private
bridge network, exposed via a Cloudflare tunnel at `matchup.potuzhno.in.ua` (app)
and `matchup-api.potuzhno.in.ua` (API). This provides a real, demonstrable,
internet-accessible deployment that the iOS application and external testers use
today.

*Planned state.* The next infrastructure milestone is a clean separation into
dedicated dev, staging, and production environments and a continuous-deployment
pipeline that builds and rolls out the Docker images automatically. The
prerequisites for that promotion already exist: multi-stage Dockerfiles that produce
tiny static binaries, a migration container that the API correctly waits on,
fully externalised configuration via environment variables, and S3-swappable
storage. Promoting from "tunnel-exposed Compose" to "managed production with
continuous deployment" therefore touches infrastructure, not application code — a
property that follows directly from the externalised-configuration and
single-binary decisions made earlier in this chapter.

== Cross-Cutting Concerns <sec:cross>

=== Security

Authentication uses JWT sessions whose validity is bound to a per-user nonce stored
on the `users` record; rotating the nonce invalidates every outstanding session,
which is the mechanism behind "log out everywhere" and a clean response to
credential compromise. A Redis-backed rate limiter guards sensitive routes, CORS is
enforced in the middleware chain, and the `moderation` module provides
block, report, and ban primitives backed by dedicated tables. Avatar uploads are
screened by the optional Rekognition integration. Collectively these address the
security requirement (NFR5) without introducing a separate identity provider, which
keeps the trust boundary inside the system the team controls.

=== Observability

Observability is treated as a first-class concern (NFR6) and is, by design, more
comprehensive than is typical at this stage. Prometheus scrapes both the API and
cron processes, exposing custom counters for swipes, matches, and per-tier
recommendation hits alongside database connection-pool gauges. Grafana provides a
provisioned "API overview" dashboard covering request rate, 5xx rate, P50/P95/P99
latency, pool health, and swipe/match volume. OpenTelemetry exports traces to
Tempo with spans down to the database; structured JSON logs emitted through Go's
`slog` are shipped to Loki by Promtail; Sentry captures errors and crashes on both
the backend and the frontend (including native iOS crashes via Sentry Capacitor);
and PostHog tracks the product funnel — swipes, matches, authentication events, and
page views. Crucially, the three pillars are cross-linked in Grafana: a log line
links to its trace, a trace links back to its logs and to a service map, so a single
request can be followed across logs, metrics, and traces when something breaks.

=== Localisation

The client defaults to Ukrainian with English available beneath it, implemented via
`sveltekit-i18n`. Combined with the uniform response envelope (@sec:decomposition),
this allows any backend error to be surfaced to the user as a localised message,
satisfying NFR7 without scattering locale logic across endpoints.

== Chapter Summary <sec:summary>

This chapter has presented an architecture engineered around the two defining
properties of the dance-partner market: density over coverage, and a fixed social
expansion chain. The modular-monolith backend in Go provides strict domain
boundaries and a deferred, low-cost path to distribution; the schema-first,
type-safe data pipeline carries correctness from the database to the API contract;
and the three-tier recommendation engine guarantees a never-empty feed while quietly
collecting the data for a future learned ranking model. A consolidated SvelteKit and
Capacitor client serves web and iOS from one bundle; every external integration is
real but swappable and fail-open; and observability spans metrics, traces, logs,
errors, and product analytics. The deployment is live today through a containerised,
tunnel-exposed topology that is already structured for promotion to managed
production. Each decision in this chapter has been justified against an explicit
requirement and its trade-off stated plainly; the following chapter documents how
these designs were realised across three development sprints, including the points
at which the implementation revised the design.

// ============================================================================
// BIBLIOGRAPHY STUB — verify and format these against the KSE template's bib file.
// These are real, verifiable project/concept sources; finalise author/date/access
// fields per IEEE style in your template. Do NOT ship unverified entries.
//
// @go            The Go Programming Language.            https://go.dev
// @gin           Gin Web Framework.                      https://gin-gonic.com
// @postgresql    PostgreSQL 18 Documentation.            https://www.postgresql.org/docs/
// @pgx           pgx — PostgreSQL Driver and Toolkit (Go). https://github.com/jackc/pgx
// @sqlc          sqlc — Compile SQL to type-safe code.   https://sqlc.dev
// @redis         Redis.                                  https://redis.io
// @valkey        Valkey.                                 https://valkey.io
// @sveltekit     SvelteKit.                              https://svelte.dev/docs/kit
// @capacitor     Capacitor.                              https://capacitorjs.com
// @leaflet       Leaflet — an open-source JS library for maps. https://leafletjs.com
// @revenuecat    RevenueCat Documentation.               https://www.revenuecat.com/docs
// @rekognition   Amazon Rekognition.                     https://aws.amazon.com/rekognition/
// @opentelemetry OpenTelemetry.                          https://opentelemetry.io
// @c4model       S. Brown, "The C4 model for visualising software architecture." https://c4model.com
// ============================================================================
