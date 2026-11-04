import type { CatalogProduct } from "./types";

export const ABO_SOURCE = "Amazon Berkeley Objects dataset (Amazon.com & UC Berkeley; CC BY 4.0)";
export const ABO_DATASET_URL = "https://amazon-berkeley-objects.s3.amazonaws.com/index.html";

type RecordLike = Record<string, unknown>;

function isRecord(value: unknown): value is RecordLike {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function localizedText(value: unknown): string {
  if (typeof value === "string") return value.trim();
  if (Array.isArray(value)) {
    const records = value.filter(isRecord);
    const preferred = records.find(item => item.language_tag === "en_US")
      ?? records.find(item => typeof item.language_tag === "string" && String(item.language_tag).startsWith("en_"))
      ?? records[0];
    return preferred ? localizedText(preferred.value ?? preferred.text ?? preferred.name) : "";
  }
  if (isRecord(value)) return localizedText(value.value ?? value.text ?? value.name);
  return "";
}

function flattenText(value: unknown): string {
  if (typeof value === "string" || typeof value === "number") return String(value);
  if (Array.isArray(value)) return value.map(flattenText).filter(Boolean).join(" ");
  if (isRecord(value)) {
    return Object.entries(value)
      .filter(([key]) => /name|path|value|label|title|department|category|node/i.test(key))
      .map(([, item]) => flattenText(item))
      .filter(Boolean)
      .join(" ");
  }
  return "";
}

function parseRawListing(row: RecordLike): RecordLike {
  const raw = row.raw_listing_json;
  if (isRecord(raw)) return raw;
  if (typeof raw === "string") {
    try {
      const parsed: unknown = JSON.parse(raw);
      if (isRecord(parsed)) return parsed;
    } catch {
      // The normalized columns remain usable even when a raw payload is malformed.
    }
  }
  return {};
}

export function classifyAboCategory(input: RecordLike): string {
  const raw = parseRawListing(input);
  const text = [
    input.department, input.product_type, input.product_type_readable,
    input.hierarchy_path, input.node_paths, input.node, input.title,
    input.item_name, raw.department, raw.product_type, raw.product_type_name,
    raw.node, raw.node_path, raw.node_paths, raw.item_name
  ].map(flattenText).join(" ").toLowerCase().replace(/[_/.-]+/g, " ");

  const rules: Array<[string, RegExp]> = [
    ["mobile-accessories", /phone case|mobile case|screen protector|power bank|phone charger|wall charger|wireless charger|charging cable|usb c cable|phone stand|phone holder/],
    ["smartphones", /smartphone|mobile phone|cell phone|iphone|ipad|galaxy (s|z|a|note)|android phone|tablet computer/],
    ["laptops", /laptop|notebook computer|macbook|chromebook|thinkpad|desktop computer|computer workstation/],
    ["mens-watches", /wristwatch|smartwatch|watch band|watch strap|timepiece|watches/],
    ["fragrances", /fragrance|perfume|eau de parfum|eau de toilette|cologne/],
    ["beauty", /beauty|personal care|cosmetic|makeup|skin care|skincare|hair care|shampoo|conditioner|moisturizer|sunscreen/],
    ["furniture", /furniture|sofa|couch|armchair|recliner|bed frame|dining table|coffee table|bookcase|bookshelf|mattress/],
    ["home-decor", /home decor|home decoration|wall art|wall decor|rug|carpet|curtain|candle holder|vase|decorative/],
    ["kitchen", /kitchen|cookware|bakeware|cutlery|flatware|dinnerware|kitchen appliance|coffee maker|toaster|blender/],
    ["sports-outdoors", /sport|fitness|exercise|workout|camping|hiking|outdoor|cycling|bicycle|yoga|running gear|athletic/],
    ["shoes", /footwear|shoe|sneaker|boot|sandal|slipper/],
    ["bags", /handbag|backpack|luggage|suitcase|wallet|purse|duffel bag|tote bag/],
    ["jewelry", /jewelry|jewellery|necklace|bracelet|earring|ring|brooch/],
    ["fashion", /clothing|apparel|garment|dress|shirt|t shirt|jeans|trouser|pants|jacket|coat|sweater|hoodie|skirt/],
    ["toys", /toy|toys|board game|puzzle|doll|action figure|playset|stuffed animal/],
    ["books", /book|books|novel|textbook|paperback|hardcover|magazine/],
    ["automotive", /automotive|auto parts|car accessory|vehicle|motorcycle|car care/],
    ["tools", /tool|hardware|power tool|hand tool|drill|screwdriver|wrench/],
    ["baby", /baby|infant|toddler|nursery/],
    ["pets", /pet supplies|pet food|dog|cat|aquarium|bird cage/],
    ["office", /office supplies|stationery|printer|paper products|desk organizer/],
    ["music", /musical instrument|guitar|keyboard instrument|drum|microphone|music equipment/],
    ["electronics", /electronics|headphone|earbud|speaker|television|monitor|camera|audio|router|computer accessory|projector|gaming console/]
  ];

  for (const [category, pattern] of rules) {
    if (pattern.test(text)) return category;
  }
  return "other-products";
}

function safeAboImageUrl(row: RecordLike, raw: RecordLike): string {
  const candidates = [row.main_image_url, raw.main_image_url, row.image_url, raw.image_url];
  for (const candidate of candidates) {
    if (typeof candidate !== "string") continue;
    try {
      const url = new URL(candidate);
      if (url.protocol === "https:" && (
        url.hostname === "amazon-berkeley-objects.s3.amazonaws.com" ||
        url.hostname === "amazon-berkeley-objects.s3.us-east-1.amazonaws.com"
      )) return url.toString();
    } catch {
      // Continue to the path-based fallback.
    }
  }

  let path = localizedText(row.main_image_path ?? raw.main_image_path);
  if (!path) {
    const imageId = localizedText(row.main_image_id ?? raw.main_image_id);
    if (imageId) path = /\.(?:jpg|jpeg|png|webp)$/i.test(imageId) ? imageId : imageId + ".jpg";
  }
  if (!path) return "";
  if (/^https:\/\//i.test(path)) {
    try {
      const url = new URL(path);
      if (url.protocol === "https:" && (
        url.hostname === "amazon-berkeley-objects.s3.amazonaws.com" ||
        url.hostname === "amazon-berkeley-objects.s3.us-east-1.amazonaws.com"
      )) return url.toString();
    } catch {
      return "";
    }
    return "";
  }

  const normalized = path
    .replace(/^\/+/, "")
    .replace(/^(?:images\/)?small\//i, "")
    .replace(/^images\/small\//i, "");
  if (!normalized || normalized.split("/").some(part => part === "..")) return "";
  return "https://amazon-berkeley-objects.s3.amazonaws.com/images/small/" + normalized;
}

export function stableAboProductId(sourceId: string): number {
  let hash = 2166136261;
  for (let index = 0; index < sourceId.length; index += 1) {
    hash ^= sourceId.charCodeAt(index);
    hash = Math.imul(hash, 16777619);
  }
  return ((hash >>> 0) & 0x7fffffff) || 1;
}

export type AboMappedProduct = CatalogProduct & { imageUrl: string; imageIds: string[] };

function normalizeImageIds(primary: unknown, additional: unknown): string[] {
  const values = [
    ...(typeof primary === "string" ? [primary] : Array.isArray(primary) ? primary : []),
    ...(Array.isArray(additional) ? additional : typeof additional === "string" ? [additional] : [])
  ];
  return [...new Set(values
    .filter((value): value is string => typeof value === "string")
    .map(value => value.trim())
    .filter(value => /^[A-Za-z0-9+/_-]{8,20}$/.test(value)))];
}

export function mapAboListing(value: unknown, importedAt = new Date().toISOString()): AboMappedProduct | null {
  if (!isRecord(value)) return null;
  const raw = parseRawListing(value);
  const sourceId = localizedText(value.item_id ?? value.asin ?? raw.item_id ?? raw.asin).toUpperCase();
  const title = localizedText(value.title ?? value.item_name ?? raw.item_name ?? raw.title);
  const imageUrl = safeAboImageUrl(value, raw);
  const imageIds = normalizeImageIds(
    value.main_image_id ?? raw.main_image_id,
    value.other_image_id ?? raw.other_image_id
  );

  if (!/^[A-Z0-9]{10}$/.test(sourceId) || title.length < 3 || (!imageUrl && imageIds.length === 0)) return null;

  const brand = localizedText(value.brand ?? raw.brand) || "Unbranded";
  const description = localizedText(value.product_description ?? value.description ?? raw.product_description ?? raw.description)
    || [value.bullet_point, raw.bullet_point].map(localizedText).filter(Boolean).join(" · ")
    || title;
  // Product pages are intentionally presented as GlobalCart listings. Source
  // attribution lives in the site-wide data credits, not as outbound product links.
  const sourceUrl = "";

  return {
    id: stableAboProductId(sourceId),
    title,
    description: description.slice(0, 2400),
    category: classifyAboCategory(value),
    price: 0,
    discountPercentage: 0,
    rating: 0,
    stock: 0,
    brand,
    thumbnail: imageUrl,
    images: [],
    source: ABO_SOURCE,
    sourceId,
    sourceUrl,
    importedAt,
    priceAvailable: false,
    ratingAvailable: false,
    stockAvailable: false,
    imageUrl,
    imageIds
  };
}
