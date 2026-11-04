# Real-product catalog migration gate

**Status:** Investigation only — no real-retailer feed is connected.  
**Checked:** 2026-11-04  
**Scope:** GlobalCart portfolio simulation; no real orders, payments, inventory reservations or fulfillment.

## Owner requirement

Replace obviously synthetic/sample listings with recognizable, genuine commercial products: authentic product names, model identifiers, brands and source-provided images. Keep the storefront experience and visual identity GlobalCart-owned. The local cart and checkout remain simulated.

The UI may omit a provider's technical platform branding where the source terms permit it. It must not falsely claim that a third-party seller owns, stocks or fulfills a product. Seller/source attribution and outbound links must remain wherever the provider's rules or license require them. A hidden provider logo is not permission to hide a merchant relationship.

## Findings

### Shopify Global Catalog: strongest broad-category live candidate

Shopify now documents two distinct catalog interfaces. The **Global Catalog** searches across eligible products from Shopify merchants; the **Storefront Catalog** remains scoped to one known merchant. The Global Catalog uses Shopify's UCP/MCP endpoint and an agent profile. The official catalog overview labels the auth model as an agent profile with no API key, but the hands-on [authentication tutorial](https://shopify.dev/docs/agents/get-started/authentication) explicitly instructs developers to create an API key in the Dev Dashboard and exchange its Client ID/Secret for a bearer token. Treat credential-free access as **unverified/ambiguous**, not as an established no-account path. It supports search, lookup and product-detail retrieval, including seller information and merchant checkout links. See the [official catalog overview](https://shopify.dev/docs/agents/catalog), [search tutorial](https://shopify.dev/docs/agents/get-started/search-catalog), [authentication tutorial](https://shopify.dev/docs/agents/get-started/authentication) and [agent-profile documentation](https://shopify.dev/docs/agents/profiles).

Important constraints:
- Shopify says not to cache catalog search results and not to cache, download, or re-host merchant product images. Images must be rendered in real time in connection with the related merchant listing.
- Keyless requests are rate-limited; the docs say keyless access does not support rate-limit increases.
- The API is designed for AI-agent/UCP commerce flows. A production web storefront integration still needs a technical proof of concept for browser/server transport, CORS, profile hosting, request limits and stable response mapping.
- Results represent third-party merchant offers. Keep seller identity/domain, seller price/availability context and merchant outbound/checkout URL visible where appropriate. Do not imply GlobalCart owns or fulfills those goods.
- This source cannot feed the existing static catalog snapshot or be treated as a persistent local product database. A compliant integration would be a separate **runtime discovery layer** that fetches fresh results and images for each view; the local demo cart must remain clearly simulated and must not reserve merchant inventory or place orders.

This is materially more promising than the Storefront API for broad real-product discovery, but it is not yet integrated. Do not start implementation until a small proof of concept resolves the official documentation's authentication discrepancy: verify whether a truly credential-free Global Catalog request is supported, or whether Dev Dashboard credentials and a bearer token are mandatory. Also verify a hosted public agent profile, real response shape, seller/price/availability fields, image rendering, CORS/server transport and practical rate limits. If credentials are mandatory, stop for owner approval before adding that dependency.

The Storefront API remains a separate option only when a specific merchant store and permission are known. It is not a global product directory. Do not hard-code a random merchant or present its inventory as GlobalCart-owned.

### Curated real-product API candidate: Daily Dose Tech

[Daily Dose Tech's public API](https://dailydose.tech/docs) advertises read-only JSON endpoints with no authentication or API key. Its documented catalog contains actual named technology products (examples include LG OLED M5, Sony WH-1000XM6 and Anker Prime), plus curated product descriptions, retailer names/links, price history and deal data. The API supports pagination, category and text search, and asks consumers to cache for at least one hour. Its about page discloses affiliate commissions.

This is a strong candidate for a **technology-only discovery collection**, not a general retail catalog. Its public API documentation does not clearly establish reusable product-image URLs or image licensing, and does not prove that listed retailer prices or stock are live at the time GlobalCart displays them. Therefore:
- Do not yet import it or copy its content/images into GlobalCart.
- Preserve retailer names and outbound links if the source is eventually used.
- Treat its prices as publisher-reported editorial data with a source timestamp, not guaranteed current retailer prices.
- Ask the publisher to confirm image rights, permitted caching/display, attribution and redistribution terms before integration.

### Open Food Facts: genuine barcode products, but food-only

[Open Food Facts API v3](https://openfoodfacts.github.io/documentation/docs/Product-Opener/api/) exposes real packaged-food products by barcode and product search. Read requests do not require an API key, but the project asks API consumers to use a descriptive User-Agent and complete its API usage form. The published limits include 15 product reads/minute/IP and 10 search requests/minute/IP; it specifically discourages search-as-you-type. API v3 is current, while v2 structured search remains available but is deprecated for new integrations.

Licensing is unusually clear but has obligations: the database is ODbL, individual contents use DbCL, and product images are CC BY-SA. Attribution/share-alike requirements and possible third-party graphical rights must be reviewed. The project also warns that community-contributed data can be incomplete or inaccurate.

This can provide authentic food/grocery SKUs and real package photos, but it is not a general retailer inventory feed: it does not provide reliable live retailer price, stock, cart or fulfillment. It is a valid **food-only open-data source candidate** if GlobalCart can satisfy the licenses and show source attribution.


### Shopify Global Catalog runtime integration — implementation decision

GlobalCart now uses Shopify Global Catalog MCP as a **runtime discovery source**, not as a persisted catalog feed. The public static storefront calls a small Cloudflare Worker proxy that forwards only validated `search_catalog` requests to Shopify. The proxy does not store results, cache responses, download images or expose credentials. Product image URLs are rendered directly from the live response; merchant names and seller listing links are shown in the UI. Search results are held only in React memory. Live product objects are excluded from localStorage persistence; only the demo cart's non-live legacy entries can be retained.

The Worker is restricted to the canonical GitHub Pages origin (plus localhost for development), enforces query/limit validation, and returns no-store headers. Its public endpoint is `https://globalcart-catalog.pooyagame141.workers.dev/api/catalog`. The profile is served by the same Cloudflare Worker at `https://globalcart-catalog.pooyagame141.workers.dev/agent-profile.json` (the source copy remains in `public/ucp/agent-profile.json`).

The local cart and checkout remain simulated: no merchant inventory is reserved, no Shopify cart/checkout is created, and no payment or fulfillment occurs. The seller listing link is the only route to a merchant's actual listing. Shopify branding is not used as storefront branding, but seller attribution is retained.

### Open Facts sibling databases: beauty, pet food and other products

The official Product Opener documentation identifies three sibling databases that use related API infrastructure:
- [Open Beauty Facts](https://world.openbeautyfacts.org): cosmetics and personal-care products.
- [Open Pet Food Facts](https://world.openpetfoodfacts.org): pet-food products.
- [Open Products Facts](https://world.openproductsfacts.org): products outside the food, beauty and pet-food scopes.

The official guide calls these APIs experimental/specific and notes that several shared features are not yet supported consistently (including Search-a-licious and Open Prices). They are category databases, not retailer storefront feeds. Their existence therefore improves the discovery map but does not establish broad fashion/home/toys coverage, current seller prices or stock, or unrestricted image reuse. Each dataset's current terms, image rights, completeness, endpoint behavior and rate limits must be checked before importing. Do not infer that every Open Food Facts license statement automatically resolves third-party rights for every sibling dataset.

Open Pet Food Facts is described by the project support documentation as largely API-compatible with Open Food Facts, with some feature differences. Treat this as a separate pet-food candidate, not a general merchandise source.

### Category coverage is still incomplete

A technology editorial feed and an open food database do not provide a balanced general merchandise catalog. No verified no-account source has yet been found that simultaneously provides broad retail categories, authentic product images, clear reuse rights, seller attribution, current prices and stock, and stable API access.

Do not fill fashion, home, toys or general electronics gaps with invented records or by scraping retailer sites. A true retailer catalog may require an approved affiliate/partner API, seller account, or a permissioned feed.

### Amazon and large retailer feeds

Amazon is not connected. Do not scrape product pages or copy the catalog. A future integration is gated on the owner's eligibility for the relevant official program/API, credentials, permitted display/storage duration, image handling, price/availability refresh, attribution and linking rules. Affiliate discovery is not wholesale inventory access.

## Integration acceptance criteria

A source may replace the current sample adapter only after all of these are confirmed:

1. **Real identity:** recognizable commercial products with manufacturer/model identifiers or barcode where available; no generated product names, brands, prices or stock.
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
- Daily Dose Tech is a technology discovery candidate only; image and redistribution rights must be confirmed before implementation.
- Open Food Facts is a food-only candidate; its attribution/share-alike terms and usage guidance must be followed.
- Do not present either source as a live retailer inventory or claim current retailer stock without an authoritative seller feed.
- Once a qualifying source is selected, implement it behind the existing adapter, add source-specific contract tests and update `docs/DATA_SOURCES.md`, `ROADMAP.md` and `PROJECT_CONTEXT.md`.

## Next technical action

Prioritize a permission-confirmed technology source and a licensed barcode-food source as separate category adapters. Continue discovery for fashion, home and general merchandise. If no source meets the requirements without credentials, report that constraint rather than shipping a fake "live retailer" integration. A real retailer catalog may require the owner to authorize a developer/affiliate account or provide a store/feed endpoint.
