export type DisplayablePrice = {
  price: number;
  priceAvailable?: boolean;
  priceKind?: "estimated" | "verified";
};

/**
 * A positive local estimate is displayable even when the source did not
 * provide a verified seller price. Legacy snapshots may omit priceKind,
 * so priceAvailable=false alone must not hide an existing estimate.
 */
export function hasDisplayPrice(product: DisplayablePrice): boolean {
  if (!Number.isFinite(product.price) || product.price <= 0) return false;
  if (product.priceKind === "verified") return product.priceAvailable !== false;
  return true;
}

/** True whenever the amount is illustrative rather than source-verified. */
export function isEstimatedPrice(product: DisplayablePrice): boolean {
  return product.priceKind === "estimated" || product.priceAvailable === false;
}

export type PricingInput = {
  sourcePrice: number;
  fxRate: number;
  discountPercentage?: number;
};

export type PricingResult = {
  sourcePrice: number;
  discountedUsd: number;
  fxRate: number;
  toman: number;
};

const USD_CENTS = 100;
const FX_SCALE = 1_000_000;
const DISCOUNT_SCALE = 10_000;
const TOMAN_DENOMINATOR = BigInt(USD_CENTS * FX_SCALE);

function toSafeScaledInteger(value: number, scale: number, field: string): bigint {
  const scaled = Math.round(value * scale);
  if (!Number.isSafeInteger(scaled)) {
    throw new RangeError(field + " exceeds the supported precision or range");
  }
  return BigInt(scaled);
}

/**
 * Calculates an illustrative storefront price using integer minor units.
 * This is deterministic demo math, not a live FX quote or financial service.
 */
export function calculateDemoPrice(input: PricingInput): PricingResult {
  if (!Number.isFinite(input.sourcePrice) || input.sourcePrice < 0) {
    throw new RangeError("sourcePrice must be a finite non-negative number");
  }
  if (!Number.isFinite(input.fxRate) || input.fxRate <= 0) {
    throw new RangeError("fxRate must be a finite positive number");
  }

  const rawDiscount = input.discountPercentage ?? 0;
  if (!Number.isFinite(rawDiscount)) {
    throw new RangeError("discountPercentage must be finite");
  }

  const discountPercentage = Math.min(100, Math.max(0, rawDiscount));
  const discountBasisPoints = BigInt(Math.round(discountPercentage * 100));
  const sourceCents = toSafeScaledInteger(input.sourcePrice, USD_CENTS, "sourcePrice");
  const fxMicros = toSafeScaledInteger(input.fxRate, FX_SCALE, "fxRate");

  if (fxMicros <= 0n) {
    throw new RangeError("fxRate is below the supported precision");
  }

  // Round discounted USD to the nearest cent before FX conversion.
  const discountedCents =
    (sourceCents * (BigInt(DISCOUNT_SCALE) - discountBasisPoints) + 5_000n) /
    BigInt(DISCOUNT_SCALE);

  // Convert cents × micro-Toman/USD to whole Toman, rounding half up.
  const tomanBigInt =
    (discountedCents * fxMicros + TOMAN_DENOMINATOR / 2n) /
    TOMAN_DENOMINATOR;

  if (tomanBigInt > BigInt(Number.MAX_SAFE_INTEGER)) {
    throw new RangeError("calculated Toman price exceeds the supported range");
  }

  return {
    sourcePrice: Number(sourceCents) / USD_CENTS,
    discountedUsd: Number(discountedCents) / USD_CENTS,
    fxRate: Number(fxMicros) / FX_SCALE,
    toman: Number(tomanBigInt)
  };
}
