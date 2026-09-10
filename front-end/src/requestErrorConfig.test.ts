import { setLocale } from '@umijs/max';
import { errorConfig } from './requestErrorConfig';

test('sends the selected language on first-party requests, including login', () => {
  const intercept = errorConfig.requestInterceptors![0] as (config: any) => any;
  localStorage.setItem('token', 'session');
  setLocale('en-US', false);
  expect(intercept({ url: '/api/v1/user/login' }).headers).toEqual({
    'Accept-Language': 'en-US',
  });
  expect(intercept({ url: '/api/v1/sys/menu/getMenu' }).headers).toEqual({
    'Accept-Language': 'en-US',
    'x-token': 'session',
  });
  const outside = { url: 'https://example.com/api/v1/sys/menu/getMenu' };
  expect(intercept(outside)).toEqual(outside);
  expect(intercept(outside).headers).toBeUndefined();
  setLocale('zh-CN', false);
  expect(
    intercept({ url: '/api/v1/user/login' }).headers['Accept-Language'],
  ).toBe('zh-CN');
});
