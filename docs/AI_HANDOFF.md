# AI / Developer Handoff Contract

This document is mandatory reading for every assistant, coding agent or human contributor taking over GlobalCart.

## Read first
1. PROJECT_CONTEXT.md — current state, constraints and next task.
2. ROADMAP.md — release milestones and verified completion checklist.
3. CHANGELOG.md — implementation history.
4. docs/ARCHITECTURE.md — boundaries and data flow.
5. docs/DECISIONS.md — accepted decisions and constraints.
6. package.json and the current source files involved in the task.

## Non-negotiable requirements
- Work only in pooya16pg-spec/globalcart.
- Do not connect this project to another portfolio repository or shared runtime.
- Treat the product as a high-fidelity portfolio simulation, never as a real retailer.
- Do not add live payment, fulfillment, inventory reservation or seller payout behavior.
- Do not scrape Amazon. Verify authorized API/program access before any connector work.
- Keep Persian RTL and English LTR behavior intact.
- Preserve original catalog fields and source provenance.
- Do not represent untranslated text as Persian translation.
- Keep secrets server-side and out of Git history.
- GitHub Pages is static hosting; never add server-only runtime code to the client bundle.
- Ask the owner before adding a Go/Python service, PostgreSQL database, external hosting, or paid API dependency.
- Every menu, button and form must perform a real action or show a clear unavailable state.

## Engineering workflow
1. Inspect latest main commit and GitHub Actions run.
2. Identify the earliest unchecked task in ROADMAP.md.
3. Implement a coherent slice, not a superficial placeholder.
4. Add types, validation and tests for business logic.
5. Run dependency install, typecheck, tests and production build.
6. Inspect the diff and confirm GitHub Pages base path.
7. Update ROADMAP.md, PROJECT_CONTEXT.md and CHANGELOG.md with verified status.
8. Never claim the deployment is live unless Actions succeeds and the URL is tested.

## Version progression
- 0.1.x: fix and complete initial storefront and pricing foundations.
- 0.2.0: complete catalog connectors, translation and scheduled catalog generation.
- 0.3.0: pricing/FX engine and detailed product experience.
- 0.4.0: full local commerce simulation.
- 0.5.0: admin and seller workspaces.
- 0.6.0: advanced discovery and recommendations.
- 0.7.0: quality, security, accessibility and performance hardening.
- 1.0.0: documented public portfolio release.

A task is complete only when its code exists, its acceptance criteria pass, and the roadmap/context/changelog reflect the evidence. The repository, not chat memory, is the canonical project state.
