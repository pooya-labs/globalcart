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


## Current continuation checkpoint — 2026-11-08 (v1.0.0 stable release published)

- Canonical source of truth: `pooya-labs/globalcart`, branch `main`. The stable [GitHub Release v1.0.0](https://github.com/pooya-labs/globalcart/releases/tag/v1.0.0) is published; tag target is [`dfe9a98`](https://github.com/pooya-labs/globalcart/commit/dfe9a98c70a9bbde3f5444c6f2cad2abd74269ac).
- Official release workflow [42494416761](https://github.com/pooya-labs/globalcart/actions/runs/42494416761), attempt 2, passed all release gates: manifest, TypeScript, unit tests, production static build, full cross-browser Playwright/axe (**151 passed / 9 intentional skips / 0 failed**), historical Gitleaks (**no leaks found**), public Pages smoke and final target comparison. The duplicate post-Lighthouse run [42495711540](https://github.com/pooya-labs/globalcart/actions/runs/42495711540) also succeeded and left the already-published release unchanged.
- Latest six-report published Lighthouse gate [42498127131](https://github.com/pooya-labs/globalcart/actions/runs/42498127131) passed unchanged budgets on retry 2: mobile Performance home 95 / product 96 / case study 99; desktop 100 on all routes; home mobile LCP 2.5 s / TBT 200 ms; product mobile 2.5 s / 140 ms; case-study mobile 0.9 s / 90 ms; CLS 0; Accessibility / Best Practices / SEO all 100.
- The first home-mobile TBT sample was 353 ms; retry 2 measured 200 ms against the unchanged 300 ms budget. The successful raw report identifies a 154 ms long task in shared Next.js chunk `4bd1b696-c023c6e3521b1417.js`, and an estimated 21.8 KiB unused portion in chunk `493-fb878163f2eec184.js` (~60 ms potential savings). M11 is a post-release investigation backlog; don't relax budgets or remove working storefront behavior.
- Pages deployment [42495371492](https://github.com/pooya-labs/globalcart/actions/runs/42495371492), live desktop/mobile image QA [42495556297](https://github.com/pooya-labs/globalcart/actions/runs/42495556297), Shopify Worker smoke [42495556252](https://github.com/pooya-labs/globalcart/actions/runs/42495556252) and Security [42495371493](https://github.com/pooya-labs/globalcart/actions/runs/42495371493) passed.
- Security exception: `.gitleaksignore` contains exactly one fingerprint for the historical `generic-api-key` false positive on the documented PostgreSQL identifier `carts.owner_user_id` at `PROJECT_CONTEXT.md:775` in commit `2ebdc241842a6366482ba6865ea7f748ab0eedb7`. Keep this exact suppression only; never broaden it to a path or rule. The release Gitleaks scan reported no leaks.
- Catalog and price rules remain unchanged: 1,000-item offline-first ABO snapshot; deterministic illustrative estimates retain `priceKind: "estimated"` and `priceAvailable: false`; no verified seller prices, stock or ratings; TGJU is only an indicative USD-to-Toman conversion/freshness source; translation stays paused.
- Manual external QA remains open: physical desktop/phone, actual platform screen-reader, complete human RTL/LTR and eligible normal-network/CDN field Core Web Vitals. Automated browser/Lighthouse lab results must not be presented as physical-device or field validation.
- Do not redo the completed gallery, catalog-price provenance, M06 backend foundations or v1 release. Continue from M11: reduce homepage mobile long-task work while maintaining the existing budget, and keep physical-device/screen-reader/owner RTL-LTR checks explicitly open.


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


## Latest continuation checkpoint — canonical Go-module guard and published Lighthouse (2026-11-03)

- Canonical repository/branch: `pooya-labs/globalcart` / `main`. Go module migration [eaef837](https://github.com/pooya-labs/globalcart/commit/eaef837d6027c5c6085340f883567ecc764ec092) changed the module and all 27 Go source/test files to `github.com/pooya-labs/globalcart`; Backend CI [41630685778](https://github.com/pooya-labs/globalcart/actions/runs/41630685778) passed module tidy, sqlc generation, govulncheck, vet, race tests, PostgreSQL backup/restore comparison and container builds. A new Backend CI guard now asserts the canonical module declaration and rejects the former module path in backend Go sources.
- Security [41631491938](https://github.com/pooya-labs/globalcart/actions/runs/41631491938) and Pages build/deploy/public smoke [41631491927](https://github.com/pooya-labs/globalcart/actions/runs/41631491927) passed on documentation checkpoint `b8fb5be`.
- Latest published Lighthouse [41631496522](https://github.com/pooya-labs/globalcart/actions/runs/41631496522) passed all six audits: mobile Performance home 91 / LCP 3.2 s / TBT 180 ms, product 94 / 2.9 s / 70 ms, case study 99 / 0.9 s / 120 ms; desktop Performance 100 across all three routes. Accessibility, Best Practices and SEO are 100 in all six. GitHub-runner lab only, not field Core Web Vitals.
- Latest full Browser E2E remains 74/74 on Chromium desktop/mobile, Firefox and WebKit ([41622800540](https://github.com/pooya-labs/globalcart/actions/runs/41622800540)); backend module-path change is server-side only.
- M09.4 owner evidence remains open: physical desktop/mobile, actual screen-reader, complete RTL/LTR and normal-network/eligible field measurements. Keep `0.2.3-dev`; do not tag `1.0.0` or represent CI emulation as owner QA.


## Latest continuation checkpoint — uncached PostgreSQL CI verification (2026-11-03)

- Backend CI now runs `go test -race -p 1 -count=1 ./...` so database-backed integration tests cannot be satisfied by Go's test-result cache. The first post-guard run [41632176463](https://github.com/pooya-labs/globalcart/actions/runs/41632176463) exposed the issue: cached integration tests had not created `schema_migrations`, so the backup check correctly failed on an empty database. The workflow was corrected and the fresh uncached run [41632402711](https://github.com/pooya-labs/globalcart/actions/runs/41632402711) passed in full: canonical module guard, sqlc generation, tidy/format, govulncheck, vet, race-enabled tests, PostgreSQL logical backup/restore comparison, API/importer builds and Docker image build.
- Security [41632402705](https://github.com/pooya-labs/globalcart/actions/runs/41632402705) and Pages deployment/public smoke [41632402717](https://github.com/pooya-labs/globalcart/actions/runs/41632402717) passed.
- Latest completed published Lighthouse at this checkpoint: [41631664623](https://github.com/pooya-labs/globalcart/actions/runs/41631664623). Mobile Performance: storefront 92 / LCP 3.1 s / TBT 170 ms; product 92 / 3.0 s / 70 ms; case study 98 / 0.9 s / 180 ms. Desktop Performance is 100 on all routes; Accessibility, Best Practices and SEO are 100 in all six modes. GitHub-runner lab only, not field data.
- Canonical Go module guard remains active; no former-owner imports remain in backend Go sources. Continue on `main`. Do not repeat completed module migration, SQL generation, DB ownership tests or previous deployment work.
- The only 1.0.0 release gate remains M09.4 owner evidence: real desktop/mobile, real screen-reader, full manual RTL/LTR and normal-network/eligible field metrics. Keep `0.2.3-dev`; do not create a 1.0.0 tag or mark external QA complete based on CI.



## Latest continuation checkpoint — public Lighthouse 41633195411 and mainline state (2026-11-03)

- Canonical repository: `pooya-labs/globalcart`; canonical Pages: `https://pooya-labs.github.io/globalcart/`; continue on `main` only. The latest inspected main head before this documentation checkpoint was `e5a22ca54dd8a8659462560e1b11b81ea4e4ec58`; its change is generated catalog timestamps plus the committed npm lockfile, not application logic.
- Latest public Lighthouse [41633195411](https://github.com/pooya-labs/globalcart/actions/runs/41633195411) completed all six desktop/mobile route audits successfully. Mobile Performance: storefront 94 (LCP 2.8 s / TBT 140 ms), product 96 (2.6 s / 80 ms), case study 99 (0.9 s / 120 ms). Desktop Performance 100 on all routes; Accessibility, Best Practices and SEO 100 in all six; CLS 0 throughout. These are GitHub-hosted lab readings, not field Core Web Vitals. Raw JSON artifact is attached to the run.
- Latest verified backend gate remains [Backend CI 41632402711](https://github.com/pooya-labs/globalcart/actions/runs/41632402711): canonical module guard, uncached race-enabled Go tests, PostgreSQL backup/restore state comparison and API/importer/Docker builds all passed. Security [41632825747](https://github.com/pooya-labs/globalcart/actions/runs/41632825747) and Pages deployment/public smoke [41632825984](https://github.com/pooya-labs/globalcart/actions/runs/41632825984) passed on the latest human-authored documentation checkpoint; scheduled Pages smoke [41632904986](https://github.com/pooya-labs/globalcart/actions/runs/41632904986) also passed.
- Latest full Browser E2E remains [41622800540](https://github.com/pooya-labs/globalcart/actions/runs/41622800540), 74/74 across Chromium desktop/mobile, Firefox and WebKit. Later mainline changes are documentation, CI guard/test-invocation maintenance and generated catalog/lockfile snapshots; don't claim that the latest head was re-run through the entire browser matrix.
- No open PRs or issues were found in the repository at this checkpoint. Leave the highly divergent `feat/m04-go-postgres-foundation` branch untouched.
- M09.4 is still blocked on actual owner evidence: physical desktop and phone, real platform screen-reader, comprehensive manual Persian RTL/English LTR, and normal-network/published delivery plus eligible field metrics. Automation cannot supply these observations. Keep `0.2.3-dev`; do not tag `1.0.0`.
- Do not repeat M06.1–M06.6, M09.1–M09.3, TGJU parsing/publication relay, ERD, screenshot publishing or completed app features. Translation remains paused; production database/API, public seller/admin routes, real payments and fulfillment remain disabled.




## Latest continuation checkpoint — 76/76 Browser E2E and canonical Pages verification (2026-11-03)

- Canonical repo/site: `pooya-labs/globalcart` / `https://pooya-labs.github.io/globalcart/`; continue on `main`. At the start of this checkpoint main was `05e6de8d915b03a4803ba266604be71619984e0a`; its only difference from the tested parent `2df4c724fb25a50ea82ba10e6b9394c898e21fe2` is the bot-generated `data/catalog.json` snapshot. No application code changed after the browser test.
- Added `e2e/locale-regression.spec.ts` in [commit 2df4c72](https://github.com/pooya-labs/globalcart/commit/2df4c724fb25a50ea82ba10e6b9394c898e21fe2). It verifies Persian RTL and English LTR preference persistence across storefront, product, recently-viewed and case-study route transitions, reloads, and document horizontal-overflow checks.
- [Browser E2E 41635693445](https://github.com/pooya-labs/globalcart/actions/runs/41635693445) passed **76/76** in 2.4 minutes: Chromium desktop/mobile, Firefox accessibility and WebKit accessibility. The newly added test passed in Chromium desktop and mobile. TypeScript, unit tests and static production build also passed before Playwright execution; the HTML report and screenshots are retained in the workflow artifact.
- Security [41635693431](https://github.com/pooya-labs/globalcart/actions/runs/41635693431) passed. Pages build/deploy and public smoke for homepage, Next.js assets, product deep link and bilingual case study [41635694057](https://github.com/pooya-labs/globalcart/actions/runs/41635694057) passed.
- Latest public Lighthouse [41635876534](https://github.com/pooya-labs/globalcart/actions/runs/41635876534) passed all six audits. Mobile Performance: home 91 / LCP 3.3 s / TBT 90 ms; product 97 / 2.6 s / 60 ms; case study 99 / 1.0 s / 110 ms. Desktop Performance 100 on all three routes; Accessibility, Best Practices and SEO 100 across all six; CLS 0 throughout. These are GitHub-hosted simulated lab results, not field Core Web Vitals.
- Latest backend verification remains [Backend CI 41632402711](https://github.com/pooya-labs/globalcart/actions/runs/41632402711): canonical Go module guard, uncached race-enabled Go suite, PostgreSQL logical backup/restore comparison and API/importer/container builds all passed. No backend source changes followed that run.
- No open PRs or issues were found at this checkpoint. Leave the highly divergent legacy `feat/m04-go-postgres-foundation` branch untouched. Do not repeat completed backend foundation, TGJU publication relay, ERD or portfolio screenshot work.
- M09.4 is not closed: real physical desktop/phone inspection, actual platform screen-reader use, comprehensive human RTL/LTR review and normal-network/field-delivery evidence still require real-world owner observations. CI and simulated Lighthouse cannot supply these. Keep `0.2.3-dev`; do not tag `1.0.0`.
- Translation remains paused. Production database/API hosting, public seller/admin routes, real payment and fulfillment remain disabled.




## Latest continuation checkpoint — Lighthouse regression budgets and pre-release redesign hold (2026-11-03)

- Canonical repo/site: pooya-labs/globalcart / https://pooya-labs.github.io/globalcart/; continue only on main. At this checkpoint, main was 9dc7efbe74eefb560fe92e538817af9e4635b810; the latest scheduled catalog snapshot commit only refreshed data/catalog.json.
- Added scripts/check-lighthouse-budget.ts, unit coverage in tests/lighthouse-budget.test.ts, and an enforcement step in .github/workflows/lighthouse-public.yml in [commit e7e5008](https://github.com/pooya-labs/globalcart/commit/e7e5008ea3ff4fca5894866968797bae3888a993). The evaluator requires all six public route/mode reports; mobile LCP ≤4.0 s / TBT ≤300 ms, desktop LCP ≤2.5 s / TBT ≤200 ms, CLS ≤0.10, and Accessibility/Best Practices/SEO ≥95. Unit tests cover accepted limits, violations, missing data and incomplete report matrices.
- Catalog CI [41639557537](https://github.com/pooya-labs/globalcart/actions/runs/41639557537) passed dependency audit, TypeScript, catalog/domain unit tests (including the new budget tests) and production build. Security [41639557550](https://github.com/pooya-labs/globalcart/actions/runs/41639557550) and Pages build/deploy/public smoke [41639557674](https://github.com/pooya-labs/globalcart/actions/runs/41639557674) passed.
- Latest public Lighthouse [41639799398](https://github.com/pooya-labs/globalcart/actions/runs/41639799398) passed all six audits and the new budget enforcement step. Mobile Performance: home 88 / LCP 3.5 s / TBT 80 ms; product 96 / 2.6 s / 40 ms; case study 100 / 1.2 s / 60 ms. Desktop Performance: 100 / 99 / 100; Accessibility, Best Practices and SEO are 100 across all six; max CLS 0.041. These are simulated GitHub-hosted lab metrics only, not field Core Web Vitals.
- Latest full application Browser E2E is still [41635693445](https://github.com/pooya-labs/globalcart/actions/runs/41635693445), 76/76 across Chromium desktop/mobile, Firefox and WebKit. Commits after that E2E run changed docs, CI guard/tests or generated catalog metadata; no application UI code changed. Do not claim that the latest repository head has been run through all E2E scenarios.
- Backend CI latest successful verification remains [41632402711](https://github.com/pooya-labs/globalcart/actions/runs/41632402711); Go/PostgreSQL sources were not changed after that check. Leave the highly divergent feat/m04-go-postgres-foundation branch untouched; no open PRs/issues were found at the previous repository check.
- OWNER-REQUESTED PRE-RELEASE HOLD: immediately before any final product publication or 1.0.0 tag, stop and notify the owner. First coordinate and complete a new redesign of the storefront shell/visual skin, then rerun visual, responsive, RTL/LTR, accessibility and performance checks and proceed only after owner approval. Routine CI/Pages preview deploys are not the final product release.
- Keep version 0.2.3-dev. M09.4 still requires physical desktop/mobile checks, actual platform screen-reader use, comprehensive human RTL/LTR review and normal-network/eligible field evidence. Translation remains paused; public API/database hosting, seller/admin routes and real payments/fulfillment remain disabled.




## Latest continuation checkpoint — stale TGJU UI regression, 78/78 browser suite (2026-11-03)

- Canonical repo/site: pooya-labs/globalcart / https://pooya-labs.github.io/globalcart/; continue on main. Main at checkpoint: 153cb93314b15178d53670cd9d4d06be7e09a23f. The tested implementation commit was 17bd8c6; the subsequent bot commit only refreshed data/catalog.json. No application UI code changed.
- Added a browser regression to e2e/storefront.spec.ts: compute expected FX freshness from the committed snapshot, assert the storefront data-status/label and the product price-breakdown warning; when TGJU provides a date-only marker it must be Stale. Corrected one erroneous test expectation from Toman / USDT to Toman / USD. Commit [17bd8c6](https://github.com/pooya-labs/globalcart/commit/17bd8c6bc624c44c17038891af72245d02e3b6ab).
- The first push had an invalidly escaped URL regex; its TypeScript validation failed in [run 41641812126](https://github.com/pooya-labs/globalcart/actions/runs/41641812126). Corrected the regex in a follow-up commit. Do not count the failed attempt as a pass.
- Corrected Browser E2E [41642120165](https://github.com/pooya-labs/globalcart/actions/runs/41642120165) passed **78/78**. The new stale-TGJU test passed in Chromium desktop (1.1 s) and mobile (992 ms). TypeScript, unit tests, static Pages build, full Chromium/Firefox/WebKit journeys and Playwright report upload all passed.
- Security [41642120200](https://github.com/pooya-labs/globalcart/actions/runs/41642120200) and Pages deployment/public smoke [41642120270](https://github.com/pooya-labs/globalcart/actions/runs/41642120270) passed. Portfolio screenshots [41642510070](https://github.com/pooya-labs/globalcart/actions/runs/41642510070) and public Lighthouse [41642306099](https://github.com/pooya-labs/globalcart/actions/runs/41642306099) also passed.
- Latest Lighthouse [41642306099](https://github.com/pooya-labs/globalcart/actions/runs/41642306099): mobile Performance 89 / 97 / 99 and desktop 99 / 100 / 100 (home / product / case study); Accessibility, Best Practices and SEO 100 in all six; CLS 0 throughout. The six-report LCP/TBT/CLS/score budget gate passed. Metrics are simulated GitHub-hosted lab readings, not field CWV.
- The six-report budget guard remains in scripts/check-lighthouse-budget.ts and .github/workflows/lighthouse-public.yml: mobile LCP ≤4.0 s / TBT ≤300 ms, desktop LCP ≤2.5 s / TBT ≤200 ms, CLS ≤0.10 and Accessibility/Best Practices/SEO ≥95. Seven evaluator/report-matrix unit tests passed in Catalog CI [41639557537](https://github.com/pooya-labs/globalcart/actions/runs/41639557537).
- Keep 0.2.3-dev. M09.4 physical desktop/mobile, real platform screen-reader, comprehensive human RTL/LTR and normal-network/field evidence remain open. Owner-requested pre-release hold: before any final publication/1.0.0 tag, stop and notify the owner to coordinate and complete the new storefront-shell redesign; obtain owner approval before release. Routine preview deployments are not a final release.




## Latest continuation checkpoint — tablet browser coverage and pre-release approval (2026-11-03)

- Canonical repo/site: pooya-labs/globalcart / https://pooya-labs.github.io/globalcart/; default branch main. At this documentation checkpoint main was 6ab898da23bc3bd3200b8aa50b4fbc499e5d0c41; after the tested commit, scheduled sync only refreshed data/catalog.json.
- Added a dedicated Chromium tablet project using iPad (7th generation) and ran the accessibility/performance suites on it. The tablet-specific regression checked the storefront, product detail and case-study routes at tablet width and asserted no horizontal document/body overflow. Browser E2E [41644963684](https://github.com/pooya-labs/globalcart/actions/runs/41644963684) passed 97 tests; the tablet overflow test and tablet axe coverage passed. Two tablet-only test skips on non-tablet projects are intentional.
- The same full E2E run had one WebKit homepage axe test exceed the old 30 s timeout on its first attempt (33.6 s), then pass on retry (12.2 s). This was a test-runner timeout, not an axe finding; page snapshot was fully rendered and the RTL/LTR audit eventually passed. Increased that test's limit to 60 s in [commit 81f447a](https://github.com/pooya-labs/globalcart/commit/81f447aaf9b07ca0a9c8a94bc05c3e6346943ec5), with all assertions unchanged. Re-run [41646243770](https://github.com/pooya-labs/globalcart/actions/runs/41646243770) was still at Playwright Chromium/Firefox/WebKit installation at this checkpoint, so the new timeout change had not yet been reverified in a completed E2E run.
- Pages deployment and public Smoke Test [41644963414](https://github.com/pooya-labs/globalcart/actions/runs/41644963414) passed; Security [41644963653](https://github.com/pooya-labs/globalcart/actions/runs/41644963653) passed.
- Latest public Lighthouse [41646444864](https://github.com/pooya-labs/globalcart/actions/runs/41646444864) passed all six route/mode audits and the lab-budget gate: mobile Performance 92 / 97 / 100 and desktop 100 / 100 / 100 (home / product / case study); Accessibility, Best Practices and SEO 100 in all six; CLS 0. An intervening run [41645326322](https://github.com/pooya-labs/globalcart/actions/runs/41645326322) had home-mobile LCP 4.196 s and correctly failed the 4.0 s budget; the next run measured 3.3 s and passed. Its breakdown showed ~2.10 s hero element render delay, with image transfer ~94 ms and TTFB ~108 ms; the hero request was already high-priority, eager and discoverable. Do not relax the budget just to mask the outlier; revisit above-the-fold rendering in the requested shell redesign.
- Automated QA that can be executed without physical hardware is approved: 97 passing browser tests in the completed run (including tablet), keyboard/focus-loop checks, axe coverage, route/locale persistence, FX stale disclosure, published Pages smoke, security and six-route Lighthouse budgets. This approval does not claim physical-device or real screen-reader testing.
- Remaining release gates: actual physical desktop/mobile verification; actual platform screen-reader review; human whole-site RTL/LTR pass; normal-network/CDN and eligible field-data review; then post-redesign regression and package/release preparation. Do not tag 1.0.0 before the requested shell redesign and owner approval. Keep version 0.2.3-dev.




## Latest continuation checkpoint — compact-phone and tablet-orientation QA (2026-11-03)

- Canonical repo/site: pooya-labs/globalcart / https://pooya-labs.github.io/globalcart/; continue only on main. Implementation test commit [a834889](https://github.com/pooya-labs/globalcart/commit/a834889d9d4663650c5baed52264f10136bd364d); subsequent bot snapshot only refreshed data/catalog.json. Main at documentation checkpoint: 912843c423ff28f4e76e66ca14c218ac73974da1.
- Added compact-phone width coverage at 320 px and 360 px in e2e/performance.spec.ts. It checks the storefront, product detail and case-study routes for document/body overflow on the mobile project.
- Expanded the iPad 7th-generation tablet route-width test to run portrait 810×1080 and landscape 1080×810, checking the storefront, product detail and case study in both orientations.
- Browser E2E [41652758312](https://github.com/pooya-labs/globalcart/actions/runs/41652758312) passed **99 tests / 4 intentional skips**, with no failed/flaky/retried test in the final summary. The compact phone test passed on the mobile project; tablet portrait/landscape route overflow passed on the tablet project. Two tablet-only checks skip on desktop/mobile projects and the compact-phone check skips on non-mobile projects by design.
- Pages deployment/public Smoke Test [41652758510](https://github.com/pooya-labs/globalcart/actions/runs/41652758510) and Security [41652758613](https://github.com/pooya-labs/globalcart/actions/runs/41652758613) passed.
- Latest Lighthouse [41652966386](https://github.com/pooya-labs/globalcart/actions/runs/41652966386) passed all six published reports and budgets: mobile Performance 91 / 94 / 99, desktop 100 / 100 / 100 (home / product / case study); all Accessibility, Best Practices and SEO scores 100; CLS 0 throughout. Mobile LCP 3.3 s / 3.1 s / 1.0 s and TBT 70 ms / 60 ms / 100 ms. GitHub-hosted lab only, not field Core Web Vitals.
- Automated keyboard/focus coverage includes logical Tab/Shift+Tab loops, Escape close, focus containment/restoration for command palette, cart/favorites drawers, quick view, compare and checkout. Cross-browser run [41646243770](https://github.com/pooya-labs/globalcart/actions/runs/41646243770) completed 98 passed / 4 intentional skips; the WebKit home dual-locale axe check passed in 24.8 s within its 60 s allowance, no retry needed. This is automated keyboard event/focus testing, not an actual screen-reader session.
- All remotely executable pre-release software QA is now green. Remaining release work is physical-device verification, actual platform screen-reader review, comprehensive human RTL/LTR inspection, normal-network/eligible field evidence, and the requested shell redesign followed by regression and owner approval. Keep 0.2.3-dev; do not tag 1.0.0 before the redesign and approval.



## Latest continuation checkpoint — M09.6 catalog galleries (2026-11-04)
- Work only in `pooya-labs/globalcart`, branch `main`; keep version `0.2.3-dev`.
- Previous import [run 41866314414](https://github.com/pooya-labs/globalcart/actions/runs/41866314414) reached the 320 MiB budget after 550 of 1,000 product galleries and preserved the prior snapshot. Do not assume full galleries have shipped.
- Current app code [ccbb48a](https://github.com/pooya-labs/globalcart/commit/ccbb48a6538ddd10db4ba118d1fc06adbcc2b039): WebP main 1,100 px/q72, fallback 960 px/q64, 650 KiB per image, 320 MiB aggregate. No gallery count truncation.
- Main product-detail cart now persists the merged cart; main image gets eager/high priority; other images remain lazy. Removed old external image preconnects. E2E data fixtures use the actual ABO catalog and test cart persistence. Case study correctly states the CC BY 4.0 ABO source and missing retail commercial facts.
- Security [41874481525](https://github.com/pooya-labs/globalcart/actions/runs/41874481525) passed. Browser E2E [41874481635](https://github.com/pooya-labs/globalcart/actions/runs/41874481635) passed TypeScript, unit tests and build, with desktop/mobile journeys still running at last check. Pages [41873827996](https://github.com/pooya-labs/globalcart/actions/runs/41873827996) remained in catalog sync; [41874481571](https://github.com/pooya-labs/globalcart/actions/runs/41874481571) was queued.
- Next: read final Playwright test results; inspect catalog import totals/files; verify public Pages gallery smoke. Mark M09.6 complete only after successful logs. Never claim queued/in-progress workflow as a pass. Don't tag 1.0.0 or claim manual physical-device, real screen-reader, human RTL/LTR or field Core Web Vitals QA.
