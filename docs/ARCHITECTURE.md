# GlobalCart Architecture

## Target flow

~~~text
Public/demo catalog sources
          |
          v
   Source connector adapters
          |
          v
   Product normalization ---- Translation adapter
          |                         |
          +------------+------------+
                       v
                 Catalog model
                       |
            +----------+----------+
            |                     |
            v                     v
       Pricing engine       Search/discovery
            |                     |
            +----------+----------+
                       v
                 Storefront UI
                       |
            Cart / Favorites / Checkout
                       |
                Local demo state

Future GitHub Actions: scheduled import -> generated catalog JSON -> static deployment
~~~

## Current runtime
- Next.js App Router, React, TypeScript, client-side interactive storefront.
- React state for transient interface state.
- localStorage for demo cart/favorite preferences only.
- Typed demo catalog adapter and local fallback.
- Clearly labelled illustrative FX rate.
- Local-only checkout simulation.
- GitHub Pages serves static export; there is no production backend.

## Planned boundaries
**Catalog connector:** returns records and provenance, independent of UI. Retain source name, source ID, original URL, currency, source price, image URL, and fetched timestamp.

**Translation:** provider adapter preserves original text and stores translated text separately with provider, locale, timestamp, and status. If unavailable, show original text with a clear language label.

**Pricing:** deterministic function over source price, demo FX snapshot, margin, and illustrative fees. Use safe decimal/integer arithmetic and show assumptions.

**Commerce:** no payment, stock reservation, payout, or shipment API.

**Hosting:** static export at /globalcart/. Server-only capabilities require adapters and explicit limitations until a backend is approved.
