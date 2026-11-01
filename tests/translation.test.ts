import test from "node:test";
import assert from "node:assert/strict";
import { localizeText } from "../lib/translation";

test("does not pretend untranslated content is localized", async () => {
  const result = await localizeText("Wireless headphones", null);
  assert.equal(result.translated, null);
  assert.equal(result.status, "unavailable");
  assert.equal(result.original, "Wireless headphones");
});

test("stores provider and timestamp for successful translations", async () => {
  const provider = { name: "test-provider", translate: async (text: string) => `ترجمه: ${text}` };
  const result = await localizeText("Wireless headphones", provider);
  assert.equal(result.translated, "ترجمه: Wireless headphones");
  assert.equal(result.status, "translated");
  assert.equal(result.provider, "test-provider");
  assert.ok(result.translatedAt);
});
