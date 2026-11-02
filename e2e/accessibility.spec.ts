import AxeBuilder from "@axe-core/playwright";
import { test, expect, type Page } from "@playwright/test";

const wcagTags = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"];

test.beforeEach(async ({ page }) => {
  // Audit only the app itself. The committed static catalog must work without
  // remote APIs, fonts or product image hosts.
  await page.route("**/*", route => {
    const url = new URL(route.request().url());
    if (url.hostname === "127.0.0.1" || url.hostname === "localhost") {
      return route.continue();
    }
    return route.abort();
  });
  await page.goto("./");
  await expect(page.locator(".product-card").first()).toBeVisible();
});

async function expectNoWcagViolations(page: Page) {
  const results = await new AxeBuilder({ page }).withTags(wcagTags).analyze();
  const violations = results.violations.flatMap(violation =>
    violation.nodes.map(node => ({
      rule: violation.id,
      impact: violation.impact,
      help: violation.help,
      target: node.target,
      summary: node.failureSummary
    }))
  );
  expect(violations, JSON.stringify(violations, null, 2)).toEqual([]);
}

test("storefront homepage meets WCAG 2.1 AA automated checks", async ({ page }) => {
  await expectNoWcagViolations(page);
});

test("command palette dialog remains accessible when open", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();
  await page.keyboard.press("Control+k");
  await expect(page.getByRole("dialog", { name: "Quick commands" })).toBeVisible();
  await expectNoWcagViolations(page);
});

test("static product detail page meets WCAG 2.1 AA automated checks", async ({ page }) => {
  await page.locator(".product-card h3 a").filter({ hasText: "Essence Mascara Lash Princess" }).click();
  await expect(page.locator(".detail-information h1")).toHaveText("Essence Mascara Lash Princess");
  await expectNoWcagViolations(page);
});
