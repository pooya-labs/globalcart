import test from "node:test";
import assert from "node:assert/strict";
import { parseProductTranslation } from "../lib/translation/gemini";

test("parses valid JSON product translations", () => {
  const result = parseProductTranslation(JSON.stringify({
    title: "هدفون بی‌سیم",
    description: "توضیحات محصول",
    category: "صوتی",
    brand: "Acme"
  }));
  assert.equal(result.title, "هدفون بی‌سیم");
  assert.equal(result.brand, "Acme");
});

test("accepts a fenced JSON response but rejects missing fields", () => {
  assert.equal(parseProductTranslation('```json\n{"title":"الف","description":"ب","category":"ج","brand":"د"}\n```').title, "الف");
  assert.throws(() => parseProductTranslation('{"title":"الف"}'));
});

test("rejects non-JSON model output", () => {
  assert.throws(() => parseProductTranslation("Here is your translation"));
});
