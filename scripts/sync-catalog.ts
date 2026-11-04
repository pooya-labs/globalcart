import { mkdir, readdir, rename, rm, stat, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { ABO_SOURCE, mapAboListing, type AboMappedProduct } from "../lib/catalog/abo";
import { assertCatalogSnapshot, readCatalogSnapshot, writeCatalogSnapshotAtomically, type CatalogSnapshot } from "../lib/catalog/snapshot";

const DATASET = "hyper3labs/amazon-berkeley-objects";
const ROWS_ENDPOINT = "https://datasets-server.huggingface.co/rows";
const EXPECTED_ROW_COUNT = 147_702;
const TARGET_PRODUCTS = 1_000;
const ROWS_PER_PAGE = 100;
const SAMPLE_PAGE_COUNT = 56;
const MAX_CANDIDATE_ROWS = SAMPLE_PAGE_COUNT * ROWS_PER_PAGE;
const IMAGE_DIRECTORY = join(process.cwd(), "public", "images", "catalog");
const OUTPUT_FILE = join(process.cwd(), "data", "catalog.json");
const IMPORTED_AT = new Date().toISOString();
const MAX_IMAGE_BYTES = 350 * 1024;
const MAX_TOTAL_IMAGE_BYTES = 75 * 1024 * 1024;
const CONCURRENCY = 8;

type DatasetRow = { row?: unknown; [key: string]: unknown };
type RowsPayload = { rows?: DatasetRow[]; num_rows_total?: number; num_rows?: number; error?: string };
type Candidate = { product: AboMappedProduct; imageUrl: string };

function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function fetchJson(url: string, label: string): Promise<RowsPayload> {
  let lastError: unknown;
  for (let attempt = 0; attempt < 5; attempt += 1) {
    try {
      const response = await fetch(url, {
        headers: { "user-agent": "GlobalCart catalog importer/1.0 (CC BY 4.0 dataset attribution)" },
        signal: AbortSignal.timeout(35_000)
      });
      if (!response.ok) throw new Error(label + " returned HTTP " + response.status);
      const payload = await response.json() as RowsPayload;
      if (!Array.isArray(payload.rows)) throw new Error(label + " did not return a rows array: " + (payload.error || "unknown response"));
      return payload;
    } catch (error) {
      lastError = error;
      if (attempt < 4) await sleep(800 * (attempt + 1) ** 2);
    }
  }
  throw lastError instanceof Error ? lastError : new Error(label + " failed");
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
    bucket.push({ product, imageUrl: product.imageUrl });
    buckets.set(product.category, bucket);
  }
  for (const bucket of buckets.values()) {
    bucket.sort((left, right) => left.product.sourceId.localeCompare(right.product.sourceId));
  }

  const categories = [...buckets.keys()].sort((a, b) => a.localeCompare(b));
  const candidateQueue: Candidate[] = [];
  const perCategoryCap = Math.max(100, Math.ceil(TARGET_PRODUCTS / Math.max(1, categories.length)) * 2);
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

function extensionFor(contentType: string, bytes: Buffer): string | null {
  const mime = contentType.split(";")[0].trim().toLowerCase();
  if (mime === "image/jpeg" && bytes[0] === 0xff && bytes[1] === 0xd8 && bytes[2] === 0xff) return ".jpg";
  if (mime === "image/png" && bytes[0] === 0x89 && bytes[1] === 0x50 && bytes[2] === 0x4e && bytes[3] === 0x47) return ".png";
  if (mime === "image/webp" && bytes.subarray(0, 4).toString("ascii") === "RIFF" && bytes.subarray(8, 12).toString("ascii") === "WEBP") return ".webp";
  return null;
}

async function existingImage(sourceId: string): Promise<string | null> {
  const prefix = "abo-" + sourceId + ".";
  let names: string[];
  try {
    names = await readdir(IMAGE_DIRECTORY);
  } catch {
    return null;
  }
  const name = names.find(item => item.startsWith(prefix) && /\.(?:jpg|png|webp)$/i.test(item));
  if (!name) return null;
  const path = join(IMAGE_DIRECTORY, name);
  try {
    const metadata = await stat(path);
    if (metadata.size > 512 && metadata.size <= MAX_IMAGE_BYTES) return "/images/catalog/" + name;
  } catch {
    return null;
  }
  return null;
}

async function downloadImage(candidate: Candidate): Promise<{ path: string; bytes: number } | null> {
  const cached = await existingImage(candidate.product.sourceId);
  if (cached) {
    const metadata = await stat(join(process.cwd(), "public", cached));
    return { path: cached, bytes: metadata.size };
  }

  for (let attempt = 0; attempt < 3; attempt += 1) {
    try {
      const response = await fetch(candidate.imageUrl, {
        headers: { "user-agent": "GlobalCart static catalog asset importer/1.0" },
        signal: AbortSignal.timeout(25_000)
      });
      if (!response.ok) throw new Error("image HTTP " + response.status);
      const bytes = Buffer.from(await response.arrayBuffer());
      if (bytes.length < 512 || bytes.length > MAX_IMAGE_BYTES) throw new Error("image size outside accepted range");
      const extension = extensionFor(response.headers.get("content-type") || "", bytes);
      if (!extension) throw new Error("unrecognized or mismatched image format");
      const name = "abo-" + candidate.product.sourceId + extension;
      const finalPath = join(IMAGE_DIRECTORY, name);
      const temporaryPath = finalPath + "." + process.pid + ".tmp";
      await mkdir(IMAGE_DIRECTORY, { recursive: true });
      await writeFile(temporaryPath, bytes, { flag: "wx" });
      await rename(temporaryPath, finalPath).catch(async error => {
        await rm(temporaryPath, { force: true });
        throw error;
      });
      return { path: "/images/catalog/" + name, bytes: bytes.length };
    } catch {
      if (attempt < 2) await sleep(500 * (attempt + 1));
    }
  }
  return null;
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
      if (product && !mappedById.has(product.sourceId)) mappedById.set(product.sourceId, product);
    }
    const candidates = balancedCandidates([...mappedById.values()]);
    console.log("ABO sampling:", JSON.stringify({
      dataset: DATASET,
      totalRows: total,
      pages: offsets.length,
      rowsRead: rows.length,
      uniqueMappableProducts: mappedById.size,
      candidateQueue: candidates.length,
      categories: [...new Set([...mappedById.values()].map(item => item.category))].length
    }));

    if (candidates.length < TARGET_PRODUCTS) {
      throw new Error("ABO sample produced only " + candidates.length + " image-bearing candidate rows; cannot safely publish 1,000 products.");
    }

    await mkdir(IMAGE_DIRECTORY, { recursive: true });
    const downloaded = new Map<string, { path: string; bytes: number }>();
    let nextCandidate = 0;
    let totalImageBytes = 0;
    let exhausted = false;
    const worker = async () => {
      while (!exhausted && downloaded.size < TARGET_PRODUCTS) {
        const index = nextCandidate;
        nextCandidate += 1;
        if (index >= candidates.length) return;
        const candidate = candidates[index];
        const image = await downloadImage(candidate);
        if (!image) continue;
        if (downloaded.size < TARGET_PRODUCTS) {
          if (totalImageBytes + image.bytes > MAX_TOTAL_IMAGE_BYTES) {
            exhausted = true;
            throw new Error("Local product imagery exceeded the 75 MiB safety budget before reaching 1,000 products.");
          }
          totalImageBytes += image.bytes;
          downloaded.set(candidate.product.sourceId, image);
        }
        if (downloaded.size > 0 && downloaded.size % 100 === 0) {
          console.log("Downloaded verified ABO images:", downloaded.size, "/", TARGET_PRODUCTS);
        }
      }
    };
    await Promise.all(Array.from({ length: CONCURRENCY }, worker));

    if (downloaded.size !== TARGET_PRODUCTS) {
      throw new Error("Only " + downloaded.size + " valid local product images were downloaded; the 1,000-product publication threshold was not met.");
    }

    const selectedProducts: AboMappedProduct[] = [];
    for (const candidate of candidates) {
      const asset = downloaded.get(candidate.product.sourceId);
      if (!asset) continue;
      selectedProducts.push({ ...candidate.product, thumbnail: asset.path, images: [asset.path] });
      if (selectedProducts.length === TARGET_PRODUCTS) break;
    }

    const usedIds = new Set<number>();
    const products = selectedProducts.map(product => {
      let id = product.id;
      while (usedIds.has(id)) id = id >= 2_147_483_646 ? 1 : id + 1;
      usedIds.add(id);
      const { imageUrl: _imageUrl, ...catalogProduct } = product;
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

    const referencedImages = new Set(products.map(product => product.thumbnail.replace(/^\/images\/catalog\//, "")));
    for (const name of await readdir(IMAGE_DIRECTORY)) {
      if (/^abo-[A-Z0-9]{10}\.(?:jpg|png|webp)$/i.test(name) && !referencedImages.has(name)) {
        await rm(join(IMAGE_DIRECTORY, name), { force: true });
      }
    }

    const categories = Object.entries(products.reduce<Record<string, number>>((counts, product) => {
      counts[product.category] = (counts[product.category] || 0) + 1;
      return counts;
    }, {})).sort((left, right) => right[1] - left[1]);
    console.log("ABO catalog snapshot published:", JSON.stringify({
      products: products.length,
      verifiedLocalImages: downloaded.size,
      localImageMegabytes: Number((totalImageBytes / 1024 / 1024).toFixed(2)),
      categories,
      generatedAt: IMPORTED_AT
    }));
  } catch (error) {
    if (previous?.source === ABO_SOURCE && previous.products.length >= 900) {
      console.warn("ABO refresh failed; keeping the previous complete ABO snapshot:", error);
      return;
    }
    throw error;
  }
}

main().catch(error => {
  console.error("ABO catalog import failed:", error);
  process.exitCode = 1;
});
