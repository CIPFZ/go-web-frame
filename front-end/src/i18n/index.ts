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

// 格式化器创建很慢（图表里每个点都要格式化一次）：按语言与选项缓存
const dateFormats = new Map<string, Intl.DateTimeFormat>();

export function formatDate(
  value: string | number | Date,
  options?: Intl.DateTimeFormatOptions,
): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '—';
  const locale = currentLocale();
  const key = `${locale}|${options ? JSON.stringify(options) : ''}`;
  let format = dateFormats.get(key);
  if (!format) {
    format = new Intl.DateTimeFormat(
      locale,
      options ?? {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        hour12: false,
      },
    );
    dateFormats.set(key, format);
  }
  return format.format(date);
}
