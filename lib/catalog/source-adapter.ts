import { normalizeCatalog } from "./normalize";
import type { CatalogProduct, CatalogSourceAdapter, CatalogSourceRecord } from "./types";

export type CatalogImportOptions = {
  signal?: AbortSignal;
  importedAt?: string;
};

/**
 * Loads one supplier feed through its source-specific mapper, then applies the
 * shared catalog validation, provenance and de-duplication rules.
 *
 * Adapters must map supplier records to the canonical raw product shape.
 * They must never fabricate variant records when the upstream feed omits them.
 */
export async function loadCatalogFromSource(
  adapter: CatalogSourceAdapter,
  options: CatalogImportOptions = {}
): Promise<CatalogProduct[]> {
  const source = adapter.source.trim();
  if (!source) throw new Error("Catalog source adapter requires a non-empty source name");

  const records = await adapter.fetchRecords(options.signal);
  if (!Array.isArray(records) || records.length === 0) {
    throw new Error(`${source} returned an empty or invalid product collection`);
  }

  const mapped: CatalogSourceRecord[] = [];
  for (const record of records) {
    const product = adapter.mapProduct(record);
    if (product) mapped.push(product);
  }

  if (mapped.length === 0) {
    throw new Error(`${source} did not return any mappable product records`);
  }

  const importedAt = options.importedAt ?? new Date().toISOString();
  const products = normalizeCatalog(mapped, source, importedAt);
  if (products.length === 0) {
    throw new Error(`${source} returned no valid products after normalization`);
  }

  return products;
}
