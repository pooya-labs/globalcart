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

**Latest full run:** [Lighthouse 41604368513](https://github.com/pooya16pg-spec/globalcart/actions/runs/41604368513), Lighthouse 13.6.0, public Pages routes, GitHub-hosted runner. Download the raw six-report artifact from the run's Artifacts section.

| Published route / mode | Performance | Accessibility | Best practices | SEO | FCP | LCP | TBT | CLS |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Storefront — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.7 s | 10 ms | 0 |
| Storefront — mobile | 87 | 100 | 100 | 100 | 1.2 s | 3.2 s | 250 ms | 0 |
| Product 1 — desktop | 100 | 100 | 100 | 100 | 0.2 s | 0.6 s | 10 ms | 0 |
| Product 1 — mobile | 96 | 100 | 100 | 100 | 0.9 s | 2.6 s | 80 ms | 0 |
| Case study — desktop | 100 | 100 | 100 | 100 | 0.3 s | 0.3 s | 0 ms | 0 |
| Case study — mobile | 99 | 100 | 100 | 100 | 0.9 s | 0.9 s | 140 ms | 0 |

The initial storefront mobile audit (run 41601101775) recorded Performance 74, LCP 4.9 s and TBT 320 ms. The latest run records 87, 3.2 s and 250 ms respectively. The 1.7 s LCP reduction is approximately 35%; these remain lab comparisons, not field measurements.

The final relevant CI set is [Browser E2E 41603301095](https://github.com/pooya16pg-spec/globalcart/actions/runs/41603301095) (44/44 tests), [Pages 41603301049](https://github.com/pooya16pg-spec/globalcart/actions/runs/41603301049) (production build and public-route smoke), and [Security 41603300997](https://github.com/pooya16pg-spec/globalcart/actions/runs/41603300997).

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
