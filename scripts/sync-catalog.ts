import { createHash } from "node:crypto";
import { readFile } from "node:fs/promises";
import { join } from "node:path";
import { dummyJsonCatalogAdapter } from "../lib/catalog/dummyjson";
import { loadCatalogFromSourceWithReport } from "../lib/catalog/source-adapter";
import { writeCatalogSnapshotAtomically } from "../lib/catalog/snapshot";
import { assertCatalogSnapshotContinuity, summarizeCatalogQuality } from "../lib/catalog/quality";
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

async function readLastKnownGood(): Promise<CatalogSnapshot | null> {
  try {
    return JSON.parse(await readFile(outputFile, "utf8")) as CatalogSnapshot;
  } catch {
    return null;
  }
}

async function main(): Promise<void> {
  const previous = await readLastKnownGood();

  try {
    const importedAt = new Date().toISOString();
    const importResult = await loadCatalogFromSourceWithReport(dummyJsonCatalogAdapter, {
      importedAt,
      requiredSnapshotMode: "full"
    });
    const importedProducts = importResult.products;
    const previousById = new Map((previous?.products ?? []).map(product => [product.id, product]));
    assertCatalogSnapshotContinuity(importedProducts.length, previous?.products?.length ?? 0);

    const products = importedProducts.map(product => {
      const prior = previousById.get(product.id);
      if (!prior?.localization) return product;
      const sourceFields = { title: product.title, description: product.description, category: product.category, brand: product.brand };
      const sourceHash = createHash("sha256").update(JSON.stringify(sourceFields)).digest("hex");
      return prior.localization.sourceHash === sourceHash
        ? { ...product, localization: prior.localization }
        : product;
    });

    const quality = summarizeCatalogQuality(products);
    const snapshot: CatalogSnapshot = {
      schemaVersion: 1,
      source: dummyJsonCatalogAdapter.source,
      generatedAt: importedAt,
      productCount: products.length,
      products
    };

    await writeCatalogSnapshotAtomically(snapshot, outputFile);
    console.log("Catalog snapshot generated:", products.length, "products at", importedAt);
    console.log("Catalog feed import report:", JSON.stringify(importResult.report));
    console.log("Catalog quality report:", JSON.stringify(quality));
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
