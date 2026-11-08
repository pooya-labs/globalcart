"use client";

import { useEffect, type Dispatch, type SetStateAction } from "react";
import Link from "next/link";
import {
  ArrowLeft, GitCompareArrows, Heart, Minus, Plus, ShieldCheck, ShoppingBag, Star, X
} from "lucide-react";

import { calculateDemoPrice, hasDisplayPrice } from "../lib/pricing";
import type { FxRateStatus } from "../lib/fx/tgju";
import type { Locale } from "../lib/locale";
import { getPersianProductCopy } from "../lib/catalog/editorial";
import type { CartLine, Product, StorefrontLabels } from "./StorefrontClient";

type BooleanSetter = Dispatch<SetStateAction<boolean>>;

type StorefrontDialogsProps = {
  t: StorefrontLabels;
  locale: Locale;
  rtl: boolean;
  cart: CartLine[];
  count: number;
  cartOpen: boolean;
  setCartOpen: BooleanSetter;
  changeQuantity: (id: number, delta: number) => void;
  removeCartLine: (line: CartLine, index: number) => void;
  favorites: number[];
  favoriteOpen: boolean;
  setFavoriteOpen: BooleanSetter;
  favoriteProducts: Product[];
  toggleFavorite: (id: number) => void;
  quick: Product | null;
  setQuick: Dispatch<SetStateAction<Product | null>>;
  isE2EFixtureCatalog: boolean;
  formatPrice: (product: Product) => string;
  catalogImageSrc: (source: string) => string;
  usd: (value: number) => string;
  money: (value: number, locale: Locale) => string;
  addToCart: (product: Product) => void;
  compareIds: number[];
  setCompareIds: Dispatch<SetStateAction<number[]>>;
  compareOpen: boolean;
  setCompareOpen: BooleanSetter;
  compareProducts: Product[];
  toggleCompare: (id: number) => void;
  checkout: boolean;
  setCheckout: BooleanSetter;
  subtotalHasUnavailablePrice: boolean;
  subtotalLabel: string;
  subtotalToman: number | null;
  subtotalUsd: number;
  rate: number | null;
  fxStatus: FxRateStatus;
  fxStatusLabel: string;
  placeDemoOrder: () => void;
  onReady: () => void;
};

export default function StorefrontDialogs({
  t, locale, rtl, cart, count, cartOpen, setCartOpen, changeQuantity, removeCartLine,
  favorites, favoriteOpen, setFavoriteOpen, favoriteProducts, toggleFavorite,
  quick, setQuick, isE2EFixtureCatalog, formatPrice, catalogImageSrc, usd, money, addToCart,
  compareIds, setCompareIds, compareOpen, setCompareOpen, compareProducts, toggleCompare,
  checkout, setCheckout, subtotalHasUnavailablePrice, subtotalLabel, subtotalToman, subtotalUsd, rate,
  fxStatus, fxStatusLabel, placeDemoOrder, onReady
}: StorefrontDialogsProps) {
  useEffect(() => {
    onReady();
  }, [onReady]);

  return (
    <>
      {cartOpen&&<div className="overlay" onMouseDown={e=>{if(e.target===e.currentTarget)setCartOpen(false);}}><aside className="side-drawer" role="dialog" aria-modal="true" aria-labelledby="cart-dialog-title" tabIndex={-1}><div className="drawer-head"><div><span className="section-kicker">YOUR SELECTION</span><h3 id="cart-dialog-title">{t.cart} <small>({count})</small></h3></div><button className="icon-button" onClick={()=>setCartOpen(false)} aria-label={t.close}><X/></button></div><p className="drawer-demo-note" role="note">{t.noPayment}</p>{cart.length===0?<div className="drawer-empty"><span className="empty-state-icon"><ShoppingBag size={25}/></span><b>{t.emptyCart}</b><p>{t.emptyCartHint}</p><button className="button-primary" onClick={()=>setCartOpen(false)}>{t.continue}</button></div>:<><div className="drawer-items">{cart.map((line,index)=><div className="drawer-item" key={line.product.id}><img src={catalogImageSrc(line.product.thumbnail)} alt={line.product.title} loading="lazy" decoding="async" fetchPriority="low" width={1440} height={1440}/><div className="drawer-item-info"><b>{rtl?(getPersianProductCopy(line.product.id)?.title??line.product.title):line.product.title}</b><strong>{formatPrice(line.product)}</strong><div className="quantity"><button aria-label={t.decrease} onClick={()=>changeQuantity(line.product.id,-1)}><Minus size={13}/></button><span>{line.quantity}</span><button aria-label={t.increase} onClick={()=>changeQuantity(line.product.id,1)}><Plus size={13}/></button></div></div><button className="remove-item" aria-label={t.remove} onClick={()=>removeCartLine(line,index)}><X size={16}/></button></div>)}</div><div className="drawer-bottom"><div><span>{t.subtotal}</span><b>{subtotalHasUnavailablePrice?(rtl?"قیمت در منبع موجود نیست":"Price not provided by source"):rtl&&subtotalToman!==null?money(subtotalToman,locale)+" تومان":usd(subtotalUsd)}</b><small className="drawer-fx-note" data-status={fxStatus}>{rtl?"وضعیت نرخ تبدیل TGJU: ":"TGJU conversion rate: "}{fxStatusLabel}</small></div><button className="button-primary full-button" onClick={()=>{setCartOpen(false);setCheckout(true);}}>{t.checkout}<ArrowLeft size={16}/></button></div></>}</aside></div>}

      {favoriteOpen&&<div className="overlay" onMouseDown={e=>{if(e.target===e.currentTarget)setFavoriteOpen(false);}}><aside className="side-drawer" role="dialog" aria-modal="true" aria-labelledby="favorites-dialog-title" tabIndex={-1}><div className="drawer-head"><div><span className="section-kicker">YOUR SHORTLIST</span><h3 id="favorites-dialog-title">{t.favorites} <small>({favorites.length})</small></h3></div><button className="icon-button" onClick={()=>setFavoriteOpen(false)} aria-label={t.close}><X/></button></div>{favoriteProducts.length===0?<div className="drawer-empty"><span className="empty-state-icon"><Heart size={25}/></span><b>{t.emptyFav}</b><p>{t.emptyFavHint}</p><button className="button-primary" onClick={()=>setFavoriteOpen(false)}>{t.continue}</button></div>:<div className="drawer-items">{favoriteProducts.map(p=><div className="drawer-item" key={p.id}><img src={catalogImageSrc(p.thumbnail)} alt={p.title} loading="lazy" decoding="async" referrerPolicy="no-referrer"/><div className="drawer-item-info"><b>{rtl?(getPersianProductCopy(p.id)?.title??p.title):p.title}</b><strong>{formatPrice(p)}</strong><Link className="text-action" href={`/product/${p.id}`} onClick={()=>setFavoriteOpen(false)}>{t.quick}</Link></div><button className="remove-item" aria-label={t.remove} onClick={()=>toggleFavorite(p.id)}><X size={16}/></button></div>)}</div>}</aside></div>}

      {quick&&<div className="overlay modal-overlay" onMouseDown={e=>{if(e.target===e.currentTarget)setQuick(null);}}><div className="quick-modal" role="dialog" aria-modal="true" aria-labelledby="quick-dialog-title" tabIndex={-1}><button className="modal-close icon-button" onClick={()=>setQuick(null)} aria-label={t.close}><X/></button><div className="quick-image"><img src={catalogImageSrc((quick.images&&quick.images[0])||quick.thumbnail)} alt={quick.title} loading="lazy" decoding="async" fetchPriority="low" width={1440} height={1440}/></div><div className="quick-content">{isE2EFixtureCatalog?<><span className="section-kicker">{quick.brand||"GLOBAL SELECT"} / PRODUCT DETAIL</span><h2 id="quick-dialog-title">{quick.title}</h2><span className="quick-rating"><Star size={14} fill="currentColor"/>{quick.rating.toFixed(1)} · {quick.stock} {t.inStock}</span><p>{quick.description}</p><div className="quick-price"><b>{formatPrice(quick)}</b><small>{usd(quick.price)}</small></div><small className="source-label">{t.source}: {quick.source||"DummyJSON public demo API"}</small><button className="button-primary full-button" onClick={()=>{addToCart(quick);setQuick(null);}}>{t.add}<Plus size={17}/></button></>:<><span className="section-kicker">{quick.brand||"GLOBALCART"} / QUICK VIEW</span><h2 id="quick-dialog-title">{rtl?(getPersianProductCopy(quick.id)?.title??quick.title):quick.title}</h2>{rtl&&getPersianProductCopy(quick.id)&&<small className="quick-source-title" dir="auto">{quick.title}</small>}<span className="quick-rating">{quick.rating>0?<><Star size={14} fill="currentColor"/>{quick.rating.toFixed(1)}</>:null} {quick.stockAvailable===false?(rtl?"موجودی ثبت نشده":"Stock not provided"):quick.availability===false?(rtl?"ناموجود":"Unavailable"):(rtl?"اعلام موجودی فروشنده":"Seller reports available")}</span><p>{rtl?(getPersianProductCopy(quick.id)?.description??quick.description):quick.description}</p><div className="quick-price"><b>{formatPrice(quick)}</b>{!quick.isLive&&quick.priceKind==="estimated"&&<small>{usd(quick.price)}</small>}</div><button className="button-primary full-button" onClick={()=>{addToCart(quick);setQuick(null);}}>{t.add}<Plus size={17}/></button></>}</div></div></div>}

      {compareIds.length>0&&<div className="compare-tray"><div><GitCompareArrows size={17}/><b>{rtl?"مقایسه محصولات":"Product comparison"}</b><span>{compareIds.length}/4</span></div><button className="compare-tray-open" disabled={compareIds.length<2} onClick={()=>setCompareOpen(true)}>{rtl?"مقایسه":"Compare"} <ArrowLeft size={15}/></button><button className="compare-tray-clear" onClick={()=>setCompareIds([])}>{rtl?"پاک کردن":"Clear"}</button></div>}

      {compareOpen&&<div className="overlay modal-overlay compare-overlay" onMouseDown={e=>{if(e.target===e.currentTarget)setCompareOpen(false);}}><section className="compare-modal" role="dialog" aria-modal="true" aria-labelledby="compare-dialog-title" tabIndex={-1}><div className="compare-modal-head"><div><span className="section-kicker">GLOBALCART / SIDE BY SIDE</span><h2 id="compare-dialog-title">{t.compareTitle}</h2></div><button className="icon-button" onClick={()=>setCompareOpen(false)} aria-label={t.close}><X/></button></div><div className="compare-table-wrap"><table className="compare-table"><thead><tr><th>{rtl?"مشخصات":"Details"}</th>{compareProducts.map(p=><th key={p.id}><img src={catalogImageSrc(p.thumbnail)} alt="" loading="lazy" decoding="async"/><b>{rtl?(getPersianProductCopy(p.id)?.title??p.title):p.title}</b>{rtl&&getPersianProductCopy(p.id)&&<small className="compare-source-title" dir="auto">{p.title}</small>}<button onClick={()=>toggleCompare(p.id)}>{rtl?"حذف":"Remove"}</button></th>)}</tr></thead><tbody><tr><th>{rtl?"قیمت فعلی":"Current price"}</th>{compareProducts.map(p=><td key={p.id}>{formatPrice(p)}</td>)}</tr><tr><th>{rtl?"قیمت اصلی":"Original price"}</th>{compareProducts.map(p=><td key={p.id}>{!hasDisplayPrice(p)?(rtl?"ثبت نشده":"Not provided"):rtl&&rate!==null?money(calculateDemoPrice({sourcePrice:p.price,fxRate:rate}).toman,locale)+" تومان":usd(p.price)}</td>)}</tr><tr><th>{rtl?"امتیاز":"Rating"}</th>{compareProducts.map(p=><td key={p.id}>{p.ratingAvailable===false?(rtl?"ثبت نشده":"Not provided"):<span className="compare-rating"><Star size={13} fill="currentColor"/>{p.rating.toFixed(1)}</span>}</td>)}</tr><tr><th>{rtl?"تخفیف":"Discount"}</th>{compareProducts.map(p=><td key={p.id}>{!hasDisplayPrice(p)?(rtl?"ثبت نشده":"Not provided"):Math.round(p.discountPercentage)+"%"}</td>)}</tr><tr><th>{rtl?"برند":"Brand"}</th>{compareProducts.map(p=><td key={p.id}>{p.brand||"—"}</td>)}</tr><tr><th>{rtl?"وضعیت فروشنده":"Seller availability"}</th>{compareProducts.map(p=><td key={p.id}>{p.stockAvailable===false?(rtl?"ثبت نشده":"Not provided"):p.availability===false?(rtl?"ناموجود":"Unavailable"):(rtl?"موجود":"Available")}</td>)}</tr></tbody></table></div><div className="compare-modal-foot"><span>{rtl?"اطلاعات برای مقایسه نمایشی هستند.":"Figures are provided for demo comparison only."}</span><button onClick={()=>{setCompareIds([]);setCompareOpen(false);}}>{t.compareClear}</button></div></section></div>}

            {checkout&&<div className="overlay modal-overlay"><div className="checkout-modal" role="dialog" aria-modal="true" aria-labelledby="checkout-dialog-title" tabIndex={-1}><button className="modal-close icon-button" onClick={()=>setCheckout(false)} aria-label={t.close}><X/></button><span className="section-kicker">GLOBALCART / DEMO FLOW</span><h2 id="checkout-dialog-title">{t.demoCheckout}</h2><p>{t.demoText}</p><div className="checkout-total"><span>{t.subtotal}</span><b>{subtotalLabel}</b></div><div className="checkout-warning"><ShieldCheck size={17}/>{t.noPayment}</div><button className="button-primary full-button" onClick={placeDemoOrder}>{t.confirm}<ArrowLeft size={16}/></button></div></div>}


    </>
  );
}
