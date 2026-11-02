import test from "node:test";
import assert from "node:assert/strict";
import { getFxRateStatus, normalizeWallexMinuteHistory } from "../lib/fx/wallex";
const NOW = Date.parse("2026-11-02T12:00:00.000Z");
const validHistory = { s: "ok", t: [Math.floor(NOW / 1000) - 120, Math.floor(NOW / 1000) - 60], c: ["272100.0000000000000000", "272250.0000000000000000"] };
test("normalizes the latest USDTTMN one-minute close in Toman", () => {
 const q=normalizeWallexMinuteHistory(validHistory,NOW);
 assert.equal(q.provider,"Wallex"); assert.equal(q.market,"USDTTMN"); assert.equal(q.baseCurrency,"USDT");
 assert.equal(q.quoteCurrency,"TMN"); assert.equal(q.rawQuoteUnit,"TMN"); assert.equal(q.priceSource,"ONE_MINUTE_CANDLE_CLOSE");
 assert.equal(q.candleResolutionMinutes,1); assert.equal(q.tomanPerUsdt,272250);
 assert.equal(q.candleStartedAt,new Date((Math.floor(NOW/1000)-60)*1000).toISOString());
});
test("selects newest candle by timestamp, not array position",()=>{
 const q=normalizeWallexMinuteHistory({s:"ok",t:[Math.floor(NOW/1000)-60,Math.floor(NOW/1000)-180,Math.floor(NOW/1000)-120],c:["272250","272000","272100"]},NOW);
 assert.equal(q.tomanPerUsdt,272250);
});
test("rejects bad status, empty or mismatched candle arrays",()=>{
 assert.throws(()=>normalizeWallexMinuteHistory({...validHistory,s:"error"},NOW),/status/);
 assert.throws(()=>normalizeWallexMinuteHistory({s:"no_data",t:[],c:[]},NOW),/status/);
 assert.throws(()=>normalizeWallexMinuteHistory({s:"ok",t:[],c:[]},NOW),/complete candle series/);
 assert.throws(()=>normalizeWallexMinuteHistory({s:"ok",t:[Math.floor(NOW/1000)],c:[]},NOW),/complete candle series/);
});
test("rejects malformed timestamps and invalid or implausible candle closes",()=>{
 assert.throws(()=>normalizeWallexMinuteHistory({s:"ok",t:["not-a-time"],c:["272250"]},NOW),/candle timestamp/);
 assert.throws(()=>normalizeWallexMinuteHistory({s:"ok",t:[Math.floor(NOW/1000)],c:["0"]},NOW),/candle close/);
 assert.throws(()=>normalizeWallexMinuteHistory({s:"ok",t:[Math.floor(NOW/1000)],c:["9999"]},NOW),/supported Toman range/);
 assert.throws(()=>normalizeWallexMinuteHistory({s:"ok",t:[Math.floor(NOW/1000)],c:["10000001"]},NOW),/supported Toman range/);
});
test("rejects future candle timestamps and invalid clock values",()=>{
 assert.throws(()=>normalizeWallexMinuteHistory({s:"ok",t:[Math.floor(NOW/1000)+120],c:["272250"]},NOW),/future/);
 assert.throws(()=>normalizeWallexMinuteHistory(validHistory,Number.NaN),/finite timestamp/);
});
test("classifies freshness from exchange candle-start timestamp",()=>{
 const q=normalizeWallexMinuteHistory(validHistory,NOW);
 assert.equal(getFxRateStatus(q,NOW),"fresh");
 assert.equal(getFxRateStatus({...q,candleStartedAt:new Date(NOW-5*60*1000-1).toISOString()},NOW),"aging");
 assert.equal(getFxRateStatus({...q,candleStartedAt:new Date(NOW-15*60*1000-1).toISOString()},NOW),"stale");
});
test("reports missing, malformed and future quotes as unavailable",()=>{
 assert.equal(getFxRateStatus(null,NOW),"unavailable");
 assert.equal(getFxRateStatus({tomanPerUsdt:null,candleStartedAt:null},NOW),"unavailable");
 assert.equal(getFxRateStatus({tomanPerUsdt:Number.NaN,candleStartedAt:new Date(NOW).toISOString()},NOW),"unavailable");
 assert.equal(getFxRateStatus({tomanPerUsdt:250000,candleStartedAt:"not-a-date"},NOW),"unavailable");
 assert.equal(getFxRateStatus({tomanPerUsdt:250000,candleStartedAt:new Date(NOW+2*60*1000).toISOString()},NOW),"unavailable");
});
