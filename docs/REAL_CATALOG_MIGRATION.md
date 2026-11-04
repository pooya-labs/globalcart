# Real-product catalog migration gate

**Status:** Investigation only — no real-retailer feed is connected.  
**Checked:** 2026-11-04  
**Scope:** GlobalCart portfolio simulation; no real orders, payments, inventory reservations or fulfillment.

## Owner requirement

Replace obviously synthetic/sample listings with recognizable, genuine commercial products: authentic product names, model identifiers, brands and source-provided images. Keep the storefront experience and visual identity GlobalCart-owned. The local cart and checkout remain simulated.

The UI may omit a provider's technical platform branding where the source terms permit it. It must not falsely claim that a third-party seller owns, stocks or fulfills a product. Seller/source attribution and outbound links must remain wherever the provider's rules or license require them. A hidden provider logo is not permission to hide a merchant relationship.

## Findings

### Shopify is not a universal public product catalog

Shopify's Storefront API is scoped to a specific `{shop}.myshopify.com` store. Its tokenless mode can query that shop's products, collections, search and cart, but it does not expose a global directory of products from all Shopify merchants. Token-based access is required for additional features. See the [official Storefront API reference](https://shopify.dev/docs/api/storefront/latest) and [Storefront API guide](https://shopify.dev/docs/storefronts/headless/building-with-the-storefront-api).

Therefore, Shopify cannot currently replace DummyJSON as a universal, no-account catalog for GlobalCart. Do not hard-code a random merchant's shop domain or present its inventory as GlobalCart-owned.

### Curated real-product API candidate: Daily Dose Tech

[Daily Dose Tech's public API](https://dailydose.tech/docs) advertises read-only JSON endpoints with no authentication or API key. Its documented catalog contains actual named technology products (examples include LG OLED M5, Sony WH-1000XM6, Anker Prime and Keychron Q1 Pro), plus curated product descriptions, retailer links, price history and deal data. The API supports pagination, category and text search, and asks consumers to cache for at least one hour.

This is a materially better lead than a synthetic product generator for a **technology-only collection**, but it is not a general retail catalog. The public API documentation does not clearly establish reusable product-image URLs or image licensing, and the price/availability freshness and redistribution terms need confirmation. It also states that the publisher may earn affiliate commissions. Therefore:
- Do not yet use it as a full catalog or copy its content/images into GlobalCart.
- Do not remove retailer/source links or imply that GlobalCart is the seller.
- Treat it as a candidate for a curated electronics/technology section only after image rights, permitted caching/display, attribution and commercial-data semantics are confirmed with the publisher.

### Amazon and large retailer feeds

Amazon is not connected. Do not scrape product pages or copy product images/prices. A future adapter is gated on the owner's eligibility for the relevant official program/API, credentials, permitted display/storage duration, image handling, price/availability refresh, attribution and linking rules. Affiliate discovery is not wholesale inventory access.

Other major retailer and marketplace APIs commonly require a developer/partner account, app credentials, an approved affiliate relationship, or a seller/store identifier. A source must be evaluated from its official documentation and actual access response before it is described as public or integrated.

### Public open-data alternatives are narrower

Open Food Facts provides genuine packaged-food product records and product images through a public API, but it is a food-product database, not a general merchandise retailer catalog. It could support a dedicated grocery/food category or a separately labelled open-data demo, subject to its current API usage and image/database licenses. It is not a credible replacement for electronics, fashion, home goods and general retail all at once.

## Integration acceptance criteria

A source may replace the current sample adapter only after all of these are confirmed:

1. **Real identity:** recognizable commercial products with manufacturer/model identifiers where available; no generated product names, brands, prices or stock.
2. **Rights:** documented permission for every field used, including image hotlinking, caching, transformations and display.
3. **Access:** stable, documented API/feed access that works from GitHub Actions or an approved server-side runtime without scraping or bypassing controls.
4. **Freshness semantics:** clearly documented update cadence and source timestamps; never label a fetch timestamp as the merchant's last modification time.
5. **Commercial facts:** price, currency, availability, seller and variant data are shown only when actually provided and permitted. Missing fields stay missing.
6. **Attribution:** retain any required merchant/source credit and product link. Do not disguise a third-party seller as GlobalCart.
7. **Operational safety:** bounded retries, timeouts, schema validation, mapping-rejection threshold, full/incremental declaration, last-known-good snapshot and tests.
8. **Demo boundary:** cart/checkout remain local simulations; no claim of purchase, stock reservation or fulfillment.

## Current implementation decision

- Keep the existing DummyJSON adapter and snapshot pipeline intact as an explicitly identified **temporary sample source** until a qualifying real-product feed is available.
- Do not silently promote sample products to real products or merely rename their provenance.
- Keep the shared `CatalogSourceAdapter` as the only ingestion extension point; do not create a parallel catalog pipeline.
- Do not integrate Shopify's Storefront API until a specific store domain and permission to use its catalog are supplied.
- Do not replace the broad marketplace catalog with a food-only source without an explicit product-scope decision.
- Once a qualifying source is selected, implement it behind the existing adapter, add source-specific contract tests and update `docs/DATA_SOURCES.md`, `ROADMAP.md` and `PROJECT_CONTEXT.md`.

## Next technical action

Verify whether Daily Dose Tech can explicitly permit the product images and data fields needed for a demo integration. In parallel, continue source discovery against official documentation, prioritizing a genuinely public, broad-category product dataset with explicit commercial-product identity and image-use terms. If no such source meets the criteria without credentials, report that constraint rather than shipping a fake "live retailer" integration. A real retailer catalog may require the owner to authorize a developer/affiliate account or provide a store/feed endpoint.
