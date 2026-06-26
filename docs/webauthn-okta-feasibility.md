# WebAuthn / Okta Feasibility Assessment

## Context

MatchUp currently uses email+password + Google OAuth (via `user_identities` table). Both WebAuthn and Okta would add onto that foundation using the same `user_identities` table and existing JWT cookie issuance.

---

## 1. WebAuthn (Passkeys)

### What it is

WebAuthn (Web Authentication API) lets users authenticate with biometrics or a hardware key instead of a password. "Passkeys" are the sync-friendly, platform-managed version (iCloud Keychain, Google Password Manager).

### Required changes

| Layer | Work |
|-------|------|
| **DB** | `webauthn_credentials(id, user_id, credential_id BYTEA UNIQUE, public_key BYTEA, sign_count INT, created_at)` table |
| **Backend** | Add `github.com/go-webauthn/webauthn` package; implement 4 routes: `POST /auth/webauthn/register/begin`, `/finish`, `POST /auth/webauthn/login/begin`, `/finish` |
| **Session** | Reuse existing `CreateJwtToken` + `auth_token` cookie after successful assertion |
| **Frontend web** | `@simplewebauthn/browser` package; call begin → native browser prompt → call finish |
| **Frontend native (Capacitor)** | iOS: Associated Domains (`webcredentials:yourapp.com`) + native Passkeys API via `Capacitor.Plugins.WebAuthn` or a community plugin; Android: Digital Asset Links + native credentials manager |

### Capacitor caveat

Passkeys on iOS/Android require:
- **iOS:** `Associated Domains` entitlement with the backend domain in `apple-app-site-association` (AASA). App binary must be signed with the entitlement.
- **Android:** `assetlinks.json` served from the backend domain; app SHA-256 fingerprint registered.

These are non-trivial setup steps that go beyond code changes.

### Effort estimate

| Phase | Effort |
|-------|--------|
| Backend (4 routes + credential table) | ~3-4 days |
| Frontend web | ~1-2 days |
| Capacitor native (iOS + Android) | ~3-5 days (domained credential setup, testing across devices) |
| Total | **~1-2 weeks** |

### Recommendation

Best added as an optional second factor or "passwordless" upgrade path after Google login is stable. The `user_identities` table already exists — WebAuthn credentials would be a separate `webauthn_credentials` table since they hold cryptographic material (public key bytes, sign counter), not just provider+subject pairs.

---

## 2. Okta (OIDC)

### What it is

Okta is an enterprise identity provider that issues OIDC (OpenID Connect) tokens. Used primarily for B2B / enterprise tenants who want SSO.

### Required changes

| Layer | Work |
|-------|------|
| **DB** | Reuse existing `user_identities(provider='okta', provider_subject=sub)` — no new table needed |
| **Backend** | Add `github.com/coreos/go-oidc/v3` package; 2 routes: `GET /auth/okta` (redirect to Okta authorization URL) + `GET /auth/okta/callback` (exchange code, verify ID token, find/create user) |
| **Config** | `OKTA_DOMAIN`, `OKTA_CLIENT_ID`, `OKTA_CLIENT_SECRET` — one set per Okta tenant |
| **Frontend** | "Sign in with Okta" button that navigates to `GET /auth/okta` server-side redirect |
| **Session** | Same JWT cookie issuance |

### Effort estimate

| Phase | Effort |
|-------|--------|
| Backend (OIDC redirect flow) | ~2-3 days |
| Frontend (SSO button) | ~0.5 days |
| Per-tenant config management (if multi-tenant) | Extra ~2-3 days |
| Total | **~3-5 days** (single-tenant) |

### Recommendation

Straightforward to add after Google. The architecture is essentially identical to Google but via a server-side redirect (code exchange) rather than an ID token POST. If multi-tenant (each enterprise customer has its own Okta org), add a `tenant_id` column to `user_identities` or maintain a tenant config table.

---

## Summary

| Option | Effort | Complexity | Recommended when |
|--------|--------|------------|-----------------|
| **WebAuthn passkeys** | 1-2 weeks | High (native entitlements, crypto) | Have significant iOS/Android user base wanting passwordless login |
| **Okta OIDC** | 3-5 days | Low-Medium | Enterprise/B2B customers who already use Okta for SSO |

Both can be added with minimal disruption to existing auth since `user_identities` decouples identity providers from the user table. Google login (already implemented) should be validated first before adding either.

