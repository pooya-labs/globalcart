"use client";

import { useEffect, useMemo, useState } from "react";
import {
  ArrowDown, ArrowLeft, ArrowRight, ArrowUpDown, Check, ChevronDown, ChevronLeft,
  ChevronRight, CircleHelp, Heart, LayoutGrid, Menu, Minus, Package, Plus,
  Search, ShieldCheck, ShoppingBag, SlidersHorizontal, Sparkles, Star, Truck,
  X, Zap
} from "lucide-react";

type Product = {
  id: number; title: string; description: string; category: string; price: number;
  discountPercentage: number; rating: number; stock: number; brand?: string;
  thumbnail: string; images?: string[]; source?: string;
};
type CartLine = { product: Product; quantity: number };
type Locale = "fa" | "en";

const copy = {
  fa: {
    top: "تجربه‌ای تازه برای کشف چیزهای خوب", search: "جست‌وجوی محصول، برند یا دسته‌بندی...",
    nav1: "فروشگاه", nav2: "دسته‌بندی‌ها", nav3: "پیشنهادهای منتخب", nav4: "درباره پروژه",
    heroEyebrow: "THE ART OF DISCOVERY", heroTitle: "چیزهای خوب،\nدرست همین‌جا پیدا می‌شوند.", heroDesc: "یک تجربه خرید مدرن برای کشف، مقایسه و انتخاب هوشمندانه؛ از محصولات روزمره تا انتخاب‌های خاص.",
    explore: "کشف محصولات", curated: "انتخاب‌های سردبیر", categories: "بر اساس علاقه بگرد", categoriesDesc: "از جایی شروع کن که به تو نزدیک‌تر است.",
    all: "همه محصولات", catalog: "ویترین جهانی", catalogDesc: "کالکشن نمونه با اطلاعات واردشده از منبع داده نمایشی.",
    sort: "مرتب‌سازی", popular: "محبوب‌ترین", low: "قیمت: کم به زیاد", high: "قیمت: زیاد به کم", rating: "بالاترین امتیاز",
    filter: "فیلترها", category: "دسته‌بندی", clear: "پاک کردن فیلتر", noResults: "محصولی با این مشخصات پیدا نشد.",
    add: "افزودن به سبد", quick: "جزئیات", cart: "سبد خرید", favorites: "علاقه‌مندی‌ها", emptyCart: "سبدت هنوز خالی است.",
    subtotal: "جمع جزء", checkout: "ادامه و ثبت سفارش نمایشی", demo: "نسخه نمایشی", noPayment: "پرداخت واقعی غیرفعال است؛ این پروژه فروشگاه واقعی نیست.",
    demoCheckout: "تکمیل سفارش آزمایشی", demoText: "این مرحله فقط برای نمایش تجربه کاربری است. هیچ پرداخت یا سفارش واقعی ثبت نمی‌شود.",
    confirm: "ثبت سفارش آزمایشی", success: "سفارش نمایشی ساخته شد", order: "شناسه سفارش", continue: "بازگشت به فروشگاه",
    source: "منبع", rate: "نرخ نمایشی هر دلار (تومان)", inStock: "موجود در کاتالوگ نمونه", off: "تخفیف",
    searchResults: "نتایج جست‌وجو", loading: "در حال دریافت کاتالوگ...", sourceFallback: "منبع آنلاین پاسخ نداد؛ محصولات نمونه محلی نمایش داده شده‌اند.",
    why: "چرا GlobalCart؟", whyText: "یک پروژه مستقل برای نمایش مهندسی محصول، معماری فرانت‌اند، یکپارچه‌سازی داده و طراحی تجربه تجارت الکترونیک.",
    feature1: "کشف هوشمند", feature1d: "جست‌وجو و فیلتر سریع", feature2: "اطلاعات شفاف", feature2d: "نمایش منبع و قیمت", feature3: "تجربه امن", feature3d: "بدون تراکنش واقعی",
    footer: "یک تجربه تجارت الکترونیک مستقل، ساخته‌شده برای نمایش مهندسی محصول.", owner: "طراحی و توسعه: پویا گودرزی",
    count: "محصول", menu: "منو", compare: "مقایسه", recently: "تازه بازدیدشده", emptyFav: "هنوز محصولی ذخیره نکرده‌ای.",
    min: "کمترین قیمت", max: "بیشترین قیمت", result: "نتیجه", close: "بستن"
  },
  en: {
    top: "A fresh perspective on finding good things", search: "Search products, brands or categories...",
    nav1: "Shop", nav2: "Categories", nav3: "Editor's picks", nav4: "About",
    heroEyebrow: "THE ART OF DISCOVERY", heroTitle: "Good things,\nright where you need them.", heroDesc: "A modern shopping experience to discover, compare and choose with confidence — from everyday essentials to thoughtful finds.",
    explore: "Explore products", curated: "Editor's selection", categories: "Browse by interest", categoriesDesc: "Start with what speaks to you.",
    all: "All products", catalog: "The global edit", catalogDesc: "A sample collection imported from a public demo data source.",
    sort: "Sort by", popular: "Most popular", low: "Price: low to high", high: "Price: high to low", rating: "Top rated",
    filter: "Filters", category: "Category", clear: "Clear filters", noResults: "No products match your search.",
    add: "Add to bag", quick: "Details", cart: "Your bag", favorites: "Saved items", emptyCart: "Your bag is still empty.",
    subtotal: "Subtotal", checkout: "Continue to demo checkout", demo: "DEMO", noPayment: "Real payments are disabled. This is not a live store.",
    demoCheckout: "Demo checkout", demoText: "This step demonstrates the user experience only. No real order or payment will be created.",
    confirm: "Place demo order", success: "Demo order created", order: "Order reference", continue: "Back to store",
    source: "Source", rate: "Illustrative USD rate (Toman)", inStock: "Sample catalog item", off: "OFF",
    searchResults: "Search results", loading: "Loading catalog...", sourceFallback: "Live source did not respond; showing the local sample catalog.",
    why: "Why GlobalCart?", whyText: "An independent portfolio project showcasing product engineering, frontend architecture, data integration and commerce experience design.",
    feature1: "Smart discovery", feature1d: "Fast search and filters", feature2: "Clear information", feature2d: "Source and price context", feature3: "Safe experience", feature3d: "No real transactions",
    footer: "An independent commerce experience, built to showcase product engineering.", owner: "Designed & engineered by Pooya Goodarzi",
    count: "products", menu: "Menu", compare: "Compare", recently: "Recently viewed", emptyFav: "You have not saved anything yet.",
    min: "Min price", max: "Max price", result: "Results", close: "Close"
  }
} as const;

const fallback: Product[] = [
  {id:101,title:"هدفون بی‌سیم استودیویی",description:"صدای شفاف با حذف نویز فعال، طراحی سبک و شارژدهی طولانی برای استفاده روزمره.",category:"electronics",price:89.99,discountPercentage:18,rating:4.8,stock:24,brand:"SONORA",thumbnail:"https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=900",images:[],source:"Local demo fixture"},
  {id:102,title:"ساعت هوشمند Active",description:"نمایشگر AMOLED، پایش فعالیت و طراحی مینیمال با بند راحت.",category:"mens-watches",price:129.5,discountPercentage:12,rating:4.7,stock:18,brand:"NORTH",thumbnail:"https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=900",images:[],source:"Local demo fixture"},
  {id:103,title:"دوربین بدون‌آینه",description:"ثبت جزئیات دقیق با حسگر پیشرفته و لنز همه‌کاره برای خلاقیت روزانه.",category:"mobile-accessories",price:549,discountPercentage:9,rating:4.9,stock:8,brand:"FRAME",thumbnail:"https://images.unsplash.com/photo-1516035069371-29a1b244cc32?w=900",images:[],source:"Local demo fixture"},
  {id:104,title:"کیبورد مکانیکی بی‌سیم",description:"کلیدهای دقیق، اتصال چندگانه و بدنه جمع‌وجور برای میز کار حرفه‌ای.",category:"keyboards",price:109,discountPercentage:15,rating:4.6,stock:32,brand:"KEYFORM",thumbnail:"https://images.unsplash.com/photo-1587829741301-dc798b83add3?w=900",images:[],source:"Local demo fixture"},
  {id:105,title:"کوله‌پشتی شهری",description:"فضای محافظ لپ‌تاپ، پارچه مقاوم و فرم ساده برای رفت‌وآمد روزانه.",category:"mens-bags",price:74,discountPercentage:20,rating:4.5,stock:41,brand:"FIELD",thumbnail:"https://images.unsplash.com/photo-1553062407-98eeb64c6a62?w=900",images:[],source:"Local demo fixture"},
  {id:106,title:"اسپیکر قابل حمل",description:"صدای پرقدرت در بدنه‌ای مقاوم و قابل حمل؛ مناسب خانه و سفر.",category:"mobile-accessories",price:64.9,discountPercentage:11,rating:4.4,stock:15,brand:"SONORA",thumbnail:"https://images.unsplash.com/photo-1608043152269-423dbba4e7e1?w=900",images:[],source:"Local demo fixture"}
];

const categories = [
  {id:"all",fa:"همه",en:"Everything",symbol:"✳"},
  {id:"beauty",fa:"زیبایی و مراقبت",en:"Beauty",symbol:"✦"},
  {id:"fragrances",fa:"عطر و رایحه",en:"Fragrance",symbol:"◌"},
  {id:"furniture",fa:"خانه و دکور",en:"Home",symbol:"⌂"},
  {id:"groceries",fa:"خوراکی",en:"Grocery",symbol:"◒"},
  {id:"laptops",fa:"لپ‌تاپ",en:"Laptops",symbol:"▣"},
  {id:"smartphones",fa:"موبایل",en:"Phones",symbol:"▯"},
  {id:"mens-watches",fa:"ساعت",en:"Watches",symbol:"◷"},
  {id:"electronics",fa:"الکترونیک",en:"Electronics",symbol:"⌘"},
  {id:"mobile-accessories",fa:"لوازم جانبی",en:"Accessories",symbol:"⌁"}
];

const money = (value:number, locale:Locale) => new Intl.NumberFormat(locale==="fa"?"fa-IR":"en-US",{maximumFractionDigits:0}).format(value);
const usd = (value:number) => "$"+value.toFixed(2);

export default function Home() {
  const [locale,setLocale]=useState<Locale>("fa");
  const [products,setProducts]=useState<Product[]>(fallback);
  const [loading,setLoading]=useState(true);
  const [usingFallback,setUsingFallback]=useState(false);
  const [query,setQuery]=useState("");
  const [category,setCategory]=useState("all");
  const [sort,setSort]=useState("popular");
  const [cart,setCart]=useState<CartLine[]>([]);
  const [favorites,setFavorites]=useState<number[]>([]);
  const [recent,setRecent]=useState<number[]>([]);
  const [cartOpen,setCartOpen]=useState(false);
  const [favoriteOpen,setFavoriteOpen]=useState(false);
  const [quick,setQuick]=useState<Product|null>(null);
  const [checkout,setCheckout]=useState(false);
  const [order,setOrder]=useState("");
  const [mobileMenu,setMobileMenu]=useState(false);
  const [toast,setToast]=useState("");
  const [rate,setRate]=useState(69000);
  const [hydrated,setHydrated]=useState(false);
  const [minPrice,setMinPrice]=useState(0);
  const [maxPrice,setMaxPrice]=useState(1000);
  const [showFilters,setShowFilters]=useState(false);
  const t=copy[locale];
  const rtl=locale==="fa";

  useEffect(()=>{
    let active=true;
    fetch("https://dummyjson.com/products?limit=100")
      .then(r=>{if(!r.ok)throw new Error("Catalog request failed");return r.json();})
      .then(data=>{
        if(!active||!Array.isArray(data.products))return;
        const mapped:Product[]=data.products.map((p:any)=>({
          id:p.id,title:p.title,description:p.description,category:p.category,
          price:Number(p.price)||0,discountPercentage:Number(p.discountPercentage)||0,
          rating:Number(p.rating)||0,stock:Number(p.stock)||0,brand:p.brand||"GLOBAL",
          thumbnail:p.thumbnail,images:Array.isArray(p.images)?p.images:[],source:"DummyJSON public demo API"
        }));
        if(mapped.length){setProducts(mapped);setUsingFallback(false);}
      })
      .catch(()=>{if(active){setProducts(fallback);setUsingFallback(true);}})
      .finally(()=>{if(active)setLoading(false);});
    return()=>{active=false;};
  },[]);

  useEffect(()=>{
    try{
      const raw=localStorage.getItem("globalcart-state-v1");
      if(!raw)return;
      const saved=JSON.parse(raw);
      if(Array.isArray(saved.cart))setCart(saved.cart);
      if(Array.isArray(saved.favorites))setFavorites(saved.favorites);
      if(Array.isArray(saved.recent))setRecent(saved.recent);
      if(Number.isFinite(saved.rate)&&saved.rate>0)setRate(saved.rate);
    }catch{/* Ignore invalid local demo state. */}
    finally{setHydrated(true);}
  },[]);

  useEffect(()=>{
    if(!hydrated)return;
    try{localStorage.setItem("globalcart-state-v1",JSON.stringify({cart,favorites,recent,rate}));}catch{/* Storage can be unavailable. */}
  },[hydrated,cart,favorites,recent,rate]);

  const filtered=useMemo(()=>{
    let list=products.filter(p=>{
      const text=(p.title+" "+p.description+" "+(p.brand||"")+" "+p.category).toLocaleLowerCase();
      return (category==="all"||p.category===category)&&text.includes(query.toLocaleLowerCase())&&p.price>=minPrice&&p.price<=maxPrice;
    });
    if(sort==="low")list.sort((a,b)=>a.price-b.price);
    else if(sort==="high")list.sort((a,b)=>b.price-a.price);
    else if(sort==="rating")list.sort((a,b)=>b.rating-a.rating);
    else list.sort((a,b)=>(b.rating*b.discountPercentage)-(a.rating*a.discountPercentage));
    return list;
  },[products,category,query,sort,minPrice,maxPrice]);

  const count=cart.reduce((sum,line)=>sum+line.quantity,0);
  const subtotal=cart.reduce((sum,line)=>sum+line.product.price*rate*line.quantity,0);
  const formatPrice=(p:Product)=>rtl?money(p.price*(1-p.discountPercentage/100)*rate,locale)+" تومان":usd(p.price*(1-p.discountPercentage/100));
  const notify=(message:string)=>{setToast(message);window.setTimeout(()=>setToast(""),2100);};
  const addToCart=(product:Product)=>{
    setCart(old=>{const found=old.find(x=>x.product.id===product.id);return found?old.map(x=>x.product.id===product.id?{...x,quantity:x.quantity+1}:x):[...old,{product,quantity:1}];});
    notify(locale==="fa"?"به سبد خرید اضافه شد":"Added to your bag");
  };
  const changeQuantity=(id:number,delta:number)=>setCart(old=>old.map(x=>x.product.id===id?{...x,quantity:x.quantity+delta}:x).filter(x=>x.quantity>0));
  const openProduct=(p:Product)=>{setQuick(p);setRecent(old=>[p.id,...old.filter(id=>id!==p.id)].slice(0,8));};
  const toggleFavorite=(id:number)=>setFavorites(old=>old.includes(id)?old.filter(x=>x!==id):[...old,id]);
  const placeDemoOrder=()=>{setOrder("GC-"+Math.random().toString(36).slice(2,8).toUpperCase());setCart([]);setCheckout(false);};
  const favoriteProducts=products.filter(p=>favorites.includes(p.id));

  return (
    <main className={rtl?"store rtl":"store ltr"} dir={rtl?"rtl":"ltr"}>
      <div className="topline"><span>{t.top}</span><span className="topline-right"><ShieldCheck size={13}/>{t.noPayment}</span></div>
      <header className="header">
        <button className="mobile-toggle icon-button" onClick={()=>setMobileMenu(!mobileMenu)} aria-label={t.menu}><Menu size={21}/></button>
        <a href="#" className="brand"><span className="brand-mark">g<span>.</span></span><span className="brand-name">globalcart<small>THE WORLD, WELL FOUND.</small></span></a>
        <nav className={mobileMenu?"main-nav nav-open":"main-nav"}>
          <a href="#catalog" onClick={()=>setMobileMenu(false)}>{t.nav1}</a>
          <a href="#categories" onClick={()=>setMobileMenu(false)}>{t.nav2}</a>
          <a href="#editors" onClick={()=>setMobileMenu(false)}>{t.nav3}</a>
          <a href="#about" onClick={()=>setMobileMenu(false)}>{t.nav4}</a>
        </nav>
        <div className="header-actions">
          <label className="search-box"><Search size={17}/><input value={query} onChange={e=>setQuery(e.target.value)} placeholder={t.search}/></label>
          <button className="locale-button" onClick={()=>setLocale(rtl?"en":"fa")}>{rtl?"EN":"فا"}</button>
          <button className="icon-button favorite-trigger" onClick={()=>setFavoriteOpen(true)} aria-label={t.favorites}><Heart size={19}/><span>{favorites.length}</span></button>
          <button className="bag-button" onClick={()=>setCartOpen(true)}><ShoppingBag size={17}/><span>{t.cart}</span><b>{count}</b></button>
        </div>
      </header>

      <section className="hero">
        <div className="hero-copy">
          <div className="eyebrow"><span></span>{t.heroEyebrow}</div>
          <h1>{t.heroTitle.split("\n").map((line,i)=><span key={i}>{line}</span>)}</h1>
          <p>{t.heroDesc}</p>
          <a className="button-primary" href="#catalog">{t.explore}<ArrowLeft size={16}/></a>
          <div className="hero-meta"><span>01</span><i></i><span>DISCOVER&nbsp; / &nbsp;COMPARE&nbsp; / &nbsp;CHOOSE</span></div>
        </div>
        <div className="hero-visual">
          <img src="https://images.unsplash.com/photo-1490312278390-ab64016e0aa9?w=1500&q=88" alt="A considered collection of lifestyle objects"/>
          <div className="hero-image-caption"><small>GLOBAL EDITION / 001</small><b>Objects with a point of view.</b><span>CURATED FOR EVERYDAY LIFE</span></div>
          <div className="hero-index">GC<span>®</span></div>
          <div className="hero-side-note">A BETTER WAY TO BROWSE</div>
        </div>
      </section>

      <section className="benefits">
        <div><span className="benefit-icon"><Sparkles size={18}/></span><span><b>{t.feature1}</b><small>{t.feature1d}</small></span></div>
        <div><span className="benefit-icon"><ShieldCheck size={18}/></span><span><b>{t.feature2}</b><small>{t.feature2d}</small></span></div>
        <div><span className="benefit-icon"><Truck size={18}/></span><span><b>{t.feature3}</b><small>{t.feature3d}</small></span></div>
        <label className="fx-control"><span>{t.rate}</span><input type="number" min="1" value={rate} onChange={e=>setRate(Math.max(1,Number(e.target.value)||1))}/></label>
      </section>

      <section className="section categories-section" id="categories">
        <div className="section-heading"><div><span className="section-kicker">01 / FIND YOUR INTEREST</span><h2>{t.categories}</h2><p>{t.categoriesDesc}</p></div><a href="#catalog">{t.all}<ArrowLeft size={15}/></a></div>
        <div className="category-strip">
          {categories.slice(1,9).map((c,i)=><button key={c.id} className={"category-tile tile-"+i} onClick={()=>{setCategory(c.id);document.getElementById("catalog")?.scrollIntoView({behavior:"smooth"});}}><span className="category-symbol">{c.symbol}</span><span>{rtl?c.fa:c.en}</span><ArrowLeft size={14}/></button>)}
        </div>
      </section>

      <section className="section catalog-section" id="catalog">
        <div className="section-heading catalog-heading"><div><span className="section-kicker">02 / THE COLLECTION</span><h2>{query?t.searchResults:t.catalog}</h2><p>{t.catalogDesc}</p></div><div className="catalog-tools"><button className="filter-toggle" onClick={()=>setShowFilters(!showFilters)}><SlidersHorizontal size={16}/>{t.filter}</button><label className="sort-select"><ArrowUpDown size={15}/><select value={sort} onChange={e=>setSort(e.target.value)}><option value="popular">{t.popular}</option><option value="low">{t.low}</option><option value="high">{t.high}</option><option value="rating">{t.rating}</option></select><ChevronDown size={14}/></label></div></div>
        <div className="catalog-layout">
          <aside className={showFilters?"filter-panel filter-visible":"filter-panel"}>
            <div className="filter-title"><b><SlidersHorizontal size={16}/>{t.filter}</b><button onClick={()=>{setCategory("all");setMinPrice(0);setMaxPrice(1000);setQuery("");}}>{t.clear}</button></div>
            <div className="filter-group"><b>{t.category}</b>{categories.map(c=><button key={c.id} className={category===c.id?"filter-option active":"filter-option"} onClick={()=>setCategory(c.id)}>{rtl?c.fa:c.en}<span>{c.id==="all"?products.length:products.filter(p=>p.category===c.id).length}</span></button>)}</div>
            <div className="filter-group price-filter"><b>{rtl?"بازه قیمت (دلار)": "Price range (USD)"}</b><label>{t.min}<input type="number" min="0" value={minPrice} onChange={e=>setMinPrice(Math.max(0,Number(e.target.value)||0))}/></label><label>{t.max}<input type="number" min="0" value={maxPrice} onChange={e=>setMaxPrice(Math.max(0,Number(e.target.value)||0))}/></label></div>
            <div className="filter-note"><InfoIcon/>{rtl?"داده‌ها نمایشی هستند و قیمت‌ها تعهد خرید نیستند.":"Demo data only. Prices are not purchase offers."}</div>
          </aside>
          <div className="product-area">
            {(loading||usingFallback)&&<div className="catalog-status">{loading?<><span className="loader"></span>{t.loading}</>:<><InfoIcon/>{t.sourceFallback}</>}</div>}
            {order&&<div className="order-success"><Check size={17}/>{t.success} · {t.order}: {order}<button onClick={()=>setOrder("")}><X size={15}/></button></div>}
            {filtered.length===0?<div className="no-results"><Search size={30}/><b>{t.noResults}</b><button onClick={()=>{setQuery("");setCategory("all");setMinPrice(0);setMaxPrice(1000);}}>{t.clear}</button></div>:<div className="product-grid">
              {filtered.map((p,index)=><article className="product-card" key={p.id}>
                <div className="product-media" onClick={()=>openProduct(p)}><img src={p.thumbnail} alt={p.title} loading="lazy"/><span className="product-number">GC / {String(index+1).padStart(2,"0")}</span>{p.discountPercentage>0&&<span className="discount-tag">{Math.round(p.discountPercentage)}% {t.off}</span>}<button className={favorites.includes(p.id)?"save-button saved":"save-button"} aria-label={t.favorites} onClick={e=>{e.stopPropagation();toggleFavorite(p.id);}}><Heart size={17} fill={favorites.includes(p.id)?"currentColor":"none"}/></button><button className="quick-view" onClick={e=>{e.stopPropagation();openProduct(p);}}>{t.quick}<ArrowUpLeftIcon/></button></div>
                <div className="product-details"><span className="product-brand">{p.brand||"GLOBAL SELECT"}</span><h3 onClick={()=>openProduct(p)}>{p.title}</h3><div className="product-rating"><span><Star size={12} fill="currentColor"/>{p.rating.toFixed(1)}</span><small>{t.inStock}</small></div><div className="product-price"><div><b>{formatPrice(p)}</b><small>{usd(p.price)}</small></div><button onClick={()=>addToCart(p)} aria-label={t.add}><Plus size={19}/></button></div></div>
              </article>)}
            </div>}
            <div className="catalog-foot"><span>{t.result}: {filtered.length} / {products.length}</span><span>{t.source}: {usingFallback?"Local demo fixture":"DummyJSON public demo API"}</span></div>
          </div>
        </div>
      </section>

      <section className="editorial" id="editors"><div className="editorial-image"><img src="https://images.unsplash.com/photo-1494438639946-1ebd1d20bf85?w=1500&q=85" alt="Quiet modern interior"/></div><div className="editorial-copy"><span className="section-kicker">03 / A MORE CONSIDERED EXPERIENCE</span><h2>{rtl?"کمتر بگرد.\nبهتر انتخاب کن.":"Less scrolling.\nMore considered choices."}</h2><p>{t.whyText}</p><a href="#catalog" className="button-primary">{t.explore}<ArrowLeft size={16}/></a></div><span className="editorial-index">03</span></section>

      <section className="about-section" id="about"><div><span className="section-kicker">04 / BUILT AS A PORTFOLIO PROJECT</span><h2>{t.why}</h2><p>{t.whyText}</p></div><div className="about-stats"><div><b>01</b><span>INDEPENDENT PRODUCT</span></div><div><b>RTL + LTR</b><span>BILINGUAL EXPERIENCE</span></div><div><b>DEMO</b><span>NO REAL TRANSACTIONS</span></div></div></section>

      <footer className="footer"><div className="footer-main"><div><a href="#" className="brand footer-brand"><span className="brand-mark">g<span>.</span></span><span className="brand-name">globalcart<small>THE WORLD, WELL FOUND.</small></span></a><p>{t.footer}</p><span className="footer-owner">{t.owner}</span></div><div className="footer-links"><b>EXPLORE</b><a href="#catalog">{t.nav1}</a><a href="#categories">{t.nav2}</a><a href="#about">{t.nav4}</a><a href="https://github.com/pooya16pg-spec/globalcart" target="_blank" rel="noreferrer">GitHub <ArrowUpLeftIcon/></a></div><div className="footer-disclaimer"><span><ShieldCheck size={17}/>{t.demo}</span><p>{t.noPayment}</p><small>PRODUCT DATA IS PROVIDED FOR DEMONSTRATION PURPOSES ONLY.</small></div></div><div className="footer-bottom"><span>© 2026 GLOBALCART · POOYA GOODARZI</span><span>INDEPENDENTLY BUILT. THOUGHTFULLY ENGINEERED.</span></div></footer>

      {cartOpen&&<div className="overlay" onMouseDown={e=>{if(e.target===e.currentTarget)setCartOpen(false);}}><aside className="side-drawer"><div className="drawer-head"><div><span className="section-kicker">YOUR SELECTION</span><h3>{t.cart} <small>({count})</small></h3></div><button className="icon-button" onClick={()=>setCartOpen(false)}><X/></button></div>{cart.length===0?<div className="drawer-empty"><ShoppingBag size={38}/><b>{t.emptyCart}</b><button className="button-primary" onClick={()=>setCartOpen(false)}>{t.continue}</button></div>:<><div className="drawer-items">{cart.map(line=><div className="drawer-item" key={line.product.id}><img src={line.product.thumbnail} alt={line.product.title}/><div className="drawer-item-info"><b>{line.product.title}</b><strong>{formatPrice(line.product)}</strong><div className="quantity"><button onClick={()=>changeQuantity(line.product.id,-1)}><Minus size={13}/></button><span>{line.quantity}</span><button onClick={()=>changeQuantity(line.product.id,1)}><Plus size={13}/></button></div></div><button className="remove-item" onClick={()=>setCart(old=>old.filter(x=>x.product.id!==line.product.id))}><X size={16}/></button></div>)}</div><div className="drawer-bottom"><div><span>{t.subtotal}</span><b>{rtl?money(subtotal,locale)+" تومان":usd(subtotal/rate)}</b></div><button className="button-primary full-button" onClick={()=>{setCartOpen(false);setCheckout(true);}}>{t.checkout}<ArrowLeft size={16}/></button><small>{t.noPayment}</small></div></>}</aside></div>}

      {favoriteOpen&&<div className="overlay" onMouseDown={e=>{if(e.target===e.currentTarget)setFavoriteOpen(false);}}><aside className="side-drawer"><div className="drawer-head"><div><span className="section-kicker">YOUR SHORTLIST</span><h3>{t.favorites} <small>({favorites.length})</small></h3></div><button className="icon-button" onClick={()=>setFavoriteOpen(false)}><X/></button></div>{favoriteProducts.length===0?<div className="drawer-empty"><Heart size={38}/><b>{t.emptyFav}</b></div>:<div className="drawer-items">{favoriteProducts.map(p=><div className="drawer-item" key={p.id}><img src={p.thumbnail} alt={p.title}/><div className="drawer-item-info"><b>{p.title}</b><strong>{formatPrice(p)}</strong><button className="text-action" onClick={()=>{setFavoriteOpen(false);openProduct(p);}}>{t.quick}</button></div><button className="remove-item" onClick={()=>toggleFavorite(p.id)}><X size={16}/></button></div>)}</div>}</aside></div>}

      {quick&&<div className="overlay modal-overlay" onMouseDown={e=>{if(e.target===e.currentTarget)setQuick(null);}}><div className="quick-modal"><button className="modal-close icon-button" onClick={()=>setQuick(null)}><X/></button><div className="quick-image"><img src={(quick.images&&quick.images[0])||quick.thumbnail} alt={quick.title}/></div><div className="quick-content"><span className="section-kicker">{quick.brand||"GLOBAL SELECT"} / PRODUCT DETAIL</span><h2>{quick.title}</h2><span className="quick-rating"><Star size={14} fill="currentColor"/>{quick.rating.toFixed(1)} · {quick.stock} {t.inStock}</span><p>{quick.description}</p><div className="quick-price"><b>{formatPrice(quick)}</b><small>{usd(quick.price)}</small></div><small className="source-label">{t.source}: {quick.source||"DummyJSON public demo API"}</small><button className="button-primary full-button" onClick={()=>{addToCart(quick);setQuick(null);}}>{t.add}<Plus size={17}/></button></div></div></div>}

      {checkout&&<div className="overlay modal-overlay"><div className="checkout-modal"><button className="modal-close icon-button" onClick={()=>setCheckout(false)}><X/></button><span className="section-kicker">GLOBALCART / DEMO FLOW</span><h2>{t.demoCheckout}</h2><p>{t.demoText}</p><div className="checkout-total"><span>{t.subtotal}</span><b>{rtl?money(subtotal,locale)+" تومان":usd(subtotal/rate)}</b></div><div className="checkout-warning"><ShieldCheck size={17}/>{t.noPayment}</div><button className="button-primary full-button" onClick={placeDemoOrder}>{t.confirm}<ArrowLeft size={16}/></button></div></div>}

      {toast&&<div className="toast"><Check size={15}/>{toast}</div>}
      <div className="demo-ribbon">{t.demo}</div>
    </main>
  );
}

function InfoIcon(){return <CircleHelp size={15}/>;}
function ArrowUpLeftIcon(){return <ArrowUpDown size={13}/>;}
