import { createRequire } from "node:module";
import { mkdir, readdir, rename, rm, stat, writeFile } from "node:fs/promises";
import { gunzipSync } from "node:zlib";
import { join } from "node:path";
import { ABO_SOURCE, mapAboListing, type AboMappedProduct } from "../lib/catalog/abo";
import { assertCatalogSnapshot, readCatalogSnapshot, writeCatalogSnapshotAtomically, type CatalogSnapshot } from "../lib/catalog/snapshot";

const DATASET = "hyper3labs/amazon-berkeley-objects";
const ROWS_ENDPOINT = "https://datasets-server.huggingface.co/rows";
const IMAGE_METADATA_URL = "https://amazon-berkeley-objects.s3.amazonaws.com/images/metadata/images.csv.gz";
const EXPECTED_ROW_COUNT = 147_702;
const TARGET_PRODUCTS = 1_000;
const ROWS_PER_PAGE = 100;
const SAMPLE_PAGE_COUNT = 56;
const MAX_CANDIDATE_ROWS = SAMPLE_PAGE_COUNT * ROWS_PER_PAGE;
const IMAGE_DIRECTORY = join(process.cwd(), "public", "images", "catalog");
const OUTPUT_FILE = join(process.cwd(), "data", "catalog.json");
const IMPORTED_AT = new Date().toISOString();
const MAX_SOURCE_IMAGE_BYTES = 18 * 1024 * 1024;
const MAX_OPTIMIZED_IMAGE_BYTES = 850 * 1024;
const MAX_TOTAL_IMAGE_BYTES = 320 * 1024 * 1024;
const CONCURRENCY = 6;
const IMAGE_CONCURRENCY_PER_PRODUCT = 4;

type DatasetRow = { row?: unknown; [key: string]: unknown };
type RowsPayload = { rows?: DatasetRow[]; num_rows_total?: number; num_rows?: number; error?: string };
type ImageMetadata = { path: string; width: number; height: number };
type Candidate = { product: AboMappedProduct };
type GalleryAsset = { path: string; bytes: number };
type SharpPipeline = {
  rotate(): SharpPipeline;
  resize(options: { width: number; height: number; fit: "inside"; withoutEnlargement: boolean }): SharpPipeline;
  webp(options: { quality: number; effort: number; smartSubsample?: boolean }): SharpPipeline;
  toBuffer(): Promise<Buffer>;
};
type SharpFactory = (input: Buffer) => SharpPipeline;

// The deployment workflow installs this pinned, server-only optimizer before import.
// It is not bundled into the storefront or downloaded by visitors.
const sharp = createRequire(join(process.cwd(), "package.json"))("sharp") as SharpFactory;

function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function fetchBytes(url: string, label: string, timeoutMs = 35_000): Promise<{ bytes: Buffer; contentType: string }> {
  let lastError: unknown;
  for (let attempt = 0; attempt < 5; attempt += 1) {
    try {
      const response = await fetch(url, {
        headers: { "user-agent": "GlobalCart catalog importer/1.0 (ABO CC BY 4.0 attribution)" },
        signal: AbortSignal.timeout(timeoutMs)
      });
      if (!response.ok) throw new Error(label + " returned HTTP " + response.status);
      return {
        bytes: Buffer.from(await response.arrayBuffer()),
        contentType: response.headers.get("content-type") || ""
      };
    } catch (error) {
      lastError = error;
      if (attempt < 4) await sleep(700 * (attempt + 1) ** 2);
    }
  }
  throw lastError instanceof Error ? lastError : new Error(label + " failed");
}

async function fetchJson(url: string, label: string): Promise<RowsPayload> {
  const { bytes } = await fetchBytes(url, label);
  const payload = JSON.parse(bytes.toString("utf8")) as RowsPayload;
  if (!Array.isArray(payload.rows)) {
    throw new Error(label + " did not return a rows array: " + (payload.error || "unknown response"));
  }
  return payload;
}

async function getRows(offset: number): Promise<{ rows: unknown[]; total: number }> {
  const url = new URL(ROWS_ENDPOINT);
  url.searchParams.set("dataset", DATASET);
  url.searchParams.set("config", "listings");
  url.searchParams.set("split", "train");
  url.searchParams.set("offset", String(offset));
  url.searchParams.set("length", String(ROWS_PER_PAGE));
  const payload = await fetchJson(url.toString(), "ABO dataset page at offset " + offset);
  return {
    rows: payload.rows!.map(entry => {
      if (entry && typeof entry === "object" && "row" in entry) return (entry as DatasetRow).row;
      return entry;
    }),
    total: Number.isSafeInteger(payload.num_rows_total) && Number(payload.num_rows_total) > 0
      ? Number(payload.num_rows_total)
      : EXPECTED_ROW_COUNT
  };
}

async function loadImageMetadata(requiredIds: Set<string>): Promise<Map<string, ImageMetadata>> {
  console.log("Downloading ABO original-image metadata:", requiredIds.size, "image IDs");
  const { bytes } = await fetchBytes(IMAGE_METADATA_URL, "ABO original-image metadata", 90_000);
  const csv = gunzipSync(bytes).toString("utf8");
  const lines = csv.split(/\r?\n/);
  const header = (lines[0] || "").split(",");
  const idColumn = header.indexOf("image_id");
  const widthColumn = header.indexOf("width");
  const heightColumn = header.indexOf("height");
  const pathColumn = header.indexOf("path");
  if ([idColumn, widthColumn, heightColumn, pathColumn].some(index => index < 0)) {
    throw new Error("ABO image metadata CSV has an unexpected header.");
  }

  const metadata = new Map<string, ImageMetadata>();
  for (let index = 1; index < lines.length && metadata.size < requiredIds.size; index += 1) {
    const fields = lines[index].split(",");
    const id = fields[idColumn];
    const path = fields[pathColumn];
    if (!requiredIds.has(id) || !path || path.startsWith("/") || path.split("/").some(part => part === "..")) continue;
    const width = Number(fields[widthColumn]);
    const height = Number(fields[heightColumn]);
    if (!Number.isSafeInteger(width) || !Number.isSafeInteger(height) || width < 1 || height < 1) continue;
    if (!/^[a-f0-9]{2}\/[a-f0-9]+\.(?:jpg|jpeg|png)$/i.test(path)) continue;
    metadata.set(id, { path, width, height });
  }

  const missing = [...requiredIds].filter(id => !metadata.has(id));
  if (missing.length > 0) {
    throw new Error("ABO image metadata is missing " + missing.length + " required image IDs; examples: " + missing.slice(0, 8).join(", "));
  }
  console.log("Resolved original-image metadata:", metadata.size, "of", requiredIds.size);
  return metadata;
}

async function mapWithConcurrency<T, R>(
  values: readonly T[],
  concurrency: number,
  mapper: (value: T, index: number) => Promise<R>
): Promise<R[]> {
  const output = new Array<R>(values.length);
  let next = 0;
  const workers = Array.from({ length: Math.min(concurrency, values.length) }, async () => {
    while (true) {
      const index = next;
      next += 1;
      if (index >= values.length) return;
      output[index] = await mapper(values[index], index);
    }
  });
  await Promise.all(workers);
  return output;
}

function balancedCandidates(products: AboMappedProduct[]): Candidate[] {
  const buckets = new Map<string, Candidate[]>();
  for (const product of products) {
    const bucket = buckets.get(product.category) ?? [];
    bucket.push({ product });
    buckets.set(product.category, bucket);
  }
  for (const bucket of buckets.values()) {
    bucket.sort((left, right) => left.product.sourceId.localeCompare(right.product.sourceId));
  }

  const categories = [...buckets.keys()].sort((a, b) => a.localeCompare(b));
  const candidateQueue: Candidate[] = [];
  const perCategoryCap = Math.max(120, Math.ceil(TARGET_PRODUCTS / Math.max(1, categories.length)) * 3);
  const cursors = new Map(categories.map(category => [category, 0]));
  let remaining = true;
  while (remaining && candidateQueue.length < MAX_CANDIDATE_ROWS) {
    remaining = false;
    for (const category of categories) {
      const bucket = buckets.get(category)!;
      const cursor = cursors.get(category)!;
      if (cursor >= Math.min(bucket.length, perCategoryCap)) continue;
      candidateQueue.push(bucket[cursor]);
      cursors.set(category, cursor + 1);
      remaining = true;
    }
  }
  return candidateQueue;
}

function hasJpegOrPngSignature(bytes: Buffer): boolean {
  return (bytes[0] === 0xff && bytes[1] === 0xd8 && bytes[2] === 0xff) ||
    (bytes[0] === 0x89 && bytes[1] === 0x50 && bytes[2] === 0x4e && bytes[3] === 0x47);
}

async function existingGalleryImage(sourceId: string, ordinal: number): Promise<GalleryAsset | null> {
  const name = "abo-" + sourceId + "-" + String(ordinal + 1).padStart(2, "0") + ".webp";
  const path = join(IMAGE_DIRECTORY, name);
  try {
    const metadata = await stat(path);
    if (metadata.size > 1000 && metadata.size <= MAX_OPTIMIZED_IMAGE_BYTES) {
      return { path: "/images/catalog/" + name, bytes: metadata.size };
    }
  } catch {
    return null;
  }
  return null;
}

async function downloadOriginalImage(
  sourceId: string,
  imageId: string,
  metadata: ImageMetadata,
  ordinal: number
): Promise<GalleryAsset | null> {
  const cached = await existingGalleryImage(sourceId, ordinal);
  if (cached) return cached;

  const originalUrl = "https://amazon-berkeley-objects.s3.amazonaws.com/images/original/" + metadata.path;
  const extension = metadata.path.slice(metadata.path.lastIndexOf(".") + 1).toLowerCase();
  const cdnUrl = "https://m.media-amazon.com/image/I/" + encodeURIComponent(imageId) + "." + extension;
  const urls = [originalUrl, cdnUrl];

  for (let attempt = 0; attempt < 3; attempt += 1) {
    for (const url of urls) {
      try {
        const response = await fetch(url, {
          headers: { "user-agent": "GlobalCart static image importer/1.0" },
          signal: AbortSignal.timeout(30_000)
        });
        if (!response.ok) continue;
        const input = Buffer.from(await response.arrayBuffer());
        if (input.length < 512 || input.length > MAX_SOURCE_IMAGE_BYTES || !hasJpegOrPngSignature(input)) continue;

        let output = await sharp(input)
          .rotate()
          .resize({ width: 1440, height: 1440, fit: "inside", withoutEnlargement: true })
          .webp({ quality: 84, effort: 5, smartSubsample: true })
          .toBuffer();

        if (output.length > MAX_OPTIMIZED_IMAGE_BYTES) {
          output = await sharp(input)
            .rotate()
            .resize({ width: 1280, height: 1280, fit: "inside", withoutEnlargement: true })
            .webp({ quality: 78, effort: 5, smartSubsample: true })
            .toBuffer();
        }
        if (output.length < 1000 || output.length > MAX_OPTIMIZED_IMAGE_BYTES) continue;

        const name = "abo-" + sourceId + "-" + String(ordinal + 1).padStart(2, "0") + ".webp";
        const finalPath = join(IMAGE_DIRECTORY, name);
        const temporaryPath = finalPath + "." + process.pid + ".tmp";
        await mkdir(IMAGE_DIRECTORY, { recursive: true });
        await writeFile(temporaryPath, output, { flag: "wx" });
        await rename(temporaryPath, finalPath).catch(async error => {
          await rm(temporaryPath, { force: true });
          throw error;
        });
        return { path: "/images/catalog/" + name, bytes: output.length };
      } catch {
        // Retry the official S3 original first and the documented Amazon image CDN second.
      }
    }
    if (attempt < 2) await sleep(500 * (attempt + 1));
  }
  return null;
}

async function downloadGallery(
  candidate: Candidate,
  metadata: Map<string, ImageMetadata>
): Promise<GalleryAsset[] | null> {
  const ids = candidate.product.imageIds;
  if (ids.length === 0) return null;
  const assets = await mapWithConcurrency(ids, IMAGE_CONCURRENCY_PER_PRODUCT, async (imageId, index) => {
    const imageMetadata = metadata.get(imageId);
    if (!imageMetadata) return null;
    return await downloadOriginalImage(candidate.product.sourceId, imageId, imageMetadata, index);
  });
  if (assets.some(asset => asset === null)) return null;
  return assets as GalleryAsset[];
}

async function main(): Promise<void> {
  const previous = await readCatalogSnapshot(OUTPUT_FILE);
  try {
    const first = await getRows(0);
    const total = Math.max(1, first.total);
    const offsets = [0];
    for (let index = 1; index < SAMPLE_PAGE_COUNT; index += 1) {
      const offset = Math.min(total - 1, Math.floor(index * total / SAMPLE_PAGE_COUNT));
      if (!offsets.includes(offset)) offsets.push(offset);
    }
    const pages = await mapWithConcurrency(offsets.slice(1), 4, offset => getRows(offset));
    const rows = [...first.rows, ...pages.flatMap(page => page.rows)];

    const mappedById = new Map<string, AboMappedProduct>();
    for (const row of rows) {
      const product = mapAboListing(row, IMPORTED_AT);
      if (product && product.imageIds.length > 0 && !mappedById.has(product.sourceId)) {
        mappedById.set(product.sourceId, product);
      }
    }
    const candidates = balancedCandidates([...mappedById.values()]);
    const requiredImageIds = new Set(candidates.flatMap(candidate => candidate.product.imageIds));
    console.log("ABO candidate set:", JSON.stringify({
      dataset: DATASET,
      totalRows: total,
      pages: offsets.length,
      rowsRead: rows.length,
      uniqueImageBearingProducts: mappedById.size,
      candidateQueue: candidates.length,
      requiredImageIds: requiredImageIds.size,
      categories: [...new Set([...mappedById.values()].map(item => item.category))].length
    }));

    if (candidates.length < TARGET_PRODUCTS) {
      throw new Error("ABO sample produced only " + candidates.length + " multi-image candidate products; cannot publish 1,000 products.");
    }
    const imageMetadata = await loadImageMetadata(requiredImageIds);
    await mkdir(IMAGE_DIRECTORY, { recursive: true });

    const downloaded = new Map<string, GalleryAsset[]>();
    let nextCandidate = 0;
    let totalImageBytes = 0;
    let totalImageCount = 0;
    let exhausted = false;
    const worker = async () => {
      while (!exhausted && downloaded.size < TARGET_PRODUCTS) {
        const index = nextCandidate;
        nextCandidate += 1;
        if (index >= candidates.length) return;
        const candidate = candidates[index];
        const gallery = await downloadGallery(candidate, imageMetadata);
        if (!gallery) continue;
        const galleryBytes = gallery.reduce((sum, asset) => sum + asset.bytes, 0);
        if (downloaded.size < TARGET_PRODUCTS) {
          if (totalImageBytes + galleryBytes > MAX_TOTAL_IMAGE_BYTES) {
            exhausted = true;
            throw new Error("The 320 MiB optimized-image budget was reached before 1,000 complete product galleries could be stored.");
          }
          totalImageBytes += galleryBytes;
          totalImageCount += gallery.length;
          downloaded.set(candidate.product.sourceId, gallery);
          if (downloaded.size % 50 === 0) {
            console.log("Imported complete product galleries:", downloaded.size, "/", TARGET_PRODUCTS,
              "· optimized images:", totalImageCount,
              "· image payload MiB:", Number((totalImageBytes / 1024 / 1024).toFixed(1)));
          }
        }
      }
    };
    await Promise.all(Array.from({ length: CONCURRENCY }, worker));

    if (downloaded.size !== TARGET_PRODUCTS) {
      throw new Error("Only " + downloaded.size + " complete product galleries were downloaded; 1,000 products are required.");
    }

    const selectedProducts: AboMappedProduct[] = [];
    for (const candidate of candidates) {
      const gallery = downloaded.get(candidate.product.sourceId);
      if (!gallery) continue;
      selectedProducts.push({
        ...candidate.product,
        thumbnail: gallery[0].path,
        images: gallery.slice(1).map(asset => asset.path)
      });
      if (selectedProducts.length === TARGET_PRODUCTS) break;
    }

    const usedIds = new Set<number>();
    const products = selectedProducts.map(product => {
      let id = product.id;
      while (usedIds.has(id)) id = id >= 2_147_483_646 ? 1 : id + 1;
      usedIds.add(id);
      const { imageUrl: _imageUrl, imageIds: _imageIds, ...catalogProduct } = product;
      return catalogProduct;
    });

    const snapshot: CatalogSnapshot = {
      schemaVersion: 1,
      source: ABO_SOURCE,
      generatedAt: IMPORTED_AT,
      productCount: products.length,
      products
    };
    assertCatalogSnapshot(snapshot, OUTPUT_FILE);
    await writeCatalogSnapshotAtomically(snapshot, OUTPUT_FILE);

    const referencedImages = new Set(products.flatMap(product => [product.thumbnail, ...product.images])
      .map(path => path.replace(/^\/images\/catalog\//, "")));
    for (const name of await readdir(IMAGE_DIRECTORY)) {
      if (/^abo-[A-Z0-9]{10}(?:-\d{2})?\.(?:jpg|png|webp)$/i.test(name) && !referencedImages.has(name)) {
        await rm(join(IMAGE_DIRECTORY, name), { force: true });
      }
    }

    const categories = Object.entries(products.reduce<Record<string, number>>((counts, product) => {
      counts[product.category] = (counts[product.category] || 0) + 1;
      return counts;
    }, {})).sort((left, right) => right[1] - left[1]);
    const productsWithMultipleImages = products.filter(product => product.images.length > 0).length;
    console.log("ABO high-resolution catalog published:", JSON.stringify({
      products: products.length,
      productsWithMultipleImages,
      optimizedImageFiles: totalImageCount,
      optimizedImageMegabytes: Number((totalImageBytes / 1024 / 1024).toFixed(2)),
      averageImagesPerProduct: Number((totalImageCount / products.length).toFixed(2)),
      categories,
      generatedAt: IMPORTED_AT
    }));
  } catch (error) {
    if (previous?.source === ABO_SOURCE && previous.products.length >= 900) {
      console.error("ABO high-resolution refresh failed; the last-known-good snapshot remains on disk:", error);
    }
    throw error;
  }
}

main().catch(error => {
  console.error("ABO high-resolution catalog import failed:", error);
  process.exitCode = 1;
});
