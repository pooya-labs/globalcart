# Changelog

## 0.1.1 — 2026-11-01

### Changed
- Reworked demo pricing to use integer USD cents and fixed-point FX units.
- Used BigInt intermediate arithmetic to avoid floating-point drift during conversion.
- Made rounding behavior explicit: discounted USD rounds to cents, then Toman rounds half-up to a whole unit.
- Rejected non-finite discounts and unsafe input/output ranges.
- Expanded pricing tests for fractional values, rounding, clamping and invalid inputs.
- Removed the duplicate Pages starter workflow, which attempted npm ci without a lockfile and competed with the project-owned deployment workflow.
- Updated project context and roadmap to document that Go/Python services and PostgreSQL require an external runtime/host beyond static GitHub Pages.

### Verification
- Previous Actions run 41294088588 passed dependency installation, TypeScript validation, its then-current unit tests, production build and Pages artifact upload.
- The new decimal-pricing tests require a fresh Actions run.
- GitHub Pages deployment previously returned HTTP 404 because Pages was not enabled/configured for the repository at that time.

# Changelog

## 0.1.0 — 2026-11-01

### Added
- Initial independent GlobalCart portfolio repository and owner attribution.
- Next.js 15, React 19 and TypeScript foundation.
- Responsive editorial marketplace storefront with Persian RTL and English LTR.
- DummyJSON public sample catalog integration and local fallback.
- Search, categories, filtering, sorting, quick view and favorites.
- Browser-local cart, quantity controls and simulated checkout.
- Configurable illustrative USD-to-Toman pricing and discount-aware totals.
- Typed catalog model, normalization, malformed-record rejection, source identity, timestamps and deduplication.
- Translation provider contract with explicit unavailable/error states and original-text preservation.
- Pure pricing calculation foundation with input validation.
- Initial Node test-runner coverage for catalog normalization, pricing and translation orchestration.
- GitHub Pages workflow, milestone roadmap, architecture and AI handoff documentation.

### Verification notes
- Initial Actions failures were caused by npm caching requiring a lockfile that did not exist.
- Workflow changed to install without cache. A build/typecheck/artifact-upload run succeeded.
- The deploy job returned HTTP 404 because GitHub Pages has not yet been enabled/configured for this repository.
- A later workflow revision added unit tests; its latest result is pending verification.
- Automatic Persian product translation, scheduled catalog generation, live FX and real Amazon API integration are not implemented.
- No real transactions or fulfillment are performed.
