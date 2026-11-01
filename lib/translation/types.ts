export type TranslationStatus = "translated" | "unchanged" | "unavailable" | "error";

export type LocalizedText = {
  original: string;
  translated: string | null;
  sourceLocale: string;
  targetLocale: string;
  provider: string;
  status: TranslationStatus;
  translatedAt: string | null;
};

export interface TranslationProvider {
  readonly name: string;
  translate(text: string, sourceLocale: string, targetLocale: string): Promise<string>;
}
