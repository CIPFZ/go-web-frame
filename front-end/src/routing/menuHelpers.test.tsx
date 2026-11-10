jest.mock('./keepAlive', () => ({ KEEP_ALIVE_GROUPS: new Set(['/reports']) }));
jest.mock('@/components/KeepAliveOutlet', () => {
  const KeepAliveOutlet = () => null;
  return { __esModule: true, default: KeepAliveOutlet };
});

import KeepAliveOutlet from '@/components/KeepAliveOutlet';
import { buildRoutes, menuGroupPaths } from './menuHelpers';

const menus = [
  { path: '/dashboard/workplace', component: 'dashboard/workplace' },
  {
    path: '/sys',
    component: 'components/RouterLayout',
    routes: [
      { path: 'https://example.com/docs', component: '' },
      { path: '/sys/user', component: 'sys/user' },
      { path: '/sys/authority', component: 'sys/authority' },
    ],
  },
  {
    path: '/reports',
    component: 'components/RouterLayout',
    routes: [{ path: '/reports/daily', component: 'sys/user' }],
  },
];

describe('menuHelpers', () => {
  it('lists menu groups (items with children) at every depth', () => {
    expect([...menuGroupPaths(menus)]).toEqual(['/sys', '/reports']);
    expect(menuGroupPaths(undefined).size).toBe(0);
  });

  it('opens the first page of a group instead of a blank group page', () => {
    const sys = buildRoutes(menus).find((r: any) => r.path === '/sys');
    const index = sys.children[0];
    expect(index.index).toBe(true);
    expect(index.element.props).toMatchObject({ to: '/sys/user', replace: true });
  });

  it('keeps the pages of registered groups alive', () => {
    const routes = buildRoutes(menus);
    const element = (path: string) => routes.find((r: any) => r.path === path).element;
    expect(element('/reports').type).toBe(KeepAliveOutlet);
    expect(element('/sys').type).not.toBe(KeepAliveOutlet);
  });
});
