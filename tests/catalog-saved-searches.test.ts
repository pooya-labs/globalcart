import test from "node:test";
import assert from "node:assert/strict";
import { hasSameSavedSearchFilters, normalizeSavedCatalogSearches, type SavedCatalogSearch } from "../lib/catalog/saved-searches";

const saved: SavedCatalogSearch = {
  id: "saved-1",
  name: "Budget headphones",
  query: "headphones",
  category: "audio",
  brand: "Auralab",
  minPrice: 10,
  maxPrice: 100,
  minRating: 4,
  sort: "rating",
  createdAt: "2026-11-02T00:00:00.000Z"
};

test("normalizes saved searches and ignores malformed browser storage entries", () => {
  const result = normalizeSavedCatalogSearches([saved, null, { ...saved, id: "", minRating: 8 }, { ...saved, id: "saved-2", name: "  Clean  " }]);
  assert.deepEqual(result.map(item => item.id), ["saved-1", "saved-2"]);
  assert.equal(result[1].name, "Clean");
});

test("saved search collection is bounded to twelve valid entries", () => {
  const items = Array.from({ length: 20 }, (_, index) => ({ ...saved, id: "saved-"+index }));
  const result = normalizeSavedCatalogSearches(items);
  assert.equal(result.length, 12);
  assert.equal(result[0].id, "saved-0");
});

test("duplicate detection compares all filter and sort dimensions but ignores case in text facets", () => {
  const candidate = { ...saved, query: " HEADPHONES ", brand: " auralab " };
  assert.equal(hasSameSavedSearchFilters(saved, candidate), true);
  assert.equal(hasSameSavedSearchFilters(saved, { ...candidate, minRating: 4.5 }), false);
  assert.equal(hasSameSavedSearchFilters(saved, { ...candidate, sort: "low" }), false);
});

test("invalid collections and invalid maximums return an empty list", () => {
  assert.deepEqual(normalizeSavedCatalogSearches(null), []);
  assert.deepEqual(normalizeSavedCatalogSearches([saved], 0), []);
});
