# GlobalCart Roadmap

**Source of truth:** this file. Check an item only after implementation and verification. Every release also updates CHANGELOG.md, PROJECT_CONTEXT.md, and package.json.

- Current version: 0.1.0
- Current milestone: M01 — foundation and storefront
- Hosting: GitHub Pages static export
- Product type: portfolio simulation; no real commerce.

## M00 — Product definition and governance
- [x] Define portfolio/demo scope and non-production boundaries.
- [x] Define owner attribution and repository identity.
- [x] Define source connector approach; prohibit unauthorized scraping.
- [x] Establish versioned roadmap, changelog, and AI handoff protocol.
Acceptance: a new contributor understands intent, boundaries, and continuation steps.

## M01 — Foundation and storefront (v0.1.x)
- [x] Next.js App Router, React, TypeScript foundation.
- [x] GitHub Pages static export and deployment workflow.
- [x] Responsive retail-inspired storefront.
- [x] Persian RTL and English LTR locale switch.
- [x] Demo catalog source adapter and local fallback.
- [x] Search, category filters, sorting, quick view.
- [x] Favorites and cart interactions.
- [x] Simulated checkout with explicit demo label.
- [x] Illustrative USD-to-Toman price calculation.
- [ ] Keyboard accessibility and reduced-motion audit.
- [ ] Verify first GitHub Actions deployment.
Acceptance: usable storefront on mobile/desktop; data provenance visible; no fake real-payment claims.

## M02 — Catalog intelligence (v0.2.x)
- [ ] Typed connector interface and source registry.
- [ ] Product normalization, stable source IDs, deduplication, category mapping.
- [ ] Image provenance and resilient fallback.
- [ ] Translation provider adapter with caching/retries.
- [ ] Persian title, description, category, and attribute localization.
- [ ] Preserve original text and translation provenance.
- [ ] Import timestamps, diagnostics, and error reporting.
- [ ] Scheduled Actions catalog refresh into generated JSON.
- [ ] Fixture tests for malformed, partial, duplicate records.
Acceptance: repeatable catalog sync; no manual per-product entry; unavailable translation is not falsely presented as translated.

## M03 — Pricing and product detail (v0.3.x)
- [ ] Safe decimal/integer currency model.
- [ ] Configurable demo USD-to-Toman FX source and timestamp.
- [ ] Margin, shipping, illustrative fees, and stale-rate guard.
- [ ] Product detail routes, gallery, specs, reviews, source disclosure.
- [ ] Product comparison and recently viewed items.
- [ ] Unit tests for conversion, rounding, and edge cases.
Acceptance: displayed converted price includes demo rate and assumptions.

## M04 — Commerce simulation (v0.4.x)
- [ ] Persistent local cart, favorites, and checkout draft.
- [ ] Multi-step checkout validation and demo address book.
- [ ] Simulated payment outcomes; no payment network calls.
- [ ] Order confirmation and local order history.
- [ ] Coupon/promotion simulation.
- [ ] Returns/refund workflow simulation.
Acceptance: complete demo journey without sending money or placing a real order.

## M05 — Admin and seller workspaces (v0.5.x)
- [ ] Admin overview with catalog/import/price diagnostics.
- [ ] Product CRUD against local demo state.
- [ ] Supplier connector configuration using non-secret demo settings.
- [ ] Seller profile and product-management simulation.
- [ ] Order management and demo analytics.
- [ ] Clearly simulated role states.
Acceptance: admin/seller flows work locally and do not imply server persistence.

## M06 — Search and discovery (v0.6.x)
- [ ] Faceted filters, price range, brand, rating, availability.
- [ ] Search suggestions and keyboard command palette.
- [ ] Explainable ranking and recommendations.
- [ ] Saved searches and personalization preferences.
- [ ] Search empty/error states and local analytics.
Acceptance: keyboard-accessible discovery with visible data provenance.

## M07 — Quality and release hardening (v0.7.x)
- [ ] Unit/component tests and Playwright critical path.
- [ ] Accessibility audit: keyboard, labels, contrast, focus, RTL.
- [ ] Performance and image-loading audit.
- [ ] Dependency and secret scan.
- [ ] CI quality gates and release notes.
- [ ] Verify clean build and GitHub Pages deployment.
Acceptance: reproducible build, passing checks, known limitations documented.

## M08 — Portfolio release (v1.0.0)
- [ ] English case study and Persian overview.
- [ ] Screenshots, feature tour, architecture diagrams.
- [ ] Versioned release and reviewed changelog.
- [ ] README matches actual implementation.
- [ ] Cross-link from Pooya.OS portfolio.
- [ ] Final public demo smoke test.
Acceptance: reviewer can understand, inspect, and try the project.

## Release ledger
| Version | Milestone | Status | Evidence |
|---|---|---|---|
| 0.1.0 | M00 + M01 | In progress | Initial scaffold commit; Actions pending |
