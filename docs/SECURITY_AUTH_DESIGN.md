# Authentication, Sessions and Authorization Design

**Status:** Proposed design; implementation is gated on owner approval of API hosting and runtime.

**Implementation checkpoint (2026-11-02):** migration 011 adds normalized identity, role, seller-membership, digest-only session/action-token and append-only audit storage. The internal `backend/internal/auth` package now provides Argon2id PHC password hashing/verification and rehash checks, plus 256-bit opaque token generation, SHA-256 digests and constant-time digest comparison. New users receive only the customer role. These are internal foundations only: the package now includes a PostgreSQL session lifecycle and transaction-bound bulk session revocation with typed reasons, per-session redacted audit events and user-row locking against concurrent session creation. The internal `ChangePassword` function re-checks the password hash previously verified by the caller while holding the user row lock, then updates the Argon2id hash, revokes all active sessions and records a redacted audit event atomically. `CreateSession` requires the same verified hash snapshot and rejects stale credentials after a concurrent password change. Password, role or account-state mutations that call `RevokeUserSessions` must include those changes in the same transaction. No authentication HTTP handler, CSRF middleware, cookies, email delivery or public routes are enabled. Backend CI and security scans verified the stale-credential check and atomic password-change/session-revocation primitive; password-reset-token consumption and delivery remain future internal work. The internal password-reset token service reuses migration 011, stores only 32-byte SHA-256 digests, bounds expiry, rotates prior tokens and consumes them once under a user-first row-lock order. Completion changes the password, revokes sibling reset tokens and sessions, and appends a redacted audit event atomically. PostgreSQL race-enabled Backend CI [run 41395571464](https://github.com/pooya16pg-spec/globalcart/actions/runs/41395571464) and Security scans [run 41395571531](https://github.com/pooya16pg-spec/globalcart/actions/runs/41395571531) verified this internal lifecycle. The internal email-verification service reuses the same action-token table, rotates and consumes tokens once, and activates only the matching pending account while preserving the customer-only role. Backend CI [run 41396081833](https://github.com/pooya16pg-spec/globalcart/actions/runs/41396081833) and Security scans [run 41396081844](https://github.com/pooya16pg-spec/globalcart/actions/runs/41396081844) verified the implementation. No email delivery or public reset/verification endpoint is enabled. Benchmark Argon2id costs and add rate limiting on the approved runtime. ADR-010 remains proposed pending same-site HTTPS hosting, email delivery and secret-management approval.

The current Argon2id baseline is `m=19456 KiB, t=2, p=1`, matching OWASP's recommended minimum configuration. Verification accepts only bounded PHC resource costs (up to 64 MiB, five passes and four lanes) and supports rehashing bounded legacy hashes. See the [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html) and [RFC 9106](https://www.rfc-editor.org/rfc/rfc9106).
**Scope:** M04 identity boundary for the GlobalCart portfolio simulation.
**Security posture:** deny by default; no real commerce, payment credentials or seller payouts.

## 1. Deployment prerequisite

The current storefront is a GitHub Pages static site. Do not enable authenticated browser flows against a cross-site API from the current `github.io` origin: browser third-party-cookie restrictions make cookie sessions unreliable. Before enabling login, choose a custom frontend domain and expose the API on a same-site origin (for example, `www.example` and `api.example`), then configure HTTPS, exact CORS origins and secret management. No hosted service is provisioned by this design.

## 2. Identity model

- `users`: UUID primary key, normalized email, password hash (versioned Argon2id PHC string), email verification timestamp, status, created/updated timestamps.
- Enforce unique normalized email with a unique index on `lower(email)`; normalize and validate input in the application as well.
- Passwords are never stored, logged, emailed or returned. Use a maintained Argon2id implementation and parameters reviewed against current OWASP guidance; keep parameters encoded with each hash so they can be upgraded.
- Default role is customer. Registration cannot choose privileged roles.
- Email verification and password-reset tokens are cryptographically random, single-use, short-lived and stored only as hashes. Do not implement email delivery until an approved provider and secret store exist.

## 3. Session model

Use opaque server-side sessions, not browser-stored JWTs.

- Generate at least 32 cryptographically random bytes per session; return the base64url token once and store only its SHA-256 digest in PostgreSQL.
- Session records contain UUID, user ID, token digest, created/last-seen/expiry timestamps and revoked timestamp. Index token digest uniquely and user/expiry for cleanup.
- Use short idle and bounded absolute expiry. Rotate the session after authentication and privilege changes. Revoke sessions on logout, password change/reset and account disable.
- Cookie: host-only `__Host-globalcart_session`, `Secure`, `HttpOnly`, `Path=/`, `SameSite=Lax`, no Domain attribute. Local HTTP development may use a clearly isolated non-production cookie name/configuration.
- Do not put session identifiers in localStorage, sessionStorage, URLs, analytics or logs.
- Protect state-changing requests with a synchronizer CSRF token bound to the session and sent in a custom header; also validate Origin/Referer against the exact allow-list. SameSite is defense-in-depth, not the only CSRF control.

## 4. Roles and resource authorization

- Roles: `customer`, `seller`, `admin`. The role catalog is controlled by migrations; users cannot self-assign roles.
- `user_roles` records role assignments. A seller additionally requires an active seller membership/account relationship; seller access is scoped to that seller ID.
- Admin grants/revokes are privileged audited operations. Never trust a role, seller ID, owner ID or permission supplied by the browser.
- Middleware authenticates the session, loads current account status and effective roles, then authorizes each operation. Resource ownership/seller scope is checked in the same transaction as the data operation where possible.
- Deny by default. Return 401 for missing/invalid authentication, 403 for authenticated but unauthorized access, and avoid leaking whether another user's private resource exists.
- Start with database-backed authorization on each request. Add caching only after measurement and with explicit invalidation/revocation semantics.

## 5. Audit and abuse controls

- Append-only audit events for login success/failure, logout/revocation, password/email changes, role grants/revokes, account disable and privileged catalog/order actions.
- Never store passwords, raw session tokens, reset tokens, full authorization headers or payment data in audit records.
- Apply per-IP and per-account rate limits to login, registration, verification and reset endpoints; use generic login/reset responses to reduce account enumeration.
- Use HTTPS only, strict security headers, request-size limits, bounded timeouts and exact CORS allow-lists. CORS is not authorization.
- No passwordless/social login, MFA or external identity provider until the owner chooses and approves the operational provider.

## 6. Initial API boundary

Proposed public endpoints:
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/session`
- `POST /api/v1/auth/verify-email`
- `POST /api/v1/auth/password/forgot`
- `POST /api/v1/auth/password/reset`

Protected endpoints must use explicit route policy metadata. Registration/login/session are the only unauthenticated identity operations; verification/reset are single-use token flows. Do not implement any of these until same-site HTTPS hosting, email delivery, secret storage and runtime approval are available.

## 7. Required tests before release

- Password hash verification, parameter upgrade and no plaintext persistence.
- Session token entropy/digest storage, expiry, rotation, logout and global revocation.
- Cookie attributes and CSRF rejection for missing/invalid tokens and untrusted origins.
- Registration cannot grant seller/admin; inactive users cannot authenticate.
- Role and seller-scope matrix tests, including cross-seller IDOR attempts.
- Concurrent role revocation/session use and transaction-bound ownership checks.
- Rate-limit behavior, generic auth errors, audit redaction and no secret leakage.
- Integration tests against PostgreSQL migrations and end-to-end browser tests on the approved same-site domain.

## 8. Explicitly out of scope

No production hosting, paid email provider, payment integration, real order fulfillment, seller payouts, external identity provider or user data collection is authorized by this design.


## 9. Interim seller-service boundary (before identity implementation)

The current commerce package contains vendor-scoped operations that accept a `vendorID` as an internal selector. That selector is **not proof of caller authorization**. Until session authentication and active seller memberships are implemented:

- Do not expose `CreateVendor`, `UpsertVendorProduct`, `UpdateVendorPrice`, `AdjustVendorInventory` or seller order reads through public HTTP routes.
- Do not copy a `vendor_id`, role, owner ID or membership ID from request JSON and treat it as trusted authority.
- A future authenticated handler must resolve the current user from the verified server-side session, load the active seller membership from PostgreSQL, and derive the vendor scope server-side.
- Pass the resolved scope through a narrow application authorization boundary; keep the membership check and protected mutation in one transaction where practical.
- Keep customer cart capability operations separate from seller identity. A cart capability authorizes only that cart and must never imply seller privileges.
- Every seller read/write must constrain the target by the resolved vendor scope and have negative cross-vendor tests. Missing and foreign resources should use a non-enumerating response policy appropriate to the endpoint.

This interim restriction is intentional: internal service methods are not yet safe to publish as seller APIs merely because they accept a vendor ID or enforce a SQL `vendor_id` predicate.
