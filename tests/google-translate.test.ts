import test from "node:test";
import assert from "node:assert/strict";
import { parseGoogleTranslateResponse } from "../lib/translation/google-apps-script";

test("accepts a complete Google Translate bridge response", () => {
  const fields = parseGoogleTranslateResponse({
    ok: true,
    fields: { title: "هدفون", description: "توضیحات", category: "صوتی", brand: "Acme" }
  });
  assert.equal(fields.title, "هدفون");
  assert.equal(fields.brand, "Acme");
});

test("rejects failed and incomplete bridge responses", () => {
  assert.throws(() => parseGoogleTranslateResponse({ ok: false }));
  assert.throws(() => parseGoogleTranslateResponse({ ok: true, fields: { title: "عنوان" } }));
});
