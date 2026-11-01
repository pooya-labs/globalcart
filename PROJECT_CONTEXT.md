# GlobalCart — Project Context / Single Source of Truth

**Owner:** Pooya Goodarzi ([GitHub](https://github.com/pooya16pg-spec))  
**Repository:** https://github.com/pooya16pg-spec/globalcart  
**Branch:** main  
**Current release:** 0.1.0  
**Active milestone:** M02  
**Last context update:** 2026-11-01

## Product definition

GlobalCart is an independent portfolio engineering project: a realistic marketplace demonstration built to showcase engineering skill. It is not a live store. All checkout, payment, inventory, shipping, seller, order, refund and analytics flows are simulated unless a future decision explicitly says otherwise.

## Independence — mandatory

This repository must remain standalone. Do not add runtime imports, packages, shared deployment dependencies, cross-repo workflows or build dependencies on Pooya.OS, ORBIT, AuraPulse, NexusCanvas, LuminaGlass or any other portfolio project. Cross-links in portfolio documentation are allowed only if later approved. Do not reuse their source code without explicit user approval.

## Product and safety invariants

1. Never collect card numbers, bank credentials or real payment details.
2. Never call a real payment gateway or create real orders, reservations, shipments, inventory changes or payouts.
3. Label demo data, local fallback, illustrative currency and simulated outcomes honestly.
4. Do not scrape Amazon or copy its catalog. An Amazon adapter needs authorized access and review of current terms.
5. Preserve original product text and source provenance when translating.
6. Never expose translation, FX or other provider secrets in client JavaScript or repository files.
7. Persian is the default locale and must use real RTL layout. English is also supported.
8. All visible controls must work or have a clear disabled/unavailable state.
9. Keep owner attribution visible in the product and repository.
10. Do not mark roadmap work complete until it is implemented and verified.

## Current repository map

- app/page.tsx — interactive storefront, catalog view, cart, favorites and simulated checkout.
- app/globals.css — responsive marketplace visual system.
- app/layout.tsx — document metadata and global styles.
- lib/catalog/types.ts — normalized catalog contracts.
- lib/catalog/normalize.ts — record validation, normalization and deduplication.
- lib/catalog/dummyjson.ts — DummyJSON source adapter.
- lib/catalog/index.ts — catalog public exports.
- lib/pricing.ts — pure demo pricing calculation foundation.
- lib/translation/types.ts — provider contract and localization metadata.
- lib/translation/index.ts — safe localization orchestration with honest unavailable/error states.
- .github/workflows/deploy.yml — GitHub Pages build/deploy pipeline.
- ROADMAP.md — milestone acceptance checklist.
- docs/AI_HANDOFF.md — mandatory instructions for the next AI/developer.

## Current implementation status

### M01 — storefront
- Implemented: responsive storefront, Persian/English shell, search/filter/sort, quick view, favorites, recently viewed IDs, local cart, simulated checkout, illustrative FX, owner/source disclosure.
- Pending verification: successful Actions run, live Pages smoke test, keyboard/accessibility and responsive QA.

### M02 — catalog intelligence
- Implemented: typed product contract; normalization; required-field validation; numeric bounds; source IDs; import timestamp; duplicate suppression; separated DummyJSON adapter; translation provider interface; translation metadata and explicit unavailable/error states.
- Pending: actual provider integration, translation cache/retries, scheduled static catalog generation, tests, admin diagnostics.
- Important: imported descriptions are not yet automatically translated. Do not claim they are.

## Recent CI blocker and correction

The original deployment workflow used setup-node npm caching and npm ci, but the repository had no package-lock.json. GitHub Actions failed before dependency installation with “Dependencies lock file is not found”. The workflow has been changed to remove npm cache dependency and run npm install. Verify the newest run before claiming CI is fixed. Next engineering improvement: commit a generated package-lock.json and restore npm ci/cache for deterministic builds.

## Mandatory continuation protocol

For every new AI assistant or developer:
1. Read this file, ROADMAP.md, docs/AI_HANDOFF.md, docs/ARCHITECTURE.md, docs/DECISIONS.md and CHANGELOG.md.
2. Inspect current main tree, package.json, relevant source files and latest Actions run.
3. Continue from the earliest unchecked acceptance item; do not rewrite working features without reason.
4. Keep the repository independent and static Pages compatible unless an architecture decision is recorded.
5. Run npm install, npm run typecheck and npm run build; add and run relevant tests.
6. Check the actual diff and the deployed site before reporting completion.
7. Update ROADMAP.md, CHANGELOG.md and this file after verified work. Update package.json only at a release boundary.
8. State exact completed work, verification evidence, known limitations and next action.

## Immediate next actions

1. Verify the newest GitHub Actions run after the missing-lockfile workflow fix.
2. If build passes, verify the public GitHub Pages site and basePath assets.
3. Generate and commit package-lock.json; return CI to npm ci.
4. Add a concrete server-side translation provider adapter and scheduled catalog generation, preserving translation provenance.
5. Add unit tests for normalizeCatalog, localizeText and calculateDemoPrice.
