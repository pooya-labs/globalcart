"use client";

import { useEffect, useId, useMemo, useRef, useState, type FormEvent, type KeyboardEvent as ReactKeyboardEvent } from "react";
import Link from "next/link";
import { fetchDummyJsonCatalog } from "../lib/catalog";
import { filterCatalogProducts } from "../lib/catalog/filtering";
import { buildCatalogSuggestions, type CatalogSearchSuggestion } from "../lib/catalog/suggestions";
import { getCatalogRecommendationReason, rankCatalogRecommendations } from "../lib/catalog/recommendations";
import { hasSameSavedSearchFilters, normalizeSavedCatalogSearches, type SavedCatalogSearch } from "../lib/catalog/saved-searches";
import { getCatalogSourceStatus, type CatalogSourceStatus } from "../lib/catalog/source-status";
import { getCatalogFreshness, type CatalogFreshnessStatus } from "../lib/catalog/freshness";
import catalogSnapshot from "../data/catalog.json";
import fxSnapshot from "../data/fx-rate.json";
import { calculateDemoPrice } from "../lib/pricing";
import { getFxRateStatus, type FxRateStatus } from "../lib/fx/nobitex";
import {
  ArrowDown, ArrowLeft, ArrowRight, ArrowUpDown, BookmarkCheck, BookmarkPlus, Check, ChevronDown, ChevronLeft,
  ChevronRight, CircleHelp, Heart, LayoutGrid, Menu, Minus, Package, Plus,
  Search, ShieldCheck, ShoppingBag, SlidersHorizontal, Sparkles, Star, Trash2, Truck, GitCompareArrows,
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
    productSuggestion: "محصول", brandSuggestion: "برند", categorySuggestion: "دسته‌بندی",
    commandPalette: "فرمان‌های سریع", commandSearch: "دنبال چه کاری می‌گردی؟",
    commandHint: "برای اجرا Enter و برای بستن Esc را بزن", noCommands: "فرمانی پیدا نشد",
    commandCatalog: "رفتن به ویترین", commandCategories: "نمایش دسته‌بندی‌ها",
    commandCart: "باز کردن سبد خرید", commandFavorites: "علاقه‌مندی‌ها",
    commandCompare: "باز کردن مقایسه", commandClear: "پاک کردن فیلترها",
    commandLanguage: "تغییر زبان", keyboardHint: "Ctrl / ⌘ K",
    nav1: "فروشگاه", nav2: "دسته‌بندی‌ها", nav3: "پیشنهادهای منتخب", nav4: "درباره پروژه",
    heroEyebrow: "THE ART OF DISCOVERY", heroTitle: "چیزهای خوب،\nدرست همین‌جا پیدا می‌شوند.", heroDesc: "یک تجربه خرید مدرن برای کشف، مقایسه و انتخاب هوشمندانه؛ از محصولات روزمره تا انتخاب‌های خاص.",
    explore: "کشف محصولات", curated: "انتخاب‌های سردبیر", categories: "بر اساس علاقه بگرد", categoriesDesc: "از جایی شروع کن که به تو نزدیک‌تر است.",
    all: "همه محصولات", catalog: "ویترین جهانی", catalogDesc: "کالکشن نمونه با اطلاعات واردشده از منبع داده نمایشی.",
    sort: "مرتب‌سازی", popular: "پیشنهاد‌شده", low: "قیمت: کم به زیاد", high: "قیمت: زیاد به کم", rating: "بالاترین امتیاز",
    recommendationBasis: "رتبه‌بندی نمایشی بر پایه امتیاز و تخفیف نمونه؛ نه شخصی‌سازی بر اساس سابقه کاربر.",
    recommendationBoth: "امتیاز بالا + تخفیف", recommendationRating: "امتیاز بالای کاتالوگ",
    recommendationDiscount: "تخفیف قابل‌توجه", recommendationBalanced: "ترکیب متعادل امتیاز و قیمت",
    filter: "فیلترها", category: "دسته‌بندی", clear: "پاک کردن فیلتر", noResults: "محصولی با این مشخصات پیدا نشد.",
    add: "افزودن به سبد", quick: "جزئیات", cart: "سبد خرید", favorites: "علاقه‌مندی‌ها", emptyCart: "سبدت هنوز خالی است.",
    subtotal: "جمع جزء", checkout: "ادامه و ثبت سفارش نمایشی", demo: "نسخه نمایشی", noPayment: "پرداخت واقعی غیرفعال است؛ این پروژه فروشگاه واقعی نیست.",
    demoCheckout: "تکمیل سفارش آزمایشی", demoText: "این مرحله فقط برای نمایش تجربه کاربری است. هیچ پرداخت یا سفارش واقعی ثبت نمی‌شود.",
    confirm: "ثبت سفارش آزمایشی", success: "سفارش نمایشی ساخته شد", order: "شناسه سفارش", continue: "بازگشت به فروشگاه",
    source: "منبع", fxTitle: "نرخ بازار تتر", fxProvider: "نوبیتکس · USDT/IRT", fxFresh: "به‌روز", fxAging: "در حال کهنه‌شدن", fxStale: "قدیمی", fxUnavailable: "نرخ در دسترس نیست", fxUnknown: "زمان نامشخص", fxCaveat: "تتر دقیقاً معادل دلار نقدی نیست.", fxUnavailableNote: "تا دریافت نرخ معتبر، قیمت به دلار نمایش داده می‌شود.", inStock: "موجود در کاتالوگ نمونه", off: "تخفیف",
    searchResults: "نتایج جست‌وجو", loading: "در حال دریافت کاتالوگ...", sourceFallback: "منبع آنلاین در دسترس نیست؛ داده‌های نمونه محلی نمایش داده شده‌اند.", snapshotFallback: "به‌روزرسانی آنلاین در دسترس نیست؛ آخرین کاتالوگ تولیدشده نمایش داده می‌شود.",
    catalogLastChecked: "آخرین بررسی منبع", snapshotGenerated: "زمان تولید کاتالوگ", freshnessFresh: "تازه", freshnessAging: "در حال کهنه‌شدن", freshnessStale: "قدیمی", freshnessUnknown: "زمان نامشخص",
    why: "چرا GlobalCart؟", whyText: "یک پروژه مستقل برای نمایش مهندسی محصول، معماری فرانت‌اند، یکپارچه‌سازی داده و طراحی تجربه تجارت الکترونیک.",
    feature1: "کشف هوشمند", feature1d: "جست‌وجو و فیلتر سریع", feature2: "اطلاعات شفاف", feature2d: "نمایش منبع و قیمت", feature3: "تجربه امن", feature3d: "بدون تراکنش واقعی",
    footer: "یک تجربه تجارت الکترونیک مستقل، ساخته‌شده برای نمایش مهندسی محصول.", owner: "طراحی و توسعه: پویا گودرزی",
    count: "محصول", menu: "منو", compare: "مقایسه", recently: "تازه بازدیدشده", emptyFav: "هنوز محصولی ذخیره نکرده‌ای.",
    min: "کمترین قیمت", max: "بیشترین قیمت", result: "نتیجه", brand: "برند", allBrands: "همه برندها", minRatingLabel: "حداقل امتیاز", anyRating: "همه امتیازها",
    saveSearch: "ذخیره جست‌وجو", savedSearches: "جست‌وجوهای ذخیره‌شده", savedSearchName: "نام جست‌وجو", savedSearchPlaceholder: "مثلاً هدفون اقتصادی",
    save: "ذخیره", savedSearchEmpty: "هنوز جست‌وجویی ذخیره نشده است.", savedSearchDefault: "جست‌وجوی من",
    savedSearchSuccess: "جست‌وجو در همین مرورگر ذخیره شد.", savedSearchDuplicate: "این ترکیب فیلترها قبلاً ذخیره شده است.",
    applySavedSearch: "اعمال", deleteSavedSearch: "حذف", close: "بستن", compareAdd: "افزودن به مقایسه", compareTitle: "مقایسه محصولات", compareNeed: "برای مقایسه حداقل دو محصول انتخاب کن.", compareClear: "پاک کردن مقایسه", amazonEyebrow: "AMAZON / PRODUCT SOURCING", amazonTitle: "محصولت را در آمازون پیدا کردی؟", amazonDesc: "لینک مستقیم محصول را وارد کن تا درخواست بررسی آن در نسخه نمایشی ثبت شود.", amazonPlaceholder: "https://www.amazon.com/dp/...", amazonSubmit: "بررسی لینک محصول", amazonNote: "ثبت لینک در این نسخه نمایشی است؛ دریافت قیمت و موجودی زنده به دسترسی رسمی آمازون نیاز دارد.", amazonSuccess: "درخواست لینک محصول ذخیره شد", amazonInvalid: "لینک معتبر محصول آمازون با شناسه ASIN ده‌کاراکتری وارد کن.", decrease:"کاهش تعداد", increase:"افزایش تعداد", remove:"حذف محصول", searchClose:"بستن جست‌وجو"
  },
  en: {
    top: "A fresh perspective on finding good things", search: "Search products, brands or categories...",
    productSuggestion: "Product", brandSuggestion: "Brand", categorySuggestion: "Category",
    commandPalette: "Quick commands", commandSearch: "What would you like to do?",
    commandHint: "Press Enter to run · Esc to close", noCommands: "No commands found",
    commandCatalog: "Go to catalog", commandCategories: "Browse categories",
    commandCart: "Open shopping bag", commandFavorites: "Open saved items",
    commandCompare: "Open comparison", commandClear: "Clear all filters",
    commandLanguage: "Switch language", keyboardHint: "Ctrl / ⌘ K",
    nav1: "Shop", nav2: "Categories", nav3: "Editor's picks", nav4: "About",
    heroEyebrow: "THE ART OF DISCOVERY", heroTitle: "Good things,\nright where you need them.", heroDesc: "A modern shopping experience to discover, compare and choose with confidence — from everyday essentials to thoughtful finds.",
    explore: "Explore products", curated: "Editor's selection", categories: "Browse by interest", categoriesDesc: "Start with what speaks to you.",
    all: "All products", catalog: "The global edit", catalogDesc: "A sample collection imported from a public demo data source.",
    sort: "Sort by", popular: "Recommended", low: "Price: low to high", high: "Price: high to low", rating: "Top rated",
    recommendationBasis: "Demo ranking uses sample ratings and discounts; it is not personalized from browsing or purchase history.",
    recommendationBoth: "High rating + discount", recommendationRating: "Top catalog rating",
    recommendationDiscount: "Strong sample discount", recommendationBalanced: "Balanced rating and price",
    filter: "Filters", category: "Category", clear: "Clear filters", noResults: "No products match your search.",
    add: "Add to bag", quick: "Details", cart: "Your bag", favorites: "Saved items", emptyCart: "Your bag is still empty.",
    subtotal: "Subtotal", checkout: "Continue to demo checkout", demo: "DEMO", noPayment: "Real payments are disabled. This is not a live store.",
    demoCheckout: "Demo checkout", demoText: "This step demonstrates the user experience only. No real order or payment will be created.",
    confirm: "Place demo order", success: "Demo order created", order: "Order reference", continue: "Back to store",
    source: "Source", fxTitle: "Live Tether market quote", fxProvider: "Nobitex · USDT/IRT", fxFresh: "Fresh", fxAging: "Aging", fxStale: "Stale", fxUnavailable: "Rate unavailable", fxUnknown: "Timestamp unavailable", fxCaveat: "USDT is not identical to cash USD.", fxUnavailableNote: "Prices remain in USD until a valid exchange quote is available.", inStock: "Sample catalog item", off: "OFF",
    searchResults: "Search results", loading: "Loading catalog...", sourceFallback: "Live source is unavailable; showing local demo fixtures.", snapshotFallback: "Live refresh is unavailable; showing the last generated catalog snapshot.",
    catalogLastChecked: "Last source check", snapshotGenerated: "Snapshot generated", freshnessFresh: "Fresh", freshnessAging: "Aging", freshnessStale: "Stale", freshnessUnknown: "Unknown",
    why: "Why GlobalCart?", whyText: "An independent portfolio project showcasing product engineering, frontend architecture, data integration and commerce experience design.",
    feature1: "Smart discovery", feature1d: "Fast search and filters", feature2: "Clear information", feature2d: "Source and price context", feature3: "Safe experience", feature3d: "No real transactions",
    footer: "An independent commerce experience, built to showcase product engineering.", owner: "Designed & engineered by Pooya Goodarzi",
    count: "products", menu: "Menu", compare: "Compare", recently: "Recently viewed", emptyFav: "You have not saved anything yet.",
    min: "Min price", max: "Max price", result: "Results", brand: "Brand", allBrands: "All brands", minRatingLabel: "Minimum rating", anyRating: "Any rating",
    saveSearch: "Save search", savedSearches: "Saved searches", savedSearchName: "Search name", savedSearchPlaceholder: "e.g. Budget headphones",
    save: "Save", savedSearchEmpty: "No saved searches yet.", savedSearchDefault: "My search",
    savedSearchSuccess: "Saved in this browser.", savedSearchDuplicate: "These filters are already saved.",
    applySavedSearch: "Apply", deleteSavedSearch: "Delete", close: "Close", compareAdd: "Add to compare", compareTitle: "Compare products", compareNeed: "Choose at least two products to compare.", compareClear: "Clear comparison", amazonEyebrow: "AMAZON / PRODUCT SOURCING", amazonTitle: "Found it on Amazon?", amazonDesc: "Paste the product link to save a sourcing request in this demo experience.", amazonPlaceholder: "https://www.amazon.com/dp/...", amazonSubmit: "Validate product link", amazonNote: "Link intake is simulated; live product data and pricing require authorized Amazon API access.", amazonSuccess: "Product link request saved", amazonInvalid: "Enter a valid Amazon product URL containing a 10-character ASIN.", decrease:"Decrease quantity", increase:"Increase quantity", remove:"Remove item", searchClose:"Close search"
  }
} as const;

const snapshotProducts = Array.isArray(catalogSnapshot.products)
  ? (catalogSnapshot.products as unknown as Product[])
  : [];

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
const formatCatalogTimestamp = (value:string|null, locale:Locale) => {
  if(!value)return null;
  const timestamp=Date.parse(value);
  if(!Number.isFinite(timestamp))return null;
  return new Intl.DateTimeFormat(locale==="fa"?"fa-IR":"en-GB",{
    year:"numeric",month:"short",day:"numeric",hour:"2-digit",minute:"2-digit"
  }).format(new Date(timestamp));
};

const parseAmazonProductLink=(raw:string):{asin:string;url:string}|null=>{
  try{
    const url=new URL(raw.trim());
    const host=url.hostname.toLowerCase().replace(/^www\./,"");
    const domains=["amazon.com","amazon.co.uk","amazon.de","amazon.fr","amazon.it","amazon.es","amazon.ca","amazon.co.jp","amazon.in","amazon.com.au","amazon.com.br","amazon.com.mx","amazon.nl","amazon.se","amazon.pl","amazon.com.tr","amazon.ae","amazon.sa","amazon.sg","amazon.eg","amazon.com.be"];
    if(!domains.includes(host))return null;
    const match=url.pathname.match(/\/(?:dp|gp\/product|gp\/aw\/d)\/([a-z0-9]{10})(?:[/?]|$)/i);
    if(!match)return null;
    return {asin:match[1].toUpperCase(),url:url.toString()};
  }catch{return null;}
};

export default function Home() {
  const [locale,setLocale]=useState<Locale>("fa");
  const [products,setProducts]=useState<Product[]>(snapshotProducts.length ? snapshotProducts : fallback);
  const [catalogSourceState,setCatalogSourceState]=useState<CatalogSourceStatus>(
    getCatalogSourceStatus(snapshotProducts.length>0,"pending")
  );
  const [catalogTimestamp,setCatalogTimestamp]=useState<string|null>(
    typeof catalogSnapshot.generatedAt === "string" ? catalogSnapshot.generatedAt : null
  );
  const [query,setQuery]=useState("");
  const [category,setCategory]=useState("all");
  const [sort,setSort]=useState<SavedCatalogSearch["sort"]>("popular");
  const [cart,setCart]=useState<CartLine[]>([]);
  const [favorites,setFavorites]=useState<number[]>([]);
  const [compareIds,setCompareIds]=useState<number[]>([]);
  const [compareOpen,setCompareOpen]=useState(false);
  const [amazonUrl,setAmazonUrl]=useState("");
  const [amazonStatus,setAmazonStatus]=useState<"success"|"error"|null>(null);
  const [amazonAsin,setAmazonAsin]=useState("");
  const [recent,setRecent]=useState<number[]>([]);
  const [cartOpen,setCartOpen]=useState(false);
  const [favoriteOpen,setFavoriteOpen]=useState(false);
  const [quick,setQuick]=useState<Product|null>(null);
  const [checkout,setCheckout]=useState(false);
  const [order,setOrder]=useState("");
  const [mobileMenu,setMobileMenu]=useState(false);
  const [mobileSearchOpen,setMobileSearchOpen]=useState(false);
  const [categoryMenu,setCategoryMenu]=useState(false);
  const [toast,setToast]=useState("");
  const rate = typeof fxSnapshot.tomanPerUsdt === "number" && Number.isFinite(fxSnapshot.tomanPerUsdt) && fxSnapshot.tomanPerUsdt > 0 ? fxSnapshot.tomanPerUsdt : null;
  const [fxStatus,setFxStatus]=useState<FxRateStatus>("unavailable");
  const [hydrated,setHydrated]=useState(false);
  const [minPrice,setMinPrice]=useState(0);
  const [maxPrice,setMaxPrice]=useState(1000);
  const [brandFilter,setBrandFilter]=useState("");
  const [minRating,setMinRating]=useState(0);
  const [savedSearches,setSavedSearches]=useState<SavedCatalogSearch[]>([]);
  const [savedSearchesHydrated,setSavedSearchesHydrated]=useState(false);
  const [savedSearchPanelOpen,setSavedSearchPanelOpen]=useState(false);
  const [saveSearchFormOpen,setSaveSearchFormOpen]=useState(false);
  const [saveSearchName,setSaveSearchName]=useState("");
  const [savedSearchNotice,setSavedSearchNotice]=useState("");
  const [showFilters,setShowFilters]=useState(false);
  const [searchSuggestionsOpen,setSearchSuggestionsOpen]=useState(false);
  const [activeSuggestionIndex,setActiveSuggestionIndex]=useState(-1);
  const [commandPaletteOpen,setCommandPaletteOpen]=useState(false);
  const [commandQuery,setCommandQuery]=useState("");
  const [activeCommandIndex,setActiveCommandIndex]=useState(0);
  const searchInputRef=useRef<HTMLInputElement>(null);
  const commandInputRef=useRef<HTMLInputElement>(null);
  const commandPaletteReturnFocusRef=useRef<HTMLElement|null>(null);
  const nextDialogReturnFocusRef=useRef<HTMLElement|null>(null);
  const searchListboxId=useId();
  const commandInputId=useId();
  const t=copy[locale];
  const rtl=locale==="fa";
  const fxTimestampLabel = fxSnapshot.marketUpdatedAt
    ? new Intl.DateTimeFormat(locale==="fa"?"fa-IR":"en-GB",{timeZone:"Asia/Tehran",year:"numeric",month:"short",day:"numeric",hour:"2-digit",minute:"2-digit"}).format(new Date(fxSnapshot.marketUpdatedAt))
    : null;
  const fxStatusLabel = fxStatus==="fresh"?t.fxFresh:fxStatus==="aging"?t.fxAging:fxStatus==="stale"?t.fxStale:t.fxUnavailable;

  useEffect(()=>{
    const handleEscape=(event:KeyboardEvent)=>{
      if(event.key!=="Escape")return;
      if(commandPaletteOpen){closeCommandPalette();return;}
      if(searchSuggestionsOpen){setSearchSuggestionsOpen(false);setActiveSuggestionIndex(-1);return;}
      if(checkout){setCheckout(false);return;}
      if(compareOpen){setCompareOpen(false);return;}
      if(quick){setQuick(null);return;}
      if(favoriteOpen){setFavoriteOpen(false);return;}
      if(cartOpen){setCartOpen(false);return;}
      if(categoryMenu){setCategoryMenu(false);return;}
      if(mobileMenu)setMobileMenu(false);
    };
    window.addEventListener("keydown",handleEscape);
    return()=>window.removeEventListener("keydown",handleEscape);
  },[commandPaletteOpen,searchSuggestionsOpen,checkout,compareOpen,quick,favoriteOpen,cartOpen,categoryMenu,mobileMenu]);

  const modalActive=commandPaletteOpen||checkout||compareOpen||Boolean(quick)||favoriteOpen||cartOpen;
  const overlayModalActive=checkout||compareOpen||Boolean(quick)||favoriteOpen||cartOpen;

  const openCommandPalette=()=>{
    commandPaletteReturnFocusRef.current=document.activeElement instanceof HTMLElement?document.activeElement:null;
    setCommandQuery("");
    setActiveCommandIndex(0);
    setCommandPaletteOpen(true);
  };
  useEffect(()=>{
    const handleCommandShortcut=(event:KeyboardEvent)=>{
      if(!(event.metaKey||event.ctrlKey)||event.key.toLowerCase()!=="k")return;
      event.preventDefault();
      if(modalActive)return;
      setSearchSuggestionsOpen(false);
      setActiveSuggestionIndex(-1);
      openCommandPalette();
    };
    window.addEventListener("keydown",handleCommandShortcut);
    return()=>window.removeEventListener("keydown",handleCommandShortcut);
  },[modalActive]);
  useEffect(()=>{
    if(!modalActive)return;
    const dialog=document.querySelector<HTMLElement>('[role="dialog"][aria-modal="true"]');
    if(!dialog)return;
    const requestedReturnTarget=nextDialogReturnFocusRef.current;
    nextDialogReturnFocusRef.current=null;
    const previouslyFocused=requestedReturnTarget?.isConnected
      ?requestedReturnTarget
      :document.activeElement instanceof HTMLElement?document.activeElement:null;
    const selector='a[href],button:not([disabled]),input:not([disabled]),select:not([disabled]),textarea:not([disabled]),[tabindex]:not([tabindex="-1"])';
    const getFocusable=()=>Array.from(dialog.querySelectorAll<HTMLElement>(selector)).filter(el=>!el.hasAttribute("hidden")&&el.getAttribute("aria-hidden")!=="true");
    const focusable=getFocusable();
    (focusable[0]??dialog).focus();
    const trapTab=(event:KeyboardEvent)=>{
      if(event.key!=="Tab")return;
      const items=getFocusable();
      if(items.length===0){event.preventDefault();dialog.focus();return;}
      const first=items[0],last=items[items.length-1];
      if(event.shiftKey&&(document.activeElement===first||!dialog.contains(document.activeElement))){event.preventDefault();last.focus();}
      else if(!event.shiftKey&&(document.activeElement===last||!dialog.contains(document.activeElement))){event.preventDefault();first.focus();}
    };
    document.addEventListener("keydown",trapTab);
    return()=>{
      document.removeEventListener("keydown",trapTab);
      if(previouslyFocused?.isConnected)previouslyFocused.focus();
    };
  },[modalActive,commandPaletteOpen,checkout,compareOpen,quick,favoriteOpen,cartOpen]);

  useEffect(()=>{
    if(commandPaletteOpen){
      commandInputRef.current?.focus();
      return;
    }
    const target=commandPaletteReturnFocusRef.current;
    if(!target)return;
    commandPaletteReturnFocusRef.current=null;
    // A command may replace the palette with another modal (cart/favorites/compare).
    // Let that modal own focus instead of restoring it outside the new dialog.
    if(overlayModalActive)return;
    requestAnimationFrame(()=>{if(target.isConnected)target.focus();});
  },[commandPaletteOpen,overlayModalActive]);

  useEffect(()=>{
    if(mobileSearchOpen)searchInputRef.current?.focus();
  },[mobileSearchOpen]);

  const submitAmazonLink=(event:FormEvent<HTMLFormElement>)=>{
    event.preventDefault();
    const parsed=parseAmazonProductLink(amazonUrl);
    if(!parsed){setAmazonStatus("error");setAmazonAsin("");return;}
    try{
      const raw=localStorage.getItem("globalcart-amazon-requests-v1");
      const requests=raw?JSON.parse(raw):[];
      const next=Array.isArray(requests)?requests:[];
      next.push({...parsed,requestedAt:new Date().toISOString()});
      localStorage.setItem("globalcart-amazon-requests-v1",JSON.stringify(next.slice(-20)));
      setAmazonAsin(parsed.asin);
      setAmazonStatus("success");
      setAmazonUrl("");
    }catch{setAmazonStatus("error");setAmazonAsin("");}
  };

  useEffect(()=>{
    try{
      const raw=localStorage.getItem("globalcart-saved-searches-v1");
      if(raw)setSavedSearches(normalizeSavedCatalogSearches(JSON.parse(raw)));
    }catch{/* Invalid local saved-search state is ignored safely. */}
    finally{setSavedSearchesHydrated(true);}
  },[]);

  useEffect(()=>{
    if(!savedSearchesHydrated)return;
    try{localStorage.setItem("globalcart-saved-searches-v1",JSON.stringify(savedSearches));}catch{/* Browser storage can be unavailable. */}
  },[savedSearchesHydrated,savedSearches]);

  useEffect(()=>{setFxStatus(getFxRateStatus(fxSnapshot));},[]);

  useEffect(()=>{
    let active=true;
    fetchDummyJsonCatalog()
      .then(records=>{
        if(!active)return;
        const nextState=getCatalogSourceStatus(snapshotProducts.length>0,"success",records.length);
        setCatalogSourceState(nextState);
        if(nextState!=="live")return;
        setCatalogTimestamp(new Date().toISOString());
        const mapped:Product[]=records.map(p=>({
          id:p.id,title:p.title,description:p.description,category:p.category,
          price:p.price,discountPercentage:p.discountPercentage,rating:p.rating,
          stock:p.stock,brand:p.brand,thumbnail:p.thumbnail,images:p.images,
          source:p.source
        }));
        setProducts(mapped);
      })
      .catch(()=>{
        if(active){
          // Preserve the generated snapshot or fixture and surface its freshness state.
          setCatalogSourceState(getCatalogSourceStatus(snapshotProducts.length>0,"error"));
        }
      });
    return()=>{active=false;};
  },[]);

  useEffect(()=>{
    try{
      const raw=localStorage.getItem("globalcart-state-v1");
      if(!raw)return;
      const saved=JSON.parse(raw);
      if(Array.isArray(saved.cart))setCart(saved.cart);
      if(Array.isArray(saved.favorites))setFavorites(saved.favorites);
      if(Array.isArray(saved.compareIds))setCompareIds(saved.compareIds.slice(0,4));
      if(Array.isArray(saved.recent))setRecent(saved.recent);
    }catch{/* Ignore invalid local demo state. */}
    finally{setHydrated(true);}
  },[]);

  useEffect(()=>{
    if(!hydrated)return;
    try{localStorage.setItem("globalcart-state-v1",JSON.stringify({cart,favorites,recent,compareIds}));}catch{/* Storage can be unavailable. */}
  },[hydrated,cart,favorites,recent,compareIds]);

  const filtered=useMemo(()=>{
    const list=filterCatalogProducts(products,{category,query,minPrice,maxPrice,brand:brandFilter,minRating});
    if(sort==="low")list.sort((a,b)=>a.price-b.price);
    else if(sort==="high")list.sort((a,b)=>b.price-a.price);
    else if(sort==="rating")list.sort((a,b)=>b.rating-a.rating);
    else return rankCatalogRecommendations(list);
    return list;
  },[products,category,query,sort,minPrice,maxPrice,brandFilter,minRating]);

  const count=cart.reduce((sum,line)=>sum+line.quantity,0);
  const subtotalUsd=cart.reduce((sum,line)=>sum+line.product.price*(1-line.product.discountPercentage/100)*line.quantity,0);
  const subtotalToman=rate===null?null:calculateDemoPrice({sourcePrice:subtotalUsd,fxRate:rate}).toman;
  const formatPrice=(p:Product)=>{
    const discountedUsd=p.price*(1-p.discountPercentage/100);
    return rtl&&rate!==null?money(calculateDemoPrice({sourcePrice:p.price,fxRate:rate,discountPercentage:p.discountPercentage}).toman,locale)+" تومان":usd(discountedUsd);
  };
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
  const compareProducts=products.filter(p=>compareIds.includes(p.id));
  const dealPicks=useMemo(()=>[...products].filter(p=>p.discountPercentage>0).sort((a,b)=>b.discountPercentage-a.discountPercentage).slice(0,4),[products]);
  const topRatedPicks=useMemo(()=>[...products].sort((a,b)=>b.rating-a.rating).slice(0,4),[products]);
  const discoveryPicks=useMemo(()=>[...products].slice(0,4),[products]);
  const recentProducts=useMemo(()=>recent.map(id=>products.find(p=>p.id===id)).filter((p):p is Product=>Boolean(p)),[recent,products]);
  const toggleCompare=(id:number)=>{
    if(compareIds.includes(id)){setCompareIds(old=>old.filter(x=>x!==id));return;}
    if(compareIds.length>=4){notify(locale==="fa"?"حداکثر چهار محصول قابل مقایسه است":"You can compare up to four products");return;}
    setCompareIds(old=>[...old,id]);
  };
  const catalogCategories=useMemo(()=>Array.from(new Set(products.map(p=>p.category).filter(Boolean))).sort((a,b)=>a.localeCompare(b)),[products]);
  const catalogBrands=useMemo(
    ()=>Array.from(new Set(products.map(p=>p.brand?.trim()).filter((brand):brand is string=>Boolean(brand)))).sort((a,b)=>a.localeCompare(b)),
    [products]
  );
  const categoryLabel=(id:string)=>{const known=categories.find(c=>c.id===id);return known?(rtl?known.fa:known.en):id.split("-").map(word=>word.charAt(0).toUpperCase()+word.slice(1)).join(" ");};
  const selectCategory=(id:string)=>{setCategory(id);setCategoryMenu(false);setMobileMenu(false);document.getElementById("catalog")?.scrollIntoView({behavior:"smooth"});};
  const searchSuggestions=useMemo(()=>buildCatalogSuggestions(products,query),[products,query]);
  const selectSearchSuggestion=(suggestion:CatalogSearchSuggestion)=>{
    setSearchSuggestionsOpen(false);
    setActiveSuggestionIndex(-1);
    setMobileSearchOpen(false);
    if(suggestion.kind==="product"){
      const product=products.find(item=>item.id===suggestion.productId);
      setQuery("");
      if(product)openProduct(product);
      return;
    }
    setQuery("");
    if(suggestion.kind==="brand"){
      setCategory("all");
      setBrandFilter(suggestion.value);
      document.getElementById("catalog")?.scrollIntoView({behavior:"smooth"});
      return;
    }
    selectCategory(suggestion.value);
  };
  const currentSearchState=():Omit<SavedCatalogSearch,"id"|"name"|"createdAt">=>({
    query:query.trim(),category,brand:brandFilter,minPrice,maxPrice,minRating,sort
  });
  const saveCurrentSearch=(event:FormEvent<HTMLFormElement>)=>{
    event.preventDefault();
    const state=currentSearchState();
    const candidate:SavedCatalogSearch={
      ...state,
      id:Date.now().toString(36)+"-"+Math.random().toString(36).slice(2,8),
      name:saveSearchName.trim().slice(0,60)||state.query||state.brand||categoryLabel(state.category)||t.savedSearchDefault,
      createdAt:new Date().toISOString()
    };
    if(savedSearches.some(existing=>hasSameSavedSearchFilters(existing,candidate))){
      setSavedSearchNotice(t.savedSearchDuplicate);
      return;
    }
    setSavedSearches(current=>[candidate,...current].slice(0,12));
    setSavedSearchNotice(t.savedSearchSuccess);
    setSaveSearchName("");
    setSaveSearchFormOpen(false);
    setSavedSearchPanelOpen(true);
  };
  const applySavedSearch=(saved:SavedCatalogSearch)=>{
    setQuery(saved.query);setCategory(saved.category);setBrandFilter(saved.brand);
    setMinPrice(saved.minPrice);setMaxPrice(saved.maxPrice);setMinRating(saved.minRating);setSort(saved.sort);
    setSavedSearchPanelOpen(false);setSaveSearchFormOpen(false);setSavedSearchNotice("");
    document.getElementById("catalog")?.scrollIntoView({behavior:"smooth"});
  };
  const deleteSavedSearch=(id:string)=>setSavedSearches(current=>current.filter(saved=>saved.id!==id));

  const closeCommandPalette=()=>{
    setCommandPaletteOpen(false);
    setCommandQuery("");
    setActiveCommandIndex(0);
  };
  const commandItems=[
    {id:"catalog",label:t.commandCatalog,keywords:"shop products storefront catalog ویترین فروشگاه محصول"},
    {id:"categories",label:t.commandCategories,keywords:"category categories browse دسته بندی گروه"},
    {id:"cart",label:t.commandCart,keywords:"bag basket cart سبد خرید"},
    {id:"favorites",label:t.commandFavorites,keywords:"saved favorite wishlist علاقه مندی ذخیره"},
    {id:"compare",label:t.commandCompare,keywords:"comparison compare مقایسه"},
    {id:"clear",label:t.commandClear,keywords:"reset filters پاک کردن فیلتر بازنشانی"},
    {id:"language",label:t.commandLanguage,keywords:"locale فارسی english زبان"}
  ].filter(item=>(item.label+" "+item.keywords).toLocaleLowerCase().includes(commandQuery.trim().toLocaleLowerCase()));
  const runCommand=(id:string)=>{
    if(id==="cart"||id==="favorites"||(id==="compare"&&compareIds.length>=2)){
      nextDialogReturnFocusRef.current=commandPaletteReturnFocusRef.current;
    }
    closeCommandPalette();
    if(id==="catalog"){document.getElementById("catalog")?.scrollIntoView({behavior:"smooth"});return;}
    if(id==="categories"){setCategoryMenu(true);return;}
    if(id==="cart"){setCartOpen(true);return;}
    if(id==="favorites"){setFavoriteOpen(true);return;}
    if(id==="compare"){
      if(compareIds.length>=2)setCompareOpen(true);
      else notify(t.compareNeed);
      return;
    }
    if(id==="clear"){
      setCategory("all");setMinPrice(0);setMaxPrice(1000);setBrandFilter("");setMinRating(0);setQuery("");
      document.getElementById("catalog")?.scrollIntoView({behavior:"smooth"});
      return;
    }
    if(id==="language")setLocale(current=>current==="fa"?"en":"fa");
  };
  const handleSearchKeyDown=(event:ReactKeyboardEvent<HTMLInputElement>)=>{
    if(event.key==="Escape"&&searchSuggestionsOpen){
      event.preventDefault();event.stopPropagation();setSearchSuggestionsOpen(false);setActiveSuggestionIndex(-1);
      return;
    }
    if(searchSuggestions.length===0)return;
    if(event.key==="ArrowDown"){
      event.preventDefault();setSearchSuggestionsOpen(true);
      setActiveSuggestionIndex(current=>current<searchSuggestions.length-1?current+1:0);
    }else if(event.key==="ArrowUp"){
      event.preventDefault();setSearchSuggestionsOpen(true);
      setActiveSuggestionIndex(current=>current>0?current-1:searchSuggestions.length-1);
    }else if(event.key==="Enter"&&searchSuggestionsOpen&&activeSuggestionIndex>=0){
      event.preventDefault();selectSearchSuggestion(searchSuggestions[activeSuggestionIndex]);
    }else if(event.key==="Enter"){
      setSearchSuggestionsOpen(false);setActiveSuggestionIndex(-1);
      document.getElementById("catalog")?.scrollIntoView({behavior:"smooth"});
    }
  };

  return (
    <main id="top" className={rtl?"store rtl":"store ltr"} dir={rtl?"rtl":"ltr"}>
      <div className="topline"><span>{t.top}</span><span className="topline-right"><ShieldCheck size={13}/>{t.noPayment}</span></div>
      <header className="header">
        <button className="mobile-toggle icon-button" onClick={()=>setMobileMenu(!mobileMenu)} aria-label={t.menu} aria-expanded={mobileMenu} aria-controls="primary-navigation"><Menu size={21}/></button>
        <a href="#top" className="brand"><span className="brand-mark">g<span>.</span></span><span className="brand-name">globalcart<small>THE WORLD, WELL FOUND.</small></span></a>
        <nav id="primary-navigation" className={mobileMenu?"main-nav nav-open":"main-nav"}>
          <a href="#catalog" onClick={()=>setMobileMenu(false)}>{t.nav1}</a>
          <button className={categoryMenu?"nav-category-trigger active":"nav-category-trigger"} onClick={()=>setCategoryMenu(v=>!v)} aria-expanded={categoryMenu} aria-controls="category-menu-panel"><LayoutGrid size={15}/>{t.nav2}<ChevronDown size={13}/></button>
          <a href="#editors" onClick={()=>setMobileMenu(false)}>{t.nav3}</a>
          <a href="#about" onClick={()=>setMobileMenu(false)}>{t.nav4}</a>
        </nav>
        <div className="header-actions">
          <button type="button" className="mobile-search-trigger" onClick={()=>{setMobileSearchOpen(v=>!v);setSearchSuggestionsOpen(true);}} aria-label={mobileSearchOpen?t.searchClose:t.search} aria-expanded={mobileSearchOpen}><Search size={19}/></button>
          <div className={mobileSearchOpen?"search-area search-area-open":"search-area"} onBlur={event=>{if(!event.currentTarget.contains(event.relatedTarget as Node|null)){setSearchSuggestionsOpen(false);setActiveSuggestionIndex(-1);}}}>
            <label className={mobileSearchOpen?"search-box search-open":"search-box"}>
              <Search size={17}/>
              <input
                id={searchListboxId+"-input"} ref={searchInputRef} autoFocus={mobileSearchOpen}
                role="combobox" aria-label={t.search} aria-autocomplete="list"
                aria-expanded={searchSuggestionsOpen&&searchSuggestions.length>0}
                aria-controls={searchSuggestionsOpen&&searchSuggestions.length>0?searchListboxId:undefined}
                aria-activedescendant={searchSuggestionsOpen&&activeSuggestionIndex>=0?searchListboxId+"-option-"+activeSuggestionIndex:undefined}
                value={query}
                onFocus={()=>setSearchSuggestionsOpen(true)}
                onChange={event=>{setQuery(event.target.value);setSearchSuggestionsOpen(true);setActiveSuggestionIndex(-1);}}
                onKeyDown={handleSearchKeyDown}
                placeholder={t.search}
              />
              <kbd aria-hidden="true">{t.keyboardHint}</kbd>
              {mobileSearchOpen&&<button type="button" className="search-close" onClick={()=>{setMobileSearchOpen(false);setSearchSuggestionsOpen(false);}} aria-label={t.searchClose}><X size={17}/></button>}
            </label>
            {searchSuggestionsOpen&&searchSuggestions.length>0&&<div className="search-suggestions" id={searchListboxId} role="listbox" aria-label={rtl?"پیشنهادهای جست‌وجو":"Search suggestions"}>
              {searchSuggestions.map((suggestion,index)=><button
                type="button" role="option" id={searchListboxId+"-option-"+index}
                aria-selected={activeSuggestionIndex===index}
                className={activeSuggestionIndex===index?"search-suggestion active":"search-suggestion"}
                key={suggestion.key}
                onMouseDown={event=>event.preventDefault()}
                onClick={()=>selectSearchSuggestion(suggestion)}
              >
                <span className="search-suggestion-kind">{suggestion.kind==="product"?t.productSuggestion:suggestion.kind==="brand"?t.brandSuggestion:t.categorySuggestion}</span>
                <span className="search-suggestion-copy"><b>{suggestion.kind==="category"?categoryLabel(suggestion.label):suggestion.label}</b>{suggestion.detail&&<small>{suggestion.detail}</small>}</span>
                <ArrowLeft size={14}/>
              </button>)}
            </div>}
          </div>
          <button className="command-shortcut" type="button" onClick={openCommandPalette} aria-label={t.keyboardHint+" "+t.commandPalette}><span>{t.keyboardHint}</span></button>
          <button className="locale-button" onClick={()=>setLocale(rtl?"en":"fa")}>{rtl?"EN":"فا"}</button>
          <button className="icon-button favorite-trigger" onClick={()=>setFavoriteOpen(true)} aria-label={t.favorites+" "+favorites.length}><Heart size={19}/><span>{favorites.length}</span></button>
          <button className="bag-button" onClick={()=>setCartOpen(true)}><ShoppingBag size={17}/><span>{t.cart}</span><b>{count}</b></button>
        </div>
      </header>

      <nav className="marketplace-categorybar" aria-label={rtl?"دسته‌بندی‌های سریع":"Quick categories"}>
        <span className="categorybar-label">{rtl?"دسته‌بندی‌ها":"Shop by category"}</span>
        {catalogCategories.slice(0,8).map(id=><button key={id} onClick={()=>selectCategory(id)}>{categoryLabel(id)}</button>)}
        <button className="categorybar-all" onClick={()=>selectCategory("all")}>{rtl?"مشاهده همه":"View all"} <ArrowLeft size={13}/></button>
      </nav>
      {categoryMenu&&<div className="mega-menu-backdrop" onMouseDown={e=>{if(e.target===e.currentTarget)setCategoryMenu(false);}}>
        <section id="category-menu-panel" className="mega-menu" aria-label={rtl?"همه دسته‌بندی‌ها":"All product categories"} role="region">
          <div className="mega-menu-head"><div><span className="section-kicker">GLOBALCART / DISCOVER</span><h3>{rtl?"همه دسته‌بندی‌ها":"Explore all categories"}</h3><p>{rtl?"مستقیم به گروه کالایی موردنظرت برو.":"Jump straight into the categories available in the live catalog."}</p></div><button className="icon-button" onClick={()=>setCategoryMenu(false)} aria-label={t.close}><X size={18}/></button></div>
          <div className="mega-menu-grid">
            <button className={category==="all"?"mega-category selected":"mega-category"} onClick={()=>selectCategory("all")}><span className="mega-category-icon"><LayoutGrid size={18}/></span><span><b>{rtl?"همه محصولات":"All products"}</b><small>{products.length} {t.count}</small></span><ArrowLeft size={14}/></button>
            {catalogCategories.map((id,i)=><button key={id} className={category===id?"mega-category selected":"mega-category"} onClick={()=>selectCategory(id)}><span className="mega-category-icon">{["✳","◈","⌂","▣","◉","◇"][i%6]}</span><span><b>{categoryLabel(id)}</b><small>{products.filter(x=>x.category===id).length} {t.count}</small></span><ArrowLeft size={14}/></button>)}
          </div>
          <div className="mega-menu-foot"><span>{rtl?"دسته‌ها بر اساس کاتالوگ فعلی ساخته می‌شوند.":"Categories are generated from the current catalog."}</span><button onClick={()=>selectCategory("all")}>{rtl?"مشاهده همه محصولات":"Browse the full catalog"}<ArrowLeft size={14}/></button></div>
        </section>
      </div>}

      <section className="hero">
        <div className="hero-copy">
          <div className="eyebrow"><span></span>{t.heroEyebrow}</div>
          <h1>{t.heroTitle.split("\n").map((line,i)=><span key={i}>{line}</span>)}</h1>
          <p>{t.heroDesc}</p>
          <a className="button-primary" href="#catalog">{t.explore}<ArrowLeft size={16}/></a>
          <div className="hero-meta"><span>01</span><i></i><span>DISCOVER&nbsp; / &nbsp;COMPARE&nbsp; / &nbsp;CHOOSE</span></div>
        </div>
        <div className="hero-visual">
          <img src="https://images.unsplash.com/photo-1490312278390-ab64016e0aa9?w=1200&q=80" alt="A considered collection of lifestyle objects" loading="eager" fetchPriority="high" decoding="async"/>
          <div className="hero-image-caption"><small>GLOBAL EDITION / 001</small><b>Objects with a point of view.</b><span>CURATED FOR EVERYDAY LIFE</span></div>
          <div className="hero-index">GC<span>®</span></div>
          <div className="hero-side-note">A BETTER WAY TO BROWSE</div>
        </div>
      </section>

      <section className="amazon-request-section" id="amazon-request">
        <div className="amazon-request-copy"><span className="section-kicker">{t.amazonEyebrow}</span><h2>{t.amazonTitle}</h2><p>{t.amazonDesc}</p><small><ShieldCheck size={14}/>{t.amazonNote}</small></div>
        <form className="amazon-link-form" onSubmit={submitAmazonLink}>
          <label htmlFor="amazon-product-url">{rtl?"لینک محصول آمازون":"Amazon product URL"}</label>
          <div className="amazon-link-input"><input id="amazon-product-url" type="url" dir="ltr" required value={amazonUrl} onChange={e=>{setAmazonUrl(e.target.value);setAmazonStatus(null);}} placeholder={t.amazonPlaceholder}/><button type="submit"><Package size={16}/>{t.amazonSubmit}<ArrowLeft size={15}/></button></div>
          {amazonStatus==="success"&&<div className="amazon-link-feedback success" role="status" aria-live="polite"><Check size={15}/>{t.amazonSuccess} · ASIN: <b>{amazonAsin}</b></div>}
          {amazonStatus==="error"&&<div className="amazon-link-feedback error" role="alert"><CircleHelp size={15}/>{t.amazonInvalid}</div>}
        </form>
      </section>

      <section className="benefits">
        <div><span className="benefit-icon"><Sparkles size={18}/></span><span><b>{t.feature1}</b><small>{t.feature1d}</small></span></div>
        <div><span className="benefit-icon"><ShieldCheck size={18}/></span><span><b>{t.feature2}</b><small>{t.feature2d}</small></span></div>
        <div><span className="benefit-icon"><Truck size={18}/></span><span><b>{t.feature3}</b><small>{t.feature3d}</small></span></div>
        <div className="fx-control" data-status={fxStatus} role="status" aria-live="polite">
          <span><b>{t.fxTitle}</b><small>{t.fxProvider}</small></span>
          <strong>{rate===null?t.fxUnavailable:money(rate,locale)+" تومان / USDT"}</strong>
          <small>{fxTimestampLabel??t.fxUnknown} · {fxStatusLabel}</small>
          <small className="fx-caveat">{t.fxCaveat}</small>
          {rate===null&&<small>{t.fxUnavailableNote}</small>}
        </div>
      </section>

      <section className="section categories-section" id="categories">
        <div className="section-heading"><div><span className="section-kicker">01 / FIND YOUR INTEREST</span><h2>{t.categories}</h2><p>{t.categoriesDesc}</p></div><a href="#catalog">{t.all}<ArrowLeft size={15}/></a></div>
        <div className="category-strip">
          {categories.slice(1,9).map((c,i)=><button key={c.id} className={"category-tile tile-"+i} onClick={()=>{setCategory(c.id);document.getElementById("catalog")?.scrollIntoView({behavior:"smooth"});}}><span className="category-symbol">{c.symbol}</span><span>{rtl?c.fa:c.en}</span><ArrowLeft size={14}/></button>)}
        </div>
      </section>

      <section className="merchandising-section">
        <div className="merchandising-intro"><div><span className="section-kicker">02 / THE GLOBAL EDIT</span><h2>{rtl?"انتخاب‌های هوشمندانه، یک‌جا":"A better way to discover"}</h2><p>{rtl?"پیشنهادها و محصولات برتر بر اساس داده‌های کاتالوگ فعلی.":"Curated product rails powered by the current sample catalog."}</p></div><a href="#catalog">{rtl?"دیدن همه محصولات":"Explore all products"} <ArrowLeft size={15}/></a></div>
        <div className="merchandising-grid">
          {[{id:"deals",title:rtl?"فرصت‌های جذاب":"Deals worth a look",label:rtl?"بیشترین تخفیف":"TOP DISCOUNTS",items:dealPicks},{id:"rated",title:rtl?"محبوب‌ترین‌ها":"Top-rated picks",label:rtl?"امتیازهای برتر":"HIGHLY RATED",items:topRatedPicks},{id:"discover",title:rtl?"تازه‌های ویترین":"Fresh discoveries",label:rtl?"برای کشف بیشتر":"EDITOR'S EDIT",items:discoveryPicks}].map(group=><section className="merchandising-panel" key={group.id}><div className="merchandising-panel-head"><div><small>{group.label}</small><h3>{group.title}</h3></div><span>0{group.id==="deals"?1:group.id==="rated"?2:3}</span></div><div className="merchandising-items">{group.items.map(p=><Link href={`/product/${p.id}`} className="merchandising-item" key={p.id}><img src={p.thumbnail} alt={p.title} loading="lazy" decoding="async"/><span><b>{p.title}</b><small><Star size={11} fill="currentColor"/>{p.rating.toFixed(1)} · {formatPrice(p)}</small></span></Link>)}</div></section>)}
        </div>
      </section>

      {recentProducts.length>0&&<section className="recently-viewed-section"><div className="recently-viewed-head"><div><span className="section-kicker">03 / PICK UP WHERE YOU LEFT OFF</span><h2>{rtl?"بازدیدهای اخیر":"Recently viewed"}</h2><p>{rtl?"محصولاتی که اخیراً بررسی کرده‌ای؛ ذخیره‌شده فقط در همین مرورگر.":"Products you opened recently, stored only in this browser."}</p></div><div className="recently-viewed-actions"><Link href="/recently-viewed">{rtl?"مشاهده تاریخچه":"View history"} <ArrowLeft size={14}/></Link><button onClick={()=>setRecent([])}>{rtl?"پاک کردن تاریخچه":"Clear history"}</button></div></div><div className="recently-viewed-grid">{recentProducts.map(p=><Link href={`/product/${p.id}`} key={p.id} className="recently-viewed-item"><img src={p.thumbnail} alt={p.title} loading="lazy" decoding="async"/><span><b>{p.title}</b><small>{formatPrice(p)}</small></span><ArrowLeft size={15}/></Link>)}</div></section>}

      <section className="section catalog-section" id="catalog">
        <div className="section-heading catalog-heading"><div><span className="section-kicker">02 / THE COLLECTION</span><h2>{query?t.searchResults:t.catalog}</h2><p>{t.catalogDesc}</p>{sort==="popular"&&<small className="recommendation-basis">{t.recommendationBasis}</small>}</div><div className="catalog-tools">
          <button type="button" className="saved-search-trigger" onClick={()=>{setSaveSearchFormOpen(value=>!value);setSavedSearchPanelOpen(false);setSavedSearchNotice("");}} aria-expanded={saveSearchFormOpen}><BookmarkPlus size={15}/>{t.saveSearch}</button>
          <button type="button" className="saved-search-trigger" onClick={()=>{setSavedSearchPanelOpen(value=>!value);setSaveSearchFormOpen(false);setSavedSearchNotice("");}} aria-expanded={savedSearchPanelOpen}><BookmarkCheck size={15}/>{t.savedSearches} <span>{savedSearches.length}</span></button>
          <button className="filter-toggle" onClick={()=>setShowFilters(!showFilters)}><SlidersHorizontal size={16}/>{t.filter}</button><label className="sort-select"><ArrowUpDown size={15}/><select aria-label={t.sort} value={sort} onChange={e=>setSort(e.target.value as SavedCatalogSearch["sort"])}><option value="popular">{t.popular}</option><option value="low">{t.low}</option><option value="high">{t.high}</option><option value="rating">{t.rating}</option></select><ChevronDown size={14}/></label></div></div>
        {(saveSearchFormOpen||savedSearchPanelOpen)&&<div className="saved-search-panel">
          {saveSearchFormOpen&&<form className="saved-search-form" onSubmit={saveCurrentSearch}>
            <label htmlFor="saved-search-name">{t.savedSearchName}</label>
            <input id="saved-search-name" value={saveSearchName} maxLength={60} onChange={event=>{setSaveSearchName(event.target.value);setSavedSearchNotice("");}} placeholder={t.savedSearchPlaceholder}/>
            <button type="submit"><BookmarkPlus size={15}/>{t.save}</button>
          </form>}
          {savedSearchPanelOpen&&<div className="saved-search-list">
            {savedSearches.length===0?<p className="saved-search-empty">{t.savedSearchEmpty}</p>:savedSearches.map(saved=><div className="saved-search-item" key={saved.id}>
              <div><b>{saved.name}</b><small>{[saved.query,saved.brand,saved.category!=="all"?categoryLabel(saved.category):""].filter(Boolean).join(" · ")||t.catalog}</small></div>
              <button type="button" className="saved-search-apply" onClick={()=>applySavedSearch(saved)}>{t.applySavedSearch}</button>
              <button type="button" className="saved-search-delete" onClick={()=>deleteSavedSearch(saved.id)} aria-label={t.deleteSavedSearch+" "+saved.name}><Trash2 size={15}/></button>
            </div>)}
          </div>}
          {savedSearchNotice&&<p className="saved-search-notice" role="status">{savedSearchNotice}</p>}
        </div>}
        <div className="catalog-layout">
          <aside className={showFilters?"filter-panel filter-visible":"filter-panel"}>
            <div className="filter-title"><b><SlidersHorizontal size={16}/>{t.filter}</b><button onClick={()=>{setCategory("all");setMinPrice(0);setMaxPrice(1000);setBrandFilter("");setMinRating(0);setQuery("");}}>{t.clear}</button></div>
            <div className="filter-group"><b>{t.category}</b>{categories.map(c=><button key={c.id} className={category===c.id?"filter-option active":"filter-option"} onClick={()=>setCategory(c.id)}>{rtl?c.fa:c.en}<span>{c.id==="all"?products.length:products.filter(p=>p.category===c.id).length}</span></button>)}</div>
            <div className="filter-group price-filter"><b>{rtl?"بازه قیمت (دلار)": "Price range (USD)"}</b><label>{t.min}<input type="number" min="0" value={minPrice} onChange={e=>setMinPrice(Math.max(0,Number(e.target.value)||0))}/></label><label>{t.max}<input type="number" min="0" value={maxPrice} onChange={e=>setMaxPrice(Math.max(0,Number(e.target.value)||0))}/></label></div>
            <div className="filter-group facet-select-group"><b>{t.brand}</b><select aria-label={t.brand} value={brandFilter} onChange={e=>setBrandFilter(e.target.value)}><option value="">{t.allBrands}</option>{catalogBrands.map(brand=><option key={brand} value={brand}>{brand}</option>)}</select></div>
            <div className="filter-group facet-select-group"><b>{t.minRatingLabel}</b><select aria-label={t.minRatingLabel} value={minRating} onChange={e=>setMinRating(Number(e.target.value))}><option value={0}>{t.anyRating}</option><option value={3}>3.0+</option><option value={4}>4.0+</option><option value={4.5}>4.5+</option></select></div>
            <div className="filter-note"><InfoIcon/>{rtl?"داده‌ها نمایشی هستند و قیمت‌ها تعهد خرید نیستند.":"Demo data only. Prices are not purchase offers."}</div>
          </aside>
          <div className="product-area">
            {catalogSourceState==="loading"&&<div className="catalog-status" role="status" aria-live="polite"><span className="loader"></span>{t.loading}</div>}
            {catalogSourceState==="snapshot-fallback"&&<div className="catalog-status snapshot-fallback" role="status" aria-live="polite"><InfoIcon/>{t.snapshotFallback}</div>}
            {catalogSourceState==="fixture-fallback"&&<div className="catalog-status fixture-fallback" role="status" aria-live="polite"><InfoIcon/>{t.sourceFallback}</div>}
            {order&&<div className="order-success" role="status"><Check size={17}/>{t.success} · {t.order}: {order}<button onClick={()=>setOrder("")}><X size={15}/></button></div>}
            {filtered.length===0?<div className="no-results"><Search size={30}/><b>{t.noResults}</b><button onClick={()=>{setQuery("");setCategory("all");setMinPrice(0);setMaxPrice(1000);setBrandFilter("");setMinRating(0);}}>{t.clear}</button></div>:<div className="product-grid">
              {filtered.map((p,index)=><article className="product-card" key={p.id}>
                <div className="product-media"><Link className="product-image-link" href={`/product/${p.id}`} aria-label={rtl?"مشاهده صفحه محصول":"View product details"}><img src={p.thumbnail} alt={p.title} loading="lazy" decoding="async"/></Link><span className="product-number">GC / {String(index+1).padStart(2,"0")}</span>{p.discountPercentage>0&&<span className="discount-tag">{Math.round(p.discountPercentage)}% {t.off}</span>}<button className={favorites.includes(p.id)?"save-button saved":"save-button"} aria-label={t.favorites} onClick={e=>{e.stopPropagation();toggleFavorite(p.id);}}><Heart size={17} fill={favorites.includes(p.id)?"currentColor":"none"}/></button><button className={compareIds.includes(p.id)?"compare-card-button selected":"compare-card-button"} aria-label={t.compareAdd} title={t.compareAdd} onClick={e=>{e.stopPropagation();toggleCompare(p.id);}}><GitCompareArrows size={15}/></button><button className="quick-view" onClick={e=>{e.stopPropagation();openProduct(p);}}>{t.quick}<ArrowUpLeftIcon/></button></div>
                <div className="product-details"><span className="product-brand">{p.brand||"GLOBAL SELECT"}</span><h3><Link href={`/product/${p.id}`}>{p.title}</Link></h3><div className="product-rating"><span><Star size={12} fill="currentColor"/>{p.rating.toFixed(1)}</span><small>{t.inStock}</small></div><div className="recommendation-reason">{(()=>{const reason=getCatalogRecommendationReason(p);return reason==="rating-and-discount"?t.recommendationBoth:reason==="rating"?t.recommendationRating:reason==="discount"?t.recommendationDiscount:t.recommendationBalanced;})()}</div><div className="product-price"><div><b>{formatPrice(p)}</b><small>{usd(p.price)}</small></div><button onClick={()=>addToCart(p)} aria-label={t.add}><Plus size={19}/></button></div></div>
              </article>)}
            </div>}
            <div className="catalog-foot">
              <span>{t.result}: {filtered.length} / {products.length}</span>
              <span>{t.source}: {catalogSourceState==="live"?"DummyJSON public demo API":catalogSourceState==="fixture-fallback"?"Local demo fixture":"Generated catalog snapshot"}</span>
              {(() => {
                const freshness:CatalogFreshnessStatus=getCatalogFreshness(catalogTimestamp);
                const freshnessLabel=freshness==="fresh"?t.freshnessFresh:freshness==="aging"?t.freshnessAging:freshness==="stale"?t.freshnessStale:t.freshnessUnknown;
                const timestampLabel=catalogSourceState==="live"?t.catalogLastChecked:t.snapshotGenerated;
                const formattedTimestamp=formatCatalogTimestamp(catalogTimestamp,locale);
                return <span className="catalog-freshness" data-freshness={freshness}>
                  {formattedTimestamp
                    ? <>{timestampLabel}: <time dateTime={catalogTimestamp??undefined}>{formattedTimestamp}</time> · {freshnessLabel}</>
                    : <>{timestampLabel}: {t.freshnessUnknown}</>}
                </span>;
              })()}
            </div>
          </div>
        </div>
      </section>

      <section className="editorial" id="editors"><div className="editorial-image"><img src="https://images.unsplash.com/photo-1494438639946-1ebd1d20bf85?w=1200&q=80" alt="Quiet modern interior" loading="lazy" decoding="async"/></div><div className="editorial-copy"><span className="section-kicker">03 / A MORE CONSIDERED EXPERIENCE</span><h2>{rtl?"کمتر بگرد.\nبهتر انتخاب کن.":"Less scrolling.\nMore considered choices."}</h2><p>{t.whyText}</p><a href="#catalog" className="button-primary">{t.explore}<ArrowLeft size={16}/></a></div><span className="editorial-index">03</span></section>

      <section className="marketplace-services">
        <div className="section-heading"><div><span className="section-kicker">04 / MADE FOR REAL DISCOVERY</span><h2>{rtl?"از پیدا کردن تا انتخاب، ساده‌تر":"From discovery to a confident choice"}</h2><p>{rtl?"ابزارهای کاربردی برای مسیر انتخاب محصول؛ با مرز روشن بین تجربه نمایشی و خدمات واقعی.":"Useful tools for the product journey, with a clear boundary between demo features and live services."}</p></div></div>
        <div className="service-cards">
          <a href="#amazon-request" className="service-card"><span><Package size={19}/></span><b>{rtl?"درخواست کالای آمازون":"Request an Amazon item"}</b><p>{rtl?"لینک محصول را وارد کن تا درخواست نمونه ذخیره شود.":"Paste a product link to save a sourcing request in the demo."}</p><strong>{rtl?"ثبت لینک محصول":"Submit a product link"} <ArrowLeft size={14}/></strong></a>
          <a href="#catalog" className="service-card"><span><GitCompareArrows size={19}/></span><b>{rtl?"مقایسه هوشمند":"Side-by-side comparison"}</b><p>{rtl?"تا چهار محصول را از نظر قیمت، امتیاز و مشخصات کنار هم ببین.":"Compare up to four products by price, rating and key facts."}</p><strong>{rtl?"رفتن به ویترین":"Explore the catalog"} <ArrowLeft size={14}/></strong></a>
          <div className="service-card service-card-planned"><span><LayoutGrid size={19}/></span><b>{rtl?"پنل فروشندگان":"Seller workspace"}</b><p>{rtl?"مدیریت کالا و ابزارهای تأمین‌کننده در نقشه راه نسخه بعدی قرار دارد.":"Catalog management and supplier tools are planned for a later milestone."}</p><strong>{rtl?"در نقشه راه · M06":"Planned · M06"}</strong></div>
          <a href="https://github.com/pooya16pg-spec/globalcart/issues" target="_blank" rel="noreferrer" className="service-card"><span><CircleHelp size={19}/></span><b>{rtl?"راهنما و بازخورد":"Help & feedback"}</b><p>{rtl?"پیشنهاد یا ایراد رابط کاربری را از مخزن پروژه ثبت کن.":"Share feedback or report an issue through the project repository."}</p><strong>{rtl?"مرکز بازخورد":"Open feedback"} <ArrowLeft size={14}/></strong></a>
        </div>
      </section>

      <section className="faq-section" id="faq">
        <div className="faq-intro"><span className="section-kicker">05 / GOOD TO KNOW</span><h2>{rtl?"سوالات پرتکرار":"Frequently asked questions"}</h2><p>{rtl?"شفافیت درباره امکانات فعلی و چیزهایی که هنوز فعال نشده‌اند.":"A clear guide to what works today and what is still on the roadmap."}</p></div>
        <div className="faq-list">
          <details><summary>{rtl?"آیا GlobalCart فروشگاه واقعی است؟":"Is GlobalCart a live store?"}</summary><p>{rtl?"خیر؛ این پروژه یک شبیه‌سازی مستقل برای نمایش تجربه تجارت الکترونیک است. پرداخت، سفارش، موجودی و ارسال واقعی انجام نمی‌شود.":"No. This is an independent commerce experience demo. No real payments, orders, inventory reservations or shipping are processed."}</p></details>
          <details><summary>{rtl?"آیا قیمت و موجودی آمازون به‌صورت زنده دریافت می‌شود؟":"Does the site fetch live Amazon prices and stock?"}</summary><p>{rtl?"در حال حاضر خیر. برای دریافت اطلاعات زنده باید دسترسی رسمی و مجاز به API آمازون فراهم شود؛ این پروژه از scraping یا کپی غیرمجاز داده استفاده نمی‌کند.":"Not yet. Live data requires authorized Amazon API access. The project does not scrape or bulk-copy Amazon content."}</p></details>
          <details><summary>{rtl?"چطور لینک محصول آمازون را ثبت کنم؟":"How do I submit an Amazon product link?"}</summary><p>{rtl?"لینک صفحه محصول را در بخش درخواست کالای آمازون وارد کن. نسخه فعلی دامنه و شناسه ASIN را اعتبارسنجی و درخواست را فقط در مرورگر ذخیره می‌کند.":"Paste a product-page URL into the Amazon request module. The current demo validates the domain and ASIN, then stores the request locally in your browser."}</p></details>
          <details><summary>{rtl?"مقایسه محصولات چطور کار می‌کند؟":"How does product comparison work?"}</summary><p>{rtl?"آیکون مقایسه روی کارت محصولات را انتخاب کن؛ بعد از انتخاب حداقل دو محصول، نوار مقایسه باز می‌شود و می‌توانی حداکثر چهار محصول را کنار هم ببینی.":"Select the compare icon on product cards. Once at least two items are selected, open the comparison tray to compare up to four products."}</p></details>
        </div>
      </section>

      <section className="about-section" id="about"><div><span className="section-kicker">04 / BUILT AS A PORTFOLIO PROJECT</span><h2>{t.why}</h2><p>{t.whyText}</p></div><div className="about-stats"><div><b>01</b><span>INDEPENDENT PRODUCT</span></div><div><b>RTL + LTR</b><span>BILINGUAL EXPERIENCE</span></div><div><b>DEMO</b><span>NO REAL TRANSACTIONS</span></div></div></section>

      <footer className="footer"><div className="footer-main"><div><a href="#top" className="brand footer-brand"><span className="brand-mark">g<span>.</span></span><span className="brand-name">globalcart<small>THE WORLD, WELL FOUND.</small></span></a><p>{t.footer}</p><span className="footer-owner">{t.owner}</span></div><div className="footer-links"><b>EXPLORE</b><a href="#catalog">{t.nav1}</a><a href="#categories">{t.nav2}</a><Link href="/recently-viewed">{t.recently}</Link><a href="#about">{t.nav4}</a><a href="https://github.com/pooya16pg-spec/globalcart" target="_blank" rel="noreferrer">GitHub <ArrowUpLeftIcon/></a></div><div className="footer-disclaimer"><span><ShieldCheck size={17}/>{t.demo}</span><p>{t.noPayment}</p><small>PRODUCT DATA IS PROVIDED FOR DEMONSTRATION PURPOSES ONLY.</small></div></div><div className="footer-bottom"><span>© 2026 GLOBALCART · POOYA GOODARZI</span><span>INDEPENDENTLY BUILT. THOUGHTFULLY ENGINEERED.</span></div></footer>

      {commandPaletteOpen&&<div className="overlay modal-overlay command-palette-overlay" onMouseDown={event=>{if(event.target===event.currentTarget)closeCommandPalette();}}>
        <section className="command-palette" role="dialog" aria-modal="true" aria-labelledby={commandInputId+"-title"} aria-describedby={commandInputId+"-hint"} tabIndex={-1}>
          <div className="command-palette-heading"><div><span className="section-kicker">GLOBALCART / COMMANDS</span><h2 id={commandInputId+"-title"}>{t.commandPalette}</h2></div><button className="icon-button" onClick={closeCommandPalette} aria-label={t.close}><X size={18}/></button></div>
          <label className="command-palette-search" htmlFor={commandInputId}><Search size={18}/><input
            id={commandInputId} ref={commandInputRef} autoFocus value={commandQuery}
            role="combobox" aria-autocomplete="list" aria-controls={commandInputId+"-list"}
            aria-expanded={commandItems.length>0}
            aria-activedescendant={commandItems[activeCommandIndex]?commandInputId+"-option-"+activeCommandIndex:undefined}
            onChange={event=>{setCommandQuery(event.target.value);setActiveCommandIndex(0);}}
            onKeyDown={event=>{
              if(event.key==="ArrowDown"){event.preventDefault();setActiveCommandIndex(current=>commandItems.length?Math.min(current+1,commandItems.length-1):0);}
              else if(event.key==="ArrowUp"){event.preventDefault();setActiveCommandIndex(current=>Math.max(current-1,0));}
              else if(event.key==="Enter"){event.preventDefault();const chosen=commandItems[activeCommandIndex]??commandItems[0];if(chosen)runCommand(chosen.id);}
            }}
            placeholder={t.commandSearch}
          /></label>
          <div className="command-palette-list" id={commandInputId+"-list"} role="listbox" aria-label={t.commandPalette}>
            {commandItems.length===0?<p className="command-palette-empty" role="status">{t.noCommands}</p>:commandItems.map((command,index)=><button type="button" role="option" id={commandInputId+"-option-"+index} aria-selected={index===activeCommandIndex} className={index===activeCommandIndex?"command-option active":"command-option"} key={command.id} onMouseEnter={()=>setActiveCommandIndex(index)} onClick={()=>runCommand(command.id)}>
              <span className="command-option-icon">{command.id==="catalog"?<Search size={17}/>:command.id==="categories"?<LayoutGrid size={17}/>:command.id==="cart"?<ShoppingBag size={17}/>:command.id==="favorites"?<Heart size={17}/>:command.id==="compare"?<GitCompareArrows size={17}/>:command.id==="clear"?<X size={17}/>:<Zap size={17}/>}</span>
              <b>{command.label}</b><span className="command-option-key">{index===activeCommandIndex?"↵":index+1}</span>
            </button>)}
          </div>
          <div className="command-palette-footer" id={commandInputId+"-hint"}><span>{t.commandHint}</span><span><kbd>↑</kbd><kbd>↓</kbd><kbd>Enter</kbd><kbd>Esc</kbd></span></div>
        </section>
      </div>}

      {cartOpen&&<div className="overlay" onMouseDown={e=>{if(e.target===e.currentTarget)setCartOpen(false);}}><aside className="side-drawer" role="dialog" aria-modal="true" aria-labelledby="cart-dialog-title" tabIndex={-1}><div className="drawer-head"><div><span className="section-kicker">YOUR SELECTION</span><h3 id="cart-dialog-title">{t.cart} <small>({count})</small></h3></div><button className="icon-button" onClick={()=>setCartOpen(false)} aria-label={t.close}><X/></button></div>{cart.length===0?<div className="drawer-empty"><ShoppingBag size={38}/><b>{t.emptyCart}</b><button className="button-primary" onClick={()=>setCartOpen(false)}>{t.continue}</button></div>:<><div className="drawer-items">{cart.map(line=><div className="drawer-item" key={line.product.id}><img src={line.product.thumbnail} alt={line.product.title} loading="lazy" decoding="async"/><div className="drawer-item-info"><b>{line.product.title}</b><strong>{formatPrice(line.product)}</strong><div className="quantity"><button aria-label={t.decrease} onClick={()=>changeQuantity(line.product.id,-1)}><Minus size={13}/></button><span>{line.quantity}</span><button aria-label={t.increase} onClick={()=>changeQuantity(line.product.id,1)}><Plus size={13}/></button></div></div><button className="remove-item" aria-label={t.remove} onClick={()=>setCart(old=>old.filter(x=>x.product.id!==line.product.id))}><X size={16}/></button></div>)}</div><div className="drawer-bottom"><div><span>{t.subtotal}</span><b>{rtl&&subtotalToman!==null?money(subtotalToman,locale)+" تومان":usd(subtotalUsd)}</b></div><button className="button-primary full-button" onClick={()=>{setCartOpen(false);setCheckout(true);}}>{t.checkout}<ArrowLeft size={16}/></button><small>{t.noPayment}</small></div></>}</aside></div>}

      {favoriteOpen&&<div className="overlay" onMouseDown={e=>{if(e.target===e.currentTarget)setFavoriteOpen(false);}}><aside className="side-drawer" role="dialog" aria-modal="true" aria-labelledby="favorites-dialog-title" tabIndex={-1}><div className="drawer-head"><div><span className="section-kicker">YOUR SHORTLIST</span><h3 id="favorites-dialog-title">{t.favorites} <small>({favorites.length})</small></h3></div><button className="icon-button" onClick={()=>setFavoriteOpen(false)} aria-label={t.close}><X/></button></div>{favoriteProducts.length===0?<div className="drawer-empty"><Heart size={38}/><b>{t.emptyFav}</b></div>:<div className="drawer-items">{favoriteProducts.map(p=><div className="drawer-item" key={p.id}><img src={p.thumbnail} alt={p.title} loading="lazy" decoding="async"/><div className="drawer-item-info"><b>{p.title}</b><strong>{formatPrice(p)}</strong><button className="text-action" onClick={()=>{setFavoriteOpen(false);openProduct(p);}}>{t.quick}</button></div><button className="remove-item" aria-label={t.remove} onClick={()=>toggleFavorite(p.id)}><X size={16}/></button></div>)}</div>}</aside></div>}

      {quick&&<div className="overlay modal-overlay" onMouseDown={e=>{if(e.target===e.currentTarget)setQuick(null);}}><div className="quick-modal" role="dialog" aria-modal="true" aria-labelledby="quick-dialog-title" tabIndex={-1}><button className="modal-close icon-button" onClick={()=>setQuick(null)} aria-label={t.close}><X/></button><div className="quick-image"><img src={(quick.images&&quick.images[0])||quick.thumbnail} alt={quick.title} loading="lazy" decoding="async"/></div><div className="quick-content"><span className="section-kicker">{quick.brand||"GLOBAL SELECT"} / PRODUCT DETAIL</span><h2 id="quick-dialog-title">{quick.title}</h2><span className="quick-rating"><Star size={14} fill="currentColor"/>{quick.rating.toFixed(1)} · {quick.stock} {t.inStock}</span><p>{quick.description}</p><div className="quick-price"><b>{formatPrice(quick)}</b><small>{usd(quick.price)}</small></div><small className="source-label">{t.source}: {quick.source||"DummyJSON public demo API"}</small><button className="button-primary full-button" onClick={()=>{addToCart(quick);setQuick(null);}}>{t.add}<Plus size={17}/></button></div></div></div>}

      {compareIds.length>0&&<div className="compare-tray"><div><GitCompareArrows size={17}/><b>{rtl?"مقایسه محصولات":"Product comparison"}</b><span>{compareIds.length}/4</span></div><button className="compare-tray-open" disabled={compareIds.length<2} onClick={()=>setCompareOpen(true)}>{rtl?"مقایسه":"Compare"} <ArrowLeft size={15}/></button><button className="compare-tray-clear" onClick={()=>setCompareIds([])}>{rtl?"پاک کردن":"Clear"}</button></div>}

      {compareOpen&&<div className="overlay modal-overlay compare-overlay" onMouseDown={e=>{if(e.target===e.currentTarget)setCompareOpen(false);}}><section className="compare-modal" role="dialog" aria-modal="true" aria-labelledby="compare-dialog-title" tabIndex={-1}><div className="compare-modal-head"><div><span className="section-kicker">GLOBALCART / SIDE BY SIDE</span><h2 id="compare-dialog-title">{t.compareTitle}</h2></div><button className="icon-button" onClick={()=>setCompareOpen(false)} aria-label={t.close}><X/></button></div><div className="compare-table-wrap"><table className="compare-table"><thead><tr><th>{rtl?"مشخصات":"Details"}</th>{compareProducts.map(p=><th key={p.id}><img src={p.thumbnail} alt="" loading="lazy" decoding="async"/><b>{p.title}</b><button onClick={()=>toggleCompare(p.id)}>{rtl?"حذف":"Remove"}</button></th>)}</tr></thead><tbody><tr><th>{rtl?"قیمت فعلی":"Current price"}</th>{compareProducts.map(p=><td key={p.id}>{formatPrice(p)}</td>)}</tr><tr><th>{rtl?"قیمت اصلی":"Original price"}</th>{compareProducts.map(p=><td key={p.id}>{rtl&&rate!==null?money(calculateDemoPrice({sourcePrice:p.price,fxRate:rate}).toman,locale)+" تومان":usd(p.price)}</td>)}</tr><tr><th>{rtl?"امتیاز":"Rating"}</th>{compareProducts.map(p=><td key={p.id}><span className="compare-rating"><Star size={13} fill="currentColor"/>{p.rating.toFixed(1)}</span></td>)}</tr><tr><th>{rtl?"تخفیف":"Discount"}</th>{compareProducts.map(p=><td key={p.id}>{Math.round(p.discountPercentage)}%</td>)}</tr><tr><th>{rtl?"برند":"Brand"}</th>{compareProducts.map(p=><td key={p.id}>{p.brand||"—"}</td>)}</tr><tr><th>{rtl?"موجودی منبع نمونه":"Source stock"}</th>{compareProducts.map(p=><td key={p.id}>{p.stock}</td>)}</tr></tbody></table></div><div className="compare-modal-foot"><span>{rtl?"اطلاعات برای مقایسه نمایشی هستند.":"Figures are provided for demo comparison only."}</span><button onClick={()=>{setCompareIds([]);setCompareOpen(false);}}>{t.compareClear}</button></div></section></div>}

            {checkout&&<div className="overlay modal-overlay"><div className="checkout-modal" role="dialog" aria-modal="true" aria-labelledby="checkout-dialog-title" tabIndex={-1}><button className="modal-close icon-button" onClick={()=>setCheckout(false)} aria-label={t.close}><X/></button><span className="section-kicker">GLOBALCART / DEMO FLOW</span><h2 id="checkout-dialog-title">{t.demoCheckout}</h2><p>{t.demoText}</p><div className="checkout-total"><span>{t.subtotal}</span><b>{rtl&&subtotalToman!==null?money(subtotalToman,locale)+" تومان":usd(subtotalUsd)}</b></div><div className="checkout-warning"><ShieldCheck size={17}/>{t.noPayment}</div><button className="button-primary full-button" onClick={placeDemoOrder}>{t.confirm}<ArrowLeft size={16}/></button></div></div>}

      {toast&&<div className="toast" role="status" aria-live="polite"><Check size={15}/>{toast}</div>}
      <div className="demo-ribbon">{t.demo}</div>
    </main>
  );
}

function InfoIcon(){return <CircleHelp size={15}/>;}
function ArrowUpLeftIcon(){return <ArrowUpDown size={13}/>;}
