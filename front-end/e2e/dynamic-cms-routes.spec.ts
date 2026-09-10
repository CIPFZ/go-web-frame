import { expect, test } from '@playwright/test';

type Menu = { path: string; component?: string; hideInMenu?: boolean; routes?: Menu[] };
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
