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
- Typed catalog model, normalization, malformed record rejection, source identity, timestamps and deduplication foundation.
- Translation provider contract with explicit unavailable/error states and original-text preservation.
- Pricing calculation foundation.
- GitHub Pages deployment workflow and engineering handoff documentation.

### Verification notes
- Initial Actions failures were caused by npm caching requiring a lockfile that did not exist.
- Workflow changed to install without cache pending lockfile generation; latest run must be checked.
- Automatic Persian product translation is not implemented yet.
- Live FX and real Amazon API integration are not implemented.
- No real transactions or fulfillment are performed.
