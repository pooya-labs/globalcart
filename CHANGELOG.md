# Changelog

## Unreleased — 2026-11-03

### Verified — latest six-route Lighthouse on canonical Pages

- Lighthouse [41620637413](https://github.com/pooya-labs/globalcart/actions/runs/41620637413) passed all six audits on the renamed Pages host. Mobile Performance: home 93 / LCP 3.2 s / TBT 80 ms, product 95 / LCP 2.9 s / TBT 50 ms, case study 99 / LCP 0.9 s / TBT 100 ms. Desktop Performance is 100 across all routes; Accessibility, Best Practices and SEO are 100 in all six modes.
- Raw diagnostics identify the homepage hero and product main image as LCP elements. Both are immediately discoverable, eager and high-priority; the homepage hero's request-load delay is ~11 ms and resource-load duration ~238 ms in this run. No extra preload or unsupported image URL transformation was added.
- These are repeatable GitHub-hosted lab results, not field Core Web Vitals. Manual physical-device, platform screen-reader, whole-site RTL/LTR and normal-network release QA remains open.


### Verified — account rename, TGJU changed-quote relay and new Pages host

- Latest post-documentation Lighthouse [41619642994](https://github.com/pooya-labs/globalcart/actions/runs/41619642994) succeeded on the new host: mobile Performance home 94 (LCP 3.0 s / TBT 90 ms), product 99 (2.3 s / 60 ms), case study 100 (0.9 s / 90 ms); desktop Performance 100 on all three routes. All six Accessibility, Best Practices and SEO scores are 100. Lab only.

- A subsequent TGJU refresh read 271,500 Toman/USD from the page's date-only `11 آبان` marker. Sync [41618464249](https://github.com/pooya-labs/globalcart/actions/runs/41618464249) persisted the validated snapshot and Pages event [41618569410](https://github.com/pooya-labs/globalcart/actions/runs/41618569410) passed. Because the source supplied no exact clock, the committed quote is intentionally stale, not live.

- The canonical GitHub owner changed to pooya-labs. Public Pages now resolves to https://pooya-labs.github.io/globalcart/; repository-dispatch Pages run [41616698725](https://github.com/pooya-labs/globalcart/actions/runs/41616698725) passed build, deployment and public homepage/assets/product/case-study smoke.
- A genuine TGJU source update changed the quote from 271,500 to 274,500 Toman/USD (2,745,000 Rial; Tehran clock 12:20:14). Sync [41616645902](https://github.com/pooya-labs/globalcart/actions/runs/41616645902) persisted [snapshot commit 624eeb9](https://github.com/pooya-labs/globalcart/commit/624eeb994dbf21971fdec0d396d374af9364483c) and requested repository_dispatch; the event-driven Pages run completed successfully. Its source timestamp is stale under the 15-minute policy.
- Corrected the public Lighthouse base URL after the old account-hosted Pages domain began returning 404. The workflow now tests the canonical Pages host.
- Migrated remaining first-party workflow actions to Node 24-compatible major versions; this is tooling-only and does not change the application runtime (Node.js 22).
- M09.4 manual desktop/mobile, screen-reader, comprehensive RTL/LTR and normal-network/field-delivery evidence remains open. Version stays 0.2.3-dev.

### Verified — latest published Lighthouse and API guard checkpoint

- Public Lighthouse [41613950109](https://github.com/pooya-labs/globalcart/actions/runs/41613950109) completed all six route/mode audits: mobile Performance 92 (home), 97 (product) and 100 (case study); desktop Performance 100 on all three. Accessibility, Best Practices and SEO remain 100 in every mode.
- Raw image diagnostics estimate savings of 160/108 KiB on homepage desktop/mobile and 82/84 KiB on product detail. These are mostly oversized catalog thumbnails and alternate gallery images. No supported transformation contract for existing DummyJSON product CDN URLs was established, so no speculative URL rewriting or third-party proxy was added.
- Reconciled M09.4.1: its public API surface guard tests passed Backend CI [41353463006](https://github.com/pooya-labs/globalcart/actions/runs/41353463006); the roadmap heading now accurately marks this slice CI-verified. M09.4 manual physical-device, screen-reader, RTL/LTR and field-delivery requirements remain open.
- Updated canonical repository/profile links after the GitHub account username changed to pooya-labs. The currently configured Pages URL was retained because its public deployment smoke test still succeeds.


### Performance and delivery hardening — CI verified

- Scoped preconnect hints to the routes that actually request their origins: homepage (Unsplash + DummyJSON), product detail (DummyJSON only), and no external image preconnects on the case study. Lighthouse [41613166735](https://github.com/pooya-labs/globalcart/actions/runs/41613166735) confirms the corresponding mobile routes have no unused-preconnect warnings.
- Added a route-transition regression to Browser E2E. The final run [41612960428](https://github.com/pooya-labs/globalcart/actions/runs/41612960428) passed **70/70** desktop/mobile/cross-browser tests. Pages deployment and public smoke [41612960450](https://github.com/pooya-labs/globalcart/actions/runs/41612960450) and Security [41612960405](https://github.com/pooya-labs/globalcart/actions/runs/41612960405) also passed.
- Added a validated-snapshot publication relay: on a changed TGJU quote, the FX workflow requests `globalcart-fx-updated` via `repository_dispatch`, and Pages rebuilds from the new committed snapshot. The current refresh was unchanged, so it correctly did not dispatch. First changed-quote event verification remains pending; no value or timestamp was fabricated.
- Latest six-route Lighthouse lab run [41613166735](https://github.com/pooya-labs/globalcart/actions/runs/41613166735): all Accessibility, Best Practices and SEO scores are 100; mobile Performance is 88 (home), 97 (product) and 98 (case study). Results are lab data, not field Core Web Vitals.


### Verified — latest published QA recheck
- Re-ran the six public Lighthouse audits in [run 41608446036](https://github.com/pooya-labs/globalcart/actions/runs/41608446036): mobile Performance 91 (storefront), 95 (product), 98 (case study); all six modes recorded 100 Accessibility, Best Practices and SEO. This supersedes the older figures as the latest recorded run without deleting the original baseline. The results are lab data, not field Core Web Vitals.
- Synchronized the README, roadmap, project context and AI handoff with the latest cross-browser (66/66), Pages, Security and Lighthouse results. Kept the 1.0.0 release gated by actual manual QA.
- Reconfirmed the saved TGJU quote is stale by its source timestamp; no stale quote is presented as current and no alternative provider or hard-coded FX fallback was added.


### Added
- Completed the first public Lighthouse baseline after targeted mobile and font optimizations. Latest run [41604368513](https://github.com/pooya-labs/globalcart/actions/runs/41604368513) measured storefront mobile Performance 87 / LCP 3.2 s / TBT 250 ms (initial baseline: 74 / 4.9 s / 320 ms); product mobile 96 and case-study mobile 99. All six lab audits recorded Accessibility 100 and SEO 100. Results are lab-only, not CrUX/field Core Web Vitals. Final 44/44 Browser E2E [41603301095](https://github.com/pooya-labs/globalcart/actions/runs/41603301095), Pages [41603301049](https://github.com/pooya-labs/globalcart/actions/runs/41603301049) and Security [41603300997](https://github.com/pooya-labs/globalcart/actions/runs/41603300997) passed.
- Added a post-deployment and weekly Lighthouse lab workflow for the public storefront, product and case-study pages in mobile and desktop modes. Lighthouse 13.6.0 is pinned; JSON reports are retained as a GitHub Actions artifact and a summary records category scores and lab metrics. Added an evidence-focused manual QA checklist. The workflow explicitly does not claim field Core Web Vitals or physical-device verification.
- Added `docs/PERFORMANCE_QA.md` and `scripts/summarize-lighthouse.mjs`; no third-party report server, API key or production secret was introduced.

### Added
- Added a bilingual `/case-study/` portfolio route with shared Persian/English locale preference, responsive feature tour, architecture/data-flow diagram, route metadata and a storefront footer entry. It documents the actual static Pages storefront, local Go/PostgreSQL foundation, TGJU source handling and demo-only commerce boundaries. Browser E2E [41577703335](https://github.com/pooya-labs/globalcart/actions/runs/41577703335), Security [41577703300](https://github.com/pooya-labs/globalcart/actions/runs/41577703300) and Pages deployment plus published smoke [41577703347](https://github.com/pooya-labs/globalcart/actions/runs/41577703347) passed; this is not a 1.0 release.
- Added full-page Persian case-study screenshots for desktop and mobile Chromium in `docs/screenshots/`. The capture test runs against the production static export, and the guarded publisher commits the artifact only after a green Browser E2E run. Screenshot publication [41597084953](https://github.com/pooya-labs/globalcart/actions/runs/41597084953) succeeded.
- Audited README claims against the actual implementation, separated completed search/recommendation work from external prerequisites, removed the unimplemented Python-worker claim and clarified the local-only backend boundary.
- Added `docs/DATABASE_ERD.md`, derived from the 15 PostgreSQL migrations and documenting all 27 tables, declared relationships and important integrity constraints in four Mermaid ERDs. Added a bilingual responsive database summary and ERD link to the case-study route, with desktop/mobile RTL/LTR accessibility assertions. Browser E2E [41596731945](https://github.com/pooya-labs/globalcart/actions/runs/41596731945) passed all 40 desktop/mobile tests, Security [41596731890](https://github.com/pooya-labs/globalcart/actions/runs/41596731890) passed, and Pages [41596731889](https://github.com/pooya-labs/globalcart/actions/runs/41596731889) passed the published case-study/ERD smoke test.

### Added
- Extended the image-loading/performance regression audit to the recently-viewed route. Product visits now have a Playwright desktop/mobile journey that verifies history rendering, the GitHub Pages base path, lazy loading, asynchronous decoding, accessible alt text and reserved image dimensions. Browser E2E [41575646034](https://github.com/pooya-labs/globalcart/actions/runs/41575646034) passed all 36 tests; Pages deployment/smoke [41575646035](https://github.com/pooya-labs/globalcart/actions/runs/41575646035) and Security [41575646222](https://github.com/pooya-labs/globalcart/actions/runs/41575646222) also passed on `c358970b31d2272db98434a761c699355a9dbc57`. The initial-route CI measurement was 520,131 JS bytes/7 requests and 70,540 CSS bytes/1 request.

### Fixed
- Extended the accessibility regression matrix to Firefox and WebKit desktop while retaining full Chromium desktop/mobile journeys. Browser E2E [41607322346](https://github.com/pooya-labs/globalcart/actions/runs/41607322346) passed 66/66 (48 Chromium scenarios + 9 Firefox + 9 WebKit accessibility scenarios); Pages [41607322372](https://github.com/pooya-labs/globalcart/actions/runs/41607322372) and Security [41607322262](https://github.com/pooya-labs/globalcart/actions/runs/41607322262) passed.
- Extended route-level axe/WCAG tests to the recently viewed page in populated and empty Persian RTL / English LTR states; fixed contrast issues and button accessible-name mismatches surfaced by the new test. Final Browser E2E [41606569768](https://github.com/pooya-labs/globalcart/actions/runs/41606569768) passed 48/48, Pages [41606569808](https://github.com/pooya-labs/globalcart/actions/runs/41606569808) and Security [41606569782](https://github.com/pooya-labs/globalcart/actions/runs/41606569782) passed.
- Preserved the `/globalcart` Pages base path in the new recently-viewed browser journey by deriving it from the product route instead of resolving a multi-level relative URL.

### Fixed
- Serialized Backend CI Go test packages with `go test -race -p 1 ./...` because PostgreSQL integration packages intentionally share `TEST_DATABASE_URL` and the checksum-drift regression temporarily alters the shared migration ledger; race detection remains enabled within each package.
- Fixed locale continuity across the homepage, product-detail and recently-viewed routes. The history page now reads and writes the shared guarded `globalcart-locale-v1` preference, and its language-toggle accessible name is localized correctly; unsupported values or unavailable storage still fall back safely to Persian. Browser E2E [41462200408](https://github.com/pooya-labs/globalcart/actions/runs/41462200408) passed all 34 desktop/mobile tests; Pages deployment and published-site smoke test [41462200496](https://github.com/pooya-labs/globalcart/actions/runs/41462200496) passed.
- Aligned the data-source provenance description with the active Wallex USDTTMN Toman candle-close feed; removed the obsolete Nobitex midpoint description.
- Moved the Wallex scheduled refresh from minute zero to `2-59/5` to avoid the busy hour boundary. Push-triggered refresh [41413175324](https://github.com/pooya-labs/globalcart/actions/runs/41413175324) passed, validated and persisted the 08:27 UTC candle at 272,892 Toman per USDT. A scheduled refresh has now also completed successfully in [Sync Wallex FX run 41558848340](https://github.com/pooya-labs/globalcart/actions/runs/41558848340) on 2026-11-03; recurring cadence reliability remains under observation.

### Changed
- TGJU now parses both live Tehran clock cells and date-only market markers. Date-only quotes retain the raw price but are always classified as stale; a time-only clock ahead of the current Tehran clock is anchored to its latest prior occurrence. Latest verified snapshot: 2,715,000 Rial (271,500 Toman per USD), marker `11 آبان` with no exact time, therefore shown as stale rather than fresh. Refresh [41568568791](https://github.com/pooya-labs/globalcart/actions/runs/41568568791), Security [41568713623](https://github.com/pooya-labs/globalcart/actions/runs/41568713623), Pages [41568713584](https://github.com/pooya-labs/globalcart/actions/runs/41568713584), and desktop/mobile Browser E2E [41568713639](https://github.com/pooya-labs/globalcart/actions/runs/41568713639) passed. Five-minute scheduled cadence remains under observation.
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Implemented and CI-verified internal M06.6 seller workspace: vendor-scoped profile read, offer and order-state aggregates, last-30-day order count and per-currency paid-simulated totals; display-name-only profile edits use optimistic concurrency and append an audit event. No customer details or real-revenue claims are exposed. Paid totals now group only orders in `paid_simulated`. Backend CI [41572747044](https://github.com/pooya-labs/globalcart/actions/runs/41572747044), Security [41572747081](https://github.com/pooya-labs/globalcart/actions/runs/41572747081) and Pages [41572746914](https://github.com/pooya-labs/globalcart/actions/runs/41572746914) passed on `fe24119685bfa8487cb8fa2abfb8ad85ed430904`.
- Implemented and CI-verified internal M06.5 seller offer create/edit/archive: draft-only creation, revision-checked SKU/price/stock edits, pause and terminal archive. Migration 015 advances revisions on all offer updates and protects append-only snapshots; price/inventory history and actor-attributed audit events are transactional. Backend CI [41571374758](https://github.com/pooya-labs/globalcart/actions/runs/41571374758), Security [41571374594](https://github.com/pooya-labs/globalcart/actions/runs/41571374594) and Pages [41571374873](https://github.com/pooya-labs/globalcart/actions/runs/41571374873) passed on `3a9a328fd7d7b1262d73dfefbc2e7a81fa994782`.
- Added the internal M06.4 seller product work queue with transaction-bound authorization, current account/session/role/vendor/membership revalidation, effective catalog text, source provenance, bounded search/filter/sort/pagination and seller-owned offer values. PostgreSQL tests cover customer/foreign-vendor denial, revoked memberships, filtering, paging and hostile input. Backend CI [41564104334](https://github.com/pooya-labs/globalcart/actions/runs/41564104334), Security scans [41564104323](https://github.com/pooya-labs/globalcart/actions/runs/41564104323) and Pages deployment/smoke test [41564104379](https://github.com/pooya-labs/globalcart/actions/runs/41564104379) passed. No schema migration, seller HTTP route, dashboard or hosted service was added.\n- Added the internal M06.3 administrator product work queue: bounded search, active/archived/all status, source/category/effective-brand filters, allow-listed sorting and pagination. The filter validator rejects invalid UTF-8, PostgreSQL-incompatible NUL characters and overlong values. Final Backend CI [41562708698](https://github.com/pooya-labs/globalcart/actions/runs/41562708698), Security scans [41562708607](https://github.com/pooya-labs/globalcart/actions/runs/41562708607) and Pages deployment [41562708709](https://github.com/pooya-labs/globalcart/actions/runs/41562708709) passed, including PostgreSQL integration/race tests, backup/restore, API/importer and Docker builds, and published-site smoke checks. No HTTP route or dashboard is exposed.
- Added internal M06.2 manual product creation, editorial editing, archive/restore and revision-history services with migrations 013–014. Manual products receive IDs from a reserved sequence, `admin_manual` provenance and a GlobalCart Direct demo offer with opening price/stock history in one transaction.
- Kept admin title/description/brand overrides separate from supplier-owned fields; explicit clearing returns the current supplier value, while re-imports cannot clobber editorial work. Catalog list/detail/search/brand filters/counts apply effective values consistently and exclude archived products.
- Required an expected revision for edit/archive/restore to reject stale writes. Append-only before/after snapshots preserve effective and source values; matching audit events are recorded transactionally. Bounded history reads are admin-only.
- Added PostgreSQL integration tests for manual creation/offer creation, input validation, admin-only authorization, supplier refresh preservation, revision conflicts, archive/restore visibility and immutable history. Backend CI [41473127094](https://github.com/pooya-labs/globalcart/actions/runs/41473127094) and Security scans [41473127072](https://github.com/pooya-labs/globalcart/actions/runs/41473127072) passed on implementation head `e288cba2f0abe891656388d6e50638bff5fadc3c`. These remain internal services; no public admin route or UI is exposed.
- Added an internal, read-only M06 administrator overview service. It resolves opaque server-side sessions, requires an active admin role, and revalidates account/session/role under row locks in the same transaction as aggregate reads.
- Added aggregate catalog/source quality, product/variant, vendor/offer, seller-membership, account-state, price-history and audit-activity diagnostics. Translation is explicitly reported as paused; no emails, token material, password hashes or audit payloads are returned.
- Added PostgreSQL integration tests for non-admin denial, administrator access, revoked-role/session races and aggregate diagnostics. Backend CI [41465375163](https://github.com/pooya-labs/globalcart/actions/runs/41465375163) and Security scans [41465375326](https://github.com/pooya-labs/globalcart/actions/runs/41465375326) passed on head `7ca99c04a348256aa0a6051241864213968f7f84`. The service remains internal; no route or dashboard was exposed.
- Added fail-closed backend runtime configuration: explicit APP_ENV modes, PostgreSQL URL/production TLS validation, host/port listener validation, and exact normalized CORS origins with HTTPS-only production policy. Added unit tests; Backend CI [41458549348](https://github.com/pooya-labs/globalcart/actions/runs/41458549348) and Security scans [41458549236](https://github.com/pooya-labs/globalcart/actions/runs/41458549236) passed.
- Added immutable PostgreSQL migration checksums: new ledger rows store SHA-256 of exact embedded SQL bytes, startup rejects changed or missing applied migrations, and pre-checksum NULL ledger entries are adopted once. Added PostgreSQL regression tests for backfill and drift detection.
- Added docs/DATABASE_MIGRATION_POLICY.md covering append-only migrations, expand-and-contract rollout, forward-fix defaults, transactional-runner limitations, backup/restore boundaries and incident recovery.
- Verified M04.13: Backend CI [41415943162](https://github.com/pooya-labs/globalcart/actions/runs/41415943162), Security scans [41415943090](https://github.com/pooya-labs/globalcart/actions/runs/41415943090) and GitHub Pages deployment [41415943078](https://github.com/pooya-labs/globalcart/actions/runs/41415943078) all passed; PostgreSQL checksum adoption/drift tests and logical backup/restore passed.
- Switched the FX feed from Nobitex (HTTP 403 from GitHub-hosted runners) to Wallex's public USDTTMN one-minute UDF candle close. Initial live snapshot succeeded in [Sync Wallex FX run 41409982177](https://github.com/pooya-labs/globalcart/actions/runs/41409982177); the UI identifies the Tether/Toman pair and candle timestamp, uses 5/15-minute freshness labels, and has no hard-coded fallback.
- Added catalog source freshness diagnostics to the storefront: show the last successful fetch or snapshot generation timestamp and classify its age as fresh, aging, stale or unknown in Persian and English. The displayed timestamp explicitly refers to GlobalCart fetch/generation time, not supplier modification time.
- Verified M02 freshness diagnostics on implementation commit `4c0365a82908551b81f52f9301b5356b0db6f041`: Catalog CI [41406450515](https://github.com/pooya-labs/globalcart/actions/runs/41406450515), Browser E2E [41406450651](https://github.com/pooya-labs/globalcart/actions/runs/41406450651), Pages deployment + smoke test [41406450478](https://github.com/pooya-labs/globalcart/actions/runs/41406450478), and Security scans [41406450499](https://github.com/pooya-labs/globalcart/actions/runs/41406450499) all passed.
- Added a committed npm lockfile and lockfile-aware install flow: Pages, Catalog CI and Browser E2E use `npm ci` once the lockfile exists. The bootstrap path persists the generated lockfile with the catalog snapshot and validates it with a clean `npm ci` before the frontend build.
- Added a post-deployment GitHub Pages smoke test for the HTTPS homepage, a Next.js JavaScript asset under the `/globalcart/` base path and one product detail deep link chosen from the generated catalog.
- Added a PostgreSQL 17 custom-format logical backup/restore regression to Backend CI. It restores into a fresh database and compares the migration ledger, key table counts, validated constraints and `pg_trgm` version.
- Added M04.12 customer-account ownership operations for cart binding, cart read/add/set/remove, parent and vendor-child order reads, and unpaid simulated-order cancellation. Every account path revalidates the active customer session and canonical PostgreSQL owner inside the protected transaction.
- Added PostgreSQL integration regression coverage for foreign and guest-only cart/order IDOR, customer cart mutations, vendor child-order reads, first-writer-wins binding races, expired carts, revoked sessions and idempotent cancellation.
- Preserved the independent guest-cart capability path; cart capability possession does not create account identity or grant customer/seller/admin roles.
- Added M04 migration 011 for identity/session/RBAC storage: normalized users, controlled roles, seller memberships, opaque session and action-token digests, and append-only audit events.
- Added database constraints for Argon2id hash shape, verified account activation, session expiry/revocation, one active action token per purpose, and immutable audit history.
- Added PostgreSQL-aware migration statement splitting for quoted strings, comments and dollar-quoted function bodies, plus unit and integration regression coverage.
- Added internal Argon2id password hash/verify and rehash-check helpers with strict PHC parsing and bounded verification resource costs.
- Added 256-bit cryptographically random opaque token generation, SHA-256 digest helpers and constant-time digest comparison.
- Added an internal PostgreSQL-backed session lifecycle: active/verified account creation, separate session and CSRF tokens, digest-only persistence, 15-minute idle refresh bounded by a 24-hour absolute expiry, and current DB-backed role/vendor-scope resolution.
- Added session-bound CSRF verification and transactional logout revocation/audit events, with PostgreSQL integration coverage for expiry, disabled accounts and seller membership state.


- Added Playwright axe coverage for the storefront in both Persian RTL and English LTR.
- Added a desktop/mobile `prefers-reduced-motion` regression test for smooth scrolling, image transitions and modal overlay animation.

### Verification
- M04.12 implementation head `928c3867e2d8e109e4e6595f9ac0898cf51b25b9` passed [Backend CI 41403135351](https://github.com/pooya-labs/globalcart/actions/runs/41403135351), [Security scans 41403135364](https://github.com/pooya-labs/globalcart/actions/runs/41403135364) and [Pages deployment 41403135344](https://github.com/pooya-labs/globalcart/actions/runs/41403135344). CI completed sqlc generation, gofmt, dependency vulnerability scan, vet, race-enabled tests, API/importer builds and Docker image build. Formatting-only generated commit: `5a5659d9b35789cf12072141152b0ab4be217c33`.
- M08.12 Backend CI [41404292955](https://github.com/pooya-labs/globalcart/actions/runs/41404292955) passed on `0e8c8556fa04fa43d735b46820685632eef2df0a`, including the PostgreSQL 17 logical backup/restore comparison; Security scans [41404293071](https://github.com/pooya-labs/globalcart/actions/runs/41404293071) also passed.
- M08.13 Pages deployment [41405289322](https://github.com/pooya-labs/globalcart/actions/runs/41405289322) passed clean npm lockfile installation, TypeScript, unit tests, static export, deployment and published-site smoke checks. Security scans [41405289317](https://github.com/pooya-labs/globalcart/actions/runs/41405289317) also passed on workflow implementation commit `8fec0d0c9bbb845fb7b1dc0788970929828e0a2a`.

### Fixed
- Increased the stale-catalog status foreground contrast from 4.32:1 to above 5:1 against its light background.
- Made reduced-motion overlay coverage use the touch-safe cart drawer rather than a hover-dependent quick-view control.

### Verification
- Backend CI [41385892498](https://github.com/pooya-labs/globalcart/actions/runs/41385892498), Security scans [41385892505](https://github.com/pooya-labs/globalcart/actions/runs/41385892505) and Pages deployment [41385892515](https://github.com/pooya-labs/globalcart/actions/runs/41385892515) passed on auth primitive implementation head `862ffb30edefb23695e34e407bfe7ea8180162ef`.
- Backend CI [41384871633](https://github.com/pooya-labs/globalcart/actions/runs/41384871633), Security scans [41384871597](https://github.com/pooya-labs/globalcart/actions/runs/41384871597) and GitHub Pages deployment [41384871618](https://github.com/pooya-labs/globalcart/actions/runs/41384871618) passed on the identity schema implementation commit.
- Browser E2E [41383420470](https://github.com/pooya-labs/globalcart/actions/runs/41383420470) passed all 30 Playwright tests across desktop Chromium and Pixel 7 mobile on `9653fc7d2d03e45f726b7e6a0e47fd7cbd1295dc`.
- Security scans [41383420409](https://github.com/pooya-labs/globalcart/actions/runs/41383420409) and GitHub Pages deployment [41383420345](https://github.com/pooya-labs/globalcart/actions/runs/41383420345) passed on the same implementation head.

### Remaining manual QA
- Screen-reader review, real-device checks, full-site RTL/LTR review and live CDN/Core Web Vitals measurement remain outstanding.


## Unreleased — 2026-11-02

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Added migration 010 for one full-order return request per paid simulated order, fixed reason codes, review timestamps and terminal `refunded_simulated` states.
- Added internal Go services for owner-scoped return requests, approve/reject review and idempotent full-order simulated refund completion.
- Added PostgreSQL integration coverage for ownership isolation, unpaid-order rejection, invalid/duplicate reasons, request/review/refund idempotency, parent/vendor child state transitions, event history and unchanged inventory.

### Fixed
- Explicitly typed JSON event parameters as text to avoid PostgreSQL's indeterminate parameter type error in `jsonb_build_object`.

### Safety boundary
- Return review and refund completion are not exposed through HTTP. No payment provider, money movement, shipment or stock mutation occurs.

### Verification
- Backend CI [41382309964](https://github.com/pooya-labs/globalcart/actions/runs/41382309964), Security scans [41382310050](https://github.com/pooya-labs/globalcart/actions/runs/41382310050) and Pages deployment [41382310024](https://github.com/pooya-labs/globalcart/actions/runs/41382310024) passed on `bc97475c0cce7f8de6aebe72071813c942b6a3d0`.
- PR #1 was squash-merged into `main` at `dc9b520b026a508825eb3dd1f6cbc3655be5f039`.


## Unreleased — 2026-11-02

### Fixed
- Preserve the original invoking control when Command Palette actions transition directly into cart, favorites or comparison dialogs.
- Keep dialog focus contained and restore focus after Escape dismissal across cart, favorites, quick view, comparison and checkout.

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Playwright regression coverage for Tab/Shift+Tab wrapping, Escape dismissal, focus restoration and the mobile Command Palette-to-saved-items path.
- Touch-appropriate Quick View interaction in mobile tests; the test no longer assumes a hover-capable pointer on a touch viewport.

### Verification
- Browser E2E push run [41380662904](https://github.com/pooya-labs/globalcart/actions/runs/41380662904) and PR run [41380669829](https://github.com/pooya-labs/globalcart/actions/runs/41380669829) passed: 28 Playwright tests, including desktop Chromium and Pixel 7 mobile.
- On final feature head `2182bbb1c55ad265e4264c1a0389fb0353f0010a`, Catalog CI [41380669723](https://github.com/pooya-labs/globalcart/actions/runs/41380669723), Backend CI [41380669722](https://github.com/pooya-labs/globalcart/actions/runs/41380669722) and Security scans [41380669865](https://github.com/pooya-labs/globalcart/actions/runs/41380669865) passed.


## 0.1.3 — 2026-11-01

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Server-side Gemini 2.5 Flash product translation adapter with strict JSON validation.
- Optional scheduled Persian catalog localization job, safely skipped when GEMINI_API_KEY is absent.
- Per-product translation provenance, source-content fingerprint caching, sequential throttling, bounded retries and error states.
- Storefront uses verified Persian product fields when available while preserving original English fields.

### Verification
- CI run 41295683515 passed catalog sync, optional translation step (safely skipped because no key is configured), TypeScript validation, unit tests, production build and GitHub Pages deployment.
- Run: https://github.com/pooya-labs/globalcart/actions/runs/41295683515

## 0.1.2 — 2026-11-01

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Scheduled and manually triggerable catalog refresh in GitHub Actions.
- Typed static catalog snapshot generator with retries, validation and atomic replacement.
- Last-known-good snapshot protection when the upstream sample API is unavailable.
- Storefront initialization from the generated catalog snapshot.
- Target architecture decision: Go REST/OpenAPI, PostgreSQL with SQL migrations/sqlc, Python data workers and optional Redis.

### Verification
- Actions run 41295039930 passed catalog sync (100 products), TypeScript validation, all unit tests, production build, Pages artifact upload and deployment.
- Run: https://github.com/pooya-labs/globalcart/actions/runs/41295039930

## 0.1.1 — 2026-11-01

### Changed
- Reworked demo pricing to use integer USD cents and fixed-point FX units.
- Used BigInt intermediate arithmetic to avoid floating-point drift during conversion.
- Explicit rounding: discounted USD rounds to cents, then Toman rounds half-up to a whole unit.
- Rejected non-finite discounts and unsafe input/output ranges.
- Expanded pricing tests for fractional values, rounding, clamping and invalid inputs.
- Removed the duplicate Pages starter workflow.

### Verification
- Actions run 41294484252 passed dependency installation, TypeScript validation, unit tests, production build, artifact upload and Pages deployment.

## [0.2.0-dev] - 2026-11-01
### Changed
- Re-scoped the storefront toward a full marketplace experience inspired by common Amazon-style discovery and shopping patterns, with an independent GlobalCart identity.
- Paused translation and removed the translation job and credentials from the GitHub Pages deployment workflow.
- Added a dedicated quick-category navigation rail and began the visual-system migration toward a cleaner, denser marketplace layout.

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Responsive catalog-driven category mega-menu with live category counts, direct filtering, backdrop dismissal and mobile layout.
- Quick category rail now reflects categories present in the current catalog rather than only the initial hand-picked list.

## 0.2.1-dev — 2026-11-01
### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Catalog-driven category mega-menu and dynamic quick-category rail.
- Dedicated static product routes with image gallery, product attributes, source disclosure, related products and demo cart actions.
- Side-by-side comparison shortlist for up to four products.
- Product cards now navigate directly to dedicated detail pages while retaining quick view.
### Fixed
- Corrected escaped comparison-state declarations and duplicate category-name key found during CI validation.
- Made catalog snapshot commits rebase on the latest main branch to avoid races with UI commits.


## 0.2.2-dev — 2026-11-01
### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Validated Amazon product-link intake with supported marketplace domains, ASIN extraction and a browser-local sourcing request queue.
- Marketplace service shortcuts for sourcing requests, product comparison, planned seller tools and GitHub feedback.
- FAQ content clarifying demo checkout, Amazon data access and comparison behavior.
- Responsive styles for the Amazon request module, service cards and FAQ.
- Keyboard-accessibility improvements: Escape dismissal, dialog semantics, visible focus indicators, live feedback announcements and localized quantity labels.
### Notes
- Amazon link submission is a local demonstration only. Live product details, pricing and availability require authorized Amazon API access.
- Translation remains paused and excluded from CI.
### Verification
- Actions run 41299580437 (Run #134) passed catalog refresh, TypeScript validation, unit tests, production build, Pages artifact upload and deployment.
- Run: https://github.com/pooya-labs/globalcart/actions/runs/41299580437


## 0.2.3-dev — 2026-11-01
### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Data-driven homepage merchandising rails for top discounts, top-rated products and fresh discoveries.
- Browser-local recently viewed product rail, updated from both quick view and dedicated product pages, with a clear-history action.
- Product detail price breakdown showing source USD price, catalog discount, illustrative Toman/USD rate and calculated demo total.
### Notes
- FX remains user-configured and illustrative; no live provider, timestamp, shipping, tax or markup is represented.
- Recently viewed data is stored locally in the browser; no account-level history is created.
### Verification
- Actions run 41299322140 (Run #126) passed catalog refresh, TypeScript validation, unit tests, production build, Pages artifact upload and deployment.
- Run: https://github.com/pooya-labs/globalcart/actions/runs/41299322140


### Recently viewed page follow-up
- Added a dedicated /recently-viewed route with empty state, clear-history action and Persian/English layout.
- Added a permanent footer entry so the history page remains discoverable even before the first product visit.
- Actions run 41299677941 (Run #135) passed TypeScript validation, unit tests, production build, Pages artifact upload and deployment.
- Run: https://github.com/pooya-labs/globalcart/actions/runs/41299677941


### Accessibility interaction pass
- Added Escape-key dismissal across homepage overlays and navigation menus.
- Added dialog roles, modal semantics and accessible title associations to cart, favorites, quick view, comparison and demo checkout.
- Added aria-expanded/aria-controls relationships for mobile and category navigation; gallery thumbnails now expose aria-pressed.
- Full focus trapping/restoration and manual assistive-technology QA remain open.

- Added modal focus management: initial focus, Tab/Shift+Tab trapping and focus restoration when dialogs close.


### M04 backend foundation — 2026-11-01
- Added a Go 1.23 REST API with health/readiness endpoints and read-only catalog list/detail routes.
- Added PostgreSQL 17 local Docker Compose, loopback-only port bindings and persistent development storage.
- Added embedded transactional SQL migrations for the initial products catalog schema.
- Added a one-shot CLI to import the existing normalized catalog snapshot into PostgreSQL.
- Added an OpenAPI 3.1 contract and backend CI for formatting, vet, race-enabled tests, builds and container build.
- This is local development infrastructure only. No hosted database/API, credentials, paid resources, authentication, real orders or payments were provisioned.

- Backend CI run #22 passed: PostgreSQL service startup, Go formatting, vet, race-enabled unit/integration tests, API/importer builds and Docker image build. Run: https://github.com/pooya-labs/globalcart/actions/runs/41301389812


### M04 catalog platform extension — 2026-11-01
#### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Normalized category table with product foreign-key backfill and category product counts.
- Product variants schema with SKU/source identity constraints, non-negative price/stock and JSONB attributes.
- sqlc v2 generation targeting PostgreSQL and pgx/v5; generated typed product, category and variant queries.
- Product list/detail handlers now use generated sqlc methods; importer retains explicit transactional upsert logic.
- Read-only product-variant endpoint and normalized category listing endpoint.
- Authentication/session/RBAC design covering Argon2id, opaque server-side sessions, CSRF, cookie policy, role scopes and required security tests.

#### Verification
- Actions run 41302783939 passed sqlc generation, Go formatting, vet, unit and PostgreSQL integration tests, API/importer builds and Docker image build.
- Run: https://github.com/pooya-labs/globalcart/actions/runs/41302783939
- Authentication remains design-only pending approval of same-site HTTPS hosting, email delivery and secret management. No user tables or login flows were added.


### Catalog pagination follow-up — 2026-11-01
- Category filtering now matches normalized category names case-insensitively and ignores surrounding whitespace.
- Added a typed total-count query sharing the same search and category predicates as product listing.
- Product list response now includes `total` for pagination while `count` continues to report items in the returned page.
- Added PostgreSQL integration coverage for category normalization and total count. Pull-request CI run 41303108214 passed sqlc generation, formatting, vet, race-enabled PostgreSQL tests, API/importer builds and Docker image build.


### Demo checkout service — 2026-11-01
- Added an internal Go transaction service for demo checkout with canonical catalog revalidation, row-level stock locks, expiring reservations, immutable order snapshots and idempotent replay.
- Added PostgreSQL integration coverage for competing carts, replay and idempotent reservation expiry; CI run 41346127244 passed.
- Service is not exposed as an HTTP endpoint; no real payments or live stock adjustments are performed.


### Guest cart lifecycle — 2026-11-01
- Added internal guest cart creation, item add/increment, quantity update, item removal and cart retrieval.
- Enforced digest-only guest capability inputs, per-cart transactional locking, canonical catalog price/title/stock reads and bounded quantities.
- Added PostgreSQL integration tests for lifecycle operations, variants, stock limits and expiry.
- Verified by CI push run 41346723656; no public cart endpoints or browser cookie flow enabled.


### Order detail and cancellation — 2026-11-01
- Added capability-scoped internal order detail reads with immutable item snapshots and event history.
- Added owner-scoped demo cancellation, atomic reservation release, cancellation event and idempotent replay.
- Added PostgreSQL integration tests for cross-cart isolation, order snapshots, event history and cancellation behavior.
- Verified by CI push run 41347325289; no public order endpoints enabled.


### Simulated payment lifecycle — 2026-11-02
- Added an owner-scoped, idempotent internal demo payment completion transition.
- Requires a live stock reservation; atomically releases holds, updates order state and appends a simulation event.
- Added PostgreSQL integration coverage for ownership, repeat completion, reservation release and invalid post-payment cancellation.
- No external payment provider or public payment endpoint is involved; Verified by CI run 41348079983, including PostgreSQL integration/race tests and Docker image build.


### Internal coupon foundation — 2026-11-02
- Added hashed coupon-code storage and internal coupon creation/application/removal methods.
- Added percentage and fixed discount calculation, validity windows, minimum subtotal, maximum discount and redemption limits.
- Checkout revalidates and records redemption atomically; order totals snapshot subtotal, discount and final total.
- Added cart coupon state and discounted order detail fields; no public routes or verified customer identity are enabled.


### M08 vendor foundation — 2026-11-02
- Added migration 005 with vendor lifecycle/type, vendor-specific base and variant offers, price history and inventory event ledger.
- Backfilled existing catalog products and variants into the active GlobalCart Direct platform vendor; recorded opening price/stock snapshots.
- Added PostgreSQL integration coverage for independent stock across two vendors, audit records and duplicate-offer constraints.
- Compatibility preserved: existing cart and checkout still use legacy catalog stock; multi-vendor cart/order splitting is not yet implemented.
- No seller membership, public seller API, dashboard, commission, payout or fulfillment was enabled.


### M09 — Vendor-aware cart foundation
- Added additive vendor-specific cart line and parent/child vendor-order schema.
- Added internal vendor-aware cart operations and legacy-checkout safety guard.
- Added integration coverage and architecture notes; split checkout remains pending.


### M09.2 — Vendor split checkout
- Implemented atomic parent/child order creation and per-vendor reservations.
- Added parent-level coupon redemption with exact child allocation.
- Propagated simulated payment, cancellation and expiry to child orders.
- Added PostgreSQL integration tests for checkout and inventory contention.


### M09.3 — Owner-scoped vendor child-order reads — 2026-11-02
- Added `GetVendorOrder` for reading a single vendor child order and its immutable item snapshots through the owning guest cart capability digest.
- Kept unknown child IDs and cross-owner access indistinguishable via `ErrOrderNotFound`.
- Added PostgreSQL integration coverage for child metadata, parent linkage, item snapshots, exact totals and cross-owner isolation.
- Backend CI verification is pending; see PR #1 before treating this slice as complete.


### M09.3 CI correction — 2026-11-02
- Inspected Backend CI run 41352752412 and identified the exact `go vet` failure: an unused `time` import in `backend/internal/commerce/vendor_orders.go`.
- Removed the unused import in commit `cb0f20816afdbf3f2dc23628e8a5bd4b7d1bce46`.
- The original failed run predates the correction. A fresh CI run is still required; M09.3 remains unverified until it passes.


## 0.2.3-dev — 2026-11-02
- M09.3 owner-scoped vendor child-order retrieval is CI-verified in Backend CI run `41352898725` (head `1234d02`).
- The vendor child-order read requires the guest capability digest and masks cross-owner IDs as not found.
- Authentication and seller RBAC remain gated by same-site HTTPS, approved email delivery and secret-management prerequisites; no public seller endpoints were added.

- Added PostgreSQL integration coverage for cross-vendor offer reads, price updates and inventory adjustments; verifies a foreign vendor scope cannot mutate the owning vendor's stock. CI verification pending.

- M09.4 cross-vendor isolation tests passed Backend CI run `41353137589` on code commit `e18a01d`; the full workflow including Docker image build succeeded.
- Documented the interim seller-service boundary: a vendor ID is not caller authorization and public seller routes remain disabled until verified sessions and active memberships are enforced.


## 0.2.4-dev — 2026-11-02
- Added HTTP handler regression tests to keep seller order/product routes, cart writes and checkout writes unavailable until verified identity and authorization exist.
- Added tests that catalog routes reject mutation methods and configured CORS origins receive only read-only methods. Backend CI verification pending.
- Reconciled roadmap and project context: M09.3 and M09.4 are CI-verified; M09.4.1 is the active regression-guard slice.

- Tightened pre-auth CORS policy by no longer advertising the `Authorization` request header; added a regression assertion. Latest code CI verification pending.


## 0.2.5-dev — 2026-11-02
- Catalog imports now create/update the GlobalCart Direct base-product offer transactionally alongside product/category data.
- Added append-only price and inventory audit events for initial offers and subsequent source changes; unchanged imports remain idempotent.
- Added PostgreSQL integration coverage for initial, repeated and changed catalog imports. CI verification pending.
- Verified public API regression guards and pre-auth CORS restrictions in Backend CI run `41353463006`.

- Platform catalog synchronization now includes product variants, mapping source SKU, price, stock and active/paused state into vendor offers. Added PostgreSQL regression coverage for idempotency and price/inventory audit events. CI verification pending.


### Platform offer synchronization verified — 2026-11-02
- Backend CI run `41354208659` passed all steps for base-product and existing-variant platform offer synchronization, including PostgreSQL integration tests and Docker image build.

### End-to-end source variant ingestion — in progress
- Added optional source variants to the normalized TypeScript catalog contract and Go importer model.
- Importer now transactionally upserts variant identity, SKU, title, price, stock, attributes and active state before synchronizing the GlobalCart Direct offer.
- Added normalization and PostgreSQL integration coverage for variants delivered in the product snapshot.
- The current DummyJSON source does not provide variant records; ingestion is optional and awaits an authorized supplier source. CI verification pending.


### M08.2 source variant ingestion — CI verified (2026-11-02)
- Extended the normalized catalog snapshot and Go importer contract to carry optional source variants.
- Persisted variant source identity, SKU, title, price, stock, attributes and active state in the same import transaction as the parent product.
- Added PostgreSQL integration coverage for variant creation, idempotent re-import, changed offer price/stock audit events and inactive-state synchronization.
- Backend CI run `41354535515` passed all steps, including PostgreSQL-backed race tests, API/importer builds and Docker image build.

### M08.3 supplier adapter contract — implementation in progress (2026-11-02)
- Added a common adapter interface for fetching supplier-specific records and mapping them into the canonical catalog input.
- Refactored DummyJSON and scheduled snapshot generation to use the shared ingestion boundary while retaining retry and last-known-good behavior.
- Corrected variant de-duplication to use the same source/source-ID identity pair enforced by PostgreSQL.
- Added synthetic supplier fixture coverage and dedicated Catalog CI for typecheck, tests and production build.
- No real supplier feed is connected; current verification is pending the new Catalog CI run.


### M08.3 supplier adapter contract — CI verified (2026-11-02)
- Added the shared `CatalogSourceAdapter` interface and `loadCatalogFromSource` ingestion boundary.
- Refactored DummyJSON and scheduled snapshot sync to use the shared adapter while retaining retries and last-known-good snapshot protection.
- Aligned variant deduplication with the database's `(source, source_id)` identity pair.
- Added a synthetic-only supplier fixture with two variants and tests for source mapping, provenance, SKU, attributes, inactive state, source-scoped identity and empty/unmappable feed rejection.
- Added dedicated Catalog CI for TypeScript validation, catalog/domain unit tests and production build.
- Verification: Catalog CI push run [41355090849](https://github.com/pooya-labs/globalcart/actions/runs/41355090849) and PR run [41355095558](https://github.com/pooya-labs/globalcart/actions/runs/41355095558) both succeeded on head `9bfab0c80dec906f52fb56a22b85d36b7eec99f9`.

### M08.4 authorized supplier onboarding — blocked
- Next integration requires a named supplier and owner-approved API/feed access and usage rights.
- No real supplier connection, credentials, scraped inventory or fabricated live variants are included pending that decision.


### Variant identity collision hardening — 2026-11-02
- Replaced delimiter-concatenated variant identity keys with an unambiguous tuple encoding of `(source, sourceId)`.
- Added a regression test proving that pairs such as `("supplier:one", "id")` and `("supplier", "one:id")` remain distinct.
- Catalog CI run `41355272122` passed TypeScript validation, all unit tests and production build on head `eaebbc0a7b8d1c823c3dcad4a7fed0d10a99edc9`.


### M08.4 catalog snapshot integrity — implementation in progress (2026-11-02)
- Added a conservative full-snapshot continuity guard that rejects a product-count drop greater than 50% compared with the last-known-good snapshot.
- The guard executes before atomic snapshot replacement, so suspiciously incomplete refreshes preserve the existing catalog snapshot.
- Added a diagnostic report for product/variant coverage and common data-quality signals; these are reported without fabricating or automatically deleting catalog data.
- Added unit tests for threshold boundaries, invalid settings and quality-report counts. Final Catalog CI verification is pending.


### M08.4 catalog snapshot continuity and quality diagnostics — CI verified (2026-11-02)
- Added a full-snapshot continuity guard that rejects a normalized product-count drop greater than 50% compared with the last-known-good snapshot, before the new snapshot can replace it.
- Added a quality report covering total products/variants, product-only coverage, zero-price/zero-stock products, missing descriptions and uncategorized products.
- Diagnostic metrics do not fabricate inventory or auto-delete data; product-only sources are valid.
- Verification: Catalog CI push run [41355552153](https://github.com/pooya-labs/globalcart/actions/runs/41355552153) and PR run [41355595494](https://github.com/pooya-labs/globalcart/actions/runs/41355595494) passed. Backend CI PR run [41355595650](https://github.com/pooya-labs/globalcart/actions/runs/41355595650) also passed SQL generation, Go/PostgreSQL tests, builds, Docker image build and cleanup.


### M08.5 variant omission and inventory retention — implementation in progress (2026-11-02)
- Added PostgreSQL integration coverage for a source that previously supplied variants but later sends a product-only snapshot.
- The regression test asserts that existing product variants are retained, platform offer values/status remain unchanged, independent merchant offers are untouched, and no spurious audit events are written.
- No automatic variant deletion, stock-zeroing or offer archival behavior was added. Backend CI verification is pending.


### M08.5 variant omission and inventory retention — CI verified (2026-11-02)
- Added PostgreSQL integration coverage for a source that later sends a product-only snapshot after previously supplying variants.
- Verified that existing variants and platform offers are retained; an independent merchant offer is unchanged and no spurious audit events are appended.
- Backend CI PR run [41356201201](https://github.com/pooya-labs/globalcart/actions/runs/41356201201) passed SQL generation, Go vet, race-enabled PostgreSQL tests, API/importer builds, Docker image build and cleanup.

### M08.6 variant source identity ownership — implementation in progress (2026-11-02)
- Added a PostgreSQL regression test for attempted reassignment of a supplier's existing variant source identity to a different product.
- The test checks that the whole multi-product import batch rolls back and the original variant owner, SKU, price and stock remain intact.
- Catalog CI PR run [41363342797](https://github.com/pooya-labs/globalcart/actions/runs/41363342797), Backend CI PR run [41363342758](https://github.com/pooya-labs/globalcart/actions/runs/41363342758), and Backend CI push run [41363337241](https://github.com/pooya-labs/globalcart/actions/runs/41363337241) passed on implementation commit `cc36339bcac77d08abc37af3f88f08607a090367`.


### M08.5 variant omission and inventory retention — CI verified
- Backend CI PR run [41356201201](https://github.com/pooya-labs/globalcart/actions/runs/41356201201) passed the PostgreSQL regression proving product-only refreshes do not delete stored variants, alter platform/merchant offers, or append duplicate audit events.

### M08.6 variant source identity ownership and batch rollback — CI verified (2026-11-02)
- Added PostgreSQL integration coverage proving a source variant identity cannot be reassigned to a different product.
- Verified that the entire multi-product import rolls back on identity conflict and the original variant owner/SKU/price/stock remain unchanged.
- Backend CI push run [41356390061](https://github.com/pooya-labs/globalcart/actions/runs/41356390061) and PR run [41356442182](https://github.com/pooya-labs/globalcart/actions/runs/41356442182) passed, including race-enabled PostgreSQL tests and Docker image build.

### M08.7 explicit feed completeness semantics — implementation in progress
- Next adapter hardening will distinguish complete supplier snapshots from incremental/delta feeds and prevent deltas from replacing the last-known-good full catalog.


### M08.7 explicit feed completeness semantics — implementation in progress (2026-11-02)
- Added a required `snapshotMode` declaration to the source adapter contract: `full` or `incremental`.
- Classified DummyJSON as a full-snapshot source and required full mode in the scheduled static snapshot writer.
- Feed-mode mismatches are rejected before fetching records, preserving the last-known-good snapshot.
- Added a test proving an incremental feed can be normalized for a future merger but cannot enter the full snapshot path. Catalog CI verification is pending.


### M08.7 explicit feed completeness semantics — CI verified (2026-11-02)
- Added required `snapshotMode` metadata (`full` or `incremental`) to every catalog source adapter.
- Classified DummyJSON as a full snapshot and made the scheduled snapshot writer explicitly require that mode.
- Mode mismatches are rejected before fetching, protecting the last-known-good catalog from accidentally replacing it with a delta batch.
- Generic ingestion may normalize incremental records for a future dedicated merger, but this change does not merge or persist delta records.
- Verification: Catalog CI PR run [41356770768](https://github.com/pooya-labs/globalcart/actions/runs/41356770768) and Backend CI PR run [41356770801](https://github.com/pooya-labs/globalcart/actions/runs/41356770801) both passed on the same branch head, including production and Docker builds.


### M08.8 feed mapping quality and request safety — CI verified (2026-11-02)
- Added a default 20% maximum mapping-rejection fraction to shared ingestion, with a validated per-source override. Duplicate canonical identities remain handled by normal de-duplication.
- Preserved distinct errors for empty and wholly unmappable feeds so existing failure semantics remain clear.
- Combined caller cancellation with a 20-second per-attempt DummyJSON timeout; explicitly cancelled requests stop immediately and are not retried.
- Limited retries to transient network/time-out failures and HTTP 408, 429 and 5xx responses, with at most three attempts. Permanent 4xx responses, invalid JSON/schema and empty collections fail fast.
- Added regression tests for mapping thresholds, invalid configuration, composed request signals, cancellation, permanent HTTP errors and transient 503 recovery.
- Verification on head `a4a0d16c2aa06cfc83558b6516c9ac282cec1701`: Catalog CI [41358956297](https://github.com/pooya-labs/globalcart/actions/runs/41358956297) and Backend CI [41358956380](https://github.com/pooya-labs/globalcart/actions/runs/41358956380) both passed.
- M08.9 authorized supplier onboarding remains blocked pending explicit owner approval of a named source and usage rights.


### M08.8.1 source-agnostic feed import diagnostics — CI verified (2026-11-02)
- Added a structured import result with aggregate source, snapshot mode, timestamp, fetched/mapped/unmapped counts, mapping rejection rate, normalized product count and records discarded during normalization.
- Kept the original products-only loader API unchanged for existing consumers.
- Scheduled catalog sync now emits the aggregate report without logging raw source records.
- Added a regression test covering mapper rejection, malformed normalized products and duplicate identities.
- Catalog CI [41359432556](https://github.com/pooya-labs/globalcart/actions/runs/41359432556) and Backend CI [41359432659](https://github.com/pooya-labs/globalcart/actions/runs/41359432659) passed on code head `29956b0423328bcf9ce28c8fbe83d391aff22a98`.


### M08.8.2 collision-safe atomic snapshot publication — CI verified (2026-11-02)
- Extracted static snapshot persistence into a reusable helper using a unique, exclusively-created temporary file in the destination directory.
- Serialize before touching disk; rename only after the full JSON file has been written, then clean up the invocation's temp file on success or failure.
- Added tests for concurrent publication, rename failure cleanup, and preserving the previous snapshot when serialization fails.
- Catalog CI [41360183218](https://github.com/pooya-labs/globalcart/actions/runs/41360183218) and Backend CI [41360183283](https://github.com/pooya-labs/globalcart/actions/runs/41360183283) passed on code head `062aabdf563626116a50ecba0d7f534e6af3f260`.


### M08.8.3 last-known-good snapshot validation — CI verified (2026-11-02)
- Added strict runtime validation for the versioned catalog snapshot and its normalized product records before reusing it as last-known-good data.
- Only ENOENT is treated as a first-run missing snapshot; JSON corruption, schema mismatch and other filesystem errors now fail closed rather than silently triggering fixture fallback.
- Added tests for missing, malformed and structurally inconsistent snapshots.
- On implementation commit `ff5cf42cc094bc4e35d317ad52f970762280ba2f`, Catalog CI PR run [41361214644](https://github.com/pooya-labs/globalcart/actions/runs/41361214644), Catalog CI push run [41361210412](https://github.com/pooya-labs/globalcart/actions/runs/41361210412), and Backend CI PR run [41361214721](https://github.com/pooya-labs/globalcart/actions/runs/41361214721) passed.


### M08.8.4 snapshot identity and database-bound validation — CI verified (2026-11-02)
- Snapshot validation now checks global product identity and variant source-identity uniqueness, and requires each product to match the declared snapshot source.
- Added PostgreSQL-compatible validation for ID safety, price precision/range, discount/rating bounds and integer stock limits.
- Added regression tests for duplicate identities, wrong source attribution and invalid numeric/stock values.
- Catalog CI PR run [41362606331](https://github.com/pooya-labs/globalcart/actions/runs/41362606331), Catalog CI push run [41362602994](https://github.com/pooya-labs/globalcart/actions/runs/41362602994), and Backend CI PR run [41362606267](https://github.com/pooya-labs/globalcart/actions/runs/41362606267) passed on implementation commit `e9d864f6bdbf0013137929e02b6258007200db29`.

- Refinement: validate the newly built snapshot before writing it and reject duplicate non-empty variant SKUs, matching the PostgreSQL partial unique index.


### M08.8.5 PostgreSQL rating domain constraint — CI verified (2026-11-02)
- Added migration 007 enforcing the canonical maximum product rating of 5 at the database boundary.
- Added a PostgreSQL batch-rollback regression for an out-of-range 5.01 rating.
- Catalog CI PR run [41363342797](https://github.com/pooya-labs/globalcart/actions/runs/41363342797), Backend CI PR run [41363342758](https://github.com/pooya-labs/globalcart/actions/runs/41363342758) and Backend CI push run [41363337241](https://github.com/pooya-labs/globalcart/actions/runs/41363337241) all passed on implementation commit `cc36339bcac77d08abc37af3f88f08607a090367`.


## Unreleased — Catalog API and storefront facets (2026-11-02)

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Added normalized brand and inclusive source-price/rating facets to the read-only PostgreSQL product API. The paginated list and total count now share the same predicates.
- Added strict decimal/range validation, stable API errors, OpenAPI documentation, PostgreSQL indexes and integration coverage.
- Added bilingual storefront brand and minimum-rating filters, with a shared reset path for all active filters.
- Extracted local catalog filtering into a pure helper and added regression tests for combined filters, normalized brand matching, inclusive rating bounds and non-mutation.

### Verification
- Implementation head `3372040d3cd71cd55d99724033065cf0dfa4537e` passed [Catalog CI push](https://github.com/pooya-labs/globalcart/actions/runs/41365136582), [Catalog CI pull request](https://github.com/pooya-labs/globalcart/actions/runs/41365141727) and [Backend CI pull request](https://github.com/pooya-labs/globalcart/actions/runs/41365141780).


## Unreleased — Accessible discovery and typo-tolerant search (2026-11-02)

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Added bounded, normalized product/brand/category suggestions with accessible combobox/listbox keyboard operation.
- Added a Ctrl/⌘+K command palette for frequently used storefront navigation and local demo actions.
- Added PostgreSQL trigram fallback search for product-title and brand typos, with exact full-text results ranked first.
- Added GIN indexes, a forward-only pg_trgm migration, OpenAPI documentation and PostgreSQL integration tests for fuzzy title/brand search and count consistency.

### Verification
- Catalog CI push 41366011979 and PR 41366016027 passed for the storefront discovery feature.
- Final combined head 87bc930da16367bd725b68fec460bae2fda6f26e passed Backend CI push 41366589804, Backend CI PR 41366594050 and Catalog CI PR 41366593947.


## Unreleased — Explainable recommendations (2026-11-02)

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Replaced the storefront's opaque rating/discount multiplication with a documented 70/30 sample-signal score, a 50% discount contribution cap and stable six-decimal precision.
- Added per-product explanation labels and a clear disclosure that the demo ranking is not personalized.
- Added recommendation helper tests for weights, cap, reasons, tie-breakers, invalid inputs and source-array non-mutation.

### Verification
- Implementation head `09fd73ea1f1dcb1f01e5022c3f4380ec2832855a` passed Catalog CI push `41366975455`, Catalog CI PR `41366979976` and Backend CI PR `41366979999`.


## Unreleased — Saved searches and resilient catalog states (2026-11-02)

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Added browser-local saved searches for query, category, brand, price/rating facets and sort order, with naming, apply/delete, duplicate prevention, validation and a twelve-entry cap.
- Added explicit loading, live, stale-snapshot and fixture-fallback status handling. Empty online responses no longer silently preserve a seemingly fresh state or overwrite the last-known-good snapshot.
- Added pure helpers and tests for saved-search normalization and source-state transitions.

### Verification
- Implementation head ff80ebcfcf4bcec4cdbbf9217c866070614524be passed Catalog CI push 41367502278, Catalog CI PR 41367539860 and Backend CI PR 41367539848.


## Unreleased — Browser journey coverage (2026-11-02)

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Added Playwright desktop and mobile browser tests against the generated GitHub Pages static export, with external network calls blocked for repeatability.
- Added browser coverage for locale/product navigation, keyboard search suggestions, command palette, saved-search persistence, cart drawer and simulated checkout.
- Added an accessible label to the responsive cart trigger so the icon-only mobile button keeps a useful accessible name.

### Verification
- Commit `08c86aeb909228ecc1ba03d35195ad4fba6ab987` passed Browser E2E push `41369733216`, Browser E2E PR `41369737767`, Catalog CI PR `41369737769` and Backend CI PR `41369737810`.


## Unreleased — Accessibility improvements (2026-11-02)

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Added axe-core Playwright audits for the homepage, command palette and static product detail page on desktop Chromium and Pixel 7 mobile.
- Improved muted-text, rating and discount contrast; enlarged FAQ and footer navigation targets; fixed the responsive cart's accessible name and labeled the sort select.

### Verification
- Commit `e515d696873aa2811c3edb35fc89ace74ae874df` passed Browser E2E push `41370817992`, Browser E2E PR `41370823048`, Catalog CI PR `41370822987` and Backend CI PR `41370823026`.
- The automated scan covers selected states and is not a replacement for manual assistive-technology review.


## Unreleased — Keyboard focus restoration (2026-11-02)

### Fixed
- The Command Palette explicitly restores focus to the invoking control after dismissal, including when opened by Ctrl/⌘+K from the search field.

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Added Playwright keyboard-only checks for Tab/Shift+Tab focus containment and Escape focus restoration on desktop Chromium and Pixel 7 mobile.

### Verification
- Commit `8937b54f0f631695ee02e51a128cc42815988f43` passed Browser E2E push `41375898314` and PR `41375903109` (18 passed), Catalog CI PR `41375903056`, and Backend CI PR `41375903095`.


## Unreleased — Image loading and performance budgets (2026-11-02)

### Improved
- Prioritized the homepage hero and product-detail main images, reduced the hero source request to 1200px/q80, and deferred non-critical images with lazy loading and async decoding.
- Added desktop/mobile regression checks for image loading policy, reserved media geometry and first-party JS/CSS transfer budgets.

### Verification
- Commit `a741cf60ab44a17acf31bba43f47eb40204f2be5` passed Browser E2E push `41376758145` and PR `41376763890` (24 passed), Catalog CI PR `41376763985`, and Backend CI PR `41376763923`.
- Baseline was 512,814 bytes of JS and 67,702 bytes of CSS on the initial route. External CDN image timing and Core Web Vitals were not measured because browser tests block external hosts.


## Unreleased — transaction-bound bulk session revocation (CI verified, 2026-11-02)

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Added an allow-listed `SessionRevocationReason` and `RevokeUserSessions` helper for password changes/resets, account disablement, privilege changes, administrative revocation and security rotation.
- The helper accepts the caller's `pgx.Tx`, locks the account row to serialize against session creation, revokes all currently active sessions and appends one redacted audit event per session.
- Added PostgreSQL integration coverage for caller rollback, invalid reasons, per-user isolation, audit attribution scoped specifically to revocation events, and idempotent repeated calls.
- Added a fast unit test for the full allow-list, including explicit rejection of `logout` (handled by single-session revocation) and unknown reasons.
- Kept the helper internal; no auth HTTP route, cookie or hosted infrastructure was enabled.

### Verification
- Backend CI [run 41393919125](https://github.com/pooya-labs/globalcart/actions/runs/41393919125) and Security scans [run 41393919064](https://github.com/pooya-labs/globalcart/actions/runs/41393919064) passed on code head `6e044b63048b0911c5a7078279e38fe354898605`, including race-enabled PostgreSQL tests and Docker image build.


## Unreleased — atomic password changes and stale-login protection (CI verified, 2026-11-02)

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Require `CreateSession` callers to supply the password-hash snapshot they verified; the service compares it under the account row lock before issuing an opaque session.
- Added internal `ChangePassword`, which verifies the expected old hash, stores a fresh Argon2id hash, revokes all active sessions and records a redacted password-change audit event in one transaction.
- Added PostgreSQL integration coverage for multi-session invalidation, current-password verification and rejection of a stale login credential snapshot.
- Kept the feature internal; no password-reset-token flow or public auth endpoint was enabled.

### Verification
- Backend CI [run 41395032305](https://github.com/pooya-labs/globalcart/actions/runs/41395032305), Security scans [run 41395032374](https://github.com/pooya-labs/globalcart/actions/runs/41395032374) and Pages deployment [run 41395032292](https://github.com/pooya-labs/globalcart/actions/runs/41395032292) passed on code head `fe2ec6b9bfce6ec4aa20ea004029e2a68d5b05ab`.


## Unreleased — one-time password-reset token lifecycle (CI verified, 2026-11-02)

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Reused migration 011 `auth_action_tokens` to issue digest-only, opaque password-reset tokens with a bounded lifetime and rotation of previous outstanding tokens.
- Added atomic token consumption, Argon2id password update, sibling-token revocation, active-session revocation and redacted audit events.
- Normal password changes now invalidate outstanding password-reset tokens.
- Added PostgreSQL tests for digest-only storage, expiry, rotation, single use and concurrent double consumption.
- No email delivery or public reset route was enabled.

### Verification
- Backend CI [run 41395571464](https://github.com/pooya-labs/globalcart/actions/runs/41395571464), Security scans [run 41395571531](https://github.com/pooya-labs/globalcart/actions/runs/41395571531) and Pages deployment [run 41395571666](https://github.com/pooya-labs/globalcart/actions/runs/41395571666) passed on code head `8a504dca8ae23ca2c06143808452898281216abe`.


## Unreleased — one-time email-verification token lifecycle (CI verified, 2026-11-02)

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Reused migration 011 `auth_action_tokens` for digest-only email-verification tokens; no new table or migration was added.
- Added bounded issuance/rotation for pending accounts and one-time transactional activation with sibling-token revocation.
- Verification preserves the database-assigned customer role and cannot grant seller/admin permissions.
- Added PostgreSQL tests for digest-only storage, expiry, rotation, disabled accounts, replay and concurrent consumption.
- No email provider or public registration/verification route was enabled.

### Verification
- Backend CI [run 41396081833](https://github.com/pooya-labs/globalcart/actions/runs/41396081833) and Security scans [run 41396081844](https://github.com/pooya-labs/globalcart/actions/runs/41396081844) passed on code head `d720ffdebd77c6b28c82ce55567ea3bba90d95c8`.


## Unreleased — audited administrator-role operations (CI verified, 2026-11-02)

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Added internal `GrantAdminRole` and `RevokeAdminRole` using the existing migration 011 role tables.
- Require a different active, verified administrator actor; reject self-change, duplicate grant and missing revoke.
- Serialize role changes using an advisory transaction lock and stable user-row locking.
- Atomically revoke the target's sessions and append actor-attributed audit events for every role change.
- Added PostgreSQL coverage for unauthorized/disabled actors, pending targets, session invalidation, audit attribution and concurrent grants.
- No public role-management endpoint or self-bootstrap mechanism was added.

### Verification
- Backend CI [run 41396823353](https://github.com/pooya-labs/globalcart/actions/runs/41396823353), Security scans [run 41396823371](https://github.com/pooya-labs/globalcart/actions/runs/41396823371) and Pages deployment [run 41396823403](https://github.com/pooya-labs/globalcart/actions/runs/41396823403) passed on code head `aef92e7526cfbc9c51f0d803bcba8493e131d8d2`.


## Unreleased — internal seller-membership lifecycle (Backend CI and Security verified, 2026-11-02)

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Added internal seller-membership request, activation, suspension and revocation using migration 011; no additional schema migration was needed.
- Pending requests grant no seller role. Admin activation grants seller only for an active, verified account and active vendor.
- Preserve vendor scopes across multiple memberships and revoke seller role only after the last active membership is removed.
- Revoke affected sessions and append actor-attributed membership/role audit events transactionally.
- Added PostgreSQL coverage for multi-vendor scopes, disabled-account cleanup, inactive vendor/requester rejection and concurrent activation.
- No seller HTTP route, dashboard, self-approval or payout feature was added.

### Verification
- Backend CI [run 41397502226](https://github.com/pooya-labs/globalcart/actions/runs/41397502226) and Security scans [run 41397502232](https://github.com/pooya-labs/globalcart/actions/runs/41397502232) passed on code head `f0756e0a3360b949b384aef1ff8e3a646192257b`.
- GitHub Pages deployment [run 41397502248](https://github.com/pooya-labs/globalcart/actions/runs/41397502248) also passed.

## Unreleased — deny-by-default internal authorization policy (CI verified, 2026-11-02)

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Added a pure internal action/resource policy evaluator for customer-owned resources, seller vendor scope and administrator-only operations.
- Require a `SessionPrincipal` resolved through the server-side session store; deny missing identity, unassigned vendor scopes and unknown actions by default.
- Added table-driven tests for ownership, scope boundaries, admin-only operations, missing identity and client-supplied scope attempts.
- No authentication middleware, cookies, CSRF route or public protected endpoint was enabled.

### Verification
- Backend CI [run 41398064895](https://github.com/pooya-labs/globalcart/actions/runs/41398064895), Security scans [run 41398064885](https://github.com/pooya-labs/globalcart/actions/runs/41398064885) and Pages deployment [run 41398064905](https://github.com/pooya-labs/globalcart/actions/runs/41398064905) passed on implementation head `c1b4b0ccfbd132ee95f8dc757a3518ff9b04899a`.
 
## Unreleased — transactional seller service authorization (CI verified, 2026-11-02)

### Changed
- Replaced Wallex's USDT/Toman candle feed with the requested TGJU free-market USD page. The adapter selects `price_dollar_rl`, preserves the source's Rial quote, converts it to Toman per USD, and derives freshness from TGJU's Tehran update clock. Blocked pages, CAPTCHA, schema drift, date-only timestamps and invalid prices fail closed without overwriting the last-known-good snapshot.
- Updated the storefront, product breakdown, recently-viewed page, E2E expectations, architecture, source policy, README and handoff docs to identify TGJU accurately; removed the Tether-specific wording. The scheduled refresh remains a low-frequency direct page request and does not bypass source access controls.

### Added
- Added a transaction-bound seller guard that rechecks active account, session, seller role, active vendor and exact active membership while holding PostgreSQL row locks.
- Added seller-scoped offer read, price update and stock adjustment methods that derive the vendor from the canonical offer record rather than client input.
- Reused existing commerce price-history and inventory-event helpers to keep behavior consistent.
- Added PostgreSQL integration coverage for own-vendor success, cross-vendor IDOR denial, stale/revoked session denial and membership-revocation serialization against an in-flight seller operation.
- No public seller route, browser cookie or middleware was enabled.

### Verification
- Backend CI [run 41399542624](https://github.com/pooya-labs/globalcart/actions/runs/41399542624), Security scans [run 41399542539](https://github.com/pooya-labs/globalcart/actions/runs/41399542539) and Pages deployment [run 41399542545](https://github.com/pooya-labs/globalcart/actions/runs/41399542545) passed on code head `f06de264488bd59f1e76a7e3d6c67f26d40c0eb1`.
