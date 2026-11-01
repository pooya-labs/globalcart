# GlobalCart — Project Context / Single Source of Truth

**Owner:** Pooya Goodarzi  
**Repository:** https://github.com/pooya16pg-spec/globalcart  
**Branch:** feat/m04-go-postgres-foundation (M04 backend foundation; PR pending)  
**Package version:** 0.2.3-dev  
**Active stage:** M01.5 marketplace redesign / M03 product experience  
**Last updated:** 2026-11-01

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
