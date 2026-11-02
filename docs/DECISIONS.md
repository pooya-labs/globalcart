# Architecture Decision Records

## ADR-001 — Portfolio simulation, not production commerce
**Status:** Accepted · 2026-11-01

GlobalCart demonstrates realistic commerce workflows but does not sell goods. Payment, inventory, fulfillment, seller payouts, and orders are simulations. No sensitive payment details are collected.

## ADR-002 — Static-first deployment
**Status:** Accepted

Public preview uses Next.js static export on GitHub Pages. Browser interactions are local. Server-only capabilities use scheduled jobs/adapters and documented limitations until a backend is explicitly approved.

## ADR-003 — Source connectors, not scraping
**Status:** Accepted

Catalog ingestion is connector-based. DummyJSON is the initial demo source. Amazon is not scraped; future use requires authorized access and current terms review.

## ADR-004 — Persian-first bilingual experience
**Status:** Accepted

Persian is default with true RTL and locale-aware formatting. English remains available. New UI strings must be added in both languages.

## ADR-005 — Translation is provenance-aware and server-side
**Status:** Accepted

Translation is performed by an adapter, not by pretending source text is translated or by exposing API credentials in the browser. Original and translated text, locales, provider, status and timestamp must be stored separately. Missing provider configuration produces an explicit unavailable state.

## ADR-006 — Repository independence
**Status:** Accepted

GlobalCart has its own code, dependencies, roadmap, workflows and deployment. No runtime dependency or code coupling with other portfolio projects is allowed.

## ADR-007 — PostgreSQL as primary database
**Status:** Accepted · 2026-11-01

PostgreSQL is the planned system of record for relational marketplace entities and transactional invariants. Schema changes use reviewed, versioned SQL migrations; Go data access uses sqlc-generated typed methods and parameterized SQL. Redis is optional for measured cache/queue needs and is never the source of truth.

## ADR-008 — Go API and Python data workers
**Status:** Accepted · 2026-11-01

Persistent backend capabilities use a Go REST API with OpenAPI for application/domain operations and Python workers for ingestion, enrichment and translation. Keep Next.js as the frontend. M04 is gated on owner approval of runtimes and separate hosting because GitHub Pages is static.

## ADR-009 — Static catalog snapshots with last-known-good fallback
**Status:** Accepted · 2026-11-01

The storefront consumes a generated JSON snapshot. Scheduled/manual GitHub Actions refreshes it from permitted sources before build. A failed refresh preserves the previous valid snapshot; live browser refresh must not discard a valid snapshot when the source is unavailable.


## ADR-010 — Opaque server-side sessions and deny-by-default RBAC
**Status:** Proposed · 2026-11-01

Use versioned Argon2id password hashes, opaque high-entropy server-side sessions with only token digests stored, host-only Secure/HttpOnly/SameSite cookies, synchronizer CSRF protection and exact Origin/CORS allow-lists. Roles are customer/seller/admin; seller operations are additionally scoped to an active seller membership. Privileged role grants are audited and never self-assigned. Authorization is database-backed and deny-by-default. Do not implement browser login until the owner approves HTTPS same-site frontend/API hosting, email delivery and secret management; the current GitHub Pages origin is not a suitable cross-site cookie deployment.
See `docs/SECURITY_AUTH_DESIGN.md` for the complete schema, flow, endpoint and test proposal.

**Implementation note — 2026-11-02:** migration 011 adds identity/session/RBAC storage and `backend/internal/auth` adds isolated Argon2id password and opaque-token primitives. Neither is wired to public routes. ADR-010 remains Proposed; login, cookies, CSRF middleware and protected endpoints stay disabled pending owner approval of same-site HTTPS hosting, email delivery and secret management.


## ADR-011 — Guest cart capability and demo-only order ledger
**Status:** Accepted for local schema design; write API remains gated

Guest carts use high-entropy opaque bearer capabilities whose SHA-256 digests alone are persisted. Checkout is designed as one PostgreSQL transaction with cart and stock-row locks, immutable order-line snapshots, expiring reservations and per-cart idempotency. Payment outcomes are simulated only; reservation completion never mutates imported stock or triggers fulfillment. Do not expose browser write routes until same-site HTTPS hosting, CSRF, secret management and authorization are approved. See `docs/COMMERCE_DESIGN.md`.
