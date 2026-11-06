import assert from "node:assert/strict";
import test from "node:test";
import catalog from "../data/catalog.json";
import { EDITORIAL_PRODUCT_GROUPS, getPersianProductCopy, selectEditorialProducts } from "../lib/catalog/editorial";

test("editorial groups contain twelve unique products from the committed catalog", () => {
  const ids = Object.values(EDITORIAL_PRODUCT_GROUPS).flat();
  assert.equal(ids.length, 12);
  assert.equal(new Set(ids).size, 12);
  const selected = selectEditorialProducts(catalog.products, ids);
  assert.equal(selected.length, 12);
  assert.deepEqual(selected.map(product => product.id), ids);
});

test("every editorial product has a Persian title and concise description", () => {
  for (const id of Object.values(EDITORIAL_PRODUCT_GROUPS).flat()) {
    const copy = getPersianProductCopy(id);
    assert.ok(copy?.title, `missing Persian title for ${id}`);
    assert.ok(copy?.description, `missing Persian description for ${id}`);
    assert.ok(copy.title.length <= 48, `title too long for ${id}`);
    assert.ok(copy.description.length <= 130, `description too long for ${id}`);
  }
});
