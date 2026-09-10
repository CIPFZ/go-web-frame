import { expect, test } from '@playwright/test';

type Menu = { path: string; component?: string; icon?: string; hideInMenu?: boolean; routes?: Menu[] };
const flatten = (menus: Menu[]): Menu[] => menus.flatMap(menu => [menu, ...flatten(menu.routes || [])]);

test.beforeEach(async ({ page, request }) => {
  const login = await request.post('/api/v1/user/login', {
    data: {
      username: process.env.E2E_ADMIN_USER || 'admin',
      password: process.env.E2E_ADMIN_PASS || 'Admin@123456',
    },
  });
  const body = await login.json();
  expect(body.code).toBe(0);
  await page.addInitScript(token => {
    localStorage.setItem('token', token);
    localStorage.setItem('umi_locale', 'zh-CN');
  }, body.data.token);
});

test('backend menus expose the basic CMS and removed pages return 404', async ({ page }) => {
  const menuResponse = page.waitForResponse(response => response.url().includes('/sys/menu/getMenu'));
  await page.goto('/#/dashboard/workplace');
  const menuBody = await (await menuResponse).json();
  expect(menuBody.code).toBe(0);
  const menus = flatten(menuBody.data);
  expect(menus.map(menu => menu.path)).toEqual(expect.arrayContaining([
    '/dashboard/workplace', '/sys/user', '/sys/authority', '/sys/menu', '/sys/api',
    '/sys/api-token', '/sys/operation', '/sys/notice', '/account/settings',
  ]));
  expect(menus.some(menu => /plugin|poetry/.test(`${menu.path} ${menu.component}`))).toBe(false);
  expect(menus.find(menu => menu.path === '/account/settings')?.hideInMenu).toBe(true);
  await expect(page.getByText('我的通知', { exact: true })).toBeVisible();
  await expect(page.getByRole('link', { name: /插件中心|诗词管理/ })).toHaveCount(0);

  await page.goto('/#/sys/api-token');
  await expect(page.getByText('API Token 控制台', { exact: true })).toBeVisible();
  await page.reload();
  await expect(page.getByText('API Token 控制台', { exact: true })).toBeVisible();
  await page.goto('/#/account/settings');
  await expect(page.getByText('个人简介', { exact: true })).toBeVisible();

  for (const path of ['/plugins', '/plugins/1', '/plugin/project/1',
    '/plugin/work-order/1/1', '/sys/plugin-master', '/poetry/poem']) {
    await page.goto(`/#${path}`);
    await expect(page.locator('.ant-result-title')).toHaveText('404');
  }
});

test('a route path supplied in the menu response determines the rendered page', async ({ page }) => {
  // Change only the response delivered to this browser. The database and static
  // route configuration do not define this path.
  await page.route('**/api/v1/sys/menu/getMenu', async route => {
    const response = await route.fetch();
    const body = await response.json();
    const menu = flatten(body.data).find(item => item.component === 'sys/api-token');
    expect(menu).toBeDefined();
    menu!.path = '/sys/dynamic-token-proof';
    await route.fulfill({ response, json: body });
  });
  await page.goto('/#/sys/dynamic-token-proof');
  await expect(page.getByText('API Token 控制台', { exact: true })).toBeVisible();
  await page.goto('/#/sys/api-token');
  await expect(page.locator('.ant-result-title')).toHaveText('404');
});


test('menu editor separates readable names from translation keys', async ({ page }) => {
  await page.goto('/#/sys/menu');
  const row = page.getByRole('row').filter({ hasText: '/dashboard/workplace' });
  await expect(row.getByText('工作台', { exact: true })).toBeVisible();
  await expect(row).not.toContainText('menu.dashboard.workplace');
  await row.getByText('编辑', { exact: true }).click();
  const dialog = page.getByRole('dialog');
  await expect(dialog.getByLabel('展示名称', { exact: true })).toHaveValue('工作台');
  await expect(dialog.locator('input[id="locale"]')).toHaveValue('menu.dashboard.workplace');
});

test('sidebar shows ordered backend menus with icons when expanded and collapsed', async ({ page }) => {
  await page.goto('/#/sys/menu');
  const sidebar = page.locator('.ant-layout-sider .ant-menu-root').first();
  const roots = sidebar.locator(':scope > .ant-menu-item > .ant-menu-title-content, :scope > .ant-menu-submenu > .ant-menu-submenu-title');
  await expect(roots).toHaveText(['工作台', '系统管理', '服务器状态', '关于']);
  const system = sidebar.locator(':scope > .ant-menu-submenu').filter({ hasText: '系统管理' });
  const children = system.locator('.ant-menu-item');
  await expect(children).toHaveText(['用户管理', '角色管理', '菜单管理', 'API 管理', 'API Token', '通知公告', '操作日志']);
  const icons = ['user', 'team', 'menu', 'api', 'key', 'notification', 'history'];
  for (const [index, icon] of icons.entries()) {
    await expect(children.nth(index).locator(`svg[data-icon="${icon}"]`)).toBeVisible();
    await expect(children.nth(index).locator('svg')).toHaveCount(1);
  }
  await expect(roots.nth(1).locator('svg[data-icon="setting"]')).toHaveCount(1);

  await page.locator('.ant-pro-sider-collapsed-button').click();
  await system.locator('.ant-menu-submenu-title').hover();
  const popup = page.locator('.ant-menu-submenu-popup:visible');
  const userLink = popup.getByRole('link', { name: '用户管理', exact: true });
  await expect(userLink.locator('svg[data-icon="user"]')).toBeVisible();
  await userLink.click();
  await expect(page).toHaveURL(/#\/sys\/user$/);
});

test('nested submenu and third-level page use icons supplied by the backend', async ({ page }) => {
  await page.route('**/api/v1/sys/menu/getMenu', async route => {
    const response = await route.fetch();
    const body = await response.json();
    const system = body.data.find((item: Menu) => item.path === '/sys');
    const user = system.routes.find((item: Menu) => item.path === '/sys/user');
    system.routes = system.routes.filter((item: Menu) => item.path !== '/sys/user');
    system.routes.unshift({
      path: '/sys/accounts', name: '账号维护', icon: 'TeamOutlined',
      component: 'components/RouterLayout',
      routes: [{ ...user, path: '/sys/accounts/users', icon: 'RocketOutlined' }],
    });
    await route.fulfill({ response, json: body });
  });
  await page.goto('/#/sys/accounts/users');
  const sidebar = page.locator('.ant-layout-sider .ant-menu-root').first();
  const group = sidebar.locator('.ant-menu-submenu-title').filter({ hasText: '账号维护' });
  await expect(group.locator('svg[data-icon="team"]')).toBeVisible();
  const user = sidebar.locator('.ant-menu-item').filter({ hasText: '用户管理' });
  await expect(user.locator('svg[data-icon="rocket"]')).toBeVisible();
  await expect(user.locator('svg')).toHaveCount(1);
});
