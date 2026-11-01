# GlobalCart — Intelligent Global Marketplace

**Portfolio project 06 · Independent repository**  
**Owner / product engineer:** [Pooya Goodarzi](https://github.com/pooya16pg-spec)  
**Source repository:** https://github.com/pooya16pg-spec/globalcart  
**Public preview:** https://pooya16pg-spec.github.io/globalcart/

GlobalCart is a commerce-grade marketplace **simulation** built to demonstrate product engineering: responsive retail UX, catalog ingestion, localization, pricing, discovery, shopping flows, dashboards, automated quality checks and deployment. It is not a live retailer. No real goods are sold and no payment, inventory reservation, shipment, seller payout or refund is processed.

## Current status — v0.2.1-dev · marketplace redesign

Implemented:
- Next.js 15 App Router, React 19 and TypeScript.
- Responsive marketplace storefront with Persian RTL and English LTR, category-first navigation and refreshed retail design system.
- Dedicated statically generated product detail pages for all imported catalog entries, including gallery, pricing panel, product facts and related products.
- Typed catalog import foundation plus a scheduled/manual snapshot generator with retry and last-known-good fallback (100 products imported; CI, tests, build and deployment passed in run 41295039930).
- Typed catalog normalization, validation, source identity, timestamps and deduplication foundation.
- Search, category filtering, price range, sorting, quick view, favorites and recently viewed.
- Browser-local cart, quantity changes, removal and simulated checkout.
- Configurable illustrative USD-to-Toman rate; discounted prices are reflected in cart totals.
- Decimal-safe pricing core using integer cents, fixed-point FX and explicit rounding.
- Source/demo disclosures and visible owner attribution.
- GitHub Pages deployment workflow and project handoff documentation.
- Target architecture: Go REST/OpenAPI, PostgreSQL + sqlc, Python workers and optional Redis at the backend milestone.

In progress / not yet claimed as complete:
- Translation is intentionally paused and excluded from CI; original source product text is displayed.
- Production-grade FX feed, price history and product detail pages.
- Compare, admin/seller workspaces, order history and advanced search.
- Expanded domain unit tests; accessibility and browser tests remain pending.
- Visual/accessibility QA and browser tests; deployment is now successful.

\n## Demo boundaries

| Area | Current behavior |
|---|---|
| Product catalog | DummyJSON public sample API; local fixtures if unavailable |
| Images | Source-provided remote image URLs |
| Translation | Paused; source text is preserved |
| Currency | User-configurable illustrative rate; not a live exchange quote |
| Cart / favorites | Local browser state |
| Checkout | Local simulation; never contacts a payment provider |
| Inventory / shipping | Not real, not reserved, not fulfilled |
| Amazon | No scraping; future integration only through authorized access and current terms |

Amazon's public storefront must not be scraped or bulk-copied. Any future adapter requires verified program/API eligibility and compliance with current content, image, price-refresh, attribution and linking rules. Affiliate product discovery is not an unrestricted catalog feed.

## Technology

- **Languages:** TypeScript and CSS (plus HTML through React).
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
- [PROJECT_CONTEXT.md](PROJECT_CONTEXT.md) — current implementation state and next action.
- [docs/AI_HANDOFF.md](docs/AI_HANDOFF.md) — mandatory instructions for a replacement AI/developer.
- [CHANGELOG.md](CHANGELOG.md) — version history.
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — system boundaries and data flow.
- [docs/DATA_SOURCES.md](docs/DATA_SOURCES.md) — source provenance and provider policy.
- [docs/DECISIONS.md](docs/DECISIONS.md) — accepted architecture decisions.
- [docs/RELEASE_PROCESS.md](docs/RELEASE_PROCESS.md) — verification and semantic versioning.
- [CONTRIBUTING.md](CONTRIBUTING.md) — engineering rules.

## Versioning

GlobalCart follows semantic versioning. Every release must update package.json, CHANGELOG.md, ROADMAP.md and PROJECT_CONTEXT.md in the same release commit. A roadmap checkbox is only marked complete after implementation and verification. The repository is the source of truth; chat history is not.

**Built and maintained independently by Pooya Goodarzi.**
