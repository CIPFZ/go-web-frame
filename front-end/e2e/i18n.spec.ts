import { expect, test, type Page } from '@playwright/test';

test.use({ actionTimeout: 15000 });

async function switchLanguage(page: Page, language: 'English' | '简体中文') {
  // Dispatch allows checking an already-open blocking dialog without dismissing it.
  await page
    .locator(
      'button[aria-label="切换语言"], button[aria-label="Change language"]',
    )
    .dispatchEvent('click');
  await page.getByRole('menuitem', { name: language, exact: true }).click();
  await expect(page.locator('html')).toHaveAttribute(
    'lang',
    language === 'English' ? 'en-US' : 'zh-CN',
  );
}

async function admin(page: Page, request: any) {
  const result = await (
    await request.post('/api/v1/user/login', {
      data: { username: 'admin', password: process.env.E2E_ADMIN_PASS },
    })
  ).json();
  expect(result.code).toBe(0);
  await page.addInitScript((token) => {
    localStorage.setItem('token', token);
    if (!localStorage.getItem('umi_locale'))
      localStorage.setItem('umi_locale', 'zh-CN');
  }, result.data.token);
  await page.route('**/sys/user/ui-config', (route) =>
    route.fulfill({ json: { code: 0 } }),
  );
}

test('login language switches without reload, preserves input and translates validation and API errors', async ({
  page,
}) => {
  await page.goto('/#/user/login');
  await page.getByRole('button', { name: /登\s*录/ }).click();
  await expect(
    page.getByText('用户名是必填项！', { exact: true }),
  ).toBeVisible();
  await page.evaluate(
    () => ((window as any).__languageMarker = 'same-document'),
  );
  await switchLanguage(page, 'English');
  await expect(
    page.getByText('Please input your username!', { exact: true }),
  ).toBeVisible();
  await page
    .getByPlaceholder('Username', { exact: true })
    .fill('i18n-invalid-user');
  await page
    .getByPlaceholder('Password', { exact: true })
    .fill('invalid-password');
  const responsePromise = page.waitForResponse((r) =>
    r.url().endsWith('/user/login'),
  );
  await page.getByRole('button', { name: /Login/ }).click();
  const response = await responsePromise;
  expect(response.request().headers()['accept-language']).toBe('en-US');
  expect((await response.json()).msg).toBe('Incorrect username or password');
  await expect(
    page.getByText('Incorrect username or password', { exact: true }),
  ).toBeVisible();
  await switchLanguage(page, '简体中文');
  await expect(page.getByPlaceholder('用户名', { exact: true })).toHaveValue(
    'i18n-invalid-user',
  );
  expect(await page.evaluate(() => (window as any).__languageMarker)).toBe(
    'same-document',
  );
  await switchLanguage(page, 'English');
  await page.reload();
  await expect(
    page.getByPlaceholder('Username', { exact: true }),
  ).toBeVisible();
  await page.goto('/#/user/register');
  await expect(
    page.getByText(
      'Registration is closed. Contact an administrator to create an account.',
      { exact: true },
    ),
  ).toBeVisible();
  await switchLanguage(page, '简体中文');
  await expect(
    page.getByText('注册已关闭，请联系管理员创建账号', { exact: true }),
  ).toBeVisible();
});

test('every CMS page and its creation dialogs respond to live language changes', async ({
  page,
  request,
}, info) => {
  test.setTimeout(180_000);
  await admin(page, request);
  const pages = [
    ['/dashboard/workplace', '快捷入口', 'Quick access', '', ''],
    ['/state', 'Go 运行环境', 'Go runtime', '', ''],
    ['/about', '技术栈', 'Technology stack', '', ''],
    ['/sys/user', '用户管理', 'Users', '新建用户', 'New user'],
    ['/sys/authority', '角色名称', 'Role name', '新增角色', 'New role'],
    [
      '/sys/menu',
      '英文展示名称',
      'English display name',
      '新建菜单',
      'New menu',
    ],
    ['/sys/api', 'API 描述', 'API description', '新建 API', 'New API'],
    [
      '/sys/api-token',
      'API Token 控制台',
      'API token console',
      '新建 Token',
      'New token',
    ],
    ['/sys/notice', '发布通知', 'Publish notice', '发布通知', 'Publish notice'],
    ['/sys/operation', '操作时间', 'Operation time', '', ''],
    ['/account/settings', '个人简介', 'Bio', '', ''],
  ];
  for (const [path, zh, en, createZh, createEn] of pages) {
    await page.goto('/#' + path);
    await expect(
      page.getByText(zh, { exact: true }).filter({ visible: true }).first(),
    ).toBeVisible();
    await switchLanguage(page, 'English');
    await expect(
      page.getByText(en, { exact: true }).filter({ visible: true }).first(),
    ).toBeVisible();
    if (createEn) {
      await page
        .getByRole('button', {
          name: new RegExp(path === '/sys/menu' ? 'New root menu' : createEn),
        })
        .click();
      const dialog = page.getByRole('dialog');
      await expect(dialog).toBeVisible();
      await switchLanguage(page, '简体中文');
      await expect(
        dialog.getByText(createZh, { exact: true }).first(),
      ).toBeVisible();
      await switchLanguage(page, 'English');
      await expect(
        dialog.getByText(createEn, { exact: true }).first(),
      ).toBeVisible();
      await page.screenshot({
        path: info.outputPath(path.replaceAll('/', '-') + '-en.png'),
      });
      await dialog
        .getByRole('button', { name: /Cancel|Close/ })
        .first()
        .click();
    }
    await switchLanguage(page, '简体中文');
  }
  await page.goto('/#/unknown-page');
  await switchLanguage(page, 'English');
  await expect(
    page.getByText('Sorry, the page you visited does not exist.'),
  ).toBeVisible();
});

test('open forms preserve entered values and update existing errors', async ({
  page,
  request,
}) => {
  await admin(page, request);
  await page.goto('/#/sys/user');
  await page.getByRole('button', { name: '新建用户' }).click();
  const dialog = page.getByRole('dialog');
  await dialog.getByPlaceholder('登录账号', { exact: true }).fill('keep-input');
  await dialog.getByRole('button', { name: /确\s*[认定]|提\s*交/ }).click();
  await expect(dialog.getByText('请输入密码', { exact: true })).toBeVisible();
  await switchLanguage(page, 'English');
  await expect(
    dialog.getByPlaceholder('Login account', { exact: true }),
  ).toHaveValue('keep-input');
  await expect(
    dialog.getByText('Enter a password.', { exact: true }),
  ).toBeVisible();
  await expect(dialog.getByText('请输入密码', { exact: true })).toHaveCount(0);
});

test('permission tree translates loaded menu nodes while preserving selection', async ({
  page,
  request,
}) => {
  await admin(page, request);
  await page.goto('/#/sys/authority');
  await page.getByText('设置权限', { exact: true }).first().click();
  const dialog = page.getByRole('dialog');
  await expect(
    dialog.getByText('工作台 (/dashboard/workplace)', { exact: true }),
  ).toBeVisible();
  const checkedBefore = await dialog
    .locator('.ant-tree-checkbox-checked')
    .count();
  await switchLanguage(page, 'English');
  await expect(
    dialog.getByText('Workplace (/dashboard/workplace)', { exact: true }),
  ).toBeVisible();
  expect(await dialog.locator('.ant-tree-checkbox-checked').count()).toBe(
    checkedBefore,
  );
  await expect(
    dialog.getByRole('tab', { name: 'Menu permissions' }),
  ).toBeVisible();
});

test('English dynamic menu names and paths stay controlled by backend data', async ({
  page,
  request,
}, info) => {
  await admin(page, request);
  await page.route('**/sys/menu/getMenu', async (route) => {
    const response = await route.fetch();
    const body = await response.json();
    const walk = (items: any[]): any[] =>
      items.flatMap((item) => [item, ...walk(item.routes || [])]);
    const menu = walk(body.data).find((item) => item.component === 'sys/user');
    menu.name = '团队成员';
    menu.nameEn = 'Team members';
    menu.path = '/sys/team-members';
    await route.fulfill({ response, json: body });
  });
  await page.goto('/#/dashboard/workplace');
  await switchLanguage(page, 'English');
  const nav = page.getByRole('navigation', { name: 'Workplace shortcuts' });
  await expect(
    nav.getByRole('link', { name: 'Team members', exact: true }),
  ).toHaveAttribute('href', '#/sys/team-members');
  await page.setViewportSize({ width: 390, height: 844 });
  await expect
    .poll(async () =>
      page.evaluate(() => document.documentElement.scrollWidth <= innerWidth),
    )
    .toBe(true);
  await page.screenshot({
    path: info.outputPath('workplace-en-mobile.png'),
    fullPage: true,
  });
  await page.setViewportSize({ width: 1440, height: 1100 });
  await nav.getByRole('link', { name: 'Team members', exact: true }).click();
  await expect(page.getByRole('button', { name: 'New user' })).toBeVisible();
});
