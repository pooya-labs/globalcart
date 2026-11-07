# GlobalCart — Intelligent Global Marketplace

**Portfolio project 06 · Independent repository**  
**Owner / product engineer:** [Pooya Goodarzi](https://github.com/pooya-labs)  
**Source repository:** https://github.com/pooya-labs/globalcart  
**Public preview:** https://pooya-labs.github.io/globalcart/  
**Bilingual case study:** https://pooya-labs.github.io/globalcart/case-study/


<details>
<summary>Desktop — full-page case-study capture</summary>

![GlobalCart bilingual case study, desktop Chromium](docs/screenshots/case-study-desktop.png)

</details>

<details>
<summary>Mobile — full-page case-study capture</summary>

![GlobalCart bilingual case study, mobile Chromium](docs/screenshots/case-study-mobile.png)

</details>

GlobalCart is a commerce-grade marketplace **simulation** built to demonstrate product engineering: responsive retail UX, catalog ingestion, localization, pricing, discovery, shopping flows, dashboards, automated quality checks and deployment. The editorial storefront shell redesign in [PR #2](https://github.com/pooya-labs/globalcart/pull/2) is merged; the current M09.8 polish pass is tracked in [ROADMAP.md](ROADMAP.md). It is not a live retailer. No real goods are sold and no real payment, inventory reservation, shipment, seller payout or money refund is processed. Returns/refunds exist only as an internal demo state machine.

## Current status — v0.2.3-dev · M09.8 editorial polish in progress; v1.0 remains gated by automated and manual QA

**Latest verification checkpoint — 2026-11-07:** Pages deployment/public smoke [42387896585](https://github.com/pooya-labs/globalcart/actions/runs/42387896585), the six-report Lighthouse regression gate [42367298478](https://github.com/pooya-labs/globalcart/actions/runs/42367298478), post-deploy visual QA [42367298484](https://github.com/pooya-labs/globalcart/actions/runs/42367298484), Shopify Worker smoke [42367298509](https://github.com/pooya-labs/globalcart/actions/runs/42367298509) and TGJU refresh [42387859483](https://github.com/pooya-labs/globalcart/actions/runs/42387859483) passed. The latest Browser E2E attempt [42247489435](https://github.com/pooya-labs/globalcart/actions/runs/42247489435) was cancelled; M09.8 is **not** fully verified and the version remains `0.2.3-dev`.

Implemented:
- Next.js 15 App Router, React 19 and TypeScript.
- Responsive marketplace storefront with Persian RTL and English LTR, category-first navigation and refreshed retail design system.
- Catalog-driven category mega-menu and quick-category navigation with counts derived from the embedded 1,000-item catalog.
- Product detail pages are statically generated for the 1,000-item local catalog and styled as native GlobalCart listings; product cards and titles open the in-store detail page, with no outbound marketplace/product links.
- Multi-image product galleries retain the main image and every alternate image supplied by the catalog record, with thumbnail selection and previous/next controls.
- Original catalog photography is downloaded server-side, optimized to responsive WebP assets and self-hosted under `public/images/catalog/`; the browser never calls the upstream image host.
- Persistent side-by-side comparison for up to four products.
- Three deliberately curated homepage collections (workspace, audio/connectivity and everyday carry) with 12 selected catalog items and reviewed Persian names/short descriptions; original source titles remain available for transparency.
- Guided empty states for the bag, favorites and search, repeated no-real-payment notices, and TGJU freshness beside converted product and cart prices.
- Marketplace service shortcuts and a transparent FAQ covering product sourcing, demo limits and support.
- Offline-first product catalog: a balanced 1,000-item subset of the Amazon Berkeley Objects dataset is committed to the repository with locally hosted, high-resolution WebP image galleries. Search, category filters, product cards and static detail routes work without a runtime Shopify or catalog-provider request.
- Storefront shell, typography, category navigation and product-card geometry render independently of image downloads; product imagery uses reserved aspect-ratio boxes, lazy loading and low fetch priority so slow photos do not block the page.
- The ABO source provides product metadata and images, but no seller sale prices, ratings or stock in the selected fields. GlobalCart generates stable, category-bounded USD demo estimates (`priceKind: estimated`) while keeping `priceAvailable: false`; estimates are never represented as retailer quotes. Ratings and stock remain unlisted. Dataset attribution and CC BY 4.0 details are documented in [data/CATALOG_SOURCE.md](data/CATALOG_SOURCE.md).
- Typed catalog normalization, validation, source identity, timestamps and deduplication foundation.
- Search, category filtering, price range, sorting, quick view, favorites and recently viewed.
- Browser-local cart, quantity changes, removal and simulated checkout.
- TGJU free-market USD quote for demo price estimates; the source publishes Rial and GlobalCart explicitly converts it to Toman. Source time and freshness are labelled; no hard-coded FX fallback is used.
- Decimal-safe pricing core using integer cents, fixed-point FX and explicit rounding.
- Source/demo disclosures and visible owner attribution.
- GitHub Pages deployment workflow, project handoff documentation and a bilingual portfolio case study with architecture/data-flow view and a 27-table PostgreSQL ERD.
- Real desktop/mobile full-page case-study screenshots are committed in `docs/screenshots/`.
- Search suggestions, keyboard command palette, rating-based catalog ranking and saved-search flows.
- Backend foundation: Go REST/OpenAPI, PostgreSQL 17 and sqlc are available for local development and CI. Redis is optional; no hosted business API or database is deployed. The only hosted Worker is a narrow public-catalog proxy.
- Automated axe checks, keyboard focus tests, desktop/mobile Playwright journeys and first-party asset budgets run in CI. The M09.3 screenshot-capture checkpoint passed 40/40 Browser E2E scenarios in [run 41596731945](https://github.com/pooya-labs/globalcart/actions/runs/41596731945).

Boundaries and remaining work:
- Translation is intentionally paused and excluded from CI; original source product text is displayed.
- The storefront does not link customers out to upstream product pages and does not treat the licensed reference catalog as live retailer inventory. Displayed amounts are deterministic GlobalCart estimates, not seller prices; source price, ratings and stock remain unverified or unlisted. A future real-seller feed requires a separate authorized integration and terms review. Automated TGJU page retrieval remains subject to the source site's terms and availability.
- Internal admin and seller services through M06.6 (seller profile and aggregate demo analytics) are implemented and CI-verified. Their dashboard UI, public routes and account-backed order history remain gated by approved hosting and identity infrastructure; cart, checkout, returns and refunds remain simulations/internal services, not live commerce.
- Automated axe/WCAG coverage includes storefront, product detail, recently viewed (populated/empty in both locales), case study and key dialogs. Browser E2E [41652758312](https://github.com/pooya-labs/globalcart/actions/runs/41652758312) passed **99 tests** with **4 intentional skips** across Chromium desktop/mobile, iPad 7th-gen tablet, Firefox and WebKit. Tablet portrait/landscape checks (810×1080 and 1080×810) verify storefront/product/case study have no horizontal overflow; compact-phone checks (320 px and 360 px) cover the same routes. The WebKit dual-locale axe test now passes within its 60 s allowance without a retry.
- Manual screen-reader, physical-device, comprehensive RTL/LTR and live CDN/Core Web Vitals QA remain open. Browser emulation is not a substitute for these external checks; the 1.0.0 release remains gated.

## Demo boundaries

| Area | Current behavior |
|---|---|
| Product catalog | A committed local snapshot with 1,000 real listings selected from the ABO dataset; category and search filters run locally. |
| Product pages | Every product has its own static GlobalCart detail route. Alternate photos are available in a thumbnail/arrow gallery; no outbound product link is presented. |
| Images | Original source imagery is normalized to responsive WebP and stored under `public/images/catalog/`; card, thumbnail and detail images load lazily after the layout is painted. |
| Pricing / availability | Stable category-based GlobalCart demo estimates are displayed separately from source provenance. ABO seller prices, ratings and stock are not provided; no live price or stock claim is made. |
| Translation | Paused; original source text is preserved |
| Currency | TGJU free-market USD page, Rial converted to Toman; indicative, not an executable quote |
| Cart / favorites | Local browser state; product records are not sent to a server. |
| Checkout | Local simulation; never contacts a payment provider or merchant checkout. |
| Inventory / shipping | Not represented as live inventory; GlobalCart does not reserve stock, ship or fulfill goods. |

Amazon's public storefront must not be scraped or bulk-copied. Any future adapter requires verified program/API eligibility and compliance with current content, image, price-refresh, attribution and linking rules. Affiliate product discovery is not an unrestricted catalog feed.

**Catalog source boundary:** the active offline storefront uses the committed 1,000-item Amazon Berkeley Objects snapshot and locally hosted image galleries. Shopify Global Catalog remains a separate Cloudflare Worker integration/smoke-test path; it is not the active storefront catalog source. No Shopify merchant offer is persisted in the local catalog. The cart and checkout remain simulations. Architecture and source constraints are documented in [docs/REAL_CATALOG_MIGRATION.md](docs/REAL_CATALOG_MIGRATION.md).

## Technology

- **Languages:** TypeScript, CSS, Go and SQL (with HTML rendered through React).
- **Framework:** Next.js 15 / React 19.
- **UI:** custom responsive design system, Lucide icons.
- **Hosting:** GitHub Actions + Next.js static export on GitHub Pages. A separate Cloudflare Worker exposes a validated, no-cache Shopify Global Catalog integration path for smoke testing; it is not used as the active storefront feed.
- **State:** React state for the local ABO snapshot and storefront interactions; the full snapshot is loaded on demand near the catalog or when users interact. Browser localStorage holds demo-only cart, favorites, locale, search and recently-viewed references.

This repository is deliberately independent. It has no runtime, deployment or package dependency on any other portfolio project.

## Run locally

Requirements: Node.js 20.9+ and npm.

```bash
git clone https://github.com/pooya-labs/globalcart.git
cd globalcart
npm install
npm run dev
```

Open http://localhost:3000.

```bash
npm run typecheck
npm run build
```

## Documentation and handoff

- [ROADMAP.md](ROADMAP.md) — milestone checklist, acceptance criteria and release ledger.
- [Bilingual case study](https://pooya-labs.github.io/globalcart/case-study/) — product story, feature tour and architecture overview.
- [PROJECT_CONTEXT.md](PROJECT_CONTEXT.md) — current implementation state and next action.
- [docs/AI_HANDOFF.md](docs/AI_HANDOFF.md) — mandatory instructions for a replacement AI/developer.
- [CHANGELOG.md](CHANGELOG.md) — version history.
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — system boundaries and data flow.
- [docs/DATABASE_ERD.md](docs/DATABASE_ERD.md) — migration-derived PostgreSQL ERD covering all 27 tables and key constraints.
- [docs/PERFORMANCE_QA.md](docs/PERFORMANCE_QA.md) — public Lighthouse lab baseline, methodology and manual 1.0 release checklist.
- Latest published Lighthouse [41652966386](https://github.com/pooya-labs/globalcart/actions/runs/41652966386): mobile storefront Performance 91 / LCP 3.3 s / TBT 70 ms; product 94 / 3.1 s / 60 ms; case study 99 / 1.0 s / 100 ms. Desktop Performance is 100 on all routes. Accessibility, Best Practices and SEO are 100 in all six audits; CLS is 0 throughout. All six reports passed the LCP/TBT/CLS/score regression guard. GitHub-hosted lab only, not field Core Web Vitals. Full diagnostics: [docs/PERFORMANCE_QA.md](docs/PERFORMANCE_QA.md).
- [docs/DATA_SOURCES.md](docs/DATA_SOURCES.md) — source provenance and provider policy.
- [docs/REAL_CATALOG_MIGRATION.md](docs/REAL_CATALOG_MIGRATION.md) — requirements and blockers for replacing sample items with real products.
- [docs/DECISIONS.md](docs/DECISIONS.md) — accepted architecture decisions.
- [docs/RELEASE_PROCESS.md](docs/RELEASE_PROCESS.md) — verification and semantic versioning.
- [CONTRIBUTING.md](CONTRIBUTING.md) — engineering rules.

## Versioning

GlobalCart follows semantic versioning. Every release must update package.json, CHANGELOG.md, ROADMAP.md and PROJECT_CONTEXT.md in the same release commit. A roadmap checkbox is only marked complete after implementation and verification. The repository is the source of truth; chat history is not.

**Built and maintained independently by Pooya Goodarzi.**


## Backend foundation (local development)
The M04 backend foundation is merged into `main`. It includes a Go REST API, PostgreSQL 17 Docker Compose setup, versioned SQL migrations, sqlc-generated query layer, catalog snapshot importer and OpenAPI 3.1 contract. It remains local-development infrastructure; the public GitHub Pages demo is static and no cloud database or API has been provisioned.

Historical M04 backend CI [41382309964](https://github.com/pooya-labs/globalcart/actions/runs/41382309964), Security scans [41382310050](https://github.com/pooya-labs/globalcart/actions/runs/41382310050) and Pages deployment [41382310024](https://github.com/pooya-labs/globalcart/actions/runs/41382310024) validated that foundation when it was merged; current public Pages checks are tracked in the roadmap. See [backend/README.md](backend/README.md) for local setup and [backend/openapi.yaml](backend/openapi.yaml) for the read-only API contract.
