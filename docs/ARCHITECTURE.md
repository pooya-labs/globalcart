# GlobalCart Architecture

## Current product boundary
GlobalCart is an independent portfolio marketplace simulation. The current public preview is a Next.js static export on GitHub Pages. Cart, favorites and checkout are local/demo state. No real payment, fulfillment or seller payout is performed.

## Current catalog and price data flow (ABO snapshot)

```text
Amazon Berkeley Objects dataset (selected 1,000-item sample)
        |
        v
scripts/sync-catalog.ts
  - fetches dataset rows and original-image metadata server-side
  - lib/catalog/abo.ts maps records and deterministic demo-price estimates
  - downloads/optimizes original gallery images to local WebP
        |
        +----> data/catalog.json (product metadata + provenance)
        |
        +----> public/images/catalog/*.webp (self-hosted galleries)
        |
        v
Next.js static export on GitHub Pages
        |
        +----> local search, filters, comparison and product routes
        |
        +----> browser-local bag / favorites / browsing history / demo checkout

TGJU public USD board
        |
        v
scripts/sync-tgju-fx.ts -> data/fx-rate.json
        |
        v
USD estimate -> indicative Toman conversion + freshness disclosure
```

The current storefront reads the committed local catalog snapshot; it does not fetch a retailer feed in the visitor's browser. The ABO source supplies product metadata and images, not verified seller prices, ratings or stock. GlobalCart adds stable `priceKind: "estimated"` amounts for the demo while retaining `priceAvailable: false`. TGJU is only the USD-to-Toman conversion source and does not validate a product price. Shopify Global Catalog remains a separate Cloudflare Worker integration/smoke path and is not wired to the active storefront. A failed scheduled ABO refresh preserves the last-known-good catalog snapshot.

## Database schema reference

The actual PostgreSQL schema is defined by migrations 001–015 (27 tables). See [DATABASE_ERD.md](DATABASE_ERD.md) for domain-grouped Mermaid ERDs and the declared foreign-key, composite-key and ownership constraints. The diagrams describe the local-development/CI schema and do not imply a hosted production database.

## Backend foundation (main; local development only)

The merged `main` branch contains a runnable Go API and PostgreSQL 17 Compose environment. The API applies embedded versioned migrations on startup, exposes liveness/readiness probes and read-only catalog list/detail endpoints, and uses a local CLI to import the normalized static catalog snapshot. OpenAPI is maintained in `backend/openapi.yaml`. The Compose ports bind to loopback; this is not a public or production deployment. The branch now also includes normalized catalog categories and product variants while preserving the current product API shape. The repository uses a sqlc-generated pgx/v5 typed catalog query layer. The embedded migration runner applies one file per transaction under a PostgreSQL advisory lock and records exact-byte SHA-256 checksums; it rejects changed or missing applied migration files. See docs/DATABASE_MIGRATION_POLICY.md for immutable migration, forward-fix and recovery rules. Migration 011 provides identity/session/RBAC storage scaffolding, but authentication, write endpoints and hosted infrastructure remain gated future work. PostgreSQL integration tests cover concurrent/idempotent migrations, checksum adoption/drift, catalog upserts and transactional rollback.

## Target platform architecture (M04+)

```text
Next.js / React / TypeScript
        |
        | HTTPS REST (OpenAPI)
        v
Go API
        |
        +------ PostgreSQL (system of record)
        |          |
        |          +-- versioned SQL migrations + sqlc typed queries
        |
        +------ Redis (optional cache / queue)
        |
        +------ Python workers
                   |
                   +-- supplier ingestion
                   +-- translation / enrichment
                   +-- data quality jobs
```

This target architecture is not deployed. GitHub Pages cannot host the Go API, PostgreSQL or Python workers. M04 requires owner approval for runtimes and separate hosting.

## Technology decisions
- Frontend: Next.js App Router, React and TypeScript.
- Backend: Go REST API with OpenAPI.
- Primary database: PostgreSQL for relational catalog, variants, users, roles, carts, orders and audit events.
- Data access: reviewed SQL migrations and sqlc-generated typed methods; parameterized SQL only.
- Cache/queue: Redis only when measured need exists; never canonical storage.
- Data/AI jobs: Python workers for ingestion, enrichment and translation, with idempotency, retries and provenance.
- Search: PostgreSQL full-text/trigram first; evaluate a dedicated engine only after measurement.
- API style: REST first; GraphQL only with demonstrated need.

## Current code boundaries
| Module | Responsibility |
|---|---|
| app/page.tsx | Storefront and local demo interactions |
| app/globals.css | Responsive design system, RTL/LTR |
| data/catalog.json | Static catalog snapshot contract |
| scripts/sync-catalog.ts | Scheduled/manual ABO snapshot and original-gallery generation |
| lib/catalog/abo.ts | Active ABO listing mapper and deterministic demo-price estimates |
| lib/catalog/types.ts | Normalized product contract |
| lib/catalog/normalize.ts | Shared validation and deduplication primitives |
| lib/catalog/dummyjson.ts | Legacy/sample adapter; not the active storefront source |
| lib/catalog/shopify.ts | Separate Shopify Worker smoke/integration mapping; not the active storefront feed |
| scripts/sync-tgju-fx.ts | Scheduled/manual TGJU USD quote snapshot |
| lib/pricing.ts | Displayable-price provenance, integer-cent math and fixed-point FX |
| lib/translation/* | Translation provider contract; full-catalog translation is paused by decision |
| tests/*.test.ts | Domain unit tests |

## Catalog and translation contracts
Connectors preserve stable source ID, source name, source URL and import timestamp. Invalid records are rejected and duplicates are removed. Translation is a separate server-side/scheduled pipeline; preserve original and translated strings, locales, provider, status and timestamp. Missing credentials or provider failures must never silently replace source text.

## Pricing contract
Catalog USD prices remain quantized to integer cents; FX uses fixed-point micro-units and BigInt intermediates. The UI imports the free-market USD quote extracted from TGJU's public `قیمت دلار` page into `data/fx-rate.json`. TGJU's `price_dollar_rl` row is quoted in Rial; GlobalCart divides by ten for Toman-per-USD, preserving both the raw Rial value and the Tehran source-clock timestamp. Freshness is 5 minutes fresh / 15 minutes stale. This is a public board indication, not a guaranteed executable buy/sell or bank rate. Stale values are labelled and no hard-coded fallback exists. A polite scheduled workflow fetches the page server-side; the static browser never calls TGJU directly.

## Static hosting
GitHub Pages serves the Next.js export under /globalcart. No secrets or server-only code may enter the client bundle. Scheduled ingestion generates JSON before build. A failed refresh preserves last-known-good data.

## Commerce simulation
Until M04/M05, cart, favorites and checkout are browser-local. Future persisted payment/order records remain simulated and never trigger real payment or fulfillment.


## Internal authentication primitives

`backend/internal/auth` implements Argon2id password hash/verify using bounded PHC parsing, a rehash decision helper, and 256-bit opaque-token generation/digest helpers. The package is not connected to HTTP handlers; the public API remains read-only and ADR-010 deployment gates still apply.

## Current backend read API
- `GET /healthz` — process liveness.
- `GET /readyz` — PostgreSQL readiness.
- `GET /api/v1/products` — bounded, searchable and category-filtered product list.
- `GET /api/v1/products/{id}` — product detail.
- `GET /api/v1/products/{id}/variants` — active variants with stock and JSON attributes.
- `GET /api/v1/categories` — normalized categories with parent IDs and product counts.

All HTTP endpoints are read-only. Migration 011 defines identity storage tables, but no accounts are collected, no sessions are issued and no authentication or write operations are exposed.


Product list pagination returns both `count` (items in the current page) and `total` (all matching products before limit/offset). Category filters match normalized category names case-insensitively and trim surrounding whitespace. Both list and count use the same search/category predicates.


## Demo commerce schema (local-only, not exposed)
Migration 003 introduces carts, cart items, demo orders, immutable order-line snapshots, expiring inventory reservations and order events. Guest ownership and checkout idempotency persist only SHA-256 digests of high-entropy capabilities/keys. The intended checkout is a single PostgreSQL transaction that locks the cart and relevant stock row, revalidates prices from canonical catalog data, writes immutable snapshots and creates an expiring reservation. Payment is simulated and reservation release never mutates imported stock. See `docs/COMMERCE_DESIGN.md`. No write endpoints or browser cookie flow are enabled until the same-site HTTPS, identity, authorization and CSRF gates are approved.


The internal commerce service now implements demo checkout transactionally in Go. It is not exposed by HTTP. Checkout locks the cart and canonical stock row, revalidates price/title, accounts for active unexpired reservations, and atomically creates order snapshots, reservations and an event. Same-cart idempotent retries return the existing order. PostgreSQL integration tests cover concurrent stock contention and replay (CI run 41345901209).


## Supplier ingestion boundary (M08.3)
A `CatalogSourceAdapter` owns supplier-specific fetching and mapping. It emits raw records in the canonical catalog input shape; `loadCatalogFromSource` applies shared normalization, provenance and identity de-duplication before data can become a snapshot. Empty, malformed and wholly unmappable feeds fail closed, allowing the snapshot job to preserve its last-known-good file. Variant identity is source-scoped as `(source, sourceId)`, consistent with the PostgreSQL unique key.

This section documents the historical M08.3 supplier-adapter foundation, not the active storefront source. DummyJSON remains a legacy/sample adapter; the active storefront catalog is the committed ABO snapshot described above. A synthetic supplier fixture tests the shape of a future authorized feed, but is never used as storefront inventory. Real supplier connectors remain gated by explicit authorization and terms review.
