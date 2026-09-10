import { localizeMenus, menuLabel } from './menus';

test('localizes dynamic labels without changing backend routes, permissions or data', () => {
  const menus = [
    {
      name: '系统管理',
      locale: 'menu.system',
      path: '/configured',
      access: 'admin',
      routes: [
        {
          name: '工作台',
          nameEn: 'Team workspace',
          path: '/chosen',
          locale: 'menu.dashboard.workplace',
        },
      ],
    },
  ];
  const en = localizeMenus(menus, 'en-US');
  expect(en[0].path).toBe('/configured');
  expect(en[0].access).toBe('admin');
  expect(en[0].routes[0].name).toBe('Team workspace');
  expect(en[0].routes[0].path).toBe('/chosen');
  expect(menus[0].routes[0].name).toBe('工作台');
  expect(
    menuLabel({ name: '工作台', locale: 'menu.dashboard.workplace' }, 'en-US'),
  ).toBe('Workplace');
  expect(
    menuLabel(
      { name: '自定义内容', locale: 'menu.dashboard.workplace' },
      'en-US',
    ),
  ).toBe('自定义内容');
  expect(menuLabel({ name: '自定义内容', nameEn: 'Custom' }, 'zh-CN')).toBe(
    '自定义内容',
  );
});
