# Scalability backlog (deferred from trainer/media optimisation plan)

Items below were scoped for a future iteration. None of these are urgent blockers;
the targeted changes already shipped address the most impactful issues.

## Pagination

- Unify the three pagination schemes (`limit/page`, `page/take`, `limit/offset`) onto a shared `PaginationParams` + `PaginatedResp{data, meta}` envelope.
- Add `cursor`-based pagination to chat message history and the chats inbox for infinite-scroll.

## Image processing

- Server-side thumbnail generation for MinIO uploads (e.g. `sharp` or `vips` via a sidecar or at upload time): produce 200px and 800px variants and serve via `srcset`.
- Consider WebP/AVIF transcode at the proxy for further bandwidth savings.

## Caching

- Redis read-through cache for public club/trainer/profile reads (key: `club:{slug}`, TTL 60 s, invalidated on ManageClub/VerifyClub writes).
- Deduplicate the chat inbox N+1 pattern: fetch all last messages in a single windowed query instead of a LATERAL per chat.

## Club/member search

- Push `gender`, `goal`, `program`, `city`, `age` filters from Go-level filtering into the SQL `WHERE` clause for `ListMembers`, using partial indexes on the common combinations.

## API versioning & DTO stability

- Introduce a stable versioned DTO layer (`/v1/…`) so clients can pin to a contract and the backend can evolve independently.

## Infrastructure

- Read replica for the PostgreSQL pool (heavy read queries → replica, writes → primary).
- CDN in front of MinIO public bucket to offload image bandwidth from the server.
