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
  assert.equal(product.category, "sporting-goods-athletics-running-shoes");
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
