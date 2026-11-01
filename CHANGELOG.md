# Changelog

## 0.1.3 — 2026-11-01

### Added
- Server-side Gemini 2.5 Flash product translation adapter with strict JSON validation.
- Optional scheduled Persian catalog localization job, safely skipped when GEMINI_API_KEY is absent.
- Per-product translation provenance, source-content fingerprint caching, sequential throttling, bounded retries and error states.
- Storefront uses verified Persian product fields when available while preserving original English fields.

### Verification
- CI run 41295683515 passed catalog sync, optional translation step (safely skipped because no key is configured), TypeScript validation, unit tests, production build and GitHub Pages deployment.
- Run: https://github.com/pooya16pg-spec/globalcart/actions/runs/41295683515

## 0.1.2 — 2026-11-01

### Added
- Scheduled and manually triggerable catalog refresh in GitHub Actions.
- Typed static catalog snapshot generator with retries, validation and atomic replacement.
- Last-known-good snapshot protection when the upstream sample API is unavailable.
- Storefront initialization from the generated catalog snapshot.
- Target architecture decision: Go REST/OpenAPI, PostgreSQL with SQL migrations/sqlc, Python data workers and optional Redis.

### Verification
- Actions run 41295039930 passed catalog sync (100 products), TypeScript validation, all unit tests, production build, Pages artifact upload and deployment.
- Run: https://github.com/pooya16pg-spec/globalcart/actions/runs/41295039930

## 0.1.1 — 2026-11-01

### Changed
- Reworked demo pricing to use integer USD cents and fixed-point FX units.
- Used BigInt intermediate arithmetic to avoid floating-point drift during conversion.
- Explicit rounding: discounted USD rounds to cents, then Toman rounds half-up to a whole unit.
- Rejected non-finite discounts and unsafe input/output ranges.
- Expanded pricing tests for fractional values, rounding, clamping and invalid inputs.
- Removed the duplicate Pages starter workflow.

### Verification
- Actions run 41294484252 passed dependency installation, TypeScript validation, unit tests, production build, artifact upload and Pages deployment.

## [0.2.0-dev] - 2026-11-01
### Changed
- Re-scoped the storefront toward a full marketplace experience inspired by common Amazon-style discovery and shopping patterns, with an independent GlobalCart identity.
- Paused translation and removed the translation job and credentials from the GitHub Pages deployment workflow.
- Added a dedicated quick-category navigation rail and began the visual-system migration toward a cleaner, denser marketplace layout.

### Added
- Responsive catalog-driven category mega-menu with live category counts, direct filtering, backdrop dismissal and mobile layout.
- Quick category rail now reflects categories present in the current catalog rather than only the initial hand-picked list.

## 0.2.1-dev — 2026-11-01
### Added
- Catalog-driven category mega-menu and dynamic quick-category rail.
- Dedicated static product routes with image gallery, product attributes, source disclosure, related products and demo cart actions.
- Side-by-side comparison shortlist for up to four products.
- Product cards now navigate directly to dedicated detail pages while retaining quick view.
### Fixed
- Corrected escaped comparison-state declarations and duplicate category-name key found during CI validation.
- Made catalog snapshot commits rebase on the latest main branch to avoid races with UI commits.


## 0.2.2-dev — 2026-11-01
### Added
- Validated Amazon product-link intake with supported marketplace domains, ASIN extraction and a browser-local sourcing request queue.
- Marketplace service shortcuts for sourcing requests, product comparison, planned seller tools and GitHub feedback.
- FAQ content clarifying demo checkout, Amazon data access and comparison behavior.
- Responsive styles for the Amazon request module, service cards and FAQ.
### Notes
- Amazon link submission is a local demonstration only. Live product details, pricing and availability require authorized Amazon API access.
- Translation remains paused and excluded from CI.
### Verification
- Actions run 41299130540 (Run #121) passed catalog refresh, TypeScript validation, unit tests, production build, Pages artifact upload and deployment.
- Run: https://github.com/pooya16pg-spec/globalcart/actions/runs/41299130540


## 0.2.3-dev — 2026-11-01
### Added
- Data-driven homepage merchandising rails for top discounts, top-rated products and fresh discoveries.
- Browser-local recently viewed product rail, updated from both quick view and dedicated product pages, with a clear-history action.
- Product detail price breakdown showing source USD price, catalog discount, illustrative Toman/USD rate and calculated demo total.
### Notes
- FX remains user-configured and illustrative; no live provider, timestamp, shipping, tax or markup is represented.
- Recently viewed data is stored locally in the browser; no account-level history is created.
### Verification
- Actions run 41299322140 (Run #126) passed catalog refresh, TypeScript validation, unit tests, production build, Pages artifact upload and deployment.
- Run: https://github.com/pooya16pg-spec/globalcart/actions/runs/41299322140
