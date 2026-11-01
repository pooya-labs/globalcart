"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import catalogSnapshot from "../../data/catalog.json";
import { ArrowLeft, Clock3, Trash2 } from "lucide-react";

type Product = {
  id:number;
  title:string;
  brand?:string;
  price:number;
  discountPercentage:number;
  rating:number;
  thumbnail:string;
};

const catalog = Array.isArray(catalogSnapshot.products)
  ? catalogSnapshot.products as unknown as Product[]
  : [];

export default function RecentlyViewedPage() {
  const [locale,setLocale]=useState<"fa"|"en">("fa");
  const [ids,setIds]=useState<number[]>([]);
  const [rate,setRate]=useState(69000);
  const rtl=locale==="fa";

  useEffect(()=>{
    try{
      const raw=localStorage.getItem("globalcart-state-v1");
      const state=raw?JSON.parse(raw):{};
      if(Array.isArray(state.recent))setIds(state.recent.filter((id:number)=>Number.isInteger(id)).slice(0,8));
      if(Number.isFinite(state.rate)&&state.rate>0)setRate(state.rate);
    }catch{/* Ignore invalid browser-local demo state. */}
  },[]);

  const products=useMemo(()=>ids.map(id=>catalog.find(p=>p.id===id)).filter((p):p is Product=>Boolean(p)),[ids]);
  const format=(p:Product)=>rtl
    ?new Intl.NumberFormat("fa-IR",{maximumFractionDigits:0}).format(p.price*(1-p.discountPercentage/100)*rate)+" تومان"
    :"$"+(p.price*(1-p.discountPercentage/100)).toFixed(2);

  const clearHistory=()=>{
    setIds([]);
    try{
      const raw=localStorage.getItem("globalcart-state-v1");
      const state=raw?JSON.parse(raw):{};
      localStorage.setItem("globalcart-state-v1",JSON.stringify({...state,recent:[]}));
    }catch{/* Browser storage may be unavailable. */}
  };

  return <main className={rtl?"history-page rtl":"history-page ltr"} dir={rtl?"rtl":"ltr"}>
    <header className="history-topbar">
      <Link href="/"><span className="brand-mark">g<span>.</span></span><b>globalcart</b></Link>
      <div><button onClick={()=>setLocale(rtl?"en":"fa")}>{rtl?"EN":"فا"}</button><Link href="/#catalog">{rtl?"بازگشت به فروشگاه":"Back to store"} <ArrowLeft size={15}/></Link></div>
    </header>
    <section className="history-content">
      <div className="history-heading"><div><span className="section-kicker">GLOBALCART / BROWSING HISTORY</span><h1>{rtl?"محصولات بازدیدشده":"Recently viewed products"}</h1><p>{rtl?"محصولاتی که در همین مرورگر بررسی کرده‌ای؛ تاریخچه به حساب کاربری یا سرور ارسال نمی‌شود.":"Products you opened in this browser. This history is stored locally and is not sent to an account or server."}</p></div>{products.length>0&&<button className="history-clear" onClick={clearHistory}><Trash2 size={15}/>{rtl?"پاک کردن تاریخچه":"Clear history"}</button>}</div>
      {products.length===0?<div className="history-empty"><Clock3 size={34}/><b>{rtl?"هنوز محصولی بازدید نکرده‌ای":"No recently viewed products yet"}</b><p>{rtl?"با مشاهده جزئیات محصولات، آن‌ها اینجا نمایش داده می‌شوند.":"Open a product’s detail page and it will appear here."}</p><Link href="/#catalog">{rtl?"رفتن به فروشگاه":"Explore the catalog"} <ArrowLeft size={15}/></Link></div>:<div className="history-grid">{products.map(p=><Link href={`/product/${p.id}`} key={p.id} className="history-card"><img src={p.thumbnail} alt={p.title} loading="lazy"/><div><small>{p.brand||"GLOBAL SELECT"}</small><b>{p.title}</b><span>★ {p.rating.toFixed(1)}</span><strong>{format(p)}</strong></div><ArrowLeft size={16}/></Link>)}</div>}
      <p className="history-note">{rtl?"تاریخچه فقط روی همین دستگاه و مرورگر ذخیره می‌شود. پاک‌کردن آن روی سبد خرید و علاقه‌مندی‌ها اثری ندارد.":"History is stored only in this browser on this device. Clearing it does not affect your cart or saved items."}</p>
    </section>
  </main>;
}
