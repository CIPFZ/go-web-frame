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

test('admin notice list and user workplace visibility', async ({ page }) => {
  const apiBase = process.env.E2E_API_BASE || 'http://127.0.0.1:8080';
  const bootstrapApi = await request.newContext();

  const adminToken = await apiLogin(bootstrapApi, apiBase, ADMIN_USER, ADMIN_PASS);
  const adminApi = await request.newContext({
    extraHTTPHeaders: { 'x-token': adminToken },
  });

  const stamp = Date.now();
  const username = `e2e_user_b_${stamp}`;
  const password = 'E2E@123456';
  const noticeTitle = `E2E Admin Notice ${stamp}`;

  const addUser = await adminApi.post(`${apiBase}/api/v1/sys/user/addUser`, {
    data: {
      username,
      password,
      nickName: 'E2E User B',
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
      content: 'E2E admin notice list visibility payload',
      level: 'warning',
      targetType: 'users',
      targetIds: [userId],
      isPopup: true,
      needConfirm: true,
    },
  });
  const createNoticeBody = await createNotice.json();
  expect(createNoticeBody.code).toBe(0);

  await page.goto('/#/user/login');
  await page.getByRole('textbox', { name: /username/i }).fill(ADMIN_USER);
  await page.locator('input[type="password"]').fill(ADMIN_PASS);
  await page.getByRole('button', { name: /login/i }).click();
  await page.waitForURL(/#\//, { timeout: 20_000 });

  await page.getByRole('link', { name: /通知公告|Notices/i }).first().click();
  await page.waitForURL(/#\/sys\/notice/, { timeout: 20_000 });
  await expect(page.getByText('通知公告')).toBeVisible();
  await expect(page.getByRole('button', { name: /发布通知/i })).toBeVisible();

  await page.goto('/#/user/login');
  await page.getByRole('textbox', { name: /username/i }).fill(username);
  await page.locator('input[type="password"]').fill(password);
  await page.getByRole('button', { name: /login/i }).click();
  await page.waitForURL(/#\//, { timeout: 20_000 });

  await expect(page.getByText('我的通知')).toBeVisible();
  await expect(page.getByText(noticeTitle)).toBeVisible();

  await adminApi.dispose();
  await bootstrapApi.dispose();
});
