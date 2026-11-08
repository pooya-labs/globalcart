import { test, expect, type Page } from "@playwright/test";
import catalogSnapshot from "../data/catalog.json";
import { buildCatalogSuggestions } from "../lib/catalog/suggestions";
import { getPersianProductCopy } from "../lib/catalog/editorial";
import fxSnapshot from "../data/fx-rate.json";
import { getFxRateStatus, type FxRateSnapshot } from "../lib/fx/tgju";

const firstProduct = catalogSnapshot.products[0]!;
const firstProductQuery = firstProduct.title.slice(0, Math.min(12, firstProduct.title.length));
const savedSearchName = "First product picks";

function findSearchFixture() {
  for (const product of catalogSnapshot.products) {
    const words = product.title.split(/[^A-Za-z0-9\u00C0-\uFFFF]+/).filter(word => word.length >= 5);
    for (const word of words) {
      const query = word.slice(0, Math.min(12, word.length));
      const suggestion = buildCatalogSuggestions(catalogSnapshot.products, query)[0];
      if (suggestion?.kind === "product" && suggestion.productId === product.id) return { product, query };
    }
  }
  throw new Error("Could not find a catalog product within the bounded search suggestions.");
}

const searchFixture = findSearchFixture();

function headerLocaleToggle(page: Page) {
  return page.locator("header.header .locale-button");
}

async function openFirstVisibleProduct(page: Page) {
  const link = page.locator(".product-card h3 a").first();
  const title = (await link.textContent())?.trim() ?? "";
  const href = await link.getAttribute("href");
  const id = Number(href?.split("/").filter(Boolean).pop());
  expect(id).toBeGreaterThan(0);
  await link.click();
  return { id, title };
}

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
  await loadCatalogForTest(page);
});

test("home merchandising is curated, category-balanced and bilingual", async ({ page }) => {
  await expect(page.locator(".hero-product-collage .hero-product-tile")).toHaveCount(3);
  await expect(page.locator('.hero-product-tile[href*="/product/601270869"]')).toHaveCount(1);
  const laptopStandTile=page.locator('.hero-product-tile[href*="/product/601270869"]');
  const monitorStandTile=page.locator('.hero-product-tile[href*="/product/344744539"]');
  const laptopTableTile=page.locator('.hero-product-tile[href*="/product/2039335072"]');
  await expect(laptopStandTile).toHaveCount(1);
  await expect(monitorStandTile).toHaveCount(1);
  await expect(laptopTableTile).toHaveCount(1);
  await expect(laptopStandTile).toHaveAccessibleName("تجهیزات فضای کار پایه نگهدارنده لپ‌تاپ");
  await expect(monitorStandTile).toHaveAccessibleName("تجهیزات فضای کار پایه رومیزی مانیتور");
  await expect(laptopTableTile).toHaveAccessibleName("تجهیزات فضای کار میز قابل‌حمل لپ‌تاپ");
  const categoryTiles=page.locator(".category-strip .category-tile");
  await expect(categoryTiles.locator(".category-symbol svg")).toHaveCount(await categoryTiles.count());
  await expect(page.locator(".product-title-fa").first()).toBeVisible();
  await expect(page.locator(".product-title-original").first()).toBeVisible();
  const categories = await page.locator(".product-card").evaluateAll(cards =>
    cards.map(card => card.getAttribute("data-category")).filter((value): value is string => Boolean(value))
  );
  expect(new Set(categories).size).toBeGreaterThanOrEqual(8);
  const titles = await page.locator(".product-title-original").allTextContents();
  expect(titles).toHaveLength(36);
  expect(titles.join(" ")).not.toMatch(/medicine|cold\\s*(?:and|&)\\s*flu|saline|hand sanitizer|baby wipes|peroxide cleaner|ice cream|printer ink|ipod/i);
});


test("home catalog uses the current snapshot version and keeps estimated prices visible", async ({ page }) => {
  const catalogRequestPromise = page.waitForRequest(request => {
    const url = new URL(request.url());
    return url.pathname.endsWith("/data/catalog.json");
  });
  await page.reload();
  await page.locator(".catalog-section").scrollIntoViewIfNeeded();
  const catalogRequest = await catalogRequestPromise;
  expect(new URL(catalogRequest.url()).searchParams.get("v")).toBe(catalogSnapshot.generatedAt);
  await expect(page.locator("main.store")).toHaveAttribute("data-catalog-loaded", "true", { timeout: 15_000 });

  const unavailablePriceLabels = await page.locator(
    ".product-card .product-price b, .merchandising-item span > small:last-child"
  ).evaluateAll(nodes => nodes.filter(node => node.textContent?.includes("قیمت در منبع درج نشده")).length);
  expect(unavailablePriceLabels).toBe(0);
});


test("legacy catalog rows without priceKind still show their positive amount as an estimate", async ({ page }) => {
  await page.route("**/data/catalog.json*", async route => {
    const legacySnapshot = JSON.parse(JSON.stringify(catalogSnapshot)) as {
      products: Array<Record<string, unknown>>;
      [key: string]: unknown;
    };
    for (const product of legacySnapshot.products) delete product.priceKind;
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(legacySnapshot)
    });
  });

  await page.reload();
  await loadCatalogForTest(page);
  const firstPrice = page.locator(".product-card").first().locator(".product-price");
  await expect(firstPrice.locator("b")).not.toContainText("قیمت در منبع درج نشده");
  await expect(firstPrice.locator(".price-estimate-note")).toContainText("قیمت تخمینی");
});

test("mobile product filters open as a contained bottom sheet and close on demand", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.locator(".filter-toggle").click();
  const panel = page.locator(".filter-panel");
  await expect(page.locator(".filter-backdrop")).toBeVisible();
  await expect(panel).toHaveAttribute("role", "dialog");
  await expect(panel).toHaveAttribute("aria-modal", "true");
  expect(await panel.evaluate(element => getComputedStyle(element).position)).toBe("fixed");
  await panel.locator(".filter-close").click();
  await expect(panel).not.toHaveClass(/filter-visible/);
  await expect(page.locator(".filter-backdrop")).toHaveCount(0);
});

test("curated Persian product detail shows a localized title, concise copy and original source title", async ({ page }) => {
  const response=await page.goto("./product/601270869/");
  expect(response?.status()).toBe(200);
  await expect(page.locator("main.product-page")).toHaveAttribute("dir","rtl");
  await expect(page.locator(".detail-information h1")).toHaveText("پایه نگهدارنده لپ‌تاپ");
  await expect(page.locator(".product-source-title")).toContainText("Laptop Stand");
  await expect(page.locator(".detail-description-panel p")).toContainText("پایه‌ای جمع‌وجور");
  await expect(page.locator(".detail-source-note")).toContainText("Amazon Berkeley Objects");
});

test("captures the redesigned storefront shell for desktop and mobile visual review", async ({ page }, testInfo) => {
  // All catalog and editorial imagery is self-hosted with the static export.
  await page.reload();

  await expect(page.locator("main.store")).toHaveClass(/storefront-v2/);
  await expect(page.locator(".hero-editorial-v2")).toBeVisible();
  await expect(page.locator(".hero-proofline")).toBeVisible();
  await expect(page.locator(".categories-section .category-tile").first()).toBeVisible();
  await expect(page.locator(".merchandising-panel")).toHaveCount(3);
  await expect(page.locator(".catalog-section .product-card").first()).toBeVisible();

  // Trigger image loading only for the hero and a curated sample so the
  // visual capture does not download the entire 1,000-item catalog.
  await page.locator(".merchandising-item img, .product-card img").evaluateAll(images => {
    images.slice(0, 8).forEach(image => { (image as HTMLImageElement).loading = "eager"; });
  });
  await expect.poll(
    () => page.locator(".hero-visual > img").evaluate(image => (image as HTMLImageElement).naturalWidth),
    { timeout: 15_000 }
  ).toBeGreaterThan(0);
  await expect.poll(
    () => page.locator(".merchandising-item img").first().evaluate(image => (image as HTMLImageElement).naturalWidth),
    { timeout: 15_000 }
  ).toBeGreaterThan(0);

  // The storefront progressively renders 36 cards at a time; capture the
  // complete initial catalog viewport without forcing the full 1,000-item DOM.
  const project = testInfo.project.name;
  await page.screenshot({
    path: `test-results/storefront-home-${project}.png`,
    animations: "disabled"
  });

  // Capture real product cards as well as the hero. The previous visual
  // artifact stopped above the catalog and could not reveal image alignment.
  await page.locator(".catalog-section").scrollIntoViewIfNeeded();
  await page.locator(".product-card img").evaluateAll(images => {
    images.slice(0, 36).forEach(image => {
      (image as HTMLImageElement).loading = "eager";
    });
  });
  await expect.poll(
    () => page.locator(".product-card img").evaluateAll(images =>
      images.slice(0, 8).filter(image => (image as HTMLImageElement).naturalWidth > 0).length
    ),
    { timeout: 20_000 }
  ).toBeGreaterThanOrEqual(4);
  await page.locator(".product-grid").screenshot({
    path: `test-results/product-grid-${project}.png`,
    animations: "disabled"
  });
});

test("keeps the filter heading in normal flow above the category options", async ({ page }) => {
  await page.setViewportSize({ width: 1440, height: 900 });

  const panel = page.locator(".filter-panel");
  const title = panel.locator(".filter-title");
  await expect(panel).toBeVisible();

  const layout = await Promise.all([
    title.evaluate(element => getComputedStyle(element).position),
    panel.evaluate(element => getComputedStyle(element).position)
  ]);
  expect(layout).toEqual(["static", "sticky"]);

  const titleBottom = await title.evaluate(element => element.getBoundingClientRect().bottom);
  const categoryHeadingTop = await panel.locator(".filter-group").first()
    .locator("b").evaluate(element => element.getBoundingClientRect().top);
  expect(titleBottom).toBeLessThanOrEqual(categoryHeadingTop);
});

test("renders the first 36 products and reveals the next batch on demand", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();
  await expect(page.locator(".product-card")).toHaveCount(36);
  const loadMore = page.getByRole("button", { name: "Load more products", exact: true });
  await expect(loadMore).toBeVisible();
  await loadMore.click();
  await expect(page.locator(".product-card")).toHaveCount(72);
});

test("product cards center the natural image canvas without cropping", async ({ page }) => {
  const card = page.locator(".product-card").first();
  const image = card.locator(".product-image-link img");
  await image.scrollIntoViewIfNeeded();
  await expect.poll(() => image.evaluate(node => {
    const img = node as HTMLImageElement;
    return img.complete && img.naturalWidth > 0 && img.naturalHeight > 0;
  })).toBe(true);

  const geometry = await image.evaluate(node => {
    const img = node as HTMLImageElement;
    const imageRect = img.getBoundingClientRect();
    const frameRect = img.closest(".product-media")!.getBoundingClientRect();
    const style = getComputedStyle(img);
    return {
      imageCenterX: imageRect.left + imageRect.width / 2,
      imageCenterY: imageRect.top + imageRect.height / 2,
      frameCenterX: frameRect.left + frameRect.width / 2,
      frameCenterY: frameRect.top + frameRect.height / 2,
      objectFit: style.objectFit,
      objectPosition: style.objectPosition,
      imageWidth: imageRect.width,
      imageHeight: imageRect.height,
      frameWidth: frameRect.width,
      frameHeight: frameRect.height
    };
  });

  expect(Math.abs(geometry.imageCenterX - geometry.frameCenterX)).toBeLessThan(1);
  expect(Math.abs(geometry.imageCenterY - geometry.frameCenterY)).toBeLessThan(1);
  expect(geometry.objectFit).toBe("contain");
  expect(geometry.objectPosition).toBe("50% 50%");
  expect(geometry.imageWidth).toBeLessThanOrEqual(geometry.frameWidth);
  expect(geometry.imageHeight).toBeLessThanOrEqual(geometry.frameHeight);
});

test("shows only live catalog categories, filters products locally and gives product images a real layout box", async ({ page }) => {
  const tiles = page.locator(".categories-section .category-tile");
  await expect(tiles.first()).toBeVisible();

  const firstTile = tiles.first();
  const categoryId = await firstTile.getAttribute("data-category");
  const categoryCount = Number(await firstTile.getAttribute("data-count"));
  expect(categoryId).toBeTruthy();
  expect(categoryCount).toBeGreaterThan(0);

  const imageLink = page.locator(".product-card .product-image-link").first();
  const imageGeometry = await imageLink.evaluate(element => {
    const wrapper = element.getBoundingClientRect();
    const image = element.querySelector("img")?.getBoundingClientRect();
    return {
      wrapperWidth: wrapper.width,
      wrapperHeight: wrapper.height,
      imageWidth: image?.width ?? 0,
      imageHeight: image?.height ?? 0
    };
  });
  expect(imageGeometry.wrapperWidth).toBeGreaterThan(0);
  expect(imageGeometry.wrapperHeight).toBeGreaterThan(0);
  expect(imageGeometry.imageWidth).toBeGreaterThan(0);
  expect(imageGeometry.imageHeight).toBeGreaterThan(0);

  await firstTile.click();
  await expect(page.locator(".filter-option.active[data-category]")).toHaveAttribute("data-category", categoryId!);
  await expect(page.locator(".product-card").first()).toBeVisible();
});

test("product detail gallery keeps alternate images browsable", async ({ page }) => {
  const galleryProduct = catalogSnapshot.products.find(product =>
    Array.isArray(product.images) && product.images.some(image => image !== product.thumbnail)
  );
  test.skip(!galleryProduct, "The committed snapshot has no distinct alternate image yet; the deployment smoke test validates freshly generated galleries.");

  await page.goto(`./product/${galleryProduct!.id}/`);
  await expect(page.locator("main.product-page")).toBeVisible();

  await expect(page.locator("main.product-page")).toBeVisible();
  await expect(page.locator("main.product-page h1")).toHaveText(galleryProduct!.title);
  const mainImage = page.locator(".detail-main-image img");
  await expect(mainImage).toHaveAttribute("loading", "eager");
  await expect(mainImage).toHaveAttribute("fetchpriority", "high");
  await expect(page.locator(".detail-provenance")).toHaveCount(0);
  await expect(page.locator('a[href*="amazon."]')).toHaveCount(0);

  const galleryItems = page.locator(".detail-thumbnails button");
  await expect(galleryItems).toHaveCount(galleryProduct!.images.length + 1);
  const firstImage = await mainImage.getAttribute("src");
  await page.locator(".gallery-arrow-next").click();
  await expect(mainImage).not.toHaveAttribute("src", firstImage || "");
  await expect(page.locator(".gallery-counter")).toContainText("2 /");
  await galleryItems.first().click();
  await expect(mainImage).toHaveAttribute("src", firstImage || "");
});

test("switches locale and opens a statically exported product detail page", async ({ page }) => {
  await expect(page.locator("main.store")).toHaveAttribute("dir", "rtl");
  await page.getByRole("button", { name: "EN", exact: true }).click();
  await expect(page.locator("main.store")).toHaveAttribute("dir", "ltr");
  await page.reload();
  await expect(page.locator("main.store")).toHaveAttribute("dir", "ltr");

  const opened = await openFirstVisibleProduct(page);
  await expect(page).toHaveURL(new RegExp("/product/" + opened.id + "/?$"));
  await expect(page.locator("main.product-page")).toHaveAttribute("dir", "ltr");
  await expect(page.locator(".detail-information h1")).toHaveText(opened.title);
  await expect(page.getByText("PRODUCT OVERVIEW")).toBeVisible();
});

test("labels date-only TGJU quotes as stale and keeps unprovided product facts explicit", async ({ page }) => {
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

  const opened = await openFirstVisibleProduct(page);
  await expect(page.locator(".detail-information h1")).toHaveText(opened.title);
  await expect(page.locator(".detail-price")).toBeVisible();
  await expect(page.locator(".buybox-kicker")).toContainText("ESTIMATED · DEMO PRICE");
  await expect(page.locator(".detail-buybox details.price-breakdown")).toContainText("Estimated base price (USD)");
  await expect(page.locator(".detail-rating")).toContainText("No rating supplied");
  await expect(page.locator(".buybox-stock")).toContainText("Stock not listed");
  await expect(page.locator(".detail-buybox details.price-breakdown")).toHaveCount(1);
});

test("shows the TGJU quote and keeps missing catalog commerce fields honest", async ({ page }) => {
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

  const opened = await openFirstVisibleProduct(page);
  await expect(page).toHaveURL(new RegExp("/product/" + opened.id + "/?$"));
  await expect(page.locator("main.product-page")).toHaveAttribute("dir", "ltr");
  await expect(page.locator(".detail-information h1")).toHaveText(opened.title);
  await expect(page.locator(".detail-price")).toBeVisible();
  await expect(page.locator(".buybox-kicker")).toContainText("ESTIMATED · DEMO PRICE");
  await expect(page.locator(".detail-buybox details.price-breakdown")).toBeVisible();
  await expect(page.locator(".detail-source-note")).toContainText("Amazon Berkeley Objects");
  await expect(page.locator(".detail-rating")).toContainText("No rating supplied");
  await expect(page.locator(".buybox-stock")).toContainText("Stock not listed");
  await expect(page.locator('a[href*="amazon."]')).toHaveCount(0);
});

test("search suggestions support keyboard navigation and selection", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();
  const mobileSearchButton = page.getByRole("button", { name: "Search products, brands or categories...", exact: true });
  if (await mobileSearchButton.isVisible()) await mobileSearchButton.click();
  const search = page.getByRole("combobox", { name: "Search products, brands or categories..." });
  await search.fill(searchFixture.query);

  const listbox = page.getByRole("listbox", { name: "Search suggestions" });
  await expect(listbox).toBeVisible();
  const productOption = listbox.getByRole("option").filter({ hasText: searchFixture.product.title });
  await expect(productOption).toBeVisible();

  await search.press("ArrowDown");
  await expect(search).toHaveAttribute("aria-activedescendant", /option-/);
  await search.press("Enter");
  await expect(page.getByRole("dialog")).toBeVisible();
  await expect(page.locator("#quick-dialog-title")).toHaveText(searchFixture.product.title);
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
  await search.fill(firstProductQuery);
  await page.locator("#catalog").scrollIntoViewIfNeeded();

  await page.getByRole("button", { name: "Save search" }).click();
  await page.getByLabel("Search name").fill(savedSearchName);
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(page.getByRole("status").filter({ hasText: "Saved in this browser." })).toBeVisible();

  await page.reload();
  await expect(page.locator(".product-card").first()).toBeVisible();
  await expect(page.locator("main.store")).toHaveAttribute("dir", "ltr");
  await page.getByRole("button", { name: /Saved searches/ }).click();
  await expect(page.getByText(savedSearchName)).toBeVisible();

  await page.getByRole("button", { name: "Apply" }).click();
  if (await page.getByRole("button", { name: "Search products, brands or categories..." }).isVisible()) {
    await page.getByRole("button", { name: "Search products, brands or categories..." }).click();
  }
  await expect(page.getByRole("combobox", { name: "Search products, brands or categories..." })).toHaveValue(firstProductQuery);

  await page.getByRole("button", { name: /Saved searches/ }).click();
  await page.getByRole("button", { name: "Delete " + savedSearchName }).click();
  await expect(page.getByText(savedSearchName)).toHaveCount(0);
});

test("cart and checkout remain a clearly simulated local flow", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();
  const targetProduct = page.locator(".product-card").first();
  const productTitle = (await targetProduct.locator("h3 a").textContent())?.trim() ?? "";
  await targetProduct.locator(".product-price button").click();
  await page.getByRole("button", { name: /Your bag/ }).click();
  await expect(page.getByRole("dialog").getByText(productTitle, { exact: true })).toBeVisible();

  await page.getByRole("button", { name: "Continue to demo checkout" }).click();
  await expect(page.getByRole("heading", { name: "Demo checkout" })).toBeVisible();
  await expect(page.locator(".checkout-warning")).toContainText("No real payment is processed");
  await page.getByRole("button", { name: "Place demo order" }).click();
  await expect(page.getByRole("status").filter({ hasText: "Demo order created" })).toBeVisible();
});

test("product detail add-to-bag persists the selected item in the local cart", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();
  const opened = await openFirstVisibleProduct(page);
  await expect(page.locator(".detail-information h1")).toHaveText(opened.title);
  await page.locator(".buybox-add").click();
  await expect(page.locator(".buybox-confirm")).toContainText("Added to your bag");
  await page.waitForFunction((productId: number) => {
    try {
      const state = JSON.parse(localStorage.getItem("globalcart-state-v1") ?? "{}");
      return Array.isArray(state.cart) && state.cart.some((line: { product?: { id?: number } }) => line.product?.id === productId);
    } catch {
      return false;
    }
  }, opened.id);
  await page.locator(".product-page-top > div > a").click();
  await expect(page.locator("main.store")).toBeVisible();
  await page.locator(".bag-button").click();
  const drawer = page.locator(".side-drawer[role='dialog']");
  await expect(drawer.getByText(opened.title, { exact: true })).toBeVisible();
});

 test("homepage clear history action matches its neighboring history link", async ({ page }) => {
  await openFirstVisibleProduct(page);
  await page.locator(".product-page-top > div > a").click();

  const section = page.locator(".recently-viewed-section");
  await expect(section).toBeVisible();
  const clear = section.locator(".recently-viewed-actions > button");
  const viewHistory = section.locator(".recently-viewed-actions > a");
  await expect(clear).toBeVisible();
  await expect(viewHistory).toBeVisible();

  const styles = await page.evaluate(() => {
    const button = document.querySelector(".recently-viewed-actions > button");
    const link = document.querySelector(".recently-viewed-actions > a");
    if (!button || !link) throw new Error("Recently viewed actions are missing.");
    const properties = ["display", "align-items", "gap", "color", "font-family", "font-size", "font-weight", "line-height", "text-decoration-line"];
    const read = (element: Element) => {
      const style = getComputedStyle(element);
      return Object.fromEntries(properties.map(property => [property, style.getPropertyValue(property)]));
    };
    return { button: read(button), link: read(link) };
  });
  expect(styles.button).toEqual(styles.link);

  await clear.click();
  await expect(section).toHaveCount(0);
});

test("recently viewed route inherits and persists the selected locale", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();
  await expect(page.locator("main.store")).toHaveAttribute("dir", "ltr");

  const opened = await openFirstVisibleProduct(page);
  await expect(page).toHaveURL(/product/);
  await expect(page.locator("main.product-page")).toHaveAttribute("dir", "ltr");
  await page.waitForFunction((id: number) => {
    try {
      const state = JSON.parse(localStorage.getItem("globalcart-state-v1") ?? "{}");
      return Array.isArray(state.recent) && state.recent.includes(id);
    } catch {
      return false;
    }
  }, opened.id);

  const productUrl = new URL(page.url());
  const productPath = productUrl.pathname.replace(/\/$/, "");
  const productDirectory = productPath.slice(0, productPath.lastIndexOf("/"));
  const basePath = productDirectory.slice(0, productDirectory.lastIndexOf("/"));
  const historyUrl = new URL(`${basePath}/recently-viewed/`, productUrl.origin).toString();
  await page.goto(historyUrl);
  const history = page.locator("main.history-page");
  await expect(history).toHaveAttribute("data-catalog-loaded", "true", { timeout: 15_000 });
  await expect(history).toHaveAttribute("dir", "ltr");
  await expect(history.getByRole("heading", { name: "Recently viewed products" })).toBeVisible();
  await expect(history.locator(".history-note")).toContainText("Prices are local demo estimates, not seller quotes");
  await expect(history.getByRole("link").filter({ hasText: opened.title })).toBeVisible();

  await page.getByRole("button", { name: "Switch language to Persian" }).click();
  await expect(history).toHaveAttribute("dir", "rtl");
  await expect(page.getByRole("button", { name: "تغییر زبان به انگلیسی" })).toBeVisible();
  await expect(history.getByRole("heading", { name: "محصولات بازدیدشده" })).toBeVisible();

  await page.reload();
  await expect(history).toHaveAttribute("dir", "rtl");
  await page.getByRole("link", { name: /بازگشت به فروشگاه/ }).click();
  await expect(page.locator("main.store")).toHaveAttribute("dir", "rtl");
});


test("recently viewed clear action matches shared history controls and clears local entries", async ({ page }) => {
  const opened = await openFirstVisibleProduct(page);
  await expect(page).toHaveURL(new RegExp("/product/" + opened.id + "/?$"));
  await page.waitForFunction((id: number) => {
    try {
      const state = JSON.parse(localStorage.getItem("globalcart-state-v1") ?? "{}");
      return Array.isArray(state.recent) && state.recent.includes(id);
    } catch {
      return false;
    }
  }, opened.id);
  const productUrl = new URL(page.url());
  const productPath = productUrl.pathname.replace(/\/$/, "");
  const productDirectory = productPath.slice(0, productPath.lastIndexOf("/"));
  const basePath = productDirectory.slice(0, productDirectory.lastIndexOf("/"));
  await page.goto(new URL(`${basePath}/recently-viewed/`, productUrl.origin).toString());

  const history = page.locator("main.history-page");
  await expect(history).toHaveAttribute("data-catalog-loaded", "true", { timeout: 15_000 });
  const clear = history.locator(".history-clear");
  await expect(clear).toBeVisible();

  const styles = await page.evaluate(() => {
    const clearButton = document.querySelector(".history-clear");
    const sharedButton = document.querySelector(".history-topbar button");
    if (!clearButton || !sharedButton) throw new Error("History controls are missing.");
    const properties = ["backgroundColor", "borderTopColor", "borderTopStyle", "borderTopWidth", "borderRadius", "color", "paddingTop", "paddingRight", "paddingBottom", "paddingLeft"];
    const read = (element: Element) => {
      const style = getComputedStyle(element);
      return Object.fromEntries(properties.map(property => [property, style.getPropertyValue(property.replace(/[A-Z]/g, letter => `-${letter.toLowerCase()}`))]));
    };
    return { clear: read(clearButton), shared: read(sharedButton) };
  });
  expect(styles.clear).toEqual(styles.shared);

  await clear.click();
  await expect(history.locator(".history-clear")).toHaveCount(0);
  await expect(history).toHaveAttribute("dir", "rtl");
  await expect(history.getByText("هنوز محصولی بازدید نکرده‌ای")).toBeVisible();
  await expect.poll(() => page.evaluate(() => {
    const state = JSON.parse(localStorage.getItem("globalcart-state-v1") ?? "{}");
    return Array.isArray(state.recent) && state.recent.length === 0;
  })).toBe(true);
});

test("active catalog filter chips remove one facet or clear the complete filter set", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();
  const isMobile = await page.evaluate(() => window.matchMedia("(max-width: 760px)").matches);
  if (isMobile) await page.getByRole("button", { name: "Filters" }).click();

  const brand = page.locator(".facet-select-group select").first();
  await brand.selectOption({ index: 1 });
  if (isMobile) await page.locator(".filter-panel .filter-close").click();
  await expect(page.locator(".active-filter-bar")).toBeVisible();
  await expect(page.locator(".active-filter-chip")).toContainText(await brand.locator("option:checked").textContent() ?? "");

  await page.locator(".active-filter-chip").first().click();
  await expect(page.locator(".active-filter-bar")).toBeHidden();
  if (isMobile) await page.getByRole("button", { name: "Filters" }).click();
  await brand.selectOption({ index: 1 });
  await page.locator(".facet-select-group select").nth(1).selectOption("4");
  if (isMobile) await page.locator(".filter-panel .filter-close").click();
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
  await page.locator(".filter-panel .filter-close").click();
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


test("recently viewed empty state waits for browser storage hydration", async ({ page }) => {
  await page.goto("./recently-viewed/");
  const history = page.locator("main.history-page");
  await expect(history).toHaveAttribute("data-catalog-loaded", "true", { timeout: 15_000 });
  await expect(history.locator(".history-empty b")).toBeVisible();
  await expect(history.locator(".history-grid")).toHaveCount(0);
  await expect(history.locator(".history-clear")).toHaveCount(0);
});

test("mobile header keeps locale and cart controls aligned without overlap in RTL and LTR", async ({ page }) => {
  for (const width of [360, 390]) {
    await page.setViewportSize({ width, height: 800 });
    for (const locale of ["rtl", "ltr"] as const) {
      const root = page.locator("main.store");
      if ((await root.getAttribute("dir")) !== locale) {
        await headerLocaleToggle(page).click();
      }
      await expect(root).toHaveAttribute("dir", locale);
      const header = page.locator("header.header");
      const language = header.locator(".locale-button");
      const cart = header.locator(".bag-button");
      await expect(language).toBeVisible();
      await expect(cart).toBeVisible();
      const geometry = await header.evaluate(element => {
        const rect = (selector: string) => {
          const box = element.querySelector(selector)!.getBoundingClientRect();
          return { left: box.left, right: box.right, top: box.top, bottom: box.bottom };
        };
        return {
          header: rect(".header-actions"),
          language: rect(".locale-button"),
          cart: rect(".bag-button"),
          viewportWidth: document.documentElement.clientWidth
        };
      });
      expect(geometry.language.left).toBeGreaterThanOrEqual(0);
      expect(geometry.cart.left).toBeGreaterThanOrEqual(0);
      expect(geometry.language.right).toBeLessThanOrEqual(geometry.viewportWidth);
      expect(geometry.cart.right).toBeLessThanOrEqual(geometry.viewportWidth);
      expect(
        geometry.language.right <= geometry.cart.left || geometry.cart.right <= geometry.language.left
      ).toBeTruthy();
      expect(geometry.header.right - geometry.header.left).toBeLessThanOrEqual(geometry.viewportWidth);
    }
  }
});
