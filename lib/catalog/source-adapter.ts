import { normalizeCatalog } from "./normalize";
import type { CatalogProduct, CatalogSnapshotMode, CatalogSourceAdapter, CatalogSourceRecord } from "./types";

export type CatalogImportOptions = {
  signal?: AbortSignal;
  importedAt?: string;
  /** Reject a feed unless it declares the requested completeness semantics. */
  requiredSnapshotMode?: CatalogSnapshotMode;
  /** Maximum fraction of source records that may fail source-specific mapping. Defaults to 20%. */
  maxUnmappedFraction?: number;
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

  if (options.requiredSnapshotMode && adapter.snapshotMode !== options.requiredSnapshotMode) {
    throw new Error(
      `${source} feed mode is '${adapter.snapshotMode}', but '${options.requiredSnapshotMode}' is required`
    );
  }

  const records = await adapter.fetchRecords(options.signal);
  if (!Array.isArray(records) || records.length === 0) {
    throw new Error(`${source} returned an empty or invalid product collection`);
  }

  const maxUnmappedFraction = options.maxUnmappedFraction ?? 0.2;
  if (!Number.isFinite(maxUnmappedFraction) || maxUnmappedFraction < 0 || maxUnmappedFraction >= 1) {
    throw new RangeError("maxUnmappedFraction must be a finite number in the range [0, 1)");
  }

  const mapped: CatalogSourceRecord[] = [];
  for (const record of records) {
    const product = adapter.mapProduct(record);
    if (product) mapped.push(product);
  }

  if (mapped.length === 0) {
    throw new Error(`${source} did not return any mappable product records`);
  }

  const unmappedCount = records.length - mapped.length;
  const unmappedFraction = unmappedCount / records.length;
  if (unmappedFraction > maxUnmappedFraction) {
    throw new Error(
      `${source} rejected ${unmappedCount}/${records.length} source records during mapping (${(unmappedFraction * 100).toFixed(1)}%); maximum allowed is ${(maxUnmappedFraction * 100).toFixed(1)}%`
    );
  }

  const importedAt = options.importedAt ?? new Date().toISOString();
  const products = normalizeCatalog(mapped, source, importedAt);
  if (products.length === 0) {
    throw new Error(`${source} returned no valid products after normalization`);
  }

  return products;
}
