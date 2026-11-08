# Public performance and release QA

## Latest fully captured six-report measurement — 2026-11-08 (post-release UI optimization)

- Public Lighthouse [run 42517087881](https://github.com/pooya-labs/globalcart/actions/runs/42517087881), attempt 2, audited the newly deployed site after PR [#24](https://github.com/pooya-labs/globalcart/pull/24). All six published route/mode reports passed the existing budgets; no thresholds changed. Attempt 1 received a transient HTTP 503 on the case-study mobile request and was retried; the successful attempt includes all six reports.
- Exact results (Lighthouse 13.6.0, GitHub-hosted runner):

| Route / mode | Performance | Accessibility | Best practices | SEO | FCP | LCP | TBT | CLS |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Case study — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.3 s | 0 ms | 0 |
| Case study — mobile | 99 | 100 | 100 | 100 | 0.9 s | 0.9 s | 121 ms | 0 |
| Home — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.5 s | 1 ms | 0 |
| Home — mobile | 96 | 100 | 100 | 100 | 1.1 s | 2.4 s | 174 ms | 0 |
| Product — desktop | 100 | 100 | 100 | 100 | 0.2 s | 0.5 s | 4 ms | 0 |
| Product — mobile | 96 | 100 | 100 | 100 | 1.0 s | 2.5 s | 130 ms | 0 |

- The unchanged gate still enforces mobile LCP ≤4.0 s / TBT ≤300 ms, desktop LCP ≤2.5 s / TBT ≤200 ms, CLS ≤0.10 and Accessibility / Best Practices / SEO ≥95.
- Compared with the preceding successful home-mobile measurement [run 42504858362](https://github.com/pooya-labs/globalcart/actions/runs/42504858362) (Performance 94 / LCP 2.47 s / TBT 224 ms), the new sample is 96 / 2.35 s / 174 ms. It is a better lab sample, but repeatability and causality are not established.
- Paired home-mobile network traces show total initial JavaScript transfer falling from **139,873 to 137,202 bytes** (~2.7 KB); uncompressed transfer fell from **462,771 to 449,250 bytes**. The page chunk decreased from 85,727 to 75,230 raw bytes. However, the shared `255-1dfe01161d17a1b9.js` chunk remains unchanged at 174,525 raw / 47,415 transferred bytes, and Lighthouse estimates exactly **21,737 bytes** of unused code in it in both traces. The lazy-overlay boundary is therefore real and useful, but the remaining shared chunk is the next performance target.
- The dialog regression verifies no drawer before the user opens the bag, then observes a new script request after interaction; focus enters the dialog and returns to its trigger after Escape. It passes in Chromium, Firefox and WebKit.
- These are GitHub-hosted lab measurements, not field Core Web Vitals. The owner separately confirmed real desktop/mobile review, actual platform screen-reader review and full human RTL/LTR review on 2026-11-08. Eligible normal-network/CDN field Core Web Vitals remain unmeasured and optional.

## Latest published gate — 2026-11-08 current main deployment

- The six-report public Lighthouse audit and enforced budget gate passed on [run 42498127131](https://github.com/pooya-labs/globalcart/actions/runs/42498127131), retry 2, against current main commit `5b7141c`. All six reports passed; no thresholds were relaxed.
- Exact successful retry report results (Lighthouse 13.6.0, GitHub-hosted runner):

| Route / mode | Performance | Accessibility | Best practices | SEO | FCP | LCP | TBT | CLS |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Case study — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.3 s | 0 ms | 0 |
| Case study — mobile | 99 | 100 | 100 | 100 | 0.9 s | 0.9 s | 90 ms | 0 |
| Home — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.6 s | 0 ms | 0 |
| Home — mobile | 95 | 100 | 100 | 100 | 1.0 s | 2.5 s | 200 ms | 0 |
| Product — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.5 s | 0 ms | 0 |
| Product — mobile | 96 | 100 | 100 | 100 | 1.0 s | 2.5 s | 140 ms | 0 |

- The unchanged gate enforces mobile LCP ≤4.0 s / TBT ≤300 ms, desktop LCP ≤2.5 s / TBT ≤200 ms, CLS ≤0.10 and Accessibility / Best Practices / SEO ≥95.
- The first sample of this same run measured home-mobile TBT at 353 ms and failed the budget; retry 2 measured 200 ms and passed. Do not treat the failed sample as the final result, but retain it as evidence of lab variance and continue to seek repeatable headroom. The budget remains unchanged.
- The successful raw `home-mobile.json` report recorded four long tasks: 154 ms in shared Next.js chunk `4bd1b696-c023c6e3521b1417.js`, 112 ms in the document, and 75 ms / 59 ms in chunk `493-fb878163f2eec184.js`. The `unused-javascript` audit estimated 21,762 bytes (~21.8 KiB) unused in that second chunk, with an estimated 60 ms saving. These are audit leads, not an instruction to remove the chunk wholesale.
- Follow-up: inspect the initial homepage hydration/shared chunk, evaluate safe code splitting or deferred loading for non-critical interactions, and re-run all six reports. Keep the existing thresholds, preserve the catalog/gallery and visible controls, and seek repeatable TBT headroom.
- These are GitHub-hosted lab measurements, not physical-device or field Core Web Vitals. Physical-device, actual screen-reader, full manual RTL/LTR and eligible normal-network/CDN evidence remain separate follow-up checks.

## Automated Lighthouse lab baseline

The `Public Lighthouse lab audit` workflow runs after a successful GitHub Pages deployment and can also be started manually or on its weekly schedule.

**Pinned tooling and environment**

- Lighthouse CLI: `13.6.0` (exact version, not a floating range).
- Node.js: 22; dependencies installed with the committed `package-lock.json`.
- Browser: the Chromium build installed through the repository's pinned Playwright dependency.
- Runner: GitHub-hosted Ubuntu; this is a repeatable lab environment, not an end-user device.

**Audited public routes**

| Route | Why it is included |
|---|---|
| `/` | Main storefront and initial discovery experience |
| `/product/1/` | Representative statically generated product-detail route |
| `/case-study/` | Portfolio page, architecture and database model |

Each route is audited in Lighthouse's default mobile mode and its desktop preset. The workflow saves the raw JSON reports as a GitHub Actions artifact and writes a compact summary to the run's job summary. It records Performance, Accessibility, Best Practices, SEO, FCP, LCP, TBT and CLS when available.

**Interpretation and limits**

- These are *lab* measurements from one GitHub-hosted runner. They are useful for repeatable comparisons and regressions, but network load, runner contention, source-image hosts and geographic routing can change results.
- A simulated mobile profile is not a physical phone. A desktop preset is not a substitute for a real desktop browser check.
- TBT is a lab diagnostic; it is not field INP. Do not report these results as CrUX or real-user Core Web Vitals.
- The workflow enforces the repository's documented six-report regression budgets through `scripts/check-lighthouse-budget.ts`. A passing budget is a lab regression guard, not proof of real-user performance or release readiness.
- The workflow does not upload reports to a third-party Lighthouse server and does not need a PSI API key.

## Recorded published lab baseline — 2026-11-03

The first baseline identified a mobile storefront bottleneck. The subsequent changes restored a single-column hero at mobile widths, added responsive image candidates, delayed the optional catalog revalidation until the browser is idle, and moved DM Sans/Vazirmatn to self-hosted `next/font` assets. Google Fonts render-blocking CSS was removed.

**Initial full run:** [Lighthouse 41604368513](https://github.com/pooya-labs/globalcart/actions/runs/41604368513), Lighthouse 13.6.0, public Pages routes, GitHub-hosted runner. The most recent follow-up measurement is recorded below.

| Published route / mode | Performance | Accessibility | Best practices | SEO | FCP | LCP | TBT | CLS |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Storefront — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.7 s | 10 ms | 0 |
| Storefront — mobile | 87 | 100 | 100 | 100 | 1.2 s | 3.2 s | 250 ms | 0 |
| Product 1 — desktop | 100 | 100 | 100 | 100 | 0.2 s | 0.6 s | 10 ms | 0 |
| Product 1 — mobile | 96 | 100 | 100 | 100 | 0.9 s | 2.6 s | 80 ms | 0 |
| Case study — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.3 s | 0 ms | 0 |
| Case study — mobile | 99 | 100 | 100 | 100 | 0.9 s | 0.9 s | 140 ms | 0 |

The initial storefront mobile audit (run 41601101775) recorded Performance 74, LCP 4.9 s and TBT 320 ms. The first full baseline above (run 41604368513) recorded 87, 3.2 s and 250 ms respectively, a 1.7 s (about 35%) lab LCP reduction. A later repeat measurement is listed below; repeated lab runs can vary and are not field measurements.

The final relevant CI set is [Browser E2E 41603301095](https://github.com/pooya-labs/globalcart/actions/runs/41603301095) (44/44 tests), [Pages 41603301049](https://github.com/pooya-labs/globalcart/actions/runs/41603301049) (production build and public-route smoke), and [Security 41603300997](https://github.com/pooya-labs/globalcart/actions/runs/41603300997).



### Latest follow-up audit — 2026-11-03

**Run:** [Public Lighthouse 41608446036](https://github.com/pooya-labs/globalcart/actions/runs/41608446036) · Lighthouse 13.6.0 · GitHub-hosted Chromium · public GitHub Pages routes. The run completed successfully and saved six raw JSON reports in the [globalcart-public-lighthouse artifact](https://github.com/pooya-labs/globalcart/actions/runs/41608446036).

| Published route / mode | Performance | Accessibility | Best practices | SEO | FCP | LCP | TBT | CLS |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Storefront — desktop | 100 | 100 | 100 | 100 | 0.4 s | 0.8 s | 10 ms | 0 |
| Storefront — mobile | 91 | 100 | 100 | 100 | 1.3 s | 3.3 s | 150 ms | 0 |
| Product 1 — desktop | 100 | 100 | 100 | 100 | 0.2 s | 0.7 s | 10 ms | 0 |
| Product 1 — mobile | 95 | 100 | 100 | 100 | 0.9 s | 2.9 s | 90 ms | 0 |
| Case study — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.3 s | 0 ms | 0 |
| Case study — mobile | 98 | 100 | 100 | 100 | 1.0 s | 1.0 s | 150 ms | 0 |

The repeat audit confirms 100 Accessibility, Best Practices and SEO across all six route/mode combinations. Mobile Performance is 91 for the storefront, 95 for product detail and 98 for the case study. Compared with the preceding run, individual mobile figures fluctuate slightly; do not treat a single lab run as field data or infer a real-user regression from the score alone. The homepage mobile LCP remains the hero image and its render delay is the dominant LCP subpart in this run. This is a measurement checkpoint, not a claim that manual release QA is complete.



### Latest published audit after route-scoped preconnects — 2026-11-03

**Run:** [Public Lighthouse 41613166735](https://github.com/pooya-labs/globalcart/actions/runs/41613166735) · Lighthouse CLI 13.6.0 · GitHub-hosted Chromium · published Pages routes. The run completed successfully and retained six raw JSON reports in the [globalcart-public-lighthouse artifact](https://github.com/pooya-labs/globalcart/actions/runs/41613166735).

| Published route / mode | Performance | Accessibility | Best practices | SEO | FCP | LCP | TBT | CLS |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Storefront — desktop | 99 | 100 | 100 | 100 | 0.4 s | 0.8 s | 10 ms | 0 |
| Storefront — mobile | 88 | 100 | 100 | 100 | 1.3 s | 3.2 s | 180 ms | 0 |
| Product 1 — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.6 s | 0 ms | 0 |
| Product 1 — mobile | 97 | 100 | 100 | 100 | 1.0 s | 2.6 s | 70 ms | 0 |
| Case study — desktop | 100 | 100 | 100 | 100 | 0.2 s | 0.2 s | 10 ms | 0 |
| Case study — mobile | 98 | 100 | 100 | 100 | 0.9 s | 0.9 s | 160 ms | 0 |

All six runs recorded 100 Accessibility, Best Practices and SEO. Mobile Performance is 88 for the storefront, 97 for product detail and 98 for the case study. The scores and timings vary across repeated runs; use them as comparative lab diagnostics, not field Core Web Vitals.

**Preconnect verification from the raw mobile reports:** the homepage has Unsplash and DummyJSON preconnects with no unused warnings; product detail has only the DummyJSON preconnect with no unused warning; the case study has no external image preconnects. Product 1's LCP remains the product image, discoverable in the initial HTML and eagerly loaded. Its reported LCP breakdown in this run is approximately 125 ms TTFB, 13 ms resource-load delay, 73 ms resource-load duration and 90 ms element-render delay. The preconnect is a resource-connection hint, not a guarantee of a fixed score increase.

**Browser regression:** [E2E 41612960428](https://github.com/pooya-labs/globalcart/actions/runs/41612960428) passed 70/70; TypeScript, unit tests, static export, Pages deployment/public smoke [41612960450](https://github.com/pooya-labs/globalcart/actions/runs/41612960450) and Security [41612960405](https://github.com/pooya-labs/globalcart/actions/runs/41612960405) passed. This still does not close physical-device, screen-reader, comprehensive manual RTL/LTR or field-delivery QA.


### Latest published audit — 2026-11-03 (post-documentation deployment)

**Run:** [Public Lighthouse 41613950109](https://github.com/pooya-labs/globalcart/actions/runs/41613950109) · Lighthouse CLI 13.6.0 · GitHub-hosted Chromium · public Pages routes. The run completed successfully and retained all six raw JSON reports.

| Published route / mode | Performance | Accessibility | Best practices | SEO | FCP | LCP | TBT | CLS |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Storefront — desktop | 100 | 100 | 100 | 100 | 0.4 s | 0.8 s | 0 ms | 0 |
| Storefront — mobile | 92 | 100 | 100 | 100 | 1.3 s | 3.3 s | 90 ms | 0 |
| Product 1 — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.6 s | 50 ms | 0 |
| Product 1 — mobile | 97 | 100 | 100 | 100 | 0.9 s | 2.6 s | 60 ms | 0 |
| Case study — desktop | 100 | 100 | 100 | 100 | 0.4 s | 0.4 s | 0 ms | 0 |
| Case study — mobile | 100 | 100 | 100 | 100 | 0.9 s | 0.9 s | 70 ms | 0 |

All six audits recorded 100 Accessibility, Best Practices and SEO. Desktop routes scored 100 Performance; mobile scores were 92 / 97 / 100 for storefront, product and case study. These are comparative lab results, not CrUX/field Core Web Vitals.

**Raw image-delivery findings:** estimated savings were 160 KiB on homepage desktop, 108 KiB on homepage mobile, 82 KiB on product desktop and 84 KiB on product mobile; case-study routes had no material candidates. The largest product-page candidate was a 1000×1000 alternate gallery image downloaded for a 54×54 thumbnail (about 63 KiB potential savings). Homepage/product merchandising candidates were mostly 300×300 catalog thumbnails displayed at 105–181 px. The homepage mobile hero used a 768×508 source for a 379×253 display and remains immediately discoverable, eager and high-priority. Its LCP breakdown was approximately 98 ms TTFB, 13 ms resource-load delay, 248 ms resource-load duration and 1,002 ms element-render delay.

**Why no speculative image rewrite was made:** the catalog provides 300×300 thumbnail.webp files alongside larger alternate images. The public [DummyJSON placeholder-image documentation](https://dummyjson.com/docs/image) documents generated placeholders, not transformations of existing cdn.dummyjson.com product assets. A working resize contract has not been established. Do not append undocumented width parameters, add a third-party proxy, substitute one thumbnail for every gallery angle or degrade previews. Revisit only through a source-supported transformer or controlled build-time derivative pipeline after its network, licensing, storage and maintenance costs are justified.

**Other verification:** route-scoped preconnects remain correctly scoped (homepage: Unsplash + DummyJSON; product: DummyJSON only; case study: neither), with no unused-preconnect warning. Browser E2E [41612960428](https://github.com/pooya-labs/globalcart/actions/runs/41612960428) passed 70/70; Pages [41613825451](https://github.com/pooya-labs/globalcart/actions/runs/41613825451) and Security [41613825512](https://github.com/pooya-labs/globalcart/actions/runs/41613825512) passed. This does not close physical-device, screen-reader, complete manual RTL/LTR or field-delivery QA.


### Latest audit on the canonical Pages host — 2026-11-03

**Run:** [Public Lighthouse 41618638260](https://github.com/pooya-labs/globalcart/actions/runs/41618638260) · Lighthouse 13.6.0 · GitHub-hosted Chromium · https://pooya-labs.github.io/globalcart/. The run succeeded against the renamed Pages domain and retained six raw JSON reports.

| Published route / mode | Performance | Accessibility | Best practices | SEO | FCP | LCP | TBT | CLS |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Storefront — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.7 s | 10 ms | 0 |
| Storefront — mobile | 91 | 100 | 100 | 100 | 1.2 s | 3.2 s | 150 ms | 0 |
| Product 1 — desktop | 100 | 100 | 100 | 100 | 0.2 s | 0.6 s | 10 ms | 0 |
| Product 1 — mobile | 94 | 100 | 100 | 100 | 0.9 s | 3.0 s | 100 ms | 0 |
| Case study — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.3 s | 0 ms | 0 |
| Case study — mobile | 99 | 100 | 100 | 100 | 0.9 s | 0.9 s | 110 ms | 0 |

All six audits recorded 100 Accessibility, Best Practices and SEO. The mobile Performance scores were home 91, product 94 and case study 99; all desktop route scores were 100. These are GitHub-hosted lab measurements only, not CrUX/field Core Web Vitals.

**Host migration verification:** the previous `pooya16pg-spec.github.io` Pages host returned 404 in the earlier Lighthouse runs [41616550190](https://github.com/pooya-labs/globalcart/actions/runs/41616550190) and [41616834727](https://github.com/pooya-labs/globalcart/actions/runs/41616834727). The workflow now targets `https://pooya-labs.github.io/globalcart`; this audit completed successfully on the new base, and Pages repository-dispatch deployment [41618569410](https://github.com/pooya-labs/globalcart/actions/runs/41618569410) passed public smoke.

**CI action migration:** the deployment workflow now uses checkout/setup-node v7 and upload-pages-artifact/deploy-pages v5; E2E/Lighthouse/screenshot/security workflows use Node 24-compatible checkout/setup-node/artifact actions. The later Pages deployment [41618842051](https://github.com/pooya-labs/globalcart/actions/runs/41618842051) and Security [41618841960](https://github.com/pooya-labs/globalcart/actions/runs/41618841960) both passed.

### Latest published audit on canonical Pages — 2026-11-03

**Run:** [Public Lighthouse 41619642994](https://github.com/pooya-labs/globalcart/actions/runs/41619642994) · Lighthouse 13.6.0 · GitHub-hosted Chromium · https://pooya-labs.github.io/globalcart/. The run succeeded against the renamed Pages host and uploaded six raw JSON reports.

| Published route / mode | Performance | Accessibility | Best practices | SEO | FCP | LCP | TBT | CLS |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Storefront — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.6 s | 0 ms | 0 |
| Storefront — mobile | 94 | 100 | 100 | 100 | 1.3 s | 3.0 s | 90 ms | 0 |
| Product 1 — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.6 s | 0 ms | 0 |
| Product 1 — mobile | 99 | 100 | 100 | 100 | 0.9 s | 2.3 s | 60 ms | 0 |
| Case study — desktop | 100 | 100 | 100 | 100 | 0.2 s | 0.2 s | 0 ms | 0 |
| Case study — mobile | 100 | 100 | 100 | 100 | 0.9 s | 0.9 s | 90 ms | 0 |

All six audits recorded 100 Accessibility, Best Practices and SEO. Mobile Performance is home 94, product 99 and case study 100; desktop Performance is 100 across all routes. These are lab measurements, not CrUX/field Core Web Vitals.

### Latest published audit — 2026-11-03 (canonical Pages, run 41620637413)

**Run:** [Public Lighthouse 41620637413](https://github.com/pooya-labs/globalcart/actions/runs/41620637413) · Lighthouse CLI 13.6.0 · GitHub-hosted Chromium · `https://pooya-labs.github.io/globalcart/`. The run passed and retained all six raw JSON reports in artifact [13223825714](https://github.com/pooya-labs/globalcart/actions/runs/41620637413).

| Published route / mode | Performance | Accessibility | Best practices | SEO | FCP | LCP | TBT | CLS |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Storefront — desktop | 100 | 100 | 100 | 100 | 0.4 s | 0.8 s | 0 ms | 0 |
| Storefront — mobile | 93 | 100 | 100 | 100 | 1.3 s | 3.2 s | 80 ms | 0 |
| Product 1 — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.7 s | 0 ms | 0.007 |
| Product 1 — mobile | 95 | 100 | 100 | 100 | 0.9 s | 2.9 s | 50 ms | 0 |
| Case study — desktop | 100 | 100 | 100 | 100 | 0.2 s | 0.2 s | 0 ms | 0 |
| Case study — mobile | 99 | 100 | 100 | 100 | 0.9 s | 0.9 s | 100 ms | 0 |

All six audits recorded 100 Accessibility, Best Practices and SEO. Desktop Performance is 100 on every route; mobile Performance is 93 / 95 / 99 for storefront, product and case study. A single desktop product audit reports CLS 0.007, still a very small lab movement; this is not a field CLS measurement.

**LCP request discovery:** the homepage hero `main#top > section.hero > div.hero-visual > img` is the mobile LCP element. It is visible in the initial HTML, uses `loading="eager"`, `fetchpriority="high"`, and a responsive `srcset`; Lighthouse's request-discovery checklist passes. Its LCP subparts were approximately 96 ms TTFB, 11 ms resource-load delay, 238 ms resource-load duration and 24 ms element-render delay. The product main image is likewise eager/high-priority and discoverable; its subparts were approximately 97 ms, 10 ms, 65 ms and 64 ms respectively. These numbers are subpart diagnostics from a lab run; do not infer that their sum must equal the displayed page-level LCP.

**Decision:** do not add another preload or rewrite the image CDN URLs based only on the 3.2 s storefront LCP. The LCP request is already discoverable and begins promptly; further work should target measured browser/rendering/network causes and must be compared across repeat runs. Existing image-size observations and the reason no undocumented DummyJSON resize/proxy was added remain documented above.

This remains lab-only evidence. It does not close physical-device, real screen-reader, comprehensive manual RTL/LTR or field-delivery QA.
### Latest canonical Pages audit and renamed-owner regression — 2026-11-03

**Lighthouse run:** [41622967095](https://github.com/pooya-labs/globalcart/actions/runs/41622967095) · Lighthouse 13.6.0 · GitHub-hosted Chromium · canonical Pages host. All six route/mode audits completed successfully.

| Published route / mode | Performance | Accessibility | Best practices | SEO | FCP | LCP | TBT | CLS |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Storefront — desktop | 100 | 100 | 100 | 100 | 0.4 s | 0.8 s | 0 ms | 0 |
| Storefront — mobile | 91 | 100 | 100 | 100 | 1.3 s | 3.2 s | 90 ms | 0 |
| Product 1 — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.6 s | 0 ms | 0 |
| Product 1 — mobile | 96 | 100 | 100 | 100 | 1.0 s | 2.7 s | 40 ms | 0 |
| Case study — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.3 s | 0 ms | 0 |
| Case study — mobile | 100 | 100 | 100 | 100 | 0.9 s | 0.9 s | 70 ms | 0 |

All six modes retained 100 Accessibility, Best Practices and SEO. Desktop Performance is 100 on all routes; mobile is 91 / 96 / 100 for storefront, product and case study.

**Lab variability:** four close-together published runs on the canonical host showed home-mobile Performance/LCP: [41620637413](https://github.com/pooya-labs/globalcart/actions/runs/41620637413) 93 / 3.2 s, [41622539031](https://github.com/pooya-labs/globalcart/actions/runs/41622539031) 80 / 4.7 s, [41622704373](https://github.com/pooya-labs/globalcart/actions/runs/41622704373) 87 / 3.4 s and [41622967095](https://github.com/pooya-labs/globalcart/actions/runs/41622967095) 91 / 3.2 s. Product mobile ranged 94–96 and case-study mobile 97–100. The repeated variance is consistent with runner/network sensitivity; do not treat the lowest run as a confirmed code regression or select only the best run as a representative field result.

**Canonical link regression:** Browser E2E [41622800540](https://github.com/pooya-labs/globalcart/actions/runs/41622800540) passed **74/74**. The added assertion validated homepage GitHub links, Case Study repository/ERD links and canonical metadata on Chromium desktop/mobile, Firefox and WebKit. Pages [41622800624](https://github.com/pooya-labs/globalcart/actions/runs/41622800624) and Security [41622800675](https://github.com/pooya-labs/globalcart/actions/runs/41622800675) passed.

Manual M09.4 remains open: physical desktop/mobile, a real platform screen-reader, comprehensive manual RTL/LTR, and normal-network/field-delivery evidence.
### Latest published canonical Pages audit — November 3, 2026 (run 41624263482)

**Run:** [Public Lighthouse 41624263482](https://github.com/pooya-labs/globalcart/actions/runs/41624263482) · Lighthouse 13.6.0 · GitHub-hosted Chromium · `https://pooya-labs.github.io/globalcart/`. All six route/mode audits completed successfully.

| Published route / mode | Performance | Accessibility | Best practices | SEO | FCP | LCP | TBT | CLS |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Storefront — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.7 s | 10 ms | 0 |
| Storefront — mobile | 91 | 100 | 100 | 100 | 1.2 s | 3.0 s | 200 ms | 0 |
| Product 1 — desktop | 100 | 100 | 100 | 100 | 0.2 s | 0.6 s | 10 ms | 0 |
| Product 1 — mobile | 96 | 100 | 100 | 100 | 0.9 s | 2.7 s | 90 ms | 0 |
| Case study — desktop | 100 | 100 | 100 | 100 | 0.2 s | 0.2 s | 10 ms | 0 |
| Case study — mobile | 97 | 100 | 100 | 100 | 0.9 s | 2.2 s | 160 ms | 0 |

All six audits recorded 100 Accessibility, Best Practices and SEO; all desktop routes scored 100 Performance. Compared with the immediately preceding published audit [41622967095](https://github.com/pooya-labs/globalcart/actions/runs/41622967095), case-study mobile varied from 100 / 0.9 s LCP / 70 ms TBT to 97 / 2.2 s / 160 ms. Treat this as a lab-run difference unless repeated, controlled evidence shows a reproducible application regression. These measurements are not CrUX or field Core Web Vitals.

**Verification lineage:** Browser E2E [41622800540](https://github.com/pooya-labs/globalcart/actions/runs/41622800540) passed 74/74 across Chromium desktop/mobile, Firefox and WebKit. Pages deployment/public smoke [41623961650](https://github.com/pooya-labs/globalcart/actions/runs/41623961650) and Security [41623961528](https://github.com/pooya-labs/globalcart/actions/runs/41623961528) passed. The subsequent mainline head update only persisted the generated catalog snapshot and lockfile; it did not change application behavior.

### Latest published canonical Pages audit — November 3, 2026 (run 41626880818)

**Run:** [Public Lighthouse 41626880818](https://github.com/pooya-labs/globalcart/actions/runs/41626880818) · Lighthouse 13.6.0 · GitHub-hosted Chromium · `https://pooya-labs.github.io/globalcart/`. All six route/mode audits completed successfully.

| Published route / mode | Performance | Accessibility | Best practices | SEO | FCP | LCP | TBT | CLS |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Storefront — desktop | 100 | 100 | 100 | 100 | 0.4 s | 0.7 s | 0 ms | 0 |
| Storefront — mobile | 93 | 100 | 100 | 100 | 1.3 s | 3.2 s | 40 ms | 0 |
| Product 1 — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.7 s | 0 ms | 0 |
| Product 1 — mobile | 94 | 100 | 100 | 100 | 0.9 s | 3.1 s | 40 ms | 0 |
| Case study — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.3 s | 0 ms | 0 |
| Case study — mobile | 100 | 100 | 100 | 100 | 0.9 s | 0.9 s | 60 ms | 0 |

All six audits report 100 Accessibility, Best Practices and SEO, and all desktop routes score 100 Performance. Compared with the prior audit [41624263482](https://github.com/pooya-labs/globalcart/actions/runs/41624263482), storefront mobile Performance improved from 91 to 93 and TBT from 200 ms to 40 ms; product mobile is 94 and case study mobile returned to 100. These are GitHub-runner lab measurements only, not CrUX or field Core Web Vitals.

**Verification lineage:** Browser E2E [41622800540](https://github.com/pooya-labs/globalcart/actions/runs/41622800540) passed 74/74 across Chromium desktop/mobile, Firefox and WebKit. Pages deployment/public smoke [41626708812](https://github.com/pooya-labs/globalcart/actions/runs/41626708812) and Security [41626708835](https://github.com/pooya-labs/globalcart/actions/runs/41626708835) passed. The current application implementation did not change in the documentation-only continuation.

### Latest completed published audit — November 3, 2026 (run 41631664623)

**Run:** [Public Lighthouse 41631664623](https://github.com/pooya-labs/globalcart/actions/runs/41631664623) · Lighthouse 13.6.0 · GitHub-hosted Chromium · `https://pooya-labs.github.io/globalcart/`. All six route/mode audits completed successfully.

| Published route / mode | Performance | Accessibility | Best practices | SEO | FCP | LCP | TBT | CLS |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Storefront — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.7 s | 10 ms | 0 |
| Storefront — mobile | 92 | 100 | 100 | 100 | 1.2 s | 3.1 s | 170 ms | 0 |
| Product 1 — desktop | 100 | 100 | 100 | 100 | 0.2 s | 0.6 s | 0 ms | 0 |
| Product 1 — mobile | 92 | 100 | 100 | 100 | 0.9 s | 3.0 s | 70 ms | 0 |
| Case study — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.3 s | 0 ms | 0 |
| Case study — mobile | 98 | 100 | 100 | 100 | 0.9 s | 0.9 s | 180 ms | 0 |

All six audits report 100 Accessibility, Best Practices and SEO; desktop Performance is 100 on every route. Compared with the previous completed audit [41631496522](https://github.com/pooya-labs/globalcart/actions/runs/41631496522), storefront mobile moved 91→92, product remained 94→92 and case study moved 99→98. These are GitHub-hosted lab results, not CrUX or field Core Web Vitals; one-run score variation is not evidence of a production regression.

**Verification lineage:** Backend CI [41632402711](https://github.com/pooya-labs/globalcart/actions/runs/41632402711) passed with uncached PostgreSQL integration tests and logical backup/restore. Security [41632402705](https://github.com/pooya-labs/globalcart/actions/runs/41632402705) and Pages deployment/public smoke [41632402717](https://github.com/pooya-labs/globalcart/actions/runs/41632402717) passed.

### Latest published canonical Pages audit — 2026-11-03 (run 41633195411)

**Run:** [Public Lighthouse 41633195411](https://github.com/pooya-labs/globalcart/actions/runs/41633195411) · Lighthouse CLI 13.6.0 · GitHub-hosted Chromium · `https://pooya-labs.github.io/globalcart/`. The workflow completed successfully and retained six raw Lighthouse JSON reports in the `globalcart-public-lighthouse` artifact.

| Published route / mode | Performance | Accessibility | Best practices | SEO | FCP | LCP | TBT | CLS |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Storefront — desktop | 100 | 100 | 100 | 100 | 0.4 s | 0.7 s | 10 ms | 0 |
| Storefront — mobile | 94 | 100 | 100 | 100 | 1.2 s | 2.8 s | 140 ms | 0 |
| Product 1 — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.7 s | 0 ms | 0 |
| Product 1 — mobile | 96 | 100 | 100 | 100 | 0.9 s | 2.6 s | 80 ms | 0 |
| Case study — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.3 s | 10 ms | 0 |
| Case study — mobile | 99 | 100 | 100 | 100 | 0.9 s | 0.9 s | 120 ms | 0 |

All six reports have no Lighthouse run warnings; Accessibility, Best Practices and SEO are 100 in every route/mode. Desktop Performance is 100 on all three routes. Mobile Performance is 94 / 96 / 99 for storefront, product and case study. CLS is 0 in all six reports. The improved mobile values relative to some immediately preceding runs are a useful lab observation, not proof of a sustained user-experienced gain: GitHub runner/network variation remains material.

**Interpretation:** this is a single GitHub-hosted lab sample using simulated mobile and desktop presets. TBT is not field INP, and these reports are not CrUX or real-user Core Web Vitals. The run does not satisfy normal-network delivery, physical-device, real screen-reader or comprehensive manual RTL/LTR QA.

### Latest published canonical Pages audit — 2026-11-03 (run 41652966386)

**Run:** [Public Lighthouse 41652966386](https://github.com/pooya-labs/globalcart/actions/runs/41652966386) · Lighthouse CLI 13.6.0 · GitHub-hosted Chromium · https://pooya-labs.github.io/globalcart/. All six published route/mode audits completed successfully; all six JSON reports were retained.

| Published route / mode | Performance | Accessibility | Best practices | SEO | FCP | LCP | TBT | CLS |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Storefront — desktop | 100 | 100 | 100 | 100 | 0.4 s | 0.8 s | 0 ms | 0 |
| Storefront — mobile | 91 | 100 | 100 | 100 | 1.2 s | 3.3 s | 70 ms | 0 |
| Product 1 — desktop | 100 | 100 | 100 | 100 | 0.2 s | 0.6 s | 0 ms | 0 |
| Product 1 — mobile | 94 | 100 | 100 | 100 | 0.9 s | 3.1 s | 60 ms | 0 |
| Case study — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.3 s | 0 ms | 0 |
| Case study — mobile | 99 | 100 | 100 | 100 | 1.0 s | 1.0 s | 100 ms | 0 |

Accessibility, Best Practices and SEO are 100 in every audit; CLS is zero throughout. All six reports passed the enforced budgets: mobile LCP ≤4.0 s / TBT ≤300 ms; desktop LCP ≤2.5 s / TBT ≤200 ms; CLS ≤0.10; Accessibility, Best Practices and SEO ≥95. The storefront remains the lowest mobile score, but all route LCP/TBT values are within the guardrails.

**Interpretation:** these are simulated GitHub-hosted lab diagnostics, not field Core Web Vitals. The earlier outlier [41645326322](https://github.com/pooya-labs/globalcart/actions/runs/41645326322) exceeded the mobile homepage LCP budget by 196 ms; a raw-report breakdown attributed most of that delay to hero element render delay, not image transfer. The 4.0 s mobile LCP budget remains intentionally unchanged. Revisit above-the-fold rendering during the owner-requested shell redesign rather than weakening the guard.

## Manual release checklist — owner-confirmed closeout

The owner confirmed completion of physical desktop/mobile, actual platform screen-reader and full human RTL/LTR reviews on 2026-11-08. The device/browser matrix was not supplied; this record reflects owner confirmation and does not invent specific hardware or test observations.

| Area | Required evidence | Status / notes |
|---|---|---|
| Physical desktop | Open the public storefront and case study on a real laptop/desktop browser; inspect normal and narrow windows | Owner-confirmed complete (2026-11-08; device matrix not recorded) |
| Physical mobile | Open the same routes on a real phone; inspect portrait and landscape, touch targets and overflow | Pending physical-device QA |
| Tablet hardware | Confirm both orientations and touch behavior on a real tablet; automated iPad emulation already passes | Automated iPad portrait/landscape passed in [41652758312](https://github.com/pooya-labs/globalcart/actions/runs/41652758312); physical tablet remains unverified |
| Persian RTL | Check navigation, search, filters, product detail, dialogs, cart simulation, case study and recently viewed | Owner-confirmed full human RTL/LTR pass complete (2026-11-08) |
| English LTR | Repeat the full route pass after changing language; reload and verify the preference persists | Automated route/locale regressions passed; full human visual pass pending |
| Keyboard | Verify logical tab order, visible focus, Escape-to-close, focus containment and focus restoration for dialogs/drawers | Playwright keyboard/focus regressions passed in the cross-browser suite; physical keyboard review pending |
| Screen reader | Complete real navigation and form/dialog interactions with a platform screen reader; report announcements and focus order | Owner-confirmed complete (2026-11-08; exact screen-reader not recorded) |
| Published delivery | Repeat Lighthouse from a normal network/location and capture the URL/date/method; distinguish lab from field data | GitHub-hosted lab audits and six-report guard passed; normal-network measurement remains optional and uncollected |
| Field Core Web Vitals | Record real-user/CrUX data only when the origin has eligible data; otherwise state that no field sample is available | No field sample claimed; optional until the origin has eligible data |

### Suggested manual route sequence

1. Storefront: load, switch language, use search and category filters, open a product, and inspect menu/focus behavior.
2. Product: inspect gallery, price disclosures, related items, comparison and local cart interactions.
3. Demo checkout: verify that the copy clearly says it is simulated and that no real payment is requested.
4. Recently viewed: open after visiting a product, verify persistence and the RTL/LTR direction.
5. Case study: change locale, inspect the architecture and database summaries, open the ERD link, and verify mobile layout.

### Release decision

The stable v1.0.0 portfolio/demo release is published. The owner confirmed physical-device, actual screen-reader and full-site RTL/LTR manual QA on 2026-11-08. Eligible normal-network/CDN field metrics remain unmeasured and optional. Automated axe and Playwright are regression protection, not a substitute for that separately owner-confirmed manual review.

The latest successful home-mobile Lighthouse retry measured TBT at 200 ms against a 300 ms budget; the first sample was 353 ms. Both are retained as lab-run variance evidence. The M11 backlog tracks the shared-chunk/document long tasks and unused JavaScript estimate for post-release optimization.
