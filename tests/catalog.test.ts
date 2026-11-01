import test from "node:test";
import assert from "node:assert/strict";
import { normalizeCatalog, normalizeProduct } from "../lib/catalog/normalize";
import { loadCatalogFromSource } from "../lib/catalog/source-adapter";
import type { CatalogSourceAdapter, CatalogSourceRecord } from "../lib/catalog/types";
import { syntheticSupplierRecords } from "./fixtures/synthetic-supplier-records";

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

test("normalizes optional variants with stable source identity and safe defaults", () => {
  const result = normalizeProduct({
    id: 8,
    title: "Configurable item",
    thumbnail: "https://cdn.example/item.jpg",
    variants: [
      { id: "red-s", sku: "ITEM-RED-S", title: "Red / Small", price: 15.25, stock: 4, attributes: { color: "red", size: "S" } },
      { id: "red-s", sku: "DUPLICATE", price: 99, stock: 99 },
      { sku: "BLUE-M", price: -2, stock: -1, isActive: false },
      { title: "Missing identity", price: 1, stock: 1 },
      null
    ]
  }, "fixture", "now");
  assert.ok(result);
  assert.equal(result.variants?.length, 2);
  assert.equal(result.variants?.[0].source, "fixture");
  assert.equal(result.variants?.[0].sourceId, "red-s");
  assert.equal(result.variants?.[0].attributes.color, "red");
  assert.equal(result.variants?.[0].isActive, true);
  assert.equal(result.variants?.[1].sourceId, "BLUE-M");
  assert.equal(result.variants?.[1].price, 0);
  assert.equal(result.variants?.[1].stock, 0);
  assert.equal(result.variants?.[1].isActive, false);
});

test("variant identity is scoped by source, matching the database unique key", () => {
  const result = normalizeProduct({
    id: 9,
    title: "Multi-source item",
    thumbnail: "https://cdn.example/item.jpg",
    variants: [
      { source: "supplier-a", sourceId: "shared-variant-id", sku: "A-1" },
      { source: "supplier-a", sourceId: "shared-variant-id", sku: "A-DUPLICATE" },
      { source: "supplier-b", sourceId: "shared-variant-id", sku: "B-1" },
      { source: "", sourceId: "parent-fallback", sku: "P-1" }
    ]
  }, "catalog-parent", "now");

  assert.ok(result);
  assert.equal(result.variants?.length, 3);
  assert.deepEqual(result.variants?.map(variant => variant.source), [
    "supplier-a",
    "supplier-b",
    "catalog-parent"
  ]);
  assert.deepEqual(result.variants?.map(variant => variant.sku), ["A-1", "B-1", "P-1"]);
});

test("supplier adapter maps its own payload into the canonical variant contract", async () => {
  const adapter: CatalogSourceAdapter = {
    source: "Synthetic Supplier Fixture",
    fetchRecords: async () => syntheticSupplierRecords,
    mapProduct: (record): CatalogSourceRecord | null => {
      if (!record || typeof record !== "object" || Array.isArray(record)) return null;
      const supplier = record as Record<string, unknown>;
      const options = Array.isArray(supplier.options) ? supplier.options : [];
      return {
        id: supplier.productCode,
        title: supplier.displayName,
        thumbnail: supplier.image,
        category: supplier.department,
        price: supplier.unitPrice,
        stock: supplier.availableQuantity,
        variants: options.map(option => {
          if (!option || typeof option !== "object" || Array.isArray(option)) return null;
          const variant = option as Record<string, unknown>;
          return {
            id: variant.variantCode,
            sku: variant.merchantSku,
            title: variant.label,
            price: variant.unitPrice,
            stock: variant.availableQuantity,
            attributes: variant.specifications,
            isActive: variant.enabled
          };
        }).filter(Boolean)
      };
    }
  };

  const products = await loadCatalogFromSource(adapter, {
    importedAt: "2026-11-02T00:00:00.000Z"
  });

  assert.equal(products.length, 1);
  assert.equal(products[0].source, "Synthetic Supplier Fixture");
  assert.equal(products[0].sourceId, "9001");
  assert.equal(products[0].variants?.length, 2);
  assert.equal(products[0].variants?.[0].source, "Synthetic Supplier Fixture");
  assert.equal(products[0].variants?.[0].sourceId, "BOTTLE-500-BLUE");
  assert.equal(products[0].variants?.[0].sku, "GC-BTL-500-BLU");
  assert.equal(products[0].variants?.[0].attributes.color, "blue");
  assert.equal(products[0].variants?.[1].isActive, false);
  assert.equal(products[0].importedAt, "2026-11-02T00:00:00.000Z");
});

test("supplier adapter rejects empty or wholly unmappable feeds", async () => {
  const emptyAdapter: CatalogSourceAdapter = {
    source: "Empty Fixture",
    fetchRecords: async () => [],
    mapProduct: () => null
  };
  await assert.rejects(
    () => loadCatalogFromSource(emptyAdapter),
    /empty or invalid product collection/
  );

  const unmappableAdapter: CatalogSourceAdapter = {
    source: "Unmappable Fixture",
    fetchRecords: async () => [null, "not-a-product"],
    mapProduct: () => null
  };
  await assert.rejects(
    () => loadCatalogFromSource(unmappableAdapter),
    /did not return any mappable product records/
  );
});
