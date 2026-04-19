import React from 'react';

export const componentMap: Record<string, React.LazyExoticComponent<any>> = {
  'state': React.lazy(() => import('@/pages/state')),
  'about': React.lazy(() => import('@/pages/about')),
  'dashboard/workplace': React.lazy(() => import('@/pages/dashboard/workplace')),
  'components/RouterLayout': React.lazy(() => import('@/components/RouterLayout')),

  'sys/user': React.lazy(() => import('@/pages/sys/user')),
  'sys/authority': React.lazy(() => import('@/pages/sys/authority')),
  'sys/menu': React.lazy(() => import('@/pages/sys/menu')),
  'sys/api': React.lazy(() => import('@/pages/sys/api')),
  'sys/api-token': React.lazy(() => import('@/pages/sys/api-token')),
  'sys/operation': React.lazy(() => import('@/pages/sys/operation')),
  'sys/notice': React.lazy(() => import('@/pages/sys/notice')),

  'plugin/project-management': React.lazy(() => import('@/pages/plugin/project-management')),
  'plugin/project-detail': React.lazy(() => import('@/pages/plugin/project-detail')),
  'plugin/work-order-pool': React.lazy(() => import('@/pages/plugin/work-order-pool')),
  'plugin/work-order-detail': React.lazy(() => import('@/pages/plugin/work-order-detail')),
  'sys/plugin-master': React.lazy(() => import('@/pages/sys/plugin-master')),

  'user/info': React.lazy(() => import('@/pages/account/settings')),

  'poetry/dynasty': React.lazy(() => import('@/pages/poetry/dynasty')),
  'poetry/genre': React.lazy(() => import('@/pages/poetry/genre')),
  'poetry/author': React.lazy(() => import('@/pages/poetry/author')),
  'poetry/poem': React.lazy(() => import('@/pages/poetry/poem')),
};

export const getComponent = (name?: string) => {
  if (!name) return undefined;
  return componentMap[name];
};
