import { test, expect } from "@playwright/test";
import fxSnapshot from "../data/fx-rate.json";
import { getFxRateStatus, type FxRateSnapshot } from "../lib/fx/tgju";

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

test("labels date-only TGJU quotes as stale on storefront and product detail", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();

  const expectedStatus = getFxRateStatus(fxSnapshot as unknown as FxRateSnapshot);
  const statusLabels = {
    fresh: "Fresh",
    aging: "Aging",
    stale: "Stale",
    unavailable: "Rate unavailable"
  } as const;

  const fxPanel = page.locator(".fx-control");
  await expect(fxPanel).toHaveAttribute("data-status", expectedStatus);
  await expect(fxPanel).toContainText(statusLabels[expectedStatus]);
  if (fxSnapshot.sourceTimeKind === "date_only" && fxSnapshot.sourceTimeTehran) {
    await expect(fxPanel).toContainText(fxSnapshot.sourceTimeTehran);
    expect(expectedStatus).toBe("stale");
  }

  await page.locator(".product-card h3 a")
    .filter({ hasText: "Essence Mascara Lash Princess" })
    .click();
  await expect(page).toHaveURL(/\/product\/1\/?$/);

  const breakdown = page.locator(".price-breakdown");
  await breakdown.locator("summary").click();
  if (expectedStatus === "fresh") {
    await expect(breakdown.locator(".fx-warning")).toHaveCount(0);
  } else {
    await expect(breakdown.locator(".fx-warning"))
      .toContainText(`FX freshness warning: ${statusLabels[expectedStatus]}`);
  }
  await expect(breakdown).toContainText("Toman / USD");
});

test("shows the TGJU free-market dollar quote and product provenance without a hard-coded FX fallback", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();

  const fxPanel = page.locator(".fx-control");
  await expect(fxPanel).toBeVisible();
  await expect(fxPanel).toContainText("TGJU · free market");

  const displayedRate = await fxPanel.locator("strong").innerText();
  if (displayedRate.includes("Rate unavailable")) {
    await expect(fxPanel).toContainText("Prices remain in USD until a valid exchange quote is available.");
  } else {
    await expect(fxPanel).toContainText("Toman / USD");
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
  if (await priceBreakdown.getByText("TGJU free-market USD rate", { exact: false }).count()) {
    await expect(priceBreakdown).toContainText("Toman / USD");
    await expect(priceBreakdown).toContainText("Rial to Toman");
    await expect(priceBreakdown).toContainText("TGJU");
  } else {
    await expect(priceBreakdown).toContainText("A valid TGJU dollar quote is currently unavailable");
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
  await expect(page.locator("main.store")).toHaveAttribute("dir", "ltr");
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

test("recently viewed route inherits and persists the selected locale", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();
  await expect(page.locator("main.store")).toHaveAttribute("dir", "ltr");

  await page.locator(".product-card h3 a")
    .filter({ hasText: "Essence Mascara Lash Princess" })
    .click();
  await expect(page).toHaveURL(/product/);
  await expect(page.locator("main.product-page")).toHaveAttribute("dir", "ltr");
  await expect.poll(() => page.evaluate(() => {
    try {
      const state = JSON.parse(localStorage.getItem("globalcart-state-v1") ?? "{}");
      return Array.isArray(state.recent) && state.recent.includes(1);
    } catch {
      return false;
    }
  })).toBe(true);

  const historyUrl = new URL("../recently-viewed", page.url()).toString();
  await page.goto(historyUrl);
  const history = page.locator("main.history-page");
  await expect(history).toHaveAttribute("dir", "ltr");
  await expect(history.getByRole("heading", { name: "Recently viewed products" })).toBeVisible();
  await expect(history.locator(".history-note")).toContainText("TGJU’s free-market USD quote");
  await expect(history.getByRole("link", { name: /Essence Mascara Lash Princess/ })).toBeVisible();

  await page.getByRole("button", { name: "Switch language to Persian" }).click();
  await expect(history).toHaveAttribute("dir", "rtl");
  await expect(page.getByRole("button", { name: "تغییر زبان به انگلیسی" })).toBeVisible();
  await expect(history.getByRole("heading", { name: "محصولات بازدیدشده" })).toBeVisible();

  await page.reload();
  await expect(history).toHaveAttribute("dir", "rtl");
  await page.getByRole("link", { name: /بازگشت به فروشگاه/ }).click();
  await expect(page.locator("main.store")).toHaveAttribute("dir", "rtl");
});


test("active catalog filter chips remove one facet or clear the complete filter set", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();
  const isMobile = await page.evaluate(() => window.matchMedia("(max-width: 760px)").matches);
  if (isMobile) await page.getByRole("button", { name: "Filters" }).click();

  const brand = page.locator(".facet-select-group select").first();
  await brand.selectOption({ index: 1 });
  if (isMobile) await page.getByRole("button", { name: "Close filters" }).click();
  await expect(page.locator(".active-filter-bar")).toBeVisible();
  await expect(page.locator(".active-filter-chip")).toContainText(await brand.locator("option:checked").textContent() ?? "");

  await page.locator(".active-filter-chip").first().click();
  await expect(page.locator(".active-filter-bar")).toBeHidden();
  if (isMobile) await page.getByRole("button", { name: "Filters" }).click();
  await brand.selectOption({ index: 1 });
  await page.locator(".facet-select-group select").nth(1).selectOption("4");
  if (isMobile) await page.getByRole("button", { name: "Close filters" }).click();
  await expect(page.locator(".active-filter-chip")).toHaveCount(2);
  await page.getByRole("button", { name: "Clear all" }).click();
  await expect(page.locator(".active-filter-bar")).toBeHidden();
  if (isMobile) await page.getByRole("button", { name: "Filters" }).click();
  await expect(brand).toHaveValue("");
  await expect(page.locator(".facet-select-group select").nth(1)).toHaveValue("0");
});


test("mobile filter sheet has a clear close action", async ({ page }) => {
  const isMobile = await page.evaluate(() => window.matchMedia("(max-width: 760px)").matches);
  test.skip(!isMobile, "The filter sheet is a mobile-only interaction.");

  await page.getByRole("button", { name: "EN", exact: true }).click();
  await page.getByRole("button", { name: "Filters" }).click();
  const panel = page.locator(".filter-panel.filter-visible");
  await expect(panel).toBeVisible();
  await page.getByRole("button", { name: "Close filters" }).click();
  await expect(page.locator(".filter-panel.filter-visible")).toHaveCount(0);
});


test("cart drawer supports undo after removing an item", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();
  await page.locator(".product-card").first().locator(".product-price button").click();
  await page.locator(".bag-button").click();
  const drawer = page.locator(".side-drawer[role='dialog']");
  await expect(drawer.locator(".drawer-item")).toHaveCount(1);
  await drawer.getByRole("button", { name: "Remove item" }).click();
  await expect(drawer.locator(".drawer-item")).toHaveCount(0);
  await page.getByRole("button", { name: "Undo" }).click();
  await expect(drawer.locator(".drawer-item")).toHaveCount(1);
});
