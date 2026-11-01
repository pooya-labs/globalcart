import { createHash } from "node:crypto";
import { readFile, rename, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { GoogleAppsScriptTranslator, type ProductTranslationFields } from "../lib/translation/google-apps-script";
import type { CatalogProduct } from "../lib/catalog/types";

type LocalizedProduct = {
  locale: "fa";
  status: "translated" | "error";
  provider: string;
  sourceHash: string;
  translatedAt: string | null;
  fields: ProductTranslationFields | null;
};
type Snapshot = {
  schemaVersion: number;
  source: string;
  generatedAt: string;
  productCount: number;
  products: Array<CatalogProduct & { localization?: LocalizedProduct }>;
};

const directory = join(process.cwd(), "data");
const inputFile = join(directory, "catalog.json");
const temporaryFile = join(directory, "catalog.json.tmp");
const hashSource = (fields: ProductTranslationFields) =>
  createHash("sha256").update(JSON.stringify(fields)).digest("hex");

async function main(): Promise<void> {
  const endpoint = process.env.GOOGLE_TRANSLATE_ENDPOINT?.trim();
  const token = process.env.GOOGLE_TRANSLATE_TOKEN?.trim();
  if (!endpoint || !token) {
    console.log("Google Translate is not configured; keeping the catalog snapshot unchanged.");
    return;
  }

  const snapshot = JSON.parse(await readFile(inputFile, "utf8")) as Snapshot;
  if (!Array.isArray(snapshot.products)) throw new Error("Catalog snapshot has no products array");
  const translator = new GoogleAppsScriptTranslator(endpoint, token);
  let translated = 0;
  let failed = 0;
  let cached = 0;

  for (const product of snapshot.products) {
    const sourceFields: ProductTranslationFields = {
      title: product.title,
      description: product.description,
      category: product.category,
      brand: product.brand
    };
    const sourceHash = hashSource(sourceFields);
    if (product.localization?.status === "translated" && product.localization.sourceHash === sourceHash) {
      cached += 1;
      continue;
    }

    try {
      const fields = await translator.translate(sourceFields);
      product.localization = {
        locale: "fa",
        status: "translated",
        provider: translator.name,
        sourceHash,
        translatedAt: new Date().toISOString(),
        fields
      };
      translated += 1;
    } catch (error) {
      product.localization = {
        locale: "fa",
        status: "error",
        provider: translator.name,
        sourceHash,
        translatedAt: null,
        fields: null
      };
      failed += 1;
      console.warn(`Translation failed for product ${product.id}:`, error);
    }

    // Google Apps Script has daily translate-call quotas; requests are deliberately sequential.
    await new Promise(resolve => setTimeout(resolve, 1_000));
  }

  await writeFile(temporaryFile, JSON.stringify(snapshot, null, 2) + "\n", "utf8");
  await rename(temporaryFile, inputFile);
  console.log(`Google Persian localization: ${translated} translated, ${cached} cached, ${failed} failed of ${snapshot.products.length} products.`);
}

main().catch(error => {
  console.error("Catalog translation job failed:", error);
  process.exitCode = 1;
});
