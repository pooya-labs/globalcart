# GlobalCart — Project Context / Single Source of Truth

**Owner:** Pooya Goodarzi  
**Repository:** https://github.com/pooya16pg-spec/globalcart  
**Branch:** main  
**Package version:** 0.2.2 (next development increment: 0.2.3-dev)  
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
- `app/product/[id]/ProductDetailClient.tsx` — product gallery, illustrative price breakdown, product facts, browser-local visit tracking, local cart/favorite actions and related products.
- `app/globals.css` — responsive marketplace design system, product detail, price breakdown, recently viewed rail, mega-menu, comparison and Amazon intake styling.
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
- Keyboard support includes Escape dismissal for overlays, dialog semantics, visible focus styles and live announcements for key feedback states.
- Comparison supports up to four products and is persisted in browser local storage.
- Homepage merchandising rails are derived from the current sample catalog (discounts, ratings and discovery).
- Recently viewed product IDs are stored in browser local storage and shown in a homepage rail; there is no server-side account history.
- Product detail pages expose source USD price, sample discount, illustrative user-configured FX rate and converted demo price. No live FX provider is connected.
- Amazon URL intake validates supported Amazon domains and ASINs, then saves requests in browser local storage. Live Amazon data is not fetched.
- Translation is out of scope and removed from the active CI pipeline.

## Deployment
Public preview: https://pooya16pg-spec.github.io/globalcart/  
Repository Actions: https://github.com/pooya16pg-spec/globalcart/actions

Actions run 41299322140 (Run #126) passed catalog refresh, snapshot persistence, TypeScript validation, unit tests, production static export, Pages artifact upload and GitHub Pages deployment. This run includes dynamic merchandising, browser-local recently viewed history and the illustrative product price breakdown.
https://github.com/pooya16pg-spec/globalcart/actions/runs/41299322140

## Immediate next actions
1. Complete responsive, RTL/LTR, keyboard, screen-reader, reduced-motion and published deep-link QA.
2. Add an authorized FX provider with source timestamp and stale-rate policy; keep the current rate clearly illustrative until then.
3. Improve source attributes, rating context and catalog freshness diagnostics.
4. Build seller/admin workspaces only after the catalog UX and role boundaries are defined.
5. Defer live Amazon product retrieval until approved API/feed credentials and usage rights are available.

## Backend/database architecture
- Frontend: Next.js App Router, React and TypeScript, statically exported to GitHub Pages.
- Backend target: Go REST API with OpenAPI.
- Primary database: PostgreSQL with versioned SQL migrations and sqlc.
- Python: ingestion and enrichment workers; translation remains paused.
- Redis: optional cache/queue only when measured need exists.
- Search: PostgreSQL full-text/trigram first.
- M04 requires owner approval for API hosting, database and worker runtime.

