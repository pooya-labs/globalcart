# GlobalCart — Project Context / Single Source of Truth

**Owner:** Pooya Goodarzi  
**Repository:** https://github.com/pooya16pg-spec/globalcart  
**Branch:** feat/m04-go-postgres-foundation (M04 backend foundation; PR pending)  
**Package version:** 0.2.3-dev  
**Active stage:** M08.4 authorized supplier onboarding (blocked pending owner-approved source/feed)  
**Last updated:** 2026-11-02

## Product definition
GlobalCart is an independent marketplace portfolio simulation. Its information architecture is inspired by large retail marketplaces and the Amazon-link sourcing flow of kiyandaria.com, while its brand, UI and implementation remain original. It is not a live retailer. No real payments, orders, inventory reservations, fulfillment, seller payouts or refunds occur.

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
- `app/page.tsx` — bilingual storefront, catalog search/filter/sort, category mega-menu, Amazon URL request intake, local cart/favorites/compare, quick view and simulated checkout, service shortcuts and FAQ.
- `app/product/[id]/page.tsx` — statically generated product routes from the current catalog snapshot.
- `app/product/[id]/ProductDetailClient.tsx` — product gallery, illustrative price breakdown, product facts, browser-local visit tracking, local cart/favorite actions and related products.
- `app/recently-viewed/page.tsx` — static-exportable history page reading and clearing browser-local recent product IDs.
- `app/globals.css` — responsive marketplace design system, product detail, price breakdown, recently viewed rail, mega-menu, comparison and Amazon intake styling.
- `app/layout.tsx` — metadata and global styles.
- `data/catalog.json` — generated 100-product sample catalog snapshot.
- `scripts/sync-catalog.ts` — scheduled/manual DummyJSON import, retry, normalization and last-known-good protection.
- `lib/catalog/*` — canonical catalog contracts, normalization and source adapter.
- `lib/pricing.ts` — integer-cent/fixed-point demo pricing with BigInt.
- `lib/translation/*` and `google-apps-script/*` — parked legacy translation implementation; not used by storefront or CI.
- `tests/*.test.ts` — catalog, pricing and legacy translation tests.
- `.github/workflows/deploy.yml` — CI, scheduled/manual catalog refresh, static build and Pages deployment.
- `ROADMAP.md` — authoritative milestone checklist.
- `backend/` — Go REST API, PostgreSQL migrations, local Docker Compose, catalog importer and OpenAPI contract.
- `go.mod` — Go module at repository root.
- `docs/AI_HANDOFF.md` — continuation contract.

## Verified behavior
- Catalog snapshot pipeline imports 100 sample products and preserves a last-known-good snapshot.
- The storefront supports Persian RTL and English LTR, category navigation, search, filters, sorting, favorites, recently viewed, local cart and simulated checkout.
- Product detail routes are statically generated for products in the snapshot and include gallery, pricing, facts and related products.
- Keyboard support includes Escape dismissal for overlays, dialog semantics, visible focus styles and live announcements for key feedback states.
- Comparison supports up to four products and is persisted in browser local storage.
- Homepage merchandising rails are derived from the current sample catalog (discounts, ratings and discovery).
- Homepage overlay interactions support Escape dismissal; major drawers/modals expose dialog semantics and labelled headings. Modal dialogs now trap keyboard focus and restore it to the invoking element when closed; manual assistive-technology QA remains outstanding.
- Recently viewed product IDs are stored in browser local storage, shown in a homepage rail and on /recently-viewed; the footer links to the history page. There is no server-side account history.
- Product detail pages expose source USD price, sample discount, illustrative user-configured FX rate and converted demo price. No live FX provider is connected.
- Amazon URL intake validates supported Amazon domains and ASINs, then saves requests in browser local storage. Live Amazon data is not fetched.
- Translation is out of scope and removed from the active CI pipeline.

## Deployment
Public preview: https://pooya16pg-spec.github.io/globalcart/  
Repository Actions: https://github.com/pooya16pg-spec/globalcart/actions

Actions run 41299677941 (Run #135) passed catalog refresh, snapshot persistence, TypeScript validation, unit tests, production static export, Pages artifact upload and GitHub Pages deployment. This run includes the dedicated recently viewed page, homepage history rail and browser-local history controls.
https://github.com/pooya16pg-spec/globalcart/actions/runs/41299677941

## Immediate next actions
1. Complete responsive, RTL/LTR, keyboard, screen-reader, reduced-motion and published deep-link QA.
2. Add an authorized FX provider with source timestamp and stale-rate policy; keep the current rate clearly illustrative until then.
3. Improve source attributes, rating context and catalog freshness diagnostics.
4. Build seller/admin workspaces only after the catalog UX and role boundaries are defined.
5. Defer live Amazon product retrieval until approved API/feed credentials and usage rights are available.

## Backend/database architecture
- Frontend: Next.js App Router, React and TypeScript, statically exported to GitHub Pages.
- Backend target: Go REST API with OpenAPI.
- Primary database: PostgreSQL with versioned SQL migrations and sqlc.
- Python: ingestion and enrichment workers; translation remains paused.
- Redis: optional cache/queue only when measured need exists.
- Search: PostgreSQL full-text/trigram first.
- M04 requires owner approval for API hosting, database and worker runtime.



## M04 backend foundation (branch work; not yet merged)
- Branch: `feat/m04-go-postgres-foundation`.
- Go 1.23 API uses the standard HTTP mux and pgx v5.
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
- This change is on `feat/m04-go-postgres-foundation`; CI verification is pending.


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
- Branch remains `feat/m04-go-postgres-foundation`; PR #1 remains Draft/open/unmerged.
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
- Active branch: `feat/m04-go-postgres-foundation` (PR #1 remains draft; do not merge until review and CI verification).
- M09.2 vendor split checkout is implemented: one parent order, one child per vendor, per-offer reservations, parent coupon redemption with exact child allocation, and lifecycle propagation for simulated payment/cancellation/expiry.
- Latest implementation adds `GetVendorOrder`, an owner-capability-scoped read for a single vendor child order with immutable item snapshots. Unknown IDs and wrong-owner IDs return the same not-found error.
- Added `vendor_orders_integration_test.go` for PostgreSQL-backed ownership and snapshot assertions.
- Backend CI run `41352752412` failed at `go vet` because `vendor_orders.go` imported unused `time`; the exact failure is recorded in its job log.
- Removed the unused import in commit `cb0f20816afdbf3f2dc23628e8a5bd4b7d1bce46`.
- **Immediate next action:** obtain a fresh CI run for the corrected branch head. The earlier failed run does not verify the fix; do not mark M09.3 complete until a fresh run passes sqlc generation, vet, PostgreSQL-backed race tests, builds and Docker image build.
- No public commerce or seller HTTP routes, browser session flow, real payments, hosted API/database, or production secrets are enabled.


## Latest continuation checkpoint — 2026-11-02 (CI verified)
- Branch: `feat/m04-go-postgres-foundation`; PR #1 remains Draft/open/unmerged.
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
- PR #1 remains Draft, open and unmerged. See `ROADMAP.md` for the next checkpoint.
- M09.5 is blocked pending approval of same-site HTTPS hosting, runtime, secret management and email provider. No deployment or paid service was enabled.


## Canonical continuation checkpoint — 2026-11-02
- M09.3 owner-scoped vendor child-order reads: CI verified by run `41352898725`.
- M09.4 cross-vendor offer read/price/inventory isolation: CI verified by run `41353137589` on code commit `e18a01d1a9414520334ceb312c2207990ccc71d5`; Docker build passed.
- Seller/auth implementation remains gated by same-site HTTPS hosting, runtime approval, secret management and an approved email provider. `vendorID` is never proof of caller identity.
- New in-progress regression tests in `backend/internal/httpapi/handler_test.go` assert that seller/order/cart/checkout write routes are not exposed and catalog routes reject mutation methods. Commit: `e91132997a9053943bc2e6c3f8dbe82d59f7d9b5`. CI pending.
- Latest roadmap checkpoint: `ROADMAP.md`, stage M09.4.1. Do not treat the new HTTP guard tests as verified until their CI run completes.
- PR #1 remains Draft/open/unmerged. No production hosting, payment provider or paid infrastructure enabled.

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
- Active branch remains `feat/m04-go-postgres-foundation`; PR #1 remains Draft/open/unmerged. Do not merge or deploy.


## Latest continuation checkpoint — 2026-11-02: M08.2 verified; M08.3 in progress
- Backend CI run `41354535515` completed with conclusion `success` on head `11132964594083998362a8ef1e35e646524b7e6f`. All steps passed, including sqlc generation, formatting, vet, race-enabled Go/PostgreSQL tests, API/importer builds and Docker image build.
- M08.2 end-to-end source variant ingestion is therefore CI-verified. The TypeScript catalog contract and Go importer accept optional source variants; the importer persists variant identity, SKU, title, price, stock, attributes and active state transactionally before syncing platform offers.
- M08.3 introduces `CatalogSourceAdapter` and `loadCatalogFromSource` as a reusable supplier ingestion boundary. Source adapters fetch supplier-specific records and map them into the canonical product/variant input contract.
- DummyJSON now uses the shared adapter contract. Its retry behavior and the snapshot generator's last-known-good protection are retained.
- Variant deduplication now keys by both source and source ID, matching the database unique identity pair. Equal source IDs from different suppliers are not incorrectly collapsed.
- Added a synthetic-only supplier fixture and tests for mapping two variants, source provenance, SKU, attributes, inactive status, source-scoped identity and empty/unmappable feed rejection.
- Added `.github/workflows/catalog-ci.yml` to run npm install, TypeScript validation, unit tests and a production build for relevant pushes and pull requests. Its result is the remaining M08.3 verification gate.
- No real supplier API/feed, credentials, live supplier inventory, hosting, secrets, merge or deployment were added. PR #1 remains Draft/open/unmerged.


## Latest continuation checkpoint — 2026-11-02: M08.3 CI verified
- M08.2 source variant ingestion was verified by Backend CI run `41354535515` on head `11132964594083998362a8ef1e35e646524b7e6f`.
- M08.3 supplier adapter contract is implemented and verified. Catalog CI push run `41355090849` and PR run `41355095558` both succeeded on head `9bfab0c80dec906f52fb56a22b85d36b7eec99f9`; TypeScript validation, all catalog/domain unit tests and production build passed.
- A first M08.3 run exposed only an overly narrow test-message regex (14/15 tests passed); the assertion was corrected and the fresh runs passed. No production code workaround or skipped test was used.
- Shared types: `CatalogSourceAdapter` / `CatalogSourceRecord`; shared loader: `loadCatalogFromSource`; DummyJSON now goes through the shared boundary.
- Variant de-duplication follows the database identity pair `(source, source_id)`. A synthetic-only supplier fixture validates mapping of two variants, provenance, SKU, attributes and inactive state. Empty and unmappable feeds are rejected.
- The snapshot sync continues to preserve its last-known-good snapshot on failed ingestion. Product-only sources remain valid; absence of variants in a feed is not a command to delete existing database variants or offers.
- M08.4 is blocked on the owner selecting and approving a named supplier/feed and its rights. No real supplier connection, API key, live inventory, hosting, secrets or deployment was introduced.
- Branch head at this checkpoint: `9bfab0c80dec906f52fb56a22b85d36b7eec99f9`. PR #1 remains Draft/open/unmerged.
