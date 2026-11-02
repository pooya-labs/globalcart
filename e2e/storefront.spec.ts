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
  await page.goto("/");
  await expect(page.locator(".product-card").first()).toBeVisible();
});

test("switches locale and opens a statically exported product detail page", async ({ page }) => {
  await expect(page.locator("main.store")).toHaveAttribute("dir", "rtl");
  await page.getByRole("button", { name: "EN" }).click();
  await expect(page.locator("main.store")).toHaveAttribute("dir", "ltr");

  await page.locator(".product-card h3 a").filter({ hasText: "Essence Mascara Lash Princess" }).click();
  await expect(page.getByRole("heading", { name: "Essence Mascara Lash Princess" })).toBeVisible();
  await expect(page.getByText("PRODUCT OVERVIEW")).toBeVisible();
});

test("search suggestions support keyboard navigation and selection", async ({ page }) => {
  await page.getByRole("button", { name: "EN" }).click();
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
  await expect(page.getByRole("heading", { name: "Essence Mascara Lash Princess" })).toBeVisible();
});

test("Ctrl+K command palette opens the local shopping bag", async ({ page }) => {
  await page.getByRole("button", { name: "EN" }).click();
  await page.keyboard.press("Control+k");

  const palette = page.getByRole("dialog", { name: "Quick commands" });
  await expect(palette).toBeVisible();
  await page.getByPlaceholder("What would you like to do?").fill("shopping bag");
  await page.getByPlaceholder("What would you like to do?").press("Enter");

  await expect(page.getByRole("heading", { name: /Your bag/ })).toBeVisible();
});

test("saved searches persist locally and can be applied and deleted", async ({ page }) => {
  await page.getByRole("button", { name: "EN" }).click();
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
  await page.getByRole("button", { name: "EN" }).click();
  await page.locator(".product-card").first().getByRole("button", { name: "Add to bag" }).click();
  await page.getByRole("button", { name: /Your bag/ }).click();
  await expect(page.getByText("Essence Mascara Lash Princess")).toBeVisible();

  await page.getByRole("button", { name: "Continue to demo checkout" }).click();
  await expect(page.getByRole("heading", { name: "Demo checkout" })).toBeVisible();
  await expect(page.getByText(/Real payments are disabled/)).toBeVisible();
  await page.getByRole("button", { name: "Place demo order" }).click();
  await expect(page.getByRole("status").filter({ hasText: "Demo order created" })).toBeVisible();
});
