import { renderHook } from '@testing-library/react';
import { appItem, widgetItem } from './layout';
import { registerWidget } from './registry';
import { useDeskPages } from './useDeskPages';

// 管理员把系统状态与个人设置改了地址：按组件标识仍然找得到
const mockMenus = [
  { path: '/health', component: 'state', name: 'Health' },
  { path: '/me', component: 'user/info', name: 'Me', hideInMenu: true },
  {
    path: '/sys',
    component: 'components/RouterLayout',
    name: 'System',
    routes: [{ path: '/sys/members', component: 'sys/user', name: 'Members' }],
  },
];
jest.mock('@umijs/max', () => ({
  useModel: () => ({ initialState: { menuData: mockMenus } }),
  getLocale: () => 'zh-CN',
}));
jest.mock('@/i18n', () => ({ useI18n: () => 'zh-CN' }));
jest.mock('@/i18n/menus', () => ({
  menuLabel: (m: { name: string }) => m.name,
}));
jest.mock('@/routing/componentMap', () => ({
  getComponent: (c: string) => (c ? {} : undefined),
}));

const Body = () => null;
const widget = (key: string, extra: object) =>
  registerWidget({
    key,
    group: 't',
    title: () => key,
    description: () => key,
    icon: null,
    tint: ['#000', '#000'],
    sizes: ['s'],
    Body,
    ...extra,
  });
widget('svc', { to: 'state' });
widget('acct', { to: 'user/info', public: true });
widget('orders', { to: '/orders/list' });
widget('apiPage', { to: 'sys/api' });
widget('free', {});

test('widgets and preset icons that point to a component follow the backend menus', () => {
  const { result } = renderHook(() => useDeskPages());
  const { canWidget, pathOf, resolveItems } = result.current;
  expect(pathOf('svc')).toBe('/health');
  expect(pathOf('acct')).toBe('/me');
  expect(canWidget('svc')).toBe(true);
  expect(canWidget('acct')).toBe(true);
  expect(canWidget('free')).toBe(true);
  // 没有这个页面的权限
  expect(canWidget('orders')).toBe(false);
  expect(canWidget('apiPage')).toBe(false);
  expect(canWidget('unknown')).toBe(false);

  const items = resolveItems([
    widgetItem('svc'),
    appItem('sys/user'),
    appItem('sys/api'),
    appItem('/sys/members'),
    appItem('https://example.com'),
  ]);
  expect(items.map((i) => (i.kind === 'app' ? i.path : i.widget))).toEqual([
    'svc',
    '/sys/members',
    'https://example.com',
  ]);
});

test('saved desks name pages by component so renamed menu paths keep working', () => {
  const { result } = renderHook(() => useDeskPages());
  const { toStored, resolveItems } = result.current;
  const items = [
    widgetItem('svc'),
    appItem('/sys/members'),
    appItem('/unknown'),
    appItem('https://example.com'),
  ];
  const stored = toStored(items);
  expect(stored.map((i) => (i.kind === 'app' ? i.path : i.widget))).toEqual([
    'svc',
    'sys/user',
    '/unknown',
    'https://example.com',
  ]);
  expect(
    resolveItems(stored).map((i) => (i.kind === 'app' ? i.path : i.widget)),
  ).toEqual(['svc', '/sys/members', '/unknown', 'https://example.com']);
});
