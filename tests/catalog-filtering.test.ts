import test from "node:test";
import assert from "node:assert/strict";
import { filterCatalogProducts } from "../lib/catalog/filtering";

const products = [
  { id: 1, title: "Aurora Headphones", description: "Studio audio", category: "Audio", brand: "Auralab", price: 49.99, rating: 4.8 },
  { id: 2, title: "Budget Headphones", description: "Entry audio", category: "Audio", brand: "Auralab", price: 29.99, rating: 3.5 },
  { id: 3, title: "Auralab Speaker", description: "Wireless sound", category: "Audio", brand: "Auralab", price: 129.99, rating: 4.2 },
  { id: 4, title: "Studio Keyboard", description: "Mechanical typing", category: "Accessories", brand: "Keyworks", price: 89.5, rating: 4.9 }
];

test("combines category, text, brand, price and rating filters", () => {
  const result = filterCatalogProducts(products, {
    category: "Audio",
    query: "headphones",
    brand: " AURALAB ",
    minPrice: 30,
    maxPrice: 100,
    minRating: 4
  });
  assert.deepEqual(result.map(product => product.id), [1]);
});

test("brand matching ignores case and whitespace; minimum rating is inclusive", () => {
  const result = filterCatalogProducts(products, {
    brand: " aUrAlAb ",
    minRating: 4.2
  });
  assert.deepEqual(result.map(product => product.id), [1, 3]);
});

test("empty filters return all products without mutating the catalog", () => {
  const originalIds = products.map(product => product.id);
  const result = filterCatalogProducts(products, {});
  assert.deepEqual(result.map(product => product.id), originalIds);
  assert.deepEqual(products.map(product => product.id), originalIds);
  assert.notEqual(result, products);
});
