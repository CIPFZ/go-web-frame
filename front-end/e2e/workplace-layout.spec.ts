import { expect, test } from '@playwright/test';

type Menu = {
  name: string;
  path: string;
  component?: string;
  hideInMenu?: boolean;
  routes?: Menu[];
};
const flatten = (menus: Menu[]): Menu[] =>
  menus.flatMap((menu) => [menu, ...flatten(menu.routes || [])]);

test.use({ timezoneId: 'Asia/Shanghai' });

test.beforeEach(async ({ page, request }) => {
  // Keep theme previews local to this browser, including SettingDrawer's sync.
  await page.route('**/sys/user/ui-config', (route) =>
    route.fulfill({ json: { code: 0 } }),
  );
  await page.route('**/sys/user/getSelfInfo', async (route) => {
    const response = await route.fetch();
    const body = await response.json();
    body.data.settings = { ...body.data.settings, navTheme: 'light' };
    await route.fulfill({ response, json: body });
  });
  const login = await (
    await request.post('/api/v1/user/login', {
      data: {
        username: process.env.E2E_ADMIN_USER || 'admin',
        password: process.env.E2E_ADMIN_PASS,
      },
    })
  ).json();
  expect(login.code).toBe(0);
  await page.addInitScript((token) => {
    localStorage.setItem('token', token);
    localStorage.setItem('umi_locale', 'zh-CN');
  }, login.data.token);
});

test('workplace follows backend menu names, paths and visibility', async ({
  page,
}) => {
  await page.route('**/sys/menu/getMenu', async (route) => {
    const response = await route.fetch();
    const body = await response.json();
    const menus = flatten(body.data);
    const findMenu = (component: string) => {
      const menu = menus.find((item) => item.component === component);
      if (!menu) throw new Error(`Missing menu: ${component}`);
      return menu;
    };
    const user = findMenu('sys/user');
    user.path = '/sys/workplace-members';
    user.name = '成员目录';
    findMenu('sys/authority').hideInMenu = true;
    findMenu('state').path = '/health-overview';
    findMenu('user/info').path = '/my-profile';
    await route.fulfill({ response, json: body });
  });
  await page.goto('/#/dashboard/workplace');
  const shortcuts = page.getByRole('navigation', { name: '工作台快捷入口' });
  await expect(
    shortcuts.getByRole('link', { name: '成员目录', exact: true }),
  ).toHaveAttribute('href', '#/sys/workplace-members');
  await expect(
    shortcuts.getByRole('link', { name: '角色管理', exact: true }),
  ).toHaveCount(0);
  await expect(shortcuts.locator('a[href="#/sys/user"]')).toHaveCount(0);
  await expect(
    page.getByRole('link', { name: '查看完整状态' }),
  ).toHaveAttribute('href', '#/health-overview');
  await expect(
    page.getByRole('link', { name: /账号设置/ }),
  ).toHaveAttribute('href', '#/my-profile');
  await shortcuts.getByRole('link', { name: '成员目录', exact: true }).click();
  await expect(page).toHaveURL(/#\/sys\/workplace-members$/);
  await expect(page.getByRole('button', { name: '新建用户' })).toBeVisible();
});

test('workplace handles empty and failed requests on desktop, mobile and dark theme', async ({
  page,
}, testInfo) => {
  let noticeFailure = false;
  await page.route('**/sys/notice/getMyNotices?**', (route) =>
    route.fulfill({
      json: noticeFailure
        ? { code: 1000, msg: 'temporary failure' }
        : { code: 0, data: { list: null, total: 0, page: 1, pageSize: 5 } },
    }),
  );
  await page.setViewportSize({ width: 1440, height: 1450 });
  await page.goto('/#/dashboard/workplace');
  await expect(page.getByText('暂时没有通知', { exact: true })).toBeVisible();
  await expect(
    page.getByRole('status').filter({ hasText: '服务运行正常' }),
  ).toBeVisible();
  await page.screenshot({
    path: testInfo.outputPath('workplace-desktop.png'),
    fullPage: true,
  });
  await page.setViewportSize({ width: 390, height: 844 });
  await expect(
    page.getByRole('heading', { name: '快捷入口', exact: true }),
  ).toBeVisible();
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= window.innerWidth,
    ),
  ).toBe(true);
  await page.screenshot({
    path: testInfo.outputPath('workplace-mobile.png'),
    fullPage: false,
  });
  await page
    .getByRole('region', { name: '当前账号', exact: true })
    .scrollIntoViewIfNeeded();
  await page.screenshot({
    path: testInfo.outputPath('workplace-mobile-bottom.png'),
  });
  noticeFailure = true;
  await page.getByRole('button', { name: '刷新通知' }).click();
  await expect(page.getByText('通知加载失败', { exact: true })).toBeVisible();
  await expect(page.getByText('暂时没有通知', { exact: true })).toHaveCount(0);
  noticeFailure = false;
  await page.getByRole('button', { name: /^重\s*试$/ }).click();
  await expect(page.getByText('暂时没有通知', { exact: true })).toBeVisible();
  await page.route('**/sys/system/getServerInfo', (route) =>
    route.fulfill({ json: { code: 1000, msg: 'temporary failure' } }),
  );
  await page.getByRole('button', { name: '刷新运行概览' }).click();
  await expect(
    page.getByRole('status').filter({ hasText: '状态待确认' }),
  ).toBeVisible();
  await expect(
    page
      .getByRole('region', { name: '运行概览', exact: true })
      .getByText('正常', { exact: true }),
  ).toHaveCount(0);
  await page.unroute('**/sys/system/getServerInfo');
  await page.route('**/sys/user/getSelfInfo', async (route) => {
    const response = await route.fetch();
    const body = await response.json();
    body.data.settings = { ...body.data.settings, navTheme: 'realDark' };
    await route.fulfill({ response, json: body });
  });
  await page.setViewportSize({ width: 1440, height: 1450 });
  await page.reload();
  await expect(page.getByText('暂时没有通知', { exact: true })).toBeVisible();
  await expect(
    page.getByRole('status').filter({ hasText: '服务运行正常' }),
  ).toBeVisible();
  await page.screenshot({
    path: testInfo.outputPath('workplace-dark.png'),
    fullPage: true,
  });
});
