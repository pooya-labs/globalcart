export type FxRateStatus = "fresh" | "aging" | "stale" | "unavailable";
export type FxRateSnapshot = {
  schemaVersion: 2; provider: "Wallex"; market: "USDTTMN"; baseCurrency: "USDT";
  quoteCurrency: "TMN"; rawQuoteUnit: "TMN";
  priceSource: "ONE_MINUTE_CANDLE_CLOSE"; candleResolutionMinutes: 1;
  tomanPerUsdt: number | null; candleStartedAt: string | null;
};
type UnknownRecord = Record<string, unknown>;
const MAX_FUTURE_SKEW_SECONDS = 60;
const FRESH_WINDOW_MS = 5 * 60 * 1000;
const AGING_WINDOW_MS = 15 * 60 * 1000;
const MIN_PLAUSIBLE_TOMAN_PER_USDT = 10_000;
const MAX_PLAUSIBLE_TOMAN_PER_USDT = 10_000_000;
function asRecord(value: unknown): UnknownRecord {
  if (!value || typeof value !== "object" || Array.isArray(value)) throw new TypeError("Wallex UDF response must be a JSON object");
  return value as UnknownRecord;
}
function positiveNumber(value: unknown, field: string): number {
  const parsed = typeof value === "number" ? value : typeof value === "string" && value.trim() !== "" ? Number(value) : Number.NaN;
  if (!Number.isFinite(parsed) || parsed <= 0) throw new RangeError("Wallex " + field + " must be a positive finite number");
  return parsed;
}
/** Wallex UDF timestamps are Unix seconds marking candle start; c[] is the close in Toman. This is indicative, not an order-book midpoint, executable quote, or cash-USD rate. */
export function normalizeWallexMinuteHistory(input: unknown, now = Date.now()): FxRateSnapshot {
  const payload = asRecord(input);
  if (payload.s !== "ok") throw new Error("Wallex UDF history status is not ok");
  if (!Number.isFinite(now)) throw new RangeError("now must be a finite timestamp");
  if (!Array.isArray(payload.t) || !Array.isArray(payload.c)) throw new TypeError("Wallex UDF history must include t and c arrays");
  if (payload.t.length === 0 || payload.t.length !== payload.c.length) throw new RangeError("Wallex UDF history has no complete candle series");
  const nowSeconds = Math.floor(now / 1000);
  let latestIndex = -1;
  let latestTimestamp = -1;
  for (let index = 0; index < payload.t.length; index += 1) {
    const timestamp = positiveNumber(payload.t[index], "candle timestamp");
    if (!Number.isSafeInteger(timestamp)) throw new RangeError("Wallex candle timestamp must be Unix seconds");
    if (timestamp > nowSeconds + MAX_FUTURE_SKEW_SECONDS) throw new RangeError("Wallex candle timestamp is implausibly in the future");
    if (timestamp > latestTimestamp) { latestTimestamp = timestamp; latestIndex = index; }
  }
  if (latestIndex < 0) throw new RangeError("Wallex UDF history contains no valid candle");
  const closeToman = positiveNumber(payload.c[latestIndex], "candle close");
  if (closeToman < MIN_PLAUSIBLE_TOMAN_PER_USDT || closeToman > MAX_PLAUSIBLE_TOMAN_PER_USDT) {
    throw new RangeError("Wallex USDTTMN close is outside the supported Toman range");
  }
  return {
    schemaVersion: 2, provider: "Wallex", market: "USDTTMN", baseCurrency: "USDT",
    quoteCurrency: "TMN", rawQuoteUnit: "TMN", priceSource: "ONE_MINUTE_CANDLE_CLOSE",
    candleResolutionMinutes: 1, tomanPerUsdt: closeToman,
    candleStartedAt: new Date(latestTimestamp * 1000).toISOString()
  };
}
export function getFxRateStatus(snapshot: Pick<FxRateSnapshot, "tomanPerUsdt" | "candleStartedAt"> | null | undefined, now = Date.now()): FxRateStatus {
  if (!snapshot || typeof snapshot.tomanPerUsdt !== "number" || !Number.isFinite(snapshot.tomanPerUsdt) ||
      snapshot.tomanPerUsdt <= 0 || typeof snapshot.candleStartedAt !== "string" || !Number.isFinite(now)) return "unavailable";
  const timestamp = Date.parse(snapshot.candleStartedAt);
  if (!Number.isFinite(timestamp) || timestamp > now + 60_000) return "unavailable";
  const age = Math.max(0, now - timestamp);
  if (age <= FRESH_WINDOW_MS) return "fresh";
  if (age <= AGING_WINDOW_MS) return "aging";
  return "stale";
}
