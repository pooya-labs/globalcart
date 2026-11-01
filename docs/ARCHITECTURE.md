# GlobalCart Architecture

## Product boundaries
GlobalCart is a standalone portfolio marketplace simulation. It has no dependency on other portfolio repositories and no real commerce backend. GitHub Pages serves a static export. Browser state is local and demonstrative.

## Target data flow

```text
Permitted demo/source APIs
        |
        v
Typed source connector adapters
        |
        v
Normalization + validation + deduplication
        |
        +----> Original source text / provenance
        |
        v
Translation provider adapter (server-side scheduled job)
        |
        v
Versioned normalized catalog JSON
        |
        +----> Pricing / FX assumptions
        +----> Search and discovery
        |
        v
Static storefront (Next.js export)
        |
        v
Local cart / favorites / simulated checkout
```

## Current code boundaries

| Module | Responsibility |
|---|---|
| app/page.tsx | Storefront composition, client-side interactions, cart/favorites/checkout |
| app/globals.css | Responsive design system, RTL/LTR layout |
| lib/catalog/types.ts | Canonical normalized product contract |
| lib/catalog/normalize.ts | Input validation, normalization, deduplication |
| lib/catalog/dummyjson.ts | DummyJSON source adapter |
| lib/catalog/index.ts | Public catalog exports |
| lib/pricing.ts | Pure demo price calculation and validation |
| lib/translation/types.ts | Provider contract and localization provenance |
| lib/translation/index.ts | Safe localization orchestration; no false translated state |
| tests/*.test.ts | Node test-runner coverage for core domain logic |

## Connector contract
Each source adapter must return normalized catalog records with stable source ID, source name, source URL, import timestamp, original currency and source price. Invalid required fields are rejected. Duplicate records are keyed by source plus source ID. Keep adapters independent of UI components.

## Translation contract
Translation is a separate pipeline, not an inline browser request per product. Preserve original strings. Store translated value, source locale, target locale, provider, translation status and timestamp. Missing credentials, quotas or provider failures must produce an explicit unavailable/error state and never silently replace source text. Secrets are only available in scheduled/server-side jobs.

## Pricing contract
Pricing is deterministic over source price, discount and configured FX assumptions. USD is quantized to integer cents, FX is represented in fixed-point micro-units, and BigInt intermediates avoid floating-point drift in the conversion path. The calculation rounds discounted USD to cents, then rounds converted Toman half-up to a whole Toman. Current UI FX is illustrative only; it is not a live quote. Future FX adapters must preserve source, timestamp and stale-rate status.

## Static hosting
GitHub Pages serves the Next.js static export under /globalcart. Do not add server-only runtime features to the client bundle. Scheduled ingestion may generate a JSON artifact committed or uploaded as a build artifact; secrets must not be bundled into public JavaScript.

## Backend evolution gate
Static GitHub Pages cannot run a persistent API server, PostgreSQL instance or Python worker. Keep the current frontend independent from future services through typed API/data contracts. When durable multi-user state, authenticated roles, background translation or SQL-backed search becomes necessary, obtain owner approval for Go/Python runtime access and select a separate API/database host before implementation. Do not simulate server persistence as if it were real.

## Commerce simulation
Cart, favorites and checkout are browser-local. Never call payment gateways, reserve inventory, dispatch shipments, create seller payouts or imply a real order.
