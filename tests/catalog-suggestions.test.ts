import test from "node:test";
import assert from "node:assert/strict";
import { buildCatalogSuggestions } from "../lib/catalog/suggestions";

const products = [
  { id: 1, title: "Aurora Headphones", category: "audio", brand: "Auralab" },
  { id: 2, title: "Aurora Speaker", category: "audio", brand: "Auralab" },
  { id: 3, title: "Studio Keyboard", category: "accessories", brand: "Keyworks" },
  { id: 4, title: "Portable Audio Player", category: "audio", brand: "SONORA" },
  { id: 5, title: "Aurora Case", category: "mobile-accessories", brand: " Auralab " }
];

test("returns no noisy suggestions for an empty or one-character query", () => {
  assert.deepEqual(buildCatalogSuggestions(products, ""), []);
  assert.deepEqual(buildCatalogSuggestions(products, "a"), []);
});

test("ranks title-prefix product matches ahead of broader matches", () => {
  const suggestions = buildCatalogSuggestions(products, "Aur");
  assert.equal(suggestions[0].kind, "product");
  assert.equal(suggestions[0].productId, 5);
  assert.ok(suggestions.some(item => item.kind === "brand" && item.value === "Auralab"));
});

test("deduplicates brands and categories after case and whitespace normalization", () => {
  const suggestions = buildCatalogSuggestions(products, "aur");
  assert.equal(suggestions.filter(item => item.kind === "brand").length, 1);
  assert.equal(suggestions.filter(item => item.kind === "category").length, 0);
  const audio = buildCatalogSuggestions(products, "aud").filter(item => item.kind === "category");
  assert.equal(audio.length, 1);
});

test("caps results and keeps the source product array unchanged", () => {
  const before = products.map(product => product.id);
  const bounded = buildCatalogSuggestions(products, "au", 2);
  assert.equal(bounded.length, 2);
  assert.deepEqual(products.map(product => product.id), before);
});

test("normalizes unicode compatibility forms and whitespace", () => {
  const unicodeProducts = [{ id: 1, title: "Ａｕｒｏｒａ Headphones", category: "audio" }];
  const suggestions = buildCatalogSuggestions(unicodeProducts, "Aurora");
  assert.equal(suggestions[0]?.productId, 1);
});

test("returns an empty array when the requested limit is zero or negative", () => {
  assert.deepEqual(buildCatalogSuggestions(products, "au", 0), []);
  assert.deepEqual(buildCatalogSuggestions(products, "au", -1), []);
});
