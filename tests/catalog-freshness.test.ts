import test from "node:test";
import assert from "node:assert/strict";
import { getCatalogFreshness } from "../lib/catalog/freshness";

const HOUR = 60 * 60 * 1000;
const NOW = Date.parse("2026-11-02T12:00:00.000Z");

test("classifies a current or recently generated catalog as fresh", () => {
  assert.equal(getCatalogFreshness("2026-11-02T12:00:00.000Z", NOW), "fresh");
  assert.equal(getCatalogFreshness(new Date(NOW - 36 * HOUR).toISOString(), NOW), "fresh");
});

test("classifies a catalog older than one daily cycle as aging", () => {
  assert.equal(getCatalogFreshness(new Date(NOW - 36 * HOUR - 1).toISOString(), NOW), "aging");
  assert.equal(getCatalogFreshness(new Date(NOW - 72 * HOUR).toISOString(), NOW), "aging");
});

test("classifies a catalog older than two daily cycles as stale", () => {
  assert.equal(getCatalogFreshness(new Date(NOW - 72 * HOUR - 1).toISOString(), NOW), "stale");
});

test("reports unknown for missing, malformed or implausibly future timestamps", () => {
  assert.equal(getCatalogFreshness(null, NOW), "unknown");
  assert.equal(getCatalogFreshness("", NOW), "unknown");
  assert.equal(getCatalogFreshness("not-a-date", NOW), "unknown");
  assert.equal(getCatalogFreshness(new Date(NOW + 6 * 60 * 1000).toISOString(), NOW), "unknown");
});

test("tolerates minor clock skew without reporting a future timestamp as unknown", () => {
  assert.equal(getCatalogFreshness(new Date(NOW + 2 * 60 * 1000).toISOString(), NOW), "fresh");
});
