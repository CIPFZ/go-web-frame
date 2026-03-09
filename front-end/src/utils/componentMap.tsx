import React from 'react';

/**
 * 这是一个关键的映射表
 * Key: 必须是你在数据库 'component' 列中存储的字符串
 * Value: 必须是该字符串对应的 React 页面组件
 */
export const componentMap: Record<string, React.LazyExoticComponent<any>> = {
  // --- 你必须在这里为你所有的页面添加映射 ---

  // 你现有的页面
  'state': React.lazy(() => import('@/pages/state')),
  'about': React.lazy(() => import('@/pages/about')),
  'dashboard/workplace': React.lazy(() => import('@/pages/dashboard/workplace')),
  // --- 添加父级布局组件 ---
  'components/RouterLayout': React.lazy(() => import('@/components/RouterLayout')),
  // 系统管理
  'sys/user': React.lazy(() => import('@/pages/sys/user')),
  'sys/authority': React.lazy(() => import('@/pages/sys/authority')),
  'sys/menu': React.lazy(() => import('@/pages/sys/menu')),
  'sys/api': React.lazy(() => import('@/pages/sys/api')),
  'sys/operation': React.lazy(() => import('@/pages/sys/operation')),
  'sys/notice': React.lazy(() => import('@/pages/sys/notice')),
  // 个人信息
  'user/info': React.lazy(() => import('@/pages/account/settings')),

  // 古诗词管理
  'poetry/dynasty': React.lazy(() => import('@/pages/poetry/dynasty')),
  'poetry/genre': React.lazy(() => import('@/pages/poetry/genre')),
  'poetry/author': React.lazy(() => import('@/pages/poetry/author')),
  'poetry/poem': React.lazy(() => import('@/pages/poetry/poem')),
};

/**
 * 辅助函数，用于根据名称获取组件
 */
export const getComponent = (name?: string) => {
  if (!name) return undefined;
  return componentMap[name];
};
