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
  /** Whether the upstream dataset supplied a verified seller price. */
  priceAvailable?: boolean;
  /** Local illustrative estimate or source-verified price; never infer this from priceAvailable. */
  priceKind?: "estimated" | "verified";
  ratingAvailable?: boolean;
  stockAvailable?: boolean;
  localization?: CatalogLocalization;
  variants?: CatalogVariant[];
};

export type CatalogResult = {
  products: CatalogProduct[];
  source: string;
  importedAt: string;
  fallback: boolean;
};

/** Canonical-but-untrusted product-shaped record emitted by a source mapper. */
export type CatalogSourceRecord = Record<string, unknown>;

/** Describes whether a feed is complete or contains only incremental changes. */
export type CatalogSnapshotMode = "full" | "incremental";

/**
 * Supplier-specific adapters fetch their own payload shape and map each
 * record into the shared catalog input contract before normalization.
 */
export type CatalogSourceAdapter = {
  readonly source: string;
  readonly snapshotMode: CatalogSnapshotMode;
  fetchRecords(signal?: AbortSignal): Promise<unknown[]>;
  mapProduct(record: unknown): CatalogSourceRecord | null;
};


/** Aggregate counters for one ingestion attempt; contains no raw supplier payload. */
export type CatalogFeedImportReport = {
  source: string;
  snapshotMode: CatalogSnapshotMode;
  importedAt: string;
  fetchedRecordCount: number;
  mappedRecordCount: number;
  unmappedRecordCount: number;
  unmappedFraction: number;
  maxUnmappedFraction: number;
  normalizedProductCount: number;
  discardedAfterNormalizationCount: number;
};

/** Normalized products plus a safe, source-agnostic import diagnostics summary. */
export type CatalogFeedImportResult = {
  products: CatalogProduct[];
  report: CatalogFeedImportReport;
};
