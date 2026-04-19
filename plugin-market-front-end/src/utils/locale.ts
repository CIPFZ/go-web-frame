export type MarketLocale = 'zh' | 'en';

const storageKey = 'plugin-market-standalone-locale';

export function getLocale(): MarketLocale {
  if (typeof window === 'undefined') {
    return 'zh';
  }
  const stored = window.localStorage.getItem(storageKey);
  if (stored === 'zh' || stored === 'en') {
    return stored;
  }
  return navigator.language.toLowerCase().startsWith('en') ? 'en' : 'zh';
}

export function setLocale(value: MarketLocale) {
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(storageKey, value);
  }
}

export function pickText(locale: MarketLocale, zh?: string, en?: string) {
  if (locale === 'en') {
    return en || zh || '-';
  }
  return zh || en || '-';
}
