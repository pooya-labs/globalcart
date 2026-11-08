"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import fxSnapshot from "../data/fx-rate.json";
import { calculateDemoPrice } from "../lib/pricing";
import { readLocalePreference, writeLocalePreference } from "../lib/locale";
import { getPersianProductCopy } from "../lib/catalog/editorial";
import { ArrowLeft, Clock3, Trash2 } from "lucide-react";

type Product = {
  id:number;
  title:string;
  brand?:string;
  price:number;
  discountPercentage:number;
  rating:number;
  thumbnail:string;
  priceAvailable?:boolean;
  ratingAvailable?:boolean;
  priceKind?:"estimated"|"verified";
};

export default function RecentlyViewedClient() {
  const [locale,setLocale]=useState<"fa"|"en">("fa");
  const [ids,setIds]=useState<number[]>([]);
  const [catalog,setCatalog]=useState<Product[]>([]);
  const [catalogLoadState,setCatalogLoadState]=useState<"idle"|"loading"|"loaded"|"failed">("idle");
  const [storageReady,setStorageReady]=useState(false);
  const rate = typeof fxSnapshot.tomanPerUsd === "number" && Number.isFinite(fxSnapshot.tomanPerUsd) && fxSnapshot.tomanPerUsd > 0 ? fxSnapshot.tomanPerUsd : null;
  const rtl=locale==="fa";
  const storageKey=process.env.NEXT_PUBLIC_E2E_FIXTURE_CATALOG==="1"?"globalcart-state-v1":"globalcart-state-v2";
  const catalogImageSrc=(source:string)=>(source.startsWith("/images/catalog/")?(process.env.NEXT_PUBLIC_BASE_PATH||"")+source:source);

  useEffect(()=>{
    if(ids.length===0){
      setCatalog([]);
      setCatalogLoadState("idle");
      return;
    }
    let active=true;
    setCatalogLoadState("loading");
    const basePath=process.env.NEXT_PUBLIC_BASE_PATH||"";
    fetch(basePath+"/data/catalog.json",{cache:"force-cache"})
      .then(response=>{
        if(!response.ok)throw new Error("Catalog snapshot request failed: "+response.status);
        return response.json() as Promise<{products?:unknown}>;
      })
      .then(payload=>{
        if(!active)return;
        if(Array.isArray(payload.products)){
          setCatalog(payload.products as Product[]);
          setCatalogLoadState("loaded");
        }else{
          setCatalogLoadState("failed");
        }
      })
      .catch(()=>{if(active)setCatalogLoadState("failed");});
    return ()=>{active=false;};
  },[ids.length]);

  useEffect(()=>{
    const savedLocale=readLocalePreference();
    if(savedLocale)setLocale(savedLocale);
  },[]);

  useEffect(()=>{
    try{
      const raw=localStorage.getItem(storageKey);
      const state=raw?JSON.parse(raw):{};
      if(Array.isArray(state.recent))setIds(state.recent.filter((id:number)=>Number.isInteger(id)).slice(0,8));
    }catch{/* Ignore invalid browser-local demo state. */}
    finally{setStorageReady(true);}
  },[]);

  const products=useMemo(()=>ids.map(id=>catalog.find(p=>p.id===id)).filter((p):p is Product=>Boolean(p)),[ids,catalog]);
  const format=(p:Product)=>(p.priceAvailable===false&&p.priceKind!=="estimated")
    ?(rtl?"قیمت ثبت نشده":"Price not provided")
    :rtl&&rate!==null
      ?new Intl.NumberFormat("fa-IR",{maximumFractionDigits:0}).format(calculateDemoPrice({sourcePrice:p.price,fxRate:rate,discountPercentage:p.discountPercentage}).toman)+" تومان"
      :"$"+(p.price*(1-p.discountPercentage/100)).toFixed(2);

  const toggleLocale=()=>{
    const nextLocale=rtl?"en":"fa";
    setLocale(nextLocale);
    writeLocalePreference(nextLocale);
  };

  const clearHistory=()=>{
    setIds([]);
    setCatalog([]);
    setCatalogLoadState("idle");
    try{
      const raw=localStorage.getItem(storageKey);
      const state=raw?JSON.parse(raw):{};
      localStorage.setItem(storageKey,JSON.stringify({...state,recent:[]}));
    }catch{/* Browser storage may be unavailable. */}
  };

  return <main className={rtl?"history-page rtl":"history-page ltr"} dir={rtl?"rtl":"ltr"} data-catalog-loaded={storageReady&&(catalogLoadState==="loaded"||ids.length===0)?"true":"false"}>
    <header className="history-topbar">
      <Link href="/" className="history-brand"><img src="../images/site/pooya-mark.svg" alt="Pooya Labs logo" width="34" height="34"/><span><b>globalcart</b><small>A POOYA LABS PRODUCT</small></span></Link>
      <div><button onClick={toggleLocale} aria-label={rtl?"EN — تغییر زبان به انگلیسی":"فا — Switch language to Persian"}>{rtl?"EN":"فا"}</button><Link href="/#catalog">{rtl?"بازگشت به فروشگاه":"Back to store"} <ArrowLeft size={15}/></Link></div>
    </header>
    <section className="history-content">
      <div className="history-heading"><div><span className="section-kicker">GLOBALCART / BROWSING HISTORY</span><h1>{rtl?"محصولات بازدیدشده":"Recently viewed products"}</h1><p>{rtl?"محصولاتی که در همین مرورگر بررسی کرده‌ای؛ تاریخچه به حساب کاربری یا سرور ارسال نمی‌شود.":"Products you opened in this browser. This history is stored locally and is not sent to an account or server."}</p></div>{products.length>0&&<button className="history-clear" onClick={clearHistory}><Trash2 size={15}/>{rtl?"پاک کردن تاریخچه":"Clear history"}</button>}</div>
      {ids.length>0&&catalogLoadState==="loading"?<div className="history-empty"><Clock3 size={34}/><b>{rtl?"در حال بارگذاری تاریخچه...":"Loading your browsing history..."}</b></div>:catalogLoadState==="failed"?<div className="history-empty"><Clock3 size={34}/><b>{rtl?"دریافت تاریخچه ناموفق بود. دوباره صفحه را بارگذاری کن.":"Could not load your history. Reload the page to retry."}</b></div>:products.length===0?<div className="history-empty"><Clock3 size={34}/><b>{rtl?"هنوز محصولی بازدید نکرده‌ای":"No recently viewed products yet"}</b><p>{rtl?"با مشاهده جزئیات محصولات، آن‌ها اینجا نمایش داده می‌شوند.":"Open a product’s detail page and it will appear here."}</p><Link href="/#catalog">{rtl?"رفتن به فروشگاه":"Explore the catalog"} <ArrowLeft size={15}/></Link></div>:<div className="history-grid">{products.map(p=><Link href={`/product/${p.id}`} key={p.id} className="history-card"><img src={catalogImageSrc(p.thumbnail)} alt={p.title} loading="lazy" decoding="async" fetchPriority="low" width={1440} height={1440}/><div><small>{p.brand||"GLOBAL SELECT"}</small><b>{rtl?(getPersianProductCopy(p.id)?.title??p.title):p.title}</b>{rtl&&getPersianProductCopy(p.id)&&<small className="history-source-title" dir="auto">{p.title}</small>}<span>{p.ratingAvailable===false?(rtl?"امتیاز ثبت نشده":"No rating supplied"):"★ "+p.rating.toFixed(1)}</span><strong>{format(p)}</strong></div><ArrowLeft size={16}/></Link>)}</div>}
      <p className="history-note">{rtl?"تاریخچه فقط در همین مرورگر ذخیره می‌شود. قیمت‌ها برآورد نمایشی GlobalCart هستند، نه قیمت فروشنده؛ منبع امتیاز و موجودی ارائه نمی‌کند.":"History is stored only in this browser. Prices are local demo estimates, not seller quotes; the source does not provide ratings or stock."}</p>
    </section>
  </main>;
}
