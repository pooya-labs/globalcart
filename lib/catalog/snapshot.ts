import { randomUUID } from "node:crypto";
import { mkdir, readFile, rename, rm, writeFile } from "node:fs/promises";
import type { CatalogProduct } from "./types";
import { basename, dirname, join } from "node:path";

/**
 * Publishes a complete JSON snapshot with an atomic same-directory rename.
 *
 * Each writer gets a unique, exclusively-created temporary file, so concurrent
 * invocations never write through the same temporary path. The final rename
 * makes readers observe either the previous complete file or the new one.
 */
export async function writeCatalogSnapshotAtomically<T>(
  snapshot: T,
  outputFile: string
): Promise<void> {
  if (typeof outputFile !== "string" || outputFile.trim().length === 0) {
    throw new TypeError("Catalog snapshot output path must be a non-empty string");
  }

  const serialized = JSON.stringify(snapshot, null, 2);
  if (serialized === undefined) {
    throw new TypeError("Catalog snapshot must be JSON-serializable");
  }

  const directory = dirname(outputFile);
  const temporaryFile = join(
    directory,
    `.${basename(outputFile)}.${process.pid}.${randomUUID()}.tmp`
  );

  await mkdir(directory, { recursive: true });
  try {
    await writeFile(temporaryFile, `${serialized}\n`, {
      encoding: "utf8",
      flag: "wx"
    });
    await rename(temporaryFile, outputFile);
  } finally {
    // A failed write/rename must not leave stale temporary files behind.
    await rm(temporaryFile, { force: true }).catch(() => undefined);
  }
}


export type CatalogSnapshot = {
  schemaVersion: 1;
  source: string;
  generatedAt: string;
  productCount: number;
  products: CatalogProduct[];
};

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

const MAX_DB_INTEGER = 2_147_483_647;
const MAX_NUMERIC_12_2 = 9_999_999_999.99;

function hasAtMostTwoDecimalPlaces(value: number): boolean {
  return Math.abs(value * 100 - Math.round(value * 100)) < 1e-7;
}

function isDatabasePrice(value: unknown): value is number {
  return typeof value === "number" &&
    Number.isFinite(value) &&
    value >= 0 &&
    value <= MAX_NUMERIC_12_2 &&
    hasAtMostTwoDecimalPlaces(value);
}

function isDatabaseStock(value: unknown): value is number {
  return typeof value === "number" &&
    Number.isSafeInteger(value) &&
    value >= 0 &&
    value <= MAX_DB_INTEGER;
}

function isCatalogProduct(value: unknown): value is CatalogProduct {
  if (!isRecord(value)) return false;
  const numericFields = [value.discountPercentage, value.rating];
  const stringFields = [
    value.title, value.description, value.category, value.brand, value.thumbnail,
    value.source, value.sourceId, value.sourceUrl, value.importedAt
  ];

  if (
    !Number.isSafeInteger(value.id) ||
    (value.id as number) < 0 ||
    !isDatabasePrice(value.price) ||
    !numericFields.every(field => typeof field === "number" && Number.isFinite(field)) ||
    (value.discountPercentage as number) < 0 ||
    (value.discountPercentage as number) > 100 ||
    !hasAtMostTwoDecimalPlaces(value.discountPercentage as number) ||
    (value.rating as number) < 0 ||
    (value.rating as number) > 5 ||
    !hasAtMostTwoDecimalPlaces(value.rating as number) ||
    !isDatabaseStock(value.stock) ||
    !stringFields.every(field => typeof field === "string") ||
    !(value.title as string).trim() ||
    !(value.thumbnail as string).trim() ||
    !(value.source as string).trim() ||
    !(value.sourceId as string).trim() ||
    !Number.isFinite(Date.parse(value.importedAt as string)) ||
    !Array.isArray(value.images) ||
    !value.images.every(image => typeof image === "string")
  ) return false;

  if (value.variants !== undefined && (
    !Array.isArray(value.variants) || !value.variants.every(variant => {
      if (!isRecord(variant)) return false;
      return typeof variant.sku === "string" &&
        typeof variant.title === "string" &&
        isDatabasePrice(variant.price) &&
        isDatabaseStock(variant.stock) &&
        isRecord(variant.attributes) &&
        typeof variant.source === "string" && variant.source.trim().length > 0 &&
        typeof variant.sourceId === "string" && variant.sourceId.trim().length > 0 &&
        typeof variant.isActive === "boolean";
    })
  )) return false;

  if (value.localization !== undefined) {
    const localization = value.localization;
    if (!isRecord(localization)) return false;
    if (
      localization.locale !== "fa" ||
      (localization.status !== "translated" && localization.status !== "error") ||
      typeof localization.provider !== "string" ||
      typeof localization.sourceHash !== "string" ||
      !(localization.translatedAt === null || typeof localization.translatedAt === "string")
    ) return false;
    if (localization.fields !== null) {
      const fields = localization.fields;
      if (!isRecord(fields) ||
        !["title", "description", "category", "brand"].every(key => typeof fields[key] === "string")) {
        return false;
      }
    }
  }
  return true;
}

function isCatalogSnapshot(value: unknown): value is CatalogSnapshot {
  if (!isRecord(value) ||
    value.schemaVersion !== 1 ||
    typeof value.source !== "string" || value.source.trim().length === 0 ||
    typeof value.generatedAt !== "string" || !Number.isFinite(Date.parse(value.generatedAt)) ||
    !Number.isSafeInteger(value.productCount) || (value.productCount as number) < 0 ||
    !Array.isArray(value.products) ||
    value.products.length !== value.productCount ||
    !value.products.every(isCatalogProduct)
  ) return false;

  const productIDs = new Set<number>();
  const productIdentities = new Set<string>();
  const variantIdentities = new Set<string>();
  const variantSKUs = new Set<string>();

  for (const product of value.products) {
    if (product.source !== value.source) return false;

    const productIdentity = JSON.stringify([product.source, product.sourceId]);
    if (productIDs.has(product.id) || productIdentities.has(productIdentity)) return false;
    productIDs.add(product.id);
    productIdentities.add(productIdentity);

    for (const variant of product.variants ?? []) {
      const identity = JSON.stringify([variant.source, variant.sourceId]);
      if (variantIdentities.has(identity)) return false;
      variantIdentities.add(identity);

      const sku = variant.sku.trim();
      if (sku && variantSKUs.has(sku)) return false;
      if (sku) variantSKUs.add(sku);
    }
  }

  return true;
}

/**
 * Only a genuinely missing file is treated as a first import. Existing
 * unreadable, malformed or structurally invalid snapshots fail closed.
 */
export async function readCatalogSnapshot(outputFile: string): Promise<CatalogSnapshot | null> {
  let content: string;
  try {
    content = await readFile(outputFile, "utf8");
  } catch (error) {
    if (isRecord(error) && error.code === "ENOENT") return null;
    throw error;
  }

  let parsed: unknown;
  try {
    parsed = JSON.parse(content) as unknown;
  } catch {
    throw new Error("Catalog snapshot is not valid JSON: " + outputFile);
  }
  assertCatalogSnapshot(parsed, outputFile);
  return parsed;
}

/** Throws before publication/import whenever a snapshot violates the shared contract. */
export function assertCatalogSnapshot(
  value: unknown,
  context = "input"
): asserts value is CatalogSnapshot {
  if (!isCatalogSnapshot(value)) {
    throw new Error("Catalog snapshot failed schema validation: " + context);
  }
}
