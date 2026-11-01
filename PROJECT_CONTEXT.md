# GlobalCart — Project Context / Single Source of Truth

**Owner:** Pooya Goodarzi ([GitHub](https://github.com/pooya16pg-spec))  
**Repository:** https://github.com/pooya16pg-spec/globalcart  
**Branch:** main  
**Current release:** 0.1.0  
**Active milestone:** M02  
**Last context update:** 2026-11-01

## Product definition
GlobalCart is an independent portfolio engineering project: a realistic marketplace demonstration built to showcase engineering skill. It is not a live store. Checkout, payment, inventory, shipping, seller, order, refund and analytics flows are simulations.

## Independence — mandatory
This repository must remain standalone. Do not add runtime imports, packages, shared deployment dependencies, cross-repo workflows or build dependencies on Pooya.OS, ORBIT, AuraPulse, NexusCanvas, LuminaGlass or any other portfolio project. Cross-links in portfolio documentation are allowed only if later approved.

## Product and safety invariants
1. Never collect card numbers, bank credentials or real payment details.
2. Never call a real payment gateway or create real orders, reservations, shipments, inventory changes or payouts.
3. Label demo data, local fallback, illustrative currency and simulated outcomes honestly.
4. Do not scrape Amazon or copy its catalog. An Amazon adapter needs authorized access and review of current terms.
5. Preserve original product text and source provenance when translating.
6. Never expose provider secrets in client JavaScript or repository files.
7. Persian is default and must use real RTL. English is also supported.
8. All visible controls must work or have a clear disabled/unavailable state.
9. Keep owner attribution visible.
10. Do not mark roadmap work complete until implemented and verified.

## Current repository map
- app/page.tsx — storefront, catalog view, cart, favorites and simulated checkout.
- app/globals.css — responsive marketplace visual system.
- app/layout.tsx — document metadata and global styles.
- lib/catalog/types.ts — canonical normalized product contract.
- lib/catalog/normalize.ts — record validation, normalization and deduplication.
- lib/catalog/dummyjson.ts — DummyJSON source adapter.
- lib/catalog/index.ts — catalog public exports.
- lib/pricing.ts — pure demo price calculation foundation.
- lib/translation/types.ts — provider contract and localization provenance.
- lib/translation/index.ts — safe localization orchestration with honest unavailable/error states.
- tests/catalog.test.ts — catalog normalization tests.
- tests/pricing.test.ts — pricing tests.
- tests/translation.test.ts — translation orchestration tests.
- .github/workflows/deploy.yml — CI and GitHub Pages deployment.
- ROADMAP.md — milestone acceptance checklist.
- docs/AI_HANDOFF.md — mandatory instructions for the next AI/developer.

## Current implementation status

### M01 — Storefront
Implemented: responsive storefront, Persian/English shell, search/filter/sort, quick view, favorites, recently viewed IDs, local cart, simulated checkout, illustrative FX, owner/source disclosure.
Pending verification: latest CI run, GitHub Pages configuration and live URL, keyboard/accessibility and responsive QA.

### M02 — Catalog intelligence
Implemented: typed product contract; normalization; required-field validation; numeric bounds; source IDs; import timestamp; duplicate suppression; separated DummyJSON adapter; translation provider interface; translation metadata and explicit unavailable/error states; initial unit tests.
Pending: verify test suite in Actions; actual provider integration; translation cache/retries; scheduled static catalog generation; admin diagnostics.

Important: imported product descriptions are not yet automatically translated. Never claim they are.

## CI blocker and deployment configuration
The original workflow failed because setup-node npm caching expected a lockfile that did not exist. The workflow was changed to remove the cache dependency and use npm install. A subsequent run successfully completed dependency installation, TypeScript validation, production build and artifact upload. Its deploy job then returned HTTP 404 because GitHub Pages is not enabled/configured for this repository. In repository Settings → Pages, set Build and deployment → Source to **GitHub Actions** once. Re-check the latest run after this setting.

The workflow now includes a unit-test step. Verify the latest run to confirm the new tests pass; do not infer success from an earlier run.

## Mandatory continuation protocol
1. Read this file, ROADMAP.md, docs/AI_HANDOFF.md, docs/ARCHITECTURE.md, docs/DECISIONS.md and CHANGELOG.md.
2. Inspect current main tree, package.json, relevant source files and latest Actions status.
3. Continue from the earliest unchecked roadmap acceptance item.
4. Keep repository independence and static Pages compatibility unless a documented architecture decision changes them.
5. Run npm install, npm run typecheck, npm test and npm run build.
6. Inspect the diff and verify the deployed site and basePath assets.
7. Update ROADMAP.md, CHANGELOG.md and this file after verified work. Update package.json at release boundaries.
8. State exact completed work, verification evidence, known limitations and next action.

## Immediate next actions
1. Verify newest Actions run including unit tests.
2. Enable GitHub Pages with Source = GitHub Actions, then verify deploy and public URL.
3. Generate and commit package-lock.json; return CI to npm ci for deterministic installs.
4. Add a server-side translation provider adapter and scheduled catalog generation, preserving translation provenance.
5. Add tests for product localization, API failure, currency rounding and cart state.
