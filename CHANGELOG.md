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
