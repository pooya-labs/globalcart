import { randomUUID } from "node:crypto";
import { readFile, rename, unlink, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { normalizeNobitexOrderBook } from "../lib/fx/nobitex";

const endpoint = "https://apiv2.nobitex.ir/v3/orderbook/USDTIRT";
const outputPath = join(process.cwd(), "data", "fx-rate.json");

async function main(): Promise<void> {
  const response = await fetch(endpoint, {
    headers: {
      accept: "application/json",
      "user-agent": "GlobalCart-FX-Snapshot/1.0"
    },
    signal: AbortSignal.timeout(15_000)
  });
  if (!response.ok) {
    throw new Error("Nobitex order-book request failed with HTTP " + response.status);
  }

  const payload: unknown = await response.json();
  const quote = normalizeNobitexOrderBook(payload);
  const serialized = JSON.stringify(quote, null, 2) + "\n";

  let previous = "";
  try {
    previous = await readFile(outputPath, "utf8");
  } catch (error) {
    if (!(error && typeof error === "object" && "code" in error && error.code === "ENOENT")) {
      throw error;
    }
  }
  if (previous === serialized) {
    process.stdout.write("Nobitex FX snapshot is unchanged; no commit is needed.\n");
    return;
  }

  const temporaryPath = outputPath + "." + randomUUID() + ".tmp";
  try {
    await writeFile(temporaryPath, serialized, { encoding: "utf8", flag: "wx" });
    await rename(temporaryPath, outputPath);
  } finally {
    await unlink(temporaryPath).catch(() => undefined);
  }

  process.stdout.write(
    "Updated Nobitex USDT/IRT midpoint to " +
    Math.round(quote.tomanPerUsdt ?? 0).toLocaleString("en-US") +
    " Toman per USDT; exchange timestamp " + quote.marketUpdatedAt + ".\n"
  );
}

main().catch(error => {
  process.stderr.write((error instanceof Error ? error.message : String(error)) + "\n");
  process.exitCode = 1;
});
