import { copyFileSync, mkdirSync, statSync } from "node:fs";
import { dirname, resolve } from "node:path";

const source = resolve(process.cwd(), "data/catalog.json");
const destination = resolve(process.cwd(), "public/data/catalog.json");

mkdirSync(dirname(destination), { recursive: true });
copyFileSync(source, destination);

console.log(`Prepared public catalog snapshot (${statSync(source).size.toLocaleString("en-US")} bytes).`);
