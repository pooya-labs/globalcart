const PROXY_URL = "https://globalcart-catalog.pooyagame141.workers.dev/api/catalog";

type ShopifyMoney = { amount?: number; currency?: string };
type ShopifyVariant = {
  price?: ShopifyMoney;
  availability?: { available?: boolean };
  seller?: { name?: string; url?: string };
};
type ShopifyRawProduct = {
  id?: string;
  title?: string;
  description?: { html?: string; plain?: string } | string;
  categories?: Array<{ value?: string; taxonomy?: string }>;
  category?: string | { name?: string; value?: string; path?: string };
  product_type?: string;
  metadata?: Record<string, unknown>;
  price_range?: { min?: ShopifyMoney; max?: ShopifyMoney };
  media?: Array<{ type?: string; url?: string; alt_text?: string; preview?: { url?: string } }>;
  image?: string | { url?: string; src?: string };
  images?: Array<string | { url?: string; src?: string }>;
  variants?: ShopifyVariant[];
  rating?: { value?: number };
  brand?: string | { name?: string };
  url?: string;
};

export type ShopifyCatalogProduct = {
  id: number;
  title: string;
  description: string;
  category: string;
  price: number;
  priceMax: number;
  currency: string;
  discountPercentage: number;
  rating: number;
  stock: number;
  brand: string;
  thumbnail: string;
  images: string[];
  source: string;
  sellerName: string;
  sellerUrl: string;
  sourceUrl: string;
  availability: boolean;
  isLive: true;
};

function safeHttpsUrl(value: unknown): string {
  if (typeof value !== "string") return "";
  try {
    const url = new URL(value);
    return url.protocol === "https:" ? url.toString() : "";
  } catch {
    return "";
  }
}

function minorUnitValue(money: ShopifyMoney | undefined, fallbackCurrency: string): number {
  if (!money || typeof money.amount !== "number" || !Number.isFinite(money.amount)) return 0;
  const currency = typeof money.currency === "string" && /^[A-Z]{3}$/.test(money.currency)
    ? money.currency
    : fallbackCurrency;
  let digits = 2;
  try {
    digits = new Intl.NumberFormat("en", { style: "currency", currency }).resolvedOptions().maximumFractionDigits ?? 2;
  } catch {
    digits = 2;
  }
  return money.amount / (10 ** digits);
}

function stableNumericId(value: string): number {
  let hash = 2166136261;
  for (let index = 0; index < value.length; index += 1) {
    hash ^= value.charCodeAt(index);
    hash = Math.imul(hash, 16777619);
  }
  return (hash >>> 0) || 1;
}

function plainText(value: ShopifyRawProduct["description"]): string {
  const raw = typeof value === "string"
    ? value
    : value?.plain || value?.html || "";
  return raw
    .replace(/<script[\s\S]*?<\/script>/gi, " ")
    .replace(/<style[\s\S]*?<\/style>/gi, " ")
    .replace(/<[^>]+>/g, " ")
    .replace(/&nbsp;/gi, " ")
    .replace(/&amp;/gi, "&")
    .replace(/&quot;/gi, '"')
    .replace(/&#39;/gi, "'")
    .replace(/\s+/g, " ")
    .trim();
}

const CATEGORY_RULES: Array<{ id: string; pattern: RegExp }> = [
  // Accessory-specific rules must run before general phone rules.
  { id: "mobile-accessories", pattern: /\b(phone case|mobile case|screen protector|power bank|wall charger|phone charger|wireless charger|usb[- ]c cable|charging cable|phone adapter|phone stand)\b/i },
  { id: "smartphones", pattern: /\b(smartphones?|mobile phones?|cell phones?|iphone|pixel phones?|galaxy (?:s|z|a|note|fold|flip)\s?\d+|android phones?)\b/i },
  { id: "laptops", pattern: /\b(laptops?|notebooks?|macbooks?|chromebooks?|thinkpads?|gaming pc)\b/i },
  { id: "mens-watches", pattern: /\b(watches|watch|smartwatch|wristwatch|apple watch|galaxy watch)\b/i },
  { id: "fragrances", pattern: /\b(perfumes?|fragrances?|eau de parfum|eau de toilette|colognes?)\b/i },
  { id: "beauty", pattern: /\b(beauty|personal care|skin ?care|makeup|cosmetics?|moisturizers?|serums?|shampoo|conditioner|sunscreen|hair care|body lotion)\b/i },
  { id: "furniture", pattern: /\b(furniture|home decor|sofas?|couches|armchairs|dining tables?|coffee tables?|desks|bookshelves|bed frames?|nightstands?|floor lamps?|lighting)\b/i },
  { id: "groceries", pattern: /\b(groceries|grocery|coffee beans|tea bags|snacks?|chocolate|olive oil|protein bars?|food|beverages?|honey|cereal|drink)\b/i },
  { id: "electronics", pattern: /\b(headphones?|headsets?|earbuds?|earphones?|speakers?|soundbars?|audio|bluetooth|televisions?|\btv\b|cameras?|monitors?|keyboards?|computer mice|routers?|game consoles?|electronics?|tablets?|projectors?|drones?|vacuum cleaners?|appliances?)\b/i },
  { id: "sports-outdoors", pattern: /\b(sports?|sporting goods|athletic|running shoes?|trail shoes?|hiking|camping|fitness|yoga|bicycles?|cycling|outdoors?|tennis|basketball|workout)\b/i },
  { id: "fashion", pattern: /\b(clothing|apparel|fashion|shirts?|t-shirts?|dresses|jeans|jackets?|coats?|sweaters?|hoodies|handbags?|backpacks?|sneakers?|shoes|boots)\b/i },
  { id: "toys", pattern: /\b(toys?|lego|puzzles?|board games?|dolls?|action figures?|plush|stuffed animals?|kids)\b/i }
];

function categoryHints(raw: ShopifyRawProduct): string[] {
  const values: string[] = [];
  for (const item of raw.categories || []) {
    if (typeof item.value === "string") values.push(item.value);
    if (typeof item.taxonomy === "string") values.push(item.taxonomy);
  }
  if (typeof raw.category === "string") values.push(raw.category);
  else if (raw.category && typeof raw.category === "object") {
    values.push(raw.category.name || "", raw.category.value || "", raw.category.path || "");
  }
  if (typeof raw.product_type === "string") values.push(raw.product_type);
  if (raw.metadata && typeof raw.metadata === "object") {
    for (const [key, value] of Object.entries(raw.metadata)) {
      if (/(categor|product.?type|department|tag)/i.test(key) && typeof value === "string") values.push(value);
      if (/(categor|product.?type|department|tag)/i.test(key) && Array.isArray(value)) {
        values.push(...value.filter((item): item is string => typeof item === "string"));
      }
    }
  }
  return values.filter(Boolean);
}

function categorySlug(raw: ShopifyRawProduct): string {
  const text = [raw.title || "", plainText(raw.description), ...categoryHints(raw)].join(" ");
  return CATEGORY_RULES.find(rule => rule.pattern.test(text))?.id || "other-products";
}


/** Map one live Shopify listing without inventing prices, ratings, discounts or imagery. */
export function mapShopifyCatalogProduct(value: unknown): ShopifyCatalogProduct | null {
  if (!value || typeof value !== "object" || Array.isArray(value)) return null;
  const raw = value as ShopifyRawProduct;
  if (typeof raw.id !== "string" || !raw.id.startsWith("gid://shopify/") || typeof raw.title !== "string" || !raw.title.trim()) return null;

  const mediaImages = (raw.media || [])
    .filter(item => {
      const type = item.type?.toLowerCase();
      return !type || type === "image" || type === "mediaimage";
    })
    .map(item => safeHttpsUrl(item.url || item.preview?.url));
  const topLevelImages = [
    raw.image,
    ...(Array.isArray(raw.images) ? raw.images : [])
  ].map(image => {
    if (typeof image === "string") return safeHttpsUrl(image);
    if (image && typeof image === "object") return safeHttpsUrl(image.url || image.src);
    return "";
  });
  const images = [...new Set([...mediaImages, ...topLevelImages].filter(Boolean))];
  if (images.length === 0) return null;

  const minMoney = raw.price_range?.min;
  const maxMoney = raw.price_range?.max || minMoney;
  const currency = typeof minMoney?.currency === "string" && /^[A-Z]{3}$/.test(minMoney.currency)
    ? minMoney.currency
    : "USD";
  const price = minorUnitValue(minMoney, currency);
  const priceMax = Math.max(price, minorUnitValue(maxMoney, currency));
  if (!(price > 0)) return null;

  const variants = Array.isArray(raw.variants) ? raw.variants : [];
  const availableVariants = variants.filter(variant => variant.availability?.available !== false);
  const sellerVariant = [...(availableVariants.length ? availableVariants : variants)]
    .sort((left, right) => minorUnitValue(left.price, currency) - minorUnitValue(right.price, currency))[0];
  const sellerName = sellerVariant?.seller?.name?.trim() || "Independent seller";
  const sellerUrl = safeHttpsUrl(sellerVariant?.seller?.url);
  const sourceUrl = safeHttpsUrl(raw.url) || sellerUrl;
  const rating = typeof raw.rating?.value === "number" && Number.isFinite(raw.rating.value)
    ? Math.min(5, Math.max(0, raw.rating.value))
    : 0;
  const available = variants.length === 0 || availableVariants.length > 0;
  const brand = sellerName;

  return {
    id: stableNumericId(raw.id),
    title: raw.title.trim(),
    description: plainText(raw.description),
    category: categorySlug(raw),
    price,
    priceMax,
    currency,
    discountPercentage: 0,
    rating,
    stock: available ? 1 : 0,
    brand,
    thumbnail: images[0],
    images,
    source: sellerName,
    sellerName,
    sellerUrl,
    sourceUrl,
    availability: available,
    isLive: true
  };
}

export async function searchShopifyCatalog(
  query: string,
  options: { signal?: AbortSignal; limit?: number } = {}
): Promise<ShopifyCatalogProduct[]> {
  const normalizedQuery = query.trim();
  if (normalizedQuery.length < 2 || normalizedQuery.length > 120) {
    throw new Error("Catalog search query must contain 2–120 characters");
  }
  const limit = Math.min(12, Math.max(1, Math.floor(options.limit ?? 8)));
  const response = await fetch(PROXY_URL, {
    method: "POST",
    headers: { "Content-Type": "application/json", "Accept": "application/json" },
    body: JSON.stringify({
      tool: "search_catalog",
      catalog: {
        query: normalizedQuery,
        limit,
        context: { currency: "USD", language: "en" }
      }
    }),
    cache: "no-store",
    signal: options.signal
  });
  const payload = await response.json().catch(() => null) as
    | { products?: unknown[]; error?: string; details?: string }
    | null;
  if (!response.ok) {
    throw new Error(payload?.error || `Live catalog request failed (${response.status})`);
  }
  if (!payload || !Array.isArray(payload.products)) {
    throw new Error("Live catalog returned an invalid product collection");
  }
  const products = payload.products
    .map(mapShopifyCatalogProduct)
    .filter((product): product is ShopifyCatalogProduct => product !== null);
  if (products.length === 0) {
    throw new Error("No live products with usable product images were returned");
  }
  return products;
}
