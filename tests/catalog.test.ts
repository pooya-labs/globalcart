import test from "node:test";
import assert from "node:assert/strict";
import { normalizeCatalog, normalizeProduct } from "../lib/catalog/normalize";

test("normalizes a valid source record and keeps provenance", () => {
  const result = normalizeProduct({ id: 7, title: "Headphones", thumbnail: "https://cdn.example/item.jpg", price: 12.5, discountPercentage: 10 }, "fixture", "2026-11-01T00:00:00.000Z");
  assert.ok(result);
  assert.equal(result.sourceId, "7");
  assert.equal(result.source, "fixture");
  assert.equal(result.importedAt, "2026-11-01T00:00:00.000Z");
  assert.equal(result.price, 12.5);
});

test("rejects records without essential product identity or image", () => {
  assert.equal(normalizeProduct({ id: 1, title: "Missing image" }, "fixture", "now"), null);
  assert.equal(normalizeProduct({ id: 1, thumbnail: "https://cdn.example/a.jpg" }, "fixture", "now"), null);
});

test("deduplicates by source ID and ignores malformed records", () => {
  const rows = [
    { id: 1, title: "A", thumbnail: "https://cdn.example/a.jpg" },
    { id: 1, title: "Duplicate", thumbnail: "https://cdn.example/b.jpg" },
    null,
    { id: 2, title: "B", thumbnail: "https://cdn.example/c.jpg" }
  ];
  const result = normalizeCatalog(rows, "fixture", "now");
  assert.equal(result.length, 2);
  assert.equal(result[0].title, "A");
});
