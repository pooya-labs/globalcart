# GlobalCart — Project Context / Single Source of Truth

**Owner:** Pooya Goodarzi  
**Repository:** https://github.com/pooya16pg-spec/globalcart  
**Branch:** main (M04 backend foundation merged; PR #1 closed 2026-11-02)  
**Package version:** 0.2.3-dev  
**Active stage:** M04 identity/session/RBAC internal boundary; M04.12 customer resource ownership is CI-verified; public HTTP auth/write routes remain gated  
**Last updated:** 2026-11-02

## Product definition
GlobalCart is an independent marketplace portfolio simulation. Its information architecture is inspired by large retail marketplaces and the Amazon-link sourcing flow of kiyandaria.com, while its brand, UI and implementation remain original. It is not a live retailer. No real payments, order fulfillment, inventory reservations, seller payouts or money refunds occur; return/refund outcomes are internal demo-only state transitions.

## Non-negotiable invariants
- Keep this repository independent from all other portfolio projects.
- Never collect card numbers/bank credentials or call a real payment gateway.
- Label sample data, illustrative currency and simulated outcomes honestly.
- Do not scrape or bulk-copy Amazon. Live catalog integration requires authorized API/feed access and terms review.
- Amazon product-link intake currently validates an Amazon hostname and 10-character ASIN, then stores a local demo request only. It does not fetch product data.
- Translation is paused; preserve original source text and do not add translation back to CI unless explicitly requested.
- Never expose provider secrets in browser code or Git history.
- Persian is default with true RTL; English must retain correct LTR.
- Every visible control must work or explain why unavailable.
- Keep owner attribution visible.
- Never mark roadmap work complete before verification.

## Current implementation map
- `app/page.tsx` — bilingual storefront, catalog search/filter/sort with dynamic brand and minimum-rating facets, browser-local saved searches, explicit snapshot/live/offline catalog states, explainable sample recommendation ranking/reason labels, accessible product/brand/category suggestions and Ctrl/⌘+K command palette, category mega-menu, Amazon URL request intake, local cart/favorites/compare, quick view and simulated checkout, service shortcuts and FAQ.
- `app/product/[id]/page.tsx` — statically generated product routes from the current catalog snapshot.
- `app/product/[id]/ProductDetailClient.tsx` — product gallery, illustrative price breakdown, product facts, browser-local visit tracking, local cart/favorite actions and related products.
- `app/recently-viewed/page.tsx` — static-exportable history page reading and clearing browser-local recent product IDs.
- `app/globals.css` — responsive marketplace design system, product detail, price breakdown, recently viewed rail, mega-menu, comparison and Amazon intake styling.
- `app/layout.tsx` — metadata and global styles.
- `data/catalog.json` — generated 100-product sample catalog snapshot.
- `scripts/sync-catalog.ts` — scheduled/manual DummyJSON import, retry, normalization and last-known-good protection.
- `lib/catalog/*` — canonical catalog contracts, normalization, source adapter, pure local facet filtering (filtering.ts) and bounded search-suggestion ranking (suggestions.ts).
- `lib/pricing.ts` — integer-cent/fixed-point demo pricing with BigInt.
- `lib/translation/*` and `google-apps-script/*` — parked legacy translation implementation; not used by storefront or CI.
- `tests/*.test.ts` — catalog ingestion/quality, storefront facet filtering, search-suggestion, recommendation, saved-search and source-status tests, pricing and legacy translation tests.
- `.github/workflows/deploy.yml` — CI, scheduled/manual catalog refresh, static build and Pages deployment.
- `ROADMAP.md` — authoritative milestone checklist.
- `backend/` — Go REST API, PostgreSQL migrations 001–011, local Docker Compose, catalog importer, internal demo-commerce/return services, OpenAPI contract and `backend/internal/auth`. The auth package provides Argon2id primitives, digest-only opaque tokens, and an internal PostgreSQL session lifecycle; none is wired to public authentication routes.
- `go.mod` — Go module at repository root.
- `docs/AI_HANDOFF.md` — continuation contract.

## Verified behavior
- Catalog snapshot pipeline imports 100 sample products and preserves a last-known-good snapshot.
- The storefront supports Persian RTL and English LTR, category navigation, search, filters, sorting, favorites, recently viewed, local cart and simulated checkout.
- Product detail routes are statically generated for products in the snapshot and include gallery, pricing, facts and related products.
- Keyboard support includes Escape dismissal for overlays, dialog semantics, visible focus styles and live announcements for key feedback states.
- Comparison supports up to four products and is persisted in browser local storage.
- Homepage merchandising rails are derived from the current sample catalog (discounts, ratings and discovery).
- Homepage overlay interactions support Escape dismissal; cart, favorites, quick view, comparison and checkout dialogs expose labelled dialog semantics, trap Tab/Shift+Tab focus and restore focus on close. When a mobile Command Palette action opens another dialog, the original opener is retained for focus restoration. The desktop/mobile Playwright regression suite verifies these paths; manual assistive-technology QA remains outstanding.
- Recently viewed product IDs are stored in browser local storage, shown in a homepage rail and on /recently-viewed; the footer links to the history page. There is no server-side account history.
- Product detail pages expose source USD price, sample discount, illustrative user-configured FX rate and converted demo price. No live FX provider is connected.
- Amazon URL intake validates supported Amazon domains and ASINs, then saves requests in browser local storage. Live Amazon data is not fetched.
- Translation is out of scope and removed from the active CI pipeline.

## M04.1 identity schema checkpoint — 2026-11-02

- Migration 011 adds normalized users, migration-controlled customer/seller/admin roles, vendor-scoped seller memberships, server-side session digests, one-time action-token digests and audit events.
- PostgreSQL enforces Argon2id PHC hash shape, verified-email activation, account-disable lifecycle, 32-byte token digests, idle/absolute session expiry, one active action token per user/purpose and append-only audit rows (UPDATE/DELETE/TRUNCATE rejected).
- Migration execution now recognizes SQL quoted strings, identifiers, comments and dollar-quoted function bodies; unit tests cover splitting and PostgreSQL integration tests cover the schema invariants.
- `backend/internal/auth` provides Argon2id PHC hash/verify and rehash helpers plus 256-bit opaque token generation, SHA-256 digesting and constant-time digest comparison. Verification rejects malformed and over-budget PHC parameters; bounded legacy hashes can be marked for rehash.
- Internal sessions are created only for active verified accounts, store only session/CSRF digests, use a 15-minute idle timeout capped by a 24-hour absolute lifetime, resolve effective roles and active seller vendor scopes from PostgreSQL, and revoke/audit logout atomically.
- No login/register/session HTTP endpoints, email delivery, cookie handling, live user data or hosted infrastructure was enabled. ADR-010 remains proposed; approval for same-site HTTPS frontend/API hosting, email delivery and secret management remains a prerequisite.


## M04.3 session lifecycle checkpoint — CI verified (2026-11-02)
- Internal session creation, refresh/role resolution, session-bound CSRF verification and logout revocation are implemented in `backend/internal/auth/sessions.go`.
- Backend CI [41387103123](https://github.com/pooya16pg-spec/globalcart/actions/runs/41387103123), Security scans [41387103141](https://github.com/pooya16pg-spec/globalcart/actions/runs/41387103141) and GitHub Pages deployment [41387103149](https://github.com/pooya16pg-spec/globalcart/actions/runs/41387103149) passed on the code/documentation head immediately before the catalog-only snapshot refresh.

## M04.4 bulk session revocation checkpoint — CI verified (2026-11-02)
- Added `RevokeUserSessions`, which requires the caller's open `pgx.Tx`, locks the target user row, and serializes against `CreateSession`.
- Reasons are a typed allow-list matching migration 011; every revoked session receives one append-only, redacted audit event. The optional actor ID supports self-service, administrator and system-initiated mutations.
- Integration coverage verifies caller rollback, rejected reasons, per-user isolation, audit attribution and idempotent repeated calls.
- Backend CI [run 41393919125](https://github.com/pooya16pg-spec/globalcart/actions/runs/41393919125) passed the PostgreSQL-backed race-enabled Go suite, `go vet`, dependency vulnerability scan, API/importer builds and Docker image build. Security scans [run 41393919064](https://github.com/pooya16pg-spec/globalcart/actions/runs/41393919064) also passed on code head `6e044b63048b0911c5a7078279e38fe354898605`.
- The helper is intentionally not wired to HTTP. Password, role and account-status mutation services must invoke it in the same transaction as their own database changes.


## M04.5 password-change and stale-login checkpoint — CI verified (2026-11-02)
- `CreateSession` now requires the exact password-hash snapshot verified by the caller and compares it in constant time while holding the account row lock.
- Internal `ChangePassword` re-checks the expected old hash, writes a fresh Argon2id hash, revokes all current sessions and appends an `auth.password.changed` event in one PostgreSQL transaction.
- A login that verified an old password before a concurrent change cannot create a session after the new hash commits; it receives `ErrCredentialsChanged`.
- Integration coverage verifies session revocation, new-password verification and stale snapshot rejection using a real Argon2id old-password hash.
- Backend CI [run 41395032305](https://github.com/pooya16pg-spec/globalcart/actions/runs/41395032305), Security scans [run 41395032374](https://github.com/pooya16pg-spec/globalcart/actions/runs/41395032374) and Pages deployment [run 41395032292](https://github.com/pooya16pg-spec/globalcart/actions/runs/41395032292) passed on implementation head `fe2ec6b9bfce6ec4aa20ea004029e2a68d5b05ab`.


## M04.6 password-reset token lifecycle checkpoint — CI verified (2026-11-02)
- Reuses migration 011 `auth_action_tokens`; no new table or migration was added.
- Internal token issuance stores only a SHA-256 digest, bounds expiry to 5 minutes–1 hour (30-minute default) and rotates previous outstanding reset tokens.
- Internal completion re-checks token validity while holding the user row lock, consumes once, updates the Argon2id hash, revokes sibling reset tokens and all active sessions, and appends a redacted audit event in one transaction.
- Normal password changes now revoke outstanding reset tokens too. No email delivery or HTTP reset endpoint is enabled.
- PostgreSQL coverage includes digest-only storage, rotation, expiry, repeated/concurrent consumption and session invalidation.
- Backend CI [run 41395571464](https://github.com/pooya16pg-spec/globalcart/actions/runs/41395571464), Security scans [run 41395571531](https://github.com/pooya16pg-spec/globalcart/actions/runs/41395571531) and Pages deployment [run 41395571666](https://github.com/pooya16pg-spec/globalcart/actions/runs/41395571666) passed on implementation head `8a504dca8ae23ca2c06143808452898281216abe`.


## M04.7 email-verification token lifecycle checkpoint — CI verified (2026-11-02)
- Reuses migration 011 `auth_action_tokens`; no schema migration or duplicate table was added.
- Internal issuance is pending-account-only, stores SHA-256 digests, bounds TTL to 5 minutes–72 hours (24-hour default), rotates previous tokens and appends redacted audit events.
- Consumption locks the user first, revalidates and consumes once, activates only the matching pending account, revokes sibling verification tokens and records completion atomically.
- The database-assigned customer role is preserved; verification cannot grant seller/admin permissions. No email delivery or public registration/verification route is enabled.
- Tests cover digest-only storage, rotation, expiry, disabled accounts, replay, customer-only roles and concurrent double consumption.
- Backend CI [run 41396081833](https://github.com/pooya16pg-spec/globalcart/actions/runs/41396081833) and Security scans [run 41396081844](https://github.com/pooya16pg-spec/globalcart/actions/runs/41396081844) passed on implementation head `d720ffdebd77c6b28c82ce55567ea3bba90d95c8`.


## M04.8 administrator-role grant/revocation checkpoint — CI verified (2026-11-02)
- `backend/internal/auth/admin_roles.go` provides internal `GrantAdminRole` and `RevokeAdminRole`; no role schema/migration changes were needed.
- The actor must be a different active, email-verified administrator. Targets for grants must be active and verified. Self-change, duplicate grant and missing revoke are rejected.
- A transaction-scoped advisory lock serializes privileged role changes; actor/target user rows are locked in stable order.
- Every role mutation revokes the target's current sessions and records an actor-attributed redacted audit event in the same transaction.
- PostgreSQL tests cover authorization failures, disabled actors, pending targets, self-change, session invalidation, audit attribution and concurrent grants.
- Backend CI [run 41396823353](https://github.com/pooya16pg-spec/globalcart/actions/runs/41396823353), Security scans [run 41396823371](https://github.com/pooya16pg-spec/globalcart/actions/runs/41396823371) and Pages deployment [run 41396823403](https://github.com/pooya16pg-spec/globalcart/actions/runs/41396823403) passed on code head `aef92e7526cfbc9c51f0d803bcba8493e131d8d2`.
- No HTTP endpoint or self-bootstrap flow exists. First-admin provisioning remains an operator-controlled database operation.


## M04.9 seller-membership administration checkpoint — Backend CI and Security verified (2026-11-02)
- `backend/internal/auth/seller_memberships.go` reuses migration 011 tables. It provides internal request, activate, suspend and revoke operations; no schema migration or public route was added.
- Requests are pending only and do not grant the seller role. Activation requires a different active, verified admin actor, an active/verified target and an active vendor.
- Each membership mutation is serialized with admin role operations. The user keeps seller role while at least one active membership on an active vendor remains; the role is revoked after the final active scope is removed.
- Affected sessions are invalidated in the same transaction; membership and automatic role changes append actor-attributed audit events.
- PostgreSQL tests verify two-vendor scope preservation, remaining-scope behavior after suspension, final seller-role removal, disabled-target cleanup, inactive vendor/requester rejection and concurrent activation.
- Backend CI [run 41397502226](https://github.com/pooya16pg-spec/globalcart/actions/runs/41397502226) and Security scans [run 41397502232](https://github.com/pooya16pg-spec/globalcart/actions/runs/41397502232) passed on implementation head `f0756e0a3360b949b384aef1ff8e3a646192257b`. Pages deployment [run 41397502248](https://github.com/pooya16pg-spec/globalcart/actions/runs/41397502248) also passed.
- No seller HTTP endpoint, dashboard, self-approval path or seller payout feature exists.

## M04.10 deny-by-default internal authorization policy checkpoint — CI verified (2026-11-02)
- `backend/internal/auth/policy.go` defines explicit customer-own-resource, seller-vendor-scoped and administrator-only actions. Unknown actions deny by default; missing session identity returns `ErrUnauthenticated`, and failed rules return `ErrForbidden`.
- `Authorize` accepts a `SessionPrincipal` which callers must obtain from `ResolveSession`. The evaluator is pure/internal: it does not authenticate, refresh sessions or replace canonical resource ownership/vendor checks inside the business transaction.
- Table-driven tests cover own/foreign customer resources, assigned/unassigned seller vendors, seller role without a live scope, admin actions, unknown actions, missing session identity and attempted client-supplied vendor scope.
- Backend CI [run 41398064895](https://github.com/pooya16pg-spec/globalcart/actions/runs/41398064895), Security scans [run 41398064885](https://github.com/pooya16pg-spec/globalcart/actions/runs/41398064885) and Pages deployment [run 41398064905](https://github.com/pooya16pg-spec/globalcart/actions/runs/41398064905) passed on implementation head `c1b4b0ccfbd132ee95f8dc757a3518ff9b04899a`.
- No HTTP middleware, cookie authentication, CSRF route or public protected endpoint was enabled.
## M04.11 service-boundary seller authorization checkpoint — CI verified (2026-11-02)
- `backend/internal/auth/policy_transaction.go` adds `AuthorizeSellerVendorTx`: it applies the pure M04.10 policy, then locks/revalidates active verified user, current session, active vendor, seller role and exact active membership inside the protected transaction.
- `backend/internal/commerce/vendors_authorized.go` adds seller-scoped offer read, price update and stock adjustment. Each derives vendor ID from the canonical offer row; no client vendor selector is accepted. Existing internal commerce mutation helpers are shared to avoid duplicate logic.
- PostgreSQL tests cover own-vendor operations, cross-vendor IDOR denial, stale session denial, suspended membership denial, and membership revocation blocking on the protected transaction's account lock.
- Backend CI [run 41399542624](https://github.com/pooya16pg-spec/globalcart/actions/runs/41399542624), Security scans [run 41399542539](https://github.com/pooya16pg-spec/globalcart/actions/runs/41399542539) and Pages deployment [run 41399542545](https://github.com/pooya16pg-spec/globalcart/actions/runs/41399542545) passed on implementation head `f06de264488bd59f1e76a7e3d6c67f26d40c0eb1`.
- No public seller route, dashboard, cookie authentication or CSRF middleware was enabled.

## M04.12 customer resource ownership checkpoint — CI verified (2026-11-02)
- Migration 012 adds nullable `carts.owner_user_id`. `BindCartToCustomer` requires both the existing guest capability digest and a currently valid, verified customer session; the cart row lock makes the first account binding win without disclosing a foreign cart.
- `GetCartForCustomer`, `AddItemForCustomer`, `SetItemQuantityForCustomer` and `RemoveItemForCustomer` resolve canonical ownership from PostgreSQL, revalidate the live customer session/role and hold the cart row lock in the same transaction as the protected operation. Product/variant price, title and stock remain database-derived.
- `GetOrderForCustomer`, `GetVendorOrderForCustomer` and `CancelOrderForCustomer` bind parent and child order access to the owning cart account. Foreign and guest-only order IDs return the same not-found error; cancellation preserves the existing reservation-release and simulated-order state machine.
- Guest cart capabilities remain a separate bearer capability for one cart. They do not grant user identity, customer role, seller scope or admin authority. Customer methods do not accept client-supplied owner IDs.
- `backend/internal/commerce/customer_ownership_integration_test.go` covers foreign cart/order IDOR, cart add/set/remove, parent and vendor-child order reads, cancellation/replay, concurrent first-writer-wins binding, expired carts and revoked sessions.
- Backend CI [41403135351](https://github.com/pooya16pg-spec/globalcart/actions/runs/41403135351), Security scans [41403135364](https://github.com/pooya16pg-spec/globalcart/actions/runs/41403135364) and Pages deployment [41403135344](https://github.com/pooya16pg-spec/globalcart/actions/runs/41403135344) passed on implementation head `928c3867e2d8e109e4e6595f9ac0898cf51b25b9`. CI persisted a formatting-only test-file commit as `5a5659d9b35789cf12072141152b0ab4be217c33`.
- No public customer/authentication route, cookie flow, seller/admin route or hosted backend was enabled.

## Deployment
Public preview: https://pooya16pg-spec.github.io/globalcart/  
Repository Actions: https://github.com/pooya16pg-spec/globalcart/actions

Actions run 41299677941 (Run #135) passed catalog refresh, snapshot persistence, TypeScript validation, unit tests, production static export, Pages artifact upload and GitHub Pages deployment. This run includes the dedicated recently viewed page, homepage history rail and browser-local history controls.
https://github.com/pooya16pg-spec/globalcart/actions/runs/41299677941

## Main branch and merge status — 2026-11-02
- PR #1, `feat(globalcart): consolidate marketplace and backend foundation`, was squash-merged into `main` at `dc9b520b026a508825eb3dd1f6cbc3655be5f039`.
- The follow-up internal return/refund simulation is implemented on `main` at `bc97475c0cce7f8de6aebe72071813c942b6a3d0`.
- Backend CI [41382309964](https://github.com/pooya16pg-spec/globalcart/actions/runs/41382309964), Security scans [41382310050](https://github.com/pooya16pg-spec/globalcart/actions/runs/41382310050) and GitHub Pages deployment [41382310024](https://github.com/pooya16pg-spec/globalcart/actions/runs/41382310024) all passed on that implementation commit.
- The backend remains local-development scaffolding. No hosted database/API, real payment/refund provider, production secrets or paid service was provisioned.
- PR #1 is closed and merged: https://github.com/pooya16pg-spec/globalcart/pull/1

## Latest quality checkpoint — M08.11 (CI verified, 2026-11-02)
- Implementation head: `9653fc7d2d03e45f726b7e6a0e47fd7cbd1295dc` on `main`.
- Playwright Browser E2E [41383420470](https://github.com/pooya16pg-spec/globalcart/actions/runs/41383420470) passed all 30 tests across desktop Chromium and Pixel 7 mobile.
- axe audits passed in both Persian RTL and English LTR. Reduced-motion behavior passed on both viewports, including scroll, transition and dialog-animation checks.
- Fixed the stale-catalog status contrast to exceed 5:1.
- Security scans [41383420409](https://github.com/pooya16pg-spec/globalcart/actions/runs/41383420409) and GitHub Pages deployment [41383420345](https://github.com/pooya16pg-spec/globalcart/actions/runs/41383420345) passed on the same implementation head.
- Manual screen-reader, real-device and full-site/deep-link QA remain open.

## Immediate next actions
1. Keep authentication, admin and seller HTTP routes, browser cookies and middleware disabled until same-site HTTPS hosting, email delivery, secret storage and runtime are explicitly approved. Actor IDs must come from verified server-side sessions.
2. Implement M04.12 customer resource ownership enforcement for cart/order/account services; preserve the separation between guest cart capability and authenticated identity.
3. Add PostgreSQL-backed cross-customer cart/order IDOR and ownership-race tests before any customer protected HTTP route is considered.
4. Complete manual responsive, RTL/LTR, screen-reader, real-device and published deep-link QA; automated axe, focus, reduced-motion and browser journeys are already CI-verified.
5. Obtain owner approval and documented rights for a named supplier API/feed and an authorized FX provider/rate-staleness policy before enabling either integration.

## Backend/database architecture
- Frontend: Next.js App Router, React and TypeScript, statically exported to GitHub Pages.
- Backend target: Go REST API with OpenAPI.
- Primary database: PostgreSQL with versioned SQL migrations and sqlc.
- Python: ingestion and enrichment workers; translation remains paused.
- Redis: optional cache/queue only when measured need exists.
- Search: PostgreSQL full-text/trigram first.
- M04 requires owner approval for API hosting, database and worker runtime.



## M04 backend foundation — merged into main (2026-11-02)
- PR #1 squash-merge: `dc9b520b026a508825eb3dd1f6cbc3655be5f039`.
- Go API, PostgreSQL 17 migrations, sqlc-generated catalog queries, local Compose, catalog importer and OpenAPI 3.1 contract are now on `main`.
- Read-only catalog/category/variant endpoints and local development CI are verified; hosted API/database, authentication, write endpoints and workers remain gated.

- Go API uses the standard HTTP mux and pgx v5; `go.mod` retains its Go 1.26.0 language baseline, while CI and the container builder use patched Go 1.27.3.
- PostgreSQL 17 is the primary system of record; local Compose binds ports to 127.0.0.1 only.
- Embedded transactional migration ledger creates an initial products table with source provenance, numeric constraints, JSONB image arrays and indexes.
- Read-only API routes: `GET /healthz`, `GET /readyz`, `GET /api/v1/products`, `GET /api/v1/products/{id}`.
- Product listing supports bounded limit/offset, category and full-text query parameters.
- A one-shot CLI imports the current normalized catalog snapshot; it refuses empty or incomplete imports.
- OpenAPI 3.1 contract is in `backend/openapi.yaml`.
- Backend CI run #22 passed gofmt, go vet, race-enabled unit and PostgreSQL integration tests, Go binary builds and Docker image build.
- Integration coverage verifies concurrent/idempotent migration startup, product upsert and transaction rollback after a database constraint violation.
- Verified run: https://github.com/pooya16pg-spec/globalcart/actions/runs/41301389812
- No authentication, user/order writes, payment, seller APIs or production deployment exists yet.
- Local Docker Compose is development scaffolding only; no cloud resources or paid services were created.


### M04 catalog schema extension — implementation in progress
- Added `002_catalog_taxonomy_variants.sql`: normalized category table with parent relationship, product category foreign key/backfill, and product_variants with SKU/source identity, price/stock checks and JSONB attributes.
- Existing `products.category` remains intact so current read endpoints and imported catalog contract stay backward-compatible.
- PostgreSQL integration test now checks category backfill and variant attribute round-trip.
- This change was developed on `feat/m04-go-postgres-foundation` and later merged into `main`; final verification is recorded in the M05.1 checkpoint.


### M04 API validation pass — implementation in progress
- Product list endpoint now rejects malformed/out-of-range pagination and oversized query/category filters with stable machine-readable 400 error codes.
- Added handler tests for invalid requests and pagination parsing, plus OpenAPI 400/error schema documentation.
- CI verification for this follow-up is pending.


### M04 typed SQL generation — in progress
- Added `sqlc.yaml` with PostgreSQL schema sourced from embedded migrations and pgx/v5 code generation to `backend/internal/dbgen`.
- Added parameterized product list/detail and active product variant query definitions under `backend/queries/catalog.sql`.
- Backend CI now generates and formats the typed package before vet, tests and builds. Generated package compilation is pending CI verification; API wiring remains next.


### M04 sqlc API integration — verified
- Product list and detail handlers now call generated `dbgen.Queries` methods. Numeric catalog columns are explicitly cast to float8 in query definitions to preserve the existing float64 API model.
- Added a row-to-domain mapper to preserve the established catalog JSON fields and safely decode images JSON.
- Mapper tests cover valid image arrays and malformed JSON.
- CI run 41302418821 passed sqlc generation, gofmt, go vet, race-enabled tests, API/importer builds; Docker image build was still running at last check.


### M04 product variants API — implementation in progress
- Added read-only `GET /api/v1/products/{id}/variants` endpoint, backed by sqlc-generated product existence and active variant queries.
- Variant response exposes ID, product ID, optional SKU, title, price, stock and JSON attributes; malformed attributes produce a stable internal catalog error.
- Added mapping tests and OpenAPI path documentation. CI verification pending.


### M04 product variants API — verified
- CI run 41302565323 passed sqlc generation, go vet, race-enabled unit/PostgreSQL integration tests, API/importer builds and Docker image build.
- Generated variant query returns typed float64 price; handler validates parent product existence and returns only active variants.


### M04 normalized category read API — implementation in progress
- Added `GET /api/v1/categories` backed by generated sqlc query.
- Response includes category ID/name, optional parent ID and product count.
- PostgreSQL integration test now executes generated category listing query after migration/backfill. CI run pending.


### M04 category API and auth design — 2026-11-01
- `GET /api/v1/categories` is implemented with generated sqlc query and product counts.
- PostgreSQL integration coverage exercises the generated category query after migrations/backfill. CI run 41302783939 passed sqlc generation, formatting, vet, unit/integration tests, API/importer builds and Docker image build.
- Authentication/session/RBAC architecture is documented in `docs/SECURITY_AUTH_DESIGN.md` and ADR-010. No authentication code or user tables are implemented; implementation waits for owner approval of same-site HTTPS hosting, email delivery and secret management.


### M04 catalog pagination — verified
- Product list filtering resolves normalized categories case-insensitively and ignores surrounding whitespace.
- Generated `CountProducts` query uses matching search/category predicates; list response adds `total` while retaining `count` as current-page item count.
- PostgreSQL integration test covers normalized category filter and count. Pull-request CI run 41303108214 passed all stages including Docker build.


### M05 demo commerce schema — verified
- Migration `003_demo_commerce.sql` adds guest cart capability digests, cart item snapshots, demo order/idempotency records, immutable order lines, expiring inventory reservations and order events.
- Raw cart tokens and idempotency keys are never stored; the schema stores fixed-length SHA-256 digests only.
- `docs/COMMERCE_DESIGN.md` defines transaction locking, stock reservation, price revalidation, replay behavior and demo-only state transitions.
- No commerce write endpoints, customer accounts, payment integration or live inventory mutation are enabled. CI run 41345512979 passed sqlc generation, Go formatting, vet, PostgreSQL-backed race tests, API/importer builds and Docker image build.


### M05 transactional checkout service — verified
- Added `backend/internal/commerce/checkout.go`: internal-only transactional checkout; not wired to HTTP.
- Locks cart and stock rows, reads canonical current catalog prices/titles, accounts for unexpired reservations, and atomically creates order snapshots, reservation rows and order event.
- Idempotency replay returns the original order before cart lifecycle checks. Cart expiry uses PostgreSQL's clock.
- Integration tests exercise two guest carts competing for limited product stock (one succeeds, one is rejected), successful checkout replay, and idempotent expiry of pending orders/reservations.
- CI run 41346127244 passed all stages including PostgreSQL-backed concurrency/expiry tests and Docker image build.
- No cart/order write endpoint, payment provider, live stock mutation, or browser cookie flow is enabled.


### M05 guest cart lifecycle service — CI verified
- Added `backend/internal/commerce/cart.go` with internal methods: `CreateCart`, `AddItem`, `SetItemQuantity`, `RemoveItem`, `GetCart`.
- Caller supplies a 32-byte capability digest; raw guest bearer tokens are never accepted or persisted by the service.
- Cart mutations lock the cart row, enforce open/unexpired lifecycle, read canonical product/variant price/title/stock, and enforce quantity/stock bounds.
- Currency is explicitly USD-only pending a real catalog currency/FX model.
- PostgreSQL integration tests cover create/add/increment/set/remove, stock overflow, active variants and expired carts.
- CI push run 41346723656 passed all stages including PostgreSQL-backed tests and Docker image build. The earlier PR run exposed one unused import; it was removed and the corrected commit passed.
- No cart HTTP routes, cookies, or cross-origin browser flows were enabled.


### M06 order read and cancellation services — CI verified
- Added `backend/internal/commerce/orders.go`: `GetOrder` scopes reads by the cart capability digest and order UUID, returning immutable item snapshots and ordered event history.
- Unknown order IDs and wrong owner digests intentionally return the same `ErrOrderNotFound`.
- Added `backend/internal/commerce/cancel.go`: `CancelOrder` locks the owned order, permits cancellation only from `awaiting_demo_payment`, releases active reservations and appends `demo_order_cancelled` atomically.
- Repeating cancellation returns the already-cancelled order; paid/expired orders cannot be cancelled.
- Integration tests verify correct owner access, cross-cart denial, snapshot/event reads, reservation release and idempotent cancellation.
- CI push run 41347325289 passed all stages, including PostgreSQL tests and Docker image build. An initial test fixture conflicted with a pre-existing variant reservation; the fixture was corrected to use available base-product stock, then the complete CI passed.
- No HTTP routes, cookie flow or public order access were enabled.


### M05 simulated payment lifecycle — implementation in progress
- Added `backend/internal/commerce/payment_simulation.go` with `CompleteDemoPayment`.
- The operation requires the owning cart's 32-byte capability digest, locks the order, requires `awaiting_demo_payment` and at least one unexpired active reservation, then atomically releases active holds, sets `paid_simulated` and appends one `demo_payment_simulated` event.
- Repeated completion returns the existing paid-simulated order without adding duplicate events. Wrong owner/unknown order returns `ErrOrderNotFound`; cancelled/expired states cannot be completed.
- This is explicitly a local simulation: no payment provider, credentials, webhook, card data or real-money movement.
- PostgreSQL integration coverage has been added for owner isolation, completion, replay, reservation release and rejection of cancellation after simulated payment. CI run 41348079983 passed sqlc generation, formatting, vet, race-enabled PostgreSQL tests, API/importer builds and Docker image build.
- No HTTP routes or browser checkout flow enabled.


### M07 internal coupon foundation — implementation
- Migration `004_coupons.sql` stores only SHA-256 digests of normalized coupon codes; it adds coupon rules, cart attachment, order discount snapshots and redemption records.
- Internal `CreateCoupon`, `ApplyCoupon` and `RemoveCoupon` methods support percentage/fixed discounts, validity windows, minimum subtotal, maximum discount, global usage limit and per-guest-capability usage limit.
- Checkout locks the coupon row and revalidates dates, active state, minimum subtotal and redemption limits in the same transaction as order creation. Discounts are calculated with PostgreSQL NUMERIC; order subtotal, discount and final total are persisted.
- `GetCart` exposes only `CouponApplied`; `GetOrder` includes subtotal/discount/total. No raw code or coupon identifier is returned.
- The current per-customer-like limit is scoped to the guest cart capability digest, not a verified person/account; account identity is still a prerequisite for a genuine customer-wide limit.
- PostgreSQL integration coverage includes percentage discount, global use exhaustion, minimum subtotal rejection/rollback, coupon attachment/removal and order-detail totals.
- Coupon administration and checkout remain internal-only; no public HTTP routes, hosted services or payment provider are enabled.


## M08 vendor foundation — migration 005
- Historical implementation branch: `feat/m04-go-postgres-foundation`; PR #1 was subsequently squash-merged into `main` on 2026-11-02.
- `005_vendor_foundation.sql` adds vendors, vendor_products, vendor_price_history and vendor_inventory_events.
- Current products and variants are backfilled as offers under the active GlobalCart Direct platform vendor, including initial price and opening-stock audit records.
- Offers support independent vendor-specific price, USD currency, stock, SKU and status; partial unique indexes enforce one base/variant offer per vendor.
- Integration coverage checks platform vendor backfill, two vendors offering the same variant with independent stock, audit records and duplicate-offer rejection.
- Critical: cart and checkout still use legacy products/product_variants stock and prices. Vendor offers are not yet wired to cart or checkout; no vendor order split exists.
- No seller membership, seller HTTP API, dashboard, payout, commission, shipment or production resource is enabled. Membership waits for verified identity and authorization.


## M09 — Vendor-aware cart foundation
- Migration 006 adds vendor_product_id to cart lines, child vendor-order structure, and vendor-aware order-item/reservation references.
- Added internal AddVendorItem, SetVendorItemQuantity and RemoveVendorItem methods with active vendor/offer checks, canonical snapshots and row locks.
- GetCart now returns vendor offer identity and display name; legacy cart methods explicitly scope to non-vendor rows.
- Legacy Checkout rejects vendor lines with ErrVendorCheckoutNotReady so vendor stock cannot be processed against legacy product inventory.
- Added PostgreSQL integration coverage for vendor cart lifecycle and this checkout guard.
- Split checkout and vendor child-order state propagation are not implemented yet.


## M09.2 — Vendor split checkout
- Added Service.CheckoutVendorCart as a separate internal-only transaction path; legacy Checkout continues to reject vendor lines.
- Vendor checkout creates one parent demo order and one child per vendor, snapshots canonical offer data and reserves inventory by vendor_product_id.
- Coupon is redeemed once at parent order level; proportional child discount allocation reconciles exactly to parent cents.
- Simulated payment, cancellation and reservation expiry now propagate state to vendor child orders.
- Mixed legacy/vendor carts are intentionally rejected. No payment gateway, real stock decrement, HTTP write routes, seller identity or production infrastructure.
- Integration tests cover two-vendor checkout, coupon totals, idempotent replay, legacy-stock non-mutation, simulated payment propagation and same-offer stock contention.


## Latest continuation checkpoint — 2026-11-02
- Historical implementation branch: `feat/m04-go-postgres-foundation`; PR #1 was subsequently squash-merged into `main` on 2026-11-02.
- M09.2 vendor split checkout is implemented: one parent order, one child per vendor, per-offer reservations, parent coupon redemption with exact child allocation, and lifecycle propagation for simulated payment/cancellation/expiry.
- Latest implementation adds `GetVendorOrder`, an owner-capability-scoped read for a single vendor child order with immutable item snapshots. Unknown IDs and wrong-owner IDs return the same not-found error.
- Added `vendor_orders_integration_test.go` for PostgreSQL-backed ownership and snapshot assertions.
- Backend CI run `41352752412` failed at `go vet` because `vendor_orders.go` imported unused `time`; the exact failure is recorded in its job log.
- Removed the unused import in commit `cb0f20816afdbf3f2dc23628e8a5bd4b7d1bce46`.
- **Immediate next action:** obtain a fresh CI run for the corrected branch head. The earlier failed run does not verify the fix; do not mark M09.3 complete until a fresh run passes sqlc generation, vet, PostgreSQL-backed race tests, builds and Docker image build.
- No public commerce or seller HTTP routes, browser session flow, real payments, hosted API/database, or production secrets are enabled.


## Latest continuation checkpoint — 2026-11-02 (CI verified)
- Historical implementation branch: `feat/m04-go-postgres-foundation`; PR #1 was subsequently squash-merged into `main` on 2026-11-02.
- M09.3 `GetVendorOrder` and its PostgreSQL integration tests are verified by fresh Backend CI run `41352898725` on head `1234d02b51c7c411816ff9a7ab1b64a84aec5bb9`; conclusion: `success`.
- The earlier run `41352848269` was cancelled after a newer PR run superseded it; it is not the verification record.
- Authentication/session/RBAC remains design-only. `docs/SECURITY_AUTH_DESIGN.md` requires a custom same-site HTTPS frontend/API origin, approved email delivery and secret management before identity endpoints or browser cookie flows are implemented.
- Next safe engineering slice: audit vendor-scoped service boundaries and define/test an internal authorization seam without creating public seller APIs or inventing seller identity.
- No cloud hosting, paid services, real payments, production secrets or deployment were enabled.


## M09.4 progress — 2026-11-02
- Added PostgreSQL integration assertions proving a mismatched vendor scope cannot read an offer, update its price or adjust its stock; the owning vendor's stock is asserted unchanged.
- Implementation commit: `e18a01d1a9414520334ceb312c2207990ccc71d5`.
- Roadmap updated to show the cross-vendor isolation slice implemented; CI verification is pending.

- Security guardrail added in `docs/SECURITY_AUTH_DESIGN.md`: internal `vendorID` parameters are selectors, not caller authorization; seller operations must not be exposed as HTTP routes until the server resolves a verified session and active seller membership. Roadmap records this design checkpoint.


## Latest checkpoint — 2026-11-02: M09.4 verified
- Cross-vendor isolation assertions in `backend/internal/commerce/vendors_integration_test.go` passed Backend CI run `41353137589` on code commit `e18a01d1a9414520334ceb312c2207990ccc71d5`.
- All CI steps passed, including vet, race-enabled tests, API/importer builds and Docker image build.
- Security design prohibits public seller routes until the handler resolves a verified session and active seller membership.
- PR #1 was subsequently squash-merged into `main` on 2026-11-02. See `ROADMAP.md` for the current checkpoint.
- M09.5 is blocked pending approval of same-site HTTPS hosting, runtime, secret management and email provider. No deployment or paid service was enabled.


## Canonical continuation checkpoint — 2026-11-02
- M09.3 owner-scoped vendor child-order reads: CI verified by run `41352898725`.
- M09.4 cross-vendor offer read/price/inventory isolation: CI verified by run `41353137589` on code commit `e18a01d1a9414520334ceb312c2207990ccc71d5`; Docker build passed.
- Seller/auth implementation remains gated by same-site HTTPS hosting, runtime approval, secret management and an approved email provider. `vendorID` is never proof of caller identity.
- New in-progress regression tests in `backend/internal/httpapi/handler_test.go` assert that seller/order/cart/checkout write routes are not exposed and catalog routes reject mutation methods. Commit: `e91132997a9053943bc2e6c3f8dbe82d59f7d9b5`. CI pending.
- Latest roadmap checkpoint: `ROADMAP.md`, stage M09.4.1. Do not treat the new HTTP guard tests as verified until their CI run completes.
- PR #1 was subsequently squash-merged into `main` on 2026-11-02. No production hosting, payment provider or paid infrastructure was enabled.

- Follow-up hardening: CORS no longer advertises the `Authorization` request header before an authenticated API exists; regression test asserts the restricted header policy. Code commits: `6d3f1f2` and `05e1442`. CI verification pending for the latest code head.


## Platform importer synchronization — 2026-11-02 (CI pending)
- `database.UpsertProducts` now synchronizes each imported base product into the seeded active `globalcart-direct` platform vendor within the same PostgreSQL transaction.
- New platform offers receive initial price-history and opening-inventory audit rows. Existing offers update price/stock only when the catalog source changes; price and inventory audit events are appended only for actual changes.
- Re-importing unchanged products is idempotent and does not duplicate audit events. Uncategorized products still synchronize their platform offer.
- PostgreSQL integration test covers initial import, unchanged re-import and changed price/stock updates.
- Code commits: `686c567` (implementation), `bd7d006` (integration coverage), `2a04bf3` (uncategorized-product edge case). CI verification pending.
- M09.4.1 public API regression guards are verified in Actions run `41353463006` (success). Seller routes and commerce write endpoints remain intentionally unexposed.

- Variant follow-up: importer synchronization now creates/updates `GlobalCart Direct` offers for each `product_variants` row, syncing SKU, price, stock and active/paused status. Variant reads are fully collected and closed before offer writes within the transaction. Integration coverage checks creation, unchanged re-import, price/stock change audit events, inactive pausing and fixture restoration. Commits: `5574937`, `58eaa21`, `a7eed9d`, `9eb5671`. CI pending.


## Latest continuation checkpoint — 2026-11-02: platform sync verified; source variant ingestion in progress
- M08.1 base-product and existing-database-variant synchronization passed Backend CI run `41354208659` on code head `5198c5f452b03920bc90fa1eec0a5b6e83dadd70`; all steps succeeded, including sqlc generation, formatting, vet, tests, API/importer builds and Docker image build.
- New M08.2 work adds optional `variants` to the TypeScript catalog contract and Go `catalog.Product` model. Normalization retains variant source identity, SKU, title, price, stock, object attributes and active state.
- `database.UpsertProducts` now upserts supplied variants in the same transaction before platform-offer synchronization. Source identity collisions cannot silently move a variant to another product; malformed price/stock/identity is rejected.
- PostgreSQL integration fixture now supplies the variant through the imported product payload rather than manually inserting it. Catalog unit tests cover variant normalization and safe defaults.
- Current DummyJSON sample feed does not provide variants; this is backward-compatible optional ingestion support, not a claim that live variant records are arriving. An authorized supplier feed/adapter is still needed to exercise real source variants.
- New code commits: `34e0883`, `a9eeef5`, `e925313`, `26d6e1f`, `26c8a1f`, `ee46960`. CI for this new end-to-end ingestion slice is pending.
- Historical implementation branch: `feat/m04-go-postgres-foundation`; PR #1 was subsequently squash-merged into `main` on 2026-11-02.


## Latest continuation checkpoint — 2026-11-02: M08.2 verified; M08.3 in progress
- Backend CI run `41354535515` completed with conclusion `success` on head `11132964594083998362a8ef1e35e646524b7e6f`. All steps passed, including sqlc generation, formatting, vet, race-enabled Go/PostgreSQL tests, API/importer builds and Docker image build.
- M08.2 end-to-end source variant ingestion is therefore CI-verified. The TypeScript catalog contract and Go importer accept optional source variants; the importer persists variant identity, SKU, title, price, stock, attributes and active state transactionally before syncing platform offers.
- M08.3 introduces `CatalogSourceAdapter` and `loadCatalogFromSource` as a reusable supplier ingestion boundary. Source adapters fetch supplier-specific records and map them into the canonical product/variant input contract.
- DummyJSON now uses the shared adapter contract. Its retry behavior and the snapshot generator's last-known-good protection are retained.
- Variant deduplication now keys by both source and source ID, matching the database unique identity pair. Equal source IDs from different suppliers are not incorrectly collapsed.
- Added a synthetic-only supplier fixture and tests for mapping two variants, source provenance, SKU, attributes, inactive status, source-scoped identity and empty/unmappable feed rejection.
- Added `.github/workflows/catalog-ci.yml` to run npm install, TypeScript validation, unit tests and a production build for relevant pushes and pull requests. Its result is the remaining M08.3 verification gate.
- No real supplier API/feed, credentials, live supplier inventory, hosting or secrets were added in this checkpoint. PR #1 was subsequently merged into `main` on 2026-11-02.


## Latest continuation checkpoint — 2026-11-02: M08.3 CI verified
- M08.2 source variant ingestion was verified by Backend CI run `41354535515` on head `11132964594083998362a8ef1e35e646524b7e6f`.
- M08.3 supplier adapter contract is implemented and verified. Catalog CI push run `41355090849` and PR run `41355095558` both succeeded on head `9bfab0c80dec906f52fb56a22b85d36b7eec99f9`; TypeScript validation, all catalog/domain unit tests and production build passed.
- A first M08.3 run exposed only an overly narrow test-message regex (14/15 tests passed); the assertion was corrected and the fresh runs passed. No production code workaround or skipped test was used.
- Shared types: `CatalogSourceAdapter` / `CatalogSourceRecord`; shared loader: `loadCatalogFromSource`; DummyJSON now goes through the shared boundary.
- Variant de-duplication follows the database identity pair `(source, source_id)`. A synthetic-only supplier fixture validates mapping of two variants, provenance, SKU, attributes and inactive state. Empty and unmappable feeds are rejected.
- The snapshot sync continues to preserve its last-known-good snapshot on failed ingestion. Product-only sources remain valid; absence of variants in a feed is not a command to delete existing database variants or offers.
- M08.4 is blocked on the owner selecting and approving a named supplier/feed and its rights. No real supplier connection, API key, live inventory, hosting, secrets or deployment was introduced.
- Branch head at this historical checkpoint: `9bfab0c80dec906f52fb56a22b85d36b7eec99f9`. PR #1 was subsequently merged into `main` on 2026-11-02.


## Follow-up hardening — 2026-11-02
- Variant de-duplication now encodes the source and source ID as an unambiguous JSON tuple rather than joining two strings with a delimiter. This prevents collisions when either value itself contains a colon.
- Added regression coverage for the former delimiter-collision shape.
- Latest Catalog CI run `41355272122` succeeded on code head `eaebbc0a7b8d1c823c3dcad4a7fed0d10a99edc9`: dependency install, TypeScript validation, all unit tests and production build passed.
- Latest Backend CI for this exact head is running; the backend code was unchanged by this normalization-only hardening. Recheck its final Docker image result before recording the full cross-workflow checkpoint.


## Latest continuation checkpoint — 2026-11-02: M08.4 snapshot integrity guard in progress
- Added `lib/catalog/quality.ts` with `assertCatalogSnapshotContinuity` and `summarizeCatalogQuality`.
- The scheduled catalog snapshot sync compares the new full snapshot count with the previous known-good snapshot before writing. A drop greater than 50% is rejected and the existing snapshot is preserved by the established error path.
- The threshold helper validates counts and its configured drop fraction; tests cover boundary acceptance, suspicious shrinkage, invalid settings and product/variant quality counters.
- Quality reporting is diagnostic only: product-only sources, zero-stock items, missing descriptions and uncategorized products are reported but not automatically deleted or fabricated.
- The continuity guard is intentionally attached to the current full-snapshot sync only. Future supplier adapters must distinguish complete snapshots from incremental/delta feeds before applying this policy.
- Catalog CI for the current implementation is pending. M08.5 real supplier onboarding remains blocked until a named, authorized supplier/feed and usage rights are approved.


## M08.4 final verification — 2026-11-02
- Catalog CI push run `41355552153` passed on code head `c3e5b1624223333f8db61af407fa43308d7eaf06`.
- Catalog CI PR run `41355595494` passed on PR head `0df4c5485e4a2eb391b61e13c52047ed4631942a`: TypeScript validation, all catalog/domain unit tests and production build succeeded.
- Backend CI PR run `41355595650` passed on the same PR head: SQL generation, Go formatting/module checks, vet, PostgreSQL tests, API/importer builds, Docker image build and container cleanup all succeeded.
- M08.4 is complete. The active roadmap stage is M08.5, authorized supplier onboarding; this is blocked on the owner selecting an approved supplier/feed and confirming its usage rights.


## Latest continuation checkpoint — 2026-11-02: M08.5 variant omission regression coverage
- Confirmed in `backend/internal/database/products.go` that `upsertProductVariants` only upserts supplied variants and does not delete database rows when `Product.Variants` is nil/empty. `syncPlatformVariantOffers` then reads existing persisted variants and reconciles the platform offers from those retained rows.
- Added PostgreSQL integration coverage in `TestMigrationsAndCatalogUpsert`: after a product+variant import, a later product-only snapshot is imported. The test asserts the original source variant remains, the GlobalCart Direct offer retains its price/stock/active status, a separate merchant offer remains active with its independent stock, and audit event counts do not increase.
- No importer deletion, automatic zero-stock adjustment, or offer archival was introduced. Omission is intentionally treated as “no new variant information,” not “delete.”
- Backend CI verification for this new regression coverage is pending. Latest code commit: `0288e6f94dc1c0b81cefa9b2ca6a244a92ac51d7`.
- M08.6 real supplier onboarding remains blocked until a named supplier and authorized API/feed access and reuse terms are approved.


## M08.5 verification and M08.6 continuation — 2026-11-02
- M08.5 variant omission and inventory retention is verified by Backend CI PR run `41356201201` on head `eca6f936d99582a83930f2040a46b912bba57225`. The complete PostgreSQL test suite, vet, API/importer build, Docker image build and cleanup passed.
- The new product-only-snapshot regression confirms that an omitted source variant remains in PostgreSQL, the GlobalCart Direct platform offer retains its price/stock/active state, another merchant's independent offer remains unchanged, and no extra price/inventory audit events are generated.
- M08.6 adds `TestVariantSourceIdentityCannotMoveBetweenProducts`: it imports an original source variant, attempts to reassign the same `(source, source_id)` to a second product in a multi-product batch, and verifies the conflict rejects the batch, rolls back the earlier row, and leaves the original variant's owner/SKU/price/stock intact.
- Latest test commit for M08.6: `499d76129d12f7af3fe3a030d37d15a78ba1a0da`. Backend CI verification pending.
- M08.7 real supplier onboarding remains blocked until a named supplier/feed and authorized data reuse are approved.


## M08.5 and M08.6 verification — 2026-11-02
- M08.5 product-only snapshot omission retention passed Backend CI PR run `41356201201` on head `eca6f936d99582a83930f2040a46b912bba57225`.
- M08.6 variant source identity ownership and batch rollback passed Backend CI push run `41356390061` on code head `499d76129d12f7af3fe3a030d37d15a78ba1a0da` and Backend CI PR run `41356442182` on head `3c9a01787420d56b281bae6f271eddbc56d3c57a`. Both runs passed Go vet, PostgreSQL race-enabled tests, API/importer builds and Docker image build.
- Catalog CI PR run `41356442235` also passed TypeScript validation, all unit tests and production build on head `3c9a01787420d56b281bae6f271eddbc56d3c57a`.
- A GitHub Actions formatting bot commit `259cd66e7b846e9e14c3baeeaa53b6cf19d94e72` applied the Go formatting changes produced by the successful push workflow; no behavior change was introduced.
- M08.7 now adds an explicit full-snapshot vs incremental-feed declaration to the adapter contract before any real supplier is connected. M08.8 real supplier onboarding remains blocked pending owner approval and authorized source access.


## M08.7 implementation checkpoint — 2026-11-02
- Added `CatalogSnapshotMode = "full" | "incremental"`; every `CatalogSourceAdapter` must now declare its feed completeness mode.
- DummyJSON is explicitly marked as a full snapshot source.
- `loadCatalogFromSource` supports an optional required mode and rejects mismatches before calling the source fetcher.
- Scheduled `scripts/sync-catalog.ts` now explicitly requires `full`; incremental adapters can still normalize records for a future dedicated merger, but cannot feed the static full-snapshot writer.
- Added a regression test that verifies an incremental adapter is rejected before fetch when full data is required, while the same records can be normalized through the generic loader without asserting any merge or persistence behavior.
- Catalog CI verification is pending.


## M08.7 final verification — 2026-11-02
- Catalog CI PR run `41356770768` passed on head `2176ad14a30f196f4ae1624aa4a17dbac69dc9bd`: npm dependency installation, TypeScript validation, all catalog/domain unit tests and production build succeeded.
- Backend CI PR run `41356770801` passed on the same head: SQL generation, Go formatting/module checks, vet, PostgreSQL-backed race tests, API/importer builds, Docker image build and cleanup succeeded.
- All source adapters must now declare `snapshotMode: "full" | "incremental"`. The scheduled snapshot writer requires `full` and rejects a mode mismatch before calling the source fetcher. The generic loader can normalize incremental records for a future merger, but there is no delta merger or persistence path.
- M08.8 authorized supplier onboarding is the next gated item; no live supplier endpoint, credential, or external commercial inventory feed has been connected.


### M08.8 feed mapping quality guard — implementation
- loadCatalogFromSource now rejects a feed when more than 20% of raw source records fail the adapter mapper by default.
- CatalogImportOptions.maxUnmappedFraction allows a source-specific override, validated as a finite value in [0, 1).
- The guard runs after fetching/mapping but before normalization; empty collections, wholly unmappable feeds and zero-valid-product feeds remain rejected.
- This protects against major schema drift or partial mapper failure without treating duplicate product identities as mapping failures.
- Authorized supplier onboarding is now M08.9 and remains blocked pending owner approval of a named source and usage rights.


### M08.8 feed mapping quality and request safety — CI verified
- `loadCatalogFromSource` rejects a source response when more than 20% of raw records fail source-specific mapping by default. `maxUnmappedFraction` permits a documented per-source override, validated as finite and in [0, 1).
- Empty collections retain their empty-feed error; wholly unmappable collections retain their dedicated unmappable-feed error.
- DummyJSON requests now use a 20-second timeout even when a caller provides an AbortSignal. The timeout and caller cancellation are combined; explicit caller cancellation exits immediately without retry.
- Regression coverage includes mostly-unmappable feeds, boundary overrides, invalid threshold values, preservation of the empty/unmappable errors, composed request signals and cancellation without retry.
- Verification on head `3db5764fa75f9188866392d3d3c874a5cb420684`: Catalog CI runs `41358732405` and `41358736412` passed; Backend CI run `41358736442` passed.
- Next stage is M08.9 authorized supplier onboarding. It remains blocked until the owner approves a named source, its API/feed access and catalog/image/price reuse terms.


### M08.8 retry policy — CI verified
- DummyJSON requests combine the caller's AbortSignal with a per-attempt 20-second timeout. Explicit caller cancellation exits immediately and is never retried.
- The adapter retries network failures/timeouts and HTTP 408, 429 or 5xx responses, with at most three attempts and bounded backoff.
- Permanent HTTP failures (other 4xx), invalid JSON/schema and empty collections fail fast. Mapping-quality rejection remains a separate shared-ingestion guard.
- Added tests for permanent 404 no-retry and transient 503 recovery within the bounded attempt count.
- Verification on head `a4a0d16c2aa06cfc83558b6516c9ac282cec1701`: Catalog CI run `41358956297` and Backend CI run `41358956380` both succeeded.
- M08.9 remains blocked until owner approval of a named supplier/API/feed and data/image/price reuse rights.


### M08.8.1 source-agnostic feed import diagnostics — CI verified
- Added `loadCatalogFromSourceWithReport`, returning normalized products plus safe aggregate counters: fetched, mapped, unmapped, normalized and discarded after normalization.
- Report also carries source, snapshot mode, import timestamp and mapping threshold. It does not include raw supplier records or sensitive payload fields.
- The existing `loadCatalogFromSource` remains backward-compatible and returns only products.
- Scheduled catalog sync logs the feed report alongside the existing catalog-quality summary.
- Tests verify mapping rejection counts and that malformed and duplicate records contribute to post-normalization discards.


## M08.8.2 — collision-safe atomic catalog publication (CI verified, 2026-11-02)
- Extracted writeCatalogSnapshotAtomically from the sync script. It serializes the complete snapshot before creating any temporary file, writes with exclusive-create semantics to a unique same-directory path, then renames it over the destination.
- Each invocation uses a process identifier plus a random UUID, preventing concurrent sync processes from sharing the same temporary path.
- Temporary files are removed in a finally cleanup after write/rename errors. A serialization error occurs before the existing destination is touched.
- Added regression tests for 12 simultaneous publishers, rename failure cleanup and preservation of an existing last-known-good file on serialization failure.
- Atomic rename guarantees complete-file visibility; it does not order source freshness across independent writers. The existing Pages workflow continues to serialize its scheduled/manual runs.
- Catalog CI [41360183218](https://github.com/pooya16pg-spec/globalcart/actions/runs/41360183218) and Backend CI [41360183283](https://github.com/pooya16pg-spec/globalcart/actions/runs/41360183283) both passed on code head `062aabdf563626116a50ecba0d7f534e6af3f260`, including TypeScript validation, unit tests, production build, Go vet/tests/builds and Docker image build.


## M08.8.3 — last-known-good snapshot validation (CI verified, 2026-11-02)
- Added readCatalogSnapshot to distinguish a genuinely missing file (ENOENT) from permission/I/O errors, malformed JSON and invalid snapshot schema.
- Snapshot validation checks versioned metadata, valid timestamp, exact product count, normalized product fields, optional variants and localization structure before the file is trusted as last-known-good.
- Invalid existing snapshots fail before a new feed fetch or fallback decision; only a genuinely absent snapshot follows first-import fixture behavior.
- Added regression coverage for missing file, malformed JSON, count mismatch and malformed normalized product.
- On implementation commit `ff5cf42cc094bc4e35d317ad52f970762280ba2f`, Catalog CI PR run [41361214644](https://github.com/pooya16pg-spec/globalcart/actions/runs/41361214644), Catalog CI push run [41361210412](https://github.com/pooya16pg-spec/globalcart/actions/runs/41361210412), and Backend CI PR run [41361214721](https://github.com/pooya16pg-spec/globalcart/actions/runs/41361214721) all passed, including TypeScript validation, unit tests, production build, Go vet/tests/builds and Docker image build.


## M08.8.4 — snapshot identity and database-bound validation (CI verified, 2026-11-02)
- Strengthened runtime snapshot validation to enforce source consistency, unique product IDs and unique product source identities across the entire snapshot.
- Added cross-product uniqueness checks for variant `(source, source_id)`, matching the PostgreSQL partial unique index and preventing an import batch from failing after snapshot publication.
- Added database-bound numeric validation: safe integer product IDs, non-negative two-decimal prices within NUMERIC(12,2), two-decimal discount/rating limits, percentage range 0–100, rating range 0–5 and non-negative 32-bit integer stock. Variant prices and stock use the corresponding database constraints.
- Added regression cases for product/variant identity conflicts, source mismatch, numeric precision and database range violations.
- Catalog CI PR run [41362606331](https://github.com/pooya16pg-spec/globalcart/actions/runs/41362606331), Catalog CI push run [41362602994](https://github.com/pooya16pg-spec/globalcart/actions/runs/41362602994), and Backend CI PR run [41362606267](https://github.com/pooya16pg-spec/globalcart/actions/runs/41362606267) all passed on implementation commit `e9d864f6bdbf0013137929e02b6258007200db29`, including TypeScript, unit tests, production build, Go vet/race tests, API/importer builds and Docker image build.

- The shared `assertCatalogSnapshot` contract runs before atomic publication as well as when reading the previous snapshot, so newly mapped invalid data is never written as the next last-known-good file.
- Non-empty trimmed variant SKUs are unique across the full snapshot, matching the PostgreSQL unique SKU index in addition to source-identity uniqueness.


## M08.8.5 — PostgreSQL rating domain constraint (CI verified, 2026-11-02)
- Added migration 007 with a database-level `rating <= 5` check so direct importer/API writes cannot bypass the canonical snapshot rating rule.
- Added PostgreSQL integration coverage using rating 5.01 in a multi-product import batch; the test asserts the earlier valid product is rolled back too.
- The migration is additive and forward-only; it does not rewrite existing ratings.
- Catalog CI PR run [41363342797](https://github.com/pooya16pg-spec/globalcart/actions/runs/41363342797), Backend CI PR run [41363342758](https://github.com/pooya16pg-spec/globalcart/actions/runs/41363342758), and Backend CI push run [41363337241](https://github.com/pooya16pg-spec/globalcart/actions/runs/41363337241) all passed on implementation commit `cc36339bcac77d08abc37af3f88f08607a090367`. PostgreSQL migrations, integration tests (including the isolated 5.01 rating rollback case), Go builds and Docker image build succeeded.


## M07.1 — Catalog facets verified (2026-11-02)
- The PostgreSQL read-only product list supports category, full-text query, normalized exact brand and inclusive source-price/rating bounds; list and total-count queries share the same predicates.
- Numeric filters reject malformed values, negatives, excess decimal precision, out-of-domain ratings and inverted min/max bounds before database access.
- Migration 008 adds supporting brand, price and rating indexes. OpenAPI documents the complete query contract.
- The Persian/English storefront derives its brand options from the loaded catalog and provides minimum-rating controls. Existing category, text and price filters continue to work; reset clears all filters.
- Local filtering logic is isolated in `lib/catalog/filtering.ts` and covered by focused regression tests.
- The latest implementation head `3372040d3cd71cd55d99724033065cf0dfa4537e` passed Catalog CI push run `41365136582`, Catalog CI PR run `41365141727` and Backend CI PR run `41365141780`.
- The frontend still uses the generated static catalog. The Go API is not hosted and has not been connected to the public storefront.


## M07.2 — Accessible suggestions and command palette (CI verified, 2026-11-02)
- app/page.tsx exposes bounded product, brand and category suggestions from the loaded catalog. The search field uses combobox/listbox semantics and supports keyboard navigation and selection.
- Ctrl/⌘+K opens a focus-trapped command dialog for catalog navigation, category browsing, cart, favorites, comparison, reset filters and language switching.
- lib/catalog/suggestions.ts owns deterministic NFKC-normalized ranking and de-duplication. tests/catalog-suggestions.test.ts covers ranking, normalization, result bounds and source-data non-mutation.
- Catalog CI push and PR checks passed on implementation commit 3d6b840f769a4c23de134800877de9c9b47e8298; the combined final PR head 87bc930da16367bd725b68fec460bae2fda6f26e also passed Catalog CI.

## M07.3 — PostgreSQL trigram typo tolerance (CI verified, 2026-11-02)
- Migration 009 enables pg_trgm and adds GIN expression indexes for normalized product titles and brands.
- backend/queries/catalog.sql retains simple-config full-text search as primary matching and adds trigram/word-similarity fallback for queries of at least three characters. Exact full-text matches rank above fuzzy matches.
- Product list and total-count filters use the same full-text/trigram predicate. OpenAPI documents the query length and typo-tolerance behavior.
- backend/internal/database/search_integration_test.go covers exact-match ordering, typo-tolerant title/phrase and brand search, and total-count consistency using isolated category fixtures.
- Final implementation head 87bc930da16367bd725b68fec460bae2fda6f26e passed Backend CI push 41366589804, Backend CI PR 41366594050 and Catalog CI PR 41366593947.
- The storefront remains static and local; this migration does not deploy or connect a public backend.


## M07.4 — Explainable catalog recommendations (CI verified, 2026-11-02)
- `lib/catalog/recommendations.ts` calculates a transparent demo score using 70% normalized product rating and 30% normalized discount contribution, with discount capped at 50%. The result is rounded to six decimal places for deterministic comparisons.
- Ranking tie-breakers use rating, lower sample price and product ID. The helper returns a new array and does not mutate catalog data.
- Product cards expose a concise reason label: strong rating, strong discount, both, or balanced signals. The storefront states that the ranking is not personalized and does not use browsing/purchase history.
- `tests/catalog-recommendations.test.ts` verifies score weights, discount cap, reason selection, deterministic ordering, non-mutation and invalid-signal bounds.
- Implementation head `09fd73ea1f1dcb1f01e5022c3f4380ec2832855a` passed Catalog CI push `41366975455`, Catalog CI PR `41366979976` and Backend CI PR `41366979999`.
- No user profile, browsing history or purchase history is used to rank catalog products.


## M07.5 — Browser-local saved searches and preferences (CI verified, 2026-11-02)
- Saved searches store the query, category, brand, min/max price, minimum rating and sort order, with an optional user-defined label. Users can apply or delete a saved entry from the catalog header.
- The collection is bounded to twelve entries and persisted only in localStorage under globalcart-saved-searches-v1. Existing cart/favorites storage remains separate.
- lib/catalog/saved-searches.ts validates and normalizes stored entries, rejects malformed values, and compares all filter/sort dimensions for duplicate detection. Tests cover bounds, malformed records and duplicate detection.
- The feature is browser-local and does not require login, backend hosting or remote synchronization.

## M07.6 — Catalog freshness and fallback states (CI verified, 2026-11-02)
- lib/catalog/source-status.ts models pending refresh, generated snapshot, live source, stale snapshot fallback and local fixture fallback.
- The storefront displays a loading indicator only when no generated snapshot is available. If a refresh fails or returns zero products, the last-known-good snapshot is retained and a visible status message explains the fallback.
- A fixture-only fallback is clearly distinguished from a generated snapshot. Empty filtered results continue to use the explicit no-results view.
- Tests cover every state transition; latest implementation head ff80ebcfcf4bcec4cdbbf9217c866070614524be passed Catalog CI push 41367502278, Catalog CI PR 41367539860 and Backend CI PR 41367539848.
- No source error path replaces a valid generated snapshot with an empty product collection.


## M08.1 — Playwright static storefront journeys (CI verified, 2026-11-02)
- `playwright.config.ts` builds and tests with `/globalcart/` as base path when running on GitHub Actions. `scripts/serve-static.mjs` safely serves only files beneath `out/`, rejects traversal and serves directory `index.html` output.
- `e2e/storefront.spec.ts` runs against Chromium desktop and Pixel 7 mobile, blocks external hosts, and covers locale/product navigation, search suggestion keyboard operation, Ctrl/⌘+K command palette, saved-search persistence/apply/delete, cart drawer and simulated checkout.
- Playwright runs after TypeScript validation, unit tests and a real `next build` with the GitHub Pages export configuration. Failure traces/screenshots and an HTML report are uploaded as a seven-day Actions artifact.
- Product detail tests assert the actual static route without assuming a trailing slash. The mobile cart trigger now has an explicit accessible label because its visible text is hidden at small breakpoints.
- Implementation head `08c86aeb909228ecc1ba03d35195ad4fba6ab987` passed Browser E2E push `41369733216`, Browser E2E PR `41369737767`, Catalog CI PR `41369737769` and Backend CI PR `41369737810`.
- All checkout and order assertions remain explicitly demo-only; no payment provider or production order API was added.


## M08.2 — Automated accessibility baseline (CI verified, 2026-11-02)
- `@axe-core/playwright` 4.14.0 is integrated in `e2e/accessibility.spec.ts` and audits the homepage, open command palette and static product detail page.
- The scan uses WCAG 2.0 A/AA, WCAG 2.1 A/AA and WCAG 2.2 AA tags in both Chromium desktop and Pixel 7 mobile projects.
- Contrast fixes darken the low-contrast supporting text and sample rating/discount accents while preserving the main green palette. FAQ summaries and footer links now expose larger touch/click targets.
- The cart's text label remains in the accessibility tree when visually hidden on mobile, allowing the native button name to include both cart label and count without an aria-label mismatch.
- Implementation head `e515d696873aa2811c3edb35fc89ace74ae874df` passed Browser E2E push `41370817992`, Browser E2E PR `41370823048`, Catalog CI PR `41370822987` and Backend CI PR `41370823026`.
- Automated checks are limited to the tested routes/states. Manual screen-reader and keyboard-only review remain open quality work.


## M08.3 — Keyboard-only modal focus verification (CI verified, 2026-11-02)
- `e2e/accessibility.spec.ts` verifies focus remains inside the Command Palette when Tab/Shift+Tab crosses its edges and returns to the invoking search control after Escape.
- `app/page.tsx` stores the active element when the palette opens via Ctrl/⌘+K or the command button, then restores focus after the dialog closes and is removed from the DOM.
- Browser E2E on final implementation head `8937b54f0f631695ee02e51a128cc42815988f43` passed Push `41375898314` and PR `41375903109` (18 tests passed). Catalog CI PR `41375903056` and Backend CI PR `41375903095` also passed.
- This is automated coverage for the Command Palette only. Other dialogs and manual screen-reader/browser review remain future QA.


## M08.4 — Image-loading and first-party asset baseline (CI verified, 2026-11-02)
- Homepage hero and product-detail main image use `loading="eager"`, `fetchPriority="high"` and `decoding="async"`. Other catalog, rail, recently viewed, editorial, gallery, related-product and modal images use lazy loading and async decoding.
- Fixed media containers reserve layout geometry while images are loaded. `e2e/performance.spec.ts` verifies the image-loading policy and first-party asset budgets on desktop Chromium and Pixel 7 mobile.
- CI baseline: 512,814 bytes JS / 7 requests and 67,702 bytes CSS / 1 request; enforced budgets: 750,000 bytes JS / 10 requests and 120,000 bytes CSS / 2 requests.
- Implementation head `a741cf60ab44a17acf31bba43f47eb40204f2be5` passed Browser E2E push `41376758145`, Browser E2E PR `41376763890` (24 tests), Catalog CI PR `41376763985` and Backend CI PR `41376763923`.
- E2E blocks external hosts deliberately. No claim is made about real CDN payload size, live image latency, LCP/CLS or external fonts; network-enabled measurement remains open.


## M05.1 — Internal full-order return and simulated refund (CI verified, 2026-11-02)
- Migration 010 adds one full-order return request per parent order, a fixed reason allow-list and reviewed/refunded timestamps.
- The internal Go service supports owner-scoped request, internal approve/reject review and owner-scoped completion of a simulated full-order refund.
- Approval snapshots the persisted net order total including coupon discount. Completion moves parent and vendor child orders to `refunded_simulated` and appends an audit event atomically.
- Requests/reviews/refunds are idempotent where applicable. Rejected returns and unpaid orders cannot be refunded.
- The feature is not exposed over HTTP and never calls a provider, moves money, or changes stock/reservations.
- Backend CI [41382309964](https://github.com/pooya16pg-spec/globalcart/actions/runs/41382309964), Security [41382310050](https://github.com/pooya16pg-spec/globalcart/actions/runs/41382310050) and Pages deployment [41382310024](https://github.com/pooya16pg-spec/globalcart/actions/runs/41382310024) passed on `bc97475c0cce7f8de6aebe72071813c942b6a3d0`.
