import { createHash } from "node:crypto";
import { mkdir, readFile, rename, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { dummyJsonCatalogAdapter } from "../lib/catalog/dummyjson";
import { loadCatalogFromSource } from "../lib/catalog/source-adapter";
import type { CatalogProduct } from "../lib/catalog/types";

type CatalogSnapshot = {
  schemaVersion: 1;
  source: string;
  generatedAt: string;
  productCount: number;
  products: CatalogProduct[];
};

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

async function main(): Promise<void> {
  await mkdir(outputDirectory, { recursive: true });
  const previous = await readLastKnownGood();

  try {
    const importedAt = new Date().toISOString();
    const importedProducts = await loadCatalogFromSource(dummyJsonCatalogAdapter, { importedAt });
    const previousById = new Map((previous?.products ?? []).map(product => [product.id, product]));
    const products = importedProducts.map(product => {
      const prior = previousById.get(product.id);
      if (!prior?.localization) return product;
      const sourceFields = { title: product.title, description: product.description, category: product.category, brand: product.brand };
      const sourceHash = createHash("sha256").update(JSON.stringify(sourceFields)).digest("hex");
      return prior.localization.sourceHash === sourceHash
        ? { ...product, localization: prior.localization }
        : product;
    });

    const snapshot: CatalogSnapshot = {
      schemaVersion: 1,
      source: dummyJsonCatalogAdapter.source,
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
