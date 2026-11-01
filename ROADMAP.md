# GlobalCart — Master Roadmap

> This is the authoritative progress ledger. Never mark work complete because a file was merely created: implementation and verification are both required.

**Current release:** 0.1.0  
**Active milestone:** M02 — Catalog intelligence foundations  
**Repository:** pooya16pg-spec/globalcart  
**Product boundary:** independent portfolio simulation; no real commerce.

## Release and milestone map

| Milestone | Target release | Scope | Status |
|---|---:|---|---|
| M00 | 0.1.0 | Product definition, independence, governance, handoff docs | Complete |
| M01 | 0.1.0 | Storefront, responsive UI, bilingual shell, demo cart/checkout | Implemented; CI/deploy and accessibility verification pending |
| M02 | 0.2.0 | Typed connectors, normalization, translation pipeline, scheduled catalog | Foundations implemented; provider/scheduler/tests verification pending |
| M03 | 0.3.0 | Pricing engine, FX provenance, product details, compare | Not started |
| M04 | 0.4.0 | Full commerce simulation, order history, coupons, returns | Not started |
| M05 | 0.5.0 | Admin and seller workspaces, catalog operations, analytics | Not started |
| M06 | 0.6.0 | Advanced search, facets, recommendations, saved searches | Not started |
| M07 | 0.7.0 | Automated tests, accessibility, performance, security, CI gates | Not started |
| M08 | 1.0.0 | Portfolio case study, release, final QA and public demo | Not started |

## M00 — Product definition and governance
- [x] Define portfolio simulation, not live store.
- [x] Establish independent repository and owner attribution.
- [x] Define safe source policy; prohibit unauthorized scraping.
- [x] Establish semantic versioning and AI handoff requirements.
- [x] Record architecture and product decisions.

## M01 — Foundation and storefront
- [x] Next.js App Router, React and TypeScript foundation.
- [x] Static export configuration and Pages deployment workflow.
- [x] Responsive editorial retail storefront.
- [x] Persian-first RTL and English LTR switch.
- [x] Demo catalog API with local fallback.
- [x] Search, category filters, price range and sorting.
- [x] Product quick view, favorites and recently viewed state.
- [x] Cart quantity controls, removal and local persistence.
- [x] Demo checkout with generated reference; no payment API.
- [x] Configurable illustrative USD/Toman conversion.
- [x] Discount-aware product and cart totals.
- [x] Owner attribution and demo/source disclosure.
- [ ] Confirm clean Actions build after CI workflow repair.
- [ ] Verify published Pages URL and deep-link behavior.
- [ ] Keyboard, focus, screen-reader and reduced-motion audit.
- [ ] Mobile/tablet/desktop visual QA.
**Acceptance:** complete storefront browsing and simulated cart journey; no real-commerce claims.

## M02 — Catalog intelligence (active)
- [x] Typed catalog product/result contracts.
- [x] Normalization and malformed-record rejection.
- [x] Source identity, import timestamp and duplicate suppression.
- [x] DummyJSON adapter separated from storefront UI.
- [x] Translation provider interface and explicit unavailable/error states.
- [x] Unit tests authored for normalization, pricing and translation orchestration.
- [ ] Verify test suite passes in latest Actions run.
- [ ] Add provider implementation (server-side only; no browser secrets).
- [ ] Translate title, description, category and attributes to Persian.
- [ ] Preserve original text, translated text, provider, locale and timestamp.
- [ ] Translation cache, retries, throttling and failure reporting.
- [ ] Add Amazon connector only after authorization/eligibility review.
- [ ] Generate a versioned static catalog JSON artifact.
- [ ] Scheduled GitHub Actions catalog refresh and manual dispatch.
- [ ] Catalog import summary and source diagnostics page.
**Acceptance:** repeatable automatic import and honest localization status; no manual product-by-product entry.

## M03 — Pricing and product details
- [x] Initial pure pricing calculation foundation with input validation.
- [ ] Integer minor units/decimal-safe arithmetic and rounding tests.
- [ ] FX adapter with source, timestamp and stale-rate behavior.
- [ ] Configurable demo markup/shipping/tax assumptions.
- [ ] Dedicated product detail routes and image gallery.
- [ ] Specs, review summaries, source links and stock disclaimers.
- [ ] Product comparison and recently viewed page.
**Acceptance:** every displayed price is traceable to source price, rate and assumptions.

## M04 — Commerce simulation
- [ ] Persistent cart, favorites and checkout draft.
- [ ] Multi-step checkout with validation and demo address book.
- [ ] Simulated payment success/failure/decline states.
- [ ] Local order history and printable demo receipt.
- [ ] Coupons and promotions.
- [ ] Returns/refund simulation.
**Acceptance:** full testable journey without collecting payment data or creating real orders.

## M05 — Admin and seller workspaces
- [ ] Demo role switcher with clearly simulated permissions.
- [ ] Admin overview and import/translation/price diagnostics.
- [ ] Product create/edit/archive against local demo state.
- [ ] Supplier connector configuration (no secret values).
- [ ] Seller profile and product management.
- [ ] Order operations and demo analytics.
**Acceptance:** meaningful admin/seller interactions, explicitly local and simulated.

## M06 — Search and discovery
- [ ] Faceted filters: brand, rating, price and category.
- [ ] Search suggestions and keyboard-accessible command palette.
- [ ] Explainable ranking and recommendation modules.
- [ ] Saved searches and preference controls.
- [ ] Empty, loading, offline and error states.
**Acceptance:** fast and accessible discovery on the full sample catalog.

## M07 — Quality and release hardening
- [ ] Unit tests for normalization, pricing, translation and state (initial tests exist; CI verification pending).
- [ ] Playwright tests for critical shopping paths.
- [ ] Accessibility audit for RTL/LTR, focus, labels and contrast.
- [ ] Performance and image-loading audit.
- [ ] Dependency, secret and static-analysis checks.
- [ ] CI gates for typecheck, tests and production build.
- [ ] Confirm reproducible build and deployed site.
**Acceptance:** documented, reproducible, passing quality checks.

## M08 — Portfolio release
- [ ] Persian and English case study.
- [ ] Screenshots, architecture diagrams and feature tour.
- [ ] Versioned 1.0.0 release and complete changelog.
- [ ] README verified against actual behavior.
- [ ] Final public demo smoke test.
**Acceptance:** a reviewer can understand the engineering, inspect source and use the demo.

## AI/developer handoff rule
Before any code change, read PROJECT_CONTEXT.md, this roadmap, docs/AI_HANDOFF.md, docs/ARCHITECTURE.md, docs/DECISIONS.md and CHANGELOG.md. Inspect the actual tree and Actions status. Work on the earliest incomplete acceptance item. At the end of each coherent task, update exact checkboxes and PROJECT_CONTEXT.md. Never guess that a previous AI finished something.

## Release ledger
| Version | Milestone | Status | Notes |
|---|---|---|---|
| 0.1.0 | M00 + M01 | Implemented; verification pending | Storefront and initial domain foundations committed |
