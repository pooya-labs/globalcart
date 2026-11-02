export type FxRateStatus = "fresh" | "aging" | "stale" | "unavailable";

export type FxRateSnapshot = {
  schemaVersion: 1;
  provider: "Nobitex";
  market: "USDTIRT";
  baseCurrency: "USDT";
  quoteCurrency: "IRT";
  rawQuoteUnit: "IRR";
  tomanPerUsdt: number | null;
  bestBidToman: number | null;
  bestAskToman: number | null;
  lastTradeToman: number | null;
  marketUpdatedAt: string | null;
};

type UnknownRecord = Record<string, unknown>;

const RIALS_PER_TOMAN = 10;
const MAX_FUTURE_SKEW_MS = 5 * 60 * 1000;
const FRESH_WINDOW_MS = 30 * 60 * 1000;
const AGING_WINDOW_MS = 90 * 60 * 1000;
const MIN_PLAUSIBLE_TOMAN_PER_USDT = 10_000;
const MAX_PLAUSIBLE_TOMAN_PER_USDT = 10_000_000;

function asRecord(value: unknown): UnknownRecord {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new TypeError("Nobitex response must be a JSON object");
  }
  return value as UnknownRecord;
}

function positiveNumber(value: unknown, field: string): number {
  const parsed = typeof value === "number"
    ? value
    : typeof value === "string" && value.trim() !== ""
      ? Number(value)
      : Number.NaN;
  if (!Number.isFinite(parsed) || parsed <= 0) {
    throw new RangeError("Nobitex " + field + " must be a positive finite number");
  }
  return parsed;
}

function parseLevels(value: unknown, field: string): number[] {
  if (!Array.isArray(value)) {
    throw new TypeError("Nobitex " + field + " must be an array");
  }
  const prices = value
    .filter((level): level is unknown[] => Array.isArray(level) && level.length >= 2)
    .map(level => positiveNumber(level[0], field + " price"));
  if (prices.length === 0) {
    throw new RangeError("Nobitex " + field + " has no valid price levels");
  }
  return prices;
}

/**
 * Converts the public USDT/IRT order book into a Toman-per-USDT quote.
 * Nobitex order-book prices are in Rial units; 10 Rial = 1 Toman.
 * The quote is the midpoint between the best bid and best ask, not a USD
 * cash-market quote and not a promise of an executable checkout price.
 */
export function normalizeNobitexOrderBook(
  input: unknown,
  now = Date.now()
): FxRateSnapshot {
  const payload = asRecord(input);
  if (payload.status !== "ok") {
    throw new Error("Nobitex order-book status is not ok");
  }
  if (!Number.isFinite(now)) {
    throw new RangeError("now must be a finite timestamp");
  }

  const lastUpdateMs = positiveNumber(payload.lastUpdate, "lastUpdate");
  if (lastUpdateMs > now + MAX_FUTURE_SKEW_MS) {
    throw new RangeError("Nobitex market timestamp is implausibly in the future");
  }

  const bidPrices = parseLevels(payload.bids, "bids");
  const askPrices = parseLevels(payload.asks, "asks");
  const bestBidRial = Math.max(...bidPrices);
  const bestAskRial = Math.min(...askPrices);
  if (bestBidRial > bestAskRial) {
    throw new RangeError("Nobitex order book is crossed");
  }

  const midpointToman = (bestBidRial + bestAskRial) / (2 * RIALS_PER_TOMAN);
  if (
    !Number.isFinite(midpointToman) ||
    midpointToman < MIN_PLAUSIBLE_TOMAN_PER_USDT ||
    midpointToman > MAX_PLAUSIBLE_TOMAN_PER_USDT
  ) {
    throw new RangeError("Nobitex USDT/IRT midpoint is outside the supported Toman range");
  }

  let lastTradeToman: number | null = null;
  if (payload.lastTradePrice !== undefined && payload.lastTradePrice !== null && payload.lastTradePrice !== "") {
    const lastTradeRial = positiveNumber(payload.lastTradePrice, "lastTradePrice");
    lastTradeToman = lastTradeRial / RIALS_PER_TOMAN;
  }

  return {
    schemaVersion: 1,
    provider: "Nobitex",
    market: "USDTIRT",
    baseCurrency: "USDT",
    quoteCurrency: "IRT",
    rawQuoteUnit: "IRR",
    tomanPerUsdt: midpointToman,
    bestBidToman: bestBidRial / RIALS_PER_TOMAN,
    bestAskToman: bestAskRial / RIALS_PER_TOMAN,
    lastTradeToman,
    marketUpdatedAt: new Date(lastUpdateMs).toISOString()
  };
}

export function getFxRateStatus(
  snapshot: Pick<FxRateSnapshot, "tomanPerUsdt" | "marketUpdatedAt"> | null | undefined,
  now = Date.now()
): FxRateStatus {
  if (
    !snapshot ||
    typeof snapshot.tomanPerUsdt !== "number" ||
    !Number.isFinite(snapshot.tomanPerUsdt) ||
    snapshot.tomanPerUsdt <= 0 ||
    typeof snapshot.marketUpdatedAt !== "string" ||
    !Number.isFinite(now)
  ) {
    return "unavailable";
  }

  const timestamp = Date.parse(snapshot.marketUpdatedAt);
  if (!Number.isFinite(timestamp) || timestamp > now + MAX_FUTURE_SKEW_MS) {
    return "unavailable";
  }
  const age = Math.max(0, now - timestamp);
  if (age <= FRESH_WINDOW_MS) return "fresh";
  if (age <= AGING_WINDOW_MS) return "aging";
  return "stale";
}
