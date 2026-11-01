import test from "node:test";
import assert from "node:assert/strict";
import { calculateDemoPrice } from "../lib/pricing";

test("applies discount before converting to Toman", () => {
  assert.deepEqual(calculateDemoPrice({ sourcePrice: 100, fxRate: 70000, discountPercentage: 10 }), {
    sourcePrice: 100, discountedUsd: 90, fxRate: 70000, toman: 6300000
  });
});

test("uses cent precision and rounds the converted Toman amount", () => {
  assert.deepEqual(calculateDemoPrice({ sourcePrice: 0.1, fxRate: 70000.5, discountPercentage: 12.5 }), {
    sourcePrice: 0.1, discountedUsd: 0.09, fxRate: 70000.5, toman: 6300
  });
});

test("clamps discounts and handles zero and full discounts", () => {
  assert.equal(calculateDemoPrice({ sourcePrice: 10, fxRate: 1000, discountPercentage: 150 }).toman, 0);
  assert.equal(calculateDemoPrice({ sourcePrice: 10, fxRate: 1000, discountPercentage: -10 }).toman, 10000);
  assert.equal(calculateDemoPrice({ sourcePrice: 10, fxRate: 1000, discountPercentage: 100 }).discountedUsd, 0);
});

test("rejects invalid, non-finite and out-of-range inputs", () => {
  assert.throws(() => calculateDemoPrice({ sourcePrice: -1, fxRate: 1000 }), RangeError);
  assert.throws(() => calculateDemoPrice({ sourcePrice: 1, fxRate: 0 }), RangeError);
  assert.throws(() => calculateDemoPrice({ sourcePrice: 1, fxRate: 1000, discountPercentage: Number.NaN }), RangeError);
  assert.throws(() => calculateDemoPrice({ sourcePrice: Number.MAX_VALUE, fxRate: 1000 }), RangeError);
});
