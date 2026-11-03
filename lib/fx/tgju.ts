export type FxRateStatus = "fresh" | "aging" | "stale" | "unavailable";
export type FxRateSnapshot = {
  schemaVersion: 3;
  provider: "TGJU";
  instrument: "price_dollar_rl";
  baseCurrency: "USD";
  rawQuoteCurrency: "IRR";
  rawPriceRial: number | null;
  displayQuoteCurrency: "TOMAN";
  tomanPerUsd: number | null;
  sourceTimeTehran: string | null;
  sourceUpdatedAt: string | null;
};

type MarketRow = { attributes: string; html: string };
const MIN_PLAUSIBLE_RIAL_PER_USD = 100_000;
const MAX_PLAUSIBLE_RIAL_PER_USD = 100_000_000;
const MAX_FUTURE_SKEW_MS = 60_000;
const FRESH_WINDOW_MS = 5 * 60 * 1000;
const AGING_WINDOW_MS = 15 * 60 * 1000;
const TEHRAN_UTC_OFFSET_MINUTES = 210;

function normalizeDigits(value: string): string {
  return value.replace(/[۰-۹]/g, d => String("۰۱۲۳۴۵۶۷۸۹".indexOf(d)))
    .replace(/[٠-٩]/g, d => String("٠١٢٣٤٥٦٧٨٩".indexOf(d)))
    .replace(/[\u200B-\u200F\uFEFF]/g, "");
}
function readAttribute(tag: string, name: string): string | null {
  const match = tag.match(new RegExp("(?:^|\\s)" + name + "\\s*=\\s*(?:\"([^\"]*)\"|'([^']*)'|([^\\s>]+))", "i"));
  return match ? (match[1] ?? match[2] ?? match[3] ?? null) : null;
}
function htmlText(value: string): string {
  return normalizeDigits(value.replace(/<script\b[^>]*>[\s\S]*?<\/script\s*>/gi, " ")
    .replace(/<style\b[^>]*>[\s\S]*?<\/style\s*>/gi, " ").replace(/<[^>]+>/g, " ")
    .replace(/&nbsp;|&#160;|&#xA0;/gi, " ").replace(/&amp;/gi, "&").replace(/&lt;/gi, "<")
    .replace(/&gt;/gi, ">").replace(/&quot;/gi, '"').replace(/&#39;|&apos;/gi, "'")
    .replace(/\s+/g, " ").trim());
}
function findDollarRow(html: string): MarketRow {
  for (const match of html.matchAll(/<tr\b[^>]*>/gi)) {
    const attributes = match[0];
    const slug = readAttribute(attributes, "data-market-nameslug") ?? readAttribute(attributes, "data-market-row");
    if (slug !== "price_dollar_rl") continue;
    const start = (match.index ?? 0) + attributes.length;
    const close = html.toLowerCase().indexOf("</tr", start);
    if (close < 0) throw new Error("TGJU dollar row is missing its closing table-row tag");
    return { attributes, html: html.slice(start, close) };
  }
  throw new Error("TGJU dollar market row (price_dollar_rl) was not found; the page may be blocked or its markup changed");
}
function parsePositiveInteger(value: string): number {
  const normalized = normalizeDigits(value).replace(/[\s,٬،]/g, "");
  if (!/^\d+$/.test(normalized)) throw new TypeError("TGJU dollar price is not a whole-number Rial quote");
  const parsed = Number(normalized);
  if (!Number.isSafeInteger(parsed) || parsed <= 0) throw new RangeError("TGJU dollar price must be a positive safe integer");
  return parsed;
}
function parseTehranTime(value: string, now: number): { time: string; timestamp: string } {
  const normalized = normalizeDigits(value).trim();
  const match = normalized.match(/^([01]\d|2[0-3]):([0-5]\d):([0-5]\d)$/);
  if (!match) throw new TypeError("TGJU dollar row does not contain a live HH:mm:ss update time");
  const parts = new Intl.DateTimeFormat("en-US", {
    timeZone: "Asia/Tehran", year: "numeric", month: "2-digit", day: "2-digit"
  }).formatToParts(new Date(now));
  const part = (type: Intl.DateTimeFormatPartTypes) => parts.find(item => item.type === type)?.value ?? "";
  const year = Number(part("year")), month = Number(part("month")), day = Number(part("day"));
  let timestamp = Date.UTC(year, month - 1, day, Number(match[1]), Number(match[2]), Number(match[3]))
    - TEHRAN_UTC_OFFSET_MINUTES * 60_000;
  if (timestamp > now + MAX_FUTURE_SKEW_MS) {
    if (timestamp - now > 12 * 60 * 60 * 1000) timestamp -= 24 * 60 * 60 * 1000;
    else { const localClock = new Intl.DateTimeFormat("en-GB", { timeZone: "Asia/Tehran", hour: "2-digit", minute: "2-digit", second: "2-digit", hourCycle: "h23" }).format(new Date(now)); throw new RangeError("TGJU source time " + normalized + " is ahead of Tehran clock " + localClock + " (UTC now " + new Date(now).toISOString() + ")"); }
  }
  if (timestamp > now + MAX_FUTURE_SKEW_MS) throw new RangeError("TGJU source time is implausibly in the future");
  return { time: normalized, timestamp: new Date(timestamp).toISOString() };
}
/** TGJU publishes price_dollar_rl in Rial. This adapter converts it to Toman and uses the source's Tehran clock. */
export function normalizeTgjuDollarPage(input: unknown, now = Date.now()): FxRateSnapshot {
  if (typeof input !== "string" || input.length < 100) throw new TypeError("TGJU response is not a complete HTML page");
  if (!Number.isFinite(now)) throw new RangeError("now must be a finite timestamp");
  const row = findDollarRow(input);
  const rawPriceText = readAttribute(row.attributes, "data-price");
  const cells = [...row.html.matchAll(/<td\b[^>]*>([\s\S]*?)<\/td\s*>/gi)].map(m => htmlText(m[1] ?? ""));
  const rawPriceRial = parsePositiveInteger(rawPriceText ?? cells[0] ?? "");
  if (rawPriceRial < MIN_PLAUSIBLE_RIAL_PER_USD || rawPriceRial > MAX_PLAUSIBLE_RIAL_PER_USD) {
    throw new RangeError("TGJU dollar price is outside the supported Rial range");
  }
  if (cells.length < 5 || !cells[4]) throw new Error("TGJU dollar row is missing its update-time cell");
  const sourceTime = parseTehranTime(cells[4], now);
  return {
    schemaVersion: 3, provider: "TGJU", instrument: "price_dollar_rl", baseCurrency: "USD",
    rawQuoteCurrency: "IRR", rawPriceRial, displayQuoteCurrency: "TOMAN",
    tomanPerUsd: rawPriceRial / 10, sourceTimeTehran: sourceTime.time, sourceUpdatedAt: sourceTime.timestamp
  };
}
export function getFxRateStatus(
  snapshot: Pick<FxRateSnapshot, "tomanPerUsd" | "sourceUpdatedAt"> | null | undefined, now = Date.now()
): FxRateStatus {
  if (!snapshot || typeof snapshot.tomanPerUsd !== "number" || !Number.isFinite(snapshot.tomanPerUsd) ||
      snapshot.tomanPerUsd <= 0 || typeof snapshot.sourceUpdatedAt !== "string" || !Number.isFinite(now)) return "unavailable";
  const timestamp = Date.parse(snapshot.sourceUpdatedAt);
  if (!Number.isFinite(timestamp) || timestamp > now + MAX_FUTURE_SKEW_MS) return "unavailable";
  const age = Math.max(0, now - timestamp);
  if (age <= FRESH_WINDOW_MS) return "fresh";
  if (age <= AGING_WINDOW_MS) return "aging";
  return "stale";
}
