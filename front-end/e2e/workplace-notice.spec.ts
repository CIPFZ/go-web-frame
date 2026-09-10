import {
  type APIRequestContext,
  expect,
  request,
  test,
} from '@playwright/test';

test.skip(
  process.env.E2E_ISOLATED !== '1',
  'Creates test users; run with deploy/local/e2e.py in a disposable database.',
);

const ADMIN_USER = process.env.E2E_ADMIN_USER || 'admin';
const ADMIN_PASS = process.env.E2E_ADMIN_PASS || 'Admin@123456';

async function apiLogin(
  api: APIRequestContext,
  apiBase: string,
  username: string,
  password: string,
): Promise<string> {
  const resp = await api.post(`${apiBase}/api/v1/user/login`, {
    data: { username, password },
  });
  const body = await resp.json();
  expect(body.code).toBe(0);
  return body.data.token;
}

test('workplace respects role menus and supports notice pagination and confirmation', async ({
  page,
}, testInfo) => {
  const apiBase = process.env.E2E_API_BASE || 'http://127.0.0.1:8080';
  const bootstrapApi = await request.newContext();

  const adminToken = await apiLogin(
    bootstrapApi,
    apiBase,
    ADMIN_USER,
    ADMIN_PASS,
  );
  const adminApi = await request.newContext({
    extraHTTPHeaders: { 'x-token': adminToken },
  });

  const stamp = Date.now();
  const username = `e2e_user_${stamp}`;
  const password = 'E2E@123456';
  const noticeTitle = `E2E Notice ${stamp}`;

  const addUser = await adminApi.post(`${apiBase}/api/v1/sys/user/addUser`, {
    data: {
      username,
      password,
      nickName: 'E2E User',
      authorityIds: [888],
      status: 1,
    },
  });
  const addUserBody = await addUser.json();
  expect(addUserBody.code).toBe(0);

  const userList = await adminApi.post(
    `${apiBase}/api/v1/sys/user/getUserList`,
    {
      data: { page: 1, pageSize: 20, username },
    },
  );
  const userListBody = await userList.json();
  expect(userListBody.code).toBe(0);
  const userId = userListBody.data?.list?.[0]?.ID;
  expect(userId).toBeTruthy();

  for (let index = 0; index < 5; index++) {
    const extra = await adminApi.post(
      `${apiBase}/api/v1/sys/notice/createNotice`,
      {
        data: {
          title: `Earlier notice ${stamp} ${index}`,
          content: 'Earlier directed notice',
          level: 'info',
          targetType: 'users',
          targetIds: [userId],
          isPopup: false,
          needConfirm: false,
        },
      },
    );
    expect((await extra.json()).code).toBe(0);
  }

  const createNotice = await adminApi.post(
    `${apiBase}/api/v1/sys/notice/createNotice`,
    {
      data: {
        title: noticeTitle,
        content: 'E2E directed notice payload',
        level: 'info',
        targetType: 'users',
        targetIds: [userId],
        isPopup: false,
        needConfirm: true,
      },
    },
  );
  const createNoticeBody = await createNotice.json();
  expect(createNoticeBody.code).toBe(0);

  await page.goto('/#/user/login');
  await page.getByRole('textbox', { name: /username/i }).fill(username);
  await page.locator('input[type="password"]').fill(password);
  await page.getByRole('button', { name: /login/i }).click();

  await page.waitForURL(/#\//, { timeout: 20_000 });

  await expect(page.getByText('我的通知')).toBeVisible();
  await expect(page.getByText(noticeTitle)).toBeVisible();
  const inbox = page.getByRole('region', { name: '我的通知', exact: true });
  await expect(inbox.getByText('总数 6', { exact: true })).toBeVisible();
  await expect(inbox.getByText('本页 5 条未读 · 未读优先')).toBeVisible();
  const shortcuts = page.getByRole('navigation', { name: '工作台快捷入口' });
  await expect(shortcuts.getByRole('link', { name: '用户管理' })).toHaveCount(
    0,
  );
  await expect(shortcuts.getByRole('link', { name: '角色管理' })).toHaveCount(
    0,
  );
  await expect(shortcuts.getByRole('link', { name: '系统状态' })).toBeVisible();
  await expect(
    shortcuts.getByRole('link', { name: '关于', exact: true }),
  ).toBeVisible();
  await page.screenshot({
    path: testInfo.outputPath('workplace-notices.png'),
    fullPage: true,
  });
  await inbox
    .getByRole('article', { name: noticeTitle, exact: true })
    .getByRole('button', { name: '查看详情' })
    .click();
  const dialog = page.getByRole('dialog');
  await expect(dialog.getByText('E2E directed notice payload')).toBeVisible();
  // A failed mutation must leave the notice available for retry.
  await page.route('**/sys/notice/markRead', (route) =>
    route.fulfill({ json: { code: 1000, msg: 'temporary failure' } }),
  );
  await dialog.getByRole('button', { name: /确认已读/ }).click();
  await expect(page.getByText('标记已读失败，请重试')).toBeVisible();
  await expect(dialog).toBeVisible();
  await page.unroute('**/sys/notice/markRead');
  await dialog.getByRole('button', { name: /确认已读/ }).click();
  await expect(dialog).not.toBeVisible();
  await expect(inbox.getByText('总数 6', { exact: true })).toBeVisible();
  await inbox.locator('.ant-pagination-item-2').click();
  await expect(inbox.getByText('本页 0 条未读 · 未读优先')).toBeVisible();
  await expect(
    inbox.getByRole('article', { name: noticeTitle, exact: true }),
  ).toContainText('已读');
  await page.reload();
  await expect(inbox.getByText('总数 6', { exact: true })).toBeVisible();
  await inbox.locator('.ant-pagination-item-2').click();
  await inbox
    .getByRole('article', { name: noticeTitle, exact: true })
    .getByRole('button', { name: '查看详情' })
    .click();
  await expect(dialog.getByRole('button', { name: /确认已读/ })).toHaveCount(0);
  await expect(dialog.getByText('已读', { exact: true })).toBeVisible();

  await adminApi.dispose();
  await bootstrapApi.dispose();
});
