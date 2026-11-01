import test from "node:test";
import assert from "node:assert/strict";
import { normalizeCatalog, normalizeProduct } from "../lib/catalog/normalize";
import { loadCatalogFromSource, loadCatalogFromSourceWithReport } from "../lib/catalog/source-adapter";
import { dummyJsonCatalogAdapter } from "../lib/catalog/dummyjson";
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


test("variant tuple identity cannot collide when source or source ID contains a colon", () => {
  const result = normalizeProduct({
    id: 10,
    title: "Collision-resistant identity",
    thumbnail: "https://cdn.example/item.jpg",
    variants: [
      { source: "supplier:one", sourceId: "id", sku: "FIRST" },
      { source: "supplier", sourceId: "one:id", sku: "SECOND" }
    ]
  }, "catalog-parent", "now");

  assert.ok(result);
  assert.equal(result.variants?.length, 2);
  assert.deepEqual(result.variants?.map(variant => [variant.source, variant.sourceId]), [
    ["supplier:one", "id"],
    ["supplier", "one:id"]
  ]);
});

test("returns aggregate feed diagnostics without exposing source records", async () => {
  const adapter: CatalogSourceAdapter = {
    source: "Diagnostics Fixture",
    snapshotMode: "full",
    fetchRecords: async () => [
      { id: 1, title: "First", thumbnail: "https://cdn.example/1.jpg" },
      { id: 1, title: "Duplicate", thumbnail: "https://cdn.example/duplicate.jpg" },
      { id: 2, title: "Missing image" },
      null,
      { id: 3, title: "Third", thumbnail: "https://cdn.example/3.jpg" }
    ],
    mapProduct: record => record && typeof record === "object" && !Array.isArray(record)
      ? record as CatalogSourceRecord
      : null
  };
  const result = await loadCatalogFromSourceWithReport(adapter, {
    importedAt: "2026-11-02T00:00:00.000Z"
  });
  assert.deepEqual(result.products.map(product => product.sourceId), ["1", "3"]);
  assert.deepEqual(result.report, {
    source: "Diagnostics Fixture",
    snapshotMode: "full",
    importedAt: "2026-11-02T00:00:00.000Z",
    fetchedRecordCount: 5,
    mappedRecordCount: 4,
    unmappedRecordCount: 1,
    unmappedFraction: 0.2,
    maxUnmappedFraction: 0.2,
    normalizedProductCount: 2,
    discardedAfterNormalizationCount: 2
  });
  assert.equal("records" in result.report, false);
});

test("supplier adapter maps its own payload into the canonical variant contract", async () => {
  const adapter: CatalogSourceAdapter = {
    source: "Synthetic Supplier Fixture",
    snapshotMode: "full",
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


test("rejects suspicious full-snapshot shrinkage but allows an approved threshold", async () => {
  const { assertCatalogSnapshotContinuity } = await import("../lib/catalog/quality");

  assert.doesNotThrow(() => assertCatalogSnapshotContinuity(50, 100));
  assert.doesNotThrow(() => assertCatalogSnapshotContinuity(75, 100, 0.25));
  assert.throws(
    () => assertCatalogSnapshotContinuity(49, 100),
    /product count dropped from 100 to 49; minimum accepted is 50/
  );
  assert.doesNotThrow(() => assertCatalogSnapshotContinuity(1, 0));
});

test("validates the configured catalog continuity threshold", async () => {
  const { assertCatalogSnapshotContinuity } = await import("../lib/catalog/quality");
  assert.throws(() => assertCatalogSnapshotContinuity(10, 10, -0.1), RangeError);
  assert.throws(() => assertCatalogSnapshotContinuity(10, 10, 1), RangeError);
  assert.throws(() => assertCatalogSnapshotContinuity(-1, 10), RangeError);
  assert.throws(() => assertCatalogSnapshotContinuity(1.5, 10), RangeError);
});

test("reports catalog completeness without treating product-only feeds as invalid", async () => {
  const { summarizeCatalogQuality } = await import("../lib/catalog/quality");
  const products = normalizeCatalog([
    {
      id: 21, title: "Variant item", description: "Configured", category: "accessories",
      price: 10, stock: 4, thumbnail: "https://cdn.example/variant.jpg",
      variants: [
        { id: "v1", sku: "V-1", price: 10, stock: 2 },
        { id: "v2", sku: "V-2", price: 11, stock: 2 }
      ]
    },
    {
      id: 22, title: "Product-only item", description: "", category: "uncategorized",
      price: 0, stock: 0, thumbnail: "https://cdn.example/simple.jpg"
    }
  ], "quality-fixture", "now");

  assert.deepEqual(summarizeCatalogQuality(products), {
    productCount: 2,
    variantCount: 2,
    productsWithVariants: 1,
    productsWithoutVariants: 1,
    zeroPriceProducts: 1,
    zeroStockProducts: 1,
    missingDescriptionProducts: 1,
    uncategorizedProducts: 1
  });
});


test("incremental feeds cannot replace a complete catalog snapshot", async () => {
  let fetchCalls = 0;
  const adapter: CatalogSourceAdapter = {
    source: "Incremental Supplier Fixture",
    snapshotMode: "incremental",
    fetchRecords: async () => {
      fetchCalls += 1;
      return [{
        id: 31,
        title: "Changed record",
        thumbnail: "https://cdn.example/changed.jpg",
        price: 20,
        stock: 3
      }];
    },
    mapProduct: record => record && typeof record === "object" && !Array.isArray(record)
      ? record as CatalogSourceRecord
      : null
  };

  await assert.rejects(
    () => loadCatalogFromSource(adapter, { requiredSnapshotMode: "full" }),
    /incremental.*full.*required/
  );
  assert.equal(fetchCalls, 0, "mode mismatch must be rejected before fetching records");

  const incrementalRecords = await loadCatalogFromSource(adapter);
  assert.equal(incrementalRecords.length, 1, "incremental records may be normalized for a dedicated merger");
  assert.equal(incrementalRecords[0].title, "Changed record");
});

test("rejects feeds whose source records mostly fail mapping", async () => {
  let mapCalls = 0;
  const adapter: CatalogSourceAdapter = {
    source: "Partially Broken Fixture",
    snapshotMode: "full",
    fetchRecords: async () => [1, 2, 3, 4, 5],
    mapProduct: record => {
      mapCalls += 1;
      if (record === 1 || record === 2) return null;
      return { id: record, title: `Product ${record}`, thumbnail: `https://cdn.example/${record}.jpg` };
    }
  };
  await assert.rejects(() => loadCatalogFromSource(adapter), /rejected 2\/5 source records.*40\.0%.*maximum allowed is 20\.0%/);
  assert.equal(mapCalls, 5);
});

test("allows a small mapping rejection rate and supports an explicit threshold", async () => {
  const adapter: CatalogSourceAdapter = {
    source: "Mostly Valid Fixture",
    snapshotMode: "full",
    fetchRecords: async () => [1, 2, 3, 4, 5],
    mapProduct: record => record === 1 ? null : ({
      id: record, title: `Product ${record}`, thumbnail: `https://cdn.example/${record}.jpg`
    })
  };
  const products = await loadCatalogFromSource(adapter);
  assert.equal(products.length, 4);
  await assert.rejects(() => loadCatalogFromSource(adapter, { maxUnmappedFraction: 0.1 }), /rejected 1\/5 source records/);
  await assert.doesNotReject(() => loadCatalogFromSource(adapter, { maxUnmappedFraction: 0.2 }));
});

test("validates the source mapping rejection threshold", async () => {
  const adapter: CatalogSourceAdapter = {
    source: "Threshold Fixture", snapshotMode: "full",
    fetchRecords: async () => [{ id: 1, title: "A", thumbnail: "https://cdn.example/a.jpg" }],
    mapProduct: record => record as CatalogSourceRecord
  };
  for (const value of [-0.1, 1, Number.NaN, Number.POSITIVE_INFINITY]) {
    await assert.rejects(() => loadCatalogFromSource(adapter, { maxUnmappedFraction: value }), RangeError);
  }
});

test("combines caller cancellation with a request timeout", async () => {
  const originalFetch = globalThis.fetch;
  const controller = new AbortController();
  const observed: { signal?: AbortSignal | null } = {};
  globalThis.fetch = (async (_input: RequestInfo | URL, init?: RequestInit) => {
    observed.signal = init?.signal ?? null;
    return { ok: true, json: async () => ({ products: [{ id: 41 }] }) } as Response;
  }) as typeof fetch;
  try {
    await dummyJsonCatalogAdapter.fetchRecords(controller.signal);
    assert.ok(observed.signal instanceof AbortSignal);
    assert.notEqual(observed.signal, controller.signal, "request deadline must remain active when caller supplies a signal");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("caller cancellation aborts the request without retrying", async () => {
  const originalFetch = globalThis.fetch;
  const controller = new AbortController();
  let calls = 0;
  globalThis.fetch = (async (_input: RequestInfo | URL, init?: RequestInit) => {
    calls += 1;
    const requestSignal = init?.signal;
    return await new Promise<Response>((_resolve, reject) => {
      if (requestSignal?.aborted) { reject(requestSignal.reason ?? new Error("aborted")); return; }
      requestSignal?.addEventListener("abort", () => reject(requestSignal.reason ?? new Error("aborted")), { once: true });
    });
  }) as typeof fetch;
  try {
    const pending = dummyJsonCatalogAdapter.fetchRecords(controller.signal);
    controller.abort(new Error("caller cancelled"));
    await assert.rejects(pending, /caller cancelled/);
    assert.equal(calls, 1, "explicit cancellation must not be retried");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("does not retry permanent HTTP errors", async () => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  globalThis.fetch = (async () => {
    calls += 1;
    return { ok: false, status: 404 } as Response;
  }) as typeof fetch;
  try {
    await assert.rejects(dummyJsonCatalogAdapter.fetchRecords(), /request failed: 404/);
    assert.equal(calls, 1);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("retries transient server errors within the bounded attempt count", async () => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  globalThis.fetch = (async () => {
    calls += 1;
    if (calls < 3) return { ok: false, status: 503 } as Response;
    return { ok: true, json: async () => ({ products: [{ id: 42 }] }) } as Response;
  }) as typeof fetch;
  try {
    const records = await dummyJsonCatalogAdapter.fetchRecords();
    assert.equal(records.length, 1);
    assert.equal(calls, 3);
  } finally {
    globalThis.fetch = originalFetch;
  }
});
test("supplier adapter rejects empty or wholly unmappable feeds", async () => {
  const emptyAdapter: CatalogSourceAdapter = {
    source: "Empty Fixture",
    snapshotMode: "full",
    fetchRecords: async () => [],
    mapProduct: () => null
  };
  await assert.rejects(
    () => loadCatalogFromSource(emptyAdapter),
    /empty or invalid product collection/
  );

  const unmappableAdapter: CatalogSourceAdapter = {
    source: "Unmappable Fixture",
    snapshotMode: "full",
    fetchRecords: async () => [null, "not-a-product"],
    mapProduct: () => null
  };
  await assert.rejects(
    () => loadCatalogFromSource(unmappableAdapter),
    /did not return any mappable product records/
  );
});
