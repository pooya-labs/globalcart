"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { ArrowLeft, Check, ChevronLeft, Heart, Minus, Plus, ShieldCheck, ShoppingBag, Star, Truck } from "lucide-react";

type Product = {
  id:number; title:string; description:string; category:string; price:number;
  discountPercentage:number; rating:number; stock:number; brand?:string;
  thumbnail:string; images?:string[]; source?:string; sourceUrl?:string;
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
  "mens-watches":{fa:"ساعت مردانه",en:"Men's watches"}
};

export default function ProductDetailClient({product,related}:Props) {
  const [locale,setLocale]=useState<"fa"|"en">("fa");
  const [activeImage,setActiveImage]=useState(0);
  const [quantity,setQuantity]=useState(1);
  const [added,setAdded]=useState(false);
  const [saved,setSaved]=useState(false);
  const [rate,setRate]=useState(69000);
  const rtl=locale==="fa";
  const images=useMemo(()=>Array.from(new Set([product.thumbnail,...(product.images||[])].filter(Boolean))),[product]);
  const image=images[Math.min(activeImage,images.length-1)]||product.thumbnail;
  const discounted=product.price*(1-product.discountPercentage/100);
  useEffect(()=>{
    try{
      const raw=localStorage.getItem("globalcart-state-v1");
      if(!raw)return;
      const state=JSON.parse(raw);
      if(Number.isFinite(state.rate)&&state.rate>0)setRate(state.rate);
      if(Array.isArray(state.favorites))setSaved(state.favorites.includes(product.id));
    }catch{/* Ignore invalid local demo state. */}
  },[product.id]);
  const format=(value:number)=>rtl
    ? new Intl.NumberFormat("fa-IR",{maximumFractionDigits:0}).format(value*rate)+" تومان"
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
        <div className="detail-main-image"><img src={image} alt={product.title}/><span className="detail-discount">{Math.round(product.discountPercentage)}% {rtl?"تخفیف":"OFF"}</span></div>
        <div className="detail-thumbnails">{images.map((src,index)=><button key={src} className={activeImage===index?"active":""} onClick={()=>setActiveImage(index)} aria-label={rtl?`تصویر ${index+1}`:`Image ${index+1}`}><img src={src} alt=""/></button>)}</div>
      </div>
      <div className="detail-information">
        <span className="detail-brand">{product.brand||"GLOBAL SELECT"}</span>
        <h1>{product.title}</h1>
        <div className="detail-rating"><span><Star size={15} fill="currentColor"/>{product.rating.toFixed(1)}</span><i/> <span>{rtl?"شناسه محصول":"Product ID"}: {product.id}</span><i/><span>{rtl?category.fa:category.en}</span></div>
        <p className="detail-description">{product.description}</p>
        <div className="detail-highlights"><div><ShieldCheck/><span><b>{rtl?"اطلاعات شفاف":"Clear information"}</b><small>{rtl?"قیمت و منبع داده مشخص است":"Price and data source disclosed"}</small></span></div><div><Truck/><span><b>{rtl?"تجربه خرید نمایشی":"Demo shopping experience"}</b><small>{rtl?"بدون ارسال یا موجودی واقعی":"No real shipping or inventory"}</small></span></div></div>
        <div className="detail-specs"><div><span>{rtl?"دسته‌بندی":"Category"}</span><b>{rtl?category.fa:category.en}</b></div><div><span>{rtl?"برند":"Brand"}</span><b>{product.brand||"Global Select"}</b></div><div><span>{rtl?"امتیاز نمونه":"Sample rating"}</span><b>{product.rating.toFixed(1)} / 5</b></div><div><span>{rtl?"موجودی در منبع نمونه":"Source stock figure"}</span><b>{product.stock}</b></div></div>
      </div>
      <aside className="detail-buybox">
        <span className="buybox-kicker">{rtl?"قیمت نمایشی":"ILLUSTRATIVE PRICE"}</span>
        <div className="detail-price">{format(discounted)}</div>
        <div className="detail-old-price"><s>{format(product.price)}</s><span>{rtl?"قیمت قبل از تخفیف":"Before discount"}</span></div>
        <div className="buybox-stock"><span/> {rtl?"موجود در کاتالوگ نمونه":"Listed in sample catalog"}</div>
        <div className="buybox-quantity"><span>{rtl?"تعداد":"Quantity"}</span><div><button onClick={()=>setQuantity(q=>Math.max(1,q-1))} aria-label="Decrease quantity"><Minus size={15}/></button><b>{quantity}</b><button onClick={()=>setQuantity(q=>Math.min(10,q+1))} aria-label="Increase quantity"><Plus size={15}/></button></div></div>
        <button className="buybox-add" onClick={addToCart}><ShoppingBag size={17}/>{rtl?"افزودن به سبد خرید":"Add to bag"}</button>
        <button className="buybox-save" onClick={toggleSaved}><Heart size={16} fill={saved?"currentColor":"none"}/>{saved?(rtl?"حذف از علاقه‌مندی‌ها":"Remove from saved"):(rtl?"ذخیره در علاقه‌مندی‌ها":"Save for later")}</button>
        {added&&<div className="buybox-confirm"><Check size={16}/>{rtl?"به سبد خرید اضافه شد":"Added to your bag"}</div>}
        <div className="buybox-disclaimer"><ShieldCheck size={16}/><span>{rtl?"این یک فروشگاه نمایشی است؛ پرداخت و ارسال واقعی انجام نمی‌شود.":"Demo storefront only. No real payment or shipping."}</span></div>
        <small className="buybox-source">{rtl?"منبع اطلاعات":"Data source"}: {product.source||"DummyJSON public demo API"}</small>
      </aside>
    </section>
    <section className="detail-description-panel"><span className="section-kicker">PRODUCT OVERVIEW</span><h2>{rtl?"درباره محصول":"About this product"}</h2><p>{product.description}</p><div><span>{rtl?"شناسه منبع":"Source ID"}</span><b>{product.id}</b><span>{rtl?"آخرین قیمت نمونه":"Sample price"}</span><b>{format(discounted)}</b></div></section>
    {related.length>0&&<section className="detail-related"><div><span className="section-kicker">YOU MAY ALSO LIKE</span><h2>{rtl?"محصولات مشابه":"Related products"}</h2></div><div className="related-grid">{related.map(item=><Link href={`/product/${item.id}`} key={item.id}><img src={item.thumbnail} alt={item.title}/><span>{item.brand||"GLOBAL SELECT"}</span><b>{item.title}</b><strong>{format(item.price*(1-item.discountPercentage/100))}</strong></Link>)}</div></section>}
    <footer className="product-page-footer"><Link href="/">globalcart</Link><span>{rtl?"پروژه مستقل نمایشی · توسعه و طراحی پویا گودرزی":"Independent portfolio demo · Designed & built by Pooya Goodarzi"}</span></footer>
  </main>;
}
