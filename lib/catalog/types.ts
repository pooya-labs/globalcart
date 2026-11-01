export type CatalogLocalization = {
  locale: "fa";
  status: "translated" | "error";
  provider: string;
  sourceHash: string;
  translatedAt: string | null;
  fields: { title: string; description: string; category: string; brand: string } | null;
};

export type CatalogVariant = {
  sku: string;
  title: string;
  price: number;
  stock: number;
  attributes: Record<string, unknown>;
  source: string;
  sourceId: string;
  isActive: boolean;
};

export type CatalogProduct = {
  id: number;
  title: string;
  description: string;
  category: string;
  price: number;
  discountPercentage: number;
  rating: number;
  stock: number;
  brand: string;
  thumbnail: string;
  images: string[];
  source: string;
  sourceId: string;
  sourceUrl: string;
  importedAt: string;
  localization?: CatalogLocalization;
  variants?: CatalogVariant[];
};

export type CatalogResult = {
  products: CatalogProduct[];
  source: string;
  importedAt: string;
  fallback: boolean;
};
