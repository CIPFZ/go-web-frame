import { getIcon } from '../utils/iconMap';
import KeepAliveOutlet from '@/components/KeepAliveOutlet';
import { getComponent } from './componentMap';
import { KEEP_ALIVE_GROUPS } from './keepAlive';
import {MenuDataItem} from "@ant-design/pro-components";
import { Navigate } from '@umijs/max';
import React from "react";

// @ts-ignore
/**
 * 递归处理菜单数据，将 'icon' 字符串转换为 ReactNode
 * @param menuData 从后端获取的原始菜单树
 * @returns 供 ProLayout 使用的、包含 ReactNode 图标的菜单树
 */
export function processMenuData(menuData: any[] | null): any[] {
  if (!menuData || menuData.length === 0) {
    return [];
  }

  return menuData.map((item) => {
    const { icon, routes, ...rest } = item;

    // 转换 icon
    const newItem = {
      ...rest,
      icon: getIcon(icon),
    };

    // 递归处理子菜单
    if (routes && routes.length > 0) {
      newItem.routes = processMenuData(routes);
    }

    return newItem;
  });
}

/** 有子菜单的分组路径（如 /sys）：它们没有自己的页面，面包屑里只显示文字、不可点击。 */
export function menuGroupPaths(menuData: any[] | null | undefined, out = new Set<string>()): Set<string> {
  for (const item of menuData || []) {
    if (item.routes?.length) {
      out.add(item.path);
      menuGroupPaths(item.routes, out);
    }
  }
  return out;
}

/** 子路由的 path 是否在父路由之下（相对路径总是在）。 */
const isUnder = (child: string | undefined, parent: string) =>
  !child?.startsWith('/') || child.startsWith(parent.endsWith('/') ? parent : `${parent}/`);

/**
 * 递归构建 React Router v6 路由：把后端菜单 { path, component, routes } 转成 { path, element, children }。
 *
 * 菜单树是给侧栏看的，管理员可以把任何菜单挂到任何菜单下面；路由只在真正嵌套时才嵌套：
 * 父菜单是布局分组（components/RouterLayout 或保活分组，它们渲染子页面），且子菜单的地址在父菜单地址之下。
 * 其余的子菜单（挂在普通页面下、或地址不在父菜单之下）注册成同级路由，页面照常打开。
 */
export function buildRoutes(menuData: any[] | null): any[] {
  if (!menuData || menuData.length === 0) return [];

  return menuData.flatMap((item) => {
    // 'routes' 是你 Go 后端 JSON 标签 (json:"routes")
    const { path, component, routes, access, ...rest } = item;

    // 1. 跳过外链
    if (path.startsWith('http')) {
      return [];
    }

    // 2. 查找要渲染的 React 组件（保活分组用 KeepAliveOutlet 代替普通的 Outlet）
    const keepAlive = KEEP_ALIVE_GROUPS.has(path) && routes?.length;
    const Component = keepAlive ? KeepAliveOutlet : getComponent(component);
    const isLayout = keepAlive || component === 'components/RouterLayout';

    // 3. 构建 React Router v6 路由对象
    const newRoute: any = {
      path: path,
      // 关键：将 component 字符串 转换为 element 属性
      element: Component ? React.createElement(Component) : null,
      // 传递所有其他属性 (name, icon, hideInMenu 等)
      ...rest,
    };

    // 4. 处理权限 (修复 403 问题)
    // 只有当 access 字段 *不是* 空字符串时，才将其添加到路由中
    if (access) {
      newRoute.access = access;
    }

    // 5. 警告缺失的映射
    if (component && !Component) {
      console.warn(`[动态路由] 未找到组件映射: "${component}"。请检查 src/routing/componentMap.tsx`);
    }

    // 6. 递归处理子路由 (json:"routes" -> v6 "children")；不能嵌套的子菜单注册成同级路由
    const nested: any[] = [];
    const siblings: any[] = [];
    for (const child of routes || []) {
      (isLayout && isUnder(child.path, path) ? nested : siblings).push(child);
    }
    if (nested.length > 0) {
      newRoute.children = buildRoutes(nested);
      // 菜单分组（如“系统管理”）本身没有页面：直接访问分组地址时跳到第一个子页面，而不是一片空白
      const first = nested.find((r: any) => r.path && !r.path.startsWith('http'));
      if (first) {
        newRoute.children.unshift({ index: true, element: React.createElement(Navigate, { to: first.path, replace: true }) });
      }
    }

    return [newRoute, ...buildRoutes(siblings)];
  });
}
