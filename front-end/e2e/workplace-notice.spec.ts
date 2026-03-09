import { APIRequestContext, expect, request, test } from '@playwright/test';

const ADMIN_USER = process.env.E2E_ADMIN_USER || 'admin';
const ADMIN_PASS = process.env.E2E_ADMIN_PASS || 'Admin@123456';

async function apiLogin(api: APIRequestContext, apiBase: string, username: string, password: string): Promise<string> {
  const resp = await api.post(`${apiBase}/api/v1/user/login`, {
    data: { username, password },
  });
  const body = await resp.json();
  expect(body.code).toBe(0);
  return body.data.token;
}

test('login and render workplace notice list', async ({ page }) => {
  const apiBase = process.env.E2E_API_BASE || 'http://127.0.0.1:8080';
  const bootstrapApi = await request.newContext();

  const adminToken = await apiLogin(bootstrapApi, apiBase, ADMIN_USER, ADMIN_PASS);
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

  const userList = await adminApi.post(`${apiBase}/api/v1/sys/user/getUserList`, {
    data: { page: 1, pageSize: 20, username },
  });
  const userListBody = await userList.json();
  expect(userListBody.code).toBe(0);
  const userId = userListBody.data?.list?.[0]?.ID;
  expect(userId).toBeTruthy();

  const createNotice = await adminApi.post(`${apiBase}/api/v1/sys/notice/createNotice`, {
    data: {
      title: noticeTitle,
      content: 'E2E directed notice payload',
      level: 'info',
      targetType: 'users',
      targetIds: [userId],
      isPopup: false,
      needConfirm: true,
    },
  });
  const createNoticeBody = await createNotice.json();
  expect(createNoticeBody.code).toBe(0);

  await page.goto('/#/user/login');
  await page.getByRole('textbox', { name: /username/i }).fill(username);
  await page.locator('input[type="password"]').fill(password);
  await page.getByRole('button', { name: /login/i }).click();

  await page.waitForURL(/#\//, { timeout: 20_000 });

  await expect(page.getByText('我的通知')).toBeVisible();
  await expect(page.getByText(noticeTitle)).toBeVisible();
  await expect(page.getByText('总数')).toBeVisible();

  await adminApi.dispose();
  await bootstrapApi.dispose();
});
