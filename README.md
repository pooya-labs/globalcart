# GlobalCart — Intelligent Global Marketplace

**Project 06 · Portfolio Engineering System**  
**Owner & maintainer:** [Pooya Goodarzi](https://github.com/pooya16pg-spec)  
**Repository:** https://github.com/pooya16pg-spec/globalcart  
**Live preview:** https://pooya16pg-spec.github.io/globalcart/

GlobalCart is a commerce-grade marketplace demonstration built to showcase product engineering, catalog ingestion, localization, pricing logic, discovery, and realistic shopping workflows. It is **not a live store**. Checkout, payments, inventory reservations, seller payouts, shipping, and orders are simulations only.

## Product goals
- A polished responsive marketplace experience inspired by modern retail products.
- Automatic product and image import from permitted demo/public catalog sources.
- Persian localization with preserved original text and translation provenance.
- Illustrative USD-to-Toman conversion with configurable assumptions.
- Functional navigation, search, filters, sorting, quick view, favorites, cart, and simulated checkout.
- Explicit distinction between demo data, external data, and real integrations.
- Durable versioned roadmap and AI/developer handoff notes.

## Demo boundaries

| Capability | Behavior |
|---|---|
| Catalog | DummyJSON demo source with local fallback |
| Images | Imported source image URLs; availability depends on source |
| Persian translation | Adapter planned; no fabricated claim of AI translation |
| FX | Configurable illustrative USD-to-Toman rate |
| Cart/favorites | Browser-local demo state |
| Checkout | Simulated; no payment provider is contacted |
| Inventory/shipping | Illustrative only; no stock or shipment is promised |
| Amazon | Authorized adapter boundary only; no scraping or bulk catalog copying |

**Amazon note:** Do not scrape Amazon. Any future connector requires eligible authorized API/program access and compliance with current content, image, price-refresh, attribution, and linking rules. Affiliate product discovery is not a wholesale inventory feed.

## Stack
Next.js 15 · React 19 · TypeScript · custom CSS design system · Lucide React · GitHub Actions · GitHub Pages static export.

## Run locally
Requirements: Node.js 20.9+ and npm.

~~~bash
git clone https://github.com/pooya16pg-spec/globalcart.git
cd globalcart
npm install
npm run dev
~~~

Open http://localhost:3000.

Commands: npm run dev, npm run typecheck, npm run build.

## Documentation
- ROADMAP.md — milestones, checklists, acceptance gates, and release evidence.
- PROJECT_CONTEXT.md — durable project state and instructions for the next AI/developer.
- CHANGELOG.md — version-by-version implementation history.
- docs/ARCHITECTURE.md — system boundaries and data flow.
- docs/DATA_SOURCES.md — catalog, image, translation, and FX provenance.
- docs/DECISIONS.md — accepted architecture decisions.
- docs/RELEASE_PROCESS.md — versioning and handoff protocol.
- CONTRIBUTING.md — engineering workflow.

## Versioning and handoff
GlobalCart uses semantic versions. Every release must update package.json, CHANGELOG.md, ROADMAP.md, and PROJECT_CONTEXT.md together. A roadmap item is complete only when code exists and relevant checks pass. Another assistant must read PROJECT_CONTEXT.md, ROADMAP.md, CHANGELOG.md, and architecture docs before editing.

GlobalCart is a portfolio project authored and maintained by **Pooya Goodarzi**. Third-party product content and images remain subject to their source terms.
