# Data Sources & Provenance

## Demo catalog
DummyJSON Products API: https://dummyjson.com/products
- Purpose: sample catalog import without manual per-product entry.
- This is external demo content, not Amazon inventory or commercial availability.
- Image URLs are supplied by the source and can change or become unavailable.
- Preserve source product IDs and import timestamps.

## Amazon
Not active in v0.1.0. Do not scrape Amazon product pages or copy its catalog. Any future connector requires eligible authorized API/program access and review of current content, image, price refresh, attribution, and linking rules. Affiliate discovery is not wholesale inventory.

## Translation
No translation provider is assumed available by default. A provider adapter must preserve original fields, store translation separately, identify provider/locale/timestamp/status, cache results, handle limits, and never expose secrets in browser code.

## Foreign exchange
Initial UI uses a configurable illustrative USD-to-Toman rate. It is not a live quote or purchase promise. Future sources must include timestamp and stale-rate policy.

## UI labels
- Demo source: external sample catalog.
- Local fallback: repository fixture used if import fails.
- Illustrative calculation: arithmetic based on configured sample inputs.
- Live external data: only when provider response and source/time are shown.
