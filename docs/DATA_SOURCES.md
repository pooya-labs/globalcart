# Data Sources, Translation and Provenance

## Current catalog source
**DummyJSON Products API:** https://dummyjson.com/products?limit=100

- Used as a public sample catalog to demonstrate automatic ingestion.
- Not Amazon inventory and not a commercial offer.
- The adapter normalizes records and keeps source identity/import time.
- Images are remote source-provided URLs and may change or fail.
- If the network source is unavailable, the UI uses a small local fixture.

## Amazon policy
Amazon is not connected. Do not scrape product pages or copy the catalog. A future integration is allowed only after verifying the account's eligibility and authorized API/program access and reviewing current terms for content reuse, image URLs, price refresh, attribution and linking. An affiliate API is not a wholesale inventory feed.

## Translation status
The code currently defines a provider interface and safe orchestration metadata. It does **not** yet automatically translate the full catalog. Until a provider adapter is configured, translated text is null and status is unavailable; source text must remain visible. A future scheduled job must translate title, description, category and attributes, preserve original fields, cache results, handle quotas/retries and record provider/locale/timestamp. API keys belong in GitHub Actions secrets and must never enter browser bundles or committed files.

## Foreign exchange
The storefront currently uses a user-configurable illustrative USD-to-Toman rate. It is not a live market quote. Future FX providers must return source, currency pair, timestamp and rate; stale values must be labelled.

## Provenance labels
- **External demo source:** a sample record returned by DummyJSON.
- **Local fallback:** fixture data included for offline resilience.
- **Translated:** only when a provider returns a non-empty translation.
- **Translation unavailable/error:** original text is retained.
- **Illustrative price:** converted using a manually configured sample rate.
- **Live external value:** only when its provider and timestamp are displayed.


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
- Product-only catalogs are valid. Missing variants in a source response do not imply deleting variants or vendor offers already stored in PostgreSQL.


**M08.4 verification:** Catalog CI PR run `41355595494` passed TypeScript validation, all tests and production build. The full-snapshot continuity guard is covered by tests for the 50% boundary, rejection of a 51% drop, and invalid thresholds.


## Variant omission and retention semantics (M08.5)
- A product snapshot with no variants is valid for a product-only source. It is not an instruction to delete previously imported variants or vendor offers.
- The Go importer upserts only variant rows actually supplied, then synchronizes the platform offer from all persisted variants for that product. Consequently, an omitted variant remains present with its previous state until a future explicit and authorized retirement protocol is designed.
- This is intentionally conservative: a partial page, temporary source omission or feed schema regression must not destroy stock or seller offers. Future source adapters must declare whether they provide a complete variant set and must provide a documented retirement signal before destructive reconciliation can be considered.
- M08.5 adds PostgreSQL regression coverage for retained source variants, retained platform offer values/status, independent merchant offer isolation and unchanged audit-event counts.


## Variant source identity ownership (M08.6)
- A variant's `(source, source_id)` pair is an ownership key. If an existing pair is presented under a different product, the importer rejects the transaction rather than moving the row.
- The M08.6 PostgreSQL test also verifies batch atomicity: an otherwise valid product earlier in the same batch must not persist when a later variant identity conflict aborts the import.
- Original variant ownership and mutable catalog fields remain unchanged after rejection.


## Feed completeness declaration (M08.7)
- A source adapter must declare whether its payload is a complete catalog snapshot or an incremental/delta batch.
- The scheduled static snapshot writer may consume only complete snapshots. Incremental records require a source-specific merge/reconciliation workflow and must never replace the complete catalog directly.
- An incremental declaration alone does not authorize deleting products or variants, zeroing stock or archiving offers.


**M08.7 implementation detail:** `CatalogSourceAdapter.snapshotMode` is mandatory. The current scheduled snapshot writer passes `requiredSnapshotMode: "full"`, so an incremental/delta adapter is rejected before its fetch method runs. The generic loader may normalize incremental records for a separately designed merger, but this implementation does not merge or persist delta feeds.
