import { getIcon } from './iconMap';
import { getComponent } from './componentMap';
import {MenuDataItem} from "@ant-design/pro-components";
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

/**
 * [新函数] 递归构建 React Router v6 路由
 * 将 { path: "...", component: "state", routes: [...] }
 * 转换为 { path: "...", element: <StatePage />, children: [...] }
 */
export function buildRoutes(menuData: any[] | null): any[] {
  if (!menuData || menuData.length === 0) return [];

  return menuData.map(item => {
    // 'routes' 是你 Go 后端 JSON 标签 (json:"routes")
    const { path, component, routes, access, ...rest } = item;

    // 1. 跳过外链
    if (path.startsWith('http')) {
      return null;
    }

    // 2. 查找要渲染的 React 组件
    const Component = getComponent(component);

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
      console.warn(`[动态路由] 未找到组件映射: "${component}"。请检查 src/utils/componentMap.tsx`);
    }

    // 6. 递归处理子路由 (json:"routes" -> v6 "children")
    if (routes && routes.length > 0) {
      newRoute.children = buildRoutes(routes);
    }

    return newRoute;
  }).filter(Boolean); // 过滤掉 null (外链)
}
