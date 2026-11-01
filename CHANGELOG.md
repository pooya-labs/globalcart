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
