export type ProductTranslationFields = {
  title: string;
  description: string;
  category: string;
  brand: string;
};

export function parseGoogleTranslateResponse(raw: unknown): ProductTranslationFields {
  if (!raw || typeof raw !== "object") throw new Error("Google Translate bridge returned an invalid response");
  const payload = raw as { ok?: unknown; fields?: Record<string, unknown> };
  if (payload.ok !== true || !payload.fields) throw new Error("Google Translate bridge reported a translation error");
  for (const field of ["title", "description", "category", "brand"] as const) {
    if (typeof payload.fields[field] !== "string" || !(payload.fields[field] as string).trim()) {
      throw new Error(`Google Translate bridge returned an invalid ${field}`);
    }
  }
  return {
    title: (payload.fields.title as string).trim(),
    description: (payload.fields.description as string).trim(),
    category: (payload.fields.category as string).trim(),
    brand: (payload.fields.brand as string).trim()
  };
}

export class GoogleAppsScriptTranslator {
  readonly name = "Google Translate (Apps Script Language Service)";

  constructor(private readonly endpoint: string, private readonly token: string) {
    if (!/^https:\/\/script\.google\.com\/macros\/s\/.+\/exec(?:\?.*)?$/.test(endpoint)) {
      throw new Error("GOOGLE_TRANSLATE_ENDPOINT must be a deployed Google Apps Script web app URL");
    }
    if (token.length < 24) throw new Error("GOOGLE_TRANSLATE_TOKEN must contain at least 24 characters");
  }

  async translate(fields: ProductTranslationFields): Promise<ProductTranslationFields> {
    let lastError: unknown;
    for (let attempt = 0; attempt < 3; attempt += 1) {
      try {
        const response = await fetch(this.endpoint, {
          method: "POST",
          headers: { "content-type": "application/json" },
          body: JSON.stringify({ token: this.token, fields }),
          signal: AbortSignal.timeout(45_000)
        });
        if (!response.ok) throw new Error(`Google Translate bridge returned HTTP ${response.status}`);
        return parseGoogleTranslateResponse(await response.json());
      } catch (error) {
        lastError = error;
        if (attempt < 2) await new Promise(resolve => setTimeout(resolve, 1_000 * (2 ** attempt)));
      }
    }
    throw lastError instanceof Error ? lastError : new Error("Google Translate request failed");
  }
}
