# Changelog

## 0.1.3 — 2026-11-01

### Added
- Server-side Gemini 2.5 Flash product translation adapter with strict JSON validation.
- Optional scheduled Persian catalog localization job, safely skipped when GEMINI_API_KEY is absent.
- Per-product translation provenance, source-content fingerprint caching, sequential throttling, bounded retries and error states.
- Storefront uses verified Persian product fields when available while preserving original English fields.

### Verification
- CI run 41295683515 passed catalog sync, optional translation step (safely skipped because no key is configured), TypeScript validation, unit tests, production build and GitHub Pages deployment.
- Run: https://github.com/pooya16pg-spec/globalcart/actions/runs/41295683515

## 0.1.2 — 2026-11-01

### Added
- Scheduled and manually triggerable catalog refresh in GitHub Actions.
- Typed static catalog snapshot generator with retries, validation and atomic replacement.
- Last-known-good snapshot protection when the upstream sample API is unavailable.
- Storefront initialization from the generated catalog snapshot.
- Target architecture decision: Go REST/OpenAPI, PostgreSQL with SQL migrations/sqlc, Python data workers and optional Redis.

### Verification
- Actions run 41295039930 passed catalog sync (100 products), TypeScript validation, all unit tests, production build, Pages artifact upload and deployment.
- Run: https://github.com/pooya16pg-spec/globalcart/actions/runs/41295039930

## 0.1.1 — 2026-11-01

### Changed
- Reworked demo pricing to use integer USD cents and fixed-point FX units.
- Used BigInt intermediate arithmetic to avoid floating-point drift during conversion.
- Explicit rounding: discounted USD rounds to cents, then Toman rounds half-up to a whole unit.
- Rejected non-finite discounts and unsafe input/output ranges.
- Expanded pricing tests for fractional values, rounding, clamping and invalid inputs.
- Removed the duplicate Pages starter workflow.

### Verification
- Actions run 41294484252 passed dependency installation, TypeScript validation, unit tests, production build, artifact upload and Pages deployment.

## [0.2.0-dev] - 2026-11-01
### Changed
- Re-scoped the storefront toward a full marketplace experience inspired by common Amazon-style discovery and shopping patterns, with an independent GlobalCart identity.
- Paused translation and removed the translation job and credentials from the GitHub Pages deployment workflow.
- Added a dedicated quick-category navigation rail and began the visual-system migration toward a cleaner, denser marketplace layout.

### Added
- Responsive catalog-driven category mega-menu with live category counts, direct filtering, backdrop dismissal and mobile layout.
- Quick category rail now reflects categories present in the current catalog rather than only the initial hand-picked list.

## 0.2.1-dev — 2026-11-01
### Added
- Catalog-driven category mega-menu and dynamic quick-category rail.
- Dedicated static product routes with image gallery, product attributes, source disclosure, related products and demo cart actions.
- Side-by-side comparison shortlist for up to four products.
- Product cards now navigate directly to dedicated detail pages while retaining quick view.
### Fixed
- Corrected escaped comparison-state declarations and duplicate category-name key found during CI validation.
- Made catalog snapshot commits rebase on the latest main branch to avoid races with UI commits.


## 0.2.2-dev — 2026-11-01
### Added
- Validated Amazon product-link intake with supported marketplace domains, ASIN extraction and a browser-local sourcing request queue.
- Marketplace service shortcuts for sourcing requests, product comparison, planned seller tools and GitHub feedback.
- FAQ content clarifying demo checkout, Amazon data access and comparison behavior.
- Responsive styles for the Amazon request module, service cards and FAQ.
- Keyboard-accessibility improvements: Escape dismissal, dialog semantics, visible focus indicators, live feedback announcements and localized quantity labels.
### Notes
- Amazon link submission is a local demonstration only. Live product details, pricing and availability require authorized Amazon API access.
- Translation remains paused and excluded from CI.
### Verification
- Actions run 41299580437 (Run #134) passed catalog refresh, TypeScript validation, unit tests, production build, Pages artifact upload and deployment.
- Run: https://github.com/pooya16pg-spec/globalcart/actions/runs/41299580437


## 0.2.3-dev — 2026-11-01
### Added
- Data-driven homepage merchandising rails for top discounts, top-rated products and fresh discoveries.
- Browser-local recently viewed product rail, updated from both quick view and dedicated product pages, with a clear-history action.
- Product detail price breakdown showing source USD price, catalog discount, illustrative Toman/USD rate and calculated demo total.
### Notes
- FX remains user-configured and illustrative; no live provider, timestamp, shipping, tax or markup is represented.
- Recently viewed data is stored locally in the browser; no account-level history is created.
### Verification
- Actions run 41299322140 (Run #126) passed catalog refresh, TypeScript validation, unit tests, production build, Pages artifact upload and deployment.
- Run: https://github.com/pooya16pg-spec/globalcart/actions/runs/41299322140


### Recently viewed page follow-up
- Added a dedicated /recently-viewed route with empty state, clear-history action and Persian/English layout.
- Added a permanent footer entry so the history page remains discoverable even before the first product visit.
- Actions run 41299677941 (Run #135) passed TypeScript validation, unit tests, production build, Pages artifact upload and deployment.
- Run: https://github.com/pooya16pg-spec/globalcart/actions/runs/41299677941


### Accessibility interaction pass
- Added Escape-key dismissal across homepage overlays and navigation menus.
- Added dialog roles, modal semantics and accessible title associations to cart, favorites, quick view, comparison and demo checkout.
- Added aria-expanded/aria-controls relationships for mobile and category navigation; gallery thumbnails now expose aria-pressed.
- Full focus trapping/restoration and manual assistive-technology QA remain open.

- Added modal focus management: initial focus, Tab/Shift+Tab trapping and focus restoration when dialogs close.


### M04 backend foundation — 2026-11-01
- Added a Go 1.23 REST API with health/readiness endpoints and read-only catalog list/detail routes.
- Added PostgreSQL 17 local Docker Compose, loopback-only port bindings and persistent development storage.
- Added embedded transactional SQL migrations for the initial products catalog schema.
- Added a one-shot CLI to import the existing normalized catalog snapshot into PostgreSQL.
- Added an OpenAPI 3.1 contract and backend CI for formatting, vet, race-enabled tests, builds and container build.
- This is local development infrastructure only. No hosted database/API, credentials, paid resources, authentication, real orders or payments were provisioned.

- Backend CI run #22 passed: PostgreSQL service startup, Go formatting, vet, race-enabled unit/integration tests, API/importer builds and Docker image build. Run: https://github.com/pooya16pg-spec/globalcart/actions/runs/41301389812


### M04 catalog platform extension — 2026-11-01
#### Added
- Normalized category table with product foreign-key backfill and category product counts.
- Product variants schema with SKU/source identity constraints, non-negative price/stock and JSONB attributes.
- sqlc v2 generation targeting PostgreSQL and pgx/v5; generated typed product, category and variant queries.
- Product list/detail handlers now use generated sqlc methods; importer retains explicit transactional upsert logic.
- Read-only product-variant endpoint and normalized category listing endpoint.
- Authentication/session/RBAC design covering Argon2id, opaque server-side sessions, CSRF, cookie policy, role scopes and required security tests.

#### Verification
- Actions run 41302783939 passed sqlc generation, Go formatting, vet, unit and PostgreSQL integration tests, API/importer builds and Docker image build.
- Run: https://github.com/pooya16pg-spec/globalcart/actions/runs/41302783939
- Authentication remains design-only pending approval of same-site HTTPS hosting, email delivery and secret management. No user tables or login flows were added.


### Catalog pagination follow-up — 2026-11-01
- Category filtering now matches normalized category names case-insensitively and ignores surrounding whitespace.
- Added a typed total-count query sharing the same search and category predicates as product listing.
- Product list response now includes `total` for pagination while `count` continues to report items in the returned page.
- Added PostgreSQL integration coverage for category normalization and total count. Pull-request CI run 41303108214 passed sqlc generation, formatting, vet, race-enabled PostgreSQL tests, API/importer builds and Docker image build.


### Demo checkout service — 2026-11-01
- Added an internal Go transaction service for demo checkout with canonical catalog revalidation, row-level stock locks, expiring reservations, immutable order snapshots and idempotent replay.
- Added PostgreSQL integration coverage for competing carts, replay and idempotent reservation expiry; CI run 41346127244 passed.
- Service is not exposed as an HTTP endpoint; no real payments or live stock adjustments are performed.


### Guest cart lifecycle — 2026-11-01
- Added internal guest cart creation, item add/increment, quantity update, item removal and cart retrieval.
- Enforced digest-only guest capability inputs, per-cart transactional locking, canonical catalog price/title/stock reads and bounded quantities.
- Added PostgreSQL integration tests for lifecycle operations, variants, stock limits and expiry.
- Verified by CI push run 41346723656; no public cart endpoints or browser cookie flow enabled.


### Order detail and cancellation — 2026-11-01
- Added capability-scoped internal order detail reads with immutable item snapshots and event history.
- Added owner-scoped demo cancellation, atomic reservation release, cancellation event and idempotent replay.
- Added PostgreSQL integration tests for cross-cart isolation, order snapshots, event history and cancellation behavior.
- Verified by CI push run 41347325289; no public order endpoints enabled.


### Simulated payment lifecycle — 2026-11-02
- Added an owner-scoped, idempotent internal demo payment completion transition.
- Requires a live stock reservation; atomically releases holds, updates order state and appends a simulation event.
- Added PostgreSQL integration coverage for ownership, repeat completion, reservation release and invalid post-payment cancellation.
- No external payment provider or public payment endpoint is involved; Verified by CI run 41348079983, including PostgreSQL integration/race tests and Docker image build.


### Internal coupon foundation — 2026-11-02
- Added hashed coupon-code storage and internal coupon creation/application/removal methods.
- Added percentage and fixed discount calculation, validity windows, minimum subtotal, maximum discount and redemption limits.
- Checkout revalidates and records redemption atomically; order totals snapshot subtotal, discount and final total.
- Added cart coupon state and discounted order detail fields; no public routes or verified customer identity are enabled.


### M08 vendor foundation — 2026-11-02
- Added migration 005 with vendor lifecycle/type, vendor-specific base and variant offers, price history and inventory event ledger.
- Backfilled existing catalog products and variants into the active GlobalCart Direct platform vendor; recorded opening price/stock snapshots.
- Added PostgreSQL integration coverage for independent stock across two vendors, audit records and duplicate-offer constraints.
- Compatibility preserved: existing cart and checkout still use legacy catalog stock; multi-vendor cart/order splitting is not yet implemented.
- No seller membership, public seller API, dashboard, commission, payout or fulfillment was enabled.


### M09 — Vendor-aware cart foundation
- Added additive vendor-specific cart line and parent/child vendor-order schema.
- Added internal vendor-aware cart operations and legacy-checkout safety guard.
- Added integration coverage and architecture notes; split checkout remains pending.


### M09.2 — Vendor split checkout
- Implemented atomic parent/child order creation and per-vendor reservations.
- Added parent-level coupon redemption with exact child allocation.
- Propagated simulated payment, cancellation and expiry to child orders.
- Added PostgreSQL integration tests for checkout and inventory contention.


### M09.3 — Owner-scoped vendor child-order reads — 2026-11-02
- Added `GetVendorOrder` for reading a single vendor child order and its immutable item snapshots through the owning guest cart capability digest.
- Kept unknown child IDs and cross-owner access indistinguishable via `ErrOrderNotFound`.
- Added PostgreSQL integration coverage for child metadata, parent linkage, item snapshots, exact totals and cross-owner isolation.
- Backend CI verification is pending; see PR #1 before treating this slice as complete.


### M09.3 CI correction — 2026-11-02
- Inspected Backend CI run 41352752412 and identified the exact `go vet` failure: an unused `time` import in `backend/internal/commerce/vendor_orders.go`.
- Removed the unused import in commit `cb0f20816afdbf3f2dc23628e8a5bd4b7d1bce46`.
- The original failed run predates the correction. A fresh CI run is still required; M09.3 remains unverified until it passes.
