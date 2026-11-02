import test from "node:test";
import assert from "node:assert/strict";
import { getCatalogSourceStatus } from "../lib/catalog/source-status";

test("uses the generated snapshot immediately while a background refresh is pending", () => {
  assert.equal(getCatalogSourceStatus(true, "pending"), "snapshot");
});

test("shows loading only when no snapshot or fixture can provide the initial catalog", () => {
  assert.equal(getCatalogSourceStatus(false, "pending"), "loading");
});

test("switches to live status only when the source returns products", () => {
  assert.equal(getCatalogSourceStatus(true, "success", 10), "live");
  assert.equal(getCatalogSourceStatus(false, "success", 10), "live");
});

test("preserves the last generated snapshot after errors or empty source responses", () => {
  assert.equal(getCatalogSourceStatus(true, "error"), "snapshot-fallback");
  assert.equal(getCatalogSourceStatus(true, "success", 0), "snapshot-fallback");
});

test("reports local demo fixtures when no generated snapshot exists", () => {
  assert.equal(getCatalogSourceStatus(false, "error"), "fixture-fallback");
  assert.equal(getCatalogSourceStatus(false, "success", 0), "fixture-fallback");
});
