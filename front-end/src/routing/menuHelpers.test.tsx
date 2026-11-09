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

describe('route nesting follows paths, not just the menu tree', () => {
  const tree = [
    {
      path: '/vms',
      component: 'sys/user',
      // 挂在普通页面下：父页面不渲染子页面，注册成同级路由
      routes: [{ path: '/vms-agent', component: 'sys/menu' }],
    },
    {
      path: '/ops',
      component: 'components/RouterLayout',
      routes: [
        { path: '/ops/jobs', component: 'sys/api' },
        // 地址不在分组之下：同级路由
        { path: '/reports/daily', component: 'sys/notice' },
      ],
    },
  ];
  const routes = buildRoutes(tree);
  const top = routes.map((r: any) => r.path);

  it('registers children that cannot nest as siblings', () => {
    expect(top).toEqual(['/vms', '/vms-agent', '/ops', '/reports/daily']);
    expect(routes[0].children).toBeUndefined();
  });

  it('keeps real nesting and the redirect to the first nested page', () => {
    const ops = routes.find((r: any) => r.path === '/ops');
    expect(ops.children.map((r: any) => r.path ?? 'index')).toEqual(['index', '/ops/jobs']);
    expect(ops.children[0].element.props).toMatchObject({ to: '/ops/jobs' });
  });
});
