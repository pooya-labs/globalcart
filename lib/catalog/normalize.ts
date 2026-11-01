import type { CatalogProduct } from "./types";

type UnknownRecord = Record<string, unknown>;

const numberOr = (value: unknown, fallback = 0): number => {
  const parsed = typeof value === "number" ? value : Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const stringOr = (value: unknown, fallback = ""): string =>
  typeof value === "string" ? value.trim() : fallback;

export function normalizeProduct(raw: UnknownRecord, source: string, importedAt: string): CatalogProduct | null {
  const id = numberOr(raw.id, -1);
  const title = stringOr(raw.title);
  const thumbnail = stringOr(raw.thumbnail);
  if (id < 0 || !title || !thumbnail) return null;

  const images = Array.isArray(raw.images)
    ? raw.images.filter((item): item is string => typeof item === "string" && /^https:\/\//.test(item))
    : [];

  return {
    id,
    title,
    description: stringOr(raw.description),
    category: stringOr(raw.category, "uncategorized").toLowerCase(),
    price: Math.max(0, numberOr(raw.price)),
    discountPercentage: Math.min(100, Math.max(0, numberOr(raw.discountPercentage))),
    rating: Math.max(0, numberOr(raw.rating)),
    stock: Math.max(0, numberOr(raw.stock)),
    brand: stringOr(raw.brand, "GLOBAL"),
    thumbnail,
    images,
    source,
    sourceId: String(id),
    sourceUrl: stringOr(raw.url),
    importedAt
  };
}

export function normalizeCatalog(records: unknown[], source: string, importedAt = new Date().toISOString()): CatalogProduct[] {
  const seen = new Set<string>();
  const products: CatalogProduct[] = [];
  for (const record of records) {
    if (!record || typeof record !== "object" || Array.isArray(record)) continue;
    const product = normalizeProduct(record as UnknownRecord, source, importedAt);
    if (!product) continue;
    const key = `${source}:${product.sourceId}`;
    if (seen.has(key)) continue;
    seen.add(key);
    products.push(product);
  }
  return products;
}
