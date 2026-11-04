export type CatalogFacetValues = {
  query?: string;
  category?: string;
  brand?: string;
  minPrice?: number;
  maxPrice?: number;
  minRating?: number;
};

export type CatalogFacetProduct = {
  title: string;
  description: string;
  category: string;
  brand?: string | null;
  price: number;
  rating: number;
  priceAvailable?: boolean;
  ratingAvailable?: boolean;
};

/**
 * Apply the storefront's local discovery filters without mutating the source
 * catalog. Brand matching is normalized; numeric bounds are inclusive.
 */
export function filterCatalogProducts<T extends CatalogFacetProduct>(
  products: readonly T[],
  filters: CatalogFacetValues
): T[] {
  const query = (filters.query ?? "").trim().toLocaleLowerCase();
  const category = (filters.category ?? "all").trim();
  const brand = (filters.brand ?? "").trim().toLocaleLowerCase();
  const minPrice = filters.minPrice ?? Number.NEGATIVE_INFINITY;
  const maxPrice = filters.maxPrice ?? Number.POSITIVE_INFINITY;
  const minRating = filters.minRating ?? 0;

  return products.filter(product => {
    const searchableText = [
      product.title,
      product.description,
      product.brand ?? "",
      product.category
    ].join(" ").toLocaleLowerCase();
    const productBrand = (product.brand ?? "").trim().toLocaleLowerCase();

    return (!category || category === "all" || product.category === category) &&
      (!query || searchableText.includes(query)) &&
      (!brand || productBrand === brand) &&
      (product.priceAvailable === false || (product.price >= minPrice && product.price <= maxPrice)) &&
      (product.ratingAvailable === false ? minRating <= 0 : product.rating >= minRating);
  });
}
