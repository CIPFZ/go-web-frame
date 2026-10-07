import React from 'react';

export const componentMap: Record<string, React.LazyExoticComponent<any>> = {
  state: React.lazy(() => import('@/pages/state')),
  about: React.lazy(() => import('@/pages/about')),
  'dashboard/workplace': React.lazy(() => import('@/pages/dashboard/workplace')),
  'components/RouterLayout': React.lazy(() => import('@/components/RouterLayout')),

  'sys/user': React.lazy(() => import('@/pages/sys/user')),
  'sys/authority': React.lazy(() => import('@/pages/sys/authority')),
  'sys/menu': React.lazy(() => import('@/pages/sys/menu')),
  'sys/api': React.lazy(() => import('@/pages/sys/api')),
  'sys/api-token': React.lazy(() => import('@/pages/sys/api-token')),
  'sys/operation': React.lazy(() => import('@/pages/sys/operation')),
  'sys/notice': React.lazy(() => import('@/pages/sys/notice')),
  'novel/books': React.lazy(() => import('@/pages/novel/books')),
  'virtualization/agent': React.lazy(() => import('@/pages/virtualization/agent')),
  'virtualization/vms': React.lazy(() => import('@/pages/virtualization/vms')),
  'virtualization/storage': React.lazy(() => import('@/pages/virtualization/storage')),

  'user/info': React.lazy(() => import('@/pages/account/settings')),
};

export const getComponent = (name?: string) => {
  if (!name) return undefined;
  return componentMap[name];
};
