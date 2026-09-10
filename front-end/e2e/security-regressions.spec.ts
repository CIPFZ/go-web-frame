import { test, expect } from '@playwright/test';

test('registration is closed in both UI and API; status page is a dependency overview', async ({ page, request }, testInfo) => {
  const config = await (await request.get('/api/v1/public/config')).json();
  expect(config.data.registrationEnabled).toBe(false);
  await page.goto('/#/user/register');
  await expect(page.getByText('注册已关闭，请联系管理员创建账号')).toBeVisible();
  expect((await (await request.post('/api/v1/user/register', {data: {username: 'closed-registration', password: 'password123'}})).json()).code).not.toBe(0);
  const login = await (await request.post('/api/v1/user/login', {data: {username: 'admin', password: process.env.E2E_ADMIN_PASS}})).json();
  expect(login.code).toBe(0);
  await page.evaluate(token => localStorage.setItem('token', token), login.data.token);
  await page.goto('/#/state');
  await page.reload();
  await expect(page.getByText('服务可用性', {exact: true})).toBeVisible();
  await expect(page.getByText('平台信息', {exact: true})).toBeVisible();
  await expect(page.getByText('观测平台未启用', {exact: true})).toBeVisible();
  await expect(page.getByRole('status').filter({hasText: '平台运行正常'})).toBeVisible();
  await page.screenshot({path: testInfo.outputPath('status-desktop.png'), fullPage: true});
  await page.setViewportSize({width: 390, height: 844});
  await expect(page.getByRole('article', {name: '数据库：正常'})).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  await page.screenshot({path: testInfo.outputPath('status-mobile.png'), fullPage: true});
  await page.route('**/sys/system/getServerInfo', async route => {
    const response = await route.fetch(); const body = await response.json();
    body.data.server.checks.database = 'unavailable';
    await route.fulfill({response, json: body});
  });
  await page.getByRole('button', {name: '刷新状态'}).click();
  await expect(page.getByText('部分服务需要关注', {exact: true})).toBeVisible();
  await expect(page.getByRole('article', {name: '数据库：不可用'})).toBeVisible();
  await page.unroute('**/sys/system/getServerInfo');
  await page.setViewportSize({width: 1440, height: 1000});
  await page.route('**/sys/user/getSelfInfo', async route => {
    const response = await route.fetch(); const body = await response.json();
    body.data.settings = {...body.data.settings, navTheme: 'realDark'};
    await route.fulfill({response, json: body});
  });
  await page.reload();
  await expect(page.getByText('平台运行正常', {exact: true})).toBeVisible();
  await page.screenshot({path: testInfo.outputPath('status-dark.png'), fullPage: true});


});

test('logout revokes one login and password reset revokes all remaining logins', async ({ request }) => {
  test.skip(process.env.E2E_ISOLATED !== '1', 'Requires disposable database');
  const login = async (username: string, password: string) => {
    const res = await (await request.post('/api/v1/user/login', {data: {username,password}})).json();
    expect(res.code).toBe(0); return res.data.token as string;
  };
  const admin = await login('admin', process.env.E2E_ADMIN_PASS!);
  const adminHeaders = {'x-token': admin};
  const username = `session_${Date.now()}`;
  const created = await (await request.post('/api/v1/sys/user/addUser', {headers: adminHeaders, data: {username,password:'password123',authorityIds:[888],status:1}})).json();
  expect(created.code).toBe(0);
  const users = await (await request.post('/api/v1/sys/user/getUserList', {headers:adminHeaders,data:{username,page:1,pageSize:10}})).json();
  const id = users.data.list[0].ID;
  const a = await login(username,'password123'); const b = await login(username,'password123');
  const self = async (token: string) => (await (await request.get('/api/v1/sys/user/getSelfInfo',{headers:{'x-token':token}})).json()).code;
  expect(await self(a)).toBe(0); expect(await self(b)).toBe(0);
  expect((await (await request.post('/api/v1/sys/user/logout',{headers:{'x-token':a}})).json()).code).toBe(0);
  expect(await self(a)).toBe(1003); expect(await self(b)).toBe(0);
  expect((await (await request.post('/api/v1/sys/user/resetPassword',{headers:adminHeaders,data:{id,password:'changed-password'}})).json()).code).toBe(0);
  expect(await self(b)).toBe(1003);
  await request.delete('/api/v1/sys/user/deleteUser',{headers:adminHeaders,data:{id}});
});
