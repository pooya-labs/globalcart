import type { CatalogProduct } from "./types";

export type CatalogQualityReport = {
  productCount: number;
  variantCount: number;
  productsWithVariants: number;
  productsWithoutVariants: number;
  zeroPriceProducts: number;
  zeroStockProducts: number;
  missingDescriptionProducts: number;
  uncategorizedProducts: number;
};

/**
 * Rejects a suspiciously incomplete full-catalog snapshot before it can
 * replace the last-known-good file. A 50% maximum drop is intentionally
 * conservative; intentional large catalog reductions require an explicit
 * source/threshold review rather than silently publishing a partial feed.
 */
export function assertCatalogSnapshotContinuity(
  currentCount: number,
  previousCount: number,
  maxDropFraction = 0.5
): void {
  if (!Number.isInteger(currentCount) || currentCount < 0) {
    throw new RangeError("Current catalog count must be a non-negative integer");
  }
  if (!Number.isInteger(previousCount) || previousCount < 0) {
    throw new RangeError("Previous catalog count must be a non-negative integer");
  }
  if (!Number.isFinite(maxDropFraction) || maxDropFraction < 0 || maxDropFraction >= 1) {
    throw new RangeError("Maximum catalog drop fraction must be in the range [0, 1)");
  }
  if (previousCount === 0) return;

  const minimumAccepted = Math.ceil(previousCount * (1 - maxDropFraction));
  if (currentCount < minimumAccepted) {
    throw new Error(
      `Catalog snapshot rejected: product count dropped from ${previousCount} to ${currentCount}; minimum accepted is ${minimumAccepted}`
    );
  }
}

export function summarizeCatalogQuality(products: readonly CatalogProduct[]): CatalogQualityReport {
  let variantCount = 0;
  let productsWithVariants = 0;
  let zeroPriceProducts = 0;
  let zeroStockProducts = 0;
  let missingDescriptionProducts = 0;
  let uncategorizedProducts = 0;

  for (const product of products) {
    const variants = product.variants ?? [];
    variantCount += variants.length;
    if (variants.length > 0) productsWithVariants += 1;
    if (!Number.isFinite(product.price) || product.price <= 0) zeroPriceProducts += 1;
    if (!Number.isFinite(product.stock) || product.stock <= 0) zeroStockProducts += 1;
    if (!product.description.trim()) missingDescriptionProducts += 1;
    if (!product.category.trim() || product.category === "uncategorized") uncategorizedProducts += 1;
  }

  return {
    productCount: products.length,
    variantCount,
    productsWithVariants,
    productsWithoutVariants: products.length - productsWithVariants,
    zeroPriceProducts,
    zeroStockProducts,
    missingDescriptionProducts,
    uncategorizedProducts
  };
}
