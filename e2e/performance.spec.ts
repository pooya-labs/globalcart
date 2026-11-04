import { test, expect } from "@playwright/test";
import catalogSnapshot from "../data/catalog.json";

const firstProduct = catalogSnapshot.products[0]!;

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
  await expect(page.locator("main.store")).toHaveAttribute("data-catalog-loaded", "true", { timeout: 15_000 });
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
    expect(imageBox!.y).toBeGreaterThanOrEqual(copyBox!.y + copyBox!.height);
    expect(imageBox!.width).toBeGreaterThan(viewport!.width * 0.8);
    const documentWidth = await page.evaluate(() => ({
      scrollWidth: document.documentElement.scrollWidth,
      clientWidth: document.documentElement.clientWidth
    }));
    expect(documentWidth.scrollWidth).toBeLessThanOrEqual(documentWidth.clientWidth);
  }
});

test("catalog and product detail images are self-hosted without obsolete external preconnects", async ({ page }) => {
  const externalHints = async () => page.locator('head link[rel="preconnect"]').evaluateAll(nodes =>
    nodes.map(node => new URL((node as HTMLLinkElement).href).origin)
      .filter(origin => origin !== location.origin)
  );

  expect(await externalHints()).toEqual([]);
  const cardImage = page.locator(".product-card .product-image-link img").first();
  const cardSource = await cardImage.getAttribute("src");
  expect(cardSource).toBeTruthy();
  expect(new URL(cardSource!, page.url()).origin).toBe(new URL(page.url()).origin);
  expect(new URL(cardSource!, page.url()).pathname).toContain("/images/catalog/");

  await page.goto(`./product/${firstProduct.id}/`);
  await expect(page.locator(".detail-information h1")).toHaveText(firstProduct.title);
  const detailImage = await page.locator(".detail-main-image > img").getAttribute("src");
  expect(detailImage).toBeTruthy();
  expect(new URL(detailImage!, page.url()).origin).toBe(new URL(page.url()).origin);
  expect(new URL(detailImage!, page.url()).pathname).toContain("/images/catalog/");

  await page.goto("./case-study/");
  await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
  expect(await externalHints()).toEqual([]);
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
  const productLink = page.locator(".product-card h3 a").first();
  const productTitle = await productLink.innerText();
  await productLink.click();
  await expect(page.locator(".detail-information h1")).toHaveText(productTitle);

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

test("site icon resolves under the GitHub Pages base path", async ({ page }) => {
  const iconLink = page.locator('link[rel="icon"]').first();
  await expect(iconLink).toHaveAttribute("href", /icon\.svg/);
  const href = await iconLink.getAttribute("href");
  expect(href).toBeTruthy();

  const iconUrl = new URL(href!, page.url());
  expect(iconUrl.pathname).toBe("/globalcart/icon.svg");

  const response = await page.request.get(iconUrl.href);
  expect(response.status()).toBe(200);
  expect(response.headers()["content-type"]).toContain("image/svg+xml");
  expect(await response.text()).toContain("<svg");
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

test("keeps the full catalog out of the initial storefront DOM", async ({ page }) => {
  await expect(page.locator("main.store")).toHaveAttribute("data-catalog-loaded", "true");
  await expect(page.locator(".product-card")).toHaveCount(36);
  await expect(page.getByRole("button", { name: "Load more products", exact: true })).toBeVisible();
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


test("tablet viewport fits the storefront, product and case-study routes without horizontal overflow", async ({ page }, testInfo) => {
  test.skip(!testInfo.project.name.includes("tablet"), "Tablet-only responsive regression check.");

  const assertNoHorizontalOverflow = async (routeName: string, orientation: string) => {
    const dimensions = await page.evaluate(() => ({
      viewportWidth: document.documentElement.clientWidth,
      documentWidth: document.documentElement.scrollWidth,
      bodyWidth: document.body.scrollWidth
    }));
    expect(dimensions.viewportWidth).toBeGreaterThanOrEqual(768);
    expect(
      Math.max(dimensions.documentWidth, dimensions.bodyWidth),
      `${routeName} overflows the tablet ${orientation} viewport: ${JSON.stringify(dimensions)}`
    ).toBeLessThanOrEqual(dimensions.viewportWidth);
  };

  const orientations = [
    { name: "portrait", width: 810, height: 1080 },
    { name: "landscape", width: 1080, height: 810 }
  ];

  for (const orientation of orientations) {
    await page.setViewportSize({ width: orientation.width, height: orientation.height });

    await page.goto("./");
    await expect(page.locator("main.store")).toBeVisible();
    await assertNoHorizontalOverflow("Storefront", orientation.name);

    const productResponse = await page.goto(`./product/${firstProduct.id}/`);
    expect(productResponse?.status()).toBe(200);
    await expect(page.locator("main.product-page")).toBeVisible();
    await expect(page.locator(".detail-information h1")).toBeVisible();
    await assertNoHorizontalOverflow("Product detail", orientation.name);

    const caseStudyResponse = await page.goto("./case-study/");
    expect(caseStudyResponse?.status()).toBe(200);
    await expect(page.locator("main.case-study")).toBeVisible();
    await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
    await assertNoHorizontalOverflow("Case study", orientation.name);
  }
});


test("compact phone widths fit all public routes without horizontal overflow", async ({ page }, testInfo) => {
  test.skip(!testInfo.project.name.includes("mobile"), "Compact phone viewport regression runs on the mobile project.");

  const widths = [
    { name: "narrow 320 px", width: 320, height: 740 },
    { name: "compact 360 px", width: 360, height: 800 }
  ];

  for (const viewport of widths) {
    await page.setViewportSize({ width: viewport.width, height: viewport.height });
    const assertFits = async (route: string) => {
      const dimensions = await page.evaluate(() => ({
        viewport: document.documentElement.clientWidth,
        document: document.documentElement.scrollWidth,
        body: document.body.scrollWidth
      }));
      expect(
        Math.max(dimensions.document, dimensions.body),
        `${route} overflows ${viewport.name}: ${JSON.stringify(dimensions)}`
      ).toBeLessThanOrEqual(dimensions.viewport);
    };

    await page.goto("./");
    await expect(page.locator("main.store")).toBeVisible();
    await assertFits("Storefront");

    const product = await page.goto(`./product/${firstProduct.id}/`);
    expect(product?.status()).toBe(200);
    await expect(page.locator("main.product-page")).toBeVisible();
    await assertFits("Product detail");

    const caseStudy = await page.goto("./case-study/");
    expect(caseStudy?.status()).toBe(200);
    await expect(page.locator("main.case-study")).toBeVisible();
    await assertFits("Case study");
  }
});


test("site-owned hero and editorial images decode from same-origin assets on every viewport", async ({ page }) => {
  const images = [
    page.locator(".hero-visual > img"),
    page.locator(".editorial-image > img")
  ];

  for (const image of images) {
    await image.scrollIntoViewIfNeeded();
    await expect.poll(() => image.evaluate(node => {
      const element = node as HTMLImageElement;
      return element.complete && element.naturalWidth > 0 && element.naturalHeight > 0;
    })).toBe(true);

    const currentSrc = await image.evaluate(node => (node as HTMLImageElement).currentSrc);
    const resolved = new URL(currentSrc);
    expect(resolved.origin).toBe(new URL(page.url()).origin);
    expect(resolved.pathname).toContain("/images/site/");
  }

  const heroSources = await page.locator(".hero-visual > img").evaluate(node => {
    const element = node as HTMLImageElement;
    return [element.getAttribute("src") ?? "", ...((element.getAttribute("srcset") ?? "")
      .split(",")
      .map(candidate => candidate.trim().split(" ")[0]))];
  });
  expect(heroSources).toHaveLength(4);
  expect(heroSources.every(source => source.startsWith("images/site/"))).toBe(true);
});


test("redesigned lower sections and footer keep a readable responsive layout", async ({ page }, testInfo) => {
  const viewports = testInfo.project.name.includes("mobile")
    ? [{ width: 320, height: 740 }, { width: 390, height: 844 }]
    : testInfo.project.name.includes("tablet")
      ? [{ width: 810, height: 1080 }, { width: 1080, height: 810 }]
      : [{ width: 1440, height: 1000 }];

  for (const viewport of viewports) {
    await page.setViewportSize(viewport);
    await page.goto("./");
    await expect(page.locator(".marketplace-services .service-cards")).toBeVisible();
    await expect(page.locator(".faq-section .faq-list details").first()).toBeVisible();
    await expect(page.locator(".about-section .about-stats")).toBeVisible();
    await expect(page.locator("footer.footer .footer-main")).toBeVisible();

    const layout = await page.evaluate(() => {
      const selectors = [
        ".marketplace-services",
        ".faq-section",
        ".about-section",
        "footer.footer",
        ".footer-main",
        ".footer-bottom"
      ];
      const boxes = Object.fromEntries(selectors.map(selector => {
        const element = document.querySelector(selector);
        if (!element) return [selector, null];
        const rect = element.getBoundingClientRect();
        return [selector, {
          left: Math.round(rect.left),
          right: Math.round(rect.right),
          width: Math.round(rect.width),
          height: Math.round(rect.height)
        }];
      }));
      return {
        viewportWidth: document.documentElement.clientWidth,
        documentWidth: document.documentElement.scrollWidth,
        bodyWidth: document.body.scrollWidth,
        boxes,
        footerColumns: getComputedStyle(document.querySelector(".footer-main")!).gridTemplateColumns,
        serviceColumns: getComputedStyle(document.querySelector(".service-cards")!).gridTemplateColumns,
        faqColumns: getComputedStyle(document.querySelector(".faq-section")!).gridTemplateColumns
      };
    });

    expect(
      Math.max(layout.documentWidth, layout.bodyWidth),
      `Lower-page horizontal overflow at ${viewport.width}px: ${JSON.stringify(layout)}`
    ).toBeLessThanOrEqual(layout.viewportWidth);

    for (const [selector, box] of Object.entries(layout.boxes)) {
      expect(box, `Missing lower-page region: ${selector}`).not.toBeNull();
      expect(box!.width, `${selector} has no layout width`).toBeGreaterThan(0);
      expect(box!.height, `${selector} has no layout height`).toBeGreaterThan(0);
    }

    if (viewport.width <= 760) {
      expect(layout.footerColumns.trim().split(/\s+/)).toHaveLength(1);
      expect(layout.serviceColumns.trim().split(/\s+/)).toHaveLength(1);
      expect(layout.faqColumns.trim().split(/\s+/)).toHaveLength(1);
    } else if (viewport.width <= 1000) {
      expect(layout.footerColumns.trim().split(/\s+/)).toHaveLength(2);
    } else {
      expect(layout.footerColumns.trim().split(/\s+/)).toHaveLength(3);
      expect(layout.serviceColumns.trim().split(/\s+/)).toHaveLength(3);
      expect(layout.faqColumns.trim().split(/\s+/)).toHaveLength(2);
    }
  }
});
