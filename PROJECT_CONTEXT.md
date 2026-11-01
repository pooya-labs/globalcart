# GlobalCart — Project Context / Single Source of Truth

**Owner:** Pooya Goodarzi  
**Repository:** https://github.com/pooya16pg-spec/globalcart  
**Branch:** main  
**Package version:** 0.2.1  
**Active stage:** M01.5 marketplace redesign / M03 product experience  
**Last updated:** 2026-11-01

## Product definition
GlobalCart is an independent marketplace portfolio simulation. Its information architecture is inspired by large retail marketplaces and the Amazon-link sourcing flow of kiyandaria.com, while its brand, UI and implementation remain original. It is not a live retailer. No real payments, orders, inventory reservations, fulfillment, seller payouts or refunds occur.

## Non-negotiable invariants
- Keep this repository independent from all other portfolio projects.
- Never collect card numbers/bank credentials or call a real payment gateway.
- Label sample data, illustrative currency and simulated outcomes honestly.
- Do not scrape or bulk-copy Amazon. Live catalog integration requires authorized API/feed access and terms review.
- Amazon product-link intake currently validates an Amazon hostname and 10-character ASIN, then stores a local demo request only. It does not fetch product data.
- Translation is paused; preserve original source text and do not add translation back to CI unless explicitly requested.
- Never expose provider secrets in browser code or Git history.
- Persian is default with true RTL; English must retain correct LTR.
- Every visible control must work or explain why unavailable.
- Keep owner attribution visible.
- Never mark roadmap work complete before verification.

## Current implementation map
- `app/page.tsx` — bilingual storefront, catalog search/filter/sort, category mega-menu, Amazon URL request intake, local cart/favorites/compare, quick view and simulated checkout, service shortcuts and FAQ.
- `app/product/[id]/page.tsx` — statically generated product routes from the current catalog snapshot.
- `app/product/[id]/ProductDetailClient.tsx` — product gallery, pricing, product facts, local cart/favorite actions and related products.
- `app/globals.css` — responsive marketplace design system, product detail, mega-menu, comparison and Amazon intake styling.
- `app/layout.tsx` — metadata and global styles.
- `data/catalog.json` — generated 100-product sample catalog snapshot.
- `scripts/sync-catalog.ts` — scheduled/manual DummyJSON import, retry, normalization and last-known-good protection.
- `lib/catalog/*` — canonical catalog contracts, normalization and source adapter.
- `lib/pricing.ts` — integer-cent/fixed-point demo pricing with BigInt.
- `lib/translation/*` and `google-apps-script/*` — parked legacy translation implementation; not used by storefront or CI.
- `tests/*.test.ts` — catalog, pricing and legacy translation tests.
- `.github/workflows/deploy.yml` — CI, scheduled/manual catalog refresh, static build and Pages deployment.
- `ROADMAP.md` — authoritative milestone checklist.
- `docs/AI_HANDOFF.md` — continuation contract.

## Verified behavior
- Catalog snapshot pipeline imports 100 sample products and preserves a last-known-good snapshot.
- The storefront supports Persian RTL and English LTR, category navigation, search, filters, sorting, favorites, recently viewed, local cart and simulated checkout.
- Product detail routes are statically generated for products in the snapshot and include gallery, pricing, facts and related products.
- Comparison supports up to four products and is persisted in browser local storage.
- Amazon URL intake validates supported Amazon domains and ASINs, then saves requests in browser local storage. Live Amazon data is not fetched.
- Translation is out of scope and removed from the active CI pipeline.

## Deployment
Public preview: https://pooya16pg-spec.github.io/globalcart/  
Repository Actions: https://github.com/pooya16pg-spec/globalcart/actions

The last fully verified feature run before the Amazon intake and FAQ additions is recorded in the repository's Actions history. Do not claim the newest commit is deployed until its matching Actions run completes successfully.

## Immediate next actions
1. Verify the latest Amazon-link intake, service/FAQ and comparison changes through TypeScript, unit tests, static export and Pages deployment.
2. Add a richer homepage merchandising system: curated collections, trending products and deal rails using only authorized catalog data.
3. Add transparent product price breakdown and FX provenance/staleness handling.
4. Complete responsive, RTL/LTR, keyboard, screen-reader, reduced-motion and browser QA.
5. Build seller/admin workspaces only after the catalog UX and role boundaries are defined.
6. Defer live Amazon product retrieval until approved API/feed credentials and usage rights are available.

## Backend/database architecture
- Frontend: Next.js App Router, React and TypeScript, statically exported to GitHub Pages.
- Backend target: Go REST API with OpenAPI.
- Primary database: PostgreSQL with versioned SQL migrations and sqlc.
- Python: ingestion and enrichment workers; translation remains paused.
- Redis: optional cache/queue only when measured need exists.
- Search: PostgreSQL full-text/trigram first.
- M04 requires owner approval for API hosting, database and worker runtime.

