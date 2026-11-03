# GlobalCart — Intelligent Global Marketplace

**Portfolio project 06 · Independent repository**  
**Owner / product engineer:** [Pooya Goodarzi](https://github.com/pooya16pg-spec)  
**Source repository:** https://github.com/pooya16pg-spec/globalcart  
**Public preview:** https://pooya16pg-spec.github.io/globalcart/  
**Bilingual case study:** https://pooya16pg-spec.github.io/globalcart/case-study/

GlobalCart is a commerce-grade marketplace **simulation** built to demonstrate product engineering: responsive retail UX, catalog ingestion, localization, pricing, discovery, shopping flows, dashboards, automated quality checks and deployment. It is not a live retailer. No real goods are sold and no real payment, inventory reservation, shipment, seller payout or money refund is processed. Returns/refunds exist only as an internal demo state machine.

## Current status — v0.2.3-dev · M09.2 database ERD and portfolio release in progress

Implemented:
- Next.js 15 App Router, React 19 and TypeScript.
- Responsive marketplace storefront with Persian RTL and English LTR, category-first navigation and refreshed retail design system.
- Catalog-driven category mega-menu and quick-category navigation with live category counts.
- Dedicated statically generated product detail pages for imported catalog entries, including image gallery, pricing panel, product facts and related products.
- Persistent side-by-side comparison for up to four products.
- Amazon product-link intake demo with domain and ASIN validation; requests are stored locally and do not fetch live Amazon data.
- Marketplace service shortcuts and a transparent FAQ covering product sourcing, demo limits and support.
- Typed catalog import foundation plus a scheduled/manual snapshot generator with retry and last-known-good fallback (100 products imported; CI, tests, build and deployment passed in run 41295039930).
- Typed catalog normalization, validation, source identity, timestamps and deduplication foundation.
- Search, category filtering, price range, sorting, quick view, favorites and recently viewed.
- Browser-local cart, quantity changes, removal and simulated checkout.
- TGJU free-market USD quote for demo price estimates; the source publishes Rial and GlobalCart explicitly converts it to Toman. Source time and freshness are labelled; no hard-coded FX fallback is used.
- Decimal-safe pricing core using integer cents, fixed-point FX and explicit rounding.
- Source/demo disclosures and visible owner attribution.
- GitHub Pages deployment workflow and project handoff documentation.
- Target architecture: Go REST/OpenAPI, PostgreSQL + sqlc, Python workers and optional Redis at the backend milestone.

In progress / not yet claimed as complete:
- Translation is intentionally paused and excluded from CI; original source product text is displayed.
- Live Amazon product retrieval and supplier price/stock synchronization require authorized provider access. Automated TGJU page retrieval remains subject to the source site's terms and availability.
- Internal admin and seller services through M06.6 (seller profile and aggregate demo analytics) are implemented and CI-verified. Their dashboard UI, public routes and account-backed order history remain gated by approved hosting and identity infrastructure; cart, checkout, returns and refunds remain simulations/internal services, not live commerce.
- Search suggestions, the keyboard command palette, explainable sample recommendations and saved-search flows are implemented.
- Automated axe checks, keyboard focus tests, desktop/mobile Playwright journeys (including recently viewed image loading), and first-party asset budgets run in CI; manual screen-reader/real-device RTL/LTR QA and live CDN/Core Web Vitals review remain open.


## Demo boundaries

| Area | Current behavior |
|---|---|
| Product catalog | DummyJSON public sample API; local fixtures if unavailable |
| Images | Source-provided remote image URLs |
| Translation | Paused; source text is preserved |
| Currency | TGJU free-market USD page, Rial converted to Toman; indicative, not an executable quote |
| Cart / favorites | Local browser state |
| Checkout | Local simulation; never contacts a payment provider |
| Inventory / shipping | Not real, not reserved, not fulfilled |
| Amazon links | Domain/ASIN validation and local demo request storage; no live product retrieval |
| Amazon catalog | No scraping; future integration only through authorized access and current terms |

Amazon's public storefront must not be scraped or bulk-copied. Any future adapter requires verified program/API eligibility and compliance with current content, image, price-refresh, attribution and linking rules. Affiliate product discovery is not an unrestricted catalog feed.

## Technology

- **Languages:** TypeScript, CSS, Go and SQL (with HTML rendered through React).
- **Framework:** Next.js 15 / React 19.
- **UI:** custom responsive design system, Lucide icons.
- **Hosting:** GitHub Actions + Next.js static export on GitHub Pages.
- **State:** React state and browser localStorage for demo-only preferences.

This repository is deliberately independent. It has no runtime, deployment or package dependency on any other portfolio project.

## Run locally

Requirements: Node.js 20.9+ and npm.

```bash
git clone https://github.com/pooya16pg-spec/globalcart.git
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
- [Bilingual case study](https://pooya16pg-spec.github.io/globalcart/case-study/) — product story, feature tour and architecture overview.
- [PROJECT_CONTEXT.md](PROJECT_CONTEXT.md) — current implementation state and next action.
- [docs/AI_HANDOFF.md](docs/AI_HANDOFF.md) — mandatory instructions for a replacement AI/developer.
- [CHANGELOG.md](CHANGELOG.md) — version history.
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — system boundaries and data flow.
- [docs/DATABASE_ERD.md](docs/DATABASE_ERD.md) — migration-derived PostgreSQL ERD covering all 27 tables and key constraints.
- [docs/DATA_SOURCES.md](docs/DATA_SOURCES.md) — source provenance and provider policy.
- [docs/DECISIONS.md](docs/DECISIONS.md) — accepted architecture decisions.
- [docs/RELEASE_PROCESS.md](docs/RELEASE_PROCESS.md) — verification and semantic versioning.
- [CONTRIBUTING.md](CONTRIBUTING.md) — engineering rules.

## Versioning

GlobalCart follows semantic versioning. Every release must update package.json, CHANGELOG.md, ROADMAP.md and PROJECT_CONTEXT.md in the same release commit. A roadmap checkbox is only marked complete after implementation and verification. The repository is the source of truth; chat history is not.

**Built and maintained independently by Pooya Goodarzi.**


## Backend foundation (local development)
The M04 backend foundation is merged into `main`. It includes a Go REST API, PostgreSQL 17 Docker Compose setup, versioned SQL migrations, sqlc-generated query layer, catalog snapshot importer and OpenAPI 3.1 contract. It remains local-development infrastructure; the public GitHub Pages demo is static and no cloud database or API has been provisioned.

Backend CI [41382309964](https://github.com/pooya16pg-spec/globalcart/actions/runs/41382309964), Security scans [41382310050](https://github.com/pooya16pg-spec/globalcart/actions/runs/41382310050) and GitHub Pages deployment [41382310024](https://github.com/pooya16pg-spec/globalcart/actions/runs/41382310024) passed on the current `main` implementation. See [backend/README.md](backend/README.md) for local setup and [backend/openapi.yaml](backend/openapi.yaml) for the read-only API contract.
