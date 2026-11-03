# AI / Developer Handoff Contract

Mandatory reading for every assistant, coding agent or human contributor taking over GlobalCart.

## Read first
1. PROJECT_CONTEXT.md — current state and next action.
2. ROADMAP.md — the single authoritative milestone plan.
3. CHANGELOG.md — implementation history.
4. docs/ARCHITECTURE.md — current and target architecture.
5. docs/DECISIONS.md — accepted decisions.
6. docs/SECURITY_AUTH_DESIGN.md before identity/session/RBAC implementation.
7. docs/DATABASE_MIGRATION_POLICY.md before changing, applying or recovering PostgreSQL migrations.
8. package.json and relevant source files.

## Non-negotiable requirements
- Work only in pooya16pg-spec/globalcart; keep it independent from other portfolio projects.
- Treat GlobalCart as a high-fidelity portfolio simulation, never a live retailer.
- Do not add real payments, fulfillment, inventory reservations or seller payouts.
- Do not scrape Amazon; verify authorized API/feed access before any connector.
- Keep Persian RTL and English LTR correct.
- Preserve original catalog fields and source provenance.
- Never label untranslated content as translated.
- Keep secrets server-side and out of Git history.
- GitHub Pages is static; never add server-only runtime code to the client bundle.
- Every visible control must work or show a clear unavailable state.
- Prioritize polished, modern, accessible UX and responsive behavior over decorative effects.
- PostgreSQL is the planned primary database; Redis is optional and not a source of truth.
- Applied SQL migrations are immutable; use forward-fix migrations and preserve the migration ledger/checksums.
- Ask the owner before adding Go/Python services, PostgreSQL hosting, external infrastructure or paid APIs. The owner can enable extra runtimes when requested.

## Target stack
- Frontend: Next.js, React and TypeScript.
- Backend (M04): Go REST API + OpenAPI.
- Primary database (M04): PostgreSQL + versioned SQL migrations + sqlc.
- Data workers: Python for ingestion and enrichment; translation is currently out of scope.
- Optional cache/queue: Redis only when justified.
- Search baseline: PostgreSQL full-text/trigram.
Introduce each technology only at the roadmap stage where it solves a real requirement.

## Engineering workflow
1. Inspect latest main commit and GitHub Actions.
2. Continue from the earliest incomplete stage in ROADMAP.md.
3. Implement a coherent vertical slice with types, validation and tests.
4. Run install, typecheck, tests and production build.
5. Inspect diff, basePath, RTL/LTR and responsive behavior.
6. Update ROADMAP.md, PROJECT_CONTEXT.md and CHANGELOG.md with verified facts.
7. Never claim deployment is live unless Actions succeeds and the published URL is checked.
8. Report the completed stage, exact commit/run and next stage to the owner.

The repository, not chat history, is the canonical source of truth.


## Current continuation checkpoint (2026-11-03)
- M06.1–M06.4 are complete; do not repeat.
- M06.5 seller offer create/edit/archive is implemented and CI-verified on main. Backend CI `41571374758`, Security `41571374594` and Pages `41571374873` passed.
- New offers begin as drafts; seller edits require the current queue revision; self-publish is blocked and archive is terminal.
- Continue with internal seller profile and demo analytics. Keep public HTTP/UI gated by existing identity, hosting, email and secret approvals.
- TGJU is the active FX source; do not reintroduce Wallex's USDT proxy.


## Current continuation checkpoint — M06.6 (2026-11-03)
- M06.1–M06.5 are CI-verified and complete; do not redo them.
- M06.6 internal seller profile and aggregate demo analytics are implemented and CI-verified on main. Backend CI `41572747044`, Security `41572747081` and Pages `41572746914` passed.
- Profile changes are display-name-only and optimistic; order aggregates are vendor-scoped and paid totals include only simulated payments.
- M06.1–M06.6 internal foundations are complete; keep public seller UI/routes gated by the existing identity/hosting approvals. Supplier connector setup still requires explicit source rights and secret-management approval.


## Current continuation checkpoint — M08.14 (2026-11-03)
- M06.1–M06.6 internal admin/seller services are complete and CI-verified; do not repeat them. Public workspace UI/routes remain gated by same-site HTTPS hosting, identity, email and secret-management approval.
- M08.14 adds asynchronous decoding to the recently-viewed product thumbnails and a base-path-aware desktop/mobile browser regression for visit persistence and image loading/layout policy.
- Browser E2E [41575646034](https://github.com/pooya16pg-spec/globalcart/actions/runs/41575646034) passed all 36 tests; Security [41575646222](https://github.com/pooya16pg-spec/globalcart/actions/runs/41575646222) and Pages deployment/published smoke [41575646035](https://github.com/pooya16pg-spec/globalcart/actions/runs/41575646035) passed on implementation head `c358970b31d2272db98434a761c699355a9dbc57`.
- Next: complete manual screen-reader, physical-device, full-site RTL/LTR and published CDN/Core Web Vitals QA, then proceed with the M09 bilingual portfolio case study and feature tour. Do not claim that browser emulation completes manual QA.
- Latest main may include a later bot commit for the scheduled catalog/lockfile snapshot. Refresh main before writing and use compare-and-swap updates.


## Current continuation checkpoint — M09.1 (2026-11-03)
- M06.1–M06.6 internal admin/seller foundations are complete and CI-verified; do not repeat them. Public workspace UI/routes remain gated by same-site HTTPS hosting, identity, email and secret-management approval.
- M08.14 browser-emulated image/history coverage passed 36 desktop/mobile tests on implementation head `c358970b31d2272db98434a761c699355a9dbc57`; physical-device, screen-reader, full-site RTL/LTR and live CDN/Core Web Vitals QA remain open.
- Added `/case-study/` as a bilingual static-exportable portfolio page with feature tour and architecture/data-flow diagram. It reuses the shared locale preference and is linked from the storefront footer. This work is committed on `main`; CI and published-route verification are pending.
- Next: verify the new route through typecheck, unit tests, static export, desktop/mobile browser checks and published Pages smoke. Then capture actual screenshots and add the PostgreSQL ERD. Do not mark M09 complete or tag 1.0.0 prematurely.
- TGJU remains the only active USD source; never reintroduce Wallex's USDT proxy. Translation remains paused.
