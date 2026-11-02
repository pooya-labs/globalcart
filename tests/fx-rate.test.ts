import test from "node:test";
import assert from "node:assert/strict";
import { getFxRateStatus, normalizeNobitexOrderBook } from "../lib/fx/nobitex";

const NOW = Date.parse("2026-11-02T12:00:00.000Z");
const validOrderBook = {
  status: "ok",
  lastUpdate: NOW - 2_000,
  lastTradePrice: "2723500",
  asks: [["2725000", "120"], ["2726000", "20"]],
  bids: [["2720000", "95"], ["2719000", "30"]]
};

test("normalizes the best USDT/IRT bid-ask midpoint from Rial into Toman", () => {
  const quote = normalizeNobitexOrderBook(validOrderBook, NOW);
  assert.equal(quote.provider, "Nobitex");
  assert.equal(quote.market, "USDTIRT");
  assert.equal(quote.bestBidToman, 272000);
  assert.equal(quote.bestAskToman, 272500);
  assert.equal(quote.tomanPerUsdt, 272250);
  assert.equal(quote.lastTradeToman, 272350);
  assert.equal(quote.marketUpdatedAt, new Date(NOW - 2_000).toISOString());
});

test("does not rely on order-book array ordering when selecting best levels", () => {
  const quote = normalizeNobitexOrderBook({
    ...validOrderBook,
    asks: [["2727000", "2"], ["2725000", "1"]],
    bids: [["2719000", "1"], ["2721000", "2"]]
  }, NOW);
  assert.equal(quote.bestBidToman, 272100);
  assert.equal(quote.bestAskToman, 272500);
  assert.equal(quote.tomanPerUsdt, 272300);
});

test("rejects bad status, missing sides, crossed books and implausible prices", () => {
  assert.throws(() => normalizeNobitexOrderBook({ ...validOrderBook, status: "error" }, NOW), /status/);
  assert.throws(() => normalizeNobitexOrderBook({ ...validOrderBook, bids: [] }, NOW), /bids/);
  assert.throws(() => normalizeNobitexOrderBook({
    ...validOrderBook,
    bids: [["2726000", "1"]],
    asks: [["2725000", "1"]]
  }, NOW), /crossed/);
  assert.throws(() => normalizeNobitexOrderBook({
    ...validOrderBook,
    bids: [["99900", "1"]],
    asks: [["99910", "1"]]
  }, NOW), /supported Toman range/);
  assert.throws(() => normalizeNobitexOrderBook({ ...validOrderBook, lastUpdate: "bad" }, NOW), /lastUpdate/);
});

test("rejects timestamps too far in the future", () => {
  assert.throws(() => normalizeNobitexOrderBook({
    ...validOrderBook,
    lastUpdate: NOW + 6 * 60 * 1000
  }, NOW), /future/);
});

test("classifies exchange quotes using the exchange timestamp", () => {
  const quote = normalizeNobitexOrderBook(validOrderBook, NOW);
  assert.equal(getFxRateStatus(quote, NOW), "fresh");
  assert.equal(getFxRateStatus({
    ...quote,
    marketUpdatedAt: new Date(NOW - 30 * 60 * 1000 - 1).toISOString()
  }, NOW), "aging");
  assert.equal(getFxRateStatus({
    ...quote,
    marketUpdatedAt: new Date(NOW - 90 * 60 * 1000 - 1).toISOString()
  }, NOW), "stale");
});

test("reports missing, malformed and future quotes as unavailable", () => {
  assert.equal(getFxRateStatus(null, NOW), "unavailable");
  assert.equal(getFxRateStatus({ tomanPerUsdt: null, marketUpdatedAt: null }, NOW), "unavailable");
  assert.equal(getFxRateStatus({
    tomanPerUsdt: Number.NaN,
    marketUpdatedAt: new Date(NOW).toISOString()
  }, NOW), "unavailable");
  assert.equal(getFxRateStatus({
    tomanPerUsdt: 250_000,
    marketUpdatedAt: "not-a-date"
  }, NOW), "unavailable");
  assert.equal(getFxRateStatus({
    tomanPerUsdt: 250_000,
    marketUpdatedAt: new Date(NOW + 6 * 60 * 1000).toISOString()
  }, NOW), "unavailable");
});
