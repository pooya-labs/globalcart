import test from "node:test";
import assert from "node:assert/strict";
import { parseLocalePreference } from "../lib/locale";

test("accepts supported Persian and English locale preferences", () => {
  assert.equal(parseLocalePreference("fa"), "fa");
  assert.equal(parseLocalePreference("en"), "en");
});

test("ignores missing or unsupported locale preferences", () => {
  assert.equal(parseLocalePreference(null), null);
  assert.equal(parseLocalePreference("de"), null);
  assert.equal(parseLocalePreference(""), null);
});
