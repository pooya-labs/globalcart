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

async function expectDialogFocusLoop(page: Page, dialog: ReturnType<Page["locator"]>) {
  const buttons = dialog.locator("button:visible:not([disabled])");
  await expect(buttons.first()).toBeVisible();
  const first = buttons.first();
  const last = buttons.last();

  await expect(first).toBeFocused();
  await last.focus();
  await page.keyboard.press("Tab");
  await expect(first).toBeFocused();

  await page.keyboard.press("Shift+Tab");
  await expect(last).toBeFocused();
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

test("command palette traps keyboard focus and restores the invoking search field", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();

  const mobileSearchTrigger = page.getByRole("button", {
    name: "Search products, brands or categories..."
  });
  if (await mobileSearchTrigger.isVisible()) {
    await mobileSearchTrigger.click();
  }

  const search = page.getByRole("combobox", {
    name: "Search products, brands or categories..."
  });
  await search.focus();
  await expect(search).toBeFocused();

  await page.keyboard.press("Control+k");
  const dialog = page.getByRole("dialog", { name: "Quick commands" });
  await expect(dialog).toBeVisible();

  const firstFocusable = dialog.locator(".command-palette-heading .icon-button");
  const lastOption = dialog.getByRole("option").last();
  await lastOption.focus();
  await page.keyboard.press("Tab");
  await expect(firstFocusable).toBeFocused();

  await page.keyboard.press("Shift+Tab");
  await expect(lastOption).toBeFocused();

  await page.keyboard.press("Escape");
  await expect(dialog).toBeHidden();
  await expect(search).toBeFocused();
});


test("shopping drawers and modals trap focus and restore their opener", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();

  const cartTrigger = page.locator(".bag-button");
  await cartTrigger.focus();
  await cartTrigger.click();
  let dialog = page.locator(".side-drawer[role='dialog']");
  await expect(dialog).toBeVisible();
  await expectDialogFocusLoop(page, dialog);
  await page.keyboard.press("Escape");
  await expect(dialog).toBeHidden();
  await expect(cartTrigger).toBeFocused();

  const favoritesTrigger = page.locator(".favorite-trigger");
  let favoritesReturnTarget = cartTrigger;
  if (await favoritesTrigger.isVisible()) {
    favoritesReturnTarget = favoritesTrigger;
    await favoritesTrigger.focus();
    await favoritesTrigger.click();
  } else {
    // The mobile header hides the dedicated icon; use the visible command-palette action.
    await cartTrigger.focus();
    await page.keyboard.press("Control+k");
    const palette = page.getByRole("dialog", { name: "Quick commands" });
    await expect(palette).toBeVisible();
    await page.getByRole("option", { name: /Open saved items/ }).click();
  }
  dialog = page.locator(".side-drawer[role='dialog']");
  await expect(dialog).toBeVisible();
  await expectDialogFocusLoop(page, dialog);
  await page.keyboard.press("Escape");
  await expect(dialog).toBeHidden();
  await expect(favoritesReturnTarget).toBeFocused();

  const firstProduct = page.locator(".product-card").first();
  const quickViewTrigger = firstProduct.locator(".quick-view");
  const isMobileViewport = await page.evaluate(() => window.matchMedia("(max-width: 760px)").matches);
  if (!isMobileViewport) await firstProduct.hover();
  await quickViewTrigger.click();
  dialog = page.locator(".quick-modal[role='dialog']");
  await expect(dialog).toBeVisible();
  await expectDialogFocusLoop(page, dialog);
  await page.keyboard.press("Escape");
  await expect(dialog).toBeHidden();
  await expect(quickViewTrigger).toBeFocused();

  const productCards = page.locator(".product-card");
  await productCards.nth(0).locator(".compare-card-button").click();
  await productCards.nth(1).locator(".compare-card-button").click();
  const compareTrigger = page.locator(".compare-tray-open");
  await expect(compareTrigger).toBeEnabled();
  await compareTrigger.focus();
  await compareTrigger.click();
  dialog = page.locator(".compare-modal[role='dialog']");
  await expect(dialog).toBeVisible();
  await expectDialogFocusLoop(page, dialog);
  await page.keyboard.press("Escape");
  await expect(dialog).toBeHidden();
  await expect(compareTrigger).toBeFocused();
});

test("checkout dialog keeps focus contained and returns focus to the cart trigger", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();

  await page.locator(".product-card").first().locator(".product-price button").click();
  const cartTrigger = page.locator(".bag-button");
  await cartTrigger.focus();
  await cartTrigger.click();

  const checkoutTrigger = page.locator(".drawer-bottom .full-button");
  await expect(checkoutTrigger).toBeVisible();
  await checkoutTrigger.click();

  const dialog = page.locator(".checkout-modal[role='dialog']");
  await expect(dialog).toBeVisible();
  await expectDialogFocusLoop(page, dialog);
  await page.keyboard.press("Escape");
  await expect(dialog).toBeHidden();
  await expect(cartTrigger).toBeFocused();
});
