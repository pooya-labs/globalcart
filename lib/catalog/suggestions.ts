export type SearchableCatalogProduct = {
  id: number;
  title: string;
  category: string;
  brand?: string | null;
};

export type CatalogSearchSuggestion = {
  key: string;
  kind: "product" | "brand" | "category";
  label: string;
  value: string;
  productId?: number;
  detail?: string;
};

const normalize = (value: string): string =>
  value.normalize("NFKC").trim().replace(/\s+/g, " ").toLocaleLowerCase();

/**
 * Build deterministic, bounded search suggestions from the catalog already
 * loaded by the storefront. No network requests or source data mutation occur.
 */
export function buildCatalogSuggestions<T extends SearchableCatalogProduct>(
  products: readonly T[],
  rawQuery: string,
  requestedLimit = 8
): CatalogSearchSuggestion[] {
  const query = normalize(rawQuery);
  const safeLimit = Number.isFinite(requestedLimit) ? Math.floor(requestedLimit) : 8;
  const limit = Math.max(0, Math.min(12, safeLimit));
  if (query.length < 2 || limit === 0) return [];

  const candidates: Array<CatalogSearchSuggestion & { score: number; kindOrder: number }> = [];
  const add = (suggestion: CatalogSearchSuggestion, score: number, kindOrder: number) => {
    candidates.push({ ...suggestion, score, kindOrder });
  };

  const brands = new Map<string, string>();
  const categories = new Map<string, string>();

  for (const product of products) {
    const title = normalize(product.title);
    const brand = (product.brand ?? "").trim();
    const normalizedBrand = normalize(brand);
    const category = product.category.trim();
    const normalizedCategory = normalize(category);

    if (title.includes(query)) {
      const score = title === query ? 0 : title.startsWith(query) ? 1 : 2;
      add({
        key: "product:" + product.id,
        kind: "product",
        label: product.title,
        value: product.title,
        productId: product.id,
        detail: [brand, category].filter(Boolean).join(" · ") || undefined
      }, score, 1);
    }
    if (brand && normalizedBrand.includes(query)) brands.set(normalizedBrand, brand);
    if (category && normalizedCategory.includes(query)) categories.set(normalizedCategory, category);
  }

  for (const [normalizedBrand, brand] of brands) {
    add({
      key: "brand:" + normalizedBrand,
      kind: "brand",
      label: brand,
      value: brand
    }, normalizedBrand.startsWith(query) ? 2 : 4, 0);
  }
  for (const [normalizedCategory, category] of categories) {
    add({
      key: "category:" + normalizedCategory,
      kind: "category",
      label: category,
      value: category
    }, normalizedCategory.startsWith(query) ? 3 : 5, 2);
  }

  candidates.sort((a, b) =>
    a.score - b.score ||
    a.kindOrder - b.kindOrder ||
    a.label.localeCompare(b.label)
  );

  const seen = new Set<string>();
  const suggestions: CatalogSearchSuggestion[] = [];
  for (const candidate of candidates) {
    if (seen.has(candidate.key)) continue;
    seen.add(candidate.key);
    const { score: _score, kindOrder: _kindOrder, ...suggestion } = candidate;
    suggestions.push(suggestion);
    if (suggestions.length >= limit) break;
  }
  return suggestions;
}
