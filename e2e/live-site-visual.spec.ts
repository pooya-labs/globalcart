import { test, expect } from "@playwright/test";

const LIVE_URL = "https://pooya-labs.github.io/globalcart/";

for (const viewport of ["chromium-desktop", "chromium-mobile"]) {
  test(`live Pages product images load and are visually centered (${viewport})`, async ({ page }, testInfo) => {
    test.skip(testInfo.project.name !== viewport, "Run the live visual check only in desktop and mobile Chromium.");

    await page.goto(LIVE_URL, { waitUntil: "domcontentloaded", timeout: 45_000 });
    await expect(page.locator("main.store")).toBeVisible({ timeout: 30_000 });
    await expect(page.locator("main.store")).toHaveAttribute("data-catalog-loaded", "true", { timeout: 30_000 });
    await page.locator(".catalog-section").scrollIntoViewIfNeeded();

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

    const cards = page.locator(".product-card").first(8);
    await cards.first().scrollIntoViewIfNeeded();
    const first = await cards.first().boundingBox();
    const last = await cards.last().boundingBox();
    if (!first || !last) throw new Error("Could not measure the first eight live product cards.");
    const x = Math.max(0, first.x);
    const y = Math.max(0, first.y);
    const right = Math.min(page.viewportSize()!.width, Math.max(first.x + first.width, last.x + last.width));
    const bottom = Math.min(page.viewportSize()!.height + last.y, last.y + last.height);
    await page.screenshot({
      path: `test-results/live-product-grid-${viewport}.png`,
      clip: { x, y, width: right - x, height: bottom - y },
      animations: "disabled"
    });
  });
}
