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
};

export type CatalogResult = {
  products: CatalogProduct[];
  source: string;
  importedAt: string;
  fallback: boolean;
};
