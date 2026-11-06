import { test, expect } from "@playwright/test";

const LIVE_URL = "https://pooya-labs.github.io/globalcart/";

for (const viewport of ["chromium-desktop", "chromium-mobile"]) {
  test(`live Pages product images load and are visually centered (${viewport})`, async ({ page }, testInfo) => {
    test.skip(process.env.RUN_LIVE_VISUAL !== "1", "Run only after the GitHub Pages deployment completes.");
    test.skip(testInfo.project.name !== viewport, "Run the live visual check only in desktop and mobile Chromium.");

    await page.goto(LIVE_URL, { waitUntil: "domcontentloaded", timeout: 45_000 });
    await expect(page.locator("main.store")).toBeVisible({ timeout: 30_000 });
    await expect(page.locator("main.store")).toHaveAttribute("data-catalog-loaded", "true", { timeout: 30_000 });
    await page.locator(".catalog-section").scrollIntoViewIfNeeded();

    const visiblePrices = page.locator(".product-card .product-price b");
    await expect(visiblePrices.first()).toBeVisible();
    const missingPriceLabels = await visiblePrices.evaluateAll(nodes =>
      nodes.filter(node => node.textContent?.includes("قیمت در منبع درج نشده")).length
    );
    expect(missingPriceLabels).toBe(0);
    await expect(page.locator(".product-card .price-estimate-note").first())
      .toContainText("قیمت تخمینی");

    const images = page.locator(".product-card .product-image-link img");
    await expect(images).toHaveCount(36);
    await images.evaluateAll(nodes => {
      nodes.slice(0, 8).forEach(node => {
        (node as HTMLImageElement).loading = "eager";
        (node as HTMLImageElement).fetchPriority = "high";
      });
    });

    await expect.poll(async () => images.evaluateAll(nodes =>
      nodes.slice(0, 8).filter(node => {
        const image = node as HTMLImageElement;
        return image.complete && image.naturalWidth > 0 && image.naturalHeight > 0;
      }).length
    ), { timeout: 30_000 }).toBe(8);

    // naturalWidth can be available before the decoded pixels are painted.
    await images.evaluateAll(nodes => Promise.all(
      nodes.slice(0, 8).map(node => (node as HTMLImageElement).decode())
    ));

    const metrics = await images.evaluateAll(nodes => nodes.slice(0, 8).map((node, index) => {
      const image = node as HTMLImageElement;
      const media = image.closest(".product-media")!;
      const imageRect = image.getBoundingClientRect();
      const frameRect = media.getBoundingClientRect();
      const style = getComputedStyle(image);
      return {
        index: index + 1,
        title: image.alt,
        src: image.currentSrc,
        natural: [image.naturalWidth, image.naturalHeight],
        imageRect: [imageRect.x, imageRect.y, imageRect.width, imageRect.height],
        frameRect: [frameRect.x, frameRect.y, frameRect.width, frameRect.height],
        centerDelta: [
          Number((imageRect.left + imageRect.width / 2 - frameRect.left - frameRect.width / 2).toFixed(2)),
          Number((imageRect.top + imageRect.height / 2 - frameRect.top - frameRect.height / 2).toFixed(2))
        ],
        objectFit: style.objectFit,
        objectPosition: style.objectPosition
      };
    }));

    console.log("LIVE_PAGES_PRODUCT_IMAGE_METRICS", JSON.stringify(metrics));

    for (const metric of metrics) {
      expect(Math.abs(metric.centerDelta[0])).toBeLessThan(1);
      expect(Math.abs(metric.centerDelta[1])).toBeLessThan(1);
      expect(metric.imageRect[2]).toBeLessThanOrEqual(metric.frameRect[2] + 1);
      expect(metric.imageRect[3]).toBeLessThanOrEqual(metric.frameRect[3] + 1);
    }

    // Keep the visual artifact focused on the eight loaded images, not the
    // remaining lazy-loaded cards farther down the long catalog.
    await page.locator(".product-card").evaluateAll(cards => {
      cards.slice(8).forEach(card => {
        (card as HTMLElement).style.display = "none";
      });
    });
    await page.locator(".product-grid").screenshot({
      path: `test-results/live-product-grid-${viewport}.png`,
      animations: "disabled"
    });
  });
}
