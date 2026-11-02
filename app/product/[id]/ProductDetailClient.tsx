"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { ArrowLeft, Check, ChevronLeft, Heart, Minus, Plus, ShieldCheck, ShoppingBag, Star, Truck } from "lucide-react";
import { calculateDemoPrice } from "../../../lib/pricing";
import fxSnapshot from "../../../data/fx-rate.json";
import { getFxRateStatus, type FxRateStatus } from "../../../lib/fx/wallex";

type ProductVariant = {
  sku:string; title:string; price:number; stock:number;
  attributes?:Record<string,unknown>; source:string; sourceId:string; isActive:boolean;
};
type Product = {
  id:number; title:string; description:string; category:string; price:number;
  discountPercentage:number; rating:number; stock:number; brand?:string;
  thumbnail:string; images?:string[]; source?:string; sourceId?:string;
  sourceUrl?:string; importedAt?:string; variants?:ProductVariant[];
};
type Props = { product:Product; related:Product[] };
type CartLine = { product:Product; quantity:number };

const categoryNames:Record<string,{fa:string;en:string}> = {
  beauty:{fa:"زیبایی و مراقبت",en:"Beauty & personal care"},
  fragrances:{fa:"عطر و رایحه",en:"Fragrances"},
  furniture:{fa:"خانه و دکور",en:"Home & living"},
  groceries:{fa:"خوراکی",en:"Groceries"},
  laptops:{fa:"لپ‌تاپ",en:"Laptops"},
  smartphones:{fa:"موبایل",en:"Smartphones"},
  "mens-watches":{fa:"ساعت",en:"Watches"},
  electronics:{fa:"الکترونیک",en:"Electronics"},
  "mobile-accessories":{fa:"لوازم جانبی",en:"Accessories"},
  "home-decoration":{fa:"دکوراسیون",en:"Home decoration"},
  "kitchen-accessories":{fa:"لوازم آشپزخانه",en:"Kitchen accessories"},
  "skin-care":{fa:"مراقبت پوست",en:"Skin care"},
  "sports-accessories":{fa:"ورزش",en:"Sports accessories"},
  "womens-bags":{fa:"کیف زنانه",en:"Women's bags"},
  "womens-dresses":{fa:"لباس زنانه",en:"Women's dresses"},
  "womens-jewellery":{fa:"زیورآلات",en:"Women's jewellery"},
  "womens-shoes":{fa:"کفش زنانه",en:"Women's shoes"},
  "mens-shirts":{fa:"پیراهن مردانه",en:"Men's shirts"},
  "mens-shoes":{fa:"کفش مردانه",en:"Men's shoes"}
};

export default function ProductDetailClient({product,related}:Props) {
  const [locale,setLocale]=useState<"fa"|"en">("fa");
  const [activeImage,setActiveImage]=useState(0);
  const [quantity,setQuantity]=useState(1);
  const [added,setAdded]=useState(false);
  const [saved,setSaved]=useState(false);
  const rate = typeof fxSnapshot.tomanPerUsdt === "number" && Number.isFinite(fxSnapshot.tomanPerUsdt) && fxSnapshot.tomanPerUsdt > 0 ? fxSnapshot.tomanPerUsdt : null;
  const [fxStatus,setFxStatus]=useState<FxRateStatus>("unavailable");
  const rtl=locale==="fa";
  const images=useMemo(()=>Array.from(new Set([product.thumbnail,...(product.images||[])].filter(Boolean))),[product]);
  const image=images[Math.min(activeImage,images.length-1)]||product.thumbnail;
  const discounted=product.price*(1-product.discountPercentage/100);
  const priceBreakdown=rate===null?null:calculateDemoPrice({sourcePrice:product.price,fxRate:rate,discountPercentage:product.discountPercentage});
  const fxStatusLabel=fxStatus==="fresh"?(rtl?"به‌روز":"Fresh"):fxStatus==="aging"?(rtl?"در حال کهنه‌شدن":"Aging"):fxStatus==="stale"?(rtl?"قدیمی":"Stale"):(rtl?"در دسترس نیست":"Unavailable");
  const fxTimestampLabel=fxSnapshot.candleStartedAt?new Intl.DateTimeFormat(rtl?"fa-IR":"en-GB",{timeZone:"Asia/Tehran",year:"numeric",month:"short",day:"numeric",hour:"2-digit",minute:"2-digit"}).format(new Date(fxSnapshot.candleStartedAt)):null;
  useEffect(()=>{setFxStatus(getFxRateStatus(fxSnapshot));},[]);
  useEffect(()=>{
    try{
      const raw=localStorage.getItem("globalcart-state-v1");
      const state=raw?JSON.parse(raw):{};
      if(Array.isArray(state.favorites))setSaved(state.favorites.includes(product.id));
      const recentIds:number[]=Array.isArray(state.recent)?state.recent:[];
      localStorage.setItem("globalcart-state-v1",JSON.stringify({...state,recent:[product.id,...recentIds.filter(id=>id!==product.id)].slice(0,8)}));
    }catch{/* Ignore invalid local demo state. */}
  },[product.id]);
  const format=(value:number)=>rtl&&rate!==null
    ? new Intl.NumberFormat("fa-IR",{maximumFractionDigits:0}).format(calculateDemoPrice({sourcePrice:value,fxRate:rate}).toman)+" تومان"
    : "$"+value.toFixed(2);
  const category=categoryNames[product.category]||{fa:product.category,en:product.category.replaceAll("-"," ")};
  const addToCart=()=>{
    try{
      const raw=localStorage.getItem("globalcart-state-v1");
      const state=raw?JSON.parse(raw):{};
      const cart:CartLine[]=Array.isArray(state.cart)?state.cart:[];
      const existing=cart.find(line=>line.product.id===product.id);
      const next=existing
        ?cart.map(line=>line.product.id===product.id?{...line,quantity:line.quantity+quantity}:line)
        :[...cart,{product,quantity}];
      localStorage.setItem("globalcart-state-v1",JSON.stringify({...state,cart}));
      setAdded(true);
    }catch{
      setAdded(false);
    }
  };
  const toggleSaved=()=>{
    try{
      const raw=localStorage.getItem("globalcart-state-v1");
      const state=raw?JSON.parse(raw):{};
      const favorites:number[]=Array.isArray(state.favorites)?state.favorites:[];
      const next=favorites.includes(product.id)?favorites.filter(id=>id!==product.id):[...favorites,product.id];
      localStorage.setItem("globalcart-state-v1",JSON.stringify({...state,favorites:next}));
      setSaved(next.includes(product.id));
    }catch{/* Browser storage may be unavailable. */}
  };

  return <main className={rtl?"product-page rtl":"product-page ltr"} dir={rtl?"rtl":"ltr"}>
    <div className="product-page-top"><Link href="/"><span className="brand-mark">g<span>.</span></span><b>globalcart</b></Link><div><button onClick={()=>setLocale(rtl?"en":"fa")}>{rtl?"EN":"فا"}</button><Link href="/#catalog">{rtl?"بازگشت به فروشگاه":"Back to store"} <ArrowLeft size={15}/></Link></div></div>
    <div className="product-breadcrumb"><Link href="/">{rtl?"خانه":"Home"}</Link><ChevronLeft size={13}/><Link href="/#categories">{rtl?"دسته‌بندی‌ها":"Categories"}</Link><ChevronLeft size={13}/><span>{rtl?category.fa:category.en}</span><ChevronLeft size={13}/><b>{product.title}</b></div>
    <section className="product-detail-layout">
      <div className="detail-gallery">
        <div className="detail-main-image"><img src={image} alt={product.title} loading="eager" fetchPriority="high" decoding="async"/><span className="detail-discount">{Math.round(product.discountPercentage)}% {rtl?"تخفیف":"OFF"}</span></div>
        <div className="detail-thumbnails">{images.map((src,index)=><button key={src} className={activeImage===index?"active":""} onClick={()=>setActiveImage(index)} aria-label={rtl?`تصویر ${index+1}`:`Image ${index+1}`} aria-pressed={activeImage===index}><img src={src} alt="" loading="lazy" decoding="async"/></button>)}</div>
      </div>
      <div className="detail-information">
        <span className="detail-brand">{product.brand||"GLOBAL SELECT"}</span>
        <h1>{product.title}</h1>
        <div className="detail-rating"><span><Star size={15} fill="currentColor"/>{product.rating.toFixed(1)}</span><i/> <span>{rtl?"شناسه محصول":"Product ID"}: {product.id}</span><i/><span>{rtl?category.fa:category.en}</span></div>
        <p className="detail-description">{product.description}</p>
        <div className="detail-highlights"><div><ShieldCheck/><span><b>{rtl?"اطلاعات شفاف":"Clear information"}</b><small>{rtl?"قیمت و منبع داده مشخص است":"Price and data source disclosed"}</small></span></div><div><Truck/><span><b>{rtl?"تجربه خرید نمایشی":"Demo shopping experience"}</b><small>{rtl?"بدون ارسال یا موجودی واقعی":"No real shipping or inventory"}</small></span></div></div>
        <div className="detail-specs"><div><span>{rtl?"دسته‌بندی":"Category"}</span><b>{rtl?category.fa:category.en}</b></div><div><span>{rtl?"برند":"Brand"}</span><b>{product.brand||"Global Select"}</b></div><div><span>{rtl?"امتیاز نمونه":"Sample rating"}</span><b>{product.rating.toFixed(1)} / 5</b></div><div><span>{rtl?"موجودی در منبع نمونه":"Source stock figure"}</span><b>{product.stock}</b></div></div>
        <section className="detail-provenance" aria-label={rtl?"منشأ و ویژگی‌های داده":"Data provenance and attributes"}>
          <div className="detail-provenance-heading"><b>{rtl?"منشأ و ویژگی‌های داده":"Data provenance & attributes"}</b><span>{product.source||"Unknown source"}</span></div>
          <div className="detail-provenance-grid">
            <div><span>{rtl?"شناسه در منبع":"Source record ID"}</span><b>{product.sourceId||String(product.id)}</b></div>
            <div><span>{rtl?"زمان ورود به کاتالوگ":"Imported into catalog"}</span><b>{product.importedAt?new Intl.DateTimeFormat(rtl?"fa-IR":"en-GB",{timeZone:"Asia/Tehran",dateStyle:"medium",timeStyle:"short"}).format(new Date(product.importedAt)):(rtl?"نامشخص":"Unknown")}</b></div>
            <div><span>{rtl?"زمینه امتیاز":"Rating context"}</span><b>{rtl?"امتیاز نمونه منبع؛ تعداد نظر ارائه نشده":"Source sample rating; review count not supplied"}</b></div>
            {product.sourceUrl&&/^https:\/\//i.test(product.sourceUrl)&&<div><span>{rtl?"پیوند منبع":"Source record URL"}</span><a href={product.sourceUrl} target="_blank" rel="noreferrer">{rtl?"مشاهده رکورد منبع":"View source record"}</a></div>}
          </div>
          <div className="detail-attributes">
            <b>{rtl?"ویژگی‌ها و گزینه‌ها":"Attributes & options"}</b>
            {(product.variants?.length??0)>0
              ? product.variants?.map(variant=><div className="detail-variant" key={variant.source+"-"+variant.sourceId}>
                  <strong>{variant.title||variant.sku||variant.sourceId}</strong>
                  {Object.entries(variant.attributes??{}).map(([key,value])=><span key={key}>{key}: {typeof value==="string"||typeof value==="number"||typeof value==="boolean"?String(value):JSON.stringify(value)}</span>)}
                  <small>{variant.isActive?(rtl?"فعال":"Active"):(rtl?"غیرفعال":"Inactive")} · {format(variant.price)} · {rtl?"موجودی منبع":"Source stock"}: {variant.stock}</small>
                </div>)
              : <p>{rtl?"منبع نمونه برای این محصول، گزینه یا ویژگی ساختاریافته‌ای ارسال نکرده است.":"The sample source did not provide structured variants or product attributes for this item."}</p>}
          </div>
        </section>
      </div>
      <aside className="detail-buybox">
        <span className="buybox-kicker">{rtl?"برآورد قیمت با نرخ بازار":"EXCHANGE-BASED ESTIMATE"}</span>
        <div className="detail-price">{priceBreakdown?(rtl?new Intl.NumberFormat("fa-IR",{maximumFractionDigits:0}).format(priceBreakdown.toman)+" تومان":"$"+priceBreakdown.discountedUsd.toFixed(2)):"$"+discounted.toFixed(2)}</div>
        <div className="detail-old-price"><s>{format(product.price)}</s><span>{rtl?"قیمت قبل از تخفیف":"Before discount"}</span></div>
        <details className="price-breakdown"><summary>{rtl?"منبع و جزئیات نرخ تبدیل":"FX source and calculation"}</summary>
          <div><span>{rtl?"قیمت منبع (USD)":"Source catalog price (USD)"}</span><b>${product.price.toFixed(2)}</b></div>
          <div><span>{rtl?"تخفیف نمونه":"Catalog sample discount"}</span><b>{product.discountPercentage.toFixed(2)}%</b></div>
          {priceBreakdown?<>
            <div><span>{rtl?"قیمت بسته‌شدن کندل یک‌دقیقه‌ای والکس":"Wallex one-minute candle close"}</span><b>{new Intl.NumberFormat(rtl?"fa-IR":"en-US",{maximumFractionDigits:0}).format(priceBreakdown.fxRate)} {rtl?"تومان / تتر":"Toman / USDT"}</b></div>
            <div><span>{rtl?"زمان شروع آخرین کندل":"Latest candle start time"}</span><b>{fxTimestampLabel??fxStatusLabel}</b></div>
            <small>{rtl?"مبنای تبدیل، قیمت بسته‌شدن آخرین کندل یک‌دقیقه‌ای بازار USDTTMN والکس است. قیمت برحسب تومان و مربوط به تتر است؛ با دلار نقدی تفاوت دارد و نرخ قابل اجرای خرید/فروش محسوب نمی‌شود.":"Conversion uses the latest one-minute candle close for Wallex USDTTMN. The quote is Toman-denominated and refers to Tether, which can differ from cash USD; it is not an executable buy/sell quote."}</small>
            {fxStatus!=="fresh"&&<small className="fx-warning">{rtl?"هشدار تازگی نرخ: "+fxStatusLabel:"FX freshness warning: "+fxStatusLabel}</small>}
          </>:<small>{rtl?"نرخ معتبر صرافی فعلاً در دسترس نیست؛ مبلغ اصلی به دلار نمایش داده می‌شود.":"A valid exchange quote is currently unavailable, so the source USD price is shown."}</small>}
          <small>{rtl?"قیمت‌ها نمایشی‌اند و شامل هزینه ارسال، مالیات، کارمزد پرداخت یا تعهد خرید نیستند.":"Demo prices only; no shipping, tax, payment fees or purchase commitment are included."}</small>
        </details>
        <div className="buybox-stock"><span/> {rtl?"موجود در کاتالوگ نمونه":"Listed in sample catalog"}</div>
        <div className="buybox-quantity"><span>{rtl?"تعداد":"Quantity"}</span><div><button onClick={()=>setQuantity(q=>Math.max(1,q-1))} aria-label={rtl?"کاهش تعداد":"Decrease quantity"}><Minus size={15}/></button><b>{quantity}</b><button onClick={()=>setQuantity(q=>Math.min(10,q+1))} aria-label={rtl?"افزایش تعداد":"Increase quantity"}><Plus size={15}/></button></div></div>
        <button className="buybox-add" onClick={addToCart}><ShoppingBag size={17}/>{rtl?"افزودن به سبد خرید":"Add to bag"}</button>
        <button className="buybox-save" onClick={toggleSaved}><Heart size={16} fill={saved?"currentColor":"none"}/>{saved?(rtl?"حذف از علاقه‌مندی‌ها":"Remove from saved"):(rtl?"ذخیره در علاقه‌مندی‌ها":"Save for later")}</button>
        {added&&<div className="buybox-confirm"><Check size={16}/>{rtl?"به سبد خرید اضافه شد":"Added to your bag"}</div>}
        <div className="buybox-disclaimer"><ShieldCheck size={16}/><span>{rtl?"این یک فروشگاه نمایشی است؛ پرداخت و ارسال واقعی انجام نمی‌شود.":"Demo storefront only. No real payment or shipping."}</span></div>
        <small className="buybox-source">{rtl?"منبع اطلاعات":"Data source"}: {product.source||"DummyJSON public demo API"}</small>
      </aside>
    </section>
    <section className="detail-description-panel"><span className="section-kicker">PRODUCT OVERVIEW</span><h2>{rtl?"درباره محصول":"About this product"}</h2><p>{product.description}</p><div><span>{rtl?"شناسه منبع":"Source ID"}</span><b>{product.sourceId||product.id}</b><span>{rtl?"آخرین قیمت نمونه":"Sample price"}</span><b>{format(discounted)}</b></div></section>
    {related.length>0&&<section className="detail-related"><div><span className="section-kicker">YOU MAY ALSO LIKE</span><h2>{rtl?"محصولات مشابه":"Related products"}</h2></div><div className="related-grid">{related.map(item=><Link href={`/product/${item.id}`} key={item.id}><img src={item.thumbnail} alt={item.title} loading="lazy" decoding="async"/><span>{item.brand||"GLOBAL SELECT"}</span><b>{item.title}</b><strong>{format(item.price*(1-item.discountPercentage/100))}</strong></Link>)}</div></section>}
    <footer className="product-page-footer"><Link href="/">globalcart</Link><span>{rtl?"پروژه مستقل نمایشی · توسعه و طراحی پویا گودرزی":"Independent portfolio demo · Designed & built by Pooya Goodarzi"}</span></footer>
  </main>;
}
