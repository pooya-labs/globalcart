import fs from "node:fs";
import path from "node:path";

const reportPaths = process.argv.slice(2).filter(file => fs.existsSync(file));
if (reportPaths.length === 0) {
  throw new Error("No Lighthouse JSON reports were supplied.");
}

const score = (report, key) => {
  const value = report.categories?.[key]?.score;
  return Number.isFinite(value) ? String(Math.round(value * 100)) : "—";
};

const metric = (report, key) => {
  const audit = report.audits?.[key];
  if (!audit || !Number.isFinite(audit.numericValue)) return "—";
  return audit.displayValue || String(Math.round(audit.numericValue));
};

const rows = reportPaths.map(file => {
  const report = JSON.parse(fs.readFileSync(file, "utf8"));
  return {
    page: path.basename(file, ".json"),
    lighthouse: report.lighthouseVersion ?? "unknown",
    performance: score(report, "performance"),
    accessibility: score(report, "accessibility"),
    bestPractices: score(report, "best-practices"),
    seo: score(report, "seo"),
    fcp: metric(report, "first-contentful-paint"),
    lcp: metric(report, "largest-contentful-paint"),
    tbt: metric(report, "total-blocking-time"),
    cls: metric(report, "cumulative-layout-shift")
  };
}).sort((a, b) => a.page.localeCompare(b.page));

const lines = [
  "## Public Lighthouse lab results",
  "",
  "| Route / mode | Lighthouse | Performance | Accessibility | Best practices | SEO | FCP | LCP | TBT | CLS |",
  "|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|",
  ...rows.map(row =>
    `| ${row.page} | ${row.lighthouse} | ${row.performance} | ${row.accessibility} | ${row.bestPractices} | ${row.seo} | ${row.fcp} | ${row.lcp} | ${row.tbt} | ${row.cls} |`
  ),
  "",
  "Scores are Lighthouse lab scores from a GitHub-hosted runner. They are comparative diagnostics, not real-user field metrics."
];

const markdown = lines.join("\n");
console.log(markdown);
if (process.env.GITHUB_STEP_SUMMARY) {
  fs.appendFileSync(process.env.GITHUB_STEP_SUMMARY, markdown + "\n");
}
