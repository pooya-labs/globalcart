import test from "node:test";
import assert from "node:assert/strict";
import {
  getCatalogRecommendationReason,
  getCatalogRecommendationScore,
  rankCatalogRecommendations
} from "../lib/catalog/recommendations";

const products = [
  { id: 1, title: "Premium Headphones", rating: 5, discountPercentage: 0, price: 100 },
  { id: 2, title: "Value Speaker", rating: 4, discountPercentage: 40, price: 60 },
  { id: 3, title: "Balanced Keyboard", rating: 4.5, discountPercentage: 10, price: 80 },
  { id: 4, title: "Editor's Pick", rating: 4.8, discountPercentage: 15, price: 120 }
];

test("scores use the documented 70/30 rating and discount weights", () => {
  assert.equal(getCatalogRecommendationScore(products[0]), 0.7);
  assert.equal(getCatalogRecommendationScore(products[1]), 0.8);
  assert.equal(getCatalogRecommendationScore({ ...products[1], discountPercentage: 100 }), 0.86);
});

test("recommendation reasons identify the strongest visible catalog signals", () => {
  assert.equal(getCatalogRecommendationReason(products[3]), "rating-and-discount");
  assert.equal(getCatalogRecommendationReason(products[0]), "rating");
  assert.equal(getCatalogRecommendationReason(products[1]), "discount");
  assert.equal(getCatalogRecommendationReason({ ...products[2], rating: 4.2, discountPercentage: 5 }), "balanced");
});

test("ranking is deterministic, stable and does not mutate the original product array", () => {
  const original = products.map(product => product.id);
  const ranked = rankCatalogRecommendations(products);
  assert.deepEqual(ranked.map(product => product.id), [2, 4, 1, 3]);
  assert.deepEqual(products.map(product => product.id), original);
  assert.notEqual(ranked, products);
});

test("invalid rating and discount inputs are safely bounded", () => {
  assert.equal(getCatalogRecommendationScore({ id: 9, rating: Number.NaN, discountPercentage: Number.POSITIVE_INFINITY, price: 1 }), 0);
  assert.equal(getCatalogRecommendationScore({ id: 10, rating: 6, discountPercentage: -10, price: 1 }), 0.7);
});
