export type Locale = "fa" | "en";

export const LOCALE_PREFERENCE_KEY = "globalcart-locale-v1";

export function parseLocalePreference(value: string | null): Locale | null {
  return value === "fa" || value === "en" ? value : null;
}

export function readLocalePreference(): Locale | null {
  if (typeof window === "undefined") return null;
  try {
    return parseLocalePreference(window.localStorage.getItem(LOCALE_PREFERENCE_KEY));
  } catch {
    return null;
  }
}

export function writeLocalePreference(locale: Locale): void {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(LOCALE_PREFERENCE_KEY, locale);
  } catch {
    // The storefront remains usable when browser storage is unavailable.
  }
}
