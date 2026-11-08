"use client";

import { useEffect, useState } from "react";
import "./case-study.css";
import Link from "next/link";
import { ArrowDownLeft, ArrowLeft, ArrowRight, ArrowUpLeft, ArrowUpRight, CheckCircle2, Database, Globe2, Layers3, Search, ShieldCheck, Sparkles, Workflow } from "lucide-react";
import { readLocalePreference, writeLocalePreference, type Locale } from "../../lib/locale";

const content = {
  fa: {
    eyebrow: "PORTFOLIO CASE STUDY · POOYA LABS · 2026",
    title: "یک بازار جهانی؛ از تجربه کاربری تا زیرساخت داده",
    intro: "GlobalCart یک شبیه‌سازی مستقل و کامل از تجربه یک بازارگاه مدرن است؛ مطالعه‌ای عملی از طراحی محصول دوزبانه، کاتالوگ آفلاین، تجربه کشف و مقایسه، شفافیت قیمت و زیرساخت مهندسی. محصول و نگه‌داری آن متعلق به Pooya Labs است و نسخهٔ پایدار نمایشی ۱.۰.۰ آن منتشر شده است.",
    explore: "مشاهده فروشگاه", github: "مشاهده کد در GitHub", language: "English",
    stats: [["1,000", "محصول نمونه"], ["24", "دسته‌بندی کالا"], ["2", "زبان و جهت نوشتار"], ["0", "تراکنش واقعی"]],
    storyTitle: "از یک ویترین ساده تا یک سیستم چندلایه",
    story: "هدف فقط ساختن چند کارت محصول نبود. پروژه مرحله‌به‌مرحله از ویترین واکنش‌گرا به کاتالوگ واردشونده، تجربه کشف و مقایسه، قیمت‌گذاری شفاف و هسته بک‌اند محلی گسترش پیدا کرد؛ در حالی که مرزهای نسخه نمایشی روشن باقی ماند.",
    featuresTitle: "تور قابلیت‌ها",
    features: [
      ["01", "کشف و جست‌وجو", "پیشنهادهای جست‌وجو، فیلترهای ترکیبی، مرتب‌سازی، جست‌وجوی ذخیره‌شده و فرمان‌پالت صفحه‌کلید.", "رفتن به ویترین", "/#catalog"],
      ["02", "تجربه محصول", "صفحه مستقل محصول، گالری قابل پیمایش تصاویر، محصولات مرتبط و مقایسه هم‌زمان تا چهار محصول.", "دیدن محصولات", "/#catalog"],
      ["03", "داده و قیمت", "کاتالوگ ۱۰۰۰ محصولی از مجموعه‌دادهٔ Amazon Berkeley Objects با مجوز CC BY 4.0؛ برآوردهای نمایشی قیمت به‌صورت جداگانه و با برچسب تخمینی نمایش داده می‌شوند. TGJU فقط تبدیل دلار به تومان و تازگی نرخ ارز را فراهم می‌کند.", "روش کار داده", "#architecture"],
      ["04", "مسیر خرید نمایشی", "سبد خرید، علاقه‌مندی‌ها، بازدیدهای اخیر و ثبت سفارش آزمایشی؛ بدون پرداخت یا ارسال واقعی.", "بازدیدهای اخیر", "/recently-viewed/"]
    ],
    architectureTitle: "معماری امروز، نه معماری خیالی",
    architectureIntro: "مسیر عمومی روی GitHub Pages یک خروجی ایستا است. Go و PostgreSQL در محیط توسعه و CI اعتبارسنجی می‌شوند؛ سرویس‌های مدیریتی و فروشنده هنوز endpoint عمومی ندارند.",
    databaseTitle: "مدل داده و روابط پایگاه‌داده",
    databaseIntro: "مدل ERD بر اساس ۱۵ مهاجرت نسخه‌بندی‌شده PostgreSQL استخراج شده است: ۲۷ جدول در چهار حوزه، همراه با روابط کلید خارجی و محدودیت‌های مهم.",
    databaseGroups: [
      ["07", "کاتالوگ و بازارگاه", "products · categories · variants · vendors · offers"],
      ["10", "تجارت نمایشی", "carts · orders · reservations · coupons · returns"],
      ["07", "هویت و دسترسی", "users · roles · seller memberships · sessions"],
      ["03", "مدیریت و تاریخچه", "admin overrides · product revisions · seller revisions"]
    ],
    databaseLink: "مشاهده نمودار کامل ERD در GitHub",
    source: "منابع و ورودی‌ها", normalize: "نرمال‌سازی و اعتبارسنجی", snapshot: "کاتالوگ ایستا", storefront: "ویترین Next.js", local: "وضعیت محلی مرورگر",
    backend: "Go REST + PostgreSQL 17", backendNote: "زیرساخت محلی؛ API فقط خواندنی و سرویس‌های داخلی",
    ci: "GitHub Actions", ciNote: "تست، اسکن امنیتی، ساخت و انتشار Pages",
    boundaryTitle: "مرزهای محصول و تصمیم‌های مهندسی",
    boundaries: [
      ["داده نمونه و برآورد قیمت", "مجموعه‌دادهٔ Amazon Berkeley Objects قیمت فروش، امتیاز یا موجودی فروشنده را ارائه نمی‌کند. GlobalCart برای تجربهٔ نمایشی، برآوردهای ثابت و دسته‌محور می‌سازد و آن‌ها را با وضعیت تخمینی از دادهٔ منبع جدا نگه می‌دارد؛ این ارقام قیمت فروشنده نیستند و قیمت منبع همچنان ناموجود است."],
      ["نرخ ارز", "منبع نرخ دلار آزاد TGJU است. مقدار خام ریال حفظ و برای نمایش به تومان بر ده تقسیم می‌شود؛ نرخ نمایشی، قیمت اجرایی معامله نیست."],
      ["خرید", "سبد، پرداخت، سفارش، مرجوعی و بازپرداخت صرفاً شبیه‌سازی هستند؛ هیچ پول یا کالایی جابه‌جا نمی‌شود."],
      ["تأمین‌کننده", "اتصال زنده به آمازون یا فروشندگان بدون دسترسی رسمی و مجوز داده انجام نمی‌شود؛ scraping در برنامه نیست."],
      ["زیرساخت", "PostgreSQL پایگاه داده اصلی است؛ Redis اختیاری است. میزبانی عمومی API، ایمیل و هویت تا تأیید مالک غیرفعال‌اند."]
    ],
    stackTitle: "پشته فنی",
    stack: ["Next.js 15", "React 19", "TypeScript", "Go", "PostgreSQL 17", "sqlc", "Playwright", "GitHub Actions"],
    nextTitle: "انتشار پایدار و کیفیت نهایی",
    next: "نسخهٔ پایدار نمایشی ۱.۰.۰ در ۸ نوامبر ۲۰۲۶ منتشر شد و محدودهٔ نسخهٔ اول بسته شده است. آزمون‌های مرورگر روی Chromium، Firefox و WebKit، ۱۶۳ آزمون موفق، ۹ مورد کنارگذاری عمدی و صفر شکست ثبت کردند. هر شش گزارش Lighthouse بودجه‌های بدون تغییر را پاس کردند: امتیاز کارایی موبایل برای خانه، محصول و کیس‌استادی ۹۶، ۹۶ و ۹۹؛ امتیاز دسکتاپ هر سه مسیر ۱۰۰ بود. دسترس‌پذیری، بهترین‌روش‌ها و SEO در هر شش گزارش ۱۰۰ و CLS صفر ثبت شد. اسکن امنیتی، استقرار Pages، بازبینی تصویری زنده و کنترل کاتالوگ نیز موفق بودند. مالک، بازبینی روی دستگاه‌های واقعی دسکتاپ و موبایل، صفحه‌خوان واقعی و بررسی کامل فارسی RTL / انگلیسی LTR را تأیید کرده است. بهینه‌سازی بیشتر یک قطعهٔ مشترک JavaScript و داده‌های میدانی Core Web Vitals پیگیری‌های اختیاری آینده‌اند؛ دادهٔ میدانی هنوز جمع‌آوری نشده و هیچ‌کدام مانع انتشار نسخهٔ نمایشی نیستند.",
    footer: "محصول و نگه‌داری: Pooya Labs · طراحی و مهندسی: پویا گودرزی",
    back: "بازگشت به فروشگاه"
  },
  en: {
    eyebrow: "PORTFOLIO CASE STUDY · POOYA LABS · 2026",
    title: "A global marketplace, from product experience to data infrastructure",
    intro: "GlobalCart is an independent, end-to-end marketplace simulation and a practical study in bilingual product design, an offline-first catalog, discovery and comparison, transparent pricing, and production-minded engineering. Owned and maintained by Pooya Labs, its stable 1.0.0 portfolio release is now published.",
    explore: "Explore the storefront", github: "View source on GitHub", language: "فارسی",
    stats: [["1,000", "sample products"], ["24", "product categories"], ["2", "languages & directions"], ["0", "real transactions"]],
    storyTitle: "From a storefront to a layered system",
    story: "The goal was never just to render product cards. The project evolved into a responsive storefront, catalog ingestion pipeline, discovery and comparison experience, transparent pricing and a local backend foundation—while keeping the demo boundary explicit.",
    featuresTitle: "Feature tour",
    features: [
      ["01", "Discovery & search", "Search suggestions, combined facets, sorting, saved searches and a keyboard command palette.", "Explore the catalog", "/#catalog"],
      ["02", "Product experience", "Dedicated product pages, browsable multi-image galleries, related products and side-by-side comparison for up to four items.", "Browse products", "/#catalog"],
      ["03", "Data & pricing", "A 1,000-product catalog built from the CC BY 4.0 Amazon Berkeley Objects dataset. Clearly labelled, deterministic demo estimates are kept separate from unavailable source prices; TGJU supplies USD-to-Toman conversion and FX freshness only.", "How data flows", "#architecture"],
      ["04", "Demo commerce journey", "Cart, favorites, recently viewed and simulated order completion—with no real payment or fulfillment.", "Recently viewed", "/recently-viewed/"]
    ],
    architectureTitle: "The architecture that exists—not an imaginary one",
    architectureIntro: "The public experience is a static GitHub Pages export. Go and PostgreSQL are validated in local development and CI; admin and seller services do not expose public HTTP routes.",
    databaseTitle: "Database model and relationships",
    databaseIntro: "The ERD is derived from the 15 versioned PostgreSQL migrations: 27 tables across four domains, including declared foreign keys and key integrity rules.",
    databaseGroups: [
      ["07", "Catalog & marketplace", "products · categories · variants · vendors · offers"],
      ["10", "Demo commerce", "carts · orders · reservations · coupons · returns"],
      ["07", "Identity & access", "users · roles · seller memberships · sessions"],
      ["03", "Administration & history", "admin overrides · product revisions · seller revisions"]
    ],
    databaseLink: "Open the full PostgreSQL ERD on GitHub",
    source: "Sources & inputs", normalize: "Normalize & validate", snapshot: "Static catalog", storefront: "Next.js storefront", local: "Browser-local state",
    backend: "Go REST + PostgreSQL 17", backendNote: "Local foundation; read-only API and internal services",
    ci: "GitHub Actions", ciNote: "Tests, security scans, build and Pages deployment",
    boundaryTitle: "Product boundaries & engineering decisions",
    boundaries: [
      ["Sample data & estimates", "The Amazon Berkeley Objects dataset does not supply seller sale prices, ratings or stock. GlobalCart generates stable category-bounded demo estimates and labels them as estimates, separately from source provenance; no estimate is presented as a verified seller price."],
      ["Foreign exchange", "TGJU is the USD free-market source. Its raw Rial quote is retained and divided by ten for Toman display; this is not an executable quote."],
      ["Commerce", "Cart, payment, orders, returns and refunds are simulations. No money or goods move."],
      ["Suppliers", "Amazon or seller integrations require authorized access and data rights. Scraping is not part of the plan."],
      ["Infrastructure", "PostgreSQL is the primary database; Redis is optional. Public API hosting, email and identity remain disabled pending owner approval."]
    ],
    stackTitle: "Technology",
    stack: ["Next.js 15", "React 19", "TypeScript", "Go", "PostgreSQL 17", "sqlc", "Playwright", "GitHub Actions"],
    nextTitle: "Stable release & quality closeout",
    next: "The stable 1.0.0 portfolio/demo release was published on November 8, 2026, and the v1.0 scope is closed. Cross-browser runs on Chromium, Firefox and WebKit recorded 163 passed tests, 9 intentional skips and zero failures. All six Lighthouse reports passed the unchanged budgets: mobile Performance scored 96, 96 and 99 for the home, product and case-study routes; desktop Performance was 100 on all three. Accessibility, Best Practices and SEO scored 100 in every report, with CLS at zero throughout. Security scanning, Pages deployment, live visual QA and catalog smoke checks also passed. The owner confirmed real desktop/mobile review, an actual platform screen-reader review and a full human RTL/LTR pass. Further shared-chunk JavaScript reduction and field Core Web Vitals collection are optional follow-ups; field data has not been collected, and neither item blocks this static portfolio/demo release.
    footer: "Owned and maintained by Pooya Labs · Designed and engineered by Pooya Goodarzi",
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
            <a href="https://github.com/pooya-labs/globalcart" target="_blank" rel="noreferrer" className="cs-secondary">{t.github}<ArrowUpRight size={15}/></a>
          </div>
        </div>
        <div className="cs-hero-art">
          <div className="cs-art-orbit orbit-one" aria-hidden="true"/><div className="cs-art-orbit orbit-two" aria-hidden="true"/>
          <a className="cs-art-core" href="https://pooya-labs.github.io/pooya-labs-platform/" target="_blank" rel="noopener noreferrer">
            <img src="../images/site/pooya-monogram.webp" alt="" aria-hidden="true" width="76" height="113"/>
            <b>GLOBALCART</b><small>OWNED BY POOYA LABS</small>
          </a>
          <div className="cs-art-chip chip-a" aria-hidden="true">RTL ↔ LTR</div><div className="cs-art-chip chip-b" aria-hidden="true">DATA → UX</div><div className="cs-art-chip chip-c" aria-hidden="true">DEMO / SAFE</div>
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
            <div className="cs-flow-node"><span><Database size={18}/></span><b>{t.source}</b><small>ABO dataset · TGJU</small></div>
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

      <section className="cs-database" id="database">
        <div className="cs-section-heading">
          <span className="cs-kicker">04 / DATA MODEL</span>
          <h2>{t.databaseTitle}</h2>
          <p>{t.databaseIntro}</p>
        </div>
        <div className="cs-database-grid">
          {t.databaseGroups.map(([count,title,description]) => (
            <article className="cs-database-card" key={title}>
              <strong>{count}</strong><h3>{title}</h3><p>{description}</p>
            </article>
          ))}
        </div>
        <a className="cs-database-link" href="https://github.com/pooya-labs/globalcart/blob/main/docs/DATABASE_ERD.md" target="_blank" rel="noreferrer">
          {t.databaseLink}<ArrowUpRight size={16}/>
        </a>
      </section>

      <section className="cs-boundaries">
        <div className="cs-section-heading"><span className="cs-kicker">05 / TRUST & ENGINEERING</span><h2>{t.boundaryTitle}</h2></div>
        <div className="cs-boundary-list">
          {t.boundaries.map(([title,description],index) => <article key={title}><span>{String(index+1).padStart(2,"0")}</span><div><h3>{title}</h3><p>{description}</p></div><CheckCircle2 size={18}/></article>)}
        </div>
      </section>

      <section className="cs-stack">
        <div><span className="cs-kicker">06 / TOOLKIT</span><h2>{t.stackTitle}</h2></div>
        <div className="cs-stack-list">{t.stack.map(item=><span key={item}>{item}</span>)}</div>
      </section>

      <section className="cs-next">
        <span className="cs-kicker">{rtl ? "07 / انتشار و کیفیت" : "07 / RELEASE & QUALITY"}</span><h2>{t.nextTitle}</h2><p>{t.next}</p>
        <Link href="/">{t.back}{rtl ? <ArrowLeft size={16}/> : <ArrowRight size={16}/>}</Link>
      </section>

      <footer className="cs-footer"><Link href="/" className="cs-brand"><span className="cs-mark">g.</span><span>globalcart</span></Link><p>{t.footer}</p><a href="https://github.com/pooya-labs/globalcart" target="_blank" rel="noreferrer">GitHub <ArrowUpRight size={14}/></a></footer>
    </main>
  );
}
