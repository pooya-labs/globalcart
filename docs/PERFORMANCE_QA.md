# Public performance and release QA

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
- The workflow deliberately does not fail on an arbitrary Lighthouse score threshold before a stable baseline has been reviewed. Use repeated runs and inspect the raw reports before setting budgets.
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
## Manual release checklist — fill with actual observations

Do not tick these items based on CI emulation. Record actual device, browser/OS version, viewport, date and findings in the table before closing M09.4.

| Area | Required evidence | Status / notes |
|---|---|---|
| Physical desktop | Open the public storefront and case study on a real laptop/desktop browser; inspect normal and narrow windows | Pending owner QA |
| Physical mobile | Open the same routes on a real phone; inspect portrait and landscape, touch targets and overflow | Pending owner QA |
| Persian RTL | Check navigation, search, filters, product detail, dialogs, cart simulation, case study and recently viewed | Pending owner QA |
| English LTR | Repeat the full route pass after changing language; reload and verify the preference persists | Pending owner QA |
| Keyboard | Verify logical tab order, visible focus, Escape-to-close, focus containment and focus restoration for dialogs/drawers | Pending owner QA |
| Screen reader | Complete real navigation and form/dialog interactions with a platform screen reader; report announcements and focus order | Pending owner QA |
| Published delivery | Repeat Lighthouse from a normal network/location and capture the URL/date/method; distinguish lab from field data | Pending owner QA |
| Field Core Web Vitals | Record real-user/CrUX data only when the origin has eligible data; otherwise state that no field sample is available | Pending owner QA |

### Suggested manual route sequence

1. Storefront: load, switch language, use search and category filters, open a product, and inspect menu/focus behavior.
2. Product: inspect gallery, price disclosures, related items, comparison and local cart interactions.
3. Demo checkout: verify that the copy clearly says it is simulated and that no real payment is requested.
4. Recently viewed: open after visiting a product, verify persistence and the RTL/LTR direction.
5. Case study: change locale, inspect the architecture and database summaries, open the ERD link, and verify mobile layout.

### Release decision

M09.4 remains open until physical-device, actual screen-reader, full-site RTL/LTR and published-delivery/field-metric checks are recorded. Automated axe and Playwright coverage is regression protection, not evidence that those manual checks have happened.
