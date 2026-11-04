import { test, expect, type Page } from "@playwright/test";

test.beforeEach(async ({ page }) => {
  // Keep this regression deterministic and prove that the committed static
  // catalog is sufficient for route and locale behavior.
  await page.route("**/*", route => {
    const url = new URL(route.request().url());
    if (url.hostname === "127.0.0.1" || url.hostname === "localhost") {
      return route.continue();
    }
    return route.abort();
  });
  await page.goto("./");
  await expect(page.locator(".product-card").first()).toBeVisible();
  await expect(page.locator("main.store")).toHaveAttribute("data-catalog-loaded", "true", { timeout: 15_000 });
});

async function expectNoHorizontalOverflow(page: Page, routeName: string) {
  const dimensions = await page.evaluate(() => ({
    viewport: document.documentElement.clientWidth,
    document: document.documentElement.scrollWidth
  }));
  expect(
    dimensions.document,
    `${routeName} overflows horizontally: ${JSON.stringify(dimensions)}`
  ).toBeLessThanOrEqual(dimensions.viewport);
}

test("Persian RTL and English LTR persist across public routes without horizontal overflow", async ({ page }) => {
  // Storefront: start in Persian, switch to English, and verify persistence.
  const storefront = page.locator("main.store");
  await expect(storefront).toHaveAttribute("dir", "rtl");
  await expectNoHorizontalOverflow(page, "Persian storefront");

  await page.locator(".locale-button").click();
  await expect(storefront).toHaveAttribute("dir", "ltr");
  await expectNoHorizontalOverflow(page, "English storefront");

  await page.reload();
  await expect(storefront).toHaveAttribute("dir", "ltr");
  await expectNoHorizontalOverflow(page, "reloaded English storefront");

  // Product detail: the route inherits the selected locale and can change it.
  const productLink = page.locator(".product-card h3 a").first();
  const productTitle = await productLink.innerText();
  const productId = Number((await productLink.getAttribute("href"))?.split("/").filter(Boolean).pop());
  expect(productId).toBeGreaterThan(0);
  await productLink.click();
  await expect(page).toHaveURL(new RegExp("/product/" + productId + "/?$"));
  const product = page.locator("main.product-page");
  await expect(product).toBeVisible();
  await expect(product).toHaveAttribute("dir", "ltr");
  await expectNoHorizontalOverflow(page, "English product detail");

  await page.locator(".product-page-top button").click();
  await expect(product).toHaveAttribute("dir", "rtl");
  await expectNoHorizontalOverflow(page, "Persian product detail");

  await page.reload();
  await expect(product).toHaveAttribute("dir", "rtl");
  await expectNoHorizontalOverflow(page, "reloaded Persian product detail");

  // The product visit should appear in browser-local history, without a server.
  const historyUrl = new URL("../recently-viewed/", page.url()).toString();
  const historyResponse = await page.goto(historyUrl);
  expect(historyResponse?.status()).toBe(200);
  const history = page.locator("main.history-page");
  await expect(history).toHaveAttribute("dir", "rtl");
  await expect(history.locator(".history-card").first()).toBeVisible();
  await expectNoHorizontalOverflow(page, "Persian recently viewed");

  await page.locator(".history-topbar button").click();
  await expect(history).toHaveAttribute("dir", "ltr");
  await expectNoHorizontalOverflow(page, "English recently viewed");

  await page.reload();
  await expect(history).toHaveAttribute("dir", "ltr");
  await expectNoHorizontalOverflow(page, "reloaded English recently viewed");

  // Case study: the same persisted preference must drive its independent page.
  const caseStudyUrl = new URL("../case-study/", page.url()).toString();
  const caseStudyResponse = await page.goto(caseStudyUrl);
  expect(caseStudyResponse?.status()).toBe(200);
  const caseStudy = page.locator("main.case-study");
  await expect(caseStudy).toHaveAttribute("dir", "ltr");
  await expect(page.getByRole("heading", { level: 1 })).toContainText("A global marketplace");
  await expectNoHorizontalOverflow(page, "English case study");

  await page.getByRole("button", { name: "تغییر زبان به فارسی" }).click();
  await expect(caseStudy).toHaveAttribute("dir", "rtl");
  await expect(page.getByRole("heading", { level: 1 })).toContainText("یک بازار جهانی");
  await expectNoHorizontalOverflow(page, "Persian case study");

  // Returning to the storefront must preserve the last selected language.
  const homeUrl = new URL("../", page.url()).toString();
  const homeResponse = await page.goto(homeUrl);
  expect(homeResponse?.status()).toBe(200);
  await expect(storefront).toHaveAttribute("dir", "rtl");
  await expectNoHorizontalOverflow(page, "Persian storefront after route round-trip");
});
