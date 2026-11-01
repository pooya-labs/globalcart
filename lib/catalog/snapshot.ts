import { randomUUID } from "node:crypto";
import { mkdir, rename, rm, writeFile } from "node:fs/promises";
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
