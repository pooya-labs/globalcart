# GlobalCart — Master Roadmap

> The single source of truth. A stage is complete only after implementation and verification.

**Current release:** 0.2.3-dev  
**Active stage:** M09.4 — 1.0.0 release readiness; M08.15 lab baseline and M09.1–M09.3 portfolio work are complete. Manual physical-device, real screen-reader, comprehensive RTL/LTR and field-delivery QA remain release gates. Public admin/seller workspaces remain gated by M04 hosting and identity prerequisites  
**Repository:** pooya16pg-spec/globalcart  
**Owner:** Pooya Goodarzi

## Engineering north star
Build a polished, bilingual global marketplace with excellent usability, accessible interaction, responsive layouts, reliable data and credible production-style architecture. Visual effects must never compromise readability, performance or accessibility.

## Target technology and database architecture

| Layer | Target | Purpose | Stage |
|---|---|---|---|
| Frontend | Next.js App Router, React, TypeScript | Storefront, product pages, admin/seller UX | Current |
| Backend | Go REST API + OpenAPI | Domain services, auth, catalog and orders | M04 |
| Primary database | PostgreSQL | Relational data and transactional consistency | M04 |
| Data access | SQL migrations + sqlc | Versioned schema and typed queries | M04 |
| Data/AI workers | Python | Supplier ingestion, enrichment and translation | M02/M04 |
| Cache/queue | Redis, optional | Cache and job coordination when justified | M04+ |
| Search | PostgreSQL FTS/trigram first | Search and relevance | M07 |

**Database decision:** PostgreSQL is the primary system of record. Marketplace entities have relationships and transactional invariants, so relational constraints, joins and ACID transactions are the right default. Redis is optional and never the canonical database. Do not add MongoDB or multiple databases without a measured use case.

**API decision:** REST + OpenAPI first. GraphQL only if real client query patterns justify it. Use versioned migrations and parameterized SQL.

**Runtime gate:** Local Go/PostgreSQL development and CI are implemented on `main`. Public API/database hosting, Python worker deployment, production secrets and paid services remain gated on explicit owner approval.

## Release and stage map

| Stage | Target | Scope | Status |
|---|---:|---|---|
| M00 | 0.1.0 | Product definition, independence, governance and handoff | Complete |
| M01 | 0.1.x | Storefront foundation, bilingual shell and deployment | Implemented; redesign supersedes original visual direction |
| M02 | 0.2.x | Automated catalog import, normalization and snapshots | Snapshot pipeline implemented; authorized supplier feeds remain gated; translation paused |
| M03 | 0.3.0 | Pricing/FX provenance, product details, gallery and comparison | TGJU free-market USD page feed, Rial-to-Toman normalization, stale/date-only handling and storefront deployment are CI-verified; five-minute recurring cadence remains under observation |
| M04 | 0.4.0 | Go API, PostgreSQL, migrations, auth boundary and Python workers | Read-only local Go/PostgreSQL foundation implemented; authentication, workers and hosting gates remain |
| M05 | 0.5.0 | Persistent demo commerce journeys, orders, coupons and returns | Internal demo cart, checkout, order, coupon, vendor-split and full-order return/refund simulation implemented; public API exposure gated |
| M06 | 0.6.0 | Admin/seller workspaces, import operations and analytics | Internal admin and seller services through profile/demo analytics CI-verified; dashboard UI and public routes remain gated |
| M07 | 0.7.0 | Advanced search, relevance, recommendations and saved searches | Implemented in storefront and read-only catalog API; saved preferences remain browser-local |
| M08 | 0.8.0 | Browser tests, accessibility, security and performance | Automated browser, axe, focus, asset-budget and CI security checks implemented; manual QA and operational recovery remain |
| M09 | 1.0.0 | Portfolio case study, release and public demo QA | Planned |

## M00 — Product and governance
- [x] Portfolio simulation, not a live retailer.
- [x] Independent repository and owner attribution.
- [x] Safe source policy; no unauthorized scraping.
- [x] Versioning, roadmap and AI handoff rules.
- [x] Architecture decisions recorded.

## M01 — Storefront foundation
- [x] Next.js, React and TypeScript foundation.
- [x] Static export and Pages deployment workflow.
- [x] Responsive editorial marketplace storefront.
- [x] Persian-first RTL and English LTR.
- [x] Persist the selected Persian/English locale across homepage, product-detail and recently-viewed routes and full reloads, with guarded browser storage and a safe Persian fallback; Browser E2E [41462200408](https://github.com/pooya16pg-spec/globalcart/actions/runs/41462200408) passed all 34 desktop/mobile tests, and Pages deployment [41462200496](https://github.com/pooya16pg-spec/globalcart/actions/runs/41462200496) passed published-site smoke checks.
- [x] Search, categories, filters and sorting.
- [x] Quick view, favorites and recently viewed.
- [x] Local cart and simulated checkout.
- [x] Owner attribution and demo/source disclosures.
- [x] GitHub Actions build, typecheck, tests, artifact upload and Pages deploy passed in run 41294484252.
- [ ] Visual QA on phone, tablet and desktop.
- [ ] Keyboard, focus, screen-reader, contrast and reduced-motion audit.
- [x] Verify published homepage, base-path assets and product deep links with the automated Pages smoke test [41413878826](https://github.com/pooya16pg-spec/globalcart/actions/runs/41413878826).

## M01.5 — Marketplace redesign and UX foundation
- [x] Re-scope reference to a broad Amazon-style marketplace information architecture without copying proprietary code, branding or assets.
- [x] Pause translation and remove it from the deployment pipeline.
- [x] Rebuild the storefront shell around category-first discovery, prominent search and shopping utilities.
- [x] Add a responsive category mega-menu generated from the live catalog, with product counts and direct category filtering.
- [x] Add data-driven homepage modules for top discounts, top-rated products and fresh discoveries.
- [x] Add Amazon product-link intake with domain/ASIN validation and a local-only demo request queue.
- [x] Add marketplace service shortcuts and a transparent FAQ for sourcing, support and demo boundaries.
- [x] Improve product cards, price hierarchy, ratings, availability labels and quick actions.
- [x] Add dedicated static product detail routes for the imported catalog, image gallery, product information and related products.
- [x] Add side-by-side comparison for up to four products, with price, rating, discount, brand and source-stock context.
- [x] Add keyboard dialog semantics, Escape dismissal, visible focus indicators and status announcements.
- [ ] Verify desktop, tablet, mobile, RTL/LTR, screen-reader and reduced-motion behavior.
- [x] Add clear demo disclosures for Amazon link intake, seller tools and support paths.
- [ ] Confirm all controls remain functional across the full site.

## M02 — Automated catalog pipeline (active)
- [x] Typed catalog contracts and normalization.
- [x] Invalid-record rejection, source identity and duplicate suppression.
- [x] DummyJSON source adapter.
- [x] Translation provider interface with honest unavailable/error states.
- [x] Scheduled and manual catalog refresh trigger.
- [x] Retry-capable static snapshot generator with last-known-good protection.
- [x] Storefront initializes from generated snapshot and keeps it if live refresh fails.
- [x] Verify snapshot generation (100 products), TypeScript, unit tests, production build and Pages deployment in Actions run 41295039930.
- [x] Add a Google Apps Script Language Service translation bridge using Google Translate (no Gemini).
- [x] Translation intentionally paused; no translation secrets or external translation service required for CI.
- [x] Translate title, description, category and brand while preserving original source fields and provenance.
- [x] Translation cache via source-content fingerprints, sequential throttling, bounded retries and per-product error reporting.
- [x] Show the last successful source-fetch time or generated snapshot timestamp and classify its age as fresh (≤36h), aging (>36h to 72h), stale (>72h) or unknown in Persian and English. Verified by Catalog CI [41406450515](https://github.com/pooya16pg-spec/globalcart/actions/runs/41406450515), Browser E2E [41406450651](https://github.com/pooya16pg-spec/globalcart/actions/runs/41406450651), Pages deploy + published smoke test [41406450478](https://github.com/pooya16pg-spec/globalcart/actions/runs/41406450478), and Security scans [41406450499](https://github.com/pooya16pg-spec/globalcart/actions/runs/41406450499). The timestamp measures GlobalCart fetch/generation time, not supplier modification time.
- [ ] Add other suppliers only through authorized APIs/feeds.
**Acceptance:** products import automatically, the storefront survives source outages, and original source text remains intact. Translation is out of scope until explicitly reactivated.

**Freshness semantics:** the public catalog job runs daily. A 36-hour fresh window allows one delayed/missed schedule; 72 hours marks two missed daily cycles as stale. This is snapshot-generation/fetch age only and is not a claim about upstream product modification timestamps.

## M03 — Pricing and product experience
- [x] Integer-cent pricing, fixed-point FX and BigInt intermediate arithmetic.
- [x] Explicit rounding and invalid-input tests.
- [x] Replace the former Wallex USDT proxy with the requested TGJU free-market USD page. The `price_dollar_rl` row is validated; the raw Rial quote is preserved and divided by ten for Toman per USD; Tehran clock and date-only markers are handled conservatively; blocked/invalid pages preserve last-known-good data; 5/15-minute freshness states are tested. TGJU refresh [41568568791](https://github.com/pooya16pg-spec/globalcart/actions/runs/41568568791), Security [41568713623](https://github.com/pooya16pg-spec/globalcart/actions/runs/41568713623), Pages deployment/smoke [41568713584](https://github.com/pooya16pg-spec/globalcart/actions/runs/41568713584), and desktop/mobile Browser E2E [41568713639](https://github.com/pooya16pg-spec/globalcart/actions/runs/41568713639) passed. Current source marker is date-only (`11 آبان`), so the stored 271,500 Toman/USD quote is deliberately labelled stale. Recurring schedule reliability remains under observation.
- [x] Transparent source-price, discount and exchange-derived Tether/Toman estimate breakdown on product pages (no shipping, tax, payment fees or executable checkout quote).
- [x] Dedicated static product details for the imported catalog, responsive image gallery and related products.
- [x] Show source record ID, import time, source URL when present, provided variants/attributes, and disclose when source review count is not supplied.
- [x] Product comparison modal and persistent comparison shortlist.
- [x] Dedicated recently viewed page, homepage rail and browser-local visit history.

## M04 — Backend and database platform
- [ ] Owner approves separate public API hosting for the existing Go/PostgreSQL runtime after selecting a same-site HTTPS origin, provider/region, budget and secret-management approach.
- [ ] Owner enables Python runtime or an approved worker runner.
- [x] Add local-only PostgreSQL 17 Docker Compose service with loopback-only port binding and persistent development volume.
- [ ] Provision hosted PostgreSQL and production secret management after owner selects provider, region and budget.
- [x] Define initial catalog schema with provenance, numeric constraints, JSONB images and category/search indexes.
- [x] Extend catalog schema with normalized categories, product-to-category relationships and constrained product variants.
- [x] Add embedded, versioned SQL migrations and transactional migration ledger for the initial catalog schema.
- [x] Record SHA-256 checksums for applied migration files; reject changed or missing applied migrations and adopt legacy NULL checksums once.
- [x] Define reviewed rollback/forward-fix policy for shared environments in docs/DATABASE_MIGRATION_POLICY.md.
- [x] Add Go REST API foundation, OpenAPI 3.1 contract, health/readiness probes, CORS allow-list and read-only catalog list/detail endpoints.
- [x] Validate catalog pagination bounds and query/category lengths with stable 400 error codes.
- [x] Document the API error response shape and list endpoint validation in OpenAPI.
- [x] Add PostgreSQL integration coverage for category backfill and product variant JSONB persistence.
- [x] Add sqlc configuration, parameterized catalog/variant query definitions and CI code generation for pgx/v5.\n- [x] Wire generated sqlc queries into product list/detail API handlers; importer remains transaction-oriented and uses explicit upsert SQL.
- [x] Add PostgreSQL identity/session/RBAC schema foundation (migration 011); no authentication or commerce write routes are exposed.
- [x] Design the internal authentication/session and deny-by-default role-based authorization boundary (M04.1–M04.12); public middleware remains deployment-gated.
- [ ] Add idempotent Python ingestion/translation workers.
- [ ] Add Redis only for measured cache/queue requirements.
- [x] Add Go unit tests and backend CI for formatting, vet, race tests, builds and Docker image build.
- [x] Add PostgreSQL integration tests for concurrent/idempotent migrations, catalog upserts and transaction rollback.
- [x] Add PostgreSQL authorization integration tests for seller vendor-scope and customer cart/order ownership, cross-tenant IDOR, revoked sessions and ownership races.

### M04.1 — Identity/session/RBAC database foundation (2026-11-02)
- [x] Add migration 011 for users, migration-controlled customer/seller/admin roles and vendor-scoped seller memberships.
- [x] Enforce normalized unique emails, Argon2id PHC password-hash format, verified-email activation and explicit account-disable state.
- [x] Store only 32-byte SHA-256 session and CSRF token digests with bounded idle/absolute expiry and explicit revocation reasons.
- [x] Store only hashed email-verification/password-reset action tokens; enforce expiry, single consumption and one active token per user/purpose.
- [x] Add bounded/redacted audit metadata and database triggers rejecting UPDATE, DELETE and TRUNCATE.
- [x] Make the migration runner safe for semicolons inside quoted values, comments and PostgreSQL dollar-quoted function bodies; add unit and PostgreSQL integration coverage.
- [ ] Add auth/session HTTP handlers, cookies, CSRF middleware and authorization tests only after same-site HTTPS hosting, email delivery and secret storage are approved.

**Boundary:** migration 011 is storage scaffolding only. It does not authenticate requests, issue browser cookies, expose seller/admin endpoints or connect a hosted database. ADR-010 remains proposed pending owner approval.

### M04.2 — Internal authentication cryptographic primitives (CI verified, 2026-11-02)
- [x] Pin maintained `golang.org/x/crypto` Argon2id implementation and hash passwords with a unique 16-byte random salt using the OWASP baseline `m=19456 KiB, t=2, p=1`.
- [x] Encode hashes as Argon2id PHC strings; strictly parse versions, parameters and canonical base64 before verification.
- [x] Bound verification work to at most 64 MiB, 5 passes and 4 lanes; allow bounded legacy hashes to verify and flag for rehash rather than downgrading new hashes.
- [x] Compare derived password keys in constant time and reject empty, invalid UTF-8 or oversized password inputs.
- [x] Generate canonical opaque tokens from 256 bits of cryptographic randomness, persist only SHA-256 digests and compare digests in constant time.
- [x] Add tests for correct/incorrect passwords, malformed/over-budget PHC strings, legacy rehash decisions, invalid password inputs, token entropy/shape and digest validation.
- [x] Verify Backend CI [41385892498](https://github.com/pooya16pg-spec/globalcart/actions/runs/41385892498), Security scans [41385892505](https://github.com/pooya16pg-spec/globalcart/actions/runs/41385892505) and GitHub Pages deployment [41385892515](https://github.com/pooya16pg-spec/globalcart/actions/runs/41385892515) on implementation head `862ffb30edefb23695e34e407bfe7ea8180162ef`.

**Boundary:** these primitives and session-store methods remain internal; no registration/login handler, cookie, CSRF middleware, email delivery or public HTTP route is enabled. Benchmark hash parameters and add rate limits before authentication is activated.

### M04.3 — Internal server-side session lifecycle (CI verified, 2026-11-02)
- [x] Create sessions only for active, email-verified accounts; generate independent 256-bit session and CSRF tokens and persist only SHA-256 digests.
- [x] Use a 15-minute sliding idle timeout capped by a 24-hour absolute expiry.
- [x] Resolve session validity and current user roles in one database statement; reject expired, revoked, disabled or unverified accounts.
- [x] Grant effective seller access only when the user has an active seller role, active membership and an active vendor; refresh scope from PostgreSQL rather than browser claims.
- [x] Validate the session-bound CSRF token using constant-time digest comparison.
- [x] Revoke sessions on logout and append creation/revocation audit events in the same transaction.
- [x] Add PostgreSQL integration coverage for raw-token non-persistence, CSRF, pending/active seller memberships, expiry, disabled/unverified accounts, idempotency boundary and audit events.
- [x] Add transaction-bound bulk revocation for password change/reset and privilege/account-status changes; verify rollback, owner isolation, allow-listed reasons and per-session audit events.
- [ ] Wire this internal service into HTTP authentication middleware only after same-site HTTPS hosting, cookie configuration, email delivery and secret storage are approved.

**Boundary:** internal store/service only. No browser cookie, route, middleware, public authentication flow or hosted database has been enabled.

### M04.4 — Transaction-bound bulk session revocation (CI verified, 2026-11-02)
- [x] Add an allow-listed revocation-reason type matching migration 011 constraints.
- [x] Lock the account row and serialize bulk revocation with session creation.
- [x] Accept the caller's PostgreSQL transaction so account/password/privilege mutation and session revocation can commit or roll back together.
- [x] Revoke every active session for one user and append one redacted audit event per session, with an optional actor identity.
- [x] Return the affected count; repeated revocation is a no-op and does not create duplicate audit events.
- [x] Add integration coverage for rollback, invalid reasons, user isolation, audit actor/reason and idempotency.
- [x] Verify the complete race-enabled PostgreSQL Backend CI suite and security scans: Backend CI [run 41393919125](https://github.com/pooya16pg-spec/globalcart/actions/runs/41393919125) and Security scans [run 41393919064](https://github.com/pooya16pg-spec/globalcart/actions/runs/41393919064) both passed on code head `6e044b63048b0911c5a7078279e38fe354898605`.
- [x] Verify the concurrent session-creation/account-disable regression: Backend CI [run 41394508053](https://github.com/pooya16pg-spec/globalcart/actions/runs/41394508053) passed the PostgreSQL race-enabled suite on implementation head `4e74a223e87ee55f84127039be99735eb30b089e`.
- [x] Re-run the full current auth integration suite, including the session/account-disable race, password change and password-reset revocation paths: Backend CI [run 41395571464](https://github.com/pooya16pg-spec/globalcart/actions/runs/41395571464) passed on code head `8a504dca8ae23ca2c06143808452898281216abe`.

**Boundary:** this is an internal transaction helper, not a password-change endpoint or authorization middleware. Callers must include the relevant credential/role/account update in the same transaction. No HTTP route, browser cookie, email flow or hosted database is enabled.

### M04.5 — Atomic password change and stale-login rejection (CI verified, 2026-11-02)
- [x] Require session creation to present the exact password-hash snapshot that was verified; compare it under the account row lock before issuing tokens.
- [x] Add an internal password-change service that re-checks the expected old hash, updates the Argon2id hash, revokes every active session and audits the change in one transaction.
- [x] Reject login attempts whose previously verified password hash became stale while waiting for the account lock.
- [x] Add PostgreSQL integration coverage for multi-session revocation, old-password rejection, new-password verification and stale credential snapshots.
- [x] Verify Backend CI [run 41395032305](https://github.com/pooya16pg-spec/globalcart/actions/runs/41395032305), Security scans [run 41395032374](https://github.com/pooya16pg-spec/globalcart/actions/runs/41395032374) and GitHub Pages deployment [run 41395032292](https://github.com/pooya16pg-spec/globalcart/actions/runs/41395032292) on code head `fe2ec6b9bfce6ec4aa20ea004029e2a68d5b05ab`; all passed.

**Boundary:** this is an internal password-change primitive, not a public endpoint or password-reset-token flow. The caller must verify the old password against the supplied hash snapshot before invoking it. No login/register HTTP route, browser cookie, email delivery or hosted database is enabled.

### M04.6 — One-time password-reset token lifecycle (CI verified, 2026-11-02)
- [x] Reuse migration 011 `auth_action_tokens`; do not create a duplicate token table or migration.
- [x] Issue 256-bit opaque reset tokens and persist only SHA-256 digests with a bounded 5-minute to 1-hour TTL (30-minute default).
- [x] Rotate prior outstanding reset tokens on issuance, including expired rows that still occupy the partial unique index.
- [x] Consume tokens once and atomically update the Argon2id password, revoke sibling reset tokens and active sessions, and append a redacted audit event.
- [x] Lock the user before token rows in issue/consume/change flows to avoid lock-order deadlocks; re-check expiry and single-use state inside the transaction.
- [x] Invalidate outstanding reset tokens when a normal password change succeeds.
- [x] Add PostgreSQL tests for digest-only storage, rotation, expiry, single use, concurrent double consumption and session invalidation.
- [x] Verify PostgreSQL race-enabled Backend CI [run 41395571464](https://github.com/pooya16pg-spec/globalcart/actions/runs/41395571464), Security scans [run 41395571531](https://github.com/pooya16pg-spec/globalcart/actions/runs/41395571531) and GitHub Pages deployment [run 41395571666](https://github.com/pooya16pg-spec/globalcart/actions/runs/41395571666) on code head `8a504dca8ae23ca2c06143808452898281216abe`; all passed.

**Boundary:** no email is sent and no public password-reset endpoint is enabled. The raw token is returned once to an internal caller for delivery by a future approved email worker; HTTP responses must not expose account existence.

### M04.7 — Email-verification token lifecycle and pending-account activation (CI verified, 2026-11-02)
- [x] Reuse migration 011 `auth_action_tokens` with purpose `email_verification`; no duplicate token table or migration was added.
- [x] Add internal issuance with digest-only token storage, bounded expiry, token rotation and an explicit no-public-endpoint boundary.
- [x] Consume a verification token once, activate only the matching pending account, and revoke stale verification tokens transactionally.
- [x] Audit issuance and successful verification without storing raw token/email secrets.
- [x] Test expiry, duplicate/replayed tokens, token rotation, disabled-account behavior, customer-only roles and concurrent verification attempts in PostgreSQL.
- [x] Verify Backend CI [run 41396081833](https://github.com/pooya16pg-spec/globalcart/actions/runs/41396081833) and Security scans [run 41396081844](https://github.com/pooya16pg-spec/globalcart/actions/runs/41396081844) on code head `d720ffdebd77c6b28c82ce55567ea3bba90d95c8`; both passed, including PostgreSQL race-enabled tests.
- [x] Keep email delivery and public registration/verification routes disabled pending approved hosting, email provider and secret management.

**Boundary:** pending-account activation only; this does not grant seller/admin roles or enable a public registration endpoint.

### M04.8 — Audited internal administrator-role grant/revocation (CI verified, 2026-11-02)
- [x] Reuse migration 011 `user_roles`; no new role table or migration was added.
- [x] Add `GrantAdminRole`: require a different active, email-verified administrator actor and an active, verified target; reject self-grants and duplicate grants.
- [x] Add `RevokeAdminRole`: require a different active, email-verified administrator actor; reject self-revocation and missing grants.
- [x] Serialize privileged role mutations with a transaction-scoped advisory lock and lock actor/target rows in stable order.
- [x] Revoke every target session in the same transaction as the role mutation; append actor-attributed, redacted audit events.
- [x] Add PostgreSQL tests for authorization, disabled actors, pending targets, self-change rejection, duplicate operations, session invalidation and concurrent admin grants.
- [x] Verify Backend CI [run 41396823353](https://github.com/pooya16pg-spec/globalcart/actions/runs/41396823353), Security scans [run 41396823371](https://github.com/pooya16pg-spec/globalcart/actions/runs/41396823371) and GitHub Pages deployment [run 41396823403](https://github.com/pooya16pg-spec/globalcart/actions/runs/41396823403) on code head `aef92e7526cfbc9c51f0d803bcba8493e131d8d2`; all passed.
- [x] Keep public/admin HTTP routes disabled. The first administrator must be provisioned through a reviewed operator-controlled database procedure; no self-bootstrap API exists.

**Boundary:** internal privileged-role operations only. The trusted caller must derive `actorUserID` from a verified server-side session, never from request JSON or a browser-supplied role claim. Seller membership lifecycle and public authorization middleware are not included.

### M04.9 — Seller-membership administration and scoped authorization (CI verified, 2026-11-02)
- [x] Add internal `RequestSellerMembership`, `ActivateSellerMembership`, `SuspendSellerMembership` and `RevokeSellerMembership`; reuse migration 011 `seller_memberships` and `user_roles` without a new migration.
- [x] Require active, verified accounts and active vendors for requests/activation; the seller role is granted only on admin approval, never from browser-supplied role/vendor claims.
- [x] Serialize admin review with the existing privileged-mutation lock, preserve other vendor scopes, revoke the seller role only after the last active vendor scope is removed, and invalidate target sessions transactionally.
- [x] Append actor-attributed audit events for membership requests, activation, suspension/revocation and automatic seller-role grant/revoke.
- [x] Add PostgreSQL tests for multi-vendor scope retention, no-scope role removal, disabled target cleanup, inactive vendors, unauthorized actors and concurrent activation.
- [x] Verify Backend CI [run 41397502226](https://github.com/pooya16pg-spec/globalcart/actions/runs/41397502226), Security scans [run 41397502232](https://github.com/pooya16pg-spec/globalcart/actions/runs/41397502232) and GitHub Pages deployment [run 41397502248](https://github.com/pooya16pg-spec/globalcart/actions/runs/41397502248) on code head `f0756e0a3360b949b384aef1ff8e3a646192257b`; all three passed.



**Boundary:** internal membership administration only; no seller self-approval, payouts, seller dashboard or public seller endpoint.

### M04.10 — Deny-by-default internal authorization policy (CI verified, 2026-11-02)
- [x] Define explicit `AuthorizationAction` and `AuthorizationResource` types for customer ownership, seller vendor scope and administrator-only actions.
- [x] Require a `SessionPrincipal` resolved by `ResolveSession`; browser role claims or request-body owner/vendor IDs must never construct trusted principal state.
- [x] Deny unknown actions/resources by default; return typed `ErrUnauthenticated`/`ErrForbidden` without revealing cross-tenant resource existence.
- [x] Add table-driven policy tests for customer ownership, assigned/unassigned vendor scopes, seller role without a live scope, admin-only actions, unknown actions, missing session identity and client-supplied vendor scope.
- [x] Keep the evaluator pure/internal; no middleware, public protected route or cookie flow was enabled.

- [x] Verify Backend CI [run 41398064895](https://github.com/pooya16pg-spec/globalcart/actions/runs/41398064895), Security scans [run 41398064885](https://github.com/pooya16pg-spec/globalcart/actions/runs/41398064885) and GitHub Pages deployment [run 41398064905](https://github.com/pooya16pg-spec/globalcart/actions/runs/41398064905) on implementation head `c1b4b0ccfbd132ee95f8dc757a3518ff9b04899a`; all passed.

**Boundary:** pure/internal authorization policy only. Its caller must pass a `SessionPrincipal` returned from `ResolveSession`; this evaluator does not authenticate, refresh sessions or replace transaction-level ownership checks. Public handlers, cookies, CSRF middleware and deployment remain blocked on ADR-010 owner approval.
### M04.11 — Service-boundary authorization and cross-tenant IDOR enforcement (CI verified, 2026-11-02)
- [x] Add `AuthorizeSellerVendorTx` to revalidate a resolved session, active account, live seller role, active vendor and exact membership under row locks inside the caller's transaction.
- [x] Add seller-scoped read, price-update and inventory-adjustment methods that derive vendor scope from the canonical offer row; the caller cannot select or override a vendor ID.
- [x] Reuse the existing price-history/inventory-event transaction helpers rather than duplicating commerce mutation logic; preserve the legacy internal untrusted-ID methods as non-HTTP helpers.
- [x] Add PostgreSQL coverage for own-offer success, cross-vendor read/write denial, stale/revoked session rejection, suspended membership and membership-revocation serialization with a protected seller transaction.
- [x] Keep protected HTTP routes disabled; no cookies, middleware, CSRF route or public seller endpoints were enabled.

- [x] Verify Backend CI [run 41399542624](https://github.com/pooya16pg-spec/globalcart/actions/runs/41399542624), Security scans [run 41399542539](https://github.com/pooya16pg-spec/globalcart/actions/runs/41399542539) and GitHub Pages deployment [run 41399542545](https://github.com/pooya16pg-spec/globalcart/actions/runs/41399542545) on implementation head `f06de264488bd59f1e76a7e3d6c67f26d40c0eb1`; all passed, including PostgreSQL race tests, API/importer builds, Docker image, static build and Pages deploy.

- [x] Verify Backend CI [run 41399542624](https://github.com/pooya16pg-spec/globalcart/actions/runs/41399542624), Security scans [run 41399542539](https://github.com/pooya16pg-spec/globalcart/actions/runs/41399542539) and GitHub Pages deployment [run 41399542545](https://github.com/pooya16pg-spec/globalcart/actions/runs/41399542545) on implementation head `f06de264488bd59f1e76a7e3d6c67f26d40c0eb1`; all passed, including PostgreSQL race tests, API/importer builds, Docker image, static build and Pages deploy.

**Boundary:** service/application-layer enforcement only; no login, cookie, CSRF middleware or public route is enabled.
### M04.12 — Customer resource ownership enforcement (CI verified, 2026-11-02)
- [x] Inventory internal customer cart/order resource operations and keep bearer-capability guest operations distinct from authenticated account operations.
- [x] Add `carts.owner_user_id` (migration 012); bind a guest cart only when both its 32-byte capability digest and a live verified customer session are proven. The first account to bind wins.
- [x] Add customer-owned cart read/add/set/remove operations and parent/vendor-child order reads plus unpaid-order cancellation. Resolve the owner from PostgreSQL and revalidate the session, customer role and resource ownership inside the protected transaction.
- [x] Keep guest cart capabilities separate from account identity; the capability authorizes only its own cart and never grants a role or seller/admin scope.
- [x] Add PostgreSQL integration coverage for foreign/guest-only cart and order denial, cart mutations, parent and vendor child-order reads, first-writer-wins binding races, expired carts, revoked sessions and idempotent cancellation.
- [x] Keep customer HTTP routes, browser cookies and middleware disabled until same-site HTTPS hosting, email delivery, secret management and runtime are explicitly approved.

**Verification on implementation head `928c3867e2d8e109e4e6595f9ac0898cf51b25b9`:** [Backend CI 41403135351](https://github.com/pooya16pg-spec/globalcart/actions/runs/41403135351), [Security scans 41403135364](https://github.com/pooya16pg-spec/globalcart/actions/runs/41403135364), and [GitHub Pages deployment 41403135344](https://github.com/pooya16pg-spec/globalcart/actions/runs/41403135344) all succeeded. Backend CI completed sqlc generation, Go formatting, dependency vulnerability scan, `go vet`, race-enabled Go tests, API/importer builds and Docker image build. The CI formatting-only commit is `5a5659d9b35789cf12072141152b0ab4be217c33`.

**Boundary:** account-backed service methods remain internal. No customer authentication, cart/order write, seller or admin HTTP route, cookie flow or public API was enabled.


### M04.13 — Immutable migration checksums and recovery policy (CI verified, 2026-11-02)
- [x] Extend schema_migrations with a SHA-256 checksum of each exact embedded SQL file; preserve compatibility with the original two-column ledger.
- [x] Adopt pre-existing NULL checksums once, then enforce NOT NULL; reject checksum drift and applied ledger entries whose source migration file is missing.
- [x] Add regression coverage for deterministic/content-sensitive checksums, legacy-ledger adoption and fail-closed drift detection.
- [x] Document append-only migration rules, expand-and-contract rollout, transaction limitations, forward-fix decisions, backup/restore boundaries and an incident runbook in docs/DATABASE_MIGRATION_POLICY.md.
- [x] Verify Backend CI [41415943162](https://github.com/pooya16pg-spec/globalcart/actions/runs/41415943162), Security scans [41415943090](https://github.com/pooya16pg-spec/globalcart/actions/runs/41415943090) and Pages deployment [41415943078](https://github.com/pooya16pg-spec/globalcart/actions/runs/41415943078); all passed on implementation/docs head `2cd5aaf72bbcddd14bfb1312768b6c708648f9d4`.

**Boundary:** CI/local PostgreSQL only. This policy does not provision a hosted database or claim production backups, PITR, RPO/RTO or an approved rollback window.

### M04.14 — Fail-closed backend runtime configuration (CI verified, 2026-11-02)
- [x] Parse `APP_ENV` into explicit development, test and production modes; reject unknown values.
- [x] Validate PostgreSQL URL scheme, host, database name, fragment and supported `sslmode`; require TLS in production.
- [x] Validate `HTTP_ADDR` as a host/port listener with a non-ephemeral port in the valid TCP range.
- [x] Accept only exact CORS origins; reject wildcard/null, credentials, paths, queries and fragments; permit plain HTTP only for loopback in development/test and require HTTPS in production.
- [x] Add unit coverage for defaults, malformed database URLs, invalid listener addresses, normalized origins, unsafe CORS values and production TLS/HTTPS requirements.
- [x] Verify Backend CI [41458549348](https://github.com/pooya16pg-spec/globalcart/actions/runs/41458549348) and Security scans [41458549236](https://github.com/pooya16pg-spec/globalcart/actions/runs/41458549236); both passed, including PostgreSQL 17 integration/race tests, logical backup/restore, API/importer builds and the container image.

**Boundary:** configuration validation only. It does not provision a hosted database, terminate HTTPS, create secrets, or enable public authentication/commerce routes. The operator must still set `APP_ENV=production`, supply TLS-capable PostgreSQL settings and configure the exact production frontend origin.

## M05 — Commerce simulation
- [x] Add local schema foundation for guest carts, bounded cart items and expiring capability digests (migration 003; API not exposed).
- [x] Internal multi-step checkout foundation: cart validation, stock reservations, immutable order snapshots and idempotent checkout (demo address book remains pending).
- [x] Define demo-only order states and reservation release semantics; no payment provider is contacted.
- [x] Add demo order ledger, immutable line snapshots, per-cart idempotency and append-oriented events (schema only; API not exposed).
- [x] Internal demo coupon validation/application, hashed coupon codes, usage limits and discount allocation during vendor split-checkout (CI-verified; public administration remains unavailable).
- [x] Full-order return request, internal approval/rejection and simulated refund completion (migration 010; owner-scoped, idempotent, no payment provider, no stock mutation, no HTTP route).

### M05.1 — Internal full-order return and simulated refund — CI verified (2026-11-02)
- [x] Add one return request per paid simulated parent order with a fixed reason allow-list and database-enforced lifecycle.
- [x] Enforce cart-capability ownership for requests and refund completion; reject unpaid orders, invalid reasons, duplicate conflicting requests and rejected returns.
- [x] Add internal approve/reject review. Approval snapshots the order's net persisted total, including coupon discount.
- [x] Atomically update parent order, vendor child orders, return state and append-only event on simulated refund completion.
- [x] Keep review and completion off the public HTTP router; never call a payment provider or mutate vendor/product stock.
- [x] Add PostgreSQL integration tests for owner isolation, idempotency, approval/rejection, refund state, event history and unchanged stock.
- [x] Verify Backend CI [41382309964](https://github.com/pooya16pg-spec/globalcart/actions/runs/41382309964), Security scans [41382310050](https://github.com/pooya16pg-spec/globalcart/actions/runs/41382310050) and GitHub Pages deployment [41382310024](https://github.com/pooya16pg-spec/globalcart/actions/runs/41382310024) on final implementation head `bc97475c0cce7f8de6aebe72071813c942b6a3d0`.

**Boundary:** this is a full-order portfolio simulation only. No real money, refund provider, customer login, item-level return, shipment or stock adjustment exists.

## M06 — Admin and seller workspaces
- [ ] Demo role switcher with real authorization boundaries where API exists; no browser role impersonation is allowed before authenticated routes exist.
- [ ] Admin overview UI and catalog/translation/price diagnostics; the internal aggregate service is implemented in M06.1 below.

- [x] Admin-only manual product creation with a reserved ID sequence, `admin_manual` provenance and a transactional GlobalCart Direct demo offer; M06.2 below.
- [x] Admin-only product editorial edit, archive and restore with immutable revision history; M06.2 below.
- [x] Internal admin product work queue with effective-value search, status/source/category/brand filters, safe sorting and bounded pagination; M06.3 below. Backend CI [41562259407](https://github.com/pooya16pg-spec/globalcart/actions/runs/41562259407) and Security scans [41562285730](https://github.com/pooya16pg-spec/globalcart/actions/runs/41562285730) passed.
- [x] Internal seller product work queue with live vendor-membership authorization, effective catalog search, offer status/source/category/brand filters, safe sorting and bounded pagination; M06.4 below. Backend CI [41564104334](https://github.com/pooya16pg-spec/globalcart/actions/runs/41564104334) and Security scans [41564104323](https://github.com/pooya16pg-spec/globalcart/actions/runs/41564104323) passed.
- [ ] Supplier connector configuration without secret exposure.
- [ ] Seller dashboard UI and browser role switcher remain gated; M06.6 internal seller profile and demo analytics services are CI-verified.

### M06.1 — Internal admin overview service foundation — CI verified (2026-11-02)
- [x] Add a dedicated deny-by-default `admin.overview.read` authorization action.
- [x] Resolve an opaque server-side session token from PostgreSQL and require an active administrator role; customer and seller sessions are denied.
- [x] Revalidate the active account, session and administrator role under row locks in the same transaction as the overview read, preventing stale-role/session authorization races.
- [x] Aggregate catalog/product/category/variant quality and source diagnostics; vendor and offer state; seller membership and account state; price-history coverage and last update; and audit event volume/recency.
- [x] Report translation as intentionally paused; do not imply a worker/provider is active.
- [x] Return aggregate-only data without account emails, password hashes, raw session/CSRF tokens, token digests or audit payloads.
- [x] Add PostgreSQL integration tests for customer/seller denial, active-admin reads, fixture diagnostics, stale role/session revocation and sensitive-field non-disclosure.
- [x] Verify Backend CI [41465375163](https://github.com/pooya16pg-spec/globalcart/actions/runs/41465375163) and Security scans [41465375326](https://github.com/pooya16pg-spec/globalcart/actions/runs/41465375326) on implementation head `7ca99c04a348256aa0a6051241864213968f7f84`.

**Boundary:** `backend/internal/admin.ReadOverview` is an internal Go service only. No HTTP route, UI, public role switcher, account bootstrap, hosted service or seller operation was enabled. The future dashboard must call this service only through approved authenticated middleware after same-site HTTPS hosting, email delivery and secret management are approved.

### M06.2 — Internal product creation, editorial management and archive — CI verified (2026-11-02)
- [x] Add migrations 013–014 for separate administrator overrides/revisions and a reserved high-range manual-product ID sequence; preserve supplier-owned product fields and provenance.
- [x] Add explicit `admin.products.manage` permission and revalidate the verified active account, exact live session and active admin role under row locks in the protected transaction.
- [x] Create manual products with generated IDs and `admin_manual` source provenance; atomically create the platform demo offer and opening price/stock history.
- [x] Support bounded title, description and brand overrides, including explicit clearing to restore the latest supplier value.
- [x] Require the caller's expected product revision on edit/archive/restore; reject stale mutations instead of silently overwriting newer work.
- [x] Archive without deleting the source product, variants, vendor offers or order references; restore the original product and retain editorial overrides.
- [x] Record append-only before/after snapshots with effective editorial/source values, creation details, actor ID, action and monotonic revision; append a matching redacted audit event in the same transaction.
- [x] Add bounded, administrator-only revision-history reads.
- [x] Apply effective title/description/brand overrides to catalog list, detail, search ranking and brand filters; apply the same predicates to result counts and exclude archived products from list/detail/category counts.
- [x] Verify supplier re-imports update source values without replacing editorial overrides.
- [x] Add PostgreSQL integration tests for customer/seller denial, manual creation/offer creation, validation, override/reset behavior, source preservation, stale revision conflicts, archive/restore visibility, history order and immutability.
- [x] Verify Backend CI [41473127094](https://github.com/pooya16pg-spec/globalcart/actions/runs/41473127094) and Security scans [41473127072](https://github.com/pooya16pg-spec/globalcart/actions/runs/41473127072) on implementation head `e288cba2f0abe891656388d6e50638bff5fadc3c`.

**Boundary:** `backend/internal/admin` remains an internal Go service. Manual creation, edit, archive/restore and history are available only to an authenticated active administrator at the service boundary. No admin HTTP routes, browser dashboard, role switcher, seller interface or public catalog write path is enabled.


### M06.3 — Internal administrator product work queue — CI verified (2026-11-03)
- [x] Add `ListManagedProducts` as an internal service guarded by the existing `admin.products.manage` permission and transaction-time account/session/role checks.
- [x] Search effective editorial title, description and brand, plus category, source, source ID and exact product ID; keep supplier provenance filters anchored to the source record.
- [x] Filter active (default), archived or all products; support bounded category/brand/source filters and a fixed allow-list of sort keys/directions.
- [x] Enforce a default page size of 20, maximum page size of 100 and maximum offset of 10,000; return total count and stable ID tie-breaking.
- [x] Add PostgreSQL integration coverage for customer denial, editorial-value matching, source/status filters, archive visibility, sort order, pagination, invalid bounds and NUL rejection.
- [x] Verify final Backend CI [41562708698](https://github.com/pooya16pg-spec/globalcart/actions/runs/41562708698), Security scans [41562708607](https://github.com/pooya16pg-spec/globalcart/actions/runs/41562708607) and Pages deployment [41562708709](https://github.com/pooya16pg-spec/globalcart/actions/runs/41562708709) on `115fbac9dc716b09911151fb19e8e44aeede110f`; race-enabled Go/PostgreSQL tests (including NUL rejection), backup/restore, API/importer builds, Docker image, security scans and published-site smoke checks passed.

**Boundary:** this is a backend service only. It adds no HTTP handler, admin page, browser role switcher, public product endpoint or hosted dependency. No schema migration is required; the existing migrations 013–014 are unchanged.



### M06.4 — Internal seller product work queue — CI verified (2026-11-03)
- [x] Add `ListSellerProductsForSeller` as a read-only service using the existing vendor offer and catalog schema; no migration or dependency is required.
- [x] Treat the requested vendor ID only as a selector; revalidate the active verified account, exact live session, seller role, active vendor and exact membership under row locks in the same transaction as the list.
- [x] Return only the selected vendor's offers and distinguish seller-owned SKU/price/currency/stock/status from effective catalog title/description/brand and original source provenance.
- [x] Add bounded free-text search across effective product/variant fields, category, source, source ID, SKU and product/offer IDs; add status/category/source/brand filters.
- [x] Add fixed allow-listed sorting and stable offer-ID tie-breaking; default page size 20, maximum 100, maximum offset 10,000.
- [x] Add PostgreSQL integration coverage for customer denial, cross-vendor isolation, membership revocation/stale sessions, effective editorial search, filters, sorting, pagination and invalid input.
- [x] Verify Backend CI [41564104334](https://github.com/pooya16pg-spec/globalcart/actions/runs/41564104334), Security scans [41564104323](https://github.com/pooya16pg-spec/globalcart/actions/runs/41564104323) and Pages deployment [41564104379](https://github.com/pooya16pg-spec/globalcart/actions/runs/41564104379) on implementation commit `b9c35829afa14d37c690fb4abad59e617140564e`; all passed, including PostgreSQL race tests, backup/restore, API/importer and Docker builds, security scan and published-site smoke test.

**Boundary:** `backend/internal/commerce` exposes this as an internal read service only. No seller HTTP handler, dashboard, browser role switcher, public seller route, migration, production database or hosting was added. Existing identity and deployment approval gates still apply.

### M06.5 — Internal seller offer create, edit and archive — CI verified (2026-11-03)
- [x] Add migration 015 with a database-maintained offer revision counter, append-only before/after history and protection against revision-history mutation.
- [x] Create seller offers for existing catalog products as drafts and record opening price and stock.
- [x] Revalidate active verified account, exact live session, seller role, active vendor and exact membership under row locks in the same transaction as every write.
- [x] Support revision-checked edits of seller SKU, USD price and stock; write price/inventory history and immutable before/after state atomically.
- [x] Allow sellers to pause or return an offer to draft; self-publish to active is blocked. Archive is terminal.
- [x] Apply revision/audit semantics to the existing authorized seller price and stock methods.
- [x] Add PostgreSQL lifecycle coverage for duplicate offers/SKUs, customer and foreign-vendor denial, stale revisions after system updates, archive and append-only history.
- [x] Verify Backend CI [41571374758](https://github.com/pooya16pg-spec/globalcart/actions/runs/41571374758), Security scans [41571374594](https://github.com/pooya16pg-spec/globalcart/actions/runs/41571374594) and Pages deploy/smoke [41571374873](https://github.com/pooya16pg-spec/globalcart/actions/runs/41571374873) on implementation head `3a9a328fd7d7b1262d73dfefbc2e7a81fa994782`; race-enabled PostgreSQL lifecycle tests, migration/backup/restore, vet, vulnerability scan, API/importer build and Docker build passed.

**Boundary:** internal Go service only. No seller HTTP route, dashboard, self-publish, payout or production hosting. Seller profile and demo analytics are implemented separately in M06.6.

### M06.6 — Internal seller profile and demo analytics — CI verified (2026-11-03)
- [x] Add a vendor-scoped seller workspace read that revalidates the exact active membership, account and session under row locks in the same transaction.
- [x] Return only selected vendor profile fields, offer-state/active-stock aggregates and child-order state counts; do not expose customer identity or line-item details.
- [x] Aggregate paid amounts by currency only from `paid_simulated` child orders, explicitly naming them simulated totals; include a 30-day order count.
- [x] Allow the seller to edit only the vendor display name, using `updated_at` as an optimistic revision token while preserving slug, type and vendor lifecycle state.
- [x] Append an actor-attributed audit event for profile changes in the same transaction.
- [x] Add PostgreSQL coverage for seller/customer/foreign-vendor isolation, profile optimistic concurrency and audit, and a real demo vendor checkout/payment flow reflected in workspace analytics.
- [x] Verify Backend CI [41572747044](https://github.com/pooya16pg-spec/globalcart/actions/runs/41572747044), Security scans [41572747081](https://github.com/pooya16pg-spec/globalcart/actions/runs/41572747081) and Pages deploy/smoke [41572746914](https://github.com/pooya16pg-spec/globalcart/actions/runs/41572746914) on implementation head `fe24119685bfa8487cb8fa2abfb8ad85ed430904`; race-enabled PostgreSQL workspace/profile/demo-checkout tests, migration/backup/restore, vet, vulnerability scan, API/importer and Docker builds, TypeScript/unit tests, production build and published-site smoke all passed.

**Boundary:** internal backend only. No seller HTTP route, dashboard, payout, real-revenue claim, customer-data disclosure or production hosting is added.

## M07 — Search and discovery
- [x] Faceted filters for brand, rating, price and category across the storefront and read-only API.
- [x] Accessible product/brand/category suggestions and keyboard command palette (Ctrl/⌘+K), with Arrow/Enter/Escape controls.
- [x] PostgreSQL full-text search baseline and indexed query path.
- [x] PostgreSQL trigram typo tolerance for product titles and brands, with exact full-text matches ranked first and integration coverage.
- [x] Explainable catalog ranking and recommendation reasons based on documented sample rating/discount signals.
- [x] Browser-local saved searches and search preferences (query, category, brand, price/rating facets and sort), with validation, duplicate detection and bounded storage.
- [x] Explicit empty, initial-loading, live-source, stale-snapshot and local-fixture fallback states with status announcements.

### M07.1 — Catalog facets: category, brand, price and rating — CI verified (2026-11-02)
- [x] Add case/whitespace-normalized exact brand filtering and inclusive minimum/maximum source-price and rating filters to the PostgreSQL catalog API.
- [x] Apply identical filtering predicates to the paginated product list and total-result count.
- [x] Validate decimal syntax, non-negative values, two-decimal precision, database bounds and inverted ranges before database access.
- [x] Add supporting brand, price and rating indexes, OpenAPI query documentation and PostgreSQL integration coverage.
- [x] Add bilingual storefront controls for dynamic brand selection and minimum rating; preserve existing category/price/search filters.
- [x] Make clear-filter and no-results recovery reset every active filter.
- [x] Extract pure client-side filtering logic and add regression tests for combined filters, normalized brand matching and non-mutating behavior.
- [x] Verify TypeScript, Catalog unit tests, production build, sqlc generation, Go vet, PostgreSQL integration/race tests, API/importer builds and Docker image.

**Verification:** latest implementation head `3372040d3cd71cd55d99724033065cf0dfa4537e` passed [Catalog CI push](https://github.com/pooya16pg-spec/globalcart/actions/runs/41365136582), [Catalog CI pull request](https://github.com/pooya16pg-spec/globalcart/actions/runs/41365141727) and [Backend CI pull request](https://github.com/pooya16pg-spec/globalcart/actions/runs/41365141780).

**Boundary:** storefront facets currently run against the local generated catalog snapshot. The Go API and PostgreSQL remain local/unhosted; no public API connection or production deployment is implied.

## M08 — Quality and hardening
- [x] Playwright tests for critical shopping journeys.
- [x] Automated RTL/LTR axe checks, dialog keyboard/focus regression tests and reduced-motion behavior verified in desktop Chromium and Pixel 7 mobile; manual screen-reader, physical-device and broader keyboard review remains open.
- [x] Automated image-loading policy (priority for the main image, lazy loading/async decoding for non-critical images, and reserved media boxes) plus first-party JavaScript/CSS transfer and request budgets; the recently-viewed page is covered by M08.14.
- [ ] Live CDN image timing and real-world Core Web Vitals review; automated CI measurements do not replace published-session and device profiling.
- [x] Automated repository secret scan (Gitleaks), Go dependency vulnerability scan and `go vet` in CI; hosted API penetration/security review remains gated.
- [x] PostgreSQL 17 logical backup/restore and migration-ledger verification in disposable CI; hosted backup policy and operational recovery drills remain gated on provider approval.
- [x] Lockfile-based frontend dependency install (`npm ci`) and post-deploy GitHub Pages smoke test; byte-for-byte export reproducibility is not claimed while the scheduled external catalog snapshot changes.

### M08.10 — Shopping-dialog focus regression coverage (CI verified, 2026-11-02)
- [x] Keep Tab and Shift+Tab navigation inside cart, favorites, quick-view, comparison and checkout dialogs.
- [x] Dismiss dialogs with Escape and restore focus to the invoking control.
- [x] Preserve the original opener when a Command Palette action opens a replacement dialog, including the mobile saved-items route.
- [x] Use touch-appropriate Quick View interaction in mobile browser tests instead of assuming pointer hover.
- [x] Verify implementation head `2182bbb1c55ad265e4264c1a0389fb0353f0010a` before merge: Browser E2E push [41380662904](https://github.com/pooya16pg-spec/globalcart/actions/runs/41380662904) and PR [41380669829](https://github.com/pooya16pg-spec/globalcart/actions/runs/41380669829) passed (28 Playwright tests).
- [x] On the same pre-merge PR head, Catalog CI [41380669723](https://github.com/pooya16pg-spec/globalcart/actions/runs/41380669723), Backend CI [41380669722](https://github.com/pooya16pg-spec/globalcart/actions/runs/41380669722) and Security scans [41380669865](https://github.com/pooya16pg-spec/globalcart/actions/runs/41380669865) passed.

**Boundary:** this is automated keyboard/focus coverage for the tested dialogs and desktop/mobile viewports. Manual screen-reader, physical-device, broader RTL/LTR and reduced-motion review remains open.

### M08.11 — RTL/LTR accessibility and reduced-motion regression coverage (CI verified, 2026-11-02)
- [x] Run axe WCAG 2.0/2.1/2.2 A/AA scans on the storefront in Persian RTL and English LTR.
- [x] Verify `prefers-reduced-motion: reduce` disables smooth scrolling and limits image transitions and overlay animation.
- [x] Correct the catalog stale-snapshot status text to exceed the 4.5:1 normal-text contrast threshold.
- [x] Keep reduced-motion overlay interaction touch-safe on Pixel 7; the check uses the cart drawer rather than a hover-dependent quick-view control.
- [x] Browser E2E run [41383420470](https://github.com/pooya16pg-spec/globalcart/actions/runs/41383420470) passed all 30 desktop/mobile Playwright tests on implementation head `9653fc7d2d03e45f726b7e6a0e47fd7cbd1295dc`.
- [x] Security scans [41383420409](https://github.com/pooya16pg-spec/globalcart/actions/runs/41383420409) and GitHub Pages deployment [41383420345](https://github.com/pooya16pg-spec/globalcart/actions/runs/41383420345) passed on the same head.

**Boundary:** automated scans cover the tested storefront states and desktop/mobile emulation. Manual screen-reader, real-device, full-site RTL/LTR and published URL/deep-link QA remain open.

### M08.12 — PostgreSQL logical backup/restore CI regression (CI verified, 2026-11-02)
- [x] Create a PostgreSQL 17 custom-format `pg_dump` of the disposable integration database after the race-enabled suite completes.
- [x] Restore into a fresh `globalcart_restore` database with `pg_restore --exit-on-error`, excluding runner-specific owner and ACL metadata.
- [x] Compare the ordered `schema_migrations` ledger, product/user/cart/vendor-offer row counts, validated constraint count and `pg_trgm` extension version between source and restored databases.
- [x] Clean up the temporary dump and restore database on both success and failure.
- [x] Verify Backend CI [run 41404292955](https://github.com/pooya16pg-spec/globalcart/actions/runs/41404292955) and Security scans [run 41404293071](https://github.com/pooya16pg-spec/globalcart/actions/runs/41404293071) on implementation commit `0e8c8556fa04fa43d735b46820685632eef2df0a`; both passed.

**Boundary:** this proves logical backup and restore on disposable CI PostgreSQL only. It does not create retained/off-site backups, test encryption or point-in-time recovery, define production RPO/RTO, or replace a provider-backed restore drill after hosted PostgreSQL is approved.

### M08.13 — Locked frontend dependencies and published Pages smoke test (CI verified, 2026-11-02)
- [x] Add and commit the npm lockfile generated from the existing package manifest; keep Node 22 as the CI runtime.
- [x] Make Pages, Catalog CI and Browser E2E use `npm ci` whenever the lockfile exists; retain a one-time bootstrap path for branches without it.
- [x] Persist `package-lock.json` alongside the refreshed catalog snapshot, then validate it with a clean `npm ci` before typecheck, tests and build.
- [x] After Pages deployment, request the published homepage, resolve and request a Next.js JavaScript asset beneath the `/globalcart/` base path, and request one statically generated product deep link selected from the built catalog.
- [x] Verify Pages deployment [run 41405289322](https://github.com/pooya16pg-spec/globalcart/actions/runs/41405289322) and Security scans [run 41405289317](https://github.com/pooya16pg-spec/globalcart/actions/runs/41405289317) on workflow implementation commit `8fec0d0c9bbb845fb7b1dc0788970929828e0a2a`; both passed. The lockfile and generated catalog snapshot were persisted on `main` by `a9aa72606839d04329837c570fd2cbf492499299`.

**Boundary:** the lockfile makes dependency resolution reproducible for the committed manifest and Node/npm environment; it does not promise byte-identical exports because the scheduled public demo catalog can change content and generation time. The smoke test covers the published homepage, one JS asset and one product route, not every route, browser/assistive-technology behavior or live CDN performance.

### M08.14 — Recently viewed image loading and layout regression — CI verified (2026-11-03)
- [x] Apply asynchronous image decoding to recently viewed product cards while retaining lazy loading, accessible alt text and fixed media dimensions.
- [x] Add a desktop/mobile Playwright journey that opens a product, confirms the browser-local visit is recorded, follows the GitHub Pages base path to the history route, and verifies the route response, image loading/decoding attributes, accessible names and reserved layout box.
- [x] Keep the performance audit explicit: the latest Browser E2E run measured the initial route at 520,131 bytes of JavaScript across 7 requests and 70,540 bytes of CSS in 1 request, within the existing 750,000/120,000-byte budgets.
- [x] Verify Browser E2E [41575646034](https://github.com/pooya16pg-spec/globalcart/actions/runs/41575646034) passed all 36 desktop/mobile tests, Security scans [41575646222](https://github.com/pooya16pg-spec/globalcart/actions/runs/41575646222) passed, and Pages deployment plus published homepage/assets/product deep-link smoke test [41575646035](https://github.com/pooya16pg-spec/globalcart/actions/runs/41575646035) passed on implementation head `c358970b31d2272db98434a761c699355a9dbc57`.

**Boundary:** these checks validate browser-emulated loading policy, layout reservation and first-party asset budgets. They do not measure remote image CDN timing, field Core Web Vitals, physical-device performance or screen-reader output; those remain manual QA.


### M08.15 — Published Lighthouse lab baseline (verified, 2026-11-03)
- [x] Add a scheduled and post-deployment Lighthouse workflow using the exact CLI version 13.6.0 and the committed Node 22/Playwright Chromium toolchain.
- [x] Audit the public storefront, representative product detail and bilingual case-study route in mobile and desktop lab modes; preserve raw JSON reports as a private GitHub Actions artifact.
- [x] Summarize Performance, Accessibility, Best Practices, SEO, FCP, LCP, TBT and CLS without conflating lab results with field Core Web Vitals.
- [x] Document manual physical-device, real screen-reader, comprehensive RTL/LTR and field-data acceptance steps in [PERFORMANCE_QA.md](docs/PERFORMANCE_QA.md).
- [x] Complete the first post-deployment Lighthouse run and review all six report results. Latest published lab report: [run 41604368513](https://github.com/pooya16pg-spec/globalcart/actions/runs/41604368513), against main snapshot `2eb11c745757f86ee9d5a3bd4d1fa8561c9e4cf1`.

**Latest recorded result (Lighthouse 13.6.0, GitHub-hosted Chromium; six published lab audits):**

| Published route / mode | Performance | Accessibility | Best practices | SEO | FCP | LCP | TBT | CLS |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Storefront — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.7 s | 10 ms | 0 |
| Storefront — mobile | 87 | 100 | 100 | 100 | 1.2 s | 3.2 s | 250 ms | 0 |
| Product 1 — desktop | 100 | 100 | 100 | 100 | 0.2 s | 0.6 s | 10 ms | 0 |
| Product 1 — mobile | 96 | 100 | 100 | 100 | 0.9 s | 2.6 s | 80 ms | 0 |
| Case study — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.3 s | 0 ms | 0 |
| Case study — mobile | 99 | 100 | 100 | 100 | 0.9 s | 0.9 s | 140 ms | 0 |

Raw JSON artifact: [globalcart-public-lighthouse, run 41604368513](https://github.com/pooya16pg-spec/globalcart/actions/runs/41604368513).

**Measured storefront mobile improvement:** the initial public run recorded Performance 74, LCP 4.9 s and TBT 320 ms. After restoring the mobile single-column hero, using responsive hero image candidates, deferring the noncritical sample-catalog revalidation until idle, and self-hosting DM Sans/Vazirmatn via `next/font`, the latest run recorded Performance 87, LCP 3.2 s and TBT 250 ms. That is a 1.7 s (about 35%) lab LCP reduction and a 13-point Performance score increase. The run environment and emulation are the same general GitHub-hosted Lighthouse lab workflow; repeated runs can still vary.

**Verified code checks:** [Browser E2E run 41603301095](https://github.com/pooya16pg-spec/globalcart/actions/runs/41603301095) passed 44/44 desktop/mobile tests, including mobile hero stacking, absence of external Google Fonts requests, and asset budgets. [Pages deployment 41603301049](https://github.com/pooya16pg-spec/globalcart/actions/runs/41603301049) passed production build and public-route smoke; [Security 41603300997](https://github.com/pooya16pg-spec/globalcart/actions/runs/41603300997) passed.

**Boundary:** Lighthouse desktop/mobile presets are lab emulation. They do not close physical-device, actual screen-reader or CrUX/field-data release gates. No third-party reporting server, API key or production secret is required.

## M09 — Portfolio release
- [x] Persian and English case study (M09.1; automated desktop/mobile RTL/LTR and axe checks passed).
- [x] Architecture/data-flow diagram and bilingual feature tour.
- [x] PostgreSQL ERD and relationship notes derived from migrations 001–015 (27 tables; M09.2).
- [x] Real full-page desktop/mobile screenshots captured from the production static export and committed under `docs/screenshots/` (M09.3).
- [x] README audited against the implemented behavior, public/static hosting boundary and remaining prerequisites.
- [x] Final public demo smoke test: homepage, JavaScript asset, product deep link, and published `/case-study/` with ERD content.
- [ ] Versioned 1.0.0 release and release changelog. Keep gated until the remaining M08 manual device, screen-reader, full-site RTL/LTR and live CDN/Core Web Vitals reviews are completed.

### M09.1 — Bilingual portfolio case study (CI verified, 2026-11-03)
- [x] Add a dedicated statically exportable `/case-study/` route with Persian RTL and English LTR content, reusing the shared guarded locale preference.
- [x] Explain the implemented product journey, real technology stack, catalog/TGJU data flow, local backend boundary and non-transactional demo limits without claiming unshipped infrastructure.
- [x] Add a responsive architecture/data-flow diagram and a feature tour linking to existing storefront journeys.
- [x] Add route metadata and a discoverable link from the storefront footer.
- [x] Verify typecheck, unit tests, production static export, desktop/mobile browser coverage and published Pages smoke. Browser E2E [41577703335](https://github.com/pooya16pg-spec/globalcart/actions/runs/41577703335), Security [41577703300](https://github.com/pooya16pg-spec/globalcart/actions/runs/41577703300) and Pages [41577703347](https://github.com/pooya16pg-spec/globalcart/actions/runs/41577703347) passed on the first implementation head.
- [x] Superseding regression coverage includes the database section and ERD link in both locales; the current suite now contains 40 desktop/mobile scenarios.

**Boundary:** no new commerce behavior, hosted service, supplier integration, translation runtime or live transaction path was introduced. M09.1 does not close M08 manual screen-reader, physical-device, full-site RTL/LTR or live CDN/Core Web Vitals QA.

### M09.2 — PostgreSQL ERD and public project documentation (CI verified, 2026-11-03)
- [x] Inspect all 15 append-only PostgreSQL migrations (001–015) and derive the table/foreign-key model from the SQL source of truth.
- [x] Document all 27 tables in domain-grouped Mermaid ERDs: catalog/marketplace, demo commerce, identity/access and editorial/seller revision history.
- [x] Record critical composite foreign keys, optional guest/account cart ownership, vendor-order parent/child invariants, append-only audit/revision history and the boundary between local DB state and browser-local state.
- [x] Add a responsive four-domain database summary and a direct ERD link to the bilingual case-study page.
- [x] Add desktop/mobile RTL/LTR accessibility regression assertions for the new database section and ERD link.
- [x] Verify TypeScript, unit tests, static build, desktop/mobile browser E2E (40 passed), Security scans and published Pages smoke. Browser E2E [41596731945](https://github.com/pooya16pg-spec/globalcart/actions/runs/41596731945), Security [41596731890](https://github.com/pooya16pg-spec/globalcart/actions/runs/41596731890), Pages [41596731889](https://github.com/pooya16pg-spec/globalcart/actions/runs/41596731889).

**Source:** [docs/DATABASE_ERD.md](docs/DATABASE_ERD.md). The document describes the schema declared by migrations; it does not claim a production database is deployed. No SQL migration, database table, runtime dependency or API endpoint was changed.

**Boundary:** M09.2 documents existing schema only. It does not expose the internal admin/seller/commerce services over HTTP or imply live payment, fulfillment or inventory operations. M08 manual screen-reader, physical-device and live CDN/Core Web Vitals QA remains open.

### M09.3 — Screenshot evidence and README accuracy (completed, 2026-11-03)
- [x] Add a Playwright capture test that takes full-page Persian screenshots from the production static export in desktop and mobile Chromium projects.
- [x] Publish the verified screenshot artifact to `docs/screenshots/case-study-desktop.png` and `docs/screenshots/case-study-mobile.png` through the guarded [Publish portfolio screenshots workflow](.github/workflows/publish-portfolio-screenshots.yml).
- [x] Confirm the capture suite: 40/40 browser scenarios passed on [run 41596731945](https://github.com/pooya16pg-spec/globalcart/actions/runs/41596731945); screenshot publication [run 41597084953](https://github.com/pooya16pg-spec/globalcart/actions/runs/41597084953) succeeded and committed the two assets.
- [x] Audit README feature claims, move completed search/command-palette/recommendation work into the implemented list, remove the unimplemented Python-worker claim, and document the true local-only backend boundary.
- [x] Extend the Pages deployment smoke test to fetch the published `/case-study/` URL and assert its main heading, database section and ERD link. The final Pages smoke [41596731889](https://github.com/pooya16pg-spec/globalcart/actions/runs/41596731889) passed.
- [x] Update AI handoff, project context, release process and changelog to match verified main.

**Release gate:** do not tag 1.0.0 until M08 manual real-device, screen-reader, comprehensive RTL/LTR and live CDN/Core Web Vitals QA are performed. Browser emulation is useful regression coverage, not a substitute for those checks.

### M09.4 — Stable 1.0.0 release gate (blocked by manual M08 QA)
- [ ] Verify the storefront on at least one physical desktop/laptop browser and one physical mobile browser; record browser/device/viewport and any regressions.
- [x] Add automated axe/WCAG regression coverage to the public recently viewed route in populated and empty states in Persian RTL and English LTR, alongside the storefront, product detail, case study and key dialogs. Chromium desktop/mobile passed 48/48 tests on [run 41606569768](https://github.com/pooya16pg-spec/globalcart/actions/runs/41606569768).
- [x] Extend the accessibility suite to Firefox and WebKit desktop while retaining the full Chromium desktop/mobile suite. Cross-browser Browser E2E [41607322346](https://github.com/pooya16pg-spec/globalcart/actions/runs/41607322346) passed **66/66**: 48 Chromium scenarios + 9 Firefox + 9 WebKit accessibility scenarios. Pages [41607322372](https://github.com/pooya16pg-spec/globalcart/actions/runs/41607322372) and Security [41607322262](https://github.com/pooya16pg-spec/globalcart/actions/runs/41607322262) also passed on the same implementation commit.
- [ ] Complete an actual keyboard and screen-reader review of primary navigation, dialogs/drawers, product detail, checkout simulation and the bilingual case-study route.
- [ ] Perform a manual RTL/LTR pass across the full public site, including narrow screens, text wrapping, menus, focus order and persisted language preference.
- [ ] Measure published CDN response and field/lab Core Web Vitals; record methodology, URL, date and observed values. Do not infer field metrics from Playwright timings.
- [ ] Resolve and document any findings, rerun the relevant automated checks, then update package version, changelog, roadmap and project context together and create the 1.0.0 release/tag.

**Why this is still open:** the automated CI suite runs Chromium desktop/mobile emulation and axe checks; it cannot physically operate a user's device or verify the actual assistive-technology experience and field CDN metrics. Keep the release at `0.2.3-dev` until these checks are genuinely performed.

## Handoff rule
Every AI/developer must read PROJECT_CONTEXT.md, this roadmap, CHANGELOG.md, docs/AI_HANDOFF.md, docs/ARCHITECTURE.md and docs/DECISIONS.md before changing code. Continue from the earliest incomplete item, inspect current Actions, and update documentation only with verified facts.

## Release ledger
| Version | Stage | Status | Notes |
|---|---|---|---|
| 0.1.0 | M00 + M01 | Implemented | Initial independent storefront |
| 0.1.1 | Pricing foundation | Implemented | Integer-cent and fixed-point demo pricing |
| 0.1.4 | M02 catalog pipeline | Implemented | Catalog import and snapshot pipeline; translation bridge parked and removed from CI |
| 0.2.0-dev | M01.5 marketplace redesign | In progress | Category-first shell and updated marketplace visual system |
| 0.2.1-dev | M03 product experience | In progress | Static product detail routes, gallery, buy box and related products added; verification pending |


## 0.2.3-dev verification update — 2026-11-01
- [x] Data-driven homepage merchandising rails for deals, top-rated products and discoveries.
- [x] Browser-local recently viewed history, including visits to dedicated product pages.
- [x] Product price breakdown using the existing fixed-point pricing helper.
- [x] GitHub Actions run 41299322140 passed catalog refresh, TypeScript, unit tests, production build, Pages artifact upload and deployment.
- [ ] Full manual responsive, accessibility and published deep-link QA remains open.


## Accessibility follow-up — 2026-11-01
- [x] Added Escape-key dismissal for the homepage drawers, quick view, comparison, checkout, category menu and mobile navigation.
- [x] Added dialog semantics, modal state and heading associations for cart, favorites, quick view, compare and demo checkout.
- [x] Added expanded/control relationships for mobile navigation and category navigation.
- [x] Exposed selected product-gallery thumbnail state to assistive technology.
- [x] Added keyboard focus trapping and focus restoration for modal dialogs.
- [ ] Manual keyboard, screen-reader and browser QA remains open.

- [x] Focus is moved into each modal on open, trapped while active and returned to the invoking control on close.


### M04 catalog schema extension — 2026-11-01
- [x] Add migration 002 for normalized category records, product category foreign keys and product variants.
- [x] Backfill categories from existing product rows without changing the existing catalog API response shape.
- [x] Enforce non-negative variant price/stock, object-shaped JSONB attributes, product ownership, optional unique SKU and source identity.
- [x] Extend PostgreSQL integration tests to verify category backfill and variant attribute persistence.
- [ ] Run the backend CI for this migration before marking the slice verified.


### M04 API validation pass — 2026-11-01
- [x] Reject malformed/out-of-range limit and offset values instead of silently replacing them.
- [x] Bound free-text query and category filter lengths.
- [x] Add HTTP tests proving invalid requests return 400 before database access.
- [x] Run backend CI for the validation changes before marking verified.


### M04 typed SQL generation — 2026-11-01
- [x] Add root sqlc v2 configuration targeting PostgreSQL and pgx/v5.
- [x] Define typed list/detail product queries and active product-variant query with parameterized inputs.
- [x] Add deterministic sqlc generation to backend CI.
- [x] Verify generated package compiles in CI and wire it into API/database code.


### M04 sqlc API integration — 2026-11-01
- [x] Replace handwritten product list/detail read queries in HTTP handlers with generated pgx/v5 sqlc methods.
- [x] Keep the external catalog JSON response shape stable through an explicit database-row mapper.
- [x] Add mapper tests for JSON images and catalog field mapping.
- [x] Verify sqlc generation, go vet, race-enabled unit/integration tests and Go builds in CI.
- [x] Complete read API for active product variants and category listing.


### M04 product variants read API — 2026-11-01
- [x] Add `GET /api/v1/products/{id}/variants` using generated sqlc product existence and active-variant queries.
- [x] Return a stable variant response with SKU, price, stock and decoded JSON attributes.
- [x] Distinguish invalid product IDs and missing parent products.
- [x] Add mapper tests and OpenAPI endpoint documentation.
- [x] Verify this endpoint and generated variant query in CI.


- CI verification for the product variants read endpoint completed successfully (run 41302565323): sqlc generation, go vet, race-enabled tests, API/importer builds and Docker image build all passed.


### M04 normalized category API — 2026-11-01
- [x] Add generated sqlc category listing query with parent relationship and product counts.
- [x] Expose `GET /api/v1/categories` as a read-only endpoint.
- [x] Add OpenAPI documentation and response serialization coverage.
- [x] Extend PostgreSQL integration tests to exercise the generated category query against migrated schema.
- [x] Verify category query and PostgreSQL integration test in CI.


Category API CI verification: run 41302783939 passed sqlc generation, formatting, vet, unit and PostgreSQL integration tests, API/importer builds, and Docker image build.


### M04 catalog pagination and normalized filtering — 2026-11-01
- [x] Normalize category filters against the category table using case-insensitive, whitespace-trimmed matching.
- [x] Add generated `CountProducts` query using the same category and full-text predicates as the product list.
- [x] Return `total` independently from the current page's `count`.
- [x] Add PostgreSQL integration coverage for normalized category matching and total count.
- [x] Verify generated count query and API pagination changes in CI.


Catalog pagination CI verification: pull-request run 41303108214 passed sqlc generation, formatting, vet, PostgreSQL-backed race tests, API/importer builds and Docker image build. Push run 41303144768 also successfully generated and persisted the updated typed query files; final image build is completing.


### M05 commerce schema foundation — 2026-11-01
- [x] Add migration 003 for guest carts, cart items, demo orders, immutable order lines, inventory reservations and order events.
- [x] Enforce digest lengths, bounded quantities, nonnegative price snapshots, valid statuses, variant/product ownership and cart-scoped idempotency.
- [x] Specify transactional checkout locking, server-side price revalidation, reservation expiry and replay semantics in `docs/COMMERCE_DESIGN.md`.
- [x] Verify migration 003 and PostgreSQL integration tests in CI (run 41345512979).
- [x] Implement domain transaction service for demo checkout, canonical price snapshots, row-locked stock reservations and idempotent replay.
- [ ] Expose persistent cart/order endpoints only after same-site HTTPS, identity/authorization and CSRF prerequisites are approved.


### M05 transactional checkout service — 2026-11-01
- [x] Implement internal Go checkout service; deliberately not exposed as an HTTP write route.
- [x] Lock guest cart and catalog stock rows, revalidate canonical price/title/stock, and write order snapshots/reservations/events in one transaction.
- [x] Use database time for cart expiry and expiring reservations; store only caller-provided SHA-256 digests.
- [x] Add PostgreSQL concurrency test: two carts competing for limited product stock yield one success and one insufficient-stock result.
- [x] Add idempotent replay test: repeated checkout returns the original order without duplicate reservations.
- [x] Implement and integration-test idempotent expiry cleanup for pending demo orders and their active reservations.
- [x] CI run 41346127244 passed sqlc generation, formatting, vet, PostgreSQL-backed tests (including reservation-expiry cleanup), API/importer builds and Docker image build.
- [x] Implement internal guest cart lifecycle operations (create, add/increment, set quantity, remove, read) with catalog price revalidation.
- [ ] Implement safe demo order read APIs after authorization boundary is ready.


### M05 guest cart lifecycle service — 2026-11-01
- [x] Implement internal cart creation using 32-byte SHA-256 capability digests; never persist raw bearer tokens.
- [x] Implement add/increment, set quantity, remove line and retrieve cart operations.
- [x] Serialize mutations per cart using row locks; validate lifecycle state and expiry.
- [x] Read canonical product/variant title, price and stock from PostgreSQL; never trust client-supplied prices.
- [x] Explicitly constrain carts to USD until catalog currency/FX support is designed.
- [x] Add PostgreSQL integration tests for cart creation, add/increment, quantity updates, stock bounds, removal, variants and expired-cart rejection.
- [x] CI push run 41346723656 passed sqlc generation, formatting, vet, PostgreSQL-backed tests, API/importer builds and Docker image build.
- [ ] Keep all cart operations internal until same-site HTTPS, CSRF and identity/authorization prerequisites are approved.


### M06 order read and cancellation services — 2026-11-01
- [x] Implement capability-scoped order detail reads, returning immutable item snapshots and ordered event history.
- [x] Hide order existence behind the same not-found error for unknown IDs and wrong cart capability digests.
- [x] Implement owner-scoped cancellation for unpaid demo orders; atomically release active reservations and append a cancellation event.
- [x] Make repeated cancellation idempotent; reject cancellation of paid or expired orders.
- [x] Add PostgreSQL integration coverage for order ownership, snapshots, event history, cancellation, reservation release and replay.
- [x] CI push run 41347325289 passed sqlc generation, formatting, vet, PostgreSQL-backed tests, API/importer builds and Docker image build.
- [ ] Keep order read/cancel methods internal until same-site HTTPS, CSRF and identity/authorization prerequisites are approved.


### M05 simulated payment lifecycle — 2026-11-02
- [x] Add internal `CompleteDemoPayment` transition from `awaiting_demo_payment` to `paid_simulated`.
- [x] Require the owning cart capability; unknown and wrong-owner order IDs share the same not-found response.
- [x] Require a live reservation before simulated completion; expired or non-pending orders cannot be completed.
- [x] Atomically release demo reservations, update order state and append one simulation event.
- [x] Make repeat completion idempotent; no payment provider, credentials, webhook or real-money flow exists.
- [x] Add PostgreSQL integration coverage for ownership, payment replay, reservation release and expired-reservation rejection.
- [x] Verify sqlc generation, formatting, vet, race-enabled PostgreSQL tests, API/importer builds and Docker image build in CI run 41348079983.
- [x] Keep all payment/order operations internal until the deployment, identity and CSRF gates are approved.


### M07 internal coupon and discount foundation — 2026-11-02
- [x] Add hashed coupon-code storage and percentage/fixed discount rules.
- [x] Support start/end dates, minimum subtotal, maximum discount, global usage limit and guest-capability usage limit.
- [x] Attach/remove coupons on open carts and expose only a boolean applied-state in cart reads.
- [x] Revalidate coupon eligibility and limits under a row lock during checkout; calculate money using PostgreSQL NUMERIC.
- [x] Snapshot subtotal, discount and final total on orders and expose those totals in owner-scoped order reads.
- [x] Add PostgreSQL integration coverage for normalized codes, percentage calculation, usage exhaustion, minimum subtotal rollback, cart attachment state and order detail totals.
- [ ] Add verified customer identity before treating guest-capability limits as true per-customer limits.
- [ ] Keep coupon administration and cart/checkout operations internal until auth, same-site deployment and CSRF prerequisites are approved.


### M08 — Multi-vendor foundation (backend-only; migration 005)
- [x] Add vendor records with platform/merchant type and constrained lifecycle states.
- [x] Add vendor-specific base-product and variant offers with independent price, USD currency, stock, SKU and status.
- [x] Enforce one offer per vendor/product-or-variant and vendor-scoped SKU uniqueness.
- [x] Backfill current products and variants into the explicit GlobalCart Direct platform vendor.
- [x] Add append-only price history and inventory event tables with initial import/opening balance records.
- [x] Add PostgreSQL integration assertions for vendor backfill, independent stock across vendors, audit records and duplicate-offer rejection.
- [x] Integrate catalog importer so newly imported base products and supplied variants create/update GlobalCart Direct platform offers transactionally; price/inventory audit events are appended only for actual changes (Backend CI [41404292955](https://github.com/pooya16pg-spec/globalcart/actions/runs/41404292955) passed).
- [x] Implement internal vendor and offer services with transactional price/inventory audit writes (covered by `TestVendorServiceTransactions`; Backend CI [41404292955](https://github.com/pooya16pg-spec/globalcart/actions/runs/41404292955) passed).
- [x] Migrate cart lines, reservations and checkout to vendor-offer identity; split each checkout into vendor order groups (migration 006, vendor cart lifecycle and split-checkout PostgreSQL tests; Backend CI [41404292955](https://github.com/pooya16pg-spec/globalcart/actions/runs/41404292955) passed).
- [ ] Add verified seller membership and tenant-scoped authorization only after the approved identity/authentication boundary.
- [ ] Keep seller APIs, seller UI, commission, payouts and fulfillment disabled until their separate security and product requirements are approved.

**Compatibility note:** migration 005 adds vendor/offer storage; migration 006 adds vendor-aware cart lines and parent/child order references. Internal vendor split checkout, per-vendor reservations and child-order state propagation are implemented and CI-verified. Public seller/customer HTTP routes remain disabled pending the approved identity, same-site hosting and CSRF boundary.


### M09 — Vendor-aware cart foundation
- [x] Add vendor offer identity to cart lines without changing legacy line ownership.
- [x] Support internal add, quantity update and removal of vendor-specific offers.
- [x] Snapshot canonical offer price and vendor identity; serialize cart/offer mutations with row locks.
- [x] Add parent/child vendor order schema and vendor-aware order-item/reservation references.
- [x] Prevent legacy checkout from mispricing or reserving vendor cart lines.
- [x] Implement atomic split checkout, per-vendor reservations and coupon allocation (M09.2; PostgreSQL integration and competing-checkout race tests passed in Backend CI [41404292955](https://github.com/pooya16pg-spec/globalcart/actions/runs/41404292955)).
- [x] Propagate cancellation, expiry and simulated payment state to vendor child orders (parent/child lifecycle SQL and simulated-payment integration coverage; Backend CI [41404292955](https://github.com/pooya16pg-spec/globalcart/actions/runs/41404292955)).


### M09.2 — Vendor split checkout
- [x] Add internal atomic vendor checkout with one parent order and one child order per vendor.
- [x] Reserve inventory against vendor offers, without mutating legacy catalog stock.
- [x] Revalidate vendor, offer, variant, currency and available stock under deterministic row locks.
- [x] Apply coupon redemption once at parent level and allocate discounts across child orders with cent-exact reconciliation.
- [x] Propagate simulated payment, cancellation and reservation expiry to child order states.
- [x] Add PostgreSQL integration coverage for split totals, idempotency, coupon allocation, simulated payment and competing checkouts.
- [x] Add owner-scoped split-order detail reads and vendor-specific internal query methods (verified in M09.3).
- [ ] Add seller authorization/membership model before any seller-facing APIs.


### M09.3 — Owner-scoped vendor child-order reads
- [x] Add `GetVendorOrder` to retrieve one vendor child order and its immutable item snapshots using the owning cart capability digest.
- [x] Conceal cross-owner child-order existence behind the same `ErrOrderNotFound` result used for unknown IDs.
- [x] Add PostgreSQL integration coverage for child/parent identity, vendor metadata, exact totals, item ownership, invalid capabilities and cross-owner isolation.
- [x] Backend CI run `41352898725` passed the vendor child-order integration tests and Docker image build.


### M09.3 CI correction checkpoint — 2026-11-02
- [x] Inspect failed Backend CI run 41352752412 and retrieve the exact job log.
- [x] Identify the blocking `go vet` error: unused `time` import in `backend/internal/commerce/vendor_orders.go`.
- [x] Remove the unused import in commit `cb0f20816afdbf3f2dc23628e8a5bd4b7d1bce46`.
- [x] Fresh Backend CI run `41352898725` on corrected code passed; the earlier failed run tested the previous commit only.
- [x] M09.3 marked complete after fresh CI verification; seller membership implementation remains gated by deployment prerequisites.


## M09.4 — Seller authorization boundary and vendor ownership hardening — CI verified
- [x] Add wrong-vendor negative integration checks for offer read, price update and inventory adjustment.
- [x] Verify a foreign vendor scope cannot mutate the actual owner's stock.
- [x] Document that `vendorID` is an internal selector, not caller authorization; seller operations must not be exposed over HTTP until verified session + active membership resolution exists.
- [x] Backend CI run `41353137589` on code commit `e18a01d` passed: SQL generation, formatting, `go vet`, race-enabled tests, API/importer builds and Docker image build.

## M09.5 — Seller membership implementation readiness (blocked)
- [ ] Select and approve custom frontend domain + same-site HTTPS API origin.
- [ ] Approve runtime/hosting, region, budget and secret-management approach.
- [ ] Select and approve email delivery provider before verification/reset flows.
- [ ] Then implement users, seller memberships, role grants, sessions, CSRF and audit events according to `docs/SECURITY_AUTH_DESIGN.md`.


## M09.4.1 — Public API surface regression guards (in progress)
- [x] Add handler tests proving seller order/product routes, cart writes and checkout writes remain unavailable over HTTP.
- [x] Add tests proving catalog routes reject mutation methods and allowed CORS origins receive only GET/OPTIONS.
- [x] Verify these handler tests and pre-auth CORS restrictions in Backend CI run `41353463006` (success).


## M08.1 — Platform catalog importer synchronization (CI verified)
- [x] Extend the existing catalog upsert transaction to create/update the `GlobalCart Direct` base-product offer.
- [x] Synchronize every imported product variant into a matching platform vendor offer, including SKU, price, stock and active/paused state.
- [x] Record initial price and opening inventory audit events; record subsequent price/stock changes with `system_sync` / `correction` audit reasons.
- [x] Keep product/category writes and platform-offer synchronization in one database transaction.
- [x] Make repeated imports idempotent without duplicate offer or audit rows.
- [x] Add PostgreSQL integration coverage for base-product initial import, unchanged re-import and changed price/stock synchronization.
- [x] Add PostgreSQL integration coverage for variant offer creation, idempotent re-import, price/stock changes and inactive-variant pausing.
- [x] Verify the importer integration slice in Backend CI, including Docker image build.

- [x] Verify base-product and existing-variant offer synchronization in Backend CI run `41354208659` (all steps passed, including PostgreSQL tests and Docker image build).


## M08.2 — End-to-end source variant ingestion (CI verified)
- [x] Extend the TypeScript catalog contract and normalization layer to retain optional source variants, stable source identity, SKU, attributes, price, stock and active state.
- [x] Extend the Go catalog model and PostgreSQL importer to upsert supplied variants transactionally before synchronizing platform offers.
- [x] Enforce source identity ownership so an existing variant identity cannot silently move to another product; reject invalid prices, negative stock and missing identity.
- [x] Add catalog normalization and PostgreSQL integration coverage using variants supplied in the imported product snapshot.
- [x] Verify the complete Go workflow, PostgreSQL integration tests and Docker image build in Backend CI run `41354535515`.
- [x] Keep variant ingestion optional for product-only sources; DummyJSON remains a sample feed without variants, and no live supplier inventory is claimed.


## M08.3 — Supplier adapter contract and fixture-backed variant ingestion (CI verified)
- [x] Define a shared source adapter contract: each adapter fetches its own raw records and maps each record into the canonical catalog input shape.
- [x] Add a reusable ingestion boundary that rejects empty, invalid and wholly unmappable feeds before snapshot replacement.
- [x] Route the existing DummyJSON adapter and scheduled snapshot sync through the shared ingestion boundary while preserving retries and last-known-good behavior.
- [x] Scope variant deduplication to the database identity pair `(source, source_id)`, allowing equal IDs from different suppliers without dropping valid rows.
- [x] Encode the source/ID tuple unambiguously and add a regression test for delimiter-containing values; verified by Catalog CI run `41355272122`.
- [x] Add a clearly synthetic supplier fixture with base product and variant mapping, including attributes, SKU, stock and inactive state.
- [x] Add tests for adapter mapping, provenance, source-scoped identity and empty/unmappable feeds.
- [x] Add a dedicated Catalog CI workflow for TypeScript validation, unit tests and production build on pushes and pull requests.
- [x] Verify the complete Catalog CI workflow on the feature branch and pull-request head.
- [x] Keep live supplier onboarding gated; do not manufacture data or add credentials without owner-approved API/feed access.

**M08.3 verification:** Catalog CI push run `41355090849` and PR run `41355095558` passed on head `9bfab0c80dec906f52fb56a22b85d36b7eec99f9`. Follow-up collision-hardening run `41355272122` also passed on latest code head `eaebbc0a7b8d1c823c3dcad4a7fed0d10a99edc9`, including TypeScript validation, all unit tests and production build.\n\n**M08.3 safety boundary:** the fixture is synthetic test data only. No supplier endpoint, API key, live offer or commercial inventory feed has been connected.


## M08.8 — Authorized supplier onboarding (blocked pending owner decision)
- [ ] Select a supplier and confirm the account has documented API/feed access rights and permitted catalog/image/price reuse.
- [ ] Review supplier API documentation, rate limits, pagination, delta/full-snapshot semantics and variant identity guarantees.
- [ ] Implement the supplier-specific fetch and mapping adapter using the M08.3 shared contract.
- [ ] Add source-specific fixture and contract tests for pagination, partial failures, duplicate IDs, SKU changes, inactive variants and omitted variants.
- [ ] Verify end-to-end ingestion through snapshot generation, Go importer, PostgreSQL variant persistence and platform offer audit events.
- [ ] Store credentials only in an approved server-side secret store; never commit credentials or expose them to the browser.

**Gate:** do not connect a real supplier, scrape retail pages, generate placeholder inventory or treat the synthetic M08.3 fixture as live data until the owner approves a named supplier and its authorized feed/API.


## M08.4 — Catalog snapshot continuity and quality diagnostics (CI verified)
- [x] Add a full-snapshot continuity guard: reject a refresh that loses more than 50% of the previous known-good product count.
- [x] Run the guard before writing the new snapshot so the existing last-known-good file is retained on suspicious drops.
- [x] Add a compact quality report for product/variant counts, products without variants, zero price/stock, missing descriptions and uncategorized items.
- [x] Add unit tests for exact threshold behavior, invalid thresholds and quality report counts.
- [x] Verify TypeScript validation, all unit tests and production build in Catalog CI PR run `41355595494` on head `0df4c5485e4a2eb391b61e13c52047ed4631942a`.
- [ ] Revisit source-specific continuity policy when a real supplier is approved; incremental/delta feeds must declare their semantics and must not use the full-snapshot drop guard blindly.

**M08.4 verification:** Catalog CI PR run `41355595494` and push run `41355552153` passed TypeScript validation, all unit tests and production build. Backend CI PR run `41355595650` also passed SQL generation, formatting, vet, tests, API/importer build, Docker image build and cleanup on the documented branch head.\n\n**M08.4 safety boundary:** this guard applies to the current full-snapshot catalog sync. Product-only feeds remain valid; absence of variants is not treated as a catalog failure and does not delete previously imported variants/offers.


## M08.5 — Variant omission and inventory retention semantics (in progress)
- [x] Audit importer behavior when a later product snapshot contains no variants: absent rows are not deleted from `product_variants`; platform offer sync reads the retained database rows.
- [x] Add PostgreSQL integration coverage importing a product-only snapshot after a prior variant snapshot.
- [x] Assert the source variant remains stored and platform offer price/stock/status remain unchanged.
- [x] Assert an independent merchant offer remains active and unchanged when the platform catalog importer receives a product-only snapshot.
- [x] Assert no duplicate price-history or inventory events are created by the omission-only import.
- [x] Verify the PostgreSQL omission regression, race-enabled Go suite and Docker build in Backend CI PR run `41356201201` on head `eca6f936d99582a83930f2040a46b912bba57225`.

**M08.5 safety rule:** absence is not deletion. A source must explicitly communicate a complete variant set and an authorized retirement signal before any future variant/offer retirement workflow is introduced. This stage does not add deletion or stock-zero behavior.


**M08.6 gate:** real supplier integration requires an owner-approved supplier name, authorized API/feed access, documented terms, and source-specific pagination/snapshot semantics. Do not scrape retail pages or invent live inventory.


## M08.6 — Variant source identity ownership and batch rollback (in progress)
- [x] Add a PostgreSQL integration test proving an existing `(source, source_id)` cannot be reassigned to a different product.
- [x] Include a valid earlier product in the same import batch and assert the identity conflict rolls back that earlier write too.
- [x] Assert the original variant owner, SKU, price and stock remain unchanged and only one row exists for the source identity.
- [x] Verify `TestVariantSourceIdentityCannotMoveBetweenProducts`, the complete race-enabled Go integration suite and Docker image build in Backend CI [41404292955](https://github.com/pooya16pg-spec/globalcart/actions/runs/41404292955).

**M08.6 safety boundary:** source identity is immutable ownership, not merely a deduplication key. Conflicts fail the transaction; they never silently move a variant between products.


**M08.7 gate:** real supplier integration requires an owner-approved supplier name, authorized API/feed access, documented terms, and source-specific pagination/snapshot semantics. Do not scrape retail pages or invent live inventory.


## M08.7 — Explicit feed completeness semantics (CI verified)
- [x] Require every source adapter to declare whether it provides a complete snapshot or incremental/delta records (`snapshotMode` is mandatory; the scheduled snapshot writer requires `full` and rejects incremental mode before fetch; Catalog CI [41356770768](https://github.com/pooya16pg-spec/globalcart/actions/runs/41356770768) and Backend CI [41356770801](https://github.com/pooya16pg-spec/globalcart/actions/runs/41356770801) passed).
- [x] Keep the static catalog snapshot writer restricted to complete snapshots; reject incremental feeds before fetching or replacing the last-known-good file.
- [x] Add contract tests proving incremental records can be normalized for a merger but cannot be written as a full snapshot.
- [x] Verify Catalog CI PR run `41356770768` and Backend CI PR run `41356770801` on head `2176ad14a30f196f4ae1624aa4a17dbac69dc9bd`; both passed all jobs.

**M08.7 safety boundary:** no automatic merging or deletion is implied by an incremental feed declaration. A source-specific delta merger must be designed and tested before incremental records can change the canonical snapshot.


**M08.8 gate:** real supplier integration requires an owner-approved supplier name, authorized API/feed access, documented terms, and source-specific pagination/snapshot semantics. Do not scrape retail pages or invent live inventory.


**M08.7 verification:** Catalog CI PR run `41356770768` and Backend CI PR run `41356770801` both passed on head `2176ad14a30f196f4ae1624aa4a17dbac69dc9bd`. TypeScript validation, unit tests, production build, sqlc generation, Go vet, PostgreSQL race tests, API/importer builds and Docker image build all succeeded.


## M08.8 — Feed mapping quality and request safety — CI verified
- [x] Reject a feed when more than 20% of raw records fail source-specific mapping by default.
- [x] Allow a source-specific threshold override; validate finite values in [0, 1).
- [x] Preserve the distinct error for empty feeds and feeds with no mappable records.
- [x] Enforce a 20-second timeout for each DummyJSON request, including when a caller supplies its own AbortSignal.
- [x] Propagate caller cancellation immediately and do not retry explicitly cancelled requests.
- [x] Retry network/time-out failures and HTTP 408, 429 and 5xx responses with a bounded three-attempt policy.
- [x] Fail fast on permanent HTTP errors, invalid JSON/schema and empty source collections.
- [x] Verify Catalog CI and Backend CI on the same PR head.
- [x] Keep supplier onboarding separate and gated by owner approval.

**Verification:** Catalog CI [41358956297](https://github.com/pooya16pg-spec/globalcart/actions/runs/41358956297) and Backend CI [41358956380](https://github.com/pooya16pg-spec/globalcart/actions/runs/41358956380) succeeded on head `a4a0d16c2aa06cfc83558b6516c9ac282cec1701`.

## M08.9 — Authorized supplier onboarding (blocked)
- Select a named supplier and confirm API/feed access rights and catalog/image/price reuse terms.
- Review authentication, rate limits, pagination, full-vs-delta semantics, stable product/variant identity and retirement signals.
- Implement only the approved adapter and source-specific failure/identity tests.
- Keep credentials server-side; do not scrape or invent live inventory.


## M08.8.1 — Source-agnostic feed import diagnostics — CI verified
- [x] Expose aggregate fetched, mapped, rejected, normalized and post-normalization-discard counts without retaining raw supplier payloads.
- [x] Include source name, snapshot mode, import timestamp and the configured mapping rejection threshold.
- [x] Preserve the existing products-only loader API for backward compatibility.
- [x] Emit the aggregate report during scheduled catalog snapshot generation.
- [x] Add regression coverage for mapper rejection, malformed normalized rows and duplicate product IDs.
- [x] Verify Catalog CI and Backend CI on the same code head.

**Verification:** Catalog CI [41359432556](https://github.com/pooya16pg-spec/globalcart/actions/runs/41359432556) and Backend CI [41359432659](https://github.com/pooya16pg-spec/globalcart/actions/runs/41359432659) succeeded on code head `29956b0423328bcf9ce28c8fbe83d391aff22a98`.


## M08.8.2 — Collision-safe atomic snapshot publication — CI verified
- [x] Extract snapshot persistence into a shared helper that serializes before writing and renames only after the complete file has been written.
- [x] Give every writer a unique, exclusively-created temporary file in the destination directory.
- [x] Clean up temporary files after write/rename failures without changing the last-known-good destination.
- [x] Add concurrent-writer, failed-rename cleanup and serialization-failure regression tests.
- [x] Verify Catalog CI, production build and Backend CI on the same code commit.

**Verification:** Catalog CI [41360183218](https://github.com/pooya16pg-spec/globalcart/actions/runs/41360183218) and Backend CI [41360183283](https://github.com/pooya16pg-spec/globalcart/actions/runs/41360183283) succeeded on code head `062aabdf563626116a50ecba0d7f534e6af3f260`.

**Safety note:** atomic rename protects readers from partial JSON, but it does not impose freshness ordering between independent writers; the last successful rename wins. The scheduled Pages workflow remains serialized.


## M08.8.3 — Last-known-good snapshot validation — CI verified
- [x] Treat only ENOENT as a missing first-run snapshot; propagate filesystem permission and I/O failures.
- [x] Validate snapshot metadata, product count and normalized product structure before using a file as last-known-good data.
- [x] Reject malformed JSON and structurally invalid snapshots instead of silently treating them as absent and switching to fixtures.
- [x] Add regression tests for missing files, malformed JSON, count mismatch and malformed product rows.
- [x] Verify Catalog CI, production build and Backend CI on the same code head.

**Verification on implementation commit `ff5cf42cc094bc4e35d317ad52f970762280ba2f`:** Catalog CI PR run [41361214644](https://github.com/pooya16pg-spec/globalcart/actions/runs/41361214644), Catalog CI push run [41361210412](https://github.com/pooya16pg-spec/globalcart/actions/runs/41361210412), and Backend CI PR run [41361214721](https://github.com/pooya16pg-spec/globalcart/actions/runs/41361214721) all succeeded.

**Safety note:** a missing snapshot may use the documented first-import behavior; an existing but invalid snapshot now fails closed and requires repair rather than being silently discarded.


## M08.8.4 — Snapshot identity and database-bound validation — CI verified
- [x] Validate product IDs and source identities are unique across the complete snapshot.
- [x] Require all product records to match the snapshot's declared source.
- [x] Reject duplicate variant `(source, source_id)` identities across products before the PostgreSQL importer can encounter a batch conflict.
- [x] Reject duplicate non-empty variant SKUs across the snapshot, matching PostgreSQL's partial unique SKU index.
- [x] Enforce database-compatible product/variant price precision and bounds, percentage/rating ranges, safe product IDs and PostgreSQL integer stock bounds.
- [x] Assert the complete contract before publishing a new snapshot and when reading the previous snapshot.
- [x] Add regression tests for cross-product identity collisions, SKU conflicts, source mismatch and invalid numeric values.
- [x] Verify Catalog CI and Backend CI on the same code head.

**Verification on implementation commit `e9d864f6bdbf0013137929e02b6258007200db29`:** Catalog CI PR run [41362606331](https://github.com/pooya16pg-spec/globalcart/actions/runs/41362606331), Catalog CI push run [41362602994](https://github.com/pooya16pg-spec/globalcart/actions/runs/41362602994), and Backend CI PR run [41362606267](https://github.com/pooya16pg-spec/globalcart/actions/runs/41362606267) all succeeded.

**Safety note:** validation mirrors the PostgreSQL identity and numeric constraints plus the catalog's 0–5 rating domain; it rejects a snapshot rather than mutating, deduplicating or fabricating supplier data.


## M08.8.5 — PostgreSQL rating domain constraint — CI verified
- [x] Add a forward-only migration enforcing `products.rating <= 5` in PostgreSQL, preserving the existing non-negative constraint.
- [x] Add a PostgreSQL integration regression proving an out-of-range 5.01 rating aborts the entire import batch, including earlier valid rows.
- [x] Ensure the regression isolates the rating constraint from variant/SKU constraints.
- [x] Verify Catalog CI, Backend CI, PostgreSQL migration/integration tests and Docker build on the same implementation head.

**Verification on implementation commit `cc36339bcac77d08abc37af3f88f08607a090367`:** Catalog CI PR run [41363342797](https://github.com/pooya16pg-spec/globalcart/actions/runs/41363342797), Backend CI PR run [41363342758](https://github.com/pooya16pg-spec/globalcart/actions/runs/41363342758), and Backend CI push run [41363337241](https://github.com/pooya16pg-spec/globalcart/actions/runs/41363337241) all succeeded.

**Safety note:** this tightens the existing canonical rating domain; it does not rewrite existing rows or fabricate rating values.


### M07.2 — Accessible suggestions and keyboard command palette — CI verified (2026-11-02)
- [x] Build deterministic, bounded product, brand and category suggestions from the catalog already loaded by the storefront.
- [x] Add accessible combobox/listbox semantics, active-option tracking and Arrow Up/Down, Enter and Escape behavior.
- [x] Add a Ctrl/⌘+K command palette for catalog, category, cart, favorites, comparison, reset-filters and language actions.
- [x] Reuse the modal focus trap and restore focus when the command dialog closes; explicitly focus mobile search when expanded.
- [x] Add unit tests for prefix ranking, normalization, de-duplication, bounded results and source-array non-mutation.
- [x] Verify TypeScript, Catalog unit tests and production build.

**Verification:** the feature was implemented in 3d6b840f769a4c23de134800877de9c9b47e8298; Catalog CI push run [41366011979](https://github.com/pooya16pg-spec/globalcart/actions/runs/41366011979) and PR run [41366016027](https://github.com/pooya16pg-spec/globalcart/actions/runs/41366016027) passed. The combined current PR head 87bc930da16367bd725b68fec460bae2fda6f26e also passed [Catalog CI PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41366593947).

**Scope:** suggestions and commands run against the browser-loaded sample catalog and local demo state; they do not call the unhosted API or create real orders/payments.

### M07.3 — PostgreSQL trigram typo tolerance — CI verified (2026-11-02)
- [x] Enable the trusted pg_trgm extension in a forward-only migration and add GIN indexes for normalized product titles and brands.
- [x] Keep full-text search as the primary path; use trigram/word-similarity matching as a fallback only for queries of at least three characters.
- [x] Rank exact full-text matches ahead of fuzzy title/brand matches.
- [x] Keep list and count search predicates aligned, including fuzzy matches.
- [x] Document the 200-character query limit and fuzzy-search behavior in OpenAPI.
- [x] Add PostgreSQL integration tests for exact-match precedence, typo-tolerant title/phrase and brand search, and count/list consistency.
- [x] Verify sqlc generation, Go vet, PostgreSQL 17 integration/race tests, API/importer builds and Docker image build.

**Verification on final implementation head 87bc930da16367bd725b68fec460bae2fda6f26e:** [Backend CI push](https://github.com/pooya16pg-spec/globalcart/actions/runs/41366589804), [Backend CI PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41366594050) and [Catalog CI PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41366593947) all succeeded.

**Scope:** this changes only the local PostgreSQL search path. The public GitHub Pages storefront continues to search its generated static catalog; no hosted API connection is implied.


### M07.4 — Explainable catalog recommendation ranking — CI verified (2026-11-02)
- [x] Replace the opaque rating-times-discount order with a documented weighted score: 70% normalized rating and 30% normalized discount (discount contribution capped at 50%).
- [x] Round scores to a stable six-decimal precision and use deterministic rating/price/ID tie-breakers without mutating the catalog.
- [x] Show a short per-product reason (top rating, strong discount, both signals or balanced signals) in every product card.
- [x] Explain that the recommendation order is a demo heuristic and is not personalized from browsing or purchase history.
- [x] Add unit tests for score weights, discount cap, reason labels, deterministic ranking, invalid-signal bounds and non-mutation.
- [x] Verify TypeScript, catalog/domain unit tests and production build in Catalog CI.

**Verification on implementation head 09fd73ea1f1dcb1f01e5022c3f4380ec2832855a:** [Catalog CI push](https://github.com/pooya16pg-spec/globalcart/actions/runs/41366975455) and [Catalog CI PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41366979976) passed. [Backend CI PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41366979999) also passed, including PostgreSQL 17 tests and Docker image build.

**Boundary:** ranking uses only the sample rating and discount fields already in the loaded catalog. It does not use personal data, purchase history, real popularity or live behavioral analytics.


### M07.5 — Browser-local saved searches and preferences — CI verified (2026-11-02)
- [x] Save the current query, category, brand, price range, minimum rating and sort order under a user-defined name.
- [x] Persist up to twelve saved searches in browser localStorage only; provide apply and delete actions.
- [x] Prevent duplicate filter/sort combinations and validate browser-stored data before using it.
- [x] Ignore malformed entries safely and keep saved-search state independent from cart/favorites data.
- [x] Add unit coverage for storage normalization, invalid records, collection bounds and duplicate detection.
- [x] Verify TypeScript, catalog/domain tests and production build.

**Verification on implementation head ff80ebcfcf4bcec4cdbbf9217c866070614524be:** [Catalog CI push](https://github.com/pooya16pg-spec/globalcart/actions/runs/41367502278) and [Catalog CI PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41367539860) passed; [Backend CI PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41367539848) also passed.

**Privacy boundary:** saved searches live in this browser only. No account, server sync, analytics profile or backend endpoint is involved.

### M07.6 — Explicit catalog empty/loading/offline states — CI verified (2026-11-02)
- [x] Keep the generated last-known-good catalog visible immediately while a background source refresh is pending.
- [x] Show a loading state only when no generated snapshot is available yet.
- [x] Surface a clear stale-snapshot status when the online source fails or returns an empty collection; never replace a valid snapshot with empty data.
- [x] Distinguish the last-known-good snapshot from the local fixture fallback used only when no snapshot exists.
- [x] Retain the existing explicit no-results state for empty filtered results and announce source states through a status region.
- [x] Add pure state-resolution tests for pending, live, empty and error outcomes with and without a snapshot.
- [x] Verify TypeScript, catalog/domain tests and production build.

**Verification on implementation head ff80ebcfcf4bcec4cdbbf9217c866070614524be:** [Catalog CI push](https://github.com/pooya16pg-spec/globalcart/actions/runs/41367502278) and [Catalog CI PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41367539860) passed; [Backend CI PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41367539848) passed.

**Safety:** a failed or empty refresh does not clear or overwrite the last-known-good generated snapshot.


## M08.1 — Static storefront browser journeys — CI verified (2026-11-02)
- [x] Add Playwright desktop Chromium and Pixel 7 mobile projects against the actual Next.js static export, served locally with the GitHub Pages `/globalcart/` base path.
- [x] Block third-party network requests during browser tests so the committed catalog snapshot is the deterministic source of truth.
- [x] Verify Persian/English locale switching and opening a pre-generated static product detail route.
- [x] Verify accessible search suggestions with Arrow/Enter selection and the Ctrl/⌘+K command palette.
- [x] Verify saved searches survive page reload, can be applied, and can be deleted from browser-local storage.
- [x] Verify cart drawer and clearly simulated checkout/order completion without any real payment/shipping path.
- [x] Add a static file server for E2E that enforces the configured base path and rejects path traversal.
- [x] Add Playwright traces, screenshots on failure and a short-lived CI artifact for debugging.

**Verification on implementation head 08c86aeb909228ecc1ba03d35195ad4fba6ab987:** [Browser E2E push](https://github.com/pooya16pg-spec/globalcart/actions/runs/41369733216) and [Browser E2E PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41369737767) passed, each running five journeys on desktop and mobile. [Catalog CI PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41369737769) and [Backend CI PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41369737810) also passed, including PostgreSQL 17 checks and Docker build.

**Coverage boundary:** this suite validates the existing demo storefront and does not imply live inventory, real payment, real order submission, or a hosted backend.


## M08.2 — Automated accessibility checks and remediation — CI verified (2026-11-02)
- [x] Add axe-core Playwright checks for the storefront homepage, open command-palette dialog and static product detail page.
- [x] Run checks on Chromium desktop and Pixel 7 mobile, using WCAG 2.0 A/AA, WCAG 2.1 A/AA and WCAG 2.2 AA rule tags.
- [x] Fix detected low-contrast muted text in catalog/product metadata, recommendation explanations, discount/rating badges, modal footer and product pricing.
- [x] Increase FAQ summary and footer navigation target sizes to at least the automated 24 CSS-pixel target-size threshold.
- [x] Make responsive cart text available to assistive technology without overriding its natural accessible name; align other icon/button labels with visible content and name the sort select.
- [x] Verify Browser E2E Push and PR after remediation; Catalog CI and Backend CI PR checks also pass.

**Verification on implementation head e515d696873aa2811c3edb35fc89ace74ae874df:** [Browser E2E push](https://github.com/pooya16pg-spec/globalcart/actions/runs/41370817992), [Browser E2E PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41370823048), [Catalog CI PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41370822987) and [Backend CI PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41370823026) all succeeded.

**Coverage boundary:** this is an automated axe scan of the three listed states at two viewport profiles. It is not a legal accessibility certification and does not replace manual screen-reader, keyboard-only usability and human contrast review.


## M08.3 — Keyboard-only modal focus verification — CI verified (2026-11-02)
- [x] Add Playwright keyboard-only coverage for the command palette's Tab and Shift+Tab focus trap in desktop Chromium and Pixel 7 mobile.
- [x] Verify Escape closes the command palette and restores focus to the exact invoking search field.
- [x] Capture the invoking element before opening through Ctrl/⌘+K or the command button; restore focus after the modal has been removed from the DOM.
- [x] Run all 18 browser tests on desktop/mobile with zero failures after the focus restoration fix.
- [x] Verify Browser E2E Push and PR, Catalog CI PR and Backend CI PR on the same final implementation head.

**Verification on implementation head 8937b54f0f631695ee02e51a128cc42815988f43:** [Browser E2E push](https://github.com/pooya16pg-spec/globalcart/actions/runs/41375898314), [Browser E2E PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41375903109), [Catalog CI PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41375903056), and [Backend CI PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41375903095) all succeeded. Browser E2E reports 18 passed across desktop and mobile.

**Coverage boundary:** this automated keyboard test covers the Command Palette, not every dialog or every screen-reader combination. Manual keyboard walkthrough and assistive-technology review remain open.


## M08.4 — Image loading policy and first-party asset budgets — CI verified (2026-11-02)
- [x] Prioritize the homepage hero image with eager loading, high fetch priority and asynchronous decoding; reduce its requested width/quality to a responsive 1200px / q80 source.
- [x] Prioritize the product-detail main image and defer gallery thumbnails, related products, catalog cards, merchandising rails, recent items and modal imagery with lazy loading and asynchronous decoding.
- [x] Preserve fixed media-container dimensions so image loading does not change reserved layout geometry.
- [x] Add desktop/mobile browser tests that enforce one critical eager image on the homepage and one on product detail, with all non-critical images lazy and asynchronously decoded.
- [x] Measure first-party initial-route JavaScript and CSS transfer sizes in Chromium and assert budgets of 750 KB JS, 120 KB CSS, no more than 10 JS requests and 2 CSS requests.
- [x] Verify the full 24-test Browser E2E suite on Push and PR, plus Catalog CI and Backend CI PR, on the same head.

**Measured baseline on implementation head `a741cf60ab44a17acf31bba43f47eb40204f2be5`:** 512,814 bytes of JavaScript across 7 requests and 67,702 bytes of CSS across 1 request. Browser E2E [Push](https://github.com/pooya16pg-spec/globalcart/actions/runs/41376758145) and [PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41376763890) passed with 24 tests; [Catalog CI PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41376763985) and [Backend CI PR](https://github.com/pooya16pg-spec/globalcart/actions/runs/41376763923) also passed.

**Coverage boundary:** E2E intentionally blocks remote hosts to stay deterministic. These budgets measure first-party JavaScript/CSS and markup loading policy, not actual CDN image payloads, real-network image latency, LCP/CLS or third-party font delivery. Those checks remain open for a network-enabled performance audit.
