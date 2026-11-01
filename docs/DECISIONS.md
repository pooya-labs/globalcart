# Architecture Decision Records

## ADR-001 — Portfolio simulation, not production commerce
**Status:** Accepted · 2026-11-01

GlobalCart demonstrates realistic commerce workflows but does not sell goods. Payment, inventory, fulfillment, seller payouts, and orders are simulations. No sensitive payment details are collected.

## ADR-002 — Static-first deployment
**Status:** Accepted

Public preview uses Next.js static export on GitHub Pages. Browser interactions are local. Server-only capabilities use adapters and documented demo behavior until a backend is explicitly approved.

## ADR-003 — Source connectors, not scraping
**Status:** Accepted

Catalog ingestion is connector-based. DummyJSON is the initial demo source. Amazon is not scraped; future use requires authorized access and current terms review.

## ADR-004 — Persian-first bilingual experience
**Status:** Accepted

Persian is default with true RTL and locale-aware formatting. English remains available. New UI strings must be added in both languages.
