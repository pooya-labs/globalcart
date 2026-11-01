import type { LocalizedText, TranslationProvider } from "./types";

export async function localizeText(
  text: string,
  provider: TranslationProvider | null,
  sourceLocale = "en",
  targetLocale = "fa"
): Promise<LocalizedText> {
  if (!text.trim()) {
    return { original: text, translated: text, sourceLocale, targetLocale, provider: "none", status: "unchanged", translatedAt: null };
  }
  if (sourceLocale === targetLocale) {
    return { original: text, translated: text, sourceLocale, targetLocale, provider: "none", status: "unchanged", translatedAt: null };
  }
  if (!provider) {
    return { original: text, translated: null, sourceLocale, targetLocale, provider: "not-configured", status: "unavailable", translatedAt: null };
  }
  try {
    const translated = await provider.translate(text, sourceLocale, targetLocale);
    if (!translated.trim()) throw new Error("Translation provider returned an empty result");
    return { original: text, translated, sourceLocale, targetLocale, provider: provider.name, status: "translated", translatedAt: new Date().toISOString() };
  } catch {
    return { original: text, translated: null, sourceLocale, targetLocale, provider: provider.name, status: "error", translatedAt: null };
  }
}

export { GeminiProductTranslator, parseProductTranslation } from "./gemini";
export type { ProductTranslationFields } from "./gemini";
export type { LocalizedText, TranslationProvider, TranslationStatus } from "./types";
