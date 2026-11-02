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
