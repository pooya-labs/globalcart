# GlobalCart — Intelligent Global Marketplace

**Project 06 · Independent Portfolio Engineering Project**  
**Owner & maintainer:** [Pooya Goodarzi](https://github.com/pooya16pg-spec)  
**Repository:** https://github.com/pooya16pg-spec/globalcart  
**Live preview:** https://pooya16pg-spec.github.io/globalcart/

GlobalCart is an independent, commerce-grade marketplace demonstration. It showcases product engineering, catalog ingestion, localization, pricing, search and realistic shopping workflows. It is **not a live store**: no real products are sold and no payment, shipment, inventory reservation or seller payout is performed.

## Current release — v0.1.0

The first storefront foundation is implemented:
- Editorial retail design system with responsive desktop and mobile layouts.
- Persian RTL as the default language and English LTR language switch.
- Product catalog imported automatically from the DummyJSON public demo API, with a local fallback fixture.
- Product search, category navigation and filtering, price range, sorting and product quick view.
- Favorites and recently viewed state; cart quantity controls and removal.
- Simulated checkout with generated demo order reference.
- Configurable illustrative USD-to-Toman conversion; discount is applied consistently in product and cart totals.
- Explicit demo/source labels and owner attribution.

**Not implemented yet:** automatic AI translation pipeline, scheduled catalog generation, live FX provider, product detail routes, compare workflow, admin/seller dashboards, and automated browser/accessibility tests. These are tracked in ROADMAP.md. The current Persian UI shell does not mean every imported English product description has already been translated.

## Demo boundaries and source policy

| Capability | Current behavior |
|---|---|
| Catalog | DummyJSON public sample catalog, up to 100 records; local fallback |
| Product images | Image URLs provided by the selected data source |
| Persian translation | Planned for M02; original source text is retained for now |
| FX | User-configurable illustrative USD-to-Toman rate; not a live quote |
| Cart and favorites | Browser-local demo state |
| Checkout | Simulation only; no payment provider is contacted |
| Inventory and shipping | Illustrative only; no stock or shipment is promised |
| Amazon | No scraping. Authorized connector only if eligibility and current terms permit |

**Amazon note:** GlobalCart must not scrape Amazon pages or bulk-copy its catalog. Any future Amazon integration requires eligible authorized API/program access and compliance with current rules for product content, images, price refresh, attribution and linking. Affiliate discovery is not a wholesale inventory feed.

## Technology

- Next.js 15 App Router
- React 19
- TypeScript 5
- Custom responsive CSS design system
- Lucide React icons
- GitHub Actions + GitHub Pages static export

The current deployment target is static hosting. Do not add server-only behavior or expose provider secrets in browser code. Future ingestion/translation automation is planned as a controlled GitHub Actions build step that generates a static catalog artifact.

## Run locally

Requirements: Node.js 20.9+ and npm.

~~~bash
git clone https://github.com/pooya16pg-spec/globalcart.git
cd globalcart
npm ci
npm run dev
~~~

Open http://localhost:3000.

Useful commands:

~~~bash
npm run dev
npm run typecheck
npm run build
~~~

## Documentation — read these before changing code

- ROADMAP.md — milestone checklist, acceptance criteria and release ledger.
- PROJECT_CONTEXT.md — authoritative state and AI/developer handoff.
- CHANGELOG.md — version history.
- docs/ARCHITECTURE.md — system boundaries and data flow.
- docs/DATA_SOURCES.md — catalog, images, translation and FX provenance.
- docs/DECISIONS.md — accepted architecture decisions.
- docs/RELEASE_PROCESS.md — verification and versioning.
- CONTRIBUTING.md — engineering workflow.

## Versioning and AI handoff

GlobalCart uses semantic versioning. Every release must update package.json, CHANGELOG.md, ROADMAP.md and PROJECT_CONTEXT.md together. A roadmap item is complete only when its implementation exists and the relevant checks pass.

A new AI assistant or developer must read PROJECT_CONTEXT.md, ROADMAP.md, CHANGELOG.md, docs/ARCHITECTURE.md and docs/DECISIONS.md before editing. The repository is the source of truth, not chat history.

GlobalCart is an independent portfolio project authored and maintained by **Pooya Goodarzi**. Third-party product content and images remain subject to their source terms.
