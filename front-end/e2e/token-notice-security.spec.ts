import { expect, test, type APIRequestContext } from '@playwright/test';

test.skip(
  process.env.E2E_ISOLATED !== '1',
  'Requires the disposable CMS database.',
);
test.describe.configure({ timeout: 150_000 });
const password = 'Security@123456';
const stamp = () => `${Date.now()}_${Math.random().toString(36).slice(2, 6)}`;
async function login(
  request: APIRequestContext,
  username = 'admin',
  pass = process.env.E2E_ADMIN_PASS!,
) {
  const submit = () => request.post('/api/v1/user/login', { data: { username, password: pass } });
  let response = await submit();
  if (response.status() === 429) {
    // Test logins share one IP. Keep production throttling active and honor it.
    expect((await response.json()).code).toBe(7);
    const seconds = Number(response.headers()['retry-after'] || 60);
    await new Promise(resolve => setTimeout(resolve, Math.min(seconds, 60) * 1000 + 100));
    response = await submit();
  }
  const body = await response.json();
  expect(body.code).toBe(0);
  return body.data.token as string;
}
function client(request: APIRequestContext, token: string) {
  return async (path: string, data?: any, method = 'POST') =>
    await (
      await request.fetch(`/api/v1${path}`, {
        method,
        headers: { 'x-token': token },
        ...(data ? { data } : {}),
      })
    ).json();
}
async function addUser(
  api: ReturnType<typeof client>,
  roles: number[],
  status = 1,
) {
  const username = `security_${stamp()}`;
  expect(
    (
      await api('/sys/user/addUser', {
        username,
        nickName: username,
        password,
        authorityIds: roles,
        status,
      })
    ).code,
  ).toBe(0);
  const users = await api('/sys/user/getUserList', {
    username,
    page: 1,
    pageSize: 10,
  });
  return { username, id: users.data.list[0].ID as number };
}

test('API Token real endpoint, CMS separation, rejection and lifecycle', async ({
  request,
}) => {
  const admin = await login(request);
  const api = client(request, admin);
  const options = await api('/sys/api-token/options', {
    page: 1,
    pageSize: 100,
  });
  expect(options.code).toBe(0);
  expect(options.data.list).toHaveLength(1);
  const endpoint = options.data.list[0];
  expect(endpoint.path).toBe('/api/v1/open/token-info');
  const payload = {
    name: `security-${stamp()}`,
    expiresAt: new Date(Date.now() + 3600000).toISOString(),
    maxConcurrency: 1,
    apiIds: [endpoint.ID],
  };
  const all = await api('/sys/api/getApiList', { page: 1, pageSize: 100 });
  const cms = all.data.list.find((item: any) =>
    item.path.endsWith('/sys/user/getUserList'),
  );
  expect(
    (await api('/sys/api-token/create', { ...payload, apiIds: [cms.ID] })).code,
  ).not.toBe(0);
  for (const invalid of [
    { expiresAt: undefined },
    { neverExpire: true },
    { name: '   ' },
    { maxConcurrency: -1 },
    { apiIds: [999999] },
  ]) {
    expect(
      (await api('/sys/api-token/create', { ...payload, ...invalid })).code,
    ).not.toBe(0);
  }
  const created = await api('/sys/api-token/create', payload);
  expect(created.code).toBe(0);
  const { token, ID: id } = created.data;
  const call = async (secret: string, path = endpoint.path) =>
    await (
      await request.get(path, {
        headers: { 'X-API-Token': secret, Cookie: '' },
      })
    ).json();
  expect((await call(token)).data.tokenId).toBe(id);
  expect((await call('wrong')).code).toBe(1003);
  expect((await call(token, '/api/v1/sys/user/getSelfInfo')).code).toBe(1003);
  expect(
    (
      await (
        await request.get(endpoint.path, { headers: { 'x-token': admin } })
      ).json()
    ).code,
  ).toBe(1003);
  expect(
    (
      await request.post(endpoint.path, { headers: { 'X-API-Token': token } })
    ).status(),
  ).toBe(404);
  const detail = await api(`/sys/api-token/detail?id=${id}`, undefined, 'GET');
  expect(detail.data.token).toBeUndefined();
  expect(detail.data.tokenHash).toBeUndefined();
  expect(detail.data.lastUsedAt).toBeTruthy();
  expect((await api('/sys/api-token/disable', { id })).code).toBe(0);
  expect((await call(token)).code).toBe(1003);
  expect((await api('/sys/api-token/enable', { id })).code).toBe(0);
  expect((await call(token)).code).toBe(0);
  const reset = await api('/sys/api-token/reset', { id });
  expect(reset.code).toBe(0);
  expect((await call(token)).code).toBe(1003);
  expect((await call(reset.data.token)).code).toBe(0);
  expect((await api('/sys/api-token/delete', { id }, 'DELETE')).code).toBe(0);
  expect((await call(reset.data.token)).code).toBe(1003);
  expect((await api('/sys/api-token/enable', { id })).code).not.toBe(0);
  const reader = await addUser(api, [888]);
  const readerApi = client(
    request,
    await login(request, reader.username, password),
  );
  expect((await readerApi('/sys/api-token/create', payload)).code).toBe(1004);
  expect((await readerApi('/sys/api-token/getApiTokenList', {})).code).toBe(
    1004,
  );
  const badEn = await (
    await request.post('/api/v1/sys/api-token/create', {
      headers: { 'x-token': admin, 'Accept-Language': 'en-US' },
      data: { ...payload, apiIds: [cms.ID] },
    })
  ).json();
  expect(badEn.msg).toContain('does not support API Token');
});

test('notices deliver a role snapshot; recipient, validity, popup and confirmation are enforced', async ({
  page,
  request,
}) => {
  const api = client(request, await login(request));
  const roleA = 100000 + Math.floor(Math.random() * 1000000),
    roleB = roleA + 1;
  for (const id of [roleA, roleB])
    expect(
      (
        await api('/sys/authority/createAuthority', {
          authorityId: id,
          authorityName: `Security ${id}`,
          parentId: 0,
          defaultRouter: 'dashboard/workplace',
        })
      ).code,
    ).toBe(0);
  const reader = await addUser(api, [888, roleA, roleB]);
  const outsider = await addUser(api, [888]);
  const disabled = await addUser(api, [roleA], 0);
  const deleted = await addUser(api, [roleA]);
  expect(
    (await api('/sys/user/deleteUser', { id: deleted.id }, 'DELETE')).code,
  ).toBe(0);
  const title = `Private role notice ${stamp()}`;
  const payload = {
    title,
    content: '<script>window.noticeInjected=true</script> Private role content',
    level: 'warning',
    targetType: 'roles',
    targetIds: [roleA, roleB, roleA],
    isPopup: true,
    needConfirm: true,
  };
  expect((await api('/sys/notice/createNotice', payload)).code).toBe(0);
  const list = await api('/sys/notice/getNoticeList', {
    title,
    page: 1,
    pageSize: 100,
  });
  expect(list.code).toBe(0);
  const notice = list.data.list[0];
  expect(notice.receiverCount).toBe(1);
  expect(notice.targetIds).toEqual([roleA, roleB]);
  const userToken = await login(request, reader.username, password);
  const userApi = client(request, userToken);
  const outsiderApi = client(
    request,
    await login(request, outsider.username, password),
  );
  const inbox = await outsiderApi(
    '/sys/notice/getMyNotices?page=1&pageSize=100',
    undefined,
    'GET',
  );
  expect(inbox.data.list?.some((n: any) => n.ID === notice.ID) ?? false).toBe(
    false,
  );
  expect(
    (
      await outsiderApi('/sys/notice/markRead', {
        noticeId: notice.ID,
        userId: reader.id,
      })
    ).code,
  ).not.toBe(0);
  expect((await userApi('/sys/notice/getNoticeList', {})).code).toBe(1004);
  expect((await userApi('/sys/notice/createNotice', payload)).code).toBe(1004);
  const futureTitle = `Future ${stamp()}`;
  expect(
    (
      await api('/sys/notice/createNotice', {
        ...payload,
        title: futureTitle,
        startAt: new Date(Date.now() + 3600000).toISOString(),
      })
    ).code,
  ).toBe(0);
  const future = (
    await api('/sys/notice/getNoticeList', { title: futureTitle })
  ).data.list[0];
  expect(
    (await userApi('/sys/notice/markRead', { noticeId: future.ID })).code,
  ).not.toBe(0);
  for (const invalid of [
    { targetType: 'users', targetIds: [reader.id, disabled.id] },
    { targetIds: [roleA, 999999] },
    { title: '  ' },
    { endAt: new Date(Date.now() - 1000).toISOString() },
  ])
    expect(
      (await api('/sys/notice/createNotice', { ...payload, ...invalid })).code,
    ).not.toBe(0);
  const normalized = await userApi(
    '/sys/notice/getMyNotices?page=1&pageSize=999',
    undefined,
    'GET',
  );
  expect(normalized.data.pageSize).toBe(100);
  expect(
    (await userApi('/sys/notice/getMyNotices?page=bad', undefined, 'GET')).code,
  ).not.toBe(0);
  // Membership changes revoke sessions; use fresh logins to verify notice snapshots.
  expect(
    (
      await api(
        '/sys/user/updateUser',
        { id: reader.id, authorityIds: [888], status: 1 },
        'PUT',
      )
    ).code,
  ).toBe(0);
  expect(
    (
      await api(
        '/sys/user/updateUser',
        { id: outsider.id, authorityIds: [888, roleA], status: 1 },
        'PUT',
      )
    ).code,
  ).toBe(0);
  const newReaderToken = await login(request, reader.username, password);
  const refreshed = client(request, newReaderToken);
  const newOutsider = client(
    request,
    await login(request, outsider.username, password),
  );
  expect(
    (
      await refreshed(
        '/sys/notice/getMyNotices?popupOnly=true',
        undefined,
        'GET',
      )
    ).data.list.map((n: any) => n.ID),
  ).toEqual([notice.ID]);
  expect(
    (
      await newOutsider('/sys/notice/getMyNotices', undefined, 'GET')
    ).data.list?.some((n: any) => n.ID === notice.ID) ?? false,
  ).toBe(false);
  await page.goto('/#/user/login');
  await page.evaluate(
    (token) => localStorage.setItem('token', token),
    newReaderToken,
  );
  await page.goto('/#/dashboard/workplace');
  await page.reload();
  const modal = page.getByRole('dialog', { name: title });
  await expect(modal).toBeVisible();
  expect(
    await page.evaluate(() => (window as any).noticeInjected),
  ).toBeUndefined();
  await modal.getByRole('button', { name: /关\s*闭/, exact: true }).click();
  await expect(modal).not.toBeVisible();
  expect(
    (
      await refreshed(
        '/sys/notice/getMyNotices?popupOnly=true',
        undefined,
        'GET',
      )
    ).data.total,
  ).toBe(1);
  await page.reload();
  await expect(modal).toBeVisible();
  await modal.getByRole('button', { name: '确认已读', exact: true }).click();
  await expect(modal).not.toBeVisible();
  const read = (
    await refreshed('/sys/notice/getMyNotices', undefined, 'GET')
  ).data.list.find((n: any) => n.ID === notice.ID).readAt;
  expect(read).toBeTruthy();
  expect(
    (await refreshed('/sys/notice/markRead', { noticeId: notice.ID })).code,
  ).toBe(0);
  expect(
    (
      await refreshed('/sys/notice/getMyNotices', undefined, 'GET')
    ).data.list.find((n: any) => n.ID === notice.ID).readAt,
  ).toBe(read);
  expect(
    (await api('/sys/notice/getNoticeList', { title })).data.list[0].readCount,
  ).toBe(1);
  await page.reload();
  await expect(
    page.getByRole('region', { name: '我的通知', exact: true }),
  ).toBeVisible();
  await expect(modal).not.toBeVisible();
});

test('notice publishing UI clears targets, resets on reopen and sends ISO scheduling dates', async ({
  page,
  request,
}) => {
  const api = client(request, await login(request));
  const recipient = await addUser(api, [888]);
  const token = await login(request);
  await page.goto('/#/user/login');
  await page.evaluate((value) => localStorage.setItem('token', value), token);
  await page.goto('/#/sys/notice');
  await page.reload();
  await page.getByRole('button', { name: /发布通知/ }).click();
  const modal = page.getByRole('dialog');
  await modal
    .locator('.ant-select:has(#targetType) .ant-select-selector')
    .click();
  await page.getByText('按角色', { exact: true }).last().click();
  await modal.locator('#targetIds').click();
  await page.getByTitle('CommonUser', { exact: true }).click();
  await modal.locator('#title').click();
  await modal
    .locator('.ant-select:has(#targetType) .ant-select-selector')
    .click();
  await page.getByText('指定用户', { exact: true }).last().click();
  await expect(
    modal
      .locator('.ant-select-selection-item')
      .filter({ hasText: 'CommonUser' }),
  ).toHaveCount(0);
  await modal.getByRole('button', { name: /取\s*消/ }).click();
  await page.getByRole('button', { name: /发布通知/ }).click();
  await expect(modal.locator('#targetIds')).toHaveCount(0);
  const title = `UI scheduled ${stamp()}`;
  await modal.locator('#title').fill(title);
  await modal.locator('#content').fill('Scheduled browser notice');
  await modal
    .locator('.ant-select:has(#targetType) .ant-select-selector')
    .click();
  await page.getByText('指定用户', { exact: true }).last().click();
  await modal.locator('#targetIds').click();
  await page
    .getByTitle(`${recipient.username} (${recipient.username})`, {
      exact: true,
    })
    .click();
  await modal.locator('#title').click();
  const tomorrow = new Date(Date.now() + 86400000);
  const date = `${tomorrow.getFullYear()}-${String(tomorrow.getMonth() + 1).padStart(2, '0')}-${String(tomorrow.getDate()).padStart(2, '0')} 12:00:00`;
  await modal.locator('#startAt').fill(date);
  await modal.locator('#startAt').press('Enter');
  await modal.locator('#title').click();
  const sent = page.waitForRequest('**/sys/notice/createNotice');
  const result = page.waitForResponse('**/sys/notice/createNotice');
  await modal.getByRole('button', { name: /确\s*定/ }).click();
  const payload = (await sent).postDataJSON();
  expect(payload.startAt).toMatch(/T.*Z$/);
  expect(payload.targetIds).toEqual([recipient.id]);
  expect((await (await result).json()).code).toBe(0);
  await expect(modal).not.toBeVisible();
  await expect(page.getByText(title, { exact: true })).toBeVisible();
});
