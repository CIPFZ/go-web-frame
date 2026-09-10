import { currentLocale, type AppLocale } from './index';
import en from '@/locales/en-US/menu';
import zh from '@/locales/zh-CN/menu';

export interface LocalizedMenu {
  name?: string;
  nameEn?: string;
  locale?: string | false;
  routes?: LocalizedMenu[];
  children?: LocalizedMenu[];
  [key: string]: unknown;
}

export function menuLabel(
  menu: Pick<LocalizedMenu, 'name' | 'nameEn' | 'locale'>,
  locale = currentLocale(),
): string {
  if (locale === 'en-US' && menu.nameEn?.trim()) return menu.nameEn;
  const key =
    typeof menu.locale === 'string'
      ? (menu.locale as keyof typeof zh)
      : undefined;
  const builtInName = key ? zh[key] : undefined;
  if (key && (!menu.name || menu.name === key || menu.name === builtInName)) {
    return (locale === 'en-US' ? en[key] : zh[key]) || menu.name || key;
  }
  return menu.name || '';
}

export function localizeMenus(
  menus: LocalizedMenu[],
  locale: AppLocale = currentLocale(),
): any[] {
  return menus.map((menu) => ({
    ...menu,
    name: menuLabel(menu, locale),
    locale: false,
    ...(menu.routes ? { routes: localizeMenus(menu.routes, locale) } : {}),
    ...(menu.children
      ? { children: localizeMenus(menu.children, locale) }
      : {}),
  }));
}
