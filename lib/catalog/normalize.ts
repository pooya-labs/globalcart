import type { CatalogProduct, CatalogVariant } from "./types";

type UnknownRecord = Record<string, unknown>;

const numberOr = (value: unknown, fallback = 0): number => {
  const parsed = typeof value === "number" ? value : Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const stringOr = (value: unknown, fallback = ""): string =>
  typeof value === "string" ? value.trim() : fallback;

function normalizeVariants(raw: unknown, source: string): CatalogVariant[] {
  if (!Array.isArray(raw)) return [];
  const seen = new Set<string>();
  const variants: CatalogVariant[] = [];
  for (const item of raw) {
    if (!item || typeof item !== "object" || Array.isArray(item)) continue;
    const record = item as UnknownRecord;
    const sku = stringOr(record.sku);
    const explicitID = stringOr(record.sourceId) || stringOr(record.id);
    const sourceId = explicitID || sku;
    const variantSource = stringOr(record.source) || source;
    const identity = JSON.stringify([variantSource, sourceId]);
    if (!sourceId || seen.has(identity)) continue;
    const attributes = record.attributes && typeof record.attributes === "object" && !Array.isArray(record.attributes)
      ? record.attributes as Record<string, unknown>
      : {};
    variants.push({
      sku,
      title: stringOr(record.title),
      price: Math.max(0, numberOr(record.price)),
      stock: Math.max(0, Math.floor(numberOr(record.stock))),
      attributes,
      source: variantSource,
      sourceId,
      isActive: typeof record.isActive === "boolean" ? record.isActive : true
    });
    seen.add(identity);
  }
  return variants;
}

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
    stock: Math.max(0, Math.floor(numberOr(raw.stock))),
    brand: stringOr(raw.brand, "GLOBAL"),
    thumbnail,
    images,
    source,
    sourceId: String(id),
    sourceUrl: stringOr(raw.url),
    importedAt,
    variants: normalizeVariants(raw.variants, source)
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
