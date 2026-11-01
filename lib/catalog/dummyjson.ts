import { loadCatalogFromSource } from "./source-adapter";
import type { CatalogProduct, CatalogSourceAdapter, CatalogSourceRecord } from "./types";

const ENDPOINT = "https://dummyjson.com/products?limit=100";
const SOURCE = "DummyJSON public demo API";

function isRecord(value: unknown): value is CatalogSourceRecord {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

async function fetchDummyJsonRecords(signal?: AbortSignal): Promise<unknown[]> {
  let lastError: unknown;
  for (let attempt = 1; attempt <= 3; attempt += 1) {
    try {
      const response = await fetch(ENDPOINT, {
        headers: { accept: "application/json" },
        signal: signal ?? AbortSignal.timeout(20_000),
        next: { revalidate: 3600 }
      });
      if (!response.ok) throw new Error("DummyJSON catalog request failed: " + response.status);
      const payload: unknown = await response.json();
      if (!isRecord(payload) || !Array.isArray(payload.products)) {
        throw new Error("DummyJSON returned an invalid products collection");
      }
      if (payload.products.length === 0) throw new Error("DummyJSON returned an empty products collection");
      return payload.products;
    } catch (error) {
      lastError = error;
      if (signal?.aborted) throw error;
      if (attempt < 3) await new Promise(resolve => setTimeout(resolve, attempt * 1_000));
    }
  }
  throw lastError instanceof Error ? lastError : new Error("DummyJSON catalog import failed");
}

/** DummyJSON is a sample source and currently supplies no variant records. */
export const dummyJsonCatalogAdapter: CatalogSourceAdapter = {
  source: SOURCE,
  fetchRecords: fetchDummyJsonRecords,
  mapProduct: record => isRecord(record) ? record : null
};

export async function fetchDummyJsonCatalog(signal?: AbortSignal): Promise<CatalogProduct[]> {
  return loadCatalogFromSource(dummyJsonCatalogAdapter, { signal });
}
