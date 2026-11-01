import { normalizeCatalog } from "./normalize";
import type { CatalogProduct } from "./types";

const ENDPOINT = "https://dummyjson.com/products?limit=100";

export async function fetchDummyJsonCatalog(signal?: AbortSignal): Promise<CatalogProduct[]> {
  const response = await fetch(ENDPOINT, { signal, next: { revalidate: 3600 } });
  if (!response.ok) throw new Error(`DummyJSON catalog request failed: ${response.status}`);
  const payload: unknown = await response.json();
  if (!payload || typeof payload !== "object" || !("products" in payload) || !Array.isArray(payload.products)) {
    throw new Error("DummyJSON returned an invalid catalog payload");
  }
  return normalizeCatalog(payload.products, "DummyJSON public demo API");
}
