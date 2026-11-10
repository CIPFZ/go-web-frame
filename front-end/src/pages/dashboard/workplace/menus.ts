import { menuLabel } from '@/i18n/menus';
import { getComponent } from '@/routing/componentMap';
import type React from 'react';

export interface WorkplaceMenu {
  path: string;
  name?: string;
  nameEn?: string;
  locale?: string;
  component?: string;
  icon?: React.ReactNode;
  hideInMenu?: boolean;
  routes?: WorkplaceMenu[];
}

/** 菜单树里的全部页面（有组件的叶子），按菜单顺序。 */
export function pageMenus(
  menus: WorkplaceMenu[],
  includeHidden = false,
): WorkplaceMenu[] {
  return menus.flatMap((menu) => {
    if (
      (!includeHidden && menu.hideInMenu) ||
      !menu.path?.startsWith('/') ||
      menu.path.startsWith('//')
    )
      return [];
    if (menu.routes?.length) return pageMenus(menu.routes, includeHidden);
    return menu.component &&
      getComponent(menu.component) &&
      menu.component !== 'components/RouterLayout'
      ? [menu]
      : [];
  });
}

/** 按顶层菜单分组的页面（快捷入口的挑选列表）：有子菜单的顶层菜单是一组，顶层的单个页面归入“其他”。 */
export function pageGroups(
  menus: WorkplaceMenu[],
  otherLabel: string,
): { key: string; label: string; pages: WorkplaceMenu[] }[] {
  const groups: { key: string; label: string; pages: WorkplaceMenu[] }[] = [];
  const loose: WorkplaceMenu[] = [];
  for (const menu of menus) {
    if (menu.hideInMenu) continue;
    if (menu.routes?.length) {
      const pages = pageMenus(menu.routes);
      if (pages.length)
        groups.push({ key: menu.path, label: menuLabel(menu), pages });
    } else loose.push(...pageMenus([menu]));
  }
  if (loose.length)
    groups.push({ key: '_other', label: otherLabel, pages: loose });
  return groups;
}
