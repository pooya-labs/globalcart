# GlobalCart — Project Context / Single Source of Truth

**Owner:** Pooya Goodarzi  
**Repository:** https://github.com/pooya16pg-spec/globalcart  
**Branch:** main  
**Current release:** 0.1.3  
**Active stage:** M02 — Automated catalog pipeline  
**Last updated:** 2026-11-01

## Product definition
GlobalCart is an independent, high-fidelity marketplace portfolio simulation. The goal is a modern, accessible, bilingual commerce experience with strong UX, reliable data and credible engineering. It is not a live retailer. Payments, fulfillment, inventory reservation and seller payouts remain simulated.

## Non-negotiable invariants
- Keep this repository independent from all other portfolio projects.
- Never collect card numbers/bank credentials or call a real payment gateway.
- Label sample data, illustrative currency and simulated outcomes honestly.
- Do not scrape Amazon; require authorized API/feed access and terms review.
- Preserve original text and source provenance during translation.
- Never expose provider secrets in browser code or Git history.
- Persian is default with true RTL; English must retain correct LTR.
- Every visible control must work or explain why unavailable.
- Keep owner attribution visible.
- Never mark roadmap work complete before verification.

## Current implementation map
- app/page.tsx — storefront, search/filter/sort, local cart/favorites/checkout; initializes from static snapshot and attempts live refresh.
- app/globals.css — responsive visual system and RTL/LTR.
- app/layout.tsx — metadata and global styles.
- data/catalog.json — static snapshot contract; empty local seed replaced by CI sync.
- scripts/sync-catalog.ts — scheduled/manual DummyJSON import, retry, normalization and last-known-good protection.
- lib/catalog/types.ts — canonical catalog contracts.
- lib/catalog/normalize.ts — validation, normalization and deduplication.
- lib/catalog/dummyjson.ts — live demo source adapter.
- lib/pricing.ts — integer-cent/fixed-point demo pricing with BigInt.
- lib/translation/* — provider contract, Gemini API adapter, strict JSON response validation and honest unavailable/error states.
- scripts/translate-catalog.ts — optional server-side Persian localization with source fingerprints, cache, sequential throttling and per-product failure records.
- tests/*.test.ts — catalog, pricing and translation tests.
- .github/workflows/deploy.yml — CI, scheduled/manual catalog refresh, static build and Pages deployment.
- ROADMAP.md — single authoritative roadmap.
- docs/AI_HANDOFF.md — continuation contract.

## Verified deployment state
Actions run 41294484252 passed dependency installation, TypeScript validation, unit tests, production build, Pages artifact upload and deployment:
https://github.com/pooya16pg-spec/globalcart/actions/runs/41294484252

This verification applies to the commit used by that run. Version 0.1.2 catalog pipeline is now verified in Actions run 41295039930.

## M01 status
Storefront foundation, bilingual shell, search/filter/sort, quick view, favorites, recently viewed, browser-local cart and simulated checkout are implemented. GitHub Pages deployment is successful. Visual QA, accessibility and base-path/deep-link verification remain open.

## M02 status
Typed catalog contracts, normalization, source metadata, deduplication, DummyJSON adapter and translation interface exist. No actual translation provider or automatic Persian translation is configured.
Version 0.1.2 adds:
- scripts/sync-catalog.ts: retrying catalog importer.
- data/catalog.json: generated static snapshot.
- GitHub Actions schedule and manual trigger.
- Storefront initializes from snapshot and preserves it if live refresh fails.
CI verification passed in Actions run 41295683515: catalog sync, optional translation step (safely skipped without GEMINI_API_KEY), TypeScript validation, unit tests, production static export and Pages deployment all succeeded. Run: https://github.com/pooya16pg-spec/globalcart/actions/runs/41295683515

## Target backend/database architecture
- Frontend: Next.js App Router, React and TypeScript.
- Backend: Go REST API with OpenAPI contract.
- Primary database: PostgreSQL, selected for relational marketplace entities and transactional integrity.
- Data access: versioned SQL migrations and sqlc-generated typed queries.
- Python: supplier ingestion, enrichment and translation workers.
- Redis: optional cache/queue only when measured need exists; never system of record.
- Search: PostgreSQL full-text/trigram first; dedicated engine only if evidence warrants it.
- REST first; GraphQL only if client requirements justify it.

GitHub Pages is static and cannot host a persistent Go API, Python worker or PostgreSQL database. M04 is gated on owner enabling Go/Python runtimes and approving separate API/database hosting. Do not claim static frontend has backend persistence.

## Immediate next actions
1. Add GEMINI_API_KEY as a GitHub Actions secret and run the workflow to verify live Persian catalog translations.
2. Perform visual/mobile/accessibility QA on the published site.
3. Add an approved translation provider; keep credentials server-side.
4. Continue M03 traceable pricing and product details.
5. Request Go/Python and hosted PostgreSQL access when M04 is reached.
