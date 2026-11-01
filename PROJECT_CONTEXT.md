# GlobalCart — Project Context & AI Handoff

**Owner:** Pooya Goodarzi · [@pooya16pg-spec](https://github.com/pooya16pg-spec)  
**Repository:** https://github.com/pooya16pg-spec/globalcart  
**Current version:** 0.1.0 · **Branch:** main · **Updated:** 2026-11-01

## Product intent
GlobalCart is Project 06 in Pooya Goodarzi's engineering portfolio. It is a realistic, commerce-grade, fully interactive marketplace demonstration, **not a store for selling real goods**. It showcases product engineering: catalog ingestion, localization, pricing, search, shopping workflows, dashboards, testing, documentation, and deployment.

## Non-negotiable boundaries
- Never collect real payment details or call a real payment gateway.
- Never claim an order, inventory reservation, shipment, payout, or refund is real.
- Cart and checkout are browser-local demonstrations.
- Label imported demo content, local fixtures, illustrative calculations, and live external data distinctly.
- Never scrape Amazon or bulk-copy its catalog. A future Amazon adapter requires verified eligibility, authorization, and terms compliance.
- Never expose API secrets in browser code or commit credentials.
- Preserve original product text and translation provenance.
- Automatically import permitted product data/images; do not require manual per-product entry.
- Persian is primary with true RTL; English is also supported.
- Owner attribution must remain visible.

## Portfolio ecosystem
- Pooya.OS — central engineering portfolio / web desktop.
- AuraPulse — market data and paper-trading studio.
- NexusCanvas — architecture graph and infrastructure design studio.
- LuminaGlass — optical and browser audio lab.
- ORBIT — distributed systems intelligence and observability.

Conventions to preserve: Next.js/React/TypeScript, responsive desktop/mobile, polished dark information-rich design, explicit demo/data provenance, owner attribution, detailed docs, GitHub Actions, and honest static-hosting boundaries. ORBIT's continuity docs are the reference for milestone acceptance and handoff discipline.

## Current implementation snapshot — v0.1.0
- [x] Next.js App Router, React 19, TypeScript.
- [x] Static export with GitHub Pages base path /globalcart.
- [x] Pages deployment workflow.
- [x] Persian RTL / English LTR storefront.
- [x] DummyJSON demo catalog adapter with local fallback.
- [x] Search, category filters, sorting, quick view.
- [x] Local favorites/cart and simulated checkout.
- [x] Illustrative FX calculation.
- [x] Roadmap, changelog, architecture, decisions, sources, and release docs.
- [ ] First Actions run and public deployment verification.
- [ ] Browser QA on real mobile/desktop.
- [ ] Automated tests and accessibility audit.
- [ ] Translation provider integration; fallback must remain clearly labelled.

## Mandatory continuation protocol for every AI/developer
Before editing, read this file, ROADMAP.md, CHANGELOG.md, docs/ARCHITECTURE.md, docs/DECISIONS.md, and inspect the latest repository tree and Actions status. Do not rely on chat history as source of truth.

Work only on the next incomplete roadmap milestone unless fixing a blocker. Preserve static Pages hosting unless an ADR changes it. Every visible control must work or clearly state why it is unavailable. Keep demo boundaries honest. Add tests for pricing, normalization, translation, and state changes.

Before finishing: run npm ci, npm run typecheck, npm run build, and relevant tests. Inspect changed files. Update ROADMAP.md only for verified work, CHANGELOG.md, this context file, and package.json when a release boundary is reached. Record exact next action and known limitations. Never claim deployment success without a successful Actions run and published-site check.

## Exact next action
Verify the first GitHub Actions deployment and browser experience. Then begin M02 by extracting the catalog importer into a typed connector registry, adding normalization/deduplication, and implementing a translation adapter that preserves original text and reports provider failures.
