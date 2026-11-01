# GlobalCart Architecture

## Current product boundary
GlobalCart is an independent portfolio marketplace simulation. The current public preview is a Next.js static export on GitHub Pages. Cart, favorites and checkout are local/demo state. No real payment, fulfillment or seller payout is performed.

## Current catalog data flow

```text
DummyJSON public demo API
        |
        v
scripts/sync-catalog.ts (scheduled/manual GitHub Actions)
        |
        v
Typed normalization + validation + deduplication
        |
        +----> source identity / URL / import timestamp
        |
        v
data/catalog.json (static snapshot)
        |
        v
Next.js static storefront
        |
        v
Browser-local cart / favorites / simulated checkout
```

The browser may attempt a live catalog refresh. If it fails, the storefront keeps the generated snapshot; local fixtures are used only if no snapshot exists.

## Backend foundation (M04 branch; local development only)

The branch `feat/m04-go-postgres-foundation` adds a runnable Go API and PostgreSQL 17 Compose environment. The API applies embedded versioned migrations on startup, exposes liveness/readiness probes and read-only catalog list/detail endpoints, and uses a local CLI to import the normalized static catalog snapshot. OpenAPI is maintained in `backend/openapi.yaml`. The Compose ports bind to loopback; this is not a public or production deployment. The branch now also includes normalized catalog categories and product variants while preserving the current product API shape. The repository now defines a sqlc-generated pgx/v5 typed catalog query layer; API wiring is the next step. Authentication, write endpoints and hosted infrastructure remain future work. PostgreSQL integration tests currently cover concurrent/idempotent migrations, catalog upserts and transactional rollback.

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
| scripts/sync-catalog.ts | Scheduled/manual source import and snapshot generation |
| lib/catalog/types.ts | Normalized product contract |
| lib/catalog/normalize.ts | Validation and deduplication |
| lib/catalog/dummyjson.ts | Live demo source adapter |
| lib/pricing.ts | Integer-cent pricing and fixed-point FX |
| lib/translation/* | Translation provider contract and honest unavailable/error states |
| tests/*.test.ts | Domain unit tests |

## Catalog and translation contracts
Connectors preserve stable source ID, source name, source URL and import timestamp. Invalid records are rejected and duplicates are removed. Translation is a separate server-side/scheduled pipeline; preserve original and translated strings, locales, provider, status and timestamp. Missing credentials or provider failures must never silently replace source text.

## Pricing contract
USD is quantized to integer cents, FX uses fixed-point micro-units and BigInt intermediates. Current UI FX is illustrative, not a live quote. Future feeds retain provider, timestamp and stale-rate status.

## Static hosting
GitHub Pages serves the Next.js export under /globalcart. No secrets or server-only code may enter the client bundle. Scheduled ingestion generates JSON before build. A failed refresh preserves last-known-good data.

## Commerce simulation
Until M04/M05, cart, favorites and checkout are browser-local. Future persisted payment/order records remain simulated and never trigger real payment or fulfillment.


## Current backend read API
- `GET /healthz` — process liveness.
- `GET /readyz` — PostgreSQL readiness.
- `GET /api/v1/products` — bounded, searchable and category-filtered product list.
- `GET /api/v1/products/{id}` — product detail.
- `GET /api/v1/products/{id}/variants` — active variants with stock and JSON attributes.
- `GET /api/v1/categories` — normalized categories with parent IDs and product counts.

All endpoints are read-only. Authentication, user data and write operations are not yet implemented.


Product list pagination returns both `count` (items in the current page) and `total` (all matching products before limit/offset). Category filters match normalized category names case-insensitively and trim surrounding whitespace. Both list and count use the same search/category predicates.
