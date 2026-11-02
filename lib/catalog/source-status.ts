export type CatalogSourceStatus =
  | "loading"
  | "snapshot"
  | "live"
  | "snapshot-fallback"
  | "fixture-fallback";

export type CatalogFetchOutcome = "pending" | "success" | "error";

/**
 * Describe which catalog is visible without discarding a last-known-good snapshot.
 * A successful but empty response is treated as a failed refresh, not as an
 * instruction to replace valid local catalog data with an empty collection.
 */
export function getCatalogSourceStatus(
  hasSnapshot: boolean,
  outcome: CatalogFetchOutcome,
  productCount = 0
): CatalogSourceStatus {
  if (outcome === "pending") return hasSnapshot ? "snapshot" : "loading";
  if (outcome === "success" && Number.isFinite(productCount) && productCount > 0) return "live";
  return hasSnapshot ? "snapshot-fallback" : "fixture-fallback";
}
