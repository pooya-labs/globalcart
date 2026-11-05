"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { ArrowLeft, Check, ChevronLeft, ChevronRight, Heart, Minus, Plus, ShieldCheck, ShoppingBag, Star, Truck } from "lucide-react";
import { calculateDemoPrice } from "../../../lib/pricing";
import fxSnapshot from "../../../data/fx-rate.json";
const fxQuote = fxSnapshot as unknown as FxRateSnapshot;
import { getFxRateStatus, type FxRateStatus, type FxRateSnapshot } from "../../../lib/fx/tgju";
import { readLocalePreference, writeLocalePreference, type Locale } from "../../../lib/locale";

type ProductVariant = {
  sku:string; title:string; price:number; stock:number;
  attributes?:Record<string,unknown>; source:string; sourceId:string; isActive:boolean;
};
type Product = {
  id:number; title:string; description:string; category:string; price:number;
  discountPercentage:number; rating:number; stock:number; brand?:string;
  thumbnail:string; images?:string[]; source?:string; sourceId?:string;
  importedAt?:string; variants?:ProductVariant[];
  priceAvailable?:boolean; ratingAvailable?:boolean; stockAvailable?:boolean;
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
  "mens-shoes":{fa:"کفش مردانه",en:"Men's shoes"},
  "home-decor":{fa:"دکور و خانه",en:"Home decor"},
  kitchen:{fa:"آشپزخانه",en:"Kitchen"},
  shoes:{fa:"کفش",en:"Shoes"},
  bags:{fa:"کیف و چمدان",en:"Bags & luggage"},
  jewelry:{fa:"زیورآلات",en:"Jewelry"},
  books:{fa:"کتاب",en:"Books"},
  automotive:{fa:"خودرو",en:"Automotive"},
  tools:{fa:"ابزار",en:"Tools"},
  baby:{fa:"کودک و نوزاد",en:"Baby"},
  pets:{fa:"حیوانات خانگی",en:"Pets"},
  office:{fa:"اداری و نوشت‌افزار",en:"Office"},
  music:{fa:"موسیقی",en:"Musical instruments"}
};

export default function ProductDetailClient({product,related}:Props) {
  const [locale,setLocale]=useState<Locale>("fa");
  const [activeImage,setActiveImage]=useState(0);
  const [quantity,setQuantity]=useState(1);
  const [added,setAdded]=useState(false);
  const [saved,setSaved]=useState(false);
  const rate = typeof fxQuote.tomanPerUsd === "number" && Number.isFinite(fxSnapshot.tomanPerUsd) && fxSnapshot.tomanPerUsd > 0 ? fxSnapshot.tomanPerUsd : null;
  const [fxStatus,setFxStatus]=useState<FxRateStatus>("unavailable");
  const rtl=locale==="fa";
  const toggleLocale=()=>{
    const next=rtl?"en":"fa";
    setLocale(next);
    writeLocalePreference(next);
  };
  const catalogImageSrc=(source:string)=>(source.startsWith("/images/catalog/")?(process.env.NEXT_PUBLIC_BASE_PATH||"")+source:source);
  const storageKey=process.env.NEXT_PUBLIC_E2E_FIXTURE_CATALOG==="1"?"globalcart-state-v1":"globalcart-state-v2";
  const images=useMemo(()=>Array.from(new Set([product.thumbnail,...(product.images||[])].filter(Boolean))),[product]);
  const image=images[Math.min(activeImage,images.length-1)]||product.thumbnail;
  const previousImage=()=>setActiveImage(index=>(index+images.length-1)%images.length);
  const nextImage=()=>setActiveImage(index=>(index+1)%images.length);
  const discounted=product.price*(1-product.discountPercentage/100);
  const priceBreakdown=product.priceAvailable===false||rate===null?null:calculateDemoPrice({sourcePrice:product.price,fxRate:rate,discountPercentage:product.discountPercentage});
  const fxStatusLabel=fxStatus==="fresh"?(rtl?"به‌روز":"Fresh"):fxStatus==="aging"?(rtl?"در حال کهنه‌شدن":"Aging"):fxStatus==="stale"?(rtl?"قدیمی":"Stale"):(rtl?"در دسترس نیست":"Unavailable");
  const fxTimestampLabel=fxQuote.sourceTimeKind==="date_only"&&fxQuote.sourceTimeTehran
    ?(rtl?fxQuote.sourceTimeTehran.replace(/\d+/g,n=>new Intl.NumberFormat("fa-IR").format(Number(n))):fxQuote.sourceTimeTehran)
    :typeof fxQuote.sourceUpdatedAt==="string"?new Intl.DateTimeFormat(rtl?"fa-IR":"en-GB",{timeZone:"Asia/Tehran",hour:"2-digit",minute:"2-digit",second:"2-digit",hourCycle:"h23"}).format(new Date(fxQuote.sourceUpdatedAt)):null;
  useEffect(()=>{
    const savedLocale=readLocalePreference();
    if(savedLocale)setLocale(savedLocale);
  },[]);
  useEffect(()=>{setFxStatus(getFxRateStatus(fxQuote));},[]);
  useEffect(()=>{
    try{
      const raw=localStorage.getItem(storageKey);
      const state=raw?JSON.parse(raw):{};
      if(Array.isArray(state.favorites))setSaved(state.favorites.includes(product.id));
      const recentIds:number[]=Array.isArray(state.recent)?state.recent:[];
      localStorage.setItem(storageKey,JSON.stringify({...state,recent:[product.id,...recentIds.filter(id=>id!==product.id)].slice(0,8)}));
    }catch{/* Ignore invalid local demo state. */}
  },[product.id]);
  const format=(value:number)=>product.priceAvailable===false
    ? (rtl?"قیمت ثبت نشده":"Price not listed")
    : rtl&&rate!==null
      ? new Intl.NumberFormat("fa-IR",{maximumFractionDigits:0}).format(calculateDemoPrice({sourcePrice:value,fxRate:rate}).toman)+" تومان"
      : "$"+value.toFixed(2);
  const category=categoryNames[product.category]||{fa:product.category,en:product.category.replaceAll("-"," ")};
  const addToCart=()=>{
    try{
      const raw=localStorage.getItem(storageKey);
      const state=raw?JSON.parse(raw):{};
      const cart:CartLine[]=Array.isArray(state.cart)?state.cart:[];
      const existing=cart.find(line=>line.product.id===product.id);
      const next=existing
        ?cart.map(line=>line.product.id===product.id?{...line,quantity:line.quantity+quantity}:line)
        :[...cart,{product,quantity}];
      localStorage.setItem(storageKey,JSON.stringify({...state,cart:next}));
      setAdded(true);
    }catch{
      setAdded(false);
    }
  };
  const toggleSaved=()=>{
    try{
      const raw=localStorage.getItem(storageKey);
      const state=raw?JSON.parse(raw):{};
      const favorites:number[]=Array.isArray(state.favorites)?state.favorites:[];
      const next=favorites.includes(product.id)?favorites.filter(id=>id!==product.id):[...favorites,product.id];
      localStorage.setItem(storageKey,JSON.stringify({...state,favorites:next}));
      setSaved(next.includes(product.id));
    }catch{/* Browser storage may be unavailable. */}
  };

  return <main className={rtl?"product-page rtl":"product-page ltr"} dir={rtl?"rtl":"ltr"}>
    <div className="product-page-top"><Link href="/"><span className="brand-mark">g<span>.</span></span><b>globalcart</b></Link><div><button onClick={toggleLocale}>{rtl?"EN":"فا"}</button><Link href="/#catalog">{rtl?"بازگشت به فروشگاه":"Back to store"} <ArrowLeft size={15}/></Link></div></div>
    <div className="product-breadcrumb"><Link href="/">{rtl?"خانه":"Home"}</Link><ChevronLeft size={13}/><Link href="/#categories">{rtl?"دسته‌بندی‌ها":"Categories"}</Link><ChevronLeft size={13}/><span>{rtl?category.fa:category.en}</span><ChevronLeft size={13}/><b>{product.title}</b></div>
    <section className="product-detail-layout">
      <div className="detail-gallery">
        <div className="detail-main-image">
          <img src={catalogImageSrc(image)} alt={product.title+" - "+String(activeImage+1)} loading="eager" fetchPriority="high" decoding="async" width={1440} height={1440}/>
          {product.priceAvailable!==false&&product.discountPercentage>0&&<span className="detail-discount">{Math.round(product.discountPercentage)}% {rtl?"تخفیف":"OFF"}</span>}
          {images.length>1&&<>
            <button type="button" className="gallery-arrow gallery-arrow-prev" onClick={previousImage} aria-label={rtl?"تصویر قبلی":"Previous image"}>{rtl?<ChevronRight size={21}/>:<ChevronLeft size={21}/>}</button>
            <button type="button" className="gallery-arrow gallery-arrow-next" onClick={nextImage} aria-label={rtl?"تصویر بعدی":"Next image"}>{rtl?<ChevronLeft size={21}/>:<ChevronRight size={21}/>}</button>
            <span className="gallery-counter" aria-live="polite">{activeImage+1} / {images.length}</span>
          </>}
        </div>
        {images.length>1&&<div className="detail-thumbnails" aria-label={rtl?"تمام تصاویر محصول":"All product images"}>{images.map((src,index)=><button type="button" key={src} className={activeImage===index?"active":""} onClick={()=>setActiveImage(index)} aria-label={rtl?"تصویر "+String(index+1):"Image "+String(index+1)} aria-pressed={activeImage===index}><img src={catalogImageSrc(src)} alt="" loading="lazy" decoding="async" fetchPriority="low" width={96} height={96}/></button>)}</div>}
      </div>
      <div className="detail-information">
        <span className="detail-brand">{product.brand||"GLOBALCART"}</span>
        <h1>{product.title}</h1>
        <div className="detail-rating">{product.ratingAvailable===false?<span>{rtl?"امتیاز ثبت نشده":"No rating supplied"}</span>:<span><Star size={15} fill="currentColor"/>{product.rating.toFixed(1)}</span>}<i/> <span>{rtl?"شناسه محصول":"Product ID"}: {product.id}</span><i/><span>{rtl?category.fa:category.en}</span></div>
        <p className="detail-description">{product.description}</p>
        <div className="detail-highlights"><div><ShieldCheck/><span><b>{rtl?"اطلاعات شفاف":"Clear information"}</b><small>{rtl?"قیمت این محصول تخمینی است و قیمت فروشنده نیست.":"This product price is estimated, not a retailer quote."}</small></span></div><div><Truck/><span><b>{rtl?"تجربه خرید نمایشی":"Demo shopping experience"}</b><small>{rtl?"بدون ارسال یا موجودی واقعی":"No real shipping or inventory"}</small></span></div></div>
        <div className="detail-specs"><div><span>{rtl?"دسته‌بندی":"Category"}</span><b>{rtl?category.fa:category.en}</b></div><div><span>{rtl?"برند":"Brand"}</span><b>{product.brand||"Global Select"}</b></div><div><span>{rtl?"امتیاز نمونه":"Sample rating"}</span><b>{product.ratingAvailable===false?(rtl?"ثبت نشده":"Not provided"):product.rating.toFixed(1)+" / 5"}</b></div><div><span>{rtl?"وضعیت موجودی":"Availability"}</span><b>{product.stockAvailable===false?(rtl?"ثبت نشده":"Not provided"):product.stock}</b></div></div>
      </div>
      <aside className="detail-buybox">
        <span className="buybox-kicker">{rtl?"قیمت تخمینی · نمایشی":"ESTIMATED · DEMO PRICE"}</span>
        <div className="detail-price">{product.priceAvailable===false?(rtl?"قیمت ثبت نشده":"Price not listed"):priceBreakdown?(rtl?new Intl.NumberFormat("fa-IR",{maximumFractionDigits:0}).format(priceBreakdown.toman)+" تومان":"$"+priceBreakdown.discountedUsd.toFixed(2)):"$"+discounted.toFixed(2)}</div>
        {product.priceAvailable===false
          ? <p className="price-breakdown">{rtl?"قیمت این کالا هنوز در کاتالوگ ثبت نشده است.":"A price has not been listed for this product yet."}</p>
          : <><div className="detail-old-price"><s>{format(product.price)}</s><span>{rtl?"قیمت قبل از تخفیف":"Before discount"}</span></div>
        <details className="price-breakdown"><summary>{rtl?"منبع و جزئیات نرخ تبدیل":"FX source and calculation"}</summary>
          <div><span>{rtl?"مبنای تخمینی (USD)":"Estimated base price (USD)"}</span><b>${product.price.toFixed(2)}</b></div>
          <div><span>{rtl?"تخفیف":"Discount"}</span><b>{product.discountPercentage.toFixed(2)}%</b></div>
          {priceBreakdown?<>
            <div><span>{rtl?"نرخ دلار آزاد TGJU (ریال به تومان)":"TGJU free-market USD rate (Rial to Toman)"}</span><b>{new Intl.NumberFormat(rtl?"fa-IR":"en-US",{maximumFractionDigits:0}).format(priceBreakdown.fxRate)} {rtl?"تومان / دلار":"Toman / USD"}</b></div>
            <div><span>{rtl?"نشانگر به‌روزرسانی TGJU":"TGJU source update marker"}</span><b>{fxTimestampLabel??fxStatusLabel}</b></div>
            <small>{rtl?"نرخ تبدیل دلار از دادهٔ TGJU گرفته می‌شود؛ خود قیمت کالا تخمینی و غیرقابل استناد برای خرید واقعی است.":"The USD conversion rate comes from TGJU; the product price itself is an estimate and is not a real purchase quote."}</small>
            {fxStatus!=="fresh"&&<small className="fx-warning">{rtl?"هشدار تازگی نرخ: "+fxStatusLabel:"FX freshness warning: "+fxStatusLabel}</small>}
          </>:<small>{rtl?"نرخ معتبر دلار TGJU فعلاً در دسترس نیست؛ مبلغ اصلی به دلار نمایش داده می‌شود.":"A valid TGJU dollar quote is currently unavailable, so the source USD price is shown."}</small>}
          <small>{rtl?"قیمت‌ها نمایشی‌اند و شامل هزینه ارسال، مالیات، کارمزد پرداخت یا تعهد خرید نیستند.":"Demo prices only; no shipping, tax, payment fees or purchase commitment are included."}</small>
        </details></>}
        <div className="buybox-stock"><span/> {product.stockAvailable===false?(rtl?"موجودی ثبت نشده":"Stock not listed"):(rtl?"موجود در کاتالوگ نمونه":"Listed in sample catalog")}</div>
        <div className="buybox-quantity"><span>{rtl?"تعداد":"Quantity"}</span><div><button onClick={()=>setQuantity(q=>Math.max(1,q-1))} aria-label={rtl?"کاهش تعداد":"Decrease quantity"}><Minus size={15}/></button><b>{quantity}</b><button onClick={()=>setQuantity(q=>Math.min(10,q+1))} aria-label={rtl?"افزایش تعداد":"Increase quantity"}><Plus size={15}/></button></div></div>
        <button className="buybox-add" onClick={addToCart}><ShoppingBag size={17}/>{rtl?"افزودن به سبد خرید":"Add to bag"}</button>
        <button className="buybox-save" onClick={toggleSaved}><Heart size={16} fill={saved?"currentColor":"none"}/>{saved?(rtl?"حذف از علاقه‌مندی‌ها":"Remove from saved"):(rtl?"ذخیره در علاقه‌مندی‌ها":"Save for later")}</button>
        {added&&<div className="buybox-confirm"><Check size={16}/>{rtl?"به سبد خرید اضافه شد":"Added to your bag"}</div>}
        <div className="buybox-disclaimer"><ShieldCheck size={16}/><span>{rtl?"این یک فروشگاه نمایشی است؛ پرداخت و ارسال واقعی انجام نمی‌شود.":"Demo storefront only. No real payment or shipping."}</span></div>
      </aside>
    </section>
    <section className="detail-description-panel"><span className="section-kicker">PRODUCT OVERVIEW</span><h2>{rtl?"درباره محصول":"About this product"}</h2><p>{product.description}</p><div><span>{rtl?"شناسه محصول":"Product ID"}</span><b>GC-{product.id}</b><span>{rtl?"قیمت":"Price"}</span><b>{format(discounted)}</b></div></section>
    {related.length>0&&<section className="detail-related"><div><span className="section-kicker">YOU MAY ALSO LIKE</span><h2>{rtl?"محصولات مشابه":"Related products"}</h2></div><div className="related-grid">{related.map(item=><Link href={`/product/${item.id}`} key={item.id}><img src={catalogImageSrc(item.thumbnail)} alt={item.title} loading="lazy" decoding="async"/><span>{item.brand||"GLOBAL SELECT"}</span><b>{item.title}</b><strong>{item.priceAvailable===false?(rtl?"قیمت ثبت نشده":"Price not listed"):format(item.price*(1-item.discountPercentage/100))}</strong></Link>)}</div></section>}
    <footer className="product-page-footer"><Link href="/">globalcart</Link><span>{rtl?"پروژه مستقل نمایشی · توسعه و طراحی پویا گودرزی":"Independent portfolio demo · Designed & built by Pooya Goodarzi"}</span></footer>
  </main>;
}
