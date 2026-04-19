export const publicPluginRoutePatterns = [/^\/plugins$/, /^\/plugins\/[^/]+$/];

export const isPublicPluginRoute = (pathname?: string) =>
  Boolean(pathname && publicPluginRoutePatterns.some((pattern) => pattern.test(pathname)));

export const isEnglishLocale = (locale?: string) => locale?.toLowerCase().startsWith('en') ?? false;

const isMeaningfulText = (value?: string) => {
  const normalized = (value || '').trim();
  if (!normalized) return false;
  return !/^[?？�]+$/.test(normalized);
};

export const pickLocaleText = (locale: string | undefined, zh?: string, en?: string, fallback = '-') => {
  const primary = isEnglishLocale(locale) ? en : zh;
  const secondary = isEnglishLocale(locale) ? zh : en;
  if (isMeaningfulText(primary)) return primary!.trim();
  if (isMeaningfulText(secondary)) return secondary!.trim();
  return fallback.trim();
};

export const getDisplayName = (locale: string | undefined, item?: { nameZh?: string; nameEn?: string }) =>
  pickLocaleText(locale, item?.nameZh, item?.nameEn);

export const getDisplayDescription = (
  locale: string | undefined,
  item?: { descriptionZh?: string; descriptionEn?: string },
) => pickLocaleText(locale, item?.descriptionZh, item?.descriptionEn);

export const getDisplayChangelog = (
  locale: string | undefined,
  item?: { changelogZh?: string; changelogEn?: string },
) => pickLocaleText(locale, item?.changelogZh, item?.changelogEn);

export const getDisplayOfflineReason = (
  locale: string | undefined,
  item?: { offlineReasonZh?: string; offlineReasonEn?: string },
) => pickLocaleText(locale, item?.offlineReasonZh, item?.offlineReasonEn);
