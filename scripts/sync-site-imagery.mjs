import { mkdir, readFile, rename, rm, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";

const root = process.cwd();
const destination = (name) => resolve(root, "public", "images", "site", name);
const source = "https://images.unsplash.com/";

const assets = [
  {
    name: "hero-480.webp",
    url: `${source}photo-1490312278390-ab64016e0aa9?w=480&q=76&fm=webp`
  },
  {
    name: "hero-768.webp",
    url: `${source}photo-1490312278390-ab64016e0aa9?w=768&q=78&fm=webp`
  },
  {
    name: "hero-1600.webp",
    url: `${source}photo-1490312278390-ab64016e0aa9?w=1600&q=80&fm=webp`
  },
  {
    name: "editorial-1200.webp",
    url: `${source}photo-1494438639946-1ebd1d20bf85?w=1200&q=80&fm=webp`
  }
];

function isValidWebp(bytes) {
  return bytes.length >= 10_000 &&
    bytes.toString("ascii", 0, 4) === "RIFF" &&
    bytes.toString("ascii", 8, 12) === "WEBP";
}

async function existingAssetsAreValid() {
  for (const asset of assets) {
    try {
      const bytes = await readFile(destination(asset.name));
      if (!isValidWebp(bytes)) return false;
    } catch {
      return false;
    }
  }
  return true;
}

async function downloadAsset(asset) {
  const response = await fetch(asset.url, {
    headers: { "user-agent": "GlobalCart-site-assets/1.0" },
    signal: AbortSignal.timeout(35_000)
  });
  if (!response.ok) {
    throw new Error(`${asset.name}: image host returned HTTP ${response.status}`);
  }

  const contentType = response.headers.get("content-type") ?? "";
  if (!contentType.toLowerCase().includes("image/webp")) {
    throw new Error(`${asset.name}: expected image/webp, received ${contentType || "no content type"}`);
  }

  const bytes = Buffer.from(await response.arrayBuffer());
  if (!isValidWebp(bytes)) {
    throw new Error(`${asset.name}: downloaded response is not a valid WebP image`);
  }
  return bytes;
}

const staged = [];
try {
  for (const asset of assets) {
    const bytes = await downloadAsset(asset);
    const target = destination(asset.name);
    const temporary = `${target}.download`;
    await mkdir(dirname(target), { recursive: true });
    await writeFile(temporary, bytes);
    staged.push({ target, temporary, size: bytes.length });
  }

  // Replace the set only after every image has downloaded and passed validation.
  for (const item of staged) {
    await rename(item.temporary, item.target);
  }
  const total = staged.reduce((sum, item) => sum + item.size, 0);
  console.log(`Prepared ${staged.length} self-hosted site images (${total.toLocaleString()} bytes).`);
} catch (error) {
  await Promise.all(staged.map(item => rm(item.temporary, { force: true })));
  if (await existingAssetsAreValid()) {
    console.warn("Site image refresh failed; retaining the complete last-known-good asset set.", error);
  } else {
    console.error("Site images could not be prepared and no valid last-known-good set exists.", error);
    process.exitCode = 1;
  }
}
