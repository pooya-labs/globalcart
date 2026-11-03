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
- Added `/case-study/` as a bilingual static-exportable portfolio page with feature tour and architecture/data-flow diagram. Browser E2E, Security and published Pages smoke all passed on `a1d53de17a64e3dd80b74e87eca9e4880123538f` (runs `41577703335`, `41577703300`, `41577703347`).
- Added `docs/DATABASE_ERD.md` from all 15 PostgreSQL migrations (27 tables), four domain-grouped Mermaid diagrams and relationship/constraint notes. The case-study page now links to the ERD and shows the four-domain schema summary.
- M09.2 final CI and published-route verification are pending on the newest commit. Then capture actual screenshots and continue remaining M09 release checks. Do not tag 1.0.0 prematurely.
- TGJU remains the only active USD source; never reintroduce Wallex's USDT proxy. Translation remains paused.


## Current continuation checkpoint — M09.2 (2026-11-03)
- M09.1 bilingual case study is CI-verified: Browser E2E `41577703335`, Security `41577703300` and Pages deployment/published smoke `41577703347` passed on `a1d53de17a64e3dd80b74e87eca9e4880123538f`.
- M09.2 adds `docs/DATABASE_ERD.md`, derived from migrations 001–015 and covering 27 actual PostgreSQL tables in four Mermaid ERDs. The case-study route now contains a bilingual schema summary and direct link; the accessibility E2E test asserts the section/link in RTL and LTR.
- Next: validate the latest main head with TypeScript, unit tests, static build, desktop/mobile Browser E2E, Security scans and published Pages smoke. Only mark M09.2 CI-verified after all checks pass.
- Remaining release work: real desktop/mobile screenshots, README final behavior audit, 1.0.0 and final public demo smoke. Do not repeat completed M06 work, do not add public admin/seller routes or real commerce, and do not reintroduce Wallex; TGJU remains the FX source.
- Manual M08 screen-reader, physical-device and live CDN/Core Web Vitals checks remain external QA gates.


## Current continuation checkpoint — M09.3 complete (2026-11-03)
- Main contains the bilingual case study, the 27-table PostgreSQL ERD, README accuracy updates and real desktop/mobile screenshot assets.
- Browser E2E [41596731945](https://github.com/pooya16pg-spec/globalcart/actions/runs/41596731945): 40/40 tests passed, including desktop/mobile screenshot capture, static build, unit tests and TypeScript validation.
- Security [41596731890](https://github.com/pooya16pg-spec/globalcart/actions/runs/41596731890) passed. Pages [41596731889](https://github.com/pooya16pg-spec/globalcart/actions/runs/41596731889) passed the published homepage/asset/product/case-study/ERD smoke check.
- Screenshot publishing workflow [41597084953](https://github.com/pooya16pg-spec/globalcart/actions/runs/41597084953) committed `docs/screenshots/case-study-desktop.png` and `docs/screenshots/case-study-mobile.png` (commit `73625e9a9191b6c850ee0a1c3ea3edca409e0216`). It only publishes after a successful Browser E2E push to main.
- Database ERD source is the 15 SQL migrations under `backend/internal/database/migrations/`; there are 27 tables. No schema migration or hosted backend was introduced.
- **Next action:** M08 manual physical-device, screen-reader, comprehensive RTL/LTR and live CDN/Core Web Vitals QA. These are not proven by browser emulation. Keep version at 0.2.3-dev; do not create/tag 1.0.0 before the external QA gate.
- Keep TGJU as the only USD source; do not reintroduce Wallex, restart translation, redo M06.1–M06.6, or expose public admin/seller/real-commerce routes without explicit owner approval.


## Previous continuation checkpoint — M08.15 public performance lab baseline (implementation initially added, 2026-11-03)
- Main head at start of this slice was `f7435f94329c7b155cb47f8e8dc25420125824a5`; default branch remains `main`, package version remains `0.2.3-dev`.
- Added `.github/workflows/lighthouse-public.yml`: after a successful Pages deployment, it audits the public home, `/product/1/` and `/case-study/` routes in mobile and desktop Lighthouse modes with exact Lighthouse CLI `13.6.0`, Node 22 and Chromium from the committed Playwright toolchain. It keeps JSON reports as a GitHub Actions artifact and summarizes lab scores/metrics. The workflow is also manually dispatchable and weekly.
- Added `scripts/summarize-lighthouse.mjs` and `docs/PERFORMANCE_QA.md`. No PSI API key or external report server is used. This is runner-based lab data, not CrUX/field CWV or physical-device evidence.
- This initial pending note is superseded by the completed measurements below.
- M09.4 remains blocked on actual physical desktop/mobile, real screen-reader, complete site RTL/LTR, and published delivery/field-data checks. The new doc gives the route sequence and evidence table. Keep 0.2.3-dev until those manual gates are truly recorded.
- TGJU remains the only USD source; translation remains paused; do not repeat M06 or enable public admin/seller/real-commerce routes.


## Current continuation checkpoint — M08.15 measured; M09.4 external QA remains (2026-11-03)
- Latest browser suite [41603301095](https://github.com/pooya16pg-spec/globalcart/actions/runs/41603301095): 44/44 tests passed. Includes responsive mobile hero, idle catalog refresh and self-hosted font assertions.
- Latest deploy [41603301049](https://github.com/pooya16pg-spec/globalcart/actions/runs/41603301049) passed production build and the published Pages smoke test; security [41603300997](https://github.com/pooya16pg-spec/globalcart/actions/runs/41603300997) passed.
- Latest published Lighthouse [41604368513](https://github.com/pooya16pg-spec/globalcart/actions/runs/41604368513): home mobile 87 / LCP 3.2 s / TBT 250 ms; product mobile 96; case-study mobile 99. All 6 routes/modes recorded 100 Accessibility and 100 SEO. Full table and first-baseline comparison are in `docs/PERFORMANCE_QA.md`.
- Homepage mobile improved from initial run 41601101775 (74 / 4.9 s / 320 ms) after mobile hero stacking, responsive image candidates, idle sample-catalog revalidation and local `next/font` fonts. These are lab results, not field Core Web Vitals.
- M08.15 is completed. Next/remaining is only M09.4: actual physical device, real screen-reader, comprehensive manual RTL/LTR and normal-network/field-delivery review. Do not tag 1.0.0 or claim field performance until those are recorded.
- Version stays 0.2.3-dev; TGJU is sole FX source; translation paused; don't redo M06, change DB schema, introduce hosted commerce or public admin/seller endpoints without an approved requirement.
