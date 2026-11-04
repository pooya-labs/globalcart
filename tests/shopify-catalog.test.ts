import test from "node:test";
import assert from "node:assert/strict";
import { mapShopifyCatalogProduct } from "../lib/catalog/shopify";

const sample = {
  id: "gid://shopify/p/7f3a2b8c1d9e",
  title: "Trail Runner Pro",
  description: { html: "<p>Lightweight <strong>trail</strong> running shoe.</p>" },
  url: "https://example-running.myshopify.com/products/trail-runner-pro",
  categories: [{ value: "Sporting Goods > Athletics > Running > Shoes", taxonomy: "merchant" }],
  price_range: { min: { amount: 8999, currency: "USD" }, max: { amount: 12999, currency: "USD" } },
  media: [{ type: "image", url: "https://cdn.shopify.com/products/trail-runner-pro.jpg", alt_text: "Trail Runner Pro" }],
  variants: [{
    price: { amount: 8999, currency: "USD" },
    availability: { available: true },
    seller: { name: "Example Running", url: "https://example-running.myshopify.com" }
  }],
  rating: { value: 4.5 }
};

test("maps a real merchant listing with live price, image and seller attribution", () => {
  const product = mapShopifyCatalogProduct(sample);
  assert.ok(product);
  assert.equal(product.title, "Trail Runner Pro");
  assert.equal(product.price, 89.99);
  assert.equal(product.priceMax, 129.99);
  assert.equal(product.currency, "USD");
  assert.equal(product.rating, 4.5);
  assert.equal(product.category, "sports-outdoors");
  assert.equal(product.sellerName, "Example Running");
  assert.equal(product.sellerUrl, "https://example-running.myshopify.com/");
  assert.equal(product.sourceUrl, "https://example-running.myshopify.com/products/trail-runner-pro");
  assert.equal(product.thumbnail, "https://cdn.shopify.com/products/trail-runner-pro.jpg");
  assert.equal(product.isLive, true);
  assert.equal(product.discountPercentage, 0);
  assert.match(product.description, /Lightweight trail running shoe/);
});

test("does not fabricate a listing when identity, price or product imagery is missing", () => {
  assert.equal(mapShopifyCatalogProduct({ ...sample, id: undefined }), null);
  assert.equal(mapShopifyCatalogProduct({ ...sample, price_range: undefined }), null);
  assert.equal(mapShopifyCatalogProduct({ ...sample, media: [] }), null);
  assert.equal(mapShopifyCatalogProduct({ ...sample, media: [{ type: "image", url: "javascript:alert(1)" }] }), null);
});

test("uses a stable positive numeric ID for browser-local cart references", () => {
  const first = mapShopifyCatalogProduct(sample);
  const second = mapShopifyCatalogProduct(sample);
  assert.ok(first);
  assert.ok(second);
  assert.equal(first.id, second.id);
  assert.ok(first.id > 0);
});

test("handles missing rating and availability without inventing a rating or stock count", () => {
  const product = mapShopifyCatalogProduct({
    ...sample,
    rating: undefined,
    variants: [{ price: { amount: 8999, currency: "USD" }, availability: { available: false }, seller: { name: "Store" } }]
  });
  assert.ok(product);
  assert.equal(product.rating, 0);
  assert.equal(product.availability, false);
  assert.equal(product.stock, 0);
});


test("classifies live listings when Shopify omits category fields", () => {
  const headphones = mapShopifyCatalogProduct({
    ...sample,
    categories: undefined,
    title: "Crusher Evo",
    description: { plain: "Wireless Bluetooth headphones with active noise cancellation." }
  });
  assert.ok(headphones);
  assert.equal(headphones.category, "electronics");

  const runningShoes = mapShopifyCatalogProduct({
    ...sample,
    categories: undefined,
    title: "Trail Runner Pro",
    description: { plain: "Lightweight running shoes for outdoor trails." }
  });
  assert.ok(runningShoes);
  assert.equal(runningShoes.category, "sports-outdoors");

  const galaxyWatch = mapShopifyCatalogProduct({
    ...sample,
    categories: undefined,
    title: "Samsung Galaxy Watch Ultra",
    description: { plain: "Premium smartwatch for fitness tracking." }
  });
  assert.ok(galaxyWatch);
  assert.equal(galaxyWatch.category, "mens-watches");

  const fallback = mapShopifyCatalogProduct({
    ...sample,
    categories: undefined,
    title: "A Product With No Category Clues",
    description: { plain: "A useful everyday item." }
  });
  assert.ok(fallback);
  assert.equal(fallback.category, "other-products");
});

test("renders Shopify CDN images directly without re-hosting", () => {
  const product = mapShopifyCatalogProduct(sample);
  assert.ok(product);
  assert.equal(product.thumbnail, "https://cdn.shopify.com/products/trail-runner-pro.jpg");
  assert.deepEqual(product.images, ["https://cdn.shopify.com/products/trail-runner-pro.jpg"]);
});

test("uses an image preview URL when a media record omits its direct URL", () => {
  const product = mapShopifyCatalogProduct({
    ...sample,
    media: [{ type: "image", preview: { url: "https://cdn.shopify.com/products/preview.jpg" } }]
  });
  assert.ok(product);
  assert.equal(product.thumbnail, "https://cdn.shopify.com/products/preview.jpg");
});
