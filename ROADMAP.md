# GlobalCart — Master Roadmap

> The single source of truth. A stage is complete only after implementation and verification.

**Current release:** 0.2.1-dev  
**Active stage:** M01.5 — Marketplace redesign and UX foundation  
**Repository:** pooya16pg-spec/globalcart  
**Owner:** Pooya Goodarzi

## Engineering north star
Build a polished, bilingual global marketplace with excellent usability, accessible interaction, responsive layouts, reliable data and credible production-style architecture. Visual effects must never compromise readability, performance or accessibility.

## Target technology and database architecture

| Layer | Target | Purpose | Stage |
|---|---|---|---|
| Frontend | Next.js App Router, React, TypeScript | Storefront, product pages, admin/seller UX | Current |
| Backend | Go REST API + OpenAPI | Domain services, auth, catalog and orders | M04 |
| Primary database | PostgreSQL | Relational data and transactional consistency | M04 |
| Data access | SQL migrations + sqlc | Versioned schema and typed queries | M04 |
| Data/AI workers | Python | Supplier ingestion, enrichment and translation | M02/M04 |
| Cache/queue | Redis, optional | Cache and job coordination when justified | M04+ |
| Search | PostgreSQL FTS/trigram first | Search and relevance | M07 |

**Database decision:** PostgreSQL is the primary system of record. Marketplace entities have relationships and transactional invariants, so relational constraints, joins and ACID transactions are the right default. Redis is optional and never the canonical database. Do not add MongoDB or multiple databases without a measured use case.

**API decision:** REST + OpenAPI first. GraphQL only if real client query patterns justify it. Use versioned migrations and parameterized SQL.

**Runtime gate:** GitHub Pages is static. M04 cannot begin until the owner enables Go/Python runtimes and approves separate API/database hosting.

## Release and stage map

| Stage | Target | Scope | Status |
|---|---:|---|---|
| M00 | 0.1.0 | Product definition, independence, governance and handoff | Complete |
| M01 | 0.1.x | Storefront foundation, bilingual shell and deployment | Implemented; redesign supersedes original visual direction |
| M02 | 0.2.x | Automated catalog import, normalization and snapshots | Catalog sync active; translation paused |
| M03 | 0.3.0 | Pricing/FX provenance, product details, gallery and comparison | Product detail, gallery and comparison UI implemented; pricing provenance remains |
| M04 | 0.4.0 | Go API, PostgreSQL, migrations, auth boundary and Python workers | Planned; owner runtime/hosting gate |
| M05 | 0.5.0 | Persistent demo commerce journeys, orders, coupons and returns | Planned |
| M06 | 0.6.0 | Admin/seller workspaces, import operations and analytics | Planned |
| M07 | 0.7.0 | Advanced search, relevance, recommendations and saved searches | Planned |
| M08 | 0.8.0 | Browser tests, accessibility, security and performance | Planned |
| M09 | 1.0.0 | Portfolio case study, release and public demo QA | Planned |

## M00 — Product and governance
- [x] Portfolio simulation, not a live retailer.
- [x] Independent repository and owner attribution.
- [x] Safe source policy; no unauthorized scraping.
- [x] Versioning, roadmap and AI handoff rules.
- [x] Architecture decisions recorded.

## M01 — Storefront foundation
- [x] Next.js, React and TypeScript foundation.
- [x] Static export and Pages deployment workflow.
- [x] Responsive editorial marketplace storefront.
- [x] Persian-first RTL and English LTR.
- [x] Search, categories, filters and sorting.
- [x] Quick view, favorites and recently viewed.
- [x] Local cart and simulated checkout.
- [x] Owner attribution and demo/source disclosures.
- [x] GitHub Actions build, typecheck, tests, artifact upload and Pages deploy passed in run 41294484252.
- [ ] Visual QA on phone, tablet and desktop.
- [ ] Keyboard, focus, screen-reader, contrast and reduced-motion audit.
- [ ] Verify published URL, base-path assets and deep links.

## M01.5 — Marketplace redesign and UX foundation
- [x] Re-scope reference to a broad Amazon-style marketplace information architecture without copying proprietary code, branding or assets.
- [x] Pause translation and remove it from the deployment pipeline.
- [x] Rebuild the storefront shell around category-first discovery, prominent search and shopping utilities.
- [x] Add a responsive category mega-menu generated from the live catalog, with product counts and direct category filtering.
- [ ] Add marketplace-grade home modules: curated collections, trending products and richer deal rails.
- [x] Add Amazon product-link intake with domain/ASIN validation and a local-only demo request queue.
- [x] Add marketplace service shortcuts and a transparent FAQ for sourcing, support and demo boundaries.
- [x] Improve product cards, price hierarchy, ratings, availability labels and quick actions.
- [x] Add dedicated static product detail routes for the imported catalog, image gallery, product information and related products.
- [x] Add side-by-side comparison for up to four products, with price, rating, discount, brand and source-stock context.
- [ ] Verify desktop, tablet, mobile, RTL/LTR, keyboard and reduced-motion behavior.
- [x] Add clear demo disclosures for Amazon link intake, seller tools and support paths.
- [ ] Confirm all controls remain functional across the full site.

## M02 — Automated catalog pipeline (active)
- [x] Typed catalog contracts and normalization.
- [x] Invalid-record rejection, source identity and duplicate suppression.
- [x] DummyJSON source adapter.
- [x] Translation provider interface with honest unavailable/error states.
- [x] Scheduled and manual catalog refresh trigger.
- [x] Retry-capable static snapshot generator with last-known-good protection.
- [x] Storefront initializes from generated snapshot and keeps it if live refresh fails.
- [x] Verify snapshot generation (100 products), TypeScript, unit tests, production build and Pages deployment in Actions run 41295039930.
- [x] Add a Google Apps Script Language Service translation bridge using Google Translate (no Gemini).
- [x] Translation intentionally paused; no translation secrets or external translation service required for CI.
- [x] Translate title, description, category and brand while preserving original source fields and provenance.
- [x] Translation cache via source-content fingerprints, sequential throttling, bounded retries and per-product error reporting.
- [ ] Catalog freshness and source diagnostics.
- [ ] Add other suppliers only through authorized APIs/feeds.
**Acceptance:** products import automatically, the storefront survives source outages, and original source text remains intact. Translation is out of scope until explicitly reactivated.

## M03 — Pricing and product experience
- [x] Integer-cent pricing, fixed-point FX and BigInt intermediate arithmetic.
- [x] Explicit rounding and invalid-input tests.
- [ ] FX provider, source timestamp and stale-rate policy.
- [ ] Transparent source/FX/markup/shipping/tax price breakdown.
- [x] Dedicated static product details for the imported catalog, responsive image gallery and related products.
- [ ] Attributes, ratings context and source provenance.
- [x] Product comparison modal and persistent comparison shortlist.
- [ ] Recently viewed page and richer browsing history.

## M04 — Backend and database platform
- [ ] Owner enables Go runtime and approves separate API hosting.
- [ ] Owner enables Python runtime or an approved worker runner.
- [ ] Provision PostgreSQL and secure connection management.
- [ ] Define ERD/schema for catalog, variants, users, roles, carts, orders and audit events.
- [ ] Add versioned migrations and rollback strategy.
- [ ] Build Go REST API with OpenAPI and request validation.
- [ ] Add sqlc typed query layer and parameterized SQL.
- [ ] Design authentication/session and role-based authorization.
- [ ] Add idempotent Python ingestion/translation workers.
- [ ] Add Redis only for measured cache/queue requirements.
- [ ] Integration tests for transactions, constraints and authorization.

## M05 — Commerce simulation
- [ ] Persistent demo cart and checkout draft.
- [ ] Multi-step checkout with validation and demo address book.
- [ ] Simulated payment outcomes; never contact a real payment provider.
- [ ] Persistent demo order history and printable receipt.
- [ ] Coupons and promotions.
- [ ] Returns/refund simulation.

## M06 — Admin and seller workspaces
- [ ] Demo role switcher with real authorization boundaries where API exists.
- [ ] Admin overview and catalog/translation/price diagnostics.
- [ ] Product create/edit/archive.
- [ ] Supplier connector configuration without secret exposure.
- [ ] Seller profile, product management and demo analytics.

## M07 — Search and discovery
- [ ] Faceted filters for brand, rating, price and category.
- [ ] Accessible suggestions and keyboard command palette.
- [ ] PostgreSQL full-text/trigram baseline and relevance tests.
- [ ] Explainable ranking and recommendations.
- [ ] Saved searches and preferences.
- [ ] Empty, loading, offline and error states.

## M08 — Quality and hardening
- [ ] Playwright tests for critical shopping journeys.
- [ ] RTL/LTR accessibility, keyboard, labels and contrast audit.
- [ ] Performance and image-loading audit.
- [ ] Dependency, secret, static-analysis and API security checks.
- [ ] Database backup/restore and migration verification.
- [ ] Reproducible build and deployment smoke test.

## M09 — Portfolio release
- [ ] Persian and English case study.
- [ ] Screenshots, architecture/ERD diagrams and feature tour.
- [ ] Versioned 1.0.0 release and changelog.
- [ ] README verified against actual behavior.
- [ ] Final public demo smoke test.

## Handoff rule
Every AI/developer must read PROJECT_CONTEXT.md, this roadmap, CHANGELOG.md, docs/AI_HANDOFF.md, docs/ARCHITECTURE.md and docs/DECISIONS.md before changing code. Continue from the earliest incomplete item, inspect current Actions, and update documentation only with verified facts.

## Release ledger
| Version | Stage | Status | Notes |
|---|---|---|---|
| 0.1.0 | M00 + M01 | Implemented | Initial independent storefront |
| 0.1.1 | Pricing foundation | Implemented | Integer-cent and fixed-point demo pricing |
| 0.1.4 | M02 catalog pipeline | Implemented | Catalog import and snapshot pipeline; translation bridge parked and removed from CI |
| 0.2.0-dev | M01.5 marketplace redesign | In progress | Category-first shell and updated marketplace visual system |
| 0.2.1-dev | M03 product experience | In progress | Static product detail routes, gallery, buy box and related products added; verification pending |
