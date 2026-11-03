import { randomUUID } from "node:crypto";
import { readFile, rename, unlink, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { normalizeTgjuDollarPage } from "../lib/fx/tgju";

const endpoint = "https://www.tgju.org/%D9%82%DB%8C%D9%85%D8%AA-%D8%AF%D9%84%D8%A7%D8%B1";
const outputPath = join(process.cwd(), "data", "fx-rate.json");
const MAX_PAGE_CHARS = 5 * 1024 * 1024;

async function main(): Promise<void> {
  const response = await fetch(endpoint, {
    headers: {
      accept: "text/html,application/xhtml+xml",
      "accept-language": "fa-IR,fa;q=0.9,en;q=0.7",
      "user-agent": "GlobalCart-FX-Snapshot/1.0 (+https://github.com/pooya16pg-spec/globalcart)"
    },
    signal: AbortSignal.timeout(15_000),
    cache: "no-store",
    redirect: "follow"
  });
  const finalUrl = new URL(response.url || endpoint);
  if (finalUrl.hostname !== "www.tgju.org" && finalUrl.hostname !== "tgju.org") {
    throw new Error("TGJU redirected the request outside its public website; refusing third-party content");
  }
  if (!response.ok) throw new Error("TGJU dollar page request failed with HTTP " + response.status);
  const contentType = response.headers.get("content-type") ?? "";
  if (!/text\/html|application\/xhtml\+xml/i.test(contentType)) {
    throw new Error("TGJU response is not HTML (content-type: " + (contentType || "missing") + ")");
  }
  const html = await response.text();
  if (html.length > MAX_PAGE_CHARS) throw new RangeError("TGJU page exceeded the 5 MiB response limit");
  const quote = normalizeTgjuDollarPage(html);
  const serialized = JSON.stringify(quote, null, 2) + "\n";

  let previous = "";
  try {
    previous = await readFile(outputPath, "utf8");
  } catch (error) {
    if (!(error && typeof error === "object" && "code" in error && error.code === "ENOENT")) throw error;
  }
  if (previous === serialized) {
    process.stdout.write("TGJU free-market USD quote is unchanged; no commit is needed.\n");
    return;
  }
  const temporaryPath = outputPath + "." + randomUUID() + ".tmp";
  try {
    await writeFile(temporaryPath, serialized, { encoding: "utf8", flag: "wx" });
    await rename(temporaryPath, outputPath);
  } finally {
    await unlink(temporaryPath).catch(() => undefined);
  }
  process.stdout.write("Updated TGJU free-market USD quote to " +
    Math.round(quote.tomanPerUsd ?? 0).toLocaleString("en-US") +
    " Toman per USD; source Tehran time " + quote.sourceTimeTehran + ".\n");
}
main().catch(error => {
  process.stderr.write((error instanceof Error ? error.message : String(error)) + "\n");
  process.exitCode = 1;
});
