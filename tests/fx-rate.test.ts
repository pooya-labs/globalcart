import test from "node:test";
import assert from "node:assert/strict";
import { getFxRateStatus, normalizeTgjuDollarPage } from "../lib/fx/tgju";

const NOW = Date.parse("2026-11-03T09:19:45.000Z"); // 12:49:45 in Tehran.
const fixture = (price = "1,864,000", time = "۱۲:۴۹:۴۵", otherRows = "") => `<!doctype html>
<html lang="fa"><head><title>قیمت دلار</title></head><body>
<table class="data-table market-table"><thead><tr><th>عنوان</th><th>قیمت زنده</th><th>تغییر</th><th>کمترین</th><th>بیشترین</th><th>زمان</th><th></th></tr></thead>
<tbody>${otherRows}
<tr data-market-nameslug="price_dollar_rl" data-market-row="price_dollar_rl" data-price="${price}">
<th>دلار</th><td class="nf">${price}</td><td>۱٬۲۰۰</td><td>1,860,000</td><td>1,870,000</td><td>${time}</td><td><a href="profile/price_dollar_rl"></a></td>
</tr></tbody></table></body></html>`;

test("extracts the free-market USD row and converts Rial to Toman", () => {
  const q = normalizeTgjuDollarPage(fixture(), NOW);
  assert.equal(q.provider, "TGJU");
  assert.equal(q.instrument, "price_dollar_rl");
  assert.equal(q.baseCurrency, "USD");
  assert.equal(q.rawQuoteCurrency, "IRR");
  assert.equal(q.rawPriceRial, 1_864_000);
  assert.equal(q.displayQuoteCurrency, "TOMAN");
  assert.equal(q.tomanPerUsd, 186_400);
  assert.equal(q.sourceTimeTehran, "12:49:45");
  assert.equal(q.sourceUpdatedAt, new Date(NOW).toISOString());
});
test("selects only the dollar instrument when other currency rows exist", () => {
  const other = '<tr data-market-nameslug="price_eur" data-price="2,100,000"><th>یورو</th><td>2,100,000</td><td>0</td><td>1</td><td>2</td><td>۱۲:۴۹:۴۵</td><td></td></tr>';
  assert.equal(normalizeTgjuDollarPage(fixture("1,864,000", "۱۲:۴۹:۴۵", other), NOW).tomanPerUsd, 186_400);
});
test("supports ASCII, Persian and Arabic-Indic digits and Rial separators", () => {
  assert.equal(normalizeTgjuDollarPage(fixture("۱٬۸۶۴٬۰۰۰"), NOW).tomanPerUsd, 186_400);
  assert.equal(normalizeTgjuDollarPage(fixture("١٬٨٦٤٬٠٠٠", "١٢:٤٩:٤٥"), NOW).rawPriceRial, 1_864_000);
});
test("supports single-quoted data-price and falls back to the live-price cell", () => {
  assert.equal(normalizeTgjuDollarPage(fixture().replace('data-price="1,864,000"', "data-price='1,864,000'"), NOW).tomanPerUsd, 186_400);
  assert.equal(normalizeTgjuDollarPage(fixture().replace(' data-price="1,864,000"', ""), NOW).tomanPerUsd, 186_400);
});
test("rejects blocked, incomplete, redesigned or missing-market responses", () => {
  assert.throws(() => normalizeTgjuDollarPage("<html><body>Just a moment... CAPTCHA</body></html>", NOW), /complete HTML page/);
  assert.throws(() => normalizeTgjuDollarPage("<html><body>" + "x".repeat(700) + "</body></html>", NOW), /market row/);
  assert.throws(() => normalizeTgjuDollarPage(null, NOW), /complete HTML page/);
});
test("rejects missing, malformed and implausible Rial prices", () => {
  assert.throws(() => normalizeTgjuDollarPage(fixture(""), NOW), /whole-number Rial quote/);
  assert.throws(() => normalizeTgjuDollarPage(fixture("unknown"), NOW), /whole-number Rial quote/);
  assert.throws(() => normalizeTgjuDollarPage(fixture("9,999"), NOW), /supported Rial range/);
  assert.throws(() => normalizeTgjuDollarPage(fixture("100,000,001"), NOW), /supported Rial range/);
  assert.throws(() => normalizeTgjuDollarPage(fixture("1.25"), NOW), /whole-number Rial quote/);
});
test("rejects date-only or malformed source times and safely ages a future time-only clock", () => {
  assert.throws(() => normalizeTgjuDollarPage(fixture("1,864,000", "۲۵ آبان"), NOW), /live HH:mm:ss/);
  assert.throws(() => normalizeTgjuDollarPage(fixture("1,864,000", "not-a-time"), NOW), /live HH:mm:ss/);
  assert.throws(() => normalizeTgjuDollarPage(fixture("1,864,000", "12:49"), NOW), /live HH:mm:ss/);
  const futureClock = normalizeTgjuDollarPage(fixture("1,864,000", "12:51:00"), NOW);
  assert.equal(futureClock.sourceTimeTehran, "12:51:00");
  assert.equal(futureClock.sourceUpdatedAt, new Date(Date.parse("2026-11-02T09:21:00.000Z")).toISOString());
  assert.equal(getFxRateStatus(futureClock, NOW), "stale");
  assert.throws(() => normalizeTgjuDollarPage(fixture(), Number.NaN), /finite timestamp/);
});
test("classifies freshness from TGJU source update time", () => {
  const q = normalizeTgjuDollarPage(fixture(), NOW);
  assert.equal(getFxRateStatus(q, NOW), "fresh");
  assert.equal(getFxRateStatus({ ...q, sourceUpdatedAt: new Date(NOW - 5 * 60_000 - 1).toISOString() }, NOW), "aging");
  assert.equal(getFxRateStatus({ ...q, sourceUpdatedAt: new Date(NOW - 15 * 60_000 - 1).toISOString() }, NOW), "stale");
});
test("reports missing, malformed, non-positive and future rates as unavailable", () => {
  const q = normalizeTgjuDollarPage(fixture(), NOW);
  assert.equal(getFxRateStatus(null, NOW), "unavailable");
  assert.equal(getFxRateStatus({ tomanPerUsd: null, sourceUpdatedAt: null }, NOW), "unavailable");
  assert.equal(getFxRateStatus({ tomanPerUsd: Number.NaN, sourceUpdatedAt: q.sourceUpdatedAt }, NOW), "unavailable");
  assert.equal(getFxRateStatus({ tomanPerUsd: 0, sourceUpdatedAt: q.sourceUpdatedAt }, NOW), "unavailable");
  assert.equal(getFxRateStatus({ tomanPerUsd: 200_000, sourceUpdatedAt: "not-a-date" }, NOW), "unavailable");
  assert.equal(getFxRateStatus({ tomanPerUsd: 200_000, sourceUpdatedAt: new Date(NOW + 2 * 60_000).toISOString() }, NOW), "unavailable");
});
