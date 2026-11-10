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

test('desk follows backend menu names, paths and visibility; edits are saved per user', async ({
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
  // 预置的图标按组件找到页面：改了名字和地址也在，隐藏的菜单不出现
  const members = page.getByRole('link', { name: '成员目录', exact: true });
  await expect(members).toHaveAttribute('href', '#/sys/workplace-members');
  await expect(
    page.getByRole('link', { name: '角色管理', exact: true }),
  ).toHaveCount(0);
  await expect(
    page
      .getByRole('region', { name: '运行概览', exact: true })
      .getByRole('link', { name: '运行概览' }),
  ).toHaveAttribute('href', '#/health-overview');

  // 编辑：图标不是链接；从组件库添加“当前账号”，完成后保存在本人账号上
  await page.getByRole('button', { name: '编辑工作台' }).click();
  await expect(members).toHaveCount(0);
  await page.getByRole('button', { name: '添加小组件' }).click();
  const gallery = page.getByRole('dialog');
  await gallery
    .getByRole('region', { name: '当前账号', exact: true })
    .getByRole('button', { name: /添\s*加/ })
    .click();
  await page.keyboard.press('Escape');
  const saved = page.waitForResponse(
    (r) => r.url().endsWith('/sys/user/workplace') && r.request().method() === 'PUT',
  );
  await page.getByRole('button', { name: /完\s*成/ }).click();
  expect((await (await saved).json()).code).toBe(0);
  const account = page.getByRole('region', { name: '当前账号', exact: true });
  await expect(account.getByRole('link', { name: '当前账号' })).toHaveAttribute(
    'href',
    '#/my-profile',
  );
  await page.reload();
  await expect(account).toBeVisible();

  // 恢复默认：回到角色的默认工作台
  await page.getByRole('button', { name: '编辑工作台' }).click();
  await page.getByRole('button', { name: '恢复默认' }).click();
  await page.getByRole('button', { name: /完\s*成/ }).click();
  await expect(account).toHaveCount(0);

  await members.click();
  await expect(page).toHaveURL(/#\/sys\/workplace-members$/);
  await expect(page.getByRole('button', { name: '新建用户' })).toBeVisible();
});

test('desk handles empty and failed requests on desktop, mobile and dark theme', async ({
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
  const notices = page.getByRole('region', { name: '我的通知', exact: true });
  const service = page.getByRole('region', { name: '运行概览', exact: true });
  await page.setViewportSize({ width: 1440, height: 1450 });
  await page.goto('/#/dashboard/workplace');
  await expect(notices.getByText('暂时没有通知', { exact: true })).toBeVisible();
  await expect(service.getByText('正常', { exact: true }).first()).toBeVisible();
  await page.screenshot({
    path: testInfo.outputPath('workplace-desktop.png'),
    fullPage: true,
  });
  await page.setViewportSize({ width: 390, height: 844 });
  await expect(notices).toBeVisible();
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= window.innerWidth,
    ),
  ).toBe(true);
  await page.screenshot({
    path: testInfo.outputPath('workplace-mobile.png'),
    fullPage: true,
  });

  noticeFailure = true;
  await page.reload();
  await expect(notices.getByText('通知加载失败', { exact: true })).toBeVisible();
  await expect(notices.getByText('暂时没有通知', { exact: true })).toHaveCount(0);
  noticeFailure = false;
  await notices.getByRole('button', { name: /重\s*试/ }).click();
  await expect(notices.getByText('暂时没有通知', { exact: true })).toBeVisible();

  await page.route('**/sys/system/getServerInfo', (route) =>
    route.fulfill({ json: { code: 1000, msg: 'temporary failure' } }),
  );
  await page.reload();
  await expect(service.getByText('状态待确认', { exact: true })).toBeVisible();
  await expect(service.getByText('正常', { exact: true })).toHaveCount(0);
  await page.unroute('**/sys/system/getServerInfo');
  await service.getByRole('button', { name: /重\s*试/ }).click();
  await expect(service.getByText('正常', { exact: true }).first()).toBeVisible();

  await page.route('**/sys/user/getSelfInfo', async (route) => {
    const response = await route.fetch();
    const body = await response.json();
    body.data.settings = { ...body.data.settings, navTheme: 'realDark' };
    await route.fulfill({ response, json: body });
  });
  await page.setViewportSize({ width: 1440, height: 1450 });
  await page.reload();
  await expect(notices.getByText('暂时没有通知', { exact: true })).toBeVisible();
  await page.screenshot({
    path: testInfo.outputPath('workplace-dark.png'),
    fullPage: true,
  });
});
