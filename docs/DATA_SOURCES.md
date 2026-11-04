# Data Sources, Translation and Provenance

## Current catalog source
**DummyJSON Products API:** https://dummyjson.com/products?limit=100

- Used as a public sample catalog to demonstrate automatic ingestion.
- Not Amazon inventory and not a commercial offer.
- The adapter normalizes records and keeps source identity/import time.
- Images are remote source-provided URLs and may change or fail.
- If the network source is unavailable, the UI uses a small local fixture.
- Do not present these entries as verified real-world retail listings, live prices or actual stock.

## Real-product source candidates (not yet connected)

### Daily Dose Tech
- Official docs: https://dailydose.tech/docs
- Public read-only JSON endpoints; no API key; paginated products, categories, search, retailer links, price history and deal data.
- The publisher describes a curated technology editorial catalog and discloses affiliate commissions.
- Scope: technology products only. This is not a general merchandise catalog or a guarantee of live retailer price/stock.
- Cache for at least one hour as requested by the API docs.
- Image reuse, redistribution, attribution and display rights are not sufficiently clear from the API docs; do not import until the publisher confirms them.
- Preserve retailer attribution and outbound links if integrated.

### Shopify Global Catalog (runtime discovery candidate; not connected)
- Official overview: https://shopify.dev/docs/agents/catalog
- Searches eligible products across Shopify merchants; distinct from the Storefront API, which is scoped to one known store.
- Uses the UCP/MCP Global Catalog endpoint and a hosted agent profile; Shopify's docs state that an API key is not required for keyless catalog access.
- Do not cache search results or product images; images must be rendered in real time and only with the related merchant listing.
- Keyless requests are rate-limited and do not support rate-limit increases.
- Results are third-party merchant offers. Preserve seller identity/domain and merchant purchase links; never claim GlobalCart owns or fulfills them.
- This is not compatible with the existing persisted static catalog snapshot. A possible future design is a separate runtime discovery layer, pending a proof of concept for profile hosting, request transport/CORS, response mapping, seller/price/availability fields and rate limits.
- No code or catalog records are connected yet.

### Open Food Facts
- Official API docs: https://openfoodfacts.github.io/documentation/docs/Product-Opener/api/
- Real packaged-food products indexed by barcode; read API calls do not require an API key.
- Use API v3 for product reads. Structured search is currently documented in v2, which is deprecated for new integrations; do not use the API for search-as-you-type.
- Use a descriptive User-Agent and follow the API usage guidance and rate limits (15 product reads/minute/IP; 10 search requests/minute/IP).
- Database: ODbL; individual contents: DbCL; product images: CC BY-SA. Attribution/share-alike requirements and possible third-party rights in image content must be honored.
- Community-contributed product facts may be incomplete or inaccurate.
- Scope: food/grocery records, not general retail inventory. It does not provide authoritative live retailer prices, stock, cart or fulfillment.

### Open Facts sibling candidates (not connected)

The Product Opener project documents related databases with separate hosts:
- Open Beauty Facts: https://world.openbeautyfacts.org — cosmetics/personal care.
- Open Pet Food Facts: https://world.openpetfoodfacts.org — pet food.
- Open Products Facts: https://world.openproductsfacts.org — other product types.

The official guide describes these as experimental/specific APIs; some shared capabilities are not available consistently. They are not retailer feeds and do not establish live seller price, stock or fulfillment. Verify each dataset's current terms, image rights, quality, rate limits and actual endpoint behavior independently before integration. Do not assume broad fashion/home/toys coverage or transfer license conclusions across datasets without checking the relevant terms.

Neither candidate is currently connected to the storefront. Do not replace DummyJSON until source rights, field mapping, freshness semantics, attribution and operational limits are verified. The migration decision and acceptance criteria live in [REAL_CATALOG_MIGRATION.md](REAL_CATALOG_MIGRATION.md).

## Amazon policy
Amazon is not connected. Do not scrape product pages or copy the catalog. A future integration is allowed only after verifying the account's eligibility and authorized API/program access and reviewing current terms for content reuse, image URLs, price refresh, attribution and linking. An affiliate API is not a wholesale inventory feed.

## Translation status
The code currently defines a provider interface and safe orchestration metadata. It does **not** yet automatically translate the full catalog. Until a provider adapter is configured, translated text is null and status is unavailable; source text must remain visible. A future scheduled job must translate title, description, category and attributes, preserve original fields, cache results, handle quotas/retries and record provider/locale/timestamp. API keys belong in GitHub Actions secrets and must never enter browser bundles or committed files.

## Foreign exchange
The static storefront consumes `data/fx-rate.json`, polled by `.github/workflows/fx-rate.yml` on a five-minute GitHub Actions schedule from the exact public TGJU dollar page requested for this project: `https://www.tgju.org/قیمت-دلار`. GitHub schedule delivery is best-effort rather than an exact real-time timer. The job makes one bounded HTML request and reads only the `price_dollar_rl` instrument row. When a validated quote changes, the workflow commits the last-known-good snapshot and sends the same-repository `globalcart-fx-updated` `repository_dispatch`; the Pages workflow then rebuilds the static site so the new quote reaches visitors. An unchanged quote does not trigger an unnecessary rebuild. The job does not use a third-party proxy or API aggregator, and the browser never calls TGJU directly.

**Currency and unit contract:** TGJU's `price_dollar_rl` field is the free-market US dollar quoted in Rial. The raw Rial value is preserved in the snapshot and divided by ten to produce the displayed Toman-per-USD rate. This is not a Tether/USDT quote. The page's update cell may contain either a Tehran-local clock or a Persian date-only marker. Clock values are anchored to their latest non-future occurrence; when the displayed time is ahead of the current Tehran clock, it is treated as yesterday rather than accepted as a future quote. Date-only values are preserved verbatim but always classified as stale because they have no year or exact time. UI labels show only the source marker, not an invented calendar date or timestamp. Unknown formats fail closed.

Freshness uses the source update marker: clock values are fresh up to five minutes, aging over five to 15 minutes, stale after 15 minutes; date-only markers are always stale. Missing or invalid prices/timestamps are unavailable. On fetch, status, markup or validation failure, the last-known-good file is not overwritten; the UI labels it stale as the source timestamp ages. With no valid snapshot, prices remain in USD rather than falling back to a hard-coded rate.

TGJU's public currency board uses the stable row key `price_dollar_rl` and the market page's raw prices are Rial-denominated. The page may change markup or restrict automated requests; the job does not bypass CAPTCHA, HTTP errors, rate limits or access controls. Automated use remains subject to TGJU's current terms and availability.

## Provenance labels
- **External demo source:** a sample record returned by DummyJSON.
- **Local fallback:** fixture data included for offline resilience.
- **Translated:** only when a provider returns a non-empty translation.
- **Translation unavailable/error:** original text is retained.
- **Exchange-derived estimate:** converted using the latest validated TGJU free-market USD quote, preserving the raw Rial value and deriving Toman per USD.
- **Stale exchange quote:** the last-known-good TGJU rate remains visible only with an aging/stale warning; no hard-coded fallback is used.

## Supplier adapter contract (M08.3)
- Every supplier-specific adapter implements the shared `CatalogSourceAdapter` contract: provide a stable source name, fetch raw records, and map each record into the canonical product/variant input shape.
- Shared ingestion then performs normalization, source provenance assignment and de-duplication. A feed that is empty, malformed or wholly unmappable fails before the last-known-good snapshot can be replaced.
- Variant identity is the pair `(source, sourceId)`; SKU is retained as a separate seller/source SKU field and is only a fallback identity when a source ID is absent.
- A source that does not provide variants remains a valid product-only source. The importer must not invent variants, zero stock, or retire records merely because a later snapshot omits them.
- `tests/fixtures/synthetic-supplier-records.ts` is synthetic test data used only to verify mapping. It is not a supplier connection or live commercial inventory.
- A real supplier adapter requires explicit authorization, reviewed API/feed terms and a confirmed field mapping. Credentials must remain server-side and are not part of this implementation.

## Snapshot integrity and quality diagnostics (M08.4)
- The current scheduled DummyJSON job is a full-snapshot importer. Before replacing the last-known-good snapshot, it rejects a refresh whose normalized product count is more than 50% below the previous snapshot.
- A rejected refresh follows the existing safe error path and preserves the previous file. It does not publish a partial catalog.
- The 50% threshold is a conservative guardrail for the current full-feed job, not a universal supplier rule. Any future incremental/delta source must document completeness semantics and use source-appropriate continuity checks instead.
- The quality report records product count, variant count, products with/without variants, zero-price/zero-stock products, missing descriptions and uncategorized products. These are diagnostics, not synthetic corrections.
- Product-only catalogs are valid. Missing variants in a source response do not imply deleting variants or archiving offers.

**M08.4 verification:** Catalog CI PR run `41355595494` passed TypeScript validation, all tests and production build. The full-snapshot continuity guard is covered by tests for the 50% boundary, rejection of a 51% drop, and invalid thresholds.

## Variant omission and retention semantics (M08.5)
- A product snapshot with no variants is valid for a product-only source. It is not an instruction to delete previously imported variants or vendor offers.
- The Go importer upserts only variant rows actually supplied, then synchronizes the platform offer from all persisted variants for that product. Consequently, an omitted variant remains present with its previous state until a future explicit and authorized retirement protocol is designed.
- This is intentionally conservative: a partial page, temporary source omission or feed schema regression must not destroy stock or seller offers. Future source adapters must declare whether they provide a complete variant set and must provide a documented retirement signal before destructive reconciliation can be considered.
- M08.5 adds PostgreSQL regression coverage for retained source variants, retained platform offer values/status, independent merchant offer isolation and unchanged audit-event counts.

## Variant source identity ownership (M08.6)
- A variant's `(source, source_id)` pair is an ownership key. If an existing pair is presented under a different product, the importer rejects the transaction rather than moving the row.
- The M08.6 PostgreSQL test also verifies batch atomicity: an otherwise valid product earlier in the same batch must not persist when a later variant identity conflict aborts the import.

## Feed completeness declaration (M08.7)
- A source adapter must declare whether its payload is a complete catalog snapshot or an incremental/delta batch.
- The scheduled static snapshot writer may consume only complete snapshots. Incremental records require a source-specific merge/reconciliation workflow and must never replace the complete catalog directly.
- An incremental declaration alone does not authorize deleting products or variants, zeroing stock or archiving offers.

**M08.7 implementation detail:** `CatalogSourceAdapter.snapshotMode` is mandatory. The current scheduled snapshot writer passes `requiredSnapshotMode: "full"`, so an incremental/delta adapter is rejected before its fetch method runs. The generic loader may normalize incremental records for a separately designed merger, but this implementation does not merge or persist delta feeds.

**M08.7 verification:** Catalog CI run `41356770768` and Backend CI run `41356770801` both passed on the same PR head. A full-snapshot writer cannot fetch or consume an incremental adapter; a dedicated delta merger is still required before incremental records can affect the canonical snapshot.

## Feed mapping rejection guard (M08.8)
- Shared ingestion rejects a source response when more than 20% of raw records fail the adapter mapProduct step.
- A source with a documented legitimate filtering rule may set CatalogImportOptions.maxUnmappedFraction explicitly; the value must be finite and in [0, 1).
- The check is based on raw source records, before normalization. Duplicate canonical product IDs are not counted as mapper rejections and continue through normal de-duplication.
- Empty collections, wholly unmappable collections and collections with no valid normalized products remain errors.
- This is a defensive default, not a substitute for source-specific completeness semantics or the full-snapshot continuity guard.

## Request deadlines and cancellation (M08.8)
- Every DummyJSON request receives a 20-second timeout. When the caller supplies an AbortSignal, the request uses a combined signal so the caller can cancel it without losing the timeout.
- Explicit caller cancellation is propagated immediately and is not retried. Other transient failures, including a request deadline expiry, retain the adapter's bounded retry policy.
- This protects scheduled refreshes from waiting indefinitely while preserving the last-known-good snapshot error path.

## Retry classification (M08.8)
- Each DummyJSON request has a 20-second deadline, combined with any caller-provided AbortSignal.
- Explicit caller cancellation is propagated immediately and is not retried.
- Network failures and request timeouts, plus HTTP 408, 429 and 5xx statuses, are retried up to three total attempts with bounded backoff.
- Permanent 4xx statuses other than 408/429, malformed JSON/schema and empty collections fail immediately; they are not retried as transient network faults.

## Aggregate feed import diagnostics (M08.8.1)
- The shared loader can return an aggregate report containing source name, snapshot mode, import timestamp, fetched/mapped/unmapped record counts, mapping rejection fraction and threshold, normalized product count and the count not retained after normalization.
- The post-normalization discard count combines invalid canonical product rows and duplicate source identities; it intentionally does not mislabel duplicates as mapper failures.
- The report contains counts and metadata only, never raw supplier payloads.
- The scheduled snapshot sync logs this report after a successful atomic snapshot write, next to the existing catalog-quality report.
- `loadCatalogFromSource` remains available as a backwards-compatible products-only API.

## Atomic snapshot publication (M08.8.2)
- The scheduled writer serializes the full snapshot before touching the destination, then writes to a unique temporary file in the destination directory.
- Temporary files use an exclusive create flag and a process/random identifier; concurrent invocations do not share a temporary path.
- The completed temporary file is renamed over the destination only after the write succeeds. Readers therefore see the previous complete snapshot or one complete new snapshot, not partially-written JSON.
- Failed writes/renames trigger best-effort cleanup of that invocation's temporary file. Serialization failures happen before the destination is changed.
- This protects file integrity, not source ordering: if independent processes publish different valid snapshots, the final successful rename wins. The existing Pages workflow still serializes its own scheduled/manual runs.

## Last-known-good snapshot validation (M08.8.3)
- Before a previous snapshot is used for continuity checks or localization preservation, its schema version, source/timestamp metadata, product count and normalized product records are validated at runtime.
- A missing path with ENOENT is the only condition treated as a first import. Malformed JSON, count mismatch, invalid product structure, permission errors and other I/O failures abort the sync rather than silently dropping the previous snapshot.
- This is intentionally fail-closed: an existing corrupt snapshot must be repaired explicitly instead of being mistaken for an empty catalog and replaced or hidden by fixture fallback.

## Snapshot identity and database-bound validation (M08.8.4)
- Product IDs and `(source, source_id)` keys must be unique across the whole snapshot; every product source must equal the snapshot source.
- Variant `(source, source_id)` identity must be unique across all products, matching the PostgreSQL unique index and importer ownership rule.
- Prices must be non-negative, fit the current PostgreSQL `NUMERIC(12,2)` range and have no more than two fractional decimal places. Product discounts are limited to 0–100, ratings to 0–5, and stock values to non-negative PostgreSQL 32-bit integers. Product IDs must remain JavaScript-safe integers.
- Invalid snapshots are rejected without modifying, deduplicating or fabricating records.
- Non-empty trimmed variant SKUs are also unique across the snapshot, matching the PostgreSQL unique SKU index. The full contract is asserted before publication and when reading the previous snapshot.
