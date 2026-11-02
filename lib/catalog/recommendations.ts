export type RecommendationSignals = {
  id: number;
  rating: number;
  discountPercentage: number;
  price: number;
};

export type CatalogRecommendationReason =
  | "rating-and-discount"
  | "rating"
  | "discount"
  | "balanced";

const bounded = (value: number, minimum: number, maximum: number): number =>
  Number.isFinite(value) ? Math.min(maximum, Math.max(minimum, value)) : minimum;

/**
 * Transparent demo score: rating contributes 70%, and discount contributes
 * 30% (discount is capped at 50%). It deliberately does not use personal data.
 */
export function getCatalogRecommendationScore(product: RecommendationSignals): number {
  const ratingSignal = bounded(product.rating, 0, 5) / 5;
  const discountSignal = bounded(product.discountPercentage, 0, 50) / 50;
  return ratingSignal * 0.7 + discountSignal * 0.3;
}

export function getCatalogRecommendationReason(
  product: RecommendationSignals
): CatalogRecommendationReason {
  const rating = bounded(product.rating, 0, 5);
  const discount = bounded(product.discountPercentage, 0, 100);
  if (rating >= 4.7 && discount >= 15) return "rating-and-discount";
  if (rating >= 4.5) return "rating";
  if (discount >= 20) return "discount";
  return "balanced";
}

/** Return a new deterministically sorted array without mutating the source catalog. */
export function rankCatalogRecommendations<T extends RecommendationSignals>(
  products: readonly T[]
): T[] {
  return products
    .map((product, index) => ({ product, index, score: getCatalogRecommendationScore(product) }))
    .sort((a, b) =>
      b.score - a.score ||
      bounded(b.product.rating, 0, 5) - bounded(a.product.rating, 0, 5) ||
      a.product.price - b.product.price ||
      a.product.id - b.product.id ||
      a.index - b.index
    )
    .map(entry => entry.product);
}
