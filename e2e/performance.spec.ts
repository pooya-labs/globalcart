import { test, expect } from "@playwright/test";

test.beforeEach(async ({ page }) => {
  // The storefront remains testable without remote image, font or API hosts.
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

test("mobile hero stacks the copy and image and uses a responsive source", async ({ page }, testInfo) => {
  const hero = page.locator(".hero-visual > img");
  await expect(hero).toHaveAttribute("srcset", /480w/);
  await expect(hero).toHaveAttribute("sizes", /92vw/);

  if (testInfo.project.name.includes("mobile")) {
    const copyBox = await page.locator(".hero-copy").boundingBox();
    const imageBox = await page.locator(".hero-visual").boundingBox();
    const viewport = page.viewportSize();
    expect(copyBox).not.toBeNull();
    expect(imageBox).not.toBeNull();
    expect(viewport).not.toBeNull();
    expect(imageBox!.y).toBeGreaterThan(copyBox!.y + copyBox!.height);
    expect(imageBox!.width).toBeGreaterThan(viewport!.width * 0.8);
    const documentWidth = await page.evaluate(() => ({
      scrollWidth: document.documentElement.scrollWidth,
      clientWidth: document.documentElement.clientWidth
    }));
    expect(documentWidth.scrollWidth).toBeLessThanOrEqual(documentWidth.clientWidth);
  }
});

test("homepage prioritizes its hero and defers non-critical images", async ({ page }) => {
  const hero = page.locator(".hero-visual > img");
  await expect(hero).toHaveAttribute("loading", "eager");
  await expect(hero).toHaveAttribute("fetchpriority", "high");
  await expect(hero).toHaveAttribute("decoding", "async");

  const imagePolicy = await page.locator("img").evaluateAll(images =>
    images.map(node => {
      const image = node as HTMLImageElement;
      return {
        source: image.getAttribute("src") ?? "",
        loading: image.loading,
        decoding: image.decoding,
        isHero: image.matches(".hero-visual > img")
      };
    })
  );
  const nonCritical = imagePolicy.filter(image => !image.isHero);
  expect(nonCritical.length).toBeGreaterThan(0);
  expect(nonCritical.filter(image => image.loading !== "lazy")).toEqual([]);
  expect(nonCritical.filter(image => image.decoding !== "async")).toEqual([]);
  expect(imagePolicy.filter(image => image.loading === "eager")).toHaveLength(1);

  // Fixed media containers reserve layout space even while remote images are
  // intentionally blocked in this deterministic browser suite.
  const reservedBoxes = await Promise.all([
    page.locator(".hero-visual").boundingBox(),
    page.locator(".product-media").first().boundingBox(),
    page.locator(".editorial-image").boundingBox()
  ]);
  expect(reservedBoxes.every(box => Boolean(box && box.width > 0 && box.height > 0))).toBe(true);
});

test("product detail prioritizes the main image and defers gallery extras", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();
  await page.locator(".product-card h3 a")
    .filter({ hasText: "Essence Mascara Lash Princess" })
    .click();
  await expect(page.locator(".detail-information h1"))
    .toHaveText("Essence Mascara Lash Princess");

  const mainImage = page.locator(".detail-main-image > img");
  await expect(mainImage).toHaveAttribute("loading", "eager");
  await expect(mainImage).toHaveAttribute("fetchpriority", "high");
  await expect(mainImage).toHaveAttribute("decoding", "async");

  const extras = await page.locator("img").evaluateAll(images =>
    images.filter(node => !(node as HTMLImageElement).matches(".detail-main-image > img"))
      .map(node => {
        const image = node as HTMLImageElement;
        return { loading: image.loading, decoding: image.decoding };
      })
  );
  expect(extras.length).toBeGreaterThan(0);
  expect(extras.filter(image => image.loading !== "lazy")).toEqual([]);
  expect(extras.filter(image => image.decoding !== "async")).toEqual([]);
  const mainBox = await page.locator(".detail-main-image").boundingBox();
  expect(mainBox && mainBox.width > 0 && mainBox.height > 0).toBe(true);
});

test("production fonts are self-hosted without Google Fonts requests", async ({ page }) => {
  const fontState = await page.evaluate(async () => {
    await document.fonts.ready;
    return {
      loadedFamilies: Array.from(document.fonts)
        .filter(face => face.status === "loaded")
        .map(face => face.family),
      externalGoogleRequests: (performance.getEntriesByType("resource") as PerformanceResourceTiming[])
        .filter(entry => {
          const host = new URL(entry.name).hostname;
          return host === "fonts.googleapis.com" || host === "fonts.gstatic.com";
        })
        .map(entry => entry.name)
    };
  });

  expect(fontState.loadedFamilies.join(" ")).toMatch(/Vazirmatn/i);
  expect(fontState.loadedFamilies.join(" ")).toMatch(/DM[_ ]Sans/i);
  expect(fontState.externalGoogleRequests).toEqual([]);
});

test("initial route stays within the static JavaScript and CSS transfer budgets", async ({ page }) => {
  const assets = await page.evaluate(() => {
    const entries = performance.getEntriesByType("resource") as PerformanceResourceTiming[];
    const appAssets = entries.filter(entry => {
      const url = new URL(entry.name);
      return url.origin === location.origin && url.pathname.includes("/_next/static/");
    });
    const scriptAssets = appAssets.filter(entry => /\.js$/.test(new URL(entry.name).pathname));
    const styleAssets = appAssets.filter(entry => /\.css$/.test(new URL(entry.name).pathname));
    const bytes = (items: PerformanceResourceTiming[]) =>
      items.reduce((total, item) => total + (item.transferSize || item.decodedBodySize), 0);
    return {
      scriptBytes: bytes(scriptAssets),
      styleBytes: bytes(styleAssets),
      scriptRequests: scriptAssets.length,
      styleRequests: styleAssets.length
    };
  });

  console.log("GlobalCart initial route asset budget:", JSON.stringify(assets));
  expect(assets.scriptRequests).toBeGreaterThan(0);
  // Current CI baseline is ~513 KB JS / ~68 KB CSS; leave measured headroom
  // without allowing a large dependency or stylesheet regression to slip in.
  expect(assets.scriptBytes).toBeLessThan(750_000);
  expect(assets.styleBytes).toBeLessThan(120_000);
  expect(assets.scriptRequests).toBeLessThanOrEqual(10);
  expect(assets.styleRequests).toBeLessThanOrEqual(2);
});


test("recently viewed page defers images and decodes them asynchronously", async ({ page }) => {
  await page.getByRole("button", { name: "EN", exact: true }).click();
  await page.locator(".product-card h3 a").first().click();
  await expect(page.locator(".detail-information h1")).toBeVisible();

  // The product detail route records this visit in guarded browser storage.
  await page.waitForFunction(() => {
    try {
      const state = JSON.parse(localStorage.getItem("globalcart-state-v1") ?? "{}");
      return Array.isArray(state.recent) && state.recent.length > 0;
    } catch {
      return false;
    }
  });

  // Derive the Pages base path from the product route instead of relying on
  // a relative URL: Next.js client routes omit the trailing slash in the URL,
  // so resolving "../../" would otherwise escape the /globalcart prefix.
  const productUrl = new URL(page.url());
  const productPath = productUrl.pathname.replace(/\/$/, "");
  const productDirectory = productPath.slice(0, productPath.lastIndexOf("/"));
  const basePath = productDirectory.slice(0, productDirectory.lastIndexOf("/"));
  const recentUrl = new URL(`${basePath}/recently-viewed/`, productUrl.origin);
  const historyResponse = await page.goto(recentUrl.href);
  expect(historyResponse?.status()).toBe(200);
  const historyImage = page.locator(".history-card > img").first();
  await expect(historyImage).toBeVisible();
  await expect(historyImage).toHaveAttribute("loading", "lazy");
  await expect(historyImage).toHaveAttribute("decoding", "async");
  await expect(historyImage).toHaveAttribute("alt", /.+/);

  const imageBox = await historyImage.boundingBox();
  expect(imageBox && imageBox.width > 0 && imageBox.height > 0).toBe(true);

  const historyImages = await page.locator(".history-card > img").evaluateAll(images =>
    images.map(node => {
      const image = node as HTMLImageElement;
      return {
        loading: image.loading,
        decoding: image.decoding,
        hasAccessibleName: image.hasAttribute("alt") && image.alt.trim().length > 0
      };
    })
  );
  expect(historyImages.length).toBeGreaterThan(0);
  expect(historyImages.filter(image => image.loading !== "lazy")).toEqual([]);
  expect(historyImages.filter(image => image.decoding !== "async")).toEqual([]);
  expect(historyImages.filter(image => !image.hasAccessibleName)).toEqual([]);
});
