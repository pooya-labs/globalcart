export type SavedCatalogSort = "popular" | "low" | "high" | "rating";

export type SavedCatalogSearch = {
  id: string;
  name: string;
  query: string;
  category: string;
  brand: string;
  minPrice: number;
  maxPrice: number;
  minRating: number;
  sort: SavedCatalogSort;
  createdAt: string;
};

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

const validNumber = (value: unknown): value is number =>
  typeof value === "number" && Number.isFinite(value);

export function normalizeSavedCatalogSearches(value: unknown, maximum = 12): SavedCatalogSearch[] {
  if (!Array.isArray(value) || !Number.isFinite(maximum) || maximum <= 0) return [];
  const normalized: SavedCatalogSearch[] = [];
  for (const item of value) {
    if (!isRecord(item)) continue;
    const sort = item.sort;
    if (
      typeof item.id !== "string" || !item.id ||
      typeof item.name !== "string" || !item.name.trim() ||
      typeof item.query !== "string" ||
      typeof item.category !== "string" ||
      typeof item.brand !== "string" ||
      !validNumber(item.minPrice) || item.minPrice < 0 ||
      !validNumber(item.maxPrice) || item.maxPrice < item.minPrice ||
      !validNumber(item.minRating) || item.minRating < 0 || item.minRating > 5 ||
      (sort !== "popular" && sort !== "low" && sort !== "high" && sort !== "rating") ||
      typeof item.createdAt !== "string" || !Number.isFinite(Date.parse(item.createdAt))
    ) continue;
    normalized.push({
      id: item.id.slice(0, 100),
      name: item.name.trim().slice(0, 60),
      query: item.query.trim().slice(0, 200),
      category: item.category.trim().slice(0, 120) || "all",
      brand: item.brand.trim().slice(0, 120),
      minPrice: item.minPrice,
      maxPrice: item.maxPrice,
      minRating: item.minRating,
      sort,
      createdAt: item.createdAt
    });
    if (normalized.length >= Math.min(12, Math.floor(maximum))) break;
  }
  return normalized;
}

export function hasSameSavedSearchFilters(
  first: Omit<SavedCatalogSearch, "id" | "name" | "createdAt">,
  second: Omit<SavedCatalogSearch, "id" | "name" | "createdAt">
): boolean {
  return first.query.trim().toLocaleLowerCase() === second.query.trim().toLocaleLowerCase() &&
    first.category === second.category &&
    first.brand.trim().toLocaleLowerCase() === second.brand.trim().toLocaleLowerCase() &&
    first.minPrice === second.minPrice &&
    first.maxPrice === second.maxPrice &&
    first.minRating === second.minRating &&
    first.sort === second.sort;
}
