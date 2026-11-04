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

GlobalCart is a commerce-grade marketplace **simulation** built to demonstrate product engineering: responsive retail UX, catalog ingestion, localization, pricing, discovery, shopping flows, dashboards, automated quality checks and deployment. The pre-release storefront shell redesign is tracked in [PR #2](https://github.com/pooya-labs/globalcart/pull/2). It is not a live retailer. No real goods are sold and no real payment, inventory reservation, shipment, seller payout or money refund is processed. Returns/refunds exist only as an internal demo state machine.

## Current status — v0.2.3-dev · automated release guards complete; v1.0 gated by manual QA

Implemented:
- Next.js 15 App Router, React 19 and TypeScript.
- Responsive marketplace storefront with Persian RTL and English LTR, category-first navigation and refreshed retail design system.
- Catalog-driven category mega-menu and quick-category navigation with counts derived from the embedded 1,000-item catalog.
- Product detail pages are statically generated for the local 1,000-item Amazon Berkeley Objects subset; each selected main image is self-hosted under `public/images/catalog/`.
- Persistent side-by-side comparison for up to four products.
- Amazon product-link intake demo with domain and ASIN validation; requests are stored locally and do not fetch live Amazon data.
- Marketplace service shortcuts and a transparent FAQ covering product sourcing, demo limits and support.
- Offline-first product catalog: a balanced 1,000-item subset of the Amazon Berkeley Objects dataset is stored in the repository with locally hosted 256px-class product images. Search, category filters, product cards and static detail routes work without a runtime Shopify or catalog-provider request.
- The ABO source provides product metadata and images but not sale prices, ratings or stock for the selected listing fields; GlobalCart marks those values unavailable instead of displaying fabricated zeroes. Dataset attribution and CC BY 4.0 details are documented in [data/CATALOG_SOURCE.md](data/CATALOG_SOURCE.md).
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
- Live Amazon product retrieval still requires an authorized provider integration. Shopify Global Catalog is connected for cross-merchant discovery; anonymous access is rate-limited and results are runtime-only. Additional supplier feeds need separate source and terms review. Automated TGJU page retrieval remains subject to the source site's terms and availability.
- Internal admin and seller services through M06.6 (seller profile and aggregate demo analytics) are implemented and CI-verified. Their dashboard UI, public routes and account-backed order history remain gated by approved hosting and identity infrastructure; cart, checkout, returns and refunds remain simulations/internal services, not live commerce.
- Automated axe/WCAG coverage includes storefront, product detail, recently viewed (populated/empty in both locales), case study and key dialogs. Browser E2E [41652758312](https://github.com/pooya-labs/globalcart/actions/runs/41652758312) passed **99 tests** with **4 intentional skips** across Chromium desktop/mobile, iPad 7th-gen tablet, Firefox and WebKit. Tablet portrait/landscape checks (810×1080 and 1080×810) verify storefront/product/case study have no horizontal overflow; compact-phone checks (320 px and 360 px) cover the same routes. The WebKit dual-locale axe test now passes within its 60 s allowance without a retry.
- Manual screen-reader, physical-device, comprehensive RTL/LTR and live CDN/Core Web Vitals QA remain open. Browser emulation is not a substitute for these external checks; the 1.0.0 release remains gated.

## Demo boundaries

| Area | Current behavior |
|---|---|
| Product catalog | A committed local snapshot with exactly 1,000 real product listings selected from the ABO dataset; search and category filtering run entirely in the browser. |
| Images | Product images are stored and served from the repository under `public/images/catalog/`; the hero/editorial images are also self-hosted. |
| Pricing / availability | The ABO listings used here do not provide sale prices, ratings or stock values. These are clearly labelled as not provided; no fake prices or stock claims are generated. |
| Translation | Paused; source text is preserved |
| Currency | TGJU free-market USD page, Rial converted to Toman; indicative, not an executable quote |
| Cart / favorites | Local browser state; live product records are not persisted. |
| Checkout | Local simulation; never contacts a payment provider or merchant checkout. |
| Inventory / shipping | Seller availability may be displayed; GlobalCart does not reserve stock, ship or fulfill goods. |
| Amazon links | Domain/ASIN validation and local demo request storage; no live product retrieval |
| Amazon catalog | No scraping; future integration only through authorized access and current terms |

Amazon's public storefront must not be scraped or bulk-copied. Any future adapter requires verified program/API eligibility and compliance with current content, image, price-refresh, attribution and linking rules. Affiliate product discovery is not an unrestricted catalog feed.

**Real-product catalog:** the main storefront catalog now discovers real eligible merchant listings through Shopify Global Catalog at runtime. Merchant names and seller listing links remain visible; Shopify is not used as the storefront brand. Search results and images are not cached or stored. The cart and checkout remain simulations and do not create merchant orders. Architecture and constraints are documented in [docs/REAL_CATALOG_MIGRATION.md](docs/REAL_CATALOG_MIGRATION.md).

## Technology

- **Languages:** TypeScript, CSS, Go and SQL (with HTML rendered through React).
- **Framework:** Next.js 15 / React 19.
- **UI:** custom responsive design system, Lucide icons.
- **Hosting:** GitHub Actions + Next.js static export on GitHub Pages; Cloudflare Worker provides a validated, no-cache proxy to the live catalog API.
- **State:** React state for live catalog results; browser localStorage for demo-only preferences and references (live product records are excluded).

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
