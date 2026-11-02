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
