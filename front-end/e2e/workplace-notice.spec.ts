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

test('desk respects role menus and supports notice paging and confirmation', async ({
  page,
}, testInfo) => {
  test.setTimeout(150_000);
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
  await page.getByRole('textbox', { name: '用户名' }).fill(username);
  await page.locator('input[type="password"]').fill(password);
  const submitLogin = async () => {
    const pending = page.waitForResponse(response => response.url().endsWith('/api/v1/user/login') && response.request().method() === 'POST');
    await page.getByRole('button', { name: /登\s*录/ }).click();
    return pending;
  };
  let response = await submitLogin();
  if (response.status() === 429) {
    // The suite shares an IP; respect the real server's login quota.
    const seconds = Number(response.headers()['retry-after'] || 60);
    await page.waitForTimeout(Math.min(seconds, 60) * 1000 + 100);
    response = await submitLogin();
  }
  expect((await response.json()).code).toBe(0);
  await page.waitForURL(/#\/dashboard\/workplace/, { timeout: 20_000 });

  const inbox = page.getByRole('region', { name: '我的通知', exact: true });
  await expect(inbox.getByText(noticeTitle)).toBeVisible();
  // 普通用户的桌面没有系统管理的图标
  await expect(page.getByRole('link', { name: '用户管理' })).toHaveCount(0);
  await expect(page.getByRole('link', { name: '角色管理' })).toHaveCount(0);
  await page.screenshot({
    path: testInfo.outputPath('workplace-notices.png'),
    fullPage: true,
  });
  await inbox.getByRole('button', { name: noticeTitle }).click();
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

  // 大号的通知可以翻页：六条通知分两页；尺寸保存在本人账号上
  await page.getByRole('button', { name: '编辑工作台' }).click();
  await page
    .getByRole('group', { name: '“我的通知”的尺寸' })
    .getByText('大', { exact: true })
    .click();
  await page.getByRole('button', { name: /完\s*成/ }).click();
  // 未读在前：五条未读在第一页，已读的那条在第二页
  const pager = inbox.locator('.ant-pagination');
  await expect(pager).toBeVisible();
  await expect(inbox.getByRole('button', { name: /Earlier notice/ })).toHaveCount(5);
  await pager.locator('.ant-pagination-next').click();
  await expect(inbox.getByRole('button', { name: noticeTitle })).toBeVisible();
  await page.reload();
  await expect(pager).toBeVisible();
  await pager.locator('.ant-pagination-next').click();
  await inbox.getByRole('button', { name: noticeTitle }).click();
  await expect(dialog.getByRole('button', { name: /确认已读/ })).toHaveCount(0);
  await expect(dialog.getByText('已读', { exact: true })).toBeVisible();

  await adminApi.dispose();
  await bootstrapApi.dispose();
});
