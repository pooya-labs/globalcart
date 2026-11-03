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
- Work only in the canonical repository pooya-labs/globalcart (formerly pooya16pg-spec/globalcart); keep it independent from other portfolio projects.
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
- Browser E2E [41575646034](https://github.com/pooya-labs/globalcart/actions/runs/41575646034) passed all 36 tests; Security [41575646222](https://github.com/pooya-labs/globalcart/actions/runs/41575646222) and Pages deployment/published smoke [41575646035](https://github.com/pooya-labs/globalcart/actions/runs/41575646035) passed on implementation head `c358970b31d2272db98434a761c699355a9dbc57`.
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
- Browser E2E [41596731945](https://github.com/pooya-labs/globalcart/actions/runs/41596731945): 40/40 tests passed, including desktop/mobile screenshot capture, static build, unit tests and TypeScript validation.
- Security [41596731890](https://github.com/pooya-labs/globalcart/actions/runs/41596731890) passed. Pages [41596731889](https://github.com/pooya-labs/globalcart/actions/runs/41596731889) passed the published homepage/asset/product/case-study/ERD smoke check.
- Screenshot publishing workflow [41597084953](https://github.com/pooya-labs/globalcart/actions/runs/41597084953) committed `docs/screenshots/case-study-desktop.png` and `docs/screenshots/case-study-mobile.png` (commit `73625e9a9191b6c850ee0a1c3ea3edca409e0216`). It only publishes after a successful Browser E2E push to main.
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
- Latest browser suite [41603301095](https://github.com/pooya-labs/globalcart/actions/runs/41603301095): 44/44 tests passed. Includes responsive mobile hero, idle catalog refresh and self-hosted font assertions.
- Latest deploy [41603301049](https://github.com/pooya-labs/globalcart/actions/runs/41603301049) passed production build and the published Pages smoke test; security [41603300997](https://github.com/pooya-labs/globalcart/actions/runs/41603300997) passed.
- Latest published Lighthouse [41604368513](https://github.com/pooya-labs/globalcart/actions/runs/41604368513): home mobile 87 / LCP 3.2 s / TBT 250 ms; product mobile 96; case-study mobile 99. All 6 routes/modes recorded 100 Accessibility and 100 SEO. Full table and first-baseline comparison are in `docs/PERFORMANCE_QA.md`.
- Homepage mobile improved from initial run 41601101775 (74 / 4.9 s / 320 ms) after mobile hero stacking, responsive image candidates, idle sample-catalog revalidation and local `next/font` fonts. These are lab results, not field Core Web Vitals.
- M08.15 is completed. Next/remaining is only M09.4: actual physical device, real screen-reader, comprehensive manual RTL/LTR and normal-network/field-delivery review. Do not tag 1.0.0 or claim field performance until those are recorded.
- Version stays 0.2.3-dev; TGJU is sole FX source; translation paused; don't redo M06, change DB schema, introduce hosted commerce or public admin/seller endpoints without an approved requirement.


## Current continuation checkpoint — M09.4 automated route accessibility (2026-11-03)
- Extended the axe suite to cover recently viewed history when populated and empty, in Persian RTL and English LTR. The test waits for actual browser-local history persistence before navigating, making desktop/mobile coverage deterministic.
- axe findings fixed: recently viewed description/note contrast, rating text contrast, clear-history accessible-name mismatch, and locale toggle visible-label/name alignment.
- Final Browser E2E [41606569768](https://github.com/pooya-labs/globalcart/actions/runs/41606569768) passed **48/48 tests**. Pages deployment/smoke [41606569808](https://github.com/pooya-labs/globalcart/actions/runs/41606569808) passed. Security [41606569782](https://github.com/pooya-labs/globalcart/actions/runs/41606569782) passed.
- Public Lighthouse [41606821395](https://github.com/pooya-labs/globalcart/actions/runs/41606821395) succeeded. Performance lab results remain those documented in `docs/PERFORMANCE_QA.md` (home mobile 87 / LCP 3.2 s / TBT 250 ms).
- Remaining release gate: actual assistive-technology/screen-reader and physical-device testing, full-site manual RTL/LTR review, and normal-network/field data. Automated axe/Playwright do not substitute for these. Keep the release at 0.2.3-dev.


## Current continuation checkpoint — M09.4 Firefox/WebKit accessibility (2026-11-03)
- Added focused Firefox and WebKit desktop projects for `e2e/accessibility.spec.ts`; full storefront flows continue on Chromium desktop and Pixel 7 mobile.
- CI now installs Chromium, Firefox and WebKit. Browser E2E [41607322346](https://github.com/pooya-labs/globalcart/actions/runs/41607322346) passed **66/66**: 48 Chromium desktop/mobile scenarios + 9 Firefox accessibility tests + 9 WebKit accessibility tests.
- The matrix includes RTL/LTR home/case-study, populated/empty recently viewed, WCAG axe checks, reduced motion, command palette, shopping drawers/modals, checkout and focus behavior.
- Pages [41607322372](https://github.com/pooya-labs/globalcart/actions/runs/41607322372) and Security [41607322262](https://github.com/pooya-labs/globalcart/actions/runs/41607322262) passed. Main's bot-generated current snapshot is `66467f63bb68a94967fc9be27354d89821a63c6d`.
- Still open: physical desktop/mobile QA, real screen-reader, full site manual RTL/LTR, ordinary-network CDN and field metrics. Keep 0.2.3-dev; don't tag 1.0.0 before those gates.


## Latest handoff checkpoint — M09.4 manual release evidence remains (2026-11-03)

- Repository: `pooya16pg-spec/globalcart`; branch: `main`; latest inspected pre-documentation head: `4504161bc43262383a2b690a3454df62bb94bce7`. Do not create a duplicate feature branch or repeat completed roadmap work.
- Latest Browser E2E [41607322346](https://github.com/pooya-labs/globalcart/actions/runs/41607322346) passed 66/66 across Chromium desktop/mobile, Firefox and WebKit accessibility projects. Pages [41608274084](https://github.com/pooya-labs/globalcart/actions/runs/41608274084) and Security [41608274114](https://github.com/pooya-labs/globalcart/actions/runs/41608274114) passed.
- Latest published Lighthouse [41608446036](https://github.com/pooya-labs/globalcart/actions/runs/41608446036) passed all six audits. The latest mobile Performance scores are storefront 91 (LCP 3.3 s, TBT 150 ms), product 95 (LCP 2.9 s, TBT 90 ms), and case study 98 (LCP 1.0 s); Accessibility, Best Practices and SEO are 100 in all six runs. See `docs/PERFORMANCE_QA.md` for the complete table and limitations.
- Latest `data/fx-rate.json` snapshot is 272,905 Toman/USD (2,729,050 Rial), but the TGJU source marker resolves to `2026-11-02T13:45:43Z`; as of this checkpoint it is stale under the configured 15-minute threshold. Do not manually advance its timestamp, report it as fresh, use Wallex/USDT, or add a fixed-rate fallback. Schedule is configured for every five minutes; the committed snapshot alone does not prove cadence reliability.
- **Next action:** M09.4 only — record real physical desktop/mobile, actual platform screen-reader, whole-site manual RTL/LTR and normal-network/field delivery checks. Browser emulation/Lighthouse lab runs do not satisfy these. Keep version at `0.2.3-dev` and do not tag 1.0.0 until the evidence is real.
- Do not redo M06, M09.1–M09.3, TGJU integration, the existing cross-browser tests or the Pages case study/ERD work. Translation remains paused; public admin/seller routes and real commerce remain disabled absent explicit approvals.


## Latest handoff checkpoint — M08.16 implementation verified; M09.4 external QA remains (2026-11-03)

- Repository: `pooya16pg-spec/globalcart`; branch: `main`; current inspected main before this documentation-only checkpoint: `18a426cbfcaadd1782d7d142e71dcc0613033bae`.
- Route-scoped preconnect hints: home (Unsplash + DummyJSON), product detail (DummyJSON only), case study (neither). The new route-transition Playwright test passes. Browser E2E [41612960428](https://github.com/pooya-labs/globalcart/actions/runs/41612960428) passed **70/70**; Pages [41612960450](https://github.com/pooya-labs/globalcart/actions/runs/41612960450) passed deployment and public smoke; Security [41612960405](https://github.com/pooya-labs/globalcart/actions/runs/41612960405) passed.
- Latest Lighthouse [41613166735](https://github.com/pooya-labs/globalcart/actions/runs/41613166735) completed all six published route/mode audits. Mobile Performance: home 88 / LCP 3.2 s / TBT 180 ms; product 97 / 2.6 s / 70 ms; case study 98 / 0.9 s / 160 ms. All six Accessibility, Best Practices and SEO results are 100. Raw reports confirm the route-appropriate preconnects are not flagged unused. See `docs/PERFORMANCE_QA.md`.
- FX publication: a changed validated TGJU snapshot now sends `globalcart-fx-updated` via same-repository `repository_dispatch`; the Pages workflow listens and rebuilds the static export. Current FX push run [41612209256](https://github.com/pooya-labs/globalcart/actions/runs/41612209256) found no change and correctly skipped dispatch. Do not force a price difference just to test the event; verify the dispatch-to-Pages chain on the next genuine source update.
- Snapshot: 271,500 Toman/USD (2,715,000 Rial), marker `11 آبان` with `date_only` precision; it is stale. No timestamp or live rate should be fabricated.
- Cron: `3-59/5 * * * *` (five-minute GitHub schedule, best-effort only). Actions history still shows one TGJU scheduled execution; cadence reliability remains under observation.
- **Next action:** verify the first real changed-quote dispatch and continue only M09.4 real physical desktop/mobile, platform screen-reader, full-site manual RTL/LTR and normal-network/field-delivery checks. Keep `0.2.3-dev`; do not tag 1.0.0 before manual QA is genuinely recorded.
- Do not redo M06, M09.1–M09.3, TGJU parsing, cross-browser suite, ERD, or Pages setup. Translation remains paused; public admin/seller/real-commerce routes remain disabled absent explicit approval.


## Latest handoff checkpoint — M09.4.1 CI closed; M09.4 manual evidence remains (2026-11-03)

- Canonical repository is pooya-labs/globalcart; the current branch is main. The GitHub account profile now resolves as pooya-labs.
- Pages now serves from https://pooya-labs.github.io/globalcart/. The repository-dispatch deployment [41616698725](https://github.com/pooya-labs/globalcart/actions/runs/41616698725) confirms that exact URL and passed homepage, JavaScript asset, product deep-link and case-study smoke.
- Latest Lighthouse [41613950109](https://github.com/pooya-labs/globalcart/actions/runs/41613950109): mobile Performance 92 (home), 97 (product), 100 (case study); desktop Performance 100 on all three. All six modes scored 100 Accessibility, Best Practices and SEO. Full metrics and raw image findings are in docs/PERFORMANCE_QA.md.
- Raw image findings estimate 160/108 KiB potential savings on home desktop/mobile and 82/84 KiB on product. These are mostly 300×300 catalog thumbnails and a 1000×1000 alternate gallery image displayed at tiny dimensions. No supported resize API for existing DummyJSON product CDN URLs is established; do not rewrite URLs or add a third-party proxy.
- M09.4.1 public API regression guards are CI verified: seller/cart/checkout writes stay unavailable, catalog mutation methods are rejected and pre-auth CORS permits only GET/OPTIONS. Backend CI [41353463006](https://github.com/pooya-labs/globalcart/actions/runs/41353463006) succeeded.
- Browser E2E [41612960428](https://github.com/pooya-labs/globalcart/actions/runs/41612960428) passed 70/70; Pages [41613825451](https://github.com/pooya-labs/globalcart/actions/runs/41613825451) and Security [41613825512](https://github.com/pooya-labs/globalcart/actions/runs/41613825512) passed.
- TGJU remains the only USD source. The preceding exact-clock quote was 274,500 Toman/USD at 12:20:14 Tehran; a later refresh returned 271,500 Toman/USD with date-only marker 11 آبان and no exact timestamp. The latest saved quote is therefore stale; never infer a fresh timestamp.
- Cron is 3-59/5 * * * *; only one TGJU scheduled run is visible. GitHub schedule delivery is best-effort, so do not claim reliable five-minute updates.
- Next: physical desktop/mobile, real platform screen-reader, full-site manual RTL/LTR and normal-network/field-delivery evidence. Also verify the dispatch-to-Pages chain after a genuine TGJU quote change. Keep 0.2.3-dev and do not tag 1.0.0 before genuine manual QA evidence exists.
- Do not redo M06, M09.1–M09.3, M09.4.1, TGJU parsing, image preconnect work, browser automation or the Pages relay. Translation remains paused; public admin/seller routes and real commerce remain disabled without approval.

## Latest handoff checkpoint — TGJU dispatch verified; Pages identity corrected (2026-11-03)

- Repository: pooya-labs/globalcart; branch main; current application version 0.2.3-dev.
- Canonical public site: https://pooya-labs.github.io/globalcart/. The previous pooya16pg-spec.github.io host returned 404 in Lighthouse; Pages deployment run [41616698725](https://github.com/pooya-labs/globalcart/actions/runs/41616698725) confirms the new host and passed homepage, JS asset, product deep-link and case-study smoke.
- TGJU-only FX sync: genuine 271,500 → 274,500 Toman/USD change; raw 2,745,000 Rial; Tehran source clock 12:20:14; sourceUpdatedAt 2026-11-03T08:50:14.000Z. FX run [41616645902](https://github.com/pooya-labs/globalcart/actions/runs/41616645902) succeeded and committed [624eeb9](https://github.com/pooya-labs/globalcart/commit/624eeb994dbf21971fdec0d396d374af9364483c). It then emitted globalcart-fx-updated; Pages repository_dispatch [41616698725](https://github.com/pooya-labs/globalcart/actions/runs/41616698725) completed successfully. The stored timestamp is stale under the 15-minute rule; do not present it as fresh or modify the timestamp.
- Backend CI [41616679743](https://github.com/pooya-labs/globalcart/actions/runs/41616679743), Catalog CI [41616687522](https://github.com/pooya-labs/globalcart/actions/runs/41616687522), Security [41616687542](https://github.com/pooya-labs/globalcart/actions/runs/41616687542) all succeeded after first Node 24 action upgrades.
- Lighthouse target now uses https://pooya-labs.github.io/globalcart. The current mainline change also migrates remaining deploy/E2E/Lighthouse/screenshot/security actions to Node24-compatible releases. Check fresh E2E, Pages and Lighthouse runs after this commit.
- Remaining release gate M09.4: actual physical desktop/mobile, platform screen reader, full-site manual RTL/LTR, normal network and field metrics. Do not tick these based on CI/emulation; keep version 0.2.3-dev and do not tag 1.0.0 yet.
- Avoid repeating M06, M09.1–M09.3, the now-verified M09.4.1 guard suite, TGJU parsing, route preconnect changes or the real dispatch relay.

## Latest handoff checkpoint — latest TGJU snapshot and canonical Pages audit (2026-11-03)

- Canonical repository and Pages: https://github.com/pooya-labs/globalcart and https://pooya-labs.github.io/globalcart/.
- Latest published Lighthouse on the new Pages host: [41618638260](https://github.com/pooya-labs/globalcart/actions/runs/41618638260) succeeded. Mobile Performance home 91 / LCP 3.2 s / TBT 150 ms, product 94 / 3.0 s / 100 ms, case study 99 / 0.9 s / 110 ms. Desktop Performance 100 on all three; Accessibility, Best Practices and SEO 100 in all six. Lab only.
- Latest saved TGJU snapshot: 2,715,000 Rial = 271,500 Toman/USD, source marker 11 آبان with date_only precision and sourceUpdatedAt null. Stale by design. Refresh [41618464249](https://github.com/pooya-labs/globalcart/actions/runs/41618464249) committed it; its event-driven Pages run [41618569410](https://github.com/pooya-labs/globalcart/actions/runs/41618569410) passed.
- The prior quote 2,745,000 Rial / 274,500 Toman at Tehran 12:20:14 was observed during an earlier real TGJU pull; the subsequent fetch is the latest saved record. Never adjust source timestamps yourself or add a fallback provider.
- Node 24 action migration: the deployment flow v7/v5 actions passed in Pages run [41618842051](https://github.com/pooya-labs/globalcart/actions/runs/41618842051); Security [41618841960](https://github.com/pooya-labs/globalcart/actions/runs/41618841960) passed. Browser E2E [41618464270](https://github.com/pooya-labs/globalcart/actions/runs/41618464270) completed success.
- M09.4 still requires real physical desktop/mobile, platform screen-reader, full-site manual RTL/LTR and normal-network/field evidence. Keep 0.2.3-dev and do not tag 1.0.0 before the manual checks are real.

## Latest handoff checkpoint — final Lighthouse and release status (2026-11-03)

- Canonical repository and site: https://github.com/pooya-labs/globalcart and https://pooya-labs.github.io/globalcart/.
- Latest Lighthouse [41619642994](https://github.com/pooya-labs/globalcart/actions/runs/41619642994) succeeded: mobile home 94 / LCP 3.0 s / TBT 90 ms, product 99 / 2.3 s / 60 ms, case study 100 / 0.9 s / 90 ms; desktop Performance 100 on all routes. Accessibility, Best Practices and SEO 100 in all six modes. Lab only.
- Latest TGJU JSON is 2,715,000 Rial / 271,500 Toman/USD with 11 آبان date-only source marker and no timestamp. Stale by policy. Run [41618464249](https://github.com/pooya-labs/globalcart/actions/runs/41618464249) updated it; Pages relay [41618569410](https://github.com/pooya-labs/globalcart/actions/runs/41618569410) succeeded.
- Latest Pages deploy [41619533670](https://github.com/pooya-labs/globalcart/actions/runs/41619533670) and Security [41619533919](https://github.com/pooya-labs/globalcart/actions/runs/41619533919) passed. Browser E2E [41618464270](https://github.com/pooya-labs/globalcart/actions/runs/41618464270) passed; screenshot publishing [41619153373](https://github.com/pooya-labs/globalcart/actions/runs/41619153373) succeeded.
- Remaining M09.4 gates are physical devices, real screen reader, comprehensive manual RTL/LTR and normal-network/field data. No 1.0.0 tag until those are actually performed. Current version remains 0.2.3-dev.

## Latest handoff checkpoint — Lighthouse 41620637413 (2026-11-03)

- Canonical repository: https://github.com/pooya-labs/globalcart; Pages: https://pooya-labs.github.io/globalcart/.
- Latest Lighthouse [41620637413](https://github.com/pooya-labs/globalcart/actions/runs/41620637413) passed six audits. Mobile Performance: home 93 / LCP 3.2 s / TBT 80 ms, product 95 / 2.9 s / 50 ms, case study 99 / 0.9 s / 100 ms. All desktop Performance scores 100; Accessibility, Best Practices and SEO all 100 in all six. Lab-only, not field CWV.
- Raw LCP: homepage hero and product main image are discoverable in the original HTML and use eager/high priority. Home image subparts ~96 ms TTFB / 11 ms resource delay / 238 ms resource duration / 24 ms element render. Do not add preload or unsupported resize params without new evidence.
- Latest FX file remains 2,715,000 Rial = 271,500 Toman/USD from TGJU, date-only 11 آبان, `sourceUpdatedAt:null`; stale by policy. The real changed-quote dispatch relay was verified twice and should not be redone.
- Latest Pages deployment [41620332749](https://github.com/pooya-labs/globalcart/actions/runs/41620332749) and Security [41620332620](https://github.com/pooya-labs/globalcart/actions/runs/41620332620) passed. Browser E2E [41618464270](https://github.com/pooya-labs/globalcart/actions/runs/41618464270) passed.
- Next work: only M09.4 owner evidence (physical devices, actual platform screen-reader, complete RTL/LTR, ordinary-network/field delivery). Keep 0.2.3-dev; no 1.0.0 tag yet.

## Latest handoff checkpoint — canonical links guarded, 74/74 browser checks green (2026-11-03)

- Repository/Pages: https://github.com/pooya-labs/globalcart and https://pooya-labs.github.io/globalcart/; version 0.2.3-dev.
- New regression covers public GitHub links on homepage and Case Study plus Case Study canonical metadata. Browser E2E [41622800540](https://github.com/pooya-labs/globalcart/actions/runs/41622800540) passed **74/74**; the new assertion passed Chromium desktop/mobile, Firefox and WebKit.
- Pages [41622800624](https://github.com/pooya-labs/globalcart/actions/runs/41622800624) and Security [41622800675](https://github.com/pooya-labs/globalcart/actions/runs/41622800675) passed.
- Latest Lighthouse [41622967095](https://github.com/pooya-labs/globalcart/actions/runs/41622967095): mobile home 91 / LCP 3.2 s / TBT 90 ms, product 96 / 2.7 s / 40 ms, case study 100 / 0.9 s / 70 ms; desktops 100 and all Accessibility/Best Practices/SEO 100. Recent home-mobile repeats ranged 80–93 / LCP 3.2–4.7 s; do not claim field CWV.
- Latest TGJU snapshot remains 2,715,000 Rial / 271,500 Toman/USD, date-only marker 11 آبان and no source timestamp; stale by policy. Publication relay already verified.
- Next: M09.4 owner evidence only—physical desktop/mobile, actual screen-reader, full manual RTL/LTR and normal-network/field delivery. No 1.0.0 tag before those checks.

## Latest continuation checkpoint — mainline release-readiness review (2026-11-03)

- Canonical repository: `https://github.com/pooya-labs/globalcart`; canonical GitHub Pages: `https://pooya-labs.github.io/globalcart/`. Work on `main` only; do not create a duplicate feature branch or PR for documentation-only follow-up.
- Current main at the time of this checkpoint: `a6d8d3d4b7ab74b24b317b346db9b0b254fd1836`. The verified docs checkpoint `1363ee0951b92161c4268c7667f6d9f1382c4de8` refreshed README, ROADMAP, PROJECT_CONTEXT and PERFORMANCE_QA. Pages deployment/public smoke [41626708812](https://github.com/pooya-labs/globalcart/actions/runs/41626708812) and Security [41626708835](https://github.com/pooya-labs/globalcart/actions/runs/41626708835) passed. The next bot commit only persisted the refreshed catalog snapshot and lockfile.
- Latest full Browser E2E [41622800540](https://github.com/pooya-labs/globalcart/actions/runs/41622800540) passed **74/74** across Chromium desktop/mobile, Firefox and WebKit. It was run on application head `54ca716902ceca38f9033d81960307a2cf3feb67`; subsequent changes were docs and generated catalog/lockfile snapshots. The E2E workflow path filters intentionally do not rerun browser suites for docs-only commits.
- Latest published Lighthouse [41626880818](https://github.com/pooya-labs/globalcart/actions/runs/41626880818) completed all six canonical Pages audits. Mobile: storefront 93 / LCP 3.2 s / TBT 40 ms; product 94 / 3.1 s / 40 ms; case study 100 / 0.9 s / 60 ms. Desktop Performance is 100 on all routes; Accessibility, Best Practices and SEO are 100 in all six modes. These are GitHub-hosted lab measurements, not field data.
- The committed TGJU-only FX snapshot is 2,715,000 Rial = 271,500 Toman/USD, with date-only marker `11 آبان` and no exact source timestamp; it is stale under the 15-minute policy. Never fabricate a timestamp, switch back to Wallex/USDT, or add a fixed fallback. The genuine changed-quote dispatch-to-Pages relay has already been verified; do not re-test it using invented price changes.
- A legacy branch `feat/m04-go-postgres-foundation` remains on GitHub but is highly divergent from current main (compare reported 444 commits ahead and 612 behind). Do not merge, force-update, rebase or delete it as cleanup; continue on `main` and leave the historical pointer untouched unless its contents are explicitly investigated first.
- The only release gate is M09.4 owner evidence: physical desktop/mobile, real platform screen-reader, full manual RTL/LTR, and normal-network/published-delivery or eligible field metrics. These cannot honestly be completed by CI emulation. Keep version `0.2.3-dev`; do not tag `1.0.0` until real evidence is recorded.
- Do not repeat M06.1–M06.6, M09.1–M09.3, TGJU parsing/publication relay, route-scoped preconnect changes, ERD, screenshot publishing or the existing cross-browser E2E suite. Translation stays paused; public admin/seller routes and real payment/fulfillment remain disabled.


## Latest performance checkpoint — Lighthouse 41626880818 (2026-11-03)

- Run [41626880818](https://github.com/pooya-labs/globalcart/actions/runs/41626880818) completed successfully on the canonical published Pages host. Mobile Performance: storefront 93 / LCP 3.2 s / TBT 40 ms; product 94 / 3.1 s / 40 ms; Case Study 100 / 0.9 s / 60 ms. Desktop Performance 100 in all three routes; Accessibility, Best Practices and SEO 100 across all six runs.
- This is lab evidence from GitHub-hosted Chromium, not field Core Web Vitals. Do not claim that this closes normal-network, physical-device or screen-reader QA.
- Keep 0.2.3-dev and M09.4 open pending owner evidence. Continue on main; leave the highly divergent legacy feature branch untouched.


## Latest continuation checkpoint — canonical Go module path migration (2026-11-03)

- Canonical repository: `https://github.com/pooya-labs/globalcart`; active branch: `main`. Commit [eaef837](https://github.com/pooya-labs/globalcart/commit/eaef837d6027c5c6085340f883567ecc764ec092) changed the Go module declaration and all 27 Go source/test files from the former owner path to `github.com/pooya-labs/globalcart` (63 import/declaration replacements). No business logic, database schema or public route changed.
- Backend CI [41630685778](https://github.com/pooya-labs/globalcart/actions/runs/41630685778) completed successfully: sqlc generation, module tidy/format, govulncheck, vet, full race-enabled Go suite, PostgreSQL backup/restore state comparison, API/importer builds and Docker image build.
- Security scans [41630685755](https://github.com/pooya-labs/globalcart/actions/runs/41630685755) passed. Pages deployment and published homepage/asset/product/case-study smoke [41630685929](https://github.com/pooya-labs/globalcart/actions/runs/41630685929) passed.
- Latest completed public Lighthouse audit remains [41627559678](https://github.com/pooya-labs/globalcart/actions/runs/41627559678): mobile Performance home 88 (LCP 3.2 s / TBT 200 ms), product 94 (3.1 s / 50 ms), case study 99 (0.9 s / 110 ms); all desktop Performance scores 100, and all six Accessibility/Best Practices/SEO scores are 100. Lab-only, not field data. Newer audit [41630849560](https://github.com/pooya-labs/globalcart/actions/runs/41630849560) was still in progress when this checkpoint was recorded; don't treat it as passed until it completes.
- Latest full Browser E2E remains [41622800540](https://github.com/pooya-labs/globalcart/actions/runs/41622800540), 74/74 across Chromium desktop/mobile, Firefox and WebKit. The module-path change is backend-only.
- Keep release version `0.2.3-dev`. M09.4 still requires physical desktop/mobile, real platform screen-reader, full manual RTL/LTR and normal-network/eligible field evidence. Do not tag 1.0.0 or mark these external gates complete based on CI.


## Latest published Lighthouse checkpoint — run 41630849560 (2026-11-03)

- Public Lighthouse [41630849560](https://github.com/pooya-labs/globalcart/actions/runs/41630849560) completed successfully against the canonical Pages host. Mobile Performance: storefront 89 / LCP 3.2 s / TBT 170 ms, product 93 / 3.1 s / 70 ms, case study 98 / 0.9 s / 150 ms. Desktop Performance is 100 on all three routes; Accessibility, Best Practices and SEO are 100 across all six audits.
- These are GitHub-hosted Lighthouse 13.6.0 lab results, not real-user field Core Web Vitals. They do not close the physical-device, real screen-reader, full RTL/LTR or field-delivery requirements in M09.4.
- The Go module-path migration [eaef837](https://github.com/pooya-labs/globalcart/commit/eaef837d6027c5c6085340f883567ecc764ec092) has passed Backend CI [41630685778](https://github.com/pooya-labs/globalcart/actions/runs/41630685778), Security [41630685755](https://github.com/pooya-labs/globalcart/actions/runs/41630685755) and Pages deployment/public smoke [41630685929](https://github.com/pooya-labs/globalcart/actions/runs/41630685929).
- Keep version `0.2.3-dev`; do not create a 1.0.0 tag until M09.4 owner evidence is recorded.
