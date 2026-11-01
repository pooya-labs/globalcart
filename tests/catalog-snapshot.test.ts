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
