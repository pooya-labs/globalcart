# GlobalCart — Master Roadmap

> The single source of truth. A stage is complete only after implementation and verification.

**Current release:** 0.2.3-dev  
**Active stage:** M01.5 — Marketplace redesign and UX foundation  
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

**Runtime gate:** Local Go/PostgreSQL development and CI are permitted on the isolated backend branch. Public API/database hosting, Python worker deployment, production secrets and paid services remain gated on explicit owner approval.

## Release and stage map

| Stage | Target | Scope | Status |
|---|---:|---|---|
| M00 | 0.1.0 | Product definition, independence, governance and handoff | Complete |
| M01 | 0.1.x | Storefront foundation, bilingual shell and deployment | Implemented; redesign supersedes original visual direction |
| M02 | 0.2.x | Automated catalog import, normalization and snapshots | Catalog sync active; translation paused |
| M03 | 0.3.0 | Pricing/FX provenance, product details, gallery and comparison | Product detail, gallery, comparison and illustrative price breakdown implemented; live FX provenance remains |
| M04 | 0.4.0 | Go API, PostgreSQL, migrations, auth boundary and Python workers | Planned; owner runtime/hosting gate |
| M05 | 0.5.0 | Persistent demo commerce journeys, orders, coupons and returns | Planned |
| M06 | 0.6.0 | Admin/seller workspaces, import operations and analytics | Planned |
| M07 | 0.7.0 | Advanced search, relevance, recommendations and saved searches | Planned |
| M08 | 0.8.0 | Browser tests, accessibility, security and performance | Planned |
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
- [x] Search, categories, filters and sorting.
- [x] Quick view, favorites and recently viewed.
- [x] Local cart and simulated checkout.
- [x] Owner attribution and demo/source disclosures.
- [x] GitHub Actions build, typecheck, tests, artifact upload and Pages deploy passed in run 41294484252.
- [ ] Visual QA on phone, tablet and desktop.
- [ ] Keyboard, focus, screen-reader, contrast and reduced-motion audit.
- [ ] Verify published URL, base-path assets and deep links.

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
- [ ] Catalog freshness and source diagnostics.
- [ ] Add other suppliers only through authorized APIs/feeds.
**Acceptance:** products import automatically, the storefront survives source outages, and original source text remains intact. Translation is out of scope until explicitly reactivated.

## M03 — Pricing and product experience
- [x] Integer-cent pricing, fixed-point FX and BigInt intermediate arithmetic.
- [x] Explicit rounding and invalid-input tests.
- [ ] FX provider, source timestamp and stale-rate policy.
- [x] Transparent illustrative source-price, discount and FX breakdown on product pages (browser-local rate only; no live FX, markup, shipping or tax quote).
- [x] Dedicated static product details for the imported catalog, responsive image gallery and related products.
- [ ] Attributes, ratings context and source provenance.
- [x] Product comparison modal and persistent comparison shortlist.
- [x] Dedicated recently viewed page, homepage rail and browser-local visit history.

## M04 — Backend and database platform
- [ ] Owner enables Go runtime and approves separate API hosting.
- [ ] Owner enables Python runtime or an approved worker runner.
- [x] Add local-only PostgreSQL 17 Docker Compose service with loopback-only port binding and persistent development volume.
- [ ] Provision hosted PostgreSQL and production secret management after owner selects provider, region and budget.
- [x] Define initial catalog schema with provenance, numeric constraints, JSONB images and category/search indexes.
- [x] Extend catalog schema with normalized categories, product-to-category relationships and constrained product variants; users, roles, carts, orders and audit events remain pending.
- [x] Add embedded, versioned SQL migrations and transactional migration ledger for the initial catalog schema.
- [ ] Define reviewed rollback/forward-fix policy for shared environments.
- [x] Add Go REST API foundation, OpenAPI 3.1 contract, health/readiness probes, CORS allow-list and read-only catalog list/detail endpoints.
- [x] Validate catalog pagination bounds and query/category lengths with stable 400 error codes.
- [x] Document the API error response shape and list endpoint validation in OpenAPI.
- [x] Add PostgreSQL integration coverage for category backfill and product variant JSONB persistence.
- [x] Add sqlc configuration, parameterized catalog/variant query definitions and CI code generation for pgx/v5.\n- [x] Wire generated sqlc queries into product list/detail API handlers; importer remains transaction-oriented and uses explicit upsert SQL.
- [ ] Design authentication/session and role-based authorization.
- [ ] Add idempotent Python ingestion/translation workers.
- [ ] Add Redis only for measured cache/queue requirements.
- [x] Add Go unit tests and backend CI for formatting, vet, race tests, builds and Docker image build.
- [x] Add PostgreSQL integration tests for concurrent/idempotent migrations, catalog upserts and transaction rollback.
- [ ] Add authorization integration tests after the authentication boundary exists.

## M05 — Commerce simulation
- [x] Add local schema foundation for guest carts, bounded cart items and expiring capability digests (migration 003; API not exposed).
- [x] Internal multi-step checkout foundation: cart validation, stock reservations, immutable order snapshots and idempotent checkout (demo address book remains pending).
- [x] Define demo-only order states and reservation release semantics; no payment provider is contacted.
- [x] Add demo order ledger, immutable line snapshots, per-cart idempotency and append-oriented events (schema only; API not exposed).
- [ ] Coupons and promotions.
- [ ] Returns/refund simulation.

## M06 — Admin and seller workspaces
- [ ] Demo role switcher with real authorization boundaries where API exists.
- [ ] Admin overview and catalog/translation/price diagnostics.
- [ ] Product create/edit/archive.
- [ ] Supplier connector configuration without secret exposure.
- [ ] Seller profile, product management and demo analytics.

## M07 — Search and discovery
- [ ] Faceted filters for brand, rating, price and category.
- [ ] Accessible suggestions and keyboard command palette.
- [ ] PostgreSQL full-text/trigram baseline and relevance tests.
- [ ] Explainable ranking and recommendations.
- [ ] Saved searches and preferences.
- [ ] Empty, loading, offline and error states.

## M08 — Quality and hardening
- [ ] Playwright tests for critical shopping journeys.
- [ ] RTL/LTR accessibility, keyboard, labels and contrast audit.
- [ ] Performance and image-loading audit.
- [ ] Dependency, secret, static-analysis and API security checks.
- [ ] Database backup/restore and migration verification.
- [ ] Reproducible build and deployment smoke test.

## M09 — Portfolio release
- [ ] Persian and English case study.
- [ ] Screenshots, architecture/ERD diagrams and feature tour.
- [ ] Versioned 1.0.0 release and changelog.
- [ ] README verified against actual behavior.
- [ ] Final public demo smoke test.

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
- [ ] Integrate catalog importer so newly imported products create/update platform offers transactionally.
- [ ] Implement internal vendor and offer services with transactional price/inventory audit writes.
- [ ] Migrate cart lines, reservations and checkout to vendor-offer identity; split each checkout into vendor order groups.
- [ ] Add verified seller membership and tenant-scoped authorization only after the approved identity/authentication boundary.
- [ ] Keep seller APIs, seller UI, commission, payouts and fulfillment disabled until their separate security and product requirements are approved.

**Compatibility note:** existing cart/checkout still use legacy product/variant stock. Migration 005 is additive and intentionally does not claim multi-vendor checkout is complete.


### M09 — Vendor-aware cart foundation
- [x] Add vendor offer identity to cart lines without changing legacy line ownership.
- [x] Support internal add, quantity update and removal of vendor-specific offers.
- [x] Snapshot canonical offer price and vendor identity; serialize cart/offer mutations with row locks.
- [x] Add parent/child vendor order schema and vendor-aware order-item/reservation references.
- [x] Prevent legacy checkout from mispricing or reserving vendor cart lines.
- [ ] Implement atomic split checkout, per-vendor reservations and coupon allocation.
- [ ] Propagate cancellation, expiry and simulated payment state to vendor child orders.
