import { getIntl, getLocale, useIntl } from '@umijs/max';
import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';
import 'dayjs/locale/en';
import { useEffect } from 'react';
import type messages from '@/locales/zh-CN/cms';

export type AppLocale = 'zh-CN' | 'en-US';
export const normalizeLocale = (locale?: string): AppLocale =>
  locale?.toLowerCase().startsWith('en') ? 'en-US' : 'zh-CN';
export const currentLocale = (): AppLocale => normalizeLocale(getLocale());

// Resolve at call time so async actions also use the selected language.
export function t(
  id: keyof typeof messages,
  values?: Record<string, string | number | boolean | null | undefined>,
): string {
  return getIntl(currentLocale()).formatMessage({ id }, values);
}

// Every localized component subscribes to Umi's locale context. No reload or
// component key is needed, so changing language preserves forms and drawers.
export function useI18n(): AppLocale {
  const intl = useIntl();
  const locale = normalizeLocale(intl.locale);
  useEffect(() => {
    document.documentElement.lang = locale;
    dayjs.locale(locale === 'zh-CN' ? 'zh-cn' : 'en');
  }, [locale]);
  return locale;
}

export function formatDate(
  value: string | number | Date,
  options?: Intl.DateTimeFormatOptions,
): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '—';
  return new Intl.DateTimeFormat(
    currentLocale(),
    options ?? {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
    },
  ).format(date);
}
