import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

type CategoryScore = { score?: number | null };
type AuditMetric = { numericValue?: number | null };
export type LighthouseReport = {
  categories?: Record<string, CategoryScore | undefined>;
  audits?: Record<string, AuditMetric | undefined>;
};

type Mode = "mobile" | "desktop";

const EXPECTED_REPORTS = [
  "home-mobile.json",
  "home-desktop.json",
  "product-1-mobile.json",
  "product-1-desktop.json",
  "case-study-mobile.json",
  "case-study-desktop.json"
];

const CATEGORY_MINIMUM = 0.95;
const LIMITS: Record<Mode, { lcpMs: number; tbtMs: number }> = {
  mobile: { lcpMs: 4_000, tbtMs: 300 },
  desktop: { lcpMs: 2_500, tbtMs: 200 }
};

export function validateLighthouseReportSet(reportPaths: string[]): string[] {
  const errors: string[] = [];
  const expected = new Set(EXPECTED_REPORTS);
  const seen = new Set<string>();

  for (const reportPath of reportPaths) {
    const name = path.basename(reportPath);
    if (!expected.has(name)) {
      errors.push(`Unexpected Lighthouse report: ${name}`);
      continue;
    }
    if (seen.has(name)) errors.push(`Duplicate Lighthouse report: ${name}`);
    seen.add(name);
  }

  for (const name of EXPECTED_REPORTS) {
    if (!seen.has(name)) errors.push(`Missing Lighthouse report: ${name}`);
  }
  return errors;
}

export function evaluateLighthouseReport(report: LighthouseReport, reportPath: string): string[] {
  const name = path.basename(reportPath);
  const match = /^(home|product-1|case-study)-(mobile|desktop)\.json$/.exec(name);
  if (!match) return [`Unsupported Lighthouse report name: ${name}`];

  const mode = match[2] as Mode;
  const limits = LIMITS[mode];
  const errors: string[] = [];

  for (const category of ["accessibility", "best-practices", "seo"]) {
    const score = report.categories?.[category]?.score;
    if (typeof score !== "number" || !Number.isFinite(score)) {
      errors.push(`${name}: missing or invalid ${category} score`);
    } else if (score < CATEGORY_MINIMUM) {
      errors.push(`${name}: ${category} score ${Math.round(score * 100)} is below 95`);
    }
  }

  const lcp = report.audits?.["largest-contentful-paint"]?.numericValue;
  if (typeof lcp !== "number" || !Number.isFinite(lcp)) {
    errors.push(`${name}: missing or invalid LCP numeric value`);
  } else if (lcp > limits.lcpMs) {
    errors.push(`${name}: LCP ${Math.round(lcp)} ms exceeds ${limits.lcpMs} ms ${mode} budget`);
  }

  const tbt = report.audits?.["total-blocking-time"]?.numericValue;
  if (typeof tbt !== "number" || !Number.isFinite(tbt)) {
    errors.push(`${name}: missing or invalid TBT numeric value`);
  } else if (tbt > limits.tbtMs) {
    errors.push(`${name}: TBT ${Math.round(tbt)} ms exceeds ${limits.tbtMs} ms ${mode} budget`);
  }

  const cls = report.audits?.["cumulative-layout-shift"]?.numericValue;
  if (typeof cls !== "number" || !Number.isFinite(cls)) {
    errors.push(`${name}: missing or invalid CLS numeric value`);
  } else if (cls > 0.1) {
    errors.push(`${name}: CLS ${cls.toFixed(3)} exceeds 0.100 budget`);
  }

  return errors;
}

export function inspectLighthouseFiles(reportPaths: string[]): string[] {
  const errors = validateLighthouseReportSet(reportPaths);
  if (reportPaths.length === 0) return [...errors, "No Lighthouse JSON reports were supplied."];

  for (const reportPath of reportPaths) {
    try {
      const report = JSON.parse(fs.readFileSync(reportPath, "utf8")) as LighthouseReport;
      errors.push(...evaluateLighthouseReport(report, reportPath));
    } catch (error) {
      const detail = error instanceof Error ? error.message : String(error);
      errors.push(`${path.basename(reportPath)}: could not parse report JSON: ${detail}`);
    }
  }
  return errors;
}

function main() {
  const reportPaths = process.argv.slice(2);
  const errors = inspectLighthouseFiles(reportPaths);
  if (errors.length > 0) {
    console.error("Lighthouse lab budgets failed:");
    for (const error of errors) console.error(`- ${error}`);
    process.exitCode = 1;
    return;
  }

  console.log("Lighthouse lab budgets passed for all six published route/mode reports.");
  console.log("Budgets: mobile LCP ≤4.0 s / TBT ≤300 ms; desktop LCP ≤2.5 s / TBT ≤200 ms; CLS ≤0.10; Accessibility, Best Practices and SEO ≥95.");
  console.log("These are CI lab regression guardrails, not field Core Web Vitals.");
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  main();
}
