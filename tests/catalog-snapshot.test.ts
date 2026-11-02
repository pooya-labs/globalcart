import test from "node:test";
import assert from "node:assert/strict";
import { mkdir, mkdtemp, readFile, readdir, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { writeCatalogSnapshotAtomically } from "../lib/catalog/snapshot";

test("concurrent snapshot writers publish complete JSON without sharing temp files", async () => {
  const directory = await mkdtemp(join(tmpdir(), "globalcart-snapshot-"));
  const outputFile = join(directory, "nested", "catalog.json");

  try {
    const snapshots = Array.from({ length: 12 }, (_, sequence) => ({
      sequence,
      products: Array.from({ length: 24 }, (_, index) => ({
        id: index + 1,
        title: `snapshot-${sequence}-product-${index + 1}`
      }))
    }));

    await Promise.all(
      snapshots.map(snapshot => writeCatalogSnapshotAtomically(snapshot, outputFile))
    );

    const persisted = JSON.parse(await readFile(outputFile, "utf8")) as typeof snapshots[number];
    const matchingSnapshot = snapshots.find(snapshot => snapshot.sequence === persisted.sequence);
    assert.ok(matchingSnapshot, "the published file must be one complete submitted snapshot");
    assert.deepEqual(persisted, matchingSnapshot);
    assert.deepEqual(await readdir(join(directory, "nested")), ["catalog.json"]);
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});

test("failed publication cleans its unique temporary file", async () => {
  const directory = await mkdtemp(join(tmpdir(), "globalcart-snapshot-failure-"));
  const outputFile = join(directory, "catalog.json");

  try {
    await mkdir(outputFile);
    await assert.rejects(() => writeCatalogSnapshotAtomically({ products: [] }, outputFile));
    assert.deepEqual(await readdir(directory), ["catalog.json"]);
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});

test("non-serializable snapshots fail before touching the last-known-good file", async () => {
  const directory = await mkdtemp(join(tmpdir(), "globalcart-snapshot-invalid-"));
  const outputFile = join(directory, "catalog.json");

  try {
    await writeFile(outputFile, '{"source":"last-known-good"}\n', "utf8");
    const circular: Record<string, unknown> = {};
    circular.self = circular;

    await assert.rejects(() => writeCatalogSnapshotAtomically(circular, outputFile), TypeError);
    assert.equal(await readFile(outputFile, "utf8"), '{"source":"last-known-good"}\n');
    assert.deepEqual(await readdir(directory), ["catalog.json"]);
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});


test("a missing previous snapshot is treated as a first import", async () => {
  const { readCatalogSnapshot } = await import("../lib/catalog/snapshot");
  const directory = await mkdtemp(join(tmpdir(), "globalcart-snapshot-missing-"));
  try {
    assert.equal(await readCatalogSnapshot(join(directory, "catalog.json")), null);
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});

test("malformed JSON is not silently treated as a missing snapshot", async () => {
  const { readCatalogSnapshot } = await import("../lib/catalog/snapshot");
  const directory = await mkdtemp(join(tmpdir(), "globalcart-snapshot-malformed-"));
  const outputFile = join(directory, "catalog.json");
  try {
    await writeFile(outputFile, '{"schemaVersion":', "utf8");
    await assert.rejects(() => readCatalogSnapshot(outputFile), /not valid JSON/);
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});

test("snapshot validation rejects count mismatches and malformed products", async () => {
  const { readCatalogSnapshot } = await import("../lib/catalog/snapshot");
  const directory = await mkdtemp(join(tmpdir(), "globalcart-snapshot-schema-"));
  const outputFile = join(directory, "catalog.json");
  const metadata = {
    schemaVersion: 1,
    source: "fixture",
    generatedAt: "2026-11-02T00:00:00.000Z"
  };
  try {
    await writeFile(outputFile, JSON.stringify({ ...metadata, productCount: 1, products: [] }), "utf8");
    await assert.rejects(() => readCatalogSnapshot(outputFile), /schema validation/);
    await writeFile(outputFile, JSON.stringify({
      ...metadata,
      productCount: 1,
      products: [{ id: 1, title: "incomplete" }]
    }), "utf8");
    await assert.rejects(() => readCatalogSnapshot(outputFile), /schema validation/);
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});


function validSnapshotProduct(overrides: Record<string, unknown> = {}): Record<string, unknown> {
  return {
    id: 10,
    title: "Snapshot product",
    description: "Valid fixture product",
    category: "accessories",
    price: 19.99,
    discountPercentage: 10,
    rating: 4.5,
    stock: 12,
    brand: "Fixture",
    thumbnail: "https://cdn.example/product.jpg",
    images: [],
    source: "fixture",
    sourceId: "product-10",
    sourceUrl: "",
    importedAt: "2026-11-02T00:00:00.000Z",
    ...overrides
  };
}

function wrapSnapshot(products: Record<string, unknown>[]) {
  return {
    schemaVersion: 1,
    source: "fixture",
    generatedAt: "2026-11-02T00:00:00.000Z",
    productCount: products.length,
    products
  };
}

test("snapshot validation rejects duplicate product IDs and source identities", async () => {
  const { readCatalogSnapshot } = await import("../lib/catalog/snapshot");
  const directory = await mkdtemp(join(tmpdir(), "globalcart-product-identity-"));
  const outputFile = join(directory, "catalog.json");
  try {
    await writeFile(outputFile, JSON.stringify(wrapSnapshot([
      validSnapshotProduct({ id: 10, sourceId: "first" }),
      validSnapshotProduct({ id: 10, sourceId: "second" })
    ])), "utf8");
    await assert.rejects(() => readCatalogSnapshot(outputFile), /schema validation/);

    await writeFile(outputFile, JSON.stringify(wrapSnapshot([
      validSnapshotProduct({ id: 10, sourceId: "shared" }),
      validSnapshotProduct({ id: 11, sourceId: "shared" })
    ])), "utf8");
    await assert.rejects(() => readCatalogSnapshot(outputFile), /schema validation/);
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});

test("snapshot validation rejects products attributed to a different source", async () => {
  const { readCatalogSnapshot } = await import("../lib/catalog/snapshot");
  const directory = await mkdtemp(join(tmpdir(), "globalcart-product-source-"));
  const outputFile = join(directory, "catalog.json");
  try {
    await writeFile(outputFile, JSON.stringify(wrapSnapshot([
      validSnapshotProduct({ source: "another-source" })
    ])), "utf8");
    await assert.rejects(() => readCatalogSnapshot(outputFile), /schema validation/);
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});

test("snapshot validation enforces PostgreSQL numeric and integer bounds", async () => {
  const { readCatalogSnapshot } = await import("../lib/catalog/snapshot");
  const directory = await mkdtemp(join(tmpdir(), "globalcart-db-bounds-"));
  const outputFile = join(directory, "catalog.json");
  const invalidProducts = [
    { price: -0.01 },
    { price: 10_000_000_000 },
    { price: 1.001 },
    { discountPercentage: 100.01 },
    { discountPercentage: "10" },
    { rating: "4.5" },
    { discountPercentage: 1.001 },
    { rating: 5.01 },
    { rating: 4.501 },
    { stock: 1.5 },
    { stock: 2_147_483_648 },
    { id: Number.MAX_SAFE_INTEGER + 1 }
  ];
  try {
    for (const invalid of invalidProducts) {
      await writeFile(outputFile, JSON.stringify(wrapSnapshot([
        validSnapshotProduct(invalid)
      ])), "utf8");
      await assert.rejects(
        () => readCatalogSnapshot(outputFile),
        /schema validation/,
        "invalid database-bound product values must be rejected"
      );
    }
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});

test("snapshot validation rejects duplicate variant source identities across products", async () => {
  const { readCatalogSnapshot } = await import("../lib/catalog/snapshot");
  const directory = await mkdtemp(join(tmpdir(), "globalcart-variant-identity-"));
  const outputFile = join(directory, "catalog.json");
  const variant = {
    sku: "FIXTURE-1",
    title: "Variant",
    price: 12.5,
    stock: 3,
    attributes: { color: "blue" },
    source: "supplier-fixture",
    sourceId: "variant-shared",
    isActive: true
  };
  try {
    await writeFile(outputFile, JSON.stringify(wrapSnapshot([
      validSnapshotProduct({ variants: [variant] }),
      validSnapshotProduct({ id: 11, sourceId: "product-11", variants: [variant] })
    ])), "utf8");
    await assert.rejects(() => readCatalogSnapshot(outputFile), /schema validation/);
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});

test("snapshot validation enforces variant price and stock constraints", async () => {
  const { readCatalogSnapshot } = await import("../lib/catalog/snapshot");
  const directory = await mkdtemp(join(tmpdir(), "globalcart-variant-bounds-"));
  const outputFile = join(directory, "catalog.json");
  const baseVariant = {
    sku: "FIXTURE-1",
    title: "Variant",
    price: 12.5,
    stock: 3,
    attributes: { color: "blue" },
    source: "supplier-fixture",
    sourceId: "variant-1",
    isActive: true
  };
  try {
    for (const overrides of [{ price: -1 }, { price: 1.001 }, { stock: 2.5 }, { stock: 2_147_483_648 }, { sourceId: " " }]) {
      await writeFile(outputFile, JSON.stringify(wrapSnapshot([
        validSnapshotProduct({ variants: [{ ...baseVariant, ...overrides }] })
      ])), "utf8");
      await assert.rejects(() => readCatalogSnapshot(outputFile), /schema validation/);
    }
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});


test("the shared snapshot assertion rejects invalid data before publication", async () => {
  const { assertCatalogSnapshot } = await import("../lib/catalog/snapshot");
  const invalid = wrapSnapshot([
    validSnapshotProduct({ price: 1.001 })
  ]);
  assert.throws(() => assertCatalogSnapshot(invalid, "publication"), /schema validation.*publication/);
  assert.doesNotThrow(() => assertCatalogSnapshot(wrapSnapshot([
    validSnapshotProduct({ price: 19.99, stock: 4 })
  ])));
});


test("snapshot validation rejects duplicate non-empty variant SKUs", async () => {
  const { readCatalogSnapshot } = await import("../lib/catalog/snapshot");
  const directory = await mkdtemp(join(tmpdir(), "globalcart-variant-sku-"));
  const outputFile = join(directory, "catalog.json");
  const variant = {
    sku: "GC-SHARED-SKU",
    title: "Variant",
    price: 12.5,
    stock: 3,
    attributes: {},
    source: "supplier-fixture",
    sourceId: "variant-1",
    isActive: true
  };
  try {
    await writeFile(outputFile, JSON.stringify(wrapSnapshot([
      validSnapshotProduct({ variants: [variant] }),
      validSnapshotProduct({
        id: 11,
        sourceId: "product-11",
        variants: [{ ...variant, sourceId: "variant-2" }]
      })
    ])), "utf8");
    await assert.rejects(() => readCatalogSnapshot(outputFile), /schema validation/);
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});
