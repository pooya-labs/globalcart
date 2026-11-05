import test from "node:test";
import assert from "node:assert/strict";
import {
  evaluateLighthouseReport,
  validateLighthouseReportSet,
  type LighthouseReport
} from "../scripts/check-lighthouse-budget";

const report = (mode: "mobile" | "desktop", overrides: Partial<LighthouseReport> = {}): LighthouseReport => ({
  categories: {
    accessibility: { score: 1 },
    "best-practices": { score: 1 },
    seo: { score: 1 }
  },
  audits: {
    "largest-contentful-paint": { numericValue: mode === "mobile" ? 3_900 : 2_400 },
    "total-blocking-time": { numericValue: mode === "mobile" ? 290 : 190 },
    "cumulative-layout-shift": { numericValue: 0.05 }
  },
  ...overrides
});

test("accepts mobile and desktop reports at the conservative budget boundary", () => {
  assert.deepEqual(evaluateLighthouseReport(report("mobile"), "home-mobile.json"), []);
  assert.deepEqual(evaluateLighthouseReport(report("desktop"), "home-desktop.json"), []);
});

test("rejects LCP, TBT and CLS values above their respective mode budgets", () => {
  const result = report("mobile", {
    audits: {
      "largest-contentful-paint": { numericValue: 4_001 },
      "total-blocking-time": { numericValue: 301 },
      "cumulative-layout-shift": { numericValue: 0.101 }
    }
  });
  const errors = evaluateLighthouseReport(result, "product-1554130194-mobile.json");
  assert.equal(errors.length, 3);
  assert.match(errors.join("\n"), /LCP 4001 ms exceeds 4000 ms mobile budget/);
  assert.match(errors.join("\n"), /TBT 301 ms exceeds 300 ms mobile budget/);
  assert.match(errors.join("\n"), /CLS 0.101 exceeds 0.100 budget/);
});

test("uses stricter desktop LCP and TBT limits", () => {
  const result = report("desktop", {
    audits: {
      "largest-contentful-paint": { numericValue: 2_501 },
      "total-blocking-time": { numericValue: 201 },
      "cumulative-layout-shift": { numericValue: 0 }
    }
  });
  const errors = evaluateLighthouseReport(result, "case-study-desktop.json");
  assert.equal(errors.length, 2);
  assert.match(errors.join("\n"), /LCP 2501 ms exceeds 2500 ms desktop budget/);
  assert.match(errors.join("\n"), /TBT 201 ms exceeds 200 ms desktop budget/);
});

test("requires strong automated Accessibility, Best Practices and SEO scores", () => {
  const result = report("mobile", {
    categories: {
      accessibility: { score: 0.94 },
      "best-practices": { score: 0.95 },
      seo: { score: 0.8 }
    }
  });
  const errors = evaluateLighthouseReport(result, "home-mobile.json");
  assert.equal(errors.length, 2);
  assert.match(errors.join("\n"), /accessibility score 94 is below 95/);
  assert.match(errors.join("\n"), /seo score 80 is below 95/);
});

test("fails closed when an essential metric or category is missing", () => {
  const errors = evaluateLighthouseReport({}, "home-mobile.json");
  assert.equal(errors.length, 6);
  assert.match(errors.join("\n"), /missing or invalid accessibility score/);
  assert.match(errors.join("\n"), /missing or invalid LCP numeric value/);
  assert.match(errors.join("\n"), /missing or invalid CLS numeric value/);
});

test("requires the complete six-route desktop/mobile report matrix", () => {
  const all = [
    "home-mobile.json", "home-desktop.json",
    "product-1554130194-mobile.json", "product-1554130194-desktop.json",
    "case-study-mobile.json", "case-study-desktop.json"
  ];
  assert.deepEqual(validateLighthouseReportSet(all), []);
  assert.match(validateLighthouseReportSet(all.slice(0, 5)).join("\n"), /Missing Lighthouse report: case-study-desktop.json/);
  assert.match(validateLighthouseReportSet([...all, all[0]]).join("\n"), /Duplicate Lighthouse report: home-mobile.json/);
  assert.match(validateLighthouseReportSet(["unknown-mobile.json"]).join("\n"), /Unexpected Lighthouse report: unknown-mobile.json/);
});

test("requires one consistent numeric product ID across mobile and desktop reports", () => {
  const all = [
    "home-mobile.json", "home-desktop.json",
    "product-1554130194-mobile.json", "product-1554130194-desktop.json",
    "case-study-mobile.json", "case-study-desktop.json"
  ];
  assert.deepEqual(validateLighthouseReportSet(all), []);
  const mixed = all.map(name => name === "product-1554130194-desktop.json"
    ? "product-719281743-desktop.json"
    : name);
  assert.match(validateLighthouseReportSet(mixed).join("\n"), /Inconsistent Lighthouse product route IDs/);
});

test("rejects unrecognized report filenames", () => {
  assert.match(
    evaluateLighthouseReport(report("mobile"), "unknown-mobile.json").join("\n"),
    /Unsupported Lighthouse report name/
  );
});
