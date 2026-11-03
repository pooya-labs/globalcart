"use client";

import { useEffect, useState } from "react";
import "./case-study.css";
import Link from "next/link";
import { ArrowDownLeft, ArrowLeft, ArrowRight, ArrowUpLeft, ArrowUpRight, CheckCircle2, Database, Globe2, Layers3, Search, ShieldCheck, Sparkles, Workflow } from "lucide-react";
import { readLocalePreference, writeLocalePreference, type Locale } from "../../lib/locale";

const content = {
  fa: {
    eyebrow: "PORTFOLIO CASE STUDY · 2026",
    title: "یک بازار جهانی؛ از تجربه کاربری تا زیرساخت داده",
    intro: "GlobalCart یک شبیه‌سازی مستقل از تجربه بازارگاه‌های مدرن است؛ پروژه‌ای برای نمایش مهندسی محصول، طراحی رابط دوزبانه، خط لوله داده و معماری بک‌اند قابل توسعه.",
    explore: "مشاهده فروشگاه", github: "مشاهده کد در GitHub", language: "English",
    stats: [["100", "محصول نمونه"], ["36", "تست مرورگر موفق"], ["2", "زبان و جهت نوشتار"], ["0", "تراکنش واقعی"]],
    storyTitle: "از یک ویترین ساده تا یک سیستم چندلایه",
    story: "هدف فقط ساختن چند کارت محصول نبود. پروژه مرحله‌به‌مرحله از ویترین واکنش‌گرا به کاتالوگ واردشونده، تجربه کشف و مقایسه، قیمت‌گذاری شفاف و هسته بک‌اند محلی گسترش پیدا کرد؛ در حالی که مرزهای نسخه نمایشی روشن باقی ماند.",
    featuresTitle: "تور قابلیت‌ها",
    features: [
      ["01", "کشف و جست‌وجو", "پیشنهادهای جست‌وجو، فیلترهای ترکیبی، مرتب‌سازی، جست‌وجوی ذخیره‌شده و فرمان‌پالت صفحه‌کلید.", "رفتن به ویترین", "/#catalog"],
      ["02", "تجربه محصول", "صفحه مستقل محصول، گالری تصاویر، اطلاعات منبع، محصولات مرتبط و مقایسه هم‌زمان تا چهار محصول.", "دیدن محصولات", "/#catalog"],
      ["03", "داده و قیمت", "کاتالوگ نمونه با آخرین نسخه سالم؛ نرخ دلار TGJU با تبدیل صریح ریال به تومان و برچسب تازگی.", "روش کار داده", "#architecture"],
      ["04", "مسیر خرید نمایشی", "سبد خرید، علاقه‌مندی‌ها، بازدیدهای اخیر و ثبت سفارش آزمایشی؛ بدون پرداخت یا ارسال واقعی.", "بازدیدهای اخیر", "/recently-viewed/"]
    ],
    architectureTitle: "معماری امروز، نه معماری خیالی",
    architectureIntro: "مسیر عمومی روی GitHub Pages یک خروجی ایستا است. Go و PostgreSQL در محیط توسعه و CI اعتبارسنجی می‌شوند؛ سرویس‌های مدیریتی و فروشنده هنوز endpoint عمومی ندارند.",
    source: "منابع و ورودی‌ها", normalize: "نرمال‌سازی و اعتبارسنجی", snapshot: "کاتالوگ ایستا", storefront: "ویترین Next.js", local: "وضعیت محلی مرورگر",
    backend: "Go REST + PostgreSQL 17", backendNote: "زیرساخت محلی؛ API فقط خواندنی و سرویس‌های داخلی",
    ci: "GitHub Actions", ciNote: "تست، اسکن امنیتی، ساخت و انتشار Pages",
    boundaryTitle: "مرزهای محصول و تصمیم‌های مهندسی",
    boundaries: [
      ["داده نمونه", "کاتالوگ عمومی DummyJSON است؛ نبود منبع آنلاین با آخرین snapshot معتبر یا fixture محلی مدیریت می‌شود."],
      ["نرخ ارز", "منبع نرخ دلار آزاد TGJU است. مقدار خام ریال حفظ و برای نمایش به تومان بر ده تقسیم می‌شود؛ نرخ نمایشی، قیمت اجرایی معامله نیست."],
      ["خرید", "سبد، پرداخت، سفارش، مرجوعی و بازپرداخت صرفاً شبیه‌سازی هستند؛ هیچ پول یا کالایی جابه‌جا نمی‌شود."],
      ["تأمین‌کننده", "اتصال زنده به آمازون یا فروشندگان بدون دسترسی رسمی و مجوز داده انجام نمی‌شود؛ scraping در برنامه نیست."],
      ["زیرساخت", "PostgreSQL پایگاه داده اصلی است؛ Redis اختیاری است. میزبانی عمومی API، ایمیل و هویت تا تأیید مالک غیرفعال‌اند."]
    ],
    stackTitle: "پشته فنی",
    stack: ["Next.js 15", "React 19", "TypeScript", "Go", "PostgreSQL 17", "sqlc", "Playwright", "GitHub Actions"],
    nextTitle: "قدم بعدی",
    next: "هسته فنی و تست‌های خودکار آماده‌اند. پیش از انتشار ۱.۰، بررسی دستی روی دستگاه واقعی، صفحه‌خوان، رفتار کامل RTL/LTR و سنجش واقعی CDN و Core Web Vitals باقی مانده است. سپس تصاویر مستند، نمودار ERD و انتشار نسخه نهایی تکمیل می‌شوند.",
    footer: "ساخته و مهندسی‌شده توسط پویا گودرزی · پروژه مستقل پورتفولیو",
    back: "بازگشت به فروشگاه"
  },
  en: {
    eyebrow: "PORTFOLIO CASE STUDY · 2026",
    title: "A global marketplace, from product experience to data infrastructure",
    intro: "GlobalCart is an independent simulation of a modern marketplace: a portfolio project showcasing product engineering, bilingual interface design, data pipelines and an extensible backend architecture.",
    explore: "Explore the storefront", github: "View source on GitHub", language: "فارسی",
    stats: [["100", "sample products"], ["36", "browser tests passed"], ["2", "languages & directions"], ["0", "real transactions"]],
    storyTitle: "From a storefront to a layered system",
    story: "The goal was never just to render product cards. The project evolved into a responsive storefront, catalog ingestion pipeline, discovery and comparison experience, transparent pricing and a local backend foundation—while keeping the demo boundary explicit.",
    featuresTitle: "Feature tour",
    features: [
      ["01", "Discovery & search", "Search suggestions, combined facets, sorting, saved searches and a keyboard command palette.", "Explore the catalog", "/#catalog"],
      ["02", "Product experience", "Dedicated product pages, image galleries, source details, related products and side-by-side comparison for up to four items.", "Browse products", "/#catalog"],
      ["03", "Data & pricing", "A generated sample catalog with last-known-good protection; TGJU USD pricing with explicit Rial-to-Toman normalization.", "How data flows", "#architecture"],
      ["04", "Demo commerce journey", "Cart, favorites, recently viewed and simulated order completion—with no real payment or fulfillment.", "Recently viewed", "/recently-viewed/"]
    ],
    architectureTitle: "The architecture that exists—not an imaginary one",
    architectureIntro: "The public experience is a static GitHub Pages export. Go and PostgreSQL are validated in local development and CI; admin and seller services do not expose public HTTP routes.",
    source: "Sources & inputs", normalize: "Normalize & validate", snapshot: "Static catalog", storefront: "Next.js storefront", local: "Browser-local state",
    backend: "Go REST + PostgreSQL 17", backendNote: "Local foundation; read-only API and internal services",
    ci: "GitHub Actions", ciNote: "Tests, security scans, build and Pages deployment",
    boundaryTitle: "Product boundaries & engineering decisions",
    boundaries: [
      ["Sample data", "The public sample catalog comes from DummyJSON. If the source is unavailable, the last valid snapshot or local fixtures keep the storefront usable."],
      ["Foreign exchange", "TGJU is the USD free-market source. Its raw Rial quote is retained and divided by ten for Toman display; this is not an executable quote."],
      ["Commerce", "Cart, payment, orders, returns and refunds are simulations. No money or goods move."],
      ["Suppliers", "Amazon or seller integrations require authorized access and data rights. Scraping is not part of the plan."],
      ["Infrastructure", "PostgreSQL is the primary database; Redis is optional. Public API hosting, email and identity remain disabled pending owner approval."]
    ],
    stackTitle: "Technology",
    stack: ["Next.js 15", "React 19", "TypeScript", "Go", "PostgreSQL 17", "sqlc", "Playwright", "GitHub Actions"],
    nextTitle: "What comes next",
    next: "The engineering foundation and automated checks are in place. Before the 1.0 release, manual real-device, screen-reader, full RTL/LTR and live CDN/Core Web Vitals reviews remain. Then the final screenshots, ERD and versioned release can be completed.",
    footer: "Designed and engineered by Pooya Goodarzi · Independent portfolio project",
    back: "Back to storefront"
  }
} as const;

export default function CaseStudyPage() {
  const [locale, setLocale] = useState<Locale>("fa");
  useEffect(() => {
    const saved = readLocalePreference();
    if (saved) setLocale(saved);
  }, []);
  const t = content[locale];
  const rtl = locale === "fa";
  const changeLocale = () => {
    const next: Locale = rtl ? "en" : "fa";
    setLocale(next);
    writeLocalePreference(next);
  };

  return (
    <main className={`case-study ${rtl ? "rtl" : "ltr"}`} dir={rtl ? "rtl" : "ltr"} lang={locale}>
      <header className="cs-topbar">
        <Link href="/" className="cs-brand"><span className="cs-mark">g.</span><span>globalcart<small>THE WORLD, WELL FOUND.</small></span></Link>
        <nav aria-label={rtl ? "ناوبری پرونده پروژه" : "Case study navigation"}>
          <Link href="/">{t.back}</Link>
          <button type="button" onClick={changeLocale} aria-label={rtl ? "Switch language to English" : "تغییر زبان به فارسی"}><Globe2 size={16}/>{t.language}</button>
        </nav>
      </header>

      <section className="cs-hero" id="top">
        <div className="cs-hero-copy">
          <span className="cs-eyebrow"><Sparkles size={14}/>{t.eyebrow}</span>
          <h1>{t.title}</h1>
          <p>{t.intro}</p>
          <div className="cs-actions">
            <Link href="/#catalog" className="cs-primary">{t.explore}{rtl ? <ArrowLeft size={16}/> : <ArrowRight size={16}/>}</Link>
            <a href="https://github.com/pooya16pg-spec/globalcart" target="_blank" rel="noreferrer" className="cs-secondary">{t.github}<ArrowUpRight size={15}/></a>
          </div>
        </div>
        <div className="cs-hero-art" role="img" aria-label="GlobalCart product architecture illustration">
          <div className="cs-art-orbit orbit-one"/><div className="cs-art-orbit orbit-two"/>
          <div className="cs-art-core"><span>g.</span><b>GLOBALCART</b><small>DISCOVER · COMPARE · CHOOSE</small></div>
          <div className="cs-art-chip chip-a">RTL ↔ LTR</div><div className="cs-art-chip chip-b">DATA → UX</div><div className="cs-art-chip chip-c">DEMO / SAFE</div>
        </div>
      </section>

      <section className="cs-stats" aria-label="Project metrics">
        {t.stats.map(([value,label]) => <div key={label}><strong>{value}</strong><span>{label}</span></div>)}
      </section>

      <section className="cs-story">
        <span className="cs-kicker">01 / THE IDEA</span>
        <div><h2>{t.storyTitle}</h2><p>{t.story}</p></div>
      </section>

      <section className="cs-features" id="tour">
        <div className="cs-section-heading"><span className="cs-kicker">02 / PRODUCT</span><h2>{t.featuresTitle}</h2></div>
        <div className="cs-feature-grid">
          {t.features.map(([number,title,description,action,href]) => (
            <article className="cs-feature-card" key={number}>
              <span className="cs-feature-number">{number}</span>
              <div className="cs-feature-icon">{number === "01" ? <Search/> : number === "02" ? <Layers3/> : number === "03" ? <Database/> : <ShieldCheck/>}</div>
              <h3>{title}</h3><p>{description}</p>
              <Link href={href}>{action}{rtl ? <ArrowUpLeft size={15}/> : <ArrowUpRight size={15}/>}</Link>
            </article>
          ))}
        </div>
      </section>

      <section className="cs-architecture" id="architecture">
        <div className="cs-section-heading"><span className="cs-kicker">03 / SYSTEM DESIGN</span><h2>{t.architectureTitle}</h2><p>{t.architectureIntro}</p></div>
        <div className="cs-flow" role="group" aria-label={rtl ? "نمودار جریان داده و معماری" : "Data flow and architecture diagram"}>
          <div className="cs-flow-row">
            <div className="cs-flow-node"><span><Database size={18}/></span><b>{t.source}</b><small>DummyJSON · TGJU</small></div>
            <div className="cs-flow-arrow">{rtl ? <ArrowDownLeft/> : <ArrowRight/>}</div>
            <div className="cs-flow-node"><span><Workflow size={18}/></span><b>{t.normalize}</b><small>Typed contracts · validation</small></div>
            <div className="cs-flow-arrow">{rtl ? <ArrowDownLeft/> : <ArrowRight/>}</div>
            <div className="cs-flow-node"><span><Layers3 size={18}/></span><b>{t.snapshot}</b><small>JSON · last-known-good</small></div>
            <div className="cs-flow-arrow">{rtl ? <ArrowDownLeft/> : <ArrowRight/>}</div>
            <div className="cs-flow-node"><span><Globe2 size={18}/></span><b>{t.storefront}</b><small>Static export · Pages</small></div>
            <div className="cs-flow-arrow">{rtl ? <ArrowDownLeft/> : <ArrowRight/>}</div>
            <div className="cs-flow-node"><span><ShieldCheck size={18}/></span><b>{t.local}</b><small>Cart · favorites · history</small></div>
          </div>
          <div className="cs-flow-foundations">
            <div className="cs-foundation"><b>{t.backend}</b><span>{t.backendNote}</span></div>
            <div className="cs-foundation"><b>{t.ci}</b><span>{t.ciNote}</span></div>
          </div>
          <p className="cs-diagram-caption">{rtl ? "نمای ساده‌شده؛ زیرساخت محلی به‌معنای انتشار عمومی API نیست." : "Simplified view. Local backend infrastructure does not imply a publicly deployed API."}</p>
        </div>
      </section>

      <section className="cs-boundaries">
        <div className="cs-section-heading"><span className="cs-kicker">04 / TRUST & ENGINEERING</span><h2>{t.boundaryTitle}</h2></div>
        <div className="cs-boundary-list">
          {t.boundaries.map(([title,description],index) => <article key={title}><span>{String(index+1).padStart(2,"0")}</span><div><h3>{title}</h3><p>{description}</p></div><CheckCircle2 size={18}/></article>)}
        </div>
      </section>

      <section className="cs-stack">
        <div><span className="cs-kicker">05 / TOOLKIT</span><h2>{t.stackTitle}</h2></div>
        <div className="cs-stack-list">{t.stack.map(item=><span key={item}>{item}</span>)}</div>
      </section>

      <section className="cs-next">
        <span className="cs-kicker">06 / ROAD AHEAD</span><h2>{t.nextTitle}</h2><p>{t.next}</p>
        <Link href="/">{t.back}{rtl ? <ArrowLeft size={16}/> : <ArrowRight size={16}/>}</Link>
      </section>

      <footer className="cs-footer"><Link href="/" className="cs-brand"><span className="cs-mark">g.</span><span>globalcart</span></Link><p>{t.footer}</p><a href="https://github.com/pooya16pg-spec/globalcart" target="_blank" rel="noreferrer">GitHub <ArrowUpRight size={14}/></a></footer>
    </main>
  );
}
