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


## 0.2.3-dev — 2026-11-02
- M09.3 owner-scoped vendor child-order retrieval is CI-verified in Backend CI run `41352898725` (head `1234d02`).
- The vendor child-order read requires the guest capability digest and masks cross-owner IDs as not found.
- Authentication and seller RBAC remain gated by same-site HTTPS, approved email delivery and secret-management prerequisites; no public seller endpoints were added.

- Added PostgreSQL integration coverage for cross-vendor offer reads, price updates and inventory adjustments; verifies a foreign vendor scope cannot mutate the owning vendor's stock. CI verification pending.

- M09.4 cross-vendor isolation tests passed Backend CI run `41353137589` on code commit `e18a01d`; the full workflow including Docker image build succeeded.
- Documented the interim seller-service boundary: a vendor ID is not caller authorization and public seller routes remain disabled until verified sessions and active memberships are enforced.


## 0.2.4-dev — 2026-11-02
- Added HTTP handler regression tests to keep seller order/product routes, cart writes and checkout writes unavailable until verified identity and authorization exist.
- Added tests that catalog routes reject mutation methods and configured CORS origins receive only read-only methods. Backend CI verification pending.
- Reconciled roadmap and project context: M09.3 and M09.4 are CI-verified; M09.4.1 is the active regression-guard slice.

- Tightened pre-auth CORS policy by no longer advertising the `Authorization` request header; added a regression assertion. Latest code CI verification pending.


## 0.2.5-dev — 2026-11-02
- Catalog imports now create/update the GlobalCart Direct base-product offer transactionally alongside product/category data.
- Added append-only price and inventory audit events for initial offers and subsequent source changes; unchanged imports remain idempotent.
- Added PostgreSQL integration coverage for initial, repeated and changed catalog imports. CI verification pending.
- Verified public API regression guards and pre-auth CORS restrictions in Backend CI run `41353463006`.

- Platform catalog synchronization now includes product variants, mapping source SKU, price, stock and active/paused state into vendor offers. Added PostgreSQL regression coverage for idempotency and price/inventory audit events. CI verification pending.


### Platform offer synchronization verified — 2026-11-02
- Backend CI run `41354208659` passed all steps for base-product and existing-variant platform offer synchronization, including PostgreSQL integration tests and Docker image build.

### End-to-end source variant ingestion — in progress
- Added optional source variants to the normalized TypeScript catalog contract and Go importer model.
- Importer now transactionally upserts variant identity, SKU, title, price, stock, attributes and active state before synchronizing the GlobalCart Direct offer.
- Added normalization and PostgreSQL integration coverage for variants delivered in the product snapshot.
- The current DummyJSON source does not provide variant records; ingestion is optional and awaits an authorized supplier source. CI verification pending.


### M08.2 source variant ingestion — CI verified (2026-11-02)
- Extended the normalized catalog snapshot and Go importer contract to carry optional source variants.
- Persisted variant source identity, SKU, title, price, stock, attributes and active state in the same import transaction as the parent product.
- Added PostgreSQL integration coverage for variant creation, idempotent re-import, changed offer price/stock audit events and inactive-state synchronization.
- Backend CI run `41354535515` passed all steps, including PostgreSQL-backed race tests, API/importer builds and Docker image build.

### M08.3 supplier adapter contract — implementation in progress (2026-11-02)
- Added a common adapter interface for fetching supplier-specific records and mapping them into the canonical catalog input.
- Refactored DummyJSON and scheduled snapshot generation to use the shared ingestion boundary while retaining retry and last-known-good behavior.
- Corrected variant de-duplication to use the same source/source-ID identity pair enforced by PostgreSQL.
- Added synthetic supplier fixture coverage and dedicated Catalog CI for typecheck, tests and production build.
- No real supplier feed is connected; current verification is pending the new Catalog CI run.


### M08.3 supplier adapter contract — CI verified (2026-11-02)
- Added the shared `CatalogSourceAdapter` interface and `loadCatalogFromSource` ingestion boundary.
- Refactored DummyJSON and scheduled snapshot sync to use the shared adapter while retaining retries and last-known-good snapshot protection.
- Aligned variant deduplication with the database's `(source, source_id)` identity pair.
- Added a synthetic-only supplier fixture with two variants and tests for source mapping, provenance, SKU, attributes, inactive state, source-scoped identity and empty/unmappable feed rejection.
- Added dedicated Catalog CI for TypeScript validation, catalog/domain unit tests and production build.
- Verification: Catalog CI push run [41355090849](https://github.com/pooya16pg-spec/globalcart/actions/runs/41355090849) and PR run [41355095558](https://github.com/pooya16pg-spec/globalcart/actions/runs/41355095558) both succeeded on head `9bfab0c80dec906f52fb56a22b85d36b7eec99f9`.

### M08.4 authorized supplier onboarding — blocked
- Next integration requires a named supplier and owner-approved API/feed access and usage rights.
- No real supplier connection, credentials, scraped inventory or fabricated live variants are included pending that decision.


### Variant identity collision hardening — 2026-11-02
- Replaced delimiter-concatenated variant identity keys with an unambiguous tuple encoding of `(source, sourceId)`.
- Added a regression test proving that pairs such as `("supplier:one", "id")` and `("supplier", "one:id")` remain distinct.
- Catalog CI run `41355272122` passed TypeScript validation, all unit tests and production build on head `eaebbc0a7b8d1c823c3dcad4a7fed0d10a99edc9`.


### M08.4 catalog snapshot integrity — implementation in progress (2026-11-02)
- Added a conservative full-snapshot continuity guard that rejects a product-count drop greater than 50% compared with the last-known-good snapshot.
- The guard executes before atomic snapshot replacement, so suspiciously incomplete refreshes preserve the existing catalog snapshot.
- Added a diagnostic report for product/variant coverage and common data-quality signals; these are reported without fabricating or automatically deleting catalog data.
- Added unit tests for threshold boundaries, invalid settings and quality-report counts. Final Catalog CI verification is pending.


### M08.4 catalog snapshot continuity and quality diagnostics — CI verified (2026-11-02)
- Added a full-snapshot continuity guard that rejects a normalized product-count drop greater than 50% compared with the last-known-good snapshot, before the new snapshot can replace it.
- Added a quality report covering total products/variants, product-only coverage, zero-price/zero-stock products, missing descriptions and uncategorized products.
- Diagnostic metrics do not fabricate inventory or auto-delete data; product-only sources are valid.
- Verification: Catalog CI push run [41355552153](https://github.com/pooya16pg-spec/globalcart/actions/runs/41355552153) and PR run [41355595494](https://github.com/pooya16pg-spec/globalcart/actions/runs/41355595494) passed. Backend CI PR run [41355595650](https://github.com/pooya16pg-spec/globalcart/actions/runs/41355595650) also passed SQL generation, Go/PostgreSQL tests, builds, Docker image build and cleanup.


### M08.5 variant omission and inventory retention — implementation in progress (2026-11-02)
- Added PostgreSQL integration coverage for a source that previously supplied variants but later sends a product-only snapshot.
- The regression test asserts that existing product variants are retained, platform offer values/status remain unchanged, independent merchant offers are untouched, and no spurious audit events are written.
- No automatic variant deletion, stock-zeroing or offer archival behavior was added. Backend CI verification is pending.


### M08.5 variant omission and inventory retention — CI verified (2026-11-02)
- Added PostgreSQL integration coverage for a source that later sends a product-only snapshot after previously supplying variants.
- Verified that existing variants and platform offers are retained; an independent merchant offer is unchanged and no spurious audit events are appended.
- Backend CI PR run [41356201201](https://github.com/pooya16pg-spec/globalcart/actions/runs/41356201201) passed SQL generation, Go vet, race-enabled PostgreSQL tests, API/importer builds, Docker image build and cleanup.

### M08.6 variant source identity ownership — implementation in progress (2026-11-02)
- Added a PostgreSQL regression test for attempted reassignment of a supplier's existing variant source identity to a different product.
- The test checks that the whole multi-product import batch rolls back and the original variant owner, SKU, price and stock remain intact.
- Backend CI verification is pending.


### M08.5 variant omission and inventory retention — CI verified
- Backend CI PR run [41356201201](https://github.com/pooya16pg-spec/globalcart/actions/runs/41356201201) passed the PostgreSQL regression proving product-only refreshes do not delete stored variants, alter platform/merchant offers, or append duplicate audit events.

### M08.6 variant source identity ownership and batch rollback — CI verified (2026-11-02)
- Added PostgreSQL integration coverage proving a source variant identity cannot be reassigned to a different product.
- Verified that the entire multi-product import rolls back on identity conflict and the original variant owner/SKU/price/stock remain unchanged.
- Backend CI push run [41356390061](https://github.com/pooya16pg-spec/globalcart/actions/runs/41356390061) and PR run [41356442182](https://github.com/pooya16pg-spec/globalcart/actions/runs/41356442182) passed, including race-enabled PostgreSQL tests and Docker image build.

### M08.7 explicit feed completeness semantics — implementation in progress
- Next adapter hardening will distinguish complete supplier snapshots from incremental/delta feeds and prevent deltas from replacing the last-known-good full catalog.


### M08.7 explicit feed completeness semantics — implementation in progress (2026-11-02)
- Added a required `snapshotMode` declaration to the source adapter contract: `full` or `incremental`.
- Classified DummyJSON as a full-snapshot source and required full mode in the scheduled static snapshot writer.
- Feed-mode mismatches are rejected before fetching records, preserving the last-known-good snapshot.
- Added a test proving an incremental feed can be normalized for a future merger but cannot enter the full snapshot path. Catalog CI verification is pending.


### M08.7 explicit feed completeness semantics — CI verified (2026-11-02)
- Added required `snapshotMode` metadata (`full` or `incremental`) to every catalog source adapter.
- Classified DummyJSON as a full snapshot and made the scheduled snapshot writer explicitly require that mode.
- Mode mismatches are rejected before fetching, protecting the last-known-good catalog from accidentally replacing it with a delta batch.
- Generic ingestion may normalize incremental records for a future dedicated merger, but this change does not merge or persist delta records.
- Verification: Catalog CI PR run [41356770768](https://github.com/pooya16pg-spec/globalcart/actions/runs/41356770768) and Backend CI PR run [41356770801](https://github.com/pooya16pg-spec/globalcart/actions/runs/41356770801) both passed on the same branch head, including production and Docker builds.


### M08.8 feed mapping quality and request safety — CI verified (2026-11-02)
- Added a default 20% maximum mapping-rejection fraction to shared ingestion, with a validated per-source override. Duplicate canonical identities remain handled by normal de-duplication.
- Preserved distinct errors for empty and wholly unmappable feeds so existing failure semantics remain clear.
- Combined caller cancellation with a 20-second DummyJSON request timeout; explicitly cancelled requests now stop without retry, while timeout failures retain bounded retry behavior.
- Added regression tests for mapping thresholds, invalid configuration, combined request signals and cancellation behavior.
- Verification on head `3db5764fa75f9188866392d3d3c874a5cb420684`: Catalog CI [41358732405](https://github.com/pooya16pg-spec/globalcart/actions/runs/41358732405) and [41358736412](https://github.com/pooya16pg-spec/globalcart/actions/runs/41358736412) passed; Backend CI [41358736442](https://github.com/pooya16pg-spec/globalcart/actions/runs/41358736442) passed.
- M08.9 authorized supplier onboarding remains blocked pending explicit owner approval of a named source and usage rights.
