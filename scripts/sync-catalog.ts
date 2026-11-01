import { mkdir, readFile, rename, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { normalizeCatalog } from "../lib/catalog/normalize";
import type { CatalogProduct } from "../lib/catalog/types";

type DummyJsonPayload = { products?: unknown[] };
type CatalogSnapshot = {
  schemaVersion: 1;
  source: string;
  generatedAt: string;
  productCount: number;
  products: CatalogProduct[];
};

const endpoint = "https://dummyjson.com/products?limit=100";
const outputDirectory = join(process.cwd(), "data");
const outputFile = join(outputDirectory, "catalog.json");
const temporaryFile = join(outputDirectory, "catalog.json.tmp");

async function readLastKnownGood(): Promise<CatalogSnapshot | null> {
  try {
    return JSON.parse(await readFile(outputFile, "utf8")) as CatalogSnapshot;
  } catch {
    return null;
  }
}

async function fetchCatalog(): Promise<unknown[]> {
  let lastError: unknown;
  for (let attempt = 1; attempt <= 3; attempt += 1) {
    try {
      const response = await fetch(endpoint, {
        headers: { accept: "application/json" },
        signal: AbortSignal.timeout(20_000)
      });
      if (!response.ok) throw new Error("Catalog source returned HTTP " + response.status);
      const payload = (await response.json()) as DummyJsonPayload;
      if (!payload || !Array.isArray(payload.products)) {
        throw new Error("Catalog source returned an invalid products collection");
      }
      if (payload.products.length === 0) throw new Error("Catalog source returned an empty collection");
      return payload.products;
    } catch (error) {
      lastError = error;
      if (attempt < 3) await new Promise(resolve => setTimeout(resolve, attempt * 1_000));
    }
  }
  throw lastError instanceof Error ? lastError : new Error("Catalog import failed");
}

async function main(): Promise<void> {
  await mkdir(outputDirectory, { recursive: true });
  const previous = await readLastKnownGood();

  try {
    const importedAt = new Date().toISOString();
    const rawProducts = await fetchCatalog();
    const products = normalizeCatalog(rawProducts, "DummyJSON public demo API", importedAt);
    if (products.length === 0) throw new Error("All source products were rejected during normalization");

    const snapshot: CatalogSnapshot = {
      schemaVersion: 1,
      source: "DummyJSON public demo API",
      generatedAt: importedAt,
      productCount: products.length,
      products
    };

    await writeFile(temporaryFile, JSON.stringify(snapshot, null, 2) + "\n", "utf8");
    await rename(temporaryFile, outputFile);
    console.log("Catalog snapshot generated:", products.length, "products at", importedAt);
  } catch (error) {
    if (previous?.products?.length) {
      console.warn("Catalog refresh failed; preserving last known-good snapshot:", error);
      console.warn("Snapshot timestamp:", previous.generatedAt, "products:", previous.productCount);
      return;
    }
    console.warn("Catalog refresh failed and no previous snapshot exists; storefront fixtures will be used:", error);
  }
}

main().catch(error => {
  console.error("Unexpected catalog sync failure:", error);
  process.exitCode = 1;
});
