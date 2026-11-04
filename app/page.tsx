import catalogSnapshot from "../data/catalog.json";
import fxSnapshot from "../data/fx-rate.json";
import StorefrontClient, { type Product } from "./StorefrontClient";
import type { FxRateSnapshot } from "../lib/fx/tgju";

export default function Home() {
  const products = Array.isArray(catalogSnapshot.products)
    ? catalogSnapshot.products as unknown as Product[]
    : [];
  const catalogTimestamp = typeof catalogSnapshot.generatedAt === "string"
    ? catalogSnapshot.generatedAt
    : null;

  return (
    <StorefrontClient
      initialProducts={products.slice(0, 36)}
      catalogTimestamp={catalogTimestamp}
      fxSnapshot={fxSnapshot as unknown as FxRateSnapshot}
    />
  );
}
