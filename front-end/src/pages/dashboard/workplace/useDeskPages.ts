import { useI18n } from '@/i18n';
import { useModel } from '@umijs/max';
import { useMemo } from 'react';
import type { DeskItem } from './layout';
import { pageMenus, type WorkplaceMenu } from './menus';
import { deskLinks, widgetDef } from './registry';

/**
 * 桌面能放的页面与小组件，和菜单权限绑定：默认按当前登录人的菜单（工作台本身）；
 * 传入 menus 时按这些菜单（“角色管理 → 默认工作台”传这个角色的菜单）。
 * 外部链接人人可用，和页面一样当作应用图标；小组件要有它对应页面的权限，不对应页面的（问候、通知）与标了 public 的（当前账号）人人都有。
 */
export function useDeskPages(only?: WorkplaceMenu[]) {
  const locale = useI18n();
  const { initialState } = useModel('@@initialState');
  const menus = (only ?? initialState?.menuData ?? []) as WorkplaceMenu[];
  const pages = useMemo(() => pageMenus(menus), [menus]);
  const allPages = useMemo(() => pageMenus(menus, true), [menus]);
  const links = useMemo(
    () =>
      deskLinks().map((l) => ({ path: l.url, name: l.name(), icon: l.icon })),
    // 名称随语言变化
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [locale],
  );
  const pageByPath = useMemo(
    () =>
      new Map<string, WorkplaceMenu>(
        [...pages, ...links].map((p) => [p.path, p]),
      ),
    [pages, links],
  );
  // 组件标识 → 页面（含隐藏菜单）：预置与小组件可以按组件指向页面，管理员改了菜单地址也能找到
  const byComponent = useMemo(() => {
    const out = new Map<string, WorkplaceMenu>();
    for (const p of allPages)
      if (p.component && !out.has(p.component)) out.set(p.component, p);
    return out;
  }, [allPages]);
  const resolve = (target: string) =>
    target.startsWith('/') || target.startsWith('https://')
      ? target
      : byComponent.get(target)?.path;
  /** 小组件打开的页面的实际地址；没有这个页面的权限时为空。 */
  const pathOf = (key: string) => {
    const to = widgetDef(key)?.to;
    return to ? resolve(to) : undefined;
  };
  const canWidget = (key: string) => {
    const def = widgetDef(key);
    if (!def) return false;
    if (!def.to || def.public) return true;
    const path = pathOf(key);
    return !!path && allPages.some((p) => p.path === path);
  };
  /** 把应用图标里的组件标识换成实际地址；找不到页面的、换完以后重复的去掉。 */
  const resolveItems = (items: DeskItem[]): DeskItem[] => {
    const seen = new Set<string>();
    return items.flatMap((i): DeskItem[] => {
      if (i.kind !== 'app') return [i];
      const path = resolve(i.path);
      if (!path || seen.has(path)) return [];
      seen.add(path);
      return [{ ...i, path }];
    });
  };
  /** 保存前把页面换成组件标识（组件只对应这一个页面时）：管理员之后改了菜单地址，保存的图标也找得到。 */
  const toStored = (items: DeskItem[]): DeskItem[] =>
    items.map((i) => {
      if (i.kind !== 'app' || !i.path.startsWith('/')) return i;
      const component = allPages.find((p) => p.path === i.path)?.component;
      return component && byComponent.get(component)?.path === i.path
        ? { ...i, path: component }
        : i;
    });
  return {
    menus,
    pages,
    allPages,
    links,
    pageByPath,
    canWidget,
    pathOf,
    resolveItems,
    toStored,
  };
}
