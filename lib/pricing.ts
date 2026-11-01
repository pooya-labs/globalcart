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

export function calculateDemoPrice(input: PricingInput): PricingResult {
  if (!Number.isFinite(input.sourcePrice) || input.sourcePrice < 0) throw new RangeError("sourcePrice must be a finite non-negative number");
  if (!Number.isFinite(input.fxRate) || input.fxRate <= 0) throw new RangeError("fxRate must be a finite positive number");
  const discount = Math.min(100, Math.max(0, input.discountPercentage ?? 0));
  const discountedUsd = input.sourcePrice * (1 - discount / 100);
  return {
    sourcePrice: input.sourcePrice,
    discountedUsd,
    fxRate: input.fxRate,
    toman: Math.round(discountedUsd * input.fxRate)
  };
}
