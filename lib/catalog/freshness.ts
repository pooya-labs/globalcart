/**
 * Freshness describes when GlobalCart last fetched or generated the catalog
 * currently shown. It is not a supplier-provided last-modified timestamp.
 *
 * The scheduled public snapshot runs daily. 36 hours allows one delayed cycle;
 * 72 hours means two daily cycles have been missed.
 */
export type CatalogFreshnessStatus = "fresh" | "aging" | "stale" | "unknown";

const FRESH_WINDOW_MS = 36 * 60 * 60 * 1000;
const STALE_WINDOW_MS = 72 * 60 * 60 * 1000;
const FUTURE_CLOCK_TOLERANCE_MS = 5 * 60 * 1000;

export function getCatalogFreshness(
  timestamp: string | null | undefined,
  now = Date.now()
): CatalogFreshnessStatus {
  if (typeof timestamp !== "string" || timestamp.trim().length === 0 || !Number.isFinite(now)) {
    return "unknown";
  }

  const parsed = Date.parse(timestamp);
  if (!Number.isFinite(parsed) || parsed > now + FUTURE_CLOCK_TOLERANCE_MS) {
    return "unknown";
  }

  const age = Math.max(0, now - parsed);
  if (age <= FRESH_WINDOW_MS) return "fresh";
  if (age <= STALE_WINDOW_MS) return "aging";
  return "stale";
}
