import AxeBuilder from "@axe-core/playwright";
import { test, expect, type Page } from "@playwright/test";
import { getPersianProductCopy } from "../lib/catalog/editorial";
const wcagTags = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"];

async function loadCatalogForTest(page: Page) {
  await page.locator(".catalog-section").scrollIntoViewIfNeeded();
  await expect(page.locator("main.store")).toHaveAttribute("data-catalog-loaded", "true", { timeout: 15_000 });
  await page.evaluate(() => {
    const root = document.documentElement;
    const previousBehavior = root.style.scrollBehavior;
    root.style.scrollBehavior = "auto";
    window.scrollTo(0, 0);
    root.style.scrollBehavior = previousBehavior;
  });
}

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
  await loadCatalogForTest(page);
});

async function expectNoWcagViolations(page: Page) {
  // AxeBuilder resolves Page through its playwright-core peer while the test fixture
  // can resolve a nested copy. The runtime page is the same Playwright API; cast only
  // at this library boundary so duplicate declaration files do not block typecheck.
  type AxePage = ConstructorParameters<typeof AxeBuilder>[0]["page"];
  const results = await new AxeBuilder({ page: page as unknown as AxePage }).withTags(wcagTags).analyze();
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


test("quick command palette loads on demand and preserves keyboard focus", async ({ page }) => {
  const trigger = page.locator(".locale-button");
  const initialScriptChunks = await page.evaluate(() =>
    performance.getEntriesByType("resource")
      .map(entry => entry.name)
      .filter(url => url.includes("/_next/static/chunks/") && url.endsWith(".js"))
  );
  await trigger.focus();
  await expect(page.locator(".command-palette")).toHaveCount(0);

  await page.keyboard.press("Control+k");
  const dialog = page.getByRole("dialog", { name: "فرمان‌های سریع" });
  await expect(dialog).toBeVisible();
  await expect.poll(() => page.evaluate(initial =>
    performance.getEntriesByType("resource").some(entry =>
      entry.name.includes("/_next/static/chunks/") &&
      entry.name.endsWith(".js") &&
      !initial.includes(entry.name)
    ), initialScriptChunks
  )).toBe(true);
  const input = dialog.getByRole("combobox");
  await expect(input).toBeFocused();

  await page.keyboard.press("ArrowDown");
  await expect(dialog.getByRole("option", { name: /نمایش دسته‌بندی‌ها/ })).toHaveAttribute("aria-selected", "true");
  await page.keyboard.press("Enter");
  await expect(dialog).toHaveCount(0);
  await expect(page.locator("#category-menu-panel")).toBeVisible();
  await expect(trigger).toBeFocused();

  await page.keyboard.press("Escape");
  await expect(page.locator("#category-menu-panel")).toHaveCount(0);
  await trigger.focus();
  await page.keyboard.press("Control+k");
  const reopened = page.getByRole("dialog", { name: "فرمان‌های سریع" });
  await expect(reopened).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(reopened).toHaveCount(0);
  await expect(trigger).toBeFocused();
});

test("storefront homepage meets WCAG checks in RTL and LTR", async ({ page }) => {
  // Two axe scans (RTL + LTR) can take over 30 seconds on WebKit CI runners.
  test.setTimeout(60_000);
  await expect(page.locator("main.store")).toHaveAttribute("dir", "rtl");
  await expectNoWcagViolations(page);

  await page.getByRole("button", { name: "EN", exact: true }).click();
  await expect(page.locator("main.store")).toHaveAttribute("dir", "ltr");
  await expectNoWcagViolations(page);
});

test("bilingual case study is navigable and meets WCAG checks in RTL and LTR", async ({ page }) => {
  const response = await page.goto("./case-study/");
  expect(response?.status()).toBe(200);

  const main = page.locator("main.case-study");
  await expect(main).toHaveAttribute("dir", "rtl");
  await expect(page.getByRole("heading", { level: 1 })).toContainText("یک بازار جهانی");
  await expect(page.getByRole("group", { name: "نمودار جریان داده و معماری" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "مدل داده و روابط پایگاه‌داده" })).toBeVisible();
  await expect(page.getByRole("link", { name: /مشاهده نمودار کامل ERD/ })).toBeVisible();
  await expect(page.getByRole("link", { name: /مشاهده فروشگاه/ })).toBeVisible();
  await expectNoWcagViolations(page);

  await page.getByRole("button", { name: "Switch language to English" }).click();
  await expect(main).toHaveAttribute("dir", "ltr");
  await expect(page.getByRole("heading", { level: 1 })).toContainText("A global marketplace");
  await expect(page.getByRole("group", { name: "Data flow and architecture diagram" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Database model and relationships" })).toBeVisible();
  await expect(page.getByRole("link", { name: /Open the full PostgreSQL ERD/ })).toBeVisible();
  await expectNoWcagViolations(page);
});

test("reduced-motion preference disables smooth scrolling and long UI motion", async ({ page }) => {
  await page.emulateMedia({ reducedMotion: "reduce" });

  const initialMotion = await page.evaluate(() => {
    const image = document.querySelector(".product-media img");
    return {
      scrollBehavior: getComputedStyle(document.documentElement).scrollBehavior,
      imageTransition: image ? getComputedStyle(image).transitionDuration : ""
    };
  });
  expect(initialMotion.scrollBehavior).toBe("auto");
  expect(initialMotion.imageTransition).not.toBe("");
  for (const duration of initialMotion.imageTransition.split(",")) {
    expect(Number.parseFloat(duration.trim())).toBeLessThanOrEqual(0.001);
  }

  // Use the always-visible cart trigger so this check works with touch input.
  await page.locator(".bag-button").click();
  const dialog = page.locator(".side-drawer[role='dialog']");
  await expect(dialog).toBeVisible();

  const overlayDuration = await page.locator(".overlay").first().evaluate(
    element => getComputedStyle(element).animationDuration
  );
  for (const duration of overlayDuration.split(",")) {
    expect(Number.parseFloat(duration.trim())).toBeLessThanOrEqual(0.001);
  }

  await page.keyboard.press("Escape");
  await expect(dialog).toBeHidden();
});

test("command palette dialog remains accessible when open", async ({ page }) => {
  test.setTimeout(60_000);
  await page.getByRole("button", { name: "EN", exact: true }).click();
  await page.keyboard.press("Control+k");
  await expect(page.getByRole("dialog", { name: "Quick commands" })).toBeVisible();
  await expect(page.locator(".command-palette-overlay")).toHaveCSS("animation-name", "fade-backdrop");
  await expectNoWcagViolations(page);
});

test("recently viewed page meets WCAG checks in populated and empty RTL/LTR states", async ({ page }) => {
  const productLink = page.locator(".product-card h3 a").first();
  const productTitle = (await productLink.textContent())?.trim() ?? "";
  const productId = Number((await productLink.getAttribute("href"))?.split("/").filter(Boolean).pop());
  expect(productId).toBeGreaterThan(0);
  await productLink.click();
  const localizedCopy=getPersianProductCopy(productId);
  await expect(page.locator(".detail-information h1")).toHaveText(localizedCopy?.title??productTitle);
  if(localizedCopy) await expect(page.locator(".product-source-title")).toHaveText(productTitle);
  await page.waitForFunction((id: number) => {
    try {
      const state = JSON.parse(localStorage.getItem("globalcart-state-v1") ?? "{}");
      return Array.isArray(state.recent) && state.recent.includes(id);
    } catch {
      return false;
    }
  }, productId);

  const productUrl = new URL(page.url());
  const productPath = productUrl.pathname.replace(/\/$/, "");
  const productDirectory = productPath.slice(0, productPath.lastIndexOf("/"));
  const basePath = productDirectory.slice(0, productDirectory.lastIndexOf("/"));
  const historyUrl = new URL(`${basePath}/recently-viewed/`, productUrl.origin).toString();
  const response = await page.goto(historyUrl);
  expect(response?.status()).toBe(200);

  const history = page.locator("main.history-page");
  await expect(history).toHaveAttribute("data-catalog-loaded", "true", { timeout: 15_000 });
  await expect(history).toHaveAttribute("dir", "rtl");
  await expect(history.getByRole("heading", { name: "محصولات بازدیدشده" })).toBeVisible();
  await expect(history.getByRole("link").filter({ hasText: productTitle })).toBeVisible();
  await expectNoWcagViolations(page);

  await page.getByRole("button", { name: "تغییر زبان به انگلیسی" }).click();
  await expect(history).toHaveAttribute("dir", "ltr");
  await expect(history.getByRole("heading", { name: "Recently viewed products" })).toBeVisible();
  await expectNoWcagViolations(page);

  await page.getByRole("button", { name: "Clear history" }).click();
  await expect(history.getByText("No recently viewed products yet")).toBeVisible();
  await expectNoWcagViolations(page);

  await page.getByRole("button", { name: "Switch language to Persian" }).click();
  await expect(history).toHaveAttribute("dir", "rtl");
  await expect(history.getByText("هنوز محصولی بازدید نکرده‌ای")).toBeVisible();
  await expectNoWcagViolations(page);
});

test("static product detail page meets WCAG 2.1 AA automated checks", async ({ page }) => {
  const productLink = page.locator(".product-card h3 a").first();
  const productTitle = (await productLink.textContent())?.trim() ?? "";
  expect(productTitle).not.toBe("");
  const productId = Number((await productLink.getAttribute("href"))?.split("/").filter(Boolean).pop());
  await productLink.click();
  const localizedCopy=getPersianProductCopy(productId);
  await expect(page.locator(".detail-information h1")).toHaveText(localizedCopy?.title??productTitle);
  if(localizedCopy) await expect(page.locator(".product-source-title")).toHaveText(productTitle);
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


test("public GitHub links and case-study canonical URL use the renamed repository owner", async ({ page }) => {
  const assertCanonicalGitHubLinks = async () => {
    const links = await page.locator('a[href^="https://github.com/"]').evaluateAll(nodes =>
      nodes.map(node => (node as HTMLAnchorElement).href)
    );
    expect(links.length).toBeGreaterThan(0);
    expect(links.filter(href => href.includes("pooya16pg-spec"))).toEqual([]);
    expect(links.every(href => href.includes("github.com/pooya-labs/"))).toBe(true);
  };

  await assertCanonicalGitHubLinks();
  await expect(page.locator('a[href="https://github.com/pooya-labs/globalcart"]')).toHaveCount(1);

  const response = await page.goto("./case-study/");
  expect(response?.status()).toBe(200);
  await expect(page.locator("main.case-study")).toBeVisible();
  await assertCanonicalGitHubLinks();
  await expect(page.locator('link[rel="canonical"]')).toHaveAttribute(
    "href",
    "https://pooya-labs.github.io/globalcart/case-study/"
  );
  await expect(page.locator('a[href="https://github.com/pooya-labs/globalcart/blob/main/docs/DATABASE_ERD.md"]')).toHaveCount(1);
});
