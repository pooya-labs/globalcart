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
