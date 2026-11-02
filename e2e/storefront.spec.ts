import { test, expect } from "@playwright/test";

test.beforeEach(async ({ page }) => {
  // Keep the browser suite deterministic: the app must work from its committed
  // static catalog even when remote demo APIs, images or fonts are unavailable.
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

test("switches locale and opens a statically exported product detail page", async ({ page }) => {
  await expect(page.locator("main.store")).toHaveAttribute("dir", "rtl");
  await page.getByRole("button", { name: "EN", exact: true }).click();
  await expect(page.locator("main.store")).toHaveAttribute("dir", "ltr");
  await page.reload();
  await expect(page.locator("main.store")).toHaveAttribute("dir", "ltr");

  await page.locator(".product-card h3 a").filter({ hasText: "Essence Mascara Lash Princess" }).click();
  await expect(page).toHaveURL(/\/product\/1$/);
  await expect(page.locator("main.product-page")).toHaveAttribute("dir", "ltr");
  await expect(page.locator(".detail-information h1")).toHaveText("Essence Mascara Lash Princess");
  await expect(page.getByText("PRODUCT OVERVIEW")).toBeVisible();
});

test("shows the Wallex exchange candle and product provenance without a hard-coded FX fallback", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();

  const fxPanel = page.locator(".fx-control");
  await expect(fxPanel).toBeVisible();
  await expect(fxPanel).toContainText("Wallex · USDTTMN");

  const displayedRate = await fxPanel.locator("strong").innerText();
  if (displayedRate.includes("Rate unavailable")) {
    await expect(fxPanel).toContainText("Prices remain in USD until a valid exchange quote is available.");
  } else {
    await expect(fxPanel).toContainText("Toman / USDT");
    await expect(fxPanel).toContainText(/Fresh|Aging|Stale/);
  }

  await page.locator(".product-card h3 a").filter({ hasText: "Essence Mascara Lash Princess" }).click();
  await expect(page).toHaveURL(/\/product\/1\/?$/);
  // The selected storefront locale must persist across the product route.
  await expect(page.locator("main.product-page")).toHaveAttribute("dir", "ltr");
  await expect(page.locator(".detail-provenance")).toContainText("DummyJSON public demo API");
  await expect(page.locator(".detail-provenance")).toContainText("Source record ID");
  await expect(page.locator(".detail-provenance")).toContainText("Source sample rating; review count not supplied");

  const priceBreakdown = page.locator(".price-breakdown");
  await priceBreakdown.locator("summary").click();
  if (await priceBreakdown.getByText("Wallex one-minute candle close", { exact: false }).count()) {
    await expect(priceBreakdown).toContainText("Toman / USDT");
    await expect(priceBreakdown).toContainText("cash USD");
    await expect(priceBreakdown).toContainText("one-minute candle close");
  } else {
    await expect(priceBreakdown).toContainText("A valid exchange quote is currently unavailable");
  }
});

test("search suggestions support keyboard navigation and selection", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();
  const mobileSearchButton = page.getByRole("button", { name: "Search products, brands or categories...", exact: true });
  if (await mobileSearchButton.isVisible()) await mobileSearchButton.click();
  const search = page.getByRole("combobox", { name: "Search products, brands or categories..." });
  await search.fill("Esse");

  const listbox = page.getByRole("listbox", { name: "Search suggestions" });
  await expect(listbox).toBeVisible();
  const productOption = listbox.getByRole("option").filter({ hasText: "Essence Mascara Lash Princess" });
  await expect(productOption).toBeVisible();

  await search.press("ArrowDown");
  await expect(search).toHaveAttribute("aria-activedescendant", /option-/);
  await search.press("Enter");
  await expect(page.getByRole("dialog")).toBeVisible();
  await expect(page.locator("#quick-dialog-title")).toHaveText("Essence Mascara Lash Princess");
});

test("Ctrl+K command palette opens the local shopping bag", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();
  await page.keyboard.press("Control+k");

  const palette = page.getByRole("dialog", { name: "Quick commands" });
  await expect(palette).toBeVisible();
  await page.getByPlaceholder("What would you like to do?").fill("shopping bag");
  await page.getByPlaceholder("What would you like to do?").press("Enter");

  await expect(page.getByRole("heading", { name: /Your bag/ })).toBeVisible();
});

test("saved searches persist locally and can be applied and deleted", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();
  const search = page.getByRole("combobox", { name: "Search products, brands or categories..." });
  if (await page.getByRole("button", { name: "Search products, brands or categories..." }).isVisible()) {
    await page.getByRole("button", { name: "Search products, brands or categories..." }).click();
  }
  await search.fill("Essence");
  await page.locator("#catalog").scrollIntoViewIfNeeded();

  await page.getByRole("button", { name: "Save search" }).click();
  await page.getByLabel("Search name").fill("Essence picks");
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(page.getByRole("status").filter({ hasText: "Saved in this browser." })).toBeVisible();

  await page.reload();
  await expect(page.locator(".product-card").first()).toBeVisible();
  await page.getByRole("button", { name: "EN", exact: true }).click();
  await page.getByRole("button", { name: /Saved searches/ }).click();
  await expect(page.getByText("Essence picks")).toBeVisible();

  await page.getByRole("button", { name: "Apply" }).click();
  if (await page.getByRole("button", { name: "Search products, brands or categories..." }).isVisible()) {
    await page.getByRole("button", { name: "Search products, brands or categories..." }).click();
  }
  await expect(page.getByRole("combobox", { name: "Search products, brands or categories..." })).toHaveValue("Essence");

  await page.getByRole("button", { name: /Saved searches/ }).click();
  await page.getByRole("button", { name: "Delete Essence picks" }).click();
  await expect(page.getByText("Essence picks")).toHaveCount(0);
});

test("cart and checkout remain a clearly simulated local flow", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();
  const targetProduct = page.locator(".product-card").filter({ hasText: "Essence Mascara Lash Princess" });
  await targetProduct.scrollIntoViewIfNeeded();
  await targetProduct.locator(".product-price button").click();
  await page.getByRole("button", { name: /Your bag/ }).click();
  await expect(page.getByRole("dialog").getByText("Essence Mascara Lash Princess", { exact: true })).toBeVisible();

  await page.getByRole("button", { name: "Continue to demo checkout" }).click();
  await expect(page.getByRole("heading", { name: "Demo checkout" })).toBeVisible();
  await expect(page.locator(".checkout-warning")).toContainText("Real payments are disabled");
  await page.getByRole("button", { name: "Place demo order" }).click();
  await expect(page.getByRole("status").filter({ hasText: "Demo order created" })).toBeVisible();
});
